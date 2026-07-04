package main

import (
	"log"
	"os"

	"github.com/MiaMish/elements-dh/correspondence_ingest/internal/app"
)

func main() {
	log.SetFlags(0)
	if err := app.RunCLI(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
}
