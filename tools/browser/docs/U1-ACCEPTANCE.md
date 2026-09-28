# U1 — shared static shell

Implemented 2026-09-28. Existing Studio gains one Emulators link; its asset routes and content remain intact. `/emulators/` contains four generated console illustrations and four platform routes with local file selection, shared display/transport, keyboard/on-screen controls and compatibility text. C64 firmware is packaged with checked hashes. Only the selected core is imported. The existing no-build Pages deployment can serve the checked-in JS/WASM artifacts.

Checked: JS syntax; packaged artifact sizes; all three hosted firmware hashes; Chromium landing/navigation/artwork; C64 local File import through the real browser picker and reset execution with automatically fetched firmware. Fixed the artwork's inherited HTML height during visual inspection. Artwork is generated illustration (not a hardware reference photograph); prompts and generation method are in ARTWORK.json.

This milestone packages existing execution behavior. U2 still owns generic boot/profile/format changes; U3 owns subsystem timing, unified physical gamepad handling and execution latency. The current 3DO boot retains the prototype profile until U2. Save-state files and provenance are deliberately shown as upcoming rather than working controls.
