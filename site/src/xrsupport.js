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
export const arSupported = (async () => {
  try {
    if (!navigator.xr?.isSessionSupported) return false;
    return await navigator.xr.isSessionSupported('immersive-ar');
  } catch {
    return false;
  }
})();
