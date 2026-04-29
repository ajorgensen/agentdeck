package agentdeck

import "github.com/urfave/cli/v2"

// Deck is the top-level application context. It owns the resolved
// XDG directories and loaded configuration, and is the value passed to
// command actions via the CLI context.
type Deck struct {
	Dirs     dirs
	Config   *Config
	Store    TaskStore
	Adapters *AdapterRegistry
}

// New constructs an Deck by ensuring the given directories exist and
// loading the config file from Dirs.Config.
func New(d dirs) (*Deck, error) {
	if err := ensureDirs(d); err != nil {
		return nil, err
	}

	cfg, err := LoadConfig(d.Config)
	if err != nil {
		return nil, err
	}

	return &Deck{
		Dirs:     d,
		Config:   cfg,
		Store:    NewJSONTaskStore(d.State),
		Adapters: defaultAdapterRegistry(),
	}, nil
}

func defaultAdapterRegistry() *AdapterRegistry {
	return NewAdapterRegistry(ShellAdapter{}, ClaudeAdapter{}, CodexAdapter{}, OpenCodeAdapter{})
}

// fromContext retrieves the *Deck previously attached to the CLI app.
// It panics if none is present, since that indicates a programmer error
// (Before hook didn't run or didn't attach).
func fromContext(ctx *cli.Context) *Deck {
	ac, ok := ctx.App.Metadata[metadataKey].(*Deck)
	if !ok || ac == nil {
		panic("agentdeck: Deck not attached to cli.App; did Before run?")
	}
	return ac
}
