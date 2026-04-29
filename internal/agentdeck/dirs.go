package agentdeck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// defaultRuntimeDir returns a sensible runtime directory. xdg.RuntimeDir is
// typically unset on macOS; fall back to a per-user directory under the OS
// temp dir so semantics ("cleared on reboot") still hold.
func defaultRuntimeDir() string {
	if xdg.RuntimeDir != "" {
		return filepath.Join(xdg.RuntimeDir, appName)
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", appName, os.Getuid()))
}

// ensureDirs creates each directory if it does not exist.
func ensureDirs(d dirs) error {
	for _, dir := range []struct {
		name string
		path string
		mode os.FileMode
	}{
		{"config", d.Config, 0o755},
		{"data", d.Data, 0o755},
		{"state", d.State, 0o755},
		// Runtime dir often holds sockets / PID files; tighter perms.
		{"runtime", d.Runtime, 0o700},
	} {
		if err := ensureDir(dir.name, dir.path, dir.mode); err != nil {
			return err
		}
	}
	return nil
}

func ensureDir(name, path string, mode os.FileMode) error {
	if path == "" {
		return fmt.Errorf("%s directory cannot be empty", name)
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s path %q exists but is not a directory", name, path)
		}
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check %s directory %q: %w", name, path, err)
	}

	if err := os.MkdirAll(path, mode); err != nil {
		return fmt.Errorf("create %s directory %q: %w", name, path, err)
	}

	return nil
}
