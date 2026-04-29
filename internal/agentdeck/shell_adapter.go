package agentdeck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ShellAdapter struct{}

func (a ShellAdapter) Name() string {
	return "shell"
}

func (a ShellAdapter) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	logFile, err := os.OpenFile(req.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", req.Prompt)
	cmd.Dir = req.CWD
	cmd.Env = req.Env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	_ = logFile.Close()

	return &StartResult{PID: cmd.Process.Pid}, nil
}

func (a ShellAdapter) ResumeCommand(task Task) (*exec.Cmd, error) {
	return nil, errors.New("shell adapter does not support resume")
}

func (a ShellAdapter) Status(ctx context.Context, task Task) (AdapterStatus, error) {
	return AdapterStatus{Status: refreshTaskStatus(task).Status}, nil
}

func (a ShellAdapter) Kill(ctx context.Context, task Task) error {
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
