package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivamshashank/cloud-infrastructure-platform/migrations"
)

func TestDatabaseCRUD(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to a dedicated tasks_test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	var name string
	if err := db.QueryRow(ctx, "SELECT current_database()").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "tasks_test" {
		t.Fatalf("refusing to modify database %q; use tasks_test", name)
	}
	t.Cleanup(func() {
		cleanup, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if _, err := db.Exec(cleanup, "TRUNCATE tasks RESTART IDENTITY"); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatalf("migration rerun must be safe: %v", err)
	}
	var migrationCount int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
		t.Fatal(err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one applied migration, got %d", migrationCount)
	}
	if _, err := db.Exec(ctx, "TRUNCATE tasks RESTART IDENTITY"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer((&app{db: db}).handler())
	defer func() { server.Close() }()
	client := server.Client()
	request := func(method, path, body string, want int) ([]byte, http.Header) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != want {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, resp.StatusCode, want, data)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatal("missing request ID")
		}
		return data, resp.Header
	}

	data, header := request(http.MethodPost, "/tasks", `{"title":"  Learn Go  "}`, http.StatusCreated)
	var created Task
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Title != "Learn Go" || created.Status != "todo" || created.ID < 1 {
		t.Fatalf("created task: %+v", created)
	}
	path := "/tasks/" + strconv.FormatInt(created.ID, 10)
	if header.Get("Location") != path {
		t.Fatalf("Location=%q, want %q", header.Get("Location"), path)
	}
	request(http.MethodGet, path, "", http.StatusOK)
	request(http.MethodGet, "/tasks?limit=1&offset=0", "", http.StatusOK)

	// A new handler simulates an API restart while the same database retains data.
	server.Close()
	server = httptest.NewServer((&app{db: db}).handler())
	client = server.Client()
	request(http.MethodGet, path, "", http.StatusOK)

	data, _ = request(http.MethodPatch, path, `{"status":"done"}`, http.StatusOK)
	var updated Task
	if err := json.Unmarshal(data, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Status != "done" || updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Fatalf("updated task: %+v", updated)
	}
	request(http.MethodGet, "/tasks/999999", "", http.StatusNotFound)
	request(http.MethodPost, "/tasks", `{"title":""}`, http.StatusBadRequest)
	request(http.MethodDelete, path, "", http.StatusNoContent)
	request(http.MethodDelete, path, "", http.StatusNotFound)
	request(http.MethodGet, path, "", http.StatusNotFound)
}
