// Package config centralizes every tweakable knob for the game — window,
// camera, input, graphics and debug settings — as one JSON-serializable
// struct. It's loaded once at startup (creating a default file on first
// run) and can be re-saved whenever settings change, so tuning gameplay
// feel during development never requires a recompile.
package config

import (
	"foxdot/internal/camera"
	"foxdot/internal/input"
	"foxdot/internal/storage"
)

// fileName is where the config lives, relative to the storage.Store base
// directory.
const fileName = "config.json"

// WindowConfig controls the OS window and core render loop.
type WindowConfig struct {
	Width      int32  `json:"width"`
	Height     int32  `json:"height"`
	Title      string `json:"title"`
	Fullscreen bool   `json:"fullscreen"`
	VSync      bool   `json:"vsync"`
	TargetFPS  int32  `json:"targetFps"`
	MSAA4x     bool   `json:"msaa4x"`
	Resizable  bool   `json:"resizable"`
}

// GraphicsConfig controls broad visual settings appropriate for a stylized
// low-poly look.
type GraphicsConfig struct {
	BackgroundColorHex string `json:"backgroundColorHex"`
	ShowDebugOverlay   bool   `json:"showDebugOverlay"`
	ShowGrid           bool   `json:"showGrid"`
}

// Config is the full set of runtime-tweakable settings for the game.
type Config struct {
	Window   WindowConfig    `json:"window"`
	Camera   camera.Settings `json:"camera"`
	Input    input.Config    `json:"input"`
	Graphics GraphicsConfig  `json:"graphics"`
}

// Default returns the built-in defaults used on first run.
func Default() Config {
	return Config{
		Window: WindowConfig{
			Width:      1280,
			Height:     720,
			Title:      "FoxDot",
			Fullscreen: false,
			VSync:      true,
			TargetFPS:  60,
			MSAA4x:     true,
			Resizable:  true,
		},
		Camera: camera.DefaultSettings(),
		Input:  input.DefaultConfig(),
		Graphics: GraphicsConfig{
			BackgroundColorHex: "#8FD3F4",
			ShowDebugOverlay:   true,
			ShowGrid:           true,
		},
	}
}

// Load reads the config from the store, writing (and returning) the
// defaults if no config file exists yet.
func Load(store *storage.Store) (Config, error) {
	cfg := Default()
	if !store.Exists(fileName) {
		if err := store.SaveJSON(fileName, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	if err := store.LoadJSON(fileName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save persists cfg to the store.
func Save(store *storage.Store, cfg Config) error {
	return store.SaveJSON(fileName, cfg)
}
