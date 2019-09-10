package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

type useProjectFlags struct {
	name string
}

var useProjectCmd *cobra.Command
var useProjectCmdFlags = useProjectFlags{}

var _ = RegisterCommandVar(func() {
	useProjectCmd = &cobra.Command{
		Use:     "use-project NAME",
		Aliases: []string{"use-ns"},
		Short:   "Sets the Kogito Project where your Kogito Service will be deployed",
		Long:    `use-project will set the Kubernetes Namespace where the Kogito services will be deployed. It's the Namespace/Project on Kubernetes/OpenShift world.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(useProjectCmdFlags.name) == 0 {
				if len(args) == 0 {
					if len(config.Namespace) == 0 {
						return fmt.Errorf("No Project set in the context. Use 'kogito new-project NAME' to create a new Project")
					}
					log.Debugf("Project in the context is '%s'. Use 'kogito deploy-service NAME SOURCE' to deploy a new Kogito Service.", config.Namespace)
					useProjectCmdFlags.name = config.Namespace
					return nil
				}
				useProjectCmdFlags.name = args[0]
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(useProjectCmd)
	useProjectCmd.Flags().StringVarP(&useProjectCmdFlags.name, "name", "n", "", "The Project name")
})

func appExec(cmd *cobra.Command, args []string) error {
	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(useProjectCmdFlags.name); err != nil {
		return fmt.Errorf("Error while trying to look for the namespace. Are you logged in? %s", err)
	} else if ns != nil {
		config.Namespace = useProjectCmdFlags.name
		log.Infof("Project set to '%s'", useProjectCmdFlags.name)
		return nil
	}

	return fmt.Errorf("Project '%s' not found. Try running 'kogito new-project %s' to create your Project first", useProjectCmdFlags.name, useProjectCmdFlags.name)
}
