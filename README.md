# FoxDot

A small Go + [raylib](https://www.raylib.com/) game/framework test: control a fox in
third person around a stylized, low-poly forest clearing — run, jump on things, nibble,
and playfully swipe. This is scaffolding for a bigger project, so the emphasis so far is
on a clean, testable foundation rather than content, though a full low-poly art pack
(fox, forest critters, trees, plants and props — see [Assets](#assets-internalassets))
is already wired in.

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
cmd/foxdot/            Entry point: wires config, input, storage, assets and the game together.
internal/
  engine/               Game interface + the raylib window/main-loop runner (App).
  input/                Controller abstraction: Frame, keyboard/mouse + gamepad Sources, Manager.
  camera/               Third-person orbit camera (pure math, no window dependency).
  sim/                  Fox movement/jump/action simulation — pure Go, no raylib window state.
  assets/               Catalog + loader for the low-poly art pack under /assets.
  world/                Forest layout: ground + scattered low-poly props, height queries for collision.
  game/                 FoxDot scene: composes sim + camera + world + assets + storage into an engine.Game.
  config/               One JSON-serializable Config struct for window/camera/input/graphics knobs.
  storage/              Cross-platform file layer: JSON save/load, save-game slots, generated-content cache.
assets/                 Low-poly art pack: OBJ/MTL models + tiny PNG textures (see its own README.md).
saves/                  Not used at runtime (see File storage) — kept for local experimentation.
```

### Why the split?

The main design goal called out up front was testability, especially once more
simulation shows up (procedural generation, creature AI, etc). The pattern used
throughout:

- **Pure logic packages** (`sim`, `camera`, `world`'s height queries, `input`'s merging,
  `assets`'s catalog) never call raylib's window/GPU functions — only its plain math types
  and helpers (`rl.Vector3`, `rl.Vector3Lerp`, ...), which have no side effects. That means
  they're testable with plain `go test`, no window or GPU required — see each package's
  `*_test.go`. Only the handful of things that inherently need a live GPU context
  (`assets.Store`, `world.Forest.Draw`, `game.FoxDot.Draw`) are excluded.
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

## Assets (`internal/assets`)

The `/assets` folder is a small low-poly art pack (CC0) covering everything the forest
scene currently uses: a fox plus six ambient forest critters, four tree species, plants
(bushes/grass/mushrooms), and props (rocks/log/crate/fence/campfire). Each model is an
OBJ + MTL pair referencing tiny shared PNG textures — see `assets/README.md` for the pack's
own notes on scale/orientation/license.

`internal/assets` is the bridge between that folder and the engine:

- **`catalog.go`** (pure, no raylib) hardcodes each model's authored bounding box
  (`Width`/`Height`/`Depth`/`MinY`) and a per-model yaw correction, so other packages can
  place and collide with an asset without ever loading it. The catalog is checked against
  the real files on disk in `catalog_test.go`.
- **`root.go`** (`FindRoot`) locates the `assets/` folder whether you're running via
  `go run` from the repo or a packaged binary shipped with its own `assets/` folder next
  to it (or override with the `FOXDOT_ASSETS_DIR` env var).
- **`store.go`** (`Store`, needs a live window) lazily loads and caches `rl.Model`s by
  catalog name and draws them with the right ground offset and yaw correction applied. A
  failed/missing model draws as a small magenta wire cube instead of crashing, so a typo'd
  asset name is obvious rather than fatal.

`world.Forest` places `Prop{AssetName, Position, YawDegrees, Scale}` values and asks the
catalog for their size (`Prop.Top()`/`Prop.Radius()`) rather than duplicating dimensions;
`world.NewDefaultForest()` hand-places a few landmark trees/rocks/campsite props and
scatters bushes/grass/mushrooms around them from a fixed seed (rejection-sampled so
nothing overlaps) — see `internal/world/generate.go`.

**Orientation note:** the pack's animal models face local +X (confirmed by rendering
`fox.obj` from above), while the engine treats world +Z as "forward" at yaw 0. The
catalog's per-model yaw offset (`-90°` for animals) handles this automatically in
`Store.Draw`, so `sim`/`camera`'s yaw math never needs to know about it.

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

- Terrain height (currently a flat plane) feeding into `world.Forest.HeightAt`.
- Simple animation state driven by `sim.Fox.State.Action`/velocity (the models are static
  meshes — see `assets/README.md`'s notes on rigging).
- Ambient wildlife: the catalog already has deer/rabbit/squirrel/bird/butterfly/beetle
  models, just not wired into the scene yet.
- A real pause menu (the engine already routes the Pause action to `App`, currently wired
  to just quit).
