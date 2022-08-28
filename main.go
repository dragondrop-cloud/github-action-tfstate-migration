package main

import (
	"fmt"
	"os"

	"github.com/dragondrop-cloud/github-action-tfstate-migration/statemigrator"
)

func main() {

	config := statemigrator.NewConfig()

	stateMigrator := statemigrator.NewStateMigrator(config)

	err := stateMigrator.MigrateAllWorkspaces()

	if err != nil {
		fmt.Printf("error migrating all workspace's state: %v", err)
		os.Exit(1)
	}
	fmt.Println("Successfully ran tf-state-migration job.")
}
