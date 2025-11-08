package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	backendServiceURL string

	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Define global logger
	logger *slog.Logger
)

func init() {
	// TODO:
	// Initialize the `logger` variable.
	// 1. Create a `slog.NewJSONHandler` (writing to `os.Stdout`).
	// 2. Create a `slog.New` logger using this handler.
	// 3. Add a permanent attribute: .With("service", "frontend-app")
	//
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "frontend-app")
	//

	logger.Info("Initializing metrics...")

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	logger.Info("Registering metrics...")
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	logger.Info("Metrics successfully registered.")
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		if path == "" {
			path = "unknown"
		}
		statusCodeStr := strconv.Itoa(rw.statusCode)
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, path, statusCodeStr).Inc()

	})
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("ERROR: couldn't read request body: %v", err)
		//
		logger.Error("Could not read request body", "error", err)
		//
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := http.Post(backendServiceURL+"/generate", "text/plain", bytes.NewReader(longURL))
	if err != nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("ERROR: Backend connection failed: %v", err)
		//
		logger.Error("Could not read request body", "error", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	shortLink, _ := io.ReadAll(resp.Body)
	// TODO: Replace log.Printf with slog.Info
	// Old log: log.Printf("INFO: Link shortened: %s -> %s", string(longURL), string(shortLink))
	//
	logger.Info("Link shortened", "long_url", string(longURL), "short_link", string(shortLink))
	//
	w.Write(shortLink)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	resp, err := http.Get(backendServiceURL + "/resolve/" + shortLink)
	if err != nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("ERROR: Backend connection failed: %v", err)
		//
		logger.Error("Backend call failed", "error", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		// TODO: Add slog.Warn
		// (No log.Printf here, but add a WARN log)
		// Use `logger.Warn` to log "Link not found". Add `shortLink` as an attribute.
		//
		logger.Warn("Link not found", "short_link", shortLink)
		//
		http.NotFound(w, r)
		return
	}

	longURL, _ := io.ReadAll(resp.Body)
	// TODO: Replace log.Printf with slog.Info
	// Old log: log.Printf("INFO: Redirect: %s -> %s", shortLink, string(longURL))
	//
	logger.Info("Redirecting link", "short_link", shortLink, "long_url", string(longURL))
	//
	http.Redirect(w, r, string(longURL), http.StatusFound)
}

func main() {
	backendServiceURL = os.Getenv("BACKEND_SVC_URL")
	if backendServiceURL == "" {
		backendServiceURL = "http://backend-app-svc:8081"
	}
	// TODO: Replace log.Printf with slog.Info
	// Old log: log.Printf("INFO: Backend-Service URL on: %s", backendServiceURL)
	//
	logger.Info("Backend service URL", "url", backendServiceURL)
	//

	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.HandleFunc("/{shortlink}", redirectHandler).Methods("GET")
	r.Use(prometheusMiddleware)

	go func() {
		metricsRouter := mux.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		// TODO: Replace log.Println with slog.Info
		// log.Println("INFO: Metrics server started on Port 9090")
		//
		logger.Info("Metrics server starting", "port", 9090)
		//
		if err := http.ListenAndServe(":9090", metricsRouter); err != nil {
			logger.Error("Metrics server failed", "error", err)
			os.Exit(1)
		}
	}()

	// TODO: Replace log.Println with slog.Info
	// Old log: log.Println("INFO: Frontend-Service starting on Port 8080")
	//
	logger.Info("Frontend service starting", "port", 8080)
	//
	http.ListenAndServe(":8080", r)
}
