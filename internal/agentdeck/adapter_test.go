package agentdeck

import (
	"context"
	"errors"
	"testing"
)

func TestAdapterRegistryLookup(t *testing.T) {
	adapter := FakeAdapter{NameValue: "fake"}
	registry := NewAdapterRegistry(adapter)

	got, err := registry.Get("fake")
	if err != nil {
		t.Fatalf("get adapter: %v", err)
	}
	if got.Name() != "fake" {
		t.Fatalf("adapter name = %q, want fake", got.Name())
	}
}

func TestAdapterRegistryUnknownAgent(t *testing.T) {
	registry := NewAdapterRegistry()

	_, err := registry.Get("missing")
	if !errors.Is(err, ErrUnsupportedAgent) {
		t.Fatalf("error = %v, want ErrUnsupportedAgent", err)
	}
}

func TestAdapterRegistryDuplicateAgentPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected duplicate adapter registration to panic")
		}
	}()

	_ = NewAdapterRegistry(FakeAdapter{NameValue: "fake"}, FakeAdapter{NameValue: "fake"})
}

func TestFakeAdapterStartResumeStatusKill(t *testing.T) {
	adapter := FakeAdapter{
		NameValue:       "fake",
		StartResult:     StartResult{PID: 123, NativeSessionID: "native-1"},
		ResumeArgsValue: []string{"fake-agent", "resume", "native-1"},
		StatusValue:     AdapterStatus{Status: TaskStatusRunning, Message: "still running"},
	}

	start, err := adapter.Start(context.Background(), StartRequest{CWD: "/tmp/work", Prompt: "do work", LogPath: "/tmp/task.log"})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if start.PID != 123 || start.NativeSessionID != "native-1" {
		t.Fatalf("start result = %#v", start)
	}

	cmd, err := adapter.ResumeCommand(Task{CWD: "/tmp/work", NativeSessionID: "native-1"})
	if err != nil {
		t.Fatalf("resume command: %v", err)
	}
	if cmd.Path != "fake-agent" || len(cmd.Args) != 3 || cmd.Args[2] != "native-1" || cmd.Dir != "/tmp/work" {
		t.Fatalf("command = %#v", cmd)
	}

	status, err := adapter.Status(context.Background(), Task{ID: "abc123"})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Status != TaskStatusRunning || status.Message != "still running" {
		t.Fatalf("status = %#v", status)
	}

	if err := adapter.Kill(context.Background(), Task{ID: "abc123"}); err != nil {
		t.Fatalf("kill: %v", err)
	}
}
