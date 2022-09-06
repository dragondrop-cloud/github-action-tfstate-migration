package tfvars

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/kelseyhightower/envconfig"
)

// VariableMap is a collection of variable key value pairs stored within a map.
type VariableMap map[string]string

// TFVars is an interface that allows for the extraction of
// terraform variables from a remote source.
type TFVars interface {

	// DownloadWorkspaceVariables downloads a workspace's variables from the remote source.
	DownloadWorkspaceVariables(ctx context.Context, workspaceName string) (*gabs.Container, error)

	// CreateAllWorkspaceVarsFiles extracts variables for all workspaces and saves them into
	// .tfvars files within the appropriate directory.
	CreateAllWorkspaceVarsFiles() error
}

// Config contains the variables needed to support the TFVars interface.
type Config struct {

	// TerraformCloudOrganization is the name of the Terraform Cloud organization
	TerraformCloudOrganization string `required:"false"`

	// TerraformCloudToken is a Terraform Cloud Token
	TerraformCloudToken string `required:"true"`

	// WorkspaceToDirectory is a map between workspace name and the relative directory for a workspace's
	// configuration.
	WorkspaceToDirectory map[string]string `required:"true"`
}

// NewConfig instantiates a new instance of Config
func NewConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)

	if err != nil {
		return nil, fmt.Errorf("[envconfig.Process] Error loading config: %v", err)
	}

	return &c, err
}

// NewTFVars instantiates a new implementation of the tfVars interface.
func NewTFVars() (TFVars, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("[NewConfig] %v", err)
	}

	return &tfCloud{
		config:     conf,
		httpClient: http.Client{},
	}, nil
}
