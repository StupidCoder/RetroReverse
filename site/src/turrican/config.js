// Turrican — configuration for the shared 2-D level viewer (site/FORMAT.md). The data
// (32px tiles with the explicit hflip bit, 4x4 per-tile collision with the class
// legend, object placements resolved to sprite-atlas keys) is exported by
// "Turrican (Amiga)/extract/cmd/webexport".
export default {
  base: 'public/turrican/',
  strategy: 'sliced',
  maxNativeFactor: 4,
  minFitFactor: 0.9,
  hud: (level) => {
    const n = (level.objects || []).length;
    return `${level.grid.width}×${level.grid.height} tiles` + (n ? ` · ${n} objects` : '');
  },
};
