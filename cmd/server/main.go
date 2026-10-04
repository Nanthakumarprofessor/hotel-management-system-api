package main

import (
	"flag"
	"log"
	"os"

	"hotel-updated/pkg/server"

	"go.uber.org/zap"
)

func main() {
	// Parse command line arguments
	env := flag.String("env", "dev", "Environment name (dev, qc, uat, prod)")
	flag.Parse()

	// Initialize application
	app, err := server.InitializeApp(*env)
	if err != nil {
		log.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}

	// Flush logger on exit
	defer app.Logger.Sync()

	// Run server
	if err := server.RunServer(app); err != nil {
		app.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}
