"""Owned CPU opcode metadata. Unstable silicon-dependent opcodes stay explicit."""
import json
from pathlib import Path
UNSTABLE={0x8b:'XAA',0xab:'LAX immediate',0x93:'AHX indirect,Y',0x9f:'AHX absolute,Y',0x9b:'TAS',0x9c:'SHY',0x9e:'SHX'}
def opcodes():
 root=Path(__file__).resolve().parents[4]
 text=(root/'site/emulators/opcodes6502.js').read_text()
 result={int(k):(v['mnemonic'],v['mode']) for k,v in json.loads(text.split('export const opcodes = ',1)[1].strip().removesuffix(';')).items()}
 for base,name in [(0,'SLO'),(0x20,'RLA'),(0x40,'SRE'),(0x60,'RRA'),(0xc0,'DCP'),(0xe0,'ISC')]:
  for code,mode in [(3,'izx'),(7,'zp'),(15,'abs'),(19,'izy'),(23,'zpx'),(27,'aby'),(31,'abx')]:result[base+code]=(name,mode)
 for name,pairs in [('SAX',[(0x83,'izx'),(0x87,'zp'),(0x8f,'abs'),(0x97,'zpy')]),('LAX',[(0xa3,'izx'),(0xa7,'zp'),(0xaf,'abs'),(0xb3,'izy'),(0xb7,'zpy'),(0xbf,'aby')]),('NOP',[(c,'imp') for c in [0x1a,0x3a,0x5a,0x7a,0xda,0xfa]]+[(c,'imm') for c in [0x80,0x82,0x89,0xc2,0xe2]]+[(c,'zp') for c in [4,0x44,0x64]]+[(c,'zpx') for c in [0x14,0x34,0x54,0x74,0xd4,0xf4]]+[(0x0c,'abs')]+[(c,'abx') for c in [0x1c,0x3c,0x5c,0x7c,0xdc,0xfc]])]:
  for code,mode in pairs:result[code]=(name,mode)
 for code,name in [(0x0b,'ANC'),(0x2b,'ANC'),(0x4b,'ALR'),(0x6b,'ARR'),(0xcb,'AXS'),(0xeb,'SBC')]:result[code]=(name,'imm')
 result[0xbb]=('LAS','aby')
 for code in [2,0x12,0x22,0x32,0x42,0x52,0x62,0x72,0x92,0xb2,0xd2,0xf2]:result[code]=('JAM','imp')
 assert set(result)|set(UNSTABLE)==set(range(256))
 return result
