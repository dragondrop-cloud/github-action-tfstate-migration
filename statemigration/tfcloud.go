package statemigration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Jeffail/gabs/v2"
)

// RunStatus is a struct containing information on a workspace run as is required to determine whether
// to cancel or discard a run if possible.
type RunStatus struct {

	// isCancelable is whether a run can be canceled.
	isCancelable bool

	// isDiscardable is whether a run can be discarded.
	isDiscardable bool

	// isPostConfirmation is whether a run is post user-confirmation that a plan should be applied.
	isPostConfirmation bool

	// runID is the Terraform Cloud ID for a workspace run.
	runID string
}

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
	// Get a list of all active/pending runs
	requestName := "getMostRecentRuns"
	requestPath := fmt.Sprintf("https://app.terraform.io/api/v2/workspaces/%v/runs", workspaceID)

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath)

	if err != nil {
		return fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	jsonResponseBytes, err := sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return err
	}

	runStatusSlices, err := extractRecentRunStatuses(jsonResponseBytes)
	if err != nil {
		return fmt.Errorf("[extractRecentRunStatuses] %v", err)
	}

	// TODO
	fmt.Println(runStatusSlices)

	for _, runStatus := range runStatusSlices {
		if runStatus.isPostConfirmation {
			fmt.Println(
				"There is an unfinished run that is post-confirmation. We do not discard those runs, ending " +
					"job execution.")
		}
	}
	// TODO: For each run in list, apply relevant operation (discard, cancel, or error message)
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#create-a-run
	return nil
}

// TODO: Unit test this function
// extractRecentRunStatuses extracts statuses for recent runs in the workspace.
func extractRecentRunStatuses(jsonResponseBytes []byte) ([]RunStatus, error) {
	jsonParsed, err := gabs.ParseJSON(jsonResponseBytes)
	if err != nil {
		return nil, fmt.Errorf("[gabs.ParseJSON] %v", err)
	}

	var runStatusSlice []RunStatus

	i := 0

	for jsonParsed.Exists("data", strconv.Itoa(i)) {

		status := jsonParsed.Search("data", strconv.Itoa(i), "attributes", "status").Data().(string)

		// checking if the status is in a terminal state, if so, continue and skip the next iteration
		if isStatusTerminalState(status) {
			i++
			continue
		}

		isCancelable := jsonParsed.Search("data", strconv.Itoa(i), "attributes", "actions", "is-cancelable").Data().(bool)

		isDiscardable := jsonParsed.Search("data", strconv.Itoa(i), "attributes", "actions", "is-discardable").Data().(bool)

		runID := jsonParsed.Search("data", strconv.Itoa(i), "id").Data().(string)

		isPostConfirmation := isStatusPostConfirmation(status)

		currentRS := RunStatus{
			isCancelable:       isCancelable,
			isDiscardable:      isDiscardable,
			isPostConfirmation: isPostConfirmation,
			runID:              runID,
		}

		runStatusSlice = append(runStatusSlice, currentRS)

		i++
	}

	return runStatusSlice, nil
}

// isStatusTerminalState checks to see if the received status indicates that a job is in a terminal
// state.
func isStatusTerminalState(status string) bool {
	terminalStateSet := map[string]bool{
		"policy_soft_failed":   true,
		"planned_and_finished": true,
		"applied":              true,
		"discarded":            true,
		"errored":              true,
		"canceled":             true,
		"force_canceled":       true,
	}

	return terminalStateSet[status]
}

// isStatusPostConfirmation checks to see if the received status indicates that a job
// has been approved after a plan has executed.
func isStatusPostConfirmation(status string) bool {
	postConfirmationSet := map[string]bool{
		"confirmed":           true,
		"post_plan_running":   true,
		"post_plan_completed": true,
		"apply_queued":        true,
		"applying":            true,
	}

	return postConfirmationSet[status]
}

// TODO: Implement and unit test
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
