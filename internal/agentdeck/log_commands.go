package agentdeck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func logCommand() *cli.Command {
	return &cli.Command{
		Name:      "log",
		Usage:     "print a task log",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "lines", Usage: "print only the last N lines"},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return fmt.Errorf("log requires exactly one task id")
			}
			ac := fromContext(ctx)
			task, err := ac.Store.Get(ctx.Args().First())
			if err != nil {
				return err
			}
			return printLog(ctx.App.Writer, task.LogPath, ctx.Int("lines"))
		},
	}
}

func tailCommand() *cli.Command {
	return &cli.Command{
		Name:      "tail",
		Usage:     "follow a task log",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "lines", Value: 20, Usage: "print the last N lines before following"},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return fmt.Errorf("tail requires exactly one task id")
			}
			ac := fromContext(ctx)
			task, err := ac.Store.Get(ctx.Args().First())
			if err != nil {
				return err
			}
			if err := printLog(ctx.App.Writer, task.LogPath, ctx.Int("lines")); err != nil {
				return err
			}
			return followLog(ctx.Context, ctx.App.Writer, task.LogPath)
		},
	}
}

func TaskLogPath(stateDir, taskID string) string {
	return filepath.Join(stateDir, "tasks", taskID+".log")
}

func printLog(w io.Writer, path string, lines int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(w, "Log file does not exist")
			return nil
		}
		return fmt.Errorf("read log %q: %w", path, err)
	}
	if len(data) == 0 {
		fmt.Fprintln(w, "Log file is empty")
		return nil
	}

	text := string(data)
	if lines > 0 {
		text = lastLines(text, lines)
	}
	_, err = io.WriteString(w, text)
	return err
}

func followLog(ctx context.Context, w io.Writer, path string) error {
	var offset int64
	if info, err := os.Stat(path); err == nil {
		offset = info.Size()
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			file, err := os.Open(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return fmt.Errorf("open log %q: %w", path, err)
			}
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				_ = file.Close()
				return fmt.Errorf("seek log %q: %w", path, err)
			}
			written, err := io.Copy(w, file)
			_ = file.Close()
			if err != nil {
				return fmt.Errorf("copy log %q: %w", path, err)
			}
			offset += written
		}
	}
}

func lastLines(text string, limit int) string {
	if limit <= 0 {
		return text
	}
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return strings.Join(lines, "\n") + "\n"
}
