// Package engine provides the thin runner that owns the raylib window and
// drives a Game implementation every frame. The Game interface deliberately
// separates Update (pure simulation tick, driven by an input.Frame) from
// Draw (rendering), so gameplay/simulation logic can be unit tested without
// ever touching a window or GPU context — only Draw and the App runner
// itself require a live raylib window.
package engine

import "foxdot/internal/input"

// Game is the abstraction every playable scene (the fox/forest demo today,
// menus or other scenes later) implements.
type Game interface {
	// Init is called once after the window is created, before the first
	// Update/Draw.
	Init() error

	// Update advances gameplay simulation by dt seconds using this frame's
	// input. It must not touch raylib's rendering API.
	Update(dt float32, in input.Frame) error

	// Draw renders the current state. Called once per frame between
	// rl.BeginDrawing/EndDrawing.
	Draw()

	// Shutdown releases any resources (textures, models, sounds) acquired
	// in Init.
	Shutdown()
}
