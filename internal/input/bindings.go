package input

import rl "github.com/gen2brain/raylib-go/raylib"

// KeyBindings maps gameplay actions to raylib keyboard key codes. A value of
// 0 (rl.KeyNull) means "unbound".
type KeyBindings struct {
	MoveForward  int32 `json:"moveForward"`
	MoveBackward int32 `json:"moveBackward"`
	MoveLeft     int32 `json:"moveLeft"`
	MoveRight    int32 `json:"moveRight"`
	Jump         int32 `json:"jump"`
	Sprint       int32 `json:"sprint"`
	Nibble       int32 `json:"nibble"`
	Swipe        int32 `json:"swipe"`
	Pause        int32 `json:"pause"`
	Confirm      int32 `json:"confirm"`
	Cancel       int32 `json:"cancel"`
}

// DefaultKeyBindings returns a sensible WASD + space layout.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		MoveForward:  rl.KeyW,
		MoveBackward: rl.KeyS,
		MoveLeft:     rl.KeyA,
		MoveRight:    rl.KeyD,
		Jump:         rl.KeySpace,
		Sprint:       rl.KeyLeftShift,
		Nibble:       rl.KeyE,
		Swipe:        rl.KeyF,
		Pause:        rl.KeyEscape,
		Confirm:      rl.KeyEnter,
		Cancel:       rl.KeyBackspace,
	}
}

// GamepadBindings maps gameplay actions to raylib gamepad button/axis codes.
type GamepadBindings struct {
	Jump      int32 `json:"jump"`
	Sprint    int32 `json:"sprint"`
	Nibble    int32 `json:"nibble"`
	Swipe     int32 `json:"swipe"`
	Pause     int32 `json:"pause"`
	Confirm   int32 `json:"confirm"`
	Cancel    int32 `json:"cancel"`
	MoveAxisX int32 `json:"moveAxisX"`
	MoveAxisY int32 `json:"moveAxisY"`
	LookAxisX int32 `json:"lookAxisX"`
	LookAxisY int32 `json:"lookAxisY"`
}

// DefaultGamepadBindings returns a standard Xbox/PlayStation-style layout.
func DefaultGamepadBindings() GamepadBindings {
	return GamepadBindings{
		Jump:      rl.GamepadButtonRightFaceDown,  // A / Cross
		Sprint:    rl.GamepadButtonLeftTrigger2,   // LT / L2
		Nibble:    rl.GamepadButtonRightFaceLeft,  // X / Square
		Swipe:     rl.GamepadButtonRightFaceRight, // B / Circle
		Pause:     rl.GamepadButtonMiddleRight,    // Start / Options
		Confirm:   rl.GamepadButtonRightFaceDown,  // A / Cross
		Cancel:    rl.GamepadButtonRightFaceRight, // B / Circle
		MoveAxisX: rl.GamepadAxisLeftX,
		MoveAxisY: rl.GamepadAxisLeftY,
		LookAxisX: rl.GamepadAxisRightX,
		LookAxisY: rl.GamepadAxisRightY,
	}
}

// Config holds every tweakable setting the controller layer needs: which
// bindings map to which actions, plus sensitivity/deadzone tuning. It is
// meant to be embedded in the game-wide config.Config and saved/loaded as
// part of it, so players (or we, during development) can retune input
// without touching code.
type Config struct {
	Keyboard KeyBindings     `json:"keyboard"`
	Gamepad  GamepadBindings `json:"gamepad"`

	GamepadIndex int32 `json:"gamepadIndex"`

	MoveDeadzone float32 `json:"moveDeadzone"`
	LookDeadzone float32 `json:"lookDeadzone"`

	MouseLookSensitivity float32 `json:"mouseLookSensitivity"`
	StickLookSensitivity float32 `json:"stickLookSensitivity"`

	InvertY bool `json:"invertY"`
}

// DefaultConfig returns reasonable defaults for a fresh install.
func DefaultConfig() Config {
	return Config{
		Keyboard:             DefaultKeyBindings(),
		Gamepad:              DefaultGamepadBindings(),
		GamepadIndex:         0,
		MoveDeadzone:         0.2,
		LookDeadzone:         0.15,
		MouseLookSensitivity: 0.15,
		StickLookSensitivity: 2.5,
		InvertY:              false,
	}
}
