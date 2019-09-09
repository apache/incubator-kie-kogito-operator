package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

type deleteServiceFlags struct {
	name    string
	project string
}

var deleteServiceCommand *cobra.Command
var deleteServiceCmdFlags = deleteServiceFlags{}

var _ = RegisterCommandVar(func() {
	deleteServiceCommand = &cobra.Command{
		Example: "delete-service example-drools --project kogito",
		Use:     "delete-service NAME [flags]",
		Short:   "Deletes a Kogito Runtime Service deployed in the namespace/project",
		Long:    `delete-service will exclude every OpenShift/Kubernetes resource created to deploy the Kogito Runtime Service into the namespace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteServiceExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(deleteServiceCommand)
	deleteServiceCommand.Flags().StringVarP(&deleteServiceCmdFlags.project, "project", "p", "", "The project name")
})

func deleteServiceExec(cmd *cobra.Command, args []string) error {
	deleteServiceCmdFlags.name = args[0]
	var err error
	if deleteServiceCmdFlags.project, err = checkProjecLocally(deleteServiceCmdFlags.project); err != nil {
		return err
	}
	if err := checkProjectExists(deleteServiceCmdFlags.project); err != nil {
		return err
	}
	log.Debugf("Using project %s", deleteServiceCmdFlags.project)

	if err := checkKogitoAppExists(deleteServiceCmdFlags.name, deleteServiceCmdFlags.project); err != nil {
		return err
	}
	log.Debugf("About to delete service %s in namespace %s", deleteServiceCmdFlags.name, deleteServiceCmdFlags.project)

	if err := kubernetes.ResourceC(kubeCli).Delete(&v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deleteServiceCmdFlags.name,
			Namespace: deleteServiceCmdFlags.project,
		},
	}); err != nil {
		return err
	}

	log.Infof("Successfully deleted Kogito Service %s in the Project %s", deleteServiceCmdFlags.name, deleteServiceCmdFlags.project)

	return nil
}
