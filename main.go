package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"job_tracker_be/internal/config"
	"job_tracker_be/internal/controller"
	"job_tracker_be/internal/dao"
	"job_tracker_be/internal/db"
	"job_tracker_be/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer pool.Close()

	jobDAO := dao.NewPgxJobDAO(pool)
	resumeQueueDAO := dao.NewPgxResumeQueueDAO(pool)
	jobService := service.NewJobService(jobDAO)
	resumeQueueService := service.NewResumeQueueService(jobDAO, resumeQueueDAO, cfg.N8NWebhookURL)
	router := controller.NewRouter(jobService, resumeQueueService, cfg.RequestTimeout)

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func runMigrations(databaseURL string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql db for migrations: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping sql db for migrations: %w", err)
	}

	migrationDir, err := filepath.Abs("./migrations")
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, migrationDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
