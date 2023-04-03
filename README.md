# dragondrop tfstate migration action
A GitHub Action for running dragondrop-built state migrations.

## Example Usage
### Migrations for AWS with history stored in s3
```yaml
name: infrastructure migration
on: push

jobs:
  new-resource-migration:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    permissions:
      contents: "read"
      id-token: "write"

    steps:
      - name: Checkout branch
        uses: actions/checkout@v3

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: "arn:aws:iam::MY_PROJECT_ID:role/ROLE_FOR_TERRAFORM_PLAN"
          aws-region: "MY_REGION"

      - name: Plan Migration of Remote State
        uses: "dragondrop-cloud/github-action-tfstate-migration@latest"
        if: ${{ github.ref_name != 'dev' && github.ref_name != 'prod'}}
        with:
          is-apply: false
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-workspace-sensitive-vars: ${{ secrets.TERRAFORM_WORKSPACE_SENSITIVE_VARS }}
          terraform-var-set-sensitive-vars: ${{ secrets.TERRAFORM_VAR_SET_SENSITIVE_VARS }}
          terraform-cloud-variable-name: "tfe_token"
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"

      - name: Apply Migration of Remote State - Dev
        uses: dragondrop-cloud/github-action-tfstate-migration@latest
        if: ${{ github.ref_name == 'dev'}}
        with:
          is-apply: true
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-workspace-sensitive-vars: ${{ secrets.TERRAFORM_WORKSPACE_SENSITIVE_VARS }}
          terraform-var-set-sensitive-vars: ${{ secrets.TERRAFORM_VAR_SET_SENSITIVE_VARS }}
          terraform-cloud-variable-name: "tfe_token"
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"
```

### Migrations on Terraform for GCP with history stored in GCP Cloud Storage
```yaml
name: infrastructure migration
on: push

jobs:
  new-resource-migration:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    permissions:
      contents: "read"
      id-token: "write"

    steps:
      - name: Checkout branch
        uses: actions/checkout@v3

      - name: Authenticate to Google Cloud - Dev
        id: auth
        uses: google-github-actions/auth@v0
        if: ${{ github.ref_name != 'prod'}}
        with:
          token_format: "access_token"
          create_credentials_file: 'true'
          export_environment_variables: 'true'
          service_account: 'my-service-account-name@my-gcp-project.iam.gserviceaccount.com'
          workload_identity_provider: 'projects/8675309/locations/global/workloadIdentityPools/my-workload-id-pool/providers/my-provider'

      - name: Plan Migration of Remote State
        uses: "dragondrop-cloud/github-action-tfstate-migration@latest"
        if: ${{ github.ref_name != 'dev' && github.ref_name != 'prod'}}
        with:
          is-apply: false
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-workspace-sensitive-vars: ${{ secrets.TERRAFORM_WORKSPACE_SENSITIVE_VARS }}
          terraform-var-set-sensitive-vars: ${{ secrets.TERRAFORM_VAR_SET_SENSITIVE_VARS }}
          terraform-cloud-variable-name: "tfe_token"
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"

      - name: Apply Migration of Remote State - Dev
        uses: dragondrop-cloud/github-action-tfstate-migration@latest
        if: ${{ github.ref_name == 'dev'}}
        with:
          is-apply: true
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-workspace-sensitive-vars: ${{ secrets.TERRAFORM_WORKSPACE_SENSITIVE_VARS }}
          terraform-var-set-sensitive-vars: ${{ secrets.TERRAFORM_VAR_SET_SENSITIVE_VARS }}
          terraform-cloud-variable-name: "tfe_token"
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"
```

## Inputs

### `is-apply`
**Required** Whether to attempt to apply migration statements
found by the action. If `"false"`, will only run a "plan" if the migrations will be successful.

Defaults to `"false"`.

### `terraform-cloud-organization`
**Required** Name of the Terraform Cloud organization against which migrations are to be run.

### `terraform-cloud-token`
**Required** Terraform Cloud API token with access to the specified `terraform-cloud-organization`.

### `terraform-workspace-sensitive-vars`:
Mapping between workspaces to sensitive variables, matching the parameterization of a
variable as specified within Terraform Cloud.

Example:
```"{
    "my_workspace_one": {
        "workspace_var_one": {
            "value": "",
            "category": "terraform"
        }
    }
}"
```

### `terraform-var-set-sensitive-vars`:
Mapping between variable sets to sensitive variables, matching the parameterization of a
variable as specified within Terraform Cloud.

Example:
```
"{
    "my_var_set_one": {
        "tfe_token": {
            "value": "zyha",
            "category": "terraform"
        }
    }
}"
```

### `terraform-version`
**Required** The Terraform version to use within the job. Must only be the numerical version ('1.2.3' is valid, '~>1.2.3' is not).

Example: `"1.2.3"`

Defaults to `""`

### `workspace-to-directories`
**Required** A map between workspace names and the relative path to that workspace's terraform definition.

Example: `"workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"`

There is no default value for this input.

## Outputs
None
