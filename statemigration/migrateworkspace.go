package statemigration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// MigrateAllWorkspaces runs migrations for all workspaces by coordinating calls to MigrateWorkspace.
func (sm *stateMigrator) MigrateAllWorkspaces() error {
	for _, workspace := range sm.config.WorkspaceDirectories {
		if workspace == "null" {
			continue
		}

		err := sm.MigrateWorkspace(Workspace(workspace))
		if err != nil {
			return fmt.Errorf("[sm.MigrateWorkspace] Error migrating %v workspace: %v", workspace, err)
		}

	}

	return nil
}

// MigrateWorkspace runs migrations for the workspace specified.
func (sm *stateMigrator) MigrateWorkspace(w Workspace) error {
	err := os.Chdir(string(w))

	if err != nil {
		return fmt.Errorf("[os.Chdir] %v", err)
	}

	// TODO: Can pull this into a helper function at some point and unit test it,
	// TODO: otherwise everything is an integration test.
	var tfMigrateCMD string

	if sm.config.IsApply {
		tfMigrateCMD = "apply"
	} else {
		tfMigrateCMD = "plan"
	}

	tfmigrateArgs := []string{tfMigrateCMD}

	err = executeCommand("tfmigrate", tfmigrateArgs...)
	if err != nil {
		return fmt.Errorf("[executeCommand] %v", err)
	}

	return nil
}

// executeCommand wraps os.exec.Command with capturing of std output and errors.
func executeCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	// Setting up logging objects
	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("%v\n\n%v", err, stderr.String()+out.String())
	}
	fmt.Printf("\n%s Output:\n\n%v\n", command, out.String())
	return nil
}
