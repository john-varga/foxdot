// Package game implements the FoxDot demo scene itself: a fox that can run,
// jump, nibble and swipe around a small low-poly forest clearing, viewed
// through a third-person orbit camera. It implements engine.Game by
// composing the (independently testable) sim, camera, world, assets and
// storage packages — this file is mostly wiring plus rendering.
package game

import (
	"fmt"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
	"foxdot/internal/camera"
	"foxdot/internal/config"
	"foxdot/internal/input"
	"foxdot/internal/sim"
	"foxdot/internal/storage"
	"foxdot/internal/world"
)

// autosaveSlot is where progress is stored between runs. A basic test/demo
// doesn't need multiple save slots yet, but the storage layer already
// supports them (see storage.Store.ListGameSlots).
const autosaveSlot = "autosave"

// FoxDot is the playable scene: control a fox in third person around a
// small forest clearing.
type FoxDot struct {
	cfg   config.Config
	store *storage.Store
	art   *assets.Store

	fox    *sim.Fox
	cam    *camera.ThirdPerson
	forest *world.Forest

	playtimeSecs float64
}

// New constructs the scene. assetsRoot should come from assets.FindRoot().
// Nothing touches raylib until Init/Update/Draw are called, so New itself
// is cheap and side-effect free.
func New(cfg config.Config, store *storage.Store, assetsRoot string) *FoxDot {
	return &FoxDot{
		cfg:    cfg,
		store:  store,
		art:    assets.NewStore(assetsRoot),
		fox:    sim.NewFox(sim.DefaultParams()),
		cam:    camera.NewThirdPerson(cfg.Camera),
		forest: world.NewDefaultForest(),
	}
}

// Init loads any existing autosave so re-launching the game resumes where
// you left off.
func (g *FoxDot) Init() error {
	save, err := g.store.LoadGameSlot(autosaveSlot)
	if err != nil {
		if err == storage.ErrNotExist {
			return nil
		}
		return fmt.Errorf("game: load autosave: %w", err)
	}

	g.fox.State.Position = rl.Vector3{X: save.FoxPosition.X, Y: save.FoxPosition.Y, Z: save.FoxPosition.Z}
	g.fox.State.Yaw = save.FoxYaw
	g.cam.Yaw = save.CameraYaw
	g.cam.Pitch = save.CameraPitch
	g.playtimeSecs = save.PlaytimeSecs
	return nil
}

// Update ticks the camera and fox simulation. Quicksave is bound to the
// Confirm action for this basic build.
func (g *FoxDot) Update(dt float32, in input.Frame) error {
	g.playtimeSecs += float64(dt)

	g.cam.Update(dt, g.fox.State.Position, in.Look, 0)
	g.fox.Step(dt, in, g.cam.Yaw, g.forest.HeightAt)

	if in.Confirm.Pressed {
		if err := g.save(autosaveSlot); err != nil {
			return fmt.Errorf("game: quicksave: %w", err)
		}
	}
	return nil
}

// fox scales the (real-world-ish sized) fox model down a bit so it reads as
// a small, nimble forest critter next to the trees/rocks rather than
// something deer-sized.
const foxModelScale = 0.55

// Draw renders the forest, the fox, and an optional debug overlay.
func (g *FoxDot) Draw() {
	rl.BeginMode3D(g.cam.RLCamera())
	g.forest.Draw(g.art, g.cfg.Graphics.ShowGrid)
	g.drawFox()
	rl.EndMode3D()

	if g.cfg.Graphics.ShowDebugOverlay {
		g.drawDebugOverlay()
	}
}

func (g *FoxDot) drawFox() {
	pos := g.fox.State.Position
	_ = g.art.Draw("fox", pos, g.fox.State.Yaw, foxModelScale)

	// A small marker above the fox's head shows the current action at a
	// glance (handy during development, before there's real animation).
	if g.fox.State.Action == sim.ActionNone {
		return
	}
	markerColor := rl.Yellow
	if g.fox.State.Action == sim.ActionSwipe {
		markerColor = rl.Red
	}
	facing := rl.Vector3{
		X: float32(math.Sin(float64(g.fox.State.Yaw))),
		Y: 0,
		Z: float32(math.Cos(float64(g.fox.State.Yaw))),
	}
	marker := rl.Vector3Add(rl.Vector3{X: pos.X, Y: pos.Y + 0.75, Z: pos.Z}, rl.Vector3Scale(facing, 0.4))
	rl.DrawSphere(marker, 0.08, markerColor)
}

func (g *FoxDot) drawDebugOverlay() {
	rl.DrawFPS(10, 10)

	pos := g.fox.State.Position
	lines := []string{
		fmt.Sprintf("fox pos: (%.1f, %.1f, %.1f)  yaw: %.0f deg", pos.X, pos.Y, pos.Z, rl.Rad2deg*g.fox.State.Yaw),
		fmt.Sprintf("grounded: %v  action: %s", g.fox.State.Grounded, g.fox.State.Action),
		fmt.Sprintf("camera dist: %.1f  yaw: %.0f deg  pitch: %.0f deg", g.cam.Distance(), rl.Rad2deg*g.cam.Yaw, rl.Rad2deg*g.cam.Pitch),
		fmt.Sprintf("playtime: %s", time.Duration(g.playtimeSecs*float64(time.Second)).Round(time.Second)),
		"move: WASD/stick  jump: Space/A  sprint: Shift/LT  nibble: E/X  swipe: F/B  quicksave: Enter",
	}
	for i, line := range lines {
		rl.DrawText(line, 10, int32(34+18*i), 18, rl.White)
	}
}

// Shutdown autosaves and releases loaded models on the way out so quitting
// normally never loses progress or leaks GPU resources.
func (g *FoxDot) Shutdown() {
	_ = g.save(autosaveSlot)
	g.art.Unload()
}

func (g *FoxDot) save(slot string) error {
	pos := g.fox.State.Position
	return g.store.SaveGameSlot(slot, storage.SaveGame{
		Version:      storage.CurrentSaveVersion,
		SavedAt:      time.Now().UTC(),
		FoxPosition:  storage.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
		FoxYaw:       g.fox.State.Yaw,
		CameraYaw:    g.cam.Yaw,
		CameraPitch:  g.cam.Pitch,
		PlaytimeSecs: g.playtimeSecs,
	})
}
