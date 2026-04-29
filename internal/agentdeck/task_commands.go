package agentdeck

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "list tracked tasks",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "print tasks as JSON"},
			&cli.StringFlag{Name: "agent", Usage: "filter by agent"},
			&cli.StringFlag{Name: "status", Usage: "filter by status"},
			&cli.StringFlag{Name: "dir", Usage: "filter by task directory"},
		},
		Action: func(ctx *cli.Context) error {
			ac := fromContext(ctx)
			tasks, err := ac.Store.List()
			if err != nil {
				return err
			}
			tasks, err = filterTasks(tasks, ctx.String("agent"), ctx.String("status"), ctx.String("dir"))
			if err != nil {
				return err
			}
			for i := range tasks {
				tasks[i] = refreshTaskStatus(tasks[i])
			}
			if ctx.Bool("json") {
				return json.NewEncoder(ctx.App.Writer).Encode(tasks)
			}
			if len(tasks) == 0 {
				fmt.Fprintln(ctx.App.Writer, "No tasks found")
				return nil
			}

			w := tabwriter.NewWriter(ctx.App.Writer, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TASK ID\tAGENT\tSTATUS\tCREATED\tUPDATED\tCWD\tPROMPT")
			for _, task := range tasks {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", task.ID, task.Agent, task.Status, formatTaskTime(task.CreatedAt), formatTaskTime(task.UpdatedAt), displayCWD(task.CWD), taskTitle(task.Prompt))
			}
			return w.Flush()
		},
	}
}

func statusCommand() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Aliases:   []string{"s"},
		Usage:     "show task status or agentdeck paths",
		ArgsUsage: "[task-id]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "print task as JSON"},
		},
		Action: func(ctx *cli.Context) error {
			ac := fromContext(ctx)
			if ctx.NArg() == 0 {
				printPaths(ctx.App.Writer, ac)
				return nil
			}

			task, err := ac.Store.Get(ctx.Args().First())
			if err != nil {
				return err
			}
			task = refreshTaskStatus(task)
			if ctx.Bool("json") {
				return json.NewEncoder(ctx.App.Writer).Encode(task)
			}
			printTaskStatus(ctx.App.Writer, task)
			return nil
		},
	}
}

func pathsCommand() *cli.Command {
	return &cli.Command{
		Name:  "paths",
		Usage: "show resolved agentdeck directories",
		Action: func(ctx *cli.Context) error {
			printPaths(ctx.App.Writer, fromContext(ctx))
			return nil
		},
	}
}

func printPaths(w io.Writer, ac *Deck) {
	fmt.Fprintf(w, "agentdeck is running\n")
	fmt.Fprintf(w, "  config:  %s\n", ac.Dirs.Config)
	fmt.Fprintf(w, "  data:    %s\n", ac.Dirs.Data)
	fmt.Fprintf(w, "  state:   %s\n", ac.Dirs.State)
	fmt.Fprintf(w, "  runtime: %s\n", ac.Dirs.Runtime)
}

func printTaskStatus(w io.Writer, task Task) {
	fmt.Fprintf(w, "id: %s\n", task.ID)
	fmt.Fprintf(w, "agent: %s\n", task.Agent)
	fmt.Fprintf(w, "status: %s\n", task.Status)
	fmt.Fprintf(w, "cwd: %s\n", task.CWD)
	fmt.Fprintf(w, "prompt: %s\n", task.Prompt)
	fmt.Fprintf(w, "log: %s\n", task.LogPath)
	if task.NativeSessionID != "" {
		fmt.Fprintf(w, "native session: %s\n", task.NativeSessionID)
	}
	if task.PID != 0 {
		fmt.Fprintf(w, "pid: %d\n", task.PID)
	}
	if !task.CreatedAt.IsZero() {
		fmt.Fprintf(w, "created: %s\n", task.CreatedAt.Format(timeFormat))
	}
	if !task.UpdatedAt.IsZero() {
		fmt.Fprintf(w, "updated: %s\n", task.UpdatedAt.Format(timeFormat))
	}
}

const timeFormat = "2006-01-02 15:04:05 MST"

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit-1]) + "..."
}

func taskTitle(prompt string) string {
	line := strings.TrimSpace(strings.Split(prompt, "\n")[0])
	if line == "" {
		line = "-"
	}
	return truncate(line, 40)
}

func filterTasks(tasks []Task, agent, status, dir string) ([]Task, error) {
	var absDir string
	var err error
	if dir != "" {
		absDir, err = filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("resolve dir filter %q: %w", dir, err)
		}
	}

	filtered := tasks[:0]
	for _, task := range tasks {
		if agent != "" && task.Agent != agent {
			continue
		}
		if status != "" && string(task.Status) != status {
			continue
		}
		if absDir != "" && task.CWD != absDir {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered, nil
}

func displayCWD(cwd string) string {
	wd, err := os.Getwd()
	if err != nil {
		return cwd
	}
	rel, err := filepath.Rel(wd, cwd)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return cwd
	}
	return rel
}

func formatTaskTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("2006-01-02 15:04")
}
