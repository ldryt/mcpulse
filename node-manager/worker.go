package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/ldryt/mcpulse/provider"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	db       *sql.DB
	rdb      *redis.Client
	provider provider.CloudProvider
}

func NewWorker(db *sql.DB, rdb *redis.Client, cp provider.CloudProvider) *Worker {
	return &Worker{db: db, rdb: rdb, provider: cp}
}

func (w *Worker) StartReaper(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Reaper started. Scanning for zombies and bankruptcy...")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processServers(ctx)
		}
	}
}

func (w *Worker) processServers(ctx context.Context) {
	query := `
		SELECT s.id, s.uuid, s.user_id, s.cloud_id, s.price_per_hour, s.last_billed_at, u.balance 
		FROM servers s
		JOIN users u ON s.user_id = u.id
		WHERE s.status = 'running'
	`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Failed to fetch active servers: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			sID          int
			sUUID        string
			uID          int
			cloudID      string
			pricePerHour float64
			lastBilled   time.Time
			balance      float64
		)
		if err := rows.Scan(&sID, &sUUID, &uID, &cloudID, &pricePerHour, &lastBilled, &balance); err != nil {
			continue
		}

		pulseKey := "server_pulse:" + sUUID
		exists, err := w.rdb.Exists(ctx, pulseKey).Result()
		if err != nil {
			log.Printf("Redis error for ''%s: %v", sUUID, err)
			continue
		}
		if exists == 0 {
			log.Printf("Server '%s' is a ZOMBIE. Killing...", sUUID)
			w.terminateServer(ctx, sID, cloudID, "crashed")
			continue
		}

		if time.Since(lastBilled) >= 1*time.Hour {
			if balance < pricePerHour {
				log.Printf("User '%d' is BANKRUPT. Killing server '%s'...", uID, sUUID)
				w.terminateServer(ctx, sID, cloudID, "stopped_no_credit")
				continue
			}

			err := w.chargeUser(ctx, uID, sID, pricePerHour)
			if err != nil {
				log.Printf("Billing failed for '%s': %v", sUUID, err)
			}
		}
	}
}

func (w *Worker) terminateServer(ctx context.Context, serverID int, cloudID, finalStatus string) {
	err := w.provider.DeleteServer(ctx, cloudID)
	if err != nil {
		log.Printf("Failed to delete server '%s': %v", cloudID, err)
	}

	_, err = w.db.ExecContext(ctx, `
		UPDATE servers 
		SET status = $1, stopped_at = NOW() 
		WHERE id = $2
	`, finalStatus, serverID)

	if err != nil {
		log.Printf("Failed to update status for server '%d': %v", serverID, err)
	}
}

func (w *Worker) chargeUser(ctx context.Context, userID, serverID int, amount float64) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance - $1 WHERE id = $2`, amount, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("user not found")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, server_id, amount, type, created_at)
		VALUES ($1, $2, $3, 'hourly_bill', NOW())
	`, userID, serverID, -amount)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE servers SET last_billed_at = NOW() WHERE id = $1`, serverID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
