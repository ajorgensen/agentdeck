//go:build !unix

package agentdeck

import "os/exec"

func replaceCurrentProcessWithExec(cmd *exec.Cmd) error {
	return errProcessReplaceUnsupported
}
