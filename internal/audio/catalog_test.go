package audio

import (
	"os"
	"path/filepath"
	"testing"
)

// assetsRoot resolves to the repo's real /assets directory. Tests run with
// cwd set to this package's directory (internal/audio), so ../../assets is
// the repo root's assets folder.
const assetsRoot = "../../assets"

func loadTestCatalog(t *testing.T) *Catalog {
	t.Helper()
	c, err := LoadCatalog(assetsRoot)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	return c
}

func TestLoadCatalogNonEmpty(t *testing.T) {
	c := loadTestCatalog(t)
	if c.Len() == 0 {
		t.Fatalf("expected a non-empty catalog")
	}
}

func TestLoadCatalogMissingManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadCatalog(dir); err == nil {
		t.Fatalf("expected an error for a directory with no audio/manifest.json")
	}
}

func TestCatalogFilesExistOnDisk(t *testing.T) {
	c := loadTestCatalog(t)
	for name, k := range c.byKind {
		for _, n := range k {
			track, _ := c.ByName(n)
			path := filepath.Join(assetsRoot, track.RelPath)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("%s (%s): audio file missing at %s: %v", n, name, path, err)
			}
			if track.MIDIRelPath != "" {
				midiPath := filepath.Join(assetsRoot, track.MIDIRelPath)
				if _, err := os.Stat(midiPath); err != nil {
					t.Errorf("%s: midi file missing at %s: %v", n, midiPath, err)
				}
			}
		}
	}
}

func TestCatalogKnownKindsAndExpectedCounts(t *testing.T) {
	c := loadTestCatalog(t)

	wantMinByKind := map[Kind]int{
		KindMusic:           6,
		KindAmbienceLoop:    8,
		KindAmbienceOneShot: 1,
		KindSFX:             20,
	}
	for k, min := range wantMinByKind {
		got := len(c.NamesByKind(k))
		if got < min {
			t.Errorf("kind %s: got %d entries, want at least %d", k, got, min)
		}
	}

	for name, track := range c.tracks {
		switch track.Kind {
		case KindMusic, KindAmbienceLoop, KindAmbienceOneShot, KindSFX:
		default:
			t.Errorf("%s: unrecognized kind %q", name, track.Kind)
		}
	}
}

func TestStreamedMatchesLoopableKinds(t *testing.T) {
	c := loadTestCatalog(t)
	for name, track := range c.tracks {
		want := track.Kind == KindMusic || track.Kind == KindAmbienceLoop
		if got := track.Streamed(); got != want {
			t.Errorf("%s: Streamed()=%v, want %v (kind=%s)", name, got, want, track.Kind)
		}
	}
}

func TestMusicTracksHaveMIDIAndTempo(t *testing.T) {
	c := loadTestCatalog(t)
	for _, name := range c.NamesByKind(KindMusic) {
		track, _ := c.ByName(name)
		if track.MIDIRelPath == "" {
			t.Errorf("%s: expected a midi source path", name)
		}
		if track.BPM <= 0 {
			t.Errorf("%s: expected a positive BPM, got %d", name, track.BPM)
		}
		if track.Seconds <= 0 {
			t.Errorf("%s: expected a positive duration, got %v", name, track.Seconds)
		}
	}
}

func TestByNameUnknownTrack(t *testing.T) {
	c := loadTestCatalog(t)
	if _, ok := c.ByName("does-not-exist"); ok {
		t.Fatalf("expected unknown track to be absent")
	}
}

func TestLoadCatalogRejectsDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "audio"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"assets":[
		{"name":"dup","kind":"sfx","path":"sfx/a.wav","loop":false},
		{"name":"dup","kind":"sfx","path":"sfx/b.wav","loop":false}
	]}`
	if err := os.WriteFile(filepath.Join(dir, "audio", "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(dir); err == nil {
		t.Fatalf("expected an error for duplicate manifest names")
	}
}
