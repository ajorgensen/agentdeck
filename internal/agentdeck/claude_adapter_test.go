package agentdeck

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeAdapterResumeCommandWithSession(t *testing.T) {
	task := Task{CWD: t.TempDir(), NativeSessionID: "00000000-0000-4000-8000-000000000001"}
	cmd, err := ClaudeAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if filepath.Base(cmd.Path) != "claude" || strings.Join(cmd.Args, " ") != "claude --resume 00000000-0000-4000-8000-000000000001" {
		t.Fatalf("command = %#v", cmd)
	}
	if cmd.Dir != task.CWD {
		t.Fatalf("command dir = %q, want %q", cmd.Dir, task.CWD)
	}
}

func TestClaudeAdapterResumeCommandWithoutSessionUsesPicker(t *testing.T) {
	task := Task{CWD: t.TempDir()}
	cmd, err := ClaudeAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if strings.Join(cmd.Args, " ") != "claude --resume" {
		t.Fatalf("command args = %#v", cmd.Args)
	}
}

func TestNewUUIDShape(t *testing.T) {
	id, err := newUUID()
	if err != nil {
		t.Fatalf("new uuid: %v", err)
	}
	if len(id) != 36 || id[14] != '4' {
		t.Fatalf("uuid = %q, want v4 shape", id)
	}
}

func TestClaudeAdapterStartIntegration(t *testing.T) {
	if os.Getenv("AGENTDECK_INTEGRATION") != "1" {
		t.Skip("set AGENTDECK_INTEGRATION=1 to run Claude CLI integration test")
	}
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not installed")
	}

	stateDir := t.TempDir()
	logPath := filepath.Join(stateDir, "claude.log")
	result, err := ClaudeAdapter{}.Start(context.Background(), StartRequest{
		CWD:     t.TempDir(),
		Prompt:  "Reply with exactly: ok",
		LogPath: logPath,
		Env:     os.Environ(),
	})
	if err != nil {
		t.Fatalf("start claude: %v", err)
	}
	if result.PID == 0 || result.NativeSessionID == "" {
		t.Fatalf("result = %#v", result)
	}
}
