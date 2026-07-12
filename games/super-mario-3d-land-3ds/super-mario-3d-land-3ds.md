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

### How far it runs — into the render loop

The oracle now carries the title from a cold boot all the way into its **render loop**: it runs the C
runtime, the VFP-heavy library init, the process-memory/resource-limit handshake and heap allocation;
it drives the full **OS service handshake** (`srv:` → `APT:U` applet lifecycle, `ndm:u`, `cfg:u`,
`fs:USER`); it loads its **real assets** from the RomFS; it registers the **GSP** graphics shared
memory and its VBlank interrupt relay; and it then runs its frame loop, **submitting PICA200 GPU
command lists** — 5 command lists over the first ~600M instructions, with no halt. `bootoracle` reports
the whole service transcript plus VBlanks delivered and GPU command lists submitted / buffer swaps.

Getting there took an OS bring-up in five parts, each landing on a verifiable reach:

- **A cooperative thread scheduler** (`thread.go`, `sync.go`). Horizon threads are kernel-scheduled, so
  the HLE actually runs them: each thread is a whole-`arm.CPU` context snapshot plus its own TLS page,
  run highest-priority-first for a quantum, with *real* synchronisation — events, mutexes, semaphores,
  `WaitSynchronization1/N`, and the `ArbitrateAddress` futex/condvar path libctru's LightLock and
  condition variables need. A blocked `svc` yields cleanly; a waker writes the ABI result into the
  parked thread's saved context and marks it ready, so it resumes past the `svc` with the right
  registers (the DS dual-core `deliver` shape).
- **`fs` backed by the real RomFS.** A title opens `ARCHIVE_ROMFS` and expects `fs` to have stripped
  the IVFC hash-tree wrapper and to present the **level-3 filesystem** directly (offset 0 = the RomFS
  header). Handing back the raw IVFC container instead made the game's own directory hash-chain walk
  read metadata 0x1000 bytes too early and loop forever; returning the region from the level-3 media
  offset fixes it and the game loads its assets.
- **Shared memory + the VBlank heartbeat** (`gsp_mem.go`, `gsp_vblank.go`). `svcMapMemoryBlock` binds a
  block handle to real bytes, so the GSP shared memory maps. On an instruction-paced frame boundary
  (deterministic for savestates) the machine pushes a VBlank into the shared-memory interrupt queue and
  signals the GSP event — Horizon delivers GPU interrupts as event signals through shared memory, not
  IRQ vectoring, so this rides the scheduler's event/wait path. Verified end to end by tracing: VBlank →
  GSP event → the game's GSP event thread drains the queue → it signals the render condvar → the main
  thread wakes.
- **The GX command path** (`gx.go`). The game posts GX commands to a FIFO in GSP shared memory and
  blocks on their completion interrupts. The machine emulates the GSP module's side: it drains the FIFO
  and raises each command's interrupt — **P3D** for a PICA200 command list, **PPF** for a display
  transfer, **PSC0/1** for a memory fill. Since Phase 4 the commands are also *executed* (the DMA
  copies, the fills fill, the transfers detile, and a P3D list runs on the GPU model below).
- **Two synchronisation-correctness fixes**, both traced from concrete lock-ups. `ArbitrateAddress`
  `DECREMENT_AND_WAIT_IF_LESS_THAN` must decrement *then* compare (park when `*addr <= value`), or the
  first waiter on a locked LightLock never registers and spin-starves the holder. And the `LDREX`/`STREX`
  exclusive monitor must be cleared on every context switch (as a real OS issues `CLREX`), or a `STREX`
  straddling a switch mistakes another thread's write for its own and corrupts shared lock words.

### The PICA200 GPU (Phase 4) — command lists into pixels

The render loop's `GX ProcessCommandList` submissions are now **executed**, not just acknowledged. The
work was done instrument-first: `bootoracle -gxdump` captures every GX command's raw FIFO slot — and,
for a ProcessCommandList, the command-list bytes at submission time, because the game reuses list
memory between frames — and the `picadump` tool decodes a captured list into its **register-write
stream** (a PICA200 command list is not drawing commands but a sequence of GPU register writes, the
same idea as the N64 RDP command list: parameter + header entries with a register id, byte-enable
mask, and burst extras).

What the first frame's capture shows, all derived before implementing:

- **~320 `RequestDMA` commands** stage vertex and texture data from the linear heap into **VRAM**,
  which the game addresses through its fixed virtual window `0x1F000000‑0x1F5FFFFF`; the same buffers
  appear as *physical* `0x18000000` addresses (`>>3`) in the command lists' framebuffer registers.
  The machine now maps VRAM at that window and really executes the DMAs, fills and transfers.
- An **all-state init list** (29,824 bytes, 6,590 writes, no draw trigger) touches every functional
  block once and zeroes the full 4096-word shader code memory through two upload ports.
- Per-frame lists upload the **real vertex shader** (code via `0x2CB/0x2CC`, operand descriptors via
  `0x2D5/0x2D6`), set uniforms through the float FIFO (both 32-bit and packed-24-bit modes, **w
  first** — the order the shaders' pervasive `.wzyx` matrix-row swizzles assume), configure the
  attribute buffers, and fire `0x22E` draw triggers: 17+17 quad batches into the two 400×240 top-eye
  targets and 3 into the bottom screen's, each preceded by a `MemoryFill` clear (colour `0x0000FFFF`
  = opaque blue; depth `0x00FFFFFF` = far in D24S8) and followed by a `DisplayTransfer` that detiles
  the rendered target into an RGB565 linear framebuffer.

The GPU model executes all of it (`gpu.go`, `gpu_shader.go`, `gpu_raster.go`, `gpu_tev.go`,
`gpu_texture.go`): the register-file interpreter with side-effecting upload FIFOs; an **LLE
vertex-shader VM** — running the game's own uploaded shader binary, the clean-room equivalent of
LLE'ing the N64 RSP microcode — whose instruction decode was validated by disassembling that shader
(`picadump -shader`) into a coherent transform program (matrix `dp4` rows read `.wzyx`, `abs` as
`max(a,−a)`, `ifc`-selected uniforms); a perspective-correct barycentric rasteriser (the PSX
triangle's shape plus 1/w interpolation) with the viewport/depth-map transform, D24S8 depth testing
and 8×8-Morton tiled RGBA8 writes; the six-stage TEV combiner (the N64 `(a−b)·c+d` family) with
alpha test and the full blend unit; and texture units decoding the tiled PICA formats (RGBA8 … A4,
ETC1 via the banner's decoder) with a per-address decode cache. Unknown structural features —
geometry shader, fragment lighting, unseen formats — halt loudly, and get implemented as the game
demands them (packed-f24 uniforms, then the L4 texture format, arrived exactly that way).

A full boot now runs the five first-frame command lists — 37 draws, ~267k fragments — and
**presents the frame**: `bootoracle -shot` writes both screens' framebuffers as PNGs (rotated from
the panel's 240-wide column layout to natural landscape). The presented top frame is the game's blue
clear: the capture shows the game deliberately **poisons its projection uniforms `c0‑c3` with NaN**
at frame start, so those first-frame triangles fail clipping on real hardware exactly as they do
here (the rasteriser rejects NaN/w≤0 triangles). The visible content comes in later frames, once the
game's warmup (logo timers, asset decompression) finishes — the current frontier.

Chasing why the loop stalled after its first frame found a real ARM-core bug, the Phase-4 analogue
of Phase 0's LDRD/STRD find: **VFP `VNMLS`/`VNMLA` had swapped product signs**. The game's
random-point-in-unit-sphere sampler maps uniform `u` to `2u−1` with `VNMLS` (acc=1, s17=2); the
swapped sign computed `−1−2u`, every candidate had length > 1, and the rejection loop spun forever.
The instruction-level trace of the spinning thread — an LCG at `0x0021F194` feeding a
`VSQRT`-and-retry loop at `0x002EF810` whose lengths came out above √3 — identified it; the fix is
unit-tested across all four multiply-accumulate sign conventions and the DS regression suite still
passes on the shared core.

### To the first legible frame — the wakeup, the save archive, and a dialog box

Past its warmup the game parks its whole main thread with **APT `NotifyToWait`**, expecting the APT
module's wakeup — a signal on the events `Initialize` returned, which the app's APT handler thread
answers by running its `InquireNotification`/`ReceiveParameter` path and releasing the parked
thread. Two subtleties, both traced: the wake must be **asynchronous** (signalling inside the
NotifyToWait reply races the caller, which still holds the library's cached APT session handle —
global `0x003E2668` — for ~50 more instructions; a handler woken inside that window throws the
applet-module fatal `0xE0A0CFF9`, found with a write-watch on that global), so the HLE arms it in
the reply and delivers it at the next VBlank. With the wakeup delivered the render loop runs at
full cadence — one command list per frame instead of five ever.

The game then initialises its **save data**: it enumerates a RomFS directory (`OpenDirectory` and
IDirectory `Read` now serve real `FS_DirectoryEntry` records), and creates, chunk-writes and
re-reads `/CFL_DB.dat` — 310,560 bytes, its Mii face-library database — through a writable
in-memory archive (`OpenArchive`/`CreateFile`/`DeleteFile`/`Write`; the command layouts came from
the game's own IPC wrappers, e.g. CreateFile's at `0x001EDE30`, after a first-guess layout produced
an empty path with a nonsense size). A lying `Write` ack would have failed the game's read-back
verification, so the store is real and rides the savestate.

**The run then renders the first legible frame**: the bottom screen presents a complete 3DS message
dialog — rounded grey panel over a darker backdrop, a gradient-filled yellow **Ⓐ OK** button, and
glyph-rendered text — every pixel produced by the LLE shader → rasteriser → TEV → texture pipeline
(the button label sits in an LA4 glyph atlas, the panel in ETC1). Two byproducts of the milestone:
the screen-rotation direction in the PNG capture was settled by the first mirrored text, and the
dialog's body text resolves to "NULL".

### HID input injection — the oracle drives the pad (verified into the game's key state)

The `hid:USER` shared-memory block, once a dead zero page, is now a live pad-state ring the oracle
publishes into every VBlank (the input driver's per-frame job). The layout was **not guessed**:
`bootoracle -hidtrace` tallies the game's own reads of the block by offset (instrument-first), and
the game's HID reader/copier/scan chain was disassembled (the sample copier at `0x00297698`,
`hidScanInput` at `0x001F52F0`, the pad-manager update at `0x002DA504`) to pin the exact structure —
per section, two 64-bit timestamps at `+0x00`/`+0x08` (the driver bumps them each sample; the game
diffs them against its saved copy to count new entries), the latest-entry index at `+0x10`, then an
8-slot ring of `0x10`-byte entries at `+0x28` whose first word is the current button mask (the game
derives keys-down itself by diffing successive polls). `bootoracle` gained `-keys`
(a/b/x/y/l/r/dpad/start/select) and `-hidtrace`; the block, its interrupt events and its mapped
address ride the savestate (recovered from the region name for pre-HID snapshots). **Proof it
works:** with `-keys a`, a memory watch shows the game's own derived held-buttons word (input
context `+8`, at `0x15923700` in the dialog state) flip `0 → 1` — the A bit — exactly as on
hardware. A first modelling pass wrote the button mask into the index field at `+0x10`; the
disassembled reader corrected it. Injection reaching the game's key state was validated before it
was claimed.

### The dialog is an error/notice, not a press-A prompt — the "NULL" is the frontier

With injection proven, the surprise: the dialog **ignores it**. Holding or pulsing A leaves the boot
re-rendering the identical `"NULL"` frame, instruction-for-instruction — the draw count does not
budge. So the dialog is not "press Ⓐ to continue"; it is a notice/error state whose body text is a
message lookup that failed at construction (`-findutf16 NULL` locates the dialog's UTF-16 string
object at `0x161B2954` in the snapshot, body length 8 = `"NULL"`, resolved once and cached — a
breakpoint on the message getter never fires again during the render loop). The `err:f` fatal path
is never invoked, so this is the game's **own** notice, not a Horizon fatal.

Two boot regressions were fixed chasing it. **fs `CloseArchive` (0x080E)** was missing — the game
closes its save archive after building the Mii database, and a cold boot halted there at 126M
instructions (the dialog savestate had masked it). And **cfg:u was returning a zero page**:
`GetConfigInfoBlk2` replied success while leaving the output buffer untouched, so the system-language
block (`0x000A0002`), sound-output (`0x00070001`), agreed-EULA version (`0x00130000`) and
stereo-camera (`0x00050005`) all read as zero. `writeConfigBlock` now fills them with a European
English console's values. (An earlier trace shows the game already *defaults* its internal language
to English from a zero config — mapping region/language 0 through its own table at `0x00100C7C` — so
the message folder was already `EuEnglish`, which the cartridge does contain; the honest config is a
correctness fix, and a cold boot with it **confirms the dialog is unchanged** — same `"NULL"`, the
same thirteen UTF-16 `"NULL"` occurrences at the same addresses. So the `"NULL"` was not a config
failure.)

### The dialog identified: the StreetPass first-launch welcome ("Welcome to Super Mario 3D Land.")

Tracing the message getter (`0x00266DDC`, which returns a literal `"NULL"` when a lookup fails) with
the new `-logpc` breakpoint pinned the dialog to a single builder — a generic **WindowConfirmSingle**
layout window (body pane `TxtMessage`, button pane `WindowConfirmSingle_OK`) built at `0x00225554`.
Breaking on its message-fetch call (`0x00225660`) and dumping the id-holder revealed the label it
asks for: **`StreetPassBegin00`**. Decoding the cartridge's own message archive settled what that is.
`LocalizedData/EuEnglish/MessageData/SystemMessage.szs` is Yaz0 → NARC → 24 MSBT (`MsgStdBn`) files;
a new clean-room decoder (`tools/platform/n3ds/message.go` + `cmd/msgtool`) reads them, and
`StreetPassBegin00` is `"Welcome to <title>."` — the first line of SM3DL's **StreetPass first-launch
flow**: `StreetPassBegin00/01/02` (welcome + what StreetPass is) → `StreetPassSetting` ("Would you
like to activate StreetPass for this game?" Activate / Cancel) → `StreetPassDisable` ("Did not
activate StreetPass. You can activate StreetPass at any time from **the title screen**"). So the boot
"dialog" is **not an error** — it is the normal first-launch StreetPass prompt, and the title screen
lies just the other side of it.

Why it renders `"NULL"` — **a CPU bug, found and fixed.** The message data is genuinely loaded (the
label `StreetPassBegin00` and its UTF-16 text sit in the heap at `0x158D50xx`, and the dialog's
message manager points at that very MSBT file — `0x158D4C80`, the 16-message System file). Tracing the
lookup (`-tracefrom 0x00225660`, the dialog's message fetch) through the MSBT resolver
(`0x002842F0`) showed the hash (`hash*0x492 + char`) and the software modulo both compute correctly —
`StreetPassBegin00` hashes to bucket 77, exactly where the file stores it, and the byte-compare
matches. The failure was the very last step: the resolver loads the message **index** with
`LDR r0, [r0, r1]` from `0x158D50D7` — an **unaligned** address, because an MSBT label entry is
`{len:u8, name, index:u32}` and the u32 index follows the variable-length name. The bytes there are
`06 00 00 00` = 6, but the core returned `0x600`: `read32` read the true unaligned bytes **and then**
applied the ARMv5 rotate (`6 ROR 24 = 0x600`), a hybrid that is wrong both ways. The manager's bounds
check (`count 16 <= 0x600`) then failed and the getter returned the literal `"NULL"`. ARMv6 (the
3DS's ARM11, unaligned access enabled by Horizon) must do a **true** unaligned load with no rotation;
the fix branches on the architecture, and the traced lookups go from 36 NULLs to zero. This is a
general core fix, not a message-system patch — any unaligned `LDR` was corrupt, so it would have bitten
elsewhere (the third such CPU find of the port, after the LDRD/STRD and VFP `VNMLS`/`VNMLA` bugs).

With messages resolving, the boot dialog becomes the real StreetPass welcome — the bottom screen reads
**"Welcome to SUPER MARIO 3D LAND"** — and the HID injection drives the whole first-launch flow to the
menu. A held button only produces one keys-down edge, which a dialog swallows during its open
animation, so `-keypulse N` releases the injected keys briefly every N frames to keep fresh press
edges arriving: **Welcome (Begin00) →Ⓐ→ Begin01 →Ⓐ→ Begin02 →Ⓐ→ "Would you like to activate
StreetPass?" →Ⓑ Cancel→ "Did not activate StreetPass…" →Ⓐ→ the A/B/C NEW GAME file-select menu**,
fully rendered. Two small unblocks were needed once past the dialogs: the PICA **TEV source 1
(PrimaryFragmentColor)** the menu uses — with fragment lighting disabled (enabling it halts earlier)
the primary colour passes through, and source 2 (secondary) is zero — and **APT `PreloadLibraryApplet`
(0x0016)**, which the title issues to preload a helper applet before the menu (acked; we do not run
library applets). The oracle now boots from cold, past the StreetPass onboarding, into the game's
interactive save-slot menu. (The top screen's logo art stays on the clear colour there: the game
issues no further GPU work while idle at the menu — the next frontier is selecting a file to start a
new game, and rendering the first in-game frame.)

### Selecting a file: the library-applet conversation and the save-creation path

Pressing Ⓐ on slot A surfaced a chain of unmodelled OS behaviour, each piece traced from the game's
own wrappers before implementing:

- **The menu was not idle — it was polling.** At the file-select the main thread re-queries **APT
  `GetAppletInfo` (0x0006)** every 10 ms (the loop at `0x002324FC` sleeps `svcSleepThread(10^7 ns)`
  between tries) until the reply's cmdbuf[5] and cmdbuf[6] bytes — the queried applet's *registered*
  and *loaded* flags — both come back nonzero. The stub reply left them zero, so the boot span at the
  menu forever. We ack `PreloadLibraryApplet` without running anything, so the HLE now reports the
  applet present and loaded.
- **The game then talks to library applet 0x402.** It sends a 32-byte parameter (**`SendParameter`
  0x000C**, signal 2) and parks its main thread until the applet answers: the wait loop
  (`0x00232BA0`, exit check at `0x00232C7C`) demands a received parameter with **sender = 0x402 and
  command = 3**, and maps the carried shared-memory block whole (`svcMapMemoryBlock` with size 0,
  then **`svcUnmapMemoryBlock` (0x20)**, now implemented). Then it **starts** the applet
  (`StartLibraryApplet` 0x001E, wrapper `0x00296F50`) and waits again — this time in a buffered
  receive loop (`0x002915E4`, an 0x84-byte response buffer) that accepts command values
  {1,0xA,0xB,0xC,0xD,0xE,0xF,0x11}, with 0xA mapped to its own "applet finished" class
  (`0x002917D8`). Since this HLE runs no applets, it fabricates the answers: a queue of pending
  parameters (`aptParams`) delivers `{sender 0x402, command 3, minted memory block}` after
  SendParameter and `{command 0xA, 0x84 zero bytes through the receiver's TLS static buffer}` after
  Start, each consumption re-arming the deferred APT wake while more remain. `gsp` commands
  **0x0019/0x001A** (bare-header, result-only — a save/restore pair around the applet hand-off) and
  the send-only **APT 0x0040** buffer hand-off are acknowledged.
- **The save-creation path is a state machine steered by fs error classes.** The flow is
  `GetFormatInfo (0x0845)` → `FormatSaveData (0x084C)` → open-with-create → `ControlArchive (0x080D)`
  commit. Three separate lies had to become truths:
  1. `GetFormatInfo` was blanket-acked, leaving stale request words as the "format info"; the game
     took the garbage for a valid save, opened `/GameData.bin`, got NotFound and threw fatal
     `0xC8804478` under a "Saving…" toast that never finished. It now returns the info recorded by
     `FormatSaveData` (newly implemented: records `{blocks, dirs, files, duplicateData}`, erases the
     store) — and, before any format, an error of the right **class**: the game's save layer
     (`0x001A5A68`) extracts the fs result's description field and forgives only **[0x154, 0x168)** —
     the "save is fresh" group — so the unformatted reply is `0xC8804554`, not the RomFS NotFound
     (description 0x78), which it throws as fatal.
  2. The reverse mistake looped forever: reporting the 0x154 class *after* a successful format reads
     as "format failed", and the game re-formats endlessly. Post-format misses are plain NotFound.
  3. The game never calls CreateFile for its save — it opens `/GameData.bin` with **open flags
     WRITE|CREATE (0x6)**, which the HLE had ignored. `OpenFile` now honours CREATE in the save
     archive, and `IFile SetSize (0x0805)` really resizes (the game sizes the fresh file, chunk-writes
     it, and verifies by read-back).

With those, slot A relabels to **WORLD 1-1** with a "Saving…" toast, the save commits, and the game
plays its screen transition — which needed two last GPU features: **GX `TextureCopy`** (the PPF
engine's gap-aware byte copy; the observed slot pins the dim words as `gap<<16 | width` in 2-byte
units — 400 lines of 960 bytes in, a 1024-byte stride out: the rendered frame grabbed as a
256-px-wide power-of-two texture for the transition) and **primitive mode 3** (the "geometry
primitive", meaningful only to a geometry shader; the draw-time check guarantees the geometry stage
is off, so its vertices assemble as independent triangles).

### The machine had no clock: `CPU.Instrs` is per-thread

The VBlank heartbeat and the GX completion deadlines were paced on `CPU.Instrs` — and that is not a
machine clock at all. A thread's whole `arm.CPU` *is* its context (`switchTo` does `*m.CPU = t.ctx`),
retired-instruction counter included, so the value is saved and restored with every context switch and
moves **backward** whenever the scheduler picks an older thread. The heartbeat was riding a counter
that sawtooths. It happened to look sane while the boot ran mostly on one thread, and fell apart as
soon as the frame machinery migrated onto the APT and worker threads: the frame clock stalled, which is
why the render loop starved.

`Machine` now carries a monotonic `instrs`, incremented once per retired instruction in the run loop
and never restored from a context; VBlank and GX ride it, and it is serialised (old snapshots seed it
from the live CPU). The `tick` that backs `GetSystemTick` stays separate — it is *jumped* forward by
the idle-sleep fast-forward, so it must not pace frames either. On the same cold-boot budget the fix
takes the run from 3,340 to **4,053 GPU command lists** and from ~320M to **941M pixels**. This is the
kind of bug the port keeps producing: not a missing feature, but a core assumption that was quietly
false.

### GX completion is now asynchronous — and a ruled-out hypothesis

On hardware the GSP system process services the GX command FIFO and raises each command's completion
interrupt *later* than the application posted it, while the CPU runs ahead; the application's command
runner posts, marks the slot pending, and blocks on the interrupt. The oracle now models that: a GX
command is **accepted** immediately (the FIFO index/count advance, as before) but **completes** on a
deadline — a per-command latency (a `ProcessCommandList` list costs more than a bare `RequestDMA`),
strictly in submission order — and only at that instruction boundary does it execute and raise its
interrupt. The run loop treats the nearest pending completion as a wake source, jumping the clock to
it when every thread is otherwise blocked (bounded by the VBlank heartbeat). The latencies are
nominal model parameters (hardware-plausible magnitudes, the same footing as the instruction-paced
`GetSystemTick`), not measurements; the pending queue rides the savestate. This is a faithfulness
improvement, kept because it removes GX-timing as a variable in everything downstream.

It was reached chasing the render-ring deadlock below — on the hypothesis that synchronous P3D
completion was starving the driver's retire path — and it **did not** move the deadlock: a boot with
async completion parks at the identical instruction. The negative result is itself informative. The
game's GX P3D interrupt drives a *frame-counter* object (`0x0041D1B0`, updated by the handler thunk
at `0x00107218`), which is a different structure from the stuck command ring (`0x00425814`); GX
completion timing and the ring's retirement are orthogonal, so the ring stall is not a GPU-timing
artefact.

### The current frontier: the render-command ring, filled during onboarding and never drained

Past the transition the screen clears and the main thread blocks inside the game's own render-command
driver (the singleton at `0x00425814`). The structure, mapped from the disassembly and a cold-boot
construction trace: the driver embeds several FIFOs, all built once by a generic queue constructor
(`0x003926BC` → `0x00392710`) — a pair of double-buffered command rings and a node free-list. The one
that deadlocks is a **32-entry ring at `+0xB4`** (`0x004258C8`, its backing array at driver `+0xE8`,
count at `+0x28`, capacity at `+0x20`). Commands are submitted through `0x0022B0CC` (grab a node from
the free-list at `+0x168`, fill it, push), which **try-pushes** — silently succeeding while there is
room — and retired through a separate **pop** path (`0x0022AF84`, and a pop-until-my-fence loop at
`0x0022B2AC`).

Three facts, established this session, pin the shape of the bug:

- **The ring only ever grows.** A memory-watch on the count field across the whole boot shows
  monotonic increase — 13 entries at the first dialog, 21 after the StreetPass flow, a full 32 at the
  file-select — with every write coming from the push site and *not one* from a pop. It is never
  drained, from cold boot onward.
- **There is no dedicated consumer.** The ring's own condition variables — the "space available"
  (`+0xC`, `0x004258D4`) the producer waits on when full, and the "data available" (`+0x4`) a consumer
  would wait on when empty — are never signalled *or waited on* by any thread during onboarding
  (watched: zero activity). The driver's init spawns no worker for it. The two threads that do park on
  driver-relative arbiters (`+0x678`, `+0x1348`) are an idle job-pool, unrelated to this ring and
  themselves never signalled. So the ring's only consumer is the **main thread's own inline flush**
  (`0x0022AF84` / the pop-until-fence loop) — and that flush is *never reached* during the entire
  onboarding (zero calls across the flow). The menu still renders because it submits GX directly; this
  ring belongs to the *in-game* renderer, filled but never pumped before gameplay.
- **The deadlock is that flush's absence catching up.** The post-save scene setup performs the first
  **blocking** push — a render object's finalize (a vtable `+0xC` method invoked during a refcount
  release, chain `0x0027B1B4` → … → `0x001DF5D8` → `0x0022B0CC`). Because the ring is already full of
  never-retired onboarding commands, the blocking push waits on `+0xC` for a retire that only its own
  now-parked thread can perform: a true self-deadlock. A probe (`bootoracle -poke`, a new instrument)
  that widens the ring lets that push through, after which the retire loop finally runs but, finding
  no chain matching its fence, ends waiting on pop-empty — confirming the ring was full of *stale*
  work, not live work.

So the ring is not a per-frame command queue that cycles; it accumulates roughly eight entries per
onboarding *screen* and is meant to be drained by a flush the main thread never executes during
onboarding.

Chasing *why* that flush never runs led up the chain that actually paces the renderer, and turned up
three real gaps — each fixed, none of them yet enough:

- **The applet-exit resume order.** A library applet's exit is followed on hardware by a parameter
  carrying **command 8**, the APT module's "resume the application". The game's registered APT callback
  (`0x00104200`, run on its APT thread at each wake) dispatches parameter commands `{2,5,8,9}`
  (`0x001044A4`), and **8 is the only path that re-arms the frame machinery** — `0x0028DBD0(0)` → the
  frame-request walk `0x0028B9B0` → the per-VBlank latch that paces the render thread.
  `StartLibraryApplet` now queues it after the exit answer, and it is verified firing.
- **The DSP data-register handshake.** The resume path's audio restart polls dsp::DSP register 0 —
  `RecvDataIsReady` then `RecvData` (wrappers `0x001F2FF4` / `0x001F3244`, answers read out of
  `cmdbuf[2]` as a byte and a u16) — and spins in a retry loop (`0x001F9488`) until it reads 1, the
  component's "running" word. Against a blanket ack it read a stale word and span forever. Answered,
  the restart completes. (Reporting the component *absent* instead is worse: it regresses the boot's
  own audio init — tried, and reverted.)
- **`PrepareToStartLibraryApplet` (APT 0x0018)**, which the title only reaches now that the frame clock
  runs at all.

With these the resume chain runs much further: the audio restart finishes and the render threads park
on `arb@0x00426B5C` — a frame-delivery condvar that nothing yet signals. But the ring still sits at 32
and the main thread still blocks on its space condvar, so **the deadlock is not fixed**. What that
condvar is, and who is supposed to signal it, is exactly where the next session starts. GX-completion
timing has already been ruled out, and the machine clock is no longer a confound.

One quirk is recorded and not yet explained — some `srv:GetServiceHandle` requests store the 8-byte
service name with each 32-bit word's halves rotated ("APT:U" half-swapped, "fs:USER" byte-rotated,
"ndm:u" straight, all from the same thread), so the reader tries each rotation and takes the one whose
service family it recognises.

### The DSP becomes real — and this title's audio init turns out to have been bailing

Captain Toad (the second 3DS title, see its own writeup) forced a real `dsp::DSP` model —
`tools/platform/n3ds/dsp.go`: the DSP RAM window at `0x1FF00000`, the audio-pipe state machine, the
two double-buffered shared-memory regions with their 15 announced structures, the per-source
config→status protocol, and an audio-frame clock (160 samples ≈ 1,310,720 cycles) paced on the
machine's monotonic counter and armed at `LoadComponent`. `RecvData(0)` is now a real state readout —
0 while the pipeline runs, 1 once it is shut down or sleeping — which subsumes the hard-coded 1 the
applet-resume retry loop (`0x001F9488`) was given earlier: this title only polls after ordering a
stop, when the state machine genuinely answers 1.

Two things this title taught in return:

- **Its audio init had been bailing all along.** Pre-DSP, SM3DL's observed conversation was
  `LoadComponent`, 79 × `FlushDataCache`, `RegisterInterruptEvents`, `GetSemaphoreEventHandle`,
  `GetHeadphoneStatus` — no pipe traffic at all, which looked like "this title just doesn't use the
  pipes". With coherent replies the same boot now runs the full standard protocol on its audio thread
  (t7): … `GetSemaphoreEventHandle` → `SetSemaphoreMask` → **audio-pipe `Initialize`** →
  `ReadPipeIfPossible` → 30 × `ConvertProcessAddressFromDspDram`. The old truncated list was the
  *failure signature* of the fake DSP — the init aborted before ever writing a pipe — and it is also
  why the earlier "just mint the semaphore-event handle" shortcut regressed this title: the handle was
  real but nothing would ever signal it. The lesson generalises: a service's observed command list is
  only as trustworthy as the replies it was observed against.
- **The regression guard passed with a 5× improvement.** The same 6-billion-instruction cold boot
  that produced 4,053 GPU command lists / 3,991 display transfers / ~941M pixels against the blanket
  acks now produces **20,270 lists / 20,209 transfers / ~4.89B pixels**, with 1,559 dsp::DSP requests,
  and still lands on the fully rendered StreetPass welcome dialog. The boot is not merely unharmed —
  with its audio init succeeding instead of failing-and-retrying, the game reaches its render loop
  markedly earlier and runs it at full cadence.

Whether the real DSP moves the render-ring frontier above was then **verified rather than assumed**.
The old savestates predate the DSP (restoring them resurrects a world where the game already holds
garbage DSP handles), so the onboarding flow was re-driven from a post-DSP cold boot in one pass —
`-keys a -keypulse 30` holds Ⓐ through the whole flow, which takes the *activate*-StreetPass branch
this time (cecd acked) and still reaches file-select, creates the save, and runs the post-save
transition. The result is the identical end state: the game spawns its full in-game thread
complement (14 threads), its sound thread parks healthily on the DSP event each audio frame — and
the render thread parks on **the same frame-delivery condvar `arb@0x00426B5C`**. So the DSP was
necessary (the resume chain's audio restart runs against a real state machine now) but **the
render-ring deadlock is independent of it**, and remains the frontier: find what is supposed to
signal `0x00426B5C`.

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
