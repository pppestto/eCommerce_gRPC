package main

import (
	"os"

	"github.com/pppestto/ecommerce-grpc/services/common/logger"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/app"
)

func main() {
	logger.Init("product-service", os.Getenv("ENV") == "production")

	application, err := app.New()
	if err != nil {
		logger.L().Error("failed to create app", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		logger.L().Error("failed to run app", "error", err)
		os.Exit(1)
	}
}
