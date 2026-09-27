# Container walkthrough

This exercise packages the existing Go API. It does not create AWS resources.

## 1. Understand the build

Read [`Dockerfile`](../Dockerfile) from top to bottom:

1. `golang:1.27.1-alpine` is the **build stage**. It downloads modules and
   compiles the API and migration commands. Copying `go.mod` and `go.sum`
   first lets Docker reuse the module layer when only application code changes.
2. `CGO_ENABLED=0` produces static Go binaries. `-trimpath` removes local
   source paths; `-s -w` removes symbol/debug tables to reduce image size.
3. Each final stage copies only its executable into a distroless image. The
   Go toolchain and source files are absent from the running images.
4. UID/GID `65532` runs the executables without root privileges. The API
   listens on port `8080`; ports above 1024 do not require root.

The `.dockerignore` file keeps unrelated files out of the build context.
The final images have no shell, so inspect them with `docker image inspect`
and use container logs for troubleshooting.

## 2. Start the stack

From the repository root, with a working Docker-compatible daemon:

```bash
export POSTGRES_PASSWORD=localdev
docker compose up -d --build --wait api
docker compose ps -a
```

Compose starts `db`, waits for its health check, runs `migrate`, waits for a
successful exit, and then starts `api`. The API connects to `db:5432` on the
internal Compose network. Only `127.0.0.1:8080` exposes the API to your host.
PostgreSQL also binds to host `127.0.0.1:5432` for local Go tests. Do not use
the example password outside this local exercise.

## 3. Prove it works

```bash
python3 scripts/smoke_test.py
docker image inspect cloud-infrastructure-platform-api:latest --format '{{.Config.User}}'
docker compose logs --no-color migrate api
```

The smoke test checks liveness, readiness, create/read/list/update/delete, and
request IDs. Image inspection should print `65532:65532`. The migration log
should say `migrations applied`, followed by the API listening log.

## 4. Recover and clean up

If startup fails, run `docker compose ps -a` and `docker compose logs db
migrate api`. A failed migration prevents the API from starting, so fix the
database or migration error first. To rebuild after editing Go code, repeat
`docker compose up -d --build --wait api`.

```bash
docker compose down
```

This removes the containers and network while keeping the PostgreSQL named
volume. `docker compose down -v` also deletes that volume and its data; use it
only when you deliberately want a fresh database.
