package camera

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/input"
)

func TestUpdateMaintainsDistanceFromTarget(t *testing.T) {
	cam := NewThirdPerson(DefaultSettings())
	target := rl.Vector3{X: 5, Y: 0, Z: -3}

	// Run several ticks so the smoothed follow target converges on the
	// actual target.
	for i := 0; i < 240; i++ {
		cam.Update(1.0/60.0, target, input.Vector2{}, 0)
	}

	rlCam := cam.RLCamera()
	dist := rl.Vector3Distance(rlCam.Position, rlCam.Target)
	if math.Abs(float64(dist-cam.Distance())) > 0.01 {
		t.Fatalf("expected camera to sit ~%.2f units from its look-at point, got %.4f", cam.Distance(), dist)
	}
}

func TestPitchIsClamped(t *testing.T) {
	settings := DefaultSettings()
	cam := NewThirdPerson(settings)

	// Push pitch far past the configured max.
	for i := 0; i < 100; i++ {
		cam.Update(1.0/60.0, rl.Vector3{}, input.Vector2{Y: 1}, 0)
	}
	maxPitch := settings.MaxPitchDegrees * rl.Deg2rad
	if cam.Pitch > maxPitch+1e-4 {
		t.Fatalf("expected pitch to clamp at %.4f rad, got %.4f", maxPitch, cam.Pitch)
	}

	// And far past the configured min.
	for i := 0; i < 100; i++ {
		cam.Update(1.0/60.0, rl.Vector3{}, input.Vector2{Y: -1}, 0)
	}
	minPitch := settings.MinPitchDegrees * rl.Deg2rad
	if cam.Pitch < minPitch-1e-4 {
		t.Fatalf("expected pitch to clamp at %.4f rad, got %.4f", minPitch, cam.Pitch)
	}
}

func TestZoomIsClampedToDistanceRange(t *testing.T) {
	settings := DefaultSettings()
	cam := NewThirdPerson(settings)

	for i := 0; i < 1000; i++ {
		cam.Update(1.0/60.0, rl.Vector3{}, input.Vector2{}, 1) // zoom out
	}
	if cam.Distance() > settings.MaxDistance {
		t.Fatalf("expected distance clamped to max %.2f, got %.2f", settings.MaxDistance, cam.Distance())
	}

	for i := 0; i < 1000; i++ {
		cam.Update(1.0/60.0, rl.Vector3{}, input.Vector2{}, -1) // zoom in
	}
	if cam.Distance() < settings.MinDistance {
		t.Fatalf("expected distance clamped to min %.2f, got %.2f", settings.MinDistance, cam.Distance())
	}
}
