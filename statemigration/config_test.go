package statemigration

import "testing"

func TestVersionDecoder(t *testing.T) {
	var envVar Version

	err := envVar.Decode("1.2.3")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := Version("1.2.3")

	if envVar != expectedResult {
		t.Errorf("got %v, expected %v", envVar, expectedResult)
	}

	var envVarTwo Version

	err = envVarTwo.Decode("~>1.2.3")
	if err == nil {
		t.Errorf("said '1.2.3' is valid, but it is not")
	}

	var envVarThree Version

	err = envVarThree.Decode("~>1.2")
	if err == nil {
		t.Errorf("said '~>1.2' is valid, but it is not")
	}

	var envVarFour Version

	err = envVarFour.Decode("2.6.5")
	if err == nil {
		t.Errorf("said '~>2.6.5' is valid, but it is not")
	}

}
