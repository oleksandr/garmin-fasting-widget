package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	flagPort := flag.String("port", "", "HTTP listen port (or FASTING_PORT env)")
	flagDB := flag.String("db", "", "SQLite database path (or FASTING_DB_PATH env)")
	flagToken := flag.String("token", "", "Optional bearer auth token (or FASTING_API_TOKEN env)")
	flag.Parse()

	port := cmp.Or(*flagPort, os.Getenv("FASTING_PORT"), os.Getenv("PORT"), "8080")
	dbPath := cmp.Or(*flagDB, os.Getenv("FASTING_DB_PATH"), "fasting.db")
	authToken := cmp.Or(*flagToken, os.Getenv("FASTING_API_TOKEN"), "")

	store, err := NewDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize sqlite database: %v", err)
	}
	defer store.Close()

	srv := NewServer(store, authToken)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Garmin Fasting Sync Server listening on :%s (db: %s)", port, dbPath)
		if authToken != "" {
			log.Printf("Authentication enabled (bearer token required)")
		} else {
			log.Printf("Authentication disabled (open access)")
		}

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}

	log.Println("Server stopped cleanly")
}
