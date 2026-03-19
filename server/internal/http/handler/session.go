package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/session"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
)

type SessionService interface {
	Create(input session.CreateInput) (session.Session, error)
	List() ([]session.Session, error)
	Get(id string) (session.Session, error)
	UpdateStatus(id string, status session.Status) (session.Session, error)
}

type SessionHandler struct {
	service SessionService
}

func NewSessionHandler(service SessionService) *SessionHandler {
	return &SessionHandler{service: service}
}

func (h *SessionHandler) Create(c *gin.Context) {
	var input session.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.Create(input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrConflict) {
			status = http.StatusConflict
		} else if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *SessionHandler) List(c *gin.Context) {
	records, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *SessionHandler) Get(c *gin.Context) {
	record, err := h.service.Get(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *SessionHandler) UpdateStatus(c *gin.Context) {
	var input session.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.UpdateStatus(c.Param("id"), input.Status)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}
