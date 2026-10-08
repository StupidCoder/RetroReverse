#!/usr/bin/env python3
"""Build the checked-in 3DS C++ translation and its browser API."""
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
r=Path(__file__).resolve().parent
for d in ['work','web']:(r/d).mkdir(exist_ok=True)
flags=['-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality']
subprocess.run(['clang++',*flags,str(r/'core/native.cpp'),'-o',str(r/'work/3ds-native')],check=True)
exports='input init run pad touch error status proof profile frame state_input state_save state_data state_load capture_begin capture_end capture_info pixel source resource replay_begin replay_seek replay_info replay_frame replay_for_write'.split()
exports += 'graphics_scanout graphics_enable graphics_stats graphics_begin graphics_end graphics_data graphics_info'.split()
exports += 'perf_enable perf_stats'.split()
exports += 'inspect_regions inspect_data activity_begin activity_end activity_count activity_dropped activity_data'.split()
subprocess.run([a.emcc,*flags,str(r/'core/api.cpp'),'-sASYNCIFY=1','-sASYNCIFY_STACK_SIZE=1048576','-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=134217728','-sSTACK_SIZE=4194304','-sDISABLE_EXCEPTION_CATCHING=0','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8","ccall"]','-o',str(r/'web/core.js')],check=True)
