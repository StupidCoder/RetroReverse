# DOS software-renderer capture

Pause now traces writes into discovered RAM render buffers, as well as VGA,
palette and display registers. No game addresses, executable hashes or renderer
symbols are used for discovery. Quake's existing executable compatibility option
is unchanged and independent of this capture feature.

## Using it

1. Load a game folder, reach an interesting scene, then press Pause.
2. Scrub Rendering. **Reveal buffer writes** dims source pixels until a recorded
   RAM write touches them. This makes redraws visible even when a stationary
   scene writes the same colors as the previous frame. Uncheck it for the stored
   RAM colors without that dimming.
3. Select a pixel. The inspector reports its guest RAM address, ordered writes,
   actual x86 writer PCs, old/new bytes and instruction clock. **Show rendering
   step** jumps to the selected writer.
4. **Inspect final VGA copy** shows the presentation transfer. **Follow earlier
   buffer writes** follows a captured MOVS dependency at the time of its read.
   RAM-to-RAM copies expose the same link, so intermediate buffers can be followed.

The preview projects each final VGA copy's immediate RAM source into screen
coordinates. It is an educational view of buffer construction, not historical
monitor output. Pixels freeze when their final VGA copy is replayed, so subsequent
reuse of the RAM cannot change the displayed frame. Pixel queries describe the
final capture even while the preview is at an earlier step. No function names
are inferred: identifying a particle renderer still requires analysis of its PC.

## Implementation and bounds

Capture takes an initial RAM snapshot and temporarily records CPU stores to RAM
and modeled VGA storage. DOS wide-write fast paths fall back to observable byte
stores only during capture. MOVSB/MOVSW/MOVSD, including REP, attach source-byte
addresses and a read-time cutoff to their writes. Direction flags, segmented
16-bit offset wrapping, 20-bit bus wrapping and VGA plane selection retain the
CPU/device model's behavior. Real-mode PCs include CS rather than using the
protected-mode segment-base cache.

The window waits for two VGA write bursts, separated by at least 20,000 retired
instructions without VGA stores. This gives a renderer that was already partway
through drawing when Pause was pressed another opportunity to draw and present.
This is a heuristic, not VSync or a guarantee of exactly two game frames. A window
ends after 140 synthetic display intervals or when the raw record bound is hit.

At the end, discovery starts from the last recorded copy to each displayed VGA
byte and walks RAM-to-RAM MOVS edges backwards. Only writes to the contributing
RAM bytes survive in the replay, along with VGA/register/palette writes. Retained
RAM pages are packed separately from modeled VGA planes and metadata; the UI
translates them back to guest RAM addresses. Page padding is storage, not evidence
that every byte on that page was watched. Each CPU instruction contributing to
these bytes is a replay step; REP remains one instruction.

The raw limit is 8,388,608 records (256 MiB), the RAM-page limit is 16 MiB, and the
shared replay limits are 2,097,152 writes / 131,072 VGA events. Dropped records
mark evidence incomplete. Raw records and the full initial RAM snapshot are
released after compaction; WASM's committed heap does not shrink. In one browser
Quake test the reported heap reached 938 MiB, including temporary capture/state
allocations; retained evidence plus the two checkpoints was 140.8 MiB.

## Current limits

- Dependency edges cover literal MOVS transfers. General `MOV` register chains,
  LODS/STOS pairs, arithmetic conversions, SSE copies and VGA latch-copy ancestry
  are not propagated. Their actual destination stores are still recorded; if no
  literal source is known for a displayed pixel, the inspector shows VGA history.
- Multiple literal-copy buffers can be followed in source history. The preview
  projects the immediate final source; it does not recursively visualize every
  earlier compositing surface or infer transformed/computed dependencies.
- Capture cannot recover writes predating its initial snapshot. Unchanged HUD
  areas that were not copied during this window may have no discovered producer.
- This preserves the existing VGA model, including its mode-3-as-mode-0
  approximation. It is not a VGA accuracy fix. Existing Underworld display
  corruption outside the 3D view remains visible in the test checkpoint.

## Validation, 2026-09-29

The same private gameplay checkpoints were restored in Node/WASM. Full replay
matched the final display byte-for-byte. State restore and uninterrupted versus
captured execution matched CPU/RAM/display proofs, and scrubbing left the paused
machine unchanged.

| Game | RAM-backed displayed pixels | Rendering steps | Retained writes | Dropped records |
| --- | ---: | ---: | ---: | ---: |
| Quake | 48,640 | 78,483 | 102,499 | 0 |
| Ultima Underworld | 19,264 | 104,320 | 109,223 | 0 |

Quake exposed 45 distinct RAM-writer PCs in the native probe. In the browser,
pixel (160,90) mapped to RAM `0x029c75dc`, with a writer at `0x00111e28`, and its
final VGA copy was at `0x00110fe0`. That browser capture ran slightly later than
the deterministic checkpoint test and contained 86,978 steps; capture took
0.39 seconds. Source-history navigation, jump-to-write, partial reveal and the
final display were checked in the actual browser.

Normal Quake execution was unchanged within measurement noise: three alternating
fresh-process samples of 150 synthetic intervals had medians of 1,928 ms before
and 1,924 ms after this change. CPU/RAM/display proofs matched in all six samples.

Public native and actual-WASM tests cover direct stores, CS-adjusted PCs, planar
copies, independent and chained buffers, reversed REP, offset/bus wrapping,
post-copy reuse, masked/direct VGA fallbacks, equal-value redraws, timeout and
trace overflow. The complete public multi-platform suite also passed. Detailed
private-media counts and hashes are in `../results/dos-render-buffers.json`;
no game files, states or screenshots are published.

```sh
python3 tools/browser/check.py
python3 tools/browser/tests/check-x86-wasm.py --emcc /path/to/em++
node tools/browser/tests/console-check.mjs dos /local/quake.exe /local/game.state
```
