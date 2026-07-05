# disasm/ — Ultima Underworld code-knowledge store

This directory holds the paired `<name>.asm` (generated disassembly) +
`<name>.annotations.txt` (hand-maintained knowledge) files for the UW.EXE code,
following the repository's annotation-store convention.

## Contents

- **`uw.asm`** — the Microsoft-C run-time startup, code-traced from the DOS entry
  point `$EC50` and reduced to the reachable code (10 routines, 1011 bytes). Its
  header records the exact regen command. Generated with the repo's own tracer:

  ```
  go run retroreverse.com/tools/cmd/codetracex86 \
      -skip 0x3200 -base 0 -entry EC50 \
      -o "Ultima Underworld (PC)/disasm/uw.raw.asm" \
      "Ultima Underworld (PC)/game/UW.EXE"
  ```

  (the raw output is the whole 549 KB image with the code marked; `uw.asm` keeps
  only the traced routines).

- **`uw.annotations.txt`** — the knowledge: those startup routines annotated,
  plus the rest of the startup chain (game handoff, overlay manager, hardware
  detection, video/arena init, input ISRs, the fixed-point renderer's
  divide-error handler) documented with the runtime addresses the execution-core
  oracle pinned. See `../Ultima_Underworld.md` Part II §3 and Part III.

## Extending the store

The static trace stops at the C-runtime's indirect handoff into the game — the
game's own code runs in **relocated and overlay-paged** segments whose raw file
bytes differ from their run-time content, so it cannot be traced statically. To
disassemble a resident routine or a paged overlay correctly, capture it from the
oracle's live memory instead:

```
go run ./cmd/bootoracle -game ../game -irq -dis SEG:OFF:LEN   # from "<Game>/extract"
```

Future `resident_core.asm` / `overlay_*.asm` listings for UW should be captured
that way and their regen command recorded in the matching annotations file,
following the Turrican `disasm/` store's resident-vs-overlay split.
