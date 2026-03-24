package database

import (
	"context"
	"fmt"
	"poc-gin/config"
	"poc-gin/pkg/logger"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func New(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, formatConnectionError(cfg, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")
	return &Database{db}, nil
}

func formatConnectionError(cfg *config.DatabaseConfig, err error) error {
	if hint := connectionHint(cfg, err); hint != "" {
		return fmt.Errorf("failed to connect to database: %w; hint: %s", err, hint)
	}

	return fmt.Errorf("failed to connect to database: %w", err)
}

func connectionHint(cfg *config.DatabaseConfig, err error) string {
	message := strings.ToLower(err.Error())

	switch {
	case strings.Contains(message, "sqlstate 28p01"),
		strings.Contains(message, "password authentication failed"):
		return "check DB_USER and DB_PASSWORD. If you are using the local docker-compose and the Postgres volume was already initialized with older credentials, recreate it from go-api with `docker compose down -v` then `docker compose up -d db` (this deletes local database data)"
	case strings.Contains(message, "role") &&
		strings.Contains(message, "does not exist"):
		return "check DB_USER and DB_NAME. If you changed them after the local Postgres volume was first created, recreate the volume from go-api with `docker compose down -v` then `docker compose up -d db`"
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "no connection could be made"),
		strings.Contains(message, "actively refused it"):
		return fmt.Sprintf("make sure Postgres is running on %s:%s. For local development, start it from go-api with `docker compose up -d db`", cfg.Host, cfg.Port)
	case strings.Contains(message, "database") &&
		strings.Contains(message, "does not exist"):
		return "check DB_NAME. If you changed it after the local Postgres volume was first created, recreate the volume from go-api with `docker compose down -v` then `docker compose up -d db`"
	default:
		return ""
	}
}

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
