package agentdeck

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenCodeAdapterResumeCommandWithSession(t *testing.T) {
	task := Task{CWD: t.TempDir(), NativeSessionID: "ses_22982ecc2ffebj2ko80WA2FEjS"}
	cmd, err := OpenCodeAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if filepath.Base(cmd.Path) != "opencode" || strings.Join(cmd.Args, " ") != "opencode --session ses_22982ecc2ffebj2ko80WA2FEjS "+task.CWD {
		t.Fatalf("command = %#v", cmd)
	}
	if cmd.Dir != task.CWD {
		t.Fatalf("command dir = %q, want %q", cmd.Dir, task.CWD)
	}
}

func TestOpenCodeAdapterResumeCommandWithoutSessionUsesContinue(t *testing.T) {
	task := Task{CWD: t.TempDir()}
	cmd, err := OpenCodeAdapter{}.ResumeCommand(task)
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if strings.Join(cmd.Args, " ") != "opencode --continue "+task.CWD {
		t.Fatalf("command args = %#v", cmd.Args)
	}
}

func TestOpenCodeAdapterStartIntegration(t *testing.T) {
	if os.Getenv("AGENTDECK_INTEGRATION") != "1" {
		t.Skip("set AGENTDECK_INTEGRATION=1 to run OpenCode CLI integration test")
	}
	if _, err := exec.LookPath("opencode"); err != nil {
		t.Skip("opencode CLI not installed")
	}

	stateDir := t.TempDir()
	logPath := filepath.Join(stateDir, "opencode.log")
	result, err := OpenCodeAdapter{}.Start(context.Background(), StartRequest{
		CWD:     t.TempDir(),
		Prompt:  "Reply with exactly: ok",
		LogPath: logPath,
		Env:     os.Environ(),
	})
	if err != nil {
		t.Fatalf("start opencode: %v", err)
	}
	if result.PID == 0 {
		t.Fatalf("result = %#v", result)
	}
}
