package api

import (
	"time"

	"github.com/charlesfan/rules-engine/internal/store"
	"github.com/google/uuid"
)

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	DSL         map[string]interface{} `json:"dsl" binding:"required"`
}

// UpdateEventRequest represents a request to update an event
type UpdateEventRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	DSL         map[string]interface{} `json:"dsl"`
	Status      *string                `json:"status"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	DSL         map[string]interface{} `json:"dsl"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EventListResponse represents a list of events with pagination
type EventListResponse struct {
	Data   []*EventResponse `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// CalculateRequest represents a request to calculate price
type CalculateRequest struct {
	Context map[string]interface{} `json:"context" binding:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ToResponse converts a store.Event to EventResponse
func ToResponse(event *store.Event) *EventResponse {
	return &EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		DSL:         event.DSL,
		Status:      event.Status,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
}

// ToResponseList converts a slice of store.Event to EventResponse slice
func ToResponseList(events []*store.Event) []*EventResponse {
	result := make([]*EventResponse, len(events))
	for i, event := range events {
		result[i] = ToResponse(event)
	}
	return result
}
