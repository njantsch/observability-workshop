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
)

var (
	ctx = context.Background()
	rdb *redis.Client

	logger *slog.Logger
)

func generateHandler(w http.ResponseWriter, r *http.Request) {
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

	shortLink := fmt.Sprintf("id%d", rand.Intn(10000))

	err = rdb.Set(ctx, shortLink, string(longURL), time.Hour*24).Err()
	if err != nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("ERROR: Redis Set failed: %v", err)
		//
		logger.Error("Redis SET failed", "error", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Replace log.Printf with slog.Error
	// Old log: log.Printf("INFO: Mapping created: %s -> %s", shortLink, string(longURL))
	//
	logger.Info("Mapping created", "short_link", shortLink, "long_url", string(longURL))
	//
	w.Write([]byte(shortLink))
}

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	longURL, err := rdb.Get(ctx, shortLink).Result()
	if err == redis.Nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("WARN: Link not found: %s", shortLink)
		//
		logger.Warn("Link not found", "short_link", shortLink)
		//
		http.NotFound(w, r)
		return
	} else if err != nil {
		// TODO: Replace log.Printf with slog.Error
		// Old log: log.Printf("ERROR: Redis Get failed: %v", err)
		//
		logger.Error("Redis GET failed", "error", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(longURL))
}

func main() {

	// TODO:
	// Initialize the `logger` variable (just like in the frontend)
	// 1. Create a `slog.NewJSONHandler` (writing to `os.Stdout`).
	// 2. Create a `slog.New` logger using this handler.
	// 3. Add a permanent attribute: .With("service", "backend-app")
	//
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "backend-app")
	//

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis-svc:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	// TODO: Replace log.Printf with slog.Info
	// Old log: log.Printf("INFO: Connecting with Redis on %s", redisAddr)
	//
	logger.Info("Connecting to Redis", "address", redisAddr)
	//

	r := mux.NewRouter()
	r.HandleFunc("/generate", generateHandler).Methods("POST")
	r.HandleFunc("/resolve/{shortlink}", resolveHandler).Methods("GET")

	// TODO: Replace log.Printf with slog.Info
	// Old log: log.Println("INFO: Backend-Service starting on Port 8081")
	//
	logger.Info("Backend service starting", "port", 8081)
	//
	http.ListenAndServe(":8081", r)
}
