package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type file struct {
	path string
}

func NewFile(path string) Metrics {
	return &file{path: filepath.FromSlash(path)}
}

func (f *file) Record(run Run) error {
	runs, err := f.Runs()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(append(runs, run), "", "  ")
	if err != nil {
		return fmt.Errorf("can't serialize metrics: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0755); err != nil {
		return fmt.Errorf("can't create a folder for metrics '%s': %w", f.path, err)
	}
	if err := os.WriteFile(f.path, data, 0600); err != nil {
		return fmt.Errorf("can't save metrics to '%s': %w", f.path, err)
	}
	return nil
}

func (f *file) Runs() ([]Run, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Run{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("can't read metrics from '%s': %w", f.path, err)
	}
	var runs []Run
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, fmt.Errorf("can't parse metrics from '%s': %w", f.path, err)
	}
	return runs, nil
}
