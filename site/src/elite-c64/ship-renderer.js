// Elite ship renderer — the `elite-ship` plugin for the shared Viewer3D/Stage3D seam. It renders
// one decoded wireframe blueprint and reproduces the game's own hidden-surface removal, then owns
// the frame so it can composite through Elite's bespoke low-res / CRT post-process. The HSR math,
// the shaders and the LORES/vignette constants are moved VERBATIM from the old elite/viewer.js
// ShipViewer — the plugin is only a new shell around exactly the same visuals.
//
// The ship sits at the origin and the camera orbits it (OrbitControls), so model space is world
// space. Each frame we back-face test every face by the sign of dot(normal, eye - faceCenter) —
// the face is visible when it points toward the eye from where the face sits — and draw an edge
// only when one of its two adjacent faces is visible (Elite.md Part IV §1). Using the face's own
// centre (not the ship's) is what correctly hides faces on the far side; the eye is kept outside
// the hull so faces never flip abruptly.
import * as THREE from 'three';

const WHITE = 0xffffff;
const FACE_NONE = 15; // edge face nibble sentinel: no face this side → always drawn

// "Old school" authentic flicker. The C64 draws ships by XOR-plotting their line heap into a
// single, live bitmap (no double buffer): to move a ship it re-draws the old lines to erase them,
// reprojects, then draws the new lines. That erase→reproject→redraw runs in the game loop, out of
// sync with the 50 Hz beam, so the beam keeps catching individual lines during the brief window
// they are erased — the signature Elite shimmer. We reproduce it by giving every edge its own
// erase-window phase and, each displayed frame, hiding the edges whose window the "beam" is
// currently sampling. FLICKER_DUTY ≈ the fraction of a redraw cycle a line spends erased;
// FLICKER_RATE drifts the sampling phase so the off-set changes (and beats) every frame. (In the
// Studio the flicker is off — kept verbatim so the effect is byte-for-byte the same when enabled.)
const FLICKER_DUTY = 0.17;
const FLICKER_RATE = 0.37;

// Low-res internal height = the C64's vertical resolution, so the ship renders in chunky C64-sized
// pixels (width follows the viewport aspect, keeping the pixels square). The Studio's global CRT
// filter sizes its scanlines/mask to this via pixelGrid().
const LORES_H = 200;
const AUTO_ROTATE_SPEED = 1.1; // idle spin
// Glow buffer height (fixed, so the blur radius is a constant fraction of the screen regardless of
// the main render resolution). Width follows the aspect.
const GLOW_H = 200;

// CRT post-process. In old-school mode the scene is rendered into a small (LORES_H-tall) texture,
// then this shader draws that texture to the full-res canvas with the tube's character: barrel
// curvature + bezel, scanlines, an aperture-grille RGB mask (needs the full output resolution to
// resolve), a cheap phosphor bloom, and a vignette. Sampling the small texture with NEAREST keeps
// the chunky C64 pixels; the mask/scanlines then live in the high-res output.
const CRT_VERT = /* glsl */`
  varying vec2 vUv;
  void main() { vUv = uv; gl_Position = vec4(position.xy, 0.0, 1.0); }
`;
// Separable Gaussian blur (5-sample, linear-weighted) used to build the glow.
// uTexel is the 1-texel step along one axis: (1/w, 0) then (0, 1/h).
const BLUR_FRAG = /* glsl */`
  varying vec2 vUv;
  uniform sampler2D tSrc;
  uniform vec2 uTexel;
  void main() {
    vec2 o1 = uTexel * 1.3846153846;
    vec2 o2 = uTexel * 3.2307692308;
    vec3 s = texture2D(tSrc, vUv).rgb * 0.2270270270
           + texture2D(tSrc, vUv + o1).rgb * 0.3162162162
           + texture2D(tSrc, vUv - o1).rgb * 0.3162162162
           + texture2D(tSrc, vUv + o2).rgb * 0.0702702703
           + texture2D(tSrc, vUv - o2).rgb * 0.0702702703;
    gl_FragColor = vec4(s, 1.0);
  }
`;
const CRT_FRAG = /* glsl */`
  varying vec2 vUv;
  uniform sampler2D tScene;
  uniform sampler2D tGlow; // pre-blurred glow (a real Gaussian, so round + ring-free)
  uniform vec2 uSceneRes;  // off-screen render-target size
  uniform float uCRT;      // 1 = full CRT, 0 = plain (chunky) upscale only
  const float PI = 3.14159265;

  vec2 curve(vec2 uv) {            // barrel screen curvature
    uv = uv * 2.0 - 1.0;
    vec2 off = abs(uv.yx) * vec2(0.045, 0.060);
    uv += uv * off * off;
    return uv * 0.5 + 0.5;
  }

  void main() {
    vec2 uv = (uCRT > 0.5) ? curve(vUv) : vUv;
    if (uCRT > 0.5 && (uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0)) {
      gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0); return; // black bezel off the tube
    }
    // Main image, snapped to texel centres (blocky when the target is low-res;
    // a sharp 1:1 copy when it's full-res).
    vec3 col = texture2D(tScene, (floor(uv * uSceneRes) + 0.5) / uSceneRes).rgb;

    if (uCRT > 0.5) {
      col += texture2D(tGlow, uv).rgb * 0.85; // smooth, round phosphor glow
      // scanlines + RGB mask at the OUTPUT pixel level (gl_FragCoord, device px),
      // so they are a fine CRT structure over the (possibly chunky) image.
      col *= 0.6 + 0.4 * cos(gl_FragCoord.y * (2.0 * PI / 3.0)); // scanlines, ~3px pitch
      float tri = mod(floor(gl_FragCoord.x), 3.0);               // aperture-grille triads
      vec3 mask = (tri < 1.0) ? vec3(1.0, 0.5, 0.5)
                : (tri < 2.0) ? vec3(0.5, 1.0, 0.5)
                              : vec3(0.5, 0.5, 1.0);
      col *= mask;
      col *= 1.6; // compensate for the mask/scanline darkening
      vec2 vd = vUv * 2.0 - 1.0;
      col *= clamp(1.0 - 0.22 * dot(vd, vd), 0.0, 1.0); // vignette
    }
    // linear scene texture, raw ShaderMaterial (no auto output encode) -> sRGB
    col = pow(clamp(col, 0.0, 1.0), vec3(1.0 / 2.2));
    gl_FragColor = vec4(col, 1.0);
  }
`;

// ShipMesh holds one ship's geometry and the per-frame visible-edge buffer. Flat typed arrays keep
// the HSR loop tight; verts/normals stay in model space (the ship never moves — the camera orbits).
// Elite's +Y is up, matching three.js; X right, Z toward the viewer (the offline montage's
// handedness). (Moved verbatim from the old ShipViewer, plus a bounding volume computed from all
// verts so the stage can fit the camera before the draw buffer is first filled.)
class ShipMesh {
  constructor(ship) {
    this.radius = ship.radius || 1;

    this.verts = new Float32Array(ship.verts.length * 3);
    for (let i = 0; i < ship.verts.length; i++) {
      this.verts[i * 3] = ship.verts[i][0];
      this.verts[i * 3 + 1] = ship.verts[i][1];
      this.verts[i * 3 + 2] = ship.verts[i][2];
    }

    this.edges = ship.edges; // [v1, v2, faceA, faceB]
    // A stable per-edge phase in [0,1) for the authentic flicker: each line gets
    // its own slot in the erase/redraw cycle, scattered so the blinking reads as
    // shimmer rather than a clean wipe.
    this.edgePhase = new Float32Array(this.edges.length);
    for (let i = 0; i < this.edges.length; i++) {
      const e = this.edges[i];
      const h = Math.sin((e[0] + 1) * 12.9898 + (e[1] + 1) * 78.233 + i * 0.613) * 43758.5453;
      this.edgePhase[i] = h - Math.floor(h);
    }
    this.faceN = new Float32Array(ship.faces.length * 3); // outward normal per face
    this.faceC = new Float32Array(ship.faces.length * 3); // a point on each face
    for (let i = 0; i < ship.faces.length; i++) {
      this.faceN[i * 3] = ship.faces[i][0];
      this.faceN[i * 3 + 1] = ship.faces[i][1];
      this.faceN[i * 3 + 2] = ship.faces[i][2];
      this.faceC[i * 3] = ship.faceC[i][0];
      this.faceC[i * 3 + 1] = ship.faceC[i][1];
      this.faceC[i * 3 + 2] = ship.faceC[i][2];
    }
    this.faceVis = new Uint8Array(ship.faces.length);

    // One LineSegments whose position buffer we refill each frame with only the
    // currently-visible edges; drawRange caps it to what we wrote.
    const positions = new Float32Array(this.edges.length * 6);
    const geom = new THREE.BufferGeometry();
    geom.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geom.setDrawRange(0, 0);
    // Bounding volume from ALL verts. The draw buffer starts empty (drawRange 0), so give the
    // stage a true bounding box/sphere to fit the camera to instead of the (initially empty)
    // position attribute. HSR is unaffected — this only feeds Stage3D.frame().
    const bb = new THREE.Box3();
    const p = new THREE.Vector3();
    for (let i = 0; i < ship.verts.length; i++) {
      bb.expandByPoint(p.set(ship.verts[i][0], ship.verts[i][1], ship.verts[i][2]));
    }
    geom.boundingBox = bb;
    geom.boundingSphere = bb.getBoundingSphere(new THREE.Sphere());
    this.geom = geom;
    this.positions = positions;
    this.object = new THREE.LineSegments(geom, new THREE.LineBasicMaterial({ color: WHITE }));
    this.object.frustumCulled = false;
  }

  // updateForCamera rebuilds the visible-edge list for a camera at camPos
  // (THREE.Vector3). A face is visible when its outward normal points toward the
  // eye from the face's own position: dot(normal, camPos - faceCenter) > 0.
  // Testing from the face centre (not the origin) is what correctly culls faces
  // on the far side instead of leaving them showing "below the horizon".
  // flickerPhase < 0 disables the effect; otherwise it is the current sampling
  // phase, and an edge is dropped this frame when the "beam" falls inside its
  // erase window — reproducing the C64's single-buffer XOR redraw flicker.
  updateForCamera(camPos, flickerPhase = -1) {
    const { verts, faceN, faceC, faceVis, edges, edgePhase } = this;
    for (let i = 0; i < faceVis.length; i++) {
      const dot = faceN[i * 3] * (camPos.x - faceC[i * 3])
        + faceN[i * 3 + 1] * (camPos.y - faceC[i * 3 + 1])
        + faceN[i * 3 + 2] * (camPos.z - faceC[i * 3 + 2]);
      faceVis[i] = dot > 0 ? 1 : 0;
    }
    const pos = this.positions;
    const flicker = flickerPhase >= 0;
    let n = 0;
    for (let ei = 0; ei < edges.length; ei++) {
      const e = edges[ei];
      const fa = e[2], fb = e[3];
      const vis = fa === FACE_NONE || fb === FACE_NONE || faceVis[fa] || faceVis[fb];
      if (!vis) continue;
      if (flicker) {
        let d = edgePhase[ei] - flickerPhase;
        d -= Math.floor(d);
        if (d < FLICKER_DUTY) continue; // this line is mid-erase right now → blink off
      }
      const a = e[0] * 3, b = e[1] * 3;
      pos[n++] = verts[a]; pos[n++] = verts[a + 1]; pos[n++] = verts[a + 2];
      pos[n++] = verts[b]; pos[n++] = verts[b + 1]; pos[n++] = verts[b + 2];
    }
    this.geom.setDrawRange(0, n / 3);
    this.geom.attributes.position.needsUpdate = true;
  }

  dispose() {
    this.geom.dispose();
    this.object.material.dispose();
  }
}

// fitDistance returns a camera distance that frames a ship of the given radius (used for the
// orbit zoom limits; the initial pose comes from Stage3D.frame).
function fitDistance(radius, fovDeg) {
  return (radius * 1.6) / Math.sin((fovDeg * Math.PI) / 360);
}

// makeStarfield returns a Points cloud of dim dots scattered on a large sphere. It lives in world
// space (the ship is fixed; the camera orbits), so the stars wheel around with the ship.
// sizeAttenuation:false keeps them a constant pixel size, so zooming the ship doesn't change them.
function makeStarfield(count, radius) {
  const pos = new Float32Array(count * 3);
  for (let i = 0; i < count; i++) {
    const u = Math.random() * 2 - 1;
    const t = Math.random() * Math.PI * 2;
    const r = Math.sqrt(1 - u * u);
    pos[i * 3] = Math.cos(t) * r * radius;
    pos[i * 3 + 1] = u * radius;
    pos[i * 3 + 2] = Math.sin(t) * r * radius;
  }
  const geom = new THREE.BufferGeometry();
  geom.setAttribute('position', new THREE.BufferAttribute(pos, 3));
  const mat = new THREE.PointsMaterial({ color: 0xdfe6f2, size: 2, sizeAttenuation: false });
  const pts = new THREE.Points(geom, mat);
  pts.frustumCulled = false;
  return pts;
}

// The elite-ship renderer plugin. build() fetches one ship's HSR blueprint (the same per-ship shape
// as one entry of the old ships.json), builds a ShipMesh + starfield, and takes over the frame so
// the render goes through Elite's low-res post-process — the exact path the old ShipViewer ran in
// the Studio (lowRes on, internal CRT off; the global screen filter draws the tube).
export default {
  kind: 'elite-ship',
  async build({ item, base, stage }) {
    const ship = await fetch(base + item.data).then((r) => r.json());
    const mesh = new ShipMesh(ship);
    const stars = makeStarfield(500, 60000);
    stars.material.size = 1; // one chunky pixel in lo-res, not a blob (lowRes is always on here)

    const { scene, camera, controls, renderer } = stage;
    stage.add(mesh.object);
    stage.add(stars);

    // ---- post pipeline (render targets, blur/CRT materials, blit + glow — moved verbatim) ----
    const rt = (w, h) => new THREE.WebGLRenderTarget(w, h, {
      minFilter: THREE.LinearFilter, magFilter: THREE.LinearFilter, depthBuffer: true,
    });
    const postTarget = rt(320, LORES_H);   // the (low-res) scene
    const glowA = rt(320, GLOW_H);         // glow ping-pong buffers (fixed size)
    const glowB = rt(320, GLOW_H);
    const blurMaterial = new THREE.ShaderMaterial({
      uniforms: { tSrc: { value: null }, uTexel: { value: new THREE.Vector2() } },
      vertexShader: CRT_VERT, fragmentShader: BLUR_FRAG, depthTest: false, depthWrite: false,
    });
    const postMaterial = new THREE.ShaderMaterial({
      uniforms: {
        tScene: { value: postTarget.texture },
        tGlow: { value: glowA.texture },
        uSceneRes: { value: new THREE.Vector2(320, LORES_H) },
        uCRT: { value: 0 },
      },
      vertexShader: CRT_VERT, fragmentShader: CRT_FRAG, depthTest: false, depthWrite: false,
    });
    const postCam = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
    const quad = new THREE.Mesh(new THREE.PlaneGeometry(2, 2), postMaterial);
    const quadScene = new THREE.Scene();
    quadScene.add(quad);

    // _blit draws a material to a target via the shared full-screen quad.
    const blit = (material, target) => {
      quad.material = material;
      renderer.setRenderTarget(target);
      renderer.render(quadScene, postCam);
    };
    // _buildGlow renders the scene small and Gaussian-blurs it (two separable iterations) into
    // glowA — a smooth, round, ring-free halo source. (Only used when the internal CRT is on; the
    // Studio keeps it off, so this stays exactly as it was, ready if the effect is ever enabled.)
    const buildGlow = () => {
      const gw = glowA.width, gh = glowA.height;
      renderer.setRenderTarget(glowA);
      renderer.render(scene, camera); // a small render of the same scene
      for (let i = 0; i < 2; i++) {
        blurMaterial.uniforms.tSrc.value = glowA.texture;
        blurMaterial.uniforms.uTexel.value.set(1 / gw, 0);
        blit(blurMaterial, glowB);
        blurMaterial.uniforms.tSrc.value = glowB.texture;
        blurMaterial.uniforms.uTexel.value.set(0, 1 / gh);
        blit(blurMaterial, glowA);
      }
    };
    void buildGlow; // preserved verbatim; not invoked while the internal CRT is off (Studio default)

    // Size the off-screen targets: low (LORES_H tall) for the chunky C64 pixels, width following
    // the viewport aspect (square pixels). Stage3D handles the canvas + camera-aspect resize; here
    // we only track the target sizes. (The old _applyResolution's lowRes branch, verbatim.)
    const applyResolution = () => {
      const w = stage.el.clientWidth, h = stage.el.clientHeight;
      if (!w || !h) return;
      const tw = Math.max(2, Math.round(LORES_H * w / h));
      const th = LORES_H;
      postTarget.setSize(tw, th);
      postMaterial.uniforms.uSceneRes.value.set(tw, th);
      const gw = Math.max(2, Math.round(GLOW_H * w / h));
      glowA.setSize(gw, GLOW_H);
      glowB.setSize(gw, GLOW_H);
    };
    applyResolution();
    const ro = new ResizeObserver(() => applyResolution());
    ro.observe(stage.el);

    // ---- orbit controls: Elite's feel (idle spin that stops once the user grabs the ship) ----
    controls.enablePan = false;
    controls.rotateSpeed = 0.9;
    controls.zoomSpeed = 4.0;
    controls.autoRotate = true;
    controls.autoRotateSpeed = AUTO_ROTATE_SPEED;
    const onGrab = () => { controls.autoRotate = false; };
    controls.addEventListener('start', onGrab);

    // Fit the camera, then apply Elite's zoom limits. Keep the eye just outside the bounding
    // sphere: with the per-face test, entering the hull is what makes faces flip abruptly.
    stage.frame(mesh.object);
    const dist = fitDistance(mesh.radius, camera.fov);
    controls.minDistance = mesh.radius * 1.1;
    controls.maxDistance = dist * 3;
    controls.autoRotate = true;
    controls.update();

    // The plugin owns the frame: recompute HSR for this camera, render the scene into the low-res
    // target, then composite it to the canvas via the post shader in plain chunky-upscale mode
    // (uCRT = 0 — the global screen filter draws the CRT tube on top). Verbatim from the old tick's
    // taken branch (lowRes on, internal crt off, flicker off → updateForCamera(pos, -1)).
    stage.render = (s) => {
      mesh.updateForCamera(s.camera.position);
      s.renderer.setRenderTarget(postTarget);
      s.renderer.render(s.scene, s.camera);
      postMaterial.uniforms.uCRT.value = 0;
      blit(postMaterial, null);
    };

    // The native C64-resolution pixel grid the low-res render presents, for the global CRT filter
    // to match its scanlines/mask to. One native pixel = clientHeight / LORES_H CSS px (no pan).
    stage.pixelGrid = () => {
      const w = stage.el.clientWidth, h = stage.el.clientHeight;
      if (!w || !h) return null;
      return { cell: h / LORES_H, ox: 0, oy: 0, ref: w };
    };

    // HUD detail line (as the old loadShip set it).
    stage.hud = `${ship.name}  ·  type ${ship.type}  ·  ${ship.verts.length} verts  ${ship.edges.length} edges  ${ship.faces.length} faces`;

    // Teardown: free this ship's GPU resources and unhook the observers/listeners before the next
    // item builds (the Viewer calls this before stage.clear()).
    stage.disposePlugin = () => {
      ro.disconnect();
      controls.removeEventListener('start', onGrab);
      mesh.dispose();
      stars.geometry.dispose();
      stars.material.dispose();
      postTarget.dispose();
      glowA.dispose();
      glowB.dispose();
      blurMaterial.dispose();
      postMaterial.dispose();
      quad.geometry.dispose();
      renderer.setRenderTarget(null); // leave the renderer targeting the screen for the next plugin
    };

    return mesh.object;
  },
};
