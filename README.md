# BlankOn Telemetry Backend

A telemetry backend service built with Go, Chi router, and **TimescaleDB** for time-series analytics.

## Features

- **TimescaleDB hypertables** for automatic time-based partitioning
- **Automatic compression** for data older than 7 days
- **Continuous aggregates** for fast dashboard queries (hourly/daily)
- Clean architecture (delivery/usecase/repo layers)

## Architecture

```
cmd/server/          - Entry point
internal/
  ├── delivery/http/ - HTTP handlers and router (Chi)
  ├── usecase/       - Business logic
  └── repo/          - Database repository (TimescaleDB)
pkg/models/          - Shared data models
migrations/          - SQL migrations
```

## Requirements

- Go 1.22+
- TimescaleDB (PostgreSQL 16 + TimescaleDB extension)
- Docker & Docker Compose (optional)

## Quick Start

### Using Docker Compose

```bash
docker-compose up -d
```

### Manual Setup

1. Install TimescaleDB and create database:
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

### Events

#### Create Event
```bash
POST /events
Content-Type: application/json

{
  "event_name": "app_launch",
  "timestamp": "2026-02-06T18:40:00Z",
  "payload": {
    "version": "1.0.0",
    "os": "linux",
    "user_id": "user123"
  }
}
```

#### List Events
```
GET /events
GET /events?event_name=app_launch
GET /events?from=2026-02-01T00:00:00Z&to=2026-02-28T23:59:59Z
GET /events?limit=50&offset=0
```

#### Get Event by ID
```
GET /events/{id}
```

### Analytics (TimescaleDB Continuous Aggregates)

#### Hourly Stats
```
GET /analytics/hourly
GET /analytics/hourly?event_name=app_launch
GET /analytics/hourly?from=2026-02-01T00:00:00Z&to=2026-02-07T00:00:00Z
```

Response:
```json
{
  "data": [
    {
      "bucket": "2026-02-06T18:00:00Z",
      "event_name": "app_launch",
      "event_count": 1523,
      "unique_users": 342
    }
  ]
}
```

#### Daily Stats
```
GET /analytics/daily
GET /analytics/daily?event_name=app_launch
GET /analytics/daily?from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | *(built from parts below)* | TimescaleDB connection string (takes precedence if set) |
| POSTGRES_USER | postgres | Database user |
| POSTGRES_PASSWORD | postgres | Database password |
| POSTGRES_HOST | localhost | Database host |
| POSTGRES_PORT | 5432 | Database port |
| POSTGRES_DB | telemetry | Database name |
| POSTGRES_SSLMODE | disable | SSL mode |
| PORT | 8080 | Server port |

## TimescaleDB Features Used

- **Hypertables**: Auto-partitioning by timestamp
- **Compression**: Data older than 7 days is compressed automatically
- **Continuous Aggregates**: Pre-computed hourly/daily stats
- **Retention Policies**: (Optional) Auto-delete old data

## Running Tests

```bash
go test ./... -v
```

## License

MIT
