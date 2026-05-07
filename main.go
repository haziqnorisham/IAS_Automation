package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	influxdb_utils "ias/automation/db/influxdb"
	redis_utils "ias/automation/db/redis"
	ingest_http "ias/automation/ingest/http"
	ingest_mqtt "ias/automation/ingest/mqtt"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	initLogger()
	loadEnv()
	rdb := initRedis()
	defer rdb.Close()
	initInfluxDB()
	buildSTICacheIfEnabled(rdb)
	setupHCBackendIfEnabled()
	startHTTPServerIfEnabled(rdb)
	startMQTTIfEnabled()
	waitForShutdown()
}

func initLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	slog.Info("Application is starting")
}

func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("Failed to load .env file", "error", err)
		os.Exit(1)
	}
	slog.Info("Environment variables loaded", "process", "main")
}

func initRedis() *redis.Client {
	slog.Info("Initializing Redis connection", "process", "main")
	rdb := redis_utils.NewRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Redis ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Redis connection established", "process", "main")
	return rdb
}

func initInfluxDB() {
	slog.Info("Initializing InfluxDB connection", "process", "main")
	if err := influxdb_utils.InitInfluxService(os.Getenv("INFLUXDB_ORG")); err != nil {
		slog.Error("Failed to initialize InfluxDB", "error", err)
		os.Exit(1)
	}
	slog.Info("InfluxDB connection established", "process", "main")
}

func buildSTICacheIfEnabled(rdb *redis.Client) {
	if os.Getenv("STI_AUTOMATION_ENABLE") != "true" {
		return
	}
	slog.Info("Building InfluxDB Cache for STI", "process", "main")
	ingest_http.BuildSTICache(rdb)
	slog.Info("InfluxDB Cache built successfully", "process", "main")
}

func setupHCBackendIfEnabled() {
	if os.Getenv("IAS_HC_BACKEND_ENABLE") != "true" {
		slog.Info("IAS HC Backend Server is disabled", "process", "main")
		return
	}
	slog.Info("IAS HC Backend Server is enabled", "process", "main")
	if err := ingest_http.SetupHcSchema(); err != nil {
		slog.Error("Failed to setup HC schema", "error", err)
		os.Exit(1)
	}
}

func startHTTPServerIfEnabled(rdb *redis.Client) {
	if os.Getenv("HTTP_SERVER_AUTOSTART") != "true" {
		return
	}
	slog.Info("Starting HTTP server", "process", "main")
	ingest_http.SetupRoutes(rdb)
	ingest_http.StartServer()
}

func startMQTTIfEnabled() {
	if os.Getenv("MQTT_ENABLED") != "true" {
		return
	}
	slog.Info("MQTT sensor monitoring is enabled, connecting to broker", "process", "main")

	var mqttHandlers []ingest_mqtt.MessageHandler
	if os.Getenv("IAS_HC_BACKEND_ENABLE") == "true" {
		slog.Info("HC raw ingest handler attached to MQTT subscription", "process", "main")
		mqttHandlers = append(mqttHandlers, ingest_mqtt.HcDbHandler())
	}

	if err := ingest_mqtt.ConnectAndSubscribe(mqttHandlers...); err != nil {
		slog.Error("Failed to start MQTT client", "error", err, "process", "main")
		os.Exit(1)
	}
	slog.Info("MQTT client connected and subscribed", "process", "main")
}

func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("Received signal, shutting down", "signal", sig.String())

	if ingest_http.IsRunning {
		slog.Info("Shutting down HTTP server", "process", "main")
		ingest_http.StopServer()
	}

	if ingest_mqtt.IsRunning() {
		slog.Info("Shutting down MQTT client", "process", "main")
		ingest_mqtt.StopClient()
	}
}
