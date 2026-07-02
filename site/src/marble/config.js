// Marble Madness — configuration for the shared 2-D level viewer (map view). The 3-D
// slope view (slopes.js) sits outside the common format; viewer.js composes the two.
export default {
  base: 'public/marble/',
  strategy: 'sliced',
  maxNativeFactor: 3,
  hud: (level) => `${level.name} · ${level.grid.width * 8}x${level.grid.height * 8}`,
};
