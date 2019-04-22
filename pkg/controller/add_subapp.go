package controller

import (
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, subapp.Add)
}
