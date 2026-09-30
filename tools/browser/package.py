#!/usr/bin/env python3
"""Package existing WASM builds into the site's static emulator subtree."""
from pathlib import Path
import argparse,hashlib,json,shutil,subprocess
repo=Path(__file__).resolve().parents[2]
out=repo/'site/emulators'
parser=argparse.ArgumentParser()
parser.add_argument('--site-only',action='store_true',help='Reuse verified packaged cores; refresh only site code/data and release')
parser.add_argument('--core',action='append',default=[],help='Refresh only these core slugs; retain and verify other packaged cores')
args=parser.parse_args()
if args.site_only and args.core:parser.error('--site-only and --core are mutually exclusive')
(repo/'site/knowledge').mkdir(exist_ok=True)
packages=sorted((repo/'games').glob('*/knowledge.json'))
subprocess.run(['go','run','./tools/cmd/knowledgeexport','-out',str(out/'knowledge-data.js'),'-index',str(repo/'site/knowledge/index.html'),*map(str,packages)],cwd=repo,check=True)
existing=json.loads((out/'build-manifest.json').read_text()) if args.site_only or args.core else {}
subprocess.run(['go','run','./tools/cmd/dis6502export','-out',str(out/'opcodes6502.js')],cwd=repo,check=True)
manifest={}
for slug,platform in [('c64','c64'),('ps1','psx'),('n64','n64'),('3do','threedo'),('ds','nds'),('3ds','n3ds'),('psp','psp'),('gb','gameboy'),('gg','gamegear'),('amiga','amiga'),('ps2','ps2'),('gc','gc'),('gba','gba'),('dc','dc'),('dos','dos'),('xbox','xbox')]:
 reuse=args.site_only or bool(args.core and slug not in args.core)
 dest=out/'cores'/slug;dest.mkdir(parents=True,exist_ok=True)
 for name in ['core.js','core.wasm']:
  src=(dest/name) if reuse else repo/f'tools/platform/{platform}/browser/web'/name
  if not src.is_file():raise SystemExit(f'Build {platform} first: missing {src}')
  if reuse:
   if hashlib.sha256(src.read_bytes()).hexdigest()!=existing.get(f'{slug}/{name}'):raise SystemExit(f'Existing core hash mismatch: {src}')
  else:shutil.copyfile(src,dest/name)
  manifest[f'{slug}/{name}']=hashlib.sha256(src.read_bytes()).hexdigest()
for p in out.rglob('*'):
 if p.is_file() and p.stat().st_size>25*1024*1024:raise SystemExit(f'Pages asset too large: {p}')
(out/'build-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('Packaged sixteen cores; all emulator assets below Pages file limit.')

# Content-address the complete executable bundle. A page opened before a deploy
# continues to use matching worker/JS/WASM/firmware rather than mixing versions.
assets={}
for pattern in ['*.js','style.css','build-manifest.json','cores/**/*','firmware/**/*']:
 for p in out.glob(pattern):
  if p.is_file():assets[p.relative_to(out).as_posix()]=hashlib.sha256(p.read_bytes()).hexdigest()
release_id=hashlib.sha256(json.dumps(assets,sort_keys=True,separators=(',',':')).encode()).hexdigest()[:20]
release=out/'releases'/release_id
old=json.loads((out/'release.json').read_text()) if (out/'release.json').is_file() else {}
try:
 committed=json.loads(subprocess.check_output(['git','show','HEAD:site/emulators/release.json'],cwd=repo,stderr=subprocess.DEVNULL))
except (subprocess.CalledProcessError,json.JSONDecodeError):committed=old
previous=committed.get('previous') if committed.get('id')==release_id else committed.get('id')
for name in assets:
 dest=release/name;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(out/name,dest)
(out/'release.json').write_text(json.dumps({'schema':1,'id':release_id,'previous':previous,'assets':assets},indent=2)+'\n')
import re
for slug in ['c64','ps1','n64','3do','ds','3ds','psp','gb','gg','amiga','ps2','gc','gba','dc','dos','xbox']:
 p=out/slug/'index.html';html=p.read_text()
 html=re.sub(r'(?<=src=")[^" ]*/app\.js',f'../releases/{release_id}/app.js',html)
 html=re.sub(r'(?<=href=")[^" ]*/style\.css',f'../releases/{release_id}/style.css',html)
 p.write_text(html)
# Keep the prior executable bundle for already-open pages. Existing game assets
# and explanation URLs outside this managed directory are never changed.
for p in (out/'releases').iterdir():
 if p.is_dir() and p.name not in {release_id,previous}:shutil.rmtree(p)
print(f'Content-addressed release {release_id}')
