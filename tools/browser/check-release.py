#!/usr/bin/env python3
"""Offline Pages release integrity/network-contract checks; needs no game images."""
from pathlib import Path
import hashlib,json,re
root=Path(__file__).resolve().parents[2];site=root/'site';out=site/'emulators'
release=json.loads((out/'release.json').read_text());base=out/'releases'/release['id']
if release.get('previous'):assert (out/'releases'/release['previous']/'worker.js').is_file(),'Previous published bundle missing'
for name,wanted in release['assets'].items():
 p=base/name;assert p.is_file(),f'Missing {p}'
 assert hashlib.sha256(p.read_bytes()).hexdigest()==wanted,f'Hash mismatch: {p}'
 assert hashlib.sha256((out/name).read_bytes()).hexdigest()==wanted,f'Run package.py after modifying {name}'
 assert p.stat().st_size<=25*1024*1024,f'Asset too large: {p}'
for slug in ['c64','ps1','n64','3do','ds','3ds','psp','gb','gg','amiga','ps2','gc','gba','dc','dos','xbox']:
 html=(out/slug/'index.html').read_text();assert f"../releases/{release['id']}/app.js" in html
 assert f"../releases/{release['id']}/style.css" in html
 assert (base/f'cores/{slug}/core.wasm').read_bytes()[:4]==b'\0asm'
for platform in ['c64','amiga']:
 for e in json.loads((base/f'firmware/{platform}/manifest.json').read_text()):
  b=(base/f'firmware/{platform}'/e['file']).read_bytes();assert len(b)==e['bytes'] and hashlib.sha256(b).hexdigest()==e['sha256']
for name in ['prepared-start.js','prepared-panel.js','checkpoint-cache.js','panel-dom.js','viewport-model.js','viewport-workspace.js','viewport-panels.js','memory-panel.js','inspection-feed.js','emulator-session.js','experiment-worker.js','experiment-panel.js','threedo-inspector.js','dos-knowledge.js','dos-inspector.js','tour-worker.js','tour-panel.js','structured-state.js','state-panel.js','panel-layout.js','debug-worker.js','execution-gate.js','code-workspace.js','disassembly6502.js','knowledge-model.js','memory-model.js','memory-workspace.js','memory-worker.js','app.js','worker.js','media.js','dc-media.js','dos-media.js','state.js','inspector.js','replay.js','pacing.js','raster.js','amiga-raster.js','workspaces.js','ui-shell.js','ui-platforms.js','render-workspace.js','buffer-view.js','render-timeline.js','tileset.js','tileset-decode.js','ps1-vram.js','ps1-vram-decode.js','dc-texture.js','dc-texture-decode.js']:
 s=(base/name).read_text();assert not re.search(r'https?://|localhost|127\.0\.0\.1|sendBeacon|XMLHttpRequest|WebSocket',s),f'Unexpected network source: {name}'
 assert not re.search(r"method\s*:\s*['\"](?:POST|PUT|PATCH)",s),f'Upload path in {name}'
assert (site/'index.html').is_file();assert 'emulators/' in (site/'src/app.js').read_text()
assert 'Cache-Control: no-cache' in (site/'_headers').read_text()
print(f"PASS: release {release['id']}, {len(release['assets'])} pinned assets, firmware hashes, sixteen direct routes and static network contract")
