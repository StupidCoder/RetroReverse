# Presentation pacing and C64 tearing

The old worker batched 8 ms of execution, regardless of how much emulated time
that covered, and only sent a display when `now - lastPaint > 80`. The latter
alone capped visible updates at about 12.5 per second. A fast core could also
run several guest frames in one batch and then sleep to repay the accumulated
emulated-time lead. Thus nominal emulation speed did not imply smooth output.
C64 frames were copied at arbitrary execution points, while its VIC was still
writing the framebuffer: parts of adjacent raster frames could appear together.

Normal execution now stops at every declared display boundary, presents that
completed buffer and paces against emulated time. Fast-forward also stops at
boundaries, selecting completed frames for a roughly 60 Hz presentation cap.
The main thread coalesces updates on requestAnimationFrame, transfers pixel
buffers without an extra structured-clone copy, and changes canvas dimensions
only on a real resolution change. Slow cores still yield between bounded slices;
missed real-time deadlines do not accumulate unlimited catch-up credit.

The UI now shows presented frames/s independently of emulated cycles/fields.
Restored states and Run reset measurement baselines rather than counting the
restored historical frame number as newly executed work. Initial load duration
is reported separately. Paused capture work remains outside normal throughput.

Observed in Chromium on this Mac while the four soak tests were also active:

| Case | Emulated progress | Presented frames/s |
|---|---|---|
| Fort Apocalypse, normal | 100% PAL, about 50.1 raster frames/s | about 50.0 |
| Fort Apocalypse, fast-forward | about 207% PAL in this sample | about 52.0 |
| Ridge Racer, normal | 100%, about 59.8 synthetic fields/s | about 60.0 |

Presentation rate does not mean a game produces a different picture every
field. Ridge Racer may render at a lower game frame rate; synthetic PS1 timing
is still an emulator approximation. The structural C64 mixed-raster source is
removed by boundary-only presentation; this is not a claim of physical-display
scanout instrumentation or exhaustive hardware conformance.

A related inspection issue was double buffering: a single PS1 synthetic field
could write only the non-displayed buffer. PS1 Pause now captures four synthetic
fields, explicitly including producer context for the final displayed image.
Three Ridge Racer captures now verified 207, 223 and 202 historical texture/copy
source queries across the sampling grid, with correct reconstructed colors and
no trace overflow. The capture retains the same bounded write budget. Other
platforms retain their one-complete-interval capture.
