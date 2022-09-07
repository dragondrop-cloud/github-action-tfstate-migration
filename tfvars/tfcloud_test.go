package tfvars

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetWorkspaceVariables(t *testing.T) {
	_, isRemote := os.LookupEnv("TerraformCloudToken")
	if !isRemote {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("[godotenv.Load] Unexpected error: %v", err)
		}
	}

	tfc := tfCloud{
		config: &Config{
			TerraformCloudToken: os.Getenv("TerraformCloudToken"),
		},
		httpClient: http.Client{},
	}

	_, err := tfc.getWorkspaceVariables(context.Background(), os.Getenv("workspaceID"))

	if err != nil {
		t.Errorf("[tfc.getWorkspaceVariables] unexpected error: %v", err)
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

func TestExtractVarSetInformation(t *testing.T) {
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

	expectedOutputMapToSet := map[string]map[string]bool{
		"varset-mio9UUFyFMjU33S4": {
			"ws-abcd12345": true,
			"ws-abcd12346": true,
		},
		"varset-tuyo9UUFyFMjU33S4": {
			"ws-xyze12345": true,
			"ws-xyze12346": true,
		},
	}

	tfc := tfCloud{}

	outputMapToSet, err := tfc.extractVarSetInformation(inputResponse)
	if err != err {
		t.Errorf("unexpected error in tfc.extractVarSetInformation: %v", err)
	}
	
	if !reflect.DeepEqual(outputMapToSet, expectedOutputMapToSet) {
		t.Errorf("got %v\n expected %v", outputMapToSet, expectedOutputMapToSet)
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
