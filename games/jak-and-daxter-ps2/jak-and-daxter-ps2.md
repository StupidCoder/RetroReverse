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
