package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/herpiko/blankon-telemetry-backend/internal/usecase"
	"github.com/herpiko/blankon-telemetry-backend/pkg/models"
)

type Handler struct {
	eventUC     usecase.EventUsecase
	analyticsUC usecase.AnalyticsUsecase
}

func NewHandler(eventUC usecase.EventUsecase, analyticsUC usecase.AnalyticsUsecase) *Handler {
	return &Handler{
		eventUC:     eventUC,
		analyticsUC: analyticsUC,
	}
}

type response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response{Data: data})
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response{Error: message})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateEvent: invalid request body: %v", err)
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.eventUC.CreateEvent(r.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidEvent) {
			log.Printf("CreateEvent: invalid event: %v", err)
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("CreateEvent: failed to create event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	h.respondJSON(w, http.StatusCreated, event)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("GetEvent: invalid event id %q: %v", idStr, err)
		h.respondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.eventUC.GetEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrEventNotFound) {
			h.respondError(w, http.StatusNotFound, "event not found")
			return
		}
		log.Printf("GetEvent: failed to get event %d: %v", id, err)
		h.respondError(w, http.StatusInternalServerError, "failed to get event")
		return
	}

	h.respondJSON(w, http.StatusOK, event)
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	filter := models.EventFilter{}

	if name := r.URL.Query().Get("event_name"); name != "" {
		filter.EventName = name
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			filter.From = &from
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			filter.To = &to
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	events, err := h.eventUC.ListEvents(r.Context(), filter)
	if err != nil {
		log.Printf("ListEvents: failed to list events: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	h.respondJSON(w, http.StatusOK, events)
}

func (h *Handler) GetHourlyStats(w http.ResponseWriter, r *http.Request) {
	eventName := r.URL.Query().Get("event_name")

	var from, to time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	stats, err := h.analyticsUC.GetHourlyStats(r.Context(), eventName, from, to)
	if err != nil {
		log.Printf("GetHourlyStats: failed: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to get hourly stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	eventName := r.URL.Query().Get("event_name")

	var from, to time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	stats, err := h.analyticsUC.GetDailyStats(r.Context(), eventName, from, to)
	if err != nil {
		log.Printf("GetDailyStats: failed: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to get daily stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}
