#!/usr/bin/env python3
"""Local browser acceptance server. Never part of the published application."""
import argparse,json,re
from pathlib import Path
from http.server import SimpleHTTPRequestHandler,ThreadingHTTPServer
from urllib.parse import urlparse,parse_qs
parser=argparse.ArgumentParser();parser.add_argument('--port',type=int,default=8767);parser.add_argument('--results',type=Path,default=Path('/private/tmp/rr-viewport-results'));args=parser.parse_args()
root=Path(__file__).resolve().parents[3];args.results.mkdir(parents=True,exist_ok=True)
class Handler(SimpleHTTPRequestHandler):
 def __init__(self,*a,**kw):super().__init__(*a,directory=str(root),**kw)
 def do_POST(self):
  url=urlparse(self.path);name=parse_qs(url.query).get('id',[''])[0]
  if url.path!='/__viewport_result__' or not re.fullmatch('[a-z-]{1,32}',name):self.send_error(404);return
  length=int(self.headers.get('Content-Length','0'))
  if not 0<length<=1048576:self.send_error(413);return
  try:data=json.loads(self.rfile.read(length));assert data['result'] in ['PASS','FAIL']
  except (ValueError,KeyError,AssertionError):self.send_error(400);return
  (args.results/(name+'.json')).write_text(json.dumps(data,indent=2)+'\n')
  self.send_response(200);self.end_headers();self.wfile.write(b'ok')
 def log_message(self,*args):pass
ThreadingHTTPServer(('127.0.0.1',args.port),Handler).serve_forever()
