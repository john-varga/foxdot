// Package world holds the forest the fox runs around in: a flat ground
// plane plus a scattered layout of low-poly props from the asset pack
// (trees, rocks, logs, bushes, ...) that double as things to jump on, plus
// a handful of decorative ambient creatures. Props/creatures are spawned as
// entities in an *ecs.World (see internal/comp for their components) rather
// than kept in a bespoke Forest object, so collision/height queries are
// plain functions over that data — testable without raylib, and reusable
// for any future entity that needs "what's under me" (only Draw touches
// the renderer).
package world

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
	"foxdot/internal/comp"
	"foxdot/internal/ecs"
)

// DefaultGroundY is the height of the forest's flat ground plane.
const DefaultGroundY float32 = 0

// top returns the height of an asset's top surface when placed at
// position with the given scale, or position.Y if the asset name isn't in
// the catalog (defensive default, so a typo'd name degrades to "flat
// ground" instead of breaking collision entirely).
func top(assetName string, position rl.Vector3, scale float32) float32 {
	meta, ok := assets.ByName(assetName)
	if !ok {
		return position.Y
	}
	return position.Y + meta.Height*scale
}

// radius returns an asset's collision radius (a circle sized to its
// widest horizontal extent) at the given scale, or 0 if unknown.
func radius(assetName string, scale float32) float32 {
	meta, ok := assets.ByName(assetName)
	if !ok {
		return 0
	}
	return meta.Footprint() * scale
}

// contains reports whether world-space point (x, z) falls within an
// asset's footprint when placed at position with the given scale.
func contains(assetName string, position rl.Vector3, scale float32, x, z float32) bool {
	dx := x - position.X
	dz := z - position.Z
	r := radius(assetName, scale)
	return dx*dx+dz*dz <= r*r
}

// HeightAt returns the height of the highest walkable surface — the
// ground plane at groundY, or the top of any Prop-tagged entity — below
// world position (x, z). This is the collision hook internal/sim's
// movement system uses to know how tall to stand.
func HeightAt(w *ecs.World, groundY float32, x, z float32) float32 {
	height := groundY
	ecs.Each3(w, func(_ ecs.Entity, t *comp.Transform, m *comp.ModelRender, _ *comp.Prop) {
		if !contains(m.AssetName, t.Position, m.EffectiveScale(), x, z) {
			return
		}
		if h := top(m.AssetName, t.Position, m.EffectiveScale()); h > height {
			height = h
		}
	})
	return height
}

// PropAt returns the Prop-tagged entity occupying (x, z), if any, mostly
// useful for deciding what a nibble/swipe action is aimed at.
func PropAt(w *ecs.World, x, z float32) (ecs.Entity, bool) {
	found := ecs.Entity(0)
	ok := false
	ecs.Each3(w, func(e ecs.Entity, t *comp.Transform, m *comp.ModelRender, _ *comp.Prop) {
		if ok || !contains(m.AssetName, t.Position, m.EffectiveScale(), x, z) {
			return
		}
		found, ok = e, true
	})
	return found, ok
}
