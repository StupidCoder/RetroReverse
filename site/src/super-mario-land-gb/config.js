// Super Mario Land — configuration for the shared 2-D level viewer (FORMAT2.md).
// The data (grid, objects in px with anchored sprite icons, tile-range collision) is
// exported by "Super Mario Land (GB)/extract/cmd/webexport".
//
// objectInfo turns the placed objects into clickable info cards: the shared viewer
// raycasts the Objects layer and shows objectInfo["type-<n>"] on a click, plus the raw
// ids (object type, the hard/second-quest flag) the card always prints. Placed objects
// carry a numeric `type` (0–127) that indexes the ROM's per-type script table at $3495;
// each entry below is read off that type's own behaviour script (extract/cmd/objscript)
// and its base metasprite frame, folded from Part V §5. Only type $00 (Goomba, named in
// objects.go) and $0A/$0B (level-end moving platforms, §1/§3) are confidently *named*;
// the rest use descriptive titles from their scripts rather than inventing retail names.
const objectInfo = {
  player: {
    title: 'Mario — the player',
    text: 'The level spawn point. Mario is drawn from his own sprite set (not the object-script table), and '
      + 'his run/jump physics are the subject of Part V. The viewer shows him as a static placement at the '
      + 'start position the ROM sets for this level.',
  },
  'type-0': {
    title: 'Goomba (type $00)',
    text: 'The two-frame walker that opens 1-1 — objects.go and Part V both name it the Goomba. Its script '
      + 'is the whole language in miniature: set frame $00, set a 3-frame step divider, then alternate frames '
      + '$00/$01 (the same tile mirrored) while creeping forward forever. Placed across worlds 1, 2 and 4.',
  },
  'type-2': {
    title: 'Ambusher (type $02)',
    text: 'Stands with frame $04 and waits for the player (a $F6 gate), then lunges out (move $10), flickers '
      + 'between frames $04/$05, and turns back. An in-place ambusher rather than a free walker. Placed in '
      + 'every world (1-4).',
  },
  'type-3': {
    title: 'Stationary thrower (type $03)',
    text: 'Holds frame $1F and, once per cycle, spawns a type $47 lobbed projectile that arcs toward Mario. '
      + 'It never moves — just throws. Placed in world 3.',
  },
  'type-4': {
    title: 'Walker (type $04)',
    text: 'A second basic ground enemy: base frame $06, a forward two-frame gait like the Goomba. The most '
      + 'heavily placed direct enemy in the game, appearing in all four worlds.',
  },
  'type-6': {
    title: 'Small hopper (type $06)',
    text: 'Base frame $65; shuffles back and forth in short hops. Placed in world 4. Identification is '
      + 'behavioural only — the script gives the gait, not a retail name.',
  },
  'type-8': {
    title: 'Spitter (type $08)',
    text: 'A fixed emplacement (frame $48) that periodically fires $1B projectiles, with a queued sound '
      + 'effect. Does not walk. Placed in world 1.',
  },
  'type-9': {
    title: 'Flyer / bomber (type $09)',
    text: 'Hovers with frame $56 and drops a $51 arcing bomb/shot (with a sound effect). An aerial attacker '
      + 'placed in world 4.',
  },
  'type-10': {
    title: 'Moving platform — vertical (type $0A)',
    text: 'A rideable moving platform / level-end lift. Part V §3 reads its script as "set frame $12; move; '
      + 'coast ~60 frames; reverse; coast ~60; restart" — a slow vertical shuttle. Part V §1 flags $0A/$0B '
      + 'as the level-end fixtures. Placed in all worlds.',
  },
  'type-11': {
    title: 'Moving platform — horizontal (type $0B)',
    text: 'The horizontal sibling of $0A: same frame $12, but shuttles left/right instead of up/down. Also '
      + 'one of the level-end fixtures (Part V §1). Placed in all worlds.',
  },
  'type-12': {
    title: 'Sliding / sinking platform (type $0C)',
    text: 'Base frame $13. Sits still, then drifts a long way (a sliding or sinking platform). Placed in '
      + 'worlds 1 and 4.',
  },
  'type-14': {
    title: 'Leaper (type $0E)',
    text: 'Bounces along in an arc, animating between frames $28 and $29. Placed in worlds 1 and 3.',
  },
  'type-16': {
    title: 'Hopper (type $10)',
    text: 'Jumps around the terrain, alternating frames $42/$43. Placed in world 2. Behavioural id only.',
  },
  'type-22': {
    title: 'Walking shooter (type $16)',
    text: 'Walks (base frame $32), fires a $17 projectile, then cools down as state $18 before resuming. A '
      + 'mobile shooter placed in world 2.',
  },
  'type-26': {
    title: 'Fire-spitting swimmer (type $1A)',
    text: 'Moves with frame $4F and fires $1F shots — a spitting swimmer of the underwater world 2. '
      + 'Identification is behavioural (movement + projectile spawn).',
  },
  'type-29': {
    title: 'Drifter / swimmer (type $1D)',
    text: 'Bobs along, alternating frames $2A/$2B. Placed in world 2 (the underwater stage).',
  },
  'type-32': {
    title: 'Bobbing enemy (type $20)',
    text: 'Rises and falls in place, frames $28/$29. Placed in world 2. Behavioural id only.',
  },
  'type-36': {
    title: 'Hopping shooter (type $24)',
    text: 'Hops (frame $2A) and fires $23 projectiles — an octopus-like attacker of world 2. Name is inferred '
      + 'from the hop-and-shoot script, not confirmed.',
  },
  'type-37': {
    title: 'Charger (type $25)',
    text: 'Frame $32; waits for Mario (a $F6 gate), then runs a long distance in one burst. A charging enemy '
      + 'placed in world 3.',
  },
  'type-47': {
    title: 'Segment-layer crawler (type $2F)',
    text: 'Crawls with frame $2C, laying $30 body segments, then becomes a $30 crawler itself — a multi-part '
      + 'segmented enemy of world 2 (built via spawn/become chains).',
  },
  'type-49': {
    title: 'Fast skitterer (type $31)',
    text: 'Zips along quickly, frame $2A — a spider-like skitterer. Placed in world 3. Behavioural id only.',
  },
  'type-50': {
    title: 'Cannon / plant (type $32)',
    text: 'A fixed cannon or plant (frames $54/$55) that fires $33 shots. Placed in world 3.',
  },
  'type-53': {
    title: 'Drop hazard (type $35)',
    text: 'Frame $20; hangs until Mario is beneath (a $F6 gate), then descends. A ceiling drop-hazard of '
      + 'world 3.',
  },
  'type-54': {
    title: 'Stationary hazard / decoration (type $36)',
    text: 'Holds a single static frame ($21) in place — a fixed hazard or decoration. The most common '
      + 'placement overall (all four worlds). Its script only sets a frame, so whether it is scenery or a '
      + 'spike is not decidable from our data alone.',
  },
  'type-56': {
    title: 'Diagonal bouncer (type $38)',
    text: 'Bounces along a diagonal path, frame $12. Placed in worlds 3 and 4.',
  },
  'type-57': {
    title: 'Diagonal bouncer, mirrored (type $39)',
    text: 'The mirror-image counterpart of $38 — same diagonal bounce (frame $12), opposite lean. Placed in '
      + 'worlds 3 and 4.',
  },
  'type-58': {
    title: 'Vertical patroller (type $3A)',
    text: 'Patrols up and down, frame $22. Placed in worlds 3 and 4.',
  },
  'type-59': {
    title: 'Horizontal patroller (type $3B)',
    text: 'The horizontal counterpart of $3A: patrols left/right, frame $22. Placed in world 3.',
  },
  'type-60': {
    title: 'Hopper / flapper (type $3C)',
    text: 'Runs an elaborate hop-and-flap routine cycling frames $23–$25 with many small moves. Placed in '
      + 'world 3. Behavioural id only.',
  },
  'type-63': {
    title: 'Spitter (type $3F)',
    text: 'A fixed spitter (frame $2A) that fires $23 projectiles with a sound effect. Placed in worlds 1 '
      + 'and 4.',
  },
  'type-66': {
    title: 'Flapping flyer (type $42)',
    text: 'Flaps in place (frames $30/$31) and drops a $45 projectile mid-cycle. An aerial dropper placed in '
      + 'world 1.',
  },
  'type-71': {
    title: 'Lobbed projectile (type $47)',
    text: 'An arcing projectile — the shot thrown by the type $03 thrower (frames $31/$47 as it rises and '
      + 'falls). Also placed directly in world 3.',
  },
  'type-72': {
    title: 'Diagonal flyer / drifter (type $48)',
    text: 'Drifts diagonally, frames $4C/$4D. Placed in world 2. Behavioural id only.',
  },
  'type-73': {
    title: 'Slow floater (type $49)',
    text: 'Bobs slowly along, frame $51. Placed in worlds 3 and 4.',
  },
  'type-75': {
    title: 'Cruising projectile (type $4B)',
    text: 'A projectile in cruise, frames $52/$53. Appears once in our data, as a second-quest (hard) '
      + 'placement in world 4.',
  },
  'type-82': {
    title: 'Creeping shooter (type $52)',
    text: 'Creeps forward (frame $31), then fires a $50 shot with a sound effect before restarting. Placed in '
      + 'world 4.',
  },
  'type-83': {
    title: 'Bouncing charger (type $53)',
    text: 'Frames $32/$33; if Mario is not close it bounds along in a long charge (a $FB proximity loop). '
      + 'Part V calls it the world-4 boss bullet — it is both placed directly and fired by the $61 boss.',
  },
  'type-84': {
    title: 'Parabola flyer (type $54)',
    text: 'A missile/shot that flies a parabolic arc, frame $45. Placed in world 4.',
  },
  'type-85': {
    title: 'Hovering flyer (type $55)',
    text: 'Hovers and drifts, alternating frames $5F/$60. Placed in world 4.',
  },
  'type-86': {
    title: 'Leaper (type $56)',
    text: 'Arcs forward in a leap (frames $28/$29) — a world-4 variant of the $0E leaper. Placed in world 4.',
  },
  'type-89': {
    title: 'Spinner (type $59)',
    text: 'Cycles frames $61/$62 continuously as it moves — a spinning enemy of world 4.',
  },
  'type-97': {
    title: 'Boss launcher — Tatanga-like (type $61)',
    text: 'Animates between frames $66/$67 and repeatedly spawns $53 bullets. Part V flags the world-4 cluster '
      + '$50–$62 as the Tatanga boss fight, with $61 as the launcher. Placed once, in world 4.',
  },
};

export default {
  base: 'public/super-mario-land-gb/',
  strategy: 'baked', // per-tile canvas textures so tileAnims (torch/waterfall) repaint

  maxNativeFactor: 6,
  minFitFactor: 0.9,
  markerCat: () => 'enemy',
  hud: (level) => {
    const n = (level.objects || []).length;
    return `${level.name} · ${level.grid.width}×${level.grid.height} tiles · ${n} objects`;
  },
  objectInfo,
};
