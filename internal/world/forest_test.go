package world

import (
	"sort"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
	"foxdot/internal/comp"
	"foxdot/internal/ecs"
)

func spawnSingleProp(t *testing.T, assetName string, position rl.Vector3, scale float32) *ecs.World {
	t.Helper()
	w := ecs.NewWorld()
	e := w.NewEntity()
	ecs.Set(w, e, comp.Transform{Position: position})
	ecs.Set(w, e, comp.ModelRender{AssetName: assetName, Scale: scale})
	ecs.Set(w, e, comp.Prop{})
	return w
}

func TestHeightAtReturnsGroundAwayFromProps(t *testing.T) {
	w := spawnSingleProp(t, "rock_01", rl.Vector3{X: 3, Z: 2}, 0)

	if got := HeightAt(w, 0, 100, 100); got != 0 {
		t.Fatalf("expected ground height 0 far from props, got %v", got)
	}
}

func TestHeightAtReturnsPropTopWhenInsideFootprint(t *testing.T) {
	crate, ok := assets.ByName("crate")
	if !ok {
		t.Fatalf("expected crate to be in the asset catalog")
	}
	w := spawnSingleProp(t, "crate", rl.Vector3{X: 3, Z: 2}, 0)

	if got := HeightAt(w, 0, 3, 2); got != crate.Height {
		t.Fatalf("expected crate top height %v at its center, got %v", crate.Height, got)
	}

	// Just outside the crate's circular footprint.
	outside := 3 + crate.Footprint() + 0.5
	if got := HeightAt(w, 0, outside, 2); got != 0 {
		t.Fatalf("expected ground height 0 just outside footprint, got %v", got)
	}
}

func TestHeightAtPicksTallestOverlappingProp(t *testing.T) {
	fence, _ := assets.ByName("fence") // shorter
	campfire, _ := assets.ByName("campfire")

	w := ecs.NewWorld()
	for _, name := range []string{"fence", "campfire"} {
		e := w.NewEntity()
		ecs.Set(w, e, comp.Transform{})
		ecs.Set(w, e, comp.ModelRender{AssetName: name})
		ecs.Set(w, e, comp.Prop{})
	}

	want := fence.Height
	if campfire.Height > want {
		want = campfire.Height
	}
	if got := HeightAt(w, 0, 0, 0); got != want {
		t.Fatalf("expected tallest overlapping prop top %v, got %v", want, got)
	}
}

func TestHeightAtUnknownAssetDefaultsToGroundHeight(t *testing.T) {
	w := spawnSingleProp(t, "does-not-exist", rl.Vector3{X: 1, Y: 2, Z: 3}, 0)

	// Zero radius means only the exact center point counts as "inside",
	// where top() falls back to the prop's own base Y (2) rather than a
	// synthesized height, matching Top()'s old defensive default.
	if got := HeightAt(w, 0, 1, 3); got != 2 {
		t.Fatalf("expected unknown asset to default to its base Y (2) at its exact center, got %v", got)
	}
	// Anywhere else, its zero radius means it never affects HeightAt.
	if got := HeightAt(w, 0, 1.5, 3); got != 0 {
		t.Fatalf("expected unknown asset to have no footprint away from its exact center, got %v", got)
	}
}

func TestScaleAffectsTopAndRadius(t *testing.T) {
	crate, _ := assets.ByName("crate")

	if got := top("crate", rl.Vector3{}, 1); got != crate.Height {
		t.Fatalf("expected scale-1 top to equal catalog height, got %v", got)
	}
	if got := top("crate", rl.Vector3{}, 2); got != crate.Height*2 {
		t.Fatalf("expected doubling scale to double height, got %v", got)
	}
	if got := radius("crate", 2); got != crate.Footprint()*2 {
		t.Fatalf("expected doubling scale to double radius, got %v", got)
	}
}

func TestPropAt(t *testing.T) {
	w := ecs.NewWorld()
	SpawnDefaultForest(w)

	if _, ok := PropAt(w, 1000, 1000); ok {
		t.Fatalf("expected no prop far away")
	}

	first := landmarks[0]
	e, ok := PropAt(w, first.Position.X, first.Position.Z)
	if !ok {
		t.Fatalf("expected to find a prop at the first landmark's position")
	}
	m, _ := ecs.Get[comp.ModelRender](w, e)
	if m.AssetName != first.AssetName {
		t.Fatalf("expected %q at landmark position, got %q", first.AssetName, m.AssetName)
	}
}

// propSnapshot is Position, AssetName and effective scale flattened for
// easy sorting/comparison in the determinism/overlap tests below.
type propSnapshot struct {
	AssetName string
	Position  rl.Vector3
	Scale     float32
}

func snapshotProps(w *ecs.World) []propSnapshot {
	var out []propSnapshot
	ecs.Each3(w, func(_ ecs.Entity, t *comp.Transform, m *comp.ModelRender, _ *comp.Prop) {
		out = append(out, propSnapshot{AssetName: m.AssetName, Position: t.Position, Scale: m.EffectiveScale()})
	})
	sort.Slice(out, func(i, j int) bool {
		if out[i].Position.X != out[j].Position.X {
			return out[i].Position.X < out[j].Position.X
		}
		return out[i].Position.Z < out[j].Position.Z
	})
	return out
}

func TestGeneratedForestPropsDontOverlap(t *testing.T) {
	w := ecs.NewWorld()
	SpawnGeneratedForest(w, 42)
	props := snapshotProps(w)

	for i, a := range props {
		for j, b := range props {
			if i == j {
				continue
			}
			dx := a.Position.X - b.Position.X
			dz := a.Position.Z - b.Position.Z
			minDist := radius(a.AssetName, a.Scale) + radius(b.AssetName, b.Scale)
			if dx*dx+dz*dz < minDist*minDist-1e-3 {
				t.Fatalf("props %d (%s) and %d (%s) overlap: dist=%v minDist=%v",
					i, a.AssetName, j, b.AssetName, dx*dx+dz*dz, minDist*minDist)
			}
		}
	}
}

func TestGeneratedForestIsDeterministic(t *testing.T) {
	wa := ecs.NewWorld()
	SpawnGeneratedForest(wa, 7)
	wb := ecs.NewWorld()
	SpawnGeneratedForest(wb, 7)

	a, b := snapshotProps(wa), snapshotProps(wb)
	if len(a) != len(b) {
		t.Fatalf("expected same prop count for same seed, got %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("expected identical prop %d for same seed, got %+v vs %+v", i, a[i], b[i])
		}
	}
}

func isScatterAsset(name string) bool {
	for _, n := range scatterAssets {
		if n == name {
			return true
		}
	}
	return false
}

func TestSpawnAreaStaysClearOfScatteredProps(t *testing.T) {
	w := ecs.NewWorld()
	SpawnGeneratedForest(w, 1)

	ecs.Each3(w, func(_ ecs.Entity, tr *comp.Transform, m *comp.ModelRender, _ *comp.Prop) {
		// Landmarks are hand-placed and may legitimately sit inside the
		// nominal "spawn clear" radius (e.g. the campfire); only scattered
		// filler needs to avoid it.
		if !isScatterAsset(m.AssetName) {
			return
		}
		distSqr := tr.Position.X*tr.Position.X + tr.Position.Z*tr.Position.Z
		if distSqr < spawnClearRadius*spawnClearRadius {
			t.Fatalf("scattered prop %+v landed inside the spawn clearing (radius %v)", m, spawnClearRadius)
		}
	})
}

func TestSpawnAmbientCreaturesHaveNoPropTag(t *testing.T) {
	w := ecs.NewWorld()
	SpawnAmbientCreatures(w)

	count := 0
	ecs.Each[comp.CreatureVoice](w, func(e ecs.Entity, _ *comp.CreatureVoice) {
		count++
		if ecs.Has[comp.Prop](w, e) {
			t.Fatalf("expected ambient creature entity %v to not be tagged Prop (must not block movement)", e)
		}
		if !ecs.Has[comp.Transform](w, e) || !ecs.Has[comp.ModelRender](w, e) {
			t.Fatalf("expected ambient creature entity %v to have Transform+ModelRender", e)
		}
	})
	if count == 0 {
		t.Fatalf("expected at least one ambient creature to be spawned")
	}
}

func TestSpawnDefaultForestMarksCampfireSource(t *testing.T) {
	w := ecs.NewWorld()
	SpawnDefaultForest(w)

	found := 0
	ecs.Each[comp.CampfireSource](w, func(e ecs.Entity, _ *comp.CampfireSource) {
		found++
		m, ok := ecs.Get[comp.ModelRender](w, e)
		if !ok || m.AssetName != "campfire" {
			t.Fatalf("expected CampfireSource entity to render as \"campfire\", got %+v (ok=%v)", m, ok)
		}
	})
	if found != 1 {
		t.Fatalf("expected exactly 1 CampfireSource entity, got %d", found)
	}
}
