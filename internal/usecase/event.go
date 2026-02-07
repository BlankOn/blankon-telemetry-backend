package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/herpiko/blankon-telemetry-backend/internal/repo"
	"github.com/herpiko/blankon-telemetry-backend/pkg/models"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidEvent  = errors.New("invalid event data")
)

type EventUsecase interface {
	CreateEvent(ctx context.Context, req models.CreateEventRequest) (*models.Event, error)
	GetEvent(ctx context.Context, id int64) (*models.Event, error)
	ListEvents(ctx context.Context, filter models.EventFilter) ([]models.Event, error)
}

type eventUsecase struct {
	repo repo.EventRepository
}

func NewEventUsecase(repo repo.EventRepository) EventUsecase {
	return &eventUsecase{repo: repo}
}

func (u *eventUsecase) CreateEvent(ctx context.Context, req models.CreateEventRequest) (*models.Event, error) {
	if req.EventName == "" {
		return nil, ErrInvalidEvent
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	event := &models.Event{
		EventName: req.EventName,
		Timestamp: req.Timestamp,
		Payload:   req.Payload,
	}

	if err := u.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (u *eventUsecase) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	event, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}

func (u *eventUsecase) ListEvents(ctx context.Context, filter models.EventFilter) ([]models.Event, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	return u.repo.List(ctx, filter)
}
