// Package ecs is a small, dependency-free Entity Component System.
//
// It is intentionally minimal: entities are opaque IDs, components are plain
// Go structs stored in per-type tables, and "systems" are just ordinary
// functions that take a *World and run each tick (see internal/game for the
// systems that make up FoxDot). There's no archetype storage or parallel
// scheduling — a few hundred entities is plenty for this game, and a simple
// map-per-component-type keeps the package easy to read, test, and extend.
//
// Singleton state that isn't per-entity (the clock, the input frame, the
// camera, ...) lives in "resources", which are just one value per Go type
// stored on the World alongside the component tables.
package ecs

import "reflect"

// Entity is an opaque handle to a bundle of components. The zero Entity is
// never issued by World.NewEntity and can be used as a "no entity" sentinel.
type Entity uint32

type componentTable interface {
	remove(e Entity)
}

// World owns all entities, their components, and any resources.
type World struct {
	nextID    Entity
	alive     map[Entity]struct{}
	tables    map[reflect.Type]componentTable
	resources map[reflect.Type]any
}

// NewWorld creates an empty World.
func NewWorld() *World {
	return &World{
		alive:     make(map[Entity]struct{}),
		tables:    make(map[reflect.Type]componentTable),
		resources: make(map[reflect.Type]any),
	}
}

// NewEntity allocates a fresh Entity with no components.
func (w *World) NewEntity() Entity {
	w.nextID++
	e := w.nextID
	w.alive[e] = struct{}{}
	return e
}

// Alive reports whether e was created and has not been destroyed.
func (w *World) Alive(e Entity) bool {
	_, ok := w.alive[e]
	return ok
}

// Destroy removes an entity and all of its components.
func (w *World) Destroy(e Entity) {
	if !w.Alive(e) {
		return
	}
	delete(w.alive, e)
	for _, t := range w.tables {
		t.remove(e)
	}
}

// Entities returns every live entity, in no particular order.
func (w *World) Entities() []Entity {
	out := make([]Entity, 0, len(w.alive))
	for e := range w.alive {
		out = append(out, e)
	}
	return out
}

// Len reports the number of live entities.
func (w *World) Len() int {
	return len(w.alive)
}

func typeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

type table[T any] struct {
	data map[Entity]*T
}

func (t *table[T]) remove(e Entity) {
	delete(t.data, e)
}

func tableFor[T any](w *World) *table[T] {
	t := typeOf[T]()
	if existing, ok := w.tables[t]; ok {
		return existing.(*table[T])
	}
	nt := &table[T]{data: make(map[Entity]*T)}
	w.tables[t] = nt
	return nt
}

// Set attaches (or overwrites) component T on e and returns a pointer to the
// stored value for further in-place mutation.
func Set[T any](w *World, e Entity, c T) *T {
	tbl := tableFor[T](w)
	ptr := new(T)
	*ptr = c
	tbl.data[e] = ptr
	return ptr
}

// Get returns a mutable pointer to e's component T, if present.
func Get[T any](w *World, e Entity) (*T, bool) {
	tbl := tableFor[T](w)
	v, ok := tbl.data[e]
	return v, ok
}

// Has reports whether e has a component of type T.
func Has[T any](w *World, e Entity) bool {
	_, ok := Get[T](w, e)
	return ok
}

// Remove detaches component T from e, if present.
func Remove[T any](w *World, e Entity) {
	tableFor[T](w).remove(e)
}

// Count returns how many entities currently have a component of type T.
func Count[T any](w *World) int {
	return len(tableFor[T](w).data)
}

// Each calls fn for every entity that has a component of type T.
func Each[T any](w *World, fn func(Entity, *T)) {
	for e, c := range tableFor[T](w).data {
		fn(e, c)
	}
}

// Each2 calls fn for every entity that has both A and B.
func Each2[A, B any](w *World, fn func(Entity, *A, *B)) {
	as := tableFor[A](w)
	bs := tableFor[B](w)
	for e, a := range as.data {
		if b, ok := bs.data[e]; ok {
			fn(e, a, b)
		}
	}
}

// Each3 calls fn for every entity that has A, B, and C.
func Each3[A, B, C any](w *World, fn func(Entity, *A, *B, *C)) {
	as := tableFor[A](w)
	bs := tableFor[B](w)
	cs := tableFor[C](w)
	for e, a := range as.data {
		b, ok := bs.data[e]
		if !ok {
			continue
		}
		c, ok := cs.data[e]
		if !ok {
			continue
		}
		fn(e, a, b, c)
	}
}

// SetResource stores a singleton value of type T on the World, replacing any
// existing value of that type.
func SetResource[T any](w *World, v T) *T {
	t := typeOf[T]()
	ptr := new(T)
	*ptr = v
	w.resources[t] = ptr
	return ptr
}

// GetResource returns a mutable pointer to the singleton value of type T, if
// one has been set.
func GetResource[T any](w *World) (*T, bool) {
	t := typeOf[T]()
	v, ok := w.resources[t]
	if !ok {
		return nil, false
	}
	return v.(*T), true
}

// MustGetResource is like GetResource but panics if the resource is missing.
// Intended for resources a System requires to have been set up during game
// Init, where a missing resource is a programmer error, not a runtime one.
func MustGetResource[T any](w *World) *T {
	v, ok := GetResource[T](w)
	if !ok {
		var zero T
		panic("ecs: missing required resource " + typeOf[T]().String() + " (want " + reflect.TypeOf(zero).String() + ")")
	}
	return v
}
