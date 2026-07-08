// sprites3d — the shared directional-billboard renderer. It turns a first-class
// *sprite spec* into a textured three.js quad and updates it every frame: it picks
// the right directional view from the camera-to-object bearing (Doom-style 8-view
// sprites), advances the animation, orients the quad, and applies the blend mode.
// It is engine-agnostic — Ultima Underworld's creatures and (later) Doom's things
// both describe themselves with the same spec and share this one code path, so no
// game needs its own billboard/animation loop.
//
// Sprite spec (the canonical shape; every field optional except `sheet`/`frames`):
//   { billboard: "camera"|"yaw"|"none", // orientation (default "camera")
//     views: 8,          // directional angle buckets; 1 = a plain (non-directional) billboard
//     heading: 0,        // object facing in RADIANS (base angle for bucket 0); ignored when views==1
//     size: [w,h],       // world-unit quad size
//     anchor: "center",  // what pos means: "center" (default) | "bottom" (feet at pos)
//     sheet: "sprites/x.png",   // atlas URL (the caller resolves it as base+sheet)
//     frames: [[x,y,w,h], ...], // atlas rects in PIXELS; length == views*perView,
//                               //   laid out as `views` blocks of `perView` frames
//     perView: 1,        // animation frames per view
//     fps: 0,            // animation rate (0 = static, first frame of each view)
//     blend: "opaque"|"alpha"|"additive" }  // default "opaque"
//
// THREE is imported lazily inside makeSprite (the browser resolves it through the
// page's importmap) so the pure helpers below — pickFrame/quantizeBucket — import
// and unit-test under plain `node` with no WebGL and no three module present.

// ---- pure math (WebGL-free; unit-tested with `node`) ----

// quantizeBucket maps a relative angle (radians) to one of `views` equal buckets
// around the circle: bucket 0 is centred on angle 0 (the spec's `heading`), and
// buckets increase CCW. The result is always in [0, views).
export function quantizeBucket(angle, views) {
  if (views <= 1) return 0;
  const step = (2 * Math.PI) / views;
  let b = Math.round(angle / step) % views;
  if (b < 0) b += views;
  return b;
}

// pickFrame returns the index into spec.frames for the current camera bearing and
// time. `angleToCamera` is the horizontal angle (radians) from the object to the
// camera; the view bucket is quantize(angleToCamera - heading, views) and the
// animation frame is floor(t*fps) % perView (frame 0 when fps<=0). The frames array
// is laid out as `views` blocks of `perView`, so the index is bucket*perView+frame.
export function pickFrame(spec, angleToCamera, t) {
  const views = spec.views > 0 ? spec.views : 1;
  const perView = spec.perView > 0 ? spec.perView : 1;
  const fps = spec.fps > 0 ? spec.fps : 0;
  const heading = spec.heading || 0;
  const bucket = views === 1 ? 0 : quantizeBucket(angleToCamera - heading, views);
  const frame = fps > 0 ? Math.floor(t * fps) % perView : 0;
  return bucket * perView + frame;
}

// test-notes (run under node against the pure helpers above):
//   spec = { views: 8, perView: 1, fps: 0, heading: 0 }
//   pickFrame(spec, 0, 0)            === 0   // camera dead ahead of bucket 0
//   pickFrame(spec, Math.PI/4, 0)    === 1   // +45° => next bucket CCW
//   pickFrame(spec, -Math.PI/4, 0)   === 7   // -45° wraps to the last bucket
//   pickFrame(spec, Math.PI, 0)      === 4   // opposite view
//   heading rotates the buckets: pickFrame({...spec, heading: Math.PI/4}, Math.PI/4, 0) === 0
//   animation wraps: pickFrame({views:1, perView:4, fps:10}, 0, 0.35) === 3, at 0.45 === 0
//   plain billboard: pickFrame({views:1, perView:1}, anyAngle, 0) === 0

// ---- the three.js object (browser only) ----

// makeSprite builds the quad for a sprite spec and returns { object3d, update }.
// `texture` is the loaded atlas THREE.Texture (the caller loads base+spec.sheet and
// passes it). The material is unlit and NearestFilter (retro-crisp); blend selects
// opaque / alpha-masked (alphaTest, depthWrite off) / additive (AdditiveBlending,
// depthWrite off). Call update(cameraPos, t) each frame with the camera world
// position and elapsed seconds.
export async function makeSprite(spec, texture) {
  const THREE = await import('three');
  const [w, h] = spec.size || [1, 1];
  const blend = spec.blend || 'opaque';

  // Own the texture transform per sprite (clone shares the image but not the
  // repeat/offset), so many sprites can read different sub-rects of one atlas.
  const map = texture.clone();
  map.needsUpdate = true;
  map.magFilter = THREE.NearestFilter;
  map.minFilter = THREE.NearestFilter;
  map.generateMipmaps = false;

  const matOpts = { map, side: THREE.DoubleSide, toneMapped: false };
  if (blend === 'additive') {
    matOpts.transparent = true;
    matOpts.blending = THREE.AdditiveBlending;
    matOpts.depthWrite = false;
  } else if (blend === 'alpha') {
    matOpts.transparent = true;
    matOpts.alphaTest = 0.5; // masked edges — keep the depth buffer sane
    matOpts.depthWrite = false;
  } else {
    matOpts.alphaTest = 0.5; // opaque quads still cut out their transparent border
  }
  const material = new THREE.MeshBasicMaterial(matOpts);

  // anchor selects what the object's pos means: "center" (default) puts the quad's
  // centre at pos; "bottom" puts the quad's bottom edge at pos (feet on the floor —
  // the usual choice for standing creatures/props). We shift the geometry up by h/2
  // for "bottom" so the mesh origin (the billboard's rotation pivot) sits at the feet.
  const geo = new THREE.PlaneGeometry(w, h);
  if ((spec.anchor || 'center') === 'bottom') geo.translate(0, h / 2, 0);
  const object3d = new THREE.Mesh(geo, material);
  object3d.frustumCulled = false; // billboards jump around; their static bounds lie

  // setFrame points the cloned texture at one atlas rect (pixels). three's UV
  // origin is bottom-left, so the Y offset flips. Guard until the image loads.
  const frames = spec.frames || [];
  let lastIdx = -1;
  const setFrame = (idx) => {
    if (idx === lastIdx) return;
    const img = map.image;
    const rect = frames[idx];
    if (!img || !img.width || !rect) return;
    const [x, y, rw, rh] = rect;
    map.repeat.set(rw / img.width, rh / img.height);
    map.offset.set(x / img.width, 1 - (y + rh) / img.height);
    map.needsUpdate = true;
    lastIdx = idx;
  };
  setFrame(0);

  const billboard = spec.billboard || 'camera';
  const _v = new THREE.Vector3();
  const update = (cameraPos, t) => {
    const dx = cameraPos.x - object3d.position.x;
    const dz = cameraPos.z - object3d.position.z;
    setFrame(pickFrame(spec, Math.atan2(dx, dz), t || 0));
    if (billboard === 'camera') {
      object3d.lookAt(_v.copy(cameraPos)); // PlaneGeometry faces +Z; lookAt aims it at the camera
    } else if (billboard === 'yaw') {
      object3d.rotation.set(0, Math.atan2(dx, dz), 0); // spin about world-up only (stays upright)
    } // "none": leave the quad's fixed orientation
  };

  return { object3d, update };
}
