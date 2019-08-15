package main

import (
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

var _ = RegisterCommandVar(func() {
	rootCmd = &cobra.Command{
		Use:   "kogito",
		Short: "Kogito CLI",
		Long:  `The Kogito CLI application to deploy your Kogito Services into an OpenShift cluster`,
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
})
