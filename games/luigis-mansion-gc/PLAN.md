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

## Phase 2 — furniture, complete with animation ✅ core (2026-07-27, cc0514a9 / 3cbdf75b)

1. ~~**Sweep all furniture bins**~~ — `lmtool -mansion` sweeps all 75 arcs: 451 unique
   furniture GLBs (content-hash dedupe, game names kept, numbered variants on collision),
   all parsing cleanly under the exact attribute-mask decode. **Bonus: the placement
   database fell out of this phase** — `Map/map2.szp jmp/furnitureinfo` (JMap format,
   `lm/jmp.go`) places every piece per room; `placements.json` + the Studio's `lm-room`
   renderer assemble rooms as shell + instanced furniture (the SM64DS model).
2. ~~**The `.anm` format**~~ — cracked and verified (binder `0x8001DE90`, evaluator
   `0x8001E04C`): 9 hermite channels per graph node over three pools, TRS-replacing;
   the rest clip reproduces the `.bin` graph bit-exact. Furniture GLBs carry one glTF
   animation per clip; the gallery click-cycles them (chest drawer confirmed visually).
3. **The `.bas` files** — still open (interaction metadata; decode opportunistically).
4. ~~**Acceptance**~~: the Studio gallery (tea table, chair, toilet, chest, mirror,
   chandelier, tool cabinet) with animations playing where the game has them.

Remaining Phase 2 polish: the `[13]/[14]` material-stage sections + TEX1 blending used
by the six two-stage rooms (63/65/66/67, b1_c_67, gyara_00).

## Phase 3 — room placement: assembling the mansion ✅ core (2026-07-27)

1. ~~**Find the placement data.**~~ Verified: all 75 room shells are mansion-global
   (`extract/cmd/roombounds` — floors tile at y −600/−50/500/1050; the building-sized
   shells are the exterior walls/grounds `room_11/16/23/37/72` and roof `room_59/60`).
   Assembly is mere concatenation, as hoped.
2. ~~**Export the assembled shell**~~: one GLB per room + `placements.json` (already the
   `lmtool -mansion` output); ALL rooms + furniture now ship in the Studio, streamed by
   the new `lm-mansion` renderer with a shared furniture cache and an exterior layer
   (on = the mansion on its grounds, off = the dollhouse cutaway).
3. ~~**Doors**~~: FOUND — the DOL's map-info header (0x8030377C) points at a 72-record
   door list: position, axis, size, type (→ door model via the 20-byte table at
   0x802FF95C), and the two rooms each opening connects. All 54 leafed doors now hang
   in the assembled mansion, swing open on click (held at the apex — the game's clips
   are full open-and-shut cycles) and shut on the next.
4. ~~**Acceptance**~~: the mansion orbits/flies as one building, rooms in true position;
   the foyer's archway opens into the parlor behind it.

## Phase 4 — the full sweep and the Studio centrepiece (doors + floors done 2026-07-27)

1. ~~**All 74 rooms + furniture placed inside them**~~ — shipped with Phase 3 (the
   furniture placement came straight from `jmp/furnitureinfo`, no `.bas` needed).
2. **Viewer work**: ~~`lm-mansion` renderer~~ ✅; ~~floor slider~~ ✅ (storey layers:
   basement/ground/first/attic peel the dollhouse, exterior+roof separate);
   ~~doors~~ ✅ (placed from the DOL list, click to swing). Still open: a camera-track
   fly-through, per-room highlight on entry.
3. **Performance**: shells stream 6-at-a-time with a shared furniture cache (~58 MB
   total, in line with the site's bigger games); KTX2 only if it ever hurts.
4. **Stretch**: the dark-ambient flashlight preset; the `[13]/[14]` material-stage
   sections + TEX1 blending for the six two-stage rooms; `.bas` interaction records.

## Standing items

- Update `luigis-mansion-gc.md` per phase (Part IX grows; new parts for `.anm`/placement).
- Verify against the oracle at each step — the sampler-stride bug shows eyeballing an
  export is not acceptance; compare against the game's own frames.
- Commit per milestone with pathspec-only commits; push (user follows remotely).
- Out of scope for now (separate tracks): the demo `.scd` actor binding, `.txp` texture
  patterns, gameplay/audio formats, the `Kawano` menu resources, `Gadgets`/ghost models.
