# M0 contracts: knowledge and instruction debugging

Decision record, 2026-09-30. These contracts are the implementation target for
M1/M2, not APIs already provided by the worker. Scope and rollout:
[master plan](../../../../CODE-KNOWLEDGE-TOURS-PLAN.md).

## 1. Decisions fixed in M0

- One `games/<slug>/knowledge.json`; schema version integer `1`, positive integer
  content `revision`. Schema path `tools/knowledge/schema/game-knowledge-v1.schema.json`.
- Top-level collections and location tags are those in the master plan. Collections
  are objects keyed by stable IDs matching `[a-z0-9]+([.-][a-z0-9]+)*`.
  References are local IDs; external identity is `{package, collection, id}`.
  System IDs match `site/emulators/platforms.js` (`ps1`, `3do`, `dos`, etc.), while
  game slugs retain repository conventions (`ridge-racer-psx`, `*-pc`, etc.).
- Canonical addresses/offsets are lowercase `0x` strings with no leading zeroes
  except `0x0`. The master's padded examples are illustrative input, not canonical
  serialized output. Counts/widths/strides are bounded safe JSON integers.
- `schemaVersion`, `id`, `revision`, `game`, `releases` are required. Other known
  collections default to empty. Unknown fields fail except a passive `extensions`
  object whose keys are namespaced; extensions cannot change execution semantics.
- Required type vocabulary: `u8/u16/u32/u64`, `i8/i16/i32/i64` (two's complement),
  explicit endian for multibyte values, struct, array, enum, bitfield, pointer,
  fixed-point and an explicit sign/magnitude representation. Do not interpret
  Elite's separate sign byte as ordinary signed 24-bit two's complement.
- Empty, unavailable, inactive, mismatched and numeric zero are distinct.
- Initial browser decoder choice: generate 6502 opcode/mode data from repository
  decoder definitions and validate parity. Do not ship the incomplete prototype
  decoder as a full 6510 decoder. Illegal opcode behavior needs explicit coverage.
  x86 follows in M5 with parity against Decode/Decode32; binding/port details are
  deferred until that ISA is implemented, not assumed settled by M0.
- Checkpoints start session-local. Existing downloadable states remain separate;
  M0 does not introduce persistent tour checkpoints or state migration.
- No JavaScript tour scripts, arbitrary expression evaluation or decoder URLs.

## 2. Identity normalization, precisely

`releases[release].media` declares raw media roles or a logical file set. Raw
identity is SHA-256 of exact bytes plus exact size. A multi-disc set assigns
explicit roles (`disc-1`, etc.); disc order never follows a file chooser's order.
Lowercase 64-character hex SHA-256 is mandatory. Preserve MD5 only as evidence.

A logical file set declares `pathPolicy: "exact" | "ascii-insensitive"`:

1. Replace backslash separators with `/` and normalize Unicode to NFC.
2. Reject absolute paths, drive prefixes, NUL/control characters, empty path
   components, `.` and `..`; do not silently collapse these components.
3. With `ascii-insensitive`, fold only ASCII A–Z. Do not apply locale-dependent
   or full-Unicode case conversion. Retain original names only for presentation.
4. Reject duplicate canonical paths, even if bytes match. Require every declared
   identity member; extra files do not alter this explicitly selected subset but
   never authorize annotations for unverified extras.
5. Sort paths by unsigned UTF-8 byte ordering. Every entry is
   `[canonicalPath, size, lowercaseSHA256]`. `size` is a nonnegative safe integer.
6. Serialize this array with JSON's compact representation (no optional spaces),
   literal UTF-8 Unicode, ordinary JSON string escaping, no BOM, no trailing LF.
7. Hash UTF-8 `rr-media-set-v1\n` concatenated with those bytes. Store algorithm
   `rr-media-set-v1`, path policy, members and resulting SHA-256.

Raw-disc and file-set matches are separate authorization scopes. Never silently
substitute one for the other. A release may explicitly accept either identity
with distinct location bindings. Multiple matches with conflicting bindings are
`ambiguous`; no automatic tour activation. Firmware and core hashes constrain
runtime recipes separately from static assets.

## 3. Location resolution and validation

Location output:

```
{status, snapshotId, mappingGeneration, spans:[{space, offset, length}],
 sourceChain:[{location, decoder?, version?}], reason?}
```

`status` is resolved/inactive/unavailable/mismatch/ambiguous. Offsets are canonical
hex; lengths are bounded integers; spans are half-open. Module/overlay locations
bind load generation and verified signature. Historical locations use historical
mapping, never today's bank. CPU logical peek is distinct from physical storage.

Initial M1 supports CPU/physical/image and relative locations; additional tags
remain recognized-but-unsupported until implemented. A required unsupported tag
fails activation, rather than returning plausible bytes. Pointer references are
resolved with explicit width, endian, address space and null convention. No bus
reads with side effects. I/O requires a safe latched observation or unavailable.

Validation stages: schema → references/type bounds → release identity → static
location bounds/signatures → runtime applicability/signatures → recipe capability
checks. Passive descriptions remain readable when runtime activation fails.
A signature is mandatory for a code-mutating experiment. A function without a
verified runtime binding may be listed as research but cannot arm a guided stop.

## 4. Worker protocol v1

Do not overload the existing `step` (frame) command. New commands:

| Command | Request fields beyond envelope | Result |
|---|---|---|
| `debug-capabilities` | none | per-processor feature flags and precise limits |
| `debug-snapshot` | processor, selection, normalizeBoundary=false | atomic snapshot |
| `debug-read` | snapshotId, resolved span | bytes/validity at that snapshot |
| `debug-step-instruction` | processor, expectedSnapshotId | next boundary snapshot |
| `debug-run-until` | processor, condition, budget, resumePolicy | job + final snapshot |
| `debug-cancel` | jobId | acknowledged final boundary, or already-terminal |
| `debug-subscribe` | selections, intervalMs | samples at most 5 Hz initially |

Envelope: `{protocol:1, session, generation, requestId, type, ...}`. Execution
requests produce a monotonic `jobId`; replies echo requestId and jobId. `session`
changes on media load; `generation` changes on restore/reset/edit/branch. Requests
with stale identifiers return a structured rejection before touching the core.
Snapshot IDs are unique within session/generation and never mean a live pointer.

Snapshot: `{snapshotId, processor, boundary, nextPC, lastExecutedPC, registers,
 clocks, mappingGeneration, spans, stopReason, imageContext}`. `boundary` is
instruction/device-event/partial; non-instruction stops do not pretend nextPC is
ready. Register bit width/ISA mode is explicit. Clocks are tagged counters encoded
as decimal strings, not imprecise JS numbers; cycles/instructions/frames are not
interchangeable. Selected bytes and registers come from the same worker turn at
that boundary. Lazy reads use retained snapshot data or return `snapshot-expired`;
never silently read the now-advanced machine.

A read-only snapshot never normalizes implicitly. `normalizeBoundary:true` is an
execution request, reports cycles advanced and device/interrupt events, and uses
job arbitration. Code entry asks for this explicitly when needed.

### Exact stopping and stepping

- `nextPC` denotes the next opcode that will execute in the declared mode/mapping.
  C64's current status PC demonstrably fails this contract; M2 must derive a
  verified fetch boundary, account for RDY and interrupt entry, and test it.
- Execution breakpoints check before executing target opcode. Matching includes
  processor, mapping/module binding and condition, not only numeric address.
- `resumePolicy: "stop-if-current"` returns immediately if current boundary matches.
  `"next-match"` bypasses only the already-reported boundary once, then rearms.
- One instruction step retires one guest instruction and advances associated
  devices. If an interrupt/exception is entered before an instruction retires,
  return that entry boundary with `retiredInstructions:0` and an event marker;
  do not silently count the handler's first instruction. Adapters unable to
  expose this precision must declare it and cannot claim the capability.
- Halt/wait is distinguished from illegal/unsupported execution. Yield with a
  reason and device progress; do not spin forever awaiting retirement.
- Multi-CPU stepping declares selected CPU and other-CPU scheduler behavior.
  No global machine pause claim based on stopping only one CPU.
- Trace events retain fetched bytes and mapping. Watchpoint read/write timing is
  explicit; a write watch normally reports after committed write, not before it.
- Terminal reasons: hit, stepped, interrupt-entry, exception-entry, paused,
  cancelled, budget-exhausted, halted, unsupported, mapping-invalid, fault, error.
  Invalid requests are rejections, not core faults. Breakpoint resume never patches
  guest opcodes by default.

### Job arbitration

Single mutating job state: idle, play, step, run-until, capture, memory-record,
seek, save, restore, experiment. The worker owns all transitions. UI disabling is
not a synchronization mechanism. Each accepted job gets exactly one terminal
result, including cancellation; every rejected request receives a reason.

| Incoming operation | Existing job | Action |
|---|---|---|
| Pause/cancel | play, step, run-until, capture, memory-record | request stop; finish at declared boundary |
| Pause/cancel | transactional restore/edit | latch request; complete or roll back atomically |
| New execution/capture/save | any non-idle job | reject `busy` with job ID; no invisible queue |
| Snapshot/read | running job | deliver consistent bounded sample only if supported; otherwise `busy` |
| Load media | any | new candidate session; publish only after initialization succeeds |
| Input | normal play | ordered existing input queue |
| Input | guided deterministic replay | reject live game input; tour navigation still allowed |

Save is serialized and bounded but not interruptible mid-serialization; publish
its latency and preserve the last coherent machine. Seek cancellation restores
its anchor when transactional. Ordinary run-until cancellation leaves the reached
state. Host key release after interactive restore differs from exact experiment
input replay; use distinct policies. Clear queued audio and stale visual captures
on generation changes. Browsing panels does not create an execution job.

## 5. Initial resource budgets

These are limits to implement/measure, not current guarantees:

| Resource | Initial limit/policy |
|---|---|
| Package JSON | 8 MiB; reject larger before parse |
| AST | 64 nodes/condition, depth 16, no loops; checked operations |
| Location chain | depth 16, 256 spans/request, cycle rejection |
| Decompression | 64 MiB output per job, explicit cancel and decoder work budget |
| Selected snapshot/read | 256 KiB per request; larger views page |
| Live display | at most 5 samples/s, virtualized lists |
| Recording | 524,288 events AND 32 MiB total; first limit wins |
| Recent execution trace | 8,192 entries within recording budget |
| Tour checkpoints | max 4 retained, aggregate 256 MiB; each <=128 MiB |
| Cancellable work | target <=8 ms worker slice; pause acknowledgement target <=250 ms |
| Run-until | explicit emulated budget + wall-time cap, default 30 s, user may retry |

Count limits do not imply all platform payloads fit the byte limit. Stop or mark
truncation according to recipe policy. A prepared experiment requiring more than
its checkpoint budget is unsupported until deliberate policy changes; never drop
its recovery anchor silently. Disabled debugging should add no memory-access
logging; benchmark its dispatch/check cost before release.

## 6. UI decisions

Extract panels incrementally, keep Code/Memory entry points. Stable preview and
transport; primary code/memory/buffer pane plus supporting state/records. Above
1,100 CSS px use columns, 720–1,099 use main + tabbed support, below 720 stack
preview and one selected inspector. These are starting breakpoints to validate,
not a forced global redesign. Tour layouts change only at stops and respect pins.

Code work baseline is not a completed UI benchmark: M0 measures core step/status
as a proxy. M2 must add actual decode/resolve/render timings and play overhead
with debugger enabled/disabled on the same fixtures.
