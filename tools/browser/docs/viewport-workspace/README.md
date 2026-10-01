# Viewport workspace

Play, Render, Memory and Code are now layouts over one machine session. The Loader preset places Game / Session above Storage / Memory atlas.
Guided tour uses separate Code, lesson, RAM atlas and Storage panes. See the
[Storage follow-up](../storage/README.md) for supported media formats and the
[prepared lessons](../prepared-lessons/README.md) for the verified Elite prefix.

## Interaction

The shared toolbar selects the system and layout, loads media and controls
execution. Play places Game on the left (70%) and Session / controls on the right.
Load game, Run, Pause and Next frame use icons with accessible names and tooltips. Load game opens the appropriate file/folder and firmware form.
Session exposes save/load state, tape controls, keyboard/gamepad help, on-screen
buttons, compatibility and performance information.

Each pane has a content selector and Layout menu. Split horizontally or vertically
(up to four panes), resize with the divider, maximize/restore, or close a pane to
merge its space with its sibling. Dividers support arrow keys. Each preset keeps
its customization in local storage; Reset layout restores its default. Layout
changes do not execute, pause, reset or reload the machine. Narrow windows show
one pane with a viewport chooser while preserving the desktop split tree.

Game, Code, RAM atlas, byte view, tape pulses and State support independent
instances. Code panes keep their addresses and selections; memory panes keep
separate regions/offsets; state panes keep separate pins and disclosures. Game
panes share output but each supports keyboard, pointer and gamepad focus. Code
snapshots are owned by inspector identity, so refreshing one pane does not expire
another pane's unchanged stopped context. Real machine changes still invalidate
execution requests. At most 64 inspector contexts are retained.

Rendering's preset links the captured output, source and detail views to the same
capture cursor. Additional Rendering/source/detail inspectors have their own UI,
namespaced DOM, request IDs and cursor selection over the session's capture.
They do not create extra emulator workers. The capture itself remains a single
bounded session resource and is replaced when a new capture is requested.
Lesson/experiment, Memory recording and Session are session-owned control panels;
the content picker prevents duplicating those controllers within a layout.
The Memory recording timeline is shared; live memory panes identify whether they
are showing current storage or historical recording data.

The State pane uses compact rows. Pin important values, expand records/arrays,
and select a row for raw bytes and research evidence. Values update in place to
preserve focus and selection; changed sampled values are highlighted. Addresses
and related functions navigate other panes. Sampling still does not imply that
all transitions between snapshots were observed.

## Session and resource boundaries

Only one emulator worker remains active after a system switch. Switching saves a
paused checkpoint, releases inputs, terminates the worker and mounts the selected
system. Returning restores the checkpoint with the existing media/core/firmware
identity checks. Locally selected File objects and checkpoint blobs are retained
only for the lifetime of the page; this is not persistent game storage. Explicit
Save state remains available. The suspension budget is 256 MiB (excluding the
active core); exceeding it refuses the switch rather than evicting progress.
Switches are disabled while a capture, save, recording or experiment owns the
machine. Core loading stays lazy and browser-cacheable. Direct system URLs remain
valid and update when switching in the top-level app.

The memory feed samples once per session, not once per pane. Hidden panes stop
sampling. State watches have a shared sampler independent of Code visibility.
Tape windows expose at most 256 pulse lengths; the UI plots 128. A tape panel can
follow the read head or inspect a fixed pulse index. It displays pulse lengths
in CPU cycles; the atlas remains a separate view of physical RAM.

Layout persistence stores panel types, split ratios and IDs only. It contains no
media, RAM bytes, checkpoint or private game state. Parked panels retain their
navigation and remain addressable by older specialized adapters; disposing a
session removes listeners, timers, workers and observers.

## Acceptance

- `node tools/browser/tests/viewport-model.mjs`: preset validity, splits, merging,
  IDs and persisted-layout validation.
- `node tools/browser/tests/debug.mjs`: native WASM debugging, stale/busy/cancelled
  jobs, bounded heap and independent inspector contexts across 100 peer refreshes.
- `python3 tools/browser/check.py`: existing public schema, native, WASM, input,
  memory, rendering and release checks, plus the viewport model.
- `python3 tools/browser/tests/serve-viewport.py`: serve the repository on
  localhost port 8767 and collect optional test reports in `/private/tmp`.
  Open `tools/browser/tests/viewport-browser.html?report=chrome` (or `safari`,
  `firefox`) for the media-free synthetic acceptance fixture. It exercises
  duplicate Code and Rendering panes, the loader preset, split/merge/keyboard
  resize, ordinary playback, rendering capture, system suspend/restore, unique
  DOM IDs and laptop/mobile layouts.
- Private Fort acceptance: generate a current checkpoint with
  `node tools/browser/tests/experiment-fort.mjs games/fort-apocalypse-c64/Fort_Apocalypse.tap tools/platform/c64/browser/work/viewport-experiment.rrstate`,
  then open `tools/browser/tests/viewport-fort-browser.html`. No image or state is
  included in the release. This verifies the existing experiment controls,
  original/modified branches and return/cancellation within the viewport shell.

Prepared lesson starts are a separate follow-on: this workspace does not publish
or cache authored lesson checkpoints. The existing prerequisite checks remain.

## Recorded validation — 2026-10-01

Final release `28a1e43e72bd7360457b` passed the complete viewport fixture in
headless Chrome 154 and Firefox 157. Chrome also passed the private Fort
experiment, including deterministic original replays, the modified branch,
return, cancellation rollback and keyboard focus in the narrow Lesson pane.
Reports are adjacent JSON files and contain no game image or checkpoint.

The full public suite passed during implementation (`public-check.log`, release
`3960b2a78acbc40321df`). After the final UI refinements, viewport-model, debug,
memory, ui-shell and release checks were rerun successfully against the final
sources/package, together with both browser fixtures. All sixteen core hashes
are unchanged. Release rotation retained the prior release unchanged, removed
only its predecessor and verified all 95 pinned asset hashes.

Safari and final manual visual acceptance were not completed: computer use was
blocked by the locked Mac. Earlier visual inspection informed the compact pane
header and embedded Code spacing fixes; the final automated fixtures check
1280×780 bounds, unique IDs and the 375px single-pane presentation.

## Compact toolbar follow-up — 2026-10-01

Release `9d4abc47ebb6c77fbf4c` passed the viewport browser fixture in Chrome 154
and Firefox 157, plus the shell, layout-model and pinned-release checks. The
in-app browser confirms Play exposes Game, Session / controls and a 70% divider,
with layout selection and labelled transport icons in the top bar. Existing
customized layouts are preserved; the previous single-Game default is migrated.
