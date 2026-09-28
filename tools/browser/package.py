#!/usr/bin/env python3
"""Package existing WASM builds into the site's static emulator subtree."""
from pathlib import Path
import hashlib,json,shutil
repo=Path(__file__).resolve().parents[2]
out=repo/'site/emulators'
manifest={}
for slug,platform in [('c64','c64'),('ps1','psx'),('n64','n64'),('3do','threedo')]:
 dest=out/'cores'/slug;dest.mkdir(parents=True,exist_ok=True)
 for name in ['core.js','core.wasm']:
  src=repo/f'tools/platform/{platform}/browser/web'/name
  if not src.is_file():raise SystemExit(f'Build {platform} first: missing {src}')
  shutil.copyfile(src,dest/name)
  manifest[f'{slug}/{name}']=hashlib.sha256(src.read_bytes()).hexdigest()
for p in out.rglob('*'):
 if p.is_file() and p.stat().st_size>25*1024*1024:raise SystemExit(f'Pages asset too large: {p}')
(out/'build-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('Packaged four cores; all emulator assets below Pages file limit.')
