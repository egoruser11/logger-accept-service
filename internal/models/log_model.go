package models

type LogRequest struct {
	ServiceName string `json:"service_name" query:"service_name"`
	LogLevel    string `json:"log_level" query:"log_level"`
	Message     string `json:"message" query:"message"`
	Timestamp   int64  `json:"timestamp" query:"timestamp"`
}

type LogSearchRequest struct {
	ServiceName string `query:"service_name" json:"service_name"`
	LogLevel    string `query:"log_level" json:"log_level"`
	Message     string `query:"message" json:"message"`
	Q           string `query:"q" json:"q"`
	From        int64  `query:"from" json:"from"`
	To          int64  `query:"to" json:"to"`
}
