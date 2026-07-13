package input

import "testing"

// fakeSource is a stub Source for testing Manager without touching raylib.
type fakeSource struct {
	frame Frame
}

func (f fakeSource) Poll(cfg Config) Frame {
	return f.frame
}

func TestManagerMergesSources(t *testing.T) {
	keyboard := fakeSource{frame: Frame{Move: Vector2{X: 0, Y: 1}, Jump: ButtonState{Pressed: true}}}
	gamepad := fakeSource{frame: Frame{Move: Vector2{X: 0, Y: 0}}}

	mgr := NewManager(DefaultConfig(), keyboard, gamepad)
	got := mgr.Poll()

	if got.Move != (Vector2{X: 0, Y: 1}) {
		t.Fatalf("expected merged move from keyboard, got %+v", got.Move)
	}
	if !got.Jump.Pressed {
		t.Fatalf("expected jump to be pressed")
	}
}

func TestManagerSetConfig(t *testing.T) {
	mgr := NewManager(DefaultConfig())
	custom := DefaultConfig()
	custom.InvertY = true
	mgr.SetConfig(custom)
	if !mgr.Config().InvertY {
		t.Fatalf("expected updated config to stick")
	}
}
