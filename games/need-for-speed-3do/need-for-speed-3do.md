# Need for Speed (3DO) — technical reference

**Image:** `Need for Speed.bin` — 771,408,960 bytes, MD5 `b213789b67a3368207a2ebf2c222847a`. Not committed (copyright, 771 MB); supply your own copy.

The 3DO Interactive Multiplayer runs an **ARM60 (ARMv3, big-endian)** CPU and is
CD-based. The toolchain comprises a dedicated big-endian `tools/cpu/arm60` CPU
core and the CD low layer in `tools/platform/psx`. Everything here is derived
from the disc image plus public 3DO platform documentation — never from
game-specific sources (clean-room rule).

- Image: `Need for Speed.bin` — 771,408,960 bytes, MD5
  `b213789b67a3368207a2ebf2c222847a`, single `MODE1_RAW` track (2352-byte
  sectors → 327,980 sectors). Not committed (771 MB); pinned in the top README.

Tooling lives in `tools/platform/threedo` (platform package) and `tools/cpu/arm60` (CPU);
per-game commands in `extract/` (module `needforspeed/extract`).

---

## Part I — the disc: the Opera file system

3DO discs use the **Opera** file system, not ISO 9660. The low layer is the same
as any CD (`tools/platform/psx/cd.go`): the data track is raw 2352-byte sectors whose
2048-byte user area is exposed through a `block()` accessor. On top of that sits
Opera, decoded in `tools/platform/threedo/operafs.go`. Three things distinguish it:

1. **Big-endian.** The ARM60 runs big-endian; every Opera field is a plain
   big-endian word (ISO stores numbers "both-endian").
2. **Volume label at block 0**, identified by record byte `0x01` + five `0x5A`
   sync bytes + version `0x01` (ISO's "CD001" descriptor is at block 16).
3. **Avatars.** Every file and directory exists as one or more identical copies
   scattered across the disc, to cut seek time and add redundancy. A directory
   entry lists all avatar block numbers; avatar 0 is read.

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

Two subtleties on this disc: the copy count is stored as the **highest index**
(so `+1` for the count), and the directory `next` field is a **relative** index,
not an absolute block.

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

- `tools/platform/threedo/operafs.go` — `Open`, `ReadDir`, `Walk`, `ReadFile`, `Entry`.
- `tools/platform/threedo/cmd/operainfo` — `-label`, `-ls`, `-extract PATH -o FILE`.

```
go run ./threedo/cmd/operainfo -ls   "Need for Speed (3DO)/Need for Speed.bin"
go run ./threedo/cmd/operainfo -extract LaunchMe -o LaunchMe "…/Need for Speed.bin"
```

Verified: `operafs_test.go` round-trips a synthetic image (no dependency on the
committed-out disc); on the real disc, extraction is byte-exact (`ANSX.TDDyn` →
1143 bytes of readable physics text; `AppStartup` → the 160-byte startup script).

---

## Part II — image assets: cels and SHPM shapes

Two image formats carry the disc's 2D art. Both decode in `tools/platform/threedo` and
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

- `tools/platform/threedo/cel.go` — `ParseCel`, `Cel.Image()`; `shpm.go` — `ParseSHPM`.
- `tools/platform/threedo/cmd/celdump` — single-file and `-all` batch to PNG.
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

## Part III — the ARM60 toolchain

The 3DO CPU is an **ARM60** — an ARMv3 core, and (crucially) wired **big-endian**,
so instructions and data are most-significant-byte first. `tools/cpu/arm60` is a
dedicated core for this big-endian target.

As an ARMv3 core, the ARM60 has **no** Thumb, no
BX/BLX, no halfword/signed loads, no CLZ/DSP/saturating ops, and `cond==1111`
means the old **"never" (NV)** condition, not the ARMv5 unconditional space. It
runs the classic set: conditional data-processing through the barrel shifter,
MRS/MSR, MUL/MLA (+ long multiplies), SWP, LDR/STR, LDM/STM, B/BL, SWI. The
game's AIF header pins **32-bit mode** (address mode 0x20), so the ARMv3 26-bit
modes are not modelled.

- `tools/cpu/arm60/arm60.go` — decoder + `Disassemble`, big-endian `word()`, `Flow`
  classification; NV instructions are decoded but flow-neutralized so the tracer
  never follows a dead branch.
- `tools/cpu/arm60/cpu.go`, `exec.go` — the execution core: banked registers + mode
  switching, CPSR/SPSR, big-endian memory, barrel-shifter carry, exceptions/IRQ,
  and a `SWI` hook for Portfolio HLE. `Bus` is the same byte interface as `arm`.
- `tools/cmd/disarm60` — linear disassembler; `tools/cmd/codetracearm60` —
  recursive-descent tracer (entry points → reachable code, data gaps marked).

Validated on `LaunchMe` itself. The AIF header disassembles exactly (`MOV r0,r0`
no-op decompress, `BL` self-reloc/zero-init/entry, `SWI #0x11` exit). The entry
at 0x100 is coherent 3DO code: the **Portfolio folio-call convention** — indirect
calls through negative offsets from `r9`, the kernel/data base (`LDR pc,[r9,#-0x78]`)
— a signed-divide helper (`CMP r1,#0x80000000 / RSBCS / RRX`), and stack thunks.
The tracer follows the entry `BL` into the AIF self-relocation loop
(`SUB r12,lr,pc / ADD r12,pc,r12`). `cpu_test.go`/`decode_test.go` cover
data-processing, conditional execution, big-endian load/store, LDM/STM, branch
and return, the SWI hook, mode banking and exception entry.

## Part IV — booting LaunchMe: the ARM60 oracle

`tools/platform/threedo` now includes an AIF loader and an ARM60 machine that boots and
traces the game's code: an HLE'd OS over a stubbed hardware map.

- `aif.go` — `ParseAIF` decodes the executable header (validated on `LaunchMe`:
  image base 0, 32-bit mode, RO 0x3DB4C, RW 0x99C4, BSS 0x14944, self-relocating).
- `machine.go` / `run.go` — a `Machine` (2 MiB DRAM at 0, 1 MiB VRAM at 0x200000,
  Madam @0x3300000 / Clio @0x3400000 stubbed) driving `tools/cpu/arm60`, with the
  `OnStep`/`OnWrite`/`WatchLo..Hi`/`TTY` instrumentation.
- `Need for Speed (3DO)/extract/cmd/bootoracle` — loads `LaunchMe` from the disc
  and runs it.

**Portfolio HLE by synthetic vectors.** The 3DO kernel enters an app with a
register pointing at the folio/kernel base; the app then calls OS services
indirectly through negative offsets from it (`LDR pc, [r9, #-0x78]`). Lacking the
OS, the oracle plants a vector table just below a synthetic base (`r7`=0x180000)
whose every slot jumps into a reserved HLE address window; the run loop intercepts
a PC landing there, logs the folio offset + arguments, stubs a result and returns.

**What the oracle does.** Booting `LaunchMe` from the disc, it executes the AIF
sequence for real: the no-op decompress, then the **self-relocation** routine
(`SUB r12,lr,pc` load-delta, `SWI #0x10`, the relocation-table scan), then the
**BSS zero-init** loop (~63k instructions clearing 84 KB — a good soak test of the
core), then the entry at 0x100. From there it traces the Portfolio startup: the
first folio call (`folio[-0x78]` from 0x118), item calls (`folio[-0x30]` with type
tags 0x10E/0x104), and a memory-probe loop (`folio[-0x1C]` from 0x320 with
descending power-of-two sizes — largest-free-block probing). It runs hundreds of
thousands of instructions without a decode/exec fault.

The machine runs the real game code and reaches the point where it needs OS
services not yet reimplemented. Identifying each folio offset (they are Portfolio
kernel/graphics/memory functions) and reimplementing the ones the game depends on
is the next step, and turns this trace into a full boot.

## Part V — reimplementing the Portfolio folios

Booting further means the oracle must answer the OS calls the game makes, not
just log them. The calls all go through one mechanism, recovered from LaunchMe's
own code: a global at `0x3E1F4` holds the **kernel/folio base**; the clib stubs
(a table around `0x300`–`0x400`) load it and jump `LDR pc, [base, #-N]`, so each
service is identified by its negative offset `N`. (`SWI #0x100xx` is a second,
trap-based path used for a few calls.)

**Functions identified from the game's use of them** (clean-room — read off the
disassembly, not external sources), reimplemented in `tools/platform/threedo/folio.go`:

| offset | function | how it was pinned |
|-------:|----------|-------------------|
| `-0x1C` | `AllocMem(memlist, size, flags)` → ptr | the routine at `0x36000` is a **binary search for the largest allocatable block** — allocate a size, on success raise the low bound, on failure lower the high bound |
| `-0x20` | `FreeMem(memlist, ptr, size)` | the same search frees each probe (`0x36068`) between tries |
| `-0x30` | a lookup returning a struct ptr | 17 call sites; callers immediately dereference the result (`LDR r0,[ret+0x78]`) |

`AllocMem`/`FreeMem` are backed by a real first-fit heap with coalescing, split
into **two pools** — the game keeps VRAM (`flags 0x80000`) and DRAM (`flags
0x10000`) separate: it allocates an 884 KB VRAM framebuffer, then binary-searches
the DRAM pool for its working set. With the pools split to match the hardware (2
MiB DRAM + 1 MiB VRAM), the probe returns consistent real pointers and startup
proceeds through memory setup.

**The boot-retry loop, and the null-base handling.** After memory setup the game
opens **more than one folio**, and only the kernel base is stored where the HLE
plants it; a *second* folio's base global stays 0, so `LDR pc, [0, #-0xC0]` reads
address `0xFFFFFF40`, gets 0, and jumps to 0 — back to the AIF header, re-running
the *whole* AIF header sequence (relocate → zero-init → entry) ~13 times (visible
with `bootoracle -break`, register/`from` logging). In `machine.go`, a read in the
top page returns `hleBase + N` for an access at `-N`, so a folio call through *any*
base — even a null one — lands in the HLE trap window with the correct offset. The
boot then leaves the header loop and runs real initialisation code (some of it copied
into and executed from VRAM), exercising 18 distinct folios.

**The current frontier: async I/O.** It now stalls in a spin-wait —
`LDRB r1,[r4,#0x18]; TEQ r1,#0; BEQ .` — on a completion flag. The preceding code
builds a request and submits it through the SWI folio dispatcher (`SendIO`-style):
on hardware an I/O server task or interrupt sets `[req+0x18]` when done; with no
task/interrupt model the flag never sets. Completing these I/O requests (the 3DO
task/message model, or synchronous HLE of each I/O) is the next milestone —
`bootoracle -hot`/`-break` locate each such wait. The mechanism, the memory folios
and the multi-folio dispatch are done; async I/O + the graphics folio are the
surface between here and a rendered frame.

## Part VI — the car model format (ORI3)

The in-race 3D car models are **ORI3** objects ("3DO object", magic `ORI3`),
carried inside the `DriveArt/CarArt/*.WrapFam` files (an outer `wwww` "wrap"
container) and the player-car `CarData/*.WrapFam` files; standalone `.3OR` files
(e.g. the car shadow) use the same object. Decoded in `tools/platform/threedo/model.go`.

Layout (big-endian, relative to the `ORI3` magic):

| off | field |
|----:|-------|
| +0x04 | u32 object size |
| +0x10 | u32 vertex count |
| +0x14 | u32 vertex offset — vertices are `int32 x,y,z` (12 bytes) |
| +0x20 | u32 face count |
| +0x24 | u32 face offset — faces are 24-byte records |
| +0x28 | char[8] name (`viper000`, `NSX00000`, …) |

Each **face is a quad** (the Madam cel engine textures quads): a 24-byte record
`u32 material, u32 id, u32 v0,v1,v2,v3`. Materials index the "wrap" texture set
in the enclosing WrapFam.

**How it was pinned (clean-room, from the disc):** the Viper header gives 78
vertices at +0x78 and faces at +0x420, and `0x78 + 78·12 = 0x420` exactly — the
vertex array ends precisely where the face array begins. Parsing those 78
vertices, **all 78 are perfectly left-right (X) mirror-symmetric** and their
silhouette is a Viper. A WrapFam can hold several LOD models (the NSX carries
three: 46/87/87 verts).

- `tools/platform/threedo/model.go` — `ParseModels` → `[]*Model{Name, Verts, Faces}`.
- `tools/platform/threedo/cmd/modelrender` — orthographic side/top/front wireframe → PNG.
- `model_test.go` round-trips a synthetic ORI3 (header offsets, vertex triples,
  quad face record) inside a container.

Verified on the real disc: `VIPER.WrapFam` → the Viper, `CarData/ANSX.WrapFam` →
the NSX (three LODs), `Ferrari.WrapFam` → an "AXXESS00" body. (Several opponent
`CarArt/*.WrapFam` slots are byte-identical placeholder `viper000` models — a
game-data quirk, faithfully reproduced.) Texturing the quads with the WrapFam
materials, and decoding the track geometry, are the next steps.

## Part VII — the track format (wwww resource packets)

The tracks decode straight from the disc, like the car
models — no boot required. Each course (Alpine, Coastal, City) is split into
three packets `DriveData/DriveArt/{Al,Cl,Cy}{1,2,3}_PKT_000`, and every packet is
a **`wwww` resource container** (decoded in `tools/platform/threedo/wrap.go`):

- A `wwww` node is `"wwww"`, a u32 count, then that many child offsets **relative
  to the node's own base** (the outer node sits at 0, so its offsets look
  absolute); each child is another `wwww` node or a leaf. `0` is an empty slot.
- Leaves are exactly the formats already decoded: **cels** (`CCB `, road and
  scenery textures), **models** (`ORI3`, the `TRSL` "track-slice" objects) and
  **shapes** (`SHPM`, the horizon backdrops).

Inventory of all nine packets (via `cmd/packetinfo`): 226–297 **cels** each, plus
3–6 ORI3 track-slice models and 3–6 SHPM backdrops per packet — every resource
classified, none unknown. So a whole track's art extracts today:

```
packetinfo -image DISC -path DriveData/DriveArt/Al1_PKT_000     # inventory
celdump   -scan Al1_PKT_000 -o cels/                            # all 269 cels -> PNG
modelrender -f Al1_PKT_000 -o slices.png                        # the TRSL models
```

The Alpine packet decodes to 269 scenery/road textures, three `TRSL` track-slice
meshes and three backdrops. What remains for a full 3D reconstruction is the
**road layout** — the segment/spline table that positions the slices and maps
textures along the course. That reference table is the piece most likely to need
the oracle to boot the race code and watch it consume the packet; the asset
formats themselves are done.

- `tools/platform/threedo/wrap.go` — `ParseWrap`, `Inventory`; `cmd/packetinfo`.
- `cmd/celdump -scan` — extract every embedded cel from a raw packet.

## Part VIII — pushing the boot: the async frontier

Past the memory-setup and the header-retry loop (Part V), the boot reaches the
game's async waits. The relevant kernel gates were identified from the code:
`SWI #0x10001` = **WaitSignal**, `#0x10002` = **SendSignal**, `#0x10007` = the
folio-call trap. The recurring stall is a busy-wait: a routine clears a byte flag
`[ctx+0x18]`, submits an operation, then spins `LDRB r1,[r4,#0x18]; TEQ r1,#0;
BEQ .` until an **interrupt** sets the flag. With no interrupt/task model the flag
never sets.

A `bootoracle -spinbreak` mode was added to explore past these: when the run
detects a closed loop containing a *compare-against-zero*, it pokes the polled
byte to 1 (standing in for the interrupt). It does advance — from 18 to 22 folios
and into more init — but the exploration proved an important point: **skipping a
wait is not the same as completing the work**. Downstream the game walks its own
allocator free-list (`[node+0x20]` = next) and finds garbage, because the awaited
operation never actually produced its result. So `-spinbreak` is a diagnostic
aid, off by default; a plain run stops honestly at the first async wait.

**The road to a real boot** (each a milestone, in order):

1. **A valid kernel base + OS `MemList`.** The game reads the memory list through
   the kernel base struct (`[base+0x98]`→`[+0xA8]`) to build its own allocator;
   giving it a real, consistent list is what stops the free-list corruption.
2. **Interrupt/VBlank/timer delivery** — so the `[ctx+0x18]` completion flags get
   set for real (synthetic VBlank delivery).
3. **Signals + a minimal task model** (`WaitSignal`/`SendSignal`) so cross-context
   waits resolve.
4. **The graphics folio** (VDL, the cel engine) to reach a rendered frame.

The asset formats (cels, SHPM, ORI3 models, the wwww track packets) are all
decoded; this OS surface is what stands between the oracle and watching the game
consume a track — which is where the **road layout table** (Part VII's remaining
piece) will finally be read.
