// Package storage provides a small cross-platform file-interaction layer.
//
// It resolves an OS-appropriate application directory (using os.UserConfigDir,
// which maps to ~/Library/Application Support on macOS, %AppData% on Windows,
// and $XDG_CONFIG_HOME or ~/.config on Linux) and exposes generic JSON
// read/write helpers on top of it. Higher-level packages (config, save games,
// procedurally-generated content caches) build on this instead of touching
// the filesystem directly, which also makes them easy to redirect to a temp
// directory in tests.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AppDirName is the folder created inside the user's OS config directory.
const AppDirName = "FoxDot"

// ErrNotExist is returned when a requested relative path does not exist.
var ErrNotExist = errors.New("storage: not found")

// Store resolves relative paths against a base directory and provides
// generic JSON persistence helpers.
type Store struct {
	baseDir string
}

// New creates a Store rooted at an explicit directory. Useful for tests or
// for pointing the game at a non-default location.
func New(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// Default creates a Store rooted at the OS-appropriate per-user application
// directory (e.g. ~/Library/Application Support/FoxDot on macOS).
func Default() (*Store, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("storage: resolve user config dir: %w", err)
	}
	return New(filepath.Join(cfgDir, AppDirName)), nil
}

// BaseDir returns the root directory this Store operates under.
func (s *Store) BaseDir() string {
	return s.baseDir
}

// Path resolves a relative, slash-separated path against the base directory.
// It rejects paths that would escape the base directory.
func (s *Store) Path(rel string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
		return "", fmt.Errorf("storage: invalid relative path %q", rel)
	}
	return filepath.Join(s.baseDir, clean), nil
}

// EnsureDir makes sure the base directory (or a relative subdirectory of it)
// exists.
func (s *Store) EnsureDir(rel string) error {
	dir := s.baseDir
	if rel != "" {
		p, err := s.Path(rel)
		if err != nil {
			return err
		}
		dir = p
	}
	return os.MkdirAll(dir, 0o755)
}

// SaveJSON writes v as indented JSON to rel, creating parent directories as
// needed. The write is atomic (write to a temp file, then rename) so a crash
// mid-write can't corrupt existing data.
func (s *Store) SaveJSON(rel string, v any) error {
	path, err := s.Path(rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("storage: create directory for %q: %w", rel, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("storage: encode %q: %w", rel, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("storage: create temp file for %q: %w", rel, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("storage: write %q: %w", rel, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("storage: close temp file for %q: %w", rel, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("storage: finalize %q: %w", rel, err)
	}
	return nil
}

// LoadJSON reads and decodes JSON from rel into v. Returns ErrNotExist if
// the file is missing.
func (s *Store) LoadJSON(rel string, v any) error {
	path, err := s.Path(rel)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotExist
		}
		return fmt.Errorf("storage: read %q: %w", rel, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("storage: decode %q: %w", rel, err)
	}
	return nil
}

// Exists reports whether rel exists.
func (s *Store) Exists(rel string) bool {
	path, err := s.Path(rel)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Delete removes rel. It is not an error if the file is already absent.
func (s *Store) Delete(rel string) error {
	path, err := s.Path(rel)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete %q: %w", rel, err)
	}
	return nil
}

// List returns the base file names (without extension) of regular files
// directly inside the relative directory relDir, sorted alphabetically. It
// returns an empty slice (not an error) if the directory does not exist.
func (s *Store) List(relDir string) ([]string, error) {
	dir, err := s.Path(relDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("storage: list %q: %w", relDir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".tmp-") {
			continue
		}
		names = append(names, strings.TrimSuffix(name, filepath.Ext(name)))
	}
	sort.Strings(names)
	return names, nil
}
