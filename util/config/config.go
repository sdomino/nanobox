package config

import "github.com/spf13/viper"

//
var config *viper.Viper

// Get fetches a generic value from viper config
func Get(key string) interface{} {

	// parse the config file if it's not already
	if config == nil {
		if err := parseConfig(); err != nil {
			return nil
		}
	}

	return config.Get(key)
}

// Viper returns the viper config object
func Viper() *viper.Viper {

	// parse the config file if it's not already
	if config == nil {
		if err := parseConfig(); err != nil {
			return nil
		}
	}

	return config
}

// ParseConfig will load the config file and parse it with viper
func parseConfig() error {

	// initilize a viper parser
	config = viper.New()

	// set sensible defaults

	// odin access
	config.SetDefault("production_host", "api.nanobox.io/v1/")

	// network space
	config.SetDefault("external-network-space", "192.168.99.50/24")
	config.SetDefault("internal-network-space", "192.168.0.50/16")

	// configurable options (via ~.nanobox/config.yml); these defaults are set here
	// incase for some reason an incomplete or invalid config.yml is parsed, nanobox
	// will have values to fall back on
	config.SetDefault("provider", "docker_machine")
	config.SetDefault("vm.cpus", 2)
	config.SetDefault("vm.cpu-cap", 50)
	config.SetDefault("vm.ram", 1024)
	config.SetDefault("vm.mount", "native")

	// parse config file; we attempt to parse the config.yml and pull out any values
	// that the user has provided (or one is generated with defaults; see ./files.go)
	config.SetConfigFile(configFile())

	// merge with defaults
	if err := config.MergeInConfig(); err != nil {
		return err
	}

	return nil
}
