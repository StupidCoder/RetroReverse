# DS / 3DS performance refactor

Measurements use local Super Mario 64 DS and Super Mario 3D Land cartridges on
the development Mac. `tests/benchmark-handheld.mjs` restores the same checkpoint,
warms three intervals, restores again, and measures complete modeled display
intervals. Loading, proof hashing, state serialization and browser presentation
are outside the timed region. Images and checkpoints are not distributed.

Three-run median core throughput:

| Workload | Before | After | Speedup |
|---|---:|---:|---:|
| DS / Super Mario 64 DS title, 120 intervals | 7.04/s | 54.45/s | 7.73× |
| 3DS / Super Mario 3D Land welcome, 30 intervals | 0.894/s | 4.615/s | 5.16× |

DS framebuffer checkpoints match the previous build; CPU/temporary RAM changes
reflect the corrected WFI behavior. 3DS checkpoint proofs and complete serialized
end-state SHA-256 hashes match the previous build in all six runs.

## DS

The first profile attributed 83% of execution time to the CPU/scheduler. Native
sampling showed that an instruction fetch repeatedly resolved four byte reads,
copying reference-counted slice owners each time. The optimized bus borrows
bounded memory spans, handles ordinary RAM accesses at their natural width, and
retains the original path for devices, boundary crossings and watchpoints.
There are no persistent host-pointer caches to invalidate on WRAM/TCM remapping
or state load. Tests cover mirrors, all WRAM modes, unaligned accesses, TCM edges
and watched byte order.

Local rendering lambdas retain their concrete types. VRAM page lookups borrow
their page lists. The scheduler reuses empty milestone maps and records a code
page once per contiguous visit in a quantum, rather than hashing every ARM9 PC.
These changes alone improved the initial benchmark from 7.1 to 29.3 updates/s;
CPU, RAM and screen checkpoints still matched exactly.

The next profile exposed a missing hardware behavior: CP15 wait-for-interrupt
was ignored. Almost all sampled ARM9 instructions cycled through a WFI idle
routine. The Go and C++ cores now implement `MCR p15,0,Rd,c7,c0,4`, stopping
instruction execution until the IRQ line asserts. CPSR.I still gates exception
entry; a masked exception wakes the CPU without entering the handler. BIOS
IntrWait remains separate. See the [ARM946E-S Technical Reference Manual,
page 2-24](https://documentation-service.arm.com/static/5e8e3ee588295d1e18d3aa82).
The instruction clock, timers, ARM7 and display keep advancing while ARM9 waits.

WFI is serialized. Raw browser states now use DS version 2; the development loader
can also read version 1, which has no WFI flag. Public state containers continue
to require an exact core hash. The Go snapshot adds an optional gob field.

Validation covers 1,201 checkpoints with stylus input through boot and the intro.
Optimized and unoptimized C++ with the same WFI correction match at every frame:
both CPUs, scheduler, mapped RAM and both screens. Go agrees on all CPU/RAM
checkpoints; one screen hash differs at frame 915 in both C++ variants, so that
cross-language rendering difference is not caused by the optimizations.
WASM matches the C++ reference at every frame. State continuation, capture final
pixels, arbitrary replay seeks and isolation of the live state pass. Go tests
cover WFI instruction retirement, IRQ gating, masked wake-up and saved wait state.

## 3DS

The first profile attributed about 97% of execution time to PICA rasterization.
The translator had turned small local helper functions into `std::function`:
perspective interpolation and TEV source selection allocated captures per pixel,
and blending invoked type-erased helpers repeatedly. They are now ordinary typed
lambdas/direct function aliases, which the compiler can inline. Render-target
subviews borrow memory for the duration of the draw instead of incrementing and
decrementing shared ownership for each fragment. Floating-point contraction stays
disabled; no reduced-precision rendering, frame skipping or hardware GPU path is
used. An experimental combiner rewrite did not improve the measurement and was
discarded.

Cold boot, deterministic state continuation and detailed capture/replay pass.
The exercised capture contains 4,452 events and 2,342,966 writes without overflow;
its last replay image matches exactly and replay leaves the saved machine intact.
The existing compatibility, memory and audio limitations still apply.

## Browser scheduling

The worker now uses a MessageChannel to yield between work slices. Real pacing
delays still use timers. This avoids repeatedly nesting zero-delay timers while
keeping input and cancellation on asynchronous tasks; it does not busy-spin or
skip display boundaries. Browser tests exercise Run/Pause, controller/stylus
input, capture, historical pixel inspection and rendering seeks.

Chromium observed 31–39 DS updates/s after optimization versus 4.1 before.
A DS capture completed in 0.14 s with 3,423 replay steps; its selected pixel
reconstructed exactly and linked to the captured 3D source. 3DS reached 4.6
modeled display updates/s (about 2.2 presented images/s in this scene). Its
three-interval capture completed in 2.41 s with 24,040 replay steps and a
1,717 MiB heap; transfer/source history and replay seeking remained correct.
These are observed samples, not a realtime guarantee for other scenes or hosts.

Final repeated measurements are recorded in `../results/handheld-performance.json`.
