package main

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
)

// fromStringArrayToControllerEnvs converts a string array in the format of key=value pairs to the kogitoapp controller required type
func fromStringArrayToControllerEnvs(strings []string) []v1alpha1.Env {
	if strings == nil {
		return nil
	}
	envs := []v1alpha1.Env{}
	mapstr := util.FromStringsKeyPairToMap(strings)
	for k, v := range mapstr {
		envs = append(envs, v1alpha1.Env{Name: k, Value: v})
	}
	return envs
}

func fromStringArrayToControllerResourceMap(strings []string) []v1alpha1.ResourceMap {
	if strings == nil {
		return nil
	}
	res := []v1alpha1.ResourceMap{}
	mapstr := util.FromStringsKeyPairToMap(strings)
	for k, v := range mapstr {
		res = append(res, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceKind(k), Value: v})
	}
	return res
}
