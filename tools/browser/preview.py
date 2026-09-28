#!/usr/bin/env python3
"""Local static preview, including the emulator area's Pages response headers."""
from http.server import ThreadingHTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from functools import partial
import argparse
p=argparse.ArgumentParser();p.add_argument('--port',type=int,default=8772);args=p.parse_args()
root=Path(__file__).resolve().parents[2]/'site'
class Handler(SimpleHTTPRequestHandler):
 def end_headers(self):
  if self.path.startswith('/emulators/'):
   for line in (root/'_headers').read_text().splitlines():
    if line.startswith('  ') and ':' in line:
     name,value=line.strip().split(':',1);self.send_header(name,value.strip())
  super().end_headers()
ThreadingHTTPServer(('127.0.0.1',args.port),partial(Handler,directory=root)).serve_forever()
