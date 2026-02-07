package models

import (
	"time"
)

type Event struct {
	ID        int64                  `json:"id"`
	EventName string                 `json:"event_name"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

type CreateEventRequest struct {
	EventName string                 `json:"event_name"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type EventFilter struct {
	EventName string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}
