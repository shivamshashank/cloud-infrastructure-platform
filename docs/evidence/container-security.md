# Container security evidence

Date: 2026-09-27. Scanner: Trivy 0.74.0. Scope: OS and Go package
vulnerabilities at HIGH and CRITICAL severity. Both runtime images were
exported with `docker save` and scanned as tar files; Trivy had no access to
the host Docker socket.

| Image | Before | Change | After |
|---|---|---|---|
| API | 2 CRITICAL in `github.com/jackc/pgx/v5` 5.8.0; 0 OS findings | Upgrade pgx to 5.11.0 | 0 HIGH/CRITICAL findings |
| Migration | Same 2 CRITICAL findings; 0 OS findings | Rebuild with pgx 5.11.0 | 0 HIGH/CRITICAL findings |

The initial findings were CVE-2026-33815 and CVE-2026-33816. Trivy reported
5.9.0 as the first fixed version. We used the current 5.11.0 release, which
also supports Go 1.27. Both rescans used `--exit-code 1` and exited 0.

After the upgrade, the race-enabled PostgreSQL integration test passed with
65.7% statement coverage, and the container API smoke test passed.

An all-severity scan of the rebuilt API image also found one `UNKNOWN`
severity Debian `tzdata` advisory (`DLA-4792-1`), with a newer package version
listed as a fix. Pulling the current distroless base image did not yet change
that package. It is recorded for review but is outside the HIGH/CRITICAL CI
gate; the Go binary had no findings in the all-severity scan.

This is a time-specific result. A new vulnerability database or rebuilt base
image can change the findings; CI rescans on every push.
