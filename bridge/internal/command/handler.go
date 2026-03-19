package command

import (
	"context"
	"log"
	"time"

	"codescope/bridge/internal/session"
)

type Publisher interface {
	Publish(ctx context.Context, msg session.Message) error
}

type Handler struct {
	meta   session.Metadata
	logger *log.Logger
	now    func() time.Time
	sink   PromptSink
}

func NewHandler(meta session.Metadata, logger *log.Logger, sink PromptSink) *Handler {
	if logger == nil {
		logger = log.Default()
	}

	return &Handler{
		meta:   meta,
		logger: logger,
		now:    time.Now,
		sink:   sink,
	}
}

func (h *Handler) Handle(ctx context.Context, msg session.Message, publisher Publisher) error {
	meta := h.meta
	if msg.SessionID != "" {
		meta.SessionID = msg.SessionID
	}

	switch msg.CommandType {
	case session.CommandTypeSendPrompt:
		h.logger.Printf("received send_prompt command command_id=%s", msg.CommandID)

		eventPayload := clonePayload(msg.Payload)
		eventPayload["command_id"] = msg.CommandID
		eventPayload["command_type"] = msg.CommandType
		event := session.NewEventMessage(meta, session.EventTypeCommand, eventPayload, h.now())
		if err := publisher.Publish(ctx, event); err != nil {
			return err
		}

		if h.sink == nil {
			return publisher.Publish(ctx, session.NewCommandResultMessage(meta, msg.CommandID, msg.CommandType, session.StatusFailed, mergeResultPayload(msg.Payload, map[string]any{
				"accepted": false,
				"error":    "no prompt sink configured",
			}), h.now()))
		}

		resultPayload, err := h.sink.WritePrompt(ctx, msg)
		if err != nil {
			return publisher.Publish(ctx, session.NewCommandResultMessage(meta, msg.CommandID, msg.CommandType, session.StatusFailed, mergeResultPayload(msg.Payload, map[string]any{
				"accepted": false,
				"error":    err.Error(),
			}), h.now()))
		}

		return publisher.Publish(ctx, session.NewCommandResultMessage(meta, msg.CommandID, msg.CommandType, session.StatusSuccess, mergeResultPayload(msg.Payload, resultPayload), h.now()))
	default:
		h.logger.Printf("received unsupported command type=%s", msg.CommandType)
		return publisher.Publish(ctx, session.NewCommandResultMessage(meta, msg.CommandID, msg.CommandType, session.StatusFailed, map[string]any{
			"accepted": false,
			"error":    "unsupported command",
		}, h.now()))
	}
}

func clonePayload(payload map[string]any) map[string]any {
	if len(payload) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(payload))
	for key, value := range payload {
		cloned[key] = value
	}
	return cloned
}

func mergeResultPayload(commandPayload, resultPayload map[string]any) map[string]any {
	merged := clonePayload(commandPayload)
	for key, value := range resultPayload {
		merged[key] = value
	}
	return merged
}
