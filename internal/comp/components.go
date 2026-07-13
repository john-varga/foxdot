// Package comp holds the ECS component and tag types shared across
// internal/world and internal/game (see internal/ecs for the World/Entity
// machinery itself). Keeping them in one neutral, dependency-light package
// lets world spawn prop entities and game spawn/drive the fox entity
// without either package needing to import the other's internals.
package comp

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/sim"
)

// Transform is an entity's world-space placement: ground-contact position
// plus facing (yaw, radians; 0 faces world +Z, the engine's convention —
// see internal/assets for how model-specific forward axes are corrected).
type Transform struct {
	Position rl.Vector3
	Yaw      float32
}

// Velocity is an entity's current linear velocity plus whether it's
// resting on a walkable surface (ground or a prop). Only entities driven
// by sim.Step need this; static props don't.
type Velocity struct {
	Vector   rl.Vector3
	Grounded bool
}

// ModelRender draws an entity as a catalog asset (see internal/assets),
// planted on the ground at its Transform.Position and facing
// Transform.Yaw. Scale <= 0 means "use 1", matching world.Prop's old
// convention.
type ModelRender struct {
	AssetName string
	Scale     float32
}

// EffectiveScale returns Scale, defaulting to 1 for the zero value.
func (m ModelRender) EffectiveScale() float32 {
	if m.Scale <= 0 {
		return 1
	}
	return m.Scale
}

// ActionState mirrors sim.State's one-shot action fields as a component, so
// any entity driven by sim.Step (today, just the player fox) can be pure
// ECS data instead of a bespoke Fox object.
type ActionState struct {
	Action sim.Action
	Timer  float32
}

// MoveParams tunes sim.Step's movement feel for an entity. A component
// (not a resource) so different creatures could someday have different
// feel while sharing the same movement system.
type MoveParams struct {
	Params sim.Params
}

// PlayerControlled tags the single entity driven by the local player's
// input.Frame.
type PlayerControlled struct{}

// Footsteps periodically queues a footstep sound while its entity is
// grounded and moving under player control (see the game package's
// FoxSystem), rotating through a set of variants for a touch of realism.
type Footsteps struct {
	Timer   float32 // system-owned countdown to the next footstep
	Variant int     // rotates through footstep sound variants
}

// Prop tags static forest scenery (trees, rocks, bushes, ...) that
// participates in ground-height/collision queries (see internal/world).
// Decorative ambient creatures use Transform+ModelRender without this tag,
// so they don't block movement.
type Prop struct{}

// CampfireSource tags the entity the audio system measures the player's
// proximity to, to fade the campfire ambience loop in/out (see
// internal/audio's AmbienceWeights).
type CampfireSource struct{}

// CreatureVoice periodically queues an ambient one-shot sound for a
// decorative creature entity (a bird chirp, a rabbit hop, ...), consumed by
// a system that ticks Timer down and pushes an audio.SoundEvent when it
// elapses, then rolls a new random interval.
type CreatureVoice struct {
	Sounds      []string
	MinInterval float32
	MaxInterval float32
	Volume      float32

	Timer float32 // seconds remaining until the next sound; system-owned
}
