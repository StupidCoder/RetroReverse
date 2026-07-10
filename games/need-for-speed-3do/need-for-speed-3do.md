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

## Part IX — the running game: pacing, tasks, and the drive train

Everything in Part VIII has since been built out (task model, signals, I/O
completion, the graphics folio with a software cel engine, SPORT VRAM clears,
Operamath), and the oracle now boots into a *playable* race: the chase-camera
intro runs, the race clock ticks, and the car drives under pad input. This part
records the machinery that makes the game tick — the parts that were invisible
until they were missing.

### The frame clock is the audio folio

The surprise at the heart of the race loop: the game does not pace itself on
VBlank messages or timer I/O — **it sleeps on the audio folio's clock**. The
Portfolio audio folio keeps a 240 Hz tick counter (`AudioTime`), exposes it via
`GetAudioTime()`, and offers a wake-up service: `SignalAtTime(cue, t)` arms a
**Cue** item (type `0x405`) so the folio raises the cue's signal on its owner
when the clock reaches `t`.

The race main loop (task #1, loop head `0x27C8`, state struct `0x3E5D4`) runs:

```
each frame:
  SignalAtTime(cue, GetAudioTime() + 4)     ; SWI 0x4000D — wake me in 4 ticks
  WaitSignal(cueSignal)                     ; cueSignal from GetCueSignal(cue)
  read pads (non-blocking), kick the sim task, bookkeep missed frames
```

At 240 Hz, +4 ticks is exactly one 60 Hz video field — the audio clock *is* the
field clock, kept in a domain where the music/SFX engine can sync to it. The
folio calls involved, pinned from the game's thunks against the audio folio's
user-function table (byte offset = 4 × table index):

| call | offset / SWI | use in the loop |
|------|--------------|-----------------|
| `GetAudioTime()` | vector `-0xA8` | current 240 Hz tick count |
| `GetCueSignal(cue)` | vector `-0x48` | the signal bit `WaitSignal` blocks on |
| `SignalAtTime(cue, t)` | SWI `0x4000D` | arm the wake-up |
| `OwnAudioClock` / `GetAudioRate` | `-0x4C` / `-0x3C` | clock ownership / rate (frac16 240.0) |

While these were stubbed the loop's `WaitSignal` mask was 0 and `GetAudioTime`
returned a constant — the loop spun, every time delta read zero, and the whole
world stood still while the renderer faithfully redrew the same frame. The HLE
(`audiofolio.go`) drives the clock 4 ticks per virtual field and fires due
`SignalAtTime` requests as real signals.

Two OS details mattered as much as the clock itself:

- **`WaitSignal` consumes what it returns.** On wake, the kernel clears the
  awaited bits from the task's signal word and returns them (`signal.c`); bits
  *not* being waited on stay pending. An HLE that leaves the bits set makes
  every subsequent once-per-frame wait return instantly — pacing silently
  degenerates into a free-running spin.
- **Signals only wake a task if they intersect its wait mask.** A task blocked
  on its message-port signal cannot be nudged with a different bit; the game's
  pad thread relies on a *pad event* arriving to fall out of its menu-mode
  block before its race-mode kick signal means anything.

### The per-frame task pipeline

The race is four cooperating tasks in a signal chain, one link per frame:

```
main #1          sim #4190             render #4196          VBL service
audio-clock  →   SendSignal 0x100  →   SendSignal 0x100  →   (WaitVBL field
wake, pads       world update          MapCel + CCB lists     waits, page
                 (30 Hz car physics)   for the cel engine     flips, SPORT)
```

- The **sim task** computes per-frame flags from its frame counter
  (`[0x41D2C]`: bit 0 = even frame → run the 30 Hz vehicle dynamics, bit 2 =
  every 4th frame) — the car physics deliberately runs at half the frame rate.
- The **render task** transforms the world through the Operamath folio and
  builds the CCB lists (`MapCel` at `0x3C208` stores the corner-engine
  transform into each CCB) that `DrawCels` consumes.
- A separate **pad thread** (entry `0x154E8`) owns the event-broker listener
  port; in menus it blocks on port events, in-race it is kicked at 30 Hz and
  converts button state into smoothed steering/throttle ramps.

### Operamath is two interfaces, and both matter

The math folio splits its API: the vector/matrix operations are **SWIs**
(`0x50000` MulVec3Mat33, `0x50002` MulManyVec3Mat33, …) but the scalar 16.16
helpers are plain **user-function vectors** on the folio base: `-0x4 MulUF16`,
`-0x8 MulSF16`, `-0xC DivUF16`, `-0x10 DivRemUF16(rem*, d1, d2)`, `-0x14/-0x18`
signed, `-0x1C/-0x20` reciprocals (remainders in 0.32 format; overflow returns
all-bits/max-positive). The renderer leans on the SWIs; **the car physics and
the dynamic cameras lean on the scalar vectors**. With only the SWIs
implemented the road rendered perfectly while every drivetrain torque product
and chase-camera term multiplied to zero — the interior view (a fixed cockpit
transform) worked, the chase camera collapsed to vertical lines, and the car
could rev but never move. One table of eight scalar functions separated "renders
a frame" from "plays the game".

### Input: from broker pods to the gearbox

- The game asks the event broker **what is plugged in** (`EB_DescribePods`,
  flavor 27) and expects a `PodDescriptionList` reply naming a generic control
  pad; with no pods described it decides there is no controller and never
  routes input to the car. The HLE answers with one digital pad.
- Button semantics come from the game's own **control-mapping table**
  (`0x3EE3C`): `+0x14` gas = A, `+0x18` brake = B, `+0x20/+0x24` gear up/down =
  the shoulder shifts, `+0x44` handbrake = X. The same struct holds the mode
  byte (`+0x10`): 0 = digital pad, 5 = attract-demo playback — in demo mode the
  game force-feeds `gas = 0x1E` and steers from recorded ramp data.
- **The car starts in neutral.** The Diablo's manual box means a standing start
  is `shift up, then throttle` — the canonical oracle script is
  `-pad "4000000:start,4300000:0,16000000:rs,16400000:0,17000000:a"`.

### Where the game keeps its world

The car simulation state lives in **VRAM used as plain RAM** (the 3DO lets
programs allocate VRAM like any memory): the player car struct at `0x295CE8`,
the opponent at `0x295804` — `+0x50` current track segment, `+0x1C` a 3×3
frac16 orientation matrix, `+0x3EC` gear state, `+0x3F4` throttle counter. The
race-progress struct (`0x3F970`) holds the start-line segment (1739 for City)
and the race state (1 = pre-start, 2 = racing once a car crosses the line);
the race clock is the sim frame counter (`0x41D24`).

### The GO trigger: the race arms itself off the player's first movement

The "no-input pre-start hold" turned out not to be a missing OS service at
all — it is the game's own start logic, traced end to end:

- After race entry the sim task holds for a **hardcoded 10-frame settle**
  (`[0x41D0C]`, set to `#0xA` at `0x20860`, decremented once per hold frame at
  `0x206E0`). During the hold the world updates run in a pre-start mode
  (`0x206FC` re-initialises the race-progress struct via `RaceProgressInit`
  `0x16BCC` every frame) and, in demo mode only (`[0x3EE4C]` = 5), throttle
  `0x1E` is force-fed. When the countdown hits zero the sim flips to *started*
  (`0x20740`: `[0x41D10]`=1, `[0x41A84]`=1, frame counter `[0x41D24]` reset).
- From then on the frame clock, world update (`0x17904`) and both cars' sim
  slots all run — the "frozen race" of earlier sessions is gone with the audio
  clock, Operamath and pod-description fixes in place. What still waits is
  the **launch stamp** `[0x3F97C]`: the per-frame stats block inside the
  *player's* physics function (`0x128DC..0x12918`, every 4th frame once
  started) writes `frame+1` into it the first time the player car's speed
  scalar (`[obj+0x5C]`) exceeds ~`0xCC0` — i.e. "the player has moved at all".
  That stamp is the base of the displayed race clock (`0xBAA4` reads `+0xC`)
  **and the opponent's AI gate**: the object walk (`0x17A78..0x17AAC`) skips
  the opponent's physics function unless `frame == 0` or
  `0 < stamp < frame`. Poking a brief fake player speed in a no-input run
  fires the whole cascade — stamp, race clock, opponent launching hard and
  pulling out of view — proving the chain.
- Consequences: the opponent launches ~2-3 s after the player first applies
  throttle (its AI ramps up on its own once ungated; it needs no pad input —
  it never touches the steer/gas consumers `0x134E0`/`0x13B58`), and the race
  clock starts when *you* move, which is exactly NFS's drag-style timing. No
  timer, audio-completion or streaming condition writes the stamp — every
  other `0x3F970`-referencing site is a reader (or the race-reset zeroing at
  `0xF2E0`) — so with a literally motionless player the opponent stays put
  under the game's own code. Since the whole car sim is the game's own ARM
  code executing in the oracle, a zero-input run here is bit-faithful to the
  console; the remembered real-hardware "opponent goes after ~3-4 s with no
  input" most plausibly involved an input edge (a held Start release plus a
  throttle blip) or a saved configuration difference — worth a re-test on the
  real machine.
- Canonical "watch the opponent go" script (one tap of gas in 1st):
  `-pad "4000000:start,4300000:0,8000000:rs,8400000:0,9000000:a,9600000:0"`
  — the player rolls a few feet and stops; the opponent tears past into view
  a couple of seconds later.

### Cel dimensions come from the preamble (the noise-columns bug)

The in-race "columns of noise" (crowd/grandstand area, scenery, opponent-car
glitches) were one bug with two layers in the software cel engine:

- **`ccb_Width`/`ccb_Height` are not hardware fields.** `CCB_LDSIZE` tells
  the cel DMA to load the *size registers* (HDX/HDY/VDX/VDY) from the CCB;
  the C struct's trailing width/height words are SDK conveniences the
  hardware never reads (opera_madam.c parses the CCB sequentially and takes
  dimensions from PRE0 `VCNT` / PRE1 `TLHPCNT` only). Our "honour
  ccb_Width/Height when sane" heuristic read leftover `(6,6)` on the race's
  scenery CCBs and cropped 64×64 textures to a 6-row band: the City
  smokestacks became vertical noise columns, and the entire distant-skyline
  panorama (64×64 slices drawn rotated 90°, texture columns → screen
  vertical) collapsed into invisible 6×6 blobs. PRE0/PRE1 are now
  authoritative.
- **A texel fills its projected quad.** The renderer bridged magnified texels
  only along the column (HD) direction, with a floored step count — any
  fractionally-magnified cel (the skyline tiles, the hillside strips at
  ~1.25 px texel pitch) left a moire of unwritten clear-through pixels. The
  fix walks both edge directions (HD and VD) with ceiling step counts, the
  sampled equivalent of the hardware's corner-to-corner quad fill. The
  "dithered" hillsides were this bug, not an intentional pattern: they are
  solid green on hardware, and the chase and driving views now show solid
  terrain, the skyline, and a clean grandstand.

## Part X — the track geometry: from disc to textured GLB

Everything the race renders in 3D — road, curbs, walls, terrain, roadside
props, cars — now decodes straight from the disc (`extract/nfs`), verified
bit-exact against the running game (`extract/cmd/geomoracle`), and ships as
web assets (`extract/cmd/webexport`: a textured course GLB, per-object GLBs
plus a placement manifest, and a textured car GLB).

### The .trk file: sections, slices and a pre-baked streaming window

The loader (0x158D8) reads a fixed **0x16C44-byte head block** (the
"RoadSection" allocation) and publishes the **segment array** at
`[0x4CCEC]` = block+0x13B4: **2400 records × 0x24**. A segment record is one
step of the track spline:

| off | field |
|----:|-------|
| +0x2/+0x3 | width bytes (lane code 0x28B70) |
| +0x7 | slice family — selects the cross-section topology |
| +0x8/+0xC/+0x10 | world position x,y,z (16.16) |
| +0x18 | s16 heading; `<<10` = yaw angle, full circle 0x1000000 |

Segments come 6 m apart; cy1 uses 1920 of the 2400 (the array is padded).
City stage 1 is a 9.7 km point-to-point road, not a loop.

The geometry itself streams: the head block's table at +0x98C gives one file
offset per **group** (4 segments), each a 0x3000-aligned **`TRKD` chunk**
`{magic, size, group#}`. A chunk is a **pre-baked sliding window** — groups
g..g+12 concatenated (13 × 0x258) plus five 0x9C near-window records at
+0x1E78 — so one disc read gives the renderer its whole ±13-group
neighbourhood with no stitching; consecutive chunks repeat 12/13 of their
content. (The double-seek pattern in the stream trace is the ring arena at
`[0x3F2D8]` keeping the window contiguous across its wrap.) For extraction
only each chunk's first group is canonical.

A **group block** (0x258) is a 0x18 header — bytes [2..14) are the **material
remap**: face material m → texture-group index — then 4 rows of 0x90: a
12-byte row header and **11 world-space (x,y,z) 16.16 points**, the slice
cross-section stored absolute (centre, right shoulder/curb/wall pairs, then
the left side). No extrusion math at runtime: the track bakes its own
world vertices.

### Slice topology is static in LaunchMe, shared by all nine courses

Three tables in the executable (offsets = load addresses):

- **0x3F2F0 selector**: 128-byte rows indexed by segment family; 32 slice-type
  ids by distance-from-camera — a draw-LOD ramp. Entry 0 is full detail.
- **0x41164 slice types** (76 B): total vertex count, face count/start, and
  four 0xFF-terminated byte lists — the **per-row vertex subsets**. Sparse
  rows drop the wall points (type 0 alternates 11/7/11/7), so walls repeat
  every second segment.
- **0x402BC faces** (8 B): four vertex indices, material, a cull byte, the
  texture sub-index, and an effect count. Indices address the group's
  **compacted 4-row vertex batch** (concatenated subsets: type 0 rows start
  at 0/11/18/29) and reach into the next group's first row from the type's
  vertex total — wall faces legitimately span two segments.

The world-vertex builder (0x15F70) walks exactly these lists into the batch
the Operamath transform consumes; the per-segment scan (0x1A7DC) draws the
face list with materials remapped through the group header into the packet's
texture groups.

### Textures: the packet tree, navigated like the game does

`Cy1_PKT_000`'s wwww root has five children, addressed by index at load
(0x1BEB8, via 0x1FA44 = "root child k", every offset rebased to a pointer;
**zero offsets inherit the previous entry** — that is the loader's rule, and
group 3 of the slice textures really resolves to group 0):

0. **slice texture groups** (51 slots) — each a wwww of up to 36 cels
   (animation frames); face material → group, texture sub-index → cel.
1. **object billboard textures** (168 slots = 42 × 4 mip cels).
2. **3D scenery objects** — (ORI3 model, SHPM shape) pairs, car-style.
3. **near horizon ring** — 6 panorama cels + 1 preview, 4. **far horizon
   ring** — 6 cels (see below).

### The horizon: two 6-cel panorama rings

Children 3 and 4 are the horizon backdrop, stored **rotated 90°** (cel
columns are screen rows — the engine draws them with a swapped corner
matrix). Child 4 is the opaque **far ring**: sky gradient plus distant
terrain (City's hazy range, Alpine's snowy peaks). Child 3 is the **near
ring**: the treeline/skyline silhouette with a PDEC-transparent sky, drawn
over the far ring. Child 3 carries a seventh cel — an *unrotated* 128×64
pre-composed preview of the two layers stacked (CCB flags 0x47EE4420,
unlike the rings' 0x1FEE4620) — which is what pinned the layer order.
Each ring is six 64×64 cels = a 384×64 panorama for the full circle;
Coastal ships 128-row near cels (a taller band whose ocean drops below
the horizon line) and collapses its far ring to a single 4×4
**destination-shading** cel (PIXC 0x9F00 — it dims the framebuffer, not a
picture). `cmd/webexport` bakes the rings into `models/sky-<id>.glb`, two
concentric camera-centred cylinders (band heights matched to the in-race
frame, ~7-13° above eye level; cone caps continue the rings' edge
colours) that the Studio pins to the camera every frame, DS-viewer style.

### RoadObjects: 64 defs + ~1000 placements

After the head block: `{?, size, ?}` then the "RoadObjects" payload —
**64 defs × 16 B** (`{flat flag, type 1/4/6, texture index, extents}`;
type 4 = upright billboard of W×H, flat = ground quad, type 1 = 3D model,
type 6 = two-anchor cel) and **placements × 16 B**, sorted by segment:
`{u32 segment, def, yaw byte, anim flag, s16 dx/dy/dz}`. World position =
segment position + offsets<<8 (the builder at 0x16118); billboard yaw =
−(yaw<<16) − (heading<<10). cy1 places 813 objects from 29 distinct defs.

### The car texture binding: SPoT records and the "!ori" face map

A car `.wrapFam` pairs an ORI3 model with a SHPM shape per LOD. The shape's
directory (`{char4 name, u32 offset}` after "SPoT") holds N textures plus a
**"!ori"** record: after its 0x18-byte header and `[+0x10]`×40-byte records
come 8-byte `{face, spot}` entries — the face → texture map the texset
builder (0x24800) bakes into `[car+0x4E0]`. A **SPoT record** is EA's coded
bitmap: `{u8 type, s24 PLUT offset, u16 w, u16 h, …, pixels at +0x10}` with
the type→bpp table `0,1,2,4,6,8,16` (from the cel constructor 0x3B9C4) and
a 0x20-tagged PLUT chunk; 6bpp pixels are 5-bit PLUT indexes with the
P-bit above, rows word-padded to ≥8 bytes. ORI3 model units are **1/128 m**
(the render path shifts them <<9 into 16.16 world space; the Diablo's ±292
units = 4.56 m). The extra directory names — `whl0`..`whl2`, `bkl1` — are
the wheel-spin frames and brake light the game swaps into the texset at
runtime.

The PLUT offset (and each chain hop) is **signed** 24-bit, relative to the
record — the constructor computes `ADD r9, r9, r2, ASR #8` with
`r2 = word0 << 8`. The player cars never exercise the sign; the **traffic
cars** do: their SHPM lists bare PLUT chunks as directory entries
(`plt0`..`plt3`, four 32-colour recolour schemes per file — the first 14
entries shared chrome/glass/tyres, the rest the body ramp; the game
retargets the cel PLUT pointers per spawned instance) and every texture
points *backwards* at `plt0`. Two more type-byte forms, also from 0x3B9C4:

- **type ≥ 0x80** (bit 7 set): skips the bpp table — the low bits are the
  CCB PRE0 word directly (`PRE0 = type & 0x17`: bits 0-2 bpp code, bit 4
  uncoded) and the path ORRs **CCB_PACKED** (0x200) into the template
  flags: the pixel stream is per-line RLE, decoded by the same packed-cel
  walker as the game's 2D art (`threedo.Cel`). 0x84 = packed 6bpp,
  0x85 = packed 8bpp (Jetta's grille, the 512 TR's LOD-2 set).
- **PIXC**: the constructor stamps the cel's pixel-processor word (cel
  +0x30) from a static template block at 0x472D8 (base literal 0x47314 at
  0x3B9C0): default `[L-0x4]` = **0x1F001F00** (pass-through), overridden
  by the PLUT chunk's `[+0xC]` word when nonzero (dormant — no CarData
  file uses it), overridden for **types 5/0x85** by `[L-0x38]` =
  **0x3F003F00** (`LDREQ lr, [r4,#-0x4]`): source × per-pixel AMV — the
  8bpp pixel's top 3 bits scale the palette colour (ch×(AMV+1)>>3). That
  shading is where the traffic cars' window tint and panel gradients
  live; decoded flat they show raw bright palette entries as white
  speckles (the same unapplied-PIXC class as the road-speckle bug).
  Template CCB flags = 0x1FEE4410 (BGND clear → PDEC black transparent).

With those rules every car `.wrapFam` on the disc decodes — 28 of 29; the
sole failure, `CopMust.WrapFam.old`, has type 0x16, an index past the
retail engine's own 7-entry table (dev-era leftover the shipped game could
not draw either). The decoder feeds all of it through `threedo.Cel` — the
same pixel pipeline the oracle draws with.

### The CarData shelf: 8 player cars, a traffic fleet, and the shelf of the unused

`DriveData/CarData` holds 29 car families for 8 playable cars. The traffic
fleet the player weaves through: BMW, CRX, GMC truck, Jeep, Jetta, Lemans,
Pickup, Prelude, Probe, Rodeo, Sunbird, Vandura, Wagon, Axxess — each with
the 4-scheme `plt` recolour set — plus the police Mustang. Never spawned by
anything in the game: **Scooter.WrapFam** (a green scooter with a helmeted
rider, fully textured, "SCOOT200"), **Porsche.WrapFam** (a classic-body 911
at 130 verts/87 faces — *higher* poly than the retail P911's 116/80),
**CopMust.WrapFam.new** (a development iteration of the police Mustang),
and the **Probe94.WrapFam** / **SASCO.WrapFam** pair — both contain the
same "SASCO000" model; Probe94 is a mislabeled file. (The F512TR's internal
model name is "TESTA000" — the 512 TR is bodied as a Testarossa.)

### Verification and deliverables

`cmd/geomoracle` (decode-reimplement compliant: decoders read the disc, the
oracle only checks) traps one scripted drive: the resident segment array
byte-compares 2400/2400; every cross-section point the world-vertex builder
emits (group, row, point + x,y,z at 0x1605C) matches the decoder bit-exact
(606 hits); every placement's world position (0x16184) matches bit-exact
(76 hits); the car texset resolves to the decoder's SPoT choices for all 61
faces (with the 4 wheel/brake swaps recognised as dynamic).

`cmd/webexport -image DISC -o out` emits **all nine courses** —
City/Coastal/Alpine × 3 stages, named by the disc's own announcer audio —
as `models/course-<id>.glb` (49k–70k verts each, zero skipped faces or
objects on any track) with per-course `models/obj-<id>-NN.glb` +
`<id>.objects.json` placement manifests (Ridge Racer-style instancing:
repeated props ship once), the **full 28-car fleet** as `models/car-*.glb`
(highest-detail LOD each, manifest sections Player cars / Traffic vehicles /
Unused vehicles; traffic cars additionally in all four `plt` recolour
schemes, `car-*-pltN.glb` — 64 car GLBs total) and `manifest.json`. Each
course's opening camera is derived from its spline (segment 16, 2.10 m right
of centre, 0.94 m up, 40 m look-ahead), calibrated against the City 1
driver's-eye camera captured from the running game (reproduced to 4 mm).
`cmd/geomprobe` sanity-checks the decoders standalone.
