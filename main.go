package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"kubernetes-api/internal/api"
	"kubernetes-api/internal/auth"
	"kubernetes-api/internal/database"
	"kubernetes-api/pkg/utils"

	"github.com/sirupsen/logrus"
)

// Application version
const appVersion = "1.0.0"

func main() {
	// Setup logging
	utils.SetupLogger()
	logrus.Info("Starting Kubernetes API service...")
	logrus.Infof("Version: %s", appVersion)

	// Initialize authentication
	if err := auth.InitAuth(); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize authentication")
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer database.CloseDB()

	// Setup HTTP server
	port := utils.GetEnv("PORT", "8080")
	router := api.SetupRouter()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		logrus.Infof("HTTP server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("HTTP server failed")
		}
	}()

	// Handle graceful shutdown
	utils.GracefulShutdown(func(ctx context.Context) error {
		logrus.Info("Shutting down HTTP server...")
		// Shutdown the server
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	})
}
