# M9 release: capability, compatibility and acceptance

M9 adds C64 step over/out, conditional write watchpoints, historical writer
navigation and a generated function/asset reference. It expands Fort Apocalypse
with three functions already established by the M8 experiment: native teleport,
accepted enemy spawn initialization and terrain-contact handling. No new game
compatibility or undocumented state fields are inferred from this migration.

## Published capability matrix

| Platform | Code / instruction step / run to | State and guided investigation | M9 advanced controls |
| --- | --- | --- | --- |
| C64 / 6510 | Mapped bytes; explicit partial-instruction normalization; bounded execution | Exact reference Fort image; short upward-probe tour and checked original/candidate experiment | JSR step over; balanced JSR/RTS step out; masked RAM writes, optional writer PC; recorded writer links |
| DOS / x86 | Real 16-bit and supported flat protected 32-bit; live module guards; atomic instruction budget | Verified Underworld file set and rendering tour; live buffer only in guarded context | Recorded writer links; no step over/out or conditional native watchpoints |
| 3DO / ARM60 | ARM32 big-endian; explicit scheduler/HLE boundaries; task-bound named stops | Verified Need for Speed image, code/car-layout guards and simulation-loop tour | Recorded writer links; no step over/out or conditional native watchpoints |
| 3DS | No Code workspace | Captain Toad static asset provenance and curated exports | None |
| Other packaged systems | No Code workspace | Existing system-specific Play/Render/Memory tools | None |

All sixteen existing emulator routes remain packaged. This table is about the
knowledge debugger, not a claim of complete emulation or feature parity. Elite
object slots and fast-loader tours are still research proposals, not published
tours. Tour applicability requires both the package's release identity and the
runtime's supported guards/actions. Unsupported contexts fail closed.

### Advanced control semantics

- Step over executes one non-JSR instruction, or follows nested calls until the
  original JSR returns to its sequential PC with its original stack pointer and
  address mapping. Interrupt handlers are tracked separately.
- Step out reads the current stack's return address and requires a mapped RAM JSR
  at its call site. Use it at a balanced subroutine frame, before arbitrary pushes
  or after those pushes have been undone. It does not reconstruct a call stack;
  manipulated stacks, banked call sites and unusual return conventions may be
  rejected, fail a return-context check or exhaust the bounded run budget.
- A write watchpoint tests an actual mapped-RAM bus write, including the dummy
  same-value write in read/modify/write instructions. `(byte & mask) == value`;
  mask zero matches every value. An optional writer PC narrows it further. Ports,
  ROM and I/O targets are excluded. A mapping change terminates the job.
- Watchpoints can stop mid-instruction. Normalize explicitly before stepping.
  The result contains one address/value/writer/cycle event, not an unbounded log.
- Memory and watchpoint writer links inspect **current** code bytes at a recorded
  PC. Self-modifying or remapped code may differ from the original write. These
  links do not implement reverse execution or reconstruct historical instructions.
- The same machine-ownership, stale-snapshot, cycle/wall limits and cancellation
  rules apply to all new commands. The ordinary `rr_run` path is unchanged.

## Schema and release compatibility policy

The source of truth remains each game's single `knowledge.json`. Source and
schema documentation live in `tools/knowledge`; browser data and
`site/knowledge/index.html` are deterministic derivatives. The latter now derives
function locations, applicability, release IDs and evidence as well as the asset
provenance tables. Regenerate with `tools/browser/package.py`; CI rejects stale
output. Do not hand-maintain duplicate function/address tables in write-ups.

The current knowledge schema is version 1. Increment `revision` for research or
content changes. A revision is not a claim of runtime compatibility. Additive
optional sections may remain in v1 only when their absence preserves prior
semantics and validation can reject unsupported actions. Consumers must not
ignore unknown executable operations or guess their meaning. A changed address
meaning, required-field contract or operation semantics requires a schema-version
change and explicit migration; the current validator's strict-field rejection
means an older consumer may reject a newer package rather than silently degrade.
Keep original sources/evidence and stable IDs through migrations, record the new
revision, and revalidate release matching, resolvers, references and generated
output. Never migrate a hypothesis into a confirmed fact merely to fit a schema.

Publish generated JS, matching WASM/firmware hashes and route HTML as one
content-addressed release. Packaging retains the prior executable release so
already-open pages do not mix cores. Portable states remain bound to exact core,
media, firmware and configuration identities: M9's changed C64 binary does not
load an older core's state by guessing compatibility. Recreate checkpoints using
the current core; no commercial media or private save states are shipped.

Adding a new tour requires native/runtime tests, exact local-media acceptance,
bounded trace/checkpoint evidence, cancellation/restore checks, and browser
acceptance for the published slice. General schema support alone is insufficient.
M4/M6/M7/M8 retain their game-specific evidence and limits; M9 rechecks the Fort
experiment against the rebuilt C64 core.

## Reproduce acceptance

Run `python3 tools/browser/check.py` with the repository's Go/C++/Node toolchains.
The public suite includes nested C64 JSR/RTS with an intervening IRQ handler,
masked actual writes and invalid conditions, stale/busy/cancelled jobs, all
existing system regressions, schema/export checks and release integrity.
`debug.mjs` also executes 2,000 complete call/return cycles with snapshots after
warm-up and asserts that the WASM heap does not grow. This checks retained WASM
allocations, not a claim that browser process RSS cannot fluctuate. Existing
Memory/tour/experiment tests exercise trace caps, overflow, checkpoint budgets,
rollback and bounded observation counts.

Serve the repository on localhost and open
`tools/browser/tests/code-m9-browser.html` in Chromium, Safari and Firefox. The
fixture makes its own one-pulse TAP and call-loop RAM, wraps a core checkpoint in
the normal portable-state format, and loads the actual content-addressed app in
an iframe. It checks Code controls, actual-write event links without execution,
ordinary Play run/pause, Memory atlas/recording/writer navigation, Render capture,
375px layout, form labels, focus and the named disassembly region. No game image
is needed. The optional `?report=chrome|safari|firefox` posts JSON to a test-only
local collector if supplied; the production app has no such endpoint.

These are focused accessibility checks, not a complete screen-reader/WCAG audit.
The narrow-layout checks found and fixed a full-width Code preview and grid
minimum-width overflow. The disassembly now has explicit readable foreground
color on its dark background.

## Recorded result (2026-09-30)

Release `d3ac0a286899e5625481` passed the fixture in Chromium 152 (Codex's
in-app browser), Safari 26.5.2 and Firefox 157 (official Mozilla build in a
throwaway headless profile). Reports are checked in beside this document. Each
reported a 134,217,728-byte WASM heap. Firefox's headless compositor emitted a
software-framebuffer warning; the functional and capture assertions still passed.
The native/WASM tests and complete public browser regression command passed.
The rebuilt C64 core also passed the private Fort M8 experiment: two original
replays matched, terrain contact and death were observed, the candidate survived
the bounded interval, and original core/host state restored exactly. The compact
report includes hashes and observations only, not game bytes or checkpoint data.

Prior game-specific evidence:
[M4](../code-m4/README.md), [M6](../code-m6/README.md),
[M7](../code-m7/README.md), [M8](../code-m8/README.md).
