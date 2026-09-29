#!/usr/bin/env python3
"""Build the checked-in DOS C++ translation and its browser API."""
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));p.add_argument('--reference',action='store_true',help='Build scalar byte-access reference into work/reference');a=p.parse_args()
r=Path(__file__).resolve().parent
for d in ['work','web']:(r/d).mkdir(exist_ok=True)
dest=r/'work/reference' if a.reference else r/'web';dest.mkdir(parents=True,exist_ok=True)
flags=['-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality','-Wno-trigraphs']
if a.reference:flags+=['-DRR_X86_REFERENCE','-DRR_GPU_REFERENCE']
subprocess.run(['clang++',*flags,str(r/'core/native.cpp'),'-o',str(r/'work'/('dos-native-reference' if a.reference else 'dos-native'))],check=True)
exports='input file init run key mouse pad error status proof profile frame state_input state_save state_data state_load capture_begin capture_end capture_info pixel source resource replay_begin replay_seek replay_info replay_frame replay_for_write replay_view'.split()
subprocess.run([a.emcc,*flags,'-msimd128','--profiling-funcs',str(r/'core/api.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=134217728','-sSTACK_SIZE=2097152','-fwasm-exceptions','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(dest/'core.js')],check=True)
