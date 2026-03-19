package capture

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

func TestManagedProcessSourceExecutePromptReturnsObservedOutput(t *testing.T) {
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
	source.idleWindow = 30 * time.Millisecond
	source.executionTimeout = 500 * time.Millisecond

	recorder := &sinkRecorder{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- source.Start(ctx, recorder)
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		source.stdinMu.Lock()
		ready := source.stdin != nil
		source.stdinMu.Unlock()
		if ready {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = stdoutWriter.Write([]byte("result line 1\nresult line 2\n"))
	}()

	output, err := source.ExecutePrompt(context.Background(), "continue")
	if err != nil {
		t.Fatalf("execute prompt: %v", err)
	}
	if output != "result line 1\nresult line 2" {
		t.Fatalf("unexpected output %q", output)
	}
	if stdin.String() != "continue\n" {
		t.Fatalf("unexpected stdin %q", stdin.String())
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
		t.Fatal("timed out waiting for shutdown")
	}
}
