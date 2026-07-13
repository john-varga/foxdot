package world

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	clearingRadius   = 18  // meters; scattered filler props land within this radius
	spawnClearRadius = 3.5 // meters; kept free so the fox doesn't spawn inside a bush
	scatterCount     = 40
)

// landmarks are hand-placed for a readable layout: a few trees, rocks, a
// fallen log and a small campsite. Everything else is scattered
// procedurally around them (see NewGeneratedForest).
var landmarks = []Prop{
	{AssetName: "pine", Position: rl.Vector3{X: 8, Z: 6}},
	{AssetName: "fir", Position: rl.Vector3{X: -9, Z: 5}},
	{AssetName: "oak", Position: rl.Vector3{X: -6, Z: -9}, YawDegrees: 40},
	{AssetName: "birch", Position: rl.Vector3{X: 10, Z: -7}, YawDegrees: 200},

	{AssetName: "rock_01", Position: rl.Vector3{X: 3, Z: 2}},
	{AssetName: "rock_02", Position: rl.Vector3{X: -4, Z: 3}, YawDegrees: 60},
	{AssetName: "rock_03", Position: rl.Vector3{X: 5, Z: -3}, YawDegrees: 300},
	{AssetName: "log", Position: rl.Vector3{X: -2, Z: -5}, YawDegrees: 15},

	{AssetName: "crate", Position: rl.Vector3{X: -3, Z: 6}, YawDegrees: 20},
	{AssetName: "campfire", Position: rl.Vector3{X: 0, Z: 5}},
	{AssetName: "fence", Position: rl.Vector3{X: 6, Z: 9}, YawDegrees: 90},
	{AssetName: "fence", Position: rl.Vector3{X: 8, Z: 9}, YawDegrees: 90},
}

// scatterAssets are strewn around the landmarks at random (but
// reproducible) positions to fill out the clearing.
var scatterAssets = []string{
	"bush_plain", "bush_berries", "bush_white_flowers", "bush_yellow_flowers",
	"grass_01", "grass_02", "grass_03",
	"mushroom_red", "mushroom_yellow",
}

// NewDefaultForest builds the fox's home clearing using a fixed seed, so
// every run/test sees the same layout.
func NewDefaultForest() *Forest {
	return NewGeneratedForest(1)
}

// NewGeneratedForest builds a forest clearing from a seed: the hand-placed
// landmarks above, plus filler plants scattered around them with simple
// rejection sampling so nothing overlaps. Reusing a seed always reproduces
// the same layout — this is deliberately simple procedural generation, a
// stand-in for whatever later generates real terrain/vegetation, and a
// natural fit for the generated-content cache in internal/storage
// (storage.Store.SaveGenerated/LoadGenerated) once layouts get expensive
// enough to be worth caching.
func NewGeneratedForest(seed int64) *Forest {
	f := &Forest{GroundY: 0}
	f.Props = append(f.Props, landmarks...)

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < scatterCount; i++ {
		name := scatterAssets[rng.Intn(len(scatterAssets))]
		for attempt := 0; attempt < 20; attempt++ {
			angle := rng.Float64() * 2 * math.Pi
			radius := spawnClearRadius + rng.Float64()*(clearingRadius-spawnClearRadius)
			candidate := Prop{
				AssetName: name,
				Position: rl.Vector3{
					X: float32(math.Cos(angle) * radius),
					Z: float32(math.Sin(angle) * radius),
				},
				YawDegrees: float32(rng.Float64() * 360),
			}
			if !overlapsAny(f.Props, candidate) {
				f.Props = append(f.Props, candidate)
				break
			}
		}
	}
	return f
}

func overlapsAny(props []Prop, candidate Prop) bool {
	for _, p := range props {
		dx := p.Position.X - candidate.Position.X
		dz := p.Position.Z - candidate.Position.Z
		minDist := p.Radius() + candidate.Radius()
		if dx*dx+dz*dz < minDist*minDist {
			return true
		}
	}
	return false
}
