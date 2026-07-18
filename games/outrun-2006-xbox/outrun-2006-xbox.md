# OutRun 2006: Coast 2 Coast (Original Xbox) — technical reference

**Image:** `OutRun 2006 - Coast 2 Coast (EUR).iso` — 866,779,136 bytes, MD5
`b36b3a2e4985f6a9a77f6e3cdc2b6972`. Not committed (copyright); supply your own copy.

OutRun 2006: Coast 2 Coast (Sumo Digital / SEGA, 2006) is a Direct3D 8 arcade racer for the
original Xbox. This document reconstructs the shipped disc from its bytes alone — no third-party
emulator, no released source, and nothing about the file or executable formats taken from a
game-specific database or reverse-engineering project. The **platform** formats (the XDVDFS
filesystem, the XBE executable) are general Xbox knowledge, the same way ISO 9660 or the PE header
are; every **game** fact below — which files this title ships, which kernel exports it imports —
is read out of this image with the tools in `tools/platform/xbox` and `tools/cmd/xbeinfo`.

The original Xbox is, hardware-wise, a late-1990s PC: a Pentium III (Coppermine) in flat 32-bit
protected mode, 64 MB of unified RAM, and the NV2A GPU (a GeForce 3/4-class part). Titles are XBE
executables that statically link the XDK (Direct3D 8, DirectSound, XAPI) and import only the
exports of `xboxkrnl.exe`, by ordinal. This is **Part I**: the disc image and the executable — the
static tooling that enumerates the assets and produces the kernel-import census that scopes the
machine's kernel HLE. (The CPU core, the kernel HLE, and the NV2A GPU are later phases; see the
roadmap.)

## Contents

* **Part I** — the disc image, the **XDVDFS** filesystem, and the **XBE** executable: the volume
  descriptor and its magic, the binary-tree directory format, the XBE header, its sections, the
  XOR-de-obfuscated entry point and kernel-thunk table, and the imported-ordinal census. *(this
  document)*
* **Part II** — the **machine**: 64 MB of RAM behind the Pentium III's address windows, the XBE
  loaded at its fixed base, every kernel import thunk-patched to a Go trap, and the title's own
  XDK/CRT boot code run until it reaches the C runtime's SSE object initialisation. Covers the
  memory model, the function/data-export split, the empirical way each `xboxkrnl` ordinal is
  identified from its live call site, the SSE core the boot demands, and the scheduler and
  savestate. *(this document)*
* **Part III** — the **NV2A push buffer** (Phase C, in progress): the corrected PFIFO register map,
  the DMA pusher that decodes the command stream and dispatches methods to the graphics engine, the
  register-model corrections (write-1-to-clear interrupts, FIFO-empty status) and CPU opcodes
  (SSE2/MMX) that carry the title through Direct3D device init, and the 128 Kelvin methods it
  submits. *(this document)*
* **Part IV** — the **Kelvin pipeline**: the LLE vertex-program interpreter (its 128-bit dual-issue
  ISA verified against the game's own uploaded program), the vertex front end, the register
  combiners, the texture unit (DXT/swizzled/linear), the rasteriser — and the frontier behind the
  white frame: the writable Z: cache partition whose absence had frozen the menu loader. Ends with
  the first real frames: the SEGA / Sumo Digital / Ferrari cards. *(this document)*
* **Part V** — **the title movie plays**: the five wrongs behind the post-install stall — a
  FILE_OBJECT is a dispatcher object, the per-thread TLS slot, the fictional error map, the XMV
  codec's MMX instruction set, and the `0xF0000000` write-combined window. Ends at the title
  screen, PRESS START and all. *(this document)*
* **Part VI** — the **frame debugger** (Phase D): the Xbox `framedbg` adapter, and the hunt for
  what a frame actually *is* on this console — in which three plausible boundaries are measured
  and refuted, one of them only by the scene that breaks it. *(this document)*
* **Part VII** — the **save-game path**: a reconstructed export table's +5 drift, three ordinals
  renamed off their own call sites, and a "not enough free blocks" screen that was our model lying
  to the game. Ends at LICENSE SELECT. *(this document)*
* **Part VIII** — the **analog pad**: the pressure bytes and the sticks, and nine button names taken
  with a camera rather than from memory — in which the title's own footer names A and B, its
  on-screen keyboard pins the stick axes against the d-pad bit-for-bit, a tidy bit-order pattern
  gets the d-pad backwards, and eleven controls are left deliberately unnamed. `A` is pressed and
  LICENSE SELECT is left. *(this document)*
* **Part IX** — **the save is written**: `NtSetInformationFile`'s length classes derived from the two
  sites that set them, ordinal 198 renamed off its call site, and the save file committed — ending on
  a game bug our own gap detonates, in which an unchecked HRESULT turns a NULL XONLINE singleton into
  a 5.7-million-iteration loop. *(this document)*
* **Part X** — the **swizzled render target**: its 128×128 size sits in the two nibbles a pitch
  surface leaves blank, the raster's colour/zeta/clear writes route through the texture unit's own
  Morton function, and the target turns out to be a reflection map — whose geometry collapses to
  `[0,0,0,w]` on a zero viewport, the next part's question. *(this document)*
* **Part XI** — **the viewport was in the register file all along**: `SET_VIEWPORT_SCALE`/`OFFSET`
  are methods the dispatch latched and dropped — on the NV2A they *are* transform constants c58/c59,
  the slots every 3-D program's screen-space epilogue reads. Three legs pin the aliasing from the
  image alone; the reflection targets render from the derived state, and the depth mapping the
  probe had guessed comes out of the game's own method stream. Ends identifying the next frontier:
  a depth *texture* (0x2E) sampled by a shadow-receiver pass. *(this document)*

---

## Part I — the disc image

### The XDVDFS filesystem

An Xbox disc is **not** ISO 9660. Its filesystem, XDVDFS, is a single volume:

- A **volume descriptor** sits at sector 32 of the game partition (0x10000 bytes past the
  partition base). It carries a 20-byte magic `MICROSOFT*XBOX*MEDIA` at *both* its head (offset 0)
  and its tail (offset 0x7EC), and, at offset 0x14/0x18, the start **sector** and byte **size** of
  the root directory. Sector size is 2048.

- The **partition base** is not fixed. A raw XISO puts the game partition at file offset 0; a full
  "redump" dump prefixes a video partition, so the game partition begins deep into the file. Rather
  than assume, `xbox.Open` scans for the head-and-tail magic pair (a signature no ordinary data run
  reproduces) and takes the base from where it lands. **For this image the base is 0** — it is a
  raw, trimmed XISO, not a full dual-layer dump — so every sector number below is a plain
  `sector*2048` file offset.

- A **directory is a binary search tree**, not a flat list. Each entry is a 14-byte header
  (`left`/`right` subtree pointers in 4-byte units from the extent start, start sector, byte size,
  an attribute byte with bit 0x10 = directory, a name length) followed by an inline ASCII name.
  Entries are 4-byte aligned and never cross a 2048-byte sector boundary; the slack is 0xFF pad. A
  pointer of 0 or 0xFFFF is "no child". Walking a directory means loading its whole extent and
  traversing the tree.

The tree walk reproduces the disc's catalogue: **3,435 files across 91 directories, 862,221,429
bytes**. The top-level layout is immediately legible as an arcade racer's asset set:

| Directory | Holds |
|---|---|
| `/mv` | `.xmv` full-motion video (title screen, attract, licence/tour/race intros) |
| `/Stage` | per-course geometry: `cs_*_pmt.sz`, `coli_*_bin.sz` collision, `scn_env_*` lighting, one subdir per track (`INDU`, `NIAG_R`, `PRIN_R`, …; `_R` = reversed) |
| `/Cars`, `/Chr`, `/Driver` | car models, characters, driver |
| `/Sprani`, `/Sprite`, `/Text` | 2-D UI: `ani_SPRANI_*` sprite animations, fonts, localized text |
| `/Sound`, `/RVT` | DirectSound / XACT audio banks |
| `/Anims`, `/AS`, `/BK`, `/OBJ`, `/OSO`, `/OCP` | motion (`mot_*_bin`), skeletons, backdrops, objects, splines |
| `/Scripts`, `/Common`, `/Media`, `/Ghosts` | game logic, shared assets, media descriptors, ghost-lap data |

Almost every asset carries an `.sz` suffix (a SEGA/Sumo compression container — a later part) and
the `_pmt` / `_bin` infixes distinguish packed-model-table from raw-binary payloads. These are
game formats and are left for a later part; Part I stops at enumeration.

### The XBE executable

`/default.xbe` (3,837,952 bytes) is the title's program — an XBE, a PE derivative. Its header:

```
title:  "OutRun 2006: Coast 2 Coast"  (title id 0x53450088)
base:   0x00010000   image size: 0x652ac0
entry:  0x000454de   (retail keys)
thunks: 0x00248260
```

- **`XBEH`** magic at offset 0; the image base is the retail-standard **0x00010000**.
- The **entry point** (header +0x128) and the **kernel-thunk-table pointer** (+0x158) are stored
  XOR'd with a constant that differs between retail and debug images
  (`entry ^ 0xA8FC57AB` retail / `^ 0x94859D4B` debug; `thunk ^ 0x5B6D40B6` retail /
  `^ 0xEFB1F152` debug). This is not encryption — it is a platform-spec constant, the sanctioned
  clean-room exception, exactly like the PSP's KIRK keys or the DOS phase's go32 base. The parser
  tries both keys and keeps whichever de-obfuscates to an address inside the image. **The retail
  keys win**: the entry lands in `.text` and the thunk table at the head of `.rdata`.
- The title name and id come from the **certificate** (header +0x118), whose name field is 40
  UTF-16LE code units.

The **20 sections** are the classic XDK static-link set — the game's own `.text`/`.rdata`/`.data`,
plus a section per linked XDK component:

```
SECTION    VADDR       VSIZE       FILEOFF     RAWSIZE     FLAGS
.text      0x00011000  0x0016ed54  0x00001000  0x0016ed54  X|preload
XMV        0x0017fd60  0x00027d34  0x00170000  0x00027d24  W|X|preload   (Xbox Media Video decoder)
D3D        0x001a7aa0  0x0001491c  0x00198000  0x00010e38  W|X|preload   (Direct3D 8)
D3DX       0x001bc3c0  0x00003bd5  0x001a9000  0x00003bd4  W|X|preload
XGRPH      0x001bffa0  0x000135c8  0x001ad000  0x0001294c  W|X|preload
DSOUND     0x001d3580  0x0000da7c  0x001c0000  0x0000d80c  W|X|preload   (DirectSound)
WMADEC     0x001e1000  0x0001907c  0x001ce000  0x0001907c  W|X|preload   (WMA decode)
XACTENG    0x001fa080  0x0000ad68  0x001e8000  0x0000ad14  W|X|preload   (XACT audio)
XNET       0x00204e00  0x00013218  0x001f3000  0x00013218  X|preload
XONLINE    0x00218020  0x000273bc  0x00207000  0x000273bc  X|preload
XPP        0x0023f3e0  0x00008e80  0x0022f000  0x00008e80  W|X|preload
.rdata     0x00248260  0x0007f674  0x00238000  0x0007f664  X|preload     (thunk table at its head)
.data      0x002c78e0  0x00389514  0x002b8000  0x000dbd94  W|X|preload
DOLBY      0x00650e00  0x00007180  0x00394000  0x0000716c  X|preload
.data1     0x00657f80  0x000000e0  0x0039c000  0x000000b0  W|X|preload
XON_RD     0x00658060  0x00001df8  0x0039d000  0x00001df8  X|preload
$$XTIMAGE  0x00659e60  0x00002800  0x0039f000  0x00002800  inserted
$$XSIMAGE  0x0065c660  0x00001000  0x003a2000  0x00001000  inserted
s2xrev     0x0065d660  0x00004930  0x003a3000  0x00004930  inserted
.XTLID     0x00661fa0  0x00000b18  0x003a8000  0x00000b18  inserted
```

Each section header gives a virtual address/size and a file (raw) address/size; the parser maps a
VA to a file offset by finding the containing section (or, for the header region, straight off the
base). That mapping is what lets it read the certificate, the section names, and the thunk table
out of their virtual addresses.

### The imported-ordinal census

The **kernel thunk table** at VA 0x00248260 is a NUL-terminated array of DWORDs; every entry has
its high bit set and imports an `xboxkrnl.exe` export by **ordinal** (its low 16 bits). This title
imports **151 distinct ordinals**:

```
  1   2   3   4   8  15  16  17  23  24  37  40  41  42  44  46  47  49  62  65
 67  69  71  74  81  83  84  85  86  87  95  97  98  99 100 107 109 113 119 126
127 128 129 137 139 142 143 145 149 150 151 153 156 158 159 160 161 164 165 166
167 168 169 170 171 172 173 175 176 178 179 180 181 182 184 187 189 190 193 195
196 197 198 199 200 202 203 207 210 211 215 217 218 219 222 224 225 226 228 231
233 234 236 246 247 250 252 253 255 258 259 260 269 277 279 289 291 294 301 302
304 305 308 312 322 323 324 325 326 327 328 335 336 337 338 339 340 343 344 345
346 347 349 353 354 355 356 357 358 359 360
```

This list is the concrete scope of the kernel HLE the machine will need: **exactly these 151
`xboxkrnl` exports** are the ones OutRun's boot path can reach. It is the Xbox analogue of the
PSP's NID census — an up-front, image-derived statement of the OS surface a title actually uses,
so the machine phase implements a bounded set rather than guessing at all ~366 exports.

### Tooling

- `tools/platform/xbox/xiso.go` — the XDVDFS reader: `Open` (magic scan → partition base + root),
  `ReadDir`, `Walk`, `ReadFile`/`ReadFileEntry`, `MD5`. Mirrors the GameCube and PSP disc readers.
- `tools/platform/xbox/xbe.go` — the XBE parser: `ParseXBE` → header, sections, XOR-de-obfuscated
  entry + thunk table, certificate title/id, and the sorted, de-duplicated ordinal list.
- `tools/cmd/xbeinfo` — the CLI. `xbeinfo -image DISC.iso` lists the tree and dumps the
  `default.xbe` census; `-extract /path -o out` pulls a file; `-xbe FILE.xbe` parses a bare XBE;
  `-md5` hashes the image.

```
$ xbeinfo -image "OutRun 2006 - Coast 2 Coast (EUR).iso"
$ xbeinfo -image "…iso" -extract /default.xbe -o default.xbe
```

The tests (`tools/platform/xbox/xbox_test.go`) build a synthetic XBE and a synthetic one-file
XISO in memory — so the parse logic, the XOR-key selection, and the tree walk are exercised on a
clean checkout — plus a `TestRealDisc` that opens the image when present and skips when it is not.

---

## Part II — the machine

Part I read the disc without running a single instruction. Part II runs the title's own code.
`tools/platform/xbox` is now a **machine**: 64 MB of RAM, the shared `tools/cpu/x86` core in flat
32-bit protected mode (the Pentium III), and the NV2A's MMIO aperture. It loads `default.xbe` at
its fixed base, replaces every kernel-import thunk with a trap into a Go handler, and runs the
XDK/C-runtime boot code until it reaches a facility not yet modelled — stopping there and naming
it, so each run is a concrete statement of how far the boot got.

### The address space

The Xbox is a flat 32-bit machine with a handful of fixed windows onto the same 64 MB of physical
RAM. `translate` (in `machine.go`) folds them all onto one backing slice:

```
0x00000000..0x03FFFFFF  physical RAM, identity (the title at 0x10000, stacks, heap)
0x80000000..0x83FFFFFF  cached kernel window    -> phys = va - 0x80000000
0xB0000000..0xB3FFFFFF  uncached kernel window   -> phys = va - 0xB0000000
0xD0000000..0xD3FFFFFF  physical / write-back    -> phys = va & 0x03FFFFFF
0xFD000000..0xFDFFFFFF  the NV2A MMIO aperture (nv2a.go)
```

Kernel bookkeeping — the KPCR that FS points at, the dispatcher and thread objects the HLE hands
out — lives in a reserved band at the very top of RAM, so it never collides with what the title
allocates from below. The XBE's sections load at their virtual addresses (Xbox VAs *are* low
physical addresses), the segment registers are flat (base 0) except FS, which carries the KPCR
base the way the real kernel leaves it.

### Trapping the kernel: functions and data

A title imports `xboxkrnl.exe` by ordinal — its IAT (the thunk table from Part I) is an array of
DWORDs, each the ordinal with its high bit set. The loader would overwrite each slot with the real
function pointer; the machine overwrites it with a unique **sentinel** address in a trap region
(`patchThunks`). When title code does `CALL DWORD PTR [slot]` the program counter lands on the
sentinel; the CPU's per-instruction hook recognises the trap range, decodes the ordinal, runs its
Go handler, and simulates the `__stdcall` return (pop the return address, drop the callee's
arguments). The sentinel bytes are never fetched. This is the x86 analogue of the PSP HLE's
`jr $ra; syscall` stub patch.

Not every import is a function. The kernel also exports **data** by ordinal — the OBJECT_TYPE tags
the object manager compares against, the debugger-present flags, the running tick counters, the
console's version and hardware structures. A title imports these the same way but then
*dereferences* the slot to read the data rather than CALLing through it. The first sign of one was
a fault: `MOV EAX,[slot]; MOV EAX,[EAX]` read straight into the trap region. The resolution is
clean: reads that land in the trap region return **zero** — the safe default for a flag, a handle,
or a pointer that a title mostly checks for bits or passes straight back to an `Nt*`/`Ob*` call —
while execution in that region still dispatches the function. Functions dispatch; data reads zero;
nothing is guessed. (A data export that genuinely needs a non-zero value gets an explicit populated
block instead; only the kernel version struct has needed one so far.)

### Reading the ordinal table off the running title

The `xboxkrnl` export table is fixed platform ABI, but a table reconstructed from memory drifts by
a few entries in several blocks — enough that binding a handler by a guessed number is a coin
flip, and a *wrong* binding is worse than none (a function pointed at a data block is CALLed
straight into it and executes garbage). So every ordinal is pinned the honest way: from its **live
call site**. The argument count comes from the pushes before the `CALL`; the identity comes from
the argument shapes and how the caller consumes the result. A few examples from this boot:

| Ordinal | Pinned as | The evidence |
|---|---|---|
| 255 | `PsCreateSystemThreadEx` | a 10-argument call spawning the main thread (start routine, stack size, two contexts, a CreateSuspended flag) |
| 24  | `ExQueryNonVolatileSetting` | a 5-arg tail-call reading system config: `(index 0x11, Type*, Value*, len 4, ResultLen*)` |
| 47  | `HalRegisterShutdownNotification` | `(&HAL_SHUTDOWN_REGISTRATION, TRUE)` — and it *returns* (so not `HalReturnToFirmware`, which never does) |
| 107 | `KeInitializeDpc` | `(KDPC, routine-in-.text, context)` |
| 149 | `KeSetTimer` | `(KTIMER, negative relative due time as two dwords, KDPC)` |
| 165 | `MmAllocateContiguousMemory` | a 1-arg allocation returning a pointer |
| 184 | `NtAllocateVirtualMemory` | `(base**, zerobits, size*, MEM_RESERVE=0x2000, PAGE_READWRITE=4)` |
| 187 | `NtCreateMutant` | `(handle*, ObjectAttributes, InitialOwner)` |
| 202 | `NtOpenFile` | `(handle*, access, ObjectAttributes, IoStatusBlock, share, options)` |

These pins also *measure* the table's drift: `NtCreateMutant` sits at 187 (where the reconstructed
table puts it), but `NtOpenFile` at 202 is five higher than the reconstruction, and the `Mm` block
(`MmAllocateContiguousMemory` at 165) and the `Ke` block are shifted too. So only ordinals promoted
into a verified list are trusted; the reconstructed names elsewhere are labels for a log line, and
the machine halts rather than run a handler it has not confirmed.

`NtOpenFile` is disc-backed: it resolves an Xbox object path (`\Device\CdRom0\…`, `\??\D:\…`)
against the mounted XISO and streams the file with `NtReadFile`; a path on the HDD
(`\Device\Harddisk0\…`) reports "not found", exactly as a freshly-formatted console would.

### SSE — the Xbox-only CPU piece

The boot runs cleanly through the thread spawn, the DPC/timer/critical-section setup, and the
config queries, and then hits `0F 57` — **XORPS**. The XDK C runtime zeroes an object's float
members with `XORPS xmm0, xmm0` and stores them with `MOVSS`. This is the one piece of the CPU no
DOS-era game reaches, so it is added to `tools/cpu/x86` behind the two-byte page's mandatory-prefix
decode (`sse.go`): eight 128-bit XMM registers and the scalar/packed single- and double-precision
moves, bitwise logic, arithmetic, conversions, and ordered compares. The real-mode SingleStepTests
harness and the protected-mode/x87 suites stay green; two new tests (`sse_test.go`) guard the
scalar and packed paths.

### The scheduler and the savestate

`PsCreateSystemThreadEx` materialises the title's main thread as a saved `x86.CPU` context; the
preemptive scheduler (`sched.go`) runs the highest-priority ready thread for a quantum and switches
between them, following the shape proven on the 3DS. The savestate (`state.go`) is a versioned,
gzip'd `gob` snapshot of the whole 64 MB, the CPU (registers, x87 stack, the new XMM file), the
allocators and clock, and the thread and object graphs — a deep copy that restores bit-identically
and resumes deterministically, as `TestSaveStateRoundTrip` checks against the real disc.

### Where the boot reaches — the first NV2A push (Phase B milestone)

`bootoracle -image "…OutRun 2006…iso"` now runs **130,371 instructions** of the title's own
XDK/CRT/Direct3D-8 boot code — spawning the game's main thread, retiring the launcher thread,
standing up the DPC/timer/critical-section machinery, querying system config, opening and probing
the disc and HDD, reserving the contiguous framebuffer and GPU instance-memory pools, hooking the
NV2A interrupt, and programming the PFIFO — until it **reaches the first NV2A push-buffer kick**
(`DMA_PUT` advanced past `DMA_GET` at PC `0x1B5CAB`). That is the Phase-B goal: the title's own
code has stood up its Direct3D device and submitted its first command. **Twenty-one `xboxkrnl`
ordinals** are modelled, every one pinned empirically from its live call site.

Getting here past the earlier `Mm` frontier turned on four things, each an instance of *read the
call, do not guess the name*:

- The frontier `Mm` call was **`MmAllocateContiguousMemoryEx`** (ordinal 166), a 5-arg allocation
  — the "64-bit split" that made its argument count look ambiguous was a `PUSH ESI` register-save
  the compiler tucked into the argument block so the callee's arg-pop leaves it on top for a
  matching `POP ESI`. The stack math (a `POP ESI` at the epilogue) proves the count is 5, not 6;
  four call sites agree once the save is excluded.
- The reconstructed ordinal table's **`187` was mislabelled `NtCreateMutant`** (3 args). Five live
  call sites each pass it a single handle straight after an open/create — it is **`NtClose`**, and
  the old 3-arg binding over-popped 8 bytes every call, derailing the boot thread into low memory.
- The XBE **entry is a launcher**: it spawns the game's main thread with `PsCreateSystemThreadEx`,
  `NtClose`s the handle, and returns — on hardware into a kernel thread-terminate stub. The boot
  stack needed the `threadExitAddr` sentinel seeded as that outermost return address, so the entry
  retires thread 0 cleanly and the scheduler hands the machine to the main thread.
- Six more ordinals fell out of the Direct3D device-creation path, each read off its call site:
  `MmClaimGpuInstanceMemory` (168), `MmSetAddressProtect` (182), `HalGetInterruptVector` (44),
  `KeInitializeInterrupt` (109), `KeConnectInterrupt` (98), and `HalReadWritePCISpace` (46, a
  read-modify-write of NV2A PCI config register `0x4C`, backed by a byte-addressed config map).

The `Mm`/`Nt` blocks drift a uniform **+5** and the `Hal` block **+2** off the reconstructed table;
the `Ke` block does not drift. Each binding is anchored to a verified neighbour and to the argument
shapes at its live site, never to the table alone.

## Part III — the NV2A push buffer (Phase C, in progress)

The Phase-B "first push" was a fiction. The stub had labelled PFIFO register `0x3220` as
`DMA_PUT`, but `0x3220` is **`CACHE1_DMA_PUSH`** — the pusher *enable* bit. So the title writing
`1` to enable the DMA pusher looked like a push-buffer submission (the reported `DMA_PUT=1 at PC
0x1B5CAB` was that enable write). Corrected against the NV2A register map: the real `DMA_PUT` is
`0x3240`, `DMA_GET` `0x3244`. The **genuine** first kick is a write of `PUT = 0x0128D000` through
the USER-area channel alias `0xFD800040` (`MOV [ECX+0x40],EDX` at PC `0x1AE668`). The channel's DMA
object (PRAMIN `0x11C0`) has base 0 and a ~128 MB limit, so `GET`/`PUT` are plain physical
addresses; physical address 0 holds `0x0128D001`, a **JUMP to 0x0128D000** — the first kick is a
*ring-priming jump* that submits no work and points the pusher at the push-buffer base. Real
commands arrive in later kicks.

The **PFIFO DMA pusher** (`nv2a_pfifo.go`, the analogue of the PICA200 command processor) walks the
stream from `GET` to `PUT`, decoding NV2A command headers — increasing methods
(`word & 0xE0030003 == 0`), non-increasing (`== 0x40000000`), old/new jumps, call, and return — and
dispatches each `(subchannel, method, argument)` to the graphics engine `PGRAPH` (`nv2a_pgraph.go`),
resolving the subchannel's object class through **RAMHT** (`nv2a_ramht.go`). It is driven
synchronously by the `DMA_PUT` write, so `GET` catches up to `PUT` at once and a title polling the
FIFO never sees it full. An unknown command halts loudly.

Reaching the title's first real submission from the kick took four register-model corrections, each
the *idle/empty/no-pending truth* of a synchronous model rather than a guessed value:

- The `DMA_PUT` write is a 32-bit store, delivered a byte at a time. Kicking the pusher on **each
  byte** ran it with a half-written pointer (`PUT = 0x0000D000`) that jumped off into the stack; the
  pusher now runs only once the aligned dword's top byte lands.
- The interrupt-status registers (`PMC`/`PFIFO`/`PGRAPH`/`PCRTC_INTR`) are **write-1-to-clear**. The
  driver acks by writing `0xFFFFFFFF`; storing that made `PGRAPH_INTR` read perpetually pending, and
  its ISR recursed until the stack overflowed *into the Direct3D device object*, corrupting the
  register-base pointer it held. They read **0** — nothing is ever pending, because the engine
  raises nothing.
- `CACHE1_STATUS` (`0x3214`) and `RUNOUT_STATUS` (`0x2400`) report the **`LOW_MARK` (empty)** bit:
  the synchronous pusher has always drained, and the XDK's post-kick "wait for FIFO empty" loop
  needs it.
- The `PFB` flush register (`0x100410` bit 16) **self-clears** — an instantaneous flush.

The x86 core gained the opcodes the Direct3D vertex/matrix path and the XDK fast memory copy use:
`INVD`/`WBINVD`, the `0F AE` group (`LFENCE`/`MFENCE`/`SFENCE`, `LDMXCSR`/`STMXCSR` with a real
`MXCSR`, `CLFLUSH`), the `PREFETCH`/long-`NOP` group, `SHUFPS`/`SHUFPD`, `MOVNTPS`, `MOVDQA`/`MOVDQU`,
and an **MMX register file** with `MOVQ`/`EMMS`. Several more kernel ordinals were pinned from their
call sites: `AvSendTVEncoderOption` (2), `NtFreeVirtualMemory` (199), `NtCreateSemaphore` (193),
`NtReleaseSemaphore` (222), `NtWaitForSingleObjectEx` (234), `ExAllocatePoolWithTag` (15, the
2-argument form), `ExQueryPoolBlockSize` (23), and `KfRaiseIrql`/`KfLowerIrql` (160/161).

**Result:** `bootoracle -gpu` runs **283 million instructions** of the title's own code, and its
Direct3D-8 runtime submits **128 methods across 5 objects** — validated as real NV2A: class
**`0x0097` (NV20_KELVIN_PRIMITIVE / 3D)** alongside the M2MF (`0x39`) and 2D (`0x62`) helper classes;
DMA-context bindings at Kelvin methods `0x180`–`0x1A8`; and identity 4×4 transform matrices at
`0x840`/`0x880`/`0x8C0` (the `0x3F800000 = 1.0` diagonal). These are the device-init pipeline state,
not geometry — no `BEGIN`/`END` or `DrawArrays` yet. The title then goes CPU-bound loading and
initialising resources.

### Past the audio frontier — the title runs its own game threads

The `0xFE801100` fault was the **MCPX southbridge**, a cluster of device apertures separate from
the NV2A. Three of them are modelled as latches in `apu.go` — sparse write-then-read-back stores,
unwritten registers reading `0` and logged once (`RR_APU_TRACE=1` traces all traffic), because we
render frames, not audio, and nothing downstream consumes what the sound library programs:

- the **APU** (audio processing unit) at `0xFE800000`. Its one register with behaviour is the DSP
  counter at `+0x20010` the DirectSound bring-up polls — first to `≥4`, later to `≥0x20`. It is a
  *progressing* counter, not a ready flag: a constant `4` cleared the first gate and spun the
  second forever (the fiction surfaced exactly as the log-once guard intends). It reads the machine
  clock scaled (`tick>>10`), monotonic and savestate-stable.
- the **AC'97 codec** at `0xFEC00000` (the register-access semaphore poll is satisfied by the zero
  default).
- the **USB OHCI** host controller at `0xFED00000` (`HcRevision` reads 1.0, `HcCommandStatus` reads
  0 — every self-clearing command bit already done in a synchronous model; no pads on the root hub,
  input is a later phase like the GameCube SI).

Past audio the boot runs a long tail of subsystem init, and each frontier is a kernel ordinal
pinned from its live call site (the reconstructed table drifts `+5` in the Ke/Mm/Nt/Ob blocks and
`+2` in the Hal/Io blocks — always anchored to a verified neighbour and the live argument shapes).
This session pinned **twenty more**: `FscSetCacheSize` (37), `IoCreateDevice` (65),
`KeQueryPerformanceCounter`/`Frequency` (126/127), `KeRaiseIrqlToDpcLevel` (129),
`KeSetBasePriorityThread` (143), `KeStallExecutionProcessor` (151), `MmGetPhysicalAddress` (173),
`MmLockUnlockBufferPages` (175), `MmQueryAllocationSize` (180), `NtCreateFile` (190),
`NtQueryInformationFile` (211), `NtReadFile` (219, the OVERLAPPED-shape *real* one — the earlier
provisional binding at 203 was never called and was removed), `NtResume`/`SuspendThread` (224/231),
`NtSetInformationFile` (226), `ObReferenceObjectByHandle` (246), `ObfDereferenceObject` (250),
`RtlInitAnsiString` (289), and the `IdexChannelObject` **data** export (357, whose queued-IRP list
head must point at itself).

Two of these turned a fiction into real behaviour. The wait model used to report every
`NtWaitForSingleObjectEx` as already satisfied; that spun a worker thread hot through its
wait-then-check loop while the producer starved, and the boot made no progress past resource load.
`doWaitTimed` is the honest wait now — signalled objects satisfy immediately, otherwise the thread
parks until a signal or a real relative/absolute timeout. And the CPU gained the P6 opcodes the
title's own interlocked and timing code reaches: **CMPXCHG** (`0F B0/B1`), **XADD** (`0F C0/C1`),
and **RDTSC** (`0F 31`), the last scaled by a new `CPU.TSCMul` so the time-stamp counter tells the
same time as `KeTickCount` (367 = a 733 MHz TSC against 2000 instructions/ms).

With all of it the title's **own game threads run**: it opens `d:\text\english_us.bin` (87,952
bytes) off the XISO and streams it through `NtReadFile` — the first real game asset loaded. The
`-gpu` boot now runs **284,019,155 instructions** and reaches **49 distinct ordinals**, halting at
the **crypto frontier**: ordinal 340, which the reconstructed table calls `XcKeyTable` but which is
really **`XcHMAC`** — two wrappers pass it seven arguments and copy a 20-byte SHA-1 digest out of
the last, then compare it (`REP CMPSB`). That is a content/save-integrity path (HMAC-SHA1, a
platform-spec standard algorithm), the next frontier to implement — carefully, since a wrong digest
is a fiction the game would compare against. The render loop that submits `BEGIN`/`END` + vertex
data + `DrawArrays` lies past it, and then the Kelvin pipeline itself (`nv2a_kelvin.go` is a
latch-only stub today): vertex-program interpreter, register combiners, swizzled texture sampling,
banded rasteriser → PNG. Savestate now covers the PGRAPH engine and the PFIFO pusher position so a
render pass resumes with the 128 device-init methods' register state intact.

### Through the crypto frontier — into the game's own runtime

Implementing the crypto library and the kernel/hardware surface behind it carries the title
through its content/save-integrity path, past the audio-codec reset, and into its **main runtime
state machine**.

**The crypto (`kernel_crypto.go`).** These are standard, published algorithms — the Go standard
library supplies the SHA-1 core (as `crypto/md5` already does in the XISO reader); only the *Xbox*
HMAC construction, a documented deviation from RFC 2104, is reproduced by hand. Six ordinals, each
pinned from its live call site:

- **340 `XcHMAC`** — HMAC-SHA1. The Xbox deviation: a key longer than the 64-byte block is
  **truncated**, not pre-hashed as RFC 2104 would. Two library wrappers route through it, one to
  compute-and-copy a digest (`0x20CAC9`), one to compute-and-`REP CMPSB`-compare it (`0x20CB09`).
  The key comes from a data-export slot holding a per-console secret that a disc image does not
  carry, so it reads back zero — a digest self-consistent with anything *this* run signed, which is
  the honest fresh-console outcome (a verify against a real-key disc signature legitimately fails,
  exactly as the hardware behaves when the secret cannot be recovered).
- **335/336/337 `XcSHAInit`/`Update`/`Final`** — streaming SHA-1. The opaque guest context is kept
  host-side keyed by its address and marshalled into the savestate.
- **338/339 `XcRC4Key`/`Crypt`** — standard RC4 keyed by the SHA digest, the 258-byte state kept
  host-side and serialised too.

**The kernel surface behind it,** all read off live call sites: **128 `KeQuerySystemTime`** (fills
an 8-byte time a token builder serialises), **181 `MmQueryStatistics`** (between the verified Mm
neighbours 180/182; the same token builder reads its `AvailablePages`), and an object destructor's
teardown triple — **97 `KeCancelTimer`** + **137 `KeRemoveQueueDpc`** + **17 `ExFreePool`** on a
`KTIMER`/`KDPC` the constructor had built with `KeInitializeTimerEx`/`KeInitializeDpc`.
`KeTickCount` (156) and `KeSystemTime` (154) became live data exports the scheduler advances.

**The AC'97 reset.** Past the crypto the boot spun forever at `0x1DE9EA`, writing `0x02` to a
bus-master **Control** byte (box+0x0B) and looping while the bit stayed set. That bit is **RR**
("Reset Registers"), which real hardware **self-clears** the instant the per-box reset completes; a
pure latch echoed the written 1 back. Reading a CR byte now clears bit `0x02` while leaving the
run and interrupt-enable bits alone — the reset is instantaneous in our model.

**Where it now reaches.** The boot runs into the title's **own runtime state machine** — a
13-state jump table (`0xEA3FE`, state variable `[0x6322D0]`) that drives per-frame updates. That is
deep into the game, well past boot. It halts in state 0's **audio update**: a handler that iterates
a fixed 17-entry voice array (`0x503438`) with no null guard, and the array is empty. The voices
belong to the **WMA music subsystem** (`OR2ED4.WMA`/`OR2ED5.WMA` — the game's own music tracks) and
are built by a subsystem-init dispatcher (`0x362E0` → the construction loop at `0x366E0`) that the
async resource loader (`0x8C4C0`, a 34-entry subsystem table at `0x5AEF94`) drives. Every subsystem
reports loaded (state 7), yet the audio dispatcher never reaches the voice-construction branch — it
routes on a resource object's type field that the WMA/`WMADEC` codec pipeline would populate. So
the main thread outruns music-voice construction and the game's own code dereferences a NULL voice.
The audio worker threads (`DebugThreads` shows entries `0x2AC00` and `0x2B450`) are a message-pump
and a self-suspending worker, both parked. **The next frontier is the WMA music subsystem** — the
first piece of the audio pipeline the render path cannot simply latch past, since the game's own
per-frame update crashes on it before it reaches geometry.

### The WMA frontier falls — and the machine learns to be slow

The audio crash had been read as a codec-dispatch problem; the trace says otherwise. The voice
constructor lives in its own function (`0x36590`, one caller), and the reason it produced NULL
voices was two layers down: the XDK DirectSound device init probes the **AC'97 codec** — deassert
ACLink cold reset in Global Control (`0xFEC0012C` bit 1), then poll Global Status (`0xFEC00130`)
for the primary-codec-ready bit (`0x100`), a thousand tries, twenty milliseconds apart — and our
latch read 0 forever: `DSERR_NODRIVER`, a device-less DirectSound, and a game that never checks
its `CreateSoundBuffer` HRESULTs. The codec is soldered onto every Xbox: ready now rises the
moment cold reset deasserts.

With the codec up the buffers still failed — `E_OUTOFMEMORY` — and the arithmetic is the game's
own: OutRun allocates a **fixed 0x2AE147A-byte (42.9 MiB) contiguous master arena** (the constant
is in its code at `0x8AD4F`), an 0xB4CCCD-byte committed heap arena, and its 6.3 MB image —
**63.1 MB of the console's 64**, leaving ~900 KB for the real kernel and every runtime pool
allocation. Our synthetic reservations (a 2 MiB kernel band, a 512 KiB launcher stack, thread
stacks inflated to 64 KiB) spent that margin three times over, so the title's last honest
allocations failed — one worker thread's stack silently landed at `sp=0xFFF0`, in page zero. The
band is 256 KiB now, stacks honour the caller's `KernelStackSize`, and a failed stack allocation
halts loudly instead of corrupting low memory.

Past audio, three mechanisms had to become *real* rather than convenient:

- **The GP DSP's command mailbox.** DirectSound submits work by writing a command word into the
  DSP's scratch page and spinning until the DSP firmware clears it — no MMIO in the loop at all.
  `apuTick` is that consumer: GP running, pending word, cleared on the next machine tick. The
  moment it landed, PGRAPH went from 128 device-init methods to half a million live ones.
- **The back-end semaphore.** The D3D busy-wait polls PGRAPH `0x400B10` (never CPU-written — both
  image references are reads) against the semaphore value in memory. The Kelvin release method
  (`0x1D70`) now writes its value through the bound semaphore DMA object — the DMA-object decode
  (base = `(w2 &^ 0xFFF) + (w0>>20 & 0xFFF)`) read off this title's own PRAMIN — and mirrors
  `release<<2` into `0x400B10`. Synchronous pipeline: retired before the CPU can look.
- **NT suspension is a count, not a state.** The deepest bug of the session: our NtResumeThread
  set any waiting thread ready — so a producer's `ResumeThread` "completed" the streaming pump's
  infinite wait on its message-queue semaphore. The pump popped a **NULL message**, read a
  zero-length close request through address zero, and double-released a streaming buffer slot;
  the two 64 KB slots' refcounts went negative and the whole CD-streaming engine wedged — which
  is why the UI sprite archives (`spr_font_xst.sz` first) never loaded and the game idled in its
  state 3 forever. Suspension is now an orthogonal `suspendCount`, exactly NT's semantics; a
  resume never satisfies a wait. `NtReadFile` also honours the caller's own async protocol — the
  overlapped wrapper pre-stores `STATUS_PENDING` in the IOSB, and such reads now complete a
  DVD-realistic latency later (the instant-I/O lesson, again).

The sprite decode path then executed the first **CMOVcc** of the project (`0F 42`, a P6
instruction the XDK emits freely) — missing from both the disassembler and the executor, now in
both.

### The vertical blank — the game's own interrupt code runs

What finally separated "renders clears forever" from "draws the game" was the display interrupt.
The swap path parks in `KeWaitForSingleObject` (ordinal 159 — the Ke API takes raw dispatcher
objects, not handles) on a KEVENT inside the D3D device that only the VBlank signals. Instead of
inventing that signal, `interrupt.go` delivers the interrupt and lets the title's own code do the
rest: `KeConnectInterrupt` registers the KINTERRUPT (vector 3, read off the boot's own connect);
a 60 Hz `vblankTick` raises PCRTC_INTR bit 0 — a real write-1-to-clear pending register now, not
an always-0 stub; the ISR runs as a nested frame on the current thread (context saved, stdcall
frame, sentinel return, IF masked, scheduler frozen). The ISR itself named the rest of the
protocol, one halt at a time: `KeInsertQueueDpc` (119) — DPCs run frame-chained after the ISR,
hardware order — then `KeSetEvent` (145) from the DPC onto the exact event the swap waits on,
then `AvSetDisplayMode` (3), where the kernel programs the CRTC scanout and we record
mode/format/pitch/framebuffer as the machine's display state.

**The result: the full flip protocol runs and OutRun renders continuously** — 2.1 million PGRAPH
methods in one run, 23,204 `SET_BEGIN_END` pairs, 324,856 inline-array vertex words, per-draw
vertex-attribute formats, transform-program uploads, the scanout registered at `0x0174C000`
(640×480, A8R8G8B8). `CLEAR_SURFACE` is the first Kelvin method that produces pixels
(`nv2a_frame.go`), and `bootoracle -png` exports the display's color surface: the first exported
frame shows the game's own clear painted onto the scanout. The vertex-program interpreter,
register combiners and rasteriser — the pixels between the clears — are Part IV.

## Part IV — the Kelvin pipeline: the first real frames

The GPU build proper (`nv2a_vsh.go`, `nv2a_vertex.go`, `nv2a_raster.go`, `nv2a_combiner.go`,
`nv2a_texture.go`) clones the PICA200 oracle's structure: latch state, interpret the game's own
shader programs instruction by instruction, rasterise with the honest fixed-function stages, and
halt loudly on anything unmodelled.

### The method surface, read off one survey

A `-survey` run from a mid-loading savestate says everything the loading loop uses: quad-strip
draws through `SET_BEGIN_END` (0x17FC, arg 9) with 28 `INLINE_ARRAY` (0x1818) words per draw —
four vertices of float4 position + D3DCOLOR + float2 uv, exactly what the vertex-array format
registers (0x1760+, values 0x42/0x40/0x22) declare; a 12-instruction transform program uploaded
per draw through the 0x0B00 window (two increasing bursts of 32+16 dwords — the window is a FIFO,
four dwords per instruction slot, the load cursor at 0x1E9C); constants through 0x0B80 with the
cursor at 0x1EA4; one register-combiner stage (spare0 = TEX0 × DIFFUSE — D3D's MODULATE compiled
down); SRC_ALPHA/ONE_MINUS_SRC_ALPHA blending; and a DXT1 512×256 texture — the "presented by"
card itself.

### The vertex-program ISA, verified against the game's own program

The NV2A transform program is a 128-bit dual-issue word (MAC + ILU per instruction, mux values
1=R/2=V/3=C, output mask/bank/address in dword 3, an idle output port encoded as mask 0 bank 1
address 0xFF). The field map was hand-verified against OutRun's own uploaded program before the
interpreter ran a single vertex — instruction 0 decodes as `MOV R1, v0`, instruction 1 as the
dual `MOV oD0, v3 ; RCP R1.w`, instruction 3 as `MUL R2, R1, c0 ; MOV oD1, v4` — a classic
pre-transformed-vertex passthrough, every field landing where the map says. A disasm regression
test pins those four live instruction words.

Two coordinate facts the live stream settled (both the kind of thing a "probably" would have got
wrong):

- **Positions arrive in sample space.** The loading phase's surface is a 320×240 clip with
  anti-aliasing mode 2 (2×2 supersampling) into a pitch-2560 target — but the program's output
  positions already span 640×480: the screen-space epilogue the D3D runtime appends bakes the AA
  scale into its viewport constants. Only the clip and clear rectangles scale by the AA factors.
  (That also resolves the pitch puzzle: the 320×240 pitch-1280 scanout and the pitch-2560 render
  target are simply different buffers.)
- **oPos.w preserves the clip-space w** (the epilogue divides via `RCC` and keeps w), which is
  what perspective-correct interpolation needs — for the 2D loading quads w=1 and the math
  degrades to affine exactly.

### The frontier behind the white frame: z:\MENU.PAK

With the pipeline live, every frame rendered… white. The textures decoded perfectly (pointing the
pipeline's own decoder at the two loading-screen bindings yields the SEGA/Sumo card and the
Ferrari license plate), but every logo quad's diffuse alpha sat at 0 — the fade never ran. The
trail led away from the GPU entirely: the menu loader retries `NtOpenFile("z:\MENU.PAK")`
forever. Z: is the Xbox HDD's utility partition — on a real console it always exists — and the
game unpacks its menu resources there on first boot. Our file HLE answered
STATUS_OBJECT_NAME_NOT_FOUND for the *partition*, which is a fiction (a fresh console is missing
the FILE, not the drive), so the game's installer path could never run and the loading screen
idled with everything faded out.

`kernel_file.go` now backs T:/U:/Z: with a writable in-memory store (savestated), honouring the
`NtCreateFile` dispositions; `NtWriteFile` pinned at 236 (canonical slot, the Nt block's
established +5 drift, `NtReadFile`'s 8-arg OVERLAPPED shape); open file handles joined the
savestate (a pre-existing gap the disc streaming masked — its reads pass explicit offsets — that
a held-open cache file made fatal). One more ordinal fell out of the install path:
**189 NtCreateEvent**, verified from the XAPI CreateEvent wrapper's call site (the EventType is
computed by `SETZ` from bManualReset — the NT inversion, Notification=0/Synchronization=1). And
because an unimplemented-ordinal halt stops with EIP still on the trap sentinel and nothing
mutated, `ClearHalt` + `bootoracle`'s auto-resume turn a frontier savestate into a
fix-and-continue workflow — the halted state re-runs the very call that stopped it.

**A cold boot now runs the game's own first-boot install** — MENU.pak is assembled on the cache
partition (36 KB from the disc's PAK table at 0x24E330), the WMA music beds copy to Z:, and the
frontend archives stream in.

### ★ The first real frames

With the install running, the boot's visual timeline (a filmstrip probe exporting the AA-resolved
render surface every 20M instructions) shows the whole opening sequence rendered by the pipeline:
the **"PRESENTED BY SEGA"** card (~340M instructions), the **Sumo Digital** logo (~380M), and the
**Ferrari Official Licensed Product** card (~420M) — each one the game's own DXT1/swizzled
textures through its own vertex programs and combiner setup, alpha-fading in and out over the
white base exactly as authored. Pinned exports (`SurfacePNG`, 640×480):

- `cold-0340M.png` — SEGA card — md5 `6db12bca529abc5703fbe73405acb91d`
- `cold-0380M.png` — Sumo Digital — md5 `4e85a74013f7dfc42a98ef1bd4fb3e10`
- `cold-0420M.png` — Ferrari license — md5 `87d8487e9cad5319c16d29eb8c256a78`

The frame exports split honestly: `-png` is the display scanout (what the TV shows — during the
install that is still the 320×240 loading mode), `-surfpng` the Kelvin render target, box-resolved
from samples to logical pixels when the surface is anti-aliased.

## Part V — the title movie plays

The "post-install stall" of Part IV unwound into five distinct wrongs, each named by the game
itself once the previous one fell. None of them was the completion-flag handshake the symptoms
pointed at — `[0x2D31C8]` turned out to be the *"intro sequence finished"* gate (set by the title
screen on a START press, or by the attract path's screen-transition handler at `0xE9A8A`), and
everything upstream of it was starving.

### 1. A FILE_OBJECT is a dispatcher object

The XMV movie loader opens `D:\mv\TitleScreen.xmv` (27.9 MB), queues an overlapped 4096-byte
header read, and waits — on the **file handle itself** (XAPI `GetOverlappedResult` with no event
waits on the file object, which the kernel signals at I/O completion). Our file objects never
signalled; the streaming pump (tid 1) parked forever, the last log line of a 2.2-billion-
instruction trace. Now: the guest `DISPATCHER_HEADER` at the handle reads signalled while idle,
`NtReadFile`'s async path de-signals it, and `ioTick` signals it and wakes its waiters.

### 2. Every thread read one garbage TLS slot

XAPI's per-thread state — `GetLastError` first among it — resolves as `[FS:[4] + tlsIndex*4]`,
where the "TLS index" global (`[0x57CFE8]` = −37) is a **negative dword offset from the stack
top** into a TLS area the kernel carves above the initial ESP, and `FS:[4]` (KPCR.NtTib.StackBase)
must be **swapped at every context switch**. Ours was written once at boot, so every thread's TLS
resolved into the middle of the boot thread's live stack. The thread-start thunk (`0x45069`) named
the whole contract: `MOV EAX,FS:[0x28]` (KPRCB.CurrentThread — at prcb+0, our KPCR had it at +4),
`MOV EDX,[EAX+0x28]` (KTHREAD.TlsData), self-pointer + template copy done by XAPI itself. The
kernel's whole job: carve TlsDataSize (PsCreateSystemThreadEx arg 3) at the stack top, point
KTHREAD+0x28 and the swapped NtTib at it. With this in place the first-boot install runs END TO
END: **12 files copied D:\SOUND → Z: (MENU.pak, FE.PAK, and 10 WMAs — TITLE_01, Splash Wave, Last
Wave, Beach Wave, the OR2ED set), every one verified byte-exact against its disc source.**

### 3. The rough error map was a fiction with teeth

`RtlNtStatusToDosError` was `status & 0xFFFF` — so STATUS_PENDING (0x103) became 259
(ERROR_NO_MORE_ITEMS), never 997 (ERROR_IO_PENDING), and the XMV loader's
`GetLastError()==ERROR_IO_PENDING` check after its overlapped read failed on every boot: the
movie was aborted with `0x80070103`, the unmapped 0x103 sitting in the error's own low word as a
confession. Canonical mappings now; unmapped statuses return 317 like the real kernel and log
once.

### 4. The codec's instruction set

With the open succeeding, the movie decoder ran — straight into `unimplemented 0F opcode $EF`:
**PXOR mm0,mm0**. The XMV codec is classic MMX. `tools/cpu/x86/mmxint.go` now executes the whole
packed-integer group (arithmetic, saturating adds, multiplies, pack/unpack, shifts, compares,
PSHUFW/PAVG/PMIN/PMAX/PSADBW/PMOVMSKB, both mm and xmm widths), and the sharpest find: the
no-prefix `0F 2A/2C/2D` are the **packed MMX↔SSE conversions** (CVTPI2PS / CVTPS2PI) — the old
scalar handlers read the wrong register file for one and *wrote a live GPR* for the other. The
decoder's YUV→RGB stage is built on exactly these (PMADDWD colour matrix → CVTPI2PS → MULPS/ADDPS
→ CVTPS2PI → PACKSSDW → PACKUSWB).

### 5. The 0xF0000000 window

The last halt looked like pointer corruption: `MOVQ [EDI],mm0` with EDI=`0xF25A4C00`. It was a
faithful pointer through an unmapped window — Xbox D3D hands out texture/surface pointers as
`0xF0000000 | physical` (the write-combined RAM alias) so CPU blits bypass the cache. One more
case in `translate()`.

### ★ The title screen

With all five in place, a resumed boot decodes and plays the title movie — the game's own XMV
codec running on the interpreted CPU, blitting frames through the write-combined window into a
D3D texture the Kelvin pipeline draws:

- `mv5.png` (SurfacePNG at ~2.3B instructions) — **the "OutRun 2006 Coast 2 Coast" title card,
  palm tree and all** — md5 `0bea502acd2a1f902d429097022116b5`.
- `title-0400M.png` (+400M further) — **the same card with the game's own blinking yellow
  "PRESS START"** — md5 `5439dd95c92d462d03b9c5fbbd8a6c86`. The title screen is fully alive:
  movie background, UI sprite layer, waiting for a controller that does not exist yet.

### Where it stands / next

The movie plays; the completion flag is the intro-sequence gate, set by the title screen on a
START press (`0x13C5E3`, buttons polled from the game's own pad records at `0x5E5158` — no pad
exists yet) or by the attract path's transition handler (`0xE9A8A`). Next: let the intro sequence
play out into attract (the screen-vtable at `0x274250` dispatches `0xE9870`, whose tail sets the
flag), then the frontend menus off `Z:\MENU.PAK` + `FE.PAK`, the 640×480 display-mode switch, and
the input phase (USB OHCI + XID pads) so START can be pressed honestly.

## Part VI — the frame debugger, and what a frame is

Phase D gives the Xbox a `framedbg` adapter (`tools/debug/xboxadapter`), the eighth platform in the
debugger suite and the first x86 console in it. It carries the GameCube adapter's feature set —
frame capture with per-pixel provenance and overdraw, the command scrubber, fast-forward, CPU
stepping with breakpoints and disassembly, memory watches, surfaces, the disc's filesystem,
savestates, resume — minus the pad, because there is no pad yet, and a capability is a promise the
target can back. An input panel that silently swallowed every press would be worse than no panel.

Most of it is the translation the other adapters are. One question was not.

### Which event is a frame?

The suite's standing lesson is that *"which buffer is the frame?" has a different answer on every
platform and getting it wrong always still looks plausible*. The Xbox adds a second half: **where
does a frame END?** The NV2A renders into ordinary RAM — no on-die EFB copied out and wiped as on
the GameCube, no tiled render target in a private aperture as on the 3DS — so the draw target is
readable at any moment and nothing erases it. That makes the *buffer* easy and the *boundary* hard.

Four candidates, each measured rather than reasoned about:

- **`AvSetDisplayMode`**, the kernel ordinal that registers the scanout. It is called from the D3D
  swap path — the Part II survey pinned it there — so it reads exactly like a present. It is called
  **once per boot**: one call in the first 340M instructions, against thousands of frames. It is a
  mode set.
- **`BACK_END_WRITE_SEMAPHORE_RELEASE`**, D3D's fence. It fires **twice** per frame (the values are
  odd and ascend by 2 — once when the batch finishes, once after the swap), so it would have
  reported two frames for every real one.
- **The vertical blank.** A 60 Hz scanout clock that ticks whether or not the title drew: a field,
  not a frame.
- **`SET_SURFACE_COLOR_OFFSET` moving to the next buffer of the swap chain.** This is the
  interesting one.

The game triple-buffers: the colour surface rotates through `0174C000` → `019A4000` → `01878000`.
At the logo phase that rotation is *exactly* one write per frame, in lockstep with the truth — 209
of each over 34M ticks. It renders the SEGA card correctly, scrubs correctly, and reports pixel
provenance correctly. It is wrong.

**At the title screen it fires three times per frame.** The title renders its XMV movie into an
off-screen target at `02B7B200` first (269 re-points, one per frame) and only then draws the frame
itself. A capture bounded by it stops on a buffer nothing has drawn into yet — and reports a clean,
correctly-sized, **blank white frame**. No error, no crash, 183 commands where a frame has 958. The
scene that validates a boundary is not the scene that breaks it.

### `FLIP_STALL`

The honest boundary is the Kelvin method **`NV097_FLIP_STALL` (0x0130)** — what Direct3D's
`Present` compiles to. On hardware it stalls the pusher until the CRTC's flip retires; here the
pipeline is synchronous and there is nothing to wait for, so it is a pure marker. But it is *the*
marker: the only thing in the stream that says "this frame is finished and meant for the screen".
It fires once per frame at both fixtures — 209 at the logo, 269 at the title — and the census that
says so is the whole argument.

Because the flip *precedes* the swap-chain re-point, the picture is captured **inside** the flip
hook, before the method latches, while the colour surface still names the buffer the frame was
built in. The GameCube needs the same discipline for the opposite reason (there the copy wipes the
buffer; here the register moves on). A regression test pins it by asserting a capture contains
geometry that stored pixels and is not one flat colour — reverting the boundary fails it with
exactly that diagnosis.

### Two things the logo scene could not have told us

- **The clear is a command that stores pixels.** `CLEAR_SURFACE` wrote straight to RAM without
  reporting fragments, so provenance said *"no command wrote this pixel"* across every pixel of the
  background — false, and the clear is very often the answer to "why is this pixel this colour".
  It reports now, and a title-screen pixel names both writers: `cmd 283 CLEAR_SURFACE` then
  `cmd 865 SET_BEGIN_END rgba=b6c1db` — the card's light blue over the clear.
- **Fragments arrive in sample space; the picture is in logical pixels.** Both fixtures run with
  anti-aliasing *off*, where the two coordinate systems are the same numbers and any confusion
  between them is invisible. On a 2×2 AA surface — which this pipeline supports and the loading
  phase actually used — three quarters of every frame's provenance would have fallen outside the
  picture and been dropped by a bounds check. The adapter resolves samples to pixels.

### ★ And the scanout is a white rectangle

The debugger's main view states a gap that every pinned PNG in Parts IV and V had walked past:
**`-surfpng`, the draw target, is the only picture that has ever been verified.** The CRTC's
programmed scanout is a **320×240 window on a buffer nothing draws into**, which renders blank
white. Two facts, both measured:

- the title registers a scanout through `AvSetDisplayMode` **once per boot**, with the *loading*
  phase's mode (320×240, pitch 1280, at `0174C000`), and the **640×480 switch it goes on to render
  at never happens** — the frontier already listed as pending above;
- it *does* write `PCRTC_START` (`0x600800`) once per vertical blank from its own ISR — but the
  value is `0xFFFFFB00` **every single time**: `0 - pitch`, a constant, not an address.

So the scanout registers cannot say what is on screen. **The flip can.** The title marks its
presents with `FLIP_STALL`, and the colour surface at that instant names the buffer it means; the
machine tracks that (`RenderPresented`), which is what any emulator does when it knows the present
but not the register behind it. `Display()` is the presented buffer.

The first cut of this adapter wired `Display()` to the programmed scanout and called it honesty —
the two surfaces disagree, the disagreement is the finding. That was wrong, and the way it was
wrong is the lesson. **The debugger's main view rendered a white rectangle while the game drew
perfectly**, which is not a crash, not an error, and is indistinguishable from a broken emulator;
a user ran four thousand frames of it before saying so. Reporting a fact nobody can act on, in the
one place they cannot avoid looking, is not honesty — it is a bug with a rationale. The honesty
belongs in **keeping the programmed scanout as its own surface**, where the gap stays visible and
nothing depends on it, and a regression test now asserts `Display()` is the frame's geometry and
not one flat colour.

### Tooling

- `tools/debug/xboxadapter` — the `debug.Target`. `framedbg -platform xbox` (an `.iso` is
  ambiguous — a PSX disc and an Xbox disc share the extension — so the platform is named rather
  than guessed); `-xbe` picks the executable within the disc. `games/outrun-2006-xbox/debug.json`
  registers the title with the debugger's game library.
- The platform side gained what a debugger needs and the oracle had not: `OnNVMethod` (the command
  hook), `OnFlip`, `StopRequested`, `RunStopAfterNVMethod` (the scrubber's engine — the NV2A's
  pusher runs *inside* the guest's own store to `DMA_PUT`, so stopping mid-list is an armed
  countdown, not a loop that declines the next command), execution breakpoints, `Machine.PC`
  (where the machine is *parked* — `CPU.LinearPC()` reports the instruction being executed, which
  between steps is the one just retired, and a breakpoint tested against it fires one instruction
  late), `ReadRAM` (RAM-only: a memory pane that read the register aperture would service the
  title's own interrupt ack), `RenderScanout`/`RenderDrawTarget` as images, `NVMethodName`/
  `NVMethodDecode`, the bound-texture and RAM-as-a-texture surfaces, and `Image.EntryAt` (a disc
  offset back to the file that holds it).

## Part VII — the save-game path, and a console that was never full

Phase E ended on a halt: `unimplemented xboxkrnl ordinal 218 (NtRemoveIoCompletion)`, reached by
pressing START through the whole real chain. Everything below follows from not believing that name.

### The name is wrong, and the call site says so

`kernel_ordinals.go` is a *reconstructed* export table, and its Nt block drifts +5 — a drift already
pinned by seven verified anchors (184, 189, 193, 199, 211, 222, 234). Under it, 218 is table-213:
**`NtQueryVolumeInformationFile`**. Three independent lines agree, and the third is the one that
settles it:

1. the drift itself;
2. the call site (`0x44498`) pushes five arguments — a handle, an 8-byte IOSB local, a buffer,
   length `0x18`, class `1`. `NtRemoveIoCompletion` takes `(Handle, KeyContext, ApcContext,
   IoStatusBlock, Timeout)`, and a literal `0x18` is not a `Timeout*`;
3. **its caller identifies itself.** `0x4447B` queries this, then calls ordinal 211 twice — class 6
   (`FileInternalInformation`, 8 bytes) and class `0x22` (`FileNetworkOpenInformation`, `0x38`
   bytes) — and finishes `REP MOVSD` with `ECX=13`: 52 bytes, `sizeof(BY_HANDLE_FILE_INFORMATION)`.
   It is `GetFileInformationByHandle`. The one dword it lifts out of *this* buffer is at `+8`, and
   `+8` of that struct is `dwVolumeSerialNumber` — which is `FILE_FS_VOLUME_INFORMATION`'s
   `VolumeSerialNumber`, at its own `+8`, under class 1. Every offset lines up.

The same reasoning renamed 210 (`NtQueryFullAttributesFile`, not `NtQuerySymbolicLinkObject`) and
207 (`NtQueryDirectoryFile`, not `NtQueryIoCompletion`). 258 `PsTerminateSystemThread` is the one
the table got right — the Ps block does not drift — and the site proves it anyway: it is the tail of
the thread trampoline, `CALL [EBP+8]` (the start routine), its return value pushed as the only
argument, then `INT3`. The compiler knew it never returns; `dispatchKernel` now has a `kretNone` for
calls that do not, because simulating a return would pop the *next* thread's stack.

### ★ The console was never full — we were

Past those, the title left the title screen and drew this:

> *Your Xbox doesn't have enough free blocks to save games. You need 120 more free blocks.*

A legible frame, a real screen, and **a lie our own model told it.** The log said why:

```
NtOpen/CreateFile: "U:\" (hdd "U:/") -> not found (disp 1)
NtQueryFullAttributesFile: "U:\" -> 0 bytes, dir=true
NtOpen/CreateFile: "U:\" (hdd "U:/") -> not found (disp 1)
```

The title asks whether the save partition exists — yes, a directory — and then **opens** it, and the
open fails. The HDD store is a flat `key -> bytes` map (`cacheFS`): it has files and no directory
records at all, so a handle onto a partition *root* was unopenable, and had always been. The
consequence is the shape worth remembering: the free-space query is `NtQueryVolumeInformationFile`
with `FileFsSizeInformation`, **which the title never got to make**. There was no wrong number
anywhere. A call that never happened produced a screen telling the player to delete their saves.

`statPath` + a `dir` flag on `fileObject` make partition roots and inferred directories openable, and
the chain unblocked itself one link at a time, each new call halting and naming itself:

| ordinal | what it really is | how it was pinned |
|---|---|---|
| 218 | `NtQueryVolumeInformationFile` | drift + 5 args + `GetFileInformationByHandle`'s `+8` |
| 211 class 6 | `FileInternalInformation` | 8-byte `IndexNumber` -> `nFileIndex{High,Low}` |
| 210 | `NtQueryFullAttributesFile` | 2 args, no IOSB; caller tests `+0x30` bit `0x10` |
| 258 | `PsTerminateSystemThread` | trampoline tail, one arg, `INT3` after |
| 207 | `NtQueryDirectoryFile` | 10 args, class 1, pattern `"*"` — enumerating `U:\` |
| 218 class 3 | `FileFsSizeInformation` | `IMUL [buf+0x10], [buf+0x14]` = the block size |

That last one is `GetDiskFreeSpaceEx` (`0x428FD`): it multiplies `SectorsPerAllocationUnit` by
`BytesPerSector` into a block size, 64-bit-multiplies both unit counts by it, and reports free and
total **bytes**. The title does the blocks arithmetic itself.

Then one CPU gap, and it says where the game had got to: `0F 15` (`UNPCKHPS`) at `0x1EAD4D`, inside
the **WMADEC** section — the menu music. `UNPCKLPS` was there; the high half was not.

With those in place the dialogue is gone and the title stands on **LICENSE SELECT** — "CREATE NEW
LICENSE", the stats panel, the Ferrari badge, the A/B footer (`work/license.png`, fixture
`work/states/license.state`).

### What is derived, what is chosen, and what is a tracer

Three numbers in this part are not facts about the image, and are marked as such in the code:

- **`VolumeCreationTime` is derived.** The XDVDFS volume descriptor carries a FILETIME at `+0x1C`
  that nothing had parsed; on this disc it decodes to 2006, the year the title shipped.
- **`VolumeSerialNumber` is a tracer** (`0xD15C5E41`). Nothing in the image says what a real console
  reports, and this HLE has no volume to be right about. Its only observed consumer files it into a
  struct field — and per the trap E3 recorded, that census is worthless until the struct's own
  readers are watched, so it is deliberately conspicuous rather than plausible.
- **`hddTotalUnits` (4 GiB of 16 KiB blocks) is a model choice.** Our store is an in-memory map with
  no size, and a save partition's size is a property of the console, not of the disc. Free space at
  least *tracks* the store's real contents. The only constraint the title has ever placed on it is
  120 blocks — and it only said so because it could not ask.

### E8 closed: the pad rides in the savestate

Phase E's own notes ended `DO NOT take a fixture past enumeration before closing it` — `usbDev`,
`usbDone`, `usbCtrlData`, `usbCtrlOff` were not state. Every fixture from here on is taken with a
pad plugged in, because the menus cannot be reached without one, so this had to close first. It
found two things worth keeping:

- **gob will not encode a nil element of an array.** `[usbPorts]*xidDevice` is nil on three ports of
  any real boot, so the in-memory `LoadState(SaveState())` passed while the *file* path failed. The
  existing round-trip test caught it. Ports are values beside a presence flag now, which also makes
  aliasing impossible by construction.
- **"does it resume identically" is structurally blind to aliasing.** `TestUSBSaveState` writes to
  each side after the copy and demands the other not move — and was mutation-tested by aliasing the
  slices on purpose, which it catches in both directions.

Verified end to end: `license.state` restores with the pad on port 0 **and START still held
(`0x0010`)**; the pre-pad `title-phaseE.state` restores an empty hub. The probe discriminates, so it
is not measuring nothing — the first attempt at it, grepping a trace for enumeration traffic,
returned "no enumeration" for both states because `usb_xid.go` has no trace output at all.

### Where it stands / next

- **A/B/X/Y are analog buttons** and `SetPadButtons` only carries the digital level (d-pad, START,
  BACK, sticks). The LICENSE SELECT footer says `A SELECT`, so the next press needs the XID report's
  analog bytes wired up — `-keys` cannot spell `a` yet. *(Part VIII does this.)*
- The 320×240 scanout gap (Part VI) is unchanged and unrelated: `-surfpng` is still the only
  verified picture, and the frames above are all draw targets.
- Unexercised and honest about it: `NtQueryDirectoryFile` on a *disc* directory, every
  `FileFsSizeInformation` field but the two the block-size multiply reads, and the wildcard matcher
  beyond `"*"` (unit-tested against the manual's lane semantics, not against a run).

---

## Part VIII — the analog pad, and letting the game name its own buttons

Part VII ended unable to press `A`. The pad could say eight digital bits; `A` is not one of them —
it is a **pressure byte**, and the report had eight of those plus four signed stick axes that
nothing could drive. This part wires them, and the interesting half is not the wiring. It is that
`usb_xid.go` forbids the obvious way to name them:

> *I know the shape of a Microsoft XID gamepad. Writing that shape down and watching the title
> accept it would prove nothing at all — the game would work, the screenshot would be right, and
> the model would rest on a memory instead of on evidence.*

A pad whose buttons are all mislabelled enumerates exactly as well as one whose buttons are right.
So every name below was **asked of the title**.

### The remap, read whole

`0x14630` is the title's own translation of `XINPUT_GAMEPAD` into its internal button mask, and
Part VII had only sampled it. Read end to end it names every field's *game bit*:

| gamepad | kind | title's game bit |
|---|---|---|
| `+0` bits `01 02 04 08` | d-pad | `0x40 0x20 0x80 0x100` |
| `+0` bit `10` | START | `0x01` |
| `+0` bits `20 40 80` | digital | `0x200 0x100000 0x08000000` |
| `+2 +3 +4 +5` | pressure | `0x02 0x04 0x08 0x10` |
| `+6 +7 +8 +9` | pressure | `0x800 0x400 0x1000 0x2000` |
| `+0xA` / `+0xC` | axis ± | `0x40000/0x80000` / `0x10000/0x20000` |
| `+0xE` / `+0x10` | axis ± | `0x2000000/0x4000000` / `0x800000/0x1000000` |

Two corrections to Part VII fall out. The analog buttons are not "+6..+9": **all eight** bytes at
`+2..+9` are thresholded, each against `0x1E` (`0x147E5`: `MOV CL,$1E` / `CMP DL,CL` / `JBE`). And
there are **four** axes, not one. The single `CMP [ESI+$7], CL` the earlier census noticed was one
line of eight — *a census that finds a reader is not a census that has found the readers.*

**The sticks are a Schmitt trigger**, which is the part worth keeping. Each axis test picks its
threshold off `EDI` — and `EDI` is *last frame's game-bit mask* (`0x14843: MOV EDI,[EBX]`, where
`0x14853` stored it):

```
00014686  TEST ECX, ECX / JZ        zero skips BOTH direction tests: 0 is centred
0001468A  TEST EDI, $00040000       was this direction already on?
00014692  CMP ECX, $00002FFF          ...if so it stays on past 37%
0001469D  CMP ECX, $00005FFF          ...if not it must clear 75% to come on
```

So a fresh direction needs `|axis| > 0x5FFF`. And XAPI adds no deadzone of its own: its cook walks
the eight pressure bytes and stops (`0x243906: PUSH $8`), copying the stick words through untouched.

### Nine names, taken with a camera

The method: plug a pad in at a known screen, drive **exactly one** control to full for fifteen
frames, run to a fixed frame, photograph the render surface. What the screen does is the name.

**LICENSE SELECT** names two buttons in its own footer — `A SELECT`, `BACK B` — and answers with
exactly two bytes. `+2` advances the card into ENTER NAME; `+3` leaves it backwards. The other six
come back **MD5-identical to a run that pressed nothing**, which is far stronger than "nothing
obvious happened".

**The on-screen keyboard**, three `A` presses further in, is a *grid* — and a grid has a cursor.
Each d-pad bit and one stick direction step it identically, to the MD5:

| held | frame | cursor |
|---|---|---|
| wButtons `0x01` ≡ axis1 `+` | `0efe773d…` | up (`1`→`Space`, wrapping) |
| wButtons `0x02` ≡ axis1 `−` | `5b614593…` | down (`1`→`a`) |
| wButtons `0x04` ≡ axis0 `−` | `e6802957…` | left |
| wButtons `0x08` ≡ axis0 `+` | `cfd176f5…` | right (`1`→`2`) |

Two independently-known cases per decode — which is what pins one. Axis 0 is X (**+ = right**),
axis 1 is Y (**+ = up**, the pad's convention, inverted from every screen coordinate here). Axes 2
and 3 move the cursor nowhere: both signs of both return the baseline hash, so this menu does not
listen to the second stick — a fact about the *menu*, since the remap does produce bits for them.

**The pattern was wrong and the picture was right.** The d-pad names were previously *asserted* —
Part VII admitted only START was derived — and they turn out correct, which is luck, not method. It
nearly went the other way: read the remap's game-bit *order* as one vocabulary (d-pad into
`0x20,0x40,0x80,0x100` in wButtons order `02,01,08,04`; the first stick into the adjacent run
`0x10000..0x80000` as `axis1+, axis1−, axis0+, axis0−`) and wButtons `0x02` is obviously "up". The
camera says **down**. That is the whole argument for taking the picture.

### The eleven names this pad does not have

Nine of twenty controls are named. The other eleven — six pressure bytes at `+4..+9`, three digital
bits (`0x20/0x40/0x80`), and the second stick — were **driven to full at both screens** and left
every frame MD5-identical to the baseline (`c5cd1149…`). They are therefore absent from the
vocabulary, not guessed at.

The canonical XID order would name most of them in one line, and the temptation is now *worse* than
before the experiment ran, not better: `A` and `B` landed on the two offsets a remembered ordering
would have predicted, which feels like corroboration and is not — it is two matches, and the same
reasoning-by-pattern just got the d-pad wrong. Two screens being indifferent to a byte is not
evidence of what that byte is; it is evidence that *these two screens cannot tell us*. The frontier
is a screen with more in its footer than SELECT and BACK.

### The vocabulary is the design decision

`padButtons` was `name → uint16 bit`. A pressure is a byte at an offset and a stick is a signed
word, so the table is now `name → PadControl{Kind, Bit|Index, Sign}`, and `PadStateOf(held)` turns
a set of held names into the whole level. **One table, both callers** — `-keys` and the debugger's
keyboard resolve through the same map. The GameCube split this (its stick directions live
adapter-side while its oracle's `-keys` stayed digital-only), and the split is exactly what lets
two callers grow different ideas of what "up" means.

Three details earned their comments:

- **`Fresh`/`Sent` tracked `Buttons` alone.** Generalising them was not tidying: a level change that
  moved only a pressure byte or a stick left `Fresh` clear and was **NAKed away**, so the one part
  of the pad the title reads as an *analog* value would have been the one part that could not
  change. The comparison is over the whole report now, against a `SentReport` that — for the first
  time — something reads. (New *name* as well as new type: gob rejects a field that changes type
  under it, and every savestate this port has taken holds a `Sent uint16`.)
- **The fields are exported**, so they ride `state.go`'s by-value snapshot for free, and gob's zero
  value is genuinely right: `0` is centred *because the title says so* (`0x14686` skips both tests
  on zero), not because zero is a tidy default.
- **Opposite directions accumulate, then clamp.** Assigning would make the answer depend on Go's map
  iteration order — the same input giving a different stick each run, and only sometimes.

### The diagonal is a declared choice, and the octagon is refuted

The GameCube splits a diagonal at `full/√2` because its shell has an octagonal gate — a physical
fact about a specific piece of plastic. **Nothing in this image describes the Xbox's shell**, and
the axes are signed words rather than bytes about a centre, so neither the shape nor the magnitude
transfers. But the image does refute the shape: `0x7FFF/√2 = 0x5A82`, which is **below** the
`0x5FFF` a fresh direction must clear. Split that way a diagonal registers *neither* direction, and
does it silently.

So each axis of a diagonal goes to full — a **square gate**, declared, with its cost declared too:
if the real shell is round, this hands the game a diagonal longer than a real pad can reach. Both
readers of a *direction* threshold rather than measure, so nothing reached so far can tell — but the
title also keeps the raw axis words (`0x1486A` → its per-pad record `+0xAC`), and a driving screen
that steers by magnitude is exactly where this would first be wrong. That is the run that should
revisit it.

### ★ A is pressed, and LICENSE SELECT is left

Through the shipped tooling, from the committed fixture:

```
bootoracle -image "…(EUR).iso" -loadstate work/states/license.state -gpu \
           -steps 300000000 -keys "a@30:10" -surfpng work/entername.png
```

LICENSE SELECT gives way to **ENTER NAME** — a driver portrait, a flag, the default `OR2C2C`, and a
`DONE` button. The stick rides the same chain: `-keys "a@30:10,a@80:10,a@200:10,stickright@260:15"`
opens the on-screen keyboard and steps its cursor from `1` to `2`.

`framedbg` gets the same vocabulary. The **arrows are the left stick** — the Xbox's primary
directional control, and on a menu provably interchangeable with the d-pad (same frames, above) —
so the d-pad keeps a home of its own on the numeric keypad, `8/2/4/6`. `a` and `b` are the two
buttons the title named. Every other letter falls through to a lookup that refuses it, because a key
that quietly did nothing would be worse than one the vocabulary declines. `padPace`'s one-state-per-
flip pacing is untouched: the USB frame is too fast and drains a press inside one game frame.

### Where it stands / next

- **The pad is a real pad**, minus eleven names it has no evidence for. The frontier is a screen
  that names more than SELECT and BACK.
- `pad_test.go` (xboxadapter) pins the stick the way the GameCube's does — an arrow must move the
  stick *and* leave the d-pad bits alone, opposites cancel, and the diagonal must clear the title's
  own `0x5FFF`. That test exists because the GameCube's first cut shipped its stick upside down.
- The 320×240 scanout gap (Part VI) is unchanged: `-surfpng` remains the only verified picture.
- **`+3` (B) leaves LICENSE SELECT into a disc error** — *"There's a problem with the disc you're
  using."* Backing out re-enters a path the model cannot yet serve. That is a real frontier and a
  clean repro, and it is the next thing to pull.

---

## Part IX — the save is written, and a count that was never written

Confirming the license halts on `NtSetInformationFile: unmodelled class 20`. The save path is two
ordinals short, and behind them is a game bug our model walks straight into.

### Bounding the class surface first

Rather than meet each class as a fresh surprise, scanning the image for every reference to the
ordinal's IAT slot (`0x2482B0`) bounds what this ordinal owes the title — 14 `CALL [slot]` plus two
`MOV ESI,[slot]` that call through the register. The class is argument 5, so it is pushed *first*:

| class | len | sites |
|---|---|---|
| `0x04` | `0x28` | `0x4339D`, `0x43FFD`, `0x457E4` |
| `0x0D` | 1 | `0x44031`, `0x44743`, `0x45411`, `0x469C5`, `0x490A7` (a BOOLEAN) |
| `0x0E` | 8 | `0x44378`, `0x4444E` — the position, already modelled |
| `0x14` | 8 | `0x43ED6`, and `0x44280` via `ESI` |
| `0x13` | 8 | `0x442A1` via `ESI` |

A slot scan cannot see the `CALL ESI` sites — which is exactly where `0x13` hides — so the two `MOV
ESI` functions were read by hand. Three sites remain unread. **The scan's VA arithmetic is per
section**: `.text` is `fileoff + 0x10000`, but XONLINE is `+0x11020`, and using `.text`'s formula
everywhere manufactures confident nonsense (it named a checksum routine as a global's writer).

### Class `0x14` is a settable 64-bit file length — derived, not remembered

Two sites say so, and they say different halves:

- **`0x43ED6`** builds its 8-byte buffer from the caller's own QWORD *parameter* — so the value is
  arbitrary and caller-chosen, not a fixed marker.
- **`0x44238`** is the function the frontier arrived on. It reads the current position back with
  `NtQueryInformationFile` class `0x0E` — which this port had already verified *independently* as the
  position, off its own read-back at `0x44321` — and then sets **both** `0x14` and `0x13` to that
  same QWORD. The live run reaches it on `U:\E4B7CAE3D198\common.dat` right after writing
  4 + 67684 + 20 = **67708** bytes, with the position therefore at 67708. The game is saying *the
  file ends here*, having just written exactly that much.

**Nothing in this image distinguishes `0x13` from `0x14`.** The one site that uses `0x13` sets it to
the same value as the `0x14` it issued three instructions earlier, on a file already exactly that
long — so both are no-ops there, and either could be the size and either the allocation. I know what
the NT enum calls them; writing that down and watching the save work would prove nothing (Part VIII's
lesson, one section up). They are modelled identically as *"the file is now N bytes"*, which is
provably right at the only site that exists and is the only reading a `[]byte` with one length can
express. If a caller ever sets them to different values, that is the run that names the difference.

The high dword is *read* and required to be zero rather than ignored — a silent truncation to the low
dword would turn a >4 GB set into a small file, which is the kind of wrong that looks like a working
save. And growth **zero-fills**: `append` does not clear the spare capacity a truncation just freed,
so the naive version hands the guest its own deleted bytes back and every length still checks out.

### Ordinal 198 is `NtFlushBuffersFile`, not `NtOpenSymbolicLinkObject`

34 instructions later. The reconstructed table's name is wrong, as usual; the Nt block's +5 drift
makes 198 table-193, and the call site agrees independently:

```
000445BA  PUSH EBP / MOV EBP,ESP
000445BD  PUSH ECX / PUSH ECX      eight bytes of locals: an IO_STATUS_BLOCK
000445BF  LEA EAX,[EBP-$8] / PUSH EAX
000445C3  PUSH DWORD [EBP+$8]      ...and a file handle. TWO args.
000445C6  CALL [$002482E4]         slot -> 0x8F000C60 = trapBase + 198*16
000445CC  TEST EAX,EAX / JL        NTSTATUS -> BOOL: the FlushFileBuffers wrapper
```

A two-argument file call taking a handle and an IOSB and nothing else is the flush; an
`OpenSymbolicLinkObject` needs a handle out-pointer and an `OBJECT_ATTRIBUTES`. It is a **no-op**,
and that is the truth rather than a stub: `writeFile` commits straight into the store's byte slice,
so there is no buffer between the guest's write and the bytes the HLE holds.

### ★ The save completes

```
NtCreateFile: "U:\E4B7CAE3D198\common.dat" -> hdd CREATED
NtWriteFile: off 0 len 4 / off 4 len 67684 / off 67688 len 20   (file now 67708 bytes)
NtSetInformationFile: "U:/E4B7CAE3D198/COMMON.DAT" length -> 67708 bytes   (class 0x14)
NtSetInformationFile: "U:/E4B7CAE3D198/COMMON.DAT" length -> 67708 bytes   (class 0x13)
```

No halt. The license panel slides off toward the main menu — and stops there.

### The XONLINE singleton is NULL, and the game does not check

The menu never arrives. 200M instructions with **zero kernel calls** — a pure-CPU loop at `0x3F8F0`,
a string copy inside a walk over 0x70-stride records whose cursor has reached `0x184734C` (~25 MB),
far past the 6.6 MB image. The count is the tell:

```
0003F8B0  PUSH ECX          ; MSVC's "reserve one dword" idiom — the local is UNINITIALISED,
                            ; and it happens to now hold `this` = 0x574618 = 5,719,064
0003F8C2  CALL $00218036    ; an XONLINE thunk: MOV ECX,[0x57D6BC] / JMP 0x223A7B
0003F8C7  MOV ECX,[ESP+$C]  ; read the out-count — WITHOUT CHECKING THE HRESULT
0003F8D1  MOV [EBX+$784],ECX ; ...and store it as the record count
0003F8DB  JLE $0003F91B     ; count <= 0 would skip the loop
```

The callee *does* zero the out-count (`0x223AA6: AND DWORD [EAX],$0`) — but only **after** the null
check at `0x223A87`, which bails with `0x80150005` if the singleton `[0x57D6BC]` is NULL. It is NULL.
So the count stays at the uninitialised `this`, and the loop runs 5.7M times over a 16-entry array
(`0x700` bytes, zeroed by `0x223AA1`'s `REP STOSD ECX=0x1C0`), scribbling ~91 MB of strings.

On hardware the singleton exists, the count is written as 0, `JLE` skips the loop, and the empty
license list draws. The game's missing HRESULT check is a latent bug that only fires against a model
that is missing the singleton — **ours**.

`[0x57D6BC]` has exactly **one writer among 71 readers**, and it is conditional:

```
002186E0  PUSH $1BA0 / CALL $234512     allocate 0x1BA0 bytes
002186EE  JNZ ... else E_OUTOFMEMORY
00218705  MOV ECX,ESI / CALL $00218323  Init() -> HRESULT
00218710  JGE $00218724                 success?
00218712  ...destruct, free, ESI = 0    FAILED: the global is never stored
00218724  MOV [$0057D6BC], ESI          the ONLY write
```

A breakpoint on the factory is **never hit** in 900M instructions from `crash.state`, so it runs — or
fails to — at boot.

### It was never the network: it is the raw HDD device

Each step measured on a cold boot rather than reasoned:

| where | what | verdict |
|---|---|---|
| `0xE0EF0` | the factory's only caller, in the game's own init | **reached** at 286M instructions |
| `0x218694` | the XNET gate, `TEST EAX,EAX` after `CALL 0x205322` | **`EAX = 0`** — success |
| `0x21870C` | `Init`'s return | **`EAX = 0x80004005` = `E_FAIL`** |
| `0x2183FF` | the *only* site in `Init` producing `E_FAIL` | guarded by one call |

That guard:

```
002183EF  CALL $00214699
002183F4  CMP EAX, $FFFFFFFF
002183F7  MOV [EBX+$94], EAX     ; the handle the enumerate later reads
002183FD  JNZ $00218409          ; -1 -> fall through to E_FAIL
```

and `0x214699` is `RtlInitAnsiString` (ordinal 289) + **`NtOpenFile` (ordinal 202)** on
**`\Device\Harddisk0\partition0`**, with `OR DWORD [EBP-$4],$FFFFFFFF` on failure. The log said it
outright: `NtOpenFile: "\Device\Harddisk0\partition0" -> not on disc`. We served `\Device\CdRom0`
and `Z:/T:/U:`; the raw disk was served by nothing.

**The XNET gate in the same function returns 0.** The network was never the problem, and modelling
one would have been a large pile of work aimed at the wrong thing.

### ★ Offline falls out; it is not asserted

Serving `partition0` as a blank raw device makes `Init` return S_OK and the singleton land:
`[0x57D6BC]` = `0x0066F010` on a cold boot, where it was `0x00000000`.

**`partition0` only, and that is measured, not tidiness.** A first cut served every
`\Device\Harddisk0\partitionN` and the boot **died after 1,465 instructions** on an unimplemented
ordinal: serving `partition1` diverts the XAPI's own mount away from the path five phases went into
making work.

**What the device says is an invention, and it is declared.** A disc cannot carry an HDD image, so
nothing derives its contents. Only the first half is forced:

1. **The device exists.** Every console has one; refusing it is the fiction — the same argument this
   port already makes for `T:/U:/Z:` ("a missing FILE is still honest; a missing PARTITION is not").
2. **It is blank** — a choice. `Init` and the enumerate each read a 0x1EC-byte record through the
   handle (`0x21844B`, `0x223AC0`) and test a signature (`CMP DWORD [EBX+$1C], $56525347` — `"GSRV"`).
   Zeros fail it and take the branch the code already has. The claim is *"this console has no XONLINE
   account"* — which is what the title already prints across the top of the screen it hung on:
   **NOT SIGNED IN**.

So "offline" is what the model's own shape produces, rather than something the HLE asserts. The risk
is named where the code is, because this port has paid it before: a value our own stub invents comes
back later as an observation.

**Its size is measured too.** Served 4 MiB with every access logged, a full cold boot touches it
**six times and never past `0x1400`** — so `0x10000` is sixteen times the observed footprint, and
reads past the end return short rather than inventing more:

```
NtOpenFile  \Device\Harddisk0\partition0     (twice, closed in between)
READ  off 0x1000 len 0x200      sector 8 of a 512-byte-sector device...
WRITE off 0x1000 len 0x200      ...read-modify-write, twice
WRITE off 0x1000 len 0x200
READ  off 0x1200 len 0x200      sector 9
```

**The guest writes to it**, which changes what "blank" means: this is not a read-only fiction handing
back zeros forever. The title reads sector 8, modifies it, writes it back — so the device is blank
only until the guest fills it, after which it reads back exactly what the guest put there. That is
the behaviour of a real fresh disk, and the reason it lives in the store: those writes ride the
savestate like every other writable byte on this machine.

**And the boot does not notice.** The title frame off a cold boot with the device served is
**MD5-identical** to the committed reference (`5439dd95…`) — the new device is inert on every path
except the one that was waiting for it.

One quiet bug came with it. `volumeUnits` sums **every** entry in the writable store, so the raw
device — which shares that store so it rides the savestate — would have been charged 4 units against
the *save* partition's free space. Four units of 262,144, invisible for a long time, and exactly the
shape of the Part VII bug where the title told the player their console was full over a number the
HLE made up.

### A savestate outlives the gap it was taken in

The fix landed and `crash.state` still froze — correctly. The singleton is created **once**, at boot;
a state captured after that factory has already failed has `[0x57D6BC] = 0` baked into its heap, and
no later build can retroactively create an object the boot never made. Measured under the *fixed*
build:

| state | `[0x57D6BC]` |
|---|---|
| `crash.state` (pre-fix) | `0x00000000` — still NULL, still hangs |
| `title-raw.state` / `license-raw.state` (fresh cold boot) | `0x0066F010` |

Every pre-fix state is a permanent pre-fix repro. This is the same shape as the GameCube's lost-DIRQ
bug, where the whole savestate library had to be re-taken and the writeup had to name the
replacements — so: **`work/states/{title,license}-raw.state`** are the ones that carry the singleton.

### ★ The menus play, and the race loads

From `license-raw.state` the game goes through **game-mode select → track select → opponent select →
soundtrack select** and into the real **"OutRun 2006: Coast 2 Coast — Loading / PLEASE WAIT"**
screen. The frames arrive at normal speed; the 5.7M-iteration loop is gone.

That loading screen ended on one more ordinal, and the reconstructed table misnamed it as usual.
**225 is `NtSetEvent`, not `NtSignalAndWaitForSingleObjectEx`** — and this one is *bracketed* rather
than extrapolated: the Nt block's +5 drift puts table-220 here, between the already-verified 224
(`NtResumeThread`, table-219) and 226 (`NtSetInformationFile`, table-221). A wrong answer would have
to be wrong *between two independently-pinned entries*. The call site agrees on its own:

```
00044D4F  PUSH $00000000        PreviousState* — the caller wants no readback
00044D51  PUSH DWORD [ESP+$8]   ...and an event handle. That is all: TWO args.
00044D55  CALL [$0024834C]
00044D5B  TEST EAX,EAX / JL     NTSTATUS -> BOOL: Win32 SetEvent, whose whole body this is
```

`NtSignalAndWaitForSingleObjectEx` takes five arguments and cannot be spelled with two. And `0x44D4F`
sits immediately after `0x44D25` — the XAPI's `CreateEvent` wrapper, already verified as ordinal 189.
This is that library's event block.

The semantics needed nothing new: signal, wake, report the previous state. `wakeWaiters` runs each
candidate through `satisfyWait`, which makes the auto-reset case right for free — a synchronisation
event has its signal **consumed** by the first waiter, so exactly one thread wakes and the event
clears itself, while a notification event stays signalled and releases everyone. That distinction was
read off `NtCreateEvent`'s own SETZ inversion back in Phase E, not invented here.

### Where it stands / next

`NtSetEvent` carries `gameplay.state` **7× further** — 175M instructions to **1,238,399,686** — and
the frontier leaves the kernel entirely:

```
HALT: nv2a: draw 487118 into non-pitch surface type 2 (0x208=07070228)
      — swizzled targets unmodelled
```

**487,118 draws in**, the loading screen is behind it and the title is rendering the race. The next
frontier is the Kelvin pipeline, not the HLE: the game binds a **swizzled render target** (a
render-to-texture) and the rasteriser models pitch surfaces only. That is Part X's problem, and it is
a good one to have — every remaining kernel gap on this path is now closed.

*(A `-surfpng` at that halt is a small grey square, and correctly so: the colour surface at that
instant IS the off-screen swizzled target, not the scene. The picture is the frontier, not a bug.)*

## Part X — the swizzled render target, and the geometry it exposed

The halt names one register field and one missing address mode. `SET_SURFACE_FORMAT`
(`0x0208`) is `0x07070228`: colour `A8R8G8B8` (nibble 0), zeta `Z24S8` (nibble 1), and — the
byte that stops the run — **type 2** (nibble 2), a *swizzled* surface where the rasteriser
models only **type 1**, pitch. A swizzled surface stores its texels in Morton (Z-order)
interleave rather than row-major; the fill hard-codes `colorPhys + py*pitch + px*4`, which for a
swizzled target addresses the wrong bytes on every pixel.

### The size is in the field pitch surfaces leave blank — three witnesses agree

A swizzled surface has no pitch, so it must carry its own width and height. The two nibbles a
pitch surface holds at zero — bits 16–19 and 24–27 — are `0x07`/`0x07`, and `1<<7 = 128`. Three
independent witnesses pin **128×128** and make this a derivation, not the family resemblance to
the texture format's size field (which is a nibble apart at bits 20–23/24–27, not a byte apart —
reading the surface with the texture's layout gives a 1×128, and the clip refutes it):

- the nibbles read 7/7 → 128×128;
- `SET_SURFACE_CLIP` reads 128×128, the render region;
- the title packs its three targets at `0x01D28200`, `0x01D38200`, `0x01D48200` — exactly
  `0x10000 = 128·128·4` bytes apart, the byte extent of one.

`swizzleGeom` reads the nibbles and cross-checks the extent against the clip. What it does **not**
pin is which nibble is width and which is height: every swizzled surface this title programs is
square, so the assignment is invisible (`swizzleOffset` is symmetric when `w==h`) and pinning it
would need a non-square case the image never presents. So the decode **halts on any non-square
swizzled surface** rather than guess an orientation it cannot verify — the same discipline the
pitch/type check already carried.

### The interleave was already earned — it is reused, and the reuse is the point

`swizzleOffset` (the texture unit's Morton function, `nv2a_texture.go`, derived earlier and
verified against real disc textures) is not rewritten. The colour and zeta writes route through
it — `colorAddr`/`zetaAddr` return `phys + swizzleOffset(px,py,w,h)*4` for a swizzled surface and
the untouched pitch expression otherwise. This is what makes a render-to-texture round-trip: the
raster **writes** the target with `swizzleOffset` and a later draw **samples** it back with the
same function (`decodeSwizzled`), and the picture only survives if the two permutations match. The
zeta buffer swizzles too — it shares the surface's single type field, and it is private to the
raster (written and read only by the depth test, never sampled), so its layout is
self-consistent by construction. `CLEAR_SURFACE` had to learn the swizzle as well, or it would
scatter the clear and leave the swizzle bytes the draws and the depth test actually use full of
stale data. And the export paths (`RenderDrawTarget`, `renderSwizzledSurface`) **de-swizzle** for
the dump — reading a swizzled target row-major would produce a plausible but wrong picture, the
one outcome worse than a halt.

The address math is unit- and mutation-tested on a **non-square** grid (a square target hides an
x/y transposition): a row-major mistake and a transposed interleave each make the suite fail.

### ★ The target is a reflection — and it proves the write

With the swizzle modelled, the run clears draw 487118 and advances ~180 draws to an unrelated
frontier (a texture in an unmodelled format, `0x2E`). But a `-surfpng` of the 128×128 target is
**uniform grey** — the clear colour `0x19808080`, and no geometry on top. The reason is not the
swizzle: tracing the draws, every triangle's transformed position is `[0, 0, 0, w]` — **x, y, z
collapse to zero while w is computed correctly**. The inputs are sound 3-D world coordinates; the
vertex program's disassembly names the cause exactly:

```
DP4 oPos.w,   v0, c163            ; w = dot(pos, c163)          — correct
DP4 oPos.xyz, v0, c160..c162      ; clip xyz into R12
MUL oPos.xyz, R12.xyz, c58.xyz    ; × viewport SCALE  c58  = 0  → 0
MAD oPos.xyz, R12.xyz, R1.x, c59  ; × 1/w + viewport OFFSET c59 = 0  → 0
```

The clip position is right; it is then multiplied by the viewport scale `c58` and offset by `c59`,
both **zero** in this savestate. (The 2-D passthrough pass that draws the framebuffer uses `c0`/`c1`
for its viewport, which are loaded — which is why every pitch frame this port renders is fine.)
This is a **separate frontier from swizzling** — the geometry would be degenerate whether the
target were swizzled or pitch, and the raster changes do not touch the transform. Where `c58`/`c59`
come from (they are set earlier than the reached window, via a path that leaves them zero here) is
Part XI's question.

To prove the swizzle **write** is nonetheless correct, a labelled probe supplies the one missing
piece — the 128×128 viewport in `c58`/`c59` — and the target renders a **coherent 3-D scene**: a
road surface, a horizon, a grandstand roof (`work/swizzled-rtt-reflection.png`). A small dynamic
render-to-texture 487k draws into a racer is a mirror / environment map / reflection, and this is
one. It de-swizzled coherent and upright: a wrong interleave or a transposed write would have
scattered a correct viewport's geometry into noise. The probe is a tracer, not a model — but the
picture it draws is the cross-check with teeth for the write path.

### Where it stands / next

The swizzled-render-target halt is closed; the pitch path is byte-identical (the cold-boot title
frame is still md5 `5439dd95c92d462d03b9c5fbbd8a6c86`). The frontier is now two independent GPU
gaps: the zero viewport constants that degenerate the reflection's geometry (Part XI), and the
`0x2E` texture format the run halts on ~180 draws later.

### Tooling

- `tools/platform/xbox/{machine,nv2a,kernel,kernel_ordinals,kernel_objects,kernel_data,kernel_file,thread,sched,state,ports,run}.go` — the machine and its HLE.
- `tools/platform/xbox/{nv2a_pfifo,nv2a_pgraph,nv2a_ramht,nv2a_kelvin}.go` — the Phase-C GPU front
  end: the PFIFO DMA pusher, the PGRAPH method dispatch + survey, RAMHT handle→class resolution,
  and the Kelvin (3D) method dispatch.
- `tools/platform/xbox/{nv2a_vsh,nv2a_vertex,nv2a_raster,nv2a_combiner,nv2a_texture,nv2a_frame}.go`
  — the Kelvin pipeline (Part IV): the transform-program interpreter, vertex fetch/assembly,
  rasteriser (with the `Machine.OnPixel` provenance hook), register combiners, texture unit, and
  the clear/frame-export paths. `RR_NV_VS=1` dumps program uploads + disassembly, `RR_NV_DRAW=N`
  dumps the first N draws' decoded state and vertices.
- `tools/platform/xbox/apu.go` — the MCPX audio/USB latch apertures (APU, AC'97, USB OHCI);
  `RR_APU_TRACE=1` traces their traffic.
- `games/outrun-2006-xbox/extract/cmd/bootoracle` — the boot driver, standard oracle flags
  (`-image -steps -trace -stack -ordinals -dump -savestate -loadstate -v`), plus `-gpu` (run the
  NV2A pusher on each kick, do not stop at the first push) and `-survey` (record and print the
  PGRAPH method surface). `-stack` on a halt disassembles the call site so the next ordinal's
  signature reads straight off the pushes. `RR_NV_TRACE=1` traces every NV2A MMIO access.
  `-png` exports the display scanout, `-surfpng` the Kelvin render target (AA-resolved); a
  `-loadstate` of a halted frontier state clears the halt and retries the trapped call.
  `-keys NAME@FRAME[:HOLD]` drives pad 1 from the title's own flip — the names are
  `xbox.PadControlNames()`, shared with the debugger, and cover the analog buttons and the stick
  (`-keys "a@30:10,stickright@260:15"`).
- `tools/cpu/x86/sse.go` — the SSE/SSE2 + MMX execution subset (the Xbox-only CPU addition).
- `RR_NV_VP=1` traces the viewport methods (`0x0A20`/`0x0AF0`) and any constant load landing in
  slots 56–60; `RR_NV_SURF=1` traces the surface state methods (`0x0200`–`0x0214`) — both print
  the current draw count, which is how a value is placed *before* or *after* a pass.

## Part XI — the viewport was in the register file all along

Part X left the three 128×128 reflection targets rendering only their clear: every triangle's
screen position came out `[0, 0, 0, w]`, because the program's screen-space epilogue multiplies by
a viewport scale in `c58` and adds an offset in `c59`, and both were zero. The session that found
this searched the reached window for constant loads and for "viewport methods" and saw neither, so
the constants were presumed set before the savestate by some path unknown.

Half of that search was wrong, and the way it was wrong is the whole lesson: it watched method
`0x0A30`, which is nothing, and concluded "no viewport methods". The zero-hit had no control. The
savestate's own **latched register file** — every method the dispatch doesn't model still lands in
`Regs` — held the answer the search missed:

```
0x0A20..0A24 = 0.53125, 0.53125            a half-pixel bias: a viewport OFFSET
0x0AF0..0AF8 = 320, -240, 16777215         ±half-extent and a 24-bit z range: a viewport SCALE
```

The game had been writing its viewport all along, through two dedicated Kelvin methods —
`SET_VIEWPORT_OFFSET` (`0x0A20`) and `SET_VIEWPORT_SCALE` (`0x0AF0`) — which our dispatch latched
into `Regs` and dropped. Under `RR_NV_VP` the reached window contains **1,200** such writes.

### Three legs, each forced, none optional

On the NV2A the viewport does not sit beside the transform constants — it lives **in** them. The
aliasing (`0x0AF0` → `c58`, `0x0A20` → `c59`) is pinned from the image alone:

- **The epilogue's arithmetic names the roles.** `MUL oPos.xyz, R12.xyz, c58.xyz` then
  `MAD oPos.xyz, R12.xyz, R1.x, c59.xyz` (`R1.x = RCC(w)`) computes
  `screen = clip/w × c58 + c59` — the multiplied slot is a scale, the added one an offset.
- **The values name the methods.** Per pass, `0x0AF0` receives `(w/2, -h/2, 2²⁴-1, 0)` and
  `0x0A20` receives `(x+w/2+0.53125, y+h/2+0.53125, 0, 0)`: `(320,-240)/(320.53,240.53)` for the
  640×480 pass, `(64,-64)/(64.53,64.53)` for the 128×128 reflection targets, `(256,-256)` and
  `(160,-120)` for two others — dimensionally a scale and an offset, sized to each target.
- **Nothing else can reach the slots.** Across the whole window: zero `SET_TRANSFORM_CONSTANT`
  loads targeting slots 56–60 (this time *with* a control — the same instrumented run catches the
  1,200 viewport-method writes, and the constant file shows the load path demonstrably working
  elsewhere: `c0`/`c1`, the `c160`–`163` MVP). If the methods don't alias into the file, no
  mechanism exists by which the program's `c58`/`c59` could ever hold the viewport the game
  demonstrably configures.

The model is two cases in `kelvinMethod`: the two 4-float method windows write
`Const[58]`/`Const[59]`. A unit test pins the routing and that neighbouring methods don't leak
into the slots. The 2-D passthrough passes are untouched — their programs carry their own
viewport in `c0`/`c1` (explicit const loads, visible as the last-latched upload in the same
register file) and never reference `c58`/`c59`.

### The z mapping the probe guessed, read instead

Part X's probe invented `z' = z/w × 2²³ + 2²³`. The game's own stream says otherwise, uniformly,
in every pass it programs: **scale.z `= 16777215`, offset.z `= 0`** — `z' = clip.z/w × (2²⁴-1)`,
full-range D3D depth onto the 24-bit buffer, no bias. Derived, not chosen: it is the third
component of the same two method vectors.

### ★ The reflections render from the derived state

From `gameplay.state`, no pokes: the three swizzled targets de-swizzle to a coherent road /
horizon / grandstand reflection (`work/swizzled-rtt-derived-viewport.png`, md5
`6a63813a4ff197a8a0faa15ad90c3cb2` — the same scene the Part X probe drew, now with the true
depth mapping). The pitch path did not move: the cold-boot title frame is still md5
`5439dd95c92d462d03b9c5fbbd8a6c86`. And because the epilogue is the *same* for every 3-D program
this port compiles, the fix feeds the 640×480 race view identically — the reached window now
halts inside the main pass on the next real gap, with honest geometry behind it.

### The next frontier introduces itself: a depth buffer as a texture

The new halt is texture format `0x2E` on unit 3, and the window says precisely what it is. At
draw 487176 the game binds `0x01C18180` as **both** colour and zeta offset — a 512×512,
pitch-2048 depth-render pass — and six draws later unit 3 samples that address as a linear
texture. Its content reads back as dwords of `depth<<8 | stencil` (the cleared buffer is uniform
`FFFFFF00`), which is the raster's own zeta layout: `0x2E` is the **linear X8_Y24 depth+stencil
image**. The consumer is a shadow receiver — `oT3` comes from a texture matrix (`c176`–`179`),
the stage mode is projective, the combiner adds `TEX3` into a black overlay gated by an alpha
test `GEQUAL 1`.

What a *sample* of it returns — a hardware depth-compare against `r/q`? in which channels, with
which polarity? — is not derivable yet: the only reachable map is all-far (no caster pass draws
in the window; a bounded toleration probe shows the run advances just ~250k steps before the
next unit halts on a DXT3 **cube map**, `fmt=06610E2D`). So the format is *identified* and the
halt now names it — `samples a depth buffer (LU X8_Y24) — shadow-map sampling unmodelled` —
rather than inventing a compare whose wrongness would render plausibly. Part XII's queue:
the shadow-map sample semantics (needs a window with a caster), then cube maps.

## Part XII — the shadow map that was never cast, and the fixed-function frontier behind it

Part XI's queue put the shadow-map sample first. The register-file evidence was already strong:
at draw 487176 the game binds `0x01C18180` as **both** colour and zeta (a 512×512 depth-render
pass), clears it to far (`FFFFFF00`), and six draws later unit 3 samples that address as a linear
`X8_Y24` depth texture, projectively, with a texture matrix and an alpha-tested black overlay.
What a *sample returns* — a hardware depth-compare against `r/q`, in which channels, with which
polarity — Part XI could not pin, because the only reachable map was all-far, and it named that
gap rather than inventing a compare.

Part XII set out to *find a caster* — a pass that writes real occluder depth into the map — and
in doing so proved, rather than assumed, that **none is reachable**, and that the compare is a
hardware operation the register combiners cannot stand in for.

### The census: only the map is empty; everything else that writes depth is accounted for

The instrument is a zeta-write census (`RR_SHADOW`): every depth write the raster performs is
bucketed by the zeta surface *offset* it lands in, so a shadow caster — a pass writing non-far
depth into a non-framebuffer buffer — is visible **whether or not it binds `colour==zeta`**. Run
over the whole reachable gameplay window (draws 486972→487371, ~8 frames), three buffers receive
depth and no fourth does:

```
zeta=01AD4000  pix=10,134,384  draws=167  zmin=000000 zmax=FFFFFF   the main framebuffer's Z
zeta=01D78240  pix=89,253      draws=19   zmin=F350F4 zmax=FFFFFF   the reflection RTTs' shared Z
zeta=00000000 / 01DD7780       draws=1    all-zero                  two degenerate single draws
```

The sampled shadow map `0x01C18180` is **absent from that list**: it receives *zero* depth writes
in every frame. It is cleared to far and sampled empty — confirmed at full 512×512 resolution,
`nonfar=0` at every sample. The one real-depth off-screen buffer, `0x01D78240` (`F350F4..FFFFFF`),
looked at first like a caster, but its three passes bind colour to `0x01D28200 / 01D38200 /
01D48200` — spaced exactly `0x10000 = 128×128×4` apart, the Part X/XI **reflection RTTs**. So
`0x01D78240` is the reflections' shared depth buffer, a red herring; it is never sampled as a
texture. This gameplay position simply casts no shadow into the 512-map.

### Why the sample can't be a combiner trick

Before concluding it needs a populated map, the alternative had to be ruled out: could the
"compare" live in the shader instead of the texture unit? The receiver's full state
(`dumpReceiverState`) says no. Only **unit 3 is enabled** (`ctl0` bit 30 set on unit 3 alone);
the stage program routes it projectively (`shaderStages=0x00010000`, unit-3 mode 2); the alpha
test is `GEQUAL` with `ref=0x01`, and stages 1 and 3 read register 11 (TEX3) into the output
alpha that gates the overlay. The register combiners (nv2a_combiner.go) have **only** mul / add /
dot / mux-at-0.5 — no general compare of two interpolated values. A shadow-map depth compare of
`r/q` against the stored depth therefore *cannot* be done in the combiner; it must be the texture
unit's own operation. And the filter register `filt=02022000` is **identical** to unit 0's normal
texture — there is no per-unit shadow-compare field the game toggles that would tell us the
function from the register state alone. So the compare function, its polarity, and the channels
the result lands in are pinnable only against a map that actually contains an occluder — which no
reachable frame provides. The halt now names exactly that:

```
texture unit 3 samples a depth buffer (LU X8_Y24) — shadow-map sampling unmodelled
  (map is all-far in every reachable frame; needs a populated caster to derive the compare)
```

This is the honest terminus for the shadow question: not "unmodelled because hard," but "not
derivable here because the only evidence that would distinguish the candidates is absent, and
inventing it would render plausibly over the empty map and be wrong." The unblock is a savestate
at a moment with a real caster (a car-select/garage turntable, or a track position with a dynamic
occluder), where the first sample is already populated and the census dumps the receiver over it.

### The frontier behind the halt: fixed-function T&L

Tolerating the depth sample (and the DXT3 cube map behind it) advances the run just past the first
shadow frame to a new halt at draw 487296: **fixed-function transform mode** (`0x1E94=0x00000004`,
mode bits `= 0`, not the program mode `2` every prior draw used). This is the NV2A's built-in T&L
engine, not a compiled vertex program. It was scoped in full:

- **The matrices are dedicated Kelvin state, not the program-mode constant file.** Found by
  scanning the register file: `SET_MODEL_VIEW_MATRIX` at `0x0480` (an orthonormal rotation +
  translation), its inverse at `0x0580` (the transpose — used to transform normals), and
  `SET_COMPOSITE_MATRIX` at `0x0680`, the world→clip model-view-projection. The composite's last
  row is `−(model-view row 2)`, i.e. `w = −z_view` — a right-handed perspective.
- **The convention is D3D row-vector.** Applied as `clip = pos · M` (not `M · pos`), the one real
  FF draw's vertices produce valid NDC (`x,y ∈ [−1,1]`, `z ∈ [0,1]`); the transpose gives garbage.
  This is the derived, verifiable half of the FF transform.
- **But this frontier sits *behind* the shadow halt** (487296 > 487182), so no clean run reaches
  it — it is only visible through toleration. And it lacks a verification anchor: of the four FF
  draws in the window, three (draws 487364–366) are **degenerate** — 800 vertices all at the
  origin, lighting disabled, drawing nothing — and the one real draw (487296) is a 70-vertex lit,
  specular, textured strip that maps to a sub-pixel screen footprint off the left edge, with no
  reference frame to check it against. Its lighting is *enabled* (`0x0314=1`), so a faithful render
  needs the light/material state (unfound) and the lighting equation, none of which can be verified
  here. Tolerating the FF draws to advance further diverges the CPU one frame later (an out-of-range
  read at `0x41000000`, PC `0x000208A3`) — a separate gameplay-logic frontier, not a rendering one.

So fixed-function T&L is *scoped* — the transform is derived and NDC-verified, the matrix methods
located, the lighting-enable read — but not *modelled*, because the reachable draws give nothing to
verify a full implementation against, and shipping an unverified lit render is exactly the
plausible-but-wrong outcome this port refuses. The `nv2a_transform_execution_mode` halt now points
here.

### Where Part XII leaves it

Two frontiers, both named precisely rather than papered over:

1. **Shadow-map sampling** — provably not derivable in any reachable frame (the map is cast-empty
   everywhere; the compare is a hardware op the combiner can't emulate). Needs a populated-caster
   savestate. `RR_SHADOW` dumps the zeta-write census + full receiver state at the sample halt for
   whoever has one.
2. **Fixed-function T&L** — scoped (composite `0x0680` row-vector world→clip, model-view `0x0480`,
   inverse `0x0580`, lighting enabled); unmodelled for want of an on-screen, lit FF draw to verify
   against, and it sits behind the shadow halt regardless.

The pitch path did not move (cold-boot title frame still `5439dd95c92d462d03b9c5fbbd8a6c86`); the
Part XI reflection render still stands; `go test ./tools/platform/xbox/` passes.

## Part XIII — the race loads, and the thing blocking it was our own raster

Part XII left two rendering frontiers and one question: past them, does the race actually finish
loading? The answer turned out to be yes — and the blocker was never the rendering *or* the
loading. It was a depth-write semantics bug in our raster that let a loading-screen UI quad
shred the title's own code, one phantom z=0 at a time. Every frontier on the way there fell to
the register evidence already in hand.

### The receiver's combiner, decoded — and the prior session's read corrected

The RR_SHADOW receiver dump, extended with the blend state and the per-stage combiner factors,
pins what the shadow overlay actually computes. The four stages are (factors: stage-1/2
`factor0=80A0A0A0`, stage-3 `factor1=3F000000`):

```
stage0:  spare0     = diffuse                       (color and alpha)
stage1:  spare1     = c0 + TEX3                     (c0.rgb = 0.627, c0.a = 0.502)
stage2:  spare0.rgb = spare0.rgb * spare1.rgb       (alpha: identity)
stage3:  spare0.a   = mux(spare0.a) ? 0 : c1.a      (c1.a = 0.247);  color: identity
final:   out.rgb    = fog.a*spare0.rgb + (1-fog.a)*fog.rgb + specular;  out.a = spare0.a
```

Two corrections to Part XI/XII's reading fall straight out:

- **TEX3's alpha never reaches the output.** Stage 1 puts it in `spare1.a`, which nothing reads
  again. The alpha-tested gate (`GEQUAL ref=1/255`) is fed by *diffuse alpha* muxed against a
  constant — the overlay's footprint is geometric, chosen per-vertex, independent of the map.
- **The sample enters through the color**: `out.rgb = diffuse.rgb × clamp01(0.627 + TEX3.rgb)`,
  blend **disabled**. TEX3 ≥ 0.373 clamps the factor to 1 — the unshadowed paint; TEX3 = 0
  darkens to 0.627 — the shadow.

That second line is what makes the all-far sample derivable *without* the compare's polarity:
over an all-far map every fragment is unoccluded, and "unoccluded" must be the return that
leaves the paint at its unshadowed baseline — a candidate where far texels return 0 would darken
the world by 0.627 unconditionally, which the reference frames refute. Hardware depth-compares
return 0/1 replicated, so **an all-far map samples as 1.0 (white) in every channel, under every
candidate compare function**. That case is now modelled: `texDecode` scans the bound map at
decode time, returns the all-white image if every texel is far, and **halts if any texel is
not** — the compare itself (function, polarity) still needs a populated caster, and now the
model itself verifies that precondition on every decode. (Also fixed while in there: the
combiner mux read `spare0.a >= 0.5 → AB`; the silicon's public spec says `< 0.5 → AB`. No test
pinned it and no reachable frame moved — the title hash is unchanged — but the receiver's gate
runs through that mux, so it had to be right before the gate could be reasoned about.)

### Cube maps: the game renders INTO its environment cube

The next halt behind the sample was the DXT3 cube map Part XII scoped (`fmt=06610E2D`, 64×64,
single-mip, unit 1, stage mode 3 = CUBEMAP). Modelled: six faces stored consecutively
(face stride = the block bytes for DXT), published `+X,-X,+Y,-Y,+Z,-Z` order, major-axis face
selection from the interpolated 3-component direction, per-face clamped filtering. A cube in
any other color format, non-square, or with a mip chain still halts by name — face strides for
those are unverified guesses.

Then the run produced a cube nobody had seen: **swizzled A8R8G8B8, 128×128** (`fmt=0771062D`).
That size is the Part X/XI reflection RTTs — and the game packs those exactly `0x10000 =
128×128×4` bytes apart, which *is* a cube's face stride with zero padding. The title isn't just
rendering reflections to textures; it's rendering them into the faces of an environment cube
map and sampling it on the car's bodywork. The face stride for the swizzled cube case is
therefore the game's own layout, not a guess. The glossy sweep over the Ferrari's paint in the
race frames is this path working end-to-end.

### Fixed-function T&L: model the verified half, arm a halt for the rest

The FF transform Part XII derived and NDC-verified is now real: `clip = pos · M` with
`SET_COMPOSITE_MATRIX` (0x0680) row-vector, then the viewport scale/offset from the aliased
`c58`/`c59` slots — the exact mapping every program-mode epilogue computes, reading the same
registers the hardware's viewport methods write. What is *not* modelled is FF lighting and
texgen, and the discipline is a **fragment-level halt**: a FF draw that needs either arms
`ffFragHalt`, and the raster halts the moment such a fragment would actually land. The
reachable FF draws (one sub-pixel off-left strip, three 800-vertex degenerates at the origin)
pass through without painting — honestly — while any future on-screen lit FF draw still names
the gap instead of shading something plausible.

### ★ The real blocker: a depth write that GL says never happens

With all three frontiers honestly modelled, the run reached the same CPU fault the toleration
probe had seen — `out-of-range read at 0x41000000, PC 0x208A3` — proving it was never a
toleration artifact. The disassembly at the fault decodes *zeroed* code: every corrupted dword
keeps its **low byte** and loses the other three. That signature is our own depth write —
`stored = z<<8 | stencil` — with z = 0. The `-watch` saw no writer because the raster stores to
RAM directly; an `RR_LOWWRITE` instrument caught it in one run:

```
LOWWRITE draw=487367 colorPhys=01DD7780 zetaPhys=00000000 ztest=false zwrite=true prim=QUADS
```

A loading-UI quad draws with depth **test disabled**, depth **mask enabled**, and
`SET_SURFACE_ZETA_OFFSET = 0`. Our raster honoured the mask unconditionally and wrote z=0
dwords from physical 0 upward — straight across the title's code at `0x20880` (also explaining
Part XII's mystery census buckets: `zeta=00000000` and `01DD7780`, one draw each, all-zero).
On this silicon the depth write happens *as part of* the depth test — test disabled means
nothing writes, whatever the mask says (GL semantics, and D3D's `ZENABLE=FALSE` compiles to
exactly this). The game itself is the proof: on real hardware this draw must write nothing, or
OutRun would corrupt its own loader on every loading screen. One line —
`st.depthWrite = st.depthTest && mask` — and the corruption, the fault, and the "loading
blocker" all vanish.

### ★ The race runs, and the oracle drives it

From `gameplay.state`, no pokes, no tolerations: the loading screen completes (clean frame —
logo, spinner, PLEASE WAIT, none of the banding/black-surface artifacts the toleration probes
had painted; those were the fakes' own fingerprints), the track streams in, and the game runs
**billions of instructions** into the race before meeting anything unmodelled at all. The next
and only frontier it met was `NtSetInformationFile` class 4 — FileBasicInformation, the save
path stamping times on its U:\ files (the class-4 sites the Part XII census had already read:
0x4339D live). Modelled: the 40-byte blob is stored verbatim per store key (savestate field
`FileBasic`, nil from older snapshots) so any future readback derives from what the guest
wrote; query class 4 still halts by name, and nothing invented flows back.

Past that, the run completed an 8-billion-step budget with **no halt**: grid frame with the
driver at the wheel, burnout smoke on the start line — and with `-keys "a@5:2000"` holding
accelerate, the oracle **plays the race**: full HUD (Time 78, Position 6/6, speedo,
start/goal bar), green start light, the starter waving the field off, the Ferrari pulling away
(`work/race-start-drive.png`, md5 `e9af08f57c9f9d7517d46c8127a3c90b`; the burnout frame is
`work/race-burnout.png`, `6f9c8710d00db142dc3eb9464732c38e`; `work/states/race.state` resumes
in-race). `-surfpng` captures are mid-frame at the stop point, so partially-drawn regions
(black wedges) are capture timing, not defects.

### Where Part XIII leaves it

- The **shadow compare** remains the one named rendering gap, now self-verifying: the first
  frame that casts real depth into a sampled map halts at the sample with the census in hand.
  Notably the whole 8B-step race window never populated the 512-map — this course seems not to
  cast into it at all.
- **FF lighting/texgen** halts on the first visible fragment that needs them; none has appeared.
- The cold-boot title frame is byte-identical (`5439dd95c92d462d03b9c5fbbd8a6c86`, procedure:
  `title.state` +100M steps `-surfpng`), and `go test` passes across
  `tools/platform/xbox`, `tools/cpu/x86`, and `tools/debug/...`.

## Part XIV — the race freeze: a game that measured our clock, and a watchdog our silence tripped

Part XIII ended with the oracle on the grid: Time 78, speedo 000, green light — and there it
stayed. From `race.state`, +6 billion steps and +36 billion steps produce the byte-identical
`-surfpng`; zero `FLIP_STALL`s in 300M steps (instrument controlled — the same grep counts flips
in the loading window); and yet the CPU retires 30B+ instructions without a halt. Nothing was
waiting on the kernel. That last fact took a tool to establish, and it inverted the whole search.

### The -threads dump: nobody is blocked

The HLE owns every thread, so the first deliverable was `bootoracle -threads`: per thread its
scheduler state, wait objects, and — via new signal provenance on every kobject (`noteSignal` at
each of the five signal sites) — who last signalled what it waits on. At the frozen state:

```
  tid=0 dead
* tid=1 running   PC=0002165x            <- the game's main thread, spinning
  tid=2 waiting   on 03FC05B0: semaphore signaled=false lastSignal=never
  tid=3 ready     susp=1 PC=00044F40     <- suspended ITSELF (NtSuspendThread wrapper)
  tid=4 dead      tid=5 dead
```

No thread waits on a vblank, an event, or an I/O completion that never comes. The prime suspect
list died in one line each: the vblank is *alive* (an `RR_PUMP` trace shows `KeSetEvent(001BB75C)
from 001B301E` — the D3D vblank DPC — at a clean 60 Hz throughout the freeze), tid=2 is a worker
idling on its work-queue semaphore, tid=3 is a parked worker awaiting a resume. The render thread
IS tid=1, it is RUNNING, and it never calls Present. The freeze is the game's own decision, taken
in user code, and the only instrument that could name it was a profiler — a sampling `RR_HOTPC`
(every 256th tick) added to the machine.

### The catch-up loop: the game measures the machine, honestly

The profile puts ~46% of all time in one small walker (0x21610) called from the frame function's
inner loop at 0x20BA4..0x20BDA — and a breakpoint on the loop's back-edge vs its exit shows the
loop re-entering forever and the exit never taken. Its shape, read off the disassembly:

```
0x20B14   CALL 0x20880 (60)            ; elapsed 60ths-of-a-second since race init
0x20B1C   MOV [0x41713C], EAX          ; -> the target
0x20B9A   while ([0x417130]++ < [0x41713C])
0x20BA4       run one simulation tick   ; 11 subsystem task lists, the walker per list
```

A fixed-timestep catch-up loop. And 0x20880 is a real clock: `RDTSC` deltas (via XAPI's
QueryPerformanceCounter — the wrapper at 0x44A1D is literally `RDTSC`), accumulated 64-bit, and
divided by a hardcoded frequency of 0x2BB5C755 = **733,466,453 Hz — the Xbox CPU clock**. The
menus never hang here because the menu game-modes take a *clamped* path (0x20AFA/0x20B03: modes
0x18/0x20/0x24 force a small target); the race trusts its clock. On real silicon that trust is
sound. On ours it wasn't, because `instrsPerMs` — the machine's whole declaration of its own
speed, from Phase B — said **2000 instructions per millisecond: a 2 MHz Xbox** (with
`TSCMul=367` faithfully bridging RDTSC to 733 MHz *of that slow time*). One simulation tick
costs ~130k instructions = 65 guest-milliseconds; the loop retires one tick per iteration while
its target grows by four. The game measured our machine honestly and concluded, correctly by its
own clock, that it was hopelessly behind — at the frozen state the target stood at 318,115
sixtieths (~88 guest-minutes) against 187,921 simulated. A machine 366× too slow *by its own
declared clock* cannot run a fixed-timestep race, and no amount of kernel modelling changes that.

### Fixing the declaration: one timebase, affine clocks, continuous savestates

The fix is the honest one: declare the hardware's real speed. `instrsPerMs = 733466` (733.466
MHz at one instruction per cycle — the same constant the title's XAPI divides by), and every
guest-visible clock now derives from a single timebase:

- `systemTime100ns` / KeTickCount / USB frames: affine in the tick with a persisted **clock
  epoch** (`clockBaseTick`/`clockBase100ns`);
- the TSC: no longer `Steps*TSCMul` in the CPU core but a host hook (`x86.CPU.TSCFunc` →
  `Machine.guestTSC`), because guest time can pass without retiring instructions (idle-advance,
  KeStall) and the TSC must tell the same time as everything else;
- `KeQueryPerformanceCounter/Frequency` return that same TSC and its true rate
  (`instrsPerMs*1000` = 733,466,000 Hz, 0.7 ppm from the XAPI constant).

Savestates carry the epoch (`ClockScheme=1`). A **legacy snapshot migrates on load**: the epoch
is pinned so guest time and the TSC *continue* from the exact values the old formulas produced at
the save — the guest's own memory holds baselines against both clocks (the race accumulator at
0x4170E8 is a raw RDTSC), and a clock that stepped across a restore would break every delta
computed against them.

### The second freeze: DirectSound's watchdog, tripped by our silence

With the clock honest, the loop drained its backlog — and re-diverged. The target had grown
another 24 guest-minutes across an 8B-instruction run. `RR_PUMP` found the inflation in one
line: **167,569 × `KeStallExecutionProcessor(10,000 µs)` per 100M instructions, all from
0x1D7F43** — inside DirectSound. The chain (read off the stack at a breakpoint on the stall):
the game's frame loop calls DirectSoundDoWork (0x37DB0), which runs a **watchdog** at 0x1D814B:

```
MOV EAX, [0x1E0D88]          ; -> 0xFE85A018, a word in the EP DSP's aperture window
CMP DWORD [EAX], 0x00CCCCCC  ; the encode-processor's alive marker
JZ  healthy
CALL 0x1D7E7C                ; else: full EP bring-up — reset, run, and a 10 ms settle stall
```

We execute no DSP56k microcode, so nothing ever wrote that word (an `RR_APU_TRACE` shows only
reads, always 0, always from the check) — the watchdog failed on *every* DoWork, and its
recovery charged a genuine 10 ms of guest time per frame: more than half a 60 Hz frame budget,
gone to an audio reset, every frame. That diverges the catch-up at any CPU speed. The marker
value is derived, not invented: the only 0x00CCCCCC in the image is the compare itself (a byte
scan finds no CPU-side writer), so it is DSP-published — and on hardware, where races run
without a 10 ms audio hiccup per frame, a running EP evidently keeps it in place. That is the
model (`apu.go`): while the title's own bring-up has left the EP run-control (+0x5FFFC) at 3
(run) — it writes 1, then 3, exactly as the recovery path does — the alive word (+0x5A018)
reads back 0x00CCCCCC; in any other run-state the latch answers and the watchdog's recovery
stays visible. The 10 ms stall itself was never the bug: `KeStallExecutionProcessor` advancing
the clock by the stall's worth is exactly right — the bug was the condition that made the title
stall every frame, forever.

### ★ Verified by outcome — and the race delivers Part XII's missing savestate

With both fixes and **no pokes**, judged frame by frame (a new `-stopflip N` stops the run *at*
the Nth FLIP_STALL, while the completed frame is still the bound colour surface — so a capture
is a whole presented frame, not a mid-frame slice):

- From `gameplay.state` (pre-race, so the race clock initialises fresh under the honest
  declaration): the track loads and **the race start actually plays** — 7,692 flips in 1.76B
  steps where the frozen machine produced zero, and the flip-aligned captures walk through the
  start-line cinematic: the grid close-up with the crowd and trackside banners, the burnout
  with smoke pouring off the tyres, the countdown **"3"** stamped over the car, its flash-halo
  transition toward "2" (`work/race-countdown-3.png` and neighbours; every capture a distinct
  hash). `work/states/race-countdown.state` resumes there under throttle.
- From `race.state` (a legacy state carrying ~46k sixtieths of backlog baked in by the old-clock
  era): the loop *drains* — the ~13 simulated minutes run the race clock out, the game moves
  through its post-race flow on its own, and lands on a fresh "Loading / PLEASE WAIT" — 14,223
  flips in one 10B-step run, no halt, where Part XIII's machine presented nothing at all.
- **The armed shadow halt fired — before GO.** 1.76B steps after `gameplay.state`,
  deterministically (re-halts within 137 instructions of a resume), the raster halts at the
  Part XII/XIII sample: *"texture unit 3 samples a POPULATED depth buffer (LU X8_Y24,
  fmt=00012E29, non-far texel at 206,250 = 0C195F)"*. The start-line scene itself is the first
  real shadow caster this pipeline has ever seen — the exact unblock Part XII said the shadow
  compare needed and no reachable frame could provide. The machine at that halt is
  `work/states/shadow-caster.state`. Deriving the compare from it is the next rendering
  frontier, and it now *gates the drive*: Time-counting and speedo evidence live on the far
  side of that sample, which is precisely what an armed fragment-level halt is for.

### The title pin moves — measured, explained, re-pinned

The Part V regression procedure (`title.state` +100M steps `-surfpng`) no longer produces
`5439dd95…`, and it cannot: 100M instructions used to be 50 guest-seconds and are now 0.136 —
the procedure names a *step count*, and the fix changed what a step is worth. The new frame is
the same settled title card at an earlier moment, and that is not asserted but counted: against
the pinned frame, exactly **1,731 pixels differ, every one inside x∈[268,373], y∈[373,391] —
the "PRESS START" text box** (blink-off vs blink-on); the background is byte-identical. The
render path did not move; the clock under the procedure did. Re-pinned:
`title.state` +100M steps `-surfpng` = `0bea502acd2a1f902d429097022116b5` (deterministic across
repeated runs, and identical before and after the EP-watchdog fix).
