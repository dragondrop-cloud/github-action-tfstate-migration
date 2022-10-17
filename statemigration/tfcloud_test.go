package statemigration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func CreateStateMigrator(t *testing.T) stateMigrator {
	_, isRemote := os.LookupEnv("TerraformCloudToken")
	if !isRemote {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("[godotenv.Load] Unexpected error: %v", err)
		}
	}

	tfc := stateMigrator{
		config: &Config{
			TerraformCloudToken:        os.Getenv("TerraformCloudToken"),
			TerraformCloudOrganization: os.Getenv("TerraformCloudOrganization"),
		},
		httpClient: http.Client{},
	}

	return tfc
}

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

// TODO
func TestCreatePlanOnlyRefreshRun(t *testing.T) {
}

func TestDiscardActiveRunsUnlockState(t *testing.T) {
	sm := CreateStateMigrator(t)
	ctx := context.Background()

	// Very simple test, only checking that it can run end to end
	err := sm.discardActiveRunsUnlockState(ctx, os.Getenv("TerraformCloudWorkspaceID"))
	if err != nil {
		t.Errorf("[sm.discardActiveRunsUnlockState] %v", err)
	}
}

// TODO
func TestExtractRecentRunStatuses(t *testing.T) {
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

func TestIsStatusPostConfirmation(t *testing.T) {
	status := "example"
	outputOne := isStatusPostConfirmation(status)
	if outputOne != false {
		t.Errorf("got %v, expected %v", outputOne, false)
	}

	validStatusSlice := []string{
		"confirmed",
		"post_plan_running",
		"post_plan_completed",
		"apply_queued",
		"applying",
	}
	for _, status = range validStatusSlice {
		outputTwo := isStatusPostConfirmation(status)
		if outputTwo != true {
			t.Errorf("got %v, expected %v, input status of: %v", outputTwo, true, status)
		}

	}
}

func TestIsStatusTerminalState(t *testing.T) {
	status := "example"
	outputOne := isStatusTerminalState(status)
	if outputOne != false {
		t.Errorf("got %v, expected %v", outputOne, false)
	}

	validStatusSlice := []string{
		"policy_soft_failed",
		"planned_and_finished",
		"applied",
		"discarded",
		"errored",
		"canceled",
		"force_canceled",
	}
	for _, status = range validStatusSlice {
		outputTwo := isStatusTerminalState(status)
		if outputTwo != true {
			t.Errorf("got %v, expected %v, input status of: %v", outputTwo, true, status)
		}

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
