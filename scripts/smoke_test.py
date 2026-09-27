#!/usr/bin/env python3
"""Check a deployed Task API with one synthetic task."""

import argparse
import json
import sys
import urllib.error
import urllib.request
import uuid


def request(base_url, method, path, payload=None):
    body = json.dumps(payload).encode() if payload is not None else None
    headers = {"Content-Type": "application/json"} if body is not None else {}
    req = urllib.request.Request(
        base_url.rstrip("/") + path, data=body, headers=headers, method=method
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            status = response.status
            data = response.read()
            request_id = response.headers.get("X-Request-ID")
    except urllib.error.HTTPError as error:
        status = error.code
        data = error.read()
        request_id = error.headers.get("X-Request-ID")
    if not request_id:
        raise RuntimeError(f"{method} {path}: missing X-Request-ID")
    return status, json.loads(data) if data else None


def expect(base_url, method, path, expected, payload=None):
    status, data = request(base_url, method, path, payload)
    if status != expected:
        raise RuntimeError(f"{method} {path}: expected {expected}, got {status}: {data}")
    return data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    args = parser.parse_args()
    task_id = None
    deleted = False
    try:
        expect(args.url, "GET", "/healthz", 200)
        expect(args.url, "GET", "/readyz", 200)
        created = expect(
            args.url,
            "POST",
            "/tasks",
            201,
            {"title": f"smoke-{uuid.uuid4().hex[:8]}"},
        )
        task_id = created["id"]
        path = f"/tasks/{task_id}"
        expect(args.url, "GET", path, 200)
        listed = expect(args.url, "GET", "/tasks?limit=1&offset=0", 200)
        if not isinstance(listed.get("tasks"), list):
            raise RuntimeError("GET /tasks did not return a task list")
        updated = expect(args.url, "PATCH", path, 200, {"status": "done"})
        if updated["status"] != "done":
            raise RuntimeError("PATCH did not persist the new status")
        expect(args.url, "DELETE", path, 204)
        deleted = True
        expect(args.url, "GET", path, 404)
        print(f"PASS: health, readiness, CRUD and request IDs (task {task_id})")
        return 0
    except (RuntimeError, OSError, ValueError, KeyError) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        return 1
    finally:
        if task_id is not None and not deleted:
            try:
                request(args.url, "DELETE", f"/tasks/{task_id}")
            except (OSError, ValueError):
                pass


if __name__ == "__main__":
    raise SystemExit(main())
