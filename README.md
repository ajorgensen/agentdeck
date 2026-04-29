# agentdeck

`agentdeck` is a small CLI for running and tracking coding-agent tasks across multiple folders.

It starts agents in the background, records the task metadata you need later, lets you inspect logs and status, and resumes the task in the agent's native CLI when you want to take over interactively.

## Purpose

Coding agents are most useful when they can work on several branches, worktrees, or repositories without forcing you to keep every terminal open. `agentdeck` gives you a local task index for that workflow.

It does not try to replace Claude Code, Codex, OpenCode, or other agent interfaces. Instead, it stores enough information to reopen the right native session in the right directory.

Supported adapters:

- `claude`: starts Claude Code with `claude -p` and resumes with `claude --resume`.
- `codex`: starts Codex with `codex exec` and resumes with `codex resume`.
- `opencode`: starts OpenCode with `opencode run` and resumes with `opencode --session` or `opencode --continue`.
- `shell`: development adapter that runs the prompt as `sh -c <prompt>`.

## Install

Install from source with Go:

```sh
go install github.com/ajorgensen/agentdeck/cmd/agentdeck@latest
```

For local development:

```sh
git clone https://github.com/ajorgensen/agentdeck.git
cd agentdeck
go run ./cmd/agentdeck paths
```

`agentdeck` uses XDG-style directories by default. You can override them with environment variables:

```sh
AGENTDECK_CONFIG_DIR=/path/to/config \
AGENTDECK_DATA_DIR=/path/to/data \
AGENTDECK_STATE_DIR=/path/to/state \
AGENTDECK_RUNTIME_DIR=/path/to/runtime \
agentdeck paths
```

## Usage

Start a task:

```sh
agentdeck start --agent claude --dir ../auth-worktree "Fix the login redirect bug"
```

List tracked tasks:

```sh
agentdeck list
agentdeck list --agent claude
agentdeck list --status running
agentdeck list --json
```

Inspect one task:

```sh
agentdeck status <task-id>
agentdeck log <task-id>
agentdeck log <task-id> --lines 50
agentdeck tail <task-id>
```

Resume a task in the native agent CLI:

```sh
agentdeck resume <task-id>
```

Stop or remove a tracked task:

```sh
agentdeck kill <task-id>
agentdeck forget <task-id>
```

Task IDs can be abbreviated as long as the prefix is unambiguous, similar to Git commit IDs.

## Common Patterns

Run agents in separate worktrees:

```sh
git worktree add ../auth-fix -b auth-fix
agentdeck start --agent claude --dir ../auth-fix "Fix the auth redirect bug and add regression tests"
```

Check what is running:

```sh
agentdeck list --status running
agentdeck tail <task-id>
```

Review the result and take over:

```sh
agentdeck log <task-id> --lines 100
agentdeck resume <task-id>
```

Run a quick local smoke test with the shell adapter:

```sh
agentdeck start --agent shell --dir . "go test ./..."
agentdeck tail <task-id>
```

Clean up finished tasks from the local index:

```sh
agentdeck forget <task-id>
```

## How It Works

When you run `agentdeck start`, the CLI:

1. Validates the working directory.
2. Creates a local task ID.
3. Starts the selected agent in a background or non-interactive mode.
4. Captures stdout and stderr to a task log.
5. Stores the task ID, agent, working directory, prompt, PID, log path, and native session ID when available.

When you run `agentdeck resume`, the CLI builds the adapter-specific resume command and runs it in the task directory so your terminal becomes the native agent interface.

## Development

Run tests:

```sh
go test ./...
```

Run the CLI from the checkout:

```sh
go run ./cmd/agentdeck list
```

Integration tests for vendor CLIs are opt-in:

```sh
AGENTDECK_INTEGRATION=1 go test ./...
```

## Current Limitations

- Status is mostly PID-liveness based, so short-lived tasks may show `unknown` after completion.
- Codex and OpenCode session IDs are captured from early JSON output when available. If unavailable, resume falls back to each tool's native continue or last-session behavior.
- The `shell` adapter is intended for local validation, not as a coding-agent integration.
