# FoxDot Asset Packs

Two bundled asset packs live here: a low-poly **visual** pack (models +
textures) and a **audio** pack (music, ambience loops, sound effects). They
have different licenses (see `LICENSE.txt`) and are loaded by two separate Go
packages: `internal/assets` (visual) and `internal/audio` (audio).

```
assets/
  LICENSE.txt       combined license text for both packs
  MANIFEST.txt       flat file listing of everything in this folder
  models/, textures/  visual pack
  audio/              audio pack (music/, ambience/, sfx/, manifest.json)
```

## Visual pack — low-poly forest

Every model is an OBJ with an accompanying MTL file. Materials reference tiny
8x8 PNG textures in the shared `textures/` directory.

### Contents

- **Animals**: fox, rabbit, deer, squirrel, bird, butterfly, beetle
- **Trees**: pine, fir, oak, birch
- **Plants**: plain/berry/white-flower/yellow-flower bush, three grass tufts,
  red and yellow mushroom
- **Props**: three rocks, fallen log, crate, fence section, campfire

### Scale and orientation

- Y is up.
- Models sit at or close to Y=0.
- One unit is approximately one meter.
- Trees are approximately 3 meters tall.
- Animals are intentionally stylized rather than anatomically exact.
- Animal models face **local +X** (nose/front along +X, tail/back along -X),
  confirmed by rendering `fox.obj` from above. `internal/assets/catalog.go`
  records a per-model yaw correction so the engine (which treats +Z as
  forward at yaw 0) draws them facing the right way automatically.

### Raylib usage

```c
Model fox = LoadModel("assets/models/animals/fox.obj");
DrawModel(fox, (Vector3){0, 0, 0}, 1.0f, WHITE);
```

Keep the folder structure intact so each MTL file can find the shared
textures (each `.mtl` references them as `../../textures/<name>.png`).

This repo loads the pack through `internal/assets` (a small Go catalog +
loader built on top of `raylib-go`) rather than the C snippet above — see
`internal/assets/catalog.go` for the full list of models and
`internal/assets/store.go` for the loader. It also records each model's
authored size and "forward" axis so gameplay code doesn't have to guess.

### Notes

- OBJ is broadly supported by Raylib and easy to inspect or edit in Blender.
- Textures are intentionally tiny and reusable; the low-poly geometry carries
  most of the visual style.
- These models are static and do not include skeletal animation.
- For a moving fox or deer, start with simple whole-part animation or import
  the OBJ into Blender and rig it.

## Audio pack — cozy, slightly haunted forest

Original procedural music, ambience, and sound effects. The music is droning
and restrained, with light snappy percussion and a shared eight-note theme
(D Dorian) that moves between anticipation and subtle melancholy across
tracks — see `audio/music/THEME.md` for the theme and per-track treatment.

### Contents

- **Music** (`audio/music/wav`, editable source in `audio/music/midi`):
  main menu, forest day, forest night, discovery cue, quiet danger loop,
  credits.
- **Ambience loops** (`audio/ambience/loops`): wind, rustling leaves, birds,
  crickets, river, campfire, rain, frogs.
- **Ambience one-shot** (`audio/ambience/oneshots`): thunder.
- **Sound effects** (`audio/sfx`):
  - `footsteps/{grass,dirt,stone}` — 4 variants each
  - `player` — jump, roll, item pickup, tool swing, bow release, inventory
    open/close
  - `ui` — click, confirm, back, page
  - `animals` — fox bark, rabbit hop, deer call, 3 bird chirps, butterfly
    flutter, beetle buzz

### Format

- WAV: 16-bit PCM, 22.05 kHz. Music is stereo; ambience and SFX are mono.
- MIDI files are standard MIDI, provided as editable composition data (not
  loaded by the game).
- Loops are crossfaded at their boundaries in the source files; use
  `Music` streams (not one-shot `Sound`s) for anything long or looping so
  raylib mixes them cleanly — see `internal/audio/engine.go`.

### `audio/manifest.json`

Structured metadata (kind, duration, loop flag, BPM, musical key) for every
audio asset. This is a *second*, code-facing manifest distinct from the
human-readable `MANIFEST.txt` at the repo root of this folder —
`internal/audio/catalog.go` parses it at startup instead of hardcoding the
list in Go, so new tracks/SFX just need an entry here plus the file itself.

### Go usage

`internal/audio` loads this pack:

- `catalog.go` (pure, no raylib) parses `manifest.json` into a `Catalog`.
- `engine.go` (raylib) lazily loads/caches `rl.Sound` for SFX and one-shots,
  and `rl.Music` streams for anything marked as a loop (music + ambience).
- `ambience.go` mixes multiple ambience loops by crossfading their volumes.
- `music.go` picks and crossfades music tracks by time of day / situation.
- `events.go` defines a simple queue so gameplay/UI code can request a sound
  by name without depending on the audio engine directly.

## License

See `LICENSE.txt`. The visual pack is CC0/public domain. The audio pack has
different terms (generated output owned by the user under the applicable
OpenAI terms) — both packs are usable in personal and commercial projects,
but read `LICENSE.txt` before redistributing the audio pack on its own.
