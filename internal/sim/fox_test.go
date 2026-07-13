package sim

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/input"
)

func flatGround(y float32) GroundHeightFunc {
	return func(x, z float32) float32 { return y }
}

func TestFoxFallsAndLandsOnGround(t *testing.T) {
	s := State{Position: rl.Vector3{Y: 5}}
	p := DefaultParams()
	ground := flatGround(0)

	landed := false
	for i := 0; i < 300; i++ { // plenty of time to fall and settle
		ev := Step(&s, p, 1.0/60.0, input.Frame{}, 0, ground)
		if ev.Landed {
			landed = true
		}
	}

	if !s.Grounded {
		t.Fatalf("expected fox to be grounded after falling, got airborne")
	}
	if s.Position.Y != 0 {
		t.Fatalf("expected fox to rest at ground height 0, got %v", s.Position.Y)
	}
	if !landed {
		t.Fatalf("expected a Landed event to fire while settling")
	}
}

func TestFoxJumpsWhenGroundedAndRises(t *testing.T) {
	s := State{Grounded: true}
	p := DefaultParams()
	ground := flatGround(0)

	ev := Step(&s, p, 1.0/60.0, input.Frame{Jump: input.ButtonState{Pressed: true}}, 0, ground)

	if s.Grounded {
		t.Fatalf("expected fox to leave the ground after jumping")
	}
	if s.Velocity.Y <= 0 {
		t.Fatalf("expected positive upward velocity after jump, got %v", s.Velocity.Y)
	}
	if !ev.Jumped {
		t.Fatalf("expected a Jumped event to fire")
	}
}

func TestFoxCannotJumpWhileAirborne(t *testing.T) {
	s := State{Grounded: false, Velocity: rl.Vector3{Y: 1}}
	p := DefaultParams()
	ground := flatGround(0)

	ev := Step(&s, p, 1.0/60.0, input.Frame{Jump: input.ButtonState{Pressed: true}}, 0, ground)

	// Velocity should just be gravity-integrated, not re-boosted by JumpVelocity.
	expected := float32(1) - DefaultParams().Gravity*(1.0/60.0)
	if math.Abs(float64(s.Velocity.Y-expected)) > 1e-4 {
		t.Fatalf("expected velocity %.5f from gravity alone, got %.5f", expected, s.Velocity.Y)
	}
	if ev.Jumped {
		t.Fatalf("did not expect a Jumped event while airborne")
	}
}

func TestFoxMovesForwardRelativeToCamera(t *testing.T) {
	s := State{Grounded: true}
	p := DefaultParams()
	ground := flatGround(0)

	// Camera facing +90 degrees (yaw = pi/2): "forward" should now push the
	// fox along world +X instead of +Z.
	cameraYaw := float32(math.Pi / 2)
	frame := input.Frame{Move: input.Vector2{X: 0, Y: 1}}

	for i := 0; i < 30; i++ {
		Step(&s, p, 1.0/60.0, frame, cameraYaw, ground)
	}

	if s.Position.X <= 0.1 {
		t.Fatalf("expected fox to move along +X when camera yaw is 90deg, got position %+v", s.Position)
	}
	if math.Abs(float64(s.Position.Z)) > 0.1 {
		t.Fatalf("expected negligible Z movement, got position %+v", s.Position)
	}
}

func TestFoxCanStandOnRaisedProp(t *testing.T) {
	s := State{Position: rl.Vector3{Y: 5}}
	p := DefaultParams()
	// A "rock" that's 2 units tall directly under the fox.
	raised := func(x, z float32) float32 { return 2 }

	for i := 0; i < 300; i++ {
		Step(&s, p, 1.0/60.0, input.Frame{}, 0, raised)
	}

	if s.Position.Y != 2 {
		t.Fatalf("expected fox to rest on top of the raised surface at y=2, got %v", s.Position.Y)
	}
	if !s.Grounded {
		t.Fatalf("expected fox to be grounded while standing on the prop")
	}
}

func TestNibbleActionRootsFoxAndExpires(t *testing.T) {
	s := State{Grounded: true}
	p := DefaultParams()
	ground := flatGround(0)

	ev := Step(&s, p, 1.0/60.0, input.Frame{Nibble: input.ButtonState{Pressed: true}}, 0, ground)
	if s.Action != ActionNibble {
		t.Fatalf("expected nibble action to start, got %v", s.Action)
	}
	if ev.ActionStart != ActionNibble {
		t.Fatalf("expected an ActionStart=nibble event, got %v", ev.ActionStart)
	}

	startPos := s.Position
	moveWhileNibbling := input.Frame{Move: input.Vector2{Y: 1}}
	for i := 0; i < 10; i++ {
		Step(&s, p, 1.0/60.0, moveWhileNibbling, 0, ground)
	}
	if s.Position != startPos {
		t.Fatalf("expected fox to stay rooted while nibbling, moved from %+v to %+v", startPos, s.Position)
	}

	// Run past the nibble duration and confirm the action clears.
	dur := DefaultParams().NibbleDuration
	steps := int(dur/(1.0/60.0)) + 5
	for i := 0; i < steps; i++ {
		Step(&s, p, 1.0/60.0, input.Frame{}, 0, ground)
	}
	if s.Action != ActionNone {
		t.Fatalf("expected nibble action to expire, still %v", s.Action)
	}
}

func TestSwipeActionStartsAndBlocksNibble(t *testing.T) {
	s := State{Grounded: true}
	p := DefaultParams()
	ground := flatGround(0)

	Step(&s, p, 1.0/60.0, input.Frame{
		Swipe:  input.ButtonState{Pressed: true},
		Nibble: input.ButtonState{Pressed: true},
	}, 0, ground)

	if s.Action != ActionSwipe {
		t.Fatalf("expected swipe to take priority when both actions pressed same frame, got %v", s.Action)
	}
}

func TestSprintIncreasesDistanceCovered(t *testing.T) {
	p := DefaultParams()
	ground := flatGround(0)

	walker := State{Grounded: true}
	sprinter := State{Grounded: true}
	walkFrame := input.Frame{Move: input.Vector2{Y: 1}}
	sprintFrame := input.Frame{Move: input.Vector2{Y: 1}, Sprint: input.ButtonState{Down: true}}

	for i := 0; i < 60; i++ {
		Step(&walker, p, 1.0/60.0, walkFrame, 0, ground)
		Step(&sprinter, p, 1.0/60.0, sprintFrame, 0, ground)
	}

	if sprinter.Position.Z <= walker.Position.Z {
		t.Fatalf("expected sprinting fox to cover more ground: walker=%v sprinter=%v",
			walker.Position.Z, sprinter.Position.Z)
	}
}
