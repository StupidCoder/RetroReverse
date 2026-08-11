// shadowmap.js — a static shadow map cast by the guest's own light, off the
// guest's own caster geometry.
//
// Some games ship a *depth-shadow proxy*: a coarse second model of each object
// whose only job is to be rendered from the light into a depth buffer that the
// visible surfaces then sample. Captain Toad's stages carry one per object, and
// the level document publishes them as a layer with role "shadow". Given that
// and a directional light in scene.lights, this renders the proxy once and
// multiplies every other layer's fragments by the comparison.
//
// It is deliberately its own pass rather than three.js's shadow machinery. The
// receivers are *unlit* materials — the guest's lighting is baked into their
// vertex colours, because the hardware multiplies in gamma space — and three.js
// applies shadows inside a lighting path a MeshBasicMaterial does not have.
//
// Two things make the result exact rather than decorative:
//
//   * A shadow is a MULTIPLY, and a multiply is the one operation that survives
//     the gamma/linear round trip unchanged. The injection sits after
//     <colorspace_fragment>, where gl_FragColor is already in the output's gamma
//     space — the same space the guest's combiner multiplies in.
//   * A shadow removes ONE light's contribution — the one casting it — and
//     leaves every other light at full strength. A shadowed surface here is
//     still lit by the sky. The baked vertex colour holds the whole rig,
//     `ambient + Σ light·N·L`, so the factor to apply is
//
//         (everything except the caster) / (everything)
//
//     which depends on the surface normal and is therefore computed per
//     fragment, from the same colours the bake used, rather than being a
//     constant "shadow floor". Two consequences fall out for free: the factor
//     reaches 1 exactly where the caster's N·L reaches 0, so a face already
//     turned away from that light is not darkened twice, and a second light
//     that does not cast keeps its contribution inside the shadow.
import * as THREE from 'three';

const TEX = 2048;

function srgbColor(hex) {
  const c = new THREE.Color();
  c.setStyle(hex, THREE.NoColorSpace); // these are gamma-space multipliers, not colours to convert
  return new THREE.Vector3(c.r, c.g, c.b);
}

// MAXDIR bounds the directional lights the shadow factor accounts for. Every
// rig seen so far is one key plus an ambient; a rig with more than this many
// directionals would need the factor's denominator extended, so it is a loud
// limit rather than a silent truncation.
const MAXDIR = 4;

// rigFromLights pulls a level document's light list apart into the ambient, the
// light that casts the shadow, and any other directionals that keep shining
// inside it — all in the gamma space the guest works in.
//
// The caster is the first directional. That is the guest's own arrangement here:
// a stage's light area names either the shadowing key or its "_noshadow" twin,
// which are the same colour and direction and differ only in this.
export function rigFromLights(lights) {
  if (!lights?.length) return null;
  const ambient = new THREE.Vector3();
  const dirs = [];
  for (const l of lights) {
    const k = l.keys?.[0];
    if (!k) continue;
    if (l.type === 'ambient') ambient.add(srgbColor(k.color));
    else if (l.type === 'directional' && k.dir) {
      dirs.push({ color: srgbColor(k.color), dir: new THREE.Vector3(...k.dir).normalize() });
    }
  }
  if (!dirs.length) return null;
  if (dirs.length > MAXDIR) {
    console.warn(`shadowmap: ${dirs.length} directional lights, only ${MAXDIR} accounted for`);
    dirs.length = MAXDIR;
  }
  return { ambient, key: dirs[0], others: dirs.slice(1) };
}

// buildShadowMap renders the caster group from the light and returns the
// uniforms the receivers need, or null when there is nothing to cast.
export function buildShadowMap(renderer, casterGroup, rig) {
  if (!casterGroup || !rig) return null;
  const box = new THREE.Box3().setFromObject(casterGroup);
  if (box.isEmpty()) return null;

  const centre = box.getCenter(new THREE.Vector3());
  const radius = box.getBoundingSphere(new THREE.Sphere()).radius;
  const dir = rig.key.dir;
  const eye = centre.clone().addScaledVector(dir, radius * 2);
  const up = Math.abs(dir.y) > 0.99 ? new THREE.Vector3(0, 0, 1) : new THREE.Vector3(0, 1, 0);
  const cam = new THREE.OrthographicCamera(-radius, radius, radius, -radius, 0.01, radius * 4);
  cam.position.copy(eye);
  cam.up.copy(up);
  cam.lookAt(centre);
  cam.updateMatrixWorld(true);
  cam.updateProjectionMatrix();

  const target = new THREE.WebGLRenderTarget(TEX, TEX, {
    minFilter: THREE.NearestFilter,
    magFilter: THREE.NearestFilter,
    depthBuffer: true,
  });

  const scene = new THREE.Scene();
  const clone = casterGroup.clone(true);
  clone.traverse((o) => { o.visible = true; });
  clone.visible = true;
  scene.add(clone);
  scene.overrideMaterial = new THREE.MeshDepthMaterial({
    depthPacking: THREE.RGBADepthPacking,
    side: THREE.DoubleSide,
  });

  const prevTarget = renderer.getRenderTarget();
  const prevClear = renderer.getClearColor(new THREE.Color());
  const prevAlpha = renderer.getClearAlpha();
  renderer.setRenderTarget(target);
  renderer.setClearColor(0xffffff, 1); // an untouched texel is the far plane: lit
  renderer.clear();
  renderer.render(scene, cam);
  renderer.setRenderTarget(prevTarget);
  renderer.setClearColor(prevClear, prevAlpha);
  scene.overrideMaterial.dispose();

  const matrix = new THREE.Matrix4().multiplyMatrices(cam.projectionMatrix, cam.matrixWorldInverse);
  return {
    uniforms: {
      uShadowMap: { value: target.texture },
      uShadowMatrix: { value: matrix },
      uShadowBias: { value: 2.5 / (radius * 4) },
      uAmbient: { value: rig.ambient },
      uKeyColor: { value: rig.key.color },
      uKeyDir: { value: dir },
      // The lights that keep shining inside the shadow.
      uOtherCount: { value: rig.others.length },
      uOtherColor: { value: padVecs(rig.others.map((o) => o.color)) },
      uOtherDir: { value: padVecs(rig.others.map((o) => o.dir)) },
    },
    dispose() { target.dispose(); },
  };
}

// receiveShadows patches every material under root to multiply its output by
// the shadow comparison.
function padVecs(v) {
  const out = v.slice();
  while (out.length < MAXDIR) out.push(new THREE.Vector3());
  return out;
}

export function receiveShadows(root, shadow) {
  if (!shadow) return;
  root.traverse((o) => {
    if (!o.isMesh || !o.material) return;
    for (const mat of Array.isArray(o.material) ? o.material : [o.material]) {
      if (mat.userData.shadowPatched) continue;
      mat.userData.shadowPatched = true;
      mat.onBeforeCompile = (shader) => {
        Object.assign(shader.uniforms, shadow.uniforms);
        shader.vertexShader = shader.vertexShader
          .replace('#include <common>', `#include <common>
uniform mat4 uShadowMatrix;
varying vec4 vShadowCoord;
varying vec3 vShadowNormal;`)
          .replace('#include <project_vertex>', `#include <project_vertex>
vec4 shadowWorld = modelMatrix * vec4(transformed, 1.0);
vShadowCoord = uShadowMatrix * shadowWorld;
vShadowNormal = mat3(modelMatrix) * normal;`);

        shader.fragmentShader = shader.fragmentShader
          .replace('#include <common>', `#include <common>
#include <packing>
uniform sampler2D uShadowMap;
uniform float uShadowBias;
uniform vec3 uAmbient;
uniform vec3 uKeyColor;
uniform vec3 uKeyDir;
uniform int uOtherCount;
uniform vec3 uOtherColor[${MAXDIR}];
uniform vec3 uOtherDir[${MAXDIR}];
varying vec4 vShadowCoord;
varying vec3 vShadowNormal;`)
          // After <colorspace_fragment>: gl_FragColor is in the output's gamma
          // space here, which is the space the guest's combiner multiplies in.
          .replace('#include <dithering_fragment>', `
{
  vec3 sc = vShadowCoord.xyz / vShadowCoord.w * 0.5 + 0.5;
  if (sc.x > 0.0 && sc.x < 1.0 && sc.y > 0.0 && sc.y < 1.0 && sc.z < 1.0) {
    float lit = 0.0;
    vec2 texel = vec2(1.0 / ${TEX}.0);
    for (int y = -1; y <= 1; y++) {
      for (int x = -1; x <= 1; x++) {
        float d = unpackRGBAToDepth(texture2D(uShadowMap, sc.xy + vec2(float(x), float(y)) * texel));
        lit += (sc.z - uShadowBias <= d) ? 1.0 : 0.0;
      }
    }
    lit /= 9.0;
    // What the shadow takes away is the caster's term alone. Everything else
    // in the rig — the ambient, and any light that does not cast — keeps its
    // full strength inside the shadow, which is what the hardware's shadow
    // attenuation does: it scales one light's contribution, not the sum.
    vec3 n = normalize(vShadowNormal);
    vec3 kept = uAmbient;
    for (int i = 0; i < ${MAXDIR}; i++) {
      if (i >= uOtherCount) break;
      kept += uOtherColor[i] * max(dot(n, uOtherDir[i]), 0.0);
    }
    vec3 full = min(vec3(1.0), kept + uKeyColor * max(dot(n, uKeyDir), 0.0));
    vec3 shaded = min(vec3(1.0), kept);
    vec3 ratio = shaded / max(full, vec3(1e-4));
    gl_FragColor.rgb *= mix(min(ratio, vec3(1.0)), vec3(1.0), lit);
  }
}
#include <dithering_fragment>`);
      };
      mat.needsUpdate = true;
    }
  });
}
