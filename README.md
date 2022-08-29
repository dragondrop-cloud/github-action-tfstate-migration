# dragondrop tfstate migration action
A GitHub Action for running dragondrop-built state migrations.

## Example Usage
### Migrations with history stored in s3
```yaml
jobs:
  new-resource-migration:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout branch
        uses: actions/checkout@v3

      - name: Plan Migration of Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration
        if: ${{ github.ref_name != 'dev'}}
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          is-apply: false
          terraform-version: "1.2.6"
          workspace-directories: "my/relative/directory/1,my/relative/directory/2"

      - name: Apply Migration of Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration
        if: ${{ github.ref_name == 'dev'}}
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          is-apply: true
          terraform-version: "1.2.6"
          workspace-directories: "my/relative/directory/1,my/relative/directory/2"
```

### Migrations with history stored in GCP
```yaml
jobs:
  new-resource-migration:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout branch
        uses: actions/checkout@v3

      - id: auth
        name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v0
        with:
          create_credentials_file: 'true'
          export_environment_variables: 'true'
          service_account: 'my-service-account@my-project.iam.gserviceaccount.com'
          workload_identity_provider: 'projects/my-project-id/locations/global/workloadIdentityPools/my-workload-identity-pool/providers/my-provider'

      - name: Plan Migration of Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration
        if: ${{ github.ref_name == 'dev'}}
        with:
          google-application-credentials: ${{ auth.GOOGLE_APPLICATION_CREDENTIALS }}
          is-apply: false
          terraform-version: "1.2.6"
          workspace-directories: "my/relative/directory/1,my/relative/directory/2"

      - name: Apply Migration of Remote State
        uses: dragondrop-cloud/github-action-tfstate-migration
        if: ${{ github.ref_name == 'prod'}}
        with:
          google-application-credentials: ${{ auth.GOOGLE_APPLICATION_CREDENTIALS }}
          is-apply: true
          terraform-version: "1.2.6"
          workspace-directories: "my/relative/directory/1,my/relative/directory/2"
```


## Inputs

### `aws-access-key-id`
AWS_ACCESS_KEY_ID, used for authentication with AWS services.

### `aws-secret-access-key`
AWS_SECRET_ACCESS_KEY, used for authentication with AWS services.

### `google-application-credentials`
GOOGLE_APPLICATION_CREDENTIALS, used for authentication with GCP services.

### `is-apply`
**Required** Whether to attempt to apply migration statements
found by the action. If `"false"`, will only run a "plan" if the migrations will be successful.

Defaults to `"false"`. 

### `terraform-version`
**Required** The Terraform version to use within the job.

Defaults to `"latest"`

### `workspace-directories`
**Required** A list of relative directory paths within the repo,
each one of which is associated with its own Terraform workspace.

Example: `"path/to/workspace/1,path/to/workspace/2,path/to/workspace/3"`

There is no default value for this input.

## Outputs
None
