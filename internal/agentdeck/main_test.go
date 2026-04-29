package agentdeck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/urfave/cli/v2"
)

func TestDefaultDirectoryFlagsUseAgentctlPathsAndEnvVars(t *testing.T) {
	app := newApp()

	if app.Name != appName {
		t.Fatalf("app name = %q, want %q", app.Name, appName)
	}

	tests := []struct {
		name    string
		value   string
		envName string
	}{
		{"config-dir", filepath.Join(xdg.ConfigHome, appName), configDirEnvVar},
		{"data-dir", filepath.Join(xdg.DataHome, appName), dataDirEnvVar},
		{"state-dir", filepath.Join(xdg.StateHome, appName), stateDirEnvVar},
		{"runtime-dir", defaultRuntimeDir(), runtimeDirEnvVar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := stringFlag(t, app, tt.name)
			if flag.Value != tt.value {
				t.Fatalf("default value = %q, want %q", flag.Value, tt.value)
			}
			if len(flag.EnvVars) != 1 || flag.EnvVars[0] != tt.envName {
				t.Fatalf("env vars = %#v, want [%q]", flag.EnvVars, tt.envName)
			}
			for _, env := range flag.EnvVars {
				if strings.Contains(env, "AGENT_CONTROL") {
					t.Fatalf("env var %q still uses old AGENT_CONTROL prefix", env)
				}
			}
		})
	}
}

func TestStatusUsesAgentctlEnvDirectoryOverrides(t *testing.T) {
	tmp := t.TempDir()
	dirs := map[string]string{
		configDirEnvVar:  filepath.Join(tmp, "config"),
		dataDirEnvVar:    filepath.Join(tmp, "data"),
		stateDirEnvVar:   filepath.Join(tmp, "state"),
		runtimeDirEnvVar: filepath.Join(tmp, "runtime"),
	}

	for env, path := range dirs {
		t.Setenv(env, path)
	}

	var out bytes.Buffer
	app := newApp()
	app.Writer = &out
	app.ErrWriter = &out

	if err := app.Run([]string{"agentdeck", "status"}); err != nil {
		t.Fatalf("run status: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "agentdeck is running") {
		t.Fatalf("status output = %q, want agentdeck identity", got)
	}
	if strings.Contains(got, "agent-control") {
		t.Fatalf("status output still contains old identity: %q", got)
	}

	for _, path := range dirs {
		if !strings.Contains(got, path) {
			t.Fatalf("status output = %q, want path %q", got, path)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %q: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q was not created as a directory", path)
		}
	}
}

func stringFlag(t *testing.T, app *cli.App, name string) *cli.StringFlag {
	t.Helper()

	for _, flag := range app.Flags {
		for _, flagName := range flag.Names() {
			if flagName == name {
				stringFlag, ok := flag.(*cli.StringFlag)
				if !ok {
					t.Fatalf("flag %q has type %T, want *cli.StringFlag", name, flag)
				}
				return stringFlag
			}
		}
	}

	t.Fatalf("missing flag %q", name)
	return nil
}
