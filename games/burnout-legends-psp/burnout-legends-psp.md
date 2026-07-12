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
- **Part III — To the main menu.** The directory-scan file catalogue
  (`sceIoDread`), the by-value thread argument, the movie player
  (sceMpeg/sceAtrac3plus HLE), and the scripted walk from the title screen
  through profile creation to the main menu.

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

---

## Part III — To the main menu

### 1. The directory-scan file catalogue

Past the loading screen the boot walked into a corrupt object and `jalr`'d
into the exception vector. The trail (`-bp`/`-watch` on the pointer the
crashing code loads, at `0x08C2C18C`) led back through the boot state machine
at `0x0880E0xx`: the pointer is the loaded image of `Data/PrgData.bin`, a
pointer-patched data file the game relocates in place (`0x089BB850` adds the
load base to a table of offsets — the same trick LocoRoco's `.clv` levels
use). The file "loaded" — but every one of the game's reads was **zero bytes
long**: `sceIoRead(..., 0) -> 0`.

The zero comes from the game's file catalogue. Burnout Legends does not ask
for file sizes with `sceIoGetstat` or `sceIoLseek(END)`; at boot it walks the
whole disc with **`sceIoDopen`/`sceIoDread`/`sceIoDclose`** (124 directories)
and builds its file table — names *and sizes* — from the returned
`SceIoDirent` entries. With those calls stubbed, every catalogued size was
zero. The oracle now serves the scan from the ISO directory tree (`io.go`):
each `sceIoDread` fills a dirent (a `SceIoStat` with the umd9660 driver's
start-LBN in `st_private[0]`, plus the name), returning the number of entries
still to read. After the fix the game streams every asset by raw sector
extent — `disc0:/sce_lbn0x%X_size0x%X` paths, the same umd9660 contract the
other disc uses — with correct sizes, and `PrgData.bin` relocates correctly.

### 2. The by-value thread argument

Next wall: the game's "SND ATRAC PACKET DECODER" thread dereferenced its
argument into garbage and crashed. The thread is started with a **112-byte
argument block** (`sceKernelStartThread(uid, 0x70, ptr)`) that lives on the
*creator's* stack. The real kernel copies the block onto the new thread's
stack before it first runs; the oracle's scheduler passed the original
pointer, and by the time the thread was scheduled the creator's frame was
long dead. `startThread` (`sched.go`) now copies the block below the thread's
`$k0` context area and points `$a1` at the copy — the kernel contract.

### 3. The movie player: sceMpeg + sceAtrac3plus

The boot then reached the intro movies (`ovid/englis30.pmf` + its `.at3`
audio) and parked: the game pumps its player loop off `sceMpegGetAvcAu`, and
the ATRAC packet-decoder thread spun millions of calls into stubbed
`sceAtracDecodeData`. The oracle now carries a **minimal, honest movie-player
HLE** (`mpeg.go`) — no video or audio codec, but the real streaming contract:

- **PSMF header** (big-endian): `sceMpegQueryStreamOffset`/`QueryStreamSize`
  parse the magic and the offset/size fields of the header the game hands in,
  and reject a buffer without the `PSMF` magic.
- **Ringbuffer accounting**: `sceMpegRingbufferConstruct` records (and writes
  into the guest struct) the packet capacity and the game's own packet-read
  callback; `sceMpegRingbufferPut` *runs that callback* in a nested guest
  frame (`callGuest`), so the movie data really is streamed by the game's
  file manager; `sceMpegGetAvcAu` consumes buffered packets into access units
  and reports `SCE_MPEG_ERROR_NO_DATA` when the stream drains — the signal
  the player's end-of-movie logic runs on.
- **Frames without pixels**: `sceMpegAvcDecodeYCbCr`/`sceMpegAvcCsc` report
  every frame produced but write no pixels (there is no H.264 decoder here) —
  a movie "plays" black, at the pace of the game's own pump, and terminates.
- **ATRAC3+ as silence with real accounting** (`sceAtracSetDataAndGetID`,
  `DecodeData`, `GetStreamDataInfo`, …): the RIFF header the game hands over
  names the block align and data size, so the decode loop serves the true
  number of frames — as silent PCM — sets the end flag on the last one, and
  returns `SCE_ATRAC_ERROR_ALL_DATA_DECODED` past it.

One more CPU op surfaced on the way: the sound mixer converts samples with
the VFPU's packed-integer conversions — `vi2s.q` and family (`vi2s`/`vi2us`/
`vi2c`/`vi2uc`, `vs2i`/`vus2i`) are now in `tools/cpu/allegrex/vfpu.go`.

### 4. Title screen to main menu

With movies completing, the attract sequence lands on the game's title screen
— its own rendered "PRESS START BUTTON TO CONTINUE" frame:

![The title screen](figures/title.png)

From there a `-keys` pad script (VBlank-scheduled, the same mechanism the
other PSP disc plays with) walks the front end: START → the profile dialog →
NEW PROFILE → the on-screen keyboard's default name → save-to-memory-stick
declined (the modelled savedata utility reports no memory stick data) →
autosave-off confirmed → the **main menu**: WORLD TOUR, SINGLE EVENT,
MULTIPLAYER, DRIVER DETAILS.

![The main menu](figures/menu.png)

### 5. Into the race: the idle-descriptor poll

Continuing the script — SINGLE EVENT → RACE → the race options (region, track,
rivals) → **SELECT CAR**, whose 3-D car model the software rasterizer renders
— the boot then deadlocked on the "Filesystem lock" semaphore: `user_main`
took it, and took it again without releasing.

The cause was a wrong answer from the oracle, not a missing one.
Burnout's stream code **polls a descriptor before it queues work on it**, to
ask "is an operation already running?" The oracle's `sceIoPollAsync` returned
`1` — *still in progress* — for a descriptor with nothing outstanding. The
game therefore believed a read it had never issued was in flight, waited for
it, and `sceIoWaitAsync` handed back a stale zero, which failed the game's
check against the byte count it expected; the error path then tore the stream
down and re-entered the lock-holding close. One value, four symptoms.

Both calls now report `SCE_KERNEL_ERROR_NOASYNC` (`0x80020321`) when no
operation is outstanding, which is the honest answer: *nothing is running
here*. The game then issues its reads, and the race load runs — `enviro.dat`,
`Gamedata.bgd`, `static.dat`, the track texture packs — onto the pre-race
loading screen, a full-colour rendered frame:

![The pre-race loading screen](figures/loading-tip.png)

Two diagnostics came out of the hunt and stay in the platform:
`PSP_SEMA_TRACE=<name>` logs every take and release of a named semaphore with
the thread, count and PC — the tool for a semaphore deadlock — and
`PSP_SYSCALL_TRACE=<substring>` logs matching syscalls with their arguments,
return value and caller.

**The current wall.** On the loading screen the game opens the track's
`streamed.dat` (3.9 MB), then polls that descriptor every frame forever — the
read is never issued. Its stream object's `Read` method (vtable + 24,
`0x08A44280`, which computes a length from the object's 64-bit position/limit
pair and calls `sceIoReadAsync`) is never called; the frame loop only calls
the sync/poll method (vtable + 48, `0x08A449F8`). Nothing is blocked — no
semaphore is held, no thread is waiting — so the game is waiting on a
completion its own producer never starts. The next step is to find what
should drive that producer.
