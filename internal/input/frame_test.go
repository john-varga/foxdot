package input

import "testing"

func TestMergePicksLargerAxis(t *testing.T) {
	idle := Frame{Move: Vector2{X: 0, Y: 0}}
	active := Frame{Move: Vector2{X: 0, Y: 1}}

	got := idle.Merge(active)
	if got.Move != active.Move {
		t.Fatalf("expected idle.Merge(active) to prefer the active axis, got %+v", got.Move)
	}

	got = active.Merge(idle)
	if got.Move != active.Move {
		t.Fatalf("expected active.Merge(idle) to keep the active axis, got %+v", got.Move)
	}
}

func TestMergeOrsButtons(t *testing.T) {
	a := Frame{Jump: ButtonState{Pressed: true}}
	b := Frame{Jump: ButtonState{Down: true}}

	got := a.Merge(b)
	if !got.Jump.Pressed || !got.Jump.Down {
		t.Fatalf("expected merged button to be pressed and down, got %+v", got.Jump)
	}
}
