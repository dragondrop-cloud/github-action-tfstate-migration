package tfvars

import (
	"reflect"
	"testing"
)

func TestGroupToVariablesDecode(t *testing.T) {
	gtv := GroupToVariables{}

	// Missing value
	err := gtv.Decode(`{
		"group_1": {
			"key_1": {"category": "cat_1"},
		},
}`)

	if err == nil {
		t.Errorf("Expected error of missing variable value, got nil error: %v", err)
	}

	// Missing category
	gtv = GroupToVariables{}

	err = gtv.Decode(`{
		"group_1": {
			"key_1": {"value": "val_1"},
		},
}`)

	if err == nil {
		t.Errorf("Expected error of missing category value, got nil error: %v", err)
	}

	// Invalid categories
	gtv = GroupToVariables{}

	err = gtv.Decode(`{
		"group_1": {
			"key_1": {"value": "val_1", "category": "cat_1"},
			"key_2": {"value": "val_2", "category": "val_2"}
		},
		"group_2": {
			"key_1": {"value": "val_1", "category": "cat_1"},
			"key_2": {"value": "val_2", "category": "val_2"}
		}
}`)

	if err == nil {
		t.Errorf("Expected error of invalid category value, got nil error: %v", err)
	}

	// Everything passes
	err = gtv.Decode(`{
		"group_1": {
			"key_1": {"value": "val_1", "category": "env"},
			"key_2": {"value": "val_2", "category": "env"}
		},
		"group_2": {
			"key_1": {"value": "val_1", "category": "terraform"},
			"key_2": {"value": "val_2", "category": "terraform"}
		}
	}`)

	expectedOutput := GroupToVariables{
		"group_1": Variables{
			"key_1": VariableData{
				value:    "val_1",
				category: "env",
			},
			"key_2": VariableData{
				value:    "val_2",
				category: "env",
			},
		},
		"group_2": Variables{
			"key_1": VariableData{
				value:    "val_1",
				category: "terraform",
			},
			"key_2": VariableData{
				value:    "val_2",
				category: "terraform",
			},
		},
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(gtv, expectedOutput) {
		t.Errorf("got:\n%v\nexpected:\n%v", gtv, expectedOutput)
	}

}
