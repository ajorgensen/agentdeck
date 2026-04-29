package agentdeck

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

var ErrUnsupportedAgent = errors.New("unsupported agent")

type StartRequest struct {
	CWD     string
	Prompt  string
	LogPath string
	Env     []string
	Config  *Config
}

type StartResult struct {
	PID             int
	NativeSessionID string
}

type AdapterStatus struct {
	Status  TaskStatus
	Message string
}

type AgentAdapter interface {
	Name() string
	Start(ctx context.Context, req StartRequest) (*StartResult, error)
	ResumeCommand(task Task) (*exec.Cmd, error)
	Status(ctx context.Context, task Task) (AdapterStatus, error)
	Kill(ctx context.Context, task Task) error
}

type AdapterRegistry struct {
	adapters map[string]AgentAdapter
}

func NewAdapterRegistry(adapters ...AgentAdapter) *AdapterRegistry {
	registry := &AdapterRegistry{adapters: map[string]AgentAdapter{}}
	for _, adapter := range adapters {
		if adapter == nil {
			continue
		}
		name := adapter.Name()
		if _, exists := registry.adapters[name]; exists {
			panic(fmt.Sprintf("duplicate adapter %q", name))
		}
		registry.adapters[name] = adapter
	}
	return registry
}

func (r *AdapterRegistry) Get(name string) (AgentAdapter, error) {
	adapter, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAgent, name)
	}
	return adapter, nil
}

type FakeAdapter struct {
	NameValue       string
	StartResult     StartResult
	StartErr        error
	ResumeArgsValue []string
	ResumeErr       error
	StatusValue     AdapterStatus
	StatusErr       error
	KillErr         error
}

func (a FakeAdapter) Name() string {
	return a.NameValue
}

func (a FakeAdapter) Start(ctx context.Context, req StartRequest) (*StartResult, error) {
	if a.StartErr != nil {
		return nil, a.StartErr
	}
	return &a.StartResult, nil
}

func (a FakeAdapter) ResumeCommand(task Task) (*exec.Cmd, error) {
	if a.ResumeErr != nil {
		return nil, a.ResumeErr
	}
	if len(a.ResumeArgsValue) == 0 {
		return nil, errors.New("fake adapter resume args cannot be empty")
	}
	cmd := exec.Command(a.ResumeArgsValue[0], a.ResumeArgsValue[1:]...)
	cmd.Dir = task.CWD
	return cmd, nil
}

func (a FakeAdapter) Status(ctx context.Context, task Task) (AdapterStatus, error) {
	if a.StatusErr != nil {
		return AdapterStatus{}, a.StatusErr
	}
	return a.StatusValue, nil
}

func (a FakeAdapter) Kill(ctx context.Context, task Task) error {
	return a.KillErr
}
