# Security checks and image publishing

The order is: test the Go code, build both runtime images, scan them, analyze
source quality, then publish images from a commit whose CI passed. The ECR
workflow is disabled until the AWS prerequisites are in place.

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
waits for the SonarQube quality gate. Until then, CI prints a configuration
warning and skips the analysis; if ECR publishing is enabled, a missing token
fails CI.

SonarSource currently documents full Go support through 1.25, while this
project uses Go 1.27.1. Check the first analysis for parser warnings or
missing files before treating its findings as complete.

Codecov is separate: its existing CI upload needs `CODECOV_TOKEN` in GitHub
Actions secrets.

## 3. Prepare ECR in `us-east-1`

Do this only after CI, Codecov, and SonarQube Cloud pass on `main`. The local
AWS CLI currently has no credentials configured. First authenticate with an
AWS identity that can manage ECR and IAM, then verify the account:

```bash
aws sts get-caller-identity
aws ecr create-repository --repository-name cloud-infrastructure-platform \
  --image-tag-mutability IMMUTABLE --region us-east-1
```

If the repository already exists, inspect it instead of creating a duplicate:

```bash
aws ecr describe-repositories --repository-names cloud-infrastructure-platform \
  --region us-east-1
```

Create a GitHub OIDC IAM role whose trust policy permits only
`repo:shivamshashank/cloud-infrastructure-platform:ref:refs/heads/main` and
audience `sts.amazonaws.com`. Its permissions should allow
`ecr:GetAuthorizationToken` on `*`, and only ECR layer upload, `PutImage`,
`DescribeImages`, and `DescribeRepositories` on this repository's ARN.
Follow the [AWS GitHub OIDC role guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-idp_oidc.html)
for the trust policy and provider setup. Do not store long-lived AWS keys in
GitHub Actions.

Then add repository Actions variables:

| Variable | Value |
|---|---|
| `AWS_ECR_PUBLISH_ROLE_ARN` | ARN of the narrow OIDC role |
| `ENABLE_ECR_PUBLISH` | `true` when ready; leave unset to keep publishing off |

The [publish workflow](../.github/workflows/publish-ecr.yml) runs only after a
successful CI run on `main` and only while that switch is `true`. It checks
out the exact commit that passed, builds the API and migration images, and
pushes immutable `api-<commit SHA>` and `migrate-<commit SHA>` tags. It writes
their image digests to the workflow summary. This is image publishing, not an
application deployment; Kubernetes and GitOps come later. ECR storage can
incur charges, so keep the repository small and remove unused images after
the exercise.
