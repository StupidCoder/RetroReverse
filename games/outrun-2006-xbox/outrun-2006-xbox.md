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
