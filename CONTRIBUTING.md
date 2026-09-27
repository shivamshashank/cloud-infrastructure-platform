# Contributing

Thanks for helping build Cloud Infrastructure Platform. This is a learning
project, so a clear explanation of *why* a change works is as useful as the
change itself.

## Before you start

1. Check the [README](README.md) for the current implementation and roadmap.
2. Search existing issues before opening a new one.
3. For a larger change, open an issue describing the problem, approach, and
   how you will verify it.
4. Report vulnerabilities privately under [SECURITY.md](SECURITY.md).

## Local development

Use the [quick start](README.md#-quick-start) to run PostgreSQL and the API.
Install the Git hooks once after cloning (requires
[`pre-commit`](https://pre-commit.com/#install)):

```bash
pre-commit install
pre-commit run --all-files
```

The hooks format Go files, run `go vet` and `go test` when Go files or the
module change, and check changed text files for trailing whitespace. `gofmt` may rewrite a file; review
and stage that change before committing. You can also run the checks directly:

```bash
gofmt -w cmd/api/*.go cmd/migrate/*.go migrations/*.go
go test ./...
go vet ./...
```

Add a focused test for new behaviour or a bug fix. If a test needs PostgreSQL,
describe how to run it and keep test data synthetic.

## Pull requests

- Keep each pull request focused and explain the behaviour it changes.
- Describe how you tested it; include commands and results.
- Update the README or other relevant documentation when behaviour changes.
- Mark an AWS, Kubernetes, CI, or observability feature as complete only after
  it has been exercised and evidence is available.
- Do not commit credentials, Terraform state, kubeconfig files, private keys,
  real user data, or unredacted screenshots.
- For infrastructure changes, include a cost estimate, a teardown path, and
  any expected service interruption.

Use the [pull request template](.github/PULL_REQUEST_TEMPLATE.md). Small,
reviewable commits are preferred.

## Community

Follow the [Code of Conduct](CODE_OF_CONDUCT.md). Public bug and feature
requests can use the forms in [.github/ISSUE_TEMPLATE](.github/ISSUE_TEMPLATE).
