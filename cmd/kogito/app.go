package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/inventory"

	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	"github.com/spf13/cobra"
)

var appCmd *cobra.Command
var appName string

var _ = RegisterCommandVar(func() {
	appCmd = &cobra.Command{
		Use:   "app NAME",
		Short: "Sets the Kogito application where your application will be deployed",
		Long:  `app will set the context where the Kogito services will be deployed. It's the namespace/project on Kubernetes/OpenShift world.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appExec(cmd, args)
		},
		PostRun: saveConfiguration,
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(appCmd)
	appCmd.Flags().StringVarP(&appName, "name", "n", "", "The app name")
})

func appExec(cmd *cobra.Command, args []string) error {
	if len(newAppName) == 0 {
		if len(args) == 0 {
			if len(config.Namespace) == 0 {
				return fmt.Errorf("No application set in the context. Use 'kogito new-app NAME' to create a new application")
			}
			log.Infof("Application in the context is '%s'. Use 'kogito deploy SOURCE' to deploy a new application.", config.Namespace)
			return nil
		}
		appName = args[0]
	}

	if ns, err := inventory.NamespaceC(kubeCli).Fetch(appName); err != nil {
		return fmt.Errorf("Error while trying to look for the namespace. Are you logged in? %s", err)
	} else if ns != nil {
		config.Namespace = appName
		log.Infof("Application set to '%s'", appName)
		return nil
	}

	return fmt.Errorf("Application '%s' not found. Try running 'kogito new-app %s' to create your application first", appName, appName)
}
