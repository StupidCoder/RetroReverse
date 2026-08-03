# Jak & Daxter: The Precursor Legacy (PlayStation 2) — technical reference

**Image:** `Jak and Daxter - The Precursor Legacy.iso` — 1,745,796,096 bytes,
MD5 `b6c372a4a34ff7ca0d0ccc4d8b7da879` (SCUS-971.24, NTSC-U).

**Contents**

- [Part I — The disc](#part-i--the-disc)
- [Part II — The GOAL object format and linker](#part-ii--the-goal-object-format-and-linker)

The PlayStation 2 machine model itself (EE/IOP/GS/VU, DMA, SPU2, CDVD) is shared
with the other PS2 titles and lives in `tools/platform/ps2`; the per-game boot
oracle is `extract/cmd/bootoracle`. The boot ELF ships its full symbol table
(1,494 symbols, 876 functions, C++-mangled), and the engine's GOAL symbol table
(5,376 named functions, globals and types) is recoverable at runtime with the
oracle's `-goalsyms`; both are load-bearing for everything below.

## Part I — The disc

A single-session ISO 9660 DVD. `SYSTEM.CNF` names the boot executable
`SCUS_971.24` (530,357 bytes) — the GOAL *runtime kernel*: a C/C++ program that
owns the console, hosts the linker, and pulls the actual engine and game (all
compiled GOAL code) from the archives below. `DISK_ID.DIZ` carries the build
identity. The directories:

| dir | contents |
| --- | --- |
| `DRIVERS/` | IOP modules: `OVERLORD.IRX` (Naughty Dog's I/O server: DGO streaming, RPC), `989SND.IRX` (sound driver), Sony's `LIBSD`/`MCMAN`/`MCSERV`/`PADMAN`/`SIO2MAN`, the `IOPRP221.IMG` IOP kernel update, and `SCREEN1.*` boot-screen images per region |
| `CGO/` | common archives, loaded at boot: `KERNEL.CGO`, `ENGINE.CGO`, `GAME.CGO`, `ART.CGO` (shared texture pages + common actor art), `COMMON.CGO`, plus per-hub actor packs (`JUNGLE.CGO`, `RACERP.CGO`, `L1.CGO`, …) |
| `DGO/` | one archive per level: `VI1/VI2/VI3` (villages), `JUN` (jungle), `BEA` (beach), `MIS` (Misty Island), …, and the non-level scenes `TIT.DGO` (Naughty Dog logo + title), `INT.DGO` (intro cutscene: `evilbro`/`evilsis`), `DEM.DGO` (demo), `FIN.DGO` (finale) |
| `VIS/` | per-level precomputed visibility bitsets |
| `SBK/` / `MUS/` | 989SND sound banks / music banks, per level |
| `STR/` | streamed cutscene data, one file per scene-part (spooled `art-joint-anim` chunks; the audio for a scene streams alongside) |

The boot order, observed live at the linker (Part II): the CGO set
(`KERNEL` → `GAME` → `ENGINE` → `ART` → common actor packs) first, then
`VI1.DGO` immediately — the title screen shows Sandover Village behind the
logo, so the level is resident before the menu. The TIT logo/title art
groups arrive in the same window via the `load-dir-art-group` id path
rather than an archive load, and the memory-card prompt (dismissed with ✕)
gates the NDI/logo scenes, whose animations then stream from `STR/` as
spooled objects. `INT.DGO` loads when the intro cutscene starts.

### The DGO/CGO container

`BeginLoadingDGO`/`GetNextDGO` (OVERLORD, RPC from the EE) read a trivial
catalog: a 64-byte archive header `{u32 count, char name[60]}`, then per object
a 64-byte header `{u32 size, char name[60]}` followed by `size` payload bytes,
each payload 16-byte aligned. Objects stream double-buffered into the top of EE
RAM (`0x17FC040`/`0x1BFC040` buffers) and are linked out of the buffer one at a
time. A name can repeat inside one archive: the first `evilbro` in `INT.DGO` is
the actor's GOAL behavior code (a v3 code object), the second is its art group
(a v4 data object).

`extract/cmd/dgo` lists archives (`-list`) and extracts raw (`-raw`) or linked
(`-base`, `-symtab`, `-out`) objects; `extract/goalobj` is the library.

## Part II — The GOAL object format and linker

Everything in a CGO/DGO archive is a *GOAL object file*, linked into the
running image by the kernel's `link_control` class in `SCUS_971.24` —
`begin` `0x1097F0`, `work` `0x109A50` (time-sliced from the main loop via
`link_begin`/`link_resume`, budgeted at 0x249F0 profiler ticks per slice),
`finish` `0x10AAF0`, with the wire-format workers `work_v2` `0x10A590`,
`work_v3` `0x109B38` and the relocation applicators `c_symlink2` `0x10B248`,
`c_symlink3` `0x10B3D8`, `c_rellink3` `0x10AFA8`.

Word 2 of the file selects the layout:

- **Version 4** (all art/data objects — texture pages, art groups, `*-vis`):
  16-byte header `{tag, linkDataSize, version=4, objectSize}`; the **object
  itself at file+16**, already in its final byte layout; then a **trailing link
  block** at `file + 16 + objectSize`: `{typeSlot, blockSize, wireVersion,
  stream…}`. The runtime stamps the `link-block` type (`0x1629C4`) into
  `typeSlot` while the block lives. The trailing block also carries the
  original source path (`/src/next/data/art-group6/logo-cam-ag.go`) and a
  `buildactor`/tool stamp — debug payload the linker never reads.
- **Version 3** (compiled GOAL code): the link block is at the *head*
  `{tag, linkSize, version=3}` with the object at `file + linkSize` and the
  stream at `file+12`; processed by `work_v3` (multi-segment: the segment
  table's `{offset, size}` pairs address main/debug/top-level segments). Not
  needed for asset extraction and not reimplemented.

The *wire version* next to the stream is independent of the file version (a v4
file carries a v2 stream; `dir-tpages` is file-v4/wire-v2). `work` dispatches
wire 2 and 4 to `work_v2`, wire 3 to `work_v3`.

### The v2 wire stream

Two phases, resumable across time slices:

1. **Pointer fixups.** If the first byte is 0 it is consumed and the phase is
   skipped. Otherwise: run-length bytes alternating *skip N words* / *relocate
   N words* over the object, starting in skip mode; relocating adds the
   object's base address to the word. A byte of 255 extends the current run
   without toggling the mode — unless followed by a 0 byte, which is consumed
   and toggles. After any non-255 byte, a following 0 byte is consumed and
   ends the phase.
2. **Symbol links.** Records until a 0 control byte. Control byte with bit 7
   set: intern a *type* named by the following NUL-terminated string, with
   `max(c & 0x7F, 1)` methods (`intern_type_from_c`); the patch target is the
   type object's address. Any other control byte: intern a *symbol*
   (`intern_from_c`); the target is the symbol cell's address. A control byte
   ≥ 10 is itself the first character of the name (the stream backs up one
   byte). After the name, a patch list (`c_symlink2`): delta-encoded offsets
   between patched words, starting at the object base — the low 2 bits of the
   first byte select a 1/2/3/4-byte little-endian varint and are then masked
   off. A patched word of `0xFFFFFFFF` is replaced by the target address; any
   other value keeps its high halfword and gets `target − s7` added to its low
   16 bits (an `$s7`-relative instruction operand — `s7`, the symbol base,
   is `0x149E04`). A 0 byte after a patch ends the list.

`finish` flushes the data cache over the object, pops the link-block
allocation, and (flags bit 0) calls `output_segment_load`, (flags bit 2) calls
the object's entry (`base+4`) when its first word carries the `function` type.

The GOAL runtime's own layout conventions follow from the patch kinds: a
`basic`'s pointer points 4 past its type word; `#f`/`#t` are symbol cells like
any other (unset object fields link as `#f` patches); a type reference is the
type object's address.

`extract/goalobj.Link` reimplements both phases. Verified byte-exact against
the running game: `tpage-463` (ART.CGO, 242,128 bytes, 252 pointer fixups, 251
symbol patches spanning both patch kinds and both intern kinds) linked offline
at the live base `0x90ACC0` with the live `-goalsyms` table is identical to
the RAM image the game's own linker produced (`cmp` clean over the full
object, captured at `finish` entry). `logo-cam` (TIT.DGO, 5,760 bytes at live
base `0xC6FB40`) matches on 5,743 of 5,760 bytes; the 17 differing bytes sit
in four small clusters inside its `res-lump` and are runtime state written
after linking by the art-group's `login` method, not linker output.

Two loading paths reach this format. CGO/DGO archive loads pass through
`link_control::begin` on the EE (`saved_link_control` `0x137260` holds the
background loader's control block; the earliest boot links use a stack-local
one). Art groups pulled by id via `load-dir-art-group` — the TIT logo/title
set among them — and the STR-spooled cutscene animations (`ndi-intro`,
`logo-intro`, `logo-intro-2`, … observed live) are placed by
`ultimate-memcpy` from a spool buffer instead and do not cross
`link_control::begin`; their in-RAM images obey the same object layout.

### Art-group anatomy (by relocation report)

The linker's relocation report types every sub-object, giving the art-group
skeleton for free. `logo-cam` (TIT.DGO, 5,760 linked bytes): an `art-group`
header (file-info, name, element table) → `art-joint-geo` with 4 `joint`s (a
camera rig) → a `res-lump` (`effect-name`, property symbols) → three
`art-joint-anim`s (`logo-cam-idle`, …), each referencing 4
`joint-anim-compressed` blocks. `logo` (352,928 linked bytes): 2
`art-joint-geo`s, 102 `joint`s, 2 `merc-ctrl` skinned meshes, 3
`art-joint-anim`s over 153 `joint-anim-compressed` blocks — the animated logo
assembly of the attract sequence. The formats of `merc-ctrl`,
`joint`/`art-joint-geo` and `joint-anim-compressed` are the subject of the
following Parts.

## Part III — Texture pages (in progress)

A `tpage-NNN` object is a texture page: header `{→file-info, →name, id,
texture count, word counts, 3 × {data ptr, size-in-words, dst}}` with the
descriptor pointer array at `P+0x7C`. The `texture` descriptor layout is read
from `adgif-shader<-texture!` (`0x612684`), the engine's own TEX0/TEX1
builder: `+0/+2` s16 w/h, `+4` mip count, `+5` filter, `+6` psm, `+8` clut
psm, `+10/+12/…` per-mip destination block, `+24` clut block, `+26/…` per-mip
tbw, `+0x24` →name. Blocks are page-relative; the runtime allocator adds the
page's VRAM base into the segment `dst` words (tpage-463 lives at block
`0x2141`, and every TEX0 the draw census reports is these fields plus that
base). CLUTs are CSM1, PSMCT32 or PSMCT16, stored inside the page.

`extract/tpage` + `extract/cmd/tpage` decode a page's textures to PNG through
the GS address swizzles (`tools/platform/ps2` exports its sampler arithmetic
as `TexAddr*`/`CSM1ClutXY`). Segment 0 — the always-resident far-mip tier,
which also carries the CLUTs — uploads as one 128-px-wide PSMCT32 IMAGE
transfer at the page base, verified byte-exact against live VRAM windows;
four full textures decode pixel-identical through the oracle's `-gstex`
sampler. The mid/near tiers (segments 1–2) stream in and out at runtime and
do not tile VRAM cumulatively — live VRAM shows tier-2 rows (128-word
stride) inside tier-1's nominal block range — so their VRAM mapping goes
through the engine's tier allocator; solving it is the part's open item.

## Part IV — Merc models (in progress)

Characters and dynamic models render through *merc*: a VU1 microprogram fed
per-fragment VIF chains. The microcode is the `vu-function` object
`merc-vu1-block` (`0x6E4960`): header `{0x7E5, 0x3F3}`, then 0x3F3 quadwords
(16,176 bytes — the full VU1 program memory) at `+16`.
`merc-vu1-add-vu-function` (`0x6E9E64`) uploads it in 127-quadword MPG
chunks through `dma-buffer-add-vu-function` ref tags. The program
disassembles completely with `tools/cmd/disvu`: a prologue that derives
clip/scale constants (`loi 1/255` for colours, ±65535 clamps), a main entry
at `0x128` and three secondary entries at `0x3510/0x3778/0x39E0` (matrix
re-load variants), double-buffered via `xtop`, kicking finished GIF packets
at `0x3EA8/0x3F10`.

The entry protocol: the unpacked input buffer sits at `TOP`, with a control
header quadword block at `TOP+140` (offsets/counts read as
`ilw .x/.y/.z/.w 0..2(vi12)`) and the output area at `TOP+371`. VU1 low
memory holds the per-frame constant block built by `merc-vu1-init-buffer`
(`merc-vu1-low-mem`): transform rows at quadwords `0..7`, the camera matrix
and fog constants at `132..139`.

In the `logo` art group, each `merc-ctrl` (`+0x1240`, `+0x29ce0`) carries a
header `{name?, size, #f, effect-count}`, a float block (scales/bounds), an
effect table of 32-byte records `{…, fragment count, …, →extra-info}`, and
the fragment streams — tightly-ranged byte records (compressed vertex data).
`draw-bones-merc` (`0x6EC794`) walks the control per frame and emits the
chain, defining the wire format. The effect table sits at basic+108 (count
at basic+52), one 32-byte `merc-effect` record per effect: `{+0 →geometry
stream, +4 →fragment-control stream, +18 fragment count, +22 triangle
count, +24 dvert count, +28 →extra-info}`. A `merc-fragment-control`
record is `{u8 byte-NUM, u8 lump-NUM, u8 fp-QWC, u8 matrix-count}` followed
by matrix-count `{u8 palette-index, u8 vu-dest}` pairs. The geometry
stream holds, per fragment and back to back: a V4-8 region of
`ceil(byte-NUM/4)` quadwords unpacked to `TOP+140` (the VIF immediate is
`0xC08C`: address 140, unsigned, TOPS-relative — its head is the
`merc-byte-header`, whose byte k lands in component k%4 of quadword
140+k/4), a second V4-8 "lump" region, and a V4-32 float region. Each
bone matrix is a 7-quadword ref from a 128-byte-stride palette (built by
`bones-mtx-calc`) to the VU address named in the pair; an `MSCAL` closes
every fragment, and a bucket's first fragment re-uploads the low-memory
constant block (8 quadwords from `*merc-bucket-info*` plus ctrl+28).
`extract/merc` parses all of this; both `logo` merc-ctrls tile their
geometry streams exactly — each effect's fragments end at the next
effect's start to the byte, 356 fragments in total.

### The vertex format and strip topology (complete)

The lump region packs one vertex per three quadwords — twelve source bytes
`{q0: ctlA, adc-ctl, nz, px | q1: dest1, dest2, ny, py | q2: s, t, nx, pz}`.
The w-lane bytes are the position on a fragment-local 8-bit lattice; the
fragment origin arrives in the fp region's first quadword as
integers-in-floats (bias `0x4B010000` = 8454144.0, sign-folded), so world
position = lattice byte + origin. The z-lanes are the normal, and q2's x/y
lanes are texture coordinates. The fp region otherwise opens with the
fragment's adgif (material) block, copied verbatim to the head of the
output packet.

Topology needs no emulation — the file states it, in three mechanisms read
off the microprogram (`mercmicro.dis`) and verified against it:

- **Scatter order.** The two dest bytes of q1 are quadword offsets into the
  fragment's output packet; each vertex stores its `{ST, RGBA, XYZF2}`
  triple at one or both (dest1 = dest2 for ~93% of vertices; distinct dests
  are strip stitches). The GS consumes the packet in address order, so
  ascending dest order is the strip order. Gaps between dests hold
  mid-strip A+D packets (material switches), not vertices.
- **Per-write ADC.** q0's y-lane byte, biased `0x80` (through the VIF-row
  16-bit truncation trick: the row magic's low half is `0xFF80`), controls
  the strip-restart bit per write: positive = both writes clean, zero = the
  dest1 write gets ADC and the dest2 write is clean (the microcode rebuilds
  the position register between the two stores, `0x718–0x748`), negative =
  both writes ADC.
- **Stitch-copy tables.** At fragment end (`0x3C40–0x3E90`) the microcode
  runs a copy table from the byte header: `hdr[0]` = table offset in
  quadwords, `hdr[7]` entries copy within the fragment's own output,
  `hdr[8]` entries copy from the previous fragment's output (the other
  DBF buffer half). Each 4-byte entry is `{src dest, dst dest, ?, b3}`;
  the copy's ADC is the source's ADC XOR b3 (the tail re-derives the bit
  via `itof15.w` plus an `add.w` of b3 — adding 1.0 lands exactly on the
  mantissa bit that is the ADC flag).

The remaining header bytes fall out: `hdr[1]` is the quadword offset of the
per-vertex RGBA records (alpha `0x80` = 1.0 — the "flag" byte of earlier
sessions is just alpha), immediately after the copy table; `hdr[2]` the
byte-region NUM; `hdr[3]` = byte+lump NUM (fp offset). Strips span
fragments — mid-strip packets are kicked without a fresh PRIM write — so an
effect is one continuous vertex sequence whose ADC bits gate the kicks.

`extract/merc/strips.go` implements exactly this from disc bytes, and
`extract/cmd/fragprobe` diffs it against the microprogram emulated on the
same fragments (`merc/emulate.go`, index-encoded colour records): all six
effects across both `logo` merc-ctrls agree triangle-for-triangle —
17,142 of 17,142, zero difference either way. The two merc-ctrls are two
variants of the logo model (near-identical bounds, different letter poses)
and export as separate GLB nodes; per-fragment bounding boxes all fit the
255-unit lattice, confirming every origin decode.

Open for Part IV: material binding (adgif blocks → tpage textures) and UV
scale, then the ndi/evilbro/evilsis groups and Part V joint animations.
