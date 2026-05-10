package handlers

import (
	"Realtime_Log_Aggregator/internal/hub"
	"Realtime_Log_Aggregator/internal/models"
	"Realtime_Log_Aggregator/internal/validators"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func IngestLogsHandler(
	rdb *redis.Client,
	hub *hub.Hub,
) echo.HandlerFunc {
	return func(c echo.Context) error {

		var log models.LogInputRequest
		if err := c.Bind(&log); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid body",
			})
		}

		err, errorsMap := validators.ValidateInputLog(log)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorsMap)
		}

		ctx := context.Background()
		logJSON, _ := json.Marshal(log)

		rdb.LPush(ctx, "logs:recent", logJSON)
		rdb.LTrim(ctx, "logs:recent", 0, 999)

		hub.Broadcast <- log

		return c.JSON(http.StatusOK, map[string]string{
			"message": "log accepted",
		})
	}
}

func RecentLogsHandler(rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()
		logs, err := rdb.LRange(ctx, "logs:recent", 0, 999).Result()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}
		return c.JSON(http.StatusOK, logs)
	}
}

func WebsocketHandler(hub *hub.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {

		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)

		if err != nil {
			return err
		}
		hub.Register <- conn

		go handleClientConnection(hub, conn)
		return nil
	}
}

func handleClientConnection(hub *hub.Hub, conn *websocket.Conn) {

	defer func() {
		hub.Unregister <- conn
		conn.Close()
	}()

	conn.SetReadLimit(512)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
