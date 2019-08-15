package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/inventory"
	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	"github.com/spf13/cobra"
)

var newAppCmd *cobra.Command
var newAppName string

var _ = RegisterCommandVar(func() {
	newAppCmd = &cobra.Command{
		Use:   "new-app NAME",
		Short: "Creates a new namespace for your Kogito Services",
		Long:  `new-app will create a namespace with the provided name where your Kogito Services will be deployed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return newAppExec(cmd, args)
		},
		PostRun: saveConfiguration,
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(newAppCmd)
	newAppCmd.Flags().StringVarP(&newAppName, "name", "n", "", "The app name")
})

func newAppExec(cmd *cobra.Command, args []string) error {
	if len(newAppName) == 0 {
		if len(args) == 0 {
			return fmt.Errorf("Please set a name for new-app")
		}
		newAppName = args[0]
	}

	ns, err := inventory.NamespaceC(kubeCli).Fetch(newAppName)
	if err != nil {
		return err
	}
	if ns == nil {
		ns, err := inventory.NamespaceC(kubeCli).Create(newAppName)
		if err != nil {
			return err
		}
		config.Namespace = ns.Name
		log.Infof("Namespace '%s' created successfully", ns.Name)
	} else {
		log.Infof("Namespace '%s' already exists", newAppName)
	}
	return nil
}
