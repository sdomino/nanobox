//
package commands

import (
	"fmt"
	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox/config"
	engineutil "github.com/nanobox-io/nanobox/util/engine"
	"github.com/spf13/cobra"
)

var (

	//
	devCmd = &cobra.Command{
		Use:   "dev",
		Short: "Starts the nanobox, provisions app, & opens an interactive terminal",
		Long:  ``,

		PreRun:  boot,
		Run:     dev,
		PostRun: halt,
	}

	//
	rebuild bool // force a deploy
	nobuild bool // force skip a deploy
)

//
func init() {
	devCmd.Flags().BoolVarP(&rebuild, "rebuild", "", false, "Force a rebuild")
	devCmd.Flags().BoolVarP(&nobuild, "no-build", "", false, "Force skip a rebuild")
}

// dev
func dev(ccmd *cobra.Command, args []string) {

	// PreRun: boot

	if !nobuild {

		// if the vm has no been created, deployed, or the rebuild flag is passed do
		// a deploy
		if Vagrant.Status() == "not created" || !config.VMfile.HasDeployed() || rebuild {

			fmt.Printf(stylish.Bullet("Deploying codebase..."))

			// remount the engine file at ~/.nanobox/apps/<app>/<engine> so any new scripts
			// will be used during the deploy
			if err := engineutil.RemountLocal(); err != nil {
				config.Error("[util/vagrant/init] engineutil.RemountLocal() failed", err.Error())
			}

			// run a deploy
			if err := Server.Deploy(""); err != nil {
				Config.Fatal("[commands/dev] server.Deploy() failed - ", err.Error())
			}

			// stream log output
			go Mist.Stream([]string{"log", "deploy"}, Mist.PrintLogStream)

			// listen for status updates
			errch := make(chan error)
			go func() {
				errch <- Mist.Listen([]string{"job", "deploy"}, Mist.DeployUpdates)
			}()

			// wait for a status update (blocking)
			err := <-errch

			//
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		}
	}

	//
	if err := Server.Exec("develop", ""); err != nil {
		config.Error("[commands/dev] Server.Exec failed", err.Error())
	}

	// PostRun: halt
}