# M8 — Reproducible, checked experiments

The C64 Code workspace now offers **Controlled experiments** separately from
read-only tours. The first experiment prepares Fort Apocalypse's underground
pursuit, runs the original twice, then runs a candidate one-byte upward-probe
repair. Setup and patch bytes are visible before Prepare is enabled. Nothing
writes to the tape or silently persists a patch.

## Fort result and what it establishes

The exact reference tape contains `$55 $76` at `$A000`: `EOR $76,X`. Replacing
`$55` with `$A5` makes this `LDA $76`, loading the enemy column directly. This is
a repair hypothesis, not proof of the historical cause. The 16 ASCII bytes at
`$9FF0–$9FFF` end with `Z`; the old write-up's stronger assertion about how the
adjacent signature overwrote the opcode has been corrected.

The prepared checkpoint is at cycle **128519799**, enemy-hunt entry `$9CDA`.
The player is at map (53,23), camera (34,15); the enemy is visible at the native
underground spawn (73,33), in hunting state with cleared sub-cell counters.

Each branch runs **3,931,200 PAL cycles (200 frames)** with the same released
joystick schedule. Observations stop briefly at `$A215`, immediately before the
game evaluates the visible enemy's collision latches. The debugger does not
advance emulated time while paused at an observation.

| Observation | Original, repeated twice | Candidate `LDA $76` |
| --- | --- | --- |
| Complete final state | Identical SHA-256 in both original replays | Different measured state |
| Terrain contact while hunting | +2,063,196 cycles; enemy (69,29), `$73=2`, `$68=0` | Not observed at the probe stops |
| Dying state | First observed +2,082,852 cycles | Not observed in the interval |
| End state | Idle/off-screen at (69,29), after explosion | Hunting/visible at (63,32) |

The original's upward trajectory ends against underground terrain; the actual
sprite/background latch takes the death path. This is not inferred from a
position change alone, and the recorded sprite-contact byte excludes a bullet
collision at that observation. The candidate avoids this death in this one
scenario. Neither the experiment nor its UI assumes every repair succeeds.
[acceptance.json](acceptance.json) records hashes, sample counts and the first
contact/death observations. This validates the current emulated machine; it is
not a new hardware-validation claim.

## Preparation uses original game behavior

The private boot recipe is unchanged from M4. Advance 6,000,000 cycles from its
RLE entry, then stop at the next `$9C52`: reference cycle **124450324**. The
experiment checks gameplay mode, player/enemy/camera coordinates and code bytes.
It then performs these explicit, reversible checked edits:

1. `$9D: 02 → 0A`, `$4D: 00 → 01`: request the game's teleport state and its
   arrival grace flag. Run 4,000,000 cycles and reach the next AI entry. The
   original `$9892` routine updates the camera, scroll and player coordinates.
   Verify player (53,23), camera (34,15), flying state and gameplay mode.
2. `$6E: 03 → 01`, `$43: 00 → 01`: return the off-screen enemy to idle and request
   a native spawn on its next update. Stop at `$9CA7`, after the original RNG,
   spawn table and distance test have selected/accepted (73,33).
3. Run to `$9CDA`; verify native initialization, cleared sub-cell counters and
   visibility. Save the complete prepared machine. No coordinate, sprite,
   camera, RNG or terrain bytes are fabricated.

The setup is deliberately restricted to a verified reference context. A different
RNG/timing context may fail an invariant; it restores the entire original session
and reports the failure. It does not substitute a convenient spawn silently.

## Runtime protocol and safety properties

`Prepare → prepared → Run original twice → original-complete → Run candidate
patch → completed`. During execution the game preview updates and the active
branch is labelled. Instructions/state can be inspected between phases; normal
execution, tours, tape controls, recording and saves cannot interleave with an
owned experiment. Inputs from the live UI are ignored while it owns the session.

- The worker saves the pre-experiment machine **and host input queue** before
  clearing live controls. It retains a second, verified prepared checkpoint.
  Each checkpoint is limited to 4 MiB; a transient branch endpoint is bounded
  by the same limit. Input schedules use exact cycle offsets, splitting work
  slices at each edge. The implemented profile supports joystick port 2.
- Every branch restores the same prepared machine, including devices, RNG and
  CPU phase, then applies the same declared input schedule. Original replay
  determinism compares the SHA-256 of the entire serialized machine and every
  observed value/predicate/offset before the patch button becomes available.
  A mismatch produces an inconclusive failure and restores the original session.
- Native `rr_debug_edit` validates an entire batch before changing any byte.
  Expected bytes, no overlaps, mapped RAM, address bounds and an instruction
  boundary without pending interrupt are mandatory. Ports, ROM and I/O are
  rejected. A patch to the prefetched opcode also replaces the SYNC data byte.
  Failed edits cannot leave a partially applied batch. Later preparation failures
  roll back earlier successful setup stages as part of the full transaction.
- Branch IDs include session generation, experiment serial and original-1,
  original-2 or modified identity. Every observation carries its branch ID and
  relative cycle. Old responses cannot advance a new generation; results never
  imply that original and patched memory belong to one history.
- Cancel during any phase, a failed assertion/edit, wall budget, observation
  overflow or execution error restores the original session and queued inputs.
  **Return to original session** does the same after completion. **Continue
  modified session** explicitly releases ownership while retaining a labelled
  modified branch and the return checkpoint. Returning later is still possible.
  Reset or a replacement image starts a fresh worker from immutable media.
- The worker yields between at most 1,000 cycles, with a 30-second operation
  ceiling. No arbitrary package code, host filesystem writes or network actions
  are evaluated. This first profile supports checked C64 byte edits to RAM/code;
  register edits, other platforms and portable prepared-checkpoint distribution
  remain unsupported and are rejected by the schema.

Observations are bounded samples at a declared execution point, not a continuous
memory trace or a claim that unobserved events never occurred. The UI says
“not observed at probe stops” and shows raw sampled values. A no-op or ineffective
patch still produces a completed comparison with its actual results.

## Package profile

Schema version 1 now accepts up to 16 `experiments`, each with title, description,
requirements, releases/evidence, start predicate, mapped-byte guards, setup
stages, prepared invariants, candidate patch, duration/wall budget, observation
PC, input schedule, byte watches, observation predicates and interpretation.
All are declarative data in `games/fort-apocalypse-c64/knowledge.json`.

An edit is `{kind: "code", address: 40960, before: [85], after: [165],
explanation: "…"}`. `ram` and `code` both require mapped RAM; the distinction
explains intent to the user. Setup stages contain preconditions, checked edits,
a fixed cycle advance, a bounded PC stop and postconditions. Existing bounded
C64 tour predicates are reused. At most 256 bytes may change per transaction;
setup has at most 8 stages and a 15-million-cycle combined budget. Branches are
limited to 5 million cycles, 256 ordered input edges, 64 byte watches, 32 named
predicates and 512 observations. Unknown fields, scripts and unsupported edit
kinds are rejected.

## Reproduce and validate

From the repository root, with the locally owned reference tape:

```sh
python3 tools/platform/c64/browser/build.py --emcc /path/to/em++
python3 tools/browser/package.py --core c64
node tools/browser/tests/experiment-fort.mjs \
  games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  tools/platform/c64/browser/work/m8-experiment.rrstate
python3 -m http.server 8878 --bind 127.0.0.1
```

Open `/tools/browser/tests/code-m8-browser.html`. It exercises the actual shared
app/worker, checks the displayed edit list, preparation and ownership, matching
original hashes, contact/death observations, candidate outcome, branch separation,
return/cancel rollback, narrow controls and keyboard focus. No game checkpoint
or image is committed. Browser save containers require the exact new core hash.

`python3 tools/browser/check.py` includes schema rejection tests, native atomic
edit/prefetch checks, and media-free real-WASM experiment tests. These exercise
partial-edit rejection, cancellation in preparation/original/modified execution,
stale and busy requests, nondeterminism refusal, input/session restoration,
explicit modified continuation, and an ineffective no-op repair. Existing C64,
DOS, 3DO, tour, state and release checks remain in the same suite.
