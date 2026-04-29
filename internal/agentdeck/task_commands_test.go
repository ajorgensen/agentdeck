package agentdeck

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListNoTasks(t *testing.T) {
	out, err := runTestApp(t, "list")
	if err != nil {
		t.Fatalf("run list: %v", err)
	}
	if !strings.Contains(out, "No tasks found") {
		t.Fatalf("list output = %q, want empty message", out)
	}
}

func TestListJSONAndFilters(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	store := NewJSONTaskStore(stateDir)
	for _, task := range []Task{
		{ID: "keep", Agent: "claude", Status: TaskStatusRunning, CWD: workDir, Prompt: "keep prompt"},
		{ID: "skip", Agent: "codex", Status: TaskStatusSucceeded, CWD: t.TempDir(), Prompt: "skip prompt"},
	} {
		if err := store.Put(task); err != nil {
			t.Fatalf("put task %q: %v", task.ID, err)
		}
	}

	out, err := runTestAppWithState(t, stateDir, "list", "--json", "--agent", "claude", "--status", "running", "--dir", workDir)
	if err != nil {
		t.Fatalf("run list json: %v", err)
	}
	var tasks []Task
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("decode json %q: %v", out, err)
	}
	if len(tasks) != 1 || tasks[0].ID != "keep" {
		t.Fatalf("tasks = %#v, want filtered keep task", tasks)
	}
}

func TestStatusJSON(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded, CWD: t.TempDir(), Prompt: "Fix bug"}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "status", "--json", task.ID)
	if err != nil {
		t.Fatalf("run status json: %v", err)
	}
	var got Task
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode json %q: %v", out, err)
	}
	if got.ID != task.ID || got.Agent != task.Agent {
		t.Fatalf("task = %#v", got)
	}
}

func TestListMultipleTasks(t *testing.T) {
	stateDir := t.TempDir()
	store := NewJSONTaskStore(stateDir)
	created := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	for _, task := range []Task{
		{ID: "older", Agent: "claude", Status: TaskStatusSucceeded, CWD: "/tmp/a", Prompt: "short prompt", CreatedAt: created, UpdatedAt: created},
		{ID: "newer", Agent: "codex", PID: os.Getpid(), Status: TaskStatusRunning, CWD: "/tmp/b", Prompt: "this is a very long prompt that should be truncated in list output", CreatedAt: created.Add(time.Minute), UpdatedAt: created.Add(time.Minute)},
	} {
		if err := store.Put(task); err != nil {
			t.Fatalf("put task %q: %v", task.ID, err)
		}
	}

	out, err := runTestAppWithState(t, stateDir, "list")
	if err != nil {
		t.Fatalf("run list: %v", err)
	}
	for _, want := range []string{"TASK ID", "AGENT", "STATUS", "CREATED", "UPDATED", "CWD", "PROMPT", "older", "claude", "succeeded", "2026-04-28 12:00", "newer", "codex", "running"} {
		if !strings.Contains(out, want) {
			t.Fatalf("list output = %q, want %q", out, want)
		}
	}
	if strings.Contains(out, "this is a very long prompt that should be truncated in list output") {
		t.Fatalf("list output did not truncate long prompt: %q", out)
	}
}

func TestStatusUnknownTaskID(t *testing.T) {
	_, err := runTestApp(t, "status", "missing")
	if err == nil {
		t.Fatal("status missing task returned nil error")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Fatalf("error = %v, want task not found", err)
	}
}

func TestStatusAcceptsUniqueTaskIDPrefix(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123def0", Agent: "claude", Status: TaskStatusSucceeded, CWD: "/tmp/work", Prompt: "Fix bug"}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}
	if err := NewJSONTaskStore(stateDir).Put(Task{ID: "def123abc0", Agent: "codex", Status: TaskStatusSucceeded}); err != nil {
		t.Fatalf("put second task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "status", "abc")
	if err != nil {
		t.Fatalf("status by prefix: %v", err)
	}
	if !strings.Contains(out, "id: abc123def0") {
		t.Fatalf("status output = %q, want full resolved id", out)
	}
}

func TestStatusTaskDetails(t *testing.T) {
	stateDir := t.TempDir()
	store := NewJSONTaskStore(stateDir)
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded, CWD: "/tmp/work", Prompt: "Fix bug", LogPath: filepath.Join(stateDir, "tasks", "abc123.log")}
	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	out, err := runTestAppWithState(t, stateDir, "status", task.ID)
	if err != nil {
		t.Fatalf("run status: %v", err)
	}
	for _, want := range []string{"id: abc123", "agent: claude", "status: succeeded", "cwd: /tmp/work", "prompt: Fix bug", "log: " + task.LogPath} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output = %q, want %q", out, want)
		}
	}
}

func runTestApp(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runTestAppWithState(t, filepath.Join(t.TempDir(), "state"), args...)
}

func runTestAppWithState(t *testing.T, stateDir string, args ...string) (string, error) {
	t.Helper()
	return runTestAppWithStateAndFactory(t, stateDir, New, args...)
}

func runTestAppWithStateAndFactory(t *testing.T, stateDir string, newDeck func(dirs) (*Deck, error), args ...string) (string, error) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv(configDirEnvVar, filepath.Join(tmp, "config"))
	t.Setenv(dataDirEnvVar, filepath.Join(tmp, "data"))
	t.Setenv(stateDirEnvVar, stateDir)
	t.Setenv(runtimeDirEnvVar, filepath.Join(tmp, "runtime"))

	var out bytes.Buffer
	app := newAppWithFactory(newDeck)
	app.Writer = &out
	app.ErrWriter = &out

	argv := append([]string{"agentdeck"}, args...)
	err := app.Run(argv)
	return out.String(), err
}
