# 3DO rendering and movie refactor

The 120-display Need for Speed driving benchmark runs from the same local native
checkpoint, holds A, and records complete CPU, DRAM, VRAM, item-memory and display
hashes every 30 updates. On the development Mac, the deployed baseline took 8.60 s
(13.96 updates/s); the optimized build took 2.45 s (48.95 updates/s). All four
checkpoints match exactly. This is a Node/WASM core throughput measurement; it
excludes browser presentation and does not establish cycle-accurate 3DO timing.

The dominant cost was cel rendering. Changes replace type-erased pixel callbacks
with template dispatch, decode tokens a byte at a time instead of a bit at a
time, specialize single-pixel footprints, bulk-copy bounded source snapshots,
and write unobserved VRAM clears/opaque pixels directly. Watched reads and writes
retain the bus path. Source snapshots still preserve overlapping-buffer semantics.
The 3DO build uses `-O3`; the other cores keep their previous flags.

Need for Speed's known-media profile now uses the existing Cinepak HLE decoder
instead of its incomplete native DSP/DataStreamer playback. It intercepts movies
requested by the game, shows all decoded frames in VRAM, and paces presentations
using FILM timestamps/durations. The EA Canada logo and the following intro are
visible. This is explicitly movie HLE, not a completed audio/DSP implementation;
there is still no audio output. Disabling the compatibility profile retains the
native movie path for investigation.

Movie capture records actual VRAM writes as Cinepak HLE events. Mid-movie states
save the decoder's codebooks/image and playback cursor; compressed frames are
reloaded from the same local disc rather than embedded in every checkpoint.
Completed movie payloads are released. New raw state format is version 2; public
state files still require an exact matching core hash. Legacy non-movie native
checkpoints remain accepted by development tests.

Validation:

- `benchmark-3do.mjs CORE_DIR DISC NATIVE_STATE [FRAMES]`: recorded before/after
  measurements in `results/3do-performance-{before,after}.json`.
- `fast-3do.cpp`: every bit alignment, token sizes 0–64, zero-padding past end,
  odd-length clears and write-watchpoint fallbacks.
- `movie-3do.mjs DISC`: cold boot, hundreds of movie frames, timestamp progression,
  mid-movie state continuation, movie provenance, replay end and live-state isolation.
- Existing driving capture/replay tests pass, including unchanged continuation.
- Public media-free CI suite passes.

Regeneration: run the existing translator, `instrument.py`,
`capture-instrument.py`, then `tools/platform/threedo/browser/optimize.py`,
`tools/browser/state/generate.py` and the 3DO `slice.py` before rebuilding.

Chromium integration: the restored cold-boot driving scene sustained 30 presented
updates/s in normal mode. A later fast-forward sample showed 36.5 updates/s
(122% of the 30-update target), including worker yields and presentation. Gameplay
capture completed in 0.28 s with 421 replay steps; the selected pixel's recorded
writes reconstructed its display color. Start successfully paused and resumed the
game after replacing the old display-count input clock with the emulated field
clock plus HLE movie time. This prevents a short Start press becoming stuck when
the game's own pause stops submitting displays. Movie input and pacing clocks are
serialized with the machine. Browser performance varies with scene and machine;
49 updates/s is core-only throughput for the recorded benchmark, not a universal
browser guarantee.
