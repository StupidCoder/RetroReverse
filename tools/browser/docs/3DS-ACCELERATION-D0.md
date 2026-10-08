# D0: running-demo performance baseline

This milestone establishes the fixed-work benchmark for the
[25 percent campaign](3DS-ACCELERATION-25-PERCENT.md). It adds optional browser
diagnostics and a production-worker test harness. It does not accelerate rendering.

## Workload and method

The private checkpoint starts at display interval 3240 in Super Mario 3D Land's
attract sequence. Mario is traversing the blue-wall landscape on the top screen;
the bottom screen shows Press A. The 300-interval window ends at 3540, still in
the same running sequence. Local images at 0, 75, 150, 225 and 300 confirm changing
Mario positions and camera/scenery. No buttons are held. Cartridge, checkpoint
and images remain local, outside the public result.

The harness runs the production worker's execution pump, input handling, GPU
transport and both-screen presentation. It restores the same checkpoint for
every trial and runs five alternating trials per build. Each build first runs
a separate 30-interval cold trial. Warm trials reuse compiled browser/GPU
pipelines; restoring canonical state still invalidates guest caches.

The app is foreground in headful Chrome on an Apple M3 Pro MacBook Pro
(Mac15,6, 12 CPU cores, 36 GiB), on AC power with low-power mode off. Browser,
adapter, core hashes and checkpoint hash are recorded in the
[machine-readable report](../results/3ds-acceleration-d0-demo.json).

End-state serialization, hashing and PNG encoding occur after the timed window.
Five pixel copies within the window remain included and their cost is reported.
Ordinary trials close the Performance panel and disable hardware timestamps.
Separate full-length runs measure the open panel with GPU timestamps and then
the same configuration with vertex sampling enabled at one vertex in sixteen.
Do not combine timing-only and diagnostic runs with the ordinary trial averages.

## Result

All 14 runs pass canonical end-state and sampled-image comparisons. Chrome was
154.0.8037.98 with the Apple `metal-3` adapter. The source baseline release was
`b603d833217206c8a8ef`; exact baseline/candidate WASM identities are in the report.

| Configuration | Average ms/interval | Nominal rate | p95 ms |
| --- | ---: | ---: | ---: |
| Shipped, five ordinary trials | 317.59 | 5.25% | 653.10 |
| Diagnostic build, diagnostics off, five trials | 318.63 | 5.23% | 655.58 |
| Diagnostic build, Performance panel and GPU timestamps | 321.86 | 5.18% | 659.90 |
| Same, plus sampled vertex/fallback diagnostics | 324.72 | 5.13% | 666.30 |

The first two p95 values are averages of the five trial p95 values. The ordinary
candidate trials range from 317.18 to 321.00 ms/interval: approximately 0.33%
slower on average than the shipped build. Opening Performance/timestamps costs
about 1.0% relative to the ordinary candidate average; adding diagnostics costs
another 0.9% in the separate measured run. These are observed differences, not
claims of statistically isolated overhead. The slowest candidate 60-interval
window averages 334.29 ms. The distribution alternates between cheap intervals
and expensive rendering intervals; the mean alone hides the roughly 650 ms tail.

The final ordinary candidate trial attributes 98.02 ms/interval to PICA commands
and vertices, 131.81 ms to software rasterization, 50.06 ms to ARM/Horizon and
30.57 ms to GPU round trips. Reaching 25% from this workload needs about 4.78×
the current throughput. It is not within reach of submission tuning alone.

Across the full diagnostic window, 15,340,476 vertices contain 6,771,178 unique
indices within draws: **55.9% are repeated work**. Sampled fetch/input mapping
and shader execution estimates are approximately 49 and 56 ms/interval,
respectively; clipping/setup is about 1.31 ms. The sampled estimates exceed the
exact combined PICA bucket slightly and must not be summed as exclusive times.
Both attribute fetching and shader dispatch warrant optimization.

The largest fallback combinations, measured as software time per interval, are:

| Feature family | ms/interval |
| --- | ---: |
| Lighting, LESS depth/write, texture mask 69635 | 39.36 |
| Lighting, GREATER depth, stencil, texture mask 69632 | 20.98 |
| Lighting, LESS depth/write, texture mask 69639 | 13.62 |
| Lighting, stencil and texture-0 shadow sampling | 12.94 |
| Lighting, LESS depth without write | 10.67 |
| Unlit stencil/shadow sampling, first stencil configuration | 9.91 |
| Unlit stencil/shadow sampling, second stencil configuration | 9.62 |
| Lighting, LESS depth/write, texture mask 69633 | 7.67 |

These families can overlap in required features. A lighting implementation that
still rejects every stencil/shadow combination cannot remove the entire lighting
cost. Preserve the exact raw feature keys in the report when selecting D3 work.

The diagnostic window submits 6,898 GPU operations, about 23 per interval,
with no bridge refusals. Input packets total 512,904,628 bytes; surface uploads
and readback results each total 2,786,304,000 bytes. GPU compute averages 15.74
ms/interval inside 28.77 ms of queue/map wait. This remains evidence for later
batching and surface reuse, while the larger immediate costs are on the CPU.

## Diagnostic contract

`rr_perf_enable(0)` disables and clears the host-only counters. Strides 1–1024
enable sampled vertex clocks. Draws, submitted/unique indices, shader identities,
fallback families, bounding-box work and canonical pixel/depth/shadow counter
deltas are exact counts. Fetch and shader milliseconds are estimates, scaled
within each draw; clipping/setup and software fallback durations are timed scopes.
Small draws are sampled more heavily than long draws. Sampling is intended to
rank work, not produce an additive cycle accounting model. Command parsing stays
inside the existing PICA bucket; it is not separately timed by this milestone.

The counters do not enable machine read, write, pixel, trace or capture hooks.
They are not serialized and do not affect renderer eligibility. Bounded program
and fallback tables report overflow explicitly. Program hashes include shader
code, operand descriptors and entry point; uniforms remain runtime inputs.

The worker suffix counts submitted operations and input/before/result bytes. The
current bridge performs one submission/map round trip per accepted operation;
these are transport counts, not hardware draw-call counts. GPU timestamp time is
inside completion wait and must not be added to it again.

## Validation and reproduction

The complete public browser check passes, including the new native diagnostic
fixture. The existing 405-packet public graphics stream is byte-for-byte
unchanged. The real-scene harness compares complete end-state SHA-256, guest
instruction/frame/draw status and pixel hashes at all five sample boundaries
across builds, timing modes and repeated trials. This is baseline/diagnostic
validation, not the later D6 repeated-handoff and device-loss acceptance campaign.

Run a local static server on port 8790, supply an owned cartridge and a raw
checkpoint from the same running scene, then use:

```sh
PRIVATE_OUT=tools/platform/n3ds/browser/work/demo-performance/d0 \
DIAGNOSTICS=1 \
node tools/browser/tests/3ds-demo-browser.mjs \
  /path/to/owned-cartridge.cci /path/to/running-demo.state \
  tools/browser/results/3ds-acceleration-d0-demo.json 300 5
```

Set `PLAYWRIGHT_MODULE` when Playwright is installed outside normal module
resolution. `BUILDS` accepts comma-separated `name:core-directory` pairs. A
changed workload requires fresh start/middle/end visual verification; matching
hashes alone cannot establish that the benchmark is still the running demo.
