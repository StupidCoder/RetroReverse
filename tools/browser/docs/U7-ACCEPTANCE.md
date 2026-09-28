# U7 — isolated rendering replay

The display now has First/Previous/Next/Last controls, a progress slider,
optional changed-pixel highlighting, seek cancellation, and contributor-to-step
navigation. C64 steps visible raster lines; other systems step captured GPU/RDP/
cel commands and CPU/HLE write groups. The inspector continues to describe the
final captured pixel, explicitly labeled while viewing an intermediate image.

Implementation choice: replay uses the capture's initial memory image and a
journal of actual, historically resolved writes, including CPU/HLE effects.
It does not execute a command list against final texture memory. It also does
not re-execute a second CPU or expose historical CPU registers. This narrower
rendering replay is deterministic without repeating guest I/O or HLE side
effects. C64 similarly replays recorded raster output over its initial display.
The capture start/end complete-machine checkpoints remain available; the live
machine stays at the end throughout inspection.

Scratch memory is separate. Up to 16 MiB of sparse, bounded intermediate memory
checkpoints accelerate backward/repeated seeking; uncached work yields after
32,768 writes and can be cancelled. Capture completion compares the last replay
image byte-for-byte with the frozen display and labels incomplete evidence.
Pixels whose values predate the capture remain visible at step zero. Offscreen
writes are replayed, but the preview maps the buffer used by final scanout; it
is not a historical monitor signal with changing display origins.

Validation: `tests/replay-wasm.mjs` restores each reference-game checkpoint,
captures an interval, checks final RGBA equality, repeats arbitrary/backward
seeks, compares the complete live state before/after inspection, and verifies
continued execution equals a fresh continuation from the saved end state. All
four pass; measured results are in `results/u7-*.jsonl`. Native `replay.cpp`
covers write groups, rejection, cache use and bounded incremental work; U6 tests
cover source overwrite and offscreen composition values used by this journal.
The C64 browser workflow exercised First/Next/Last, highlighting and a sprite
pixel's contributor jump. Observed C64 browser seeks were about 0.2 ms.

Limits inherited from U6 still apply: overflow/missing history, PS1 blend/mask
coverage and simplified N64 scanout. This is rendering-effect replay, not full
CPU debugging or universal source taint analysis. GPU state/memory watches are
reserved for the later UI design milestone.
