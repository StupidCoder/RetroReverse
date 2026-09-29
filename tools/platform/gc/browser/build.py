#!/usr/bin/env python3
"""Build the checked-in GC C++ translation and its browser API."""
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
r=Path(__file__).resolve().parent
for d in ['work','web']:(r/d).mkdir(exist_ok=True)
flags=['-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality']
subprocess.run(['clang++',*flags,str(r/'core/native.cpp'),'-o',str(r/'work/gc-native')],check=True)
exports='init_file run pad error status proof profile frame state_input state_save state_data state_load capture_begin capture_end capture_info pixel source resource replay_begin replay_seek replay_info replay_frame replay_for_write'.split()
subprocess.run([a.emcc,*flags,'-msimd128','--profiling-funcs',str(r/'core/api.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=134217728','-sSTACK_SIZE=2097152','-fwasm-exceptions','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(r/'web/core.js')],check=True)
