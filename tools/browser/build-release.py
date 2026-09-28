#!/usr/bin/env python3
"""Rebuild the four checked-in C++ cores with the pinned Emscripten toolchain."""
from pathlib import Path
import argparse,json,subprocess,os
root=Path(__file__).resolve().parents[2]
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
version=json.loads((root/'tools/browser/release.json').read_text())['emscripten']
actual=subprocess.check_output([a.emcc,'--version'],text=True)
if version not in actual.splitlines()[0]:raise SystemExit(f'Release builds require Emscripten {version}: {actual.splitlines()[0]}')
for platform in ['c64','psx','n64','threedo','nds','n3ds']:
 subprocess.run(['python3',str(root/f'tools/platform/{platform}/browser/build.py'),'--emcc',a.emcc],check=True,cwd=root)
subprocess.run(['python3',str(root/'tools/browser/package.py')],check=True,cwd=root)
subprocess.run(['python3',str(root/'tools/browser/check-release.py')],check=True,cwd=root)
