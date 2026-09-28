#!/usr/bin/env python3
"""Idempotent trace-only adaptations to the translated 3DO renderer."""
from pathlib import Path
root=Path(__file__).resolve().parents[2]
p=root/'tools/platform/threedo/browser/core/generated.cpp'
s=p.read_text()
if '// RR_CAPTURE_INSTRUMENTED' not in s:
 s='#include "../../../../browser/core/capture.h"\n// RR_CAPTURE_INSTRUMENTED\n'+s
 s=s.replace('rrprof::Scope rrclock(2,"Cel software rasterizer");','rrprof::Scope rrclock(2,"Cel software rasterizer");\nrrcapture::RenderScope traceRendering;')
 s=s.replace('calls++;\nauto tmp80 =', '''calls++;
if(rrcapture::trace.active){auto&t=rrcapture::trace;t.u=sx;t.v=sy;t.texel=v;t.palette=plutPtr;
 t.source=cel->Packed?0:lrform?src+uint32_t((sy/2)*(cel->Width/2)*4+sx*4+(sy&1)*2):src+uint32_t(sy*((cel->Width*cel->BPP+31)/32)*4+(sx*cel->BPP)/8);}
auto tmp80 =''')
 s=s.replace('if (transparent) {\nclearN++;\nreturn ;\n}', 'if (transparent) {\nclearN++;\nif(!rrcapture::trace.active)return ;\n}')
 s=s.replace('threedo_Machine_blendPixel(m,bm,x,y,pix,amv,pixc,flags);', '''if(transparent){auto&t=rrcapture::trace;uint32_t a=bm.buf+uint32_t((y/2)*bm.w*4+x*4+(y&1)*2);t.record(a,0,2,m->CPU->Instrs,m->CPU->cur,t.current,4);continue;}
if(rrcapture::trace.active)rrcapture::trace.texel=pix;
threedo_Machine_blendPixel(m,bm,x,y,pix,amv,pixc,flags);''')
 # Retain the selected PLUT entry address, not just the table base.
 target='uint16_t raw = cel->PLUT[idx];'
 if target in s:s=s.replace(target,'if(rrcapture::trace.active)rrcapture::trace.palette+=idx*2;\n'+target)
 else: print('PLUT exact-entry hook requires review; table base is retained')
 p.write_text(s)

# Apply the exact palette-entry hook to already instrumented translations too.
s=p.read_text()
if 'rrcapture::trace.palette+=idx*2' not in s:
 s=s.replace('uint16_t raw = cel->PLUT[idx];','if(rrcapture::trace.active)rrcapture::trace.palette+=idx*2;\nuint16_t raw = cel->PLUT[idx];')
 p.write_text(s)

s=p.read_text()
needle='cel->PDAT = data;'
if 'trace.resource(data.p,data.n)' not in s:
 s=s.replace(needle,needle+'\nif(rrcapture::trace.active&&rrcapture::trace.current)rrcapture::trace.events[rrcapture::trace.current-1].resource=rrcapture::trace.resource(data.p,data.n);')
 p.write_text(s)
