# Luigi's Mansion — plan: the fully assembled mansion

**Long-term goal:** one Studio scene of the whole mansion — all 74 rooms assembled in their
true positions, full materials, furniture in place with its animations playing — plus the
already-shipped cutscene library around it. This document is the working plan; strike items
as they land and record findings in `luigis-mansion-gc.md`.

The phases are ordered so that every one ends with something visibly better in the Studio,
and so that verification (against the game's own renderer, via the oracle) comes before
volume (74 rooms of everything).

## Phase 1 — a correct single room

The foundation everything else stands on. A room must look *right* before 74 of them do.

1. **Material/texture verification pass** — one room (the foyer), compared surface-by-surface
   against the game's render at the same spot (`foyer.state` exists; framedbg or `-shot`).
   The sampler-stride bug the user caught (12 → 20 bytes) is fixed; the remaining known
   gaps, in order:
   - **Untextured white pieces** (staircase banister ribbon in the foyer): find what those
     materials reference — suspects are the palette path (sampler `paletteIdx` ≠ −1) and
     multi-stage materials (`s16` sampler list at material+8, one per stage — only stage 0
     is exported today).
   - **CI-format textures** (`C4/C8/C14X2`): the decoder refuses them; wire the palette
     section through `decodeGXTexture` (the sampler's `paletteIdx` names the palette).
   - **Wrap modes**: honour the sampler's `wrapS/wrapT` in the GLB instead of defaulting
     to REPEAT.
   - **Vertex colours** (slot 5) and **second UV set** (slot 7): parse, and export COLOR_0
     where present — the rooms' baked lighting almost certainly lives there, and it is what
     will make rooms look like the game instead of full-bright.
2. **The two-index vertex question**: 2-index meshes are exported as (pos, normal) — verify
   against the renderer's vertex-descriptor setup (`0x8001D4C0` sets attrs 9/10/11/13–16 from
   per-mesh state; confirm which the `0x002` mask actually enables — it may be (pos, uv)).
3. **Acceptance**: a side-by-side of the exported foyer vs `foyer.png` where every visible
   surface carries the right texture; differences only in lighting.

## Phase 2 — furniture, complete with animation

1. **Sweep all furniture bins** in all room arcs (they repeat across rooms — dedupe by
   content hash, keep the game's own names: oisu the chair, otukue the tea table, …).
   Fix what the sweep turns up (the odd attribute mask, the `[13]/[14]` material-stage
   sections richer furniture uses — decode them from `0x8001DBF8`'s stage path).
2. **The `.anm` format** — furniture animation (drawers, lids, curtains; `anm/ota1_0.anm`).
   Same method as always: find the evaluator via rwatch/callers; expect another wearing of
   the hermite-channel mechanism. Export as glTF animations on the furniture GLBs.
3. **The `.bas` files** (168 B, one per animated prop in the demos too) — likely the actor's
   base/placement record; small, decode opportunistically.
4. **Acceptance**: a furniture gallery section in the Studio, animated where the game
   animates (a chest opening, the toilet lid).

## Phase 3 — room placement: assembling the mansion

1. **Find the placement data.** Rooms are modelled in mansion-global coordinates already
   (the foyer's graph TRS is identity — check whether this holds for all rooms; if so,
   assembly is mere concatenation). If not: the candidates are `Map/map2.szp` (982 KB, the
   mansion "map" the game loads with the rooms), `gidemap.szp`, and the engine's room table
   in the DOL. The GameBoy Horror's map view knows every room's bounds — that data exists
   somewhere.
2. **Export the assembled shell**: all 74 rooms into one GLB (or one GLB per room plus a
   placement JSON the viewer composes — better for streaming and per-room toggles).
3. **Doors**: the door vignettes' 8 door sets (already exported) are the connectors between
   rooms; place them from the same map data if it names them.
4. **Acceptance**: the mansion cutaway in the Studio — orbit the whole building, rooms in
   their true positions.

## Phase 4 — the full sweep and the Studio centrepiece

1. **All 74 rooms + furniture placed inside them** (room arcs contain their furniture;
   placement of furniture within a room needs the same investigation as Phase 3 — the
   `.bas` records and the room's `keyper`-style tables are the suspects).
2. **Viewer work**: an `lm-mansion` renderer — per-room visibility toggles (the game's own
   trick: rooms light up as you enter), a floor slider, and the existing camera-track
   machinery for a fly-through.
3. **Performance**: 74 rooms of GLB is tens of MB — lazy-load per room, share repeated
   furniture GLBs by hash, and consider KTX2/basis if it gets heavy (weigh against the
   site's no-build convention).
4. **Stretch**: room lighting. The game's look is flashlight + darkness; vertex colours
   (Phase 1) plus a dark-ambient viewer preset would land the mood without a lighting
   engine.

## Standing items

- Update `luigis-mansion-gc.md` per phase (Part IX grows; new parts for `.anm`/placement).
- Verify against the oracle at each step — the sampler-stride bug shows eyeballing an
  export is not acceptance; compare against the game's own frames.
- Commit per milestone with pathspec-only commits; push (user follows remotely).
- Out of scope for now (separate tracks): the demo `.scd` actor binding, `.txp` texture
  patterns, gameplay/audio formats, the `Kawano` menu resources, `Gadgets`/ghost models.
