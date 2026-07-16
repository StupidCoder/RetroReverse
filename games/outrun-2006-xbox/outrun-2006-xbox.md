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

The debugger's first honest frame capture states a gap that every pinned PNG in Parts IV and V had
walked past: **`-surfpng`, the draw target, is the only picture that has ever been verified.**
`Display()` — the scanout, what the CRTC actually reads — is a **320×240 window on a stale buffer
that renders blank white**, because the machine still has the loading phase's display mode
registered and the 640×480 mode switch has never happened. The title is drawing 640×480 frames into
buffers the TV is not reading.

That is not a bug the adapter introduced; it is the known-pending mode switch, seen for the first
time. The adapter does not paper over it — a `Display()` quietly wired to the draw target would
have looked perfect and hidden it. The two surfaces are both offered, they disagree, and the
disagreement is the finding.

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
- `tools/cpu/x86/sse.go` — the SSE/SSE2 + MMX execution subset (the Xbox-only CPU addition).
