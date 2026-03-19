package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/filebrowser"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
)

type FileService interface {
	ListTree(sessionID string) ([]filebrowser.Node, error)
	ReadContent(sessionID, requestedPath string) (filebrowser.Content, error)
}

type FileHandler struct {
	service FileService
}

func NewFileHandler(service FileService) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) Tree(c *gin.Context) {
	nodes, err := h.service.ListTree(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (h *FileHandler) Content(c *gin.Context) {
	content, err := h.service.ReadContent(c.Param("id"), c.Query("path"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}
