package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/project"

	"github.com/gin-gonic/gin"
)

type ProjectService interface {
	ListProjects() ([]project.Project, error)
	GetProject(projectID string) (project.Project, error)
	ListThreads(projectID string) ([]project.Thread, error)
	GetThread(threadID string) (project.Thread, error)
	ListMessages(threadID string) ([]project.Message, error)
}

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	records, err := h.service.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	record, err := h.service.GetProject(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *ProjectHandler) ListThreads(c *gin.Context) {
	records, err := h.service.ListThreads(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *ProjectHandler) GetThread(c *gin.Context) {
	record, err := h.service.GetThread(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *ProjectHandler) ListMessages(c *gin.Context) {
	records, err := h.service.ListMessages(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}
