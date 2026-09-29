#!/usr/bin/env python3
"""Build the Amiga hardware model and the pinned Musashi 68000 core."""
from pathlib import Path
import argparse
import json
import os
import subprocess

root = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--native-only', action='store_true')
parser.add_argument('--test', action='store_true')
parser.add_argument('--emcc', default=os.environ.get('EMXX', 'em++'))
args = parser.parse_args()
for name in ['work', 'web']:
    (root / name).mkdir(exist_ok=True)

flags = ['-O3', '-fwrapv', '-ffp-contract=off',
         '-DMUSASHI_CNF="' + str(root / 'core/cpu-config.h') + '"',
         '-I' + str(root / 'vendor/musashi')]
cpu_sources = [root / 'vendor/musashi' / name for name in
               ['m68kcpu.c', 'm68kops.c', 'softfloat/softfloat.c']]
cpu_sources.append(root / 'core/cpu-state.c')
sources = [str(root / 'core' / name) for name in
           ['machine.cpp', 'cia-disk.cpp', 'copper-blitter.cpp', 'video.cpp']]


def compile_cpu(compiler, prefix=''):
    objects = []
    for source in cpu_sources:
        output = root / 'work' / (prefix + source.stem + '.o')
        subprocess.run([compiler, *flags, '-Wno-shift-negative-value',
                        '-c', str(source), '-o', str(output)], check=True)
        objects.append(str(output))
    return objects


objects = compile_cpu(os.environ.get('CC', 'clang'))
cxx = os.environ.get('CXX', 'clang++')
subprocess.run([cxx, *flags, '-std=c++20', *sources,
                str(root / 'core/native.cpp'), *objects,
                '-o', str(root / 'work/amiga-native')], check=True)
if args.test:
    for test in ['pixel-amiga', 'raster-amiga', 'memory-amiga']:
        output = str(root / 'work' / test)
        subprocess.run([cxx, *flags, '-std=c++20', *sources,
                        str(root.parents[2] / ('browser/tests/' + test + '.cpp')),
                        *objects, '-o', output], check=True)
        subprocess.run([output], check=True)

if not args.native_only:
    objects = compile_cpu(str(Path(args.emcc).with_name('emcc')), 'wasm-')
    exports = ('inspect_regions inspect_data activity_begin activity_end activity_count activity_dropped activity_data input firmware init run pad mouse key error status proof profile '
               'frame state_input state_save state_data state_load capture_begin '
               'capture_end capture_info pixel source resource replay_begin '
               'replay_seek replay_info replay_frame replay_for_write raster_info raster_seek '
               'raster_frame raster_pixel raster_plane blit_seek blit_frame blit_pixel').split()
    subprocess.run([
        args.emcc, *flags, '-std=c++20', *sources, str(root / 'core/api.cpp'),
        *objects, '-sMODULARIZE=1', '-sEXPORT_ES6=1', '-sENVIRONMENT=web,worker',
        '-sALLOW_MEMORY_GROWTH=1', '-sINITIAL_MEMORY=67108864',
        '-sSTACK_SIZE=8388608', '-sDISABLE_EXCEPTION_CATCHING=0',
        '-sEXPORTED_FUNCTIONS=' + json.dumps(['_rr_' + name for name in exports]),
        '-sEXPORTED_RUNTIME_METHODS=["UTF8ToString","HEAPU8"]',
        '-o', str(root / 'web/core.js')], check=True)
