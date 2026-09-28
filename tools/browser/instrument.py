#!/usr/bin/env python3
"""Reapply coarse profiling hooks after regenerating a Go-to-C++ translation."""
from pathlib import Path
import re
root=Path(__file__).resolve().parents[2]
for system,hooks in {
'n64':{'n64_Machine_runRSP':(1,'RSP'),'n64_Machine_runRDP':(2,'RDP / software rasterizer')},
 'threedo':{'threedo_Machine_swi':(5,'Portfolio SWI'),'threedo_Machine_serviceKernelCall':(1,'Portfolio HLE'),'threedo_Machine_drawOneCel':(2,'Cel software rasterizer'),'threedo_Machine_flashClearRange':(3,'Flash clear')}}.items():
 p=root/f'tools/platform/{system}/browser/core/generated.cpp';s=p.read_text()
 include='#include "../../../../browser/core/profile.h"\n'
 if include not in s:s=include+s
 for fn,(bucket,name) in hooks.items():
  pattern=r'(^[^\n;]*\b'+fn+r'\([^\n]*\)\{)(?!\nrrprof::Scope)'
  s=re.sub(pattern,lambda m:m[0]+f'\nrrprof::Scope rrclock({bucket},"{name}");',s,flags=re.M)
 if system=="threedo":s=s.replace('rrprof::Scope rrclock(5,"Portfolio SWI");','rrprof::Scope rrclock(5,"Portfolio SWI",true,comment);')
 if system=="threedo":s=s.replace('threedo_Machine_fileDeviceIO(threedo_Machine* m,std::string name,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length){','threedo_Machine_fileDeviceIOLegacy(threedo_Machine* m,std::string name,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length){')
 if system=="threedo" and 'Release closed stream buffers' not in s:
  s=s.replace('removeKey(m->streams,handle);', '// Release closed stream buffers; only the tiny arena token survives reset.\nif(auto stream=get(m->streams,handle))stream->data={};\nremoveKey(m->streams,handle);')
 p.write_text(s)
