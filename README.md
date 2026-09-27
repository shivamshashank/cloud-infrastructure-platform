<div align="center">

# ☁️ Cloud Infrastructure Platform

### Build, deploy, observe, and recover a small Go service on Kubernetes and AWS

A hands-on SRE and platform engineering project built around one PostgreSQL-backed
Task API. Each stage adds an operational capability and a way to prove it works.

**Current stage:** Go API and local PostgreSQL · **AWS lab Region:** `us-east-1`

<br />

[![CI](https://img.shields.io/github/actions/workflow/status/shivamshashank/cloud-infrastructure-platform/ci.yml?branch=main&label=CI&logo=githubactions&style=flat-square)](https://github.com/shivamshashank/cloud-infrastructure-platform/actions/workflows/ci.yml)
[![Release: planned](https://img.shields.io/badge/Release-planned-6c757d?style=flat-square&logo=githubactions&logoColor=white)](#-current-status-and-learning-path)
[![Codecov](https://img.shields.io/codecov/c/github/shivamshashank/cloud-infrastructure-platform?logo=codecov&style=flat-square)](https://app.codecov.io/gh/shivamshashank/cloud-infrastructure-platform)
[![Version: unreleased](https://img.shields.io/badge/Release-unreleased-6c757d?style=flat-square)](#-current-status-and-learning-path)
[![GitHub stars](https://img.shields.io/github/stars/shivamshashank/cloud-infrastructure-platform?style=flat-square)](https://github.com/shivamshashank/cloud-infrastructure-platform/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/shivamshashank/cloud-infrastructure-platform?style=flat-square)](https://github.com/shivamshashank/cloud-infrastructure-platform/forks)
[![License: MIT](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

<br />

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker Compose](https://img.shields.io/badge/Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-232F3E?style=for-the-badge&logo=amazonwebservices&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-844FBA?style=for-the-badge&logo=terraform&logoColor=white)
![Ansible](https://img.shields.io/badge/Ansible-EE0000?style=for-the-badge&logo=ansible&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Helm](https://img.shields.io/badge/Helm-0F1689?style=for-the-badge&logo=helm&logoColor=white)
![Argo CD](https://img.shields.io/badge/Argo%20CD-EF7B4D?style=for-the-badge&logo=argo&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white)
![Amazon ECR](https://img.shields.io/badge/Amazon%20ECR-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white)
![Trivy](https://img.shields.io/badge/Trivy-1904DA?style=for-the-badge&logo=trivy&logoColor=white)
![SonarQube](https://img.shields.io/badge/SonarQube-4E9BCD?style=for-the-badge&logo=sonarqube&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![PagerDuty](https://img.shields.io/badge/PagerDuty-06AC38?style=for-the-badge&logo=pagerduty&logoColor=white)

<sub>Stack badges show the project scope. The table below tracks what is implemented.</sub>

</div>

---

## 📌 Overview

The application is intentionally small: create, list, read, update, and delete
tasks. The project uses that service to learn the operational path from a local
process to a repeatable cloud deployment:

1. Build and test the API.
2. Package and scan a container image.
3. Provision a temporary AWS lab with Terraform and configure its host with
   Ansible.
4. Deploy through Helm and Argo CD.
5. Measure service behaviour, alert on failure, and practice recovery.

This is a **learning lab**, not a claim of production availability. The AWS
environment will be created only for exercises and destroyed afterward to
preserve the available credits.

| | Current state |
|---|---|
| **Application** | Go HTTP Task API with PostgreSQL persistence |
| **Working locally** | CRUD, input validation, bounded listing, health, readiness, request IDs, structured logs, and Prometheus metrics |
| **Tests** | Handler tests and a PostgreSQL integration test using a dedicated `tasks_test` database; both passed locally |
| **Cloud deployment** | Planned for AWS `us-east-1`; no Terraform or Kubernetes deployment is in this repository yet |
| **Operational evidence** | CI runs, screenshots, SLO measurements, chaos results, and restore logs will be added when performed |

---

## 🧭 Current status and learning path

The badges above summarize the target stack. The table identifies implemented
work; an unchecked item is a planned exercise, not a completed capability.

| Stage | Deliverable | Status |
|---|---|---|
| 1. Application | Go CRUD API, SQL migration, local PostgreSQL | ✅ CRUD, persistence, and database outage behaviour verified locally |
| 2. Quality | Integration tests, formatting, linting, pre-commit | 🟡 CI workflow configured; first GitHub run and Codecov upload pending |
| 3. Container | Multi-stage Dockerfile, non-root image, Trivy scan | 🟡 Both images pass local HIGH/CRITICAL Trivy gate; GitHub CI run pending |
| 4. AWS foundation | IAM, VPC, EC2, S3, ECR, CloudWatch; Terraform and Ansible | 🟡 ECR publish workflow prepared but disabled; AWS setup pending |
| 5. Kubernetes | K3s first; Services, probes, Secrets, storage, RBAC | ⬜ Planned |
| 6. Delivery | GitHub Actions, Helm, Argo CD and immutable image promotion | ⬜ Planned |
| 7. Operations | Prometheus, Grafana, Alertmanager, SLI/SLO and PagerDuty | ⬜ Planned |
| 8. Resilience | k6, smoke tests, chaos exercise, backup/restore and runbooks | ⬜ Planned |
| 9. Extensions | SonarQube, Jenkins comparison, small Go operator, temporary EKS exercise | 🟡 SonarQube Cloud project imported; token and first CI analysis pending |

The older [project specification](01-project-specification.md) and
[ten-day plan](02-ten-day-implementation-plan.md) were drafted for a VPS and
GHCR. **Their hosting and registry assumptions are superseded here by AWS and
ECR.** Use their acceptance criteria as ideas, not as a record of completed
work.

---

## ⚡ Quick start

You need Go 1.27 and a running Docker-compatible daemon. The example password
is only for a database bound to `127.0.0.1`; choose a separate secret for any
remote environment.

```bash
git clone https://github.com/shivamshashank/cloud-infrastructure-platform.git
cd cloud-infrastructure-platform
export POSTGRES_PASSWORD=localdev
docker compose up -d --wait db
export DATABASE_URL="postgres://tasks:localdev@127.0.0.1:5432/tasks?sslmode=disable"
go run ./cmd/migrate
go run ./cmd/api
```

In a second terminal:

```bash
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
curl -i -X POST http://127.0.0.1:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Learn Go","description":"Build the first API"}'
curl -i 'http://127.0.0.1:8080/tasks?limit=20&offset=0'
curl -i -X PATCH http://127.0.0.1:8080/tasks/1 \
  -H 'Content-Type: application/json' -d '{"status":"done"}'
curl -i -X DELETE http://127.0.0.1:8080/tasks/1
curl -i http://127.0.0.1:8080/metrics
python3 scripts/smoke_test.py
```

Stop the database with `docker compose down`. Its named volume keeps data.
`docker compose down -v` removes that volume and its data.

### 🐳 Run the full stack in containers

The local Compose stack starts PostgreSQL, runs the migration once, and then
starts the API. Use the same local-only example password as above:

```bash
export POSTGRES_PASSWORD=localdev
docker compose up -d --build --wait api
python3 scripts/smoke_test.py
docker compose ps -a
docker compose down
```

`docker compose down` keeps PostgreSQL data in its named volume. See the
[container walkthrough](docs/containers.md) to learn what the Dockerfile
stages do and how to inspect failures. If you set a different password, URL
encode any reserved URL characters because Compose inserts it into a database
connection URL.

### 🔌 API contract

| Endpoint | Behaviour |
|---|---|
| `GET /healthz` | Process liveness; does not query PostgreSQL |
| `GET /readyz` | Database readiness; returns 503 when unavailable |
| `GET /metrics` | Prometheus counters, latency histogram, and in-flight gauge; keep private |
| `GET /tasks?limit=20&offset=0` | List tasks; limit is 1–100 |
| `POST /tasks` | Create a task; returns 201 and a `Location` header |
| `GET /tasks/{id}` | Read one task or return 404 |
| `PATCH /tasks/{id}` | Update title, description, or status |
| `DELETE /tasks/{id}` | Delete a task; a repeated delete returns 404 |

Status values are `todo`, `in_progress`, and `done`. JSON errors use an
`error` field. Each response includes `X-Request-ID`; JSON logs record that
ID, route, status, and duration without request bodies. The migration is
applied explicitly; starting the API does not silently change the schema.
Set `DB_MAX_CONNS` (default 5, range 1–100) and `DB_MIN_CONNS` (default 0,
at most the maximum) to tune the PostgreSQL pool. Keep `/metrics` reachable
only from a trusted network or Prometheus; it has no authentication.

---

## 🏗️ Architecture

### Running locally today

```mermaid
flowchart LR
    Client["curl / HTTP client"] --> API["Go Task API container<br/>localhost:8080"]
    API --> DB[("PostgreSQL container")]
    Migration["Migration container<br/>runs before API"] --> DB
```

### Target AWS lab — planned

```mermaid
flowchart LR
    Dev["Developer"] --> GitHub["GitHub source"]
    GitHub --> CI["GitHub Actions<br/>tests · SonarQube · Trivy"]
    CI --> ECR["Amazon ECR<br/>image digest"]
    GitHub --> GitOps["GitOps configuration"]
    GitOps --> Argo["Argo CD"]
    Terraform["Terraform"] --> AWS["AWS: IAM · VPC · EC2 · S3"]
    Ansible["Ansible"] --> Host["EC2 Linux host"]
    AWS --> Host
    Host --> K3s["K3s / containerd"]
    Argo --> K3s
    ECR --> K3s
    K3s --> App["Go API + PostgreSQL"]
    K3s --> Obs["Prometheus · Grafana<br/>Alertmanager"]
    Obs --> PD["PagerDuty"]
    App --> Backup["S3 backup exercise"]
    Host --> CW["CloudWatch"]
```

**Ownership rule:** Terraform provisions AWS resources; Ansible configures the
Linux host; Helm packages Kubernetes workloads; Argo CD reconciles releases
from Git. Each resource should have one owner.

### ☁️ Why these six AWS services?

| Service | Project use |
|---|---|
| IAM | Scoped access for people, CI, and the EC2 instance |
| VPC | Subnets, routing, and security groups |
| EC2 | Temporary K3s lab server |
| S3 | Protected Terraform state and off-server database backup |
| ECR | Container image registry for AWS deployments |
| CloudWatch | Basic AWS host metrics and carefully retained logs |

EKS is a later, short-lived comparison with K3s. No always-on managed cluster
is required for the core project.

---

## 🔁 Delivery and reliability exercises

| Exercise | Evidence to save when completed |
|---|---|
| CI quality gate | One failing check and a passing rerun |
| Image promotion | Commit SHA, image digest, and deployed revision |
| GitOps rollback | Bad release, diagnosis, Git revert, restored service |
| Observability | Request rate, error ratio, p95/p99 latency, resource dashboard |
| Alerting | Alert fires, reaches PagerDuty, resolves, and links to a runbook |
| Chaos | Controlled pod or dependency failure with detection/recovery times |
| Backup and restore | Synthetic records restored into a separate database |
| Cost teardown | Terraform destroy output and billable-resource inventory |

Numbers and screenshots belong here **after** the corresponding exercise is
run. No latency, availability, or recovery result is claimed yet.

---

## 📚 Documentation

| | File | What it contains |
|---|---|---|
| 🗺️ | [Project specification](01-project-specification.md) | Original scope and acceptance criteria; VPS/GHCR assumptions are outdated |
| 🗓️ | [Ten-day implementation plan](02-ten-day-implementation-plan.md) | Original staged learning plan; hosting assumptions are outdated |
| 🗃️ | [Initial migration](migrations/001_create_tasks.sql) | PostgreSQL schema |
| 🔄 | [Migration procedure](docs/migrations.md) | Local and planned Kubernetes release sequence |
| 🧾 | [Backend local evidence](docs/evidence/backend-local.md) | Tests, persistence, outage and smoke-test results |
| 🐳 | [Container walkthrough](docs/containers.md) | Dockerfile stages, Compose startup, smoke test and cleanup |
| 🔐 | [Security and publishing](docs/security-and-publishing.md) | Trivy, SonarQube Cloud and ECR setup |
| 🧾 | [Image scan evidence](docs/evidence/container-security.md) | Initial pgx findings and clean rescans |
| 🐳 | [Compose file](compose.yaml) | Local PostgreSQL, migration and API services |
| ✅ | [CI workflow](.github/workflows/ci.yml) | Push checks, PostgreSQL tests, and Codecov upload |
| 📦 | [ECR workflow](.github/workflows/publish-ecr.yml) | SHA-tagged image publishing after a passing CI run; disabled by default |
| 🤝 | [Contributing](CONTRIBUTING.md) | Development and pull request workflow |
| 🔐 | [Security policy](SECURITY.md) | Private vulnerability reporting |
| 📜 | [Code of conduct](CODE_OF_CONDUCT.md) | Community expectations |

Planned documentation includes architecture decisions, tested versions,
runbooks, incident notes, cost records, and restore evidence. Each will be
linked once it exists.

---

## 📂 Repository structure

```text
cmd/api/                         Go API, metrics, handler and integration tests
cmd/migrate/                     Migration command
migrations/                      Versioned SQL migration
scripts/smoke_test.py            Post-deployment API smoke test
docs/migrations.md               Migration release procedure
Dockerfile                       API and migration images
compose.yaml                     Local PostgreSQL, migration and API
docs/containers.md              Container walkthrough
docs/security-and-publishing.md Trivy, SonarQube Cloud and ECR setup
sonar-project.properties        Go analysis and coverage configuration
01-project-specification.md      Earlier project draft
02-ten-day-implementation-plan.md Earlier implementation draft
.github/ISSUE_TEMPLATE/          Bug and feature request forms
.github/PULL_REQUEST_TEMPLATE.md Pull request checklist
CONTRIBUTING.md                  Contribution guide
CODE_OF_CONDUCT.md               Community rules
SECURITY.md                      Vulnerability reporting
LICENSE                          MIT license
```

Future directories for Terraform, Ansible, Helm, GitOps, monitoring, and
recovery tests will be listed when their contents are committed.

---

## 🧪 Testing and reproduction

```bash
go test ./...
go vet ./...
gofmt -d cmd/api/*.go cmd/migrate/*.go migrations/*.go
pre-commit run --all-files
```

Run `pre-commit install` once after cloning to check each commit. The hooks
use the local Go toolchain and require no downloaded hook repositories.

Every push runs the same hooks in
[GitHub Actions](.github/workflows/ci.yml), followed by race-enabled tests
against a dedicated PostgreSQL service, two Trivy image scans, an optional
SonarQube Cloud analysis, and a Codecov coverage upload. To
enable the upload, add the repository secret `CODECOV_TOKEN` from your Codecov
repository settings under **GitHub → Settings → Secrets and variables →
Actions**. The upload step fails if Codecov rejects the report; a passing
badge should appear only after the workflow succeeds.

For the PostgreSQL integration test, create a **dedicated** test database and
run the test with its URL. The test refuses any database whose name is not
`tasks_test` and clears only that test database:

```bash
docker compose exec -T db psql -U tasks -d tasks -c 'CREATE DATABASE tasks_test'
export TEST_DATABASE_URL="postgres://tasks:localdev@127.0.0.1:5432/tasks_test?sslmode=disable"
go test ./...
```

To reproduce the current API behaviour, run the [quick start](#-quick-start)
and issue the example requests. Run `python3 scripts/smoke_test.py` against a
running API; it creates and deletes one synthetic task. Load test results and
deployment smoke-test evidence are still planned.

### 🛡️ Controls we intend to verify

| Control | Expected check |
|---|---|
| Database unavailable | `/readyz` fails within a timeout while `/healthz` remains live |
| Invalid input | Blank title, unknown status, malformed JSON, and bad IDs are rejected |
| Image provenance | Deployment references a scanned immutable digest |
| Reconciliation | A manual Kubernetes drift is corrected from Git |
| Data recovery | Backup restores records into a fresh database |
| Cloud cost | Temporary AWS resources are inventoried and destroyed |

---

## 🚧 Known limitations

| Limitation | Consequence |
|---|---|
| No authentication or rate limiting | Keep the API private; do not expose it directly to the internet |
| No completed dashboard or alert rules | `/metrics` is available, but Prometheus/Grafana deployment is still planned |
| Single local PostgreSQL instance | No high availability; the Docker volume is not an off-server backup |
| AWS/Kubernetes configuration absent | The target architecture has not been deployed or validated |
| Older planning documents describe a VPS and GHCR | Follow the AWS/ECR direction in this README for current project intent |

---

## 🤝 Contributing

Contributions and learning-focused feedback are welcome. Read
[CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Use the
[bug report](.github/ISSUE_TEMPLATE/bug_report.yml) or
[feature request](.github/ISSUE_TEMPLATE/feature_request.yml) forms for public
issues. Report vulnerabilities through the private process in
[SECURITY.md](SECURITY.md). Community interactions follow
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

---

## 📄 License and citation

This project is released under the [MIT License](LICENSE).

If you cite the software in a portfolio or technical write-up, link to this
repository and identify the commit you used. There is no archived dataset,
published experiment, or DOI for this project.

## 👤 Author

**Shivam Shashank** — MSc dissertation, University of Birmingham. Supervisor: Dr Vincent Rahli.

- 🌐 Portfolio: [shivam-shashank.me](https://www.shivam-shashank.me/)
- 💼 LinkedIn: [shivam-shashank-2b5766217](https://www.linkedin.com/in/shivam-shashank-2b5766217/)
- 📧 Email: [shivamkumar872000@gmail.com](mailto:shivamkumar872000@gmail.com)
- 🐙 GitHub: [shivamshashank](https://github.com/shivamshashank)

---

<div align="center">

### ⭐ Follow the build, run the exercises, and share what you find

</div>
