# Pilotwings 64 (Nintendo 64) — technical reference

**Image:** `Pilotwings 64 (USA).z64` — 8,388,608 bytes, MD5 `c5569227242e04138aac8457b7f83e6c`.
Not committed (copyright); supply your own copy. Despite the `.z64` extension the bytes are in
v64 order (16-bit halves swapped); the loader normalizes from the magic word, never the extension.

Pilotwings 64 (Nintendo/Paradigm Simulation, 1996) is a fully 3-D flight game and one of the
console's two launch titles. This document reconstructs the shipped cartridge from its bytes
alone, with purpose-built tools: a VR4300 disassembler/CPU (`tools/cpu/r4300`), an RSP
disassembler/CPU with its vector unit (`tools/cpu/rsp`), and a machine oracle
(`tools/platform/n64`) that boots the retail ROM through its own IPL3 and runs the whole RCP —
the graphics microcode is executed as code (LLE), never pattern-matched. The oracle renders the
attract sequence, including the 2-D title card, matching captures of real hardware.

This file currently documents the **asset load pipeline and the RDRAM→ROM traceback**; the full
toolchain and rendering write-up follows.

## The asset load pipeline

Nothing the attract sequence draws arrives in RDRAM verbatim. A scratch boot with PI-DMA logging
(`bootoracle -dmalog`, all 8,016 cartridge DMAs out to field 2100) shows the only large direct
loads are the boot's program image (cart `0x1000` → `0x200050`, 1 MiB) and a 562 KiB block at
cart `0x51E30` → `0x2CA900` whose first words decode as MIPS code — an overlay, not assets. No DMA ever touches the texture region `0x0A1978–0x0B8418`
or the model region `0x040000–0x07xxxx`. Watching those addresses (`-watch`) and walking the
write PCs back gives the machinery:

- **`load(dst, src, len)` at `0x8022A760`** — the game's one fetch primitive. `src` with bit 31
  set is a RAM pointer and the routine byte-copies; `src` below `0x80000000` is a cartridge
  offset, fetched by PI DMA in 4 KiB chunks through an alignment bounce. Thousands of 4-byte
  calls (`ra=0x8022AA50`) read directory structures straight out of the cartridge.
- **A MIO0 decompressor at `0x80231A20`.** Compressed blobs are staged at `0xDA800`, inflated
  into a window at `0x3DA800`, and slices of the window are then `load`-copied to their final
  addresses. The stream format, read off this routine's 40 instructions: 16-byte header
  (`'MIO0'`, decompressed length, back-reference stream offset, literal stream offset), then
  flag bits MSB-first in 32-bit words — set = copy one literal byte, clear = back-reference of
  length `(rec>>12)+3` at distance `(rec&0xFFF)+1`. Reimplemented in Go in `extract/mio0`
  (`sub $t1,$a1,$t2` then `lb -1($t1)`: the stored distance field is one less than the distance).
- On the cartridge each MIO0 image sits 8 bytes into an outer container: a fourCC (`COMM` for
  the attract assets, `TABL` for the first table at `0xDE74C`) and a 32-bit payload size.
- **Display-list templates ship in the same blobs** and are patched after placement: the ROM
  copy holds a `G_SETTIMG` (`FD`) with a zero address word; the loader writes the texture's
  final RDRAM address into it (e.g. the sky-band template at `0x0B4080`, three bytes
  `00 00 00 00` → `0A 39 A0`). A slice that is not byte-identical to the ROM is one of these,
  never a texture or mesh payload.
- **The frame display lists at `0x299280`/`0x2A15C0` are built by the CPU every frame**
  (double-buffered), as are transient vertex runs near them (`0x2906C0`); they exist on the
  cartridge only as the code that emits them.

### Verification

`extract/cmd/romtrace` replays the whole chain from the ROM alone: it parses a call log of the
loader and the decompressor (`bootoracle -calllog 8022A760 -calllog 80231A20`), inflates every
referenced blob with our own `mio0` package, and compares each copied slice byte-for-byte
against an RDRAM snapshot of the title scene. Result: **every texture and mesh payload is
IDENTICAL** (1,157,171 of 1,314,568 bytes across 2,371 replayed copies; the remainder is the
patched SETTIMG words in DL templates, runtime-mutated descriptor fields, and the transient
bounce buffers).

### RDRAM → ROM map (attract-scene assets, the GLB export set)

Cart offsets are file offsets in the byte-order-normalized image; every payload row verified
byte-identical. `+off` is the slice offset inside the decompressed blob.

| Object | Content | RDRAM | ROM source (MIO0 in `COMM` container) |
|---|---|---|---|
| island | mesh (13 G_VTX runs) | `0A9480–0B0080` | `0x0FF06C +0x08` |
| island | ocean tile | `0A49A0–0A51A0` | `0x24E084 +0x14` |
| island | terrain tiles | `0A1978–0A9480` | one blob per tile: `0x24A7C4`, `0x24B178`, `0x24BA74`, `0x24C2F8`, `0x24E084`, `0x2500DC`, `0x251C84`, `0x265D80`, `0x26C490`, `0x26CD44`, `0x273058`, each `+0x14` |
| island | far-water tile | `09BA20–09BCF0` | `0x04616B8 +0x0E` |
| sky | dome/horizon mesh | `077620–078100` | `0x1AAE34 +0x08` |
| sky | cloud + band tiles | `0B4798`, `0B5798`, `0B6418`, `0B7418` | `0x301AB8`, `0x302854`, `0x303060`, `0x3039AC`, each `+0x14` |
| logo | letters/wing/feather/6/4 mesh | `05D2B0–063B50` | `0x1661B0 +0x08` |
| logo | gradient texture | `041820–041A20` | `0x266E0C +0x14` |
| gyros | per-part meshes | `0670E8–076xxx` | `0x189190`, `0x18B1C0`, `0x18D148`, `0x18EFA0`, `0x190DA0`, `0x192B88`, each `+0x08` |
| gyro A | body texture | `0452A0–0462A0` | `0x2E15C8 +0x14` |
| gyro B | body texture | `0462A0–047258` | `0x2E191C +0x14` |
| — | frame display lists | `299280`, `2A15C0` | built per frame by the CPU (not on cart) |
| — | material DL templates | `0B3D80–0B4768` | trailing slices (`+0xACC`…) of the texture blobs above, SETTIMG patched at load |

The loads run at ~12 blobs per video field across fields 266–283 (and again from field 272 when
the attract's second scene rebuilds the arena), which is why the sequence fades in scene by scene.

## Verifying the display-list walk against the RDP stream

The GLB exports come from `extract/cmd/dlwalk`, which walks the frame's Fast3D display list out
of an RDRAM snapshot (the walker lives in `extract/f3d`). `extract/cmd/dlverify` verifies that
walk mechanically: for one field it projects every walked triangle through the walked modelview,
projection and viewport, and compares against the RDP triangle stream the oracle actually
executed that field (captured via OnRDPCmd up to Sync_Full, each triangle's texture source
tracked by TMEM address, its bounding box reconstructed from the edge coefficients). Tolerances
are declared up front: counts per texture source in `[inside, inside + 7·straddling + slivers]`
(the RSP clips and subdivides; near-degenerate triangles may collapse under its quantization),
1:1 nearest-neighbour bbox matching at 2 px, and a per-triangle fixed-point band above that
(the RSP concatenates MV·P in s15.16; the logo scene's rotation elements are ~0.01, one quantum
is 0.1% of the element).

Running it pinned four conventions the walker now embodies: screen y maps with **negated**
viewport scale (x positive); the cull-face sign in that flipped space (front faces wind
clockwise); the viewport applies **per vertex** — the RSP maps to the screen at G_VTX time, and
the title card reprograms the viewport between the 3-D backdrop and the logo overlay; and a draw
with G_TEXTURE off is untextured no matter what tile is bound.

Standing results: the flyby scene passes clean (692/695 triangles matched within 2 px, every
count in range); title and logo match 99.4% each, failing only on seven small untextured
triangles apiece in one static logo-cluster group — their matrix and vertex bytes are identical
between snapshot and run, both positions show rendered geometry in the frame, and they look like
near-duplicate draws pairing off against each other. Unexplained; dlverify exits non-zero on
them deliberately so the question stays visible.

The G_TEXTURE fix re-cut the export curation: the PILOTWINGS letters are **vertex-coloured**,
not textured (their 1,464 triangles draw with texturing off; only the emblem wing samples the
`041820` gradient), so the earlier letters GLB carried a texture the game never applied to them.
The rotor-blur discs and the island's surf decal render through the blender — `FORCE_BL`
(0x4000) in othermode-L, mapped to glTF `alphaMode: BLEND` — and the discs' **vertex alpha
carries the blur amount** (38 at the rim to 153 at the hub; the tail discs sit at a constant
126). Current curation, by draw-group index: title scene — ocean 0-1, sky 2-5, island 6-16,
feather 17, wing 18, letters 19, "6" 20-21, "4" 22-23; flyby scene — gyro A 17-29+44-45,
gyro B 30-43+46-47.
