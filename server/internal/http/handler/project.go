package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/project"
	"codescope/server/internal/session"

	"github.com/gin-gonic/gin"
)

type ProjectService interface {
	ListProjects() ([]project.Project, error)
	GetProject(projectID string) (project.Project, error)
	ListThreads(projectID string) ([]project.Thread, error)
	GetThread(threadID string) (project.Thread, error)
	ListMessages(threadID string) ([]project.Message, error)
	PrepareThreadLaunch(projectID string, input project.CreateThreadInput) (project.ThreadLaunch, error)
}

type ProjectCommandService interface {
	CreatePrompt(sessionID string, input session.PromptCommandInput) (session.CommandTask, error)
}

type ProjectHandler struct {
	service  ProjectService
	commands ProjectCommandService
}

func NewProjectHandler(service ProjectService, commands ProjectCommandService) *ProjectHandler {
	return &ProjectHandler{service: service, commands: commands}
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

func (h *ProjectHandler) CreateThread(c *gin.Context) {
	var input project.CreateThreadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	launch, err := h.service.PrepareThreadLaunch(c.Param("id"), input)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, project.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, project.ErrNoWritableSession):
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.commands.CreatePrompt(launch.BackingSessionID, session.PromptCommandInput{
		Content:         input.Content,
		ProjectID:       launch.Thread.ProjectID,
		ThreadID:        launch.Thread.ID,
		ThreadTitle:     launch.Thread.Title,
		SourceSessionID: launch.BackingSessionID,
	}); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, session.ErrBridgeNotConnected) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, launch.Thread)
}
