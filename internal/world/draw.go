package world

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// groundHalfExtent controls how large the flat ground plane/grid is drawn.
const groundHalfExtent = 20

// DrawGround renders the flat ground plane (and an optional debug grid) at
// groundY. Must be called between rl.BeginMode3D/EndMode3D. Prop and
// creature entities are drawn by the game package's render system, which
// treats every (Transform, ModelRender) entity uniformly regardless of
// which package spawned it.
func DrawGround(groundY float32, showGrid bool) {
	rl.DrawPlane(rl.Vector3{Y: groundY}, rl.Vector2{X: groundHalfExtent * 2, Y: groundHalfExtent * 2}, rl.Color{R: 92, G: 148, B: 63, A: 255})
	if showGrid {
		rl.DrawGrid(int32(groundHalfExtent*2), 1)
	}
}
