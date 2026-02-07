package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventStats struct {
	Bucket      time.Time `json:"bucket"`
	EventName   string    `json:"event_name"`
	EventCount  int64     `json:"event_count"`
	UniqueUsers int64     `json:"unique_users"`
}

type AnalyticsRepository interface {
	GetHourlyStats(ctx context.Context, eventName string, from, to time.Time) ([]EventStats, error)
	GetDailyStats(ctx context.Context, eventName string, from, to time.Time) ([]EventStats, error)
}

type analyticsRepo struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) AnalyticsRepository {
	return &analyticsRepo{db: db}
}

func (r *analyticsRepo) GetHourlyStats(ctx context.Context, eventName string, from, to time.Time) ([]EventStats, error) {
	query := `
		SELECT bucket, event_name, event_count, unique_users
		FROM events_hourly
		WHERE bucket >= $1 AND bucket <= $2
	`
	args := []interface{}{from, to}

	if eventName != "" {
		query += " AND event_name = $3"
		args = append(args, eventName)
	}

	query += " ORDER BY bucket DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query hourly stats: %w", err)
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		if err := rows.Scan(&s.Bucket, &s.EventName, &s.EventCount, &s.UniqueUsers); err != nil {
			return nil, fmt.Errorf("scan hourly stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (r *analyticsRepo) GetDailyStats(ctx context.Context, eventName string, from, to time.Time) ([]EventStats, error) {
	query := `
		SELECT bucket, event_name, event_count, unique_users
		FROM events_daily
		WHERE bucket >= $1 AND bucket <= $2
	`
	args := []interface{}{from, to}

	if eventName != "" {
		query += " AND event_name = $3"
		args = append(args, eventName)
	}

	query += " ORDER BY bucket DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		if err := rows.Scan(&s.Bucket, &s.EventName, &s.EventCount, &s.UniqueUsers); err != nil {
			return nil, fmt.Errorf("scan daily stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}
