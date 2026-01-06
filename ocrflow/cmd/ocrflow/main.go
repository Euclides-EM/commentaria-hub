package main

import (
	"log"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/app"
	"github.com/joho/godotenv"
)

//go:generate swag init -g main.go -d .,../../internal/ocrflow --parseInternal -o ../../internal/ocrflow/docs

// @title          	OCR Flow API
// @version         1.0
// @description     HTTP API for the OCR pipeline.
// @BasePath        /api/v1/
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	err = godotenv.Load(".env_private")
	if err != nil {
		log.Printf("failed to load the optional .env_private file, continuing without it: %v", err)
	}
	a, err := app.NewHTTPApp()
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
