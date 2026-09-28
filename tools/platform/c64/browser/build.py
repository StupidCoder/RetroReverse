#!/usr/bin/env python3
"""Build the same C++ core natively and with an explicitly selected Emscripten."""
import argparse, os, pathlib, subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
here=pathlib.Path(__file__).resolve().parent
(here/'work').mkdir(exist_ok=True);(here/'web').mkdir(exist_ok=True)
flags=['-O2','-std=c++20','-Wno-c99-designator','-Wno-address-of-temporary']
subprocess.run(['clang++',*flags,str(here/'core/core.cpp'),str(here/'core/native.cpp'),'-o',str(here/'work/c64-native')],check=True)
subprocess.run(['clang++',*flags,str(here/'core/core.cpp'),str(here/'core/lesson_native.cpp'),'-o',str(here/'work/lesson-native')],check=True)
exports=['capture_info','state_input','state_save','state_data','state_load','profile','input','init','tape','prepare','run','play','key','joystick','checkpoint','restore','trace','status','events','error','ram','bus','frame','width','height','audio','audio_count','corrupt','trace_mask','bus_flags','stop_reason','trace_reads','cycle','previous','memory','capture_begin','capture_end','pixel']
subprocess.run([a.emcc,*flags,str(here/'core/core.cpp'),'-sDISABLE_EXCEPTION_CATCHING=0','-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=134217728','-sSTACK_SIZE=4194304','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(here/'web/core.js')],check=True)
