package agentdeck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type OpenCodeAdapter struct{}

func (a OpenCodeAdapter) Name() string {
	return "opencode"
}

func (a OpenCodeAdapter) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	logFile, err := os.OpenFile(req.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "opencode", "run", "--format", "json", "--dir", req.CWD, req.Prompt)
	cmd.Dir = req.CWD
	cmd.Env = req.Env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	_ = logFile.Close()

	sessionID := waitForLogMatch(req.LogPath, openCodeSessionIDPattern, 2*time.Second)
	return &StartResult{PID: cmd.Process.Pid, NativeSessionID: sessionID}, nil
}

func (a OpenCodeAdapter) ResumeCommand(task Task) (*exec.Cmd, error) {
	args := []string{}
	if task.NativeSessionID != "" {
		args = append(args, "--session", task.NativeSessionID)
	} else {
		args = append(args, "--continue")
	}
	if task.CWD != "" {
		args = append(args, task.CWD)
	}
	cmd := exec.Command("opencode", args...)
	cmd.Dir = task.CWD
	return cmd, nil
}

func (a OpenCodeAdapter) Status(ctx context.Context, task Task) (AdapterStatus, error) {
	return AdapterStatus{Status: refreshTaskStatus(task).Status}, nil
}

func (a OpenCodeAdapter) Kill(ctx context.Context, task Task) error {
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
