# U4 — portable machine states

Implemented September 28, 2026; browser download coverage limits are noted below.

States use explicit field-wise little-endian serialization, version 1 per core.
No native addresses, struct padding or WASM memory dump is stored. The container
has a SHA-256 integrity check over metadata and payload. Metadata binds the file
to the exact WASM build, selected media files (incrementally hashed), firmware
hashes and compatibility settings. Immutable cartridge/disc/firmware data stays
outside; 3DO open disc streams restore by name and offset against mounted media.

C64 serializes chip pipelines/registers, RAM/color RAM, framebuffer, keyboard,
audio chip state and tape timing. The initialized worker supplies callbacks,
ROMs and memory maps. PS1 includes delayed CPU operations, GTE/GPU transfers,
VRAM, CD response/read queues, BIOS HLE state and interrupts. N64 includes CPU,
TLBs, RSP, RDP/TMEM/pending commands, device registers, clocks and EEPROM; CPU
bus links and shared RSP memories are rebound. 3DO includes HLE tasks, heaps,
items and shared item identities, streams, signals, NVRAM and resumable scheduler
locals; callback closures are reconstructed.

The UI loads into a separate worker; the current paused worker is retained until
media/configuration/integrity checks and core deserialization succeed. A failed
candidate is discarded. File imports are bounded to 129 MiB, metadata to 1 MiB,
core allocations to 128 MiB and graph objects to 100,000. Core build changes
intentionally invalidate old state files; migration is not implemented.

Saved input queues are encoded. Interactive restore releases host-held buttons
and pending physical-key events, leaving the machine paused; deterministic core
continuation tests restore exact device state without that UI safety release.
Debugger history and wall-time profiling are host observations and are reset.

## Recorded checks

- Native PS1 at 10k, 10M, 400M and 500M instructions: advance 37 different-sized
  slices after restore, then compare the entire serialized machine byte-for-byte.
- Native N64 at 10k, 10M, 150M and 300M instructions: same full-state continuation
  comparison, including RSP/RDP execution.
- Native 3DO at startup and displays 1, 60, 300 and 900: shared objects, stream
  reconstruction and scheduler continuation compare byte-for-byte.
- Native C64: four continuation checkpoints with active keyboard/tape state.
- WASM on Node: all four cores at three checkpoints; continuation matches and a
  truncated import is rejected without changing the original serialized state.
- Container tests reject corruption in magic, metadata, payload and digest,
  along with truncation.

Observed uncompressed core sizes: C64 230,067 bytes; PS1 about 3.16 MB; N64 about
4.21 MB; 3DO about 7.34–10.04 MB on tested paths. Initial WASM restore calls on the
reference Mac: C64 1–2 ms, PS1 12–13 ms, N64 19–20 ms, 3DO 34–44 ms. These exclude
full-file hashing, worker startup and browser presentation, and are not maxima.

Native-to-WASM imports reproduce identical serialized bytes on all four cores,
including the late PS1/N64/3DO checkpoints. In the in-app browser, a fresh C64
page restored a state with only the TAP selected: hosted ROMs loaded automatically.
PS1 state import succeeded. A corrupted C64 state was rejected while preserving
the paused machine. Save creates an explicit download link as well as requesting
the download. Download-link activation was exercised. The in-app
browser did not emit a download event despite the UI requesting a download;
Chrome automation file selection requires the extension's file-URL permission.
