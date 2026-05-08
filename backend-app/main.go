package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ctx    = context.Background()
	rdb    *redis.Client
	logger *slog.Logger
	// New OTel Tracer
	tracer trace.Tracer
)

func init() {

	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "backend-app")

	// TODO: Initialize OTel
	// Call the `initTracerProvider` function (from tracing.go)
	// to set up the OTel SDK.
	//
	// Also, get a global `Tracer` instance from OpenTelemetry's
	// global provider, so we can create manual spans.
	//
	if _, err := initTracerProvider(logger); err != nil {
		logger.Error("Failed to initialize OTel TracerProvider", "error", err)
	}
	tracer = otel.Tracer("backend-app-tracer")
	//
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Create a manual Span
	// The OTel middleware (from Task 3) will create a span for the router,
	// but we want to *also* create a more specific span for this
	// entire handler function.
	//
	// Start a new span here (e.g., "generateHandler") from the
	// request's context.
	//
	// **Crucially**: Remember to end the span when the function finishes.
	//
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "generateHandler")
	defer span.End()
	//

	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortLink := fmt.Sprintf("id%d", rand.Intn(10000))

	// TODO: Create a manual Child Span
	// The span for "generateHandler" is good, but we need more detail.
	// We suspect the database call might be slow.
	//
	// Create a *new child span* that *only* wraps the `rdb.Set(...)` call.
	// Give it a descriptive name (e.g., "redis-set").
	//
	// Also, add the "short_link" as an attribute to this new span
	// so we can identify it in our trace.
	//
	redisCtx, redisSpan := tracer.Start(ctx, "redis-set")
	redisSpan.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.key", shortLink),
	)
	//

	// Pass the context from the *new* child span to the Redis call
	err = rdb.Set(redisCtx, shortLink, string(longURL), time.Hour*24).Err()

	// TODO: (continued)
	// End the child span immediately after the Redis call finishes.
	//
	redisSpan.End()
	//

	if err != nil {
		logger.Error("Redis SET failed", "error", err, "short_link", shortLink)
		http.Error(w, "Internal error in backend-app", http.StatusInternalServerError)
		return
	}

	logger.Info("Mapping created", "short_link", shortLink, "long_url", string(longURL))
	w.Write([]byte(shortLink))
}

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: (continued)
	// Do the same as in `generateHandler`:
	// Start a new span that measures this entire function.
	//
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "resolveHandler")
	defer span.End()
	//

	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	// TODO: (continued)
	// Do the same as in `generateHandler`:
	// Create a new child span that *only* wraps the `rdb.Get(...)` call.
	// Give it a descriptive name (e.g., "redis-get")
	// and add the "short_link" as an attribute.
	//
	redisCtx, redisSpan := tracer.Start(ctx, "redis-get")
	redisSpan.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.key", shortLink),
	)
	//

	longURL, err := rdb.Get(redisCtx, shortLink).Result()

	// TODO: (continued)
	// End the child span immediately after the Redis call finishes.
	//
	redisSpan.End()
	//

	if err == redis.Nil {
		logger.Warn("Link not found", "short_link", shortLink)
		http.NotFound(w, r)
		return
	} else if err != nil {
		logger.Error("Redis GET failed", "error", err, "short_link", shortLink)
		http.Error(w, "Internal error in backend-app", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(longURL))
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis-svc:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	logger.Info("Connecting to Redis", "address", redisAddr)

	r := mux.NewRouter()
	r.HandleFunc("/generate", generateHandler).Methods("POST")
	r.HandleFunc("/resolve/{shortlink}", resolveHandler).Methods("GET")

	// TODO: Add OTel Middleware
	// Just like the frontend, our backend router needs to be wrapped.
	//
	// Apply the OTel Mux middleware here. This middleware will
	// automatically *extract* the trace headers from the frontend's
	// request and continue the trace.
	//
	r.Use(otelmux.Middleware("backend-router"))
	//

	logger.Info("Backend service starting", "port", 8081)

	if err := http.ListenAndServe(":8081", r); err != nil {
		logger.Error("Backend server failed to start", "error", err)
	}
}
