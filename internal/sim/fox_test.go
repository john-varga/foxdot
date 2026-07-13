package sim

import (
	"math"
	"testing"

	"foxdot/internal/input"
)

func flatGround(y float32) GroundHeightFunc {
	return func(x, z float32) float32 { return y }
}

func TestFoxFallsAndLandsOnGround(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Position.Y = 5
	fox.State.Grounded = false

	ground := flatGround(0)
	for i := 0; i < 300; i++ { // plenty of time to fall and settle
		fox.Step(1.0/60.0, input.Frame{}, 0, ground)
	}

	if !fox.State.Grounded {
		t.Fatalf("expected fox to be grounded after falling, got airborne")
	}
	if fox.State.Position.Y != 0 {
		t.Fatalf("expected fox to rest at ground height 0, got %v", fox.State.Position.Y)
	}
}

func TestFoxJumpsWhenGroundedAndRises(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Grounded = true
	ground := flatGround(0)

	fox.Step(1.0/60.0, input.Frame{Jump: input.ButtonState{Pressed: true}}, 0, ground)

	if fox.State.Grounded {
		t.Fatalf("expected fox to leave the ground after jumping")
	}
	if fox.State.Velocity.Y <= 0 {
		t.Fatalf("expected positive upward velocity after jump, got %v", fox.State.Velocity.Y)
	}
}

func TestFoxCannotJumpWhileAirborne(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Grounded = false
	fox.State.Velocity.Y = 1
	ground := flatGround(0)

	fox.Step(1.0/60.0, input.Frame{Jump: input.ButtonState{Pressed: true}}, 0, ground)

	// Velocity should just be gravity-integrated, not re-boosted by JumpVelocity.
	expected := float32(1) - DefaultParams().Gravity*(1.0/60.0)
	if math.Abs(float64(fox.State.Velocity.Y-expected)) > 1e-4 {
		t.Fatalf("expected velocity %.5f from gravity alone, got %.5f", expected, fox.State.Velocity.Y)
	}
}

func TestFoxMovesForwardRelativeToCamera(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Grounded = true
	ground := flatGround(0)

	// Camera facing +90 degrees (yaw = pi/2): "forward" should now push the
	// fox along world +X instead of +Z.
	cameraYaw := float32(math.Pi / 2)
	frame := input.Frame{Move: input.Vector2{X: 0, Y: 1}}

	for i := 0; i < 30; i++ {
		fox.Step(1.0/60.0, frame, cameraYaw, ground)
	}

	if fox.State.Position.X <= 0.1 {
		t.Fatalf("expected fox to move along +X when camera yaw is 90deg, got position %+v", fox.State.Position)
	}
	if math.Abs(float64(fox.State.Position.Z)) > 0.1 {
		t.Fatalf("expected negligible Z movement, got position %+v", fox.State.Position)
	}
}

func TestFoxCanStandOnRaisedProp(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Position.Y = 5
	// A "rock" that's 2 units tall directly under the fox.
	raised := func(x, z float32) float32 { return 2 }

	for i := 0; i < 300; i++ {
		fox.Step(1.0/60.0, input.Frame{}, 0, raised)
	}

	if fox.State.Position.Y != 2 {
		t.Fatalf("expected fox to rest on top of the raised surface at y=2, got %v", fox.State.Position.Y)
	}
	if !fox.State.Grounded {
		t.Fatalf("expected fox to be grounded while standing on the prop")
	}
}

func TestNibbleActionRootsFoxAndExpires(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Grounded = true
	ground := flatGround(0)

	fox.Step(1.0/60.0, input.Frame{Nibble: input.ButtonState{Pressed: true}}, 0, ground)
	if fox.State.Action != ActionNibble {
		t.Fatalf("expected nibble action to start, got %v", fox.State.Action)
	}

	startPos := fox.State.Position
	moveWhileNibbling := input.Frame{Move: input.Vector2{Y: 1}}
	for i := 0; i < 10; i++ {
		fox.Step(1.0/60.0, moveWhileNibbling, 0, ground)
	}
	if fox.State.Position != startPos {
		t.Fatalf("expected fox to stay rooted while nibbling, moved from %+v to %+v", startPos, fox.State.Position)
	}

	// Run past the nibble duration and confirm the action clears.
	dur := DefaultParams().NibbleDuration
	steps := int(dur/(1.0/60.0)) + 5
	for i := 0; i < steps; i++ {
		fox.Step(1.0/60.0, input.Frame{}, 0, ground)
	}
	if fox.State.Action != ActionNone {
		t.Fatalf("expected nibble action to expire, still %v", fox.State.Action)
	}
}

func TestSwipeActionStartsAndBlocksNibble(t *testing.T) {
	fox := NewFox(DefaultParams())
	fox.State.Grounded = true
	ground := flatGround(0)

	fox.Step(1.0/60.0, input.Frame{
		Swipe:  input.ButtonState{Pressed: true},
		Nibble: input.ButtonState{Pressed: true},
	}, 0, ground)

	if fox.State.Action != ActionSwipe {
		t.Fatalf("expected swipe to take priority when both actions pressed same frame, got %v", fox.State.Action)
	}
}

func TestSprintIncreasesDistanceCovered(t *testing.T) {
	makeFox := func() *Fox {
		fox := NewFox(DefaultParams())
		fox.State.Grounded = true
		return fox
	}
	ground := flatGround(0)

	walker := makeFox()
	sprinter := makeFox()
	walkFrame := input.Frame{Move: input.Vector2{Y: 1}}
	sprintFrame := input.Frame{Move: input.Vector2{Y: 1}, Sprint: input.ButtonState{Down: true}}

	for i := 0; i < 60; i++ {
		walker.Step(1.0/60.0, walkFrame, 0, ground)
		sprinter.Step(1.0/60.0, sprintFrame, 0, ground)
	}

	if sprinter.State.Position.Z <= walker.State.Position.Z {
		t.Fatalf("expected sprinting fox to cover more ground: walker=%v sprinter=%v",
			walker.State.Position.Z, sprinter.State.Position.Z)
	}
}
