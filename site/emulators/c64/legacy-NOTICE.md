# Frozen legacy C64 runtime

The rollback release `9dc547ff2dea4b78b50a` contains a runtime derived from
[floooh/chips](https://github.com/floooh/chips), pinned at
`9e88298ce56319953ac7a43213a1120359f7a3a6`, Copyright (c) 2018 Andre Weissflog.
See [the retained license](legacy-LICENSE.txt). RetroReverse modifications
implemented tape scheduling, inspection, checkpoint serialization and profiling.
The exact modified source remains in RetroReverse commit `d66634b1` under
`tools/platform/c64/browser/{core,vendor}`. It is no longer a build dependency.
New releases compile the independently authored `tools/platform/c64/owned` core.
