package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func getenv(key string) (val string) {
	val = os.Getenv(key)
	if val == "" {
		log.Fatalf("%s not set.", key)
	}
	return val
}

var (
	ListenAddr  = getenv("LISTEN_ADDR")
	BackupsPath = getenv("BACKUPS_PATH")
	RedisAddr   = getenv("REDIS_ADDR")

	rdb *redis.Client
)

func main() {
	rdb = redis.NewClient(&redis.Options{Addr: RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis at '%s': %v", ListenAddr, err)
	}

	if _, err := os.Stat(BackupsPath); err != nil {
		log.Fatalf("Could not stat Backups Path at '%s': %v", BackupsPath, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", handleDownload)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := &http.Server{
		Addr:         ListenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader { // no "Bearer " prefix found
		http.Error(w, "Invalid Token Format", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	serverUUID, err := rdb.Get(ctx, token).Result()

	if err == redis.Nil {
		http.Error(w, "Invalid or Expired Token", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Redis failure: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s.zip", serverUUID)
	backupPath := filepath.Join(BackupsPath, filename)

	fileInfo, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		http.Error(w, "Backup Not Found", http.StatusNotFound)
		return
	}
	if fileInfo.IsDir() {
		http.Error(w, "Invalid Backup File", http.StatusInternalServerError)
		return
	}

	log.Printf("Serving backup '%s' to '%s'", backupPath, r.RemoteAddr)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=server.zip")
	http.ServeFile(w, r, backupPath)
}
