# Consul and Vault learning flow

**Status: planned.** Neither tool is installed or connected to the Go API.
Start the local exercises after the API and container basics. Integrate with
Kubernetes only after its baseline deployment works. Repeat on a temporary
AWS K3s host only after measuring its memory use and reviewing the cost of
keeping that host running.

## Why each tool is here

| Tool | Focused exercise | Boundary |
|---|---|---|
| Consul | Register a test service with an HTTP health check; query its catalog and DNS; observe how failed health changes discovery | Kubernetes DNS remains the normal discovery path for the Go API. Consul is an additional service discovery lab, not required for a one-service cluster. |
| Vault | Store the synthetic PostgreSQL connection secret under a narrow read policy; deliver it to Kubernetes; rotate it and verify recovery | Vault is the secret source in this exercise. Git, Terraform variables/state, container images, and logs must not contain the credential. |

HashiCorp documents [Consul service discovery on Kubernetes](https://developer.hashicorp.com/consul/docs/discover/k8s)
and [Vault Secrets Operator](https://developer.hashicorp.com/vault/docs/deploy/kubernetes/vso/sources/vault).
The operator can sync Vault data into a Kubernetes Secret and trigger a rollout
restart when a consumer cannot reload credentials. The current Go process reads
`DATABASE_URL` at startup, so a restart is part of the rotation drill.

## Ordered exercises and evidence

1. **Consul locally:** Run a temporary Consul server in an isolated local lab.
   Register a disposable HTTP service with a health check, look it up through
   the catalog and Consul DNS, then stop the service. Record the healthy and
   failing lookup results. Remove the registration and stop Consul.
2. **Vault locally:** Use a temporary Vault development server with synthetic
   credentials only. Create a KV secret and a policy that can read just that
   path. Prove that a broader read is denied. Do not use development mode on
   the AWS host: it uses insecure defaults and loses data on restart, as the
   [Vault documentation](https://developer.hashicorp.com/vault/docs/concepts/dev-server)
   states. Stop the server and discard its test data.
3. **Kubernetes baseline:** Deploy the API and PostgreSQL with Helm/Argo CD,
   using a manually created Kubernetes Secret first. Confirm the migration Job,
   `/readyz`, and the smoke test work. Do this before adding another dependency.
4. **Vault delivery:** In a temporary cluster lab, install a properly
   configured Vault and Vault Secrets Operator with pinned versions, private
   access, TLS, persistent storage, and a recovery procedure. Give the
   operator a dedicated Kubernetes identity and a policy limited to the API's
   database secret. Sync that path into the Secret referenced by the API and
   migration Job. Wait for the Secret to exist before the migration Job runs.
   GitOps stores only non-secret references and configuration.
5. **Rotation:** Replace the synthetic database credential in PostgreSQL and
   Vault. Verify the synced Secret changes, restart the API to pick up the new
   `DATABASE_URL`, run the smoke test, and confirm the old credential can no
   longer connect. Record the observed outage, if any. Add automated rollout
   restart only after the manual sequence works.
6. **Consul with Kubernetes:** If the cluster has enough spare memory, register
   or sync a disposable Kubernetes service into Consul and query it from a
   test client. Compare the result with Kubernetes DNS. Do not redirect the
   API's normal service traffic through Consul just to claim another tool.
7. **Teardown:** Remove the disposable registration and clients; stop the
   Consul and Vault lab workloads; clean up test secrets and policies; then
   destroy the temporary AWS resources using the infrastructure runbook.
   Check that no persistent volume or billable host remains.

## Completion checks

- [ ] Consul catalog/DNS returns a healthy test service and reflects a failed check.
- [ ] A narrow Vault identity reads only the intended synthetic secret.
- [ ] The API and migration Job use a Vault-synced Kubernetes Secret.
- [ ] Rotation reaches the running API, the smoke test passes, and old database access fails.
- [ ] Setup, recovery, resource usage, and teardown evidence are recorded.

Keep these items unchecked until the corresponding exercise is performed.
Vault and Consul add processes, storage, and failure modes. On the single-node
AWS lab, measure their resource use before leaving them running; keep the lab
short-lived to protect the AWS credit balance.
