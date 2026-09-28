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
s=p.read_text()
marker='auto tmp80 = threedo_Machine_decodePixel(m,cel,v,flags,bgnd);'
if 't.sourceBefore=lrform' not in s:
 s=s.replace(marker,'''if(rrcapture::trace.active){auto&t=rrcapture::trace;
 t.sourceBefore=lrform?t.writes.size():(t.current?t.events[t.current-1].sourceBefore:0);
 t.paletteBefore=t.current?t.events[t.current-1].sourceBefore:0;t.paletteValue=0;
 if(lrform)t.sourceValue=((v>>8)&255)|((v&255)<<8);
 else if(!cel->Packed&&t.source>=src&&uint64_t(t.source-src)+2<=uint64_t(data.n))t.sourceValue=uint32_t(data[t.source-src])|(uint32_t(data[t.source-src+1])<<8);
}
'''+marker)
 s=s.replace('if(rrcapture::trace.active)rrcapture::trace.palette+=idx*2;', 'if(rrcapture::trace.active){rrcapture::trace.palette+=idx*2;rrcapture::trace.paletteValue=(uint32_t(cel->PLUT[idx])>>8)|((uint32_t(cel->PLUT[idx])&255)<<8);}')
 p.write_text(s)
# N64: retain the last filter tap's TMEM/TLUT location. The filtered texel remains
# separately recorded by OnPixel; this is not misrepresented as the sole tap.
p=root/'tools/platform/n64/browser/core/generated.cpp';s=p.read_text()
if '// RR_TMEM_EVIDENCE' not in s:
 s='#include "../../../../browser/core/capture.h"\n// RR_TMEM_EVIDENCE\n'+s
 target='uint32_t row = cast<uint32_t>((cast<uint32_t>((t->TMem * cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>((ty * t->Line)) * cast<uint32_t>(8ULL)))));'
 assert target in s
 s=s.replace(target,target+'\nif(rrcapture::trace.active){auto&tr=rrcapture::trace;tr.source=n64_swizzle(row+((sx*(4u<<t->Size))/8),ty)&4095;tr.palette=0;tr.sourceValue=n64_rdp_tmem16(r,tr.source);}')
 target='uint16_t n64_rdp_tlut(n64_rdp* r,uint32_t i){\n{'
 assert target in s
 s=s.replace(target,target+'\nif(rrcapture::trace.active){rrcapture::trace.palette=0x800+((i&255)*8);}')
 p.write_text(s)
