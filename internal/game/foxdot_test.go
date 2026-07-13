package game

import (
	"testing"

	"foxdot/internal/config"
	"foxdot/internal/input"
	"foxdot/internal/storage"
)

// Note: this test exercises Init/Update/Shutdown only, never Draw — Draw
// calls raylib rendering functions that require a live window/GPU context,
// which is exactly the split engine.Game is designed to make unnecessary
// for testing gameplay logic.

// testAssetsRoot points at the repo's real assets/ directory. Tests run
// with cwd set to this package's directory (internal/game).
const testAssetsRoot = "../../assets"

func newTestGame(t *testing.T) (*FoxDot, *storage.Store) {
	t.Helper()
	store := storage.New(t.TempDir())
	g := New(config.Default(), store, testAssetsRoot)
	if err := g.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return g, store
}

func TestInitWithNoSaveStartsAtOrigin(t *testing.T) {
	g, _ := newTestGame(t)
	if g.fox.State.Position.X != 0 || g.fox.State.Position.Y != 0 || g.fox.State.Position.Z != 0 {
		t.Fatalf("expected fox to start at origin with no save present, got %+v", g.fox.State.Position)
	}
}

func TestUpdateMovesFoxAndAdvancesPlaytime(t *testing.T) {
	g, _ := newTestGame(t)
	g.fox.State.Grounded = true

	frame := input.Frame{Move: input.Vector2{Y: 1}}
	for i := 0; i < 30; i++ {
		if err := g.Update(1.0/60.0, frame); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	if g.fox.State.Position.X == 0 && g.fox.State.Position.Z == 0 {
		t.Fatalf("expected fox to have moved after 30 ticks of forward input")
	}
	if g.playtimeSecs <= 0 {
		t.Fatalf("expected playtime to accumulate, got %v", g.playtimeSecs)
	}
}

func TestQuicksaveAndReloadRoundTrips(t *testing.T) {
	g, store := newTestGame(t)
	g.fox.State.Grounded = true

	frame := input.Frame{Move: input.Vector2{Y: 1}}
	for i := 0; i < 30; i++ {
		if err := g.Update(1.0/60.0, frame); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	// Confirm.Pressed on this tick triggers a quicksave.
	if err := g.Update(1.0/60.0, input.Frame{Confirm: input.ButtonState{Pressed: true}}); err != nil {
		t.Fatalf("Update (quicksave): %v", err)
	}
	savedPos := g.fox.State.Position

	if !store.Exists("saves/" + autosaveSlot + ".json") {
		t.Fatalf("expected quicksave to write %s slot", autosaveSlot)
	}

	// A fresh game instance loading from the same store should resume at
	// the saved position.
	g2 := New(config.Default(), store, testAssetsRoot)
	if err := g2.Init(); err != nil {
		t.Fatalf("Init on reload: %v", err)
	}
	if g2.fox.State.Position != savedPos {
		t.Fatalf("expected reloaded fox position %+v, got %+v", savedPos, g2.fox.State.Position)
	}
}

func TestShutdownAutosaves(t *testing.T) {
	g, store := newTestGame(t)
	g.Shutdown()
	if !store.Exists("saves/" + autosaveSlot + ".json") {
		t.Fatalf("expected Shutdown to autosave")
	}
}
