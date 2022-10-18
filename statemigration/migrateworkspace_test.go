package statemigration

import (
	"reflect"
	"testing"
)

func TestBuildTFMigrateArgs(t *testing.T) {
	sm := stateMigrator{
		config: &Config{
			IsApply: false,
		},
	}

	_, output := sm.BuildTFMigrateArgs()
	expectedOutput := []string{"plan", "--config=./dragondrop/tfmigrate/.tfmigrate.hcl"}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}

	sm = stateMigrator{
		config: &Config{
			IsApply: true,
		},
	}

	_, output = sm.BuildTFMigrateArgs()
	expectedOutput = []string{"apply", "--config=./dragondrop/tfmigrate/.tfmigrate.hcl"}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}

}
