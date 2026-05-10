package main

import (
	"Realtime_Log_Aggregator/internal/handlers"
	"Realtime_Log_Aggregator/internal/hub"
	"Realtime_Log_Aggregator/seeders"
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"sync"
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

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("server started on :8080")
	wg := &sync.WaitGroup{}
	go seeders.GenerateLogs(wg)
	wg.Wait()
	if err := e.StartServer(server); err != nil {
		log.Fatal(err)
	}
}
