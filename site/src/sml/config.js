// Super Mario Land — configuration for the shared 2-D level viewer (site/FORMAT.md).
// The data (grid, objects in px with anchored sprite icons, tile-range collision) is
// exported by "Super Mario Land (GB)/extract/cmd/webexport".
export default {
  base: 'public/sml/',
  strategy: 'sliced',
  maxNativeFactor: 6,
  minFitFactor: 0.9,
  markerCat: () => 'enemy',
  hud: (level) => {
    const n = (level.objects || []).length;
    return `${level.name} · ${level.grid.width}×${level.grid.height} tiles · ${n} objects`;
  },
};
