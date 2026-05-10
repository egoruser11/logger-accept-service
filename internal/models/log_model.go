package models

type Log struct {
	ServiceName string
	LogLevel    string
	Message     string
	Timestamp   int64
}

type LogInputRequest struct {
	ServiceName string `json:"service_name"`
	LogLevel    string `json:"log_level"`
	Message     string `json:"message"`
	Timestamp   int64  `json:"timestamp"`
}
