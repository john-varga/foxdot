package input

import rl "github.com/gen2brain/raylib-go/raylib"

// GamepadSource reads raylib gamepad state for a single gamepad slot and
// turns it into a Frame according to Config's bindings, deadzones and
// sensitivity settings. Supports any controller raylib/GLFW recognizes
// (Xbox, PlayStation, generic HID) on macOS, Windows and Linux.
type GamepadSource struct{}

// NewGamepadSource constructs a gamepad input source.
func NewGamepadSource() *GamepadSource {
	return &GamepadSource{}
}

// Poll implements Source. Returns a zero Frame if the configured gamepad
// index isn't connected.
func (g *GamepadSource) Poll(cfg Config) Frame {
	idx := cfg.GamepadIndex
	if !rl.IsGamepadAvailable(idx) {
		return Frame{}
	}
	gb := cfg.Gamepad

	move := Vector2{
		X: applyDeadzone(rl.GetGamepadAxisMovement(idx, gb.MoveAxisX), cfg.MoveDeadzone),
		// Raylib/GLFW report the stick's Y axis as positive-down; flip so
		// that positive Move.Y consistently means "forward" for callers.
		Y: -applyDeadzone(rl.GetGamepadAxisMovement(idx, gb.MoveAxisY), cfg.MoveDeadzone),
	}
	move = clampMagnitude(move, 1)

	lookX := applyDeadzone(rl.GetGamepadAxisMovement(idx, gb.LookAxisX), cfg.LookDeadzone)
	lookY := applyDeadzone(rl.GetGamepadAxisMovement(idx, gb.LookAxisY), cfg.LookDeadzone)
	if cfg.InvertY {
		lookY = -lookY
	}

	return Frame{
		Move: move,
		Look: Vector2{
			X: lookX * cfg.StickLookSensitivity,
			Y: lookY * cfg.StickLookSensitivity,
		},
		Jump:    gamepadButtonState(idx, gb.Jump),
		Sprint:  gamepadButtonState(idx, gb.Sprint),
		Nibble:  gamepadButtonState(idx, gb.Nibble),
		Swipe:   gamepadButtonState(idx, gb.Swipe),
		Pause:   gamepadButtonState(idx, gb.Pause),
		Confirm: gamepadButtonState(idx, gb.Confirm),
		Cancel:  gamepadButtonState(idx, gb.Cancel),
	}
}

func gamepadButtonState(gamepad, button int32) ButtonState {
	return ButtonState{
		Down:     rl.IsGamepadButtonDown(gamepad, button),
		Pressed:  rl.IsGamepadButtonPressed(gamepad, button),
		Released: rl.IsGamepadButtonReleased(gamepad, button),
	}
}
