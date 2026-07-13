package config

import (
	"testing"

	"foxdot/internal/storage"
)

func TestLoadCreatesDefaultsOnFirstRun(t *testing.T) {
	store := storage.New(t.TempDir())

	cfg, err := Load(store)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Default() {
		t.Fatalf("expected defaults on first run, got %+v", cfg)
	}
	if !store.Exists("config.json") {
		t.Fatalf("expected Load to persist defaults to disk")
	}
}

func TestLoadReadsBackSavedChanges(t *testing.T) {
	store := storage.New(t.TempDir())

	cfg, err := Load(store)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cfg.Window.Width = 1920
	cfg.Camera.Distance = 9
	if err := Save(store, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(store)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if got.Window.Width != 1920 || got.Camera.Distance != 9 {
		t.Fatalf("expected saved changes to round-trip, got %+v", got)
	}
}
