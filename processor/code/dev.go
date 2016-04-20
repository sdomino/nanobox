package code

import (
	"fmt"

	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/nanobox-boxfile"

	"github.com/nanobox-io/golang-docker-client"
	"github.com/nanobox-io/nanobox/processor"
	"github.com/nanobox-io/nanobox/util"
	"github.com/nanobox-io/nanobox/util/ip_control"
)

type codeDev struct {
	config processor.ProcessConfig
}

func init() {
	processor.Register("code_dev", codeDevFunc)
}

func codeDevFunc(config processor.ProcessConfig) (processor.Processor, error) {
	// confirm the provider is an accessable one that we support.

	return &codeDev{config: config}, nil
}

func (self codeDev) Results() processor.ProcessConfig {
	return self.config
}

func (self *codeDev) Process() error {
	box := boxfile.NewFromPath(util.BoxfileLocation())
	image := box.Node("build").StringValue("image")

	if image == "" {
		image = "nanobox/build:v1"
	}

	_, err := docker.ImagePull(image)
	if err != nil {
		return err
	}

	// create build container
	localIp, err := ip_control.ReserveLocal()
	if err != nil {
		return err
	}

	// return ip
	defer ip_control.ReturnIP(localIp)

	appName := util.AppName()
	config := docker.ContainerConfig{
		Name:    fmt.Sprintf("%s-dev", util.AppName()),
		Image:   image, // this will need to be configurable some time
		Network: "virt",
		IP:      localIp.String(),
		Binds: []string{
			fmt.Sprintf("/share/%s/code:/code", appName),
			fmt.Sprintf("/mnt/%s/build:/data", appName),
			fmt.Sprintf("/mnt/%s/cache:/mnt/cache", appName),
		},
	}
	for _, lib_dir := range box.Node("code.build").StringSliceValue("lib_dirs") {
		config.Binds = append(config.Binds, fmt.Sprintf("/mnt/%s/cache/lib_dirs/%s:/code/%s", util.AppName(), lib_dir, lib_dir))
	}
	// add lib_dirs
	// fmt.Sprintf("/mnt/%s/cache/lib_dirs/vendor:/code/vendor", util.AppName()),

	// start container
	container, err := docker.CreateContainer(config)
	if err != nil {
		lumber.Error("container: ", err)
		return err
	}

	// shutdown container
	defer docker.ContainerRemove(container.ID)

	// run user hook
	output, err := util.Exec(container.ID, "user", util.UserPayload())
	if err != nil {
		fmt.Println("user", output)
		return err
	}

	// console into the dev container
	err = processor.Run("console", self.config)
	if err != nil {
		fmt.Println("unable to enter console", err)
	}

	return err
}	
