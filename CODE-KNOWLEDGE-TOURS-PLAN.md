# Game knowledge packages, Code workspace, and guided investigations

Status: M0–M3 completed for the initial C64/Fort scope, 2026-09-30; see the
[M0 contracts and baseline](tools/browser/docs/code-m0/README.md) and
[M1 schema/package acceptance](tools/browser/docs/code-m1/README.md) and
[M2 debugger acceptance](tools/browser/docs/code-m2/README.md).
See also [M3 structured state acceptance](tools/browser/docs/code-m3/README.md).
Elite remains gated on verified compatibility. M4 onward remain planned. This document records the agreed product direction;
it does not claim that the interfaces or schema below are implemented. Milestones
require their own acceptance evidence before being marked complete.

Delivery: commit and push after each completed milestone, as requested.

## 1. Outcome and scope

Make RetroReverse explain the running game's internals: executable code,
identified functions, actual game-state records, assets, and the relationships
between them. Add a **Code** tab beside Play, Render, and Memory. Let authored
guided tours advance the real machine between meaningful stops, with explanations
and appropriate inspection panels. Later, support reproducible experiments that
prepare a situation, run original code, restore it, apply a documented patch,
and compare the result.

The canonical output of future reverse engineering is a **single versioned
knowledge file per game**. It contains identity, known locations and meanings,
object layouts, assets, functions, evidence, tours, and experiments. The UI and
tools consume those definitions rather than copying addresses into unrelated
JavaScript files. The narrative write-up remains the readable account of the
research; structured facts should come from the package wherever practical.

The first complete vertical slice is Fort Apocalypse on C64. Elite adds fixed
object slots and a longer loader tour. Ultima Underworld adds relocated overlays,
self-modifying rendering code, and off-screen buffers. Need for Speed on 3DO
tests a full simulation-loop tour. Captain Toad tests filesystem/archive-based
asset discovery. These are separate acceptance targets, not a promise to ship
every example in the first release.

### Product commitments

- Generic inspection works without a recognized game or authored tour.
- Show actual stored state. Do not invent velocity or other fields to fit a UI.
  Any useful computed measurement is explicitly derived, with its formula and
  emulated-time basis shown separately from memory fields.
- Selecting a function navigates the view; it never changes the program counter.
- Run-to-function stops before the selected instruction executes.
- Explanations distinguish observation, interpretation, and unverified hypotheses.
- Keep game-specific addresses and narratives outside hardware cores.
- User-selected media stays local. Packages contain knowledge, not game images.
- Default inspection is read-only. State edits belong to explicit experiments.
- A tour can arrange panels at a stop without taking away free exploration.

## 2. Existing foundations and boundaries

Relevant implementation and planning references:

- [Workspace registration](site/emulators/workspaces.js),
  [application wiring](site/emulators/app.js), and
  [worker ownership/execution](site/emulators/worker.js).
- [Memory workspace](site/emulators/memory-workspace.js),
  [memory service](site/emulators/memory-worker.js), and
  [existing image-matched labels](site/emulators/memory-labels.js).
- [Memory inspector plan](MEMORY-INSPECTOR-PLAN.md) and
  [current coverage](site/README.md#memory-inspector).
- [C64 low-level controls](tools/platform/c64/browser/core/core.h),
  [implementation](tools/platform/c64/browser/core/core.cpp). The standalone
  lesson disassembler at `tools/platform/c64/browser/web/disassemble.js` was also
  inspected locally; that prototype file is untracked at the time of this plan.
- [DOS browser execution loop](tools/platform/dos/browser/core/api.cpp),
  repository CPU disassemblers under `tools/cpu/`, and
  [native debugger contracts](tools/debug/debug.go).
- [Browser observatory plan](BROWSER-OBSERVATORY-PLAN.md). The local
  `EDUCATIONAL-DEBUGGER-PLAN.md` also informed this proposal; it is untracked at
  the time of writing. This document preserves the relevant future requirements
  without requiring that local historical plan to be published with it.

Memory already supplies a game preview, shared transport, snapshots, and bounded
recording on some platforms. Snapshot coverage is broader than access-recording
coverage. The browser's existing `step` command is a frame step, not an instruction
step. C64 has PC/opcode/write stop primitives and a separate lesson/history UI;
those are useful foundations, not an already integrated cross-platform Code tab.
The small C64 lesson decoder is incomplete, including the `$55` opcode relevant
to the proposed Fort investigation. DOS has an instruction loop but needs a
browser-facing register/step/breakpoint contract.

Preserve the distinctions between Memory's historical byte reconstruction,
Render's captured interval, and actual machine checkpoint restoration. A
historical display cursor is not permission to run from that historical state.

This plan extends the existing plans. It does not replace the repository's
[standards](STANDARDS.md), [analysis playbook](ANALYSIS-PLAYBOOK.md), or
[extracted level format](FORMAT2.md). Continue deriving game knowledge through
repository tools and implementing reusable decoders in Go. Retain existing
disassembly stores and evidence while migrating their structured knowledge.

## 3. User experience and panel composition

### Free inspection

Keep Code and Memory as distinct entry points into one machine session and a
shared panel library. Code initially pauses at an instruction boundary and
focuses on the next instruction. If reaching that boundary advances a machine
paused mid-instruction, report that normalization rather than calling it a
zero-execution operation. Reopening a panel preserves navigation where possible.

Default Code layout:

1. Function search/list and explanation, grouped by documented subsystem.
2. Disassembly with current-PC marker, raw bytes, symbols, inline commentary,
   registers/flags, and changed-value highlighting.
3. Game preview, transport, and selected game-state watches or records.

Controls: Play, Pause, Next instruction, Next frame, Run until selected function,
Cancel, Follow PC, and Return to PC. Step over/out is a later capability, not an
alias for blindly running to the next address. It must account for interrupts,
exceptions, recursion, and nonstandard control flow.

Function selection while paused only changes the viewed location. During play,
users can inspect metadata; exact live navigation/stepping waits for a pause.
Following PC is optional, so browsing does not constantly snap back. Refresh
live samples at a bounded rate (initial target 5 Hz); never stream every executed
instruction to the DOM. Single steps produce exact boundary snapshots.

Object tables expose decoded fields and expandable raw bytes. Selecting an
object identifies its actual backing range and relevant code/assets. Unknown
enum values remain numeric and visibly unknown. Pointer failures, inactive
overlays, empty slots, and unavailable state are distinct from a stored zero.

### Tour layout

Use stable panel roles: explanation/Continue, game preview/transport, a primary
inspection panel, and a supporting panel area. Stops request panels by ID and
semantic selection, such as a function, object slot, memory range, or asset.
Available panels include disassembly, registers, hex, memory atlas, object table,
watches, tape activity, rendering buffer, and recorded events.

Only change layouts at stops. Preserve compatible selections and scroll state;
allow pinning/rearranging panels and provide **Restore tour layout**. On small
screens, supporting panels become tabs. Do not require every panel to remain
visible. Show missing capabilities with an explanation or an authored fallback,
not an empty panel. Keep keyboard focus predictable and support reduced motion.

Navigation links carry semantic identity and snapshot identity: function to code,
field to memory, writer to historical instruction, asset to source location.
Do not silently combine a historical value with current registers or mappings.

## 4. Knowledge package specification: proposed v1

### 4.1 File ownership and versioning

Canonical source: `games/<slug>/knowledge.json`. Proposed schema location:
`tools/knowledge/schema/game-knowledge-v1.schema.json`. Proposed Go validator and
compiler: `tools/cmd/knowledgecheck` and `tools/cmd/knowledgeexport`, with shared
implementation under `tools/knowledge/`. These paths are planned, not existing APIs.

Use strict UTF-8 JSON, validated with JSON Schema plus semantic checks. Markdown
strings hold narrative explanations. One source file contains all game-specific
definitions; no required include tree of hand-maintained facts. Large binaries,
generated disassembly, screenshots, extracted assets, and checkpoints remain
external artifacts referenced by this file. References to shared decoder code
are allowed; they must not become a second hidden store of game-specific offsets.

Top-level fields:

| Field | Meaning |
|---|---|
| `schemaVersion` | Integer contract version; reject unsupported versions |
| `id`, `revision` | Stable game slug and package content revision |
| `game` | Display title, system ID/name, description |
| `releases` | Explicit supported media sets and revision-specific bindings |
| `sources`, `evidence` | Research provenance and verification recipes |
| `spaces`, `locations` | Address spaces and composable location definitions |
| `types` | Primitive conventions, enums, bitfields, structures, arrays |
| `regions`, `state` | Known ranges, live values, object collections |
| `functions`, `annotations` | Code identities and commentary |
| `assets` | Semantic assets, sources, formats, dependencies, export recipes |
| `tours`, `experiments` | Declarative execution and presentation recipes |

Each collection is keyed by a stable local ID. References use those IDs, not
display names or copied addresses. External consumers qualify them with package
ID. Renames preserve aliases through migration; do not reuse an ID for a different
meaning. Schema migration is explicit and deterministic. Unknown execution
operations fail validation rather than being ignored. Allow namespaced passive
extension metadata, but never execute it as a fallback.

### 4.2 Media and release identity

Each release declares game version, region/system configuration where relevant,
and media roles. Media entries contain SHA-256, exact size, and a meaningful role;
preserve existing MD5 values as provenance, not the primary matching mechanism.
Require lowercase 64-character SHA-256 hex in production packages.

Support a single ROM/tape, multiple ordered discs/tracks, and a set of required
files. Filesystem-based imports use canonical relative paths plus per-file
size/hash. Define normalization explicitly (separators, case policy, Unicode,
duplicate paths); never merge ambiguous names silently. A manifest identity is
computed from a specified sorted serialization, not a directory's incidental
iteration order. Distinguish raw-image identity from logical file identity and
state which match authorizes which annotations.

Firmware, emulator build/state-format, timing region, and compatibility options
belong in reproducibility requirements for execution evidence. They need not
block purely static asset annotations when irrelevant. Unknown media retains
generic inspection. An explicit unverified package preview may show descriptions,
but must not automatically arm tours or edits.

Multiple supported releases live in one file. Use common semantic definitions
plus explicit release bindings for locations and verified signatures; reject
ambiguous overlaps. Do not guess an address translation from similar titles.

### 4.3 Spaces, locations, and resolution

A location resolves to bytes and a context, not just an integer. All offsets and
lengths are byte-based unless a tagged kind explicitly states otherwise. Ranges
are half-open `[start, start + length)`. Store addresses/offsets as canonical hex
strings to avoid JavaScript integer precision loss; counts use bounded integers.
Runtime code uses checked integers/BigInt as appropriate.

Required tagged location kinds, introduced incrementally:

| Kind | Required meaning |
|---|---|
| `cpu` | Processor/address space, logical address, mapping applicability |
| `physical` | Named physical region/bank and offset |
| `image` | Media role and raw byte offset |
| `sectors` | Track/media, sector layout, first sector/count, payload selection |
| `file` | Parent filesystem location and exact normalized path |
| `member` | Parent container and member identity |
| `transform` | Source plus registered decoder/version and parameters |
| `relative` | Base location plus checked byte offset |
| `module` | Executable/overlay identity and module-relative offset |
| `pointer` | Typed pointer read from a location with target-space rules |

Sector descriptions distinguish raw sectors, user payloads, headers, and stride;
do not assume every disc has 2,048-byte sectors. Files can have noncontiguous
extents. Compressed data is a transformation, not a false linear memory alias.
Tape pulse indexes and emulated cycles are explicitly tagged units, separate
from TAP byte offsets. Runtime asset mappings require evidence; an archive entry
does not automatically identify every relocated copy in RAM.

A resolver returns validity (`resolved`, `inactive`, `unavailable`, `mismatch`,
or `ambiguous`), spans, source chain, mapping generation, and snapshot identity.
Bound chain depth, decoded size, pointer traversal, and decompression work. Reject
cycles, overflow, out-of-range reads, and path traversal. No implicit network
fetch, shell execution, JavaScript evaluation, or arbitrary code import.

Use safe debug peeks for CPU-visible code. Reading an I/O register through normal
emulated bus access can change the machine; expose a latched observation or mark
it unavailable instead. Physical RAM and banked CPU-visible bytes are different
spaces. DOS locations retain segment/offset and execution mode as well as linear
address. Overlay identity includes load generation/signature, not only CS.

### 4.4 Types, regions, and game state

Support explicit width/endian primitives; signed/unsigned integers; bitfields;
enums with numeric fallback; fixed-point formats with exact signedness/scale;
structs with offsets, lengths, padding and unknown bytes; arrays with element
type, stride and fixed or bounded runtime count; typed near/far/banked pointers.
Allow explicit unions/overlap only when documented. Validate field bounds,
alignment assumptions, enum widths, and count/stride limits.

Regions declare location, length, purpose, lifetime/applicability and evidence.
State entries bind a type to a location and give a label, description, units,
validity condition, and optional related function/asset IDs. Arrays model both
array-of-struct and parallel arrays. An occupancy predicate identifies empty
slots independently of a zero-valued field. Include slot incarnation in history
so a reused slot is not mistaken for the same object.

Derived displays are a separate tagged kind with bounded expressions, provenance,
and sampling semantics. Never present them as stored fields. Expressions use a
typed, side-effect-free AST (constants, field references, bounded arithmetic,
comparisons, masks, Boolean operators); no `eval`, arbitrary loops or host access.
Define overflow, signedness, unavailable operands, and integer division explicitly.
Condition evaluation has a fixed cost/depth budget.

### 4.5 Functions and annotations

A function declares an entry location, optional known code ranges, processor/ISA
and mode, label, subsystem tags, explanation, evidence, release applicability,
and validation signatures. Optional arguments, clobbers and return descriptions
are only included when established. A function need not occupy one continuous
range or return conventionally. Distinguish routine entry from internal probes.

Annotations attach to stable function IDs plus instruction-relative locations or
resolved ranges. Include branch intent, effective-address explanations, and
related state/asset IDs. Store masked signatures only where relocation or known
self-modification justifies the mask; a permissive mask is not identification.

Disassemble current mapped bytes. Historical events retain actual fetched bytes
and context, so later modifications cannot rewrite the apparent past. On variable
length ISAs, decode from a known entry/executed boundary; do not pretend arbitrary
backward decoding from PC is unambiguous. Unsupported opcodes remain explicit and
must not cause misleading instruction alignment.

### 4.6 Assets

An asset declares semantic name/category, source location(s), format and decoder
ID/version, parameters, related textures/palettes/skeletons/animations, evidence,
and optional preview/export artifacts. Record source spans and transformation
chains separately from exported file locations. Permit shared data and multiple
source fragments without duplicating the bytes.

For example, “Captain Toad's character model” should resolve through the verified
image → filesystem file → archive member → decompressor → model resource chain,
and link its textures/skeleton where known. Actual paths and formats must be
researched, not filled with plausible guesses. Other examples include ROM bank
graphics, raw disk sectors, and dynamically unpacked RAM assets.

Registered decoders are repository code. Packages provide parameters and facts,
not implementations of arbitrary decoding scripts. Static exports remain clean
reimplementations under the existing Go/tool conventions. Any browser counterpart
needs parity fixtures; do not replace extraction with scraped oracle output.

### 4.7 Evidence and authorship

Every substantive definition references evidence with status `confirmed`,
`inferred`, or `hypothesis`, source document/anchor or generated artifact,
derivation/tool recipe, tested release, and applicable limitations. Execution
claims additionally pin core/configuration, starting identity, input schedule,
and expected observations. Keep artifact hashes where appropriate. Old addresses
or contradictory prose produce validation issues, not silent precedence rules.

Small expected-byte signatures may live in the package. Whole game images,
large extracted binary payloads, and generated machine states do not. Referenced
artifacts follow existing repository copyright/output conventions. Site exports
carry package revision/content hash so stale derived views are detectable.

### 4.8 Illustrative package fragment

This is a partial design example using documented Fort addresses. It is not a
complete validated package, and does not prove the proposed bug reproduction.
The initial schema milestone must turn such fragments into executable fixtures.

```json
{
  "schemaVersion": 1,
  "id": "fort-apocalypse-c64",
  "revision": 1,
  "game": {
    "name": "Fort Apocalypse",
    "system": {"id": "c64", "name": "Commodore 64"}
  },
  "releases": {
    "reference-pal": {
      "media": [{
        "role": "tape", "size": 225817,
        "sha256": "9e444c4576bac52ba691f0ffe2c0a7efb0f62f3fe2be7cbe78dba08672dda00b"
      }]
    }
  },
  "spaces": {"main-cpu": {"kind": "cpu", "processor": "6510"}},
  "locations": {
    "enemy-state": {"kind": "cpu", "space": "main-cpu", "address": "0x006e"},
    "enemy-up": {"kind": "cpu", "space": "main-cpu", "address": "0xa000"}
  },
  "types": {
    "enemy-mode": {
      "kind": "enum", "storage": "u8",
      "values": {"1": "Idle", "3": "Hunting", "4": "Dying"}
    }
  },
  "state": {
    "enemy.mode": {
      "label": "Enemy helicopter state", "location": "enemy-state",
      "type": "enemy-mode", "releases": ["reference-pal"],
      "evidence": ["enemy-ai-writeup"]
    }
  },
  "functions": {
    "enemy.upward-probe": {
      "label": "Enemy upward terrain probe", "entry": "enemy-up",
      "isa": "mos6510", "releases": ["reference-pal"],
      "evidence": ["enemy-ai-writeup"]
    }
  },
  "evidence": {
    "enemy-ai-writeup": {
      "status": "inferred",
      "source": "fort-apocalypse-c64.md#3-enemy-helicopter-sprite-1",
      "limitations": "Revalidate live instruction bytes and gameplay applicability before enabling the bug tour."
    }
  }
}
```

## 5. Runtime architecture and contracts

### Core and worker

One worker remains the sole owner of each machine. Add a debug service beside
Memory/Render services with capability discovery, consistent snapshots, safe
peeks, instruction stepping, breakpoints, and optional recording. Capabilities
are per processor and include ISA modes, step precision, read/write recording,
checkpoint coverage and replay support. A missing capability is not inferred
from the existence of a similarly named export.

Proposed commands: `debug-capabilities`, `debug-snapshot`, `debug-read`,
`debug-step-instruction`, `debug-run-until`, `debug-cancel`, and
`debug-subscribe`. Exact naming/ABI is fixed in M0/M2. Every response carries
session, machine generation, request/job ID, processor, boundary/clock and
mapping generation. Reject stale responses after load, restore, edit, or reset.

Snapshots atomically describe registers, next PC, selected memory, mapping and
stop reason at one boundary. Distinguish the last executed PC from next PC; audit
the existing C64 `ctx.pc` latch rather than exposing it unquestioned. Define
interrupt-entry, fault, halted-CPU, and DMA behavior. Instruction stepping advances
the complete machine's required device clocks, not an isolated CPU detached from
video/timers. Multi-CPU systems require an explicit stepping/scheduling policy.

Run-until uses pre-execution core checks, never JS callbacks per instruction or
polling once per frame. Prefer side-table breakpoints over patching guest opcodes.
Define current-PC behavior: “run until next invocation” passes the current stop
once, then rearms; a plain seek-to-boundary may stop immediately. Do not globally
suppress a breakpoint after resume. Conditions use compiled bounded predicates.

Execution jobs yield in bounded slices, service cancellation, and return explicit
reasons: hit, paused/cancelled, budget exhausted, unsupported, invalid mapping,
halt/fault, or error. Initial responsiveness targets: ordinary UI feedback within
100 ms and pause/cancel serviced within 250 ms; measure on named fixtures and
hardware rather than promise universal performance. Audio presentation is muted
or flushed as needed for stepping; emulated sound-device time still progresses.

Unify arbitration for run, frame capture, memory recording, tour continuation,
save/restore and experiments. Only one mutating execution job runs at a time.
Cancellation semantics are explicit: ordinary run-until stops at the reached
boundary; transactional restore/edit work either completes or restores its anchor.
Keep input ordering deterministic and clear pending presentation/audio after seek.

### Decoder and package services

Separate platform debug adapter, disassembler, package resolver, tour engine, and
panel presentation. Reuse repository CPU decoder knowledge through generated
tables or validated ports/bindings; do not introduce external game-specific RE
sources. Define parity tests against repository decoders before selecting the
browser implementation. Fetch only the selected platform/package and requested
spans. Cache immutable media separately from live mappings and invalidate decoded
code on relevant writes/overlay generation changes.

Proposed browser modules: `code-workspace.js`, `debug-worker.js`,
`knowledge-model.js`, `knowledge-resolver.js`, and `tour-controller.js`, with
reusable panels extracted incrementally from current workspaces. Avoid a broad
rewrite before the first vertical slice. Publish generated package indexes and
version-matched knowledge files through the existing browser packaging process.

## 6. Tour contract

A tour defines ID/title, explanation, supported releases/configurations, required
capabilities, entry requirements, start policy, ordered/branched stops and exit
behavior. A stop contains stable ID, run condition, execution budget, capture
requirements, assertions, explanation, panel layout, and permitted next stops.

Start policies distinguish “continue current session”, “replay verified boot/input
recipe”, and “restore compatible generated checkpoint”. Do not inject invented RAM
while describing an authentic boot. Checkpoints are derived artifacts pinned to
media, firmware, core, state format, configuration and input history.

Conditions initially support function/address entry, field comparisons, masks,
specified object slot, bounded hit count, and conjunction/disjunction. Add caller
conditions only with dependable observed call context. Device events such as tape
edges/writes are separately tagged. Budgets include emulated work and wall-time
limits. A timeout explains the last known state and permits retry/free inspection.

Tour lifecycle: idle → preparing → running-to-stop → paused-at-stop → continuing
→ completed; also cancelled/failed. Continue evaluates the next step against the
current generation. Manual running, edits, loading or restoring invalidates tour
assumptions unless explicitly supported. Browsing paused panels does not. Offer
return to the stop anchor when available, and visibly mark a diverged session.

Illustrative stop (partial, dependent definitions supplied by its package):

```json
{
  "id": "inspect-upward-probe",
  "until": {"kind": "function-entry", "function": "enemy.upward-probe"},
  "budget": {"cycles": 10000000, "wallMs": 30000},
  "capture": {"snapshot": true, "recording": "probe-evidence"},
  "explanation": "Inspect the fetched opcode and the column used by the probe.",
  "layout": {
    "primary": {"panel": "disassembly", "selection": "enemy.upward-probe"},
    "supporting": [{"panel": "state", "selection": "enemy.mode"}]
  },
  "next": "inspect-terrain-result"
}
```

### Evidence between stops

Maintain a stop baseline and a current snapshot. **Value differences** compare
endpoints; **recorded activity** shows writes/reads/execution during the interval.
They are different features: a byte may be written and restored to its old value.
Label the distinction in both UI and data. A stop's recording request is armed
before its incoming execution interval, not after reaching the stop.

Recording declarations select ranges, event kinds, maximum events/bytes and
overflow policy. Record exact writer contexts only where supported. No full
instruction trace by default. On overflow, mark evidence incomplete; if a lesson
assertion requires completeness, fail that assertion or retry with a narrower
window. Do not silently present a truncated atlas as all changes.

Offer bounded recent trace, branch decisions, effective operands and “why did this
change?” links when recorded. A watchpoint stops on an actual write; an endpoint
value comparison is not an equivalent implementation.

## 7. Controlled experiment contract

Experiments extend tours with explicit, inspectable mutations. They declare setup
requirements, expected pre-edit values/signatures, typed memory/register/code
edits, setup invariants, baseline checkpoint, input schedule, comparison duration,
observations and cleanup. Read-only tours do not gain mutation permission by
accident. The user starts a labelled experiment and can review the changes.

Protocol:

1. Retain the pre-tour session checkpoint, subject to a visible storage budget.
2. Reach a validated baseline and prepare the demonstrated situation. Prefer
   invoking observed game behavior or a documented consistent state edit over
   changing a few coordinates and assuming all related state follows.
3. Verify occupancy, AI state, camera/world position, timers and other required
   invariants; save a complete prepared checkpoint.
4. Run original code with recorded inputs for a specified emulated interval.
5. Restore the exact prepared checkpoint, including devices/RNG/input queues.
6. Apply the checked patch transactionally and invalidate code/mapping caches.
7. Replay identical inputs for the same interval; compare declared observations.
8. Offer restore of the original session or explicit continuation of a labelled
   modified branch. Reset/cancel/failure cleanup is deterministic.

Never modify source media or silently persist patches. A write mismatch aborts
before any partial edit is left applied. Branch IDs distinguish original/patched
history and prevent stale evidence links. Repeated replay must first establish
determinism; otherwise report an inconclusive comparison. A hypothesized repair
remains a hypothesis even if a particular demonstration improves.

## 8. Reference use cases and acceptance stories

### A. Fort Apocalypse: code and state, then AI investigation

Use the [existing research](games/fort-apocalypse-c64/fort-apocalypse-c64.md).
Initial functions include enemy AI `$9C52`, hunting `$9CDA`, upward probe `$A000`,
and player movement `$A4CE`, subject to runtime/release validation. Initial watches
include enemy state `$6E`, enemy map column/row `$76/$77`, player map position
`$69/$6A`, and bank `$67`. Do not label bank or position as stored velocity.

First tour stops at the actual upward probe, explains the fetched bytes and
effective terrain lookup, then shows the resulting state/behavior. Reconcile the
signature-overlap description with binary evidence before teaching it as fact;
the earlier educational plan explicitly leaves this investigation pending.

Later experiment prepares an underground enemy spawn and visible player target,
validates all dependent state, records the original run, restores, applies a
verified candidate repair, and compares the same interval. Demonstrate a ceiling
crash only if reproduced. Do not predetermine the result or claim the patch fixes
every AI problem. Acceptance includes repeatable original behavior, transparent
edits, a meaningful comparison, and successful return to the original session.

### B. Elite: universe slots and spawning

Convert the [universe/AI research](games/elite-c64/elite-c64.md) into a fixed-slot
collection with documented occupancy, object type, raw bytes and known fields.
Do not assume every field is contiguous: model slot records, type tables and
temporary zero-page workspaces as documented. Link object types to known ship
blueprints; human-readable ship names are annotations, not necessarily strings
stored in the game. Preserve their evidence status.

Expose slot allocation and the routines that choose/populate new objects,
including the documented spawn routine `$855B` after validation. Distinguish
random selection from allocation and per-object update. Break on a selected
slot's creation, show before/after bytes and interpreted values, then observe
reuse without confusing it with the previous occupant.

Gate on browser compatibility: an earlier observatory report records incorrect
Elite rendering. Re-test current behavior and record its status before claiming
the tour works; historical documentation is not a current acceptance run.

### C. Elite: self-modifying fastloader tour

Follow authentic tape execution between named stops. Show pulse interpretation,
decoder code before/after modification, destination memory, and interval writes
on the atlas. Establish precisely which encoding/decoding rules change from
the tape and live trace; do not conflate transformed payload bytes with changed
physical pulse encoding. Explain the observed mechanism rather than relying on
the shorthand “changes encoding on the fly”.

Continue runs to the next verified event. Keep tape position, code, memory and
event timestamps synchronized. Prove actual writes, including bytes restored to
their original value, against bounded trace fixtures. No synthetic installation
of already decoded loader results while claiming authentic execution.

### D. Ultima Underworld: world to screen

Use [renderer research](games/ultima-underworld-pc/ultima-underworld-pc.md) and
the [disassembly store](games/ultima-underworld-pc/disasm/README.md), reconciling
stale introductory summaries with later detailed findings. Documented runtime
examples include projection `07F7:6148`, horizontal span setup `01A0:0296`, texture
span `01A0:02CE`, and Mode X copy `01A0:0B96`. These are observations from a
particular loaded arrangement, not globally valid addresses.

Resolve module/overlay identity first. Explain a projected vertex, a span, changes
to self-modified gradient operands, drawing into the off-screen buffer, and the
final display copy. Show documented matrix/span fields with validated numeric
formats. The last completed display frame may remain unchanged while a buffer is
being drawn; label the output and buffer times rather than forcing a frame step.

Acceptance includes a reloaded overlay, signature mismatch rejection, current
versus historical code bytes, and consistent code/buffer observations.

### E. Need for Speed (3DO): one simulation-loop cycle

Build from [driver research](games/need-for-speed-3do/need-for-speed-3do.md),
including documented road-AI and dormant-pool driver entry points, after browser
verification. Establish a precise cycle boundary, then tour spawn gates,
per-car driver dispatch and state updates until that boundary is reached again.

Do not equate one game-loop iteration with one video frame. Distinguish dormant
pool entries, active traffic, opponent and police behavior as supported by the
actual records. An iteration in which no car spawns should explain why the gates
reject it rather than fabricate a spawn. Conditions select the intended car and
loop generation; task scheduling/OS HLE limitations remain visible. Acceptance
compares traced entry order and relevant state changes to the declared cycle.

### F. Captain Toad and other asset sources

Author a verified semantic model asset and its complete location/decoder chain.
Resolve it from user-supplied matching media, link dependencies, and open a
preview/export through existing tools. Validate the filesystem/archive path and
decoded identity. Also exercise one ROM-range asset and one raw-sector asset so
the schema is not tailored exclusively to filesystems. This use case does not
require an instruction debugger for every asset's platform.

## 9. Milestones and dependencies

Each milestone ends with a small acceptance report containing exact commands,
fixture/media identities, core/package revisions, observed results and remaining
limitations. Report skipped private-media checks as skipped, never passed.

### M0 — Contracts and inventory

Completed for the initial C64-first scope. The [M0 report](tools/browser/docs/code-m0/README.md)
records the exact baseline limitations, including a core-side proxy for Code work
before the Code UI exists. Its contracts refine the illustrative specification here.

- Audit current adapters, instruction boundaries, disassemblers, checkpoint and
  recording capabilities. Separate legacy prototype functionality from shipped UI.
- Specify debug snapshot/stop semantics, job arbitration and cancellation.
- Inventory Fort/Elite/UW/NFS knowledge and conflicting/stale statements.
- Fix schema naming, identity normalization and representative fixtures.
- Record a baseline for play throughput, pause latency, heap use and Code work.

Exit: reviewed contracts, capability matrix, risk list, and runnable fixture plan.
No platform is marked supported solely because its CPU has a native debugger.

### M1 — Knowledge schema, validator and first package

Completed for the basic profile. The [M1 report](tools/browser/docs/code-m1/README.md)
records supported definitions, exact Fort media verification, generated Memory
labels, browser checks and the intentionally unsupported later-milestone features.

Depends on M0. Implement strict schema and semantic validation, stable references,
release matching, CPU/physical/image locations, basic types, regions, functions,
evidence, and Fort's initial package. Export existing applicable memory labels
from the package rather than keeping a second hand-authored address list.

Exit: positive/negative synthetic fixtures cover bad hashes, dangling references,
invalid fields/ranges, unknown enums, applicability and unsupported operations.
The Fort package validates; unknown media retains generic inspection. Round-trip
or deterministic export proves generated data is reproducible.

### M2 — C64 debug adapter and Code vertical slice

Completed for C64; see [implementation and acceptance](tools/browser/docs/code-m2/README.md).
Documented function entries remain explicitly unverified at runtime; address
breakpoints are available without claiming guided applicability.

Depends on M1. Add consistent boundary snapshots, safe mapped peeks, sufficient
6510 decoding (including the implicated opcode), instruction step, run-to-function,
cancel/budget handling, and the Code workspace with game preview/controls.

Exit: native and WASM fixtures agree on stop-before-execute state; current-PC
resume, branches, interrupt entry, banking, stalls and self-modified code are
tested. View navigation cannot change execution. Actual-browser checks cover
input, tab switching and cancellation; disabled debugging has measured overhead.

### M3 — Structured state and reusable panels

Completed for Fort watches and bounded fixed arrays/records; see the
[M3 report](tools/browser/docs/code-m3/README.md). Elite remains gated;
indirect/parallel-array joining is not claimed by this implementation.

Depends on M1/M2. Add typed arrays/records, occupancy, enum/bitfield decoding,
raw-byte expansion, state-to-memory/code links, and reusable panel layout.
Implement Fort watches and Elite universe slots once its compatibility gate passes.

Exit: live decoded values match backing bytes at the same boundary, empty/unknown
states are distinct, slot reuse works, and panel changes preserve navigation.
Code/Memory share definitions without duplicating live state or execution owners.

### M4 — Read-only tours and interval evidence

Completed for sequential read-only C64 tours and the verified Fort terrain
investigation; see the [M4 report](tools/browser/docs/code-m4/README.md).
The implemented profile uses PC/physical-RAM predicates and current paused starts.
Symbolic selectors, automated preparation and the Elite loader remain gated.

Depends on M2/M3. Implement tour lifecycle, conditions, assertions, budgets,
Continue/free-exploration behavior, layout suggestions, baselines and bounded
activity capture. Integrate reusable pieces from the C64 lesson prototype.
Ship a short Fort investigation and then the longer Elite loader tour when verified.

Exit: scripted stops are reproducible; cancel, missed predicates, stale messages,
media mismatch, divergence and trace overflow produce correct outcomes. Snapshot
diff and actual write activity visibly differ on a write-then-restore fixture.
The user can complete a tour with keyboard controls and a narrow layout.

### M5 — DOS adapter and Underworld renderer tour

Depends on M2/M4. Add x86 registers/modes and instruction stepping, live decoding,
module/overlay resolution, relocation validation and buffer panels. Build the
projection/span/copy lesson incrementally.

Exit: signature/overlay invalidation and self-modifying operands behave correctly;
code, typed state and rendering buffers refer to matching execution context.
Validate real/protected-mode distinctions instead of treating all DOS addresses
as segment-times-16. Support only the modes actually tested.

### M6 — Asset location chains and authoring integration

Depends on M1; can progress independently of M5 once the base contract is stable.
Implement sectors/filesystems/containers/transforms and bounded resolver execution,
decoder registry, dependency links, static validation/export tools and site index.
Exercise Captain Toad, a ROM range and a raw-sector fixture.

Exit: source chains resolve reproducibly, wrong media/members fail clearly,
decompression limits work, and asset dependencies/preview outputs validate.
Document the reverse-engineering authoring workflow and update repository standards
to make knowledge-package updates part of completing new research.

### M7 — 3DO adapter and one NFS loop

Depends on M4; reuse M5 lessons where applicable. Add ARM60 debug capabilities,
task/execution context where needed, car records, loop predicates and selected
recording. Research/compatibility work is explicitly budgeted if existing evidence
does not reproduce in the current browser core.

Exit: one verified simulation iteration is explained, with each selected car
identified and spawn decisions tied to actual conditions. Video-frame and
simulation-loop identities remain distinct. HLE boundaries are labelled.

### M8 — Reproducible experiments

Depends on M4 and validated checkpoint/replay support. Implement transactional
checked edits, original/modified branches, identical input replay, observations,
cleanup and return-to-session. Author the prepared Fort helicopter comparison.

Exit: repeat original runs first prove determinism; failed edits leave no partial
state; original/patched evidence cannot mix; restoring the initial session works.
The lesson reports the observed result, including a failed repair hypothesis.

### M9 — Expansion, migration and release hardening

Depends on accepted slices above; it is not necessary to wait for every optional
platform before releasing C64 support. Migrate more knowledge incrementally,
derive function/asset indexes and applicable write-up tables, add step over/out,
conditional watchpoints and historical writer links where supported.

Exit: release capability matrix, schema compatibility policy, complete packaging,
primary-browser and Safari/Firefox checks, accessibility checks, bounded-memory
soaks and regression evidence for ordinary Play/Render/Memory. Publish only tours
whose runtime capabilities and game compatibility were actually validated.

### Recommended release cuts

- First useful release: M0–M3, C64 Code plus Fort state/functions; Elite slots can
  follow if compatibility work is larger than expected.
- First guided release: M4 with a verified short investigation, then Elite loader.
- Broader knowledge release: M5/M6, Underworld rendering and asset provenance.
- Advanced release: M7/M8, simulation-loop explanations and controlled experiments.

Treat this as a multi-release program, not one large UI change. Estimate dates
after M0 and the first measured M2 vertical slice; the main uncertainties are
core correctness, research reconciliation and replay, not tab construction.

## 10. Validation, performance and authoring workflow

### Test layers

- Media-free schema/resolver fixtures: malformed/ambiguous input, release variants,
  typed layout, paths, noncontiguous extents, transformed data and cyclic refs.
- Synthetic CPU programs: exact pre-execution breaks, interrupt/fault boundaries,
  current-PC resume, changing banks/overlays, self-modification and cancellation.
- Cross-build tests: native/WASM boundary registers, selected bytes, event order,
  checkpoint continuation and deterministic input replay.
- Private-media acceptance: named authentic game scenarios with hashes and explicit
  skips when media is absent. Never commit test ROMs to make CI convenient.
- Browser flows: focus/input, responsive layout, tab/tour transitions, stale reply
  rejection, error recovery, save/load, audio after restore, and ordinary playback.
- Budget tests: event/byte caps, decompression caps, bounded conditions, large
  records, package size, disabled/active debugging overhead and UI latency.

Allocate concrete per-feature memory budgets in M0. Avoid copying an entire game
image or RAM on every refresh; request selected spans, cache immutable data and
virtualize long lists. Prefer range-filtered recording over full tracing. Cap
checkpoint counts and state their retention policy. Full rewind, automatic taint
analysis and perfect inferred call stacks are outside the initial scope.

### Research completion workflow

1. Pin the supplied image/release and derive the finding through repository tools.
2. Add/update the semantic definition, source chain, types and evidence in the
   game's single knowledge file. Leave uncertain fields explicitly unknown.
3. Validate schema/references and actual bytes or decoded assets against the
   matching local media. Add meaningful synthetic/regression fixtures.
4. For runtime knowledge, verify applicability and reproduce observations in the
   intended core. A native finding alone does not certify a browser tour.
5. Update the readable write-up; generate applicable tables/indexes from the package.
6. Re-export browser knowledge and validate consumers against the package hash.
7. Commit the knowledge, tools/tests, narrative and permitted artifacts together;
   record what is inferred, untested or blocked by compatibility.

Do not bulk-convert prose into authoritative fields without verification. Import
existing annotation stores incrementally and retain regeneration commands. When
the package becomes canonical for a field, remove or generate duplicate machine
definitions so later discoveries have one place to be corrected.

## 11. Open decisions and explicit non-goals

Decide in M0: exact shared debug ABI, browser decoder reuse strategy, canonical
media-manifest serialization, per-feature memory limits, checkpoint persistence,
and initial layout breakpoints. Decide after real overlay fixtures: generic module
recognition versus small registered platform resolvers. Decide after M4: whether
declarative tour operations need any additional extension points; arbitrary scripts
are not the default answer.

Not required for the first release: all sixteen cores having instruction control,
a full IDE, unrestricted scripting, arbitrary guest patch editing, automatic
function discovery, unlimited history, automatic explanations generated from raw
assembly, or an exhaustive migration of every game's research. The architecture
must accommodate these games without making every future capability a prerequisite
for a useful first Code tab.
