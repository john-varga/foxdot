package game

import (
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
	"foxdot/internal/audio"
	"foxdot/internal/camera"
	"foxdot/internal/comp"
	"foxdot/internal/ecs"
	"foxdot/internal/input"
	"foxdot/internal/sim"
	"foxdot/internal/world"
)

// footstepInterval is roughly how often a footstep sound plays while
// walking at normal speed; sprinting shortens it proportionally.
const footstepInterval float32 = 0.33

var footstepSounds = []string{"grass_footstep_01", "grass_footstep_02", "grass_footstep_03", "grass_footstep_04"}

// FoxSystem advances every player-controlled entity's sim.State by dt and
// queues the sound effects that follow from what happened (jump, footstep,
// swipe) — the sim package itself never touches audio, so this is the
// glue between "what the simulation did" and "what that should sound
// like".
func FoxSystem(w *ecs.World, dt float32, in input.Frame, cameraYaw float32, events *audio.EventQueue) {
	groundHeight := func(x, z float32) float32 { return world.HeightAt(w, world.DefaultGroundY, x, z) }

	type stepper struct {
		e  ecs.Entity
		t  *comp.Transform
		v  *comp.Velocity
		a  *comp.ActionState
		mp *comp.MoveParams
	}
	var steppers []stepper
	ecs.Each3(w, func(e ecs.Entity, t *comp.Transform, v *comp.Velocity, a *comp.ActionState) {
		if !ecs.Has[comp.PlayerControlled](w, e) {
			return
		}
		mp, ok := ecs.Get[comp.MoveParams](w, e)
		if !ok {
			return
		}
		steppers = append(steppers, stepper{e: e, t: t, v: v, a: a, mp: mp})
	})

	for _, s := range steppers {
		state := sim.State{Position: s.t.Position, Velocity: s.v.Vector, Yaw: s.t.Yaw, Grounded: s.v.Grounded, Action: s.a.Action, ActionTimer: s.a.Timer}
		ev := sim.Step(&state, s.mp.Params, dt, in, cameraYaw, groundHeight)

		s.t.Position, s.t.Yaw = state.Position, state.Yaw
		s.v.Vector, s.v.Grounded = state.Velocity, state.Grounded
		s.a.Action, s.a.Timer = state.Action, state.ActionTimer

		queueFoxEvents(w, s.e, s.v, s.a, ev, dt, in, events)
	}
}

func queueFoxEvents(w *ecs.World, e ecs.Entity, v *comp.Velocity, a *comp.ActionState, ev sim.Events, dt float32, in input.Frame, events *audio.EventQueue) {
	if ev.Jumped {
		events.Push("jump", 1)
	}
	if ev.ActionStart == sim.ActionSwipe {
		events.Push("fox_bark", 0.8)
	}

	fs, ok := ecs.Get[comp.Footsteps](w, e)
	if !ok {
		return
	}
	moving := v.Grounded && a.Action == sim.ActionNone && (in.Move.X != 0 || in.Move.Y != 0)
	if !moving {
		fs.Timer = 0
		return
	}
	interval := footstepInterval
	if in.Sprint.Down {
		interval *= 0.65
	}
	fs.Timer -= dt
	if fs.Timer <= 0 {
		fs.Variant = (fs.Variant + 1) % len(footstepSounds)
		events.Push(footstepSounds[fs.Variant], 0.5)
		fs.Timer = interval
	}
}

// CameraSystem re-centers the third-person camera resource on the (single)
// player-controlled entity's position.
func CameraSystem(w *ecs.World, dt float32, look input.Vector2, zoomDelta float32) {
	cam := ecs.MustGetResource[camera.ThirdPerson](w)

	var target rl.Vector3
	ecs.Each2(w, func(_ ecs.Entity, t *comp.Transform, _ *comp.PlayerControlled) {
		target = t.Position
	})
	cam.Update(dt, target, look, zoomDelta)
}

// CreatureVoiceSystem ticks every CreatureVoice's timer down and queues a
// random sound from its list once it elapses, then rolls a fresh interval
// — a stand-in for real wildlife AI that still makes the forest sound
// inhabited.
func CreatureVoiceSystem(w *ecs.World, dt float32, events *audio.EventQueue) {
	ecs.Each(w, func(_ ecs.Entity, cv *comp.CreatureVoice) {
		if len(cv.Sounds) == 0 {
			return
		}
		cv.Timer -= dt
		if cv.Timer > 0 {
			return
		}
		events.Push(cv.Sounds[rand.Intn(len(cv.Sounds))], cv.Volume)
		lo, hi := cv.MinInterval, cv.MaxInterval
		if hi <= lo {
			hi = lo + 1
		}
		cv.Timer = lo + rand.Float32()*(hi-lo)
	})
}

// campfireProximityRadius is the distance (meters) at which the campfire
// ambience loop has faded out completely.
const campfireProximityRadius = 8

// campfireProximity returns how close pos is to the nearest CampfireSource
// entity, normalized so 0 means "right on top of it" and 1+ means "far
// enough to not hear it at all" (see audio.AmbienceWeights).
func campfireProximity(w *ecs.World, pos rl.Vector3) float32 {
	best := float32(1)
	found := false
	ecs.Each2(w, func(_ ecs.Entity, t *comp.Transform, _ *comp.CampfireSource) {
		d := rl.Vector3Distance(pos, t.Position) / campfireProximityRadius
		if !found || d < best {
			best, found = d, true
		}
	})
	return best
}

// RenderSystem draws every entity that has both a Transform and a
// ModelRender, uniformly — the fox, forest props, and ambient creatures
// alike — regardless of which package spawned it.
func RenderSystem(w *ecs.World, store *assets.Store) {
	ecs.Each2(w, func(_ ecs.Entity, t *comp.Transform, m *comp.ModelRender) {
		_ = store.Draw(m.AssetName, t.Position, t.Yaw, m.EffectiveScale())
	})
}
