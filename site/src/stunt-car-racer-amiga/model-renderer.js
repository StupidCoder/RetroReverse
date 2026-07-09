// Stunt Car Racer model renderer — the `stunt-model` plugin for the shared Viewer3D/Stage3D seam.
// It replaces the old bespoke `stunt-track` wireframe builder: the geometry now arrives as a baked
// GLB (webexport cmd/webexport/models.go), so this plugin just loads the glTF, drops it on the
// stage and handles the presentation. Every circuit, the opponent car and the horizon ring are
// exported the same way — the road ribbon, side walls and stroked decal LINES are all in the file,
// coloured exactly as the game's pre-race preview renders them (byte-verified by cmd/coloracle),
// so there is no per-vertex work left to do here.
//
// Two presentation modes, chosen by the manifest entry:
//   • a circuit (item.fly): you fly through it with the shared FlyCam (WASD to move, arrows to
//     look, mouse drag orbits) — the same free-flight the other 3-D level viewers use.
//   • the car / horizon (no fly): a slow auto-rotating orbit, as the ship/model viewers present
//     small models.
// The Draw Bridge circuit ships a morph-target animation ($5A794, cmd/bridgeoracle); any GLB that
// carries animation clips is driven by an AnimationMixer each frame.
//
// Resolution: the whole viewer renders at the Amiga's native vertical resolution (200 lines). The
// scene is drawn into a 200-tall off-screen target with the width following the viewport aspect
// (square pixels), then upscaled to the canvas with NEAREST sampling for chunky Amiga pixels. The
// native pixel grid is published via stage.pixelGrid so the Studio's global CRT filter locks its
// scanlines/mask to it, exactly as Elite's low-res raster does.
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { FlyCam, flyHint } from '../shared/flycam.js';
import { Physics } from './physics.js';

// Circuit slugs in physics track-id order (physdump / package physics), so the driven car
// loads the right per-track placed state.
const TRACK_SLUGS = ['little-ramp', 'stepping-stones', 'hump-back', 'big-ramp',
  'ski-jump', 'draw-bridge', 'high-jump', 'roller-coaster'];
// physics world -> GLB world: the sim tracks the car at grid*128 (128 units/cell) in 16.16
// fixed point; the baked GLB is grid*2048/1024 = 2 units/cell. So GLB = (physics 16.16) /
// (65536*64). Verified by overlaying the drive path on the track top-down.
const PHYS_TO_GLB = 1 / (65536 * 64);

// The engine's night-sky clear colour (the pre-race preview's backdrop). Shared by every item so
// the whole game reads as one scene.
const SKY = 0x0a0d12;
// Native Amiga vertical resolution: the game runs a 320×200 lo-res playfield. Locking the render
// to 200 lines (width follows the viewport aspect for square pixels) reproduces the chunky raster.
const NATIVE_H = 200;
const AUTO_ROTATE_SPEED = 0.9; // idle spin for the car / horizon

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

// setupDrive loads a circuit's placed physics state and returns a driver that steps the sim
// each frame from keyboard input and maps the car onto the GLB track. The physics is the
// verified vehicle-control layer (physics.js driveTickCoupled: input decode -> physics with
// the $5B32E spawn -> coupling -> timer); the world<->GLB mapping and the on-road height
// (raycast) are the presentation layer. The car spawns (~240 frames) then drives.
async function setupDrive(id, base, car, track) {
  const dir = base + 'phys/';
  const [perTrack, staticBuf] = await Promise.all([
    fetch(`${dir}${id}.bin`).then((r) => { if (!r.ok) throw new Error('no bin'); return r.arrayBuffer(); }),
    fetch(`${dir}static.bin`).then((r) => { if (!r.ok) throw new Error('no static'); return r.arrayBuffer(); }),
  ]);
  const phys = new Physics();
  phys.loadTrack(perTrack, staticBuf);

  const keys = new Set();
  const kd = (e) => keys.add(e.key.toLowerCase());
  const ku = (e) => keys.delete(e.key.toLowerCase());
  window.addEventListener('keydown', kd);
  window.addEventListener('keyup', ku);

  const ray = new THREE.Raycaster();
  const down = new THREE.Vector3(0, -1, 0);
  let px = null, pz = null, yaw = 0, y = 0;

  return {
    step() {
      // build the joystick byte $1BB47: bits 0-1 throttle/brake, 2-3 steer, 4 fire.
      // IJKL for the car (the fly-cam owns WASD / arrows): I throttle, K brake, J/L steer, O fire.
      let inp = 0;
      if (keys.has('i')) inp |= 0x01;
      else if (keys.has('k')) inp |= 0x02;
      if (keys.has('j')) inp |= 0x04;
      else if (keys.has('l')) inp |= 0x08;
      if (keys.has('o')) inp |= 0x10;
      const speed = phys.driveTickCoupled(inp);

      const gx = phys.l(0x1BCD8) * PHYS_TO_GLB, gz = -phys.l(0x1BCE0) * PHYS_TO_GLB;
      if (px !== null && (gx - px) ** 2 + (gz - pz) ** 2 > 1e-6) {
        yaw = Math.atan2(-(gx - px), -(gz - pz)); // face along motion (car forward = -Z local)
      }
      px = gx; pz = gz;
      ray.set(new THREE.Vector3(gx, 200, gz), down);
      const hits = ray.intersectObject(track, true);
      if (hits.length) y = hits[0].point.y;
      car.position.set(gx, y, gz);
      car.rotation.y = yaw;
      return { gx, gz, y, yaw, speed };
    },
    dispose() {
      window.removeEventListener('keydown', kd);
      window.removeEventListener('keyup', ku);
    },
  };
}

export default {
  kind: 'stunt-model',
  async build({ item, base, stage }) {
    const gltf = await gltfLoader().loadAsync(base + item.file);
    const model = gltf.scene;

    // Hidden-line removal for the coplanar decals. The GLB carries the road/walls as TRIANGLES
    // and the curb strokes, wall struts and start/finish stripe as LINES lying exactly on those
    // faces — so without help they z-fight. glTF can't express a polygon-offset render state, so
    // (as the old stunt-track renderer did for its fill/walls) we push every triangle mesh back a
    // touch in depth; the offset-free lines then always resolve in front of the surface they sit on.
    model.traverse((o) => {
      if (!o.isMesh) return; // Mesh = TRIANGLES; LineSegments (o.isLine) stay un-offset, in front
      for (const mat of Array.isArray(o.material) ? o.material : [o.material]) {
        mat.polygonOffset = true;
        mat.polygonOffsetFactor = 1;
        mat.polygonOffsetUnits = 1;
      }
    });

    const { scene, camera, controls, renderer } = stage;
    scene.background = new THREE.Color(SKY);
    stage.add(model);

    // On a circuit, add the opponent car GLB and drive it with the verified physics (WASD /
    // arrows). If the per-track physics data loads, the car spawns and drives; otherwise it
    // falls back to a static placement at the start line (item.start).
    let car = null, drive = null;
    if (item.fly) {
      const carGltf = await gltfLoader().loadAsync(base + 'models/opponent-car.glb');
      car = carGltf.scene;
      stage.add(car);
      const slug = item.file.replace(/^.*\//, '').replace(/\.glb$/, '');
      const id = TRACK_SLUGS.indexOf(slug);
      if (id >= 0) {
        try {
          drive = await setupDrive(id, base, car, model);
        } catch (e) {
          console.warn('stunt drive: physics data unavailable, static car', e);
        }
      }
      if (!drive && item.start) {
        car.position.set(item.start.pos[0], item.start.pos[1], item.start.pos[2]);
        car.rotation.y = item.start.yaw;
      }
    }

    // ---- native-resolution post: render the scene into a 200-line target, upscale chunky ----
    // The target's texture is tagged sRGB so the scene is encoded into it and decoded back out by
    // the upscale quad — a colour-exact round trip (the GLB materials are unlit, linear-light
    // baseColorFactors). NEAREST keeps the Amiga pixels square-edged.
    const postTarget = new THREE.WebGLRenderTarget(2, NATIVE_H, {
      minFilter: THREE.NearestFilter, magFilter: THREE.NearestFilter, depthBuffer: true,
    });
    postTarget.texture.colorSpace = THREE.SRGBColorSpace;
    const postCam = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
    const quad = new THREE.Mesh(
      new THREE.PlaneGeometry(2, 2),
      new THREE.MeshBasicMaterial({ map: postTarget.texture, depthTest: false, depthWrite: false }),
    );
    const quadScene = new THREE.Scene();
    quadScene.add(quad);

    // Size the off-screen target: 200 tall, width following the viewport aspect (square pixels).
    const applyResolution = () => {
      const w = stage.el.clientWidth, h = stage.el.clientHeight;
      if (!w || !h) return;
      postTarget.setSize(Math.max(2, Math.round(NATIVE_H * w / h)), NATIVE_H);
    };
    applyResolution();
    const ro = new ResizeObserver(() => applyResolution());
    ro.observe(stage.el);

    // Fit the camera to the model from the shared 3/4 angle.
    stage.frame(model);
    const size = new THREE.Box3().setFromObject(model).getSize(new THREE.Vector3()).length() || 10;

    // ---- animation (Draw Bridge morph, and any future clip) ----
    let mixer = null;
    if (gltf.animations && gltf.animations.length) {
      mixer = new THREE.AnimationMixer(model);
      for (const clip of gltf.animations) mixer.clipAction(clip).play();
    }

    // ---- controls: circuits fly (WASD / arrows), the car / horizon orbit. A driven circuit
    // keeps the fly-cam — the physics steps the car independently, the camera stays yours. ----
    let flycam = null;
    if (item.fly) {
      controls.autoRotate = false;
      flycam = new FlyCam(camera, controls, stage.el);
      flycam.setScale(size);
      flycam.setMoveScale(1.4);
      flycam.setEnabled(true);
      stage.fly = flycam;
      stage.hud = `${item.name} · ${flyHint}${drive ? ' · IJKL drive' : ''}`;
    } else {
      controls.autoRotate = true;
      controls.autoRotateSpeed = AUTO_ROTATE_SPEED;
      stage.hud = item.name;
    }

    // Optional low-res render: the 200-line target + chunky upscale is authentic but distracting
    // when judging exactness. Driven by the Studio's "Low res" layer toggle (applyLayers ->
    // setLayer); pixelGrid()/stage.render re-query it each frame.
    let lowRes = true;
    stage.setLayer = (id, on) => { if (id === 'lowres') lowRes = on; };

    // Per-frame: fixed-50 Hz physics for the driven car (the sim is a fixed-timestep tick),
    // decoupled from the render dt via an accumulator. The camera is the fly-cam (or orbit) —
    // the car drives on its own; it does NOT drive the camera.
    let acc = 0;
    const DT = 1 / 50;
    stage.onFrame = (camPos, dt) => {
      if (flycam) flycam.update(dt);
      if (mixer) mixer.update(dt);
      if (drive) {
        acc += Math.min(dt, 0.25);
        while (acc >= DT) { drive.step(); acc -= DT; }
      }
    };
    stage.render = (s) => {
      if (lowRes) {
        s.renderer.setRenderTarget(postTarget);
        s.renderer.render(s.scene, s.camera);
        s.renderer.setRenderTarget(null);
        s.renderer.render(quadScene, postCam);
      } else {
        s.renderer.setRenderTarget(null);
        s.renderer.render(s.scene, s.camera); // full-resolution, no upscale
      }
    };

    // The native Amiga pixel grid the low-res render presents, so the global CRT filter can match
    // its scanlines/mask to it. One native pixel = clientHeight / 200 CSS px (no pan). Null when
    // the low-res render is off, so the CRT filter drops its scanlines to match.
    stage.pixelGrid = () => {
      if (!lowRes) return null;
      const w = stage.el.clientWidth, h = stage.el.clientHeight;
      if (!w || !h) return null;
      return { cell: h / NATIVE_H, ox: 0, oy: 0, ref: w };
    };

    // Teardown: free this item's GPU resources and unhook the fly-cam / observer before the next
    // item builds (the Viewer calls this before stage.clear()).
    stage.disposePlugin = () => {
      ro.disconnect();
      stage.setLayer = null;
      if (flycam) flycam.dispose();
      if (drive) drive.dispose();
      if (mixer) mixer.stopAllAction();
      controls.autoRotate = false;
      renderer.setRenderTarget(null);
      const disposeTree = (root) => root.traverse((o) => {
        if (o.geometry) o.geometry.dispose();
        if (o.material) {
          for (const m of Array.isArray(o.material) ? o.material : [o.material]) m.dispose();
        }
      });
      disposeTree(model);
      if (car) disposeTree(car);
      postTarget.dispose();
      quad.geometry.dispose();
      quad.material.dispose();
    };

    return model;
  },
};
