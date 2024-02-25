package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

func main() {
	db := sqlx.MustOpen("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()

	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-1"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	snsClient := sns.NewFromConfig(cfg)
	accessLogService := NewAccessLogService(db)
	handler := NewHandler(db, accessLogService, snsClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/verify-access", handler.VerifyAccess)
	mux.HandleFunc("/clear-access-cache", handler.ClearAccessCache)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	go func() {
		slog.Info("Starting server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down server: %v\n", err)
	}
}
