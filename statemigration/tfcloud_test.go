package statemigration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildTFCloudHTTPRequest(t *testing.T) {
	ctx := context.Background()
	config := Config{
		TerraformCloudOrganization: "dragondrop-cloud",
		TerraformCloudToken:        "example_token",
		WorkspaceToDirectory:       map[string]string{"workspace_1": "directory_1", "workspace_2": "directory_2"},
	}

	sm := stateMigrator{
		config:     &config,
		httpClient: http.Client{},
	}

	request, err := sm.buildTFCloudHTTPRequest(
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

	config := Config{
		TerraformCloudOrganization: "dragondrop-cloud",
		TerraformCloudToken:        "example_token",
		WorkspaceToDirectory:       map[string]string{"workspace_1": "directory_1", "workspace_2": "directory_2"},
	}

	sm := stateMigrator{
		config:     &config,
		httpClient: http.Client{},
	}

	request, _ := sm.buildTFCloudHTTPRequest(
		ctx, "testRequest", "GET", server.URL+"/terraform/cloud/",
	)

	output, err := sm.terraformCloudRequest(request, "testRequest")

	if err != nil {
		t.Errorf("Was expecting no error, instead received %v", err)
	}

	if string(output) != "example output" {
		t.Errorf("Got %v, expected 'example output'", string(output))
	}
}
