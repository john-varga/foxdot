# FoxDot

A small Go + [raylib](https://www.raylib.com/) game/framework test: control a fox in
third person around a stylized, low-poly forest clearing that breathes through a
day/night cycle — run, jump on things, nibble, and playfully swipe, while ambience and
music shift with the time of day. This is scaffolding for a bigger project, so the
emphasis so far is on a clean, testable foundation rather than content, though a full
low-poly art pack and an original music/ambience/SFX pack (see
[Assets](#assets-internalassets-and-internalaudio)) are already wired in, and gameplay is
built as an Entity Component System (see [Architecture](#architecture-internalecs)) rather
than a class hierarchy.

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
  ecs/                  Minimal Entity Component System core: World, Entity, components, resources, queries.
  comp/                 Shared component/tag types (Transform, ModelRender, ActionState, ...).
  engine/               Game interface + the raylib window/main-loop runner (App).
  input/                Controller abstraction: Frame, keyboard/mouse + gamepad Sources, Manager.
  camera/               Third-person orbit camera (pure math, no window dependency); an ECS resource.
  worldtime/            Pure day/night clock (time of day, phase, sun height); an ECS resource.
  sim/                  Fox movement/jump/action simulation — pure Go, no raylib window state.
  assets/               Catalog + loader for the low-poly art pack under /assets.
  audio/                Audio catalog, playback engine, ambience mixer, music director, SFX event queue.
  world/                Forest layout: spawns prop/creature entities, ground plane, height queries.
  game/                 FoxDot scene: builds the ecs.World and runs its Systems each tick (engine.Game).
  config/               One JSON-serializable Config struct for window/camera/input/graphics/audio knobs.
  storage/              Cross-platform file layer: JSON save/load, save-game slots, generated-content cache.
assets/                 Art + audio packs: models/textures (CC0) and music/ambience/SFX (see its own README.md).
saves/                  Not used at runtime (see File storage) — kept for local experimentation.
```

### Why the split?

The main design goal called out up front was testability, especially once more
simulation shows up (procedural generation, creature AI, etc). The pattern used
throughout:

- **Pure logic packages** (`sim`, `camera`, `worldtime`, `world`'s spawn/height queries,
  `input`'s merging, `assets`'s catalog, `audio`'s catalog and mix/track *decisions*) never
  call raylib's window/GPU/audio functions — only its plain math types and helpers
  (`rl.Vector3`, `rl.Vector3Lerp`, ...), which have no side effects. That means they're
  testable with plain `go test`, no window, GPU or audio device required — see each
  package's `*_test.go`. Only the handful of things that inherently need a live GPU/audio
  context (`assets.Store`, `audio.Engine`, `world.DrawGround`, `game.FoxDot.Draw`) are
  excluded.
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

## Architecture (`internal/ecs`)

Gameplay is built as an Entity Component System rather than a class hierarchy: the fox,
every forest prop (tree, rock, bush, ...) and every ambient creature are all just an
`ecs.Entity` (an opaque ID) plus whichever plain-data components (`internal/comp`) it
happens to have — there's no `Fox` or `Prop` *object* with its own methods and identity.
Systems are ordinary functions that operate on a `*ecs.World`:

- `ecs.Set`/`Get`/`Has`/`Remove` attach/query a component type on an entity.
- `ecs.Each`/`Each2`/`Each3` iterate every entity that has one, two, or three given
  component types — e.g. `RenderSystem` (`internal/game/systems.go`) draws *every* entity
  with `(Transform, ModelRender)` the same way, whether it's the fox, a pine tree, or a
  butterfly.
- `ecs.SetResource`/`GetResource` hold singleton state that isn't per-entity — the
  `worldtime.Clock`, the `camera.ThirdPerson` orbit state, and the `audio.EventQueue` all
  live as resources, read/written by whichever System needs them.

`internal/game/foxdot.go` builds the `ecs.World` (spawning the player entity and the
forest via `world.SpawnDefaultForest`) and its `Update` just calls each System in order
(`CameraSystem` → `FoxSystem` → `CreatureVoiceSystem` → audio update). `internal/sim`
stays a free function (`sim.Step`) operating on plain `sim.State`/`sim.Params` — the same
pure, unit-tested movement logic as before, just driven by ECS components instead of a
bespoke `Fox` struct, so any future creature could reuse it with its own components.

## Day/night cycle (`internal/worldtime`)

`worldtime.Clock` is a small, pure resource tracking elapsed time as a fraction of a
configurable-length day (`DaySeconds`, 6 minutes by default so a full cycle is easy to see
in a short play session). It derives:

- **`Phase()`** — `Night`/`Dawn`/`Day`/`Dusk`, from configurable dawn/dusk windows.
- **`SunHeight()`** — a continuous `[-1, 1]` value (for future lighting) so things can fade
  smoothly instead of snapping at phase boundaries.

`TimeSystem` (folded into `FoxDot.Update`) advances it every tick; `TimeOfDay`/`Day` are
persisted in saves so reloading resumes at the same time of day instead of always
restarting at midnight.

## Audio (`internal/audio`)

- **`catalog.go`** (pure, no raylib) parses `assets/audio/manifest.json` into a `Catalog`
  of `Track`s (music, ambience loops, ambience one-shots, SFX), each already knowing
  whether it should stream (`Music`, for long/looping audio) or fully decode
  (`Sound`, for short one-shots).
- **`engine.go`** (`Engine`, needs a live audio device) lazily loads/caches tracks, plays
  overlapping SFX via `rl.LoadSoundAlias` so one bark doesn't cut off another, and pumps
  every playing `Music` stream each frame via `Update`.
- **`mixer.go`** (`Mixer`) crossfades a set of streamed tracks toward target volume
  weights — the one piece of fade logic shared by:
  - **`ambience.go`** (`AmbiencePlayer`) — `AmbienceWeights(phase, campfireProximity)` is a
    pure decision function (birds by day, crickets/frogs by night, wind/rustling leaves as
    a year-round bed, campfire crackle fading in with proximity to a `CampfireSource`
    entity) fed into a `Mixer`.
  - **`music.go`** (`MusicDirector`) — `ChooseMusicTrack(situation, phase)` picks one track
    (`forest_day`/`forest_night` while exploring; `main_menu`/`discovery`/`quiet_danger`/
    `credits` once something sets a different `Situation`) and crossfades to it via the
    same `Mixer` — playing "just one track" is really just a mix with a single nonzero
    weight.
- **`events.go`** (`EventQueue`) decouples "something happened that makes noise" from
  "how it's played": any System pushes `events.Push("jump", 1)` and an `AudioSystem`-style
  step (in `FoxDot.updateAudio`) drains the queue into `Engine.PlaySFX` once per frame. The
  fox's jump/footsteps/swipe, ambient creatures' periodic chirps/hops/calls, and even the
  quicksave confirmation (`ui_confirm` — standing in for real UI SFX until there's a menu
  to click) all go through this same path.

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

## Assets (`internal/assets` and `internal/audio`)

The `/assets` folder bundles two separately-licensed packs (see `assets/README.md` and
`assets/LICENSE.txt` for the full breakdown):

- A low-poly **visual** pack (CC0): a fox plus six ambient forest critters, four tree
  species, plants (bushes/grass/mushrooms), and props (rocks/log/crate/fence/campfire).
  Each model is an OBJ + MTL pair referencing tiny shared PNG textures.
- An original **audio** pack: music (WAV + editable MIDI), seamless ambience loops, a
  thunder one-shot, and sound effects (footsteps, player actions, UI, animal calls).

`internal/assets` is the bridge between the visual pack and the engine:

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

`internal/world` spawns entities from that catalog rather than duplicating dimensions:
`world.SpawnDefaultForest` hand-places a few landmark trees/rocks/campsite props, scatters
bushes/grass/mushrooms around them from a fixed seed (rejection-sampled so nothing
overlaps — see `internal/world/spawn.go`), and spawns a handful of ambient creatures each
with a `CreatureVoice` component tied to the audio pack's animal SFX. `world.HeightAt`
queries every `Prop`-tagged entity's `(Transform, ModelRender)` for collision.

**Orientation note:** the pack's animal models face local +X (confirmed by rendering
`fox.obj` from above), while the engine treats world +Z as "forward" at yaw 0. The
catalog's per-model yaw offset (`-90°` for animals) handles this automatically in
`Store.Draw`, so `sim`/`camera`'s yaw math never needs to know about it.

## Configuration

`internal/config.Config` bundles every tweakable knob — window size/vsync/fps, camera
distance/FOV/pitch limits/smoothing, all input bindings and sensitivities, graphics
toggles, and per-category audio volumes (master/SFX/music/ambience) — into one struct
that's saved as pretty-printed JSON. On first run it's written out with defaults; edit the
file and relaunch to retune anything without recompiling.

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
quicksave (Enter/Confirm) and on `Shutdown`. Saved state includes the fox's
position/facing, the camera orbit, playtime, and the world clock's time of day/day count.

## What's next

- Terrain height (currently a flat plane) feeding into `world.HeightAt`.
- Simple animation state driven by an entity's `ActionState`/`Velocity` (the models are
  static meshes — see `assets/README.md`'s notes on rigging).
- Real wildlife AI: ambient creatures currently just stand in place and periodically make
  noise (`comp.CreatureVoice`) rather than wandering.
- A real pause menu / HUD — `internal/audio` already has UI SFX (`ui_click`, `ui_back`,
  `ui_page`, ...) loaded and ready via the same `EventQueue` the quicksave confirmation
  sound uses; there's just nothing to click yet.
- Lighting/skybox driven by `worldtime.Clock.SunHeight()`, now that day/night is tracked.
