package handler

import (
	"errors"
	"net/http"

	"codescope/server/internal/filebrowser"
	"codescope/server/internal/project"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
)

type FileService interface {
	ListTree(sessionID string) ([]filebrowser.Node, error)
	ReadContent(sessionID, requestedPath string) (filebrowser.Content, error)
	ListTreeByWorkspace(workspaceRoot string) ([]filebrowser.Node, error)
	ReadContentByWorkspace(workspaceRoot, requestedPath string) (filebrowser.Content, error)
}

type FileProjectService interface {
	GetProject(projectID string) (project.Project, error)
}

type FileHandler struct {
	service  FileService
	projects FileProjectService
}

func NewFileHandler(service FileService, projects FileProjectService) *FileHandler {
	return &FileHandler{service: service, projects: projects}
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

func (h *FileHandler) ProjectTree(c *gin.Context) {
	projectRecord, err := h.projects.GetProject(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	nodes, err := h.service.ListTreeByWorkspace(projectRecord.WorkspaceRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (h *FileHandler) ProjectContent(c *gin.Context) {
	projectRecord, err := h.projects.GetProject(c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, project.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	content, err := h.service.ReadContentByWorkspace(projectRecord.WorkspaceRoot, c.Query("path"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}
