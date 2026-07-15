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
initialising resources, and halts at a **new device region: an out-of-range read at `0xFE801100`,
the MCPX APU (audio) MMIO at `0xFE800000`**, entirely separate from the NV2A. Past the APU
bring-up lies the render loop that submits geometry, and then the Kelvin pipeline itself
(`nv2a_kelvin.go` is a latch-only stub today): vertex-program interpreter, register combiners,
swizzled texture sampling, banded rasteriser → PNG.

### Tooling

- `tools/platform/xbox/{machine,nv2a,kernel,kernel_ordinals,kernel_objects,kernel_data,kernel_file,thread,sched,state,ports,run}.go` — the machine and its HLE.
- `tools/platform/xbox/{nv2a_pfifo,nv2a_pgraph,nv2a_ramht,nv2a_kelvin}.go` — the Phase-C GPU: the
  PFIFO DMA pusher, the PGRAPH method dispatch + survey, RAMHT handle→class resolution, and the
  Kelvin (3D) object (a latch-only stub pending the pipeline).
- `games/outrun-2006-xbox/extract/cmd/bootoracle` — the boot driver, standard oracle flags
  (`-image -steps -trace -stack -ordinals -dump -savestate -loadstate -v`), plus `-gpu` (run the
  NV2A pusher on each kick, do not stop at the first push) and `-survey` (record and print the
  PGRAPH method surface). `-stack` on a halt disassembles the call site so the next ordinal's
  signature reads straight off the pushes. `RR_NV_TRACE=1` traces every NV2A MMIO access.
- `tools/cpu/x86/sse.go` — the SSE/SSE2 + MMX execution subset (the Xbox-only CPU addition).
