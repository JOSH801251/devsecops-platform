package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/josh801251/devsecops-platform/services/api/internal/jobs"
)

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_CONFIG", "configured-value")

	if got := envOr("TEST_CONFIG", "default-value"); got != "configured-value" {
		t.Fatalf("expected configured value, got %q", got)
	}

	t.Setenv("TEST_CONFIG", "")

	if got := envOr("TEST_CONFIG", "default-value"); got != "default-value" {
		t.Fatalf("expected default value, got %q", got)
	}
}

func TestResolveDSN_PrefersFile(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://environment")

	path := filepath.Join(t.TempDir(), "database-url")
	if err := os.WriteFile(path, []byte("postgres://file\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DATABASE_URL_FILE", path)

	got, err := resolveDSN()
	if err != nil {
		t.Fatal(err)
	}

	if got != "postgres://file" {
		t.Fatalf("expected file value, got %q", got)
	}
}

func TestResolveDSN_UsesEnvironmentFallback(t *testing.T) {
	t.Setenv("DATABASE_URL_FILE", "")
	t.Setenv("DATABASE_URL", "postgres://environment")

	got, err := resolveDSN()
	if err != nil {
		t.Fatal(err)
	}

	if got != "postgres://environment" {
		t.Fatalf("unexpected DSN: %q", got)
	}
}

func TestResolveDSN_ReturnsErrorWhenMissing(t *testing.T) {
	t.Setenv("DATABASE_URL_FILE", "")
	t.Setenv("DATABASE_URL", "")

	_, err := resolveDSN()
	if err == nil {
		t.Fatal("expected error when database configuration is missing")
	}
}

func TestResolveDSN_ReturnsErrorWhenFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-database-url")

	t.Setenv("DATABASE_URL_FILE", path)
	t.Setenv("DATABASE_URL", "postgres://environment")

	_, err := resolveDSN()
	if err == nil {
		t.Fatal("expected error when configured database file does not exist")
	}
}

func TestRepoPinger_PingCheck(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := jobs.NewPostgresRepo(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	pinger := repoPinger{repo: repo}

	if err := pinger.PingCheck(); err != nil {
		t.Fatalf("database health check failed: %v", err)
	}
}
