# Prepared lessons and Elite loader introduction

Prepared lessons are available from **Open prepared lessons** in Session / controls,
or in the Code layout's Lesson / experiment pane. Load your matching local tape,
choose a lesson and select **Prepare and start**. Fort offers terrain decoding and
the enemy AI experiment; Elite offers a bounded fastloader introduction.

First use boots the original machine with a verified input recipe in a candidate
worker. It runs without waiting for real-time playback. The current machine is
paused and retained until the candidate succeeds. Cancel or failure discards the
candidate; successful preparation replaces the current session with the lesson.
The AI preparation applies the experiment's documented setup, but never applies
the candidate repair automatically. The original/modified comparison remains
explicit. Prepared-start replacement is stated in each lesson's description.

Verified raw checkpoints are cached in IndexedDB on this device. Repeat launches
skip the boot recipe. Cache failure or corruption falls back to local generation.
**Clear saved lesson starts** removes the cache. Nothing is uploaded and no game
RAM/checkpoint binary is published. The first preparation still takes computation;
this is not an instant first visit. A future licensed distributable checkpoint
could avoid that, but this release supports local generation only.

## Knowledge package contract

Optional `preparedStarts` maps stable IDs to:

- `title`, `description`, exact `releases` and provenance `evidence` references;
- `target: {kind: "tour" | "experiment", id}` for the action after preparation;
- `core` SHA-256, ordered three-ROM `firmware` SHA-256 array, expected full raw
  state `sha256`, and the required pre-execution `pc`;
- `actions`: bounded `run` cycle counts, key/joystick transitions, tape `play`,
  or `until` PC with an explicit `budget`;
- optional `layout: "loader"`, which opens independent Code, lesson, RAM atlas
  and tape-pulse panes. Tour stops can also request this via Suggested layout.

This initial profile is C64-only. It cannot inject RAM, patch code, execute host
scripts, fetch assets or guess prerequisites. Schema/semantic validation checks
references, input ranges, released inputs, at most 256 actions/32 starts and a
500-million-cycle total budget. Runtime has a two-minute wall budget and yields
between bounded slices for cancellation. Normal tour/experiment guards still
run after the prepared state is installed.

Cache identity includes the complete knowledge-package hash and start ID, media
names/sizes/hashes, core hash, firmware hashes and configuration. Every cached
payload must match the authored full-state SHA-256 before restore, then pass the
instruction-boundary check. A recipe run must match that same hash before it can
replace the session or enter the cache. Cache storage is bounded to 32 MiB with
least-recently-used eviction and 4 MiB per state. Browser storage denial does not
prevent lessons. Core/firmware changes require reauthoring and revalidation.

## Elite scope and compatibility gate

The single `games/elite-c64/knowledge.json` now contains the exact tape identity,
initial loader functions/annotations, prepared recipe and a four-stop tour:

1. `$0378`: configure the CIA2 discriminator latch ($0243 / 579 cycles).
2. `$03A3`: pilot and sync recognized; begin reading the block header.
3. `$03AE`: destination $0300, exclusive end $0334 established.
4. `$03B3`: first payload byte $8B stored at $0300.

The tour shows the initial 384/744-cycle pulse encoding and the RAM write. The
first stored byte already equals $8B; write evidence correctly records the store
without inventing an endpoint change. These stops execute authentic tape code,
without KERNAL hooks or injected loader RAM.

**The full self-modifying loader tour remains gated.** With the current shipped
core, the first observed disagreement with the independently extracted block is
at `$0306`: expected `$A8`, decoded `$B1`. A longer exploratory trace eventually
halted at `$8F5E`; it is not used as lesson evidence. The introduction ends before
this divergence and states its limitation in both the selector and final stop.
The extraction driver's successful load does not certify this cycle-stepped
browser core. Elite gameplay slots, random spawning and later protocol mutations
are still disabled. Next work is to isolate this first pulse/byte disagreement,
fix core compatibility, then validate the self-modifying stages and flight scene.

## Validation and reproduction

Public checks: `node tools/browser/tests/prepared-start.mjs`, Go knowledge tests,
and the existing tour/experiment suites. Rejection fixtures cover unknown targets,
script actions, invalid digests, held inputs and excess budgets. Runtime checks
cover identity mismatch, bad cache regeneration, full-state mismatch and wrong PC.

Private media (never committed):

```
node tools/browser/tests/prepared-media.mjs fort-apocalypse-c64 games/fort-apocalypse-c64/Fort_Apocalypse.tap
node tools/browser/tests/prepared-media.mjs elite-c64 games/elite-c64/Elite.tap
node tools/browser/tests/tour-elite.mjs games/elite-c64/Elite.tap
```

The first two run each recipe twice and restore its cached state on a fresh core.
The third verifies the complete authored tour and reports the subsequent decoding
discrepancy separately from the passing prefix.

Run `python3 tools/browser/tests/serve-viewport.py`, then open
`/tools/browser/tests/prepared-browser.html?report=prepared-chrome` (or Firefox).
Add `&game=elite` for the loader introduction. The fixture exercises cancellation,
authentic cold preparation, cache reuse and automatic tour/experiment entry.
The usual viewport fixture covers ordinary playback and system switching.

### Recorded acceptance — 2026-10-01

Release `754ae11c8642e9f65378`: Fort cold starts, cancellation and cache reuse pass
in Chrome and Firefox; Elite's four-stop introduction and four-pane layout pass
in Chrome. The ordinary viewport regression also passes. Adjacent JSON reports
record these results and the repeated native-WASM recipe/cache checks. Knowledge
validation, prepared-start rejection, tour, experiment, debugger, identity, shell,
layout and release suites pass. All emulator core hashes remain unchanged.
Safari was not run for this milestone.
