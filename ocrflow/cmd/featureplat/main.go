package main

import (
	"log"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/app"
	"github.com/joho/godotenv"
)

//go:generate swag init -g main.go -d .,../../internal/api/common,../../internal/api/featureplat,../../internal/model,../../internal/model/featureplat --parseInternal --instanceName featureplat -o ../../internal/docs/featureplat

// @title          	OCR Flow API
// @version         1.0
// @description     HTTP API for the OCR pipeline.
// @BasePath        /api/v1/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide your Bearer token in the format: Bearer {token}
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	err = godotenv.Load(".env_private")
	if err != nil {
		log.Printf("failed to load the optional .env_private file, continuing without it: %v", err)
	}
	a, err := app.NewFeaturePlatform()
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	defer a.Close()
	addr := a.Env.HTTPAddr
	log.Printf("HTTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, a.Router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
