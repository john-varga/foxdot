# AGENTS.md — context for AI agents working on FoxDot

This file is for AI coding agents (and future-you). It captures the *why* behind the
codebase, not just the *what* — the README covers usage/features, this covers
conventions, gotchas, and where things are headed. Keep both in sync when you make
structural changes.

## What this project is

A Go + [raylib](https://www.raylib.com/) (`github.com/gen2brain/raylib-go`) game/framework
test: a fox explores a low-poly forest clearing in third person, with a day/night cycle
and a small audio subsystem. It's deliberately scoped as scaffolding/a testbed, not a
finished game — the emphasis throughout has been a clean, well-tested foundation over
content. No multiplayer, ever.

Gameplay is built as an **Entity Component System** (`internal/ecs`), not an
object-oriented class hierarchy — this was an explicit, deliberate architectural choice
(see "ECS, not OO" below), not a default.

## Architecture at a glance

```
cmd/foxdot/            Entry point. Wires config → input → assets → game → engine.App, then Run().
internal/
  ecs/                  Generic ECS core: World, Entity, Set/Get/Has/Remove, Each/Each2/Each3, resources.
  comp/                 Shared component/tag structs (Transform, Velocity, ModelRender, ActionState, ...).
  engine/               Game interface (Init/Update/Draw/Shutdown) + App (owns the raylib window/main loop).
  input/                Frame (device-agnostic input snapshot), Source impls, Manager (merges Sources).
  camera/               ThirdPerson orbit camera. Pure math; held as an ecs resource.
  worldtime/            Clock: day/night time-of-day, Phase(), SunHeight(). Pure; held as an ecs resource.
  sim/                  sim.Step: pure fox movement/jump/action logic over sim.State/sim.Params.
  assets/               Visual asset catalog (pure) + Store (raylib model loader/cache/draw).
  audio/                Audio catalog (pure) + Engine (raylib playback) + Mixer/AmbiencePlayer/MusicDirector + EventQueue.
  world/                Spawns prop/creature entities into an ecs.World; ground plane; HeightAt/PropAt queries.
  game/                 FoxDot: builds the ecs.World, owns Systems (systems.go), implements engine.Game.
  config/               Config: one JSON-serializable struct for every tunable (window/camera/input/graphics/audio).
  storage/              Cross-platform per-user file storage: JSON save/load, save slots, generated-content cache.
assets/                 Visual pack (models/textures, CC0) + audio pack (music/ambience/sfx). See assets/README.md.
```

### The core testability pattern

This is the single most important convention in the codebase: **pure logic is separated
from anything requiring a live window/GPU/audio device**, so almost everything can be
unit tested with plain `go test`, no display/audio hardware required.

- Packages that only use raylib's *plain math types and pure helpers* (`rl.Vector3`,
  `rl.Vector3Lerp`, `rl.Clamp`, ...) — never `InitWindow`/`DrawModel`/`InitAudioDevice`/etc.
  — are fully unit tested: `sim`, `camera`, `worldtime`, `world`'s spawn/query logic,
  `input`'s merge logic, `assets`'s catalog, `audio`'s catalog + decision functions
  (`AmbienceWeights`, `ChooseMusicTrack`, `approach`).
- The handful of things that inherently need a live context are *not* directly unit
  tested and are kept as thin as possible: `assets.Store` (models), `audio.Engine`
  (sound/music playback), `world.DrawGround`, `game.FoxDot.Draw`. Everything they depend
  on (what to draw, what to play, when) is decided by pure code elsewhere.
- `engine.Game` encodes this split at the top level: `Update(dt, input.Frame)` must never
  touch rendering; `Draw()` is the only place that does. `game.FoxDot` follows the same
  rule — `Init`/`Update`/`Shutdown` are tested directly in `foxdot_test.go` (Draw is not).

When adding a feature, default to writing the decision/logic as a pure function first,
then a thin adapter that calls into raylib. If you can't figure out how to test something,
that's usually a sign it should be split this way.

### ECS, not OO

The fox, every forest prop (tree/rock/bush/...), and every ambient creature are all just
an `ecs.Entity` (an opaque `uint32`) plus whichever components (`internal/comp`) it has —
there is deliberately no `Fox` struct or `Prop` object with its own methods/identity.
`internal/ecs` is intentionally minimal (no archetypes, no parallel scheduling — a map per
component type is plenty for a scene this size):

- `ecs.Set[T]`/`Get[T]`/`Has[T]`/`Remove[T]` — attach/query one component type.
- `ecs.Each[T]`/`Each2[A,B]`/`Each3[A,B,C]` — iterate entities that have all the given
  component types. `RenderSystem` (`internal/game/systems.go`) is the clearest example:
  it draws *every* entity with `(Transform, ModelRender)` identically, whether it's the
  fox, a pine tree, or a butterfly — no special-casing per "type".
- `ecs.SetResource[T]`/`GetResource[T]`/`MustGetResource[T]` — singleton, non-per-entity
  state: `worldtime.Clock`, `camera.ThirdPerson`, `audio.EventQueue`.

Systems are just ordinary functions taking a `*ecs.World` (see `internal/game/systems.go`):
`FoxSystem`, `CameraSystem`, `CreatureVoiceSystem`, `RenderSystem`. `FoxDot.Update` runs
them in a fixed order each tick. `internal/sim` deliberately stayed a *pure function*
(`sim.Step(*State, Params, dt, input.Frame, cameraYaw, GroundHeightFunc) Events`) rather
than becoming ECS-aware itself — `FoxSystem` is the adapter that copies ECS component
data into a `sim.State`, calls `sim.Step`, and copies the result back. This keeps `sim`
trivially testable and reusable if a second creature ever needs the same movement logic
with its own `MoveParams` component.

**If you add a new kind of entity** (a new creature, a new prop type, a UI element),
prefer composing existing components over adding a new bespoke type/object. If you need
new shared state, add a component to `internal/comp` (if per-entity) or a resource (if
singleton) rather than a field bolted onto `FoxDot`.

### Audio subsystem shape

`internal/audio` mirrors the `assets` package's pure-catalog/impure-store split:

- `catalog.go` (pure) parses `assets/audio/manifest.json` — the source of truth for what
  audio exists, not hardcoded Go data (unlike `assets/catalog.go`, which *does* hardcode
  visual model metadata because it needs bounding boxes the JSON doesn't have).
- `engine.go` (`Engine`, needs `rl.InitAudioDevice`) loads/caches: short catalog entries
  (`sfx`, `ambience_oneshot`) become `rl.Sound` played via `rl.LoadSoundAlias` so
  overlapping plays don't cut each other off; long/looping entries (`music`,
  `ambience_loop`) become `rl.Music` streams, pumped every frame via `Engine.Update`.
- `mixer.go` (`Mixer`) is one generic crossfade primitive reused by both:
  - `ambience.go`'s `AmbiencePlayer`, fed by the *pure* `AmbienceWeights(phase,
    campfireProximity)` — many loops mixed at once.
  - `music.go`'s `MusicDirector`, fed by the *pure* `ChooseMusicTrack(situation, phase)` —
    exactly one nonzero weight, so "play one track" falls out of the same crossfade code
    as "blend N ambience loops" (a music change is just a Mixer target-weight change).
- `events.go`'s `EventQueue` decouples "something happened that should make noise" from
  "how it's played" — gameplay/creature/(future UI) code calls `events.Push(name, volume)`
  by catalog name; `FoxDot.updateAudio` drains it into `Engine.PlaySFX` once per frame.
  **Always route new SFX through this queue**, don't call `Engine.PlaySFX` directly from
  a System — keeps Systems raylib-audio-agnostic and easy to test.

`Engine.Close()` is idempotent (safe to call twice) — this matters because raylib only
supports **one open audio device globally** (like the window), so tests that spin up a
second `FoxDot` (e.g. reload-from-save tests) must `Shutdown()` the first one before
`Init()`-ing the second, or `CloseAudioDevice` hangs. See `internal/game/foxdot_test.go`
for the pattern.

### Day/night cycle

`worldtime.Clock` is a pure resource (`Elapsed`, `Day`, `Settings.DaySeconds` = 360s/6min
by default so a cycle is visible in a short session). `Phase()` buckets into
Night/Dawn/Day/Dusk; `SunHeight()` is a continuous `[-1,1]` value for future lighting.
Persisted in saves (`TimeOfDay`, `Day` fields on `storage.SaveGame`) so reloading resumes
at the same time of day.

## Development practices

- **Test everything that doesn't need a window/GPU/audio device.** Every pure package has
  a `*_test.go`; keep that ratio up. Table-driven tests where it fits; otherwise plain
  `TestXxx` functions with clear failure messages (see existing tests for tone/style).
- **Run before considering anything done:**
  ```bash
  gofmt -l -w .      # or just the packages you touched
  go vet ./...
  go build ./...
  go test ./...
  ```
  `go test ./...` should complete in well under a minute; if something hangs, see
  "Known environment gotchas" below before assuming your code is at fault.
- **Doc comments matter here.** Every package has a package-level doc comment explaining
  its role and, critically, *why* it's structured the way it is (see any file in this
  repo for the tone: terse, but explains trade-offs/conventions, not just "what"). Match
  that style — this codebase is used as its own explanation.
- **No comments that narrate the obvious.** Comments should explain non-obvious intent,
  constraints, or trade-offs — not restate what the next line does.
- **Config, not constants, for anything a player/dev might want to tune.** New
  gameplay/graphics/audio knobs belong on `config.Config` (see `AudioConfig` for the most
  recent example), not hardcoded, unless they're truly internal implementation detail
  (e.g. `campfireProximityRadius`).
- **Keep the pure/impure split intact when extending existing packages.** If you're
  tempted to import `rl` (beyond its plain math types/constants) into `sim`, `worldtime`,
  a `*.catalog.go`, or a decision function like `AmbienceWeights`, stop — that logic
  belongs in a `Store`/`Engine`-style adapter instead.
- **gofmt handles struct/map alignment** — if you hand-format a struct literal or field
  list, just run gofmt rather than manually aligning columns.

### Known environment gotchas (this sandbox specifically)

- **The sandboxed display service cannot reliably open a GLFW window.** Attempts to
  actually run `go run ./cmd/foxdot` here have failed with GLFW/display-server errors
  (this has been consistent across sessions, including before any of this repo's own
  code existed — a trivial raylib smoke test hit the same issue). This means: **rely on
  unit tests, not manual visual verification, for anything gated on `Draw()`.** When you
  can't visually verify something (model orientation, layout, lighting), reason about it
  analytically and say so explicitly rather than claiming you saw it work. Ask the user to
  eyeball it on their machine when that's the only way to confirm.
- **`rl.InitAudioDevice`/`CloseAudioDevice` are a global singleton**, same as the window.
  Only one `audio.Engine` can be open at a time process-wide. Tests that need two
  `game.FoxDot` instances (reload/round-trip tests) must fully `Shutdown()` the first
  before constructing the second. `Engine.Close()` is idempotent specifically so
  `t.Cleanup` + an early manual `Shutdown()` don't conflict.
- **cgo mode, not purego.** `CGO_ENABLED=1` (the default) builds raylib's vendored C
  sources and works reliably here. The purego build tag failed in this sandbox needing a
  system `libffi` that wasn't present — don't switch build modes without re-verifying.
- Audio device init in this sandbox uses miniaudio's **Null backend** (no real hardware) —
  loading/playing/closing all work fine and are fast; this is why `internal/game`'s tests
  can safely exercise the real `audio.Engine` rather than mocking it.

## Future plans / what's next

Roughly in order of how likely they are to come up next, per the README's "What's next":

1. **Terrain height** — currently a flat plane (`world.DefaultGroundY = 0`); `HeightAt`
   is already the right seam to extend (add a heightmap/noise sample instead of a
   constant, still only affecting entities outside prop footprints).
2. **Animation** — models are static meshes. `comp.ActionState`/`comp.Velocity` already
   carry what an animation system would need to key off of (nibble/swipe/moving/grounded);
   see `assets/README.md` for rigging notes if real skeletal animation gets added.
3. **Real wildlife AI** — ambient creatures (`world.SpawnAmbientCreatures`) currently
   stand still and periodically make noise via `comp.CreatureVoice`. Wandering behavior
   would be a new System reading/writing `Transform`+`Velocity`, same shape as `FoxSystem`.
4. **A real pause menu / HUD** — `engine.App` already routes the Pause input action
   (currently just quits). `internal/audio` already has UI SFX loaded and ready
   (`ui_click`/`ui_confirm`/`ui_back`/`ui_page`) via the same `EventQueue` the quicksave
   confirmation sound uses today — wiring a menu just means pushing to that queue.
5. **Lighting/skybox driven by time of day** — `worldtime.Clock.SunHeight()` exists
   specifically to feed this later (continuous, not steppy at phase boundaries).
6. Multiple save slots (storage layer already supports `ListGameSlots` etc., just not
   exposed in `game.FoxDot` beyond the single `"autosave"` slot).
7. Procedurally-generated (rather than hand-placed + seeded-scatter) terrain/forest
   layout — `storage.SaveGenerated`/`LoadGenerated` exist as a caching seam for whenever
   generation gets expensive enough to be worth persisting.

## Where to look for more detail

- `README.md` — user-facing: quick start, controls, requirements, full architecture tour.
- `assets/README.md` — the two bundled asset packs (visual + audio): contents, scale/
  orientation conventions, format notes, licensing.
- Package-level doc comments (top of each `.go` file with one) — each explains that
  package's specific rationale in more depth than this file does.
