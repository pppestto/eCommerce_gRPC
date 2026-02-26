package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pppestto/ecommerce-grpc/services/bff/internal/app"
)

func main() {
	cfg := app.Config{
		Port:        8080,
		UserAddr:    getEnv("USER_SERVICE_ADDR", "127.0.0.1:50051"),
		ProductAddr: getEnv("PRODUCT_SERVICE_ADDR", "127.0.0.1:50052"),
		OrderAddr:   getEnv("ORDER_SERVICE_ADDR", "127.0.0.1:50053"),
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	go func() {
		if err := application.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Stop(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bff stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
