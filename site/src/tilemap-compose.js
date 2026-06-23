// composeTilemap builds a PixiJS Container of one sprite per cell from a tile-atlas
// image and a row-major `cells` array — the shared core of the tilemap viewers (Super
// Mario Land, Marble Madness). The atlas is a grid of tileSize×tileSize tiles, each
// surrounded by a `gutter`-pixel extruded border (so tiles don't bleed at fractional
// zoom). Returns the container plus the atlas TextureSource (so the caller can switch
// its scale mode on zoom).
import { Container, Rectangle, Sprite, Texture } from 'pixi.js';

export function composeTilemap(atlasImg, cells, W, H, opts = {}) {
  const tileSize = opts.tileSize ?? 8;
  const gutter = opts.gutter ?? 1;
  const atlasCols = opts.atlasCols ?? 16;
  const ntiles = opts.ntiles ?? atlasCols * Math.ceil((atlasImg.height / (tileSize + 2 * gutter)));
  const acell = tileSize + 2 * gutter;

  const src = Texture.from(atlasImg).source;
  src.scaleMode = 'nearest';
  const frames = [];
  for (let n = 0; n < ntiles; n++) {
    const sx = (n % atlasCols) * acell + gutter;
    const sy = ((n / atlasCols) | 0) * acell + gutter;
    frames.push(new Texture({ source: src, frame: new Rectangle(sx, sy, tileSize, tileSize) }));
  }

  const container = new Container();
  for (let r = 0; r < H; r++) {
    for (let x = 0; x < W; x++) {
      let n = cells[r * W + x];
      if (n < 0 || n >= frames.length) n = 0;
      const s = new Sprite(frames[n]);
      s.x = x * tileSize;
      s.y = r * tileSize;
      container.addChild(s);
    }
  }
  return { container, src };
}
