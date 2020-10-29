package controller

import (
	"openafs-csi-operator/pkg/controller/openafscsiapp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, openafscsiapp.Add)
}
