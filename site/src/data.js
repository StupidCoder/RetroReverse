// data.js — Retro-X document loading. The viewer knows the FORMAT, never a
// game: everything below is generic over index.json / manifest.json and the
// level/object/script documents they reference.

export const RETROX_VERSION = 1;

function checkHeader(doc, path) {
  if (doc.format !== 'retro-x') throw new Error(`${path}: not a Retro-X document`);
  if (!(doc.version >= 1 && doc.version <= RETROX_VERSION)) {
    throw new Error(`${path}: version ${doc.version} unsupported (viewer speaks ${RETROX_VERSION})`);
  }
  return doc;
}

async function getJSON(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(`${url}: HTTP ${r.status}`);
  return r.json();
}

// resolveRef joins a file reference to the directory of the document that
// carries it (RETROX.md §2: refs are doc-relative).
export function resolveRef(docPath, ref) {
  const dir = docPath.includes('/') ? docPath.slice(0, docPath.lastIndexOf('/') + 1) : '';
  return dir + ref;
}

export class Game {
  constructor(root, manifest) {
    this.root = root;           // "public/<id>/"
    this.man = manifest;
    this.byId = new Map(manifest.assets.map((a) => [a.id, a]));
    this._docs = new Map();     // asset/doc path -> parsed document
  }

  get id() { return this.man.id; }
  get display() { return this.man.display || { native: { w: 320, h: 240 }, tickHz: 60 }; }

  asset(id) { return this.byId.get(id); }
  assets(category) { return this.man.assets.filter((a) => a.category === category); }
  categories() {
    const seen = [];
    for (const a of this.man.assets) if (!seen.includes(a.category)) seen.push(a.category);
    return seen;
  }

  // url resolves a reference made by a document at docPath ('' = the manifest).
  url(docPath, ref) { return this.root + resolveRef(docPath, ref); }

  // doc fetches + caches a Retro-X JSON document referenced from docPath.
  async doc(docPath, ref) {
    const p = resolveRef(docPath, ref);
    if (!this._docs.has(p)) {
      this._docs.set(p, getJSON(this.root + p).then((d) => checkHeader(d, p)));
    }
    return this._docs.get(p);
  }

  // assetDoc fetches an asset's own document (levels and objects).
  assetDoc(asset) { return this.doc('', asset.file); }

  // objectFor resolves an object asset id into { asset, doc, docPath }.
  async object(id) {
    const asset = this.byId.get(id);
    if (!asset || asset.category !== 'object') throw new Error(`no object asset ${id}`);
    return { asset, doc: await this.assetDoc(asset), docPath: asset.file };
  }
}

// loadIndex loads the publication root and every game's manifest.
export async function loadIndex(root = 'public/') {
  const idx = checkHeader(await getJSON(root + 'index.json'), 'index.json');
  const games = await Promise.all(idx.games.map(async (id) => {
    try {
      const man = checkHeader(await getJSON(`${root}${id}/manifest.json`), `${id}/manifest.json`);
      return new Game(`${root}${id}/`, man);
    } catch (e) {
      console.warn('skipping game', id, e);
      return null;
    }
  }));
  return games.filter(Boolean);
}

// mulberry32 — the seedable RNG pools use (?seed=N reproduces a roll).
export function mulberry32(seed) {
  let a = seed >>> 0;
  return () => {
    a |= 0; a = (a + 0x6D2B79F5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export const CATEGORY_LABELS = {
  level: 'Levels', object: 'Objects', music: 'Music',
  sfx: 'Sound effects', picture: 'Pictures', video: 'Videos',
};
