package input

import "math"

// clampMagnitude scales v down so its length never exceeds max, preserving
// direction. Vectors shorter than max are left untouched.
func clampMagnitude(v Vector2, max float32) Vector2 {
	m := sqrMag(v)
	if m <= max*max || m == 0 {
		return v
	}
	scale := max / float32(math.Sqrt(float64(m)))
	return Vector2{X: v.X * scale, Y: v.Y * scale}
}

// applyDeadzone snaps small axis values to zero to avoid stick drift.
func applyDeadzone(v, deadzone float32) float32 {
	if v > -deadzone && v < deadzone {
		return 0
	}
	return v
}
