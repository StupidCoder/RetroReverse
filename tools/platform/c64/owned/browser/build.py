#!/usr/bin/env python3
"""Build the development adapter into an explicit scratch directory; never ship it."""
import argparse,json,os,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser()
p.add_argument('--emcc',default=os.environ.get('EMXX','em++'))
p.add_argument('--out',type=Path,required=True)
a=p.parse_args();a.out.mkdir(parents=True,exist_ok=True)
sources=[HERE.parent/(n+'.cpp') for n in ['cpu','opcodes','cia','vic','sid','board','via','disk','drive','system','state']]+[HERE/'core.cpp']
exports='capabilities input init tape drive_rom disk drive_status prepare run stop_reason play key joystick trace trace_mask trace_reads status events error profile ram bus bus_flags frame width height audio audio_count cycle previous checkpoint restore corrupt state_bind state_input state_save state_data state_load debug_begin debug_run debug_watch debug_event debug_edit debug_snapshot inspect_regions inspect_data activity_begin activity_end activity_count activity_dropped activity_data'.split()
subprocess.run([a.emcc,'-std=c++20','-O2','-Wall','-Wextra','-Werror','-fexceptions',*map(str,sources),
 '-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=web,worker,node','-sALLOW_MEMORY_GROWTH=1','-sINITIAL_MEMORY=33554432','-sSTACK_SIZE=2097152',
 '-sEXPORTED_FUNCTIONS='+json.dumps(['_rr_'+n for n in exports]),'-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]','-o',str(a.out/'core.mjs')],check=True)
