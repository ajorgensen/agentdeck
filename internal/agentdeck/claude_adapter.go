package agentdeck

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ClaudeAdapter struct{}

func (a ClaudeAdapter) Name() string {
	return "claude"
}

func (a ClaudeAdapter) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	sessionID, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("generate claude session id: %w", err)
	}

	logFile, err := os.OpenFile(req.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "claude", "-p", "--output-format", "json", "--session-id", sessionID, req.Prompt)
	cmd.Dir = req.CWD
	cmd.Env = req.Env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	_ = logFile.Close()

	return &StartResult{PID: cmd.Process.Pid, NativeSessionID: sessionID}, nil
}

func (a ClaudeAdapter) ResumeCommand(task Task) (*exec.Cmd, error) {
	args := []string{"--resume"}
	if task.NativeSessionID != "" {
		args = append(args, task.NativeSessionID)
	}
	cmd := exec.Command("claude", args...)
	cmd.Dir = task.CWD
	return cmd, nil
}

func (a ClaudeAdapter) Status(ctx context.Context, task Task) (AdapterStatus, error) {
	return AdapterStatus{Status: refreshTaskStatus(task).Status}, nil
}

func (a ClaudeAdapter) Kill(ctx context.Context, task Task) error {
	if task.PID <= 0 {
		return nil
	}
	process, err := os.FindProcess(task.PID)
	if err != nil {
		return err
	}
	err = process.Signal(syscall.SIGTERM)
	if isNoSuchProcess(err) {
		return nil
	}
	return err
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
