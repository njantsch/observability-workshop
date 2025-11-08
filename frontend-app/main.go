package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var backendServiceURL string

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

	log.Println("INFO: Frontend-Service starting on Port 8080")
	http.ListenAndServe(":8080", r)
}
