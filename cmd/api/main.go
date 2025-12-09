// @title Subscriptions API
// @version 1.0
// @description REST API for managing online subscriptions
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SoulStalker/subscribes_api/internal/config"
	"github.com/SoulStalker/subscribes_api/internal/handler"
	"github.com/SoulStalker/subscribes_api/internal/repository/db"
	"github.com/SoulStalker/subscribes_api/internal/repository/postgres"
	"github.com/SoulStalker/subscribes_api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfgPath := "./config/config.yaml"
	cfg := config.MustLoad(cfgPath)

	logger := initLogger(cfg.Log)
	defer logger.Sync()

	dbPool, err := initDB(cfg.DB)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	if err := db.RunMigrations(cfg.DB.DSN()); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	repo := postgres.NewSubscriptionRepository(dbPool, logger)
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewHandler(svc, logger)
	r := h.InitRoutes(cfg.Server.Mode)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger(cfg config.LogConfig) *zap.Logger {
	var zapCfg zap.Config

	if cfg.Encoding == "json" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	zapCfg.Level = zap.NewAtomicLevelAt(parseLogLevel(cfg.Level))

	logger, _ := zapCfg.Build()
	return logger
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func initDB(cfg config.DBConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
