// shadowmask.js — a blob shadow that belongs to a placed OBJECT and is
// projected onto the ground every frame.
//
// It is the other kind of shadow a guest casts, and the one shadowmap.js cannot
// express. That pass renders the level's own caster geometry once, from the
// level's own light, because the level does not move. A character does: it is
// animated, and Captain Toad's goal star flies off its pedestal during the
// opening clip. Several guests do not put their characters in the depth-shadow
// pass at all — Captain Toad's `InitExecutor` lists the terrain and the star and
// not the characters — and give them a mask instead: a disc hung from a named
// joint, dropped onto whatever is below, at a stated radius and colour.
//
// Three things follow from "projected onto the ground", and they are the whole
// reason this is not baked geometry:
//
//   * it sits where the GROUND is, not where the caster is, so a rising caster
//     leaves its shadow behind on the floor;
//   * it weakens with the distance it had to fall, and vanishes past the reach
//     the guest states — an object high above the world casts nothing;
//   * it follows the joint, so an animated skeleton drags it around.
//
// A shadow is a MULTIPLY. Blended with alpha instead, a grey disc lightens dark
// ground as much as it darkens light ground: fog, not shade. The disc therefore
// multiplies and fades to WHITE at its rim, because multiplying by one is what
// "no shadow here" means.
import * as THREE from 'three';

const SEGMENTS = 32;
const CORE = 0.6;  // the fraction of the radius at full strength
const LIFT = 1.5;  // clear of the surface it lies on, in world units

// srgbToLinear matches the exporter's. The colours below are written into a
// vertex-colour attribute, which three.js takes as linear and encodes back to
// gamma before blending — and gamma is the space the guest multiplies in. A
// factor of 0.2 written straight in arrives as 0.48.
function srgbToLinear(c) {
  return c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
}

function parseHex(hex) {
  const c = new THREE.Color();
  c.setStyle(hex || '#000000', THREE.NoColorSpace); // a multiplier, not a colour to convert
  return c;
}

// createShadowMask builds the disc for one placement. It returns the mesh to
// add to the scene and an update(inst, ground) to call each frame; dispose()
// releases what it owns.
export function createShadowMask(mask) {
  if (!mask || !(mask.radius > 0) || !(mask.drop > 0)) return null;

  // One triangle fan out to a solid ring, then a band that fades to white.
  const verts = [];
  const radial = []; // per vertex: 1 at full strength, 0 at the rim
  verts.push(0, 0, 0);
  radial.push(1);
  for (let i = 0; i < SEGMENTS; i++) {
    const a = (i / SEGMENTS) * Math.PI * 2;
    const cs = Math.cos(a), sn = Math.sin(a);
    verts.push(cs * CORE, 0, sn * CORE, cs, 0, sn);
    radial.push(1, 0);
  }
  const idx = [];
  const inner = (i) => 1 + (i % SEGMENTS) * 2;
  const outer = (i) => 2 + (i % SEGMENTS) * 2;
  for (let i = 0; i < SEGMENTS; i++) {
    idx.push(0, inner(i), inner(i + 1));
    idx.push(inner(i), outer(i), outer(i + 1));
    idx.push(inner(i), outer(i + 1), inner(i + 1));
  }

  const geo = new THREE.BufferGeometry();
  geo.setAttribute('position', new THREE.Float32BufferAttribute(verts, 3));
  geo.setAttribute('color', new THREE.Float32BufferAttribute(new Float32Array(radial.length * 3), 3));
  geo.setIndex(idx);

  const mat = new THREE.MeshBasicMaterial({
    vertexColors: true,
    transparent: true,
    blending: THREE.MultiplyBlending,
    depthWrite: false,
    side: THREE.DoubleSide,
    polygonOffset: true,
    polygonOffsetFactor: -1,
    polygonOffsetUnits: -1,
  });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.renderOrder = 1;
  mesh.frustumCulled = false;
  mesh.visible = false;
  mesh.scale.setScalar(mask.radius);

  const colour = parseHex(mask.color);
  const offset = new THREE.Vector3(...(mask.offset?.length === 3 ? mask.offset : [0, 0, 0]));
  const fadeExp = mask.fadeExp > 0 ? mask.fadeExp : 1;

  // Writing the strength into the vertex colours rather than into a uniform
  // keeps the material a plain MeshBasicMaterial; there are sixty-six of them.
  let lastStrength = -1;
  const setStrength = (s) => {
    if (Math.abs(s - lastStrength) < 1 / 512) return;
    lastStrength = s;
    const col = geo.getAttribute('color');
    for (let i = 0; i < radial.length; i++) {
      const f = s * radial[i];
      col.setXYZ(i,
        srgbToLinear(1 - (1 - colour.r) * f),
        srgbToLinear(1 - (1 - colour.g) * f),
        srgbToLinear(1 - (1 - colour.b) * f));
    }
    col.needsUpdate = true;
  };

  const ray = new THREE.Raycaster();
  ray.far = mask.drop + LIFT;
  const from = new THREE.Vector3();
  const down = new THREE.Vector3(0, -1, 0);
  let joint = null;

  return {
    mesh,
    // inst is the placed object's node; ground is what the ray is cast against.
    update(instNode, ground) {
      if (!ground?.length) { mesh.visible = false; return; }
      if (!joint && mask.joint) joint = instNode.getObjectByName(mask.joint);
      const anchor = joint || instNode;
      anchor.updateWorldMatrix(true, false);
      from.copy(offset).applyMatrix4(anchor.matrixWorld);

      ray.set(from, down);
      const hits = ray.intersectObjects(ground, true);
      const hit = hits.find((h) => h.object !== mesh);
      if (!hit) { mesh.visible = false; return; }

      // Where the ground is, not where the caster is.
      mesh.position.copy(hit.point);
      mesh.position.y += LIFT;
      mesh.visible = true;
      setStrength(Math.pow(1 - Math.min(hit.distance / mask.drop, 1), fadeExp));
    },
    dispose() {
      geo.dispose();
      mat.dispose();
    },
  };
}
