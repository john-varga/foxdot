package world

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/assets"
)

func TestHeightAtReturnsGroundAwayFromProps(t *testing.T) {
	f := &Forest{GroundY: 0, Props: []Prop{
		{AssetName: "rock_01", Position: rl.Vector3{X: 3, Z: 2}},
	}}

	if got := f.HeightAt(100, 100); got != 0 {
		t.Fatalf("expected ground height 0 far from props, got %v", got)
	}
}

func TestHeightAtReturnsPropTopWhenInsideFootprint(t *testing.T) {
	crate, ok := assets.ByName("crate")
	if !ok {
		t.Fatalf("expected crate to be in the asset catalog")
	}
	f := &Forest{GroundY: 0, Props: []Prop{
		{AssetName: "crate", Position: rl.Vector3{X: 3, Z: 2}},
	}}

	if got := f.HeightAt(3, 2); got != crate.Height {
		t.Fatalf("expected crate top height %v at its center, got %v", crate.Height, got)
	}

	// Just outside the crate's circular footprint.
	outside := 3 + crate.Footprint() + 0.5
	if got := f.HeightAt(outside, 2); got != 0 {
		t.Fatalf("expected ground height 0 just outside footprint, got %v", got)
	}
}

func TestHeightAtPicksTallestOverlappingProp(t *testing.T) {
	fence, _ := assets.ByName("fence") // shorter
	campfire, _ := assets.ByName("campfire")
	f := &Forest{GroundY: 0, Props: []Prop{
		{AssetName: "fence", Position: rl.Vector3{X: 0, Z: 0}},
		{AssetName: "campfire", Position: rl.Vector3{X: 0, Z: 0}},
	}}

	want := fence.Height
	if campfire.Height > want {
		want = campfire.Height
	}
	if got := f.HeightAt(0, 0); got != want {
		t.Fatalf("expected tallest overlapping prop top %v, got %v", want, got)
	}
}

func TestPropWithUnknownAssetDefaultsToGroundHeight(t *testing.T) {
	p := Prop{AssetName: "does-not-exist", Position: rl.Vector3{X: 1, Y: 2, Z: 3}}
	if got := p.Top(); got != 2 {
		t.Fatalf("expected unknown asset to default Top() to its base Y, got %v", got)
	}
	if got := p.Radius(); got != 0 {
		t.Fatalf("expected unknown asset to have zero radius, got %v", got)
	}
}

func TestScaleAffectsTopAndRadius(t *testing.T) {
	crate, _ := assets.ByName("crate")
	p1 := Prop{AssetName: "crate", Position: rl.Vector3{}}
	p2 := Prop{AssetName: "crate", Position: rl.Vector3{}, Scale: 2}

	if p2.Top() != p1.Top()*2 {
		t.Fatalf("expected doubling scale to double height: p1=%v p2=%v", p1.Top(), p2.Top())
	}
	if p2.Radius() != crate.Footprint()*2 {
		t.Fatalf("expected doubling scale to double radius, got %v", p2.Radius())
	}
}

func TestPropAt(t *testing.T) {
	f := NewDefaultForest()
	if _, ok := f.PropAt(1000, 1000); ok {
		t.Fatalf("expected no prop far away")
	}
	if len(f.Props) == 0 {
		t.Fatalf("expected default forest to have some props")
	}
	first := f.Props[0]
	if p, ok := f.PropAt(first.Position.X, first.Position.Z); !ok || p.AssetName != first.AssetName {
		t.Fatalf("expected to find prop %q at its own position, got %+v (ok=%v)", first.AssetName, p, ok)
	}
}

func TestGeneratedForestPropsDontOverlap(t *testing.T) {
	f := NewGeneratedForest(42)
	for i, a := range f.Props {
		for j, b := range f.Props {
			if i == j {
				continue
			}
			dx := a.Position.X - b.Position.X
			dz := a.Position.Z - b.Position.Z
			minDist := a.Radius() + b.Radius()
			if dx*dx+dz*dz < minDist*minDist-1e-3 {
				t.Fatalf("props %d (%s) and %d (%s) overlap: dist=%v minDist=%v",
					i, a.AssetName, j, b.AssetName, dx*dx+dz*dz, minDist*minDist)
			}
		}
	}
}

func TestGeneratedForestIsDeterministic(t *testing.T) {
	a := NewGeneratedForest(7)
	b := NewGeneratedForest(7)
	if len(a.Props) != len(b.Props) {
		t.Fatalf("expected same prop count for same seed, got %d vs %d", len(a.Props), len(b.Props))
	}
	for i := range a.Props {
		if a.Props[i] != b.Props[i] {
			t.Fatalf("expected identical prop %d for same seed, got %+v vs %+v", i, a.Props[i], b.Props[i])
		}
	}
}

func TestSpawnAreaStaysClearOfScatteredProps(t *testing.T) {
	f := NewGeneratedForest(1)
	for _, p := range f.Props[len(landmarks):] { // scattered props are appended after the landmarks
		distSqr := p.Position.X*p.Position.X + p.Position.Z*p.Position.Z
		if distSqr < spawnClearRadius*spawnClearRadius {
			t.Fatalf("scattered prop %+v landed inside the spawn clearing (radius %v)", p, spawnClearRadius)
		}
	}
}
