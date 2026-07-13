package world

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
)

// groundHalfExtent controls how large the flat ground plane/grid is drawn.
const groundHalfExtent = 20

// Draw renders the ground plane and every prop via the asset store, in
// keeping with a stylized low-poly look. Must be called between
// rl.BeginMode3D/EndMode3D. A prop whose asset fails to load draws as a
// small magenta wire cube (see assets.Store.Draw) rather than being skipped
// silently or crashing.
func (w *Forest) Draw(store *assets.Store, showGrid bool) {
	rl.DrawPlane(rl.Vector3{Y: w.GroundY}, rl.Vector2{X: groundHalfExtent * 2, Y: groundHalfExtent * 2}, rl.Color{R: 92, G: 148, B: 63, A: 255})
	if showGrid {
		rl.DrawGrid(int32(groundHalfExtent*2), 1)
	}

	for _, p := range w.Props {
		_ = store.Draw(p.AssetName, p.Position, p.YawDegrees*rl.Deg2rad, p.effectiveScale())
	}
}
