# BlankOn Telemetry Backend

A telemetry backend service built with Go, Chi router, and PostgreSQL.

## Architecture

```
cmd/server/          - Entry point
internal/
  ├── delivery/http/ - HTTP handlers and router (Chi)
  ├── usecase/       - Business logic
  └── repo/          - Database repository (PostgreSQL)
pkg/models/          - Shared data models
migrations/          - SQL migrations
```

## Requirements

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

## Quick Start

### Using Docker Compose

```bash
docker-compose up -d
```

### Manual Setup

1. Start PostgreSQL and create database:
```bash
createdb telemetry
psql -d telemetry -f migrations/001_create_events.sql
```

2. Run the server:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable"
export PORT=8080
go run ./cmd/server
```

## API Endpoints

### Health Check
```
GET /health
```

### Create Event
```
POST /events
Content-Type: application/json

{
  "event_name": "app_launch",
  "timestamp": "2026-02-06T18:40:00Z",
  "payload": {
    "version": "1.0.0",
    "os": "linux"
  }
}
```

### List Events
```
GET /events
GET /events?event_name=app_launch
GET /events?from=2026-02-01T00:00:00Z&to=2026-02-28T23:59:59Z
GET /events?limit=50&offset=0
```

### Get Event by ID
```
GET /events/{id}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable | PostgreSQL connection string |
| PORT | 8080 | Server port |

## License

MIT
