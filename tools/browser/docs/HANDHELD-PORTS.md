# DS and 3DS C++/WASM ports

Both ports join the static Pages release, shared local-media flow, pause/reset,
keyboard/on-screen/gamepad input, subsystem timings, portable states and pixel
inspection / rendering replay. They require no server or separately installed
firmware. Generated console art and its prompts are in `ARTWORK.json`.

## DS evidence

- C++ and Go agree at every checkpoint through 1,200 cold-boot display frames:
  ARM9 and ARM7 registers, mapped main RAM and both screen hashes.
- A second 700-frame reference run with stylus presses at frames 320–340 and
  420–440 also matches at every frame. The WASM run checks CPU/screen hashes,
  state continuation, final replay pixels, arbitrary seeks and live-state isolation.
- Node/WASM took 110.1 s for that interactive-title workload, including state and
  capture checks. Cold-boot 1,200-frame validation took 186.9 s. The heap remained
  at 64 MiB during ordinary emulation. These are throughput tests, not browser FPS.
- A representative title capture contains 397 events, 342,988 writes and 1,340
  replay steps with no overflow. State size is about 6.8 MB.
- Chromium: local cartridge and state import, animated title, stylus Adventure
  selection, new-file selection and Peach's opening letter were exercised.
  Roughly 3–4 display updates/s while other development checks were running.
  A later scene captured in 0.62 s with 3,915 rendering steps; clicking a pixel
  reconstructed its exact display color and jumped to its composition step.
- Existing Go limitations remain visible, including incorrect toon shading in
  some scenes. Audio, display capture, fog and edge marking are incomplete.
  This port is not a claim of compatibility with every NDS game.

## 3DS evidence

- The original ARM11/Horizon/PICA core compiles as ordinary C++20 and WASM.
  Shared source coverage includes the DSP HLE mixer, but not browser audio output.
- C++ and Go CPU registers, CPSR and instruction count match all 181 checkpoints
  in the Super Mario 3D Land reference run. Rendering is deliberately **not claimed
  bit-identical**: the first difference at checkpoint 58 is one channel byte out
  of 384,000 (red 0 versus 8). Enabling native floating-point contraction makes
  that frame match Go; the browser build retains strict, uncontracted arithmetic.
  Later image hashes can differ as well. The port preserves the original model,
  not physical PICA accuracy.
- Node/WASM cold boot reached display 60 in 70.3 s. Ordinary execution used about
  738 MiB of WASM memory, including the 512 MiB cartridge. UI presentation and
  input yielding are additional costs. This is an educational desktop build.
- A 73,788,113-byte state restores deterministically. Three subsequent frames
  match the uninterrupted run. The mounted cartridge and read-only filesystem
  backing stay outside the snapshot.
- A GPU-active capture contains 4,452 events, 2,342,966 writes, 8,632 replay steps,
  no overflow, and about 359 MiB of evidence. Replay's final image matches every
  byte, arbitrary seeks work, and replay does not alter the serialized live state.
  Peak heap at that test's capture checkpoint was about 1,588 MiB.
- Capturing three display intervals includes the delayed front-buffer producer.
  A welcome-screen check records 11,186 events, 2,909,047 writes and 24,040 replay
  steps without overflow. All four sampled top/bottom pixels link to display
  transfer commands; final replay matches exactly. Heap is about 1,496 MiB.
- Chromium: local cartridge loading, browser state download/import and the animated
  Super Mario 3D Land title/welcome dialog were exercised. The title ran around
  0.7–1.8 modeled display updates/s; a GPU draw can occupy a worker call for about
  a second. Main-page controls remain on the separate UI thread.
  A browser capture spanning displays 316–319 completed in 5.36 s with 23,822
  replay steps and a 1,637 MiB WASM heap. Pixel inspection matched the display,
  followed a GX transfer back through a memory fill and PICA writes, and sought
  to the transfer step.
- Capture records PICA register writes, raster effects/rejections, GX fills/copies
  and display transfers. A transfer's source links to its historical render-buffer
  pixel. Texture-sampler ancestry remains future work. Regions are packed in the
  recorder; the interface labels this instead of presenting offsets as physical RAM.
- CCI/NCSD parsing accepts unfamiliar decrypted cartridges, up to 1 GiB, subject
  to implemented services and GPU features. Encrypted files and CIA packages are
  not supported. No title-specific bypass is added by this port.

The shared public check suite includes media/input/state containers, display color
and coordinate mapping, replay, existing four cores, and new DS/3DS rendering
regressions. Private images/checkpoints remain local; only measured JSON summaries
are committed in `tools/browser/results`.
