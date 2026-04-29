package agentdeck

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexAdapterResumeCommandWithSession(t *testing.T) {
	task := Task{CWD: t.TempDir(), NativeSessionID: "019dd67b-ce2e-78a0-b193-29f36f57a88c"}
	cmd, err := CodexAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if filepath.Base(cmd.Path) != "codex" || strings.Join(cmd.Args, " ") != "codex resume --include-non-interactive 019dd67b-ce2e-78a0-b193-29f36f57a88c" {
		t.Fatalf("command = %#v", cmd)
	}
	if cmd.Dir != task.CWD {
		t.Fatalf("command dir = %q, want %q", cmd.Dir, task.CWD)
	}
}

func TestCodexAdapterResumeCommandWithoutSessionUsesLastNonInteractive(t *testing.T) {
	task := Task{CWD: t.TempDir()}
	cmd, err := CodexAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if strings.Join(cmd.Args, " ") != "codex resume --include-non-interactive --last" {
		t.Fatalf("command args = %#v", cmd.Args)
	}
}

func TestCodexAdapterStartIntegration(t *testing.T) {
	if os.Getenv("AGENTDECK_INTEGRATION") != "1" {
		t.Skip("set AGENTDECK_INTEGRATION=1 to run Codex CLI integration test")
	}
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex CLI not installed")
	}

	stateDir := t.TempDir()
	logPath := filepath.Join(stateDir, "codex.log")
	result, err := CodexAdapter{}.Start(context.Background(), StartRequest{
		CWD:     t.TempDir(),
		Prompt:  "Reply with exactly: ok",
		LogPath: logPath,
		Env:     os.Environ(),
	})
	if err != nil {
		t.Fatalf("start codex: %v", err)
	}
	if result.PID == 0 {
		t.Fatalf("result = %#v", result)
	}
}
