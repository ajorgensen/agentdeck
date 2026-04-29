package agentdeck

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"time"
)

const taskIDLength = 10

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusKilled    TaskStatus = "killed"
	TaskStatusUnknown   TaskStatus = "unknown"
)

type Task struct {
	ID              string     `json:"id"`
	Agent           string     `json:"agent"`
	CWD             string     `json:"cwd"`
	Prompt          string     `json:"prompt"`
	NativeSessionID string     `json:"native_session_id,omitempty"`
	PID             int        `json:"pid,omitempty"`
	Status          TaskStatus `json:"status"`
	LogPath         string     `json:"log_path"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func NewTaskID() (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
	id = strings.ToLower(id)
	return id[:taskIDLength], nil
}
