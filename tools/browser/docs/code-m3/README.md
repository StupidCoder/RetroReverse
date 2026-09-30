# M3 — Structured state and shared panels

Implemented 2026-09-30 for the initial C64/Fort scope. Elite's conditional
compatibility gate remains open; no Elite runtime package is enabled.

## Delivered

- Exact signed/unsigned integer decoding, both byte orders, enums with numeric
  unknown fallback, bitfields and exact fixed-point decimals. Values use decimal
  strings/BigInt internally so 64-bit state is not rounded by JavaScript Number.
- Nested records and fixed arrays with explicit stride/padding. Every node retains
  its raw bytes and offset. Unavailable bytes remain `null`, never coerced to zero;
  readable fields inside a partly unavailable record can still be inspected.
- Optional array occupancy checks against integer/enum fields, including nested
  record paths. Unknown enum types may be occupied; empty and unavailable slots
  are separate states. Observed empty/unavailable → occupied transitions increment
  a per-slot incarnation. This is sampled evidence, not continuous object identity:
  removal/reuse between samples is not detectable and is explicitly described.
- A shared Game state panel in Code and Memory. This is the same panel instance
  and the same worker-produced decoded snapshot, not two state implementations.
  Expansion reveals raw bytes and nested fields. State entries link to physical
  Memory when mapped and to CPU addresses or documented related functions in Code.
  These links are disabled during execution/capture/recording/save.
- Hide/show and above/below placement preserve expanded fields and workspace
  navigation. Native disclosure controls and DOM text rendering are used; package
  descriptions are never interpreted as HTML or executable expressions.
- Fort package revision 2 adds related function links for the six already
  documented watches: enemy mode/column/row and player column/row/bank. Their
  applicability and evidence remain visible. No velocity or other inferred watch
  was added, and enum labels do not assert that gameplay has initialized.

## Snapshot and resource contract

`debug-worker.js` adds `state` to the existing debugger snapshot. The state object
has the same cycle string and boundary flag as its enclosing CPU snapshot.
All mapped reads and decoding finish synchronously without yielding or executing
instructions. CPU reads use M2's safe mapped peek; physical RAM uses a copy of
backing RAM. I/O/open-bus bytes are unavailable. Location ranges cannot wrap out
of C64 storage. No new native/WASM API or execution owner was introduced.

Code's existing five-Hz sampler feeds the panel there. In Memory, state refreshes
follow Memory overview updates and paused execution updates; a second live polling
loop is not started. The Memory atlas can display a historical recording while
the state panel identifies its own current CPU snapshot/cycle. State links request
a fresh paused Memory snapshot rather than accidentally navigating old recording
bytes. Historical structured-state decoding is reserved for M4 interval evidence.

Limits are 16 KiB of selected state bytes per snapshot, 4,096 decoded nodes,
16 nesting levels and 256 elements per array. Oversized/unsupported entries show
an explicit error instead of freezing the page. Expanded raw views show at most
256 bytes with a Memory link for the remainder. Array/record descendants mount
on disclosure. Model limits apply even when the panel is hidden.

The schema-v1 additive state properties are:

```json
{
  "relatedFunctions": ["documented-function-id"],
  "occupancy": {"path": ["nested-record", "type"], "notEquals": "0"}
}
```

`occupancy` is permitted only on an array state definition; an empty path compares
its scalar element. Paths must resolve to integer/enum storage, constants must
fit that storage, and related functions must exist for all selected releases.
No generalized predicate evaluator or script execution is added. The validator
also now accepts a one-byte state range at the final byte of an address space.

## Acceptance

- `node tools/browser/tests/structured-state.mjs`: exact numeric decoding,
  partial reads, unknown enum vs empty slot, strided records, observed slot reuse,
  relative locations and byte/node/slot budgets.
- `node tools/browser/tests/debug.mjs`: existing M2 jobs plus real packaged WASM
  Fort watches checked against physical bytes and the identical CPU clock.
- `go test ./tools/knowledge ./tools/cmd/knowledgecheck ./tools/cmd/knowledgeexport`:
  schema, reference, occupancy-field and storage bounds, final-byte watch, export.
- Existing knowledge, Memory and UI-shell regressions pass. Packaged release
  integrity and static network checks pass. All packaged core hashes are unchanged.
- Actual Chromium: `code-m3-browser.html?fort` runs M2 regression first, then checks
  all six watches, raw expansion, panel hide/reorder, panel identity across Code
  and Memory, links, zero-execution navigation, live link disabling and old-state
  removal when loading an unknown image.
- Actual Chromium: `code-m3-panels.html` verifies rendered records with empty,
  unknown, occupied, reused and unavailable slots and persistent expansion.

Browser evidence is recorded in `results/code-m3/`. The local reference Fort
TAP is used for identity and reset/ROM execution, not a verified gameplay phase.

## Elite gate and next work

The current repository compatibility ledger establishes Fort gameplay but has no
verified Elite run on this browser core. M3 does not claim that historical Elite
rendering issues are fixed or still reproduced. Elite's type list, zero terminator,
indirect 37-byte records and scratch workspace require a separately validated
adapter; flattening them into these fixed arrays would be incorrect. Live Elite
slots therefore remain disabled, as allowed by the milestone's compatibility gate.

Pointer chasing, joined parallel arrays, runtime counts, generalized predicates,
continuous incarnation history, historical decoded snapshots and tour-controlled
layouts are not claimed by this milestone. The shared panel and decoder provide
bounded building blocks for their later validated implementations.
