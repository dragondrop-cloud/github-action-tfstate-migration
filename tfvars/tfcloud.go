package tfvars

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
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

	if tfc.config.TerraformCloudToken == "null" {
		fmt.Println("Job kicked off in test-mode (TerraformCloudToken == 'null').")
		return nil
	}

	workspaceToVarSetIDs, workspaceToVarSetVars, varSetIDsToName, err := tfc.getWorkspaceToVarSetVars()
	if err != nil {
		return fmt.Errorf("[tfc.getWorkspaceToVarSetVars] %v", err)
	}
	fmt.Println("Done pulling down workspace variables from variable sets.")

	for workspace := range tfc.config.WorkspaceToDirectory {
		err = tfc.PullWorkspaceVariables(ctx, workspace, workspaceToVarSetVars, workspaceToVarSetIDs, varSetIDsToName)
		if err != nil {
			return fmt.Errorf(
				"[tfc.PullWorkspaceVariables] Error in workspace %v: %v",
				workspace,
				err,
			)
		}
		fmt.Printf(
			"Done pulling down workspace variables for workspace: %v\n", workspace,
		)
	}
	return nil
}

// getWorkspaceToVarSetVars produces a map between a workspace name and variables associated
// with that workspace from variable sets.
func (tfc *tfCloud) getWorkspaceToVarSetVars() (map[string]map[string]bool, map[string]VariableMap, map[string]string, error) {
	varSetIDsToName, err := tfc.getVarSetIdsForOrg()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[tfc.getVarSetIdsForOrg] %v", err)
	}

	varSetVars, err := tfc.getVarSetVars(varSetIDsToName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[tfc.getVarSetVars] %v", err)
	}

	workspaceToVarSetIDs, err := tfc.getWorkspaceToVarSetIDs()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[tfc.getWorkspaceToVarSetIDs] %v", err)
	}

	workspaceToVarSetVars, err := tfc.createWorkspaceToVarSetVars(varSetVars, workspaceToVarSetIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("[tfc.getWorkspaceToVarSetVars] %v", err)
	}

	return workspaceToVarSetIDs, workspaceToVarSetVars, varSetIDsToName, nil
}

// getVarSetIdsForOrg returns a map between var set ids and the var set's name.
func (tfc *tfCloud) getVarSetIdsForOrg() (map[string]string, error) {
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

	varSetIDToName, err := tfc.extractVarSetIDToName(response)
	if err != nil {
		return nil, fmt.Errorf("[tfc.extractVarSetInformation] %v", err)
	}

	return varSetIDToName, nil
}

// extractVarSetIDToName extracts variable set ids from a response json
// from the Terraform Cloud API.
func (tfc *tfCloud) extractVarSetIDToName(response []byte) (map[string]string, error) {
	varSetIDToName := map[string]string{}

	container, err := gabs.ParseJSON(response)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	i := 0
	for container.Exists("data", strconv.Itoa(i)) {
		varSetID := container.Search("data", strconv.Itoa(i), "id").Data().(string)
		varSetIDToName[varSetID] = container.Search("data", strconv.Itoa(i), "attributes", "name").Data().(string)
		i++
	}

	return varSetIDToName, nil
}

// getVarSetVars pulls down from terraform cloud all variables for each variable set passed in via
// varSetIDs.
func (tfc *tfCloud) getVarSetVars(varSetIDsToName map[string]string) (map[string]VariableMap, error) {
	varSetToVars := map[string]VariableMap{}

	for varSetID := range varSetIDsToName {
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

		varSetToVars, err = tfc.extractVarsFromVarSet(
			response, varSetToVars, varSetID,
		)
		if err != nil {
			return nil, fmt.Errorf("[tfc.extractWorkspaceVars] %v", err)
		}

	}
	return varSetToVars, nil
}

// extractVarsFromVarSet extracts workspace variables from the current variable set's variables.
func (tfc *tfCloud) extractVarsFromVarSet(
	varSetVarsResponse []byte,
	varSetToVars map[string]VariableMap,
	varSetID string,
) (map[string]VariableMap, error) {
	container, err := gabs.ParseJSON(varSetVarsResponse)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	varMap := VariableMap{}
	i := 0

	for container.Exists("data", strconv.Itoa(i)) {
		varKey := container.Search("data", strconv.Itoa(i), "attributes", "key").Data().(string)
		value := container.Search("data", strconv.Itoa(i), "attributes", "value").Data()

		if value == nil {
			i++
			continue
		}

		varValue := container.Search("data", strconv.Itoa(i), "attributes", "value").Data().(string)
		varMap[varKey] = varValue
		i++
	}

	varSetToVars[varSetID] = varMap

	return varSetToVars, nil
}

// getWorkspaceToVarSetIDs produce a map of workspaces to the corresponding var set IDs.
func (tfc *tfCloud) getWorkspaceToVarSetIDs() (map[string]map[string]bool, error) {
	ctx := context.Background()

	outputMap := map[string]map[string]bool{}

	for workspace := range tfc.config.WorkspaceToDirectory {
		workspaceID, err := tfc.getWorkspaceID(ctx, workspace)
		if err != nil {
			return nil, fmt.Errorf("[tfc.getWorkspaceID] %v", err)
		}

		requestPath := fmt.Sprintf(
			"https://app.terraform.io/api/v2/workspaces/%v/varsets", workspaceID,
		)
		httpRequest, err := tfc.buildTFCloudHTTPRequest(
			context.Background(),
			"getWorkspaceVarSets",
			"GET",
			requestPath,
		)
		if err != nil {
			return nil, fmt.Errorf("[tfc.buildTFCloudHTTPRequest] %v", err)
		}

		response, err := tfc.terraformCloudRequest(httpRequest, "getWorkspaceVarSets")
		if err != nil {
			return nil, fmt.Errorf("[tfc.terraformCloudRequest] %v", err)
		}

		varSetIDSet, err := tfc.extractVarSetIDsForWorkspace(response)
		if err != nil {
			return nil, fmt.Errorf("[tfc.terraformCloudRequest] %v", err)
		}

		outputMap[workspace] = varSetIDSet

	}

	return outputMap, nil
}

// extractVarSetIDsForWorkspace extracts variable set ids from a response to a request
// for all resources within a workspace.
func (tfc *tfCloud) extractVarSetIDsForWorkspace(response []byte) (map[string]bool, error) {
	container, err := gabs.ParseJSON(response)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	outputMap := map[string]bool{}

	i := 0
	for container.Exists("data", strconv.Itoa(i)) {
		value := container.Search("data", strconv.Itoa(i), "id").Data().(string)
		outputMap[value] = true

		i++
	}

	return outputMap, nil
}

// createWorkspaceToVarSetVars takes an input of two maps: var set ids to their variables and
// workspace to var set ids and returns a map of workspace to variable maps.
func (tfc *tfCloud) createWorkspaceToVarSetVars(
	varSetVars map[string]VariableMap, workspaceToVarSetIDs map[string]map[string]bool,
) (map[string]VariableMap, error) {
	outputWorkspaceToVariable := map[string]VariableMap{}

	var workspaceNameList []string
	for workspace := range workspaceToVarSetIDs {
		workspaceNameList = append(workspaceNameList, workspace)
	}
	sort.Strings(workspaceNameList)

	for _, workspace := range workspaceNameList {
		varSetIDs := workspaceToVarSetIDs[workspace]
		currentVarMap := VariableMap{}

		for varSetID := range varSetIDs {
			currentVarMap = currentVarMap.Merge(varSetVars[varSetID])
		}

		outputWorkspaceToVariable[workspace] = currentVarMap
	}

	return outputWorkspaceToVariable, nil
}

// PullWorkspaceVariables extracts variables for a single workspace saves into a .tfvars
// file within the appropriate directory.
func (tfc *tfCloud) PullWorkspaceVariables(
	ctx context.Context,
	workspaceName string,
	workspaceToVarSetVars map[string]VariableMap,
	workspaceToVarSetIDs map[string]map[string]bool,
	varSetIDsToName map[string]string,
) error {
	workspaceVarsContainer, err := tfc.DownloadWorkspaceVariables(ctx, workspaceName)
	if err != nil {
		return fmt.Errorf("[tfc.DownloadWorkspaceVariables] %v", err)
	}

	workspaceVarsMap, err := tfc.extractWorkspaceVars(workspaceVarsContainer)
	if err != nil {
		return fmt.Errorf("[tfc.parseWorkspaceVars] %v", err)
	}

	workspaceSensitiveEnvMap, workspaceSensitiveVarsMap, err := tfc.createWorkspaceSensitiveVars(
		workspaceName, workspaceToVarSetIDs, varSetIDsToName,
	)
	if err != nil {
		return fmt.Errorf("[tfc.workspaceSensitiveVars] %v", err)
	}

	tfVarsFile, err := tfc.generateTFVarsFile(
		workspaceVarsMap, workspaceToVarSetVars[workspaceName], workspaceSensitiveVarsMap,
	)
	if err != nil {
		return fmt.Errorf("[tfc.generateTFVarsFile] %v", err)
	}

	err = tfc.updateEnvironmentVariables(workspaceSensitiveEnvMap)
	if err != nil {
		return fmt.Errorf("[tfc.updateEnvironmentVariables] %v", err)
	}

	fileName := fmt.Sprintf(
		"/github/workspace%vterraform.tfvars",
		tfc.config.WorkspaceToDirectory[workspaceName],
	)

	err = os.WriteFile(fileName, tfVarsFile, 0400)
	if err != nil {
		return fmt.Errorf("[os.WriteFile] %v", err)
	}

	return nil
}

// DownloadWorkspaceVariables downloads a workspace's variables from the remote source.
func (tfc *tfCloud) DownloadWorkspaceVariables(ctx context.Context, workspaceName string) ([]byte, error) {
	workspaceID, err := tfc.getWorkspaceID(ctx, workspaceName)
	if err != nil {
		return nil, fmt.Errorf("[tfc.getWorkspaceID] %v", err)
	}

	varsJSON, err := tfc.getWorkspaceVariables(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("[tfc.getWorkspaceVariables] %v", err)
	}

	return varsJSON, nil
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

// extractWorkspaceVars extracts workspace variables from a []byte from the Terraform Cloud
// endpoint and places them into a VariableMap.
func (tfc *tfCloud) extractWorkspaceVars(workspaceResponse []byte) (VariableMap, error) {
	container, err := gabs.ParseJSON(workspaceResponse)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	outputVarMap := VariableMap{}

	i := 0
	for container.Exists("data", strconv.Itoa(i)) {
		varKey := container.Search("data", strconv.Itoa(i), "attributes", "key").Data().(string)
		value := container.Search("data", strconv.Itoa(i), "attributes", "value").Data()

		if value == nil {
			i++
			continue
		}
		varValue := container.Search("data", strconv.Itoa(i), "attributes", "value").Data().(string)
		outputVarMap[varKey] = varValue
		i++
	}

	return outputVarMap, nil
}

// createWorkspaceSensitiveVars produces collections of sensitive workspace variables.
func (tfc *tfCloud) createWorkspaceSensitiveVars(
	workspaceName string,
	workspaceToVarSetIDs map[string]map[string]bool,
	varSetIDToName map[string]string,
) (VariableMap, VariableMap, error) {
	allVariablesEnv := VariableMap{}
	allVariablesTerraform := VariableMap{}

	for varSetID := range workspaceToVarSetIDs[workspaceName] {
		varSetName := varSetIDToName[varSetID]
		if _, ok := tfc.config.TerraformVarSetSensitiveVars[varSetName]; !ok {
			continue
		}
		currentVarSetSensitiveVars := tfc.config.TerraformVarSetSensitiveVars[varSetName]
		currentVarMapEnv, currentVarMapTerraform, err := tfc.variablesToVariableMaps(currentVarSetSensitiveVars)
		if err != nil {
			return nil, nil, fmt.Errorf("[tfc.variablesToVariableMaps] %v", err)
		}

		allVariablesTerraform = allVariablesTerraform.Merge(currentVarMapTerraform)
		allVariablesEnv = allVariablesEnv.Merge(currentVarMapEnv)
	}

	workspaceSensitiveVars := tfc.config.TerraformWorkspaceSensitiveVars[workspaceName]
	workspaceVarMapEnv, workspaceVarMapTerraform, err := tfc.variablesToVariableMaps(workspaceSensitiveVars)
	if err != nil {
		return nil, nil, fmt.Errorf("[tfc.variablesToVariableMaps] %v", err)
	}
	allVariablesEnv = allVariablesEnv.Merge(workspaceVarMapEnv)
	allVariablesTerraform = allVariablesTerraform.Merge(workspaceVarMapTerraform)
	return allVariablesEnv, allVariablesTerraform, nil
}

// variablesToVariableMaps converts a sensitive variables object to two VariableMaps,
// one for variables to be used within terraform and others for environment variables
func (tfc *tfCloud) variablesToVariableMaps(vars Variables) (VariableMap, VariableMap, error) {
	varMapEnv := VariableMap{}
	varMapTerraform := VariableMap{}

	for varKey, varData := range vars {
		switch varData.category {
		case "env":
			varMapEnv[varKey] = varData.value
		case "terraform":
			varMapTerraform[varKey] = varData.value
		default:
			return nil, nil, fmt.Errorf(
				"sensitive variables must have a category of either `env` or `terraform`, got: %v",
				varData.value,
			)
		}
	}

	return varMapEnv, varMapTerraform, nil
}

// generateTFVarsFile aggregates varset variables and workspace-specific variables to create a
// .tfvars file for the current workspace.
func (tfc *tfCloud) generateTFVarsFile(
	workspaceVars VariableMap,
	workspaceVarSetVars VariableMap,
	workspaceSensitiveVars VariableMap,
) ([]byte, error) {
	// workspace variables assigned to the variable itself has priority in Terraform cloud,
	// which is reflected here
	workspaceVariableMap := workspaceVarSetVars.Merge(workspaceVars)
	workspaceCompleteVariableMap := workspaceVariableMap.Merge(workspaceSensitiveVars)

	f := hclwrite.NewEmptyFile()
	body := f.Body()

	allKeys := []string{}

	for k := range workspaceCompleteVariableMap {
		allKeys = append(allKeys, k)
	}

	// This sorting is helpful for cleaner output and allows unit tests to be deterministic.
	sort.Strings(allKeys)

	for _, k := range allKeys {
		if workspaceCompleteVariableMap[k] == "null" {
			fmt.Printf(
				"null value has been specified for variable %v - this variable might need to be specified as a sensitive variable",
				k,
			)
		}
		body.SetAttributeValue(k, cty.StringVal(workspaceCompleteVariableMap[k]))
	}

	return f.Bytes(), nil
}

// updateEnvironmentVariables
func (tfc *tfCloud) updateEnvironmentVariables(workspaceSensitiveEnvMap VariableMap) error {
	for k, v := range workspaceSensitiveEnvMap {
		err := os.Setenv(k, v)
		if err != nil {
			return fmt.Errorf("[os.Setenv] %v", err)
		}
	}
	return nil
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
