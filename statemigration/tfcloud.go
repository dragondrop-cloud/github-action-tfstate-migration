package statemigration

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// discardActiveRunsUnlockState identifies pending/active Terraform Cloud runs and discards
// them so that tfmigrate apply can itself apply a state lock and run migrations.
func (sm *stateMigrator) discardActiveRunsUnlockState() error {

	return nil
}

// createPlanOnlyRefreshRun kicks off a new plan-only, refresh-state run for the workspace.
func (sm *stateMigrator) createPlanOnlyRefreshRun() error {

	return nil
}

// TODO: needs significant refinement
// buildTFCloudHTTPRequest structures a request to the Terraform Cloud api.
func (sm *stateMigrator) buildTFCloudHTTPRequest(ctx context.Context, requestName string, method string, requestPath string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestPath, nil)
	if err != nil {
		return nil, fmt.Errorf("[%v] error in http request instantiation: %v", requestName, err)
	}

	request.Header = http.Header{
		"Authorization": {"Bearer " + config.TerraformCloudToken},
		"Content-Type":  {"application/vnd.api+json"},
	}

	return request, nil
}

// TODO: needs significant refinement
// terraformCloudRequest build, executes, and processes an API call to the Terraform Cloud API.
func (sm *stateMigrator) terraformCloudRequest(, request *http.Request, requestName string) ([]byte, error) {

	response, err := c.httpClient.Do(request)

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
