package agentdeck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const tasksFileName = "tasks.json"

var ErrTaskNotFound = errors.New("task not found")
var ErrTaskIDAmbiguous = errors.New("ambiguous task id")

type TaskStore interface {
	List() ([]Task, error)
	Get(id string) (Task, error)
	Put(task Task) error
	Delete(id string) error
	Update(id string, update func(*Task) error) error
}

type JSONTaskStore struct {
	stateDir string
}

func NewJSONTaskStore(stateDir string) *JSONTaskStore {
	return &JSONTaskStore{stateDir: stateDir}
}

func (s *JSONTaskStore) List() ([]Task, error) {
	tasks, err := s.load()
	if err != nil {
		return nil, err
	}

	items := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, task)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (s *JSONTaskStore) Get(id string) (Task, error) {
	tasks, err := s.load()
	if err != nil {
		return Task{}, err
	}

	resolvedID, err := resolveTaskID(tasks, id)
	if err != nil {
		return Task{}, err
	}
	task, ok := tasks[resolvedID]
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (s *JSONTaskStore) Put(task Task) error {
	if task.ID == "" {
		return errors.New("task id cannot be empty")
	}

	tasks, err := s.load()
	if err != nil {
		return err
	}
	tasks[task.ID] = task
	return s.save(tasks)
}

func (s *JSONTaskStore) Delete(id string) error {
	tasks, err := s.load()
	if err != nil {
		return err
	}
	resolvedID, err := resolveTaskID(tasks, id)
	if err != nil {
		return err
	}

	delete(tasks, resolvedID)
	return s.save(tasks)
}

func (s *JSONTaskStore) Update(id string, update func(*Task) error) error {
	if update == nil {
		return errors.New("update function cannot be nil")
	}

	tasks, err := s.load()
	if err != nil {
		return err
	}
	resolvedID, err := resolveTaskID(tasks, id)
	if err != nil {
		return err
	}
	task, ok := tasks[resolvedID]
	if !ok {
		return ErrTaskNotFound
	}
	if err := update(&task); err != nil {
		return err
	}

	tasks[resolvedID] = task
	return s.save(tasks)
}

func resolveTaskID(tasks map[string]Task, id string) (string, error) {
	if id == "" {
		return "", ErrTaskNotFound
	}
	if _, ok := tasks[id]; ok {
		return id, nil
	}

	matches := make([]string, 0, 2)
	for taskID := range tasks {
		if strings.HasPrefix(taskID, id) {
			matches = append(matches, taskID)
		}
	}
	sort.Strings(matches)

	switch len(matches) {
	case 0:
		return "", ErrTaskNotFound
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%w %q matches %s", ErrTaskIDAmbiguous, id, strings.Join(matches, ", "))
	}
}

func (s *JSONTaskStore) load() (map[string]Task, error) {
	f, err := os.Open(s.path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Task{}, nil
		}
		return nil, fmt.Errorf("open task store %q: %w", s.path(), err)
	}
	defer f.Close()

	var tasks map[string]Task
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&tasks); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]Task{}, nil
		}
		return nil, fmt.Errorf("decode task store %q: %w", s.path(), err)
	}
	if tasks == nil {
		return map[string]Task{}, nil
	}
	return tasks, nil
}

func (s *JSONTaskStore) save(tasks map[string]Task) error {
	if err := os.MkdirAll(s.stateDir, 0o755); err != nil {
		return fmt.Errorf("create state dir %q: %w", s.stateDir, err)
	}

	tmp, err := os.CreateTemp(s.stateDir, ".tasks-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp task store: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(tasks); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("encode task store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp task store: %w", err)
	}
	if err := os.Rename(tmpPath, s.path()); err != nil {
		cleanup()
		return fmt.Errorf("rename temp task store to %q: %w", s.path(), err)
	}

	return nil
}

func (s *JSONTaskStore) path() string {
	return filepath.Join(s.stateDir, tasksFileName)
}
