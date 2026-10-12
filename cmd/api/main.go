package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"libratrack/internal/app"
	"libratrack/internal/repository"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	migrationsPath := getenv("MIGRATIONS_PATH", "migrations")
	addr := ":" + getenv("PORT", "8080")

	// The schema is created only by SQL migrations, applied on every start.
	if err := repository.Migrate(dsn, migrationsPath); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	db, err := repository.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           app.NewRouter(db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown on Ctrl+C / SIGTERM: finish in-flight requests first.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
