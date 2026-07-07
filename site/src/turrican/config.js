// Turrican — configuration for the shared 2-D level viewer (site/FORMAT.md). The data
// (32px tiles with the explicit hflip bit, 4x4 per-tile collision with the class
// legend, object placements resolved to sprite-atlas keys) is exported by
// "Turrican (Amiga)/extract/cmd/webexport".
//
// objectInfo turns the placed objects into clickable info cards: the shared viewer
// raycasts the Objects layer and, on a click, shows objectInfo[pick.name] plus the raw
// ids (placement type, AI handler address, sprite/frame-table key) the card always
// prints. Every placed object resolves through the scroll-triggered spawner
// (enemy_spawner $1710): a placement type byte picks an AI handler (resident table $1A60
// for type 1/2, else the scene descriptor's +$20 table), and the handler installs a
// frame table — that frame-table address is the object's `sprite` key ("w<world>/ft<addr>",
// an "r" suffix = a resident engine-wide sprite). Keys are the sprite strings; "player" is
// the spawn. Prose is folded from Turrican.md Part V §2–§6 + disasm/resident_core.asm;
// objects our RE doesn't describe get no entry (the card still prints their ids).
const objectInfo = {
  player: {
    title: 'Turrican — the player',
    text: 'The level spawn point. Turrican is not a single sprite but a multi-part composite drawn by a '
      + 'dedicated routine ($56AC inside the player update $5576): it indexes per-animation-state tables '
      + '($4D56 hot-spots, $4E56 part sizes, $4FD6 part lists) by the animation state ($594B & $1F) and '
      + 'draws three body parts ($5950/$5954/$5958) each frame. Orbiting him is the shared spinning energy '
      + 'weapon ($4CCA) — 32 frames of a full rotation plus a three-frame burst. The viewer frames the scene '
      + 'on this spawn: camera tile (desc +$08/+$0A) plus the player on-screen offset (+$0C/+$0E), the same '
      + 'start the Amiga shows.',
  },

  // Resident, engine-wide objects (placement types 1 & 2, resolved through table $1A60) —
  // one entry per world since each is drawn in that world's palette (same handler).
  'w0/ft7E7Er': {
    title: 'Rotating mine (sprite $7E7E)',
    text: 'The little rotating mine — the resident, engine-wide object seeded as placement type 1 in every '
      + 'world (handler table $1A60, AI handler $7D0E, drawn in each world’s own palette). On spawn the '
      + 'handler installs frame table $7E7E and drifts the mine along a fixed built-in offset path (the '
      + 'wobble tables at $7E98/$7ECE), keeping it inside the screen bounds and despawning it once it '
      + 'scrolls off ($1B7E).',
  },
  'w1/ft7E7Er': {
    title: 'Rotating mine (sprite $7E7E)',
    text: 'The resident rotating mine (placement type 1, AI handler $7D0E), here drawn in world 2’s palette. '
      + 'It installs frame table $7E7E, hovers along a fixed built-in offset path ($7E98/$7ECE) and despawns '
      + 'when it scrolls off screen.',
  },
  'w2/ft7E7Er': {
    title: 'Rotating mine (sprite $7E7E)',
    text: 'The resident rotating mine (placement type 1, AI handler $7D0E), drawn in world 3’s palette. '
      + 'It installs frame table $7E7E, drifts along a fixed built-in offset path ($7E98/$7ECE) and despawns '
      + 'off screen.',
  },
  'w3/ft7E7Er': {
    title: 'Rotating mine (sprite $7E7E)',
    text: 'The resident rotating mine (placement type 1, AI handler $7D0E), drawn in world 4’s palette. '
      + 'It installs frame table $7E7E, drifts along a fixed built-in offset path ($7E98/$7ECE) and despawns '
      + 'off screen.',
  },
  'w4/ft7E7Er': {
    title: 'Rotating mine (sprite $7E7E)',
    text: 'The resident rotating mine (placement type 1, AI handler $7D0E), drawn in world 5’s palette. '
      + 'It installs frame table $7E7E, drifts along a fixed built-in offset path ($7E98/$7ECE) and despawns '
      + 'off screen.',
  },
  'w0/ft80E8r': {
    title: 'Power-up box (sprite $80E8)',
    text: 'A floating collectible — the other resident, engine-wide object (placement type 2, handler table '
      + '$1A60, AI handler $7EFC), installing frame table $80E8. When Turrican touches it (collision test '
      + '$438C) the handler grants a reward chosen by the icon’s frame index (+$C) — a weapon type/upgrade '
      + '($544F, stepping $5451/$5452), a bonus item, or a special effect — plays a pickup sound through the '
      + 'resident sound API ($1A2DC) and unlinks the object.',
  },
  'w1/ft80E8r': {
    title: 'Power-up box (sprite $80E8)',
    text: 'The resident collectible/power-up object (placement type 2, AI handler $7EFC), drawn in world 2’s '
      + 'palette. On contact ($438C) its frame index (+$C) selects the reward — a weapon upgrade or bonus '
      + 'item — then it plays a pickup sound ($1A2DC) and vanishes.',
  },
  'w3/ft80E8r': {
    title: 'Power-up box (sprite $80E8)',
    text: 'The resident collectible/power-up object (placement type 2, AI handler $7EFC), drawn in world 4’s '
      + 'palette. On contact ($438C) its frame index (+$C) selects the reward, then it plays a pickup sound '
      + '($1A2DC) and vanishes.',
  },
  'w4/ft80E8r': {
    title: 'Power-up box (sprite $80E8)',
    text: 'The resident collectible/power-up object (placement type 2, AI handler $7EFC), drawn in world 5’s '
      + 'palette. On contact ($438C) its frame index (+$C) selects the reward, then it plays a pickup sound '
      + '($1A2DC) and vanishes.',
  },

  // Per-world scene enemies our RE explicitly describes (Part V §2/§4).
  'w0/ft1E23E': {
    title: 'Rotating eye (sprite $1E23E)',
    text: 'World 1’s rotating eyeball (Part V §2), and the world’s most common placement. It is placement '
      + 'type 3, resolved through AI handler $1E042, which installs frame table $1E23E and cycles the object '
      + 'through its rotation frames.',
  },
  'w0/ft1DF80': {
    title: 'World-1 enemy (sprite $1DF80)',
    text: 'A world-1 enemy driven by AI handler $1DE68. On its first run the handler sets frame table $1DF80 '
      + 'and its health (+$26), then each frame animates an 8-frame loop, applies any damage it has taken '
      + '(+$1D → +$26) and, on death, spawns the destruction burst (JSR $130) — the canonical world-enemy '
      + 'behaviour documented in Part V §4.',
  },
  'w4/ft1DE36': {
    title: 'Red-eyed mech (sprite $1DE36)',
    text: 'World 5’s red-eyed mech (Part V §2), placement type 7 resolved through AI handler $1DBF2, which '
      + 'installs frame table $1DE36. It is one of world 5’s larger enemies; the finer per-frame behaviour '
      + 'beyond the shared world-enemy model isn’t separately traced in our RE.',
  },
};

export default {
  base: 'public/turrican/',
  strategy: 'sliced',
  maxNativeFactor: 4,
  minFitFactor: 0.9,
  hud: (level) => {
    const n = (level.objects || []).length;
    return `${level.grid.width}×${level.grid.height} tiles` + (n ? ` · ${n} objects` : '');
  },
  objectInfo,
};
