// Package sim contains the gameplay simulation logic — currently just the
// fox's movement, jumping and action state machine. It depends only on
// input.Frame and raylib's plain math types/helpers (no window, no GPU, no
// global state), so it can be exercised with ordinary table-driven tests
// even though it will eventually grow into the "game" part of this
// game/framework test.
package sim

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/input"
)

// Action identifies a one-shot animation-like action the fox can perform.
type Action int

const (
	ActionNone Action = iota
	ActionNibble
	ActionSwipe
)

// String implements fmt.Stringer for readable test failures/debug overlays.
func (a Action) String() string {
	switch a {
	case ActionNibble:
		return "nibble"
	case ActionSwipe:
		return "swipe"
	default:
		return "none"
	}
}

// State is the fox's full simulated state at a point in time.
type State struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Yaw      float32 // facing direction, radians, 0 = -Z
	Grounded bool

	Action      Action
	ActionTimer float32 // seconds remaining on the current action
}

// Params tunes the fox's movement feel. Exposed so it can live under
// config.Config and be retuned without a recompile.
type Params struct {
	MoveSpeed        float32 `json:"moveSpeed"`
	SprintMultiplier float32 `json:"sprintMultiplier"`
	JumpVelocity     float32 `json:"jumpVelocity"`
	Gravity          float32 `json:"gravity"`
	TurnSmoothTime   float32 `json:"turnSmoothTime"` // seconds; how quickly facing catches up to movement
	NibbleDuration   float32 `json:"nibbleDuration"`
	SwipeDuration    float32 `json:"swipeDuration"`
}

// DefaultParams returns reasonable defaults for a small, nimble fox.
func DefaultParams() Params {
	return Params{
		MoveSpeed:        4.5,
		SprintMultiplier: 1.8,
		JumpVelocity:     6.0,
		Gravity:          18.0,
		TurnSmoothTime:   0.08,
		NibbleDuration:   0.6,
		SwipeDuration:    0.45,
	}
}

// GroundHeightFunc reports the height of the walkable surface (ground or a
// prop the fox can stand on) below world position (x, z). Injected rather
// than imported so this package never depends on the world package.
type GroundHeightFunc func(x, z float32) float32

// Fox is the stateful simulation for the player-controlled fox.
type Fox struct {
	State  State
	Params Params
}

// NewFox creates a fox at the origin with the given tuning parameters.
func NewFox(params Params) *Fox {
	return &Fox{Params: params}
}

// Step advances the simulation by dt using this frame's input. cameraYaw
// rotates movement into camera-relative space (so "forward" always means
// "away from the camera"), and groundHeight resolves collision against the
// world's ground/props.
func (f *Fox) Step(dt float32, in input.Frame, cameraYaw float32, groundHeight GroundHeightFunc) {
	if dt <= 0 {
		return
	}

	f.updateAction(dt, in)
	f.applyMovement(dt, in, cameraYaw)
	f.applyVerticalMotion(dt, in, groundHeight)
}

func (f *Fox) updateAction(dt float32, in input.Frame) {
	s := &f.State
	if s.Action != ActionNone {
		s.ActionTimer -= dt
		if s.ActionTimer <= 0 {
			s.Action = ActionNone
			s.ActionTimer = 0
		}
		return
	}
	// If both are pressed on the same frame, swipe wins (it's the more
	// "urgent" reflexive action).
	switch {
	case in.Swipe.Pressed:
		s.Action = ActionSwipe
		s.ActionTimer = f.Params.SwipeDuration
	case in.Nibble.Pressed:
		s.Action = ActionNibble
		s.ActionTimer = f.Params.NibbleDuration
	}
}

func (f *Fox) applyMovement(dt float32, in input.Frame, cameraYaw float32) {
	p := &f.Params
	s := &f.State

	// Performing an action roots the fox in place, like a real animal
	// pausing to nibble or swipe at something.
	if s.Action != ActionNone {
		return
	}

	move := in.Move
	if move.X == 0 && move.Y == 0 {
		return
	}

	speed := p.MoveSpeed
	if in.Sprint.Down {
		speed *= p.SprintMultiplier
	}

	sinYaw := float32(math.Sin(float64(cameraYaw)))
	cosYaw := float32(math.Cos(float64(cameraYaw)))

	// Rotate the input axes (X = strafe, Y = forward) by the camera's yaw
	// so movement is always relative to where the camera is looking.
	worldX := move.X*cosYaw + move.Y*sinYaw
	worldZ := -move.X*sinYaw + move.Y*cosYaw

	s.Position.X += worldX * speed * dt
	s.Position.Z += worldZ * speed * dt

	targetYaw := float32(math.Atan2(float64(worldX), float64(worldZ)))
	s.Yaw = turnTowards(s.Yaw, targetYaw, dt, p.TurnSmoothTime)
}

func (f *Fox) applyVerticalMotion(dt float32, in input.Frame, groundHeight GroundHeightFunc) {
	p := &f.Params
	s := &f.State

	ground := float32(0)
	if groundHeight != nil {
		ground = groundHeight(s.Position.X, s.Position.Z)
	}

	if s.Grounded && in.Jump.Pressed && s.Action == ActionNone {
		s.Velocity.Y = p.JumpVelocity
		s.Grounded = false
	}

	s.Velocity.Y -= p.Gravity * dt
	s.Position.Y += s.Velocity.Y * dt

	if s.Position.Y <= ground {
		s.Position.Y = ground
		s.Velocity.Y = 0
		s.Grounded = true
	} else {
		s.Grounded = false
	}
}

// turnTowards smoothly rotates `from` toward `to` (both radians), taking the
// shortest path around the circle, at a rate derived from smoothTime.
func turnTowards(from, to, dt, smoothTime float32) float32 {
	if smoothTime <= 0 {
		return to
	}
	delta := wrapAngle(to - from)
	factor := 1 - float32(math.Exp(-float64(dt/smoothTime)))
	return wrapAngle(from + delta*factor)
}

// wrapAngle normalizes an angle in radians to (-pi, pi].
func wrapAngle(a float32) float32 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a <= -math.Pi {
		a += 2 * math.Pi
	}
	return a
}
