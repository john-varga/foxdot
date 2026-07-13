package input

import rl "github.com/gen2brain/raylib-go/raylib"

// KeyboardMouseSource reads raylib keyboard and mouse state and turns it
// into a Frame according to Config's key bindings and sensitivity settings.
type KeyboardMouseSource struct{}

// NewKeyboardMouseSource constructs a keyboard/mouse input source.
func NewKeyboardMouseSource() *KeyboardMouseSource {
	return &KeyboardMouseSource{}
}

// Poll implements Source.
func (k *KeyboardMouseSource) Poll(cfg Config) Frame {
	kb := cfg.Keyboard

	move := Vector2{}
	if rl.IsKeyDown(kb.MoveForward) {
		move.Y += 1
	}
	if rl.IsKeyDown(kb.MoveBackward) {
		move.Y -= 1
	}
	if rl.IsKeyDown(kb.MoveRight) {
		move.X += 1
	}
	if rl.IsKeyDown(kb.MoveLeft) {
		move.X -= 1
	}
	move = clampMagnitude(move, 1)

	delta := rl.GetMouseDelta()
	lookY := delta.Y
	if cfg.InvertY {
		lookY = -lookY
	}

	return Frame{
		Move: move,
		Look: Vector2{
			X: delta.X * cfg.MouseLookSensitivity,
			Y: lookY * cfg.MouseLookSensitivity,
		},
		Jump:    keyState(kb.Jump),
		Sprint:  keyState(kb.Sprint),
		Nibble:  keyState(kb.Nibble),
		Swipe:   keyState(kb.Swipe),
		Pause:   keyState(kb.Pause),
		Confirm: keyState(kb.Confirm),
		Cancel:  keyState(kb.Cancel),
	}
}

func keyState(key int32) ButtonState {
	if key == rl.KeyNull {
		return ButtonState{}
	}
	return ButtonState{
		Down:     rl.IsKeyDown(key),
		Pressed:  rl.IsKeyPressed(key),
		Released: rl.IsKeyReleased(key),
	}
}
