package capture

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codescope/bridge/internal/session"
	"github.com/fsnotify/fsnotify"
)

type Sink interface {
	Emit(ctx context.Context, event ObservedEvent) error
}

type Source interface {
	Start(ctx context.Context, sink Sink) error
}

type ObservedEvent struct {
	Meta  session.Metadata
	Event session.Event
}

type NoopSource struct{}

type ReaderSource struct {
	meta      session.Metadata
	name      string
	reader    io.Reader
	eventType string
}

type MultiSource struct {
	sources []Source
}

type FileWatcherSource struct {
	meta         session.Metadata
	root         string
	ignored      map[string]struct{}
	dedupeWindow time.Duration
	recentMu     sync.Mutex
	recentEvents map[string]time.Time
}

type JSONLSource struct {
	meta   session.Metadata
	reader io.Reader
}

type jsonlRecord struct {
	MessageType string         `json:"message_type"`
	EventType   string         `json:"event_type"`
	Payload     map[string]any `json:"payload"`
}

func NewNoopSource() NoopSource {
	return NoopSource{}
}

func NewReaderSource(meta session.Metadata, name string, reader io.Reader, eventType string) ReaderSource {
	return ReaderSource{
		meta:      meta,
		name:      name,
		reader:    reader,
		eventType: eventType,
	}
}

func NewMultiSource(sources ...Source) MultiSource {
	return MultiSource{sources: sources}
}

func NewFileWatcherSource(meta session.Metadata, root string) (*FileWatcherSource, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}

	return &FileWatcherSource{
		meta: meta,
		root: absoluteRoot,
		ignored: map[string]struct{}{
			".git":         {},
			"node_modules": {},
			".codescope":   {},
			".dart_tool":   {},
			"build":        {},
			"dist":         {},
			"coverage":     {},
			"tmp":          {},
			"temp":         {},
			"target":       {},
			"out":          {},
		},
		dedupeWindow: 250 * time.Millisecond,
		recentEvents: make(map[string]time.Time),
	}, nil
}

func NewJSONLSource(meta session.Metadata, reader io.Reader) JSONLSource {
	return JSONLSource{meta: meta, reader: reader}
}

func (NoopSource) Start(ctx context.Context, _ Sink) error {
	log.Printf("capture source placeholder active")
	<-ctx.Done()
	return ctx.Err()
}

func (s ReaderSource) Start(ctx context.Context, sink Sink) error {
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		line := scanner.Text()
		if err := sink.Emit(ctx, ObservedEvent{
			Meta: s.meta,
			Event: session.Event{
				Type: s.eventType,
				Payload: map[string]any{
					"content": line,
					"source":  s.name,
				},
			},
		}); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}

	return nil
}

func (m MultiSource) Start(ctx context.Context, sink Sink) error {
	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(m.sources))
	)

	for _, source := range m.sources {
		source := source
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := source.Start(ctx, sink); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-done:
		return nil
	}
}

func (s JSONLSource) Start(ctx context.Context, sink Sink) error {
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record jsonlRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("decode jsonl event: %w", err)
		}

		eventType := strings.TrimSpace(record.EventType)
		if strings.EqualFold(strings.TrimSpace(record.MessageType), session.MessageTypeHeartbeat) {
			eventType = session.EventTypeHeartbeat
		}
		if eventType == "" {
			return errors.New("event_type is required")
		}

		if err := sink.Emit(ctx, ObservedEvent{
			Meta: s.meta,
			Event: session.Event{
				Type:    eventType,
				Payload: cloneAnyMap(record.Payload),
			},
		}); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return nil
}

func (s *FileWatcherSource) Start(ctx context.Context, sink Sink) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create file watcher: %w", err)
	}
	defer watcher.Close()

	if err := s.addRecursive(watcher, s.root); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-watcher.Errors:
			if err != nil {
				return fmt.Errorf("watcher error: %w", err)
			}
		case event := <-watcher.Events:
			if event.Name == "" || s.shouldIgnorePath(event.Name) {
				continue
			}

			if event.Has(fsnotify.Create) {
				info, statErr := os.Stat(event.Name)
				if statErr == nil && info.IsDir() {
					if err := s.addRecursive(watcher, event.Name); err != nil {
						return err
					}
					continue
				}
			}

			if isDirectory(event.Name) {
				continue
			}

			relPath, err := filepath.Rel(s.root, event.Name)
			if err != nil {
				continue
			}
			normalizedPath := filepath.ToSlash(relPath)
			op := fsnotifyOp(event.Op)
			if s.seenRecently(dedupeKey(normalizedPath, op), time.Now()) {
				continue
			}

			if err := sink.Emit(ctx, ObservedEvent{
				Meta: s.meta,
				Event: session.Event{
					Type: session.EventTypeFileChange,
					Payload: map[string]any{
						"path": normalizedPath,
						"op":   op,
					},
				},
			}); err != nil {
				return err
			}
		}
	}
}

func (s *FileWatcherSource) addRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if s.shouldIgnorePath(path) {
			return filepath.SkipDir
		}
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("watch directory %s: %w", path, err)
		}
		return nil
	})
}

func (s *FileWatcherSource) shouldIgnorePath(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(filepath.Separator))
	for _, part := range parts {
		if _, ignored := s.ignored[part]; ignored {
			return true
		}
	}
	return false
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fsnotifyOp(op fsnotify.Op) string {
	switch {
	case op.Has(fsnotify.Create):
		return "create"
	case op.Has(fsnotify.Write):
		return "write"
	case op.Has(fsnotify.Remove):
		return "remove"
	case op.Has(fsnotify.Rename):
		return "rename"
	case op.Has(fsnotify.Chmod):
		return "chmod"
	default:
		return "unknown"
	}
}

func cloneAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func dedupeKey(path, op string) string {
	return path + "|" + op
}

func (s *FileWatcherSource) seenRecently(key string, now time.Time) bool {
	s.recentMu.Lock()
	defer s.recentMu.Unlock()

	cutoff := now.Add(-s.dedupeWindow)
	for existingKey, seenAt := range s.recentEvents {
		if seenAt.Before(cutoff) {
			delete(s.recentEvents, existingKey)
		}
	}

	if seenAt, ok := s.recentEvents[key]; ok && now.Sub(seenAt) < s.dedupeWindow {
		return true
	}

	s.recentEvents[key] = now
	return false
}
