package statemigration

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Jeffail/gabs/v2"
)

// getWorkspaceID gets the workspace ID for the corresponding workspace name
// from the Terraform Cloud API.
func (sm *stateMigrator) getWorkspaceID(ctx context.Context, workspace string) (string, error) {
	requestName := "getWorkspaceID"
	requestPath := fmt.Sprintf("https://app.terraform.io/api/v2/organizations/%v/workspaces/%v", sm.config.TerraformCloudOrganization, workspace)

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath)

	if err != nil {
		return "", fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	jsonResponseBytes, err := sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return "", err
	}

	return extractWorkspaceID(jsonResponseBytes)
}

// extractWorkspaceID is a helper function that uses the gabs library to pull out the workspace ID
// from a Terraform Cloud API response.
func extractWorkspaceID(jsonBytes []byte) (string, error) {
	jsonParsed, err := gabs.ParseJSON(jsonBytes)
	if err != nil {
		return "", fmt.Errorf("[getWorkspaceID] error in parsing bytes array to json via 'gabs': %v", err)
	}

	value, ok := jsonParsed.Path("data.id").Data().(string)
	if !ok {
		return "", fmt.Errorf("[getWorkspaceID] unable to find workspace id: %v", err)
	}

	return value, nil
}

// discardActiveRunsUnlockState identifies pending/active Terraform Cloud runs and discards
// them so that tfmigrate apply can itself apply a state lock and run migrations.
func (sm *stateMigrator) discardActiveRunsUnlockState(ctx context.Context, workspaceID string) error {
	// TODO: First get a list of all active/pending runs
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#discard-a-run

	// TODO: For each run in list of active/pending, discard.
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#create-a-run
	return nil
}

// createPlanOnlyRefreshRun kicks off a new plan-only, refresh-state run for the workspace.
func (sm *stateMigrator) createPlanOnlyRefreshRun(ctx context.Context, workspaceID string) error {
	// TODO: Create a plan-only refresh run
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#list-runs-in-a-workspace
	return nil
}

// buildTFCloudHTTPRequest structures a request to the Terraform Cloud api.
func (sm *stateMigrator) buildTFCloudHTTPRequest(ctx context.Context, requestName string, method string, requestPath string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestPath, nil)
	if err != nil {
		return nil, fmt.Errorf("[%v] error in http request instantiation: %v", requestName, err)
	}

	request.Header = http.Header{
		"Authorization": {"Bearer " + sm.config.TerraformCloudToken},
		"Content-Type":  {"application/vnd.api+json"},
	}

	return request, nil
}

// terraformCloudRequest build, executes, and processes an API call to the Terraform Cloud API.
func (sm *stateMigrator) terraformCloudRequest(request *http.Request, requestName string) ([]byte, error) {

	response, err := sm.httpClient.Do(request)

	if err != nil {
		return nil, fmt.Errorf("[%v] error in http GET request to Terraform cloud: %v", requestName, err)
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("[%v] was unsuccessful, with the server returning: %v", requestName, response.StatusCode)
	}

	// Read in response body to bytes array.
	outputBytes, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("[%v] error in reading response into bytes array: %v", requestName, err)
	}

	return outputBytes, nil
}
