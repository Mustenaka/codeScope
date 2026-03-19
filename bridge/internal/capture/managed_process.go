package capture

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"codescope/bridge/internal/session"
	"github.com/creack/pty"
)

type processStarter func(ctx context.Context, command string, args []string) (io.ReadCloser, io.Writer, func() error, error)

const (
	failureStageStartup = "startup"
	failureStageRuntime = "runtime"
	failureStageExit    = "exit"
)

type ManagedProcessSource struct {
	command          string
	args             []string
	startProcess     processStarter
	restartDelay     time.Duration
	maxRestarts      int
	idleWindow       time.Duration
	executionTimeout time.Duration

	stdinMu sync.Mutex
	stdin   io.Writer

	subscribersMu sync.Mutex
	subscribers   map[uint64]chan string
	nextSubID     uint64
}

func NewManagedProcessSource(command string, args []string) *ManagedProcessSource {
	return &ManagedProcessSource{
		command:          command,
		args:             append([]string(nil), args...),
		restartDelay:     250 * time.Millisecond,
		maxRestarts:      3,
		idleWindow:       200 * time.Millisecond,
		executionTimeout: 5 * time.Second,
		subscribers:      make(map[uint64]chan string),
		startProcess: func(ctx context.Context, command string, args []string) (io.ReadCloser, io.Writer, func() error, error) {
			cmd := exec.CommandContext(ctx, command, args...)
			ptyFile, err := pty.Start(cmd)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("start managed process with pty: %w", err)
			}
			wait := func() error {
				err := cmd.Wait()
				_ = ptyFile.Close()
				return err
			}
			return ptyFile, ptyFile, wait, nil
		},
	}
}

func (s *ManagedProcessSource) ConfigureRestart(maxRestarts int, restartDelay time.Duration) {
	s.maxRestarts = maxRestarts
	s.restartDelay = restartDelay
}

func (s *ManagedProcessSource) Start(ctx context.Context, sink Sink) error {
	restarts := 0
	for {
		stage, err := s.runOnce(ctx, sink)
		if err == nil || ctx.Err() != nil {
			return ctx.Err()
		}

		restartScheduled := s.maxRestarts < 0 || restarts < s.maxRestarts
		backoff := s.restartBackoff(stage, restarts)

		_ = sink.Emit(ctx, ObservedEvent{
			Event: session.Event{
				Type: session.EventTypeError,
				Payload: map[string]any{
					"kind":              "managed_process_exit",
					"command":           s.command,
					"args":              append([]string(nil), s.args...),
					"message":           err.Error(),
					"exit_status":       exitStatusFromError(err),
					"failure_stage":     stage,
					"restart_attempt":   restarts,
					"restart_scheduled": restartScheduled,
					"restart_delay_ms":  backoff.Milliseconds(),
					"source":            "managed_process",
				},
			},
		})

		if !restartScheduled {
			return err
		}
		restarts++

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (s *ManagedProcessSource) runOnce(ctx context.Context, sink Sink) (string, error) {
	stdout, stdin, wait, err := s.startProcess(ctx, s.command, s.args)
	if err != nil {
		return failureStageStartup, err
	}
	defer stdout.Close()

	s.stdinMu.Lock()
	s.stdin = stdin
	s.stdinMu.Unlock()
	defer func() {
		s.stdinMu.Lock()
		s.stdin = nil
		s.stdinMu.Unlock()
	}()

	scanErrCh := make(chan error, 1)
	go func() {
		scanErrCh <- s.scanOutput(ctx, stdout, sink)
	}()

	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- wait()
	}()

	var scanErr error
	for {
		select {
		case <-ctx.Done():
			return failureStageExit, ctx.Err()
		case err := <-scanErrCh:
			scanErr = err
			if scanErr != nil && !errors.Is(scanErr, io.EOF) && !errors.Is(scanErr, context.Canceled) {
				return failureStageRuntime, scanErr
			}
		case err := <-waitErrCh:
			if err == nil {
				return failureStageExit, fmt.Errorf("managed process exited")
			}
			return failureStageRuntime, err
		}
	}
}

func (s *ManagedProcessSource) scanOutput(ctx context.Context, stdout io.Reader, sink Sink) error {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if err := sink.Emit(ctx, ObservedEvent{
			Event: session.Event{
				Type: session.EventTypeTerminalOutput,
				Payload: map[string]any{
					"content": line,
					"source":  "pty",
				},
			},
		}); err != nil {
			return err
		}
		s.publishLine(line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func (s *ManagedProcessSource) SendPrompt(_ context.Context, content string) error {
	s.stdinMu.Lock()
	defer s.stdinMu.Unlock()

	if s.stdin == nil {
		return fmt.Errorf("managed process stdin not ready")
	}

	if _, err := io.WriteString(s.stdin, content+"\n"); err != nil {
		return fmt.Errorf("write prompt to managed process: %w", err)
	}
	return nil
}

func (s *ManagedProcessSource) ExecutePrompt(ctx context.Context, content string) (string, error) {
	subscription, cancel := s.subscribe()
	defer cancel()

	if err := s.SendPrompt(ctx, content); err != nil {
		return "", err
	}

	timer := time.NewTimer(s.executionTimeout)
	defer timer.Stop()

	var (
		lines     []string
		idleTimer *time.Timer
		idleC     <-chan time.Time
	)

	resetIdle := func() {
		if idleTimer != nil {
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
		}
		idleTimer = time.NewTimer(s.idleWindow)
		idleC = idleTimer.C
	}
	defer func() {
		if idleTimer != nil {
			idleTimer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timer.C:
			return strings.Join(lines, "\n"), nil
		case <-idleC:
			return strings.Join(lines, "\n"), nil
		case line, ok := <-subscription:
			if !ok {
				return strings.Join(lines, "\n"), nil
			}
			lines = append(lines, line)
			resetIdle()
		}
	}
}

func (s *ManagedProcessSource) subscribe() (<-chan string, func()) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	s.nextSubID++
	id := s.nextSubID
	ch := make(chan string, 32)
	s.subscribers[id] = ch

	return ch, func() {
		s.subscribersMu.Lock()
		defer s.subscribersMu.Unlock()
		current, ok := s.subscribers[id]
		if !ok {
			return
		}
		delete(s.subscribers, id)
		close(current)
	}
}

func (s *ManagedProcessSource) publishLine(line string) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()
	for _, ch := range s.subscribers {
		select {
		case ch <- line:
		default:
		}
	}
}

func exitStatusFromError(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, io.EOF) || err.Error() == "managed process exited" {
		return "exited"
	}
	return "failed"
}

func (s *ManagedProcessSource) restartBackoff(stage string, attempt int) time.Duration {
	base := s.restartDelay
	switch stage {
	case failureStageStartup:
		multiplier := attempt + 1
		if multiplier > 4 {
			multiplier = 4
		}
		return time.Duration(multiplier) * base
	case failureStageExit:
		half := base / 2
		if half <= 0 {
			return time.Millisecond
		}
		return half
	default:
		return base
	}
}
