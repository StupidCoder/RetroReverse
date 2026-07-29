# Crazy Taxi (Sega Dreamcast)

**Image:** `Crazy Taxi (US).bin` + `Crazy Taxi (US).cue` — 1,188,056,352 bytes, MD5 `e446e7a09dc23985359c7cb9d1034639`.
Not committed (copyright); supply your own copy.

Sega's 1999 open-city taxi game, on the repo's Dreamcast platform: an SH-4
(SH7091) toolchain in `tools/cpu/sh4` (`dissh4`, `codetracesh4`, execution
core), the machine in `tools/platform/dc`, the static disc inspector
`tools/platform/dc/cmd/gdinfo`, and the oracle in
`extract/cmd/bootoracle`.

## Part I — The disc

### The container

The rip is a `.cue` in **cdrdao TOC syntax** (not CDRWIN cue) over a single
`.bin` of raw 2352-byte sectors:

```
TRACK MODE1_RAW   DATAFILE "Crazy Taxi (US).bin" 00:08:00
TRACK AUDIO       DATAFILE "Crazy Taxi (US).bin" #1058400 09:52:00
TRACK MODE1_RAW   DATAFILE "Crazy Taxi (US).bin" #2295552 112:02:00
```

`#N` is a starting byte offset in the `.bin`; the trailing MSF is the track's
*length*, and it lies: track 2's declared 104,428,800 bytes overrun track 3's
offset by two orders of magnitude — the audio track is truncated in the dump.
The reader (`tools/platform/dc/cue.go`, `disc.go`) therefore trusts only
modes and `#` offsets; lengths derive from the next track's offset and the
file size.

### The GD-ROM layout

A GD-ROM has two sessions: a low-density CD area (tracks 1-2 here) and the
high-density data area, which on every retail disc begins at absolute LBA
45000. The anchor is not taken from the TOC: each raw data sector carries its
own address as BCD minutes:seconds:frames in the 16-byte header, offset by
the 150-frame pregap, and track 3's first sector says 10:02:00 = LBA 45000.

| track | mode | file offset | bytes | absolute LBA |
|---|---|---|---|---|
| 1 | MODE1_RAW | 0 | 1,058,400 | 0–449 |
| 2 | AUDIO | 1,058,400 | 1,237,152 | — |
| 3 | MODE1_RAW | 2,295,552 | 1,185,760,800 | 45000–549149 |

### IP.BIN

The high-density area's first 16 sectors are the initial program. Sector 0
opens with the boot metadata, fixed-width space-padded text:

| field | value |
|---|---|
| hardware | `SEGA SEGAKATANA` |
| maker | `SEGA ENTERPRISES` |
| device | `EDB5 GD-ROM1/1` |
| area | ` U` |
| peripherals | `0799A10` |
| product | `MK-51035` `V1.004` |
| release | `19991219` |
| boot file | `1ST_READ.BIN` |
| title | `CRAZY TAXI` |

### The filesystem

The data area is ISO 9660 with the multi-session convention: the PVD sits at
session+16 (LBA 45016, volume `CRAZY_TAXI`), its volume-size field is
session-relative (504,150 blocks), but every directory extent is *absolute* —
the root lands at 45020. The disc therefore serves absolute LBAs as an
`iso9660.BlockSource` and the volume opens through `iso9660.OpenVolumeAt`.

51 files, flat in the root. The census in brief: `1ST_READ.BIN` (1,468,208
bytes, LBA 548283 — outer-edge placement, where the drive reads fastest),
`AICADRV.BIN` (the 57 KB sound-CPU driver), three of each per-city set —
`BINC*.AFS` archives, `COLDC*.BIN` collision, `LANDDC*.AFS`, `POLDC*.BIN`
geometry, `TEXDC*.BIN`/`TEXPVR*.BIN` textures — plus `MOTDC.BIN` motion
data, `SEQON*.BIN`, `VOICE.BIN`, `GDTEX.PVR` (the disc-icon texture), and
`CT2/` sequel-preview assets. `gdinfo -files` lists them all; `-at LBA`
names the file any raw read lands in.

## Part II — Boot chain and toolchain

### The bootstrap

`1ST_READ.BIN` loads at 8C010000 and its first instructions are a copy loop
(`dissh4 -base 8C010000`):

```
8C010000  D006  mov.l @(0x18,pc), r0      ; = AC010100
8C010002  D107  mov.l @(0x1C,pc), r1      ; = AC014000
8C010004  D207  mov.l @(0x1C,pc), r2      ; = AC004000
8C010006  6302  mov.l @r0, r3             ; copy loop
8C010008  2232  mov.l r3, @r2
8C01000A  7004  add #4, r0
8C01000C  7204  add #4, r2
8C01000E  3100  cmp/eq r0, r1
8C010010  8BF9  bf 0x8C010006
8C010012  D004  mov.l @(0x10,pc), r0      ; = AC004000
8C010014  402B  jmp @r0
8C010016  0009  nop
```

The loader relocates its second stage from 8C010100 down to 8C004000 through
the uncached P2 mirror and jumps there; the second stage then loads the main
program high (0C14xxxx and up), zeroes the vacated pages, and enters the
game. The literal pools between these functions are why the tracer treats
`mov.l @(disp,PC)` targets as first-class data (`codetracesh4` renders them
as `.word` and refuses to decode into them).

### What the machine model provides at boot

No BIOS image is used. `dc.Boot` reproduces the hand-over state the BIOS
leaves: the boot binary at 8C010000, SR `400000F0`, FPSCR `00040001`, stack
at 8C00F400, and in low RAM exactly six planted facts —

- the five syscall vectors (8C0000B0 sysinfo, B4 romfont, B8 flashrom,
  BC gdrom/misc, E0 menu), each a pointer to a trap PC the HLE services;
- the console ID at 8C000068;
- `rte; nop` at 8C000010: the BIOS's uncached return-from-exception stub.
  The game's own interrupt epilogue is the evidence — it restores the whole
  context including SSR/SPC, hand-builds the constant AC000010 (P2, so the
  `rte` executes uncached) and jumps there with nothing left to do but
  return;
- zeros at 8C00FFF0/FFF4: the game keeps a lazily-initialised work-structure
  pointer pair there and its "already initialised?" test is `tst` — a fresh
  boot must read zero.

Every other byte of the BIOS work area is poisoned `F00D` + offset, so a
dependency on an unplanted value surfaces as a self-naming register value,
never as an invented fact.

### The boot, observed

The oracle runs the retail boot for a billion instructions
(`bootoracle -steps 1000000000 -gd -v`):

1. the bootstrap relocates and the runtime initialises;
2. a serial driver probes the expansion socket at 03010000 (an empty G2
   socket reads all-ones and the probe gives up) and prints over the SCIF —
   the on-chip UART's transmit FIFO always drains, and every byte lands in
   the oracle's `serial` capture;
3. the runtime installs VBR at 8C00F400, hooks the VBlank through Holly's
   level-6 mask (INTEVT 0x320), and services fields at 60 Hz;
4. GD-ROM syscalls read the session lead-in and then stream `AICADRV.BIN`
   (28 sectors) into sound RAM; the Maple bus enumerates the port-A
   controller and block-reads for memory cards (answered "function
   unsupported"; the VMU slots answer as empty sockets);
5. 4,568 words of geometry go to the tile accelerator by ch2 DMA (counted,
   not rendered — the PVR is a later milestone);
6. the boot parks polling a handshake word in sound RAM at A080005C: it is
   waiting for the AICA's ARM7, which does not run yet. That poll is the
   frontier of this milestone, named by the machine itself.

Savestates round-trip the whole machine (a run continued past a snapshot and
a run restored from it stay byte-identical), and the unmodelled-hardware
census after a run is the worklist: the AICA block, the GD DMA registers,
the RTC words at 00710000, the PVR beyond its register file.

### The SH-4 toolchain

`tools/cpu/sh4` follows the repo's CPU-package shape: `decode.go` dispatches
the 16-bit halfwords the way the manual's tables group them (`decode_fpu.go`
holds group 1111), `exec.go`/`exec_fpu.go` interpret under the r4300-style
delay-slot pipeline with the SH-4 rules — interrupts are never accepted
between a delayed transfer and its slot, delayed conditionals run their slot
on both paths, R0-R7 swap banks on SR.RB writes, the FPU is two banks of raw
bits reinterpreted by FPSCR. The TMU, INTC, store queues and SCIF live in
the CPU package because P4 accesses never reach the external bus. Validation
is the assembled-loop suite in `cpu_test.go` plus the golden decode table
seeded with this game's own bootstrap; an env-gated per-instruction
conformance harness (`SH4_SST_DIR`) is deliberately deferred until the suite
files are in hand to pin the format against.
