package handler

import (
	"errors"
	"net/http"
	"sort"

	"codescope/server/internal/project"
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

type PromptThreadService interface {
	GetThread(threadID string) (project.Thread, error)
	ResolveThreadExecution(threadID string) (project.ThreadExecution, error)
}

type PromptHandler struct {
	prompts  PromptService
	commands CommandService
	threads  PromptThreadService
}

func NewPromptHandler(prompts PromptService, commands CommandService, threads PromptThreadService) *PromptHandler {
	return &PromptHandler{
		prompts:  prompts,
		commands: commands,
		threads:  threads,
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

func (h *PromptHandler) SendToThread(c *gin.Context) {
	var input session.PromptCommandInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.threads.ResolveThreadExecution(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, project.ErrNoWritableSession) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	task, err := h.commands.CreatePrompt(execution.WritableSessionID, session.PromptCommandInput{
		Content:         input.Content,
		ProjectID:       execution.Thread.ProjectID,
		ThreadID:        execution.Thread.ID,
		ThreadTitle:     execution.Thread.Title,
		SourceSessionID: execution.WritableSessionID,
	})
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

func (h *PromptHandler) ListThreadCommands(c *gin.Context) {
	execution, err := h.threads.ResolveThreadExecution(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	records := make([]session.CommandTask, 0, len(execution.SessionIDs))
	for _, sessionID := range execution.SessionIDs {
		sessionRecords, listErr := h.commands.ListBySession(sessionID)
		if listErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": listErr.Error()})
			return
		}
		records = append(records, sessionRecords...)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].CreatedAt.Equal(records[j].CreatedAt) {
			return records[i].ID < records[j].ID
		}
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	c.JSON(http.StatusOK, records)
}
