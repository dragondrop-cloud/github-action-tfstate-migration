package statemigration

import (
	"fmt"

	"github.com/dragondrop-cloud/github-action-tfstate-migration/tfvars"
)

// WorkspaceDirectory is a relative path to a directory that corresponds to a single Terraform workspace.
type WorkspaceDirectory string

// StateMigrator interface for running Terraform state migrations.
type StateMigrator interface {

	// MigrateAllWorkspaces runs migrations for all workspaces by coordinating calls to MigrateWorkspace.
	MigrateAllWorkspaces() error

	// MigrateWorkspace runs migrations for the workspace specified.
	MigrateWorkspace(w WorkspaceDirectory) error
}

// stateMigrator implements the StateMigrator interface.
type stateMigrator struct {

	// config is composed of environment variables needed to run StateMigrator methods.
	config *Config

	// tfVar is a struct which can extract the remote variables needed to run migration statements.
	tfVar tfvars.TFVars
}

// NewStateMigrator instantiates a new implementation of the StateMigrator interface.
func NewStateMigrator() (StateMigrator, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("[NewConfig] %v", err)
	}

	tfVar, err := tfvars.NewTFVars()
	if err != nil {
		return nil, fmt.Errorf("[NewTFVars] %v", err)
	}

	return &stateMigrator{
		config: conf,
		tfVar:  tfVar,
	}, nil
}
