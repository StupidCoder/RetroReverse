# U6 — captured pixel inspection

Implemented September 28, 2026. All four shared emulator pages now expose a
compact pixel inspector: coordinate/color, ordered contributors, a selected
event, and available historical source evidence. Numeric coordinates supplement
canvas clicks. CSS contain scaling and letterboxing are accounted for. Controls,
profiling and compatibility notes are collapsed beside the inspector; contributor
and detail panes scroll independently. The page still scrolls at small viewport
heights; the later GPU/memory workspace design is not part of this milestone.

C64 exposes VIC fetch roles, addresses, captured values, palette decisions and
historical writers with instruction bytes. Nearby CPU reads are explicitly not
claimed as proven data dependencies. PS1 exposes primitives, fills, copies and
CPU/DMA uploads, texture/CLUT bytes and historical source queries. N64 exposes
RDP events, modeled depth/alpha rejects, command state and immutable TMEM/TLUT
bytes. 3DO exposes CCB/cel events, CPU/HLE writes, PIXC, transparency rejects,
PDAT snapshots and captured PLUT values. Interleaved framebuffer sources can be
followed back to their state at the actual read.

Each source query uses a write cursor and the captured source value. A later
write cannot become the supposed origin of an earlier read. Requests carry
session/capture/request identities; selecting another event, running, resetting,
or loading invalidates old replies. Cancel query discards the reply; an already
running bounded core query completes in the worker. Capture itself is cancellable
between execution slices. Detailed tracing is disabled during ordinary running.
Capture timing and subsystem costs are shown separately from normal execution.

## Validation

- Native pixel tests cover C64 character/sprite fetch and priority history;
  PS1 fill, upload, copy, texture/CLUT, transparent rejection and later source
  overwrite; N64 depth/alpha rejection and CI/TLUT lookup; 3DO PIXC blending,
  interleaved framebuffer layout and an actual offscreen cel render followed
  by source overwrite. The shared recorder additionally tests overflow,
  unobserved-change detection, initial contents and immutable resource deduplication.
- WASM tests compare the entire serialized end state with tracing enabled and
  disabled. They reconstruct stored bytes and compare modeled display colors at
  85 coordinates over three intervals per platform. All sampled pixels pass,
  with no trace overflow in these runs. Results are in `results/u6-*.jsonl`.
  The C64 scene was booted from the real Fort tape into its game; PS1 uses a late
  Ridge Racer checkpoint, N64 a Pilotwings checkpoint, and 3DO a Need for Speed
  driving checkpoint generated with the recorded input schedule.
- Source-history correctness is tested explicitly in synthetic PS1 and 3DO
  cases. The reference-game sampling grids reported zero matching source-history
  queries; they must not be cited as independent validation of that navigation.
- Browser checks exercised Need for Speed driving, canvas/numeric selection,
  ordered writes and transparent candidates, and immutable source-byte lookup.
  A driving capture took about 0.26 s with a 100.5 ms longest call and 67.9 MiB
  evidence/checkpoints on this Mac. These observations are not worst-case limits.
  Save remained enabled after capture. Normal execution statistics excluded the
  measured capture overhead. Fort state import used hosted firmware; its browser capture completed in about
  20 ms and a selected graphics fetch matched the displayed VIC palette color.
- Coordinate/letterbox and per-platform scanout color unit checks pass. Source
  instrumentation regeneration is idempotent. Frontend syntax checks pass.

## Explicit limits

This explains the current emulator models. PS1 semi-transparency blending and
mask-bit effects remain unimplemented in that core; the UI says so. N64 coverage
and VI scanout are simplified; the recorded TMEM location is the last filter tap,
not a complete multi-tap derivation. Submission PCs identify dispatch, not the
instruction that built command data. Coded/packed 3DO data retains source bytes
and the sampled value, not a full compressed-bit-to-instruction dependency graph.

Initial framebuffer contents can predate the capture. In particular, the tested
Ridge Racer synthetic intervals contained writes to an offscreen buffer or no
writes to the currently displayed buffer; those pixels correctly show no recorded
contributors. Looking further back across buffers requires retained earlier
captures/replay, which is not provided here. Need for Speed mirror-area pixels
reconstruct correctly, but complete game-specific mirror ancestry has not been
proven; the synthetic offscreen-copy path is the verified source-history case.
Unknown history is labeled and never filled in from later RAM.

Rendering replay, stepping and scrubbing are U7. These checks do not establish
broad game compatibility, real-hardware fidelity, or Safari/Firefox support.
The U4 browser download-to-disk verification limitation remains documented there.
