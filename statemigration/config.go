package statemigration

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/kelseyhightower/envconfig"
)

// GroupToVariables is a mapping between a group name and Variables associated with that group.
type GroupToVariables map[string]Variables

// Variables is a mapping between variable keys and VariableData
type Variables map[string]VariableData

// VariableData represents the information needed for a Terraform Cloud variable.
type VariableData struct {

	// value is the variables value.
	value string

	// category is the variable category and can either be "env", meaning it is an environment
	// variable, or "terraform" meaning the variable is to be read directly into Terraform HCL code.
	category string
}

// Config contains environment variables needed to run StateMigrator methods.
type Config struct {

	// TerraformCloudOrganization is the name of the terraform cloud organization where state is maintained.
	TerraformCloudOrganization string `required:"true"`

	// TerraformCloudToken is a token to access terraform cloud remote state.
	TerraformCloudToken string `required:"true"`

	// TerraformWorkspaceSensitiveVars is a mapping between a Terraform Cloud workspace and sensitive
	// variables associated with that workspace.
	TerraformWorkspaceSensitiveVars GroupToVariables `required:"false"`

	// TerraformVarSetSensitiveVars is a mapping between a Terraform Cloud variable set and sensitive
	// variables associated with that variable set.
	TerraformVarSetSensitiveVars GroupToVariables `required:"false"`

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

// Decode parses a string variable into the format needed for a GroupToVariables
// object.
func (gtv *GroupToVariables) Decode(value string) error {
	parsedJSON, err := gabs.ParseJSON([]byte(value))
	if err != nil {
		return fmt.Errorf("[gabs.ParseJSON] Error parsing JSON: %v", err)
	}

	groupToVars := GroupToVariables{}

	for group, variables := range parsedJSON.ChildrenMap() {
		groupToVars[group] = Variables{}
		for varKey, variableData := range variables.ChildrenMap() {
			var value string
			var category string

			// extracting value
			if variableData.Exists("value") {
				value, _ = variableData.Search("value").Data().(string)
			} else {
				return fmt.Errorf(
					"no specified 'value' field in grouping %v for key %v",
					group, varKey,
				)
			}

			// extracting category
			if variableData.Exists("category") {
				category, _ = variableData.Search("category").Data().(string)
			} else {
				return fmt.Errorf(
					"no specified 'category' field in grouping %v for key %v",
					group, varKey,
				)
			}

			if category != "env" && category != "terraform" {
				return fmt.Errorf(
					"category must be either 'env' or 'terraform'. In grouping %v for key %v recieved category of %v",
					group, varKey, category,
				)
			}

			groupToVars[group][varKey] = VariableData{
				value:    value,
				category: category,
			}

		}
	}

	return nil
}
