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

	output := sm.BuildTFMigrateArgs()
	expectedOutput := []string{"plan"}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}

	sm = stateMigrator{
		config: &Config{
			IsApply: true,
		},
	}

	output = sm.BuildTFMigrateArgs()
	expectedOutput = []string{"apply"}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}

}
