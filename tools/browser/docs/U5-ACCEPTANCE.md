# U5 — bounded capture of the next complete display interval

Implemented September 28, 2026. Pause aligns to a display boundary, saves a
start checkpoint, enables instrumentation for one full declared interval, then
saves the end checkpoint and frozen output. It leaves the live machine at that
end state. C64 uses a VIC raster interval; PS1/N64 use the existing explicitly
synthetic field boundaries; 3DO uses consecutive DisplayScreen calls.

The worker owns one capture generation per session, two checkpoints (32 MiB
maximum each), frozen pixels and input state. Run, reset, replacement, restore
and cancellation invalidate the generation. Progress yields between bounded
execution slices. Cancel stops where execution reached and discards partial
evidence. A loading stall remains cancellable rather than silently becoming a
successful capture. Source snapshots and old/new writes are captured at their
historical time, not reconstructed from later live RAM.

C64 retains the existing bounded raster/fetch/writer evidence. The other cores
share byte-addressed evidence with a million-write cap, 16,384 command cap and
16 MiB metadata/source-snapshot cap. PS1 records GPU fills/primitives/transfers
and texture/CLUT selections; N64 records RDP commands, immutable deduplicated
TMEM snapshots, modeled writes and depth/alpha rejects; 3DO records CCB/cel
metadata, PDAT snapshots, PLUT values, PIXC outputs, transparent candidates and
CPU/HLE framebuffer writes. The instrumentation stays disabled during ordinary
execution. Direct source-to-writer navigation is part of U6.

Unknown ancestry is explicit: initial buffer contents predate the capture.
Unobserved changes are detected by comparing reconstructed bytes with captured
final bytes. Overflow marks evidence incomplete; it never fabricates an origin.
Submission PCs describe command dispatch, not the instruction that constructed
the command. Existing hardware-model limitations remain (notably PS1 blend/mask
behavior and N64 coverage/scanout approximations).

## Validation

`tests/capture-wasm.mjs` restores late native checkpoints, runs an untraced
interval, restores again, traces the same interval, then compares the entire
serialized end state byte-for-byte. Three intervals pass on each core. Reference
paths include Ridge Racer track rendering, Pilotwings RSP/RDP execution, Need
for Speed's native attract/movie sequence and C64 keyboard/raster state.

`tests/capture-buffer.cpp` verifies overwrite, rejection without a write,
unchanged background, deduplicated snapshots, unobserved-change detection and
hard overflow bounds. Captures are sampled through pixel queries. Out-of-bounds
coordinates are rejected explicitly.

Observed trace storage, excluding checkpoints: C64 about 10 MiB, PS1 about
22 MiB, N64 about 24 MB and 3DO about 14 MB. The initial N64 implementation
exhausted metadata by duplicating TMEM; deduplication removed that overflow on
the tested intervals. In the browser, a C64 Pause recorded display 16–17 in
about 20 ms, with no overflow. A 3DO capture was cancelled in the browser;
the machine stayed paused at display 10 and remained runnable. These are observations, not worst-case bounds.

The shared compact contributor inspector is U6. General rendering replay and
seeking remain U7. This capture does not imply full hardware accuracy or complete
history before its start checkpoint.
