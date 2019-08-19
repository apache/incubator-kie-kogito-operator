package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	defaultConfigPath      = ".kogito"
	defaultConfigFile      = "config"
	defaultConfigExt       = "yaml"
	defaultConfigFinalName = defaultConfigFile + "." + defaultConfigExt
)

// configuration is the struct for the configuration definition for the Kogito CLI application
// add all configuration needed to this struct
type configuration struct {
	// Namespace is the projet/namespace context where the application will be deployed
	Namespace string
}

func (c *configuration) initConfig(configFile string) {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// config viper
		fullPath := filepath.Join(home, defaultConfigPath)
		viper.AddConfigPath(fullPath)
		viper.SetConfigName(defaultConfigFile)
		viper.SetConfigType(defaultConfigExt)
		// ensure file
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
					log.Error("Error creating path for config file")
				} else {
					viper.WriteConfigAs(filepath.Join(fullPath, defaultConfigFinalName))
				}
			} else {
				log.Error("Error reading file")
			}
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		viper.Unmarshal(&config)
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}

// save will write all configuration data back to the configuration file
func (c *configuration) save() {
	// TODO: keep an eye on viper to come up with a solution like viper.Marshal(&c)
	if b, err := yaml.Marshal(&c); err != nil {
		log.Error("Error while marshalling config objects")
	} else {
		viper.ReadConfig(bytes.NewBuffer(b))
		viper.WriteConfig()
	}
}
