package main

import (
	"github.com/kiegroup/kogito-cloud-operator/version"
	"github.com/spf13/cobra"
)

var versionCmd *cobra.Command

var _ = RegisterCommandVar(func() {
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the kogito CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionExec(cmd, args)
		},
		PreRun: preRunF,
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(versionCmd)
})

func versionExec(cmd *cobra.Command, args []string) error {
	log.Infof("Kogito CLI version: %s", version.Version)
	return nil
}
