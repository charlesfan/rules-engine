package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/evaluator"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
	"github.com/charlesfan/rules-engine/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventsHandler handles event-related API requests
type EventsHandler struct {
	store  store.EventStore
	parser *parser.Parser
}

// NewEventsHandler creates a new EventsHandler
func NewEventsHandler(store store.EventStore) *EventsHandler {
	return &EventsHandler{
		store:  store,
		parser: parser.NewParser(),
	}
}

// Create handles POST /api/events
func (h *EventsHandler) Create(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validate DSL
	if err := h.validateDSL(req.DSL); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid DSL: " + err.Error()})
		return
	}

	event := &store.Event{
		Name:        req.Name,
		Description: req.Description,
		DSL:         req.DSL,
		Status:      "draft",
	}

	created, err := h.store.Create(c.Request.Context(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ToResponse(created))
}

// List handles GET /api/events
func (h *EventsHandler) List(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	events, total, err := h.store.List(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, EventListResponse{
		Data:   ToResponseList(events),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// Get handles GET /api/events/:id
func (h *EventsHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	event, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToResponse(event))
}

// Update handles PUT /api/events/:id
func (h *EventsHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	// Get existing event
	event, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Apply updates
	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.DSL != nil {
		if err := h.validateDSL(req.DSL); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid DSL: " + err.Error()})
			return
		}
		event.DSL = req.DSL
	}
	if req.Status != nil {
		event.Status = *req.Status
	}

	updated, err := h.store.Update(c.Request.Context(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToResponse(updated))
}

// Delete handles DELETE /api/events/:id
func (h *EventsHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Validate handles POST /api/events/:id/validate
func (h *EventsHandler) Validate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	event, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.validateDSL(event.DSL); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
}

// Calculate handles POST /api/events/:id/calculate
func (h *EventsHandler) Calculate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid event ID"})
		return
	}

	event, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Parse DSL
	dslData, _ := json.Marshal(event.DSL)
	ruleSet, err := h.parser.Parse(dslData)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid DSL: " + err.Error()})
		return
	}

	// Build context from request
	ctx := h.buildContext(req.Context, ruleSet)

	// Evaluate
	ev := evaluator.NewEvaluator(ruleSet)
	result, err := ev.Evaluate(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// validateDSL validates the DSL using the parser
func (h *EventsHandler) validateDSL(dslMap map[string]interface{}) error {
	data, err := json.Marshal(dslMap)
	if err != nil {
		return err
	}
	_, err = h.parser.Parse(data)
	return err
}

// buildContext builds a dsl.Context from the request context map
func (h *EventsHandler) buildContext(reqCtx map[string]interface{}, ruleSet *dsl.RuleSet) *dsl.Context {
	ctx := &dsl.Context{
		Variables:      ruleSet.Variables,
		ComputedValues: make(map[string]interface{}),
	}

	// Extract known fields from request context
	if user, ok := reqCtx["user"].(map[string]interface{}); ok {
		ctx.User = user
	}
	if team, ok := reqCtx["team"].(map[string]interface{}); ok {
		ctx.Team = team
	}
	if teamSize, ok := reqCtx["team_size"].(float64); ok {
		ctx.TeamSize = int(teamSize)
	}
	if addons, ok := reqCtx["addons"].(map[string]interface{}); ok {
		ctx.Addons = addons
	}

	// Handle register_date
	if regDate, ok := reqCtx["register_date"].(string); ok {
		if t, err := time.Parse(time.RFC3339, regDate); err == nil {
			ctx.RegisterDate = t
		}
	}

	// Copy data sources from rule set
	if ruleSet.DataSources != nil {
		ctx.DataSources = make(map[string]interface{})
		for k, v := range ruleSet.DataSources {
			ctx.DataSources[k] = v
		}
	}

	return ctx
}
