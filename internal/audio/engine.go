package audio

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Category groups tracks for independent volume control (a "music" slider
// separate from "sfx" and "ambience" in options, say).
type Category int

const (
	CategorySFX Category = iota
	CategoryMusic
	CategoryAmbience
)

func categoryFor(k Kind) Category {
	switch k {
	case KindMusic:
		return CategoryMusic
	case KindAmbienceLoop, KindAmbienceOneShot:
		return CategoryAmbience
	default:
		return CategorySFX
	}
}

// Engine owns the raylib audio device and lazily loads/caches catalog
// tracks: short entries decode fully into an rl.Sound (played via
// LoadSoundAlias so overlapping instances don't cut each other off); long
// and/or looping entries stream via rl.Music, which must be pumped every
// frame through Update. It must only be used after InitAudioDevice can
// succeed (i.e. never from a unit test) — see catalog_test.go for the
// window/device-free parts of this package.
type Engine struct {
	root    string
	catalog *Catalog

	sounds  map[string]rl.Sound
	aliases []rl.Sound
	streams map[string]rl.Music
	closed  bool

	MasterVolume   float32
	SFXVolume      float32
	MusicVolume    float32
	AmbienceVolume float32
}

// NewEngine opens the audio device and returns a ready-to-use Engine.
func NewEngine(root string, catalog *Catalog) *Engine {
	rl.InitAudioDevice()
	return &Engine{
		root:    root,
		catalog: catalog,
		sounds:  make(map[string]rl.Sound),
		streams: make(map[string]rl.Music),

		MasterVolume:   1,
		SFXVolume:      1,
		MusicVolume:    1,
		AmbienceVolume: 1,
	}
}

func (e *Engine) categoryVolume(c Category) float32 {
	switch c {
	case CategoryMusic:
		return e.MusicVolume
	case CategoryAmbience:
		return e.AmbienceVolume
	default:
		return e.SFXVolume
	}
}

func (e *Engine) path(track Track) string {
	return filepath.Join(e.root, track.RelPath)
}

func (e *Engine) loadSound(name string) (rl.Sound, Track, error) {
	track, ok := e.catalog.ByName(name)
	if !ok {
		return rl.Sound{}, Track{}, fmt.Errorf("audio: unknown track %q", name)
	}
	if s, ok := e.sounds[name]; ok {
		return s, track, nil
	}
	s := rl.LoadSound(e.path(track))
	if !rl.IsSoundValid(s) {
		return s, track, fmt.Errorf("audio: failed to load sound %q from %s", name, e.path(track))
	}
	e.sounds[name] = s
	return s, track, nil
}

func (e *Engine) loadStream(name string) (rl.Music, Track, error) {
	track, ok := e.catalog.ByName(name)
	if !ok {
		return rl.Music{}, Track{}, fmt.Errorf("audio: unknown track %q", name)
	}
	if m, ok := e.streams[name]; ok {
		return m, track, nil
	}
	m := rl.LoadMusicStream(e.path(track))
	if !rl.IsMusicValid(m) {
		return m, track, fmt.Errorf("audio: failed to load music stream %q from %s", name, e.path(track))
	}
	m.Looping = track.Loop
	e.streams[name] = m
	return m, track, nil
}

// PlaySFX plays a short one-shot track (sfx/* or an ambience one-shot) by
// catalog name at the given volume (0..1, multiplied by the SFX/master
// category volume). Overlapping plays of the same name don't cut each other
// off. Errors (unknown name, failed load) are logged, not fatal — a missing
// sound shouldn't crash gameplay.
func (e *Engine) PlaySFX(name string, volume float32) {
	s, track, err := e.loadSound(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	vol := clamp01(volume) * e.categoryVolume(categoryFor(track.Kind)) * e.MasterVolume
	alias := rl.LoadSoundAlias(s)
	rl.SetSoundVolume(alias, vol)
	rl.PlaySound(alias)
	e.aliases = append(e.aliases, alias)
}

// PlayLoop starts (or restarts, if stopped) a streamed track — music or an
// ambience loop — at volume 0 by default; callers (Mixer) ramp the volume
// up/down explicitly so multiple loops can be crossfaded smoothly.
func (e *Engine) PlayLoop(name string) error {
	m, _, err := e.loadStream(name)
	if err != nil {
		return err
	}
	if !rl.IsMusicStreamPlaying(m) {
		rl.PlayMusicStream(m)
	}
	return nil
}

// StopLoop stops a streamed track if it's playing.
func (e *Engine) StopLoop(name string) {
	m, ok := e.streams[name]
	if !ok {
		return
	}
	rl.StopMusicStream(m)
}

// IsLoopPlaying reports whether a streamed track is currently playing.
func (e *Engine) IsLoopPlaying(name string) bool {
	m, ok := e.streams[name]
	return ok && rl.IsMusicStreamPlaying(m)
}

// SetLoopVolume sets a streamed track's volume, scaled by its category and
// the master volume. weight is expected in [0, 1].
func (e *Engine) SetLoopVolume(name string, weight float32) {
	m, ok := e.streams[name]
	if !ok {
		return
	}
	track, _ := e.catalog.ByName(name)
	rl.SetMusicVolume(m, clamp01(weight)*e.categoryVolume(categoryFor(track.Kind))*e.MasterVolume)
}

// Update must be called once per frame: it pumps every currently-playing
// music stream (required by raylib to actually produce audio) and reclaims
// SFX aliases that have finished playing.
func (e *Engine) Update() {
	for _, m := range e.streams {
		if rl.IsMusicStreamPlaying(m) {
			rl.UpdateMusicStream(m)
		}
	}

	live := e.aliases[:0]
	for _, a := range e.aliases {
		if rl.IsSoundPlaying(a) {
			live = append(live, a)
		} else {
			rl.UnloadSoundAlias(a)
		}
	}
	e.aliases = live
}

// Close releases every loaded sound/stream and closes the audio device.
// Safe to call more than once (a no-op after the first call) and even if
// nothing was ever loaded.
func (e *Engine) Close() {
	if e.closed {
		return
	}
	e.closed = true

	for _, a := range e.aliases {
		rl.UnloadSoundAlias(a)
	}
	e.aliases = nil
	for name, s := range e.sounds {
		rl.UnloadSound(s)
		delete(e.sounds, name)
	}
	for name, m := range e.streams {
		rl.UnloadMusicStream(m)
		delete(e.streams, name)
	}
	rl.CloseAudioDevice()
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
