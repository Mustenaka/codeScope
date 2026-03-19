package capture

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"codescope/bridge/internal/session"
)

func TestManagedProcessSourceEmitsOutputAndAcceptsPrompt(t *testing.T) {
	stdoutReader, stdoutWriter := io.Pipe()
	var stdin bytes.Buffer
	waitDone := make(chan struct{})

	source := NewManagedProcessSource("agent.exe", []string{"--stdio"})
	source.startProcess = func(context.Context, string, []string) (io.ReadCloser, io.Writer, func() error, error) {
		return stdoutReader, &stdin, func() error {
			<-waitDone
			return nil
		}, nil
	}

	recorder := &sinkRecorder{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- source.Start(ctx, recorder)
	}()

	if _, err := stdoutWriter.Write([]byte("agent line\n")); err != nil {
		t.Fatalf("write stdout: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(recorder.events) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(recorder.events) == 0 {
		t.Fatal("expected managed process output event")
	}

	if recorder.events[0].Event.Type != session.EventTypeTerminalOutput {
		t.Fatalf("expected terminal_output, got %q", recorder.events[0].Event.Type)
	}

	if err := source.SendPrompt(ctx, "continue"); err != nil {
		t.Fatalf("send prompt: %v", err)
	}

	if got := stdin.String(); got != "continue\n" {
		t.Fatalf("expected prompt to be written to stdin, got %q", got)
	}

	cancel()
	_ = stdoutWriter.Close()
	close(waitDone)

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("source returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for source shutdown")
	}
}

func TestManagedProcessSourceEmitsErrorEventAndRestarts(t *testing.T) {
	var starts atomic.Int32
	var secondStdin bytes.Buffer
	waitSecond := make(chan struct{})

	source := NewManagedProcessSource("agent.exe", []string{"--stdio"})
	source.restartDelay = 10 * time.Millisecond
	source.maxRestarts = 1
	source.startProcess = func(context.Context, string, []string) (io.ReadCloser, io.Writer, func() error, error) {
		if starts.Add(1) == 1 {
			return io.NopCloser(bytes.NewReader([]byte("first run\n"))), io.Discard, func() error {
				return errors.New("exit status 1")
			}, nil
		}
		reader, writer := io.Pipe()
		_ = writer.Close()
		return reader, &secondStdin, func() error {
			<-waitSecond
			return nil
		}, nil
	}

	recorder := &sinkRecorder{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- source.Start(ctx, recorder)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		events := recorder.snapshot()
		if starts.Load() >= 2 && hasManagedProcessErrorEvent(events, "exit status 1", "failed", "runtime") {
			cancel()
			close(waitSecond)
			select {
			case err := <-errCh:
				if err != nil && err != context.Canceled {
					t.Fatalf("source returned error: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for managed process shutdown")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	close(waitSecond)
	t.Fatalf("expected restart and error event, starts=%d events=%#v", starts.Load(), recorder.snapshot())
}

func TestManagedProcessSourceClassifiesStartupFailureSeparately(t *testing.T) {
	var starts atomic.Int32

	source := NewManagedProcessSource("agent.exe", []string{"--stdio"})
	source.restartDelay = 10 * time.Millisecond
	source.maxRestarts = 0
	source.startProcess = func(context.Context, string, []string) (io.ReadCloser, io.Writer, func() error, error) {
		starts.Add(1)
		return nil, nil, nil, errors.New("binary not found")
	}

	recorder := &sinkRecorder{}
	err := source.Start(context.Background(), recorder)
	if err == nil {
		t.Fatal("expected startup failure to bubble up")
	}

	events := recorder.snapshot()
	if !hasManagedProcessErrorEvent(events, "binary not found", "failed", "startup") {
		t.Fatalf("expected startup failure event, got %#v", events)
	}
}

func hasManagedProcessErrorEvent(events []ObservedEvent, expected string, status string, stage string) bool {
	for _, event := range events {
		if event.Event.Type != session.EventTypeError {
			continue
		}
		if event.Event.Payload["message"] == expected && event.Event.Payload["exit_status"] == status && event.Event.Payload["failure_stage"] == stage {
			return true
		}
	}
	return false
}
