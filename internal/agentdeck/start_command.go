package agentdeck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func startCommand() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "start an agent task in the background",
		ArgsUsage: "<prompt>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "agent", Usage: "agent adapter to use", Required: true},
			&cli.StringFlag{Name: "dir", Usage: "working directory for the task", Value: "."},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() == 0 {
				return fmt.Errorf("start requires a prompt")
			}

			ac := fromContext(ctx)
			cwd, err := validateTaskDir(ctx.String("dir"))
			if err != nil {
				return err
			}

			adapter, err := ac.Adapters.Get(ctx.String("agent"))
			if err != nil {
				return err
			}

			id, err := NewTaskID()
			if err != nil {
				return fmt.Errorf("generate task id: %w", err)
			}

			logPath := TaskLogPath(ac.Dirs.State, id)
			if err := createLogFile(logPath); err != nil {
				return err
			}

			now := time.Now().UTC()
			task := Task{
				ID:        id,
				Agent:     adapter.Name(),
				CWD:       cwd,
				Prompt:    strings.Join(ctx.Args().Slice(), " "),
				Status:    TaskStatusPending,
				LogPath:   logPath,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := ac.Store.Put(task); err != nil {
				return err
			}

			result, err := adapter.Start(ctx.Context, StartRequest{
				CWD:     task.CWD,
				Prompt:  task.Prompt,
				LogPath: task.LogPath,
				Env:     os.Environ(),
				Config:  ac.Config,
			})
			if err != nil {
				updateErr := ac.Store.Update(task.ID, func(task *Task) error {
					task.Status = TaskStatusFailed
					task.UpdatedAt = time.Now().UTC()
					return nil
				})
				if updateErr != nil {
					return fmt.Errorf("start task: %w; additionally failed to mark task failed: %v", err, updateErr)
				}
				return fmt.Errorf("start task: %w", err)
			}

			if result == nil {
				result = &StartResult{}
			}
			if err := ac.Store.Update(task.ID, func(task *Task) error {
				task.PID = result.PID
				task.NativeSessionID = result.NativeSessionID
				task.Status = TaskStatusRunning
				task.UpdatedAt = time.Now().UTC()
				return nil
			}); err != nil {
				return err
			}

			fmt.Fprintf(ctx.App.Writer, "Started task %s\n", task.ID)
			fmt.Fprintf(ctx.App.Writer, "Next: agentdeck log %s or agentdeck status %s\n", task.ID, task.ID)
			return nil
		},
	}
}

func validateTaskDir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve dir %q: %w", path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("dir %q is not a directory: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("dir %q is not a directory", path)
	}
	return abs, nil
}

func createLogFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create log file %q: %w", path, err)
	}
	return file.Close()
}
