package main

import (
	"fmt"
	"os"

	"github.com/kiegroup/kogito-cloud-operator/pkg/inventory"
	"github.com/spf13/cobra"
)

var (
	// holder for all configuration changes accross command executions
	config      = &configuration{}
	cfgFile     string
	varInitFncs []func()
	cmdInitFncs []func()
	// saveConfiguration is the default command post hook that persists all configuration set during the function execution
	saveConfiguration = func(cmd *cobra.Command, args []string) {
		config.save()
	}

	// used by unit tests is the kube client for communication with Kubernetes API
	kubeCli = &inventory.Client{}
)

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config.initConfig(cfgFile)
}

// RegisterCommandVar is used to register with kogito the initialization function
// for the command variable.
// Something must be returned to use the `var _ = ` trick.
func RegisterCommandVar(c func()) bool {
	varInitFncs = append(varInitFncs, c)
	return true
}

// RegisterCommandInit is used to register with kogito the initialization function
// for the command flags.
// Something must be returned to use the `var _ = ` trick.
func RegisterCommandInit(c func()) bool {
	cmdInitFncs = append(cmdInitFncs, c)
	return true
}

func registerCommands() {
	// Setup all variables.
	// Setting up all the variables first will allow kogito
	// to initialize the init functions in any order
	for _, v := range varInitFncs {
		v()
	}

	// Call all plugin inits
	for _, f := range cmdInitFncs {
		f()
	}
}

// Main starts the kogito cli
func Main() error {
	// register every command
	registerCommands()
	// Execute kogito
	return rootCmd.Execute()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := Main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
