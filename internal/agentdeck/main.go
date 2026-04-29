package agentdeck

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/urfave/cli/v2"
)

// Environment variables for overriding directory locations.
const (
	configDirEnvVar  = "AGENTDECK_CONFIG_DIR"
	dataDirEnvVar    = "AGENTDECK_DATA_DIR"
	stateDirEnvVar   = "AGENTDECK_STATE_DIR"
	runtimeDirEnvVar = "AGENTDECK_RUNTIME_DIR"

	metadataKey = "agentdeck"
	appName     = "agentdeck"
)

// dirs holds the resolved XDG-compliant directories for the application.
type dirs struct {
	Config  string // user preferences, e.g. config.toml
	Data    string // persistent data: tracked folders registry
	State   string // recoverable state: task history, logs
	Runtime string // ephemeral: PIDs, sockets, lock files
}

// NewApp builds the CLI application.
func NewApp() *cli.App {
	return newAppWithFactory(New)
}

func newApp() *cli.App { return NewApp() }

func newAppWithFactory(newDeck func(dirs) (*Deck, error)) *cli.App {
	app := &cli.App{
		Name:                 appName,
		Usage:                "manage coding-agent tasks",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-dir",
				Usage:   "directory for configuration files",
				EnvVars: []string{configDirEnvVar},
				Value:   filepath.Join(xdg.ConfigHome, appName),
			},
			&cli.StringFlag{
				Name:    "data-dir",
				Usage:   "directory for persistent data (tracked folders, agent definitions)",
				EnvVars: []string{dataDirEnvVar},
				Value:   filepath.Join(xdg.DataHome, appName),
			},
			&cli.StringFlag{
				Name:    "state-dir",
				Usage:   "directory for state data (task history, logs)",
				EnvVars: []string{stateDirEnvVar},
				Value:   filepath.Join(xdg.StateHome, appName),
			},
			&cli.StringFlag{
				Name:    "runtime-dir",
				Usage:   "directory for runtime data (PIDs, sockets), cleared on reboot",
				EnvVars: []string{runtimeDirEnvVar},
				Value:   defaultRuntimeDir(),
			},
		},
		Before: func(ctx *cli.Context) error {
			ac, err := newDeck(dirsFromContext(ctx))
			if err != nil {
				return err
			}

			app := ctx.App
			if app.Metadata == nil {
				app.Metadata = map[string]any{}
			}

			app.Metadata[metadataKey] = ac

			return nil
		},
		Commands: []*cli.Command{
			startCommand(),
			listCommand(),
			statusCommand(),
			logCommand(),
			tailCommand(),
			resumeCommand(),
			killCommand(),
			forgetCommand(),
			pathsCommand(),
		},
	}

	return app
}

// dirsFromContext extracts the resolved directories from the CLI context.
func dirsFromContext(ctx *cli.Context) dirs {
	return dirs{
		Config:  ctx.String("config-dir"),
		Data:    ctx.String("data-dir"),
		State:   ctx.String("state-dir"),
		Runtime: ctx.String("runtime-dir"),
	}
}
