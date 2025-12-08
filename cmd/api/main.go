package main

import (
	"context"
	"fmt"

	"github.com/SoulStalker/subscribes_api/internal/config"
	"github.com/SoulStalker/subscribes_api/internal/repository/db"
	"github.com/SoulStalker/subscribes_api/internal/repository/postgres"

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

	fmt.Println(repo)
	fmt.Println(cfg.DB.DSN())
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
