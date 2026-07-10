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
- On the cartridge each MIO0 image sits 8 bytes inside a `GZIP` chunk of an IFF `FORM` — see
  *Part IV — the cartridge archive* below, which supersedes the "outer container" reading this
  section originally recorded.
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

## Part IV — the cartridge archive

Every asset in the game is a file in a single archive, and that archive is **IFF-85** — the same
`FORM`/chunk container the Amiga uses, byte-for-byte, on a Nintendo cartridge. It occupies cart
`0x0DE720`–`0x0618B6C`: **1,273 back-to-back `FORM` chunks**, contiguous but for one 12-byte run
of zeros after the first. Below it sits the program image; above it, the audio banks.

```
FORM <u32 size> <type>              type: UVTX, UVMD, UVCT, UPWT, ... (20 resource types)
  PAD  <u32 4> 00 00 00 00          padding, always first
  <tag> <u32 size> <bytes>          a plain chunk, or …
  GZIP <u32 size>                   … "the chunk that follows is compressed"
    <tag> <u32 uncompressedSize>
    MIO0 …                          the stream decoded in Part III
```

`GZIP` is a wrapper tag, unrelated to RFC 1952: its payload is an ordinary chunk header whose
size field carries the *decompressed* length, followed by an MIO0 stream. A `FORM` holds between
zero and thirteen of them. All 39,150 chunk sizes are even and every `FORM`'s chunk list ends
exactly on its boundary, so IFF's word-padding rule never fires here.

### The directory

The first `FORM` has type `UVRM` and holds one compressed chunk, `TABL` — the archive's table of
contents, and the very first thing the game loads (video field 10, cart `0x0DE74C`). It is
**1,272 entries of `(fourCC type, u32 totalLength)`**, one per following `FORM`, in ROM order.
The length counts the `FORM`'s own 8-byte header, which makes the directory a walkable index:
resource *i* begins at `0x0DF5B0` plus the sum of the preceding lengths. The running sum over all
1,272 entries is 5,477,820 bytes and lands **exactly** on the archive's end. Combined with an
independent walk of the `FORM` chain — which agrees with the directory on every resource's type,
length and offset, 1,272 for 1,272 — that arithmetic is what proves the container.

This is why no cartridge offset table exists anywhere in the ROM: the game does not need one. It
reads `TABL` once and walks. The thousands of 4-byte cartridge fetches noted in Part III
(`ra=0x8022AA50`) are that walk, reading chunk headers straight off the cartridge bus.

### The inventory

| Type | × | Bytes | Chunks | Contents |
|---|---|---|---|---|
| `UVTX` | 463 | 820,444 | `GZIP(COMM)`, 23 plain `COMM` | textures |
| `UVMD` | 363 | 876,404 | `GZIP(COMM)` | models |
| `UVAN` | 115 | 146,972 | `COMM`, `PART`, `GZIP(PART)` | animations |
| `UVBT` | 102 | 426,728 | `GZIP(COMM)`, 1 plain `COMM` | per-chunk terrain companion |
| `UVCT` | 101 | 544,852 | `GZIP(COMM)` | terrain chunks |
| `UPWT` | 61 | 103,008 | `NAME` `INFO` `COMM` `JPTX` `TPAD` `LPAD` `RNGS` `THER` `BALS` `TARG` `CNTG` `LSTP` `LWIN` `HOPD` `PHTS` `FALC` `HPAD` `BTGT` | missions |
| `PDAT` | 25 | 840,512 | `PHDR`, `PPOS`×7910, `RHDR`, `RPKT`×24401 | recorded flight data |
| `3VUE` | 12 | 68,832 | `COMM` `QUAT` `XLAT` | camera views |
| `UVFT` | 9 | 15,892 | `FRMT` `STRG` `GZIP(BITM)` `GZIP(IMAG)` | fonts |
| `SPTH` | 8 | 5,392 | `SCP#` `SCPP` `SCPH` `SCPR` `SCPX` `SCPY` `SCPZ` | spline paths |
| `UPWL` | 4 | 76,144 | `LEVL` `ESND` `LPAD` `WOBJ` `TOYS` `APTS` `TPTS` `BNUS` | the four levels |
| `UVLV` | 1 | 12,160 | `COMM`×136 | level index |
| `UVTR` | 1 | 8,864 | `COMM`×10 | the world map |
| `UVSX` | 1 | 1,455,088 | `.CTL` `.TBL` | audio bank |
| `ADAT` | 1 | 74,120 | `SIZE`, `NAME`×439, `DATA`×439 | named sound effects |
| `UVEN` | 1 | 1,912 | `COMM`×24 | — |
| `UVTP` | 1 | 264 | `COMM`×7 | — |
| `UVSQ` | 1 | 136 | `COMM` | — |
| `UVSY` | 1 | 72 | `COMM` | — |
| `UVLT` | 1 | 24 | (only `PAD`) | — |

Several chunk tags name themselves. `UPWT`'s `NAME` chunks are the developers' own mission
labels, in the clear: `LEVEL 1`, `CB L1 Target2`, `SD 1`. `ADAT` pairs 439 `NAME`/`DATA` chunks
(`COUNT_3`, `B_CLR_N`, `HG_P1_S2`). `UVSX` holds `.CTL` and `.TBL` — libultra's audio bank pair.
The `UVxx`/`UPWx` prefixes and the four `UPWL` forms line up with the game's four islands, but
the mapping from a chunk tag to a *meaning* is only ever settled by disassembling the code that
reads it; the table above records what the container says, not what the names suggest.

One hypothesis the chunk names already corrected: `PDAT` was assumed to be object placement, on
its size alone. Its chunks are `PHDR` + 7,910 `PPOS` and `RHDR` + 24,401 `RPKT` — the shape of
recorded flight paths, not scenery. Placement lives elsewhere (`UPWL`'s `WOBJ`, `UPWT`'s
per-mission chunks are the candidates).

### Reading it

`extract/pwad` is the reader; `extract/cmd/romls` lists and censuses the archive, and
`extract/cmd/romextract` writes all 36,723 chunks out, decompressing as it goes. Neither boots
the machine.

Four checks stand behind it, all mechanical:

1. The `FORM` walk finds exactly 1,273 forms and consumes the archive with no unexplained gap.
2. The walk and the `TABL` directory — two independent readings — agree on every resource's type,
   byte length and offset, and the directory's running sum ends exactly at the archive's end.
3. All **1,322 MIO0 streams inflate to precisely the length their chunk header declares**: 1,322
   independent checks of the codec reimplemented in Part III.
4. **The static archive is the archive the game reads.** `romls -calllog` takes a loader trace
   from a real boot and checks every cartridge fetch against the parsed chunk list: all 104
   staged (compressed) fetches land on a `GZIP` chunk body, all 1,234 in-archive direct fetches
   land on a plain chunk body, zero exceptions. The four fetches outside the archive are the
   program overlay at `0x051E30` and the post-archive audio data at `0x0618B70`.

Check 4 is the one that matters most: the first three could all pass on a self-consistent
misreading. Only the running game can testify that the container we parse is the container it
loads from.

### What the container does not contain

The frame display lists are not on the cartridge — the CPU rebuilds them every frame into the
double buffers at `0x299280`/`0x2A15C0`. The terrain's `G_TRI1`s come out of a strip builder at
`0x802206C0`, which walks one flat vertex pool through three globals at `0x80296A80` (pool
pointer, cursor, end), emitting `G_VTX` batches of ≤16 vertices and purely sequential triangles
`k, k+1, k+2` (the index emitter is the `multu` by 10 at `0x802208E4`). Terrain connectivity is
therefore *implicit*: a mesh chunk stores a vertex pool and a stream that drives those globals,
not a triangle list. Decoding that stream, and the corresponding builders for articulated models,
is what the per-format sections below do.

Two early readings of the compressed payloads, recorded here because the archive now explains
them: a texture chunk opens with `u16` payload length and `u16` format code (`0x19` RGBA16,
`0x0D` I4) and ends with material display-list templates carrying a zero-address `G_SETTIMG` that
the loader patches once the texture lands; a mesh chunk opens with `u16` vertex count and `u16`
type code (`0x0101` terrain, `0x010A` articulated), followed by that many 16-byte Fast3D vertex
records.

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

Two walker defects the verification-era work exposed in the exports themselves:

- **Set_Tile's tile index is in the low command word** (`w1>>24&7`, like Set_Tile_Size). The
  walker read it from the high word, whose bits 24-31 are the opcode — `0xF5 & 7 = 5` — so every
  Set_Tile landed in tile 5 and the tile the draws sample never received its format, line stride
  or wrap bits. Every texture decode failed silently to the white fallback: **all GLB exports
  before 2026-07-10 were vertex-coloured only**, and looked plausible because Pilotwings' terrain
  carries most of its look in vertex colours. With the index fixed, all textured groups decode
  (terrain RGBA16 tiles, I4 sky/water/gradient tiles) and the island shows its farmland, surf,
  forest and airstrip texel detail.
- A tile is a **window** into its loaded texture: SL/TL (10.2) give the origin — the ocean binds
  two windows of one water image — and the tile's coordinate shift scales s/t before windowing.
  Both now feed the texture decode and the exported UVs.

The G_TEXTURE fix re-cut the export curation: the PILOTWINGS letters are **vertex-coloured**,
not textured (their 1,464 triangles draw with texturing off; only the emblem wing samples the
`041820` gradient), so the earlier letters GLB carried a texture the game never applied to them.
The rotor-blur discs and the island's surf decal render through the blender — `FORCE_BL`
(0x4000) in othermode-L, mapped to glTF `alphaMode: BLEND` — and the discs' **vertex alpha
carries the blur amount** (38 at the rim to 153 at the hub; the tail discs sit at a constant
126). Current curation, by draw-group index: title scene — ocean 0-1, sky 2-5, island 6-16,
feather 17, wing 18, letters 19, "6" 20-21, "4" 22-23; flyby scene — gyro A 17-29+44-45,
gyro B 30-43+46-47.
