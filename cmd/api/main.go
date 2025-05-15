package main

import (
	"os"

	"mms_api/internal/bootstrap"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")

	// Initialize and start application
	app := bootstrap.NewApp(port)
	if err := app.Start(); err != nil {
		// A logger não é necessária aqui pois o erro já será logado internamente
		os.Exit(1)
	}
}
