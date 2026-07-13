// Package camera implements a simple orbiting third-person camera suitable
// for a low-poly, over-the-shoulder style game. It only depends on raylib's
// plain math types/helpers (no window/GPU state), so its orbit math can be
// unit tested like any other Go code.
package camera

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/input"
)

// Settings holds every tunable knob for the third-person camera. It is
// meant to be embedded into the game's overall config so it can be tweaked
// without a recompile. Pitch uses degrees: positive looks down at the
// target, so MinPitchDegrees >= 0 keeps the camera from orbiting under the
// follow target (and therefore under a flat ground plane).
type Settings struct {
	Distance         float32 `json:"distance"`
	MinDistance      float32 `json:"minDistance"`
	MaxDistance      float32 `json:"maxDistance"`
	HeightOffset     float32 `json:"heightOffset"`   // height of the orbit pivot above the target
	ShoulderOffset   float32 `json:"shoulderOffset"` // sideways offset for an over-the-shoulder feel
	MinPitchDegrees  float32 `json:"minPitchDegrees"`
	MaxPitchDegrees  float32 `json:"maxPitchDegrees"`
	FollowSmoothTime float32 `json:"followSmoothTime"` // seconds; higher = laggier/smoother follow
	FOVDegrees       float32 `json:"fovDegrees"`
	ZoomSpeed        float32 `json:"zoomSpeed"`
	// MinHeightAboveGround clamps the camera's world Y so it cannot dip
	// below the follow target's Y by this much (extra guard beyond pitch
	// limits, useful on uneven ground later).
	MinHeightAboveGround float32 `json:"minHeightAboveGround"`
}

// DefaultSettings returns reasonable defaults for a mid-distance third-person
// camera.
func DefaultSettings() Settings {
	return Settings{
		Distance:             6,
		MinDistance:          2.5,
		MaxDistance:          12,
		HeightOffset:         1.6,
		ShoulderOffset:       0,
		MinPitchDegrees:      8, // slight look-down floor; never orbit underground
		MaxPitchDegrees:      70,
		FollowSmoothTime:     0.12,
		FOVDegrees:           55,
		ZoomSpeed:            4,
		MinHeightAboveGround: 0.4,
	}
}

// ThirdPerson orbits a point of interest (typically just above the player
// character) at a configurable distance, yaw and pitch.
type ThirdPerson struct {
	Settings Settings

	Yaw   float32 // radians, 0 = looking down -Z
	Pitch float32 // radians, positive = looking down at the target

	distance float32 // current zoom distance, clamped to [MinDistance, MaxDistance]

	smoothedTarget rl.Vector3
	haveTarget     bool

	position rl.Vector3
	lookAt   rl.Vector3
}

// NewThirdPerson creates a camera with the given settings, a neutral yaw,
// and a slight default downward pitch.
func NewThirdPerson(settings Settings) *ThirdPerson {
	return &ThirdPerson{
		Settings: settings,
		Yaw:      0,
		Pitch:    15 * rl.Deg2rad,
		distance: settings.Distance,
	}
}

// Update advances the camera orbit by dt using per-frame look input (yaw and
// pitch deltas, already scaled by sensitivity upstream) and a zoom delta
// (e.g. mouse wheel), then re-centers on target.
func (c *ThirdPerson) Update(dt float32, target rl.Vector3, look input.Vector2, zoomDelta float32) {
	s := c.Settings

	c.Yaw += look.X
	c.Pitch += look.Y
	minPitch := s.MinPitchDegrees * rl.Deg2rad
	maxPitch := s.MaxPitchDegrees * rl.Deg2rad
	c.Pitch = rl.Clamp(c.Pitch, minPitch, maxPitch)

	c.distance -= zoomDelta * s.ZoomSpeed * dt
	c.distance = rl.Clamp(c.distance, s.MinDistance, s.MaxDistance)

	if !c.haveTarget {
		c.smoothedTarget = target
		c.haveTarget = true
	} else {
		c.smoothedTarget = rl.Vector3Lerp(c.smoothedTarget, target, followFactor(dt, s.FollowSmoothTime))
	}

	pivot := rl.Vector3Add(c.smoothedTarget, rl.Vector3{Y: s.HeightOffset})

	cosPitch := float32(math.Cos(float64(c.Pitch)))
	sinPitch := float32(math.Sin(float64(c.Pitch)))
	cosYaw := float32(math.Cos(float64(c.Yaw)))
	sinYaw := float32(math.Sin(float64(c.Yaw)))

	offset := rl.Vector3{
		X: c.distance * cosPitch * sinYaw,
		Y: c.distance * sinPitch,
		Z: c.distance * cosPitch * cosYaw,
	}

	// Shoulder offset shifts the camera sideways in view space without
	// changing what it's looking at, for an over-the-shoulder feel.
	right := rl.Vector3{X: cosYaw, Y: 0, Z: -sinYaw}
	shoulder := rl.Vector3Scale(right, s.ShoulderOffset)

	c.position = rl.Vector3Add(rl.Vector3Add(pivot, offset), shoulder)
	c.lookAt = rl.Vector3Add(pivot, shoulder)

	// Belt-and-suspenders against dipping under the world: even if pitch
	// limits were retuned aggressively in config, keep the camera above
	// the follow target by MinHeightAboveGround.
	minY := c.smoothedTarget.Y + s.MinHeightAboveGround
	if c.position.Y < minY {
		c.position.Y = minY
	}
}

// followFactor converts a smoothing time constant into a per-frame lerp
// factor so the camera's follow speed is frame-rate independent.
func followFactor(dt, smoothTime float32) float32 {
	if smoothTime <= 0 {
		return 1
	}
	factor := 1 - float32(math.Exp(-float64(dt/smoothTime)))
	return rl.Clamp(factor, 0, 1)
}

// RLCamera returns the raylib Camera3D for the current orbit state, ready
// to pass to rl.BeginMode3D.
func (c *ThirdPerson) RLCamera() rl.Camera3D {
	return rl.Camera3D{
		Position:   c.position,
		Target:     c.lookAt,
		Up:         rl.Vector3{Y: 1},
		Fovy:       c.Settings.FOVDegrees,
		Projection: rl.CameraPerspective,
	}
}

// Distance returns the camera's current (post-zoom) orbit distance.
func (c *ThirdPerson) Distance() float32 {
	return c.distance
}
