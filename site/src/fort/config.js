// Fort Apocalypse — configuration for the shared 2-D level viewer (site/FORMAT.md).
// The playfield is a horizontal cylinder (meta wrap "x"); the soft-char animations
// need the baked per-tile strategy (repaint in place); prisoners/tanks/mines/the
// enemy helicopter are randomized objectPools re-rolled per toggle, and the movers
// patrol per their engines ($992A tanks, $AABA prisoners, $94D2 mines) via the
// pools' patrol/dirStamps data. Exported by
// "Fort Apocalypse (C64)/extract/cmd/webexport".
export default {
  base: 'public/fort/',
  strategy: 'baked',
  maxNativeFactor: 3,
  minFitFactor: 1, // never zoom out past one cylinder period (objects would repeat)
  hud: (level) => `${level.grid.width}x${level.grid.height} chars · wraps`,
};
