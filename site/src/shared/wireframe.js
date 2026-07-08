// Wireframe toggle for the 3-D viewers (a Studio "Wireframe" layer). Traverses a scene and sets
// material.wireframe on every Mesh material; LineSegments / Points (Elite ships, drive lines,
// slope markers — already edges/dots) have no `wireframe` and are left alone. The Studio re-applies
// every layer after each load, so a newly-loaded model inherits the current wireframe state.
export function applyWireframe(root, on) {
  if (!root) return;
  root.traverse((o) => {
    if (!o.isMesh || !o.material) return;
    const mats = Array.isArray(o.material) ? o.material : [o.material];
    for (const m of mats) if (m && 'wireframe' in m) m.wireframe = on;
  });
}
