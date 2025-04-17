package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// SetupLogger configures the logger
func SetupLogger() {
	// Set the log format to JSON for easier parsing by log aggregators
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	// Set the log level based on the environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Parse the log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
}

// GracefulShutdown handles graceful shutdown of the server
func GracefulShutdown(shutdown func(context.Context) error) {
	// Create a channel to listen for signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-sigs
	logrus.Infof("Received signal %s, shutting down gracefully...", sig)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call the shutdown function
	if err := shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Error during shutdown")
	}

	logrus.Info("Shutdown complete")
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvWithPrefix gets an environment variable with a prefix or returns a default value
func GetEnvWithPrefix(prefix, key, fallback string) string {
	return GetEnv(prefix+"_"+key, fallback)
}

// Define custom types for context keys
type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UsernameKey contextKey = "username"
)
