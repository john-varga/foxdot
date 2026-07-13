// Package world holds the forest the fox runs around in: a flat ground
// plane plus a scattered layout of low-poly props from the asset pack
// (trees, rocks, logs, bushes, ...) that double as things to jump on.
// Collision/height-query logic is plain data manipulation so it's testable
// without raylib; only Draw touches the renderer.
package world

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
)

// Prop places one catalog asset (see internal/assets) in the world. Props
// have no mesh data of their own — Position/YawDegrees/Scale plus a lookup
// into the asset catalog is all that's needed for both collision and
// drawing.
type Prop struct {
	AssetName  string
	Position   rl.Vector3 // ground-contact point (Y is normally the forest's GroundY)
	YawDegrees float32
	Scale      float32 // <= 0 means "use 1"
}

func (p Prop) effectiveScale() float32 {
	if p.Scale <= 0 {
		return 1
	}
	return p.Scale
}

// Top returns the height of the prop's top surface, or its base position if
// its asset name isn't in the catalog (defensive default, so a typo'd name
// degrades to "flat ground" instead of breaking collision entirely).
func (p Prop) Top() float32 {
	meta, ok := assets.ByName(p.AssetName)
	if !ok {
		return p.Position.Y
	}
	return p.Position.Y + meta.Height*p.effectiveScale()
}

// Radius returns the prop's collision radius: a circle sized to the widest
// horizontal extent of its model. Circular (rather than a rotated
// rectangle) so props can be scattered with a random yaw for visual variety
// without complicating collision.
func (p Prop) Radius() float32 {
	meta, ok := assets.ByName(p.AssetName)
	if !ok {
		return 0
	}
	return meta.Footprint() * p.effectiveScale()
}

// Contains reports whether the world-space point (x, z) falls within the
// prop's footprint.
func (p Prop) Contains(x, z float32) bool {
	dx := x - p.Position.X
	dz := z - p.Position.Z
	r := p.Radius()
	return dx*dx+dz*dz <= r*r
}

// Forest is the whole playable clearing.
type Forest struct {
	GroundY float32
	Props   []Prop
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
