package main

import (
	"context"
	"fmt"
	redis_utils "ias/automation/db/redis"
	ingest_http "ias/automation/ingest/http"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

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

	// Start the HTTP server if autostart is enabled
	if often := os.Getenv("HTTP_SERVER_AUTOSTART"); often == "true" {
		ingest_http.SetupRoutes()
		ingest_http.StartServer()
	}

	for {
		fmt.Print("\nCommand (start/stop/restart/quit): ")
		var cmd string
		fmt.Scanln(&cmd)

		switch cmd {
		case "start":
			if !ingest_http.IsRunning {
				ingest_http.StartServer()
			} else {
				fmt.Println("Already running")
			}
		case "stop":
			if ingest_http.IsRunning {
				ingest_http.StopServer()
			} else {
				fmt.Println("Not running")
			}
		case "restart":
			ingest_http.StopServer()
			time.Sleep(1 * time.Second)
			ingest_http.StartServer()
		case "quit":
			ingest_http.StopServer()
			return
		}
	}
}
