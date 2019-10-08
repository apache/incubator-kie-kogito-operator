// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

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
	// DefaultConfigPath is the directory name for the kogito config files
	DefaultConfigPath = ".kogito"
	// DefaultConfigFile is the name of the kogito config file
	DefaultConfigFile = "config"
	// DefaultConfigExt is the default extension for the kogito config file
	DefaultConfigExt = "yaml"
	// DefaultConfigFinalName is the full URI for kogito config file
	DefaultConfigFinalName = DefaultConfigFile + "." + DefaultConfigExt
)

// Configuration is the struct for the configuration definition for the Kogito CLI application add all configuration needed to this struct
type Configuration struct {
	// Namespace is the projet/namespace context where the application will be deployed
	Namespace string
}

// ReadConfig will read the configuration from disk
func ReadConfig() Configuration {
	log := GetDefaultLogger()
	if rootCmd.ConfigFile() != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rootCmd.ConfigFile())
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// config viper
		fullPath := filepath.Join(home, DefaultConfigPath)
		viper.AddConfigPath(fullPath)
		viper.SetConfigName(DefaultConfigFile)
		viper.SetConfigType(DefaultConfigExt)
		// ensure file
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
					panic(fmt.Errorf("Error creating path for config file: %s ", err))
				} else {
					if err := viper.WriteConfigAs(filepath.Join(fullPath, DefaultConfigFinalName)); err != nil {
						panic(fmt.Errorf("Error while trying to write config file: %s ", err))
					}
				}
			} else {
				panic(fmt.Errorf("Error reading file: %s ", err))
			}
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		config := Configuration{}
		if err := viper.Unmarshal(&config); err != nil {
			panic(fmt.Errorf("Error while unmarshalling the config file: %s ", err))
		}
		log.Debug("Using config file:", viper.ConfigFileUsed())
		return config
	}
	return Configuration{}
}

// Save will write all configuration data back to the configuration file
func (c *Configuration) Save() {
	// TODO: keep an eye on viper to come up with a solution like viper.Marshal(&c)
	if b, err := yaml.Marshal(&c); err != nil {
		panic(fmt.Errorf("Error while marshalling config objects: %s ", err))
	} else {
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			panic(fmt.Errorf("Error while reading config file: %s ", err))
		}
		viper.WriteConfig()
	}
}
