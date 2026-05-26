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
	"strings"
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

		var log models.LogRequest
		if err := c.Bind(&log); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid body",
			})
		}
		hub.Mutex.RLock()
		if hub.IsShuttingDown {
			hub.Mutex.RUnlock()
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"error": "server is shutting down",
			})
		}
		hub.Mutex.RUnlock()
		err, errorsMap := validators.ValidateInputLog(log)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorsMap)
		}

		ctx := context.Background()
		logJSON, _ := json.Marshal(log)

		rdb.LPush(ctx, "logs:recent", logJSON)
		rdb.LTrim(ctx, "logs:recent", 0, 999)
		hub.PendingLogs.Add(1)
		hub.Broadcast <- log
		return c.JSON(http.StatusOK, map[string]string{
			"message": "log accepted",
		})
	}
}

func RecentLogsHandler(rdb *redis.Client) echo.HandlerFunc {
	return logsSearchHandler(rdb)
}

func SearchLogsHandler(rdb *redis.Client) echo.HandlerFunc {
	return logsSearchHandler(rdb)
}

func logsSearchHandler(rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req models.LogSearchRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid query params",
			})
		}

		if req.From > 0 && req.To > 0 && req.From > req.To {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "from must be less than or equal to to",
			})
		}

		ctx := context.Background()
		rawLogs, err := rdb.LRange(ctx, "logs:recent", 0, 999).Result()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		logs := make([]models.LogRequest, 0, len(rawLogs))
		for _, rawLog := range rawLogs {
			var logEntry models.LogRequest
			if err := json.Unmarshal([]byte(rawLog), &logEntry); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to parse log",
				})
			}

			logs = append(logs, logEntry)
		}

		return c.JSON(http.StatusOK, filterLogs(logs, req))
	}
}

func filterLogs(logs []models.LogRequest, req models.LogSearchRequest) []models.LogRequest {
	filtered := make([]models.LogRequest, 0, len(logs))

	for _, logEntry := range logs {
		if !matchesLog(logEntry, req) {
			continue
		}

		filtered = append(filtered, logEntry)
	}

	return filtered
}

func matchesLog(logEntry models.LogRequest, req models.LogSearchRequest) bool {
	if req.ServiceName != "" && !strings.EqualFold(logEntry.ServiceName, req.ServiceName) {
		return false
	}

	if req.LogLevel != "" && !strings.EqualFold(logEntry.LogLevel, req.LogLevel) {
		return false
	}

	if req.Message != "" && !containsFold(logEntry.Message, req.Message) {
		return false
	}

	if req.Q != "" && !containsFold(logEntry.Message, req.Q) {
		return false
	}

	if req.From > 0 && logEntry.Timestamp < req.From {
		return false
	}

	if req.To > 0 && logEntry.Timestamp > req.To {
		return false
	}

	return true
}

func containsFold(value string, search string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(search))
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
