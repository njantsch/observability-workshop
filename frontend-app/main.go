package main

import (
	"bytes"
	"io"
	"log"
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

	// Define metric variables globally
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
)

func init() {
	log.Printf("INFO: Initializing metrics")

	// TODO:
	// Create the `httpRequestsTotal` CounterVec.
	// - Name: "http_requests_total"
	// - Help: "Total number of HTTP requests."
	// - Labels: "method", "path", "code"
	//
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests.",
			ConstLabels: prometheus.Labels{"team": teamName},
		},
		[]string{"method", "path", "code"},
	)

	// Create the `httpRequestDuration` HistogramVec.
	// - Name: "http_request_duration_seconds"
	// - Help: "HTTP request duration in seconds."
	// - Labels: "method", "path"
	//
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds.",
			Buckets:     prometheus.DefBuckets,
			ConstLabels: prometheus.Labels{"team": teamName},
		},
		[]string{"method", "path"},
	)

	log.Println("INFO: Registering metrics...")
	// TODO:
	// Register both metrics with Prometheus.
	//
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	//
	log.Println("INFO: Metrics successfully registered.")
}

// responseWriter is a wrapper for http.ResponseWriter,
// allowing us to capture the status code.
// (This struct is provided for you. No changes needed.)
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

// prometheusMiddleware is our middleware that instruments every request.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Call the next handler in the chain
		next.ServeHTTP(rw, r)

		//Record metrics after the request has been handled
		duration := time.Since(start).Seconds()

		// Get the route (e.g., "/shorten" or "/{shortlink}")
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		if path == "" {
			path = "unknown"
		}

		statusCodeStr := strconv.Itoa(rw.statusCode)

		// TODO:
		// 1. Observe the request duration with the Histogram.
		//
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		//

		// 2. Increment the request counter.
		//
		httpRequestsTotal.WithLabelValues(r.Method, path, statusCodeStr).Inc()
		//

	})
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: couldn't read request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := http.Post(backendServiceURL+"/generate", "text/plain", bytes.NewReader(longURL))
	if err != nil {
		log.Printf("ERROR: Backend connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	shortLink, _ := io.ReadAll(resp.Body)
	log.Printf("INFO: Link shortened: %s -> %s", string(longURL), string(shortLink))
	w.Write(shortLink)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	resp, err := http.Get(backendServiceURL + "/resolve/" + shortLink)
	if err != nil {
		log.Printf("ERROR: Backend connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}

	longURL, _ := io.ReadAll(resp.Body)
	log.Printf("INFO: Redirect: %s -> %s", shortLink, string(longURL))
	http.Redirect(w, r, string(longURL), http.StatusFound)
}

func main() {
	backendServiceURL = os.Getenv("BACKEND_SVC_URL")
	if backendServiceURL == "" {
		backendServiceURL = "http://backend-app-svc:8081"
	}
	log.Printf("INFO: Backend-Service URL on: %s", backendServiceURL)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.HandleFunc("/{shortlink}", redirectHandler).Methods("GET")
	// TODO: Apply the `prometheusMiddleware` to the main router `r`.
	//
	r.Use(prometheusMiddleware)
	//

	// Start the /metrics server on port 9090
	// (This part is provided for you. No changes needed.)
	go func() {
		metricsRouter := mux.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		log.Println("INFO: Metrics server started on Port 9090")
		if err := http.ListenAndServe(":9090", metricsRouter); err != nil {
			log.Fatalf("FATAL: Couldn't start metrics server: %v", err)
		}
	}()

	log.Println("INFO: Frontend-Service starting on Port 8080")
	http.ListenAndServe(":8080", r)
}
