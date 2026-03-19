package session

import "time"

type CommandTaskStatus string

const (
	CommandTaskPending CommandTaskStatus = "pending"
	CommandTaskRunning CommandTaskStatus = "running"
	CommandTaskSuccess CommandTaskStatus = "success"
	CommandTaskFailed  CommandTaskStatus = "failed"
)

type CommandTask struct {
	ID        string            `json:"id"`
	SessionID string            `json:"session_id"`
	TaskType  string            `json:"task_type"`
	Payload   map[string]any    `json:"payload"`
	Status    CommandTaskStatus `json:"status"`
	Result    string            `json:"result"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
