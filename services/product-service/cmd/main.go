package main

import (
	"log"

	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
