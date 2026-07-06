// Fort Apocalypse — configuration for the shared 2-D level viewer (site/FORMAT.md).
// The playfield is a horizontal cylinder (meta wrap "x"); the soft-char animations
// need the baked per-tile strategy (repaint in place); prisoners/tanks/mines/the
// enemy helicopter are randomized objectPools re-rolled per toggle, and the movers
// patrol per their engines ($992A tanks, $AABA prisoners, $94D2 mines) via the
// pools' patrol/dirStamps data. Exported by
// "Fort Apocalypse (C64)/extract/cmd/webexport".
//
// objectInfo turns the placed objects into clickable info cards (the shared viewer
// raycasts the objects/pools layer and shows objectInfo[name] on a click, the way
// the Super Mario 64 DS viewer names its actors). The keys match the `name` each
// pool carries in the level JSON, plus "player" for the spawn placement; the prose
// is the traced behaviour, folded from the game's Markdown write-up (Part V).
const objectInfo = {
  player: {
    title: 'Rocket Copter — the player',
    text: 'Your helicopter, and the level’s spawn point. Left/right build a bank that steers it, '
      + 'aims the gun and tilts the sprite; up/down move it directly; gravity pulls it down at the '
      + 'chosen Gravity Skill. Terrain is fatal unless the cell is a legal landing surface — the '
      + 'landing pad, a fuel depot, the walkway floor, or a prisoner — where it bounces gently and the '
      + 'spot becomes the respawn checkpoint. Setting down on a depot refuels; fuel falls slowly in '
      + 'flight and “LOW ON FUEL” flashes at zero. (The title attract mode flies it by replaying a '
      + 'recorded joystick sequence.)',
  },
  prisoner: {
    title: 'Prisoner — one of the “men to rescue”',
    text: 'Eight per level, placed wherever the level builder finds a floor cell with rock directly '
      + 'above (engine $AABA). Each runs back and forth along its walkway. Fly within a few cells to '
      + 'rescue him: he boards, the rescued count rises, and the HUD tally is reprinted. He can also be '
      + 'killed — by shooting away the floor beneath him, or by an explosion or missile — so a stray '
      + 'shot, yours or the enemy’s, can kill the very men you need. Either way he leaves the count, and '
      + 'both level exits stay locked until it reaches zero.',
  },
  tank: {
    title: 'Tank',
    text: 'A character-based enemy, six per level (engine $992A). Three body cells plus a turret that '
      + 'always aims at the player. The six patrol horizontally in lockstep — one shared countdown steps '
      + 'them all together — turning back only at open air or water and driving straight through every '
      + 'other kind of terrain; they respawn at fixed home positions once cleared. Each can launch one '
      + 'homing missile when the player passes within range above it. An explosion character or a missile '
      + 'in its cell destroys it.',
  },
  mine: {
    title: 'Self-Propelled Mine',
    text: 'The manual’s name for the small drones (engine $94D2). They patrol the corridors in numbers '
      + 'set by the Pilot Skill option — 13, 26 or 39 — spawning at random empty cells, flying '
      + 'horizontally and reversing at obstacles. Unlike the tanks they do not respawn once destroyed '
      + 'until the next level. They die the same way everything character-based does: an explosion '
      + 'character or a missile in their cell.',
  },
  'enemy-helicopter': {
    title: 'Enemy Helicopter',
    text: 'Only one is active at a time. After a delay it spawns at a random patrol point — but never '
      + 'within ~34 columns or 8 rows of the player, so it can’t materialise on top of you — then hunts '
      + 'by pure per-axis pursuit: each tick it steps one cell toward the player horizontally, then '
      + 'vertically, testing the cells ahead so it only advances into clear corridor. It banks into its '
      + 'motion (which aims its shots) and fires while on screen, but can’t chase across the cylinder’s '
      + 'wrap. Its only exits are death — flying into terrain, or a player bullet through the VIC '
      + 'collision latch. (This viewer leaves it static; in play it is a pursuit AI, not a patroller.)',
  },
};

export default {
  base: 'public/fort/',
  strategy: 'baked',
  maxNativeFactor: 3,
  minFitFactor: 1, // never zoom out past one cylinder period (objects would repeat)
  hud: (level) => `${level.grid.width}x${level.grid.height} chars · wraps`,
  objectInfo,
};
