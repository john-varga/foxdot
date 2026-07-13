// Package world holds the (currently placeholder) forest the fox runs
// around in: a flat ground plane plus a handful of simple low-poly props
// (stand-ins for rocks, stumps and logs) that double as things to jump on.
// Collision/height-query logic is plain data manipulation so it's testable
// without raylib; only Draw touches the renderer.
package world

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Prop is a simple box-shaped obstacle/platform. Real meshes will replace
// these once art assets land; for now they double as jumpable "furniture"
// for the forest.
type Prop struct {
	Name     string
	Position rl.Vector3 // center of the base (Y = bottom of the box)
	Size     rl.Vector3 // width (X), height (Y), depth (Z)
	Color    rl.Color
}

// Top returns the height of the prop's top surface.
func (p Prop) Top() float32 {
	return p.Position.Y + p.Size.Y
}

// Contains reports whether the world-space point (x, z) falls within the
// prop's footprint.
func (p Prop) Contains(x, z float32) bool {
	halfX := p.Size.X / 2
	halfZ := p.Size.Z / 2
	return x >= p.Position.X-halfX && x <= p.Position.X+halfX &&
		z >= p.Position.Z-halfZ && z <= p.Position.Z+halfZ
}

// Forest is the whole (currently tiny and placeholder) playable area.
type Forest struct {
	GroundY float32
	Props   []Prop
}

// NewPlaceholderForest builds a small clearing with a few props scattered
// around so there's immediately something to run around and jump on, ahead
// of real assets landing.
func NewPlaceholderForest() *Forest {
	return &Forest{
		GroundY: 0,
		Props: []Prop{
			{Name: "stump-1", Position: rl.Vector3{X: 3, Y: 0, Z: 2}, Size: rl.Vector3{X: 1.2, Y: 0.6, Z: 1.2}, Color: rl.Color{R: 121, G: 85, B: 61, A: 255}},
			{Name: "rock-1", Position: rl.Vector3{X: -4, Y: 0, Z: 3}, Size: rl.Vector3{X: 1.6, Y: 1.0, Z: 1.4}, Color: rl.Color{R: 130, G: 130, B: 130, A: 255}},
			{Name: "log-1", Position: rl.Vector3{X: -2, Y: 0, Z: -4}, Size: rl.Vector3{X: 2.4, Y: 0.5, Z: 0.6}, Color: rl.Color{R: 101, G: 67, B: 33, A: 255}},
			{Name: "rock-2", Position: rl.Vector3{X: 5, Y: 0, Z: -3}, Size: rl.Vector3{X: 1.0, Y: 1.8, Z: 1.0}, Color: rl.Color{R: 110, G: 110, B: 110, A: 255}},
			{Name: "stump-2", Position: rl.Vector3{X: 0, Y: 0, Z: 6}, Size: rl.Vector3{X: 1.0, Y: 1.2, Z: 1.0}, Color: rl.Color{R: 121, G: 85, B: 61, A: 255}},
		},
	}
}

// HeightAt returns the height of the highest walkable surface (ground or a
// prop top) below world position (x, z). This is the collision hook the
// fox simulation uses to know how tall to stand.
func (w *Forest) HeightAt(x, z float32) float32 {
	height := w.GroundY
	for _, p := range w.Props {
		if p.Contains(x, z) {
			if top := p.Top(); top > height {
				height = top
			}
		}
	}
	return height
}

// PropAt returns the prop occupying (x, z), if any, mostly useful for
// deciding what a nibble/swipe action is aimed at.
func (w *Forest) PropAt(x, z float32) (Prop, bool) {
	for _, p := range w.Props {
		if p.Contains(x, z) {
			return p, true
		}
	}
	return Prop{}, false
}

func (w *Forest) String() string {
	return fmt.Sprintf("Forest{groundY=%.2f, props=%d}", w.GroundY, len(w.Props))
}
