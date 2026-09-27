# Security checks and image publishing

The planned Vault database secret and Consul discovery exercises are described
in the [service discovery and secrets flow](service-discovery-and-secrets.md).
No Vault token, database credential, or Consul ACL token belongs in GitHub
Actions variables, repository files, container images, or Terraform state.

CI runs on every push: test the Go code, build both runtime images, and scan
them. SonarQube analysis runs on `main` pushes so the workflow also works on
the SonarQube Cloud Free plan, which does not support ordinary feature-branch
analysis. After a pull request merges into
`main`, the separate CD workflow verifies that the resulting commit passed CI
before publishing images. The ECR workflow is disabled until the AWS
prerequisites are in place.

## 1. Repeat the local Trivy scan

The official Trivy image can scan an exported image without access to the
Docker daemon. This keeps the scanner's privileges narrow.

```bash
export POSTGRES_PASSWORD=localdev
docker compose build api migrate
docker save -o api-image.tar cloud-infrastructure-platform-api:latest
docker save -o migrate-image.tar cloud-infrastructure-platform-migrate:latest
docker volume create cloud_infra_trivy_cache
docker run --rm -v "$PWD/api-image.tar:/scan/image.tar:ro" \
  -v cloud_infra_trivy_cache:/root/.cache/ aquasec/trivy:0.74.0 \
  image --input /scan/image.tar --scanners vuln \
  --severity HIGH,CRITICAL --exit-code 1
docker run --rm -v "$PWD/migrate-image.tar:/scan/image.tar:ro" \
  -v cloud_infra_trivy_cache:/root/.cache/ aquasec/trivy:0.74.0 \
  image --input /scan/image.tar --scanners vuln \
  --severity HIGH,CRITICAL --exit-code 1
rm api-image.tar migrate-image.tar
```

Exit code 0 means the selected gate passed; it does not mean the image can
never have vulnerabilities. Review lower-severity findings separately. The
GitHub CI workflow runs the same gate on both images. See the recorded
[local scan evidence](evidence/container-security.md).

## 2. Connect SonarQube Cloud

This public repository is now imported into the `shivamshashank` organization
as project `shivamshashank_cloud-infrastructure-platform`. The verified keys
are in [`sonar-project.properties`](../sonar-project.properties). In the
project's **Administration → Analysis Method**, disable Automatic Analysis and
select CI-based analysis. SonarQube Cloud does not support running both
methods concurrently.

In GitHub repository **Settings → Secrets and variables → Actions**, create:

| Type | Name | Value |
|---|---|---|
| Secret | `SONAR_TOKEN` | Token from the SonarQube Cloud project |

The CI action reads `sonar-project.properties`, including
`sonar.go.coverage.reportPaths=coverage.out`. The test step generates that Go
coverage file before the analysis. Once `SONAR_TOKEN` is set, the workflow
waits for the SonarQube quality gate. The first CI-based analysis uploaded
coverage successfully (SonarQube displayed 50.4%) but failed the new-code
security rating. The scan identified two new workflow findings: an unnecessary
`POSTGRES_PASSWORD` variable in the CI image-build step and privileged
`workflow_run` code checkout in the ECR workflow. The current changes remove
that build variable and switch CD to a merge-only main push with an explicit
CI-success check. These changes need a new scan before the gate can be
considered passing. If `SONAR_TOKEN` is missing, CI prints a warning and
skips analysis; when ECR publishing is enabled, a missing token fails CI.

SonarSource currently documents full Go support through 1.25, while this
project uses Go 1.27.1. Check the first analysis for parser warnings or
missing files before treating its findings as complete.

Codecov is separate: its existing CI upload needs `CODECOV_TOKEN` in GitHub
Actions secrets.

## 3. Prepare ECR Public in `us-east-1`

This project publishes only the API and migration binaries, with no runtime
credentials baked into the images. **Anyone can pull these images.** ECR
Public is a better cost fit for this public learning project: [AWS currently
includes 50 GB of public repository storage per month in its always-free
tier](https://aws.amazon.com/ecr/pricing/). Storage or transfer beyond the
published free limits can be charged. The previous workflow targeted ECR
Private, whose free storage allowance is different.

The [Terraform configuration](../infra/ecr-public/main.tf) creates a public
repository, a GitHub Actions OIDC provider if the account does not already
have one, and a publish role limited to this repository and the `main` branch.
It does not create a server or deploy the application. The local AWS CLI has
no credentials configured yet, so this infrastructure has **not** been
applied.

1. Sign in to the intended AWS account with a CLI profile that can create ECR
   Public repositories and IAM resources. This machine's AWS CLI supports
   [short-term console sign-in](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sign-in.html).
   Use your IAM Identity Center profile instead if that is how your account is
   configured. Confirm the account before planning:

   ```bash
   aws login --profile cloud-lab --region us-east-1
   export AWS_PROFILE=cloud-lab
   aws sts get-caller-identity
   ```

2. Change to `infra/ecr-public` and initialize Terraform. Check whether the
   account already has an OIDC provider for
   `token.actions.githubusercontent.com`. If it does, pass its ARN as
   `-var='existing_github_oidc_provider_arn=arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com'`
   to `plan` (and to `destroy` later). Otherwise Terraform will create it. If
   the public repository already exists, import it after `init` and before
   applying rather than trying to create a duplicate:

   ```bash
   cd infra/ecr-public
   terraform init
   terraform import aws_ecrpublic_repository.app cloud-infrastructure-platform
   ```

   Skip the import if the repository does not exist. Add the same `-var`
   argument to `terraform import` when reusing an existing OIDC provider.

3. From `infra/ecr-public`, review and apply the plan:

   ```bash
   terraform fmt -check
   terraform validate
   terraform plan -out=ecr.tfplan
   terraform apply ecr.tfplan
   terraform output repository_uri
   terraform output github_ecr_publish_role_arn
   ```

   Terraform state and saved plans are ignored by Git. Keep the state file
   safe: Terraform needs it to update or remove these resources later. Do not
   create the same repository or IAM role manually after Terraform manages it.

4. Add these GitHub repository **Actions variables** under **Settings →
   Secrets and variables → Actions** when the infrastructure is ready:

   | Variable | Value |
   |---|---|
   | `AWS_ECR_PUBLISH_ROLE_ARN` | `github_ecr_publish_role_arn` Terraform output |
   | `ENABLE_ECR_PUBLISH` | `true` only when ready to publish on a PR merge |

   To publish on **this PR's merge**, apply Terraform and set both variables
   before merging. The CD workflow will wait for the new `main` CI run and
   will not publish if its SonarQube quality gate fails. You can also leave
   publishing disabled, merge first, inspect `main` CI, and enable publishing
   for a later PR merge.

The [publish workflow](../.github/workflows/publish-ecr.yml) is triggered by
pushes to `main`. Its unprivileged `verify` job checks that the commit came
from a merged pull request into `main` and waits for that exact SHA's CI run
to succeed. Direct pushes do not publish. The `publish` job then assumes the
narrow OIDC role and publishes `api-<commit SHA>` and `migrate-<commit SHA>`
tags to ECR Public. It skips an already-existing tag and records each image
digest in the workflow summary. No long-lived AWS keys are stored in GitHub.

To stop publishing, set `ENABLE_ECR_PUBLISH` to `false` or remove it. For a
full teardown, delete the repository's images in the ECR Public console and
run `terraform destroy` from `infra/ecr-public` with the same OIDC-provider
variable used for `apply`. Review the destroy plan before confirming it. A
shared OIDC provider supplied through the variable is not managed or deleted
by this configuration. ECR Public is only an image registry; application
deployment, Kubernetes, and GitOps come later.
