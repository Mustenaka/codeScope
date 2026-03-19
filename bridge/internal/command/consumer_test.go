package command

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type promptTargetRecorder struct {
	mu      sync.Mutex
	prompts []string
}

func (r *promptTargetRecorder) SendPrompt(_ context.Context, content string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts = append(r.prompts, content)
	return nil
}

func (r *promptTargetRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.prompts))
	copy(out, r.prompts)
	return out
}

func TestInboxConsumerProcessesNewPromptRecords(t *testing.T) {
	root := t.TempDir()
	inbox := filepath.Join(root, "inbox.jsonl")
	state := filepath.Join(root, "inbox.state.json")
	target := &promptTargetRecorder{}
	consumer := NewInboxConsumer(inbox, state, target, nil)
	consumer.pollInterval = 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Run(ctx)
	}()

	record := []byte("{\"payload\":{\"content\":\"continue fixing tests\"}}\n")
	if err := os.WriteFile(inbox, record, 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		prompts := target.snapshot()
		if len(prompts) == 1 && prompts[0] == "continue fixing tests" {
			cancel()
			select {
			case err := <-errCh:
				if err != nil && err != context.Canceled {
					t.Fatalf("consumer returned error: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for consumer shutdown")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("expected inbox consumer to forward prompt")
}

func TestInboxConsumerPersistsOffsetAcrossRestart(t *testing.T) {
	root := t.TempDir()
	inbox := filepath.Join(root, "inbox.jsonl")
	state := filepath.Join(root, "inbox.state.json")
	target := &promptTargetRecorder{}

	if err := os.WriteFile(inbox, []byte("{\"payload\":{\"content\":\"first\"}}\n"), 0o644); err != nil {
		t.Fatalf("write initial inbox: %v", err)
	}

	consumer := NewInboxConsumer(inbox, state, target, nil)
	if err := consumer.drain(context.Background()); err != nil {
		t.Fatalf("first drain: %v", err)
	}

	firstPrompts := target.snapshot()
	if len(firstPrompts) != 1 || firstPrompts[0] != "first" {
		t.Fatalf("expected first prompt to be consumed once, got %#v", firstPrompts)
	}

	restartedTarget := &promptTargetRecorder{}
	restartedConsumer := NewInboxConsumer(inbox, state, restartedTarget, nil)
	if err := restartedConsumer.drain(context.Background()); err != nil {
		t.Fatalf("restarted drain without new data: %v", err)
	}

	if prompts := restartedTarget.snapshot(); len(prompts) != 0 {
		t.Fatalf("expected no replay after restart, got %#v", prompts)
	}

	file, err := os.OpenFile(inbox, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open inbox for append: %v", err)
	}
	if _, err := file.WriteString("{\"payload\":{\"content\":\"second\"}}\n"); err != nil {
		_ = file.Close()
		t.Fatalf("append inbox: %v", err)
	}
	_ = file.Close()

	if err := restartedConsumer.drain(context.Background()); err != nil {
		t.Fatalf("restarted drain with new data: %v", err)
	}

	if prompts := restartedTarget.snapshot(); len(prompts) != 1 || prompts[0] != "second" {
		t.Fatalf("expected only new prompt after restart, got %#v", prompts)
	}
}

func TestInboxConsumerResetsOffsetWhenInboxIsRewritten(t *testing.T) {
	root := t.TempDir()
	inbox := filepath.Join(root, "inbox.jsonl")
	state := filepath.Join(root, "inbox.state.json")
	target := &promptTargetRecorder{}

	if err := os.WriteFile(inbox, []byte("{\"payload\":{\"content\":\"alpha\"}}\n"), 0o644); err != nil {
		t.Fatalf("write initial inbox: %v", err)
	}

	consumer := NewInboxConsumer(inbox, state, target, nil)
	if err := consumer.drain(context.Background()); err != nil {
		t.Fatalf("initial drain: %v", err)
	}

	if prompts := target.snapshot(); len(prompts) != 1 || prompts[0] != "alpha" {
		t.Fatalf("expected first prompt to be consumed, got %#v", prompts)
	}

	if err := os.WriteFile(inbox, []byte("{\"payload\":{\"content\":\"bravo\"}}\n"), 0o644); err != nil {
		t.Fatalf("rewrite inbox: %v", err)
	}

	restartedTarget := &promptTargetRecorder{}
	restartedConsumer := NewInboxConsumer(inbox, state, restartedTarget, nil)
	if err := restartedConsumer.drain(context.Background()); err != nil {
		t.Fatalf("drain rewritten inbox: %v", err)
	}

	if prompts := restartedTarget.snapshot(); len(prompts) != 1 || prompts[0] != "bravo" {
		t.Fatalf("expected rewritten prompt to be consumed after reset, got %#v", prompts)
	}
}
