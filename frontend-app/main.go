package main

import (
	"bytes"
	"io"
	"log"
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
	// Code goes here ...
	//

	// TODO: Replace the old log line
	log.Printf("INFO: Initializing metrics")
	//

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

	// TODO: Replace the old log line
	log.Println("INFO: Registering metrics...")
	//
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	// TODO: Replace the old log line
	log.Println("INFO: Metrics successfully registered.")
	//
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
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: couldn't read request body: %v", err)
		//
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := http.Post(backendServiceURL+"/generate", "text/plain", bytes.NewReader(longURL))
	if err != nil {
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: Backend connection failed: %v", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	shortLink, _ := io.ReadAll(resp.Body)
	// TODO: Replace the old log line
	//
	log.Printf("INFO: Link shortened: %s -> %s", string(longURL), string(shortLink))
	//
	w.Write(append(shortLink, '\n'))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	resp, err := http.Get(backendServiceURL + "/resolve/" + shortLink)
	if err != nil {
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: Backend connection failed: %v", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		// TODO: No log.Printf here, but add a WARN log to log "Link not found"
		// Add `shortLink` as an attribute.
		//
		// Code goes here ...
		//
		http.NotFound(w, r)
		return
	}

	longURL, _ := io.ReadAll(resp.Body)
	// TODO: Replace the old log line
	//
	log.Printf("INFO: Redirect: %s -> %s", shortLink, string(longURL))
	//
	http.Redirect(w, r, string(longURL), http.StatusFound)
}

func main() {
	backendServiceURL = os.Getenv("BACKEND_SVC_URL")
	if backendServiceURL == "" {
		backendServiceURL = "http://backend-app-svc:8081"
	}
	// TODO: Replace the old log line
	//
	log.Printf("INFO: Backend-Service URL on: %s", backendServiceURL)
	//

	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.HandleFunc("/{shortlink}", redirectHandler).Methods("GET")
	r.Use(prometheusMiddleware)

	go func() {
		metricsRouter := mux.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		// TODO: Replace the old log line
		//
		log.Println("INFO: Metrics server started on Port 9090")
		//
		if err := http.ListenAndServe(":9090", metricsRouter); err != nil {
			// TODO: Replace the old log line and exit with a non zero status code after it
			log.Fatalf("FATAL: Couldn't start metrics server: %v", err)
		}
	}()

	// TODO: Replace the old log line
	//
	log.Println("INFO: Frontend-Service starting on Port 8080")
	//
	http.ListenAndServe(":8080", r)
}
