package main

import (
	"flag"

	"go-playground/internal/app"
)

func main() {
	// Allow running multiple nodes locally by passing a port per process.
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	app.Run(*port)
}
