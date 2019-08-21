package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	verbose bool
)

var _ = RegisterCommandVar(func() {
	rootCmd = &cobra.Command{
		Use:    "kogito",
		Short:  "Kogito CLI",
		Long:   `Kogito CLI deploys your Kogito Services into an OpenShift cluster`,
		PreRun: preRunF,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		//	Run: func(cmd *cobra.Command, args []string) { },
	}
})

var _ = RegisterCommandInit(func() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kogito.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
})
