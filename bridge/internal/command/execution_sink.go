package command

import (
	"context"
	"fmt"

	"codescope/bridge/internal/session"
)

type PromptExecutor interface {
	ExecutePrompt(ctx context.Context, content string) (string, error)
}

type ExecutionPromptSink struct {
	executor PromptExecutor
}

func NewExecutionPromptSink(executor PromptExecutor) *ExecutionPromptSink {
	return &ExecutionPromptSink{executor: executor}
}

func (s *ExecutionPromptSink) WritePrompt(ctx context.Context, msg session.Message) (map[string]any, error) {
	if s.executor == nil {
		return nil, fmt.Errorf("no prompt executor configured")
	}
	content, _ := msg.Payload["content"].(string)
	result, err := s.executor.ExecutePrompt(ctx, content)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"accepted": true,
		"result":   result,
	}, nil
}
