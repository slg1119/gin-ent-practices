package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-start/src/internal/bootstrap"
	"gin-start/src/internal/database"
)

func main() {
	if err := run(); err != nil {
		slog.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startupCtx, cancelStartup := context.WithTimeout(ctx, 10*time.Second)
	client, err := database.Open(startupCtx, env("DATABASE_DSN", "file:app.db?_fk=1&_busy_timeout=5000"))
	cancelStartup()
	if err != nil {
		return err
	}
	defer client.Close()

	application := bootstrap.NewApplication(client)

	server := &http.Server{
		Addr:              env("HTTP_ADDR", "127.0.0.1:8080"),
		Handler:           application.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting API", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	case <-ctx.Done():
		// 진행 중인 요청이 끝난 다음 DB 연결이 닫히도록 기다린다.
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return fmt.Errorf("shutdown HTTP: %w", err)
		}
	}
	return nil
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
