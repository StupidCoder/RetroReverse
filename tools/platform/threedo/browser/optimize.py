#!/usr/bin/env python3
"""Reviewed C++ specializations applied after regenerating/instrumenting the Go port.

Keeps the original algorithms and watched-memory fallbacks. Optimizations remove
callback type erasure, per-bit decoding and redundant per-byte bus dispatch.
Run after instrument.py/capture-instrument.py, before slice.py and build.py.
"""
from pathlib import Path
p=Path(__file__).resolve().parent/'core/generated.cpp'
s=p.read_text()
marker='// C++ hot paths specialized by optimize.py.\n'
if s.startswith(marker):raise SystemExit(0)
# Statically dispatch locally known closures. No std::function allocations or indirect pixel calls.
import re
s=re.sub(r'std::function<[^\n]+> (\w+) = (\[&\])',r'auto \1 = \2',s)
for name in ['threedo_Cel_decodePacked','threedo_Cel_decodeUnpacked','threedo_Machine_decodeLRForm16']:
 s=re.sub(r'void '+name+r'\(([^\n]+)std::function<void\(int64_t,int64_t,uint32_t\)> set\)',r'template<class Put> void '+name+r'(\1Put&& set)',s)

a=s.index('uint32_t threedo_bitReader_read(threedo_bitReader* br,int64_t n){');b=s.index('// tools/platform/',a)
s=s[:a]+'''uint32_t threedo_bitReader_read(threedo_bitReader* br,int64_t n){
  if(n<=0)return 0;
  if(br->pos<0)throw std::runtime_error("negative bit position");
  uint32_t value=0;
  // At most five byte loads for a 32-bit token, rather than one load per bit.
  while(n>0){
    const int take=int(std::min<int64_t>(n,8-(br->pos&7)));
    const auto byte=br->pos>>3;
    const uint32_t v=byte<br->data.n?br->data.p[byte]:0;
    value=(value<<take)|((v>>(8-(br->pos&7)-take))&((1u<<take)-1));
    br->pos+=take;n-=take;
  }
  return value;
}
''' + s[b:]
# Skip costly runtime divisions when the sample footprint is a single pixel.
needle='int64_t stepsV = edgeSteps(cast<int64_t>((xv - x0)),cast<int64_t>((yv - y0)));'
pos=s.index(needle)+len(needle)
s=s[:pos]+'''
if(stepsH==1&&stepsV==1&&!m->CelDebug){
 const auto x=x0>>16,y=y0>>16;
 if(x<0||y<0||x>=bm.w||y>=bm.h){offN++;return;}
 if(transparent){auto&t=rrcapture::trace;uint32_t a=bm.buf+uint32_t((y/2)*bm.w*4+x*4+(y&1)*2);t.record(a,0,2,m->CPU->Instrs,m->CPU->cur,t.current,4);return;}
 if(rrcapture::trace.active)rrcapture::trace.texel=pix;
 threedo_Machine_blendPixel(m,bm,x,y,pix,amv,pixc,flags);written++;return;
}
''' +s[pos:]
# Snapshot source through a bulk copy if it is contained in one RAM bank. Keep
# snapshot semantics for overlapping source/destination and read watchpoints.
a=s.index('Slice<uint8_t> data = Slice<uint8_t>::make(max);',s.index('bool threedo_Machine_drawOneCel(',s.index('// tools/platform/threedo/graphicsfolio.go:351')))
b=s.index('cel->PDAT = data;',a)
s=s[:a]+'''Slice<uint8_t> data = Slice<uint8_t>::make(max);
const uint8_t* source=nullptr;
if(!m->OnRead){
 if(src<0x200000&&uint64_t(max)<=0x200000-src)source=m->dram.p+src;
 else if(src>=0x200000&&src<0x300000&&uint64_t(max)<=0x300000-src)source=m->vram.p+src-0x200000;
 else if(src>=0x400000&&src<0x800000&&uint64_t(max)<=0x800000-src)source=m->imem.p+src-0x400000;
}
if(source)std::memcpy(data.p,source,max);
else for(int64_t i=0;i<max;i++)data[i]=threedo_Machine_Read(m,src+uint32_t(i));
''' +s[b:]

needle='threedo_Machine_Write(m,a,cast<uint8_t>(shr<uint16_t>(pix,cast<int64_t>(8ULL))));\nthreedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(pix));'
assert needle in s
s=s.replace(needle,'''if(!m->OnWrite&&a>=0x200000&&a<0x2fffff){auto p=m->vram.p+a-0x200000;p[0]=pix>>8;p[1]=pix;}
else {'''+needle+'}',1)
needle='for (uint32_t a = dest;(cast<uint32_t>((a + cast<uint32_t>(1ULL))) < end);a += cast<uint32_t>(2ULL)){'
assert needle in s
s=s.replace(needle,'''if(!m->OnWrite&&dest>=0x200000&&dest<0x300000&&bytes<=0x300000-dest){
 auto p=m->vram.p+dest-0x200000;
 for(uint32_t i=0;i+1<bytes;i+=2){p[i]=hi;p[i+1]=lo;}
}else '''+needle,1)

needle='m->vram[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = cast<uint8_t>(c);'
assert needle in s
s=s.replace(needle,needle+'\nif(rrcapture::trace.active)rrcapture::trace.record(0x200000+o,uint32_t(c>>8)|uint32_t(c&255)<<8,2,m->CPU->Instrs,m->CPU->cur,rrcapture::trace.current);')
p.write_text(marker+s)
