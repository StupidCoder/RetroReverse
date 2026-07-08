# Headless FS-UAE with GDB remote debug stub (macOS arm64)

A debugger-controllable Amiga emulator — shared reverse-engineering
infrastructure for any Amiga title in this repo, not specific to one game. It
gives full run control (boot, breakpoints, continue, single-step, register/
memory read) over a real Kickstart, so the *irreducible runtime state* a program
depends on — values set by booted AmigaDOS that are absent from the disk image —
can be captured live and fed to a static Go reimplementation, which keeps the
*algorithm* in Go and uses the emulator only to supply (and verify against) the
few bytes of OS state.

Its first use was Marble Madness (`Marble Madness (Amiga)/`): the `c/zzz`
copy-protection (`sub_DAA`) folds the CPU exception/TRAP vector page-bytes and
the launcher task's `tc_ExceptCode`/`tc_TrapCode` handler pages into its
decryption keystream; those ~25 bytes were captured here and baked into the Go
decoder (`Marble Madness (Amiga)/extract/cmd/decode`). The generic pieces live
here; per-game capture scripts (e.g. Marble Madness's `modadf.go`, `capdat.py`,
`capkey.py`) live in that game's own `tools/fsuae-debug/` and import the shared
`dbg.py` from here.

Upstream: **prb28/fs-uae** — an FS-UAE fork (core derived from WinUAE 3300b2)
that adds a GDB remote-serial-protocol server (`src/remote_debug/`). Built at
commit `58feb4ccee1d0fa162ab25ef57a4529d433a5b3e`.

The patched + built working tree lives at **`~/Development/fs-uae`** (sibling of
this repo, kept out of git — it's 144 MB of third-party source + build
artifacts, and the ROM/ADF are copyrighted). The `fsuae-arm64.patch` here is the
canonical record of our changes; the build recipe below reproduces it from
scratch if that tree is ever lost.

## Build (macOS arm64 / Apple Silicon)

```sh
brew install pkg-config glib sdl2 libpng gettext libmpeg2
git clone https://github.com/prb28/fs-uae /tmp/fs-uae
cd /tmp/fs-uae
git checkout 58feb4c
git apply "<repo>/tools/amiga/fsuae-debug/fsuae-arm64.patch"

export PATH="/opt/homebrew/bin:$PATH"
export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig"
export CPPFLAGS="-I/opt/homebrew/include"
export LDFLAGS="-L/opt/homebrew/lib"

./bootstrap            # autoreconf
./configure
# configure/Makefile emit a `-pagezero_size 0x2000` link flag that produces a
# malformed arm64 Mach-O (genblitter then SIGKILLs). Strip it:
sed -i '' 's/-pagezero_size 0x2000//g' Makefile config.status configure

make -j6 \
  CFLAGS="-g -O2 -Wno-narrowing -Wno-implicit-function-declaration" \
  CXXFLAGS="-g -O2 -Wno-c++11-narrowing -Wno-narrowing -Wno-implicit-function-declaration"
```

### What the arm64 port needed (see `fsuae-arm64.patch`)

| Symptom | Cause | Fix |
|---|---|---|
| `sysdeps.h: "unrecognized CPU type"` | no aarch64 branch | add `__aarch64__`/`_M_ARM64` case (`CPU_aarch64`, `CPU_64_BIT`) |
| genblitter killed; "Malformed Mach-o file" | `-pagezero_size 0x2000` invalid on arm64 | strip the flag (sed above) |
| `blkdev.cpp` C++11 narrowing errors | clang strict initializer-list narrowing | `-Wno-c++11-narrowing -Wno-narrowing` |
| `actions.c` implicit-decl error | `render.h` not included (symbol *does* exist) | `-Wno-implicit-function-declaration` |
| link: undefined `uae_ppc_wakeup_main()` | two calls sit outside the `#ifdef WITH_PPC` guard | guard them (patch) |

### The continue-freeze fix (the important one) — `remote_debug.cpp`

Out of the box, a bare GDB `c` (continue) **froze the whole emulated chipset**: the
video beam counter stopped, no interrupts fired, and any `STOP`-based OS idle loop
deadlocked. Root cause: the stub's `Tracing` service loop in `remote_debug_()`
spins on `sleep_millis(1)` and only breaks on `step_cpu` or quit — it never checks
`s_state`. `handle_continue_exec → remote_deactivate_debugger()` sets
`s_state = Running` but does **not** break that loop, so after `c` the CPU thread
stays trapped in C code, never calling `x_do_cycles()`. The 68000 `STOP` idle loop
(`newcpu.cpp` ~3870) only advances the chipset via `x_do_cycles()` *inside*
`while (SPCFLAG_STOP && !SPCFLAG_BRK)`, so with the CPU thread parked, the chipset
never ticks, no VBlank/CIA interrupt is generated, and `STOP` never wakes —
total deadlock. Single-step worked only because it sets `step_cpu`, which *does*
break the loop.

Fix: break the `Tracing` loop when a continue switches the state back to `Running`:

```c
if (s_state == Running) {   // continue requested -> resume the CPU/chipset
    break;
}
```

With this, `c` resumes the emulation thread and the chipset runs again
(verified: chip RAM changes across a continue).

### Two more stability fixes — `remote_debug.cpp`

**Speed.** `remote_debug_init_()` enables `debugmem_trace = true` and
`debugmem_enable_stackframe(true)` ("Full stack frame tracking enabled"), which
hook every call/return and slow emulation badly. Not needed for
memory/register/breakpoint capture — disabled (`false`/`false`) so emulation
runs at a usable speed.

**Crash on disconnect/shutdown (SIGSEGV).** `remote_debug_update_()` (called
every vsync) checks `s_conn` once at the top, then calls `remote_debug_()` —
which on the `is_quiting()` path does `rconn_destroy(s_conn); s_conn = 0;`.
Control returns and the function then calls `rconn_poll_read(s_conn)` →
`rconn_connected(NULL)` → dereferences `0x8` → `EXC_BAD_ACCESS`. This fired on
every emulator quit/disconnect (visible as macOS "fs-uae crashed" reports).
Fix: re-check `if (!s_conn) return;` after `remote_debug_()`. Verified: clean
connect/continue/disconnect now produces no crash report.

### Speed: enable the CPU-context debug hook — `remote_debug.cpp`

`remote_debug_init_()` left `debugging = 0`, so `do_specialties()` never called
`debug()` → `remote_debug()`; the whole debugger ran only off the per-vsync hook
(~50 Hz), so software breakpoints were only *sampled* 50×/sec and effectively
never hit. Setting `debugging = 1` enables the per-instruction CPU-context hook
so breakpoints are checked every instruction. (NB: a bp keeps `SPCFLAG_BRK` set,
so this also needs the SOCKET_POLL_INTERVAL throttle above to avoid a `select()`
per instruction.)

### The continue/resume bug — `handle_continue_exec` (the big one)

Symptom: after *any* halt (manual `0x03` or a breakpoint hit), `c` (continue)
advanced the CPU zero instructions — the emulator was frozen. Single-step
(`s`) worked. Found by instrumenting the whole path (`m68k_go` → `m68k_run_1_ce`
→ `do_specialties` → `debug()` → `remote_debug_()`'s Tracing spin, plus
`parse_packet`/`handle_packet`): the `c` packet was *received and dispatched*
to `case 'c'`, but `handle_continue_exec` returned early **before** calling
`remote_deactivate_debugger()`, so `s_state` stayed `Tracing` and the spin loop
never broke.

Root cause: `handle_packet()` runs `remove_checksum()`, which null-terminates
the packet at `#`. So for a bare `c`, the argument handed to
`handle_continue_exec` (`packet + 1`) points at `'\0'`, **not** `'#'`. The guard
`if ((packet != NULL) && (*packet != '#'))` therefore treated the empty string
as an address argument, `sscanf("%x#", …)` failed, and it `return false`d
without resuming. (Step worked because `step` parses no argument.)

Fix: only parse an address when one is actually present —
`*packet != '#' && *packet != '\0'` — and also accept a trailing-`#`-less hex
address. Verified: breakpoints now **chain** (15/15 continue→re-hit on a 50 Hz
handler) and manual interrupt→continue→interrupt no longer freezes.

### The breakpoint re-arm "max breakpoints reached" bug — `add_breakpoint`

Symptom (first read as a breakpoint *leak*): across connect/disconnect cycles, a
`Z0` to re-arm a breakpoint already in the table would intermittently reply
`E37` (`ERROR_MAX_BREAKPOINTS_REACHED`), as if the table had filled up — yet a
count probe showed `s_breakpoint_count` never grew past 1. The contradiction
(per-add logs say count≈1, but E37 means count≥65535) was the clue: **there is
no leak**; the error came from a return-value mismatch.

`add_breakpoint()` returned two differently-based values — the **0-based** index
from `find_breakpoint()` when the breakpoint already existed, but the
**post-increment** `s_breakpoint_count` (1-based) when adding a new one. The
callers (`set_absolute_address_breakpoint`, `set_offset_seg_breakpoint`) tested
the result as a plain boolean. So whenever you re-armed the breakpoint that
happened to occupy **index 0**, `find_breakpoint` returned `0`, `add_breakpoint`
returned `0`, `if (add_breakpoint(...))` read that as failure, and the stub sent
`E37`. The table was fine; the *first* breakpoint just couldn't be re-armed. A
second latent bug: `set_offset_seg_breakpoint` resolved
`&s_breakpoints[s_breakpoint_count - 1]` (the last slot) instead of the slot the
add actually returned — wrong target whenever a seg breakpoint was found at any
index other than the last.

Fix: make `add_breakpoint` return a **consistent 0-based index** (or `-1` when
full), and have callers test `>= 0` and use the returned index. Verified: 60
re-arm/remove cycles and 2000 mixed re-add ops → **0 errors**, table stays
bounded (re-adds dedup correctly, since the static table legitimately persists
across reconnects). This pairs with the earlier `MAX_BREAKPOINT_COUNT` 512→65536
raise and the `remove_breakpoint`/`reset_breakpoint` index-0 guard fix
(`> 0` → `>= 0`, which had made breakpoint index 0 unremovable).

## What works (macOS arm64, this fork)

- Boots Kickstart 1.2 + the Marble Madness floppy to **Workbench** under the
  debugger.
- Connect, halt (raw `0x03`), read/write registers and memory, single-step.
- **Software breakpoints** fire during free-run (async `%Stop;swbreak`) and
  **`continue` resumes correctly** — full breakpoint chaining and step-through.

(NB: KS1.2 exception/TRAP vectors `$8–$BC` legitimately stay all-`$FC` (ROM)
even when fully booted — AmigaDOS doesn't redirect CPU vectors — so "0 vectors
redirected" is the normal booted state, not a sign of a failed boot.)

The patch only touches `src/include/sysdeps.h` and `src/cpuboard.cpp`; the
narrowing/implicit-decl issues are handled by the `make` flags, and the
pagezero strip is a `sed` on generated files. The GDB stub is compiled in
unconditionally (`REMOTE_DEBUGGER` is `#define`d in `src/remote_debug.h`).

## Run (debug server on tcp/6860)

FS-UAE requires a real OpenGL context (`SDL_VIDEODRIVER=dummy` fails with
`[GLAD] Failed to initialize OpenGL context`), so run it with a real window on
a logged-in GUI session.

```sh
./fs-uae \
  --kickstart_file=/path/to/kick13.rom \
  --floppy_drive_0="/path/to/Marble_Madness.adf" \
  --amiga_model=A500 \
  --remote_debugger=10        `# wait up to 9s for a debugger to connect` \
  --remote_debugger_port=6860 \
  --fullscreen=0
```

`remote_debugger=N` waits `N-1` seconds at startup for a connection; the server
keeps listening afterwards. The ROM and ADF are copyrighted and not committed.

## Drive it (`rsp.py` / `dbg.py`)

Two GDB remote-serial-protocol clients ship here, both stdlib-only and
game-agnostic:

- **`rsp.py`** — minimal client. Connects, interrupts (raw `0x03`), reads
  registers (`g` → 8×D, 8×A, SR, PC as u32), reads memory (`m<addr>,<len>` →
  hex), sets breakpoints (`Z0,addr,kind`). Its example session dumps the
  exception vector table `$0..$BF`.
- **`dbg.py`** — the fuller client used by the capture scripts: adds
  `cont_wait`/`wait_stop` (read until a `T`/`S`/`W`/`X`/`%Stop` reply),
  `stop_regs` parsing, and Amiga-exec helpers (`sysbase`, `thistask`,
  `liblist`, `find_lib`, `amiga_str`). Per-game scripts `import dbg` from here.

```sh
python3 rsp.py
```
