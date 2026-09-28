#!/usr/bin/env python3
"""Read-only verification of the deployed static emulator bundle."""
import argparse,hashlib,json,subprocess,tempfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('base');a=p.parse_args();base=a.base.rstrip('/')+'/'
def get(path):
 with tempfile.TemporaryDirectory(prefix='rr-http-') as tmp:
  headers=Path(tmp)/'headers'
  body=subprocess.check_output(['curl','--fail','--silent','--show-error','--max-time','30','--dump-header',str(headers),base+path])
  h={}
  for line in headers.read_text().splitlines():
   if ':' in line:
    key,value=line.split(':',1);h[key.lower()]=value.strip()
  return body,h
b,h=get('release.json');release=json.loads(b);rows=[]
for slug in ['c64','ps1','n64','3do']:
 body,headers=get(slug+'/');assert f"releases/{release['id']}/app.js".encode() in body
 assert headers.get('Cache-Control',headers.get('cache-control'))=='no-cache'
 rows.append({'route':slug+'/','contentType':headers.get('Content-Type',headers.get('content-type'))})
for name,wanted in release['assets'].items():
 body,headers=get('releases/'+release['id']+'/'+name);assert hashlib.sha256(body).hexdigest()==wanted,name
 content=headers.get('Content-Type',headers.get('content-type',''))
 if name.endswith('.wasm'):assert content.startswith('application/wasm'),(name,content)
 if name.endswith('.js'):assert 'javascript' in content,(name,content)
assert h.get('Cross-Origin-Opener-Policy',h.get('cross-origin-opener-policy'))=='same-origin'
assert h.get('Cross-Origin-Embedder-Policy',h.get('cross-origin-embedder-policy'))=='require-corp'
print(json.dumps({'base':base,'release':release['id'],'assetsVerified':len(release['assets']),'routes':rows,'isolationHeaders':True}))
