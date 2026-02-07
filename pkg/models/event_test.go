package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent_JSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	
	event := Event{
		ID:        1,
		EventName: "test_event",
		Timestamp: now,
		Payload: map[string]interface{}{
			"version": "1.0.0",
			"count":   float64(42),
		},
		CreatedAt: now,
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	assert.NoError(t, err)

	// Unmarshal back
	var decoded Event
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.EventName, decoded.EventName)
	assert.Equal(t, event.Payload["version"], decoded.Payload["version"])
	assert.Equal(t, event.Payload["count"], decoded.Payload["count"])
}

func TestCreateEventRequest_JSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	
	req := CreateEventRequest{
		EventName: "app_launch",
		Timestamp: now,
		Payload: map[string]interface{}{
			"os":   "linux",
			"arch": "amd64",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	assert.NoError(t, err)

	// Unmarshal back
	var decoded CreateEventRequest
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, req.EventName, decoded.EventName)
	assert.Equal(t, req.Payload["os"], decoded.Payload["os"])
}

func TestEventFilter_Defaults(t *testing.T) {
	filter := EventFilter{}
	
	assert.Equal(t, "", filter.EventName)
	assert.Nil(t, filter.From)
	assert.Nil(t, filter.To)
	assert.Equal(t, 0, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
}

func TestEventFilter_WithValues(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)
	
	filter := EventFilter{
		EventName: "app_launch",
		From:      &now,
		To:        &later,
		Limit:     50,
		Offset:    10,
	}
	
	assert.Equal(t, "app_launch", filter.EventName)
	assert.Equal(t, &now, filter.From)
	assert.Equal(t, &later, filter.To)
	assert.Equal(t, 50, filter.Limit)
	assert.Equal(t, 10, filter.Offset)
}
