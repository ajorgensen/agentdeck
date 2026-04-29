package agentdeck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type CodexAdapter struct{}

func (a CodexAdapter) Name() string {
	return "codex"
}

func (a CodexAdapter) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	logFile, err := os.OpenFile(req.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "codex", "exec", "--json", "--cd", req.CWD, "--skip-git-repo-check", req.Prompt)
	cmd.Dir = req.CWD
	cmd.Env = req.Env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	_ = logFile.Close()

	threadID := waitForLogMatch(req.LogPath, codexThreadIDPattern, 2*time.Second)
	return &StartResult{PID: cmd.Process.Pid, NativeSessionID: threadID}, nil
}

func (a CodexAdapter) ResumeCommand(task Task) (*exec.Cmd, error) {
	args := []string{"resume", "--include-non-interactive"}
	if task.NativeSessionID != "" {
		args = append(args, task.NativeSessionID)
	} else {
		args = append(args, "--last")
	}
	cmd := exec.Command("codex", args...)
	cmd.Dir = task.CWD
	return cmd, nil
}

func (a CodexAdapter) Status(ctx context.Context, task Task) (AdapterStatus, error) {
	return AdapterStatus{Status: refreshTaskStatus(task).Status}, nil
}

func (a CodexAdapter) Kill(ctx context.Context, task Task) error {
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
