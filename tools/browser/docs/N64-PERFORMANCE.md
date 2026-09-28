# N64 interpreter performance

Pilotwings 64 was measured from the same local checkpoint for 120 modeled VI
intervals. Each interval is 750,000 CPU run-loop ticks in the existing model;
these are not distinct game images or accurate hardware cycles. The test warms
three intervals, restores again, and excludes state serialization, proof hashing
and browser presentation from timing. Cartridge/checkpoint data is not published.

Three alternating runs of each build on the development Mac:

| WASM core | Median intervals/s |
|---|---:|
| Before | 13.18 |
| After | 23.71 |
| Speedup | **1.80×** |

The first profile was roughly 39% VR4300/scheduler, 44% RSP and 16% RDP. Unlike the
3DS, allocation was not the dominant problem. Targeted inlining keeps CPU and RSP
interpreter dispatch paths together instead of repeatedly passing machine state
through WASM function calls. Local helper lambdas retain their concrete types,
and the build uses `-O3`. In a representative repeated run the RSP fell from
4.09 s to 1.39 s and VR4300/scheduler from 3.51 s to 2.23 s; RDP stayed near 1.47 s.

Explicit SIMD and more aggressive rasterizer inlining were explored. Their
additional gains were not compelling enough to retain. The final build does not
require SIMD, JIT compilation or hardware graphics. It retains
`-fwrapv -ffp-contract=off`, instruction counts, interrupts, microcode execution,
rendering and capture hooks. There are no game-specific shortcuts or skipped
frames. Larger gains would require work on interpreter/vector execution or the
renderer itself.

All six benchmark runs have identical CPU/RSP/RAM/DMEM/IMEM checkpoints and
complete serialized end-state SHA-256 hashes. Validation also includes 65,536
recorded RSP vector cases under UBSan, CPU branch-delay and memory-boundary tests,
and a WASM capture with 15,523 replay steps. Final pixels, arbitrary seeks,
unchanged live state and deterministic continuation pass.

Reproduce with private media and a raw or portable checkpoint:

```sh
node tools/browser/tests/benchmark-n64.mjs /path/to/cartridge /path/to/checkpoint 120
```

`CORE_DIR` selects a directory containing a different `core.js` and `core.wasm`.
Full measurements: [n64-performance.json](../results/n64-performance.json).

After translator changes, regenerate and restore instrumentation:

```sh
go run ./tools/platform/n64/browser/portgen
python3 tools/browser/instrument.py
python3 tools/browser/capture-instrument.py
python3 tools/platform/n64/browser/build.py --emcc /path/to/em++
```
