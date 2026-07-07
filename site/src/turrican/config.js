// Turrican — configuration for the shared 2-D level viewer (site/FORMAT.md). The data
// (32px tiles with the explicit hflip bit, 4x4 per-tile collision with the class
// legend, object placements resolved to sprite-atlas keys) is exported by
// "Turrican (Amiga)/extract/cmd/webexport".
//
// objectInfo turns the placed objects into clickable info cards: the shared viewer
// raycasts the Objects layer and, on a click, shows objectInfo[pick.name] plus the raw
// ids (placement type, orientation byte, AI handler address, sprite/frame-table key) the
// card always prints. Every placed object resolves through the scroll-triggered spawner
// (enemy_spawner $1710): a placement type byte picks an AI handler (resident table $1A60
// for type 1/2, else the scene descriptor's +$20 table), and the handler installs a
// frame table — that frame-table address is the object's `name` key ("w<world>/ft<addr>",
// an "r" suffix = a resident engine-wide sprite; the `sprite` key adds ".<frame>").
//
// ORIENTATION. Each placement entry is `type.w, x.w, y.w`; the spawner uses the type
// word's high byte for the handler and writes its LOW byte into the object node at +$1E —
// the orientation/direction selector (usually $80 | index). The handler reads +$1E on its
// first (spawn-init) run and picks the frame ($C) / direction / position for that
// orientation. Because the BOB blit has no flip, every facing is a *pre-drawn frame* in
// the one frame table (e.g. the shaft cannon $1F68A: frame 0 up, 1 down, 2 right, 3 left;
// the block turret $1F9A4: frame 0 up, 7 flipped down). The exporter (extract/cmd/webexport
// + extract/scene/sim.go) resolves each object's displayed frame/position by RUNNING its
// handler's init in a 68000 interpreter — the same "run the code" method as the music — so
// the viewer shows the facing the Amiga does, not a fixed frame 0.
//
// Keys are the frame-table strings; "player" is the spawn. Prose is folded from
// Turrican.md Part V §2–§6/§8 + disasm/resident_core.asm; objects our RE doesn't describe
// get no entry (the card still prints their ids).
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

  // World 3 (internal world 2) — the machine shaft. These are the enemies whose facing
  // is chosen by the placement orientation byte (Part V §8), verified against gameplay.
  'w2/ft1F68A': {
    title: 'Four-way cannon (sprite $1F68A)',
    text: 'The rotating wall cannon that lines the machine shaft — placement type 10, AI handler $1F4FE. It '
      + 'installs frame table $1F68A and reads the placement orientation byte (node+$1E) on spawn to pick the '
      + 'barrel’s facing: orient $80 → frame 0 (up), $81 → 1 (down), $82 → 2 (right), $83 → 3 (left). The BOB '
      + 'blit can’t flip, so each facing is its own pre-drawn frame; the same cannon appears pointing every '
      + 'way around the shaft. (Before the orientation was decoded these all showed frame 0 — every cannon '
      + 'pointing up.)',
  },
  'w2/ft1F9A4': {
    title: 'Block turret (sprite $1F9A4)',
    text: 'A turret mounted on a destructible block — placement type 11, AI handler $1F7FA, frame table $1F9A4. '
      + 'The orientation byte flips it vertically: orient $80 → frame 0 (barrel up), $81 → frame 7 (a pre-drawn '
      + 'upside-down copy). This is the vertical-flip case — there is no flip bit in the blitter, the mirrored '
      + 'frame is stored in the table and the handler selects it from node+$1E.',
  },
  'w2/ft1EC38': {
    title: 'Shaft rocket (sprite $1EC38)',
    text: 'A rocket that flies across the shaft — placement type 6, AI handler $1EAE0. On spawn the orientation '
      + 'byte sets which wall it launches from and its horizontal velocity (orient $80 → enters at the right '
      + 'edge moving left, $81 → the left edge moving right, ±2 px/frame); the frame table $1EC38 carries the '
      + 'left- and right-pointing rocket frames, and the per-frame AI shows the one matching its travel '
      + 'direction. The viewer places it at its wall entry point.',
  },
  'w2/ft1E8F6': {
    title: 'Helmet turret (sprite $1E8F6)',
    text: 'The shaft’s most common enemy — installed by AI handler $1E740 (reached directly as placement type 5, '
      + 'or handed off from the type-9 launcher $1EA82). It installs frame table $1E8F6 and displays frame 7; '
      + 'the orientation byte sets a spawn flag and nudges its start position, and the handler then runs its '
      + 'drop/gravity behaviour.',
  },
  'w2/ft20A7A': {
    title: 'Directional projectile (sprite $20A7A)',
    text: 'A projectile launched in one of several fixed directions — placement type 8, AI handler $207A6. The '
      + 'orientation byte’s index feeds a jump table ($20868) that sets the movement vector (node+$2E, indexing '
      + 'the per-frame movers at $20878) and the matching frame of table $20A7A, so the shot points the way it '
      + 'travels.',
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
