# The Need for Speed (3DO) — reverse engineering notes

The 3DO Interactive Multiplayer is the repo's first 3DO project. Its CPU is an
**ARM60 (ARMv3, big-endian)** and it is CD-based, so the toolchain reuses the
shape of the ARM work (`tools/arm`, now joined by a dedicated big-endian
`tools/arm60`) and the CD-oracle work (`tools/psx`). Everything here is derived
from the disc image plus public 3DO platform documentation — never from
game-specific sources (clean-room rule).

- Image: `Need for Speed.bin` — 771,408,960 bytes, MD5
  `b213789b67a3368207a2ebf2c222847a`, single `MODE1_RAW` track (2352-byte
  sectors → 327,980 sectors). Not committed (771 MB); pinned in the top README.

Tooling lives in `tools/threedo` (platform package) and `tools/arm60` (CPU);
per-game commands in `extract/` (module `needforspeed/extract`).

---

## Part I — the disc: the Opera file system

3DO discs use the **Opera** file system, not ISO 9660. The low layer is the same
as any CD (`tools/psx/cd.go`): the data track is raw 2352-byte sectors whose
2048-byte user area we expose through a `block()` accessor. On top of that sits
Opera, decoded in `tools/threedo/operafs.go`. Three things distinguish it:

1. **Big-endian.** The ARM60 runs big-endian; every Opera field is a plain
   big-endian word (ISO stores numbers "both-endian").
2. **Volume label at block 0**, identified by record byte `0x01` + five `0x5A`
   sync bytes + version `0x01` (ISO's "CD001" descriptor is at block 16).
3. **Avatars.** Every file and directory exists as one or more identical copies
   scattered across the disc, to cut seek time and add redundancy. A directory
   entry lists all avatar block numbers; we read avatar 0.

### On-disc structures (confirmed byte-for-byte against this disc)

**Volume label** (block 0):

| off | field | value on this disc |
|----:|-------|--------------------|
| 0x00 | record type = 1 | |
| 0x01 | 5× `0x5A` sync | |
| 0x06 | version = 1 | |
| 0x28 | volume label (32) | `"CD-ROM"` |
| 0x48 | volume id (u32) | `0x06DE0DC2` |
| 0x4C | block size (u32) | `0x800` = 2048 |
| 0x50 | block count (u32) | |
| 0x58 | root dir blocks (u32) | 1 |
| 0x60 | last root-copy index (u32) | copies = index + 1 |
| 0x64 | root-copy block numbers | root dir at block `0x3551F` |

**Directory block header** (20 bytes): `next(s32)`, `prev(s32)`, `flags(u32)`,
`endOffset(u32)`, `firstEntryOffset(u32)=20`. `next`/`prev` are **relative** block
indices within the directory (start block + index), not absolute; `-1` ends the
chain. `endOffset` is the first-free byte, and bounds the entries in the block —
entries carry **no** per-entry terminator flags on this disc.

**Directory entry** (variable): `flags(u32)` (low byte: `0x02` file, `0x06`
special, `0x07` dir), `id(u32)`, `type[4]` (e.g. `Bigd`, `Wrap`, `Tire`),
`blockSize`, `byteCount`, `blockCount`, `burst`, `gap`, `filename[32]`,
`lastAvatarIndex(u32)`, then `lastAvatarIndex+1` avatar block numbers. Avatar
block numbers are **absolute** — bulk file data is pooled near the front of the
disc (blocks ~1–800) while the directory tree sits near the back (~block 218k).

Two subtleties caught by reading the real bytes: the copy count is stored as the
**highest index** (so `+1` for the count), and the directory `next` field is a
**relative** index, not an absolute block — an absolute reading walked off into a
bogus block until both were corrected.

### What's on the disc

846 files, 31 directories. Root layout:

- `LaunchMe` (296,720 bytes) — the boot ARM executable (the M4 target). The
  `AppStartup` script confirms it: it ends `… the system will start $boot/Launchme`.
- `frontovl` (120,100 bytes) — front-end overlay.
- `DriveData/` — car and physics data. `CarData/` holds per-car assets keyed by a
  4-letter tag (`ANSX`, `CAR1`, `COP1`, `CZR1`, …): `*.carSlideArt*` (153,688 B
  each), `*.dashFam`, `*.WrapFam`, `*.TireF/R`, and text tunables like `ANSX.TDDyn`
  (`#Mass 1 / #MomentOfInertia 1.4 / …`) and `AudioTweak.data`.
- `FrontEnd/`, `Movies/`, `System/`, plus `signatures`, `rom_tags`, `Disc label`.

### Tools

- `tools/threedo/operafs.go` — `Open`, `ReadDir`, `Walk`, `ReadFile`, `Entry`.
- `tools/threedo/cmd/operainfo` — `-label`, `-ls`, `-extract PATH -o FILE`.

```
go run ./threedo/cmd/operainfo -ls   "Need for Speed (3DO)/Need for Speed.bin"
go run ./threedo/cmd/operainfo -extract LaunchMe -o LaunchMe "…/Need for Speed.bin"
```

Verified: `operafs_test.go` round-trips a synthetic image (no dependency on the
committed-out disc); on the real disc, extraction is byte-exact (`ANSX.TDDyn` →
1143 bytes of readable physics text; `AppStartup` → the 160-byte startup script).

---

## Part II — image assets: cels and SHPM shapes

Two image formats carry the disc's 2D art. Both decode in `tools/threedo` and
render through `cmd/celdump` (`-image DISC -path P -o out.png`, or `-all` to
batch every `.cel`/`.3sh` into a directory).

### 3DO cels (`.cel`) — `cel.go`

The standard Madam cel-engine format: IFF-like chunks (`CCB `, optional `PLUT`,
`XTRA` metadata, `PDAT`), each an 8-byte header (4CC tag + big-endian size).
The `CCB` chunk holds a `ccbversion` word, `ccb_Flags`, and — at payload
offsets 0x38/0x3C/0x40/0x44 — the two preamble words and the width/height:

- **PRE0**: bits 2:0 = bpp code (`3`→4bpp, `6`→16bpp seen here), bits 15:6 =
  height−1 (validated: 239 ⇒ 240).
- **PRE1**: bits 10:0 = width−1 (validated: 319 ⇒ 320).
- **`CCB_PACKED`** (flags bit 9) selects packed vs unpacked source data.

*Coded* cels (bpp ≤ 8) index a **PLUT** of big-endian RGB555 colors; *uncoded*
16bpp cels carry literal RGB555 and load no PLUT. **Packed** data is per-line:
a 1-byte (bpp ≤ 8) / 2-byte (16bpp) word offset to the next line, then an
MSB-first packet bitstream — control = 2-bit type + 6-bit (count−1), types
`EOL/LITERAL/TRANSPARENT/REPEAT` = 0/1/2/3, pixels bit-packed at `bpp`. Lines are
word-aligned. Confirmed by decoding `credits/1.Cel` (a crisp 4bpp credits roll)
and `credits/bgnd.cel` (a 16bpp photographic backdrop).

### EA shapes (`.3sh`) — `shpm.go`

The bulk of the full-screen art (car photos, `TITLE`, `EALOGO`, menu backdrops)
is in EA's own **SHPM** container, decoded structurally from the disc (to be
confirmed against the `LaunchMe` loader in Part IV):

```
"SHPM", u32 size, u32 count, then a "SPoT" record:
  "SPoT", char[4] name, u32 hdrsize, u16 width, u16 height (@+0x10/+0x12),
  raw pixels @ +0x1C: width*height 16-bit big-endian RGB555.
```

The single-image 320×240 case (all eight car photos, the EA logo, etc.) decodes
cleanly. Other SHPM variants — non-320×240 (`TITLE`), multi-shape (`stats`),
and possibly indexed/packed bodies — fail cleanly and are deferred to the loader
trace, which will pin the record format exactly rather than by inference.

### Tools & tests

- `tools/threedo/cel.go` — `ParseCel`, `Cel.Image()`; `shpm.go` — `ParseSHPM`.
- `tools/threedo/cmd/celdump` — single-file and `-all` batch to PNG.
- `cel_test.go` round-trips a hand-encoded packed 4bpp cel (chunk walk + all
  packet types + PLUT); no dependency on the committed-out disc.

### Disc map (846 files, 31 dirs)

| Area | Count | Contents |
|------|------:|----------|
| `DriveData/` | 456 | per-car art (`.3sh`), physics/audio tunables (text: `#Mass 1 …`), track "Horizons" art |
| `FrontEnd/` | 117 | menus: `.3sh` shapes, `.cel`, fonts (`.3fn`), UI audio (`.aiff`/`.aifc`) |
| `Movies/` | 165 | `*.Stream` — 3DO streamed FMV (SANM video + SoundStream audio) |
| `System/` | 129 | the **Portfolio OS** itself: `Kernel/`, `Folios/`, 32 `Programs/`, `Drivers/`, `Fonts/`, DSP sound patches (`.dsp`), AIFF |

Extension census: 103 `.3sh`, 26 AIFF-family audio (`.aiff`/`.aifc`/`.aifffam`),
8 `.cel`, `.3fn` font, `.dsp` DSP patches, `.Stream` movies. Audio (AIFF-C,
`.dsp` sound patches) and the `.Stream` FMV are catalogued but not yet decoded;
they are lower priority than the code toolchain and are revisited later.

### `LaunchMe` is a standard ARM Image Format (AIF) executable

Confirming the M3/M4 approach: `LaunchMe`'s first 128 bytes are a textbook,
**big-endian** AIF header — `MOV r0,r0` (no decompress), a `BL` self-relocate,
`BL` zero-init, `BL` entry, `SWI 0x11` exit; RO(code)=0x3DB4C, RW(data)=0x99C4,
BSS=0x14944, **image base 0, address mode 0x20 = 32-bit** (so no ARM 26-bit
support is needed), code at base+0x80. `System/Programs/*`, `System/Kernel` and
`System/Folios/*` are AIF ARM60 code too — the Portfolio OS the oracle boots
through in Part IV.

## Part III — the ARM60 toolchain *(in progress)*
## Part IV — booting LaunchMe *(planned)*
