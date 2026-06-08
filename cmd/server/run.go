package server

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
	"entaintest/internal/infrastructure"
	"entaintest/internal/repository"
	"entaintest/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run starts the HTTP server and blocks until shutdown.
func Run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := connect(ctx, cfg.DSN())
	if err != nil {
		return err
	}

	defer pool.Close()

	transactor := repository.NewTransactor(pool)
	userRepository := repository.NewUserRepository(pool)
	transactionRepository := repository.NewTransactionRepository(pool)

	walletController := controller.New(
		service.NewTransactionService(transactor, userRepository, transactionRepository),
		service.NewBalanceService(userRepository),
	)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           infrastructure.NewRouter(walletController),
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

// connect dials Postgres, retrying so the server can start alongside the
// database under docker compose.
func connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	const (
		attempts = 30
		delay    = time.Second
	)

	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}

			pool.Close()
		}

		lastErr = err
		log.Printf("waiting for database (attempt %d/%d): %v", attempt, attempts, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, lastErr
}
