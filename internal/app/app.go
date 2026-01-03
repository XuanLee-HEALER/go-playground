package app

import (
	"log"

	"go-playground/internal/httpserver"
)

// Run wires up the HTTP server and starts it on the given port.
func Run(port int) {
	server := httpserver.New()

	if err := server.Run(port); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
