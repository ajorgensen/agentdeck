//go:build unix

package agentdeck

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func replaceCurrentProcessWithExec(cmd *exec.Cmd) error {
	if cmd.Path == "" {
		return errors.New("resume command path is empty")
	}

	path := cmd.Path
	if !filepath.IsAbs(path) {
		resolved, err := exec.LookPath(path)
		if err != nil {
			return err
		}
		path = resolved
	}

	if cmd.Dir != "" {
		if err := os.Chdir(cmd.Dir); err != nil {
			return err
		}
	}

	args := cmd.Args
	if len(args) == 0 {
		args = []string{cmd.Path}
	}

	env := cmd.Env
	if env == nil {
		env = os.Environ()
	}

	return syscall.Exec(path, args, env)
}
