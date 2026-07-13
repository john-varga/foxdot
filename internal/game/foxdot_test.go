package game

import (
	"math"
	"testing"

	"foxdot/internal/comp"
	"foxdot/internal/config"
	"foxdot/internal/ecs"
	"foxdot/internal/input"
	"foxdot/internal/storage"
	"foxdot/internal/worldtime"
)

// Note: this test exercises Init/Update/Shutdown only, never Draw — Draw
// calls raylib rendering functions that require a live window/GPU context,
// which is exactly the split engine.Game is designed to make unnecessary
// for testing gameplay logic. Init/Shutdown here do open/close a real
// raylib audio device (no window needed for that), which works fine in
// this sandboxed test environment.

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
	t.Cleanup(g.Shutdown)
	return g, store
}

func foxTransform(t *testing.T, g *FoxDot) comp.Transform {
	t.Helper()
	tr, ok := ecs.Get[comp.Transform](g.world, g.player)
	if !ok {
		t.Fatalf("expected player entity to have a Transform")
	}
	return *tr
}

func setFoxGrounded(g *FoxDot, grounded bool) {
	v, _ := ecs.Get[comp.Velocity](g.world, g.player)
	v.Grounded = grounded
}

func TestInitWithNoSaveStartsAtOrigin(t *testing.T) {
	g, _ := newTestGame(t)
	pos := foxTransform(t, g).Position
	if pos.X != 0 || pos.Y != 0 || pos.Z != 0 {
		t.Fatalf("expected fox to start at origin with no save present, got %+v", pos)
	}
}

func TestUpdateMovesFoxAndAdvancesPlaytime(t *testing.T) {
	g, _ := newTestGame(t)
	setFoxGrounded(g, true)

	frame := input.Frame{Move: input.Vector2{Y: 1}}
	for i := 0; i < 30; i++ {
		if err := g.Update(1.0/60.0, frame); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	pos := foxTransform(t, g).Position
	if pos.X == 0 && pos.Z == 0 {
		t.Fatalf("expected fox to have moved after 30 ticks of forward input")
	}
	if g.playtimeSecs <= 0 {
		t.Fatalf("expected playtime to accumulate, got %v", g.playtimeSecs)
	}
}

func TestQuicksaveAndReloadRoundTrips(t *testing.T) {
	g, store := newTestGame(t)
	setFoxGrounded(g, true)

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
	savedPos := foxTransform(t, g).Position

	if !store.Exists("saves/" + autosaveSlot + ".json") {
		t.Fatalf("expected quicksave to write %s slot", autosaveSlot)
	}

	// raylib only supports one open audio device at a time, so close g's
	// before opening a second game instance below (Shutdown is otherwise
	// safe to run again via t.Cleanup: Close/Unload are idempotent).
	g.Shutdown()

	// A fresh game instance loading from the same store should resume at
	// the saved position.
	g2 := New(config.Default(), store, testAssetsRoot)
	if err := g2.Init(); err != nil {
		t.Fatalf("Init on reload: %v", err)
	}
	defer g2.Shutdown()
	if got := foxTransform(t, g2).Position; got != savedPos {
		t.Fatalf("expected reloaded fox position %+v, got %+v", savedPos, got)
	}
}

func TestQuicksavePersistsTimeOfDay(t *testing.T) {
	g, store := newTestGame(t)

	for i := 0; i < 120; i++ { // advance the clock a bit
		if err := g.Update(1.0/60.0, input.Frame{}); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	if err := g.Update(1.0/60.0, input.Frame{Confirm: input.ButtonState{Pressed: true}}); err != nil {
		t.Fatalf("Update (quicksave): %v", err)
	}

	save, err := store.LoadGameSlot(autosaveSlot)
	if err != nil {
		t.Fatalf("LoadGameSlot: %v", err)
	}
	if save.TimeOfDay <= 0 {
		t.Fatalf("expected a positive persisted TimeOfDay, got %v", save.TimeOfDay)
	}

	// raylib only supports one open audio device at a time.
	g.Shutdown()

	g2 := New(config.Default(), store, testAssetsRoot)
	if err := g2.Init(); err != nil {
		t.Fatalf("Init on reload: %v", err)
	}
	defer g2.Shutdown()

	clock := ecs.MustGetResource[worldtime.Clock](g2.world)
	if math.Abs(clock.TimeOfDay()-save.TimeOfDay) > 1e-6 {
		t.Fatalf("expected reloaded TimeOfDay %v, got %v", save.TimeOfDay, clock.TimeOfDay())
	}
}

func TestShutdownAutosaves(t *testing.T) {
	g, store := newTestGame(t)
	g.Shutdown()
	if !store.Exists("saves/" + autosaveSlot + ".json") {
		t.Fatalf("expected Shutdown to autosave")
	}
}
