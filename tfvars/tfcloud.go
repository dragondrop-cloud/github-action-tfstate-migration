package tfvars

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Jeffail/gabs/v2"
)

// tfCloud implements the TFVars interface for a Terraform Cloud remote backend.
type tfCloud struct {

	// config contains the configuration needed for tfCloud methods to run.
	config *Config

	// httpClient contains an http.Client struct
	httpClient http.Client
}

// CreateAllWorkspaceVarsFiles extracts variables for all workspaces and saves them into
// .tfvars files within the appropriate directory.
func (tfc *tfCloud) CreateAllWorkspaceVarsFiles() error {
	ctx := context.Background()

	workspaceToVarSetVars, err := tfc.getVarSetVarsByWorkspace()
	if err != nil {
		return fmt.Errorf("[tfc.getVarSetVarsByWorkspace] %v", err)
	}

	for workspace, _ := range tfc.config.WorkspaceToDirectory {
		err := tfc.PullWorkspaceVariables(ctx, workspace, workspaceToVarSetVars)
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

// TODO: Implement and test end to end
// getVarSetVarsByWorkspace produces a map between a workspace name and variables associated
// with that workspace from variable sets.
func (tfc *tfCloud) getVarSetVarsByWorkspace() (map[string]VariableMap, error) {
	varSetIDToWorkspaces, err := tfc.getVarSetIdsForOrg()
	if err != nil {
		return nil, fmt.Errorf("[tfc.getVarSetIdsForOrg] %v", err)
	}

	workspaceToVarSetVars, err := tfc.getVarSetVars(varSetIDToWorkspaces)
	if err != nil {
		return nil, fmt.Errorf("[tfc.getVarSetVars] %v", err)
	}

	return workspaceToVarSetVars, nil
}

// TODO: Implement and unit test
func (tfc *tfCloud) getVarSetIdsForOrg() (map[string]map[string]bool, error) {
	// GET organizations/:organization_name/varsets
	requestPath := fmt.Sprintf(
		"https://app.terraform.io/api/v2/organizations/%v/varsets",
		tfc.config.TerraformCloudOrganization,
	)

	httpRequest, err := tfc.buildTFCloudHTTPRequest(
		context.Background(),
		"getAllVarSetIds",
		"GET",
		requestPath,
	)
	if err != nil {
		return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
	}

	response, err := tfc.terraformCloudRequest(httpRequest, "getVarSetVars")
	if err != nil {
		return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
	}

	varSetIdToWorkspaces, err := tfc.extractVarSetInformation(response)
	if err != nil {
		return nil, fmt.Errorf("[tfc.extractVarSetInformation] %v", err)
	}

	return varSetIdToWorkspaces, nil
}

func (tfc *tfCloud) extractVarSetInformation(response []byte) (map[string]map[string]bool, error) {
	varSetToWorkspaces := map[string]map[string]bool{}

	container, err := gabs.ParseJSON(response)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	i := 0
	for container.Exists("data", strconv.Itoa(i)) {
		varSetID := container.Search("data", strconv.Itoa(i), "id").Data().(string)
		varSetToWorkspaces[varSetID] = map[string]bool{}

		j := 0
		for container.Exists("data", strconv.Itoa(i), "relationships", "workspaces", "data", strconv.Itoa(j)) {
			workspaceID := container.Search(
				"data",
				strconv.Itoa(i),
				"relationships",
				"workspaces",
				"data",
				strconv.Itoa(j),
				"id").Data().(string)
			varSetToWorkspaces[varSetID][workspaceID] = true
			j++
		}

		i++
	}

	return varSetToWorkspaces, nil
}

// GET varsets/:varset_id/relationships/vars
// TODO: Implement and unit test
func (tfc *tfCloud) getVarSetVars(varSetToWorkspaces map[string]map[string]bool) (map[string]VariableMap, error) {
	workspaceToVarSetVars := map[string]VariableMap{}

	for varSetID := range varSetToWorkspaces {
		requestPath := fmt.Sprintf("https://app.terraform.io/api/v2/varsets/%v/relationships/vars", varSetID)

		httpRequest, err := tfc.buildTFCloudHTTPRequest(
			context.Background(),
			"getVarSetVars",
			"GET",
			requestPath,
		)
		if err != nil {
			return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
		}

		response, err := tfc.terraformCloudRequest(httpRequest, "getVarSetVars")
		if err != nil {
			return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
		}

		workspaceToVarSetVars, err = tfc.extractWorkspaceVars(response, varSetToWorkspaces, workspaceToVarSetVars)
		if err != nil {
			return nil, fmt.Errorf("[tfc.extractWorkspaceVars] %v", err)
		}

	}
	return nil, nil
}

// TODO: Implement and unit test
// extractWorkspaceVars extracts workspace variables from the current variable set's variables.
func (tfc *tfCloud) extractWorkspaceVars(
	varSetVarsResponse []byte,
	varSetIDToWorkspaces map[string]map[string]bool,
	workspaceToVarSetVars map[string]VariableMap,
) (map[string]VariableMap, error) {

	return nil, nil
}

// PullWorkspaceVariables extracts variables for a single workspace saves into a .tfvars
// file within the appropriate directory.
func (tfc *tfCloud) PullWorkspaceVariables(
	ctx context.Context,
	workspaceName string,
	workspaceToVarSetVars map[string]VariableMap,
) error {
	workspaceVarsContainer, err := tfc.DownloadWorkspaceVariables(ctx, workspaceName)
	if err != nil {
		return fmt.Errorf("[tfc.DownloadWorkspaceVariables] %v", err)
	}

	workspaceVarsMap, err := tfc.parseWorkspaceVars(workspaceVarsContainer)
	if err != nil {
		return fmt.Errorf("[tfc.parseWorkspaceVars] %v", err)
	}

	tfVarsFile, err := tfc.generateTFVarsFile(workspaceVarsMap, workspaceToVarSetVars)
	if err != nil {
		return fmt.Errorf("[tfc.generateTFVarsFile] %v", err)
	}

	err = os.WriteFile("terraform.tfvars", tfVarsFile, 0400)
	if err != nil {
		return fmt.Errorf("[os.WriteFile] %v", err)
	}

	return nil
}

// TODO: implement and add unit tests.
// generateTFVarsFile aggregates varset variables and workspace-specific variables to create a
// .tfvars file for the current workspace.
func (tfc *tfCloud) generateTFVarsFile(
	workspaceVars VariableMap,
	workspaceToVarSetVars map[string]VariableMap,
) ([]byte, error) {
	return nil, nil
}

// TODO: implement and add unit tests.
// parseWorkspaceVars extracts workspace variables from a []byte from the Terraform Cloud
// endpoint and places them into a VariableMap.
func (tfc *tfCloud) parseWorkspaceVars(container *gabs.Container) (VariableMap, error) {
	return nil, nil
}

// DownloadWorkspaceVariables downloads a workspace's variables from the remote source.
func (tfc *tfCloud) DownloadWorkspaceVariables(ctx context.Context, workspaceName string) (*gabs.Container, error) {
	workspaceID, err := tfc.getWorkspaceID(ctx, workspaceName)
	if err != nil {
		return nil, fmt.Errorf("[tfc.getWorkspaceID] %v", err)
	}

	varsJSON, err := tfc.getWorkspaceVariables(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("[tfc.getWorkspaceVariables] %v", err)
	}

	container, err := gabs.ParseJSON(varsJSON)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	return container, nil
}

// getWorkspaceVariables calls the Terraform Cloud API and receives workspace-specific variables
// data as a []byte.
func (tfc *tfCloud) getWorkspaceVariables(ctx context.Context, workspaceID string) ([]byte, error) {
	requestName := "getWorkspaceVars"
	requestPath := fmt.Sprintf(
		"https://app.terraform.io/api/v2/workspaces/%v/vars",
		workspaceID,
	)

	request, err := tfc.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath)
	if err != nil {
		return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
	}

	jsonResponseBytes, err := tfc.terraformCloudRequest(request, requestName)

	if err != nil {
		return nil, fmt.Errorf("[tfc.terraformCloudRequest] %v", err)
	}

	return jsonResponseBytes, nil
}

// getWorkspaceID calls the Terraform Cloud API and gets the workspace ID for the
// relevant workspace name in the relevant organization.
func (tfc *tfCloud) getWorkspaceID(ctx context.Context, workspaceName string) (string, error) {
	requestName := "getWorkspaceID"
	requestPath := fmt.Sprintf(
		"https://app.terraform.io/api/v2/organizations/%v/workspaces/%v",
		tfc.config.TerraformCloudOrganization, workspaceName,
	)

	request, err := tfc.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath)

	if err != nil {
		return "", fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	jsonResponseBytes, err := tfc.terraformCloudRequest(request, requestName)

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
