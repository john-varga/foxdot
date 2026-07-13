package storage

import (
	"fmt"
	"path"
	"regexp"
	"time"
)

const (
	savesDir     = "saves"
	generatedDir = "generated"
)

// Vec3 is a plain, JSON-friendly 3D vector used in persisted data so this
// package has no dependency on raylib or any other rendering types.
type Vec3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// SaveGame is the serialized state of an in-progress game. It intentionally
// only holds plain data so it is trivial to version and unit test.
type SaveGame struct {
	Version      int       `json:"version"`
	SavedAt      time.Time `json:"savedAt"`
	FoxPosition  Vec3      `json:"foxPosition"`
	FoxYaw       float32   `json:"foxYaw"`
	CameraYaw    float32   `json:"cameraYaw"`
	CameraPitch  float32   `json:"cameraPitch"`
	PlaytimeSecs float64   `json:"playtimeSecs"`
}

// CurrentSaveVersion should be bumped whenever the SaveGame layout changes in
// a way that requires migration.
const CurrentSaveVersion = 1

var saveNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func sanitizeName(name string) (string, error) {
	if name == "" || !saveNamePattern.MatchString(name) {
		return "", fmt.Errorf("storage: invalid save name %q (use letters, numbers, - and _)", name)
	}
	return name, nil
}

// SaveGame persists a save under the given name (no extension needed).
func (s *Store) SaveGameSlot(name string, data SaveGame) error {
	clean, err := sanitizeName(name)
	if err != nil {
		return err
	}
	return s.SaveJSON(path.Join(savesDir, clean+".json"), data)
}

// LoadGameSlot loads a previously saved game by name.
func (s *Store) LoadGameSlot(name string) (SaveGame, error) {
	var data SaveGame
	clean, err := sanitizeName(name)
	if err != nil {
		return data, err
	}
	err = s.LoadJSON(path.Join(savesDir, clean+".json"), &data)
	return data, err
}

// DeleteGameSlot removes a save slot.
func (s *Store) DeleteGameSlot(name string) error {
	clean, err := sanitizeName(name)
	if err != nil {
		return err
	}
	return s.Delete(path.Join(savesDir, clean+".json"))
}

// ListGameSlots returns the names of all available save slots.
func (s *Store) ListGameSlots() ([]string, error) {
	return s.List(savesDir)
}

// SaveGenerated persists arbitrary procedurally-generated content (e.g. a
// forest layout keyed by seed) so it doesn't need to be regenerated on
// every load.
func (s *Store) SaveGenerated(key string, v any) error {
	clean, err := sanitizeName(key)
	if err != nil {
		return err
	}
	return s.SaveJSON(path.Join(generatedDir, clean+".json"), v)
}

// LoadGenerated loads previously cached procedurally-generated content.
func (s *Store) LoadGenerated(key string, v any) error {
	clean, err := sanitizeName(key)
	if err != nil {
		return err
	}
	return s.LoadJSON(path.Join(generatedDir, clean+".json"), v)
}
