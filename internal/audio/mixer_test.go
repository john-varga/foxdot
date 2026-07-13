package audio

import "testing"

func TestApproachMovesTowardTarget(t *testing.T) {
	got := approach(0, 1, 0.5)
	if got != 0.5 {
		t.Fatalf("got %v, want 0.5", got)
	}
}

func TestApproachClampsFullStepToTarget(t *testing.T) {
	if got := approach(0.2, 0.9, 1); got != 0.9 {
		t.Fatalf("got %v, want 0.9 (full step should land exactly on target)", got)
	}
	if got := approach(0.2, 0.9, 5); got != 0.9 {
		t.Fatalf("got %v, want 0.9 (over-shoot step should clamp to target)", got)
	}
}

func TestApproachZeroStepDoesNotMove(t *testing.T) {
	if got := approach(0.3, 1, 0); got != 0.3 {
		t.Fatalf("got %v, want 0.3 unchanged", got)
	}
}
