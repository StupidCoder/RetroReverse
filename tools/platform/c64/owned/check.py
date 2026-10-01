#!/usr/bin/env python3
"""Offline CPU acceptance; native and WASM execute exactly the same tests."""
import argparse,hashlib,json,os,subprocess,sys,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser();p.add_argument('--native-only',action='store_true');p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));p.add_argument('--node',default=os.environ.get('NODE','node'));p.add_argument('--out',type=Path);a=p.parse_args()
subprocess.run([sys.executable,str(HERE/'generate-opcodes.py'),'--check'],check=True)
manifest=json.loads((HERE/'tests/vectors.json').read_text());vectors=HERE/'tests/vectors.bin'
assert hashlib.sha256(vectors.read_bytes()).hexdigest()==manifest['sha256'],'Independent vector hash mismatch'
sources=[str(HERE/name) for name in ['cpu.cpp','opcodes.cpp','tests/cpu.cpp']]
flags=['-std=c++20','-O2','-Wall','-Wextra','-Werror']
def run(out):
 out.mkdir(parents=True,exist_ok=True)
 native=out/'cpu-native'
 subprocess.run([os.environ.get('CXX','clang++'),*flags,*sources,'-o',str(native)],check=True)
 expected=subprocess.check_output([str(native),str(vectors)],text=True);print(expected,end='')
 # Undefined behavior can invalidate determinism across native and WASM compilers.
 sanitized=out/'cpu-ubsan'
 subprocess.run([os.environ.get('CXX','clang++'),*flags,'-fsanitize=undefined','-fno-sanitize-recover=all',*sources,'-o',str(sanitized)],check=True)
 assert subprocess.check_output([str(sanitized),str(vectors)],text=True)==expected
 print('PASS undefined-behavior sanitizer')
 if not a.native_only:
  wasm=out/'cpu-wasm.mjs'
  subprocess.run([a.emcc,*flags,*sources,'-sENVIRONMENT=node','-sEXIT_RUNTIME=1','-sALLOW_MEMORY_GROWTH=1','-sSTACK_SIZE=1048576','--embed-file',str(vectors)+'@/vectors.bin','-o',str(wasm)],check=True)
  runner=out/'run-wasm.mjs'
  runner.write_text("import create from './cpu-wasm.mjs'; await create({arguments:['/vectors.bin']});\n")
  actual=subprocess.check_output([a.node,str(runner)],text=True)
  assert actual==expected,(actual,expected)
  print('PASS native/WASM conformance and trace digest equivalence')
if a.out:run(a.out.resolve())
else:
 with tempfile.TemporaryDirectory(prefix='rr-c64-owned-') as temp:run(Path(temp))
