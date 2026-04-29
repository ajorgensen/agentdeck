package agentdeck

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var errProcessReplaceUnsupported = errors.New("process replacement unsupported")

var replaceCurrentProcess = replaceCurrentProcessWithExec

func resumeCommand() *cli.Command {
	return &cli.Command{
		Name:      "resume",
		Usage:     "resume a task in the native agent CLI",
		ArgsUsage: "<task-id>",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return fmt.Errorf("resume requires exactly one task id")
			}

			cmd, err := buildResumeCommand(fromContext(ctx), ctx.Args().First())
			if err != nil {
				return err
			}
			return runResumeCommand(cmd)
		},
	}
}

func buildResumeCommand(ac *Deck, taskID string) (*exec.Cmd, error) {
	task, err := ac.Store.Get(taskID)
	if err != nil {
		return nil, err
	}
	adapter, err := ac.Adapters.Get(task.Agent)
	if err != nil {
		return nil, err
	}
	cmd, err := adapter.ResumeCommand(task)
	if err != nil {
		return nil, err
	}
	if cmd.Dir == "" {
		cmd.Dir = task.CWD
	}
	return cmd, nil
}

func runResumeCommand(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := replaceCurrentProcess(cmd); err != nil {
		if errors.Is(err, errProcessReplaceUnsupported) {
			return cmd.Run()
		}
		return err
	}
	return nil
}
