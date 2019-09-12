package main

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
	"regexp"
)

const (
	defaultDeployReplicas = 1
	// see: https://github.com/docker/distribution/blob/master/reference/regexp.go
	dockerTagRegx = `(?P<namespace>.+/)?(?P<image>[^:]+)(?P<tag>:.+)?`
)

var (
	dockerTagRegxCompiled = *regexp.MustCompile(dockerTagRegx)
)

type deployCommonFlags struct {
	project  string
	replicas int32
	env      []string
	limits   []string
	requests []string
}

func commonAddDeployFlags(command *cobra.Command, flags *deployCommonFlags) {
	command.Flags().StringVarP(&flags.project, "project", "p", "", "The project name where the service will be deployed")
	command.Flags().Int32Var(&flags.replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	command.Flags().StringSliceVarP(&flags.env, "env", "e", nil, "Key/Pair value environment variables that will be set to the service runtime. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.limits, "limits", nil, "Resource limits for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.requests, "requests", nil, "Resource requests for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
}

func commonCheckDeployArgs(flags *deployCommonFlags) error {
	if err := util.ParseStringsForKeyPair(flags.env); err != nil {
		return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", flags.env)
	}
	if err := util.ParseStringsForKeyPair(flags.limits); err != nil {
		return fmt.Errorf("limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.limits)
	}
	if err := util.ParseStringsForKeyPair(flags.requests); err != nil {
		return fmt.Errorf("requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.requests)
	}
	if flags.replicas <= 0 {
		return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", flags.replicas)
	}
	return nil
}

func commonCheckImageTag(image string) error {
	if len(image) > 0 && !dockerTagRegxCompiled.MatchString(image) {
		return fmt.Errorf("invalid name for image tag. Valid format is namespace/image-name:tag. Received %s", image)
	}
	return nil
}
