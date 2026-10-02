#!/usr/bin/env python3
"""Build the production C64 core from the owned implementation."""
import argparse,os,subprocess,sys
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--emcc',default=os.environ.get('EMXX','em++'));a=p.parse_args()
here=Path(__file__).resolve().parent
subprocess.run([sys.executable,str(here.parent/'owned/browser/build.py'),'--emcc',a.emcc,'--out',str(here/'web'),'--module-name','core.js'],check=True)
