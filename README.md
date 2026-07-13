# FoxDot

A small Go + [raylib](https://www.raylib.com/) game/framework test: control a fox in
third person around a stylized, low-poly forest clearing — run, jump on things, nibble,
and playfully swipe. This is scaffolding for a bigger project, so the emphasis so far is
on a clean, testable foundation rather than content: real art assets, terrain, and more
gameplay will layer on top of this.

## Quick start

```bash
go run ./cmd/foxdot
```

First run creates a config file (see [Configuration](#configuration)) and an autosave on
quit; both live in your OS's standard per-user app data location (see
[File storage](#file-storage)).

### Controls

| Action  | Keyboard      | Gamepad         |
|---------|---------------|-----------------|
| Move    | WASD          | Left stick      |
| Look    | Mouse         | Right stick     |
| Jump    | Space         | A / Cross       |
| Sprint  | Left Shift    | Left trigger    |
| Nibble  | E             | X / Square      |
| Swipe   | F             | B / Circle      |
| Quicksave | Enter       | A / Cross (Confirm) |
| Quit    | Escape        | Start / Options |

All bindings, sensitivities and deadzones are just config values — see below.

## Requirements

- Go 1.21+
- A C compiler (cgo) — Xcode Command Line Tools on macOS, a MinGW toolchain on Windows,
  `gcc`/`build-essential` on Linux. `github.com/gen2brain/raylib-go` vendors raylib's C
  sources and compiles them via cgo, so **no system raylib install is required** on any
  of the three platforms.

  (raylib-go also has an experimental cgo-free "purego" mode that dlopen's a prebuilt
  raylib shared library at runtime. It's appealing for distributing binaries without a C
  toolchain, but on this machine it additionally required a system `libffi` that wasn't
  present, so this project sticks with the well-trodden cgo path for now.)

## Project layout

```
cmd/foxdot/            Entry point: wires config, input, storage and the game together.
internal/
  engine/               Game interface + the raylib window/main-loop runner (App).
  input/                Controller abstraction: Frame, keyboard/mouse + gamepad Sources, Manager.
  camera/               Third-person orbit camera (pure math, no window dependency).
  sim/                  Fox movement/jump/action simulation — pure Go, no raylib window state.
  world/                Placeholder forest: ground + low-poly props, height queries for collision.
  game/                 FoxDot scene: composes sim + camera + world + storage into an engine.Game.
  config/               One JSON-serializable Config struct for window/camera/input/graphics knobs.
  storage/              Cross-platform file layer: JSON save/load, save-game slots, generated-content cache.
assets/                 Placeholder for art (models, textures) once available.
saves/                  Not used at runtime (see File storage) — kept for local experimentation.
```

### Why the split?

The main design goal called out up front was testability, especially once more
simulation shows up (procedural generation, creature AI, etc). The pattern used
throughout:

- **Pure logic packages** (`sim`, `camera`, `world`'s height queries, `input`'s merging)
  never call raylib's window/GPU functions — only its plain math types and helpers
  (`rl.Vector3`, `rl.Vector3Lerp`, ...), which have no side effects. That means they're
  testable with plain `go test`, no window or GPU required — see each package's
  `*_test.go`.
- **`engine.Game`** separates `Update(dt, input.Frame)` (pure simulation tick) from
  `Draw()` (rendering). `game.FoxDot` follows the same split: its `Init`/`Update`/
  `Shutdown` never touch raylib's drawing API and are unit tested directly; only `Draw`
  (untested, by nature) issues `rl.Draw*` calls.
- **`engine.App`** is the only place that owns the raylib window lifecycle and runs the
  frame loop (poll input → Update → clear → Draw → present).

Run everything with:

```bash
go test ./...
```

## Controller abstraction (`internal/input`)

Gameplay code never touches raylib's `IsKeyDown`/`IsGamepadButtonDown` etc. directly.
Instead:

- `input.Frame` is a normalized, device-agnostic snapshot of player intent for one tick
  (`Move`, `Look`, and action buttons like `Jump`/`Nibble`/`Swipe`).
- `input.Source` implementations (`KeyboardMouseSource`, `GamepadSource`) turn raw device
  state into a `Frame` according to `input.Config` (key/button bindings, deadzones,
  sensitivities).
- `input.Manager` polls every configured `Source` each tick and merges the results
  (buttons OR together; axes take whichever source has the larger input), so keyboard and
  a gamepad can both be plugged in and used interchangeably without gameplay code caring.

Adding a new device later (e.g. a Steam Deck-specific mapping) means adding a new
`Source`, not touching `sim` or `game`.

## Configuration

`internal/config.Config` bundles every tweakable knob — window size/vsync/fps, camera
distance/FOV/pitch limits/smoothing, all input bindings and sensitivities, and a couple
of graphics toggles — into one struct that's saved as pretty-printed JSON. On first run
it's written out with defaults; edit the file and relaunch to retune anything without
recompiling.

## File storage (`internal/storage`)

`storage.Store` resolves an OS-appropriate per-user app directory via
`os.UserConfigDir()`:

- macOS: `~/Library/Application Support/FoxDot`
- Windows: `%AppData%\FoxDot`
- Linux: `$XDG_CONFIG_HOME/FoxDot` (or `~/.config/FoxDot`)

On top of generic atomic JSON read/write helpers, it exposes save-game slots
(`SaveGameSlot`/`LoadGameSlot`/`ListGameSlots`/`DeleteGameSlot`) and a similar cache for
procedurally-generated content (`SaveGenerated`/`LoadGenerated`) — useful once the forest
layout, or anything else, starts being generated rather than hand-placed.

`game.FoxDot` currently uses a single `"autosave"` slot: it loads on `Init` and saves on
quicksave (Enter/Confirm) and on `Shutdown`.

## What's next

- Swap the placeholder capsule/box primitives for real low-poly fox and forest models.
- Terrain height (currently a flat plane) feeding into `world.Forest.HeightAt`.
- Simple animation state driven by `sim.Fox.State.Action`/velocity.
- A real pause menu (the engine already routes the Pause action to `App`, currently wired
  to just quit).
