#!/usr/bin/env python3
"""Public, media-free CI checks for the unified browser application."""
from pathlib import Path
import os,subprocess,tempfile
root=Path(__file__).resolve().parents[2];node=os.environ.get('NODE','node');clang=os.environ.get('CXX','clang++')
subprocess.run(['go','test','./tools/knowledge','./tools/cmd/knowledgecheck','./tools/cmd/knowledgeexport','./tools/cpu/mos6502','./tools/cmd/dis6502export'],cwd=root,check=True)
subprocess.run(['go','run','./tools/cmd/knowledgeexport','-check','-out','site/emulators/knowledge-data.js',*map(str,sorted((root/'games').glob('*/knowledge.json')))],cwd=root,check=True)
subprocess.run(['go','run','./tools/cmd/dis6502export','-check'],cwd=root,check=True)
for name in ['debug','knowledge','memory','media','dc-media','dos-media','input','state-container','inspector','pacing','ui-shell','tileset','ps1-vram','dc-texture']:
 subprocess.run([node,f'tools/browser/tests/{name}.mjs'],cwd=root,check=True)
for p in (root/'site/emulators').glob('*.js'):
 subprocess.run([node,'--check',str(p)],check=True)
with tempfile.TemporaryDirectory(prefix='rr-check-') as d:
 for name in ['debug-c64','memory-gg','memory-c64','runtime','capture-buffer','replay','pixel-c64','raster-c64','pixel-ps1','vram-ps1','pixel-n64','pixel-3do','fast-3do','pixel-ds','fast-ds','pixel-3ds','pixel-psp','fast-psp','pixel-gb','raster-gb','pixel-gg','timing-gg','raster-gg','pixel-gc','fast-gc','pixel-ps2','fast-ps2','pixel-gba','fast-gba','pixel-dc','texture-dc','fast-dc','pixel-dos','render-dos','fast-x86','pixel-xbox','fast-xbox']:
  out=str(Path(d)/name)
  subprocess.run([clang,'-std=c++20','-O1','-fwrapv','-ffp-contract=off','-Wno-address-of-temporary','-Wno-parentheses-equality','-Wno-trigraphs',f'tools/browser/tests/{name}.cpp','-o',out]+(['-lz'] if name in ['pixel-psp','fast-psp','pixel-gb','raster-gb','pixel-gg'] else []),cwd=root,check=True)
  subprocess.run([out]+([str(Path(d)/'dc-samples.jsonl')] if name=='texture-dc' else []),cwd=root,check=True)
  if name=='texture-dc':subprocess.run([node,'tools/browser/tests/dc-texture.mjs',str(Path(d)/'dc-samples.jsonl')],cwd=root,check=True)
subprocess.run(['python3','tools/platform/amiga/browser/build.py','--native-only','--test'],cwd=root,check=True)
subprocess.run(['python3','tools/browser/check-release.py'],cwd=root,check=True)
