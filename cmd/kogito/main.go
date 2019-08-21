package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	// holder for all configuration changes accross command executions
	config        = &configuration{}
	log           = logger.GetLogger("kogito_cli")
	cfgFile       string
	varInitFncs   []func()
	cmdInitFncs   []func()
	commandOutput io.Writer
	posRunF       = func(cmd *cobra.Command, args []string) {
		config.save()
	}
	preRunF = func(cmd *cobra.Command, args []string) {
		setDefaultLog("kogito_cli", commandOutput)
	}
	// used by unit tests is the kube client for communication with Kubernetes API
	kubeCli = &client.Client{}
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

// changes the default log
func setDefaultLog(name string, output io.Writer) {
	log = logger.GetLoggerWithOptions(name, &logger.Opts{Output: output, Verbose: verbose, Console: true})
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
