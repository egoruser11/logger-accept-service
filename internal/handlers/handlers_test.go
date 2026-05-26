package handlers

import (
	"Realtime_Log_Aggregator/internal/models"
	"testing"
)

func TestFilterLogsByServiceLevelTextAndTime(t *testing.T) {
	logs := []models.LogRequest{
		{
			ServiceName: "payment-service",
			LogLevel:    "ERROR",
			Message:     "Request timeout",
			Timestamp:   1000,
		},
		{
			ServiceName: "auth-service",
			LogLevel:    "INFO",
			Message:     "User login successful",
			Timestamp:   2000,
		},
		{
			ServiceName: "payment-service",
			LogLevel:    "WARN",
			Message:     "Retrying payment",
			Timestamp:   3000,
		},
	}

	filtered := filterLogs(logs, models.LogSearchRequest{
		ServiceName: "payment-service",
		LogLevel:    "error",
		Q:           "TIMEOUT",
		From:        500,
		To:          1500,
	})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 log, got %d", len(filtered))
	}

	if filtered[0].ServiceName != "payment-service" {
		t.Fatalf("expected payment-service, got %s", filtered[0].ServiceName)
	}
}

func TestFilterLogsReturnsAllLogsWithoutFilters(t *testing.T) {
	logs := []models.LogRequest{
		{ServiceName: "payment-service", LogLevel: "ERROR", Message: "Request timeout", Timestamp: 1000},
		{ServiceName: "auth-service", LogLevel: "INFO", Message: "User login successful", Timestamp: 2000},
	}

	filtered := filterLogs(logs, models.LogSearchRequest{})

	if len(filtered) != len(logs) {
		t.Fatalf("expected %d logs, got %d", len(logs), len(filtered))
	}
}
