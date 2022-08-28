package statemigration

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config contains environment variables needed to run StateMigrator methods.
type Config struct {

	// AWSAccessKeyID is an AWS access key id.
	AWSAccessKeyID string `required:"false"`

	// AWSSecretAccessKey is an AWS secret access key.
	AWSSecretAccessKey string `required:"false"`

	// GoogleApplicationCredentials is a set of Google Default Application Credentials.
	GoogleApplicationCredentials string `required:"false"`

	// IsApply is a Boolean of whether to run `tfmigrate apply` ("true") or
	// `tfmigrate plan` ("false") for the migrations.
	IsApply bool `required:"true"`

	// WorkspaceDirectories is a list of strings. Each string is relative file path to a directory
	// which is tied to a single Terraform Cloud workspace.
	WorkspaceDirectories []string `required:"true"`
}

func NewConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)

	if err != nil {
		return nil, fmt.Errorf("[envconfig.Process] Error loading config: %v", err)
	}
	return &c, err
}
