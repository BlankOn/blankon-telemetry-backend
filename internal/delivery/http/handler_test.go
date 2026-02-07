package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/herpiko/blankon-telemetry-backend/internal/usecase"
	"github.com/herpiko/blankon-telemetry-backend/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventUsecase is a mock implementation of EventUsecase
type MockEventUsecase struct {
	mock.Mock
}

func (m *MockEventUsecase) CreateEvent(ctx context.Context, req models.CreateEventRequest) (*models.Event, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventUsecase) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventUsecase) ListEvents(ctx context.Context, filter models.EventFilter) ([]models.Event, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Event), args.Error(1)
}

func TestHealth(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	
	var resp response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Data)
}

func TestCreateEvent_Success(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	now := time.Now()
	reqBody := models.CreateEventRequest{
		EventName: "test_event",
		Timestamp: now,
		Payload:   map[string]interface{}{"key": "value"},
	}

	expectedEvent := &models.Event{
		ID:        1,
		EventName: "test_event",
		Timestamp: now,
		Payload:   map[string]interface{}{"key": "value"},
		CreatedAt: now,
	}

	mockUC.On("CreateEvent", mock.Anything, mock.AnythingOfType("models.CreateEventRequest")).Return(expectedEvent, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEvent_InvalidEvent(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	reqBody := models.CreateEventRequest{
		EventName: "", // Invalid - empty name
	}

	mockUC.On("CreateEvent", mock.Anything, mock.AnythingOfType("models.CreateEventRequest")).Return(nil, usecase.ErrInvalidEvent)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestGetEvent_Success(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	expectedEvent := &models.Event{
		ID:        1,
		EventName: "test_event",
	}

	mockUC.On("GetEvent", mock.Anything, int64(1)).Return(expectedEvent, nil)

	req := httptest.NewRequest(http.MethodGet, "/events/1", nil)
	rec := httptest.NewRecorder()

	// Need to add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetEvent(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestGetEvent_InvalidID(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/events/abc", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetEvent_NotFound(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	mockUC.On("GetEvent", mock.Anything, int64(999)).Return(nil, usecase.ErrEventNotFound)

	req := httptest.NewRequest(http.MethodGet, "/events/999", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetEvent(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestListEvents_Success(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	expectedEvents := []models.Event{
		{ID: 1, EventName: "event1"},
		{ID: 2, EventName: "event2"},
	}

	mockUC.On("ListEvents", mock.Anything, mock.AnythingOfType("models.EventFilter")).Return(expectedEvents, nil)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()

	h.ListEvents(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestListEvents_WithFilters(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	expectedEvents := []models.Event{
		{ID: 1, EventName: "app_launch"},
	}

	mockUC.On("ListEvents", mock.Anything, mock.AnythingOfType("models.EventFilter")).Return(expectedEvents, nil)

	req := httptest.NewRequest(http.MethodGet, "/events?event_name=app_launch&limit=10&offset=0", nil)
	rec := httptest.NewRecorder()

	h.ListEvents(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestListEvents_WithTimeFilters(t *testing.T) {
	mockUC := new(MockEventUsecase)
	h := NewHandler(mockUC)

	expectedEvents := []models.Event{}

	mockUC.On("ListEvents", mock.Anything, mock.AnythingOfType("models.EventFilter")).Return(expectedEvents, nil)

	req := httptest.NewRequest(http.MethodGet, "/events?from=2026-01-01T00:00:00Z&to=2026-12-31T23:59:59Z", nil)
	rec := httptest.NewRecorder()

	h.ListEvents(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}
