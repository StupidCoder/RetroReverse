# U2 — Local media and boot paths

Implemented September 28, 2026. All four adapters accept unknown images through
structural validation, without a game whitelist. This is **not** a claim of broad
game compatibility. Only the four reference titles are currently demonstrated;
additional commercial PS1/N64/3DO images are not available in this checkout.

- C64: generic PAL TAP v0/v1 import, hosted verified firmware or local overrides,
  ordinary keyboard LOAD/RUN and tape transport. No Fort Apocalypse schedule.
- PS1: ISO and raw Mode 1/2 data tracks, 64-sector cache, on-demand local file
  reads up to 4 GiB. CUE validates every named companion, one data track and its
  INDEX 01. Separate audio files may be selected but CD audio is not emulated.
  Combined audio/data BINs and explicit gaps are rejected rather than guessed.
  `HookEntryInt` now restores a guest jmp_buf, with setjmp/longjmp and
  ReturnFromException support. Ridge Racer no longer needs its hardcoded ISR.
- N64: three byte orders, region boot flag, CIC 6102 and libdragon IPL3 accepted.
  Other IPL3 variants give a specific error rather than using the wrong seed.
- 3DO: bounded cached local File slices. Generic Opera LaunchMe AIF discovery;
  no unconditional NFS VBL mirror or input gate in the shared app. Full incremental SHA-256
  enables and names that compatibility profile only for the original NFS image.
  The user can disable it before loading. It includes the inherited VBL mirror
  and queues early controller input until display 300 for this image only.
  HLE is still incomplete; arbitrary executables/other OS versions may fail.

Validation: SHA-256 against Node crypto at padding/chunk boundaries; CUE positive
and missing/ambiguous companion cases (`tools/browser/tests/media.mjs`). Native
Ridge Racer generic boot ran 580 million instructions with scripted input and
rendered a race at that checkpoint. The shared browser PS1 page uses the same
boot path with random-access reads. All three changed C++ cores build natively
and as WASM; packaged hashes are regenerated.

Unmet broader-compatibility acceptance: a second useful commercial scene per
platform has not been demonstrated. Do not describe U2 as establishing universal
compatibility. Execution errors are distinct from successful media import.

PS1 ABI reference: https://psx-spx.consoledev.net/ps1/kernelbios/misc-functions/
and https://psx-spx.consoledev.net/ps1/kernelbios/interrupt-exception-handling/.

U3 follow-up: Elite's 801,536-pulse PAL tape was additionally imported and run
from reset with ordinary LOAD/RUN keyboard input. This exposed and removed the
old 524,288-pulse cap (the existing 2 MiB TAP-file limit remains). The tape
finishes but the resulting display is corrupted, so Elite is **not** a passing
compatibility result. State observations: `../results/u2-elite.jsonl`.
