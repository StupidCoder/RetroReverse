# Super Mario 3D Land (Nintendo 3DS) — technical reference

**Image:** `Super Mario 3D Land (Europe) (En,Fr,De,Es,It,Nl,Pt,Ru) (Rev 2).cci` — 536,870,912 bytes,
MD5 `d92c6456e8cbd8dcb37d6382807df80c`. Not committed (copyright); supply your own copy. The dump is
**decrypted** — the NCCH NoCrypto flag is set — so its ExHeader, ExeFS and RomFS are plaintext; an
encrypted retail dump cannot be read here, because the AES-CTR keys are console state, not cartridge
data, and none are embedded.

Super Mario 3D Land (Nintendo EAD, 2011) is the 3DS's first 3-D Mario and one of its early
system-sellers. This document reconstructs the shipped cartridge from its bytes alone, with
purpose-built tools: an ARM11 (ARMv6K) disassembler/CPU with VFPv2 (`tools/cpu/arm`, in its `V6K`
variant), the 3DS container library (`tools/platform/n3ds` — NCSD/NCCH/ExeFS/RomFS/BLZ), and a
machine oracle (`tools/platform/n3ds`) that loads the retail `.code` and runs its ARM11 program
under a high-level-emulated Horizon kernel. This is the repository's first Nintendo 3DS title and
its first ARMv6/VFP target.

This file currently documents **Part I (the cartridge image and its containers)**, **Part II (the
boot/execution bring-up)**, **Part III (the HOME-Menu banner — its format)** and **Part IV (extracting the banner to a GLB)**;
the in-game asset formats follow in later work.

---

## Part I — The image

The medium is a **CCI** ("CTR Cart Image", a.k.a. `.3ds`): a flat dump of the cartridge framed by an
**NCSD** header. `n3dsdump` parses the whole nest in one pass and every structural claim below is
arithmetic the headers pin themselves — a wrong media-unit size or a wrong container offset does not
"look slightly off," it fails a magic check or an extent bound.

### NCSD — the cartridge

At file offset `0x100`, after a 0x100-byte RSA-2048 signature, sits the `NCSD` header. Its flag
byte 6 gives the **media unit** (`512 << flags[6]` = 512 bytes here); *every* offset and length in
both the NCSD and the NCCH headers is counted in these units. The partition table holds three used
slots:

| # | offset | size | contents |
|---|--------|------|----------|
| 0 | `0x00004000` | `0x120B6200` | the application (a "CXI" NCCH) |
| 1 | `0x120BA200` | `0x001E5200` | the electronic manual |
| 7 | `0x1229F400` | `0x021E6200` | system update data |

### NCCH — the application container

Partition 0 is an `NCCH` container. Its header names the title (`CTR-P-AREP`, program ID
`0004000000053f00`) and, crucially, `flags[7]` has bit 2 (**NoCrypto**) set — the dump is decrypted.
The header locates three regions relative to the partition start: the **ExHeader** (`0x400` bytes),
the **ExeFS** and the **RomFS**.

### ExHeader — the code-set layout

The extended header's system-control info is the loader's map of where code goes. It parses byte-exact:

```
title           CtrApp
flags           0x01  (ExeFS/.code is BLZ-compressed)
text    addr 0x00100000  674 pages (0x2a2000)  size 0x2a1784
rodata  addr 0x003a2000   63 pages (0x03f000)  size 0x03ee50
data    addr 0x003e1000   19 pages (0x013000)  size 0x0129d4
bss     addr 0x003f39d4  size 0x03d684
stack   0x20000
```

Two invariants make this self-checking, and the parser enforces both: each segment's `size` fits
inside its reserved page extent (`NumPages × 0x1000`), and the extents **tile contiguously** — rodata
begins exactly where text's extent ends (`0x100000 + 0x2a2000 = 0x3a2000`), data where rodata's does.
A violation would mean the header was misread (most likely: the container is still encrypted).

### ExeFS and the BLZ `.code`

The ExeFS is a small filesystem of four files:

```
.code   0x1a6b8c stored
banner  0x03afa8
icon    0x0036c0
logo    0x002000
```

`.code` is compressed with **BLZ** — Nintendo's "backward LZ77", an ordinary LZSS run *backwards*:
the stream decodes from the last byte toward the first, writing output from its end toward its
beginning, so the decompressor can work in place (the write pointer stays ahead of the read pointer,
and a badly-compressing prefix is simply stored verbatim). A footer at the end of the file carries
`compressedSize`/`footerSize` (bits packed into one word) and `originalBottom` (how much to add to
the file length to get the decompressed length). Decompressing gives `0x2f4000` bytes — which is
**exactly** `text.extent (0x2a2000) + rodata.extent (0x3f000) + data.extent (0x13000)`, the sum
pinned by the ExHeader that the compressed stream knows nothing about. That equality is the
round-trip proof the BLZ decode is correct; `ExeFS.Code` refuses any output that misses it.

### RomFS — the asset filesystem

The RomFS region opens with an **IVFC** hash tree: three levels of SHA-256 hashes over fixed-size
blocks, so the console can authenticate any one filesystem block via a short chain (master → L1 → L2
→ L3) without reading the whole 300 MB image. A subtlety cost one debugging round: the levels'
`LogicalOffset` fields describe a *logical* space in numerical order, but on media the **data level
(L3) is stored first**, then L1, then L2. Reading the logical offsets as media offsets lands in the
middle of a hash block; the level-3 header's own length field (`0x28`) is the check that catches it.
The real order was settled not by guessing but by SHA-256: every hash chain verifies and the levels
tile the region to the exact byte (`VerifyIVFC`, and `n3dsdump -verify`).

Level 3 is a conventional metadata-table filesystem (directory table + file table, UTF-16LE names,
`0xFFFFFFFF` link terminators). It flattens to **1,771 files in 33 directories, 298,216,940 bytes**:

```
/EffectData  /LayoutData  /LocalizedData  /ObjectData
/SoundData   /StageData    /SystemData
```

The payloads are `.szs` (Yaz0-compressed SARC archives), `.bcmdl`/`.bcmld` models, `.bcstm` audio and
the like — the standard Nintendo middleware containers. Decoding them is later work.

---

## Part II — Boot and execution bring-up

Unlike a bare-metal cartridge, a 3DS title is a **process** under the Horizon microkernel: the loader
maps its code segments, and the program reaches gameplay by talking to OS services over IPC. The
oracle (`bootoracle`) models the *process* view — it loads `.code` at the ExHeader addresses, lays out
the ARM11 userland memory map, and high-level-emulates the supervisor calls — and runs the retail code
until it needs a kernel facility not yet implemented, halting *explicitly* rather than diverging.

### The ARM11 and VFP

The 3DS's application processor is an **ARM11 MPCore — ARMv6K** with **VFPv2** hardware floating point,
a strict superset of the DS's ARMv5TE. Two ARMv6 encodings alias slots ARMv5TE assigns to other
instructions (LDREX/STREX over SWP, UMAAL over MUL), so the v6 decoder is consulted *before* the v5
one, never as a fallback. The core disassembles and executes the full ARMv6K integer additions
(LDREX/STREX, the parallel arithmetic and SEL, the pack/saturate/extend/reverse group, the signed
dual-multiplies, CPS/SETEND/CLREX) and VFPv2 (loads/stores, arithmetic, VCMP, VCVT, and the ARM↔VFP
transfers). A census of the decompressed `.code` shows the game leans on all of it — 271 PKHBT, 194
STREX, 179 LDREX, hundreds of extends and multiply-accumulates, and VFP throughout. A decode subtlety
worth recording: in the VFP data-processing encodings **bit 22 is the destination register's high bit
(D), not part of the opcode**, so the operation is `{bit23, bit21, bit20}`; folding bit 22 in makes
any instruction with a high destination register decode as the wrong op.

### The memory map

The oracle maps the userland an application process sees:

| region | address | notes |
|--------|---------|-------|
| code (text/rodata/data/bss) | `0x00100000` | one contiguous span; bss is the zero tail |
| process heap | `0x08000000` | grown by `svcControlMemory(ALLOC)` |
| main-thread stack | top `0x10000000` | grows down, size from the ExHeader (`0x20000`) |
| LINEAR heap | `0x14000000` | physically-contiguous allocations |
| config (shared) page | `0x1FF80000` | kernel→user, read-only |
| thread-local storage | `0x1FF82000` | TPIDRURO (CP15 c13) points here |

### The supervisor-call surface

`svc #n` traps to the HLE kernel. The memory and information calls are modelled for real (the runtime
depends on their results); kernel-object calls (threads, events, mutexes, sync) are stubbed to hand out
handles and report waitable objects as signalled; anything unimplemented halts with its number and PC.
The one place the emulated ABI had to be **traced, not assumed** was `svcControlMemory`. Its C prototype
order is not the kernel-entry order; disassembling the wrapper settled it:

```
00293054  PUSH {r0, r4}          ; save the out-pointer
00293058  LDR  r0, [sp, #8]      ; r0 = operation   (loaded from the stacked C arg)
0029305C  LDR  r4, [sp, #0xC]    ; r4 = permission
00293060  SVC  #1                ; r0=op, r1=addr0, r2=addr1, r3=size, r4=perm
00293064  LDR  r2, [sp]          ; *out = r1  (the mapped address)
```

### Heap sizing — a traced constant, not a guess

The runtime sizes its heap as `ReadWord(configPage + 0x40) − ResourceLimit(COMMIT)`, read straight from
the code at `0x00100750` (`LDR r0,[cfg]; LDR r0,[r0,#0x40]`) minus the value a `GetResourceLimitLimitValues`
supervisor call returns. With the config field left zero the subtraction underflowed to `0xFC000000`
(−64 MiB), which sent the allocator down an error path. Filling `config+0x40` with the application memory
budget (`APPMEMALLOC`, 64 MiB) and reporting the COMMIT figure as the committed base (code + stack, 4 MiB)
makes the heap a clean, page-aligned **60 MiB** (`0x03C00000`) — and the whole init then flows through.

### How far it runs

From a cold boot the oracle runs the C runtime's segment/BSS setup, the VFP-heavy library init, the
process-memory and resource-limit handshake, the heap allocation, and mutex/handle setup, and then
enters the **OS service handshake**: it connects to the service-manager port **`srv:`**
(`RegisterClient`, `EnableNotification`), and through it acquires and drives the first services — the
applet manager **`APT:U`** (lock handle, Initialize with its notification/resume events), the network
daemon **`ndm:u`**, and the system-config service **`cfg:u`**. `bootoracle` reports the whole service
transcript. It reaches **~725,000 instructions** — roughly triple the pre-IPC reach — before halting
on the next unimplemented service command, which is always named precisely.

The Horizon IPC layer is high-level-emulated (`ipc.go`, `ipc_services.go`): a `SendSyncRequest` reads
the caller's TLS command buffer, and each service is modelled just far enough to keep init moving.
One quirk is recorded and not yet explained — some `srv:GetServiceHandle` requests store the 8-byte
service name with each 32-bit word's halves swapped ("APT:U" arrives half-swapped, "ndm:u" straight,
both from the same thread), so the reader tries both orders and takes the one whose family it knows.

The remaining distance to a **rendered frame** is real work, not a mystery: completing the APT
applet-lifecycle handshake (the app receiving its startup parameter and transitioning to the running
state), the GSP graphics service (registering the shared command queue, then the game building GPU
command lists), and — for actual pixels — a **PICA200 GPU** implementation to execute those command
lists, an effort on the scale of this repo's N64 RDP or PSX GPU. The machine is built to reach the
point of *submitting* the first frame and to report it (`bootoracle` counts GPU command lists and
buffer swaps); producing the pixels is the next platform milestone.

---

## Part III — The HOME-Menu banner (the animated 3-D scene)

When a title is highlighted in the 3DS HOME Menu, the top screen shows a small **animated 3-D scene** —
for Super Mario 3D Land, a slowly turning logo lit and framed by a moving camera. That scene is the
ExeFS **`banner`** file, and it is a complete little 3-D program: a model, its textures and materials,
a camera, a light, a scene environment, and a skeletal animation. `bannerdump` runs the whole chain
from the cartridge in one command.

### CBMD → LZ11 → CGFX

The `banner` file is a **CBMD** ("Common Banner Model Data") container:

```
0x00 "CBMD"
0x04 u32   version
0x08 u32   offset to the common CGFX
0x0C u32[13] language-specific CGFX offsets (all 0 here — one shared scene)
0x40 u32   CWAV audio offset (0 here — silent banner)
0x88 …     the CGFX payload
```

The payload is not a raw CGFX: at `0x88` the bytes are `11 80 7d 04 …`, and the `0x11` type byte with
a 24-bit size (`0x047D80` = 294,272) is Nintendo **LZ11**, a forward LZSS. The tell that it is
compressed and not corrupt is exactly that the CGFX/DATA/DICT magics appear *interspersed* with binary:
those are LZ literals between back-references. Decoding the first flag byte confirms it — `0x00` means
eight literals, and they are `43 47 46 58 FF FE 14 00`, a clean CGFX header. LZ11 decompresses to
294,272 bytes, the size the header promised.

### CGFX — the scene graph

The decompressed blob is a **CGFX** ("CTR Graphics", Nintendo's NW4C scene format). Its structure is a
header, a **DATA** block of typed resource dictionaries, and a trailing **IMAG** block holding the raw
vertex and texture bytes. Every internal pointer is *self-relative* (a field at P holding V points to
P+V), and the dictionaries are **patricia trees** ("DICT") — but a linear walk of their nodes yields
every named entry, which is all an extractor needs. For this banner the scene graph is:

```
CGFX  revision 0x05000000  fileSize 294272
  IMAG (raw vertex/texture data) at 0x7878
    Models              1   COMMON  (CMDL)
    Textures            4   COMMON1 COMMON4 COMMON7 COMMON8  (TXOB)
    LookupTables        1   COMMON  (LUTS — lighting)
    Cameras             1   Camera1 (CCAM)
    Lights              1   Light1  (CFLT)
    Scenes              1   SceneEnvironment1 (CENV)
    SkeletalAnimations  1   COMMON  (CANM)
```

The **model** (CMDL) resolves further: it carries **8 meshes** (each binding a shape to one of the **4
materials**, MTOB) and **8 shapes** (SOBJ) that hold the geometry — a ninth SOBJ is the skeleton itself. A shape's vertex-buffer descriptor
points into the IMAG block — the first shape's vertices are a 2,922-byte run at IMAG offset `0x8`, with
16-bit (`GL_UNSIGNED_SHORT`) indices — and each material's texture reference names one of the four
TXOBs. The **CANM** is what makes it move: a skeletal/transform animation over the scene, the reason the
logo turns.

### Storage summary

```
ExeFS/banner  =  CBMD
                 └── LZ11-compressed  →  CGFX
                     ├── DATA   descriptors: CMDL (8 meshes / 8 shapes / 4 materials),
                     │          4× TXOB, LUTS, CCAM, CFLT, CENV, CANM
                     └── IMAG   raw vertex streams + tiled texture pixels
```

### The GLB export

`webexport` decodes the whole scene from the cartridge and writes one GLB —
`site/public/super-mario-3d-land-3ds/models/banner.glb` — that a standard glTF viewer plays. The
mapping is one-to-one with the CGFX: a node per bone (parented per the skeleton), a mesh per CGFX
mesh attached to its bone's node, and the CANM curves carried as glTF **CUBICSPLINE** channels — the
source keys are (frame, value, in/out-slope) hermite triples, which map onto glTF's (in-tangent,
value, out-tangent) form with no resampling. The banner's rest pose has no bone rotations, and the
scene is Y-up like glTF, so no reorientation is applied (a Z-up guess visibly tilted the logo — the
render is the check).

Each shape's **interleaved vertex buffer** de-interleaves by its attribute descriptors: position and
normal are float32 triples; vertex colour is four unsigned bytes scaled to [0,1]; the two UV sets are
float32 pairs (V-flipped to glTF's top-left origin). Indices are the 16-bit streams read straight from
IMAG. The **textures** decode from the 3DS's native layouts — the colour atlas and Mario's skin are
**ETC1** (4×4 block compression, four blocks per 8×8 tile in Z-order), the title's cutout layers are
**L4** (4-bit luminance) — all un-swizzled from their 8×8 Morton-ordered tiles.

The one place glTF cannot mirror the hardware is the **texture combiner**: the title's flat faces are
plain quads that the PICA cuts to the logo silhouette by multiplying the colour atlas (sampled by
UV0) with a 4-bit mask (sampled by UV1). glTF binds one texture per material, so the exporter **bakes
the combine per texel** — it rasterises each such mesh in mask (UV1) space, interpolates UV0
barycentrically (exact, since UV0 is affine within a triangle), and writes the atlas colour with the
mask as alpha — reproducing the fragment result exactly. Only the extruded title-side mesh drops its
secondary depth-shade layer; that is the single documented approximation. The result renders correctly:
Mario mid-run in his textures, the golden Super Leaf, the question block and the brick platforms, and
Mario and the leaf bobbing on the baked animation loop. The Studio hosts it under a new Nintendo 3DS
system entry.

---

## Tooling

| tool | role |
|------|------|
| `tools/platform/n3ds` | NCSD/NCCH/ExHeader/ExeFS/RomFS parsers, BLZ + LZ11 decompressors, CBMD/CGFX banner parsing, the machine + SVC & IPC HLE |
| `tools/platform/n3ds/cmd/n3dsdump` | list/extract the cartridge's containers (`-romfs`, `-verify`, `-code`, `-x`) |
| `tools/platform/n3ds/cmd/bannerdump` | decode the HOME-Menu banner scene (`game.cci`; `-o` writes the CGFX) |
| `games/super-mario-3d-land-3ds/extract/cmd/webexport` | export the banner to `site/public/.../banner.glb` (`-texdump` writes the decoded PNGs) |
| `tools/cpu/arm` (`V6K` variant) | ARMv6K + VFPv2 disassembler, code-tracer and execution core |
| `tools/cmd/disarm -v6`, `codetracearm -v6` | ARM11 disassembly and recursive-descent tracing |
| `games/super-mario-3d-land-3ds/extract/cmd/bootoracle` | boot and run the ARM11 program under the HLE kernel |

Every result here is one command from reproducible against the pinned image.
