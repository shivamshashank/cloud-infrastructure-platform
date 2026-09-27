package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Pointers distinguish an omitted PATCH field from a provided empty string.
type taskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

type app struct{ db *pgxpool.Pool }

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	config, err := poolConfig(databaseURL, os.Getenv("DB_MAX_CONNS"), os.Getenv("DB_MIN_CONNS"))
	if err != nil {
		slog.Error("invalid database configuration", "error", err)
		os.Exit(1)
	}
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.Ping(ctx)
	cancel()
	if err != nil {
		slog.Error("connect to database", "error", err)
		os.Exit(1)
	}

	a := &app{db: db}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           a.handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("API listening", "address", server.Addr, "db_max_conns", config.MaxConns)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "error", err)
	}
}

func (a *app) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *app) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.db.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *app) listTasks(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 100 {
			writeError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = n
	}
	offset := 0
	if s := r.URL.Query().Get("offset"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = n
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := a.db.Query(ctx, `SELECT id, title, description, status, created_at, updated_at
		FROM tasks ORDER BY id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		serverError(w, r, err)
		return
	}
	defer rows.Close()
	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := scanTask(rows, &task); err != nil {
			serverError(w, r, err)
			return
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string][]Task{"tasks": tasks})
}

func (a *app) createTask(w http.ResponseWriter, r *http.Request) {
	var input taskInput
	if !readInput(w, r, &input) {
		return
	}
	if input.Title == nil || strings.TrimSpace(*input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.Status != nil && !validStatus(*input.Status) {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	description, status := "", "todo"
	if input.Description != nil {
		description = *input.Description
	}
	if input.Status != nil {
		status = *input.Status
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var task Task
	err := scanTask(a.db.QueryRow(ctx, `INSERT INTO tasks (title, description, status)
		VALUES ($1, $2, $3) RETURNING id, title, description, status, created_at, updated_at`,
		strings.TrimSpace(*input.Title), description, status), &task)
	if err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Location", "/tasks/"+strconv.FormatInt(task.ID, 10))
	writeJSON(w, http.StatusCreated, task)
}

func (a *app) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var task Task
	err := scanTask(a.db.QueryRow(ctx, `SELECT id, title, description, status, created_at, updated_at
		FROM tasks WHERE id = $1`, id), &task)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "task not found")
	} else if err != nil {
		serverError(w, r, err)
	} else {
		writeJSON(w, http.StatusOK, task)
	}
}

func (a *app) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var input taskInput
	if !readInput(w, r, &input) {
		return
	}
	if input.Title == nil && input.Description == nil && input.Status == nil {
		writeError(w, http.StatusBadRequest, "provide a field to update")
		return
	}
	if input.Title != nil {
		*input.Title = strings.TrimSpace(*input.Title)
		if *input.Title == "" {
			writeError(w, http.StatusBadRequest, "title cannot be blank")
			return
		}
	}
	if input.Status != nil && !validStatus(*input.Status) {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var task Task
	err := scanTask(a.db.QueryRow(ctx, `UPDATE tasks SET
		title = COALESCE($2, title), description = COALESCE($3, description),
		status = COALESCE($4, status), updated_at = now()
		WHERE id = $1 RETURNING id, title, description, status, created_at, updated_at`,
		id, input.Title, input.Description, input.Status), &task)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "task not found")
	} else if err != nil {
		serverError(w, r, err)
	} else {
		writeJSON(w, http.StatusOK, task)
	}
}

func (a *app) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	result, err := a.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		serverError(w, r, err)
	} else if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "task not found")
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func scanTask(row interface{ Scan(...any) error }, task *Task) error {
	return row.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
}

func taskID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return 0, false
	}
	return id, true
}

func validStatus(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}

func readInput(w http.ResponseWriter, r *http.Request, input *taskInput) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "only one JSON object is allowed")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "request_id", r.Context().Value(requestIDKey{}), "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}
