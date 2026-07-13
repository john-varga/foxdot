// Package assets catalogs the low-poly forest art pack under /assets and
// loads it into raylib Models. The catalog itself (this file) is plain data
// with no raylib dependency, so it's trivially unit testable; only Store
// (see store.go) touches raylib, and only once a window/GPU context exists.
package assets

import "math"

// Category groups related models, mirroring the folder layout under
// assets/models.
type Category string

const (
	CategoryAnimals Category = "animals"
	CategoryTrees   Category = "trees"
	CategoryPlants  Category = "plants"
	CategoryProps   Category = "props"
)

// animalYawOffsetRadians corrects for how the pack's animal models were
// authored: they face local +X (nose along +X, tail along -X, confirmed by
// rendering fox.obj from above — see assets/README.md), while the engine's
// convention is that yaw=0 faces world +Z. Rotating -90 degrees around Y
// maps local +X to world +Z (see internal/camera and internal/sim for the
// matching convention). Non-directional props/plants/trees use zero.
const animalYawOffsetRadians = -math.Pi / 2

// Model is one catalog entry: where to find it on disk, and enough of its
// authored bounding box to place and collide with it without having to load
// it first (useful for world layout and for tests that never touch raylib).
type Model struct {
	Name     string
	Category Category
	RelPath  string // relative to the assets root, e.g. "models/animals/fox.obj"

	// Width (local X), Height (local Y) and Depth (local Z) are the
	// model's authored bounding box extents, in meters.
	Width  float32
	Height float32
	Depth  float32

	// MinY is the lowest point of the model's bounding box, relative to its
	// origin. Most models sit right at Y=0, but a few (rocks, the log)
	// dip slightly below or float slightly above it; Store.Draw uses this
	// to plant models exactly on the ground instead of clipping or
	// floating.
	MinY float32

	// YawOffsetRadians is added to any caller-supplied yaw before drawing,
	// correcting for the model's authored forward axis. Zero for models
	// with no meaningful facing direction.
	YawOffsetRadians float32
}

// Footprint returns a conservative circular collision radius (meters) for
// this model, using its widest horizontal extent. A circle is rotation
// invariant, which keeps world placement code simple since props can be
// scattered with a random yaw for visual variety without affecting
// collision.
func (m Model) Footprint() float32 {
	r := m.Width
	if m.Depth > r {
		r = m.Depth
	}
	return r / 2
}

// catalog is the full set of models shipped in assets/models. Keep it
// sorted by category then name to make diffs (and adding new assets) easy.
var catalog = []Model{
	// Animals - directional (see animalYawOffsetRadians above).
	{Name: "beetle", Category: CategoryAnimals, RelPath: "models/animals/beetle.obj", Width: 0.70, Height: 0.32, Depth: 0.49, MinY: 0.07, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "bird", Category: CategoryAnimals, RelPath: "models/animals/bird.obj", Width: 0.97, Height: 0.90, Depth: 0.79, MinY: -0.06, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "butterfly", Category: CategoryAnimals, RelPath: "models/animals/butterfly.obj", Width: 0.96, Height: 0.76, Depth: 0.11, MinY: 0.09, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "deer", Category: CategoryAnimals, RelPath: "models/animals/deer.obj", Width: 1.58, Height: 2.27, Depth: 0.62, MinY: -0.04, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "fox", Category: CategoryAnimals, RelPath: "models/animals/fox.obj", Width: 2.88, Height: 1.71, Depth: 0.69, MinY: 0.00, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "rabbit", Category: CategoryAnimals, RelPath: "models/animals/rabbit.obj", Width: 1.29, Height: 1.50, Depth: 0.66, MinY: 0.10, YawOffsetRadians: animalYawOffsetRadians},
	{Name: "squirrel", Category: CategoryAnimals, RelPath: "models/animals/squirrel.obj", Width: 1.25, Height: 1.32, Depth: 0.50, MinY: 0.04, YawOffsetRadians: animalYawOffsetRadians},

	// Plants - non-directional.
	{Name: "bush_berries", Category: CategoryPlants, RelPath: "models/plants/bush_berries.obj", Width: 1.40, Height: 1.05, Depth: 1.01, MinY: -0.02},
	{Name: "bush_plain", Category: CategoryPlants, RelPath: "models/plants/bush_plain.obj", Width: 1.40, Height: 1.05, Depth: 0.90, MinY: -0.02},
	{Name: "bush_white_flowers", Category: CategoryPlants, RelPath: "models/plants/bush_white_flowers.obj", Width: 1.40, Height: 1.05, Depth: 1.01, MinY: -0.02},
	{Name: "bush_yellow_flowers", Category: CategoryPlants, RelPath: "models/plants/bush_yellow_flowers.obj", Width: 1.40, Height: 1.05, Depth: 1.01, MinY: -0.02},
	{Name: "grass_01", Category: CategoryPlants, RelPath: "models/plants/grass_01.obj", Width: 0.48, Height: 0.61, Depth: 0.42, MinY: -0.01},
	{Name: "grass_02", Category: CategoryPlants, RelPath: "models/plants/grass_02.obj", Width: 0.44, Height: 0.61, Depth: 0.46, MinY: -0.01},
	{Name: "grass_03", Category: CategoryPlants, RelPath: "models/plants/grass_03.obj", Width: 0.45, Height: 0.61, Depth: 0.45, MinY: -0.01},
	{Name: "mushroom_red", Category: CategoryPlants, RelPath: "models/plants/mushroom_red.obj", Width: 0.56, Height: 0.61, Depth: 0.56, MinY: 0.00},
	{Name: "mushroom_yellow", Category: CategoryPlants, RelPath: "models/plants/mushroom_yellow.obj", Width: 0.56, Height: 0.61, Depth: 0.56, MinY: 0.00},

	// Props - non-directional.
	{Name: "campfire", Category: CategoryProps, RelPath: "models/props/campfire.obj", Width: 1.10, Height: 0.96, Depth: 0.91, MinY: -0.08},
	{Name: "crate", Category: CategoryProps, RelPath: "models/props/crate.obj", Width: 1.08, Height: 1.00, Depth: 1.10, MinY: 0.00},
	{Name: "fence", Category: CategoryProps, RelPath: "models/props/fence.obj", Width: 1.78, Height: 1.53, Depth: 0.28, MinY: 0.00},
	{Name: "log", Category: CategoryProps, RelPath: "models/props/log.obj", Width: 1.70, Height: 0.71, Depth: 0.64, MinY: 0.02},
	{Name: "rock_01", Category: CategoryProps, RelPath: "models/props/rock_01.obj", Width: 1.07, Height: 0.84, Depth: 0.93, MinY: 0.00},
	{Name: "rock_02", Category: CategoryProps, RelPath: "models/props/rock_02.obj", Width: 0.87, Height: 1.24, Depth: 0.83, MinY: -0.20},
	{Name: "rock_03", Category: CategoryProps, RelPath: "models/props/rock_03.obj", Width: 1.03, Height: 0.71, Depth: 0.97, MinY: 0.06},

	// Trees - non-directional.
	{Name: "birch", Category: CategoryTrees, RelPath: "models/trees/birch.obj", Width: 1.50, Height: 3.08, Depth: 1.18, MinY: 0.00},
	{Name: "fir", Category: CategoryTrees, RelPath: "models/trees/fir.obj", Width: 2.00, Height: 2.65, Depth: 2.00, MinY: 0.00},
	{Name: "oak", Category: CategoryTrees, RelPath: "models/trees/oak.obj", Width: 1.85, Height: 2.98, Depth: 1.05, MinY: 0.00},
	{Name: "pine", Category: CategoryTrees, RelPath: "models/trees/pine.obj", Width: 2.04, Height: 2.88, Depth: 2.07, MinY: 0.00},
}

var byName = func() map[string]Model {
	m := make(map[string]Model, len(catalog))
	for _, entry := range catalog {
		m[entry.Name] = entry
	}
	return m
}()

// ByName looks up a catalog entry by its name (the file's base name, e.g.
// "fox" or "rock_02").
func ByName(name string) (Model, bool) {
	m, ok := byName[name]
	return m, ok
}

// All returns every catalog entry.
func All() []Model {
	out := make([]Model, len(catalog))
	copy(out, catalog)
	return out
}

// ByCategory returns every catalog entry in the given category.
func ByCategory(cat Category) []Model {
	var out []Model
	for _, entry := range catalog {
		if entry.Category == cat {
			out = append(out, entry)
		}
	}
	return out
}
