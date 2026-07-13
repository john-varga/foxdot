package assets

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Store loads catalog models from disk on first use and caches the result,
// so repeatedly drawing e.g. "pine" across a whole forest only touches the
// filesystem/GPU once. It must only be used after a raylib window/GPU
// context exists (i.e. never from a unit test — see catalog_test.go for the
// window-free parts of this package).
type Store struct {
	root   string
	loaded map[string]rl.Model
}

// NewStore creates a Store that resolves catalog RelPaths against root
// (typically the result of FindRoot).
func NewStore(root string) *Store {
	return &Store{root: root, loaded: make(map[string]rl.Model)}
}

// get loads (or returns the cached) raylib Model plus its catalog metadata.
func (s *Store) get(name string) (rl.Model, Model, error) {
	meta, ok := ByName(name)
	if !ok {
		return rl.Model{}, Model{}, fmt.Errorf("assets: unknown model %q", name)
	}
	if m, ok := s.loaded[name]; ok {
		return m, meta, nil
	}

	path := filepath.Join(s.root, meta.RelPath)
	model := rl.LoadModel(path)
	s.loaded[name] = model

	if model.MeshCount == 0 {
		return model, meta, fmt.Errorf("assets: failed to load %q from %s", name, path)
	}
	return model, meta, nil
}

// Draw renders a catalog model by name at position, planted on the ground
// (using the model's authored MinY so it neither clips nor floats), facing
// yawRadians (already corrected for the model's own forward axis). Errors
// (unknown name, failed load) are non-fatal: a small magenta wire cube is
// drawn in the model's place so a missing/misnamed asset is obvious without
// crashing the game.
func (s *Store) Draw(name string, position rl.Vector3, yawRadians float32, scale float32) error {
	model, meta, err := s.get(name)
	if err != nil {
		drawMissingPlaceholder(position, scale)
		return err
	}

	drawPos := rl.Vector3{
		X: position.X,
		Y: position.Y - meta.MinY*scale,
		Z: position.Z,
	}
	totalYaw := (yawRadians + meta.YawOffsetRadians) * rl.Rad2deg
	rl.DrawModelEx(model, drawPos, rl.Vector3{Y: 1}, totalYaw, rl.Vector3{X: scale, Y: scale, Z: scale}, rl.White)
	return nil
}

func drawMissingPlaceholder(position rl.Vector3, scale float32) {
	size := scale
	if size <= 0 {
		size = 1
	}
	center := rl.Vector3{X: position.X, Y: position.Y + size/2, Z: position.Z}
	dims := rl.Vector3{X: size, Y: size, Z: size}
	rl.DrawCubeWiresV(center, dims, rl.Magenta)
}

// Unload releases every model this Store has loaded. Safe to call even if
// nothing was ever loaded.
func (s *Store) Unload() {
	for name, model := range s.loaded {
		if model.MeshCount > 0 {
			rl.UnloadModel(model)
		}
		delete(s.loaded, name)
	}
}
