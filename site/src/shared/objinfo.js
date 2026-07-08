// Shared object click-to-inspect: one InfoCard, one resolution, one 3D picker — across the
// 2-D and 3-D viewers. Object info comes from two places, both handled here:
//
//   - EXTRACTED, per object: the format-2 object's own fields in <level>.objects.json — its
//     `name`, `type`, `props.text` (the object's own in-game words), or an optional pre-composed
//     `info: { title?, body?, quote? }`. Comes straight from the game (e.g. Ultima Underworld's
//     STRINGS.PAK names + writings).
//   - EDITORIAL, per game: an `objectInfo` map `{ <key>: { title, text } }` of hand-authored
//     documentation, keyed by the object's `name` / `type` / `model` / `actor`. Lives in a
//     per-game config (2-D games' config.js, the DS viewers' tables) because it is written prose,
//     not extractable data.
//
// resolveObjectInfo merges the two into an InfoCard payload; installPicker wires the 3-D
// raycast/click plumbing so viewers don't each re-implement it.

import * as THREE from 'three';
import { InfoCard } from './infocard.js';

// resolveObjectInfo returns an InfoCard payload ({title, subtitle, body, quote, muted}) for a
// format-2 object, or null if there is nothing to show. Precedence: the object's own extracted
// `info`, then the editorial `objectInfo` entry, then a minimal identity from the object's fields.
export function resolveObjectInfo(obj, objectInfo) {
  const ex = obj.info || null; // extracted, per-object
  const key = obj.name ?? obj.type ?? obj.model ?? obj.actor;
  const ed = objectInfo && key != null ? objectInfo[key] : null; // editorial { title, text }

  const title = ex?.title ?? ed?.title ?? obj.name
    ?? (obj.actor != null ? `Actor ${obj.actor}` : undefined)
    ?? (obj.type != null ? `type ${obj.type}` : undefined)
    ?? (obj.id != null ? `object ${obj.id}` : undefined);
  const subtitle = ex?.subtitle ?? (obj.id != null ? `id ${obj.id}` : undefined);
  const body = ex?.body ?? ed?.text ?? undefined;
  const quote = ex?.quote ?? obj.props?.text ?? undefined; // the object's own words

  if (!title && !body && !quote) return null;
  return { title, subtitle, body, quote, muted: !ex && !ed };
}

// installPicker wires click-to-inspect on a 3-D viewer. `stage` provides `{ el, canvas, camera }`
// (a Stage3D, or any adapter exposing those). `pickables` is an array — or a getter returning one,
// for viewers whose object set changes per level — of `{ object3d, obj }` (obj = the format-2
// object for resolution). A non-drag pointerup raycasts the object3ds with the camera and shows the
// resolved payload in a shared InfoCard. Options: `objectInfo` (editorial map for the default
// resolver), `resolve(obj)` (override, for a viewer's own title composition, e.g. "Actor 202 —
// kuribo_model"), `enabled()` (skip picking when it returns false, e.g. the objects layer is off).
// Returns a teardown that removes the listeners and hides the card.
export function installPicker(stage, pickables, opts = {}) {
  const { objectInfo, resolve, enabled } = opts;
  const getItems = typeof pickables === 'function' ? pickables : () => pickables;

  const card = new InfoCard(stage.el);
  card.hide();
  const ray = new THREE.Raycaster();
  const ndc = new THREE.Vector2();
  const canvas = stage.canvas;
  let downX = 0, downY = 0;

  const onDown = (e) => { downX = e.clientX; downY = e.clientY; };
  const onUp = (e) => {
    if (enabled && !enabled()) return;
    if (Math.hypot(e.clientX - downX, e.clientY - downY) > 5) return; // ignore drags / look
    const items = getItems();
    if (!items || !items.length) { card.hide(); return; }
    const owner = new Map(); // every descendant object3d -> its pickable
    const roots = [];
    for (const p of items) { roots.push(p.object3d); p.object3d.traverse((n) => owner.set(n, p)); }
    const rect = canvas.getBoundingClientRect();
    ndc.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
    ndc.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;
    ray.setFromCamera(ndc, stage.camera);
    const hits = ray.intersectObjects(roots, true);
    const pick = hits.length ? owner.get(hits[0].object) : null;
    if (!pick) { card.hide(); return; }
    const payload = resolve ? resolve(pick.obj) : resolveObjectInfo(pick.obj, objectInfo);
    if (payload) card.show(payload); else card.hide();
  };
  canvas.addEventListener('pointerdown', onDown);
  canvas.addEventListener('pointerup', onUp);

  return {
    hide: () => card.hide(), // hide the card (e.g. on a level change)
    dispose: () => {
      canvas.removeEventListener('pointerdown', onDown);
      canvas.removeEventListener('pointerup', onUp);
      card.hide();
    },
  };
}
