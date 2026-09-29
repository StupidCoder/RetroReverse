#!/usr/bin/env python3
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args();r=Path(__file__).resolve().parent
for d in ['work','web']:(r/d).mkdir(exist_ok=True)
flags=['-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality']
subprocess.run(['clang++',*flags,str(r/'core/native.cpp'),'-o',str(r/'work/handheld-native')],check=True)
exports='inspect_regions inspect_data activity_begin activity_end activity_count activity_dropped activity_data input init run pad error status proof profile frame state_input state_save state_data state_load capture_begin capture_end capture_info pixel source resource replay_begin replay_seek replay_info replay_frame replay_for_write raster_info raster_seek raster_frame raster_pixel tileset_size tileset_data'.split()
subprocess.run([a.emcc,*flags,str(r/'core/api.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=33554432','-sSTACK_SIZE=1048576','-sDISABLE_EXCEPTION_CATCHING=0','-sEXPORTED_FUNCTIONS='+str(['_rr_'+x for x in exports]).replace("'",'"'),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(r/'web/core.js')],check=True)
