package api

import (
	"github.com/charlesfan/rules-engine/internal/store"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes for events
func RegisterRoutes(r *gin.Engine, eventStore store.EventStore) {
	handler := NewEventsHandler(eventStore)

	events := r.Group("/api/events")
	{
		events.POST("", handler.Create)
		events.GET("", handler.List)
		events.GET("/:id", handler.Get)
		events.PUT("/:id", handler.Update)
		events.DELETE("/:id", handler.Delete)
		events.POST("/:id/validate", handler.Validate)
		events.POST("/:id/calculate", handler.Calculate)
	}
}
