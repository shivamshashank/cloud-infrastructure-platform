# Cloud Infrastructure Platform: Ten-Day Implementation Plan

> Historical schedule: use the [current README](README.md) for the active AWS
> roadmap and the [Consul/Vault learning flow](docs/service-discovery-and-secrets.md)
> for the added discovery and secrets exercises. VPS/GHCR steps below are
> retained as planning history.

## Target

Complete a small Go application platform on one remote 16 GB server. Learn by implementing, breaking, diagnosing and documenting each component. Follow [the project specification](01-project-specification.md) for architecture and boundaries.

Budget: **10 days × 8 hours = 80 hours**. This is ambitious for a learner. Each day has an acceptance gate: carry unfinished essentials forward rather than claiming completion. Simplify optional features if time runs short; all named technologies remain represented by one small working exercise.

All work runs remotely. GitHub replaces GitLab entirely. GitHub Actions is primary CI; Jenkins is a separate manually triggered CI exercise. Argo CD owns application deployment after day 6.

## Prerequisites and day-one readiness

- Choose a VPS provider: 16 GB RAM, at least 4 vCPUs, suitable SSD capacity and root access.
- Have a GitHub account, SSH key, provider account and billing access. Set a spending limit/alert where supported; alerts are not hard caps.
- Select Linux and confirm your server CPU architecture matches the image build target.
- Decide how Terraform will first run without the target VPS: a manually dispatched hosted workflow with protected credentials, or manually provision and import the server later.
- Keep a recovery route outside the VPS, including provider console access and a protected state backup.
- Reuse an existing subdomain if available. A new domain is optional; private access through SSH tunnels is sufficient.
- Do not enter credentials into repository files or record them in demonstration videos.

These account and billing prerequisites must be ready before the clock starts. Day 1 includes server bootstrap, but account verification delays may extend the schedule.

## Daily working method

At the start, write down what you expect to happen. Read enough official documentation to understand the next change, implement it, then explain it without copying generated text. Commit small changes. End each day with acceptance evidence, one failure observation and a short learning note.

Suggested evidence path: `docs/evidence/day-XX.md`. Record commands/results, relevant screenshots, actual versions and unresolved issues. Never include secrets. Use pinned versions verified at implementation time.

## Day 1 — Remote setup, Go API and PostgreSQL

**Goal:** A tested Go API running on the server and connected to PostgreSQL.

### Eight-hour allocation

| Hours | Work |
|---|---|
| 1.5 | Bootstrap VPS access, install basic development tools and PostgreSQL, create GitHub repository |
| 1 | Learn request handling, contexts, database connections and project structure |
| 3 | Implement migration, CRUD handlers, validation, errors and health endpoints |
| 1.5 | Add meaningful tests and bounded database timeouts |
| 1 | Verify behaviour, document API and commit |

### Tasks

- Create a non-root administration user and verify SSH-key access before changing access rules. Keep provider console access available.
- Record manual bootstrap steps so Terraform/cloud-init can reproduce them on day 3.
- Create `cmd/api`, internal packages and an initial migration.
- Add task CRUD with bounded pagination and consistent HTTP/JSON errors.
- Load database credentials from environment/configuration outside Git.
- Add structured logs and request IDs; avoid logging credentials or task descriptions unnecessarily.
- Implement `/healthz` and `/readyz` with different semantics.
- Add graceful shutdown and connection-pool configuration.
- Test create/read/update/delete, malformed requests, unknown IDs and database failure behaviour. Use actual PostgreSQL for integration tests.

### Acceptance gate

Demonstrate CRUD with `curl`, restart the API and verify persisted data. Stop database access temporarily: readiness fails within a timeout, while liveness stays healthy. Tests pass.

**Interview prompts:** What happens between request and SQL query? Why use a context timeout? Why separate readiness from liveness?

**Outputs:** API code, migration, tests, API README and day-one evidence.

## Day 2 — Docker and K3s

**Goal:** Run the application and PostgreSQL inside Kubernetes on the same server.

| Hours | Work |
|---|---|
| 1 | Learn image construction and Kubernetes workload basics |
| 2 | Build multi-stage images and install K3s |
| 2.5 | Write initial workload, Service, Secret, configuration and PVC manifests |
| 1.5 | Run failure and persistence exercises |
| 1 | Document networking, storage and commands |

### Tasks

- Build the Go image with a minimal non-root runtime; verify port binding and termination.
- Install a pinned K3s release and configure kubeconfig permissions.
- Use namespace `task-app`; deploy PostgreSQL with persistent storage and the API with a ClusterIP Service.
- Create Secrets outside Git; commit only instructions/examples without values.
- Add requests, limits, readiness and liveness checks.
- Verify the image is available to K3s: Docker's local image cache is not automatically shared with its runtime. Import a development image explicitly or publish/pull it.
- Migrate synthetic data if needed, then stop the day-one standalone PostgreSQL instance to avoid duplicate resource use.

### Failure exercises and gate

Use a bad image tag and explain image-pull events. Break the database hostname and diagnose using logs, Services and endpoints. Delete/recreate the PostgreSQL pod and verify persisted data. Restore all settings and confirm readiness.

**Interview prompts:** Container versus pod? Service versus pod IP? PVC versus backup? What survives node failure?

**Outputs:** Dockerfile, initial Kubernetes manifests, operational notes and evidence.

## Day 3 — Terraform and reproducible server provisioning

**Goal:** Describe the VPS and firewall as code without relying on undocumented manual setup.

| Hours | Work |
|---|---|
| 1.5 | Study selected provider, state, import and cloud-init |
| 2.5 | Implement VPS/firewall configuration and bootstrap scripts |
| 1.5 | Import existing resources or validate planned provisioning |
| 1.5 | Verify state recovery, bootstrap and safe teardown procedures |
| 1 | Document costs, prerequisites and architectural decisions |

### Tasks

- Pin Terraform/provider versions and commit the lock file.
- Configure variables, outputs, SSH keys, VPS image/size and firewall rules.
- Translate host setup into cloud-init plus repeatable scripts. Never embed long-lived secrets in cloud-init, which may persist in provider metadata/state.
- If the server was provisioned manually, import it and relevant resources. Inspect plans carefully: do not blindly apply a replacement of your only server.
- Set up protected state storage or encrypted off-server backup; document locking/concurrency limits of the chosen approach.
- Keep provider credentials out of Git and public workflow artifacts.
- Ensure the plan does not unintentionally expose admin ports or replace imported infrastructure.
- Write a teardown checklist covering server, volumes, snapshots, addresses and other billable resources.

### Acceptance gate

`terraform fmt -check` and validation pass. After reconciliation, the plan reports no unexpected changes. Retrieve a state backup through the documented process. If a fresh-server rebuild cannot be performed within budget, explicitly mark it untested and complete it on day 10.

**Interview prompts:** What does state contain? What is drift? Why can stopping a server leave charges? How would you recreate a failed server?

**Outputs:** Provider-specific Terraform, cloud-init, recovery instructions and cost inventory. This day covers VPS infrastructure, not AWS/EKS.

## Day 4 — Helm, GitHub Actions and GHCR

**Goal:** Produce a tested, versioned image and deploy it with a reusable chart.

| Hours | Work |
|---|---|
| 1 | Learn Helm values and GitHub Actions permissions |
| 2 | Package existing manifests into a chart |
| 2.5 | Add primary CI and GHCR publishing |
| 1.5 | Verify upgrade, compatible rollback and fresh image pull |
| 1 | Document release workflow |

### Tasks

- Add chart templates, values, resource settings, health checks and existing-Secret references.
- Run lint/render checks; avoid conflicting resource ownership when migrating day-two manifests to Helm.
- Configure pull-request jobs for formatting, vet, tests and PostgreSQL integration tests.
- On trusted main commits, build and publish the API image to GHCR with the Git SHA. Record the image digest.
- Use narrowly scoped workflow permissions; PR jobs do not receive publishing secrets.
- Store shared build/test commands in scripts or Make targets for reuse by Jenkins.
- Test image pulls without depending on a pre-existing cache.
- Demonstrate a Helm upgrade and rollback with a backward-compatible schema.

### Acceptance gate

A deliberately failing test makes CI fail. A fixed main commit publishes an image. Deploy that image through Helm and verify the API. Save image digest and test results.

**Interview prompts:** Why an immutable digest? What does Helm render? Why is schema rollback different from image rollback?

**Outputs:** Application chart, workflow files, shared scripts and release evidence.

## Day 5 — SonarQube and image security

**Goal:** Learn code analysis and scanning through a demonstrated pass/fail policy.

| Hours | Work |
|---|---|
| 1 | Understand SonarQube, coverage import and scanner results |
| 2 | Deploy Community Build and its supported database configuration |
| 2 | Configure remote-server analysis and Trivy in CI |
| 2 | Demonstrate failed checks, fixes and successful reruns |
| 1 | Document access and gate policy |

### Tasks

- Check the selected SonarQube release's requirements, Go support and scanner compatibility before installation.
- Run SonarQube privately on the VPS, with separate database/user from the application's database.
- Configure an analysis token securely and import the Go coverage report.
- Run the scanner from the server; save a reusable script for Jenkins on day 9.
- Generate a deliberate code-quality/coverage failure and fix it.
- Add Trivy to the GitHub image build workflow before publication. Define treatment of severe findings, unavailable fixes and temporary exceptions.
- Set an analysis timeout. Do not make private SonarQube access from hosted GitHub runners an implicit dependency.

### Acceptance gate

Show one failed and one passing analysis; show the vulnerability scan result and explain its policy. Admin access is restricted, and secrets are absent from logs/Git.

**Interview prompts:** What can static analysis miss? What does coverage fail to prove? Why might an image scan require an exception?

**Outputs:** Scanner configuration, scan/gate policy, CI scan step and evidence. The automated SonarQube gate is completed in Jenkins on day 9.

## Day 6 — Argo CD and GitOps

**Goal:** Deploy and recover application versions through Git.

| Hours | Work |
|---|---|
| 1 | Learn reconciliation and ownership |
| 2 | Install Argo CD and create the companion GitOps repository |
| 2 | Configure chart source, image digest and application sync |
| 2 | Test promotion, drift and Git-based recovery |
| 1 | Document the deployment procedure |

### Tasks

- Install Argo CD with a pinned upstream chart and minimal non-HA configuration.
- Restrict its UI/API and store repository credentials appropriately.
- Define an Application referencing a tested chart revision and lab values in Git. Use a supported multi-source setup if separating chart and values repositories.
- Transfer application ownership from manual Helm management deliberately; avoid parallel management and preserve database PVC/data.
- Create a promotion PR updating the image digest; merge and sync.
- Demonstrate drift with a harmless replica change, then restore declared state.
- Introduce a bad image/configuration in the lab, inspect failure and revert the Git commit.
- Configure pruning cautiously and protect persistent data from accidental deletion.

### Acceptance gate

Git describes the running application version. A reviewed change deploys successfully, and a revert returns to the previous compatible release. Capture sync and health evidence.

**Interview prompts:** Why separate CI from CD? What happens after a manual cluster edit? Does Argo CD build images?

**Outputs:** GitOps repository, Argo configuration, release/recovery runbook and evidence.

## Day 7 — Observability and load testing

**Goal:** Detect and investigate application problems using metrics and logs.

| Hours | Work |
|---|---|
| 1 | Learn metrics, labels, histograms and alert evaluation |
| 2 | Install a minimal Prometheus/Grafana/Alertmanager stack through Helm |
| 2 | Instrument the API and add dashboard/alert definitions |
| 2 | Run k6 traffic and a controlled failure |
| 1 | Document the first incident |

### Tasks

- Configure small persistent storage, retention and requests/limits for monitoring.
- Disable unsupported K3s scrape targets; explain any remaining alerts.
- Expose request counts and duration histograms using bounded labels.
- Create a ServiceMonitor with selectors matching the Prometheus configuration.
- Commit dashboard JSON and alert rules.
- Add traffic, error ratio, p95 latency, memory/CPU and restart panels.
- Configure Alertmanager to a test webhook receiver and verify a received notification.
- Run a short k6 baseline, then introduce a database connection failure or controlled API error.
- Use structured logs to complement metrics. Do not add Loki/Tempo this week.

### Acceptance gate

Metrics targets are healthy. Dashboard data changes under traffic. An alert fires, is received, and resolves after recovery. Record load, duration and environment alongside measurements.

**Interview prompts:** Why not label metrics with task IDs? What does p95 mean? When should an alert page someone?

**Outputs:** Instrumentation, Helm values, dashboard, alert rule, k6 script and incident report 1.

## Day 8 — Small Go Kubernetes operator

**Goal:** Reconcile one fixed worker Deployment from a custom resource.

| Hours | Work |
|---|---|
| 1.5 | Learn CRDs, reconciliation and owner references |
| 1 | Scaffold a compatible Kubebuilder project |
| 3 | Implement the worker and controller |
| 1.5 | Test reconciliation, ownership and RBAC |
| 1 | Document behaviour and limitations |

### Tasks

- Create the `TaskWorker` CRD with a validated replica field and status.
- Write the small worker that polls API readiness with timeout/backoff and structured logs.
- Fix the worker image and API target through operator configuration; accept no arbitrary manifests.
- Implement create/update logic with owner references and watches for owned Deployments.
- Avoid unnecessary update loops and return actionable errors for failed operations.
- Restrict permissions to required resources/namespaces.
- Test repeated reconcile, replica changes and missing Deployment behaviour using suitable controller tests plus real-cluster verification.
- Publish worker/operator images through existing CI and deploy via GitOps.

### Acceptance gate

Create a TaskWorker, change replicas, delete its Deployment and observe recreation. Delete the TaskWorker and verify garbage collection. Explain each transition without reading generated code.

**Interview prompts:** Why must reconciliation be idempotent? Why is a finalizer unnecessary here? What is an owner reference?

**Outputs:** Operator/worker code, generated manifests, tests and demonstration evidence.

## Day 9 — Jenkins alternative CI

**Goal:** Reproduce the main build workflow and enforce the private SonarQube quality gate.

| Hours | Work |
|---|---|
| 1 | Learn Jenkins controller/executor and credentials concepts |
| 2 | Install/configure Jenkins with one executor and protected access |
| 2.5 | Implement Jenkinsfile calling shared scripts |
| 1.5 | Demonstrate failed tests/gate and successful build |
| 1 | Compare CI systems and record findings |

### Tasks

- Use a supported Jenkins release and minimal plugins. Record versions.
- Configure credentials for repository access if needed, GHCR publishing and SonarQube analysis.
- Run only trusted code; avoid unrestricted privileged execution.
- Add checkout, test, coverage, SonarQube analysis, gate wait/check, image build, Trivy scan and optional publication stages.
- If the quality-gate integration uses a webhook, configure its URL/authentication and prove delivery. Otherwise use bounded polling via the supported API.
- Publish `jenkins-<sha>` tags, without automatically changing GitOps configuration.
- Limit concurrency to one and clean workspaces/caches responsibly.
- Keep application rollout owned by Argo CD.

### Acceptance gate

A manual Jenkins run stops on a failed test or quality gate. After fixing the issue, it completes and produces a traceable image. GitHub Actions remains the main pipeline.

**Interview prompts:** When would you choose Jenkins? Where do builds run? How are secrets scoped? Why avoid two production promotion pipelines?

**Outputs:** Jenkinsfile, setup/credentials instructions without secrets and CI comparison notes.

## Day 10 — Recovery, reproducibility and interviews

**Goal:** Verify the claims you will put on the CV and leave a usable repository.

| Hours | Work |
|---|---|
| 1.5 | Follow setup instructions from a clean starting point |
| 2 | Perform database backup/restore and deployment recovery |
| 1.5 | Fix documentation and verify cleanup/state recovery |
| 1.5 | Record a short demonstration and finish incident reports |
| 1.5 | Practise interview explanations and final acceptance review |

### Tasks

- Perform a clean namespace/application redeployment using documented configuration, secrets and GHCR pulls.
- If claiming full server reproducibility, also recreate/rebuild the VPS after protecting code, state and backups, using an execution path outside that VPS. Otherwise state clearly that only application redeployment was verified.
- Create synthetic records, take a logical PostgreSQL backup and restore into a separate database. Check record counts and representative values.
- Keep a protected off-server backup; a second directory on the same VPS is not disaster recovery.
- Re-run one failed-release recovery and write incident report 2.
- Check CI, dashboards, alert delivery and operator behaviour after changes.
- Document all ongoing charges and the exact teardown procedure. Delete temporary resources. If keeping the demo online, explicitly list what remains billable.
- Scan repository history/artifacts for accidental secrets and ensure state is excluded.
- Update README with actual tested features and known limitations.

### Final acceptance checklist

- [ ] CRUD API and database integration tests pass.
- [ ] Kubernetes persists data across pod recreation.
- [ ] Terraform configuration validates and matches documented provider resources.
- [ ] Fresh image pulls work; images are tied to commits/digests.
- [ ] Helm configuration renders correctly.
- [ ] GitHub Actions tests, scans and publishes trusted changes.
- [ ] Jenkins demonstrates a SonarQube gate and alternative CI.
- [ ] Argo CD promotion and recovery through Git are demonstrated.
- [ ] Prometheus/Grafana show real application behaviour.
- [ ] Alert delivery and resolution are tested.
- [ ] Operator corrects drift and cleans up owned resources.
- [ ] Database restoration is verified.
- [ ] Architecture, costs, runbooks and two incident reports are committed.
- [ ] Reproducibility claims match what was actually rebuilt.
- [ ] CV describes this as a personal project and does not claim EKS experience.

## Interview demonstration: 5–8 minutes

1. Explain the one-server architecture and its failure boundaries.
2. Create/read a task and show persistence.
3. Show a successful CI run and the resulting image digest.
4. Show GitOps desired state and application health.
5. Generate requests and show metrics/alerts from a prepared incident.
6. Delete the operator-owned worker Deployment and show recreation.
7. Explain recovery evidence and one trade-off you would change in a real production system.

Do not attempt every disruptive exercise live. Keep recorded evidence and reproducible commands for longer operations.

## If the schedule slips

Keep core Go, Kubernetes, Terraform, delivery and monitoring gates. Keep Jenkins manual, the operator narrow, and the API private. Drop optional automation, UI polish, public ingress and additional dashboards first. If an essential gate remains incomplete, extend the schedule and label it honestly rather than replacing verification with screenshots.

## Official documentation starting points

Consult current compatible versions during implementation:

- Go: https://go.dev/doc/
- PostgreSQL: https://www.postgresql.org/docs/
- Terraform: https://developer.hashicorp.com/terraform/docs
- K3s: https://docs.k3s.io/
- Kubernetes: https://kubernetes.io/docs/
- Helm: https://helm.sh/docs/
- GitHub Actions: https://docs.github.com/en/actions
- GHCR: https://docs.github.com/en/packages
- Jenkins: https://www.jenkins.io/doc/
- SonarQube Community Build: https://docs.sonarsource.com/sonarqube-community-build/
- Argo CD: https://argo-cd.readthedocs.io/en/stable/
- Prometheus: https://prometheus.io/docs/introduction/overview/
- Grafana: https://grafana.com/docs/grafana/latest/
- Kubebuilder: https://book.kubebuilder.io/
- Trivy: https://trivy.dev/
- k6: https://grafana.com/docs/k6/latest/
