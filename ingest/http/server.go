package http

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	redis_lib "github.com/redis/go-redis/v9"
)

var IsRunning = false
var currentServer *http.Server

func SetupRoutes(rdb *redis_lib.Client) {
	http.HandleFunc("/", homeHandler)

	if os.Getenv("IAS_HC_BACKEND_ENABLE") == "true" {
		http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			time.Sleep(0 * time.Second)
			w.Write([]byte("{\"status\": \"OK\", \"delay\": \"0s\"}"))
		})
		http.HandleFunc("/api/get_all_devices", func(w http.ResponseWriter, r *http.Request) {
			GetAllDevices(w, r)
		})
		http.HandleFunc("/api/get_raw_ingest", func(w http.ResponseWriter, r *http.Request) {
			GetRawIngest(w, r)
		})
	}
	http.HandleFunc("/GET_ALL_TREE_SENSOR", func(w http.ResponseWriter, r *http.Request) {
		getAllTreeSensorHandler(w, r, rdb)
	})
	http.HandleFunc("/GET_TREE_SENSOR_BATTERY", func(w http.ResponseWriter, r *http.Request) {
		getTreeSensorBatteryHandler(w, r, rdb)
	})
	http.HandleFunc("/GET_TREE_SENSOR_ANGLE", func(w http.ResponseWriter, r *http.Request) {
		getTreeSensorAngleHandler(w, r, rdb)
	})
	http.HandleFunc("/GET_TREE_SENSOR_MAGNITUDE_MIN", func(w http.ResponseWriter, r *http.Request) {
		getTreeSensorMagnitudeMinHandler(w, r, rdb)
	})
	http.HandleFunc("/GET_TREE_SENSOR_MAGNITUDE_MAX", func(w http.ResponseWriter, r *http.Request) {
		getTreeSensorMagnitudeMaxHandler(w, r, rdb)
	})
}
func StartServer() {
	currentServer = &http.Server{Addr: ":" + os.Getenv("HTTP_SERVER_PORT")}
	go func() {
		IsRunning = true
		slog.Info("Server started on "+currentServer.Addr, "address", currentServer.Addr, "process", "main")
		if err := currentServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // Give it a moment to start
}

func StopServer() {
	if currentServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := currentServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server forced shutdown", "error", err)
	}
	IsRunning = false
	slog.Info("HTTP server stopped gracefully")
}
