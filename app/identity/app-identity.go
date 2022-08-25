package app_identity

import (
	uuid "github.com/satori/go.uuid"
)

var ( // Application identities as usual are set under CI/CD and build system
	//Namespace k8s namespace
	Namespace string

	//Name application name
	Name string

	//Version application version
	Version string

	//BuildTS application build timestamp
	BuildTS string

	//BuildBranch GIT branch
	BuildBranch string

	//BuildHash GIT hash
	BuildHash string

	//BuildTag GIT tag
	BuildTag string

	//ClientID application client ID
	ClientID string

	//InstanceID id of instance
	InstanceID uuid.UUID
)

func init() {
	if InstanceID == uuid.Nil {
		InstanceID = uuid.NewV4()
	}
}
