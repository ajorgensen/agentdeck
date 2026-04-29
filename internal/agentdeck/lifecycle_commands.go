package agentdeck

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
)

func killCommand() *cli.Command {
	return &cli.Command{
		Name:      "kill",
		Usage:     "stop a tracked task",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "force", Usage: "forcefully kill the process"},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return fmt.Errorf("kill requires exactly one task id")
			}

			ac := fromContext(ctx)
			task, err := ac.Store.Get(ctx.Args().First())
			if err != nil {
				return err
			}
			refreshed := refreshTaskStatus(task)
			if refreshed.Status == TaskStatusRunning {
				if ctx.Bool("force") {
					err = forceKillProcess(refreshed.PID)
				} else {
					var adapter AgentAdapter
					adapter, err = ac.Adapters.Get(refreshed.Agent)
					if err == nil {
						err = adapter.Kill(ctx.Context, refreshed)
					}
				}
				if err != nil {
					return err
				}
			}

			if err := ac.Store.Update(task.ID, func(task *Task) error {
				task.Status = TaskStatusKilled
				task.UpdatedAt = time.Now().UTC()
				return nil
			}); err != nil {
				return err
			}
			fmt.Fprintf(ctx.App.Writer, "Killed task %s\n", task.ID)
			return nil
		},
	}
}

func forgetCommand() *cli.Command {
	return &cli.Command{
		Name:      "forget",
		Usage:     "remove a task from the local index",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "force", Usage: "forget even if the task appears to be running"},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return fmt.Errorf("forget requires exactly one task id")
			}

			ac := fromContext(ctx)
			task, err := ac.Store.Get(ctx.Args().First())
			if err != nil {
				return err
			}
			if !ctx.Bool("force") && refreshTaskStatus(task).Status == TaskStatusRunning {
				return fmt.Errorf("task %s is running; use --force to forget it", task.ID)
			}
			if err := ac.Store.Delete(task.ID); err != nil {
				return err
			}
			fmt.Fprintf(ctx.App.Writer, "Forgot task %s\n", task.ID)
			return nil
		},
	}
}

func forceKillProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Kill()
	if isNoSuchProcess(err) {
		return nil
	}
	return err
}

func isNoSuchProcess(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
}
