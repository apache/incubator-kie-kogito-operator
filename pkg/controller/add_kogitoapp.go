package controller

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kogitoapp.Add)
}
