package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"poc-gin/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// adminDatabaseConfig connects to the default database on the same server as
// the test suite's Postgres, so a scratch database can be created and
// dropped for each test without touching the shared "ecommerce" database.
func adminDatabaseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:     testEnv("DB_HOST", "127.0.0.1"),
		Port:     testEnv("DB_PORT", "5432"),
		User:     testEnv("DB_USER", "golang"),
		Password: testEnv("DB_PASSWORD", "golang"),
		Name:     "postgres",
	}
}

func openAdminConnection(t *testing.T) *gorm.DB {
	t.Helper()

	cfg := adminDatabaseConfig()
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("postgres is required in CI: %v", err)
		}
		t.Skipf("postgres not available: %v", err)
	}

	return db
}

// createScratchDatabase provisions an empty database for a single test and
// registers its cleanup, so migrations can be exercised against a truly
// fresh schema without racing the shared integration test database.
func createScratchDatabase(t *testing.T) string {
	t.Helper()

	admin := openAdminConnection(t)
	sqlDB, err := admin.DB()
	if err != nil {
		t.Fatalf("failed to get admin sql.DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	name := fmt.Sprintf("migrate_test_%d", time.Now().UnixNano())
	if err := admin.Exec(fmt.Sprintf(`CREATE DATABASE %q OWNER %s`, name, adminDatabaseConfig().User)).Error; err != nil {
		t.Fatalf("failed to create scratch database: %v", err)
	}

	t.Cleanup(func() {
		// Terminate any lingering connection (e.g. from the migrator) before
		// dropping, otherwise Postgres refuses with "database is being accessed".
		_ = admin.Exec(fmt.Sprintf(
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s`,
			pqQuoteLiteral(name),
		)).Error
		_ = admin.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, name)).Error
	})

	return name
}

func pqQuoteLiteral(value string) string {
	return "'" + value + "'"
}

func TestMigrateAppliesSchemaToFreshDatabase(t *testing.T) {
	dbName := createScratchDatabase(t)
	adminCfg := adminDatabaseConfig()
	cfg := &config.DatabaseConfig{
		Host:     adminCfg.Host,
		Port:     adminCfg.Port,
		User:     adminCfg.User,
		Password: adminCfg.Password,
		Name:     dbName,
	}

	if err := Migrate(cfg); err != nil {
		t.Fatalf("expected migration success, got %v", err)
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to connect to migrated database: %v", err)
	}
	defer func() { _ = db.Close() }()

	expectedTables := []string{"categories", "users", "products", "promotions", "product_promotions", "orders", "order_items"}
	for _, table := range expectedTables {
		var exists bool
		if err := db.DB.Raw(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)`,
			table,
		).Scan(&exists).Error; err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist after migration", table)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	dbName := createScratchDatabase(t)
	adminCfg := adminDatabaseConfig()
	cfg := &config.DatabaseConfig{
		Host:     adminCfg.Host,
		Port:     adminCfg.Port,
		User:     adminCfg.User,
		Password: adminCfg.Password,
		Name:     dbName,
	}

	if err := Migrate(cfg); err != nil {
		t.Fatalf("expected first migration success, got %v", err)
	}
	if err := Migrate(cfg); err != nil {
		t.Fatalf("expected second migration run to be a no-op, got %v", err)
	}
}
