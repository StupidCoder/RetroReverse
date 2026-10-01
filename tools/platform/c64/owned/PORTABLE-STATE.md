# Owned C64 portable checkpoints

This is the hardware-state codec used by the C6 development browser adapter.
The browser adds a separate version-2 envelope documented in
[browser/README.md](browser/README.md). Owned prepared recipes have distinct
identities; existing production checkpoints remain incompatible and unchanged.

## API and host responsibilities

`state.h` exports `saveState` and `loadState` overloads for a tape-only `Board`
and a dual-machine `System`. Saves return bytes; loads return success and an
error string. Loads parse into a temporary object, validate it, then commit
with nonthrowing assignments. A failure leaves the running machine untouched.
Saving an invalid state throws. Emscripten builds must enable C++ exception
catching (`-fexceptions`, as in `check.py`).

The host supplies `StateIdentity`: SHA-256 digests of the core build, BASIC,
KERNAL, character ROM, drive ROM, original TAP and original disk image, plus
a configuration identifier. Zero denotes an absent input. Hash the actual
inputs; do not trust identity fields supplied by the checkpoint. Compare
against the currently loaded session's identities. Configuration must distinguish
execution-affecting choices such as drive presence and initial write protection.
Dirty disk contents are state; the disk identity continues to identify its
original source image.

The codec does not compute these SHA-256 hashes or establish authenticity.
Prepared lesson packages must additionally pin the complete checkpoint's hash,
as the existing browser recipe protocol already does. CRC32 detects accidental
corruption but is not a signature. The test harness uses explicit fixture
identities for mismatch checks; real-media test inputs are hashed by `check.py`.

ROMs and decoded tape pulses stay outside the snapshot. A restore retains the
host's existing copies; it validates the tape position and remaining duration
against the attached pulse vector. If a host edits decoded tape pulses, it must
bind those edited inputs to a distinct identity; pulse edits are not restored
by this codec. Disk track contents, speed zones and dirty
flags are included, so a restore can undo local disk writes. No source image
on the host filesystem is modified.

After a successful load, a browser adapter must pause, invalidate stale
inspection history, clear temporary breakpoint bypass/step state, and publish
fresh snapshots. This codec deliberately does not persist debugger/UI state.
For the current owned `Debugger`, call `restore(system.save())` after a successful
load to apply its existing pause/history semantics while retaining breakpoints.

## Format version 1

All integers are little-endian, booleans are one byte (only 0 or 1), and enum
widths are explicitly declared in the hardware headers. There are no pointers,
padding bytes, native `size_t` fields or WASM heap images.

| Field | Encoding |
| --- | --- |
| Archive magic | u32 `0x53525231` |
| Owned-core platform tag | u32 `0x4f343643` |
| Format version | u32 `1` |
| Machine kind | u8: `0` board, `1` board + 1541 + IEC scheduler |
| Identity | Seven 32-byte digests in API order, then u32 configuration |
| Hardware | Explicit field order in `state/fields.h` |
| Corruption check | u32 IEEE CRC32 over every preceding byte |

Each disk half-track has a u32 byte count and raw bytes, followed by a u32
speed-map count and one speed-zone byte per track byte. Counts must match and
must not exceed the G64 v0 65,535-byte limit. Fixed arrays have their declared
length; the disk track count is explicitly u32 even on different native ABIs.
The maximum checkpoint size is 12 MiB, enough for all 84 maximum-sized raw
half-tracks and speed maps. The ordinary authored D64 test checkpoint is
733,405 bytes; the browser's separate prepared-start size limit still applies.

Hardware fields include pending CPU transactions and interrupt samples, all
CIA/VIA pipelines, VIC framebuffer/fetch latches/shifters, SID oscillators and
noise, keys/joysticks, tape phase, rotating drive mechanics and the IEC samples
retained across peer clock edges. Adding or reordering fields requires a format
version change. Emulator behavior changes also require a new host core identity.
Old third-party-core states are never interchangeable with this format.

Reads reject wrong versions/types/identities, truncation, trailing data, invalid
booleans/enums, unsafe device indices, inconsistent disk maps and out-of-range
timing phases. Disk sizes are checked before allocation. Validation is not a
proof that an arbitrary combination of registers could arise on real hardware.

## Acceptance

`tests/state.cpp` exercises thirteen individually ordered clock-edge positions,
partial CPU execution, tape/graphics/SID continuation, in-flight disk writes,
G64 half-tracks with varying speed zones, and failed-load atomicity. It compares
the complete serialized end state, not only RAM or a screenshot. The WASM
executable imports a checkpoint produced by the native executable, verifies
byte identity, and reproduces its continuation. Native UBSan runs the same cases.

The authentic Fort test now restores serialized checkpoints during Novaload
and gameplay. Giana's custom-loader investigation restores both machines from
serialized state and checks its IEC trace and payload again. Elite independently
boots through the real KERNAL to `$0378`, checks all 52 stores in the initial
`$0300–$0333` vector block against raw tape pulses, and repeats that block after
restore. None of these tests injects expected game bytes into the machine.

C6 browser ABI, rendering provenance, cancellation/recording and regenerated
prepared lessons are validated separately in [the compatibility ledger](COMPATIBILITY.md).
Later Elite loader stages remain unvalidated; the initial-block gate is bounded.
