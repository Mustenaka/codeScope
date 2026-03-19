package router

import (
	"codescope/server/internal/config"
	"codescope/server/internal/event"
	"codescope/server/internal/filebrowser"
	"codescope/server/internal/http/handler"
	"codescope/server/internal/project"
	"codescope/server/internal/prompt"
	"codescope/server/internal/session"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Config         config.Config
	SessionService *session.Service
	EventService   *event.Service
	EventHub       *event.Hub
	FileService    *filebrowser.Service
	ProjectService *project.Service
	PromptService  *prompt.Service
	CommandService *session.CommandService
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	bridgeRegistry := session.NewBridgeRegistry()
	if deps.CommandService != nil {
		bridgeRegistry = deps.CommandService.Bridges()
	}

	sessionHandler := handler.NewSessionHandler(deps.SessionService)
	eventHandler := handler.NewEventHandler(deps.EventService)
	fileHandler := handler.NewFileHandler(deps.FileService, deps.ProjectService)
	projectHandler := handler.NewProjectHandler(deps.ProjectService, deps.CommandService)
	promptHandler := handler.NewPromptHandler(deps.PromptService, deps.CommandService, deps.ProjectService)
	bridgeHandler := handler.NewBridgeHandler(bridgeRegistry)
	wsHandler := handler.NewWebSocketHandler(deps.EventService, deps.SessionService, deps.CommandService, bridgeRegistry, deps.EventHub)

	api := engine.Group("/api")
	{
		api.GET("/health", handler.Health(deps.Config.AppName))
		api.GET("/bridge/status", bridgeHandler.Status)
		api.GET("/projects", projectHandler.ListProjects)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.GET("/projects/:id/threads", projectHandler.ListThreads)
		api.POST("/projects/:id/threads", projectHandler.CreateThread)
		api.GET("/projects/:id/files/tree", fileHandler.ProjectTree)
		api.GET("/projects/:id/files/content", fileHandler.ProjectContent)
		api.GET("/threads/:id", projectHandler.GetThread)
		api.GET("/threads/:id/messages", projectHandler.ListMessages)
		api.POST("/threads/:id/commands/prompt", promptHandler.SendToThread)
		api.GET("/threads/:id/commands", promptHandler.ListThreadCommands)
		api.GET("/sessions", sessionHandler.List)
		api.POST("/sessions", sessionHandler.Create)
		api.GET("/sessions/:id", sessionHandler.Get)
		api.PATCH("/sessions/:id/status", sessionHandler.UpdateStatus)
		api.GET("/sessions/:id/events", eventHandler.ListBySession)
		api.GET("/sessions/:id/files/tree", fileHandler.Tree)
		api.GET("/sessions/:id/files/content", fileHandler.Content)
		api.POST("/sessions/:id/commands/prompt", promptHandler.Send)
		api.GET("/sessions/:id/commands", promptHandler.ListCommands)
		api.GET("/prompts", promptHandler.List)
		api.POST("/prompts", promptHandler.Create)
	}

	engine.GET("/ws/bridge", wsHandler.Bridge)
	engine.GET("/ws/mobile", wsHandler.Mobile)

	return engine
}
