# M5 — DOS Code workspace and Underworld renderer tour

The shared Code workspace now supports the DOS core. It displays live x86
instructions, registers, mode, segment selectors and their effective bases.
Ultima Underworld has a single knowledge package containing required file
identities, verified runtime modules, renderer functions, three self-modifying
operand watches, a chunky-buffer view and a four-stop tour.

## Using it

Select your installed Underworld folder and `UW.EXE`, boot normally, enter the
dungeon, then pause and open Code. The tour becomes available after the required
file hashes match; Start additionally requires the resident rasterizer and
geometry overlay to pass their live signatures and relocation checks.

**From projection to texture spans and VGA** visits:

1. Perspective projection: the original multiply/divide code produces screen
   coordinates from view-space coordinates and depth.
2. A texture span: `MOVSB` samples the texture; the following additions use live
   gradient immediates which the game patches per span.
3. The chunky-to-planar copy entry: compare the off-screen buffer with the game
   output before presentation.
4. The end of the copy loop: all selected rows/planes have been copied. Actual
   VGA write counts show the work even when the colors did not change.

Continue, Cancel, Explore freely and Restore stop retain their M4 semantics.
Selecting a function, inspecting memory or changing panel layout never advances
execution. A function's module must remain valid to run to its named entry.
Raw linear-address stops remain available for unknown games, with no inferred
function identity. The off-screen view appears only at the documented span/copy
contexts, with the expected live source segment; its pixels, palette, state
watches and CPU registers carry the same instruction clock.

The lesson follows successive pipeline stages, not one isolated polygon all the
way through a complete frame. Dungeon entry is manual. No commercial game bytes,
checkpoint or boot automation is distributed. Existing DOS/VGA limitations,
including the known Underworld HUD corruption, remain visible.

## CPU adapter

- `real16`: 16-bit instruction/address defaults, with the modeled 386 prefixes.
  Instruction fetch and near targets obey 16-bit offset and 20-bit bus wrapping.
- `protected32`: the existing flat go32/DPMI model, including cached segment
  bases. Selectors are **not** multiplied by 16. This does not claim general
  protected-16, paging, privilege transitions, or arbitrary descriptor modes.
- Live disassembly reuses the repository's Go x86 decoder through the DOS C++
  generator. A bounded stack decoder avoids allocating permanent arena objects
  during observation. No game-specific decoder or external disassembler is used.
- Safe peeks read backing storage without calling VGA read handlers or changing
  latches. Device windows are unavailable. Disassembly bytes are fetched again
  on every snapshot, including self-modifying immediates. Follow-PC rows preserve
  segmented fetch wrapping; near targets are shown in linear coordinates. A raw
  address inspection uses the current CPU mode; it cannot infer another CS mode.
- Single-step retires at most one architectural instruction. IRQ delivery is
  returned separately, before the first handler instruction. A consumed device
  hook is preserved across step, ordinary run, checkpoint save and restore,
  preventing duplicate IRQ/timer/input processing.
- REP is one instruction. Debug execution rejects 32-bit string repeat counts
  above 65,536 and excessive prefix chains instead of allowing an unbounded
  slice. Cooperative jobs use at most 1,000 instructions per slice and retain
  wall/instruction budgets. Counters in DOS are instruction steps, not CPU cycles.

The core state format advances from DOS version 1 to 2 to store the pending
instruction-boundary hook. The native importer still accepts version 1. Browser
save files retain the existing exact-core identity rule: this is not a migration
mechanism for a save container pinned to an older WASM build.

Ordinary playback keeps its direct CPU loop; the debug helper is called only to
finish a pending debug boundary. With identical `clang++ -O3` settings, three
alternating 300-interval runs against M4 had median times of 0.931 s before and
0.955 s after (about 2.6% in this local sample). CPU/RAM/display proofs matched in
all six runs. The added write-observation branch remains inactive outside
recording; these measurements are not a cross-machine performance guarantee.

## Knowledge profile additions

`games/ultima-underworld-pc/knowledge.json` is the authored source. Go schema and
semantic checks reject unknown operations and invalid references/bounds.

### Identity

A DOS file-set release declares `executable`, which must be a pinned member and
must match the selected entry program. The Underworld profile pins `UW.EXE`,
`DATA/LEV.ARK`, `DATA/W64.TR`, `DATA/F32.TR` and `DATA/PALS.DAT`; unrelated extras
are allowed but are not covered by that identity. Paths follow the existing
ASCII-insensitive file-set policy; ambiguous canonical paths fail closed. Hashes
are computed locally. Merely finding `UW.EXE` beside a different selected program
does not activate its knowledge.

### Modules and locations

A `modules` entry has a label, `mode: real16`, size, immutable byte signatures and
relocation words. The implemented kinds are:

- `mz`: `paragraph` is relative to the actual DOS load segment. Signatures and
  each `relativeSegment + loadSegment` relocation word must match live RAM.
- `overlay`: scan paragraph-aligned conventional RAM for all declared signatures
  and relocation words. Exactly one matching base is required. Zero or multiple
  matches mean unavailable/ambiguous, not a guessed fixed slot.

Module signatures are limited to eight ranges of 8–128 bytes, module size to
64 KiB, relocations to 64, and modules to 16. Runtime scans stop below VGA at
`0xA0000`. The Underworld rasterizer has two signatures and a verified segment
relocation; the geometry overlay has two signatures and its projection-output
segment relocation. Mutable texture-step operands are deliberately outside the
identity signatures. Unlisted bytes are not asserted to be immutable.

Locations support `{kind: module, module, offset}` and
`{kind: relocated-segment, module, relocation, offset}`. The latter derives a
real-mode data segment from an already verified relocation word; Underworld's
chunky buffer therefore follows the load context rather than a guessed absolute
alias. Existing relative locations still work. No generic pointer chasing,
filesystem/container resolution or protected-mode module loader is introduced.

Snapshots resolve modules afresh. Named breakpoint jobs bind the segment and
base and recheck identity between slices. Tours pin their resolved module bases
and load context and fail when those checks change. This is bounded signature
validation, not a continuous trace proving every overlay change between samples.

### Tours, watches and buffers

DOS tour predicates add symbolic `location` targets, mode tests and register
comparisons. Symbolic code targets require both the resolved linear PC and CS.
Guards/capture ranges may reference locations; resolved capture overlap is
rejected. Runtime verification preserves the existing M4 bounds and lifecycle.
DOS checkpoints have a 128 MiB cap, matching the adapter's state limit.

The three operand watches use ordinary typed `u16` values at module offsets
`02EC`, `02F0`, `02F3`. They show raw installed immediates, not invented game state.
A buffer declares its location, width, height, stride, VGA palette and an explicit
`when` condition. Bounds are at most 320×200, 64 KiB per buffer and four buffers.
Underworld's 320×200 buffer is enabled at the span and copy stops only when the
live ES/DS source segment agrees with the documented profile. Pixels and palette
are copied synchronously in the same snapshot as CPU and typed state.

DOS activity recording now observes modeled RAM, VGA-plane and palette stores.
Read/fetch tracing is **not** implemented. The Memory region description says
write activity only. The tour's selected-RAM differences and writes remain
separate from additional storage writes. Region totals preserve same-value
VGA writes; the renderer's display can remain byte-identical after 19,264 copies.
Recorder overflow still marks evidence incomplete and aborts the tour.

## Validation

Public checks include Go knowledge/x86 tests, synthetic module relocation,
overlay movement/ambiguity/invalidation, executable identity, shared tours/state
and debugger regression. `debug-dos.cpp` runs both natively and in actual WASM:
512 synthetic vectors compare decoder text/length against Go in both modes,
alongside safe peeks, live operands, real-mode wrapping, selector-base semantics,
IRQ stops, next-visit policy, state restore, and oversized REP bounds. Existing
x86, DOS render-buffer and Xbox WASM tests also pass.

```sh
go test ./tools/knowledge ./tools/cpu/x86
node tools/browser/tests/dos-knowledge.mjs
node tools/browser/tests/tours.mjs
node tools/browser/tests/debug.mjs
python3 tools/browser/tests/check-x86-wasm.py --emcc /path/to/em++
python3 tools/browser/check-release.py
```

To reproduce private-media acceptance, supply an installed reference game and a
native gameplay checkpoint reached by normal execution:

```sh
node tools/browser/tests/tour-underworld.mjs \
  games/ultima-underworld-pc/game/UW.EXE \
  tools/platform/dos/browser/work/uw-input.state \
  tools/platform/dos/browser/work/m5-tour.rrstate
python3 -m http.server 8779 --bind 127.0.0.1
# Open /tools/browser/tests/code-m5-browser.html
```

The helper verifies required hashes, normalizes host input like browser restore,
creates an ignored local test container/file list, then asserts the four stops,
atomic snapshots, actual VGA writes, full-machine restoration and signature
rejection without execution. The browser test covers directory paths, exact
entry identity, Code navigation, decoded operands, shared Memory navigation,
instruction stepping, restoration, the four-stop tour and 375px buffer layout.
Fort's M4 browser tour remains a regression check. Acceptance results and normal
execution measurements are stored beside this report; no private states or
images are checked in.
