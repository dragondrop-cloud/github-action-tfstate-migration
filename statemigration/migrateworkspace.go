package statemigration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// MigrateAllWorkspaces runs migrations for all workspaces by coordinating calls to MigrateWorkspace.
func (sm *stateMigrator) MigrateAllWorkspaces() error {
	fmt.Println("Beginning to create all workspace variable files.")
	err := sm.tfVar.CreateAllWorkspaceVarsFiles()

	if err != nil {
		return fmt.Errorf("[sm.tfVar.CreateAllWorkspaceVarsFiles] %v", err)
	}
	fmt.Println("Done creating workspace variable files.")

	for workspace, directory := range sm.config.WorkspaceToDirectory {
		if directory == "null" {
			continue
		}

		fmt.Printf("Beginning to migrate the directory %v\n", directory)
		err = sm.MigrateWorkspace(workspace, WorkspaceDirectory(directory))
		if err != nil {
			return fmt.Errorf("[sm.MigrateWorkspace] Error migrating %v workspace: %v", directory, err)
		}
		fmt.Printf("Done migrating the directory %v\n", directory)
	}

	fmt.Println("Done migrating all workspaces.")
	return nil
}

// MigrateWorkspace runs migrations for the workspace specified.
func (sm *stateMigrator) MigrateWorkspace(workspace string, directory WorkspaceDirectory) error {
	ctx := context.Background()

	err := os.Chdir(fmt.Sprintf("/github/workspace%v", string(directory)))
	if err != nil {
		return fmt.Errorf("[os.Chdir] %v", err)
	}

	if sm.config.TerraformVersion != "" {
		tfSwitchArgs := []string{string(sm.config.TerraformVersion)}
		err = executeCommand("tfswitch", tfSwitchArgs...)

		if err != nil {
			return fmt.Errorf("[executeCommand `tfswitch`] %v", err)
		}
	} else {
		err = executeCommand("tfswitch", []string{}...)
		if err != nil {
			return fmt.Errorf("[executeCommand `tfswitch`] %v", err)
		}
	}

	terraformInitArgs := []string{"init"}
	err = executeCommand("terraform", terraformInitArgs...)

	if err != nil {
		return fmt.Errorf("[executeCommand `terraform init`] %v", err)
	}

	fmt.Printf("Running migrations for: %v", directory)

	planOrApply, tfMigrateArgs := sm.BuildTFMigrateArgs()

	workspaceID, err := sm.getWorkspaceID(ctx, workspace)
	if err != nil {
		return fmt.Errorf("[sm.getWorkspaceID] %v", err)
	}

	if planOrApply == "apply" {
		err = sm.discardActiveRunsUnlockState(ctx, workspaceID)
		if err != nil {
			return fmt.Errorf("[sm.clearPendingRunsAndUnlockState`] %v", err)
		}
	}

	err = executeCommand("tfmigrate", tfMigrateArgs...)

	if err != nil {
		return fmt.Errorf("[executeCommand `tfmigrate`] %v", err)
	}

	if planOrApply == "apply" {
		err = sm.createPlanOnlyRefreshRun(ctx, workspaceID)
		if err != nil {
			return fmt.Errorf("[sm.createPlanOnlyRefreshRun`] %v", err)
		}
	}

	return nil
}

// BuildTFMigrateArgs constructs a slice of strings for use within
// a tfmigrate command
func (sm *stateMigrator) BuildTFMigrateArgs() (string, []string) {
	var tfMigrateCMD string

	if sm.config.IsApply {
		tfMigrateCMD = "apply"
	} else {
		tfMigrateCMD = "plan"
	}

	tfMigrateArgs := []string{tfMigrateCMD, "--config=./dragondrop/tfmigrate/.tfmigrate.hcl"}

	return tfMigrateCMD, tfMigrateArgs
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
	fmt.Printf("\n`%s %s` output:\n\n%v\n", command, args, out.String())
	return nil
}
