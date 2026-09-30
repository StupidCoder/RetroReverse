#!/usr/bin/env python3
"""Compile the checked-in C++ core; Go is needed only to regenerate its translation."""
from pathlib import Path
import argparse, os, subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
r=Path(__file__).resolve().parent;(r/'work').mkdir(exist_ok=True);(r/'web').mkdir(exist_ok=True)
flags=['-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality']
subprocess.run(['clang++',*flags,str(r/'core/native.cpp'),'-o',str(r/'work/threedo-native')],check=True)
exports=['replay_begin','replay_seek','replay_info','replay_frame','replay_for_write','source','resource','capture_info','capture_begin','capture_end','pixel','state_input','state_save','state_data','state_load','run_slice','profile','init_config','init','run','pad','error','status','proof','frame']
exports += 'cycle ram ram_size debug_snapshot debug_begin debug_bind debug_run activity_select activity_clear_selection inspect_regions inspect_data activity_begin activity_end activity_count activity_dropped activity_data'.split()
subprocess.run([a.emcc,*flags,str(r/'core/api.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=67108864','-sSTACK_SIZE=2097152','-sDISABLE_EXCEPTION_CATCHING=0','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(r/'web/core.js')],check=True)
