package main

import "fmt"

// checkProjecLocally will verify if the project/namespace exists in the CLI context.
// Won't fecth the cluster to verify if the project/namespace exists. This is a local validation only.
func checkProjecLocally(project string) (localProject string, err error) {
	if len(project) == 0 {
		if len(config.Namespace) == 0 {
			return "", fmt.Errorf("Couldn't find any Project in the current context. Use 'kogito use-project NAME' to set the Kogito Project where the service will be deployed or pass '--project NAME' flag to this one")
		}
		return config.Namespace, nil
	}
	return project, nil
}
