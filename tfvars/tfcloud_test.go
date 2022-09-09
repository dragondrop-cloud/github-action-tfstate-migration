package tfvars

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
)

func CreateTFC(t *testing.T) tfCloud {
	_, isRemote := os.LookupEnv("TerraformCloudToken")
	if !isRemote {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("[godotenv.Load] Unexpected error: %v", err)
		}
	}

	tfc := tfCloud{
		config: &Config{
			TerraformCloudToken:        os.Getenv("TerraformCloudToken"),
			TerraformCloudOrganization: os.Getenv("TerraformCloudOrganization"),
			WorkspaceToDirectory:       map[string]string{"google-backend-api-dev": "/"},
		},
		httpClient: http.Client{},
	}

	return tfc
}

func TestCreateWorkspaceToVarSetVars(t *testing.T) {
	inputVarSetVars := map[string]VariableMap{
		"var_set_id_1": {
			"var1": "abc",
			"var2": "abc",
		},
		"var_set_id_2": {
			"var1": "edf",
			"var3": "xyz",
		},
		"var_set_id_3": {
			"var4": "123",
		},
	}

	inputWorkspaceToVarSetIDs := map[string]map[string]bool{
		"workspace_1": {"var_set_id_1": true, "var_set_id_2": true},
		"workspace_2": {"var_set_id_3": true},
	}

	expectedOutput := map[string]VariableMap{
		"workspace_1": {
			"var1": "edf",
			"var2": "abc",
			"var3": "xyz",
		},
		"workspace_2": {
			"var4": "123",
		},
	}

	tfc := CreateTFC(t)

	outputWorkspaceToVarSetVars, err := tfc.createWorkspaceToVarSetVars(
		inputVarSetVars,
		inputWorkspaceToVarSetIDs,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expectedOutput, outputWorkspaceToVarSetVars) {
		t.Errorf("got %v, expected %v", outputWorkspaceToVarSetVars, expectedOutput)
	}
}

func TestGenerateTFVarsFile(t *testing.T) {
	inputWorkspaceVars := VariableMap{
		"var_1": "val_1",
		"var_2": "val_2",
	}

	inputWorkspaceVarSetVars := VariableMap{
		"var_2": "val_xyz",
		"var_3": "val_3",
	}

	tfc := CreateTFC(t)
	byteArray, _ := tfc.generateTFVarsFile(inputWorkspaceVars, inputWorkspaceVarSetVars)

	expectedOutput := `var_2 = "val_2"
var_3 = "val_3"
var_1 = "val_1"
`

	if expectedOutput != string(byteArray) {
		t.Errorf("got:\n%v\nexpected:\n%v",
			strconv.Quote(string(byteArray)),
			expectedOutput)
	}
}

func TestGetWorkspaceToVarSetIds(t *testing.T) {
	tfc := CreateTFC(t)

	output, err := tfc.getWorkspaceToVarSetIDs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output == nil {
		t.Errorf("expected non-nil output from tfc.getWorkspaceToVarSetIDs")
	}
}

func TestGetVarSetIdsForOrg(t *testing.T) {
	tfc := CreateTFC(t)

	output, err := tfc.getVarSetIdsForOrg()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output == nil {
		t.Errorf("expected non-nil output from tfc.getVarSetIdsForOrg")
	}
}

func TestGetWorkspaceToVarSetVars(t *testing.T) {
	tfc := CreateTFC(t)

	output, err := tfc.getWorkspaceToVarSetVars()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output == nil {
		t.Errorf("expected non-nil output from tfc.getWorkspaceToVarSetVars")
	}
}

func TestGetWorkspaceVariables(t *testing.T) {
	tfc := CreateTFC(t)

	output, err := tfc.getWorkspaceVariables(context.Background(), os.Getenv("workspaceID"))

	if err != nil {
		t.Errorf("[tfc.getWorkspaceVariables] unexpected error: %v", err)
	}

	if output == nil {
		t.Errorf("expected non-nil output from tfc.getWorkspaceVariables")
	}
}

func TestGetVarSetVars(t *testing.T) {
	tfc := CreateTFC(t)

	output, err := tfc.getVarSetVars([]string{os.Getenv("VarSetID")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output == nil {
		t.Errorf("expected non-nil output from tfc.getVarSetVars")
	}
}

func TestBuildTFCloudHTTPRequest(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		TerraformCloudOrganization: "dragondrop-cloud",
		TerraformCloudToken:        "example_token",
		WorkspaceToDirectory:       map[string]string{"workspace_1": "directory_1", "workspace_2": "directory_2"},
	}

	tfc := tfCloud{
		config:     config,
		httpClient: http.Client{},
	}

	request, err := tfc.buildTFCloudHTTPRequest(
		ctx, "testRequest", "GET", "https://test.com/",
	)
	if err != nil {
		t.Errorf("Error in buildTFCloudHTTPRequest: %v", err)
	}

	outputContentType := request.Header.Get("Content-Type")
	expectedContentType := "application/vnd.api+json"
	if outputContentType != expectedContentType {
		t.Errorf("header content type: got %v, expected %v", outputContentType, expectedContentType)
	}

	outputContentType = request.Header.Get("Authorization")
	expectedContentType = "Bearer example_token"
	if outputContentType != expectedContentType {
		t.Errorf("header authorization: got %v, expected %v", outputContentType, expectedContentType)
	}

}

func TestExtractWorkspaceID(t *testing.T) {
	jsonBytes := []byte(`{
		"data" : {
			"attributes": {"attribute_1": 10},
			"id": "8675309"
		}
	}`)

	outputValue, err := extractWorkspaceID(jsonBytes)
	if err != nil {
		t.Errorf("Unexpectedly failed with %v", err)
	}
	expectedValue := "8675309"
	if outputValue != expectedValue {
		t.Errorf("Got %v, expected %v", expectedValue, outputValue)
	}
}

func TestExtractWorkspaceVars(t *testing.T) {
	inputResponse := []byte(`
{
   "data":[
      {
         "id":"var-AD4pibb9nxo1468E",
         "type":"vars",
         "attributes":{
            "key":"varKey_1",
            "value":"varVal_1",
            "hcl":false
         }
      },
      {
         "id":"var-dewc9nxoasdE",
         "type":"vars",
         "attributes":{
            "key":"varKey_2",
            "value":null,
			"sensitive": true,
            "hcl":false
         }
      },
      {
         "id":"var-SDBnxoasdE",
         "type":"vars",
         "attributes":{
            "key":"varKey_3",
            "value":"varValue_3",
			"sensitive": true,
            "hcl":false
         }
      }
   ]
}
`)

	expectedOutput := VariableMap{
		"varKey_1": "varVal_1",
		"varKey_2": "null",
		"varKey_3": "varValue_3",
	}

	tfc := tfCloud{}

	output, err := tfc.extractWorkspaceVars(inputResponse)
	if err != nil {
		t.Errorf("unexpected err in tfc.extractWorkspaceVars: %v", err)
	}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}
}

func TestExtractVarSetIDs(t *testing.T) {
	inputResponse := []byte(`{
  "data": [
    {
      "id": "varset-mio9UUFyFMjU33S4",
      "type": "varsets",
      "attributes":  {
         "name": "varset-b7af6a77",
         "workspace-count": 2
      },
      "relationships": {
        "organization": {
          "data": {"id": "organization_1", "type": "organizations"}
        },
        "vars": {
          "data": [
           {"id": "var-abcd12345", "type": "vars"}
          ]
        },
        "workspaces": {
          "data": [
           {"id": "ws-abcd12345", "type": "workspaces"},
           {"id": "ws-abcd12346", "type": "workspaces"}
          ]
        }
      }
    },
	{
      "id": "varset-tuyo9UUFyFMjU33S4",
      "type": "varsets",
      "attributes":  {
         "name": "varset-b7af6a77",
         "workspace-count": 2
      },
      "relationships": {
        "organization": {
          "data": {"id": "organization_1", "type": "organizations"}
        },
        "vars": {
          "data": [
           {"id": "var-abcd12345", "type": "vars"}
          ]
        },
        "workspaces": {
          "data": [
           {"id": "ws-xyze12345", "type": "workspaces"},
           {"id": "ws-xyze12346", "type": "workspaces"}
          ]
        }
      }
    }
  ]
}
`)

	expectedOutputMapToSet := []string{
		"varset-mio9UUFyFMjU33S4",
		"varset-tuyo9UUFyFMjU33S4",
	}

	tfc := tfCloud{}

	outputMapToSet, err := tfc.extractVarSetIDs(inputResponse)
	if err != err {
		t.Errorf("unexpected error in tfc.extractVarSetInformation: %v", err)
	}

	if !reflect.DeepEqual(outputMapToSet, expectedOutputMapToSet) {
		t.Errorf("got %v\n expected %v", outputMapToSet, expectedOutputMapToSet)
	}
}

func TestExtractVarSetIDsForWorkspace(t *testing.T) {
	inputResponse := []byte(`
	{
   "data":[
      {
         "id":"varset-yN8675309",
         "type":"varsets",
         "attributes":{
            "name":"var_set_name"
         },
         "relationships":{
            "workspaces":{
               "data":[]
            }
         }
      },
      {
         "id":"varset-W1324adf234",
         "type":"varsets",
         "relationships":{
            "organization":{
            },
            "vars":{
               "data":[]
            }
         }
      }
   ]
}`)

	expectedOutput := map[string]bool{
		"varset-W1324adf234": true,
		"varset-yN8675309":   true,
	}

	tfc := tfCloud{}

	output, err := tfc.extractVarSetIDsForWorkspace(inputResponse)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(output, expectedOutput) {
		t.Errorf("got %v, expected %v", output, expectedOutput)
	}
}

func TestExtractVarsFromVarSet(t *testing.T) {
	inputResponse := []byte(`{
  "data": [
    {
      "id": "var-134r1k34nj5kjn",
      "type": "vars",
      "attributes": {
        "key": "F115037558b045dd82da40b089e5db745",
        "value": null,
        "sensitive": false,
        "category": "terraform",
        "hcl": false,
        "created-at": "2021-10-29T18:54:29.379Z",
        "description": ""
      }
    },
	{
      "id": "var-894r1k34nj5kjn",
      "type": "vars",
      "attributes": {
        "key": "asd7558b045dd82da40b089e5db745",
        "value": "asdazxc0dfd3060e2c37890422905f",
        "sensitive": false,
        "category": "terraform",
        "hcl": false,
        "created-at": "2021-10-29T18:54:29.379Z",
        "description": ""
		}
	}
  	]
	}`)

	inputVarSetToVars := map[string]VariableMap{}

	inputVarSetID := "test-var-set-id"

	expectedOutput := map[string]VariableMap{
		"test-var-set-id": {
			"F115037558b045dd82da40b089e5db745": "null",
			"asd7558b045dd82da40b089e5db745":    "asdazxc0dfd3060e2c37890422905f",
		},
	}

	tfc := tfCloud{}
	varSetToVars, err := tfc.extractVarsFromVarSet(
		inputResponse, inputVarSetToVars, inputVarSetID,
	)
	if err != nil {
		t.Errorf("error in tfc.extractVarsFromVarSet: %v", err)
	}

	if !reflect.DeepEqual(expectedOutput, varSetToVars) {
		t.Errorf("got %v, expected %v", varSetToVars, expectedOutput)
	}
}

func TestTerraformCloudRequest(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/terraform/cloud/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`example output`))
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	config := &Config{
		TerraformCloudOrganization: "dragondrop-cloud",
		TerraformCloudToken:        "example_token",
		WorkspaceToDirectory:       map[string]string{"workspace_1": "directory_1", "workspace_2": "directory_2"},
	}

	tfc := tfCloud{
		config:     config,
		httpClient: http.Client{},
	}

	request, _ := tfc.buildTFCloudHTTPRequest(
		ctx, "testRequest", "GET", server.URL+"/terraform/cloud/",
	)

	output, err := tfc.terraformCloudRequest(request, "testRequest")

	if err != nil {
		t.Errorf("Was expecting no error, instead received %v", err)
	}

	if string(output) != "example output" {
		t.Errorf("Got %v, expected 'example output'", string(output))
	}
}
