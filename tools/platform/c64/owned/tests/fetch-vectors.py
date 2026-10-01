#!/usr/bin/env python3
"""Refresh the bounded, pinned independent NMOS instruction/bus regression corpus."""
import sys
import concurrent.futures,hashlib,json,struct,time,urllib.request
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
REV='2f6980a2d95757486c7bee24355c360e40e2a224'
BASE='https://raw.githubusercontent.com/SingleStepTests/65x02/'+REV+'/'
PER_OPCODE=64
sys.path.insert(0,str(HERE.parent))
from instruction_set import opcodes
ops={code:value for code,value in opcodes().items() if value[0]!='JAM'}
def fetch(code):
 for attempt in range(3):
  try:
   cases=[]
   with urllib.request.urlopen(BASE+f'6502/v1/{code:02x}.json',timeout=45) as response:
    for line in response:
     line=line.strip().rstrip(b',')
     if line.startswith(b'{'):cases.append(json.loads(line))
     if len(cases)==PER_OPCODE:break
   assert len(cases)==PER_OPCODE
   return code,cases
  except Exception:
   if attempt==2:raise
   time.sleep(1)
def regs(s):return struct.pack('<H5B',s['pc'],s['s'],s['a'],s['x'],s['y'],s['p'])
def memory(s):return struct.pack('<H',len(s['ram']))+b''.join(struct.pack('<HB',*p) for p in s['ram'])
with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool:
 cases=dict(pool.map(fetch,sorted(map(int,ops))))
out=bytearray(struct.pack('<II',0x31565043,len(cases)*PER_OPCODE))
for code,tests in sorted(cases.items()):
 for i,t in enumerate(tests):
  out+=struct.pack('<BH',code,i)+regs(t['initial'])+regs(t['final'])+memory(t['initial'])+memory(t['final'])
  out+=bytes([len(t['cycles'])])+b''.join(struct.pack('<HBB',a,v,access=='write') for a,v,access in t['cycles'])
license=urllib.request.urlopen(BASE+'LICENSE',timeout=45).read()
(HERE/'SingleStepTests-LICENSE').write_bytes(license)
(HERE/'vectors.bin').write_bytes(out)
(HERE/'vectors.json').write_text(json.dumps({'source':'https://github.com/SingleStepTests/65x02','revision':REV,'path':'6502/v1','selection':'first 64 cases of each of 237 documented/stable undocumented opcodes; JAM and seven unstable encodings excluded; not exhaustive','cases':len(cases)*PER_OPCODE,'sha256':hashlib.sha256(out).hexdigest()},indent=2)+'\n')
print('Pinned',len(cases)*PER_OPCODE,'cases,',len(out),'bytes',flush=True)
