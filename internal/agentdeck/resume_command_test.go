package agentdeck

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildResumeCommandThroughAdapter(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "fake", CWD: workDir, NativeSessionID: "native-1"}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}
	ac := &Deck{
		Store: NewJSONTaskStore(stateDir),
		Adapters: NewAdapterRegistry(FakeAdapter{
			NameValue:       "fake",
			ResumeArgsValue: []string{"fake-agent", "resume", "native-1"},
		}),
	}

	cmd, err := buildResumeCommand(ac, task.ID)
	if err != nil {
		t.Fatalf("build resume command: %v", err)
	}
	if cmd.Path != "fake-agent" || strings.Join(cmd.Args, " ") != "fake-agent resume native-1" {
		t.Fatalf("command args = %#v, path = %q", cmd.Args, cmd.Path)
	}
	if cmd.Dir != workDir {
		t.Fatalf("command dir = %q, want %q", cmd.Dir, workDir)
	}
}

func TestResumeMissingTaskID(t *testing.T) {
	_, err := runTestApp(t, "resume", "missing")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("error = %v, want ErrTaskNotFound", err)
	}
}

func TestResumeUnsupportedAgent(t *testing.T) {
	stateDir := t.TempDir()
	task := Task{ID: "abc123", Agent: "missing", CWD: filepath.Join(t.TempDir(), "work")}
	if err := NewJSONTaskStore(stateDir).Put(task); err != nil {
		t.Fatalf("put task: %v", err)
	}

	_, err := runTestAppWithState(t, stateDir, "resume", task.ID)
	if !errors.Is(err, ErrUnsupportedAgent) {
		t.Fatalf("error = %v, want ErrUnsupportedAgent", err)
	}
}

func TestResumeRequiresOneTaskID(t *testing.T) {
	_, err := runTestApp(t, "resume")
	if err == nil || !strings.Contains(err.Error(), "resume requires exactly one task id") {
		t.Fatalf("error = %v, want arity error", err)
	}
}

func TestRunResumeCommandPrefersProcessReplacement(t *testing.T) {
	oldReplace := replaceCurrentProcess
	t.Cleanup(func() { replaceCurrentProcess = oldReplace })

	var replaced *exec.Cmd
	replaceCurrentProcess = func(cmd *exec.Cmd) error {
		replaced = cmd
		return nil
	}

	cmd := exec.Command("fake-agent", "resume")
	if err := runResumeCommand(cmd); err != nil {
		t.Fatalf("run resume command: %v", err)
	}
	if replaced != cmd {
		t.Fatalf("replaced command = %p, want %p", replaced, cmd)
	}
	if cmd.Stdin == nil || cmd.Stdout == nil || cmd.Stderr == nil {
		t.Fatalf("stdio was not inherited")
	}
}
