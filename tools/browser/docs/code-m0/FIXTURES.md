# M0 runnable fixture plan

The commands below run from the repository root. M0 fixtures use synthetic bytes;
no commercial media is required. Fixtures describe the existing implementation
and lock contract examples. They do not certify the future M1 schema or M2 API.

## Run now

```sh
node tools/browser/tests/code-m0-contract.mjs
node tools/browser/tests/code-m0-audit.mjs
python3 -m http.server 8779 --bind 127.0.0.1
```

Open `http://127.0.0.1:8779/tools/browser/tests/code-m0-browser.html` in a browser.
It runs automatically, displays PASS/FAIL and JSON, and terminates its worker.
Save the JSON alongside browser/version, source revision and machine information.
The simple loopback server exposes the repository locally; stop it after testing.
It has no upload endpoint and no cross-origin-isolation headers, unlike production.

To refresh checked-in results deliberately:

```sh
node tools/browser/tests/code-m0-contract.mjs > tools/browser/results/code-m0/contract-vectors.json
node tools/browser/tests/code-m0-audit.mjs > tools/browser/results/code-m0/core-audit.json
```

The browser fixture measures its existing worker's request-to-ack timing at reset,
then a separate synthetic core instance. It has a 15 s response timeout and no
performance pass threshold. Hardware speed should not make correctness tests flaky.
The audit checks public bindings (including minified WASM exports through JS glue),
hashes the cores and executes a synthetic C64 boundary/PC-stop/checkpoint probe.
It does not initialize all 16 machines or claim functional validation of them.

## Concrete corpus and expected results

`tests/fixtures/code-m0/contracts.json` contains:

- NFC/backslash/case path vectors, invalid path examples and a canonical collision.
- Empty-file and `abc` SHA-256 vectors; exact canonical file-set serialization.
- Parallel type/pointer arrays with two synthetic records and zero terminator.
  Record addresses are 16/20, values 42/17, flags 3/7. This is not game data.
- Write 0→1→0: no endpoint difference, but two actual writes.
- `INC $0200; JMP $0800`: after the first instruction, next PC `$0803`, six cycles,
  memory `$0200=1`. The current C64 status latch instead reports `$0800`.

The test-only identity implementation fixes the contract's bytes; it is not a
production validator. M1 must consume these vectors through its real implementation
and remove reliance on a second authoritative normalization implementation.

## Tests to implement at their milestones

| ID | Input / action | Required observation | Gate |
|---|---|---|---|
| ID-1 | Reordered file picker members | identical manifest identity | M1 |
| ID-2 | NFC/case collision; traversal; conflicting releases | explicit rejection | M1 |
| TYPE-1 | Parallel tables in current corpus | exact two records + raw bytes | M1/M3 |
| TYPE-2 | Empty slot, unknown enum, sign-magnitude negative, invalid pointer | distinct validity and correct representation | M1/M3 |
| LOC-1 | Bank switch with same numeric PC | different physical resolution/generation | M2 |
| LOC-2 | Overlay unload/reload and moved segment | stale function invalid; new validated binding | M5 |
| ASSET-1 | Synthetic raw 2352-byte sectors, payload subranges | omit headers/gaps; preserve source provenance | M6 |
| ASSET-2 | Fragmented file → archive → compressed member | correct chain, size cap, corrupt member rejection | M6 |
| CPU-1 | Existing INC/JMP probe | normalized nextPC, exact one instruction | M2 |
| CPU-2 | Breakpoint at current PC, two resume policies | immediate stop vs exactly one bypass | M2 |
| CPU-3 | IRQ/NMI entry, RDY stall, BRK, illegal opcode | explicit entry/fault and retirement counts | M2 |
| CPU-4 | Self-modifying operand; unsupported decode | live update, historical bytes preserved, honest alignment | M2 |
| JOB-1 | Run→capture/save/step; cancel twice; stale session reply | busy/rejection, one terminal result, no stale update | M2 |
| SNAP-1 | Lazy read after machine advanced | retained old bytes or snapshot-expired, never new bytes | M2 |
| TRACE-1 | Current write-then-restore vector; cap hit | activity differs from diff; truncation explicit | M4 |
| TOUR-1 | Missed predicate, timeout, user-run divergence | bounded stop and recovery, no false lesson success | M4 |
| EXP-1 | Replay baseline twice; fail second edit precondition | deterministic outputs; zero partial mutation | M8 |
| EXP-2 | Restore original session after original/patched branches | core/device/input identity preserved as specified | M8 |

Use native/WASM synthetic programs for CPU-1..4; compare registers, selected RAM,
clocks, stop reason and fetched bytes, not only screenshots. Save fixture version
and hashes with every acceptance run. Test condition budgets and decompression
limits independently of proprietary media. Tests must call the actual production
validator/adapter once those exist; the M0 vectors alone are not acceptance.

## Private-media acceptance gates (not run in M0)

| Scenario | Reproduction prerequisite | Evidence needed |
|---|---|---|
| Fort probe | exact TAP/firmware, authentic boot/input recipe | actual `$A000` bytes/path, lookup and state |
| Elite slots | verified working flight scene | pointers, terminator/type, workspace copy-back and reuse |
| Elite loader | exact TAP, authentic loader execution | code mutations tied to pulse position and destination writes |
| UW rendering | exact file set + dungeon input/checkpoint recipe | overlay identity, span writes, buffer/display timing |
| NFS cycle | exact disc/config + City scene recipe | outer loop boundary, linked records, spawn/driver ordering |
| Captain Toad model | exact image/archive/extraction recipe | actual model path, decode and dependency identity |

Do not reuse old checkpoints without core/media/config compatibility checks.
Historical local boot schedules are research aids, not required public fixtures.
Mark absent media or incompatible state as skipped/blocked and identify the gate.

## Baseline extension required in M2

Keep the current reset/synthetic fixtures as stable comparisons. Add authentic
Fort loading and gameplay baselines with debug off/on, 5 Hz selected snapshots,
instruction + decode + resolve + panel render, and active breakpoints with/without
hits. Measure per-call and request-to-ack distribution, throughput, WASM allocation,
retained trace/checkpoint bytes, and browser memory if actually available. Record
sample counts and warm-up; do not relabel a short maximum as a p99 guarantee.
