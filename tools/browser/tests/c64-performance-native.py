#!/usr/bin/env python3
"""Private Fort native timing. Optional legacy source root from commit d66634b1."""
from pathlib import Path
import argparse,json,subprocess,tempfile
p=argparse.ArgumentParser();p.add_argument('--legacy-root',type=Path);a=p.parse_args()
root=Path(__file__).resolve().parents[3];owned=root/'tools/platform/c64/owned'
pkg=json.loads((root/'games/fort-apocalypse-c64/knowledge.json').read_text())
for label in ['owned']+(['legacy'] if a.legacy_root else []):
 base=owned if label=='owned' else a.legacy_root/'tools/platform/c64/browser'
 header=base/('browser/core.h' if label=='owned' else 'core/core.h')
 actions=[]
 for step in pkg['preparedStarts']['terrain-owned' if label=='owned' else 'terrain']['actions']:
  if 'run' in step:actions.append(f'for(int left={step["run"]};left;){{int n=std::min(left,100000);assert(rr_run(n,0,0)==n);left-=n;}}')
  elif 'key' in step:actions.append('rr_key(%d,%d);'%tuple(step['key']))
  elif 'joystick' in step:actions.append('rr_joystick(%d,%d);'%tuple(step['joystick']))
  elif 'play' in step:actions.append(f'rr_play({step["play"]});')
  elif 'until' in step:actions.append(f'assert(rr_debug_begin(2,{step["until"]},1));int r=0;for(int left={step["budget"]};left>0&&!r;left-=10000)r=rr_debug_run(std::min(left,10000));assert(r==4);')
  else:raise ValueError(step)
 with tempfile.TemporaryDirectory(prefix='rr-native-perf-') as temp:
  temp=Path(temp);src=temp/'bench.cpp';exe=temp/'bench'
  src.write_text('#include '+json.dumps(str(header))+'''\n#include <cassert>
#include <fstream>
#include <iterator>
#include <vector>
#include <algorithm>
#include <chrono>
#include <cstdio>
std::vector<unsigned char> read(const char* p){std::ifstream f(p,std::ios::binary);assert(f);return {std::istreambuf_iterator<char>(f),{}};}
int main(int,char** v){size_t offset=0;for(int i=1;i<4;i++){auto b=read(v[i]);std::copy(b.begin(),b.end(),rr_input()+offset);offset+=b.size();}assert(rr_init(8192,8192,4096));auto tape=read(v[4]);std::copy(tape.begin(),tape.end(),rr_input());assert(rr_tape(tape.size()));
'''+''.join('{'+action+'}' for action in actions)+'''
rr_trace(0,0,0);rr_run(40000,7,0);assert(rr_checkpoint(0));for(int mode=0;mode<2;mode++)for(int i=0;i<6;i++){assert(rr_restore(0));auto start=std::chrono::steady_clock::now();double cycle=rr_cycle();if(mode){rr_capture_begin();rr_run(40000,7,0);rr_capture_end();}else assert(rr_run(985248,0,0)==985248);double ms=std::chrono::duration<double,std::milli>(std::chrono::steady_clock::now()-start).count();if(i)printf("%s %s %.6f ms %.0f cycles %.3fx\\n",'''+json.dumps(label)+''',mode?"capture":"gameplay",ms,rr_cycle()-cycle,(rr_cycle()-cycle)/985248/(ms/1000));}}
''')
  sources=[owned/(n+'.cpp') for n in ['cpu','opcodes','cia','vic','sid','board','via','disk','drive','system','state']]+[owned/'browser/core.cpp'] if label=='owned' else [base/'core/core.cpp']
  subprocess.run(['clang++','-O2','-std=c++20','-Wno-address-of-temporary',str(src),*map(str,sources),'-o',str(exe)],check=True)
  subprocess.run([str(exe),*[str(root/f'site/emulators/firmware/c64/{n}.rom') for n in ['basic','kernal','chargen']],str(root/'games/fort-apocalypse-c64/Fort_Apocalypse.tap')],check=True)
