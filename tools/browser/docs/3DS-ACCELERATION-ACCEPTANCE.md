# M5: reference inspection and prototype acceptance

The browser 3DS prototype now supports a selected Play mode independently of its
current execution mode. Reference remains the default. Render capture, Memory
snapshot/record requests and frame stepping temporarily select Reference. Resume
uses the selected Play mode. Explicit selection of Reference changes that
preference. The selector and status message identify temporary reference
inspection and the mode Resume will use.

Render still finishes the current display interval and records the next three
complete intervals. Its displayed interval range identifies that later capture;
it does not reconstruct the previously displayed Play image. Pixel histories and
isolated replay belong to the reference capture. CPU instruction stepping and
breakpoints remain M6 work, and historical Play-frame reconstruction remains M7.

## State, cancellation and recovery

The portable container saves `playMode` as host metadata. Loading validates the
same media/configuration and combined WASM/renderer identity, restores canonical
state in Reference, then enables the saved preference if available. Missing
WebGPU loads in Reference. A fresh machine/reset selects Reference. No new
machine-state format or second independently maintained emulator is introduced.

The execution gate and existing generation checks protect the asynchronous
handoff. Pause suppresses a pending switch's automatic continuation; cancelling a
capture releases its ownership before subsequent execution. Input remains queued
outside suspended WASM. GPU scratch buffers are overwritten from canonical RAM
for every operation, so reference execution or restoration cannot leave a stale
GPU target or texture cache authoritative.

Adapter/shader/presentation initialization and each live GPU operation have
five-second deadlines. Late initialization results are disposed. Unsupported
WebGPU, missing adapters, failed shaders, invalid self-test output, synchronous
submission failure, timeout and device loss leave Reference available. If the
backend fails while inspection has temporarily selected Reference, Resume also
falls back to Reference. Canonical RAM always contains all completed GPU writes;
an uncommitted failed operation executes once through the reference continuation.
There is no GPU-only guest state requiring checkpoint rollback/history.

## Acceptance tests

The private endurance harness runs each available Mario browser checkpoint for
520 display intervals in Reference, then repeats it with deterministic button,
circle-pad and touch inputs and frequent mode changes. R-button pulses and corner
touches keep these scenes rendering; each 50-interval sample must show further
GPU draws. Every intermediate checkpoint compares full serialized machine bytes,
CPU/scheduler progress and draw counters. Each run includes five three-interval
reference captures, pixel evidence on both screens, replay to the end and back,
non-invasive physical-memory inspection and repeated state restore/continuation.
Scrubbing and inspection must leave the machine's state hash unchanged.

Both sustained runs pass: **1,040 switched display intervals, 540 actual mode
changes, ten matching captures and 22 matching intermediate checkpoints** (plus
the corresponding 1,040 Reference intervals). The welcome run accelerates 7,089
draws; title accelerates 8,250. GPU scratch stays at 16,842,820 bytes throughout
each switched run. WASM capacity stays at 2,132,344,832 bytes for welcome and
2,147,483,648 bytes for title during the switched passes, after the reference
passes have warmed capture/checkpoint storage. Reports:
[welcome](../results/3ds-acceleration-m5-handoff-welcome.json),
[title](../results/3ds-acceleration-m5-handoff-title.json).

The [UI harness](../tests/3ds-handoff-ui-browser.mjs) loads a local browser
checkpoint through the actual state-file control, plays with WebGPU, pauses,
captures in Reference, inspects a pixel, resumes Experimental, snapshots Memory,
saves/loads the selected preference, tests Pause during a switch, cancels capture,
advances a reference frame and resets. A second worker with WebGPU unavailable
loads the saved Experimental preference in Reference. Private screenshots remain
local.

The [recovery harness](../tests/3ds-recovery-browser.mjs) destroys a real GPU
backend while mapping an in-flight result after earlier successful commits. It
also injects a completion timeout and three milliseconds of host latency per
operation. Six-interval final machine-state and framebuffer hashes equal the
Reference run in every case. Public tests exercise missing adapters, shader
failure, startup timeout, late disposal and transition cancellation.

Reports are `../results/3ds-acceleration-m5-*.json`. These are continuation tests,
not performance measurements; use the warmed [M4 results](3DS-ACCELERATION-COVERAGE.md)
for end-to-end speed. The M4 target of 2× was missed: the measured gain is about
1.84× on welcome and 1.83× on title, well below real-time emulation.

## Support and limits

| Area | Validated status |
| --- | --- |
| Browser/GPU | Chrome 154, Apple Metal 3; other browser/GPU combinations untested |
| Graphics | Exact bounded transfers, narrow stencil, unlit integer fragment tails; see M4 support table |
| Unsupported PICA features | Existing software reference fallback; includes lighting and general depth/stencil |
| Handoff | One shared browser machine; quiescent continuation, captures, replay, save/load and input checked |
| No WebGPU / failed backend | Reference remains usable; Experimental is disabled with a reason |
| CPU | ARM interpreter; no JIT or instruction debugger added |
| Fixtures | Mario welcome and title/Mii-dialog; no verified interactive gameplay checkpoint |
| Other games | M1 Captain Toad cold-boot baseline only; no claim of accelerated gameplay compatibility |
| Memory | GPU scratch is bounded; the existing WASM/capture pipeline remains memory-heavy |

WASM heap length is a high-water capacity, not live allocation or resident memory.
The 1 GiB cartridge, machine/checkpoint copies and repeated captures bring it near
2 GiB; it does not shrink after freeing buffers. These tests monitor capacity and
GPU allocation, not a complete native-allocator leak census. M4 live presentation
uses approximately 16.6 MiB of GPU buffers; the endurance harness without scanout
presentation uses approximately 16.1 MiB. Mobile memory suitability is unproven.

The prototype is browser-only. No Go emulator optimization, parity work or Go
oracle is part of these milestones. Public `tools/browser/check.py` covers the
shared application's existing tests; the Go port generator is build plumbing.
