package statemigration

import (
	"bytes"
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

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath, nil)

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

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "GET", requestPath, nil)

	if err != nil {
		return fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	jsonResponseBytes, err := sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return err
	}

	runStatusSlice, err := extractRecentRunStatuses(jsonResponseBytes)
	if err != nil {
		return fmt.Errorf("[extractRecentRunStatuses] %v", err)
	}

	for _, runStatus := range runStatusSlice {
		if runStatus.isPostConfirmation {
			fmt.Println(
				"There is an unfinished run that is post-confirmation. We do not discard those runs, ending " +
					"job execution.")
		}
	}

	for _, runStatus := range runStatusSlice {
		if runStatus.isDiscardable {
			err = sm.discardRun(ctx, runStatus.runID)
			if err != nil {
				return fmt.Errorf("[sm.discardRun] %v", err)
			}
		} else if runStatus.isCancelable {
			err = sm.cancelRun(ctx, runStatus.runID)
			if err != nil {
				return fmt.Errorf("[sm.cancelRun] %v", err)
			}
		}
	}

	return nil
}

// TODO: Add unit test if possible
// cancelRun cancels the run specified by runID.
func (sm *stateMigrator) cancelRun(ctx context.Context, runID string) error {
	requestName := "cancelRun"
	requestPath := fmt.Sprintf("https://app.terraform.io/api/v2/runs/%v/actions/discard", runID)

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "POST", requestPath, nil)

	if err != nil {
		return fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	_, err = sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return err
	}
	return nil
}

// discardRun discards the run specified by runID.
func (sm *stateMigrator) discardRun(ctx context.Context, runID string) error {
	requestName := "discardRun"
	requestPath := fmt.Sprintf("https://app.terraform.io/api/v2/runs/%v/actions/discard", runID)

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "POST", requestPath, nil)

	if err != nil {
		return fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}

	_, err = sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return err
	}
	return nil
}

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

// createPlanOnlyRefreshRun kicks off a new plan-only, refresh-state run for the workspace.
func (sm *stateMigrator) createPlanOnlyRefreshRun(ctx context.Context, workspaceID string) error {
	requestPath := "https://app.terraform.io/api/v2/runs"

	payload, err := generateRefreshOnlyPlanPayload(workspaceID)
	if err != nil {
		return fmt.Errorf("[generateRefreshOnlyPlanPayload] %v", err)
	}

	requestName := "createPlanOnlyRefreshRun"

	request, err := sm.buildTFCloudHTTPRequest(ctx, requestName, "POST", requestPath, bytes.NewBuffer(payload))

	if err != nil {
		return fmt.Errorf("[%v] error in newRequest: %v", requestName, err)
	}
	_, err = sm.terraformCloudRequest(request, requestName)

	if err != nil {
		return err
	}
	return nil
}

// generateRefreshOnlyPlanPayload builds the JSON payload needed to
// run a refresh-only, plan-only, run within Terraform cloud for the specified workspaceID.
func generateRefreshOnlyPlanPayload(workspaceID string) ([]byte, error) {
	jsonObj := gabs.New()

	_, err := jsonObj.Set("runs", "data", "type")
	if err != nil {
		return nil, fmt.Errorf("[data: type:]%v", err)
	}

	_, err = jsonObj.Set(true, "data", "attributes", "refresh-only")
	if err != nil {
		return nil, fmt.Errorf("[data: attributes: refresh-only:]%v", err)
	}

	_, err = jsonObj.Set(true, "data", "attributes", "plan-only")
	if err != nil {
		return nil, fmt.Errorf("[data: attributes: plan-only:]%v", err)
	}

	_, err = jsonObj.Set("workspaces", "data", "relationships", "workspace", "data", "type")
	if err != nil {
		return nil, fmt.Errorf("[data: relationships: workspace: type:]%v", err)
	}

	_, err = jsonObj.Set(workspaceID, "data", "relationships", "workspace", "data", "id")
	if err != nil {
		return nil, fmt.Errorf("[data: relationship: workspace: id:]%v", err)
	}

	return jsonObj.Bytes(), nil
}

// buildTFCloudHTTPRequest structures a request to the Terraform Cloud api.
func (sm *stateMigrator) buildTFCloudHTTPRequest(ctx context.Context, requestName string, method string, requestPath string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestPath, body)
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
	if !(response.StatusCode <= 299) {
		return nil, fmt.Errorf("[%v] was unsuccessful, with the server returning: %v", requestName, response.StatusCode)
	}

	// Read in response body to bytes array.
	outputBytes, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("[%v] error in reading response into bytes array: %v", requestName, err)
	}

	return outputBytes, nil
}
