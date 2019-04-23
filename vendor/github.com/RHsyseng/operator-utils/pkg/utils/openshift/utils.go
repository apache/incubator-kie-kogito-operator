package openshift

import (
	"golang.org/x/text/language"
	"golang.org/x/text/search"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// IsOpenShift checks for the OpenShift API
func IsOpenShift(config *rest.Config) (bool, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, err
	}
	apiSchema, err := clientset.OpenAPISchema()
	if err != nil {
		return false, err
	}
	return stringSearch(apiSchema.GetInfo().Title, "openshift"), nil
}

func stringSearch(str string, substr string) bool {
	if start, _ := search.New(language.English, search.IgnoreCase).IndexString(str, substr); start == -1 {
		return false
	}
	return true
}
