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

// The engine's night-sky clear colour (the pre-race preview's backdrop). Shared by every item so
// the whole game reads as one scene.
const SKY = 0x0a0d12;
// Native Amiga vertical resolution: the game runs a 320×200 lo-res playfield. Locking the render
// to 200 lines (width follows the viewport aspect for square pixels) reproduces the chunky raster.
const NATIVE_H = 200;
const AUTO_ROTATE_SPEED = 0.9; // idle spin for the car / horizon

let _gltf = null;
const gltfLoader = () => (_gltf || (_gltf = new GLTFLoader()));

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

    // On a circuit, drop the opponent car GLB onto the start/finish line. Both GLBs share the
    // 1/1024 scale and Y-up/Z-negated convention (webexport models.go), so the car goes in at
    // scale 1: item.start.pos is the road-surface midpoint of the finish rung and item.start.yaw
    // faces the car (local forward -Z) along the direction of travel. Placeholder placement —
    // the car is static for now, not yet driven by the physics.
    let car = null;
    if (item.fly && item.start) {
      const carGltf = await gltfLoader().loadAsync(base + 'models/opponent-car.glb');
      car = carGltf.scene;
      car.position.set(item.start.pos[0], item.start.pos[1], item.start.pos[2]);
      car.rotation.y = item.start.yaw;
      stage.add(car);
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

    // ---- controls: circuits fly, the car / horizon orbit ----
    let flycam = null;
    if (item.fly) {
      controls.autoRotate = false;
      flycam = new FlyCam(camera, controls, stage.el);
      flycam.setScale(size);
      flycam.setMoveScale(1.4);
      flycam.setEnabled(true);
      stage.fly = flycam;
      stage.hud = `${item.name} · ${flyHint}`;
    } else {
      controls.autoRotate = true;
      controls.autoRotateSpeed = AUTO_ROTATE_SPEED;
      stage.hud = item.name;
    }

    // Per-frame: advance the fly-cam / animation, then the plugin owns the frame — render the scene
    // into the low-res target and upscale it to the canvas.
    stage.onFrame = (camPos, dt) => {
      if (flycam) flycam.update(dt);
      if (mixer) mixer.update(dt);
    };
    stage.render = (s) => {
      s.renderer.setRenderTarget(postTarget);
      s.renderer.render(s.scene, s.camera);
      s.renderer.setRenderTarget(null);
      s.renderer.render(quadScene, postCam);
    };

    // The native Amiga pixel grid the low-res render presents, so the global CRT filter can match
    // its scanlines/mask to it. One native pixel = clientHeight / 200 CSS px (no pan).
    stage.pixelGrid = () => {
      const w = stage.el.clientWidth, h = stage.el.clientHeight;
      if (!w || !h) return null;
      return { cell: h / NATIVE_H, ox: 0, oy: 0, ref: w };
    };

    // Teardown: free this item's GPU resources and unhook the fly-cam / observer before the next
    // item builds (the Viewer calls this before stage.clear()).
    stage.disposePlugin = () => {
      ro.disconnect();
      if (flycam) flycam.dispose();
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
