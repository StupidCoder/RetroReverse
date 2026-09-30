#!/usr/bin/env python3
"""Public media-free x86/NV2A tests in actual WASM, including explicit SIMD."""
from pathlib import Path
import argparse,os,subprocess,tempfile,json
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
with tempfile.TemporaryDirectory(prefix='rr-x86-wasm-') as tmp:
 for name in ['debug-dos','fast-x86','fast-xbox','pixel-dos','render-dos','pixel-xbox']:
  out=Path(tmp)/(name+'.mjs')
  subprocess.run([a.emcc,'-O3','-std=c++20','-fwrapv','-ffp-contract=off','-msimd128','-fwasm-exceptions','-Wno-parentheses-equality','-Wno-trigraphs','-sALLOW_MEMORY_GROWTH=1','-sSTACK_SIZE=2097152','-sENVIRONMENT=node','-sEXIT_RUNTIME=1','tools/browser/tests/'+name+'.cpp','-o',str(out)],check=True)
  subprocess.run([os.environ.get('NODE','node'),'--input-type=module','-e','await (await import('+json.dumps(out.as_uri())+')).default()'],check=True)
