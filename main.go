package main

import (
	"fmt"
	"os"

	"github.com/dragondrop-cloud/github-action-tfstate-migration/statemigration"
)

func main() {
	stateMigrator, err := statemigration.NewStateMigrator()
	if err != nil {
		fmt.Printf("error in statemigration.NewStateMigrator(config): %v", err)
		os.Exit(1)
	}

	err = stateMigrator.MigrateAllWorkspaces()

	if err != nil {
		fmt.Printf("error migrating all workspace's state: %v", err)
		os.Exit(1)
	}
	fmt.Println("Successfully ran tfstate-migration job.")
}
