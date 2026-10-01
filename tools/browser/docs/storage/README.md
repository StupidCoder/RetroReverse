# Loader and Storage

The Loader preset has four equal quadrants:

| | Left | Right |
|---|---|---|
| Top | Game | Session / controls |
| Bottom | Storage | Memory atlas |

Existing shipped Loader layouts migrate to this arrangement; customized layouts
retain their geometry. Old Tape pulses instances become Storage instances. The
prepared Elite investigation uses a separate **Guided tour** preset so a tour
never overwrites the Loader preset.

Storage appears in every system's content picker. It keeps independent image,
view, sector and expanded-folder selections in each instance. Local media is
inspected with bounded `File.slice` reads, without invoking the emulator. Select
**Inspect storage** in the Load game dialog to browse a file without booting it
or replacing the current game; this also works for D64, which the C64 execution
core does not currently load. Normal loading also makes the selected image
available in Storage.

## Supported views

- C64 TAP: existing live pulse lengths and follow-head mode, only for the actual
  tape loaded into the current machine. A different selected tape cannot borrow
  that machine's live pulse stream; it remains available as raw bytes until loaded.
- C64 D64: 35/40/42-track geometry, with optional appended error bytes; track/sector
  navigation and directory files. Byte sizes come from bounded data-sector chains.
  Sector link bytes are excluded from file sizes. PETSCII letters are mapped;
  unsupported filename glyphs are displayed as replacement characters.
- Amiga ADF: DD/HD sectors identified by cylinder, side and sector; lazy AmigaDOS
  OFS/FFS directory trees. Custom loader disks without DOS headers retain sector
  inspection. Selecting a file navigates to its header block, explicitly labelled.
- Optical images: recognized ISO 9660 primary-volume filesystems in cooked 2048,
  raw 2352/2448 and stripped 2336-byte layouts. Shows original sector bytes,
  logical payload offset, directories, file sizes and file-to-sector navigation.
  A supported single-data-track CUE is normalized using the existing media reader.
- Nintendo 3DS: NCSD partition hierarchy or standalone NCCH, decrypted ExeFS and
  RomFS directories/files. IVFC layout offsets and metadata are checked; the hash
  tree is not authenticated. Encrypted partitions are identified and remain opaque.
- Selected local folders: preserves supplied relative paths and file sizes.
- Other formats: raw 256-byte blocks and selected-file metadata. No fabricated
  filesystem hierarchy or inferred sector geometry.

File sizes use decimal KB, MB and GB with exactly one decimal digit; tooltips
retain exact bytes. Directory sizes are not guessed. Expansion reads metadata on
demand. Readers reject out-of-bounds extents, linked-list cycles and excessive
metadata/depth. Limits are 2 MiB per metadata read, 1 MiB per ISO directory,
4,096 entries/links per directory and 20,000 displayed nodes per panel.

Filesystem readers are intentionally format-specific. UDF, Joliet/Rock Ridge,
ISO multi-extent/interleaved files, OperaFS, GameCube FST, XDVDFS and other custom
filesystems are not yet implemented in this browser panel. Compressed images and
encrypted 3DS contents are not decoded. Raw media views represent the original
selected image, not live drive access or guest-written sectors. Per-file export
is deliberately deferred; the current file click navigates to source metadata or
the first data sector, rather than downloading anything.

## Sources and acceptance

ISO, AmigaDOS and 3DS readers follow the repository's `tools/lib/iso9660`,
`tools/platform/amiga/adf` and `tools/platform/n3ds` parsers. D64 geometry follows
the [VICE format documentation](https://vice-emu.sourceforge.io/vice_17.html).

`node tools/browser/tests/storage.mjs` covers sizes, D64 file chains and cycles,
ADF nested entries, cooked/raw ISO, decrypted/encrypted 3DS and local paths.
`viewport-model.mjs` also checks migration of old presets and custom tape panes.

The browser viewport fixture checks actual quadrant positions, live tape pulses,
D64 inspection without executing/replacing the current game, and ordinary
viewport behavior. `storage-browser.html` checks lazy tree expansion, sizes,
file-to-sector links and independent panel selections.

Private-media validation uses `storage-private.mjs <platform> <path>` with a
bounded file-backed reader. Captain Toad: 3,834 entries, 87 directories, 305,742
bytes read from a 536,870,912-byte image. Jak and Daxter: 347 entries in ten
directories. Marble Madness: 54 entries in four directories. The reports include
counts and read budgets; no image or extracted file bytes are committed.

Final release `0f174857fac728f111a2` passed viewport acceptance in Chrome and
Firefox, the tree UI fixture, the loaded-tape identity check and the prepared
Elite guided-tour fixture. Source syntax, storage/model/shell tests and the
100-asset release audit also pass. Emulator core binaries are unchanged.
Safari was not run for this change.

### Live tape refresh correction

The shared inspection sampler now keeps its pending 200 ms deadline when game
frames arrive. Previously each frame restarted that deadline, starving live
inspection during continuous playback. Hiding all inspectors cancels sampling;
showing one or loading another game restarts it. Tape inspection also shows an
explicit pulse-position progress bar, pulse count/percentage and sampled motor
state. The percentage measures pulse count, not estimated loading time. Follow
head recenters immediately when enabled, including while paused. Repeating pilot
pulses can still look identical; the counter and position bar show advancement.

`node tools/browser/tests/inspection-feed.mjs` exercises continuous frame traffic,
rate limiting and hide/reset lifecycles with a deterministic clock.
`tools/browser/tests/tape-live-browser.html`, served with `serve-viewport.py`,
requires the private Fort Apocalypse TAP at its repository-local path. It builds
a loading checkpoint in memory from the existing boot recipe and tests the
packaged app while the actual tape loads. No checkpoint or game bytes are saved.
Release `8abb1aa0986735222938` passed in Chrome 154 and Firefox 157: repeated
live samples, followed scrolling, fixed manual window with advancing head,
paused stability, immediate re-follow, pulse progress and live RAM atlas.
