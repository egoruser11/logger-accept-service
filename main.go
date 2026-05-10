package main

import (
	"Realtime_Log_Aggregator/internal/handlers"
	"Realtime_Log_Aggregator/internal/hub"
	"Realtime_Log_Aggregator/seeders"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection error: %v", err)
	}

	log.Println("redis connected")

	hub := hub.NewHub()
	go hub.Run()

	api := e.Group("/api")

	api.GET("/health", handlers.HealthHandler)
	api.POST("/logs", handlers.IngestLogsHandler(rdb, hub))
	api.GET("/logs/recent", handlers.RecentLogsHandler(rdb))
	api.GET("/ws", handlers.WebsocketHandler(hub))

	go seeders.GenerateLogs()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Println("server started on :8080")

	interruptSignalCh := make(chan os.Signal, 1)
	signal.Notify(interruptSignalCh, os.Interrupt)

	select {
	case <-interruptSignalCh:
		log.Println("Received interrupt signal, shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}

		hub.Shutdown()

		if err := closeRedis(rdb); err != nil {
			log.Printf("Redis close error: %v", err)
		}

		log.Println("Server stopped")
	}
}

func closeRedis(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.BgSave(ctx).Err(); err != nil {
		log.Printf("Redis BgSave failed: %v", err)
	}

	if err := rdb.Close(); err != nil {
		return fmt.Errorf("redis close error: %w", err)
	}

	log.Println("Redis connection closed")
	return nil
}
