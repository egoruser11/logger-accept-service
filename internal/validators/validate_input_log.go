package validators

import (
	"Realtime_Log_Aggregator/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

func ValidateInputLog(req models.LogRequest) (error, map[string]string) {
	errors := make(map[string]string)

	if strings.TrimSpace(req.ServiceName) == "" {
		errors["serviceName"] = "service name is required"
	} else if len(req.ServiceName) > 100 {
		errors["serviceName"] = "service name must not exceed 100 characters"
	}
	validLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
		"FATAL": true,
	}
	if strings.TrimSpace(req.LogLevel) == "" {
		errors["logLevel"] = "log level is required"
	} else if !validLevels[strings.ToUpper(req.LogLevel)] {
		errors["logLevel"] = "invalid log level, must be one of: DEBUG, INFO, WARN, ERROR, FATAL"
	}

	if strings.TrimSpace(req.Message) == "" {
		errors["message"] = "message is required"
	} else if len(req.Message) > 10000 {
		errors["message"] = "message must not exceed 10000 characters"
	}

	if req.Timestamp <= 0 {
		errors["timestamp"] = "timestamp is required and must be positive"
	} else if req.Timestamp > time.Now().Unix()*1000+86400000 {
		errors["timestamp"] = "timestamp cannot be in the future"
	}

	if len(errors) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed"), errors
	}

	return nil, errors
}
