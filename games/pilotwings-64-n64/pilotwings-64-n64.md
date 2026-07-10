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

A mesh chunk opens with `u16` vertex count and `u16` type code (`0x0101` terrain, `0x010A`
articulated), followed by that many 16-byte Fast3D vertex records.

## Part IV §2 — `UVTX`: the textures

Each of the 463 `UVTX` resources is a single `COMM` chunk — usually compressed, 23 of them raw.
Its header is read by the game at `0x80226A54`, which pulls the fields one at a time through a
cursor helper at `0x80225394`:

```
u16 dataSize      texel bytes that follow the header
u16 formatCode    an authoring code; the reader parses and discards it
u32 × 4           read and discarded
u8  texels[dataSize]
…                 a Fast3D display list: the material template
```

`dataSize` is clamped to **4096** — the reader prints an error and truncates above that — so one
TMEM is the hard limit on a Pilotwings texture, and the largest are 128×64 I4 or 64×64 I8. The
`formatCode` takes twelve values (`0x19` on 230 resources, `0x0D` on 118, …) but the game never
uses it: the authority is the template.

### The material template

The bytes after the texels are an ordinary display list. All 463 parse to completion:

```
G_TEXTURE       tile 1 (431 resources) or tile 0 (32), scale 0xFFFF/0xFFFF
G_SETOTHERMODE_H × 5, G_SETCOMBINE, G_RDPLoadSync, G_RDPTileSync
G_SETTIMG       fmt/siz, address word 0x00000000  ← the loader patches this
Set_Tile        tile 7, the load tile
Load_Block      tile 7, dxt = 0
Set_Tile        tile 1 …                          ┐ up to six mip levels,
Set_Tile_Size   tile 1 …                          ┘ tiles 1..6, halving
G_ENDDL
```

Three properties hold across every resource, and the decoder relies on all three: **no template
loads a palette** (nothing is CI4/CI8), **every `Load_Block` uses `dxt = 0`** (so the texels ship
pre-swizzled, exactly as TMEM wants them), and **every template terminates**. The texel formats
are RGBA16 (255), I4 (103), IA8 (56), I8 (20), IA4 (20) and IA16 (9).

Decoding is therefore not a matter of guessing dimensions from `dataSize`. Walk the template with
the same Fast3D interpreter the renderer uses (`extract/f3d`), take the tile `G_TEXTURE` selects,
and read the texels through it. Two wrinkles, both of which the game's own data forced:

- **The drawing tile may carry no `Set_Tile_Size`.** Thirty-two templates draw through tile 0 and
  size it never. Its extent is then either the *sibling tile at the same TMEM base and line
  stride* — the mip-mapped templates re-declare level 0 as tile 0 at the very end, so tile 1 holds
  the size — or, failing that, the tile's own wrap **mask**: 2<sup>mask</sup> texels per axis,
  where the RDP clamps. In the mask case the template configures no mip levels, so the image must
  account for every stored texel byte exactly; the decoder asserts it, because a wrong extent
  would otherwise decode into a plausible picture.
- **The odd-row swizzle is computed on the absolute byte offset** (`off ^= 4`), so it only lands
  correctly when the texels begin on an 8-byte boundary. In RDRAM they do — the game's allocator
  aligns to 8. Inside the chunk they begin at offset 20. Decoding in place silently corrupts every
  odd row; the decoder copies the texels to an aligned base first.

### Verification

`extract/cmd/texdump` decodes all 463 with **zero fallbacks** — there is no white-square path —
and `-verify` checks the ROM decode against the running game. It walks a frame's display list out
of an RDRAM snapshot and, for each textured draw group, finds the `UVTX` whose texels the game
copied to that address. Then:

- the fields that say *how the bytes are read* — format, size, line stride — must match between
  the frame's tile and the resource's own template;
- the **game's copy of the texels, decoded through the extent our ROM template gave us, must be
  pixel-identical to the ROM decode.**

Across the title card and the flyby, 44 draw groups match a resource and every one is identical.
The check is what caught the swizzle-alignment bug above; a silent fallback would fail it on the
first group. It also measured two things worth stating: three groups sample a **window** of their
texture rather than the whole of it (the ocean wraps a 65×65 window over a 64×64 image), and one
is re-issued with different wrap modes for its draw. Extents and wrap modes are sampling
behaviour and vary per use; format, size and line stride never do.

## Part IV §3 — `UVMD`: the models

Each of the 363 `UVMD` resources is one compressed `COMM` chunk. Its reader is at `0x802256B8`
and pulls every field through the same cursor helper as the textures, so the file is a strictly
linear stream and the field order below is simply the order the reader calls that helper in:

```
u16 vertexCount
u8  lodCount            levels of detail
u8  partCount
u8  faceCount           count of the 36-byte records after the matrices
u8  flag
u16 boneCount           count of the 6-byte records at the end
Vertex[vertexCount]     16 bytes each — the Fast3D vertex pool, copied verbatim to RDRAM
for each lod:
    u8 partCount, u8 flag
    for each part:
        u8 batchCount, u8 flag, u8 flag
        for each batch:
            u32 material
            u16 unknown, u16 unknown
            u16 commandCount
            command[commandCount]
    u32 lodDistance
Matrix[partCount]       64 bytes: a 4x4 float rest pose
Face[faceCount]         36 bytes  (consumer not yet traced)
u32 × 3
Bone[boneCount]         6 bytes   (consumer not yet traced)
```

The whole resource is consumed but for 0–7 bytes of padding to an 8-byte boundary, on all 363.
`vertexCount × 16` is exactly the length of the slice the loader copies to RDRAM, and the matrices
begin exactly where its second slice starts — the two placements we had already observed in Part
III fall out of the header.

### The command stream: connectivity that isn't stored

A batch is **not** a triangle list. It is a compact encoding the game expands into a Fast3D
display list once, at load time; the emitter is at `0x80225940`. Each command is a `u16`:

| bit `0x4000` | meaning |
|---|---|
| set | a triangle — three 4-bit **vertex-buffer slots** in bits 8–11, 4–7, 0–3 (the game multiplies each by 10, a `G_TRI1` index's byte stride) |
| clear | a vertex load — bits 0–13 are an index into the vertex pool, and a following `u8` packs `((n-1) << 4) \| slot`, loading *n* vertices into the 16-entry buffer at *slot* |

So the mesh's connectivity lives in a 16-slot sliding window, exactly as the RSP sees it. The
decoder replays that window to recover triangles as indices into the vertex pool. This is why the
earlier guess of a "patch stream" of unknown layout was only half right: the stream *is* the
display list, minus the addresses.

### Materials

The batch's `u32` material has one pinned field: **the low 12 bits index the archive's `UVTX`
resources in order**, 0–462 — the texture resources are contiguous (indices 471–933), so the
ordinal is all that is needed. The value `0xFFF` means **untextured**, and the batch draws with
vertex colours. That is the independent confirmation of something the RDP stream had already told
us: the PILOTWINGS letters are 1,464 untextured triangles. Across the archive 1,481 of 3,427
batches are untextured. The material's upper 20 bits are render flags the loader tests (bit 27 is
latched into a per-part byte); they are carried through raw rather than guessed at.

### Rest poses

Each part has a 64-byte 4×4 float rest pose, in part order. The evidence for that pairing — which
otherwise would rest only on the loader's allocation order — is that **all 264 single-part models
in the archive carry an identity matrix.** A part-ordering error would have to arrange for every
one of them to be identity by coincidence. The pairing holds at LOD 0; a lower LOD may drop a part
outright (model 83's LOD 1 has two parts against three matrices), so nothing is placed below it.

### Verification: the display lists, byte for byte

This is the narrow seam. `extract/cmd/mdldump -verify` takes an RDRAM snapshot, locates each
model's vertex pool by its bytes, and then — from the ROM alone — rebuilds the exact Fast3D
display list the engine's emitter would write for every batch, and requires those bytes to appear
in RAM. **12 models resident, 197 display lists, every one byte-for-byte identical, zero
mismatches**, across both the title-card and flyby snapshots.

The engine wrote those command words; we derived them from the file. Matching them verifies the
command-stream decode, the slot window, the vertex indices and the emitter's encoding all at once,
without reimplementing one line of the renderer. It also caught the one encoding error in the
first attempt: the `G_VTX` pointer is a **segmented** address (the pool's physical address), not a
KSEG0 one.

Rendering the results confirms it from the other side: `uvmd-0047` is the island, textured, and
`uvmd-0212` is the ten pieces of the PILOTWINGS logo.

## Part IV §4 — the world: `UVTR`, `UVCT`, `UVBT`

The attract sequence draws its island as a single `UVMD` model. The playable worlds are built
differently, and none of them is ever loaded before the title card — so nothing in this section is
verified by watching the machine. It is verified by assembling the world and looking at it.

### `UVTR` — ten world grids

The archive holds exactly one `UVTR` resource, carrying **ten uncompressed `COMM` chunks: ten
worlds.** Its reader is at `0x802270BC`:

```
f32 minX, minY, minZ        the world's lower bound
f32 maxX, maxY, maxZ        its upper bound
u8  cols, u8 rows
f32 cellW, f32 cellH
f32 radius
Cell[cols*rows]             row-major
```

and a cell is one byte of presence, and if present a 4×4 float matrix, a flag byte, and a `u16`:

```
u8  present                 0 -> the loader zeroes a 72-byte slot and moves on
f32[16] matrix
u8  flag
u16 chunk                   a UVCT resource's ordinal, 0..100
```

Three facts pin it, and none of them can hold by accident:

1. Every world's extent is **exactly** `cols × cellW` by `rows × cellH`.
2. Every present cell's matrix translates to that cell's **centre**,
   `(minX + (col+½)·cellW, minY + (row+½)·cellH)` — for all 120 cells of all ten worlds. The
   grid's second axis lands in the matrix's *Y* slot: the terrain's ground plane is X/Y with Z up,
   as the island model already implied.
3. The ten grids name **each of the 101 `UVCT` resources at least once**, and nothing outside them.

The worlds: a 2×2 of 600-unit cells; two 8×8 grids of 512; four 8×5 grids of 1000 sharing one
8000×5000 extent; two 2×5 grids of 1000; and a single 6000-unit cell. Nineteen chunks are named by
more than one grid — the grids that share an extent are variants of one place, re-using most of its
terrain.

### `UVCT` — the terrain chunks

Read at `0x80225FBC`, through the same cursor helper as everything else:

```
u16 vertexCount, u16 faceCount, u16 objectCount, u16 batchCount
Vertex[vertexCount]     16 bytes — the Fast3D vertex pool
Face[faceCount]         8 bytes  — three u16 vertex indices and a word
Object[objectCount]     u8 poseCount, that many 64-byte 4x4 matrices, then u16, u32×3, u16, u16
Batch[batchCount]       a UVMD batch, then a 24-byte trailer
u32 × 5
```

**A terrain batch is a model batch.** Same material word, same `u16` command stream, expanded by
the same display-list emitter at `0x80225940` — so `extract/uvmd`'s `DecodeBatch` is reused
verbatim, and the terrain's connectivity is the same 16-slot sliding window. (This retires an old
suspicion: the "strip builder" at `0x802206C0`, once assumed to draw terrain, draws something else.)

What `UVCT` adds is a trailer after each batch's commands: `u16 faceStart, u16 faceCount`, then
four more words. The loader turns `faceStart` into a pointer into the face array, so **the faces
belong to the batch drawn over them** — three vertex indices and a word, which says collision. No
code in the render path reads them, so they are carried through undecoded rather than named.

### `UVBT` — identified, not decoded

102 resources, tantalisingly close to `UVCT`'s 101. Its loader is `0x80227D34` and its parser
`0x80227260`, which reads the chunk through **64-bit shift helpers**: it is a bit-packed structure,
not an array of records, and nothing in the terrain draw path touches it. Guessing at it from the
count coincidence is exactly what this project does not do, so it is left here: named, located,
and undecoded.

### Verification: assemble the world

`extract/cmd/worldexport` decodes the ten grids and the 101 chunks from the ROM, places every
chunk by its cell's transform, and writes one continuous textured GLB per world. Because each
chunk is decoded and placed independently, a wrong transform, a wrong axis or a wrong cell stride
has nowhere to hide.

- **Containment, mechanically:** no vertex of any placed chunk escapes the cell that names it, in
  any of the ten worlds. Worst overhang: **0.0000 units.** (A chunk need not *fill* its cell — a
  coastal cell stops at the shoreline — so measuring the gap between neighbours proves nothing;
  measuring the overhang proves everything.)
- **Continuity, visually:** world 1 assembles 45 chunks into the crescent island the attract
  sequence flies over — the same landmass, from a completely different source. World 3 assembles
  38 chunks into **Little States**, unmistakably the United States, with the Rockies, the
  Mississippi and Florida.

20,846 terrain triangles across the 101 chunks.

## Part V — missions, levels, paths and recorded flights

Everything in this Part ships **uncompressed**, so the work was never decoding — it was meaning.
The mission reader lives in a gameplay overlay (not in the base program image; it loads at
`0x8034xxxx`) and is a switch over chunk tags at `0x80345D40`. Almost every arm does nothing but
`load(global, chunkPtr, chunkSize)` — copy the chunk verbatim into a fixed buffer. **A mission's
features are therefore flat record arrays, and the tag names each one.**

### `UPWT` — the 61 missions

```
NAME    the developers' own label, NUL-padded
JPTX    a short key: "A_EX_1"
INFO    a description: "shoot at the the target"
COMM    1072 bytes, the mission descriptor
TPAD    48 bytes, exactly one per mission: the takeoff pad
…       zero or more feature arrays, by tag
```

`NAME` gives all 61 missions the names their authors used: `LEVEL 1`, `CB L1 Target2`, `SD 3`,
`BIRD 3C`, `GC Exp`, `RP P2`, `HG B1`. The prefixes are the craft — cannonball, sky diving,
birdman, gyrocopter, rocket belt, hang glider.

The `COMM` descriptor's **first word decomposes into four bytes: class, vehicle, variant, level.**
The vehicle byte is what makes that reading more than a story: it agrees with the developers' own
prefix on **all 42 prefixed missions** — every `GC` is vehicle 2, every `RP` 1, every `HG` 0,
`CB` 3, `SD` 4, `HM` 5. The level byte is 0–3 and there are exactly four `UPWL` levels. The other
1068 bytes of the descriptor are carried through raw.

The feature arrays, with the strides the archive's sizes pin (a tag appearing once per mission
pins its own; the rest by common divisor):

| tag | stride | tag | stride | tag | stride |
|---|---|---|---|---|---|
| `TPAD` takeoff pad | 48 | `LPAD` landing pads | 48 | `CNTG` | 32 |
| `RNGS` rings | — | `THER` thermals | 40 | `HOPD` | 32 |
| `BALS` balloons | 104 | `TARG` targets | 32 | `BTGT` | 32 |
| `LSTP` | 40 | `HPAD` | 128 | `FALC` | 344 |
| `LWIN`, `PHTS` | — | | | | |

A dash means the sizes across the archive share no divisor big enough to claim, so the array is
carried as bytes. Beyond the stride, no record's fields are claimed: nothing in the render or
gameplay path has been traced to them.

### `UPWL` — the four levels

`LEVL` (8 bytes), `LPAD` (24-byte landing pads), `WOBJ` (16-byte world objects: a position and a
word), and further arrays `TOYS`, `APTS`, `TPTS`, `BNUS`, `ESND`. Levels 1 and 2 carry 6 and 13
world objects; levels 0 and 3 carry none.

### `SPTH` — eight spline paths

Seven chunks, one per animated channel — `SCPX`, `SCPY`, `SCPZ`, `SCPH`, `SCPP`, `SCPR`, `SCP#` —
and each is

```
u32 count
u32 unknown
Key[count]     two f32s
```

`8 + 8*count` equals the chunk's size for all 56 chunks, exactly. Which float is the value and
which the time does not follow from that, but two measurements settle it: in **every** multi-key
curve (49 of them) the *second* component rises across all keys but the last, whose second
component is **exactly 0.0** — a terminator; and in six of the eight paths, `SCPX`, `SCPY` and
`SCPZ` carry an *identical* array of second components, which is what a shared timeline looks like.
(The other two give X, Y and Z different key counts outright, so they are keyed independently.)

### `PDAT` — 25 recorded flights

`PDAT` was assumed to be object placement, on its 841 KB alone, before the chunk names were read.
It is not. Each resource is `PHDR` (8 bytes) plus 24-byte `PPOS` position samples — three floats
of position, three more of orientation — and `RHDR`/`RPKT` replay packets (24 and 16 bytes).
Across the archive: 7,910 samples and 24,401 packets. These are recorded flights.

### `3VUE` — twelve camera views, undecoded

`COMM` (32 bytes), `QUAT` and `XLAT`. `QUAT`'s size divides by 16 in all twelve. `XLAT`'s divides
by 12 in only six — so **`XLAT` is not an array of vec3s**, the obvious reading is wrong, and
`3VUE` is left named and undecoded rather than forced.

### Verification

Nothing here is verified by the machine: the attract sequence loads none of it. What verifies it
is that these formats have to agree with the worlds decoded in Part IV. The pads and flight
samples are float triples read at offsets nothing forced on us — a wrong one yields denormals or
1e38, not coordinates.

`extract/cmd/missions` checks, from the ROM alone:

- The vehicle byte agrees with the mission name on **42/42** prefixed missions.
- **61/61 takeoff pads** lie inside at least one world, and for each of the four levels there is at
  least one world containing *all* of its missions' pads. Level 2's pads fit only the four
  8000×5000 worlds — Little States — and no other level's do.
- Each `UPWL[i]`'s landing pads fit the same worlds as level *i*'s takeoff pads, which independently
  corroborates that the descriptor's level byte indexes the `UPWL` resources in order.
- **8/8** spline paths are 0.0-terminated with rising times; 6/8 share one position timeline.
- **7,872 of 7,910** recorded flight samples lie inside a world. The 38 that do not are one
  recording leaving the western edge of its 8000-unit world, contiguous and within ~313 units of
  the boundary — a flight, not a misread.

### Driving the oracle: `bootoracle -keys`

The machine already modelled the controller completely — `tools/platform/n64/si.go` answers the
joybus with the buttons and stick of `Controllers[0]`, and a controller is attached by default —
so injection needed only a field-timed script:

```
bootoracle -keys "2160:+start,2190:-start,2260:+a,2290:-a"
```

Setting `Controllers[0]` *is* the game's real input path: Pilotwings polls the joybus itself, so
nothing is injected past the hardware the way Ultima Underworld's oracle must inject past DOS. The
script applies at a field boundary, so the game's once-per-frame poll sees each change exactly
once, like a real press.

**It does not yet reach a mission.** Holding Start and A on the title card demonstrably changes the
machine's state — a run with the button held diverges from one without — but the menu does not
advance, and the reason is not yet found. Since every check in this Part is static, nothing here
depends on it. Driving into a mission stays open, and would earn its keep by confirming one
reachable mission's feature records against this decode.

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
