# M4: exact fragment acceleration

Experimental now combines reference float32 coverage, perspective interpolation
and texture sampling with an integer WebGPU fragment tail. Each pixel receives
an ordered list of fragments. A single cached compute pipeline evaluates all six
TEV stages, delayed combiner-buffer updates, alpha tests, blend equations/factors,
logic operations and channel masks. Guest uniforms and state are parameters, so
there is no guest-program pipeline cache to invalidate or asynchronous compile
queue to accumulate.

The workload census found unlit textured draws without depth writes/tests or
stencil in both available Mario checkpoints. This made integer fragment work a
useful next exact subset. General GPU float interpolation is still excluded:
matching this browser reference takes precedence over conventional raster speed.

## Support boundary

| Operation | Experimental behavior |
| --- | --- |
| Fills, non-overlapping copies, supported display conversion | Exact compute from M2–M3 |
| Stencil NEVER with bounded half-pixel vertices | Exact compute from M2–M3 |
| Unlit textured/color draws, depth and stencil disabled | Reference coverage/sampling; GPU TEV, alpha, blend/logic and masks |
| Texture formats 0–13 and normal texture coordinates | Existing reference decoder and sampler; decoded colors feed GPU |
| Lighting, depth tests/writes, general stencil, shadow/projection sampling | Full reference fallback |
| Vertex programs, clipping, ARM, kernel and DSP | Existing reference execution |
| Capture, access tracing or profiling hooks | Reference execution |
| Tiny draws, oversized preparation, unsupported TEV/blend, invalid input | Reference fallback before writes |

A draw may prepare at most 16 MiB of fragment data. Links are validated for bounds,
alignment, unique ownership and forward ordering before GPU submission. Outputs
are committed only after buffer size and fragment counters are validated. Inputs
borrow suspended WASM memory until queue uploads copy them; recording owns copies.
The bridge retains the M3 eager readback contract and one-operation queue.

The larger texture fixture exposed an existing C++ map-iteration bug in texture
invalidation. Erasure now advances the returned iterator. This fix is preserved
by the browser generator and applies to both browser modes; no Go emulator code
was changed.

## Evidence

The public fixture contains 149 operations: 147 supported and exactly matching,
plus two overlapping copies correctly refused. Its 96 fragment draws cover actual
textures in all 14 formats, six randomized TEV stages, alpha comparisons, all
blend equations/factors, 16 logic operations, partial masks and overlapping
primitives. Fragment counters match alongside every output byte. Seed: 0x3d500002.
The fixture also checks that later source writes cannot change recorded packets.

Private live trials restore the same browser checkpoint, apply deterministic
buttons/circle-pad/touch, run 30 display intervals and present both screens every
interval. Both engines warm for two intervals first; three trials alternate
execution order. Full saved machine bytes and final scanout agree, with no pixel
differences. Shader self-test/compilation and checkpoint hashing are outside the
timed region. Display intervals are not necessarily unique game frames.

See the committed [welcome](../results/3ds-acceleration-m4-live-welcome.json),
[title](../results/3ds-acceleration-m4-live-title.json) and
[public replay](../results/3ds-acceleration-m4-public.json) reports.
Median times per display interval including presentation are 277 → 150 ms for
welcome and 325 → 178 ms for title. GPU scratch plus presentation stays at
17,367,108 bytes in these trials. Welcome accelerates 819/1,699 draws; title
990/2,100. All fills and supported display transfers also run on the GPU.
The final repeated results were 1.834–1.846× (welcome) and 1.827–1.837× (title),
below the 2× target and far below real-time 60 intervals/s. Remaining CPU coverage,
interpolation and sampling, fallback draws, and per-operation upload/readback
limit this deliberately conservative implementation. Host transfer timing includes
queueing and JS work, not just GPU execution. M2 isolated replay provides GPU
query timing; live profiling disables timestamp queries to avoid measurement cost.

These are menu/welcome checkpoints, not verified interactive gameplay. Chrome
154 on Apple Metal 3 is the measured environment; no cross-browser or cross-GPU
performance claim is made. Raw cartridges, checkpoints and graphics streams are
local only. The next milestone tests repeated handoffs and reference inspection.
