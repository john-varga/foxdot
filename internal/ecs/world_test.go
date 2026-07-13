package ecs

import "testing"

type Position struct{ X, Y float32 }
type Velocity struct{ X, Y float32 }
type Tag struct{}

func TestNewEntityIsAliveAndUnique(t *testing.T) {
	w := NewWorld()
	a := w.NewEntity()
	b := w.NewEntity()

	if a == b {
		t.Fatalf("expected distinct entities, got %v and %v", a, b)
	}
	if !w.Alive(a) || !w.Alive(b) {
		t.Fatalf("expected both entities to be alive")
	}
	if w.Alive(Entity(0)) {
		t.Fatalf("zero entity should never be alive")
	}
}

func TestSetGetHasRemoveComponent(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()

	if Has[Position](w, e) {
		t.Fatalf("expected no Position before Set")
	}

	Set(w, e, Position{X: 1, Y: 2})
	pos, ok := Get[Position](w, e)
	if !ok {
		t.Fatalf("expected Position after Set")
	}
	if pos.X != 1 || pos.Y != 2 {
		t.Fatalf("got %+v, want {1 2}", *pos)
	}

	// Mutating through the returned pointer should stick.
	pos.X = 5
	pos2, _ := Get[Position](w, e)
	if pos2.X != 5 {
		t.Fatalf("expected mutation via pointer to persist, got %+v", *pos2)
	}

	Remove[Position](w, e)
	if Has[Position](w, e) {
		t.Fatalf("expected no Position after Remove")
	}
}

func TestDestroyRemovesAllComponents(t *testing.T) {
	w := NewWorld()
	e := w.NewEntity()
	Set(w, e, Position{X: 1})
	Set(w, e, Velocity{X: 2})

	w.Destroy(e)

	if w.Alive(e) {
		t.Fatalf("expected entity to be dead after Destroy")
	}
	if Has[Position](w, e) || Has[Velocity](w, e) {
		t.Fatalf("expected components to be gone after Destroy")
	}

	// Destroying again, or destroying something never created, must not panic.
	w.Destroy(e)
	w.Destroy(Entity(9999))
}

func TestEachVisitsOnlyMatchingEntities(t *testing.T) {
	w := NewWorld()
	withPos := w.NewEntity()
	Set(w, withPos, Position{X: 1})

	withoutPos := w.NewEntity()
	Set(w, withoutPos, Velocity{X: 2})

	seen := map[Entity]bool{}
	Each(w, func(e Entity, p *Position) {
		seen[e] = true
	})

	if !seen[withPos] {
		t.Fatalf("expected to visit entity with Position")
	}
	if seen[withoutPos] {
		t.Fatalf("did not expect to visit entity without Position")
	}
	if len(seen) != 1 {
		t.Fatalf("expected exactly 1 visited entity, got %d", len(seen))
	}
}

func TestEach2RequiresBothComponents(t *testing.T) {
	w := NewWorld()

	both := w.NewEntity()
	Set(w, both, Position{X: 1})
	Set(w, both, Velocity{X: 10})

	onlyPos := w.NewEntity()
	Set(w, onlyPos, Position{X: 2})

	onlyVel := w.NewEntity()
	Set(w, onlyVel, Velocity{X: 20})

	visited := 0
	Each2(w, func(e Entity, p *Position, v *Velocity) {
		visited++
		if e != both {
			t.Fatalf("expected only 'both' entity to be visited, got %v", e)
		}
		// Systems commonly integrate velocity into position; verify mutation
		// through the pointers is visible afterward.
		p.X += v.X
	})

	if visited != 1 {
		t.Fatalf("expected 1 visit, got %d", visited)
	}
	p, _ := Get[Position](w, both)
	if p.X != 11 {
		t.Fatalf("expected integrated position 11, got %v", p.X)
	}
}

func TestEach3RequiresAllThreeComponents(t *testing.T) {
	w := NewWorld()

	all := w.NewEntity()
	Set(w, all, Position{})
	Set(w, all, Velocity{})
	Set(w, all, Tag{})

	missingTag := w.NewEntity()
	Set(w, missingTag, Position{})
	Set(w, missingTag, Velocity{})

	visited := []Entity{}
	Each3(w, func(e Entity, p *Position, v *Velocity, tag *Tag) {
		visited = append(visited, e)
	})

	if len(visited) != 1 || visited[0] != all {
		t.Fatalf("expected only 'all' entity visited, got %v", visited)
	}
}

func TestCount(t *testing.T) {
	w := NewWorld()
	if Count[Position](w) != 0 {
		t.Fatalf("expected 0 before any Set")
	}
	Set(w, w.NewEntity(), Position{})
	Set(w, w.NewEntity(), Position{})
	if Count[Position](w) != 2 {
		t.Fatalf("expected 2, got %d", Count[Position](w))
	}
}

func TestResources(t *testing.T) {
	w := NewWorld()

	if _, ok := GetResource[int](w); ok {
		t.Fatalf("expected no int resource before SetResource")
	}

	SetResource(w, 42)
	v, ok := GetResource[int](w)
	if !ok || *v != 42 {
		t.Fatalf("expected resource 42, got %v ok=%v", v, ok)
	}

	*v = 100
	v2, _ := GetResource[int](w)
	if *v2 != 100 {
		t.Fatalf("expected mutation via pointer to persist, got %d", *v2)
	}

	SetResource(w, 7)
	v3, _ := GetResource[int](w)
	if *v3 != 7 {
		t.Fatalf("expected overwritten resource 7, got %d", *v3)
	}
}

func TestMustGetResourcePanicsWhenMissing(t *testing.T) {
	w := NewWorld()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for missing resource")
		}
	}()
	MustGetResource[int](w)
}

func TestEntitiesAndLen(t *testing.T) {
	w := NewWorld()
	if w.Len() != 0 {
		t.Fatalf("expected empty world")
	}
	a := w.NewEntity()
	b := w.NewEntity()
	if w.Len() != 2 {
		t.Fatalf("expected 2 entities, got %d", w.Len())
	}
	ids := w.Entities()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	found := map[Entity]bool{}
	for _, id := range ids {
		found[id] = true
	}
	if !found[a] || !found[b] {
		t.Fatalf("expected both entities in Entities(), got %v", ids)
	}
}
