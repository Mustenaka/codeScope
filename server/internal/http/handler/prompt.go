package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/prompt"
	"codescope/server/internal/session"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
)

type PromptService interface {
	Create(input prompt.CreateInput) (prompt.Template, error)
	List() ([]prompt.Template, error)
}

type CommandService interface {
	CreatePrompt(sessionID string, input session.PromptCommandInput) (session.CommandTask, error)
	ListBySession(sessionID string) ([]session.CommandTask, error)
}

type PromptHandler struct {
	prompts  PromptService
	commands CommandService
}

func NewPromptHandler(prompts PromptService, commands CommandService) *PromptHandler {
	return &PromptHandler{
		prompts:  prompts,
		commands: commands,
	}
}

func (h *PromptHandler) Create(c *gin.Context) {
	var input prompt.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record, err := h.prompts.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (h *PromptHandler) List(c *gin.Context) {
	records, err := h.prompts.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *PromptHandler) Send(c *gin.Context) {
	var input session.PromptCommandInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.commands.CreatePrompt(c.Param("id"), input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, session.ErrBridgeNotConnected) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *PromptHandler) ListCommands(c *gin.Context) {
	records, err := h.commands.ListBySession(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}
