package main

import (
	"log"
	"os"

	"kokotoba-backend/internal/server"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	if err := server.Run(addr); err != nil {
		log.Fatal(err)
	}
}
