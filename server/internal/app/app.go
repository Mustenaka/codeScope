package app

import (
	"fmt"

	"codescope/server/internal/config"
	"codescope/server/internal/event"
	"codescope/server/internal/filebrowser"
	"codescope/server/internal/http/router"
	"codescope/server/internal/project"
	"codescope/server/internal/prompt"
	"codescope/server/internal/session"
	"codescope/server/internal/store"
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

type App struct {
	deps Dependencies
}

type commandResultRecorder struct {
	events *event.Service
}

func New() *App {
	return NewWithConfig(config.Load())
}

func NewWithConfig(cfg config.Config) *App {
	sessionStore := store.NewMemorySessionStore()
	eventStore := store.NewMemoryEventStore()
	promptStore := store.NewMemoryPromptStore()
	commandTaskStore := store.NewMemoryCommandTaskStore()
	hub := event.NewHub()
	bridges := session.NewBridgeRegistry()
	eventService := event.NewService(eventStore, sessionStore, hub)

	return &App{
		deps: Dependencies{
			Config:         cfg,
			SessionService: session.NewService(sessionStore),
			EventService:   eventService,
			EventHub:       hub,
			FileService:    filebrowser.NewService(sessionStore),
			ProjectService: project.NewService(sessionStore, eventStore),
			PromptService:  prompt.NewService(promptStore),
			CommandService: session.NewCommandService(sessionStore, commandTaskStore, bridges, commandResultRecorder{events: eventService}),
		},
	}
}

func (a *App) Dependencies() Dependencies {
	return a.deps
}

func (a *App) Run() error {
	engine := router.New(router.Dependencies{
		Config:         a.deps.Config,
		SessionService: a.deps.SessionService,
		EventService:   a.deps.EventService,
		EventHub:       a.deps.EventHub,
		FileService:    a.deps.FileService,
		ProjectService: a.deps.ProjectService,
		PromptService:  a.deps.PromptService,
		CommandService: a.deps.CommandService,
	})
	address := fmt.Sprintf("%s:%d", a.deps.Config.HTTP.Host, a.deps.Config.HTTP.Port)
	return engine.Run(address)
}

func (r commandResultRecorder) RecordCommandResult(message session.BridgeMessage) error {
	_, err := r.events.Ingest(event.Message{
		MessageID:   message.MessageID,
		SessionID:   message.SessionID,
		MessageType: event.MessageType(message.MessageType),
		EventType:   event.Type(message.EventType),
		CommandID:   message.CommandID,
		CommandType: event.CommandType(message.CommandType),
		Status:      event.CommandStatus(message.Status),
		Timestamp:   message.Timestamp,
		Payload:     message.Payload,
	})
	return err
}
