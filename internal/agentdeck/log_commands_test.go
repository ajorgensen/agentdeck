package agentdeck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskLogPath(t *testing.T) {
	stateDir := "/tmp/agentdeck-state"
	got := TaskLogPath(stateDir, "abc123")
	want := filepath.Join(stateDir, "tasks", "abc123.log")
	if got != want {
		t.Fatalf("log path = %q, want %q", got, want)
	}
}

func TestLogMissingFile(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded, LogPath: TaskLogPath(stateDir, "abc123")}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "log", task.ID)
	if err != nil {
		t.Fatalf("run log: %v", err)
	}
	if !strings.Contains(out, "Log file does not exist") {
		t.Fatalf("log output = %q, want missing log message", out)
	}
}

func TestLogExistingFile(t *testing.T) {
	stateDir := t.TempDir()
	logPath := TaskLogPath(stateDir, "abc123")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}
	if err := os.WriteFile(logPath, []byte("line 1\nline 2\nline 3\n"), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded, LogPath: logPath}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "log", "--lines", "2", task.ID)
	if err != nil {
		t.Fatalf("run log: %v", err)
	}
	if strings.Contains(out, "line 1") || !strings.Contains(out, "line 2") || !strings.Contains(out, "line 3") {
		t.Fatalf("log output = %q, want last two lines", out)
	}
}

func TestLogEmptyFile(t *testing.T) {
	stateDir := t.TempDir()
	logPath := TaskLogPath(stateDir, "abc123")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded, LogPath: logPath}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "log", task.ID)
	if err != nil {
		t.Fatalf("run log: %v", err)
	}
	if !strings.Contains(out, "Log file is empty") {
		t.Fatalf("log output = %q, want empty log message", out)
	}
}
