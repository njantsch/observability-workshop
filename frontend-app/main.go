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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	backendServiceURL string
	logger            *slog.Logger
	// New OTel Tracer
	tracer trace.Tracer

	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// TODO: Define a global OTel HTTP Client
	// Right now, we are using Go's default HTTP client to call the backend.
	// This default client doesn't know anything about tracing.
	//
	// We need to define a new, global `*http.Client` variable here.
	// This client must be trace-aware, so it automatically sends
	// trace information (like Trace IDs) to any service it calls.
	//
	otelHttpClient *http.Client
	//
)

func init() {

	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "frontend-app")

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
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)

	// TODO: Initialize OTel
	// Call the `initTracerProvider` function (from tracing.go)
	// to set up the OTel SDK.
	//
	// Also, get a global `Tracer` instance from OpenTelemetry's
	// global provider, so we can create manual spans later if needed.
	//
	// (Your code here)
	if _, err := initTracerProvider(logger); err != nil {
		logger.Error("Failed to initialize OTel TracerProvider", "error", err)
	}
	tracer = otel.Tracer("frontend-app-tracer")
	//
	// TODO: (continued)
	// Now that the OTel SDK is initialized,
	// create the actual instrumented HTTP client you defined globally above.
	//
	// (Your code here)
	otelHttpClient = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
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
		logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Fehlerhafter Request", http.StatusBadRequest)
		return
	}

	// TODO: Use the OTel HTTP Client
	// Replace the default `http.Post` call below.
	//
	// Use your new, trace-aware http-client to send this request
	// to the backend.
	//
	// **Crucially**: You must pass the `context` from the incoming
	// request (`r`) to the *new* outgoing request. This is the "magic"
	// that connects the two services in a single trace.
	//
	// (Your code here, replace the line below)
	// resp, err := http.Post(backendServiceURL+"/generate", "text/plain", bytes.NewBuffer(longURL))
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, "POST", backendServiceURL+"/generate", bytes.NewBuffer(longURL))
	if err != nil {
		logger.Error("Failed to create backend request", "error", err)
		http.Error(w, "Internal error in frontend-app", http.StatusInternalServerError)
		return
	}
	resp, err := otelHttpClient.Do(req)
	//

	if err != nil {
		logger.Error("Backend call failed", "error", err)
		http.Error(w, "Internal error in frontend-app", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	shortLink, _ := io.ReadAll(resp.Body)

	logger.Info("Link shortened", "long_url", string(longURL), "short_link", string(shortLink))
	w.Write(shortLink)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	// TODO: (continued)
	// Do the same as in `shortenHandler`.
	//
	// Replace the default `http.Get` call below with a call
	// using your `otelHttpClient`.
	//
	// Remember to pass the `context` from the incoming request (`r`)
	// to the new outgoing request.
	//
	// (Your code here, replace the line below)
	// resp, err := http.Get(backendServiceURL + "/resolve/" + shortLink)
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, "GET", backendServiceURL+"/resolve/"+shortLink, nil)
	if err != nil {
		logger.Error("Failed to create backend request", "error", err)
		http.Error(w, "Internal error in frontend-app", http.StatusInternalServerError)
		return
	}
	resp, err := otelHttpClient.Do(req)
	//

	if err != nil {
		logger.Error("Backend call failed", "error", err)
		http.Error(w, "Internal error in frontend-app", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		logger.Warn("Link not found", "short_link", shortLink)
		http.NotFound(w, r)
		return
	}
	longURL, _ := io.ReadAll(resp.Body)

	logger.Info("Redirecting link", "short_link", shortLink, "long_url", string(longURL))
	http.Redirect(w, r, string(longURL), http.StatusFound)
}

func main() {
	backendServiceURL = os.Getenv("BACKEND_SVC_URL")
	if backendServiceURL == "" {
		backendServiceURL = "http://backend-app-svc:8081"
	}
	logger.Info("Backend service URL", "url", backendServiceURL)

	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.HandleFunc("/{shortlink}", redirectHandler).Methods("GET")

	// TODO: Add OTel Middleware
	// Our router `r` is "dumb" and doesn't know about traces.
	//
	// We need to apply a middleware to the router that
	// automatically creates a new trace span for every
	// incoming request.
	//
	// (Your code here)
	r.Use(otelmux.Middleware("frontend-router"))
	//

	r.Use(prometheusMiddleware)

	go func() {
		metricsRouter := mux.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		logger.Info("Metrics server starting", "port", 9090)
		if err := http.ListenAndServe(":9090", metricsRouter); err != nil {
			logger.Error("Metrics server failed to start", "error", err)
		}
	}()

	logger.Info("Frontend service starting", "port", 8080)

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Error("Frontend server failed to start", "error", err)
	}
}
