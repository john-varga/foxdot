package world

import rl "github.com/gen2brain/raylib-go/raylib"

// groundHalfExtent controls how large the flat ground plane/grid is drawn.
const groundHalfExtent = 20

// Draw renders the ground plane and every prop with simple flat-shaded
// primitives, in keeping with a stylized low-poly look. Must be called
// between rl.BeginMode3D/EndMode3D.
func (w *Forest) Draw(showGrid bool) {
	rl.DrawPlane(rl.Vector3{Y: w.GroundY}, rl.Vector2{X: groundHalfExtent * 2, Y: groundHalfExtent * 2}, rl.Color{R: 92, G: 148, B: 63, A: 255})
	if showGrid {
		rl.DrawGrid(int32(groundHalfExtent*2), 1)
	}

	for _, p := range w.Props {
		center := rl.Vector3{X: p.Position.X, Y: p.Position.Y + p.Size.Y/2, Z: p.Position.Z}
		rl.DrawCubeV(center, p.Size, p.Color)
		rl.DrawCubeWiresV(center, p.Size, rl.Color{R: 30, G: 30, B: 30, A: 255})
	}
}
