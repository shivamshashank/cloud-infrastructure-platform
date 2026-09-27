# Database migration procedure

The Task API does **not** alter its schema at startup. Run the separate
`cmd/migrate` program before starting a new API version.

## Local

Start PostgreSQL and set `DATABASE_URL` as shown in the
[README](../README.md#-quick-start), then run:

```bash
go run ./cmd/migrate
go run ./cmd/api
```

The runner embeds the numbered SQL files in `migrations/`, runs them in name
order inside a transaction, and records their SHA-256 checksums in
`schema_migrations`. An advisory lock prevents concurrent migration jobs from
applying the same version. Re-running it is safe; editing an already applied
file produces an error. Add a new numbered file for a new schema change.

## Kubernetes release sequence

1. Back up the database before a schema change and record the backup location.
2. Build the API and migration images from the same commit. The migration
   image contains its embedded SQL files. Pin both images by digest.
3. Run `migrate` as a Kubernetes Job with `DATABASE_URL` from a Secret. When
   the planned Vault Secrets Operator exercise is enabled, wait for it to sync
   that Secret before starting the Job; a missing Secret must stop the release.
   In Argo CD, use a **PreSync** hook Job for this step. Set a bounded active
   deadline and a small retry limit; require the Job to finish successfully.
4. Only after the Job succeeds, sync the API Deployment and run the smoke test.
5. If migration fails, stop the rollout, inspect the Job log, fix the cause,
   and rerun. The API version already running remains in place.

Make changes backward compatible while old and new API pods may overlap.
Prefer adding a nullable column or a new table first, then deploying code
that uses it. Remove old schema in a later release. Reverting an image does
not undo a database migration. Never put database credentials or dumps in Git.
