# Realtime Log Aggregator

Realtime Log Aggregator is a small Go service that accepts application logs, stores the latest logs in Redis, and broadcasts new log events to WebSocket clients.

## Features

- Accept logs with `service_name`, `log_level`, `message`, and `timestamp`.
- Store the latest 1000 logs in Redis.
- Read recent logs as parsed JSON objects.
- Search recent logs by service name, log level, message text, free-text query, and timestamp range.
- Stream incoming logs over WebSocket.
- Generate demo logs automatically on startup.

## Requirements

- Go 1.26+
- Docker and Docker Compose

## Start Redis

```bash
docker compose up -d redis
```

## Run The App

```bash
go run .
```

The API starts on:

```text
http://localhost:8080
```

## Endpoints

### Health Check

```http
GET /api/health
```

Example:

```bash
curl http://localhost:8080/api/health
```

### Ingest Log

```http
POST /api/logs
```

Example:

```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "payment-service",
    "log_level": "ERROR",
    "message": "Request timeout",
    "timestamp": 1710000000000
  }'
```

### Recent Logs

```http
GET /api/logs/recent
```

Returns the latest logs from Redis as JSON objects.

Example:

```bash
curl http://localhost:8080/api/logs/recent
```

### Search Logs

```http
GET /api/logs/search
```

Supported query parameters:

- `service_name` - exact service name match, case-insensitive.
- `log_level` - exact log level match, case-insensitive.
- `message` - substring search in the log message, case-insensitive.
- `q` - substring search in the log message, case-insensitive.
- `from` - minimum timestamp in milliseconds.
- `to` - maximum timestamp in milliseconds.

Example:

```bash
curl "http://localhost:8080/api/logs/search?service_name=payment-service&log_level=ERROR&q=timeout"
```

Timestamp range example:

```bash
curl "http://localhost:8080/api/logs/search?from=1710000000000&to=1710000300000"
```

The same filters also work on:

```http
GET /api/logs/recent
```

### WebSocket Stream

```text
ws://localhost:8080/api/ws
```

You can open `websocetconnect.html` in a browser to test the stream locally.

## Log Format

```json
{
  "service_name": "payment-service",
  "log_level": "ERROR",
  "message": "Request timeout",
  "timestamp": 1710000000000
}
```

`timestamp` is expected to be a Unix timestamp in milliseconds.

## Tests

```bash
go test ./...
```
