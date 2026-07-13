package world

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestHeightAtReturnsGroundAwayFromProps(t *testing.T) {
	f := &Forest{
		GroundY: 0,
		Props: []Prop{
			{Position: rl.Vector3{X: 3, Y: 0, Z: 2}, Size: rl.Vector3{X: 1, Y: 0.6, Z: 1}},
		},
	}

	if got := f.HeightAt(100, 100); got != 0 {
		t.Fatalf("expected ground height 0 far from props, got %v", got)
	}
}

func TestHeightAtReturnsPropTopWhenInsideFootprint(t *testing.T) {
	f := &Forest{
		GroundY: 0,
		Props: []Prop{
			{Position: rl.Vector3{X: 3, Y: 0, Z: 2}, Size: rl.Vector3{X: 1, Y: 0.6, Z: 1}},
		},
	}

	if got := f.HeightAt(3, 2); got != 0.6 {
		t.Fatalf("expected prop top height 0.6 at prop center, got %v", got)
	}
	// Just inside the footprint edge.
	if got := f.HeightAt(3.4, 2.4); got != 0.6 {
		t.Fatalf("expected prop top height 0.6 near footprint edge, got %v", got)
	}
	// Just outside the footprint edge.
	if got := f.HeightAt(3.6, 2); got != 0 {
		t.Fatalf("expected ground height 0 just outside footprint, got %v", got)
	}
}

func TestHeightAtPicksTallestOverlappingProp(t *testing.T) {
	f := &Forest{
		GroundY: 0,
		Props: []Prop{
			{Position: rl.Vector3{X: 0, Y: 0, Z: 0}, Size: rl.Vector3{X: 4, Y: 0.5, Z: 4}},
			{Position: rl.Vector3{X: 0, Y: 0.5, Z: 0}, Size: rl.Vector3{X: 1, Y: 2, Z: 1}},
		},
	}

	if got := f.HeightAt(0, 0); got != 2.5 {
		t.Fatalf("expected tallest overlapping prop top 2.5, got %v", got)
	}
}

func TestPropAt(t *testing.T) {
	f := NewPlaceholderForest()
	if _, ok := f.PropAt(1000, 1000); ok {
		t.Fatalf("expected no prop far away")
	}
	if len(f.Props) == 0 {
		t.Fatalf("expected placeholder forest to have some props")
	}
	first := f.Props[0]
	if p, ok := f.PropAt(first.Position.X, first.Position.Z); !ok || p.Name != first.Name {
		t.Fatalf("expected to find prop %q at its own position, got %+v (ok=%v)", first.Name, p, ok)
	}
}
