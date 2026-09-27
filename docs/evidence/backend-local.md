# Backend local verification — 2026-09-27

Environment: Go 1.27.1 on macOS arm64; PostgreSQL 17 in local Docker Compose.
No AWS resources were used.

| Check | Result |
|---|---|
| `go test ./...` with `TEST_DATABASE_URL` pointing at `tasks_test` | Passed after fixing route parameter handling |
| `go vet ./...` | Passed |
| `go test -race ./...` with `TEST_DATABASE_URL` | Passed |
| `go test -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out ./...` | Passed; local total statement coverage 65.7% |
| `python3 scripts/smoke_test.py` | Passed health, readiness, CRUD, and request IDs |
| Create task, restart API, read task | Task remained available from PostgreSQL |
| Stop PostgreSQL while API runs | `/healthz` returned 200; `/readyz` returned 503 |
| Restart PostgreSQL | `/readyz` returned 200 again |
| `GET /metrics` | Exposed request counter, duration histogram buckets, and in-flight gauge |

The synthetic persistence-check task was deleted after verification. The local
PostgreSQL volume was retained for repeatable exercises. This evidence does not
establish Kubernetes availability, cloud deployment, or a latency SLO.
