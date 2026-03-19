package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/event"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
)

type EventService interface {
	ListBySession(sessionID string) ([]event.Record, error)
}

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) ListBySession(c *gin.Context) {
	records, err := h.service.ListBySession(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}
