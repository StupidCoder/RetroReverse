Per-shot actor blocking tables A(f), captured from the running demo by
extract/cmd/actorsolve (-census to find each actor's object + world-matrix
array in a savestate, bootoracle -memrec to record them every field through
the shot, -bake to solve A(f) = world[0]·composed[0]^-1 per frame).

The demo player computes these placements in code every frame; no file on
the disc stores them (the .scd carries only camera and light channels, the
.sco only per-cut bases), so capture-and-solve against our own .key
evaluation is the honest route — the identity solves on the world-space
sets are the control that pins the whole pipeline to the game.
