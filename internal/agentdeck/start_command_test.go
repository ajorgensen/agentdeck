package agentdeck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartInvalidDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	_, err := runTestApp(t, "start", "--agent", "shell", "--dir", missing, "echo hi")
	if err == nil {
		t.Fatal("start with missing directory returned nil error")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("error = %v, want not a directory", err)
	}
}

func TestStartFailedAdapterLaunchMarksTaskFailed(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	startErr := errors.New("boom")

	_, err := runTestAppWithStateAndAdapters(t, stateDir, NewAdapterRegistry(FakeAdapter{NameValue: "fake", StartErr: startErr}), "start", "--agent", "fake", "--dir", workDir, "do work")
	if err == nil {
		t.Fatal("start with failing adapter returned nil error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %v, want adapter error", err)
	}

	tasks, err := NewJSONTaskStore(stateDir).List()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %#v, want one failed task", tasks)
	}
	if tasks[0].Status != TaskStatusFailed {
		t.Fatalf("task status = %q, want failed", tasks[0].Status)
	}
	if _, err := os.Stat(tasks[0].LogPath); err != nil {
		t.Fatalf("stat log path: %v", err)
	}
}

func TestStartSuccessfulLaunchMetadata(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	adapter := FakeAdapter{NameValue: "fake", StartResult: StartResult{PID: 123, NativeSessionID: "native-1"}}

	out, err := runTestAppWithStateAndAdapters(t, stateDir, NewAdapterRegistry(adapter), "start", "--agent", "fake", "--dir", workDir, "do", "work")
	if err != nil {
		t.Fatalf("run start: %v", err)
	}
	if !strings.Contains(out, "Started task ") || !strings.Contains(out, "agentdeck log ") {
		t.Fatalf("start output = %q, want task id and next-step hint", out)
	}

	tasks, err := NewJSONTaskStore(stateDir).List()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %#v, want one task", tasks)
	}
	task := tasks[0]
	if task.Agent != "fake" || task.Status != TaskStatusRunning || task.PID != 123 || task.NativeSessionID != "native-1" {
		t.Fatalf("task metadata = %#v", task)
	}
	if task.CWD != workDir || task.Prompt != "do work" {
		t.Fatalf("task cwd/prompt = %q/%q, want %q/do work", task.CWD, task.Prompt, workDir)
	}
	if _, err := os.Stat(task.LogPath); err != nil {
		t.Fatalf("stat log path: %v", err)
	}
}

func TestShellStartWritesLog(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()

	_, err := runTestAppWithState(t, stateDir, "start", "--agent", "shell", "--dir", workDir, "printf shell-output")
	if err != nil {
		t.Fatalf("run shell start: %v", err)
	}
	tasks, err := NewJSONTaskStore(stateDir).List()
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %#v, want one task", tasks)
	}
	assertEventually(t, func() bool {
		data, err := os.ReadFile(tasks[0].LogPath)
		return err == nil && strings.Contains(string(data), "shell-output")
	})
}

func runTestAppWithStateAndAdapters(t *testing.T, stateDir string, adapters *AdapterRegistry, args ...string) (string, error) {
	t.Helper()
	return runTestAppWithStateAndFactory(t, stateDir, func(d dirs) (*Deck, error) {
		ac, err := New(d)
		if err != nil {
			return nil, err
		}
		ac.Adapters = adapters
		return ac, nil
	}, args...)
}

func assertEventually(t *testing.T, fn func() bool) {
	t.Helper()
	for range 50 {
		if fn() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition did not become true")
}
