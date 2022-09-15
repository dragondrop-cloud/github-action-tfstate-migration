package tfvars

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// TFVars is an interface that allows for the extraction of
// terraform variables from a remote source.
type TFVars interface {

	// DownloadWorkspaceVariables downloads a workspace's variables from the remote source.
	DownloadWorkspaceVariables(ctx context.Context, workspaceName string) ([]byte, error)

	// CreateAllWorkspaceVarsFiles extracts variables for all workspaces and saves them into
	// .tfvars files within the appropriate directory.
	CreateAllWorkspaceVarsFiles() error
}

// NewTFVars instantiates a new implementation of the tfVars interface.
func NewTFVars() (TFVars, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("[NewConfig] %v", err)
	}

	// This allows the terraform command to make calls to Terraform Cloud
	err = os.Setenv("TF_TOKEN_app_terraform_io", conf.TerraformCloudToken)
	if err != nil {
		return nil, fmt.Errorf("[os.Setenv] %v", err)
	}

	return &tfCloud{
		config:     conf,
		httpClient: http.Client{},
	}, nil
}
