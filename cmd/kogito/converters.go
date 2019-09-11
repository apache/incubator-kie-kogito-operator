package main

import (
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
)

// fromStringToImage will convert a plain string into a image
func fromStringToImage(imagetag string) v1alpha1.Image {
	image := v1alpha1.Image{}
	if len(imagetag) > 0 {
		if strings.Contains(imagetag, "/") {
			splitName := strings.Split(imagetag, "/")
			image.ImageStreamNamespace = splitName[0]
			imagetag = splitName[1]
		}
		if strings.Contains(imagetag, ":") {
			splitName := strings.Split(imagetag, ":")
			image.ImageStreamName = splitName[0]
			image.ImageStreamTag = splitName[1]
			return image
		}
		image.ImageStreamName = imagetag
	}
	return image
}

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

// extractResource will read a string array in the format memory=512M, cpu=1 and will return the value for a given kind
func extractResource(kind v1alpha1.ResourceKind, resources []string) string {
	for _, res := range resources {
		resKV := strings.Split(res, "=")
		if string(kind) == resKV[0] {
			return resKV[1]
		}
	}

	return ""
}
