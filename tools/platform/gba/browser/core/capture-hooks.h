#pragma once
namespace rrgba {
constexpr uint32_t palBase=0x1000,vramBase=0x2000,oamBase=0x1a000,frameBase=0x20000,frameBytes=240*160*4,memorySize=frameBase+frameBytes*7;
struct Source{uint32_t address=0,value=0,palette=0,paletteValue=0,texel=0;int u=0,v=0;};
inline std::array<std::array<Source,240>,5>sources{};
inline int currentY=0;inline unsigned currentObject=0;
inline uint32_t objectSource=0;inline int objectU=0,objectV=0;
inline void clean(){auto&t=rrcapture::trace;t.source=t.palette=t.texel=t.sourceValue=t.paletteValue=0;t.sourceBefore=t.paletteBefore=t.writes.size();t.u=t.v=0;}
inline uint32_t event(gbamachine_Machine*m,const char*kind,int layer,int y){auto&t=rrcapture::trace;if(!t.active)return 0;clean();std::array<uint8_t,0x60>regs{};for(int i=0;i<0x60;i+=2){auto v=get(m->io,uint32_t(i));regs[i]=v;regs[i+1]=v>>8;}std::ostringstream s;s<<"{\"kind\":\""<<kind<<"\",\"layer\":"<<layer<<",\"scanline\":"<<y<<",\"object\":"<<(layer==4?int(currentObject):-1)<<",\"registerSnapshot\":\"GBA display registers 04000000–0400005F\"}";auto e=t.event(m->Steps,m->cpu->R[15],s.str(),t.resource(regs.data(),regs.size()));return e;}
inline void source(gbamachine_Machine*m,int layer,int x,uint32_t a,int idx,int bank,int u,int v){if(!rrcapture::trace.active||x<0||x>=240||layer<0||layer>=5)return;Source s;s.address=vramBase+a;s.value=m->vram[a];s.texel=idx;s.u=u;s.v=v;if(bank>=0){auto p=bank*32+idx*2;if(p>=0&&p+1<m->pal.n){s.palette=palBase+p;s.paletteValue=m->pal[p]|uint32_t(m->pal[p+1])<<8;}}else{s.value=m->vram[a]|uint32_t(m->vram[a+1])<<8;}sources[layer][x]=s;}
inline void dot(gbamachine_Machine*m,int layer,int x,uint16_t color){auto&t=rrcapture::trace;if(!t.active)return;auto s=sources[layer][x];t.source=s.address;t.sourceValue=s.value;t.palette=s.palette;t.paletteValue=s.paletteValue;t.texel=s.texel;t.u=s.u;t.v=s.v;t.sourceBefore=t.paletteBefore=t.writes.size();t.record(frameBase+(layer+1)*frameBytes+(currentY*240+x)*4,gbamachine_rgb15(color),4,m->Steps,m->cpu->R[15],t.current);}
inline void beginLine(gbamachine_Machine*m,int y){if(!rrcapture::trace.active)return;currentY=y;sources={};event(m,"PPU scanline begins",-1,y);}
inline void compose(gbamachine_Machine*m,int x,int y,int top,int second,uint16_t topC,uint16_t secondC,uint16_t color,uint16_t ctl){auto&t=rrcapture::trace;if(!t.active)return;clean();uint32_t at=frameBase+(y*240+x)*4;if(top<5){t.source=frameBase+(top+1)*frameBytes+(y*240+x)*4;t.sourceValue=gbamachine_rgb15(topC);}else{t.source=palBase;t.sourceValue=topC;}t.texel=top;t.u=x;t.v=y;
 // The source links to the actual rendered layer sample; the event keeps the
 // second layer and effect inputs for blending/brightness inspection.
 {std::ostringstream s;s<<"{\"kind\":\"PPU priority / window / color effects\",\"scanline\":"<<y<<",\"x\":"<<x<<",\"topLayer\":"<<top<<",\"secondLayer\":"<<second<<",\"topColor15\":"<<topC<<",\"secondColor15\":"<<secondC<<",\"windowMask\":"<<ctl<<",\"blendControl\":"<<get(m->io,0x50u)<<"}";auto e=t.event(m->Steps,m->cpu->R[15],s.str());}
 t.record(at,gbamachine_rgb15(color),4,m->Steps,m->cpu->R[15],t.current);
}
inline std::vector<uint8_t>memory(gbamachine_Machine*m,bool final){std::vector<uint8_t>b(memorySize);if(final&&rrcapture::trace.shadow.size()==memorySize)b=rrcapture::trace.shadow;std::memcpy(b.data()+palBase,m->pal.p,m->pal.n);std::memcpy(b.data()+vramBase,m->vram.p,m->vram.n);std::memcpy(b.data()+oamBase,m->oam.p,m->oam.n);std::memcpy(b.data()+frameBase,m->screen.data(),frameBytes);return b;}
}
inline void rrGBAMemWrite(gbamachine_Machine*m,const Slice<uint8_t>&b,uint32_t i){uint32_t a;if(b.p==m->pal.p)a=rrgba::palBase;else if(b.p==m->vram.p)a=rrgba::vramBase;else if(b.p==m->oam.p)a=rrgba::oamBase;else return;rrgba::clean();auto&t=rrcapture::trace;auto previous=t.current;t.current=0;t.record(a+i,b[i],1,m->Steps,m->cpu->R[15]);t.current=previous;}
