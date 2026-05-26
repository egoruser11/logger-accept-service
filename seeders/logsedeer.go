package seeders

import (
	"Realtime_Log_Aggregator/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func GenerateLogs() {
	time.Sleep(2 * time.Second)
	services := []string{
		"user-service",
		"auth-service",
		"order-service",
		"payment-service",
		"notification-service",
	}

	logLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	messages := []string{
		"User login successful",
		"Order created successfully",
		"Payment processed",
		"Failed to connect to database",
		"Cache cleared",
		"New user registered",
		"Password reset requested",
		"Email notification sent",
		"API rate limit exceeded",
		"Service health check passed",
		"Transaction rolled back",
		"File upload completed",
		"Invalid authentication token",
		"Request timeout",
		"Database migration completed",
	}

	log.Println("🚀 Starting log generation: 200 logs with 0.5s interval")

	for i := 1; i <= 200; i++ {
		logEntry := models.LogRequest{
			ServiceName: services[rand.Intn(len(services))],
			LogLevel:    logLevels[rand.Intn(len(logLevels))],
			Message:     messages[rand.Intn(len(messages))],
			Timestamp:   time.Now().UnixMilli(),
		}

		if err := sendLog(logEntry); err != nil {
			log.Printf("❌ Error sending log #%d: %v", i, err)

			// Выводим что именно отправляем для отладки
			debugJSON, _ := json.Marshal(logEntry)
			log.Printf("   Debug - sent JSON: %s", string(debugJSON))
		} else {
			log.Printf("✅ Log #%d sent: [%s] %s - %s",
				i,
				logEntry.LogLevel,
				logEntry.ServiceName,
				logEntry.Message,
			)
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Println("✅ Log generation completed: 200 logs sent")
}

func sendLog(logEntry models.LogRequest) error {
	url := "http://localhost:8080/api/logs"

	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
