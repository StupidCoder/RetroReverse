// Data loading for the shared 2-D level viewer: meta.json, level files, the sprite
// index and image assets, per the common format documented in site/FORMAT.md.

import { Rectangle, Texture } from 'pixi.js';

export function loadImage(src) {
  return new Promise((res, rej) => {
    const i = new Image();
    i.onload = () => res(i);
    i.onerror = rej;
    i.src = src;
  });
}

export class GameData {
  constructor(base) {
    this.base = base;             // e.g. 'public/sml/'
    this.meta = null;
    this.spriteIndex = {};        // key -> { src, frames, anchor?, durations?, steps?, path? }
    this.spriteTex = new Map();   // key -> { frames: Texture[], entry }
    this.atlasCache = new Map();  // file -> HTMLImageElement
    this.shapes = null;           // collision profiles (kind "profiles"), lazy
  }

  async loadMeta() {
    this.meta = await fetch(this.base + 'meta.json').then((r) => r.json());
    try {
      this.spriteIndex = await fetch(this.base + 'sprites/index.json').then((r) => r.json());
    } catch { /* game has no sprites */ }
    return this.meta;
  }

  loadLevel(entry) {
    return fetch(this.base + entry.file).then((r) => r.json());
  }

  async atlasImage(file) {
    if (!this.atlasCache.has(file)) this.atlasCache.set(file, await loadImage(this.base + file));
    return this.atlasCache.get(file);
  }

  // Resolve a sprite key to per-frame textures (cached). Returns null for unknown keys.
  async sprite(key) {
    if (this.spriteTex.has(key)) return this.spriteTex.get(key);
    const entry = this.spriteIndex[key];
    if (!entry) return null;
    const src = Texture.from(await loadImage(this.base + entry.src)).source;
    src.scaleMode = 'nearest';
    const frames = entry.frames.map(([x, y, w, h]) =>
      new Texture({ source: src, frame: new Rectangle(x, y, w, h) }));
    const rec = { frames, entry };
    this.spriteTex.set(key, rec);
    return rec;
  }

  // Collision height profiles (collision.kind === "profiles"), shared per game.
  async loadShapes(file) {
    if (!this.shapes) this.shapes = await fetch(this.base + file).then((r) => r.json());
    return this.shapes;
  }
}
