package agentdeck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJSONTaskStoreListAbsentFile(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())

	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list absent store: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("tasks = %#v, want empty", tasks)
	}
}

func TestJSONTaskStoreSaveAndReloadTasks(t *testing.T) {
	stateDir := t.TempDir()
	store := NewJSONTaskStore(stateDir)
	now := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	task := Task{
		ID:        "abc123",
		Agent:     "claude",
		CWD:       "/tmp/work",
		Prompt:    "Fix the login redirect bug",
		PID:       12345,
		Status:    TaskStatusRunning,
		LogPath:   filepath.Join(stateDir, "tasks", "abc123.log"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	reloaded := NewJSONTaskStore(stateDir)
	got, err := reloaded.Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.ID != task.ID || got.Agent != task.Agent || got.Status != task.Status || !got.CreatedAt.Equal(now) {
		t.Fatalf("task = %#v, want %#v", got, task)
	}
}

func TestJSONTaskStoreMalformedJSON(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stateDir, tasksFileName), []byte("not-json"), 0o644); err != nil {
		t.Fatalf("write malformed store: %v", err)
	}

	store := NewJSONTaskStore(stateDir)
	if _, err := store.List(); err == nil {
		t.Fatal("list malformed store returned nil error")
	}
}

func TestJSONTaskStoreDeleteTask(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusSucceeded}

	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}
	if err := store.Delete(task.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if _, err := store.Get(task.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("get deleted task error = %v, want ErrTaskNotFound", err)
	}
}

func TestJSONTaskStoreUpdateTask(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())
	task := Task{ID: "abc123", Agent: "claude", Status: TaskStatusPending}

	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}
	if err := store.Update(task.ID, func(task *Task) error {
		task.Status = TaskStatusRunning
		return nil
	}); err != nil {
		t.Fatalf("update task: %v", err)
	}

	got, err := store.Get(task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Status != TaskStatusRunning {
		t.Fatalf("status = %q, want %q", got.Status, TaskStatusRunning)
	}
}

func TestJSONTaskStoreGetUniqueIDPrefix(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())
	task := Task{ID: "abc123def0", Agent: "claude", Status: TaskStatusSucceeded}
	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}
	if err := store.Put(Task{ID: "def123abc0", Agent: "codex", Status: TaskStatusSucceeded}); err != nil {
		t.Fatalf("put second task: %v", err)
	}

	got, err := store.Get("abc")
	if err != nil {
		t.Fatalf("get task by prefix: %v", err)
	}
	if got.ID != task.ID {
		t.Fatalf("task id = %q, want %q", got.ID, task.ID)
	}
}

func TestJSONTaskStoreGetAmbiguousIDPrefix(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())
	for _, id := range []string{"abc123def0", "abc456def0"} {
		if err := store.Put(Task{ID: id, Agent: "claude", Status: TaskStatusSucceeded}); err != nil {
			t.Fatalf("put task %q: %v", id, err)
		}
	}

	_, err := store.Get("abc")
	if !errors.Is(err, ErrTaskIDAmbiguous) {
		t.Fatalf("error = %v, want ErrTaskIDAmbiguous", err)
	}
	for _, want := range []string{"abc", "abc123def0", "abc456def0"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want %q", err.Error(), want)
		}
	}
}

func TestJSONTaskStoreDeleteUniqueIDPrefix(t *testing.T) {
	store := NewJSONTaskStore(t.TempDir())
	task := Task{ID: "abc123def0", Agent: "claude", Status: TaskStatusSucceeded}
	if err := store.Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	if err := store.Delete("abc"); err != nil {
		t.Fatalf("delete by prefix: %v", err)
	}
	if _, err := store.Get(task.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("get deleted task error = %v, want ErrTaskNotFound", err)
	}
}

func TestNewTaskIDIsShortAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		id, err := NewTaskID()
		if err != nil {
			t.Fatalf("new task id: %v", err)
		}
		if len(id) != taskIDLength {
			t.Fatalf("id length = %d, want %d", len(id), taskIDLength)
		}
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
}
