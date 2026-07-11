# LocoRoco (PSP) — technical reference

**Image.** `LocoRoco.cso` — 528,623,264 bytes, MD5 `c69f23b9926fd2dbe613092a79002289`.
A CISO-compressed dump of a UMD (`DISC_ID` UCUS98662). The uncompressed ISO is
589,299,712 bytes. The image is not committed (see the repository copyright policy);
supply a dump with the MD5 above.

## Contents

- **Part I — The image.** The CISO container, the ISO 9660 UMD filesystem, PARAM.SFO,
  the `PSP_GAME/SYSDIR` boot files, the `~PSP` KIRK-encrypted executable and its
  decryption, the ELF/PRX module, and the Allegrex memory map.
- **Part II — Boot chain.** Module relocation and load, register/stack seeding, the
  kernel-HLE syscall-stub mechanism, and the surface reached.
- **Part III — Graphics engine (GE).** The display-list interpreter, the software
  rasterizer, and the framebuffer output.
- **Part IV — Audio.** *(future)*
- **Part V — Game data.** *(future)*

The toolchain — the Allegrex CPU core and its disassembler/tracer, the PSP machine
oracle, and the `pspinfo`/`bootoracle` front-ends — is documented in Parts I and II.

---

## Part I — The image

### 1. Container: CISO ("CSO")

A UMD dump is commonly stored as CISO, a run of fixed-size logical blocks each held
raw or raw-DEFLATE compressed, indexed for random access. The 24-byte header
(little-endian):

| offset | size | field |
|--------|------|-------|
| 0x00 | 4 | magic `CISO` |
| 0x04 | 4 | header size (reserved) |
| 0x08 | 8 | uncompressed total size — `0x23270000` = 589,299,712 |
| 0x10 | 4 | block size — `0x800` = 2048 |
| 0x14 | 1 | version — 1 |
| 0x15 | 1 | align — 0 |
| 0x18 | . | index: `(nblocks+1)` uint32 |

Each index entry's bit 31 flags a block stored uncompressed; bits 0..30 are the
block's file offset right-shifted by `align`. A block's stored length is the gap to
the next entry's offset; compressed blocks are raw DEFLATE (no zlib wrapper). The
2048-byte block size equals the ISO logical block, so the CISO block index *is* the
ISO LBA. The decoder (`tools/platform/psp/cso.go`) reads through an `io.ReaderAt`
with a small LRU of decoded blocks, so the ~562 MiB image is never fully resident.

### 2. Filesystem: ISO 9660 UMD

The volume is a plain "cooked" 2048-byte-per-sector ISO 9660: a Primary Volume
Descriptor at logical block 16 (`CD001`), both-endian numeric fields, directory
records packed within a logical block. The reader (`iso.go`) sits on a block source
satisfied by either the CISO decoder or a flat `.iso`. The boot tree:

```
PSP_GAME/
  PARAM.SFO           title metadata
  SYSDIR/
    EBOOT.BIN         the encrypted boot executable (~PSP)
    BOOT.BIN          blanked on this UMD
  USRDIR/
    data/DATA.BIN     the packed game data (445 MiB)
    modules/*.prx     media codecs (audiocodec, atrac3plus, mpeg, libfont, …)
```

### 3. Metadata: PARAM.SFO

`PARAM.SFO` (`sfo.go`) is a key/value table (`\0PSF` magic, a 16-byte index record
per entry, a key string table and a data table). For this disc:

| key | value |
|-----|-------|
| `TITLE` | LocoRoco |
| `DISC_ID` | UCUS98662 |
| `CATEGORY` | UG |
| `PSP_SYSTEM_VER` | 2.71 |
| `BOOTABLE` | 1 |

### 4. The boot executable: `~PSP` / KIRK

`PSP_GAME/SYSDIR/EBOOT.BIN` is a `~PSP` container: a 0x150-byte header over an
AES-CBC-encrypted body. `BOOT.BIN`, the historically-plaintext companion, is blanked
on this UMD, so decryption is required to reach code. Header fields used:

| offset | field | value |
|--------|-------|-------|
| 0x28 | decrypted ELF size | `0x2388F4` = 2,328,820 |
| 0x2C | `~PSP` file size | `0x238A50` = 2,329,168 |
| 0xB0 | body size | `0x2388F4` |
| 0xD0 | tag | `0xC0CB167C` |

Decryption is performed by the PSP's KIRK security coprocessor. The tag at +0xD0
selects a per-firmware XOR seed; the header is descrambled with it and rebuilt into a
KIRK command-1 header, which KIRK decrypts. The steps (`kirk.go`, `prx.go`), for the
2.xx-game tag `0xC0CB167C` (seed `g_keyEBOOT2xx`, kirk7 key id `0x5D`):

1. Reassemble a 0x150-byte structure from the header regions (the embedded SHA-1 at
   +0xD4, the "unused" block at +0xE8, the 0x90-byte key block split across +0x110
   and +0x80, and the 0x80-byte PRX header at +0x00).
2. Type-1 pre-decrypt: `kirk7` (AES-128-CBC decrypt under `keyvault[0x5D]`, zero IV) a
   0xA0-byte slice of the structure; verify the embedded SHA-1 over the seed's first
   0x14 bytes and the reassembled blocks.
3. Build a KIRK command-1 header: XOR the first 0x70 header bytes with `seed[+0x14]`,
   `kirk7`-descramble, XOR with `seed[+0x20]`.
4. KIRK command 1: AES-CBC-decrypt the header's first 0x20 bytes under the KIRK master
   key to recover the body's AES key, then AES-CBC-decrypt the body → the plaintext
   ELF (verified: `\x7fELF`, exactly `0x2388F4` bytes).

The KIRK AES key set (the master key and the `keyvault`/tag seeds) are hardware
constants of the PSP's KIRK engine — not present on the UMD and not derivable from
ciphertext. They are treated as documented platform constants (`kirk_keys.go`), the
same standing as a boot-ROM seed; the KIRK *algorithm* is reimplemented from the
format and verified against ground truth.

### 5. The module: ELF32 / PRX

The plaintext is an ELF32 little-endian MIPS relocatable module (`e_machine` 8,
`e_type` `0xFFA0` = PRX). One `PT_LOAD` segment (file `0x1B9570`, mem `0xBF7580`),
entry `0x3C500`, `gp` `0x1C1560`. The PSP-specific `.rodata.sceModuleInfo` (located
by the first program header's `p_paddr`) names the module (`LocoRoco`) and lists the
import stub tables: **29 libraries** — `ThreadManForUser`, `SysMemUserForUser`,
`IoFileMgrForUser`, `sceDisplay`, `sceGe_user`, `sceCtrl`, `sceAudio`, `sceSasCore`,
`sceAtrac3plus`, `sceMpeg`, `sceUmdUser`, `sceUtility`, `sceLibFont`, `scePower`,
and others — each a list of function NIDs and a call-stub address. (`elf.go`.)

### 6. The Allegrex and its memory map

The PSP CPU is the Allegrex: a little-endian MIPS32R2 core with a single-precision
COP1 FPU and a 128-bit COP2 vector unit (the VFPU). The core (`tools/cpu/allegrex`)
shares the R3000 skeleton of `tools/cpu/mips` — branch and load delay slots, the same
`Bus` interface — and adds the MIPS32R2 integer group (`movz`/`movn`/`rotr`, SPECIAL2
`mul`/`madd`/`clz`, SPECIAL3 `ext`/`ins`/`seb`/`seh`/`wsbh`), the FPU, and the VFPU
load/stores. Its disassembler (`disallegrex`) and tracer (`codetraceallegrex`) follow
the shared CPU-command layout; the machine oracle drives it through the memory map:

| range | region |
|-------|--------|
| `0x00010000`–`0x00013FFF` | scratchpad (16 KiB) |
| `0x04000000`–`0x041FFFFF` | VRAM (2 MiB) |
| `0x08000000`–`0x09FFFFFF` | main RAM (32 MiB; user partition at `0x08800000`) |
| `0x1C000000`–`0x1FFFFFFF` | hardware I/O |

Addresses fold through the MIPS kseg mirrors (`addr & 0x1FFFFFFF`).

`pspinfo` is the Part I inspector: `-ls` walks the disc, `-sfo` dumps the metadata,
`-exe` KIRK-decrypts and describes the module, `-extract` pulls files.

---

## Part II — Boot chain

The oracle (`tools/platform/psp`, driven by `extract/cmd/bootoracle`) boots the
decrypted module directly.

### 1. Load and relocation

The PRX is relocatable (segment virtual address 0). It is loaded into the user
partition at base `0x08804000` and relocated: the `SHT_PRXRELOC` sections
(`0x700000A0`) carry MIPS relocations (`R_MIPS_32`, `R_MIPS_26` for `jal` targets, the
`HI16`/`LO16` pairs for split immediates), all resolved against the single segment's
base. The entry becomes `0x08840500` and `gp` `0x089C5560`.

### 2. Register and stack seeding

`$sp`/`$fp` are seeded near the top of user RAM (`0x09FF0000`), `$gp` to the
relocated module gp, `$ra` to 0 (a return to `$ra=0` ends the run), and `$a0`/`$a1`
(argc/argv) to 0. A bump heap is placed above the module image for
`sceKernelAllocPartitionMemory`.

### 3. Kernel HLE: syscall-stub patching

The PSP kernel is high-level-emulated. A game calls an imported `sceXxx` function
through a call stub that ends in a `syscall`; at load time each stub is patched to

```
jr $ra
syscall <synthetic code>
```

and the code is mapped to a Go handler. Functions are identified by NID — the first
four bytes (little-endian) of SHA-1(function name) — so hashing a curated name list
gives the NID→name map used to label the trace and bind the modelled handlers. The
CPU's syscall hook dispatches by code. Handlers follow three tiers: memory/thread/
display/time calls the C runtime needs are modelled; kernel objects (semaphores,
events, mutexes) are stubbed to hand out handles and report success; everything else
logs its `(library, NID)` and returns 0, so one run enumerates the whole syscall
surface the boot path reaches.

### 4. Scheduling and the surface reached

The MIPS32R2 core has no load delay slot (loads are interlocked), which the C runtime
relies on: an init-array walk executes `lw $v0, 0($s0); jalr $v0`, expecting the loaded
pointer immediately. A cooperative scheduler (`sched.go`) carries the boot through its
thread hand-offs: `sceKernelStartThread` makes a thread runnable and lets the caller
continue; when a thread sleeps or returns (to a sentinel `$ra`) the scheduler saves its
register context and switches to the highest-priority ready thread.

With those, the boot runs the module entry and C runtime, creates and starts the main
thread, brings up synchronization (`sceKernelCreateSema`/`CreateEventFlag`/`WaitSema`),
and initializes the graphics engine — **submitting GE display lists** via
`sceGeListEnQueue`. The kernel/GE functions reached in one run include:

```
sceKernelCpuSuspendIntr / sceKernelCpuResumeIntr, sceKernelCreateThread / StartThread,
sceKernelCreateSema / CreateEventFlag / WaitSema, sceKernelChangeThreadPriority,
sceRtcGetCurrentTick, sceDisplaySetMode,
sceGeEdramGetAddr, sceGeListEnQueue, sceGeListSync, sceGeDrawSync, sceGeSetCallback
```

Execution then reaches the current wall — an unimplemented VFPU matrix op
(`vmmul`) used to build the render transforms. The VFPU results feed control flow, so a
correct VFPU is required to advance. Savestates (`state.go`) snapshot the full machine
(RAM, VRAM, scratchpad, the Allegrex register files including FPU/VFPU, the thread
contexts, kernel-object and syscall tables) with the image MD5 pinned.

---

## Part III — Graphics engine (GE)

The GE is the PSP's GPU: a game builds a display list of 32-bit commands (top 8 bits =
command, low 24 = argument) in memory and submits it with `sceGeListEnQueue`; the GE
walks the list maintaining render state and draws primitives into the framebuffer in
VRAM.

- **Capture (`ge.go`).** A submitted list is captured by following its control flow —
  `JUMP`/`CALL`/`RET`/`BASE` — to the `END`, flattening it to a command sequence.
- **Interpret (`ge_raster.go`).** The commands set the framebuffer target
  (`FRAME_BUF_PTR`/`FRAME_BUF_WIDTH`/`FRAMEBUF_PIX_FORMAT`), the viewport
  (scale/center), the world/view/projection matrices, the vertex format
  (`VERTEX_TYPE`) and the texture binding; `PRIM` triggers a draw. Vertices are decoded
  per the vertex type (position s8/s16/float, colour 565/5551/4444/8888, texcoords),
  and — for non-through primitives — transformed by model-view-projection and the
  viewport, or taken as screen-space coordinates in through mode.
- **Rasterize (`ge_draw.go`).** Triangles (lists/strips/fans) fill by barycentric
  interpolation of per-vertex colour; sprites fill axis-aligned rectangles; a bound
  texture modulates the colour. Pixels are written in the framebuffer's PSP format.
- **Output (`framebuffer.go`).** The 480×272 framebuffer is decoded from VRAM to an
  RGBA image and written to PNG (`bootoracle -shot`).

The GE pipeline is validated end to end by a rasterizer test that renders a list into
VRAM and reads the result back. The two lists Loco Roco submits before the VFPU wall
are graphics-engine *state initialization* (every register reset), not draws, so the
game's own first frame awaits the VFPU and the asset-loading path (`sceIo`, backed by
the mounted UMD volume).

---

## Part IV — Audio

*(future)* `sceSasCore`, `sceAtrac3plus` and the `USRDIR/data/stream/*.sgb` /
`*.at3` assets.

## Part V — Game data

*(future)* `USRDIR/data/DATA.BIN` (445 MiB) and the `first_us.arc` archive.
