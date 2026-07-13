// Package audio is the game's audio subsystem: a catalog of the audio pack
// under assets/audio, a thin raylib playback engine, an ambience mixer, a
// music director, and a small event queue so gameplay/UI code can request
// sounds without depending on the engine directly.
//
// Like internal/assets, the catalog (this file) is pure data parsed from
// JSON with no raylib dependency, so it's unit testable on its own; only
// engine.go touches raylib, and only once a window/audio device exists.
package audio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Kind identifies how a track should be played back.
type Kind string

const (
	KindMusic           Kind = "music"
	KindAmbienceLoop    Kind = "ambience_loop"
	KindAmbienceOneShot Kind = "ambience_oneshot"
	KindSFX             Kind = "sfx"
)

// Track is one entry from assets/audio/manifest.json.
type Track struct {
	Name        string
	Kind        Kind
	RelPath     string // relative to the assets root, e.g. "audio/sfx/ui/click.wav"
	MIDIRelPath string // editable source, not loaded by the game; empty if none
	Loop        bool
	Seconds     float64
	BPM         int
	Key         string
}

// Streamed reports whether this track should be loaded as a raylib Music
// stream (pumped every frame, suited to long and/or looping audio) rather
// than a fully-decoded Sound (fire-and-forget, suited to short one-shots).
func (t Track) Streamed() bool {
	return t.Kind == KindMusic || t.Kind == KindAmbienceLoop
}

// Catalog is the parsed, queryable form of manifest.json.
type Catalog struct {
	tracks map[string]Track
	byKind map[Kind][]string
}

type manifestAsset struct {
	Name    string  `json:"name"`
	Kind    string  `json:"kind"`
	Path    string  `json:"path"`
	MIDI    string  `json:"midi"`
	Seconds float64 `json:"seconds"`
	Loop    bool    `json:"loop"`
	BPM     int     `json:"bpm"`
	Key     string  `json:"key"`
}

type manifestFile struct {
	Assets []manifestAsset `json:"assets"`
}

// LoadCatalog parses assets/audio/manifest.json under assetsRoot (typically
// the result of assets.FindRoot).
func LoadCatalog(assetsRoot string) (*Catalog, error) {
	path := filepath.Join(assetsRoot, "audio", "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("audio: reading manifest: %w", err)
	}

	var raw manifestFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("audio: parsing manifest %s: %w", path, err)
	}

	c := &Catalog{
		tracks: make(map[string]Track, len(raw.Assets)),
		byKind: make(map[Kind][]string),
	}
	for _, a := range raw.Assets {
		if a.Name == "" {
			return nil, fmt.Errorf("audio: manifest entry with empty name in %s", path)
		}
		if _, dup := c.tracks[a.Name]; dup {
			return nil, fmt.Errorf("audio: duplicate manifest entry %q in %s", a.Name, path)
		}
		t := Track{
			Name:    a.Name,
			Kind:    Kind(a.Kind),
			RelPath: filepath.Join("audio", a.Path),
			Loop:    a.Loop,
			Seconds: a.Seconds,
			BPM:     a.BPM,
			Key:     a.Key,
		}
		if a.MIDI != "" {
			t.MIDIRelPath = filepath.Join("audio", a.MIDI)
		}
		c.tracks[a.Name] = t
		c.byKind[t.Kind] = append(c.byKind[t.Kind], a.Name)
	}
	return c, nil
}

// ByName looks up a track by its catalog name.
func (c *Catalog) ByName(name string) (Track, bool) {
	t, ok := c.tracks[name]
	return t, ok
}

// NamesByKind returns every track name of the given kind.
func (c *Catalog) NamesByKind(k Kind) []string {
	return c.byKind[k]
}

// Len returns the total number of catalog entries.
func (c *Catalog) Len() int {
	return len(c.tracks)
}
