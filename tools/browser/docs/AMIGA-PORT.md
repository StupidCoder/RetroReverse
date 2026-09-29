# Amiga 500 browser emulator

The new `tools/platform/amiga/browser` core boots the repository's original
Marble Madness and Turrican ADFs through Kickstart and their own loaders. Both
reach their first playable level. The ADFs are not patched, extracted into an
artificial boot environment, or identified by filename/hash inside the core.
Other standard ADFs can be selected, with experimental compatibility.

## Implementation

This is a new C++ chipset model using the permissively licensed
[Musashi 68000](https://github.com/kstenerud/Musashi) CPU, pinned at
`313ebf1bd9f4d0d93341eb5ce21fd8a119e9dbdd`. It is not a UAE port. Vendored notices
are retained; the browser also serves the Musashi license. Generated opcode
sources are included so builds do not need a code generator or network access.

The machine models a PAL 68000 at 7,093,790 Hz, 454 CPU cycles per line and 312
lines per frame (about 50.08 Hz), 512 KiB chip RAM and 512 KiB slow RAM. It includes:

- Kickstart overlay switching, exception vectors and interrupt priority.
- Both 8520 CIAs: ports, timers, binary TOD, interrupt latches and keyboard bytes.
- DF0 select/motor/side/step signals, standard ADF-to-MFM track encoding, sync
  detection, double-armed DSKLEN, DMA and disk interrupts.
- Copper MOVE/WAIT/SKIP, including the vertical-wrap wait used by PAL lists.
- Blitter Boolean minterms, source shifts, masks, modulos, descending operation,
  area fill and an approximate line mode.
- OCS bitplanes, lores/hires, scroll/modulo, dual playfields, EHB/HAM6, eight
  sprites, attached sprites and priority decisions.
- Audio DMA address/count/interrupt progression, without sound synthesis/output.

The existing research ROM named `kick13.rom` actually identifies as **Kickstart
1.2, version 34.5**. The hosted copy is correctly named `kick12.rom`; its SHA-256
is in `site/emulators/firmware/amiga/manifest.json`. A local 256/512 KiB override
is optional. ROM and disk identities remain outside the portable machine state.

Hardware behavior was checked against Commodore's
[Amiga Hardware Reference Manual, third edition](https://www.ikod.se/wp-content/uploads/2020/08/Amiga_Hardware_Reference_Manual_3rd_Edition.pdf).
The original reverse-engineering notes in the two game directories describe
the loaders and provide useful checkpoints.

## Browser workflow

The landing page links to `/emulators/amiga/` and includes a generated Amiga 500
illustration. Its exact prompt and generator are recorded in `ARTWORK.json`.
The route uses the shared static WASM worker, local-file picker, transports,
performance counters, state container, pixel inspector and rendering replay.
There is no server component and no game upload.

Marble Madness starts in Workbench. Double-click the disk icon, then the game
icon inside the window. Once its own menu appears, select GO. The default rear
port joystick uses arrows and Space. For the repository's Turrican disk, click
the left mouse button to leave the TRSI intro, wait for the original decompressor,
then press Fire at the title screen.

Mouse motion is relative. The speed selector defaults to 4× for Workbench;
1× suits game menus that consume mouse counts directly. The host pointer is
hidden over a running display, leaving the emulated pointer visible. Left/right
mouse buttons also have on-screen equivalents. Mouse edges are serialized in
emulated time so a quick double-click cannot collapse into one press. Motion is
spread across frames to avoid wrapping the eight-bit counters in one poll.
Keyboard position codes and joystick/gamepad controls are also available.

## Inspection and states

The Amiga now has separate Play and Render workspaces, with **Scanlines & Copper**
and **Blitter & masks** inspection lenses. Pause captures three modeled PAL
intervals to include producers for buffered graphics. Pixels lead to historical
bitplane/palette/sprite words and their CPU, Copper or blitter writers. Blits expose
actual A/B/C inputs after masking/shifting, D results and destination previews.
See [Amiga raster inspection](AMIGA-RASTER-INSPECTION.md) for the workflow,
accuracy scope, capture limits and current validation.

Capture address spaces remain chip RAM at native addresses, slow RAM compacted
to `0x80000`, custom words at `0x100000`, and 640×256 RGBA scanout at `0x110000`.
UI register addresses use `$DFFxxx`. Historical byte values use the shared
little-endian packing, while Amiga RAM/register words are big-endian.

States serialize explicit CPU registers/prefetch plus RAM, video buffers, CIA,
Copper, blitter, disk/rotation and input state. No C++ pointers or struct padding
are persisted. Raw payloads are approximately 2.3 MiB; version 2 also preserves pending palette changes; the usual
`.rrstate` wrapper enforces media/firmware/core identity and integrity. Invalid
loads retain the current machine via the shared candidate-worker flow.

## Original port validation

For current raster-workspace tests and performance, see
[Amiga raster inspection](AMIGA-RASTER-INSPECTION.md). The measurements below
refer to the original single-frame capture.

`python3 tools/platform/amiga/browser/build.py --native-only --test` runs public,
media-free checks for 68000 execution, overlay/RAM mapping, CIA timing and keyboard
encoding, all eleven MFM sectors/checksums, floppy DMA arming, all 256 blitter
minterms, cross-word shifting, descending copying, Copper's line-255 wrap,
sprite priority, captured memory reconstruction, replay and state continuation.
It is included in `tools/browser/check.py` alongside the existing nine cores.

The optional private-media `tests/validate-amiga.mjs` restores a native gameplay
checkpoint, compares 120 frames of WASM continuation byte-for-byte with native,
checks that tracing/replay leave continuation unchanged, inspects a pixel grid
and its historical sources, and verifies that joystick movement changes output.
No game data or checkpoints are published. On the two reference scenes:

| Check | Marble Madness | Turrican |
|---|---:|---:|
| Captured writes | 340,109 | 335,424 |
| Capture overflow | 0 | 0 |
| Sampled pixels | 88 | 88 |
| Historical source checks | 357 | 383 |
| Replay steps | 3,209 | 1,113 |
| Evidence memory | 28.5 MiB | 28.2 MiB |

Both produce identical native/WASM state bytes after 120 gameplay frames. The
browser was also exercised through a cold Turrican boot, intro dismissal, title,
first level, pause, pixel/source inspection and partial rendering replay.
Marble Madness was launched from Workbench using actual browser double-clicks,
and a gameplay state was imported, resumed and captured in the browser. Normal
mode measured 49.7 emulated updates/s and 50.2 presented frames/s (99% PAL speed).

Initial measurements on the development Mac, without capture: Node/V8 WASM
roughly 280–322 frames/s for Marble Madness and 250–279 for Turrican. These short
120-frame measurements establish substantial headroom over PAL; they are not a
cross-machine or cycle-accuracy claim. Turrican in Chromium fast-forward reached
about 275 emulated frames/s while presenting about 53 frames/s. Detailed capture
was around 60 ms and used about 93 MiB of WASM memory including replay; ordinary
execution used 64 MiB. CPU/chipset scheduling, scanout, Copper, blitter and MFM
encoding have separate exclusive wall-time buckets.

## Limits

This is a useful initial OCS emulator, not UAE-level compatibility. Bitplane rendering is
sampled per scanline, with palette writes retaining modeled horizontal positions; DMA bus arbitration, blitter completion time, keyboard
serial handshake and several chip edge cases are approximate. It is not suitable
as a cycle-exact oracle. Interlace, collision registers, precise mid-line effects,
ECS/AGA extensions and audio output remain incomplete or unimplemented.
Only one standard 880 KiB ADF is supported, mounted read-only in DF0; disk writes,
extended ADF/IPF, hard disks and disk swapping are not implemented. Browser save
states can preserve progress without modifying the disk. Safari/Firefox and
physical gamepad acceptance for this core remain untested.

Build the native and WASM versions with:

```sh
python3 tools/platform/amiga/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py
python3 tools/browser/check.py
```

The static release retains the prior published executable bundle as usual.
