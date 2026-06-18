# Per-world scene code

Each world's scene block (decoded at `$1B980`, see Part IV) carries its own
**per-world game code**, in two parts:

* **Scene handlers** — the `+$18` field of each scene descriptor (Part III §6). Two
  or three per world. These run the **animated parallax background** and trigger
  **ambient sounds** (they `JSR` only the resident sound API `$1A2AC`/`$1A2DC`/…,
  never the spawn routines). World 1's, for example, drives the waterfall blit
  (`$1D2DA`); they differ world to world only in which backdrop they animate.

* **Enemy AI handlers** — the `+$20` field points at a table of object update
  routines (6–18 per world). Each is a complete enemy behaviour on an `a5`-relative
  object node (the object-system struct, Part V §3, extended with AI fields):

  ```
  +$12 frame table (one of the world's sprites)   +$1A y       +$1C active
  +$1D damage taken   +$1E init/state   +$26 health   +$2E anim timer   +$30 …
  ```

  A handler typically: on first run sets its `+$12` frame table and `+$26` health,
  then each frame animates (`+$C` frame, `+$2E` timer), applies damage (`+$1D` →
  `+$26`) and, on death, `JSR $130` (spawn the burst). So the per-world code is the
  **enemy roster + behaviours**, each wired to one of the sprites in
  `rendered/sprites/world<N>_sprite_*.png` — not new mechanics; the worlds share the
  engine and differ in their enemies and backdrops.

The worlds are similar in shape (no world has a fundamentally different mechanic in
its scene code): code sizes are ~3.5–8.8 KB, all using the same object/AI and
sound interfaces.

## Regenerate

```sh
go run turrican/extract/cmd/decrunch -o /tmp/turrican.bin "Turrican (Amiga)/Turrican.adf"
# decode a world's scene block (offsets from the $46A level table; world 0 shown):
go run turrican/extract/cmd/block -off 0x3F000 -len 0x1EAAC -base 0x1B980 -o /tmp/w0.bin "Turrican (Amiga)/Turrican.adf"
# trace its scene + AI handlers (handlers = descriptor +$18 and the +$20 table):
go run stupidcoder.com/tools/cmd/codetrace68k -base 0x1B980 \
  -entry 0x1D0AC,0x1D15E,0x1D28E,<the +$20 AI handlers> -o /tmp/world0.asm /tmp/w0.bin
```

Scene handlers per world: world 0 `$1D0AC/$1D15E/$1D28E`, world 1 `$1C894/$1C7F2`,
world 2 `$1D86E/$1D8EC/$1D992`, world 3 `$1D084/$1D0FC/$1D18E`, world 4
`$1C800/$1C876`. The full dumps aren't committed (each is ~1.4 MB, mostly the
block's graphics/map data already covered in Part IV).
