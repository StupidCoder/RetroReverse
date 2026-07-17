// The Studio renderer plugin for kind "nfs-movie": the disc's streamed FMV,
// exported by webexport -movies as an H.264 MP4 (our own Cinepak decoder does
// the video; ffmpeg only re-encodes the decoded frames). It plays the clip on a
// full-frame quad, letterboxed to the movie's aspect, owning the stage's render
// so the flying-camera scene is bypassed.
import * as THREE from 'three';

export default {
  kind: 'nfs-movie',
  async build({ item, base, stage }) {
    const video = document.createElement('video');
    video.src = base + item.file;
    video.loop = true;
    video.muted = true;        // muted autoplay is allowed without a user gesture
    video.playsInline = true;
    // Chrome will not decode a detached <video> for a texture, so park it in the
    // DOM off-screen (a 1px hidden element is enough to keep it decoding).
    video.style.cssText = 'position:fixed;left:-10px;top:-10px;width:1px;height:1px;opacity:0;pointer-events:none';
    document.body.appendChild(video);
    video.play().catch(() => {});

    const tex = new THREE.VideoTexture(video);
    tex.colorSpace = THREE.SRGBColorSpace;
    tex.minFilter = THREE.LinearFilter;
    tex.magFilter = THREE.NearestFilter; // keep the 320-wide pixels crisp when upscaled

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x000000);
    const cam = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
    const quad = new THREE.Mesh(
      new THREE.PlaneGeometry(2, 2),
      new THREE.MeshBasicMaterial({ map: tex }),
    );
    scene.add(quad);

    const vw = item.w || 320, vh = item.h || 240;
    // Own the frame: letterbox the movie into the canvas each render.
    stage.render = (s) => {
      const cw = s.renderer.domElement.width || 1;
      const ch = s.renderer.domElement.height || 1;
      const canvasAspect = cw / ch, vidAspect = vw / vh;
      let sx = 1, sy = 1;
      if (canvasAspect > vidAspect) sx = vidAspect / canvasAspect;
      else sy = canvasAspect / vidAspect;
      quad.scale.set(sx, sy, 1);
      s.renderer.render(scene, cam);
    };
    stage.hud = `${vw}×${vh} · Cinepak FMV`;
    stage.disposePlugin = () => {
      video.pause();
      video.removeAttribute('src');
      video.load();
      video.remove();
      tex.dispose();
      quad.geometry.dispose();
      quad.material.dispose();
    };
    return quad;
  },
};
