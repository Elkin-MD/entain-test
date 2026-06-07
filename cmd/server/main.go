package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"entaintest/internal/config"
	"entaintest/internal/controller"
	"entaintest/internal/repository"
	"entaintest/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := connectWithRetry(ctx, cfg.DSN())
	if err != nil {
		return err
	}

	defer pool.Close()

	repo := repository.NewWalletRepository(pool)
	svc := service.New(repo)
	ctrl := controller.New(svc)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           ctrl.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.HTTPPort)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

func connectWithRetry(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	const (
		attempts = 30
		delay    = time.Second
	)

	var lastErr error

	for i := 0; i < attempts; i++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}

			pool.Close()
		}

		lastErr = err
		log.Printf("waiting for database (attempt %d/%d): %v", i+1, attempts, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, lastErr
}
