#!/usr/bin/env python3
"""Media-free differential tests against the browser core's WASM interpreter."""
from pathlib import Path
import argparse,os,subprocess
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
root=Path(__file__).resolve().parents[3];output=root/'tools/platform/n3ds/browser/work/shader-oracle.mjs';output.parent.mkdir(exist_ok=True)
subprocess.run([a.emcc,'-O3','-std=c++20','-fwrapv','-ffp-contract=off','-Wno-parentheses-equality',str(root/'tools/browser/tests/shader-3ds.cpp'),'-sMODULARIZE=1','-sEXPORT_ES6=1','-sENVIRONMENT=node,web','-sALLOW_MEMORY_GROWTH=1','-sDISABLE_EXCEPTION_CATCHING=0','-sEXPORTED_FUNCTIONS=["_shader_memory","_shader_reference"]','-sEXPORTED_RUNTIME_METHODS=["HEAPU8"]','-o',str(output)],cwd=root,check=True)
subprocess.run([os.environ.get('NODE','node'),'tools/browser/tests/shader-3ds.mjs',str(output)],cwd=root,check=True)
