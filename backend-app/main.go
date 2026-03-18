package main

import (
	"context"
	"fmt"
	"io"
	"log"
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
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: couldn't read request body: %v", err)
		//
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortLink := fmt.Sprintf("id%d", rand.Intn(10000))

	err = rdb.Set(ctx, shortLink, string(longURL), time.Hour*24).Err()
	if err != nil {
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: Redis Set failed: %v", err)
		//
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Replace the old log line
	//
	log.Printf("INFO: Mapping created: %s -> %s", shortLink, string(longURL))
	//
	w.Write([]byte(shortLink))
}

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortLink := vars["shortlink"]

	longURL, err := rdb.Get(ctx, shortLink).Result()
	if err == redis.Nil {
		// TODO: Replace the old log line
		//
		log.Printf("WARN: Link not found: %s", shortLink)
		//
		http.NotFound(w, r)
		return
	} else if err != nil {
		// TODO: Replace the old log line
		//
		log.Printf("ERROR: Redis Get failed: %v", err)
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
	// Code goes here ...
	//

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis-svc:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	// TODO: Replace the old log line
	//
	log.Printf("INFO: Connecting with Redis on %s", redisAddr)
	//

	r := mux.NewRouter()
	r.HandleFunc("/generate", generateHandler).Methods("POST")
	r.HandleFunc("/resolve/{shortlink}", resolveHandler).Methods("GET")

	// TODO: Replace the old log line
	//
	log.Println("INFO: Backend-Service starting on Port 8081")
	//
	http.ListenAndServe(":8081", r)
}
