package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"codescope/bridge/internal/capture"
	"codescope/bridge/internal/command"
	"codescope/bridge/internal/config"
	"codescope/bridge/internal/discovery"
	"codescope/bridge/internal/session"
	"codescope/bridge/internal/transport"
)

type App struct {
	cfg       config.Config
	meta      session.Metadata
	transport *transport.Client
	source    capture.Source
	commands  *command.Handler
	logger    *log.Logger
}

func New() *App {
	cfg := config.Load()
	logger := log.Default()
	meta := session.Metadata{
		AgentName:     cfg.AgentName,
		WorkspaceRoot: cfg.WorkspaceRoot,
		MachineID:     cfg.MachineID,
		SessionID:     cfg.SessionID,
	}

	source := newCaptureSource(cfg, meta, logger)
	promptSink := command.NewUnsupportedPromptSink("side-channel mode does not support prompt injection")

	return &App{
		cfg:  cfg,
		meta: meta,
		transport: transport.NewClient(
			cfg.ServerURL,
			transport.WithHeartbeatInterval(0),
			transport.WithLogger(logger),
		),
		source:   source,
		commands: command.NewHandler(meta, logger, promptSink),
		logger:   logger,
	}
}

func newCaptureSource(cfg config.Config, meta session.Metadata, logger *log.Logger) capture.Source {
	switch cfg.CaptureMode {
	case "", "discovery":
		scanner := discovery.NewProcessScanner(nil, cfg.MachineID, cfg.WorkspaceRoot)
		scanner.ConfigureSessionStabilityWindow(cfg.SessionStabilityWindow)
		discoverySource := capture.NewDiscoverySource(
			scanner,
			cfg.DiscoveryInterval,
			meta,
			logger,
			capture.ProcessSnapshotAdapter{},
			capture.NewSessionHeartbeatAdapter(cfg.SessionHeartbeatInterval),
			capture.WorkspaceFileWatcherAdapter{},
		)
		codexRoot, err := resolveCodexRoot()
		if err != nil {
			logger.Printf("codex session capture disabled: %v", err)
			return discoverySource
		}
		return capture.NewMultiSource(
			discoverySource,
			capture.NewCodexSessionSource(codexRoot, cfg.MachineID, cfg.DiscoveryInterval, logger),
		)
	case "reader", "jsonl":
		source, err := newInputSource(cfg, meta)
		if err != nil {
			logger.Fatalf("configure input source: %v", err)
		}
		fileWatcher, err := capture.NewFileWatcherSource(meta, cfg.WorkspaceRoot)
		if err != nil {
			logger.Printf("file watcher disabled root=%q err=%v", cfg.WorkspaceRoot, err)
			return source
		}
		return capture.NewMultiSource(source, fileWatcher)
	default:
		logger.Printf("unsupported capture mode=%q, falling back to discovery", cfg.CaptureMode)
		return newCaptureSource(config.Config{
			AgentName:                cfg.AgentName,
			ServerURL:                cfg.ServerURL,
			WorkspaceRoot:            cfg.WorkspaceRoot,
			MachineID:                cfg.MachineID,
			SessionID:                cfg.SessionID,
			CaptureMode:              "discovery",
			DiscoveryInterval:        cfg.DiscoveryInterval,
			SessionHeartbeatInterval: cfg.SessionHeartbeatInterval,
		}, meta, logger)
	}
}

func resolveCodexRoot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(homeDir, ".codex"), nil
}

func newInputSource(cfg config.Config, meta session.Metadata) (capture.Source, error) {
	reader, err := openSourceReader(cfg.SourceFile)
	if err != nil {
		return nil, err
	}

	switch cfg.SourceMode {
	case "", "reader":
		return capture.NewReaderSource(meta, "stdin", reader, session.EventTypeTerminalOutput), nil
	case "jsonl":
		return capture.NewJSONLSource(meta, reader), nil
	default:
		return nil, fmt.Errorf("unsupported source mode %q", cfg.SourceMode)
	}
}

func openSourceReader(sourceFile string) (io.Reader, error) {
	if sourceFile == "" {
		return os.Stdin, nil
	}

	file, err := os.Open(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("open source file %s: %w", sourceFile, err)
	}
	return file, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a.logger.Printf(
		"starting bridge agent=%q workspace=%q machine_id=%q session_id=%q target=%q",
		a.cfg.AgentName,
		a.cfg.WorkspaceRoot,
		a.cfg.MachineID,
		a.cfg.SessionID,
		a.transport.Target(),
	)

	if err := a.transport.Start(ctx); err != nil {
		return err
	}

	sourceErr := make(chan error, 1)
	go func() {
		sourceErr <- a.source.Start(ctx, a)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sourceErr:
			if err == nil || errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		case msg := <-a.transport.Commands():
			if err := a.commands.Handle(ctx, msg, a.transport); err != nil {
				return err
			}
		}
	}
}

func (a *App) Emit(ctx context.Context, observed capture.ObservedEvent) error {
	meta := observed.Meta
	if meta.SessionID == "" {
		meta = a.meta
	}
	if observed.Event.Type == session.EventTypeHeartbeat {
		return a.transport.Publish(ctx, session.NewHeartbeatMessage(meta, time.Now()))
	}
	return a.transport.Publish(ctx, session.NewEventMessage(meta, observed.Event.Type, observed.Event.Payload, time.Now()))
}
