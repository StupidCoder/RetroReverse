# U3 — Transport, inputs, profiling and scheduling

Implemented September 28, 2026. The shared static shell provides Run/Pause,
Reset with mounted File objects retained, next-display stepping, nominal pacing
and fast forward, keyboard/touch controls, and standard browser gamepads.

## Inputs

Keyboard, pointer and physical-controller states are aggregated independently.
The worker applies an ordered input queue at execution boundaries, in emulated
time. Short taps survive a slow worker receiving both edges between slices;
C64 typing has keyboard scan hold/release intervals. Reset creates a fresh
worker; blur releases inputs; hidden documents release and pause. Gamepads use
an 18% analog dead zone, release on disconnect, and only inject while the display
or controller buttons have focus. The UI identifies the device and mapping;
non-standard mappings are explicitly unsupported rather than guessed.

C64 LOAD/RUN was typed through the actual browser UI, including rapid typing,
and reached Fort's game/attract display. Tape Play, Pause, step and Reset were
exercised; reset retained the selected tape and returned the machine to cycle 0.
PS1 active-low button wiring is corrected. Ridge Racer rendered its track in the
browser; keyboard Start, Pause, next-field stepping and Reset were exercised.
3DO Start/Pause/step were exercised.
**No physical gamepad was connected**, so physical-device validation is still
open. Unit tests cover short taps, unsigned 3DO masks and ordered C64 key edges.

## Timing model

Every C++ core exports cumulative timing buckets. The UI reports interval deltas,
exclusive milliseconds and percentages, separately from frame presentation.
Nested scopes subtract their children; the CPU/device/scheduler remainder is
labeled. No instruction count is presented as measured time.

- PS1: R3000A/devices remainder, GTE, GPU/software rasterizer, local disc I/O.
- N64: VR4300/scheduler remainder, RSP, RDP/software rasterizer.
- 3DO: ARM60/scheduler remainder, Portfolio HLE, Portfolio SWI, cel rasterizer,
  flash clear and local disc I/O.
- C64: one sampled chip tick per 1,021 ticks, split into 6510, VIC-II, CIA, SID
  and board glue. Milliseconds are **sample-only**, not extrapolated totals.
  These tiny measurements are noisy and include timer overhead; they are labeled
  accordingly. Cross-origin isolation under `/emulators/` improves timer precision.

The UI also reports display update rate, CPU throughput, nominal speed,
allocated WASM heap, longest execution call and canvas-copy time. PS1/N64 fields
are explicitly synthetic; 3DO display calls are not hardware-accurate frame time.
Pacing excludes pause time, drops excessive catch-up credit, and rebases when
fast forward changes. Capture/replay is not yet implemented in this shell.

## Scheduler and file-I/O verification

3DO retains Run's seen-PC map, spin ring, progress counters, scheduler counters
and logical run step across yields. The original whole-frame Run remains as an
oracle. It resets that context only when the original run would end.

The profiler exposed whole-movie reconstruction on every small file READ, plus
a byte-by-byte implementation of Go's bulk append. Replacing the latter with
alias-safe bulk copies and serving ordinary disc READ/STATUS from sector-backed
extents removed the long SWI calls. Closed streams release their file buffers;
normal SWI/kernel diagnostic histories are bounded. These changes affect host
resource use, not guest-visible execution. Disc cache: eight 1,024-sector chunks;
PS1 cache: 64 sectors. Neither disc is copied wholesale into WASM.

Recorded native comparisons:

- PS1: 10,000 versus 17,777 instruction slices, in-memory versus cached disc,
  generic interrupt hook and scripted inputs, matching checkpoints through
  500 million instructions.
- N64: both slice sizes match through 300 million steps, including 139 million
  RSP instructions and 2.69 million RDP words. Three byte orders normalize to the
  same boot state; unsupported IPL3 is rejected.
- 3DO: 10,000/17,777 slices match original whole-frame execution through display
  300. The final 17,777-slice run with range I/O matches all 41 recorded native
  checkpoints through display 1,200, including the scripted driving sequence.
- Bulk append tests cover self-append, overlapping ranges and nontrivial types.

Results: `../results/u3-*-slices.jsonl`. Tests: `../tests/`.

## Browser responsiveness and limits

Chromium on the reference Mac, same static site with Pages headers:

| Platform / observed path | Longest execution call |
| --- | ---: |
| C64 / reset, tape load and Fort display | 4.0 ms |
| PS1 / Ridge Racer boot and track display through 3,969 synthetic fields | 5.2 ms |
| N64 / Pilotwings boot through about 153 million steps | 71.9 ms |
| 3DO / native intro and attract sequence through about 3,000 displays | 54.4 ms |

The intermediate 3DO implementation recorded 1.35-second SWI slices. The final
file-I/O changes removed those outliers in the observed run. The old full-frame
call was about 28 seconds; it is no longer the shared shell's execution API.
These observations meet the normal <250 ms target on the tested paths; they
are not upper bounds for arbitrary images, synchronous N64 RSP/RDP tasks,
slow storage, browser scheduling or one-time image loading. The UI acknowledges
pause immediately and shows the measured longest call.

The optional exact-image Need for Speed profile still supplies the inherited
VBL timing global at RAM address `0x42734`: `advanceVBlank(n)` writes `n * 100`
centifields, not the monotonic field total. This is a direct game-memory timing
workaround for incomplete Portfolio HLE. The profile also defers inputs until
display call 300 to avoid an early-input stall. It does not replace movie
decoding, rendering or gameplay. Unknown images do not receive this profile.

Remaining validation: actual physical gamepad/disconnect hardware, Safari and
Firefox for the shared shell, additional commercial titles and sustained memory
behavior across many different game scenes. Elite's corrupted display remains
a known compatibility result, not a pass. Save-state files and provenance/replay
remain U4–U6 work.

Header references: https://developers.cloudflare.com/pages/configuration/headers/
and timer precision: https://developer.mozilla.org/en-US/docs/Web/API/Performance/now.
