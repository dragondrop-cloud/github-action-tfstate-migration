package tfvars

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// TFVars is an interface that allows for the extraction of
// terraform variables from a remote source.
type TFVars interface {

	// PullAllWorkspaceVariables extracts variables for all workspaces and saves into .tfvars files.
	PullAllWorkspaceVariables() error

	// PullWorkspaceVariables extracts variables for a single workspace saves into a .tfvars file.
	PullWorkspaceVariables() error
}

// Config contains the variables needed to support the TFVars interface.
type Config struct {

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

// tfVars is a struct that implements the TFVars interface.
type tfFVars struct {
}

// NewTFVars instantiates a new implementation of the tfVars interface.
func NewTFVars() TFVars {
	return tfCloud{}
}
