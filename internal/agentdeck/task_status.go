package agentdeck

import (
	"errors"
	"os"
	"syscall"
)

func refreshTaskStatus(task Task) Task {
	if task.Status != TaskStatusRunning {
		return task
	}
	if task.PID <= 0 || !processAlive(task.PID) {
		task.Status = TaskStatusUnknown
	}
	return task
}

func processAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, os.ErrPermission)
}
