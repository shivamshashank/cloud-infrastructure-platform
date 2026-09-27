package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type requestIDKey struct{}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *responseRecorder) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (a *app) handler() http.Handler {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_api_http_requests_total",
		Help: "Total HTTP requests handled by the Task API.",
	}, []string{"method", "route", "status_class"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_api_http_request_duration_seconds",
		Help:    "Task API HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	inFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "task_api_http_requests_in_flight",
		Help: "Task API HTTP requests currently running.",
	})
	registry.MustRegister(requests, duration, inFlight)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.ready)
	mux.HandleFunc("GET /tasks", a.listTasks)
	mux.HandleFunc("POST /tasks", a.createTask)
	mux.HandleFunc("GET /tasks/{id}", a.getTask)
	mux.HandleFunc("PATCH /tasks/{id}", a.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", a.deleteTask)
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var idBytes [16]byte
		if _, err := rand.Read(idBytes[:]); err != nil {
			writeError(w, http.StatusInternalServerError, "request ID unavailable")
			return
		}
		id := hex.EncodeToString(idBytes[:])
		w.Header().Set("X-Request-ID", id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))

		_, pattern := mux.Handler(r)
		if pattern == "" {
			pattern = "unmatched"
		}
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}
		if r.URL.Path != "/metrics" {
			inFlight.Inc()
			defer inFlight.Dec()
		}
		mux.ServeHTTP(recorder, r)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}
		elapsed := time.Since(start)
		if r.URL.Path != "/metrics" {
			method := metricMethod(r.Method)
			requests.WithLabelValues(method, pattern, strconv.Itoa(recorder.status/100)+"xx").Inc()
			duration.WithLabelValues(method, pattern).Observe(elapsed.Seconds())
		}
		slog.Info("HTTP request", "request_id", id, "method", r.Method,
			"route", pattern, "status", recorder.status, "duration_ms", elapsed.Milliseconds())
	})
}

func metricMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
		return method
	default:
		return "OTHER"
	}
}
