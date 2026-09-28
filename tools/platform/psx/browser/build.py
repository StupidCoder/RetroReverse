#!/usr/bin/env python3
"""Same original C++ PS1 implementation, native and Emscripten browser targets."""
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
r=Path(__file__).resolve().parent;(r/'work').mkdir(exist_ok=True);(r/'web').mkdir(exist_ok=True)
flags=['-O2','-std=c++20']
subprocess.run(['clang++',*flags,str(r/'core/machine.cpp'),str(r/'core/native.cpp'),'-o',str(r/'work/psx-native')],check=True)
exports=['source','resource','capture_info','state_input','state_save','state_data','state_load','profile','init_file','input','init','run','pad','error','status','proof','frame','ram','save','restore','capture_begin','capture_end','pixel','diagnostics']
subprocess.run([a.emcc,*flags,str(r/'core/machine.cpp'),str(r/'core/api.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=67108864','-sSTACK_SIZE=1048576','-sDISABLE_EXCEPTION_CATCHING=0','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(r/'web/core.js')],check=True)
