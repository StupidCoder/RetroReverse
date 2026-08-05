// xrsupport.js — "is there AR here?", and nothing else.
//
// Its own module because it must import NOTHING. The shell needs the answer to
// decide whether to offer the button, on 2-D pages as much as 3-D ones, and
// asking xr.js would drag in engine3d.js and with it the whole of three.js —
// which a tilemap page has no other reason to load. The 2-D views build their
// three.js form only inside their AR content's build(), so the cost lands when
// somebody actually presses the button.
//
// One probe per page load, awaited by whoever needs it. `navigator.xr` is
// undefined outside a secure context, which is the usual reason the button
// never appears.
const probe = (mode) => (async () => {
  try {
    if (!navigator.xr?.isSessionSupported) return false;
    return await navigator.xr.isSessionSupported(mode);
  } catch {
    return false;
  }
})();

export const arSupported = probe('immersive-ar');

// The other half: walking around inside a level (xrplacer.js worldPlace) needs
// an OPAQUE session, so it asks for immersive-vr. A headset generally answers
// yes to both and a phone to neither, but they are separate questions and the
// two buttons are gated separately — a session's mode cannot be changed without
// ending it, so the choice is made before there is one.
export const vrSupported = probe('immersive-vr');
