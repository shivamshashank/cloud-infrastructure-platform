package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	(&app{}).health(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("health response: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCreateRejectsInvalidInputBeforeDatabase(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing title", `{"description":"example"}`},
		{"blank title", `{"title":"   "}`},
		{"invalid status", `{"title":"example","status":"unknown"}`},
		{"unknown field", `{"title":"example","id":99}`},
		{"multiple objects", `{"title":"example"}{"title":"another"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			(&app{}).createTask(w, r)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestGetRejectsInvalidIDBeforeDatabase(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	r.SetPathValue("id", "abc")
	(&app{}).getTask(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHealthAndReadinessWhenDatabaseIsUnavailable(t *testing.T) {
	db, err := pgxpool.New(context.Background(),
		"postgres://tasks:localdev@127.0.0.1:1/tasks?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := (&app{db: db}).handler()
	for _, tt := range []struct {
		path string
		want int
	}{
		{"/healthz", http.StatusOK},
		{"/readyz", http.StatusServiceUnavailable},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if w.Code != tt.want {
			t.Errorf("%s: got %d, want %d", tt.path, w.Code, tt.want)
		}
		if w.Header().Get("X-Request-ID") == "" {
			t.Errorf("%s: missing X-Request-ID", tt.path)
		}
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "task_api_http_requests_total") {
		t.Fatalf("metrics response: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPoolConfig(t *testing.T) {
	const url = "postgres://tasks:localdev@127.0.0.1:5432/tasks?sslmode=disable"
	config, err := poolConfig(url, "7", "2")
	if err != nil || config.MaxConns != 7 || config.MinConns != 2 {
		t.Fatalf("config=%+v err=%v", config, err)
	}
	if _, err := poolConfig(url, "2", "3"); err == nil {
		t.Fatal("expected min > max to fail")
	}
}
