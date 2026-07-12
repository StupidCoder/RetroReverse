# Burnout Legends (PSP) — technical reference

**Image.** `Burnout Legends.cso` — 260,157,913 bytes, MD5
`eaa446ea6d4847bdf486eb114441ddfd`. A CISO-compressed dump of a UMD
(`DISC_ID` ULUS10025, `PSP_SYSTEM_VER` 1.52). The image is not committed
(see the repository copyright policy); supply a dump with the MD5 above.

Burnout Legends is a 3-D arcade racer; it shares the PSP toolchain with the
other PSP target — the Allegrex CPU core (`tools/cpu/allegrex`), the machine
oracle and its format libraries (`tools/platform/psp`), and the
`pspinfo`/`bootoracle` front-ends. This document records what is specific to
this disc.

## Contents

- **Part I — The image.** The CISO/ISO layers, PARAM.SFO, and the `~PSP`
  KIRK-encrypted executable — a 1.xx-firmware EBOOT with a distinct
  decryption tag.
- **Part II — Boot chain.** Module relocation (multi-segment, segment-index
  aware), the async IO the file manager runs on, and the boot to the first
  rendered frame.

---

## Part I — The image

### 1. Container and filesystem

The CISO container and the ISO 9660 UMD filesystem decode with the shared
readers (`cso.go`, `iso.go`); no format differences from the other PSP disc.
PARAM.SFO reports `TITLE` Burnout Legends, `DISC_ID` ULUS10025, `CATEGORY`
UG, `PSP_SYSTEM_VER` 1.52. The boot tree holds the encrypted
`PSP_GAME/SYSDIR/EBOOT.BIN` (5,075,472 bytes; `BOOT.BIN` is blanked), a
`SYSDIR/UPDATE` firmware-updater payload, and `USRDIR` — the game's
`data/*.txd` texture dictionaries (`Frontend.txd`, `Global.txd`, …),
`*.kfs`/`*.bin` asset packs, per-language `Global*.bin`, and the media PRXs
under `amodule/` (`libatrac3plus`, `mpeg`, `sc_sascore`, the `pspnet`
adhoc-multiplayer stack).

### 2. The boot executable: `~PSP` / KIRK, tag `0x08000000`

`EBOOT.BIN` is a `~PSP` container whose header tag at `+0xD0` is
**`0x08000000`** — the 1.xx-firmware retail EBOOT tag, distinct from the
2.xx tag `0xC0CB167C` the other disc uses. The KIRK decryption algorithm
(`prx.go`/`kirk.go`) is unchanged; the tag selects a different XOR seed and
kirk7 key: `g_keyEBOOT1xx` (0x90 bytes) with key id `0x4B`. The seed is a
documented platform constant (`kirk_keys.go`), transcribed as the
little-endian bytes of the reference's u32 array and verified against ground
truth: the SHA-1 header check holds and the body decrypts to `\x7fELF`,
exactly `0x4D70BD` = 5,075,133 bytes.

The plaintext is an ELF32-LE MIPS PRX (`e_type` `0xFFA0`) with **two**
`PT_LOAD` segments: segment 0 (file = mem = `0x3B9B7C`, the code and
read-only data, `p_paddr` `0x29B6C4` locating the `sceModuleInfo`) and
segment 1 (file `0x1730`, mem `0x31A270` — a small initialized-data block
over a large BSS). The module names itself `Burnout` and imports **31
libraries**: `sceGe_user`, `sceDisplay`, `sceCtrl`, `sceMpeg`, `sceSasCore`,
`sceAtrac3plus`, `sceNet`/`sceNetAdhoc*`, `sceUtility`, `sceRtc`,
`ThreadManForUser`, and the rest.

`pspinfo -image "…/Burnout Legends.cso" -exe PSP_GAME/SYSDIR/EBOOT.BIN`
decrypts and describes it.

---

## Part II — Boot chain

### 1. Multi-segment PRX relocation

The module is relocatable and is loaded and relocated at `0x08804000`
(`elf.go`, `Relocate`). Its `SHT_PRXRELOC` sections (type `0x700000A0`) carry
144,831 MIPS relocations, and — unlike a single-segment module — their
`r_info` fields matter beyond the type byte: bits 8-15 name the **segment the
offset is relative to** and bits 16-23 the **segment whose load address is
added**. The same offset appears twice in this module — once as a segment-0
`HI16` (the first instruction, `lui $a0, 0x2B`) and once as a segment-1
`R_MIPS_32` — so applying every relocation to segment 0 (correct for a
single-segment module) overwrites segment-0 code with a segment-1 fixup and
corrupts the image at load. The relocator honours both indices: it reads and
writes within the segment named by the offset-base index and adds
`base + segment[addr-base].vaddr`, so segment-1 data fixups and segment-1
pointers land correctly and segment-0 code is left intact. `HI16`
relocations defer, carrying their segment, until the paired `LO16` resolves
the split immediate.

### 2. Async file IO

Burnout drives its assets through the PSP's asynchronous IO: a file-manager
thread issues `sceIoOpenAsync`/`sceIoReadAsync` and retrieves the results
with `sceIoWaitAsync`/`sceIoWaitAsyncCB`/`sceIoPollAsync`, guarded by a
"Filesystem lock" semaphore. The oracle's volume reads complete instantly,
so each async call performs its operation immediately and stores the 64-bit
result on the descriptor; the wait and poll calls hand that result back and
report completion (`io.go`). Without this the boot parks forever on the
filesystem lock; with it the game streams `data/Global.txd` and its other
packs and proceeds.

### 3. To the first frame

With the relocation and async IO correct, the boot brings up the game's
threads — `user_main`, `Callback_Handler`, `SystemControl` — and its kernel
objects (the system-control, display-list, UMD and filesystem semaphores,
the `SceGuSignal` event flag), loads and starts its **14 media modules**
(`sceKernelLoadModule`/`StartModule`), reads the pad
(`sceCtrlReadBufferPositive`), waits on VBlank (`sceDisplayWaitVblankCB`),
and submits GE display lists (`sceGeListEnQueue`) that the software
rasterizer draws — reaching the game's **LOADING screen**, its own rendered
frame.

![The Burnout Legends loading screen](figures/loading.png)

Beyond it the boot walks into an uninitialized global (a game subsystem the
current HLE surface does not yet bring up) — the next step of the boot to
reverse.
