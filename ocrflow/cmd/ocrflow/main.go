package main

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/app"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

//go:generate swag init -g main.go -d .,../../internal/ocrflow --parseInternal -o ../../internal/ocrflow/docs

// @title          	OCR Flow API
// @version         1.0
// @description     HTTP API for the OCR pipeline.
// @BasePath        /
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	a, err := app.NewHTTPApp()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	defer a.Close()
	addr := a.Env.HTTPAddr
	log.Printf("HTTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, a.Router); err != nil {
		log.Fatalf("server error: %v", err)
	}

	//config, err := config.InitEnv()
	//if err != nil {
	//	log.Fatalf("Error initializing env: %v", err)
	//}
	//
	//if err := db.InitDB(config.DBPath, config.MigrationsDir); err != nil {
	//	log.Fatalf("Error initializing DB: %v", err)
	//}
}

func AvailableEditions() []string {
	return []string{"community", "enterprise"}
}
