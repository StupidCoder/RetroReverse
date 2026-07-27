// viewlevel.js — the Level category view: dispatches on the level document's
// type to the 2D tilemap engine or the 3D scene engine.

export async function mount(ctx) {
  const doc = await ctx.game.assetDoc(ctx.asset);
  if (doc.type === 'tilemap') {
    const mod = await import('./level2d.js');
    return mod.mount(ctx, doc);
  }
  const mod = await import('./level3d.js');
  return mod.mount(ctx, doc);
}
