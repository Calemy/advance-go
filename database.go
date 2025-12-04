package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB(dsn string) {
	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to parse db config: %v", err)
	}

	// Tune for performance:
	cfg.MaxConns = 20
	cfg.MinConns = 5
	cfg.MaxConnIdleTime = time.Minute * 5
	cfg.HealthCheckPeriod = time.Minute

	DB, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
}
