#!/usr/bin/env python3
# Run clang's JSON AST on the declaration-only vendor headers, then emit scalar
# field traversal. Anonymous structs/arrays are expanded; pointers are rebound.
import json,re,sys
from pathlib import Path
ast=json.load(open(sys.argv[1]));records={};aliases={}
def walk(n):
 if n.get('kind') in ('RecordDecl','CXXRecordDecl') and n.get('completeDefinition'):records[n['id']]=n
 if n.get('kind')=='TypedefDecl':
  def ref(x):
   if x.get('kind')=='RecordType':aliases[n['name']]=x['decl']['id']
   for c in x.get('inner',[]):ref(c)
  ref(n)
 for c in n.get('inner',[]):walk(c)
walk(ast)
skip={'v.mem_cpu','v.mem_vic','v.debug','v.c1530','v.c1541','v.rom_char','v.rom_basic','v.rom_kernal'}
lines=['// Generated declaration-field traversal; pointer callbacks/maps are rebound.','#pragma once','inline void stateFields(rrstate::Archive&a,c64_t&v){']
def emit(rec,path,depth=0):
 pending=None
 for f in rec.get('inner',[]):
  if f.get('kind') in ('RecordDecl','CXXRecordDecl') and f.get('completeDefinition'):pending=f
  if f.get('kind')!='FieldDecl':continue
  if 'name' not in f:
   emit(pending,path,depth);continue
  p=path+'.'+f['name'];t=f['type']['qualType']
  if p in skip or '*' in f['type'].get('desugaredQualType',t) or '*' in t or '(' in t and 'struct' not in t:continue
  dims=re.findall(r'\[(\d+)\]',t);base=t.split('[')[0].strip()
  for d in dims:
   i='i'+str(depth);depth+=1;lines.append(f'for(size_t {i}=0;{i}<{d};++{i}){{');p+=f'[{i}]'
  if base in aliases:emit(records[aliases[base]],p,depth)
  elif 'unnamed' in base or 'anonymous' in base:
   assert pending,(p,t);emit(pending,p,depth)
  else:lines.append(f'a({p});')
  for d in dims:lines.append('}');depth-=1
emit(records[aliases['c64_t']],'v');lines.append('}')
Path('tools/platform/c64/browser/core/state-fields.h').write_text('\n'.join(lines)+'\n')
