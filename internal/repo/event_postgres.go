package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/herpiko/blankon-telemetry-backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	List(ctx context.Context, filter models.EventFilter) ([]models.Event, error)
}

type eventRepo struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) EventRepository {
	return &eventRepo{db: db}
}

func (r *eventRepo) Create(ctx context.Context, event *models.Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	query := `
		INSERT INTO events (event_name, timestamp, payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err = r.db.QueryRow(ctx, query, event.EventName, event.Timestamp, payloadJSON).
		Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

func (r *eventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	query := `
		SELECT id, event_name, timestamp, payload, created_at
		FROM events
		WHERE id = $1
	`

	var event models.Event
	var payloadJSON []byte

	err := r.db.QueryRow(ctx, query, id).
		Scan(&event.ID, &event.EventName, &event.Timestamp, &payloadJSON, &event.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get event: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	return &event, nil
}

func (r *eventRepo) List(ctx context.Context, filter models.EventFilter) ([]models.Event, error) {
	query := `
		SELECT id, event_name, timestamp, payload, created_at
		FROM events
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.EventName != "" {
		query += fmt.Sprintf(" AND event_name = $%d", argNum)
		args = append(args, filter.EventName)
		argNum++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filter.From)
		argNum++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filter.To)
		argNum++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var payloadJSON []byte

		if err := rows.Scan(&event.ID, &event.EventName, &event.Timestamp, &payloadJSON, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}
