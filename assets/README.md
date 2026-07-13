# Low-Poly Forest Starter Pack for Raylib

A small, procedural low-poly asset pack designed to be dropped into a Raylib
project. Every model is an OBJ with an accompanying MTL file. Materials reference
tiny 8x8 PNG textures in the shared `textures/` directory.

## Contents

### Animals
- Fox
- Rabbit
- Deer
- Squirrel
- Bird
- Butterfly
- Beetle

### Trees
- Pine
- Fir
- Oak
- Birch

### Plants
- Plain bush
- Berry bush
- White-flower bush
- Yellow-flower bush
- Three grass tufts
- Red mushroom
- Yellow mushroom

### Props
- Three rocks
- Fallen log
- Crate
- Fence section
- Campfire

## Scale and orientation

- Y is up.
- Models sit at or close to Y=0.
- One unit is approximately one meter.
- Trees are approximately 3 meters tall.
- Animals are intentionally stylized rather than anatomically exact.
- Animal models face **local +X** (nose/front along +X, tail/back along -X),
  confirmed by rendering `fox.obj` from above. `internal/assets/catalog.go`
  records a per-model yaw correction so the engine (which treats +Z as
  forward at yaw 0) draws them facing the right way automatically.

## Raylib usage

```c
Model fox = LoadModel("assets/models/animals/fox.obj");
DrawModel(fox, (Vector3){0, 0, 0}, 1.0f, WHITE);
```

Keep the folder structure intact so each MTL file can find the shared textures
(each `.mtl` references them as `../../textures/<name>.png`).

This repo loads the pack through `internal/assets` (a small Go catalog + loader
built on top of `raylib-go`) rather than the C snippet above — see
`internal/assets/catalog.go` for the full list of models and
`internal/assets/store.go` for the loader. It also records each model's
authored size and "forward" axis so gameplay code doesn't have to guess.

## Notes

- OBJ is broadly supported by Raylib and easy to inspect or edit in Blender.
- Textures are intentionally tiny and reusable; the low-poly geometry carries
  most of the visual style.
- These models are static and do not include skeletal animation.
- For a moving fox or deer, start with simple whole-part animation or import the
  OBJ into Blender and rig it.

## License

CC0 / Public Domain. Use, modify, redistribute, and sell projects containing
these assets without attribution.
