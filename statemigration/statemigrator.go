package statemigration

// Workspace is a relative path to a directory that corresponds to a single Terraform workspace.
type Workspace string

// StateMigrator interface for running Terraform state migrations.
type StateMigrator interface {

	// MigrateAllWorkspaces runs migrations for all workspaces by coordinating calls to MigrateWorkspace.
	MigrateAllWorkspaces() error

	// MigrateWorkspace runs migrations for the workspace specified.
	MigrateWorkspace(w Workspace) error
}

// stateMigrator implements the StateMigrator interface.
type stateMigrator struct {

	// config is composed of environment variables needed to run StateMigrator methods.
	config *Config
}

// NewStateMigrator instantiates a new implementation of the StateMigrator interface.
func NewStateMigrator(conf *Config) StateMigrator {
	return &stateMigrator{
		config: conf,
	}
}
