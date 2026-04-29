package agentdeck

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFirstLogMatchCodexThreadID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "codex.log")
	data := []byte(`{"type":"thread.started","thread_id":"019dd67b-ce2e-78a0-b193-29f36f57a88c"}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	got := firstLogMatch(path, codexThreadIDPattern)
	if got != "019dd67b-ce2e-78a0-b193-29f36f57a88c" {
		t.Fatalf("thread id = %q", got)
	}
}

func TestFirstLogMatchOpenCodeSessionID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.log")
	data := []byte(`{"type":"step_start","sessionID":"ses_22982ecc2ffebj2ko80WA2FEjS"}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	got := firstLogMatch(path, openCodeSessionIDPattern)
	if got != "ses_22982ecc2ffebj2ko80WA2FEjS" {
		t.Fatalf("session id = %q", got)
	}
}

func TestWaitForLogMatchTimesOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.log")
	got := waitForLogMatch(path, codexThreadIDPattern, time.Millisecond)
	if got != "" {
		t.Fatalf("match = %q, want empty", got)
	}
}
