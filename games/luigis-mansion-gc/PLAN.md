# Luigi's Mansion — plan: the fully assembled mansion

**Long-term goal:** one Studio scene of the whole mansion — all 74 rooms assembled in their
true positions, full materials, furniture in place with its animations playing — plus the
already-shipped cutscene library around it. This document is the working plan; strike items
as they land and record findings in `luigis-mansion-gc.md`.

The phases are ordered so that every one ends with something visibly better in the Studio,
and so that verification (against the game's own renderer, via the oracle) comes before
volume (74 rooms of everything).

## Phase 1 — a correct single room ✅ (done 2026-07-26, commits c19b912f / 9ec4beac / 24e8062d)

The foundation everything else stands on. A room must look *right* before 74 of them do.

1. ~~**Material/texture verification pass**~~ — all decode bugs found and fixed: sampler
   stride 12→20; texture data offsets are texture-SECTION-relative (the "white banister
   ribbon" was the misdecoded stair carpet); graph +0x08 is flags, not a second pair count
   (phantom window frames); **wrap modes** incl. GX mirror wired into the GLB (the floor
   medallion is one quadrant mirrored both ways).
   - **Untextured white pieces**: resolved — the remaining white material (25, the banister
     cloth) is genuinely stage-less in the game (white fabric under GX lighting); the
     exporter now leaves untextured materials lit so the viewer shades them.
   - **CI-format textures**: closed — the disc survey (`extract/cmd/binsurvey`) shows **no
     CI textures in any of the 572 .bin files**; no palette path needed.
   - **Vertex colours**: closed — **no .bin on the disc has colour sections**; there is no
     baked lighting to export. The game's look is GX lighting (Phase 4's dark-ambient
     preset remains the way to land the mood).
   - **Second UV set** (section 7 = TEX1): parsed; only the six two-stage rooms
     (63/65/66/67, b1_c_67, gyara_00) use it. TEX1/stage-1 blending in the export remains
     a Phase 2 nicety.
2. ~~**The two-index vertex question**~~: answered exactly — the true attribute mask is the
   **u32 at mesh+4**, literal `1<<GXAttr` bits read by `0x8001D5B8` (0x200 POS, 0x400 NRM,
   0x2000 TEX0, 0x4000 TEX1; byte +11 = NBT3 with three normal indices). 183 meshes
   disc-wide are (pos, uv) — the old u16@+8 heuristic had them as (pos, normal). The DL
   parser now decodes exactly, no width fallback.
3. ~~**Acceptance**~~: the foyer export rendered from the game's own camera (medallion →
   rosette arch) matches the `foyer.state` game frame surface for surface; the differences
   are the game's flashlight lighting, the door leaf (separate asset), and furniture
   (separate bins, Phase 2).

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
