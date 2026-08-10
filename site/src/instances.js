// instances.js — hardware instancing for a level's repeated static scenery.
//
// The billboard batch (billboards.js) solved the Abyss's sprite cast; this is
// the same disease in solid geometry. OutRun's Sunny Beach places 411 course
// objects — 211 of them the same two-primitive palm tree — and every placement
// is its own Object3D.clone: 663 draw calls of scenery before the course
// itself is counted, plus a per-frame depth sort of every transparent palm.
// A Quest 3's browser runs out of main-thread long before its GPU notices
// geometry like this (the report that found it: looking down the palm-lined
// track crawls, looking back at the lone car does not — the draw-call
// signature).
//
// The batch turns each (object, primitive) pair into ONE InstancedMesh whose
// per-instance matrix is the placement's composed world transform: the palms
// become 2 draw calls instead of 422. Geometry and materials are the proto's
// own — shared, not copied — so the picture is the per-clone picture.
//
// The placement groups stay exactly as they were, just invisible: they still
// carry the transform, the pick record and the info panel (the raycaster
// does not test visibility, which for once is the useful reading of that old
// three.js gotcha). Only placements with no behaviour of any kind are taken:
// no clips, no skins, no clicks, no tint, no variants, no rooms, no layers —
// scenery in the strictest sense. Everything else keeps its clone.
//
// ?nobatch=1 keeps the per-placement meshes, which is what this batch is
// checked AGAINST: the two paths have to render the same picture.
//
// Imports NOTHING (the xrfloor.js discipline): the caller hands in the two
// three.js classes the assembly step needs, so the eligibility and grouping
// rules — the part that decides what is safe to batch — are testable in node
// against the real shipped level documents.

// A placement qualifies when nothing can ever make it differ from its
// siblings: same proto, same materials, only the transform its own.
export function instanceable(pl, doc, inst = {}) {
  return (doc?.type === 'model3d')
    && !doc.skinnedClone
    && !(doc.animations || []).length
    && !inst.playAnim && !inst.update
    && !pl.onClick && !pl.anim && !pl.tint && !pl.hflip
    && !(pl.variants || []).length
    && pl.room == null && !pl.layer && !pl.behavior && !pl.route;
}

// groupInstanceable buckets eligible records by object id and keeps the
// buckets worth a batch. records: [{ pl, doc, inst }]. Pure data in, pure
// data out — this is the decision the node test pins.
export function groupInstanceable(records, minCount = 4) {
  const byObject = new Map();
  for (const r of records) {
    if (!instanceable(r.pl, r.doc ?? r.inst?.doc, r.inst)) continue;
    if (!byObject.has(r.pl.object)) byObject.set(r.pl.object, []);
    byObject.get(r.pl.object).push(r);
  }
  for (const [k, rs] of byObject) if (rs.length < minCount) byObject.delete(k);
  return byObject;
}

// buildInstances replaces each group with InstancedMeshes under one group
// node. Returns { group, stats } or null when nothing qualified. Records that
// were taken get .batched = true and their node hidden; the caller keeps them
// for picking.
export function buildInstances(records, { Group, InstancedMesh }, { minCount = 4 } = {}) {
  const byObject = groupInstanceable(records, minCount);
  const group = new Group();
  group.name = 'instances';
  let meshes = 0, taken = 0, callsBefore = 0;

  for (const [, rs] of byObject) {
    // The clones share structure by construction (Object3D.clone of one
    // proto), so one traversal order names the same primitive in each.
    const perRecord = rs.map((r) => {
      r.node.updateMatrixWorld(true);
      const ms = [];
      r.node.traverse((o) => { if (o.isMesh) ms.push(o); });
      return ms;
    });
    const nMesh = perRecord[0].length;
    if (!nMesh || perRecord.some((ms) => ms.length !== nMesh)) continue;

    for (let k = 0; k < nMesh; k++) {
      const im = new InstancedMesh(perRecord[0][k].geometry, perRecord[0][k].material, rs.length);
      // The geometry's bounding sphere knows nothing about where the
      // instances went; letting three cull on it would blink the whole
      // orchard when the first palm leaves the frustum.
      im.frustumCulled = false;
      for (let i = 0; i < rs.length; i++) im.setMatrixAt(i, perRecord[i][k].matrixWorld);
      im.instanceMatrix.needsUpdate = true;
      group.add(im);
      meshes++;
    }
    for (const r of rs) {
      r.batched = true;
      r.node.visible = false;
      callsBefore += nMesh;
    }
    taken += rs.length;
  }

  if (!meshes) return null;
  return { group, stats: { taken, meshes, callsBefore } };
}

// The instance matrix buffers are the batch's own GPU allocation; geometry
// and materials belong to the protos and are not touched here.
export function disposeInstances(group) {
  group?.traverse((o) => { if (o.isInstancedMesh) o.dispose(); });
}
