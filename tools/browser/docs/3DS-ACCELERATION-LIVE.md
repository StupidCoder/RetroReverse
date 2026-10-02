# M3: live compute bridge

Experimental is now selectable after a successful adapter, shader and actual GPU
transfer self-test. Reference remains the default. The selector identifies the
renderer and reports GPU operations and reference fallbacks separately. Switching
while playing waits for the current call and continues playing; Pause can cancel
a transition and prevents automatic resumption.

## Boundary and memory ownership

The browser 3DS build uses Emscripten Asyncify around the GX continuation. It
awaits one isolated operation, installs the complete output, and resumes the same
C++ call. It does not restart a GX command, advance a guest clock for host latency,
or publish its completion interrupt early. Smaller operations, overlapping
physical backing, unsupported draws and instrumented accesses use Reference.

M3 deliberately uses **eager materialization**: every GPU result returns to RAM
before the machine continues. This is more conservative than deferred dirty-range
ownership and has a measurable readback cost. It covers CPU accesses, physical
aliases, direct borrowed spans, texture decoding, DSP and IPC without adding a
partially audited collection of memory barriers. GPU buffers remain scratch;
there is never an authoritative guest byte that exists only on the host GPU.
The queue is bounded to one operation, with a five-second completion timeout.

The worker holds execution ownership until the suspended call completes.
Inspection and transition messages cannot call into suspended WASM. Input queues
can receive messages, but guest input is applied only at the usual slice boundary.
The old pump's generation prevents it resuming after Pause or a newer request.

Validation failure, device loss or timeout discards the uncommitted result. The
same operation then executes once in Reference against unchanged inputs. Previously
completed GPU writes are already in RAM. The worker returns to Reference at the
completed call boundary and reports the failure. This design needs no rollback
history to recover lost GPU-only state because such state is never retained.

## Presentation and states

Experimental Play uploads canonical scanout bytes and decodes/rotates both screens
on the GPU. A transferred ImageBitmap presents the result without a full RGBA
readback. The software canvas remains available for Render inspection. This is
GPU presentation, not persistent GPU ownership of guest render targets. Replaced,
stale and consumed bitmaps are explicitly closed.

State identity hashes the WASM build together with the transfer shader, live
bridge and presentation modules. A state from a different rendering build is
rejected by the existing container validation. Derived GPU buffers are overwritten
from canonical inputs for every operation; restoring a state cannot reuse stale
texture or target data from a previous machine.

## Validation

Private welcome and title/Mii-dialog runs compare entire serialized end states
after Reference, Experimental and Experimental with three milliseconds of added
host latency per operation. All match. The welcome run executes 22 GPU operations,
the title run 27; general PICA draws remain software fallbacks. Both screens match
the reference framebuffer byte for byte after GPU presentation.

The short six-interval trials improve execution time by approximately 15% versus
the same Asyncify build in Reference. This is preliminary evidence, not the M4
performance target or a gameplay claim. Asyncify also adds cost to the reference
build; comparisons with the pre-bridge baseline must account for that overhead.

The real UI test loads local media, enables Experimental, switches to Reference
while playing, checks continued execution and pauses at a safe boundary. Unit
tests cover adapter failure, exclusive GPU work, quiescence, timeouts and device
loss. M5 adds sustained switching, state-file/capture and recovery acceptance.
Results are `../results/3ds-acceleration-m3-*.json`.
