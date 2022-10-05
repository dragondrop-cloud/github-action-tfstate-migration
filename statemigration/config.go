package statemigration

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Version is a type representing a Terraform Version.
type Version string

// Config contains environment variables needed to run StateMigrator methods.
type Config struct {

	// TerraformCloudOrganization is the name of the terraform cloud organization where state is maintained.
	TerraformCloudOrganization string `required:"true"`

	// TerraformCloudToken is a token to access terraform cloud remote state.
	TerraformCloudToken string `required:"true"`

	// TerraformVersion is the default version of terraform to use for migrations. It is optional.
	TerraformVersion Version `required:"true"`

	// IsApply is a Boolean of whether to run `tfmigrate apply` ("true") or
	// `tfmigrate plan` ("false") for the migrations.
	IsApply bool `required:"true"`

	// WorkspaceToDirectory is a map between workspace name and the relative directory
	// for a workspace's configuration.
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

func (v *Version) Decode(value string) error {
	if string(value[1]) != "." {
		return fmt.Errorf("terraform version should start with 'major version[.]'")
	}

	stringComponents := strings.Split(value, ".")
	versionLength := len(stringComponents)
	if versionLength != 3 {
		return fmt.Errorf("expected three pieces of the version once split by '.', instead got %v", versionLength)
	}

	majorVersion := stringComponents[0]

	if (majorVersion != "0") && (majorVersion != "1") {
		return fmt.Errorf("terraform major version must be either '0' or '1', got %v", majorVersion)
	}

	*v = Version(value)
	return nil
}
