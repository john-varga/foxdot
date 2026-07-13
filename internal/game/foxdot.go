// Package game implements the FoxDot demo scene itself: a fox that can run,
// jump, nibble and swipe around a small low-poly forest clearing that
// breathes through a day/night cycle, viewed through a third-person orbit
// camera and scored by a small audio subsystem (ambience mix, music, SFX).
// It implements engine.Game by driving an Entity Component System world
// (see internal/ecs and internal/comp): the fox, every forest prop, and
// every ambient creature are all just entities made of the same handful of
// components, updated each tick by the System functions in systems.go —
// this file is mostly construction, save/load, and top-level Update/Draw
// wiring.
package game

import (
	"fmt"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
	"foxdot/internal/audio"
	"foxdot/internal/camera"
	"foxdot/internal/comp"
	"foxdot/internal/config"
	"foxdot/internal/ecs"
	"foxdot/internal/input"
	"foxdot/internal/sim"
	"foxdot/internal/storage"
	"foxdot/internal/world"
	"foxdot/internal/worldtime"
)

// autosaveSlot is where progress is stored between runs. A basic test/demo
// doesn't need multiple save slots yet, but the storage layer already
// supports them (see storage.Store.ListGameSlots).
const autosaveSlot = "autosave"

// foxModelScale scales the (real-world-ish sized) fox model down a bit so
// it reads as a small, nimble forest critter next to the trees/rocks
// rather than something deer-sized.
const foxModelScale = 0.55

// ambienceFadeSeconds/musicFadeSeconds set how long ambience loops and
// music tracks take to crossfade (see internal/audio.Mixer).
const (
	ambienceFadeSeconds = 2.5
	musicFadeSeconds    = 3.0
)

// FoxDot is the playable scene: control a fox in third person around a
// small forest clearing.
type FoxDot struct {
	cfg        config.Config
	store      *storage.Store
	art        *assets.Store
	assetsRoot string

	world  *ecs.World
	player ecs.Entity

	audioEngine *audio.Engine
	ambience    *audio.AmbiencePlayer
	music       *audio.MusicDirector

	playtimeSecs float64
}

// New constructs the scene. assetsRoot should come from assets.FindRoot().
// Nothing touches raylib until Init/Update/Draw are called, so New itself
// is cheap and side-effect free.
func New(cfg config.Config, store *storage.Store, assetsRoot string) *FoxDot {
	w := ecs.NewWorld()
	ecs.SetResource(w, *worldtime.NewClock(worldtime.DefaultSettings()))
	ecs.SetResource(w, *camera.NewThirdPerson(cfg.Camera))
	ecs.SetResource(w, audio.EventQueue{})

	g := &FoxDot{
		cfg:        cfg,
		store:      store,
		art:        assets.NewStore(assetsRoot),
		assetsRoot: assetsRoot,
		world:      w,
	}

	g.player = w.NewEntity()
	ecs.Set(w, g.player, comp.Transform{})
	ecs.Set(w, g.player, comp.Velocity{})
	ecs.Set(w, g.player, comp.ActionState{})
	ecs.Set(w, g.player, comp.MoveParams{Params: sim.DefaultParams()})
	ecs.Set(w, g.player, comp.ModelRender{AssetName: "fox", Scale: foxModelScale})
	ecs.Set(w, g.player, comp.PlayerControlled{})
	ecs.Set(w, g.player, comp.Footsteps{})

	world.SpawnDefaultForest(w)

	return g
}

// Init opens the audio device, loads the audio catalog, and loads any
// existing autosave so re-launching the game resumes where you left off
// (position, camera orbit, playtime, and time of day).
func (g *FoxDot) Init() error {
	catalog, err := audio.LoadCatalog(g.assetsRoot)
	if err != nil {
		return fmt.Errorf("game: load audio catalog: %w", err)
	}
	g.audioEngine = audio.NewEngine(g.assetsRoot, catalog)
	g.audioEngine.MasterVolume = g.cfg.Audio.MasterVolume
	g.audioEngine.SFXVolume = g.cfg.Audio.SFXVolume
	g.audioEngine.MusicVolume = g.cfg.Audio.MusicVolume
	g.audioEngine.AmbienceVolume = g.cfg.Audio.AmbienceVolume
	g.ambience = audio.NewAmbiencePlayer(g.audioEngine, ambienceFadeSeconds)
	g.music = audio.NewMusicDirector(g.audioEngine, musicFadeSeconds)

	save, err := g.store.LoadGameSlot(autosaveSlot)
	if err != nil {
		if err == storage.ErrNotExist {
			return nil
		}
		return fmt.Errorf("game: load autosave: %w", err)
	}
	return g.applySave(save)
}

func (g *FoxDot) applySave(save storage.SaveGame) error {
	t, ok := ecs.Get[comp.Transform](g.world, g.player)
	if !ok {
		return fmt.Errorf("game: player entity missing Transform")
	}
	t.Position = rl.Vector3{X: save.FoxPosition.X, Y: save.FoxPosition.Y, Z: save.FoxPosition.Z}
	t.Yaw = save.FoxYaw

	cam := ecs.MustGetResource[camera.ThirdPerson](g.world)
	cam.Yaw = save.CameraYaw
	cam.Pitch = save.CameraPitch

	clock := ecs.MustGetResource[worldtime.Clock](g.world)
	clock.SetTimeOfDay(save.TimeOfDay)
	clock.Day = save.Day

	g.playtimeSecs = save.PlaytimeSecs
	return nil
}

// Update ticks the world clock, camera, fox simulation, ambient creatures,
// and the audio subsystem (ambience mix, music, queued one-shot SFX).
// Quicksave is bound to the Confirm action for this basic build.
func (g *FoxDot) Update(dt float32, in input.Frame) error {
	g.playtimeSecs += float64(dt)

	clock := ecs.MustGetResource[worldtime.Clock](g.world)
	clock.Advance(float64(dt))

	CameraSystem(g.world, dt, in.Look, 0)
	cam := ecs.MustGetResource[camera.ThirdPerson](g.world)

	events := ecs.MustGetResource[audio.EventQueue](g.world)
	FoxSystem(g.world, dt, in, cam.Yaw, events)
	CreatureVoiceSystem(g.world, dt, events)

	// Quicksave doubles as this build's only "UI interaction" — there's no
	// real UI yet, but wiring its confirmation sound through the same
	// EventQueue any future menu/HUD code would use demonstrates that path
	// end to end.
	if in.Confirm.Pressed {
		events.Push("ui_confirm", 1)
		if err := g.save(autosaveSlot); err != nil {
			return fmt.Errorf("game: quicksave: %w", err)
		}
	}

	g.updateAudio(dt, clock.Phase(), events)
	return nil
}

func (g *FoxDot) updateAudio(dt float32, phase worldtime.Phase, events *audio.EventQueue) {
	t, _ := ecs.Get[comp.Transform](g.world, g.player)
	proximity := campfireProximity(g.world, t.Position)

	g.ambience.Update(dt, phase, proximity)
	g.music.Update(dt, phase)

	for _, e := range events.Drain() {
		g.audioEngine.PlaySFX(e.Name, e.Volume)
	}
	g.audioEngine.Update()
}

// Draw renders the forest, every entity in it, the fox's action marker,
// and an optional debug overlay.
func (g *FoxDot) Draw() {
	cam := ecs.MustGetResource[camera.ThirdPerson](g.world)

	rl.BeginMode3D(cam.RLCamera())
	world.DrawGround(world.DefaultGroundY, g.cfg.Graphics.ShowGrid)
	RenderSystem(g.world, g.art)
	g.drawActionMarker()
	rl.EndMode3D()

	if g.cfg.Graphics.ShowDebugOverlay {
		g.drawDebugOverlay()
	}
}

// drawActionMarker shows a small marker above the fox's head for the
// current action at a glance (handy during development, before there's
// real animation).
func (g *FoxDot) drawActionMarker() {
	t, _ := ecs.Get[comp.Transform](g.world, g.player)
	a, _ := ecs.Get[comp.ActionState](g.world, g.player)
	if a.Action == sim.ActionNone {
		return
	}

	markerColor := rl.Yellow
	if a.Action == sim.ActionSwipe {
		markerColor = rl.Red
	}
	facing := rl.Vector3{
		X: float32(math.Sin(float64(t.Yaw))),
		Y: 0,
		Z: float32(math.Cos(float64(t.Yaw))),
	}
	marker := rl.Vector3Add(rl.Vector3{X: t.Position.X, Y: t.Position.Y + 0.75, Z: t.Position.Z}, rl.Vector3Scale(facing, 0.4))
	rl.DrawSphere(marker, 0.08, markerColor)
}

func (g *FoxDot) drawDebugOverlay() {
	rl.DrawFPS(10, 10)

	t, _ := ecs.Get[comp.Transform](g.world, g.player)
	v, _ := ecs.Get[comp.Velocity](g.world, g.player)
	a, _ := ecs.Get[comp.ActionState](g.world, g.player)
	cam := ecs.MustGetResource[camera.ThirdPerson](g.world)
	clock := ecs.MustGetResource[worldtime.Clock](g.world)

	tod := clock.TimeOfDay() * 24
	lines := []string{
		fmt.Sprintf("fox pos: (%.1f, %.1f, %.1f)  yaw: %.0f deg", t.Position.X, t.Position.Y, t.Position.Z, rl.Rad2deg*t.Yaw),
		fmt.Sprintf("grounded: %v  action: %s", v.Grounded, a.Action),
		fmt.Sprintf("camera dist: %.1f  yaw: %.0f deg  pitch: %.0f deg", cam.Distance(), rl.Rad2deg*cam.Yaw, rl.Rad2deg*cam.Pitch),
		fmt.Sprintf("day %d, %02d:%02d (%s)  music: %s", clock.Day, int(tod)%24, int(tod*60)%60, clock.Phase(), g.music.Current()),
		fmt.Sprintf("playtime: %s", time.Duration(g.playtimeSecs*float64(time.Second)).Round(time.Second)),
		"move: WASD/stick  jump: Space/A  sprint: Shift/LT  nibble: E/X  swipe: F/B  quicksave: Enter",
	}
	for i, line := range lines {
		rl.DrawText(line, 10, int32(34+18*i), 18, rl.White)
	}
}

// Shutdown autosaves, releases loaded models, and closes the audio device
// on the way out so quitting normally never loses progress or leaks
// GPU/audio resources.
func (g *FoxDot) Shutdown() {
	_ = g.save(autosaveSlot)
	g.art.Unload()
	if g.audioEngine != nil {
		g.audioEngine.Close()
	}
}

func (g *FoxDot) save(slot string) error {
	t, _ := ecs.Get[comp.Transform](g.world, g.player)
	cam := ecs.MustGetResource[camera.ThirdPerson](g.world)
	clock := ecs.MustGetResource[worldtime.Clock](g.world)

	pos := t.Position
	return g.store.SaveGameSlot(slot, storage.SaveGame{
		Version:      storage.CurrentSaveVersion,
		SavedAt:      time.Now().UTC(),
		FoxPosition:  storage.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
		FoxYaw:       t.Yaw,
		CameraYaw:    cam.Yaw,
		CameraPitch:  cam.Pitch,
		PlaytimeSecs: g.playtimeSecs,
		TimeOfDay:    clock.TimeOfDay(),
		Day:          clock.Day,
	})
}
