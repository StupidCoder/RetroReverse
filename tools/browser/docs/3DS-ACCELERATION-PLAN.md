# 3DS acceleration prototype

Build an optional accelerated execution mode for the existing browser 3DS
emulator, with a verified transition back to reference execution and inspection.
The first useful result is a faster playable scene that can be paused, captured
with the software renderer, inspected, and resumed without losing machine state.
Full-speed execution is a target to measure, not a prerequisite to reporting an
experiment honestly.

Scope is exclusively the browser emulator. **Reference** means the existing
browser C++/WASM interpreter and software renderer. The Go emulator is outside
this work: no Go-emulator optimization, feature parity, oracle comparisons or
cross-language validation is required. Existing Go-based build tools may remain
build plumbing; they do not define the prototype's reference behavior.

M0–M3 are complete. M1's [state inventory and evidence](3DS-ACCELERATION-STATE.md)
record the baseline and transition interface. M2's [replay results and decision](3DS-ACCELERATION-REPLAY.md)
select a deliberately narrow exact compute subset, with general draws retained
in Reference. M3's [live bridge](3DS-ACCELERATION-LIVE.md) uses eager materialization
and GPU scanout presentation. M4–M5 are the active prototype
work. Complete, validate, commit and push each milestone to `main` separately.
Keep the default reference mode usable at every milestone.

## Product decision

Add an **Execution mode** selector to the existing `/emulators/3ds/` Play view:

- **Reference**: ARM interpreter and software PICA renderer; default.
- **Experimental**: the accelerated backend where supported, with reference
  fallbacks and an explicit indication of the backend actually in use.

Use one worker and one machine state, with interchangeable execution components.
Keep CPU execution and graphics backend selection independent internally: the
first accelerated mode retains the ARM interpreter. Avoid a second emulator page
or a second independently evolving machine model. A separately packaged WASM
module is an implementation option only if needed; it must satisfy the same
state-transfer tests and retain explicit build identity checks.

Selecting a mode while running requests a cooperative transition. Show progress
and acknowledge the effective mode only after success. Preserve mounted media,
guest time, held controls and queued input; do not reboot the game. Serialize
transitions against capture, save/load, reset, stepping and memory jobs through
the existing execution gate. Stale asynchronous replies cannot affect a newer
session or transition.

Pause into Render switches to reference execution and records the next complete
capture window. Resume returns to the selected Play mode. An explicit switch to
Reference stays selected. Preserve the current three-field 3DS capture window
until tests establish a different producer window. Clearly distinguish the last
Play image from the later reference capture. Reconstructing the image already on
screen is a later milestone, not an implied capability of Pause.

Unavailable WebGPU disables Experimental with a concise explanation. A partial
backend reports fallback use; it must not claim accelerated rendering while all
work is taking the reference path. Browser device-loss recovery must use a
verified recovery point, never silently continue with stale graphics memory.

## Existing foundation and gaps

- [The port](../../platform/n3ds/browser/README.md) is a generated C++20/WASM
  ARM11/Horizon HLE machine with software PICA rendering. Its worker pool is
  serial. Its existing browser behavior is the reference for this prototype.
  Keep browser acceleration code in maintained C++/JavaScript/WGSL modules and
  preserve their integration across any regeneration of the existing core;
  this does not require changing the Go emulator.
- [The saved performance run](../results/handheld-performance.json) reports
  approximately 217 ms per display interval in the median optimized welcome
  scene: 179 ms rasterization, 29 ms transfers and 7 ms CPU/scheduler. These are
  historical measurements, preceding subsequent rendering fixes. Rebaseline
  before choosing new optimization work. Display intervals are not unique
  presented game frames.
- [The core API](../../platform/n3ds/browser/core/api.cpp) runs synchronously.
  Rendering and GX transfers can occur inside an execution call. Adding an
  asynchronous GPU therefore requires explicit resumable boundaries; replacing
  a raster call with a fire-and-forget submission is insufficient.
- [Portable state restoration](../../platform/n3ds/browser/core/state.h)
  rebuilds memory mappings and callbacks and invalidates shader/texture caches.
  Extend that discipline to accelerated resources. State-file serialization is
  a validation tool, not a required full copy on every mode switch.
- [The shared worker](../../../site/emulators/worker.js) presents copied RGBA
  frames from WASM. GPU-resident Play presentation needs a separate transport
  path; repeatedly reading complete images back into WASM is not the target.
- Existing capture records rendering effects and register activity. It is not
  yet a self-contained graphics input stream with every historical texture,
  shader and render-target dependency. GPU replay needs that additional data.
- Browser 3DS currently exposes Render and Memory inspection. It does not expose
  the shared instruction-debugger API used by C64, DOS and 3DO. M6 adds the ARM
  debugger interface before claiming instruction stepping through compiled code.

## Correctness contract

The reference is the current browser C++/WASM machine model, including its
documented limitations. Interpreter execution is not by itself proof of hardware
accuracy.

At each supported transition boundary, both modes must agree on observable
machine state and continuation: ARM/Thumb registers, CPSR and VFP state; active
and suspended thread contexts; mapped memory and aliases; kernel objects and
waiters; scheduler clocks and pending events; DSP state; PICA programs, uniforms
and registers; GX queues and completion state; graphics memory and display
selection. Preserve host input ordering separately from guest state.

Compiled blocks, decoded shaders, host pipelines and host texture objects are
derived caches. They are discardable only after any authoritative guest data
held in them has been materialized. Retain an explicit inventory of canonical
state versus derived caches; do not use a framebuffer hash as a substitute.

A transition must:

1. Acquire exclusive execution ownership and stop at a declared guest boundary.
2. Complete the host work needed to represent that boundary, without completing
   guest operations scheduled for a later boundary or advancing guest clocks.
3. Materialize registers and guest-visible GPU writes, including color, depth,
   stencil and transfer destinations; preserve operations that remain pending.
4. Rebind or invalidate derived resources and transfer execution ownership.
5. Publish the actual mode and boundary only after the operation succeeds.

Apply the same synchronization rules before CPU access to GPU-owned memory,
software fallback, save, load, memory inspection and detailed capture. Track
overlapping physical ranges and format aliases, not just texture object IDs.
Audit direct RAM views as well as bus helpers so no access bypasses coherence.

Unsupported accelerated operations fall back before their effects are committed.
The fallback sees coherent inputs and executes exactly once. A numerical mismatch
that reaches guest memory cannot be repaired merely by switching engines later.
Strict acceleration therefore needs reference-equivalent results for its enabled
operations. Diagnose uncertain paths in an isolated comparison mode and retain
software execution until they pass. Approximate rendering, higher resolutions
and fast-math are outside this prototype's continuity guarantee.

Portable states remain bound to media, configuration and a compatible execution
build. Extend execution identity to include behavior-bearing JavaScript/WGSL
assets as appropriate; the current WASM hash alone cannot identify a future
renderer implemented outside WASM. Do not weaken existing identity validation.
Save materialized machine state; restore it first in Reference and then honor
the selected acceleration preference when supported.

## Milestones

### M0 Plan and architecture

Deliver this plan and link it from the browser documentation. Choose the shared
session selector, define the continuity contract and distinguish the first
prototype from later CPU recompilation and historical reconstruction.

Acceptance: repository paths and existing capabilities checked; documentation
links and diff reviewed; commit and push this planning milestone to `main`.

### M1 Baselines and transition interface

Inventory canonical state and GPU memory access paths. Add a 3DS-specific
execution interface for capabilities, requested/effective mode, transition
status, quiescence and cache invalidation. Integrate it with worker ownership,
input ordering and the shared UI. Add the Execution mode selector; Experimental
remains disabled until a real backend is available. A test-only second reference
backend can exercise transitions without pretending to provide acceleration.

Rebaseline the current build using local Super Mario 3D Land boot, welcome and
interactive gameplay checkpoints, plus a Captain Toad scene if available. Record
unavailable fixtures explicitly. Use public synthetic workloads for reproducible
checks without private media.

Acceptance: no-op/test-backend round trips preserve canonical state and future
execution, including pending waits and input; normal save/load/capture still
work; cancellation and reset cannot receive stale transition completion. Record
baseline frame-time distribution, subsystem timings, memory and longest calls.

### M2 Isolated WebGPU graphics replay

Create a versioned replay input stream containing ordered operations, initial
state and immutable resource versions. Capture the actual dependencies before
the guest overwrites them. Keep detailed output-write provenance separate from
the inputs needed to execute a draw again. Store private streams locally only.

Build a browser replay harness independent of live CPU execution. Start with
fills, copies and display conversion, then representative textured triangles,
depth/stencil, blending and TEV combinations. Initially retain reference vertex
processing to isolate raster costs. Compare conventional WebGPU rasterization
with a small compute implementation where hardware coverage or arithmetic fails
the reference contract; choose based on correctness and measured end-to-end cost.

Acceptance: exact public transfer fixtures; per-target color/depth/stencil diffs
for draw fixtures; reported unsupported operations; complete replay of at least
one private graphics interval when available. Measure uploads, shader warmup,
submission, GPU work, readback and presentation separately. A partial or
numerically divergent result is evidence for narrowing the next milestone, not
permission to call the live backend equivalent.

Decision: proceed with the backend that preserves the contract and demonstrates
useful headroom. If conventional rasterization fails, prefer compute or narrower
exact acceleration; do not silently lower the correctness requirement.

### M3 Live GPU operations and coherent handoff

Implement a bounded command queue and resumable GX/renderer bridge. A suspended
operation retains its continuation and is not re-executed when the worker resumes.
GPU completion latency is host time; it must not become guest time. Preserve the
existing guest-visible command completion and interrupt ordering.

Integrate the operations proven in M2, initially fills/transfers and a narrow
draw subset. Track CPU/GPU ownership, dirty ranges and aliases; support ordered
software fallback and guest reads. Add GPU-resident Play presentation for both
screens while retaining the software inspection canvas. Enable Experimental only
when an accelerated operation genuinely runs, with coverage/fallback counters.

Acceptance: live Reference to Experimental to Reference at supported boundaries;
CPU reads after GPU writes, CPU writes before GPU sampling, overlapping copies,
save/load and input sequences match the reference. An artificially delayed GPU
must preserve guest event order. Queue pressure is bounded and yields to Pause.
Unsupported features run through the coherent reference path exactly once.

### M4 PICA coverage and measurable acceleration

Expand the selected backend using the workload census: textures and formats,
TEV, lighting, clipping, depth/stencil, blend modes and render-to-texture. Add
PICA vertex-program translation only after the raster/transfer path is stable.
Cache generated shaders and pipelines by all relevant program/state inputs.
Invalidate on program writes, uniform/resource changes where relevant and load.

Retain reference fallbacks. Report how much work is accelerated so improvement
cannot be attributed to skipped draws or missing features. Shader compilation
must remain asynchronous with bounded pending work and a correct fallback.

Acceptance: enabled paths pass exact output and machine-continuation comparisons
over the fixture corpus. Record an end-to-end improvement with matching guest
work; target at least 2x on a representative GPU-heavy scene before describing
the mode as faster. Treat real-time 60 display intervals/s as a stretch target.
Report a missed performance target honestly and profile the remaining cost.

### M5 Reference inspection and prototype acceptance

Connect Pause/Render to coherent handoff, the existing three-field reference
capture, pixel inspection and isolated replay. Resume back into the selected
Play mode after invalidating resources affected by reference execution or edits.
Reference stays the default; both available modes are selectable in the UI.

Exercise mode changes interleaved with pause/resume, touch and buttons, state
save/load, memory inspection, cancellation, repeated captures and reset. Use
hundreds of switches and at least 1,000 display intervals across the available
private scenes, with deterministic seeds and inputs. Check memory growth.

Test missing WebGPU, adapter creation failure, shader failure and device loss.
Maintain a recoverable materialized checkpoint and the history needed to replay
from it in Reference. If current GPU-only state cannot be recovered, stop and
offer the last valid checkpoint with its actual position; do not claim seamless
continuation after losing authoritative data. Bound recovery storage and cost.

Acceptance: switched runs match reference state and subsequent execution;
reference capture pixels and histories match their own displayed capture;
scrubbing does not mutate the live machine; saved states continue correctly in
either supported mode. UI tests verify progress, effective mode, both screens,
fallback messages and unavailable capabilities. Publish benchmark and support
matrix results with precise limitations. This completes the first graphics
acceleration prototype.

### M6 ARM debugger and bounded WASM recompiler

First expose reference ARM register/memory inspection, instruction stepping and
PC breakpoints through a platform-appropriate browser debugger adapter. Do not
assume the existing 6502-oriented UI accepts ARM addresses and register layouts.
Specify watchpoint behavior and exception/event boundaries in the same adapter.

Then compile a bounded subset of hot ARM/Thumb regions into WASM sharing canonical
memory. Unsupported instructions execute through the interpreter. Preserve
instruction budgets, guest clock accounting, scheduler boundaries, condition
flags and VFP behavior; cache keys include code content/mapping and ISA state.
Invalidate on executable writes, aliases and state load. Breakpoints split or
invalidate compiled regions; stepping and watched accesses initially use the
reference path. Compile in the background with bounded module/cache growth.

Acceptance: instruction/block differential tests cover ARM/Thumb transitions,
flags, memory boundaries, SVC, exceptions and self-modifying code. Breakpoints
stop before the requested instruction; reference single-step then compiled
resume has the same continuation. Repeat M5 switching tests with all four
CPU/renderer combinations. Profile again before expanding instruction coverage.

### M7 Inspect recently displayed frames

Add bounded checkpoints and graphics input history sufficient to reconstruct a
selected recently presented frame in a scratch reference machine. Retain older
resource producers when display buffering or render-to-texture spans frames.
CPU history additionally requires deterministic input and external-event logs.
Associate evidence with the reconstructed frame and compare it with the recorded
Play image. Explicitly report missing history, overflow and mismatches.

Acceptance: reconstruction remains valid after live textures are overwritten;
backward/forward inspection is repeatable and leaves live execution unchanged.
Measure recording overhead and memory. Keep M5 next-window capture available
when historical reconstruction is unavailable. This is an extension beyond the
first prototype, not a dependency for shipping M5.

## Validation and delivery

Use exact comparisons of a normalized canonical state inventory. Exclude only
documented host timing, instrumentation and derived caches; include all guest
state. Compare continued execution as well as the transition instant. Full
serialized-state hashes are useful where stable but must not mask missing fields
or replace comparisons after continuation. Compare experimental execution with
the browser reference backend. Native builds of the same browser C++ sources
may assist unit testing, but browser WASM/WebGPU runs establish acceptance;
the Go emulator is not a validation dependency.

Public checks should cover synthetic ARM programs, graphics commands, resource
aliasing, malformed streams and transition races. Private scene checks should
cover boot, title/welcome, active input and geometry, offscreen targets and recent
stencil/clipping regressions. Publish hashes and aggregate measurements, not
cartridges, private checkpoints or captured game resource streams.

Measure cold and warm runs separately, using identical initial state and guest
work, deterministic input and multiple sequential trials. Report display
intervals and presented images separately; include median/p95 interval cost,
transition/capture latency, CPU time, GPU timings where available, wall time,
uploads/readbacks, fallback coverage, compilation and peak memory. Distinguish
Node reference throughput from actual browser WebGPU measurements. Initial
acceptance uses a specified desktop Chromium/adapter; record actual Safari and
Firefox capability/results before claiming support there.

Implementation touchpoints are the C++ sources under
`tools/platform/n3ds/browser/core`, `site/emulators/worker.js`, the execution gate,
shared UI metadata/shell and a 3DS-specific renderer module. Limit any port
generator changes to preserving browser-specific integration hooks; do not
propagate accelerated behavior into the Go emulator. Keep console-specific logic
in these browser components rather than copying the shared application. Add
artifacts through the established build and release tooling.

For each implementation milestone:

1. Update this milestone's status and add a concise result with commands,
   correctness evidence, performance and unresolved limitations.
2. Run relevant tests of the browser C++ sources, WASM and WebGPU, plus the
   shared browser public checks. Rebuild the 3DS binary when necessary, package
   with `--core 3ds` (or `--site-only` for JS-only changes), and run the release
   audit. Review package-generated changes so unrelated source/build work cannot
   enter the milestone.
3. Stage only the milestone's source, evidence, documentation and required
   release artifacts. The working tree already contains unrelated changes.
4. Commit the completed milestone on `main`, push `origin main`, and verify the
   remote commit. Each push can trigger the existing site deployment. Preserve
   Reference as the default and keep unfinished paths capability-gated.
5. If `main` advances, integrate without overwriting unrelated work and repeat
   affected checks. Never force-push. Regressions use a follow-up fix or revert.

M0 changes documentation only: check links and whitespace, commit and push;
there is no reason to rebuild or republish emulator bundles for the plan itself.
