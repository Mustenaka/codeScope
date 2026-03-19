package capture

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codescope/bridge/internal/session"
)

func TestFileWatcherSourceEmitsFileChangeEvent(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "main.go")
	if err := os.WriteFile(filePath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	source, err := NewFileWatcherSource(session.Metadata{SessionID: "session-1"}, root)
	if err != nil {
		t.Fatalf("new file watcher source: %v", err)
	}
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- source.Start(ctx, recorder)
	}()

	deadline := time.Now().Add(2 * time.Second)
	var found bool
	for time.Now().Before(deadline) {
		if err := os.WriteFile(filePath, []byte("package main\n// changed\n"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		for _, event := range recorder.events {
			if event.Event.Type != session.EventTypeFileChange {
				continue
			}
			if event.Event.Payload["path"] == "main.go" {
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("watcher returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for watcher shutdown")
	}

	if !found {
		t.Fatal("expected file_change event for modified file")
	}
}

func TestFileWatcherSourceIgnoresBuildAndDartToolDirectories(t *testing.T) {
	root := t.TempDir()
	buildDir := filepath.Join(root, "mobile", "build")
	dartToolDir := filepath.Join(root, ".dart_tool")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		t.Fatalf("mkdir build dir: %v", err)
	}
	if err := os.MkdirAll(dartToolDir, 0o755); err != nil {
		t.Fatalf("mkdir dart tool dir: %v", err)
	}

	source, err := NewFileWatcherSource(session.Metadata{SessionID: "session-1"}, root)
	if err != nil {
		t.Fatalf("new file watcher source: %v", err)
	}

	if !source.shouldIgnorePath(filepath.Join(buildDir, "artifact.txt")) {
		t.Fatal("expected build path to be ignored")
	}
	if !source.shouldIgnorePath(filepath.Join(dartToolDir, "state.json")) {
		t.Fatal("expected .dart_tool path to be ignored")
	}
}

func TestFileWatcherSourceDeduplicatesBurstWrites(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "main.go")
	if err := os.WriteFile(filePath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	source, err := NewFileWatcherSource(session.Metadata{SessionID: "session-1"}, root)
	if err != nil {
		t.Fatalf("new file watcher source: %v", err)
	}

	key := dedupeKey(filepath.ToSlash("main.go"), "write")
	if source.seenRecently(key, time.Now()) {
		t.Fatal("expected first event not to be deduped")
	}
	if !source.seenRecently(key, time.Now().Add(50*time.Millisecond)) {
		t.Fatal("expected second burst event to be deduped")
	}
	if source.seenRecently(key, time.Now().Add(500*time.Millisecond)) {
		t.Fatal("expected event outside dedupe window to pass through")
	}
}
