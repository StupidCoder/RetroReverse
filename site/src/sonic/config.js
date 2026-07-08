// Sonic the Hedgehog (Game Gear) — configuration for the shared 2-D level viewer
// (site/FORMAT.md). The richest data set: block-indirected tilemap, ring/flower tile
// animations, the $50 background cell animators, animated object sprites with
// movement paths (index.json steps/durations/path), height-profile collision
// (shapes.json) and the palette effects (water/lava cycle + the Labyrinth underwater
// split). Exported by "Sonic (GG)/extract/cmd/webexport" + cmd/spriterip.
//
// objectInfo turns the placed objects into clickable info cards: the shared viewer
// raycasts the Objects layer and shows objectInfo[pick.name] on a click (pick.name = the
// object's `name`, the objNames map in the webexport), plus the raw ids (object type,
// sprite key) the card always prints. Prose is folded from Sonic.md Part V §1 "Objects".
// Unnamed placement types have no entry — the card still prints their "type $XX" id.
const objectInfo = {
  player: {
    title: 'Sonic — the player',
    text: 'The level spawn point (object type $00). At load the engine settles him onto the ground at this '
      + 'position; from there his momentum-based run/jump/roll physics take over — the subject of Part V. '
      + 'The viewer shows him as a static placement.',
  },

  // --- Enemies ---
  crab: {
    title: 'Crab (type $08)',
    text: 'An enemy walker (~0.16 px/frame) that periodically stops to fire a projectile to each side; four '
      + 'appear in Green Hills Act 1 (handler bank 1 $65F9). It runs a four-step claw-walk animation '
      + '(frames 0,1,2,1 held ~13 frames each) through the shared $7C75 routine, and starts out walking '
      + 'toward the player.',
  },
  beetle: {
    title: 'Beetle (type $10)',
    text: 'An enemy walker that marches back and forth at 1 px/frame with no attack, driven by the same '
      + 'script engine as the crab (handler bank 1 $6E65). Its two-frame walk (frames 0,1 held ~9 frames '
      + 'each) plays through the shared animation routine. Its 16-px height happens to equal its block’s '
      + 'surface offset, so it sits exactly on the ground where it is placed.',
  },
  bird: {
    title: 'Bird (type $0E)',
    text: 'A flying enemy (handler bank 1 $6BD9). It plays a fast two-frame wing-flap (frames 0,1 held ~3 '
      + 'frames each) through the shared $7C75 animation routine. Beyond that flap our notes do not detail '
      + 'its flight path.',
  },
  fish: {
    title: 'Fish (type $26)',
    text: 'An enemy that leaps 128 px (four blocks) straight up and back down on roughly a 2.6-second cycle '
      + '(handler bank 1 $7D25).',
  },
  porcupine: {
    title: 'Porcupine (type $2D)',
    text: 'A spiky walker (handler bank 2 $82FB) that patrols about 160 px each way at 0.25 px/frame with no '
      + 'attack. Its handler has gravity, so where it spawns over a gap (Bridge Act 1, at 480,320) it drops '
      + '48 px onto the lower ground the moment it activates. It starts facing right.',
  },

  // --- Platforms ---
  'swing platform': {
    title: 'Swinging platform (type $09)',
    text: 'A pendulum platform that carries Sonic (handler bank 1 $6747): it swings a 180° arc of about '
      + '51-px radius on roughly a 3.7-second cycle. Its metasprite layout is chosen by zone ($6910 in '
      + 'Green Hills, else $6922).',
  },
  'sinking platform': {
    title: 'Sinking platform (type $0B)',
    text: 'A platform that gives under Sonic’s weight (handler bank 1 $69ED): it sinks 16 px at ½ px/frame '
      + 'while he stands on it, and floats back up when he steps off.',
  },
  'moving platform': {
    title: 'Horizontal moving platform (type $0F)',
    text: 'A back-and-forth platform that carries Sonic (handler bank 1 $6DCA). It glides at exactly '
      + '1 px/frame with no sub-pixel fraction, keeping a 16-bit phase counter and a direction byte; when '
      + 'the counter reaches $A0 = 160 it resets and flips sign — a symmetric 160-px (5-block) out-and-back '
      + 'that reverses every 160 frames. In the Jungle this same slot is instead a floating log.',
  },
  'bobbing platform': {
    title: 'Bobbing platform (type $3B)',
    text: 'A free-floating platform with no terrain contact (the Bridge’s floating logs). Its Y-velocity '
      + 'ramps ±$10 sub-px/frame on a 160-frame phase, clamped to ±2 px/frame: from its spawn it sinks about '
      + '160 px, then bobs in a 96-px, 160-frame cycle. Sonic can ride it (via $7CF5).',
  },
  'floating log': {
    title: 'Floating log (type $29)',
    text: 'The Jungle’s rideable log (handler bank 2 $7EFC). It floats on the water — spawning lifted 24 px, '
      + 'then a gentle 5-px bob on a 40-frame cycle. Ridden, it becomes a log-roll: Sonic steers it at half '
      + 'his speed (the handler moves the log by vel/2 and writes the log’s X back to Sonic), the roll phase '
      + 'picking one of three rotating-grain layouts, and it stops against solid terrain probed beside it.',
  },
  seesaw: {
    title: 'Seesaw (type $4E)',
    text: 'A tilt-arm catapult (handler bank 2 $8681): its launch height scales with Sonic’s landing impact, '
      + 'transferring his downward momentum into the launch.',
  },

  // --- Items / pickups (item monitors) ---
  bonus: {
    title: 'Bonus monitor (types $01–$03)',
    text: 'An item monitor (TV), part of the pickup family (handlers bank 1 $5DE1/$5EB1/$5EDD). Each lazily '
      + 'streams its own 16×16 screen icon into sprite tiles $5C–$5F ($01 = $15200, $02 = $15280) over the '
      + 'common item-box base. Its “blink” is actually a sprite-per-scanline budget trick — the game drops '
      + 'an icon-covered cell most frames — so it renders and plays as static.',
  },
  shield: {
    title: 'Shield monitor (type $04)',
    text: 'The shield power-up monitor (handler bank 1 $5FAF), part of the pickup/item-box family. Like the '
      + 'other monitors it composites a streamed 16×16 icon over the common item-box base tiles.',
  },
  emerald: {
    title: 'Chaos Emerald (type $06)',
    text: 'A Chaos Emerald pickup (handler bank 1 $6183). Like the other monitors it streams its own 16×16 '
      + 'icon (from $15480) into sprite tiles $5C–$5F over its item-box base.',
  },
  checkpoint: {
    title: 'Checkpoint (type $51)',
    text: 'The mid-act checkpoint (handler bank 1 $6010). On contact it writes its own block position into '
      + 'the respawn table at $D32F + act×2 (the $6034 respawn-save code), so Sonic restarts here after a '
      + 'death.',
  },
  'continue': {
    title: 'Continue power-up (type $52)',
    text: 'A special-stage pickup, tagged in our objNames map as the Continue power-up. It belongs to the '
      + 'icon-overlay pickup family; our notes list $52 among the types whose behaviour is not yet '
      + 'reverse-engineered beyond the name.',
  },

  // --- Level fixtures ---
  goal: {
    title: 'Goal sign (type $07)',
    text: 'The end-of-act goal sign (handler bank 1 $61F8). On its first frame it loads its own sprite sheet '
      + 'over VRAM $2000 (bank 9 file $27AB8) plus palette $0E, and clamps the camera ($D26D/$D26F) so the '
      + 'act ends at the sign. It idles static on the “?” plate; when Sonic crosses it, it hops and spins '
      + '(plates 0→3→2→4 × 6 frames each) and stops on the outcome plate.',
  },
  capsule: {
    title: 'Capsule (type $25)',
    text: 'The animal-prison capsule that ends each world: Sonic jumps on it to free the animals (handler '
      + 'bank 1 $736B). It draws with the boss’s runtime tiles (own-graphics), so the extractor skips it and '
      + 'the viewer shows a labelled marker instead of an extracted sprite.',
  },
  teleporter: {
    title: 'Teleporter (type $13)',
    text: 'Named from play-testing as a teleporter. Scrap Brain and Sky Base carry teleporter-only '
      + 'sub-scenes (engine zone 7) that our loader lists as hidden scenes; this object is understood to '
      + 'warp between them, but its exact behaviour is otherwise not detailed in our notes.',
  },
  bumper: {
    title: 'Bumper (type $21)',
    text: 'Named from play-testing as a special-stage bumper. It falls in the unnamed $13–$24 dispatch band '
      + 'in our notes, so its behaviour is not yet reverse-engineered beyond the name.',
  },

  // --- Bosses (own-graphics set-pieces; the viewer shows a marker) ---
  'world 1 boss': {
    title: 'World 1 boss (type $12)',
    text: 'Robotnik’s pod for World 1 (handler bank 1 $7065): a bytecode-scripted set-piece that sweeps '
      + 'across the arena and takes 8 hits to defeat (hit counter $D2ED). It decompresses its own graphics '
      + 'over the zone sprite sheet, so the extractor skips it and the viewer shows a labelled marker.',
  },
  'world 2 boss': {
    title: 'World 2 boss (type $48)',
    text: 'The World 2 boss (handler bank 2 $84AB). Like the other bosses it is a self-contained set-piece '
      + 'that loads its own graphics over the zone sprite sheet, so the extractor skips it and the viewer '
      + 'shows a labelled marker; its script is not further detailed in our notes.',
  },
  'world 3 boss': {
    title: 'World 3 boss (type $2C)',
    text: 'The World 3 boss (handler bank 2 $806B). Like the other bosses it is a self-contained set-piece '
      + 'that loads its own graphics over the zone sprite sheet, so the extractor skips it and the viewer '
      + 'shows a labelled marker; its script is not further detailed in our notes.',
  },
  'world 4 boss': {
    title: 'World 4 boss (type $49)',
    text: 'The World 4 boss (handler bank 2 $9271). Like the other bosses it is a self-contained set-piece '
      + 'that loads its own graphics over the zone sprite sheet, so the extractor skips it and the viewer '
      + 'shows a labelled marker; its script is not further detailed in our notes.',
  },
};

const CAT = {
  crab: 'enemy', beetle: 'enemy', fish: 'enemy', porcupine: 'enemy', bird: 'enemy',
  bonus: 'item', shield: 'item', emerald: 'item', goal: 'item', capsule: 'item',
  'swing platform': 'platform', 'moving platform': 'platform', seesaw: 'platform',
  'bobbing platform': 'platform', 'floating log': 'platform',
  'world 1 boss': 'boss', 'world 2 boss': 'boss', 'world 3 boss': 'boss', 'world 4 boss': 'boss',
  checkpoint: 'ctrl', 'bg animator': 'ctrl',
};

export default {
  base: 'public/sonic-gg/',
  maxNativeFactor: 1, // GG 1:1 — never magnify past the original viewport
  markerCat: (o) => CAT[o.name] || 'default',
  hud: (level) => `${level.grid.width}x${level.grid.height} blocks`,
  objectInfo,
};
