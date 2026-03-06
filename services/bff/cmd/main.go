package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/app"
	"github.com/pppestto/ecommerce-grpc/services/common/logger"
)

func main() {
	logger.Init("bff", os.Getenv("ENV") == "production")

	cfg := app.Config{
		Port:        8080,
		UserAddr:    getEnv("USER_SERVICE_ADDR", "127.0.0.1:50051"),
		ProductAddr: getEnv("PRODUCT_SERVICE_ADDR", "127.0.0.1:50052"),
		OrderAddr:   getEnv("ORDER_SERVICE_ADDR", "127.0.0.1:50053"),
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
	}

	application, err := app.New(cfg)
	if err != nil {
		logger.L().Error("failed to create app", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := application.Run(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Stop(ctx); err != nil {
		logger.L().Error("shutdown error", "error", err)
	}
	logger.L().Info("bff stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
