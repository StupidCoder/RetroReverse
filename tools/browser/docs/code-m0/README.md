# M0 — contracts, inventory and baseline

Completed 2026-09-30 for the initial C64-first scope. At M0 exit, M1/M2 were
unimplemented; subsequent [M1 acceptance](../code-m1/README.md) is recorded separately.
Source baseline: `2284b8c8` (plan commit), inspected with existing unrelated local
changes left untouched. No production emulator or UI code was changed for M0.

## Deliverables

- [Contracts](CONTRACTS.md): schema naming, exact file-set normalization, stable
  identifiers, boundary snapshots, stop/resume semantics, job arbitration,
  cancellation, initial resource budgets, and decoder/persistence decisions.
- [Inventory](INVENTORY.md): all 16 browser bindings, execution seams, native versus
  browser/prototype distinctions, recording limits and four-game research risks.
- [Runnable fixture plan](FIXTURES.md): synthetic contract corpus, current commands,
  future acceptance cases and explicit private-media gates.
- [Core audit](../../results/code-m0/core-audit.json),
  [contract vectors](../../results/code-m0/contract-vectors.json), and
  [browser baseline](../../results/code-m0/browser-baseline.json).

## Findings that affect implementation

1. C64's existing opcode stop reaches the correct bus fetch address, but status
   PC is stale by one instruction in the synthetic probe: `$0800` versus `$0803`.
   M2 must expose an explicitly normalized next-PC boundary, not reuse status.
2. Exported activity methods exist on all cores, but only C64/GG/Amiga currently
   enable Memory activity. Capability discovery cannot inspect export names alone.
3. Save/load exports and native debugger interfaces do not establish browser
   instruction stepping, historical mapping or deterministic tours.
4. Elite needs parallel tables, pointers, zero termination and scratch-workspace
   semantics; a flat array-of-struct design would misrepresent it.
5. The Fort bug story needs byte reconciliation; UW needs overlay binding and
   corrected progress/numeric-format descriptions; NFS needs the outer loop traced.
6. Current worker job control is spread over flags and epochs. A single arbiter
   and per-request terminal results are required before tour execution is added.

## Measured baseline

Environment: Apple M3 Pro, macOS arm64. Node version and exact artifact hashes are
in the core audit. Browser: Codex in-app Chromium 152, no cross-origin isolation.
The browser user-agent's Intel token does not describe the host CPU.

| Measurement | Observed result | Interpretation |
|---|---|---|
| Existing worker reset/BASIC turbo throughput | 4.34 million C64 cycles/s over 12 short intervals | Includes worker status/framebuffer work; no game |
| Pause request → acknowledgement | median 3.15 ms, observed max 5.0 ms | Includes messaging/frame copy, excludes input hardware/DOM paint |
| Largest reported core call in worker run | 7.7 ms | Cumulative maximum in this short run only |
| WASM heap | 134,217,728 bytes (128 MiB) | Allocated heap, not browser RSS |
| Synthetic core-only throughput, browser | 6.91–7.07 million cycles/s | INC/JMP, dummy ROM, display disabled |
| 1,000 instruction steps + status, browser | 239.9 ms total | Status hashes RAM/framebuffer; no decode or Code UI |
| Synthetic serialized state | 230,067 bytes | Current C64 core serialization |
| Instruction boundary observation | six cycles; next bus PC `$0803`, status `$0800` | M2 contract incompatibility reproduced |

All correctness assertions in the Node audit and browser probe passed. The PC
mismatch is recorded as an existing semantic gap, not a failure of this audit.
The browser JSON was read from the rendered PASS page; its checked-in result
preserves selected raw fields and explicitly records that transcription.

These are engineering baselines, not game compatibility or worst-case guarantees.
Actual Code UI timing is unavailable because that UI does not exist; the measured
step/status cost is a labelled proxy and M2 has explicit replacement measurements.
No Fort/Elite/UW/NFS gameplay acceptance, Safari/Firefox comparison, native/WASM
CPU parity, or long-duration soak was performed for M0. Those remain named gates
in the fixture plan, not silently completed work.

## Validation performed

```sh
node tools/browser/tests/code-m0-contract.mjs
node tools/browser/tests/code-m0-audit.mjs
python3 -m http.server 8779 --bind 127.0.0.1
# Open /tools/browser/tests/code-m0-browser.html — PASS
```

Contract vectors validate canonical file-set bytes and synthetic state examples;
they do not implement the future production schema validator. Review checked
local links, JSON syntax and JS syntax before committing these artifacts.

## M0 exit review

- [x] Current browser seams and all core bindings inventoried, with evidence tiers.
- [x] Snapshot/stop/cancel/arbitration semantics fixed for M2.
- [x] Fort/Elite/UW/NFS research inventory and contradictions recorded.
- [x] Schema naming, identity normalization and synthetic representative corpus fixed.
- [x] Repeatable throughput/pause/heap and pre-Code-work baselines recorded.
- [x] Runnable fixture plan, resource budgets, decisions and milestone-owned risks.

This is a source-and-execution review by the implementing agent, not a separate
independent review. Next action is M1: implement the schema/semantic validator and
Fort package against these vectors, then integrate image-matched Memory labels.
