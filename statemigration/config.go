package statemigration

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config contains environment variables needed to run StateMigrator methods.
type Config struct {

	// TerraformCloudOrganization is the name of the terraform cloud organization where state is maintained.
	TerraformCloudOrganization string `required:"true"`

	// TerraformCloudToken is a token to access terraform cloud remote state.
	TerraformCloudToken string `required:"true"`

	// IsApply is a Boolean of whether to run `tfmigrate apply` ("true") or
	// `tfmigrate plan` ("false") for the migrations.
	IsApply bool `required:"true"`

	// WorkspaceToDirectory is a map between workspace name and the relative directory for a workspace's
	// configuration.
	WorkspaceToDirectory map[string]string `required:"true"`
}

// NewConfig instantiates a new instance of the Config struct.
func NewConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)

	if err != nil {
		return nil, fmt.Errorf("[envconfig.Process] Error loading config: %v", err)
	}

	return &c, err
}
