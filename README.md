# dragondrop tfstate migration action
A GitHub Action for running dragondrop-built state migrations.

## Example Usage
### Migrations with history stored in s3
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

    env:
      AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
      AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

    steps:
      - name: Checkout branch
        uses: actions/checkout@v3

      - name: Plan Migration of Remote State
        uses: "dragondrop-cloud/github-action-tfstate-migration@latest"
        if: ${{ github.ref_name != 'dev' && github.ref_name != 'prod'}}
        with:
          is-apply: false
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"

      - name: Apply Migrations to Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration@latest
        if: ${{ github.ref_name == 'dev'}}
        with:
          is-apply: true
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_API_TOKEN }}
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"
```

### Migrations with history stored in GCP
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

      - name: Authenticate to Google Cloud
        id: auth
        uses: google-github-actions/auth@v0
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
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_TOKEN }}
          terraform-version: "1.2.6"
          workspace-to-directories: "workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"

      - name: Apply Migration of Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration@latest
        if: ${{ github.ref_name == 'dev'}}
        with:
          is-apply: true
          terraform-cloud-organization: ${{ secrets.TERRAFORM_CLOUD_ORG }}
          terraform-cloud-token: ${{ secrets.TERRAFORM_CLOUD_TOKEN }}
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

### `terraform-version`
**Required** The Terraform version to use within the job.

Defaults to `"latest"`

### `workspace-to-directories`
**Required** A map between workspace names and the relative path to that workspace's terraform definition.

Example: `"workspace_1:/my/relative/directory/1/,workspace_2:/my/relative/directory/2/"`

There is no default value for this input.

## Outputs
None
