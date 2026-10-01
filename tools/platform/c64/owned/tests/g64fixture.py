#!/usr/bin/env python3
"""Independent offline GCR oracle; print identities, never inject guest bytes."""
import hashlib,json,struct,sys
from pathlib import Path
image=Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(image).hexdigest()=='5ce29ce04786eca6518fb08dfe659abb3eee079b4135a3f7606f9d17a501ec77'
assert image[:9]==b'GCR-1541\0'
codes=[10,11,18,19,14,15,22,23,9,25,26,27,13,29,30,21]
sectors={}
for track in range(1,36):
 offset=struct.unpack_from('<I',image,12+8*(track-1))[0]
 length=struct.unpack_from('<H',image,offset)[0]
 bits=''.join(f'{v:08b}' for v in image[offset+2:offset+2+length]);bits+=bits[:3000]
 def decode(at,count):
  return bytes(codes.index(int(bits[at+i:at+i+5],2))*16+codes.index(int(bits[at+i+5:at+i+10],2)) for i in range(0,count*10,10))
 pos=0;sector=None
 while True:
  pos=bits.find('11111111110',pos)
  if pos<0:break
  pos+=10
  try:
   kind=decode(pos,1)[0]
   if kind==8:
    sector=None;h=decode(pos,8)
    if h[3]==track and h[1]==h[2]^h[3]^h[4]^h[5]:sector=h[2]
   elif kind==7 and sector is not None:
    d=decode(pos,260);checksum=0
    for v in d[1:257]:checksum^=v
    if checksum==d[257]:sectors[track,sector]=d[1:257]
    sector=None
  except (ValueError,IndexError):pass
  pos+=1
entries=[];directory=sectors[18,1]
for i in range(8):
 e=directory[i*32:i*32+32]
 if not e[2]:continue
 track,sector=e[3:5];data=b'';visited=set()
 while track:
  assert (track,sector) not in visited;visited.add((track,sector));d=sectors[track,sector]
  data+=d[2:] if d[0] else d[2:d[1]+1];track,sector=d[:2]
 h=2166136261
 for v in data[2:]:h=((h^v)*16777619)&0xffffffff
 entries.append(dict(name=e[5:21].rstrip(b'\xa0').decode('ascii'),address=int.from_bytes(data[:2],'little'),payloadBytes=len(data)-2,fnv1a=h,sha256=hashlib.sha256(data).hexdigest()))
print(json.dumps(dict(imageSha256=hashlib.sha256(image).hexdigest(),validSectors=len(sectors),files=entries),indent=2))
