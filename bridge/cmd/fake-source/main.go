package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type record struct {
	MessageType string         `json:"message_type,omitempty"`
	EventType   string         `json:"event_type,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

func main() {
	var (
		projectName   = flag.String("project-name", "codeScope", "project name injected into payload")
		workspaceRoot = flag.String("workspace-root", ".", "workspace root injected into payload")
		machineID     = flag.String("machine-id", "fake-machine", "machine id injected into payload")
		interval      = flag.Duration("interval", 500*time.Millisecond, "interval between emitted records")
		count         = flag.Int("count", 1, "number of scripted batches to emit")
		includeBad    = flag.Bool("include-invalid", false, "append one malformed line after valid records")
	)
	flag.Parse()

	encoder := json.NewEncoder(os.Stdout)
	for i := 0; i < *count; i++ {
		for _, item := range scriptedRecords(*projectName, *workspaceRoot, *machineID, i+1) {
			if err := encoder.Encode(item); err != nil {
				fmt.Fprintf(os.Stderr, "encode record: %v\n", err)
				os.Exit(1)
			}
			time.Sleep(*interval)
		}
	}

	if *includeBad {
		fmt.Fprintln(os.Stdout, `{"payload":{"content":"missing event type"}}`)
	}
}

func scriptedRecords(projectName, workspaceRoot, machineID string, batch int) []record {
	meta := map[string]any{
		"project_name":   projectName,
		"workspace_root": workspaceRoot,
		"machine_id":     machineID,
	}

	return []record{
		{
			EventType: "terminal_output",
			Payload: merge(meta, map[string]any{
				"content": fmt.Sprintf("batch %d: running go test ./...", batch),
				"stream":  "stdout",
			}),
		},
		{
			EventType: "ai_output",
			Payload: merge(meta, map[string]any{
				"content": fmt.Sprintf("batch %d: updating server websocket flow", batch),
			}),
		},
		{
			EventType: "file_change",
			Payload: merge(meta, map[string]any{
				"path": "server/internal/http/handler/ws.go",
				"op":   "write",
			}),
		},
		{
			MessageType: "heartbeat",
			Payload:     meta,
		},
	}
}

func merge(base, extra map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(extra))
	for key, value := range base {
		out[key] = value
	}
	for key, value := range extra {
		out[key] = value
	}
	return out
}
