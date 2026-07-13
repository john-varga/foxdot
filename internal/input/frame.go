// Package input provides a controller abstraction layer that turns raw
// keyboard/mouse/gamepad state into a normalized, engine-agnostic Frame.
// Game and simulation code should depend only on Frame, never on raylib's
// input functions directly, so gameplay logic stays testable and adding a
// new input device (or remapping bindings) never touches gameplay code.
package input

// Vector2 is a small 2D vector used for input axes. It intentionally avoids
// depending on raylib so this package's core types remain engine-agnostic.
type Vector2 struct {
	X, Y float32
}

// ButtonState captures a single action button's state for one frame.
// Pressed/Released are edge triggers (true for exactly one frame).
type ButtonState struct {
	Down     bool
	Pressed  bool
	Released bool
}

// Frame is a normalized snapshot of player input for a single game tick.
// It's what Sources produce and what gameplay/simulation code consumes.
type Frame struct {
	// Move is the desired movement direction in local/ground space:
	// X = strafe (-1 left .. 1 right), Y = forward (-1 back .. 1 forward).
	// Magnitude is clamped to [0, 1].
	Move Vector2

	// Look is camera/orbit rotation input for this frame: X = yaw delta,
	// Y = pitch delta, already scaled by sensitivity.
	Look Vector2

	Jump    ButtonState
	Sprint  ButtonState
	Nibble  ButtonState // primary interact/nibble action
	Swipe   ButtonState // playful swipe action
	Pause   ButtonState
	Confirm ButtonState
	Cancel  ButtonState
}

// Merge combines this frame with another, treating it as the earlier /
// lower-priority source. Buttons are OR'd together; axis values take
// whichever of the two has the larger magnitude, so an idle device never
// overrides an active one.
func (f Frame) Merge(other Frame) Frame {
	out := f
	out.Move = pickLarger(f.Move, other.Move)
	out.Look = pickLarger(f.Look, other.Look)
	out.Jump = orButton(f.Jump, other.Jump)
	out.Sprint = orButton(f.Sprint, other.Sprint)
	out.Nibble = orButton(f.Nibble, other.Nibble)
	out.Swipe = orButton(f.Swipe, other.Swipe)
	out.Pause = orButton(f.Pause, other.Pause)
	out.Confirm = orButton(f.Confirm, other.Confirm)
	out.Cancel = orButton(f.Cancel, other.Cancel)
	return out
}

func pickLarger(a, b Vector2) Vector2 {
	if sqrMag(b) > sqrMag(a) {
		return b
	}
	return a
}

func sqrMag(v Vector2) float32 {
	return v.X*v.X + v.Y*v.Y
}

func orButton(a, b ButtonState) ButtonState {
	return ButtonState{
		Down:     a.Down || b.Down,
		Pressed:  a.Pressed || b.Pressed,
		Released: a.Released || b.Released,
	}
}
