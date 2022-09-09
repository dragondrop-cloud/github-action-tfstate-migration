package tfvars

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	varMap := VariableMap{
		"var_1": "val_1",
		"var_2": "val_2",
		"var_3": "val_3",
	}

	inputOther := VariableMap{
		"var_3": "new_val",
		"var_4": "val_4",
	}

	expectedOutput := VariableMap{
		"var_1": "val_1",
		"var_2": "val_2",
		"var_3": "new_val",
		"var_4": "val_4",
	}

	newMap := varMap.Merge(inputOther)

	if !reflect.DeepEqual(expectedOutput, newMap) {
		t.Errorf("got %v, expected %v", newMap, expectedOutput)
	}
}
