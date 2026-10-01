#!/usr/bin/env python3
"""Offline native/WASM CPU and board acceptance; optional local firmware boot."""
import argparse,hashlib,json,os,subprocess,sys,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser();p.add_argument('--native-only',action='store_true');p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));p.add_argument('--node',default=os.environ.get('NODE','node'));p.add_argument('--out',type=Path);p.add_argument('--firmware-dir',type=Path);p.add_argument('--functional-rom',type=Path);p.add_argument('--fort-tap',type=Path);p.add_argument('--fort-fixtures',type=Path);a=p.parse_args()
subprocess.run([sys.executable,str(HERE/'generate-opcodes.py'),'--check'],check=True)
manifest=json.loads((HERE/'tests/vectors.json').read_text());vectors=HERE/'tests/vectors.bin'
assert hashlib.sha256(vectors.read_bytes()).hexdigest()==manifest['sha256'],'Independent vector hash mismatch'
flags=['-std=c++20','-O2','-Wall','-Wextra','-Werror']
firmware=[]
if a.firmware_dir:
 for name,size in [('basic',8192),('kernal',8192),('chargen',4096)]:
  path=(a.firmware_dir/(name+'.rom')).resolve();data=path.read_bytes();assert len(data)==size
  firmware.append(path);print(name+' SHA256 '+hashlib.sha256(data).hexdigest(),flush=True)
if a.functional_rom:
 a.functional_rom=a.functional_rom.resolve()
 assert hashlib.sha256(a.functional_rom.read_bytes()).hexdigest()=='fa12bfc761e6f9057e4cc01a665a7b800ff01ae91f598af1e39a1201d01953fd','Unexpected functional test image'
fort=[]
if a.fort_tap or a.fort_fixtures:
 if not (a.fort_tap and a.fort_fixtures and firmware):p.error('Fort acceptance requires --fort-tap, --fort-fixtures and --firmware-dir')
 fort=[a.fort_tap.resolve(),*[(a.fort_fixtures/name).resolve() for name in ['expected.bin','pages.bin','graphics.bin','graphics-mask.bin']]]
 for path in fort:print(path.name+' SHA256 '+hashlib.sha256(path.read_bytes()).hexdigest(),flush=True)
def target(out,name,sources,files,images=False):
 sources=[str(HERE/source) for source in sources]
 native=out/(name+'-native')
 subprocess.run([os.environ.get('CXX','clang++'),*flags,*sources,'-o',str(native)],check=True)
 arguments=[*map(str,files)]+([str(out/name)] if images else [])
 expected=subprocess.check_output([str(native),*arguments],text=True);print(expected,end='',flush=True)
 sanitized=out/(name+'-ubsan')
 subprocess.run([os.environ.get('CXX','clang++'),*flags,'-fsanitize=undefined','-fno-sanitize-recover=all',*sources,'-o',str(sanitized)],check=True)
 assert subprocess.check_output([str(sanitized),*arguments],text=True)==expected
 print('PASS '+name+' undefined-behavior sanitizer',flush=True)
 if not a.native_only:
  wasm=out/(name+'-wasm.mjs');embedded=[];arguments=[]
  for f in files:embedded+=['--embed-file',str(f)+'@/'+f.name];arguments.append('/'+f.name)
  subprocess.run([a.emcc,*flags,*sources,'-sENVIRONMENT=node','-sEXIT_RUNTIME=1','-sALLOW_MEMORY_GROWTH=1','-sSTACK_SIZE=8388608',*embedded,'-o',str(wasm)],check=True)
  runner=out/('run-'+name+'-wasm.mjs')
  runner.write_text('import create from '+json.dumps('./'+wasm.name)+'; await create({arguments:'+json.dumps(arguments)+'});\n')
  actual=subprocess.check_output([a.node,str(runner)],text=True)
  assert actual==expected,(actual,expected)
  print('PASS '+name+' native/WASM output and trace equivalence',flush=True)
def run(out):
 out.mkdir(parents=True,exist_ok=True)
 target(out,'cpu',['cpu.cpp','opcodes.cpp','tests/cpu.cpp'],[vectors])
 if a.functional_rom:target(out,'functional',['cpu.cpp','opcodes.cpp','tests/functional.cpp'],[a.functional_rom])
 core=['cpu.cpp','opcodes.cpp','cia.cpp','vic.cpp','sid.cpp','board.cpp']
 target(out,'board',core+['tests/board.cpp'],firmware)
 target(out,'video',core+['tests/video.cpp'],[])
 if fort:target(out,'fort',core+['tests/fort.cpp'],firmware+fort,images=True)
if a.out:run(a.out.resolve())
else:
 with tempfile.TemporaryDirectory(prefix='rr-c64-owned-') as temp:run(Path(temp))
