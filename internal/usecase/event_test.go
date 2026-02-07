package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/herpiko/blankon-telemetry-backend/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventRepository is a mock implementation of EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	if args.Error(0) == nil {
		event.ID = 1
		event.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventRepository) List(ctx context.Context, filter models.EventFilter) ([]models.Event, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Event), args.Error(1)
}

func TestCreateEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	req := models.CreateEventRequest{
		EventName: "test_event",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"key": "value"},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Event")).Return(nil)

	event, err := uc.CreateEvent(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, req.EventName, event.EventName)
	assert.Equal(t, req.Payload, event.Payload)
	mockRepo.AssertExpectations(t)
}

func TestCreateEvent_EmptyEventName(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	req := models.CreateEventRequest{
		EventName: "",
		Timestamp: time.Now(),
	}

	event, err := uc.CreateEvent(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEvent, err)
	assert.Nil(t, event)
}

func TestCreateEvent_DefaultTimestamp(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	req := models.CreateEventRequest{
		EventName: "test_event",
		// No timestamp - should default to now
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Event")).Return(nil)

	event, err := uc.CreateEvent(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.False(t, event.Timestamp.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestCreateEvent_RepoError(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	req := models.CreateEventRequest{
		EventName: "test_event",
		Timestamp: time.Now(),
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Event")).Return(errors.New("db error"))

	event, err := uc.CreateEvent(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, event)
	mockRepo.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	expected := &models.Event{
		ID:        1,
		EventName: "test_event",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"key": "value"},
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(expected, nil)

	event, err := uc.GetEvent(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, event)
	mockRepo.AssertExpectations(t)
}

func TestGetEvent_NotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, nil)

	event, err := uc.GetEvent(ctx, 999)

	assert.Error(t, err)
	assert.Equal(t, ErrEventNotFound, err)
	assert.Nil(t, event)
	mockRepo.AssertExpectations(t)
}

func TestListEvents_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	expected := []models.Event{
		{ID: 1, EventName: "event1"},
		{ID: 2, EventName: "event2"},
	}

	filter := models.EventFilter{Limit: 10}
	mockRepo.On("List", ctx, filter).Return(expected, nil)

	events, err := uc.ListEvents(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, events)
	mockRepo.AssertExpectations(t)
}

func TestListEvents_DefaultLimit(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	expected := []models.Event{}
	
	// When limit is 0, it should default to 100
	mockRepo.On("List", ctx, models.EventFilter{Limit: 100}).Return(expected, nil)

	events, err := uc.ListEvents(ctx, models.EventFilter{Limit: 0})

	assert.NoError(t, err)
	assert.NotNil(t, events)
	mockRepo.AssertExpectations(t)
}

func TestListEvents_MaxLimit(t *testing.T) {
	mockRepo := new(MockEventRepository)
	uc := NewEventUsecase(mockRepo)
	ctx := context.Background()

	expected := []models.Event{}
	
	// When limit exceeds 1000, it should cap at 1000
	mockRepo.On("List", ctx, models.EventFilter{Limit: 1000}).Return(expected, nil)

	events, err := uc.ListEvents(ctx, models.EventFilter{Limit: 5000})

	assert.NoError(t, err)
	assert.NotNil(t, events)
	mockRepo.AssertExpectations(t)
}
