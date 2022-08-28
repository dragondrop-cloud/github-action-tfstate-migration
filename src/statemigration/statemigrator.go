package statemigration

type StateMigrator interface {
	MigrateAllWorkspaces() error

	MigrateWorkspace() error
}

type stateMigrator struct {
}

func NewStateMigrator() StateMigrator {

	return &stateMigrator{}
}
