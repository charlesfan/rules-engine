package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event represents an event with its DSL rules
type Event struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	DSL         map[string]interface{} `json:"dsl"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EventStore defines the interface for event storage operations
type EventStore interface {
	// Create creates a new event and returns the created event with ID
	Create(ctx context.Context, event *Event) (*Event, error)

	// GetByID retrieves an event by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// List retrieves events with optional search query and pagination
	List(ctx context.Context, query string, limit, offset int) ([]*Event, int, error)

	// Update updates an existing event
	Update(ctx context.Context, event *Event) (*Event, error)

	// Delete deletes an event by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Close closes the store connection
	Close() error
}
