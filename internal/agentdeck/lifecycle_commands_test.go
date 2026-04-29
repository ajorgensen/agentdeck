package agentdeck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForgetMissingTask(t *testing.T) {
	_, err := runTestApp(t, "forget", "missing")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("error = %v, want ErrTaskNotFound", err)
	}
}

func TestForgetRefusesRunningTask(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "shell", PID: os.Getpid(), Status: TaskStatusRunning, CWD: t.TempDir()}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	_, err := runTestAppWithState(t, stateDir, "forget", task.ID)
	if err == nil || !strings.Contains(err.Error(), "is running") {
		t.Fatalf("error = %v, want running refusal", err)
	}
	if _, err := NewJSONTaskStore(stateDir).Get(task.ID); err != nil {
		t.Fatalf("task should remain in store: %v", err)
	}
}

func TestForgetCompletedTaskDeletesIndexAndPreservesLog(t *testing.T) {
	stateDir := t.TempDir()
	logPath := TaskLogPath(stateDir, "abc123")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}
	if err := os.WriteFile(logPath, []byte("log"), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	task := Task{ID: "abc123", Agent: "shell", Status: TaskStatusSucceeded, LogPath: logPath}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "forget", task.ID)
	if err != nil {
		t.Fatalf("run forget: %v", err)
	}
	if !strings.Contains(out, "Forgot task abc123") {
		t.Fatalf("forget output = %q", out)
	}
	if _, err := NewJSONTaskStore(stateDir).Get(task.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("get forgotten task err = %v, want ErrTaskNotFound", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log should be preserved: %v", err)
	}
}

func TestKillAlreadyExitedProcessMarksKilled(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "shell", PID: 0, Status: TaskStatusRunning, CWD: t.TempDir()}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "kill", task.ID)
	if err != nil {
		t.Fatalf("run kill: %v", err)
	}
	if !strings.Contains(out, "Killed task abc123") {
		t.Fatalf("kill output = %q", out)
	}
	got, err := NewJSONTaskStore(stateDir).Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Status != TaskStatusKilled {
		t.Fatalf("status = %q, want killed", got.Status)
	}
}

func TestKillRunningTaskUsesAdapter(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "fake", PID: os.Getpid(), Status: TaskStatusRunning, CWD: t.TempDir()}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	_, err := runTestAppWithStateAndAdapters(t, stateDir, NewAdapterRegistry(FakeAdapter{NameValue: "fake"}), "kill", task.ID)
	if err != nil {
		t.Fatalf("run kill: %v", err)
	}
	got, err := NewJSONTaskStore(stateDir).Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Status != TaskStatusKilled {
		t.Fatalf("status = %q, want killed", got.Status)
	}
}
