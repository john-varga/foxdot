package world

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/comp"
	"foxdot/internal/ecs"
)

const (
	clearingRadius   = 18  // meters; scattered filler props land within this radius
	spawnClearRadius = 3.5 // meters; kept free so the fox doesn't spawn inside a bush
	scatterCount     = 40
)

// placement is plain layout data for a prop before it becomes an entity —
// kept separate from comp.Transform/comp.ModelRender so the landmark table
// and the overlap-rejection sampler below don't need a live *ecs.World.
type placement struct {
	AssetName  string
	Position   rl.Vector3
	YawDegrees float32
}

// landmarks are hand-placed for a readable layout: a few trees, rocks, a
// fallen log and a small campsite. Everything else is scattered
// procedurally around them (see SpawnGeneratedForest).
var landmarks = []placement{
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

// ambientCreatures are decorative, non-colliding wildlife: each gets a
// fixed (seed-independent) spot in the clearing and a CreatureVoice tied to
// the matching catalog sound effects, so the world periodically sounds
// alive without any real AI yet. The audio pack has no dedicated squirrel
// sound, so it borrows rabbit_hop as the closest match.
var ambientCreatures = []struct {
	AssetName string
	Position  rl.Vector3
	Scale     float32
	Sounds    []string
}{
	{AssetName: "rabbit", Position: rl.Vector3{X: 4, Z: -4}, Scale: 0.5, Sounds: []string{"rabbit_hop"}},
	{AssetName: "squirrel", Position: rl.Vector3{X: -5, Z: -2}, Scale: 0.45, Sounds: []string{"rabbit_hop"}},
	{AssetName: "deer", Position: rl.Vector3{X: -11, Z: -4}, Scale: 0.6, Sounds: []string{"deer_call"}},
	{AssetName: "bird", Position: rl.Vector3{X: 2, Z: 8}, Scale: 0.4, Sounds: []string{"bird_chirp_1", "bird_chirp_2", "bird_chirp_3"}},
	{AssetName: "butterfly", Position: rl.Vector3{X: -2, Z: 2}, Scale: 0.35, Sounds: []string{"butterfly_flutter"}},
	{AssetName: "beetle", Position: rl.Vector3{X: 6, Z: 3}, Scale: 0.3, Sounds: []string{"beetle_buzz"}},
}

// SpawnDefaultForest spawns the fox's home clearing using a fixed seed, so
// every run/test sees the same layout.
func SpawnDefaultForest(w *ecs.World) {
	SpawnGeneratedForest(w, 1)
}

// SpawnGeneratedForest spawns a forest clearing from a seed as entities in
// w: the hand-placed landmarks above, plus filler plants scattered around
// them with simple rejection sampling so nothing overlaps, plus a handful
// of ambient (non-colliding) creatures. Reusing a seed always reproduces
// the same layout — this is deliberately simple procedural generation, a
// stand-in for whatever later generates real terrain/vegetation, and a
// natural fit for the generated-content cache in internal/storage
// (storage.Store.SaveGenerated/LoadGenerated) once layouts get expensive
// enough to be worth caching.
func SpawnGeneratedForest(w *ecs.World, seed int64) {
	placed := make([]placement, 0, len(landmarks)+scatterCount)
	for _, pl := range landmarks {
		spawnProp(w, pl)
		placed = append(placed, pl)
	}

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < scatterCount; i++ {
		name := scatterAssets[rng.Intn(len(scatterAssets))]
		for attempt := 0; attempt < 20; attempt++ {
			angle := rng.Float64() * 2 * math.Pi
			r := spawnClearRadius + rng.Float64()*(clearingRadius-spawnClearRadius)
			candidate := placement{
				AssetName: name,
				Position: rl.Vector3{
					X: float32(math.Cos(angle) * r),
					Z: float32(math.Sin(angle) * r),
				},
				YawDegrees: float32(rng.Float64() * 360),
			}
			if !overlapsAny(placed, candidate) {
				spawnProp(w, candidate)
				placed = append(placed, candidate)
				break
			}
		}
	}

	SpawnAmbientCreatures(w)
}

// SpawnAmbientCreatures spawns the fixed set of decorative wildlife. Split
// out from SpawnGeneratedForest so tests (and a future "wildlife density"
// setting) can spawn them independently of the prop layout.
func SpawnAmbientCreatures(w *ecs.World) {
	for _, c := range ambientCreatures {
		e := w.NewEntity()
		ecs.Set(w, e, comp.Transform{Position: c.Position})
		ecs.Set(w, e, comp.ModelRender{AssetName: c.AssetName, Scale: c.Scale})
		ecs.Set(w, e, comp.CreatureVoice{
			Sounds:      c.Sounds,
			MinInterval: 6,
			MaxInterval: 16,
			Volume:      0.6,
			Timer:       2 + rand.Float32()*10,
		})
	}
}

func spawnProp(w *ecs.World, pl placement) ecs.Entity {
	e := w.NewEntity()
	ecs.Set(w, e, comp.Transform{Position: pl.Position, Yaw: pl.YawDegrees * rl.Deg2rad})
	ecs.Set(w, e, comp.ModelRender{AssetName: pl.AssetName})
	ecs.Set(w, e, comp.Prop{})
	if pl.AssetName == "campfire" {
		ecs.Set(w, e, comp.CampfireSource{})
	}
	return e
}

func overlapsAny(placed []placement, candidate placement) bool {
	for _, p := range placed {
		dx := p.Position.X - candidate.Position.X
		dz := p.Position.Z - candidate.Position.Z
		minDist := radius(p.AssetName, 1) + radius(candidate.AssetName, 1)
		if dx*dx+dz*dz < minDist*minDist {
			return true
		}
	}
	return false
}
