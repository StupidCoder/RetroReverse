# 3DS browser execution state

M1 defines the transition boundary as a completed browser `_rr_run` call. That
call returns after saving the active thread context and never leaves reference
rendering outstanding. The worker execution gate owns the machine throughout an
asynchronous mode transition. Cancelling the request prevents activation but
keeps ownership until any pending host quiescence completes. Reset replaces the
worker; session IDs prevent predecessor messages reaching the new machine.

## State inventory

The browser `core/state-fields.h` and `state.h` encode the following groups.
The M1 differential harness compares the complete serialized state, including
pending events, before/after selected transitions and after continuation.

| Group | Canonical data | Derived or host-only data |
| --- | --- | --- |
| ARM11 | R0–R15, CPSR/flags, architecture/Thumb state, VFP registers and control state, instruction counts, halt state | Host callbacks and bus pointers |
| Memory | Mapped region bases, contents and backing aliases, heaps, stacks and TLS | Page lookup table reconstructed from regions |
| Threads and Horizon | All thread contexts, current thread, priorities, waits/deadlines, handles, kernel objects, services, file/save positions and state | Closures rebound to the restored machine; immutable mounted media retained |
| Time and input | Instruction/guest tick counters, VBlank deadlines, DSP progression, HID rings and guest input latches | Host input queue, sequence and held controls remain in the worker; no transition releases or reapplies them |
| PICA | Registers, code/opdesc, uniforms, attributes, lights/LUTs, draw state and counters | Decoded shader epochs/cache, decoded textures, worker-pool objects |
| GX and display | Pending commands, GSP memory/events, completion state, color/depth/stencil bytes, transfer destinations, framebuffer selections and display counters | RGBA presentation conversion, captured evidence and replay scratch |
| DSP | Decoder/mixer/voice state, buffers, timing and pending guest events | Browser audio presentation, not currently exposed |

Future host GPU buffers may hold authoritative canonical bytes. Such bytes are
not discardable caches until materialized. Each accelerated operation must state
its exact physical ranges and read/write dependencies, including overlapping
views and transfers between linear memory and VRAM. Saving or inspecting must
wait for those dependencies without advancing the guest.

## Memory access audit

CPU byte/word accesses enter `n3ds_Machine_Read`, `ReadWord`, `Write` and
`WriteWord`; 16-bit adapters combine byte accesses. Also audit `ReadBytes`,
`directRange` and `copyRange`, which return or consume borrowed backing spans.
`regionOf` and indexed pages resolve guest addresses; GPU physical addresses
map into the same backing through the 3DS graphics-memory mapping helpers.

Raster targets use borrowed spans inside the triangle renderer. Texture decode
and vertex fetch can read mapped storage directly. GX fills and transfers write
guest memory, while command-list decode can chain to new command buffers. DSP,
Horizon IPC/file operations and the state loader also modify mapped regions.
The Memory workspace intentionally reads backing storage directly; it does not
execute CPU/device read handlers. All these paths must be coherent before M3
can leave data resident exclusively on the GPU.

Reference restore already rebinds bus/coprocessor/SVC callbacks, reconstructs
pages, invalidates decoded shaders and clears decoded textures. M3 extends
invalidation to host resources. M1 changes no core binaries, serialization or
guest execution semantics. Its test-only second reference engine has no caches
to invalidate and is never exposed as accelerated execution in production.

## M1 evidence

`tests/execution-3ds.mjs` covers availability, unknown modes, exclusive ownership,
preparation/quiescence order, failed preparation, cancellation and stale-session
completion. `tests/benchmark-3ds-acceleration.mjs` restores private browser
checkpoints, performs three sequential measured runs, checks identical complete
end-state hashes and compares 32 test-backend switches against uninterrupted
reference execution with identical button, circle-pad and touch inputs.

The browser acceptance harness verifies the selector and effective-renderer
description. Experimental is disabled until a real backend is integrated. A
local Chromium 154 adapter exposes WebGPU on Apple Metal 3; this establishes
availability for M2, not WebGPU rendering correctness or other-browser support.

Results are stored in `../results/3ds-acceleration-m1-*.json`. The welcome scene
is approximately 5.2 modeled display intervals/s; the independently checked
title/Mii-dialog scene approximately 4.4/s. These are Node/WASM execution-only
measurements, not browser canvas rates or gameplay claims. Cold boot is measured
separately. Existing inspected checkpoints did not establish an interactive
gameplay scene; that coverage remains explicitly outstanding for later acceptance.
Private media and checkpoints are not committed.

The full public `tools/browser/check.py` suite and release audit pass. Three
successive private reference captures also pass complete state-continuation,
pixel evidence and replay comparisons (`capture-wasm.mjs`); each capture spans
the existing three display fields. No renderer or core binary changes in M1.
