package tfvars

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Jeffail/gabs/v2"
)

// tfCloud implements the TFVars interface for a Terraform Cloud remote backend.
type tfCloud struct {

	// config contains the configuration needed for tfCloud methods to run.
	config *Config

	// httpClient contains an http.Client struct
	httpClient http.Client
}

// PullAllWorkspaceVariables extracts variables for all workspaces and saves them into
// .tfvars files within the appropriate directory.
func (tfc *tfCloud) PullAllWorkspaceVariables() error {
	ctx := context.Background()

	for workspace, _ := range tfc.config.WorkspaceToDirectory {
		err := tfc.PullWorkspaceVariables(ctx, workspace)
		if err != nil {
			return fmt.Errorf(
				"[tfc.PullWorkspaceVariables] Error in workspace %v: %v",
				workspace,
				err,
			)
		}
	}
	return nil
}

// PullWorkspaceVariables extracts variables for a single workspace saves into a .tfvars
// file within the appropriate directory.
func (tfc *tfCloud) PullWorkspaceVariables(ctx context.Context, workspaceName string) error {
	varContainer, err := tfc.DownloadWorkspaceVariables(ctx, workspaceName)
	if err != nil {
		return fmt.Errorf("[tfc.DownloadWorkspaceVariables] %v", err)
	}
	// TODO: parse out variables into a reasonable format for .tfvars

	// TODO: Save out that file into the appropriate directory.

	return nil
}

// TODO: CI/CD test should work here for making a request. Both locally and in Github Action pipeline
// DownloadWorkspaceVariables downloads a workspace's variables from the remote source.
func (tfc *tfCloud) DownloadWorkspaceVariables(ctx context.Context, workspaceName string) (*gabs.Container, error) {
	// TODO: get workspace id (require organization name as config component)

	// TODO: Make request to get variables associated with the workspace

	// TODO: parse as gabs and return
	return nil, nil
}

// TODO: Pull in unit tests for this and add to CI/CD pipeline.
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

// terraformCloudRequest build, executes, and processes an API call to the Terraform Cloud API.
func (tfc *tfCloud) terraformCloudRequest(request *http.Request, requestName string) ([]byte, error) {

	response, err := tfc.httpClient.Do(request)

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

// TODO: pull in unit tests for this function
// buildTFCloudHTTPRequest structures a request to the Terraform Cloud api.
func (tfc *tfCloud) buildTFCloudHTTPRequest(
	ctx context.Context, requestName string, method string, requestPath string,
) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestPath, nil)
	if err != nil {
		return nil, fmt.Errorf("[%v] error in http request instantiation: %v", requestName, err)
	}

	request.Header = http.Header{
		"Authorization": {"Bearer " + tfc.config.TerraformCloudToken},
		"Content-Type":  {"application/vnd.api+json"},
	}

	return request, nil
}
