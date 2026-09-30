# M7 — 3DO Code and a Need for Speed simulation iteration

The 3DO Code workspace provides live big-endian ARM60 disassembly, registers,
current Portfolio task, single-step and bounded address/function stops. Its game
preview and play controls use the shared workspace. The NFS knowledge package
contains the reference disc identity, eight guarded functions, seven decoded car
records and a 17-stop guided investigation. All game addresses remain in that
single package, not in the adapter.

## The investigation

Load the reference disc, enter the City race and pause at player segment 26.
The private fixture below provides a reproducible route from normal boot to this
point. Start checks the disc hash, original code signatures, car self pointers,
list links and the dormant driver's countdown-record pointer. Other allocations
or game phases fail explicitly; raw debugging remains available.

The tour follows world-update entry `$17904`, each dispatch at `$17A60`, the
selected driver entries and the world-update return `$17EB4`. At its first stop
it binds the current task and big-endian simulation counter at `$41D24`. Every
subsequent stop must belong to that same task and iteration. A changed counter
fails the tour instead of presenting the explanation for another iteration.

The reference run visits pool indices **3, 0, 2, 1, 4, 6, 5**:

| Pool indices | Observed dispatch |
| --- | --- |
| 3, 0, 2, 1 | Dormant traffic; original countdown and spawn gates |
| 4 | Cop behavior flags, but activity word 3 skips driver execution |
| 6 | Player driver `$F8D8` |
| 5 | Opponent using shared road AI `$284E0` |

For the first dormant entry, the tour pauses at the spacing branch `$2A540`:
`R3 = 25 + 2 × ordinal = 31`, while `R1 = player segment = 26`. The original
`BGT` takes the rejection path to `$2A56C`, which rearms the countdown. The later
density/activation gates are not reached in this case. This demonstrates an
actual failed spawn attempt; it does not claim that a traffic car spawned.
Model and texture identifiers remain raw IDs with unknown names.

The recorded run starts its verified iteration at scheduler step **61061349**
and reaches the return at **61069815**. All 17 stops have simulation counter
**1476**, task **4218**, and display **1250**. The task differs from the old
research write-up; it is discovered at runtime rather than hard-coded.
[acceptance.json](acceptance.json) records each stop and its interval evidence.
The tour reaches the function's return instruction before executing it.

## Adapter and evidence contract

- The existing ARM decoder reads current backing bytes on every snapshot.
  DRAM, VRAM and simulated kernel item memory can be inspected without device
  reads. Unavailable windows and virtual Portfolio entries are labelled.
- This is the core's ARM32 model. Thumb, ARM26 and hardware-accurate kernel
  execution are not advertised. Virtual Portfolio calls are HLE boundaries;
  SWIs retire one guest instruction and report their atomic HLE service.
  Movie HLE cannot be instruction-stepped. Task and display transitions without
  a retired instruction are distinct step results.
- Scheduler units, retired guest instructions, display count, field count and
  the game's simulation counter are separate values. A video frame is not a
  simulation-loop identifier.
- Named function stops require the live layout guards and bind the selected
  task. Raw address stops remain available without game knowledge.
- Tour recording selects only its car-pool and counter ranges. DRAM/VRAM bus
  writes carry scheduler timestamps; endpoint bytes are captured atomically
  with registers and decoded state. Same-value writes are retained. Direct HLE
  bulk stores are **not** bus-write events; endpoint differences still expose
  their lasting changes. Evidence completeness applies to this stated scope.
- Snapshot/state navigation is read-only. Explore and Restore stop retain the
  shared tour behavior. Checkpoints preserve whether the scheduler prelude has
  already run, avoiding a duplicate field/input tick after a breakpoint.
  Native state version 3 reads versions 1/2; browser containers still require
  the exact core hash. No commercial image or checkpoint is distributed.
- Layout guards deliberately restrict this package to the verified City pool.
  Manual preparation may produce different counters/tasks, but the documented
  first spacing predicate must still match or the explanation is rejected.

## Package additions

Schema version 1 gains two optional, bounded, data-only fields. Existing packages
are unchanged. Go validation rejects these capabilities on other platforms.

```json
{
  "liveGuards": [{"address": 96516, "bytes": [225, 160, 192, 13]}],
  "tours": {
    "one-iteration": {"iteration": {"counter": 269604}}
  }
}
```

This fragment illustrates the field shapes, not a complete package.
`liveGuards` checks direct DRAM/VRAM bytes before resolving named functions or
state. `iteration.counter` identifies a four-byte big-endian game counter; the
first verified stop captures its value and current task. ARM tour predicates
accept `mode: "arm32"` and registers `R0` through `R15` or `CPSR`. CPU PC targets
must be instruction-aligned; captures and signatures remain bounded by the 3 MiB
DRAM/VRAM view. No task ID, callbacks or executable code are stored in a package.

## Reproduce

Run from the repository root. Supply the exact locally owned disc named below.
The fixture executes 1250 normal display frames and recorded controller input;
it does not patch game state or inject cars.

```sh
python3 tools/platform/threedo/browser/build.py --emcc /path/to/em++
clang++ -std=c++20 -O1 -fwrapv -ffp-contract=off \
  -Wno-address-of-temporary -Wno-parentheses-equality \
  tools/browser/tests/nfs-fixture.cpp -o /tmp/nfs-fixture
/tmp/nfs-fixture 'games/need-for-speed-3do/Need for Speed.bin' \
  tools/browser/tests/nfs-race-inputs.txt /tmp/nfs-race.state
python3 tools/browser/package.py --core 3do
node tools/browser/tests/tour-nfs.mjs \
  'games/need-for-speed-3do/Need for Speed.bin' /tmp/nfs-race.state \
  tools/platform/threedo/browser/work/m7-tour.rrstate
python3 -m http.server 8877 --bind 127.0.0.1
```

Open `/tools/browser/tests/code-m7-browser.html` on that local server. The fixture
loads the private image and checkpoint, checks function navigation is read-only,
steps/restores an ARM instruction, follows all 17 stops in the same task/iteration,
checks Memory navigation, and exercises a 375px panel and keyboard focus. It
reports PASS/FAIL in the page. The large disc is loaded only from localhost.

`python3 tools/browser/check.py` includes media-free native ARM tests, 3DO
knowledge/iteration tests, schema rejection tests and the shared C64/DOS
regressions. Native tests also cover virtual HLE versus SWI steps, task-bound
stops, selected VRAM writes, idempotent recording cleanup and checkpoint prelude
restoration. Normal playback was compared with the pre-M7 native core over the
fixture's 1250-frame preparation: all 13 sampled CPU/DRAM/VRAM/item-memory/pixel
proofs and instruction/display counts matched. The state format gained one byte
for the pending-prelude flag.

This milestone does not implement game edits, patched replay experiments or a
successful traffic-spawn lesson. Those require their own verified scenarios.
