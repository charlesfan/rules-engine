package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements EventStore using PostgreSQL
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Create creates a new event
func (s *PostgresStore) Create(ctx context.Context, event *Event) (*Event, error) {
	dslJSON, err := json.Marshal(event.DSL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DSL: %w", err)
	}

	query := `
		INSERT INTO events (name, description, dsl_json, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	status := event.Status
	if status == "" {
		status = "draft"
	}

	err = s.pool.QueryRow(ctx, query,
		event.Name,
		event.Description,
		dslJSON,
		status,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	event.Status = status
	return event, nil
}

// GetByID retrieves an event by its ID
func (s *PostgresStore) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	query := `
		SELECT id, name, description, dsl_json, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	var event Event
	var dslJSON []byte

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Description,
		&dslJSON,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := json.Unmarshal(dslJSON, &event.DSL); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DSL: %w", err)
	}

	return &event, nil
}

// List retrieves events with optional search query and pagination
func (s *PostgresStore) List(ctx context.Context, query string, limit, offset int) ([]*Event, int, error) {
	// Default pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var (
		rows pgx.Rows
		err  error
	)

	// Count total
	var total int
	if query != "" {
		countQuery := `
			SELECT COUNT(*) FROM events
			WHERE name ILIKE $1 OR description ILIKE $1
		`
		err = s.pool.QueryRow(ctx, countQuery, "%"+query+"%").Scan(&total)
	} else {
		err = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Fetch events
	if query != "" {
		listQuery := `
			SELECT id, name, description, dsl_json, status, created_at, updated_at
			FROM events
			WHERE name ILIKE $1 OR description ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = s.pool.Query(ctx, listQuery, "%"+query+"%", limit, offset)
	} else {
		listQuery := `
			SELECT id, name, description, dsl_json, status, created_at, updated_at
			FROM events
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = s.pool.Query(ctx, listQuery, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		var dslJSON []byte

		if err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Description,
			&dslJSON,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(dslJSON, &event.DSL); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal DSL: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, total, nil
}

// Update updates an existing event
func (s *PostgresStore) Update(ctx context.Context, event *Event) (*Event, error) {
	dslJSON, err := json.Marshal(event.DSL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DSL: %w", err)
	}

	query := `
		UPDATE events
		SET name = $1, description = $2, dsl_json = $3, status = $4
		WHERE id = $5
		RETURNING updated_at
	`

	err = s.pool.QueryRow(ctx, query,
		event.Name,
		event.Description,
		dslJSON,
		event.Status,
		event.ID,
	).Scan(&event.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", event.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

// Delete deletes an event by its ID
func (s *PostgresStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// Close closes the database connection pool
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
