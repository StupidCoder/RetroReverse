# disasm/ — Ultima Underworld code-knowledge store

This directory holds the paired `<name>.asm` (generated disassembly) +
`<name>.annotations.txt` (hand-maintained knowledge) files for the UW.EXE code,
following the repository's annotation-store convention.

The x86 toolchain now exists (`tools/x86` + `tools/cmd/codetracex86`), so the
store can be populated. Regenerate a listing from the MZ load module (skip the
12,800-byte header, base 0) with, e.g.:

```
go run retroreverse.com/tools/cmd/codetracex86 \
    -skip 0x3200 -base 0 -entry EC50 \
    -annotate "Ultima Underworld (PC)/disasm/uw.annotations.txt" \
    -o "Ultima Underworld (PC)/disasm/uw.asm" \
    "Ultima Underworld (PC)/game/UW.EXE"
```

The near-static trace from the DOS entry (`$EC50`) reaches the C-runtime startup
but stops at the indirect/far handoff into the game; a committed `.asm` covering
the whole engine waits on the Part II execution-core oracle (see
`../Ultima_Underworld.md` Part II §3), which will supply the real entry points to
seed `-entry`/`-table` with. The annotations file's own header should record the
exact regen command, as with the other games' `disasm/` stores.
