package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ldryt/mcpulse/provider"
	"github.com/redis/go-redis/v9"
)

type CallbackRequest struct {
	Status string `json:"status"` // "started", "ping", "timeout", "crash"
}

type Handler struct {
	db       *sql.DB
	rdb      *redis.Client
	provider provider.CloudProvider
	worker   *Worker
}

func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		http.Error(w, "Missing Token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	serverUUID, err := h.rdb.Get(ctx, "server_token:"+token).Result()
	if err == redis.Nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	var req CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Printf("Server '%s' signals '%s'", serverUUID, req.Status)
	switch req.Status {
	case "started", "ping":
		pulseKey := "server_pulse:" + serverUUID
		err := h.rdb.Set(ctx, pulseKey, "alive", 130*time.Second).Err()
		if err != nil {
			log.Printf("Failed to update pulse for '%s': %v", serverUUID, err)
		}

		if req.Status == "started" {
			h.markServerRunning(ctx, serverUUID)
		}
	case "timeout", "crash":
		h.handleExplicitShutdown(ctx, serverUUID, req.Status)
	default:
		http.Error(w, "Unknown Status", http.StatusBadRequest)
	}
}

func (h *Handler) markServerRunning(ctx context.Context, uuid string) {
	_, err := h.db.ExecContext(ctx, `
		UPDATE servers 
		SET status = 'running', started_at = NOW(), last_billed_at = NOW()
		WHERE uuid = $1 AND status != 'running'
	`, uuid)
	if err != nil {
		log.Printf("Failed to mark running '%s': %v", uuid, err)
	}
}

func (h *Handler) handleExplicitShutdown(ctx context.Context, uuid, reason string) {
	var id int
	var cloudID string
	err := h.db.QueryRowContext(ctx, `SELECT id, cloud_id FROM servers WHERE uuid = $1`, uuid).Scan(&id, &cloudID)
	if err != nil {
		log.Printf("Failed to find server '%s' for shutdown: %v", uuid, err)
		return
	}

	finalStatus := "stopped"
	if reason == "crash" {
		finalStatus = "crashed"
	}

	h.worker.terminateServer(ctx, id, cloudID, finalStatus)

	h.rdb.Del(ctx, "server_pulse:"+uuid)
}
