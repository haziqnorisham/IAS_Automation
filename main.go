package main

import (
	"context"
	"fmt"
	influxdb_utils "ias/automation/db/influxdb"
	redis_utils "ias/automation/db/redis"
	ingest_http "ias/automation/ingest/http"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Create JSON logger writing to stdout (Docker captures this)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Change to LevelDebug for development
	}))

	// Replace global logger (optional but convenient)
	slog.SetDefault(logger)

	// Use the logger
	slog.Info("Application is starting")

	slog.Info("Loading environment variables from .env file", "process", "main")
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	slog.Info("Environment variables loaded successfully", "process", "main")

	slog.Info("Initializing Redis connection", "process", "main")
	// Establish Connection to Redis & test connection.
	rdb := redis_utils.NewRedisClient()
	defer rdb.Close()

	ctx := context.Background()

	err = rdb.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("foo", val)
	slog.Info("Redis connection established and tested successfully", "process", "main")

	slog.Info("Initializing InfluxDB connection", "process", "main")
	// Setup InfluxDB Connection
	influxdb_utils.InitInfluxService(os.Getenv("INFLUXDB_ORG"))
	influxdb_utils.TestQuery()
	slog.Info("InfluxDB connection established and tested successfully", "process", "main")

	slog.Info("Building InfluxDB Cache for STI", "process", "main")
	// Build InfluxDB Cache for STI
	ingest_http.BuildSTICache(rdb)
	slog.Info("InfluxDB Cache for STI built successfully", "process", "main")

	if os.Getenv("IAS_HC_BACKEND_ENABLE") == "true" {
		slog.Info("IAS HC Backend Server is enabled", "process", "main")
		err := ingest_http.SetupHcSchema()
		if err != nil {
			slog.Error("Failed to setup HC schema", "error", err)
			return
		}
	} else {
		slog.Info("IAS HC Backend Server is disabled", "process", "main")
	}

	slog.Info("Starting HTTP server if autostart is enabled", "process", "main")
	// Start the HTTP server if autostart is enabled
	if often := os.Getenv("HTTP_SERVER_AUTOSTART"); often == "true" {
		ingest_http.SetupRoutes(rdb)
		ingest_http.StartServer()
	}
	select {}
}
