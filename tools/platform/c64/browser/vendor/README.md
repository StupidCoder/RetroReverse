# Hardware component provenance

Source: https://github.com/floooh/chips
Pinned commit: `9e88298ce56319953ac7a43213a1120359f7a3a6`
License: [zlib/libpng](LICENSE), Copyright (c) 2018 Andre Weissflog.

These headers implement the CPU, CIA, VIC-II, SID, keyboard, memory map and C64 board. They are C headers compiled as C++20. Their authorship is retained; they are not claimed as an original RetroReverse rewrite.

Local portability edits are marked inline:

- `chips/m6569.h`: cast the framebuffer `void*` to `uint8_t*` for C++.
- `chips/kbd.h`: explicitly convert the sticky-key time expression to `uint32_t`.

Clang's C compound-literal/address-of-temporary extension is used for the upstream constructors. Native and Emscripten use the same source and flags.

The upstream datasette implementation is **disabled**. RetroReverse validates the complete TAP stream itself and schedules one CIA FLAG event after each supplied interval, freezes the countdown when the motor is off, and guards the end of the buffer. This avoids upstream c1530's first-edge/countdown/end-boundary behavior and supports TAP v0/v1 deliberately. No tape byte decoder exists in the machine.

The included c1530/c1541/m6522 headers are required by c64.h's declarations; floppy support is disabled in the descriptor. ROMs are not included.


C4 also adds optional `RR_*` observation hooks in `chips/m6569.h` and `systems/c64.h`. With no observer they compile to no-ops. They report actual video fetches/shifter output/pixels and CPU writes, without replacing device behavior. The VIC write hook runs inside `_m6569_write`, after that cycle's rendering, to preserve the correct historical register version. CPU RAM/color hooks run before mutation to retain the old byte. CIA2 output-pin latency is handled in the observer. The unchanged authentic C3 schedule and targeted UBSan tests check that these hooks do not alter machine output.
