# M4 — Read-only tours and interval evidence

Implemented in the shared C64 Code workspace. The first package-owned tour,
**How Fort expands its terrain**, follows two invocations of the original RLE
routine. It makes no claim about the unresolved enemy-helicopter bug.

## Using the tour

Load the exact Fort reference tape and boot normally. At the title screen, pause,
inspect `$8CDB` in Code and choose **Run to address**. Click the Code preview and
press Fire while that bounded debugger job is running. Release controls after the
breakpoint. The required first terrain entry has source `$7000`, destination
`$0503`, and mode `$21 = 0`. A later decompressor invocation fails the start check.
Alternatively load your own saved state at that boundary. No copyrighted game
checkpoint is shipped. Booting or approaching the entry is currently manual;
the private acceptance recipe below provides reproducible test preparation.

Choose the tour and **Start**, then **Continue** through the two runs. Each stop
shows an explanation, disassembly and shared state snapshot. **Suggested layout**
selects the package's code address and game-state visibility; it never executes.
Code/Memory navigation and read-only refresh leave the tour intact. **Explore
freely**, instruction stepping, playback, Render capture, recording, tape control
or meaningful input invalidate continuation. **Restore stop** explicitly restores
the last verified machine checkpoint; **Continue** is then available again.
Cancelling leaves the machine where it stopped and retains the preceding anchor.
Final stops remain restorable. Native buttons/selects support normal Tab,
Shift-Tab, Enter/Space and arrow-key interaction; controls wrap on narrow screens.

## Implemented package profile

`tours` is no longer reserved. The schema and Go semantic validator accept:

- `title`, `description`, supported `releases`, evidence references,
  `startPolicy: "current-paused"`, human-readable `requirements`, `start`,
  mapped-RAM byte `guards`, and an ordered nonempty `stops` list.
- Conditions: `{ "pc": 36059 }`, `{ "address": 33, "value": 0 }`, optional
  byte `mask`, or nested `all` / `any`. Memory conditions read **physical C64
  RAM**, excluding CPU ports 0/1; PC conditions require a RAM fetch boundary
  without a pending interrupt. These numeric addresses are intentionally a
  small runtime profile, independent of future symbolic/relocated selectors.
- Stops: unique `id`, `title`, `explanation`, `until`, `assertions`, `hitCount`,
  `cycleBudget`, `wallBudget`, physical-RAM `capture` ranges, `eventBudget`,
  and `layout: {address, gameState}`. Assertions must all pass before presenting
  a stop as verified. The next interval skips the current predicate occurrence.
- At most 64 tours/package and 64 stops/tour; condition depth 8 and 64 nodes;
  32 guards of at most 256 bytes; 16 nonoverlapping capture ranges totalling
  16 KiB; 29,557,440 cycles and 30 seconds per stop; 524,288 recorded events;
  one checkpoint of at most 4 MiB. Schema rejects unknown fields and scripts.

Only exact single-image C64 matches activate tours. Code guards are checked at
start and every verified stop. The core's target mapping check aborts a run if a
target changes mapping. The current firmware/core are inherited from the loaded
machine and the checkpoint is local to that worker; this profile does not import
portable tour checkpoints or claim arbitrary firmware equivalence.

Named function/state/slot condition selectors, branches, automated boot recipes,
portable checkpoint starts, continuous signature watching, non-C64 adapters and
patch experiments remain future profile extensions. The verified boot/input
schedule from `lesson_native.cpp` is reused **for acceptance**, not exposed as an
untested generic script engine. Elite's loader remains gated on verified research.

## Lifecycle and execution ownership

`idle → preparing → running-to-stop → paused-at-stop → continuing`, ending in
`completed`, `cancelled` or `failed`; exploration produces `diverged`. Requests
carry session generation, protocol version and monotonically increasing request
IDs. A replacement image gets a new worker and cannot inherit an old tour.

One tour job owns execution and the recorder. Play, debugger execution, capture,
save/seek and Memory recording cannot interleave with it. Cooperative slices are
at most 1,000 cycles; cancellation is checked between slices. The PC fast path
retains its armed native debugger job across slices. Compound predicates with no
mandatory PC use instruction boundaries. Runs obey both wall and cycle budgets;
missing targets stop with `budget`, not a successful explanation.

Starting/continuing releases host-held keys/joystick and clears queued input.
No tour changes game RAM. A successful stop saves full machine state; Continue
compares the serialized machine against that anchor before running. Read-only
observation does not invalidate it. Restore clears pending host input, restores
the complete machine and rechecks signatures. These anchors are private worker
memory, not new save-state files or downloadable game assets.

## Interval evidence

Capture starts before execution of the incoming interval. Each selected range
has an endpoint baseline. Independently, the existing core recorder collects
RAM writes, including same-value writes and write-then-restore. The UI separately
reports endpoint byte differences, actual write counts, and per-address old/new
values and counts. Pointer bookkeeping writes are included if their range was
selected; the Fort package captures both pointers and 256 terrain bytes.

The recorder has a global event limit; a tour may impose a smaller limit. A slice
can overshoot that requested threshold by at most its bounded work. Hitting the
limit aborts with `trace-overflow` and marks evidence **INCOMPLETE**. Dropped events
are explicit. At most 2,048 event details are retained; that display truncation
is distinct from incomplete capture (per-address counts still include all
recorded events). Endpoint differences remain exact for the observed endpoints,
but do not prove which addresses were written between them. M4 records writes,
not a complete instruction/read trace, atlas timeline or off-screen render log.

## Validation

Public, game-free checks:

```sh
go test ./tools/knowledge ./tools/cmd/knowledgecheck ./tools/cmd/knowledgeexport
node tools/browser/tests/tours.mjs
node tools/browser/tests/debug.mjs
node tools/browser/tests/structured-state.mjs
node tools/browser/tests/knowledge.mjs
node tools/browser/tests/memory.mjs
node tools/browser/tests/ui-shell.mjs
python3 tools/browser/check-release.py
```

The real WASM synthetic program writes 1 then restores 0 at `$0200`: two actual
writes, zero endpoint differences. Tests cover verified stops, exact restore,
mutation divergence, cancellation, missed predicates, assertions, hit counts,
overflow, identity/signature failures, stale request/generation and ownership.
Go tests cover injection/unknown fields, predicate depth/masks, duplicate stops,
unknown releases, invalid budgets and capture overlap/bounds.

Private-media acceptance (requires locally supplied reference TAP; never commit
its generated state):

```sh
node tools/browser/tests/tour-fort.mjs games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  tools/platform/c64/browser/work/m4-tour.rrstate
python3 -m http.server 8779 --bind 127.0.0.1
# Open /tools/browser/tests/code-m4-browser.html
```

Fresh replay from real firmware and tape, without RAM injection, passed on
2026-09-30 using core SHA-256
`65969dd2bdfd1512cafe9e74d9b1d7c21ee395b468d87e0cf3aaa8d23bc7546c`:

| Stop | Machine cycle | Interval cycles | Terrain writes | Terrain differences |
| --- | ---: | ---: | ---: | ---: |
| Entry | 118,436,151 | 0 | 0 | 0 |
| First run | 118,441,859 | 5,708 | 215 | 215 |
| Second run | 118,442,960 | 1,101 | 40 | 0 |

Browser acceptance exercises the real shared app/worker: exact-image state load,
all stops, native button focus, suggested layout, Code/Memory navigation without
execution, free exploration, instruction stepping and exact checkpoint restore,
write/diff distinction, 375px picker sizing, and replacement with unknown media.
The fixture reports the content-addressed release it tested. Game bytes and the
private checkpoint are excluded from version control.
