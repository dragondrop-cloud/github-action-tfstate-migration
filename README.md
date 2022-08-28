# dragondrop tfstate migration action
A GitHub Action for running dragondrop-built state migrations.

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

Defaults to `"1.2.6"`

### `workspace-directories`
**Required** A list of relative directory paths within the repo,
each one of which is associated with its own Terraform workspace.

Example: `"path/to/workspace/1,path/to/workspace/2,path/to/workspace/3"`

There is no default value for this input.

## Outputs
None

## Example Usage
### Migrations with history stored in s3
```yaml

```

### Migrations with history stored in GCP
```yaml

```
