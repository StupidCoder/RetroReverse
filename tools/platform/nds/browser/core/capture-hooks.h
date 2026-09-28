#pragma once
#include "../../../../browser/core/capture.h"
#include "../../../../browser/core/profile.h"
#include <map>
#include <unordered_set>
namespace rrds {
// Capture-local surfaces, not guest physical addresses: engine A, engine B, 3D.
constexpr uint32_t planeSize=256*192*4,plane3D=2*planeSize,total=3*planeSize;
inline std::map<std::tuple<int,int,int,int,bool>,uint32_t> events;
inline std::unordered_set<uint32_t> render3DEvents;
inline uint32_t rgba(uint32_t v){return __builtin_bswap32(v|255);}
inline uint32_t frag(dsmachine_rfrag c){return uint32_t(dsmachine_c6to8(c.r))|uint32_t(dsmachine_c6to8(c.g))<<8|uint32_t(dsmachine_c6to8(c.b))<<16|uint32_t(dsmachine_a5to8(c.a))<<24;}
inline void clean(){auto&t=rrcapture::trace;t.source=t.palette=t.texel=t.sourceValue=t.paletteValue=t.sourceBefore=t.paletteBefore=0;t.u=t.v=0;}
inline uint32_t event(dsmachine_engine*e,const char*kind){auto&t=rrcapture::trace;clean();std::ostringstream s;s<<"{\"kind\":\""<<kind<<"\",\"engine\":\""<<(e->isB?"B":"A")<<"\",\"addressSpace\":\"capture-local RGBA surfaces\",\"DISPCNT\":"<<dsmachine_engine_reg32(e,0)<<",\"BLDCNT\":"<<dsmachine_engine_reg16(e,80)<<",\"MASTER_BRIGHT\":"<<dsmachine_engine_reg16(e,108)<<"}";return t.event(e->m->Steps,e->m->ARM9->cpu->R[15],s.str());}
inline void write(dsmachine_engine*e,int64_t index,uint32_t v){auto&t=rrcapture::trace;if(!t.active)return;uint32_t a=(e->isB?planeSize:0)+index*4;if(t.value(t.shadow,a,4)==rgba(v))return;t.record(a,rgba(v),4,e->m->Steps,e->m->ARM9->cpu->R[15],t.current);}
inline void composite(dsmachine_engine*e,int64_t x,int64_t y,dsmachine_cand top,dsmachine_cand below,int64_t effect,bool fx){auto&t=rrcapture::trace;if(!t.active)return;clean();auto key=std::make_tuple(int(e->isB),int(top.layer),int(below.layer),int(effect),fx);auto id=events[key];if(!id){std::ostringstream s;s<<"{\"kind\":\"2D priority / blend\",\"engine\":\""<<(e->isB?"B":"A")<<"\",\"topLayer\":"<<top.layer<<",\"underLayer\":"<<below.layer<<",\"effect\":"<<effect<<",\"effectsEnabled\":"<<(fx?"true":"false")<<",\"DISPCNT\":"<<dsmachine_engine_reg32(e,0)<<",\"BLDCNT\":"<<dsmachine_engine_reg16(e,80)<<",\"BLDALPHA\":"<<dsmachine_engine_reg16(e,82)<<",\"hasSource\":"<<((top.layer==0&&e->is3D)?"true":"false")<<",\"sourceKind\":\"capture-local 3D plane; layer IDs BG0-3, OBJ=4, backdrop=5\"}";id=t.event(e->m->Steps,e->m->ARM9->cpu->R[15],s.str());events[key]=id;}
 t.current=id;if(top.layer==0&&e->is3D){auto sx=(x+dsmachine_engine_reg16(e,16))&255;t.source=plane3D+(y*256+sx)*4;t.sourceBefore=t.writes.size();t.sourceValue=t.value(t.shadow,t.source,4);}t.texel=rgba(dsmachine_bgr555RGBA(top.c));t.record((e->isB?planeSize:0)+(y*256+x)*4,rgba(dsmachine_bgr555RGBA(e->line[x])),4,e->m->Steps,e->m->ARM9->cpu->R[15],id);}
inline void clear3D(dsmachine_gpu3d*g,dsmachine_Machine*m){auto&t=rrcapture::trace;if(!t.active)return;clean();render3DEvents.insert(t.event(m->Steps,m->ARM9->cpu->R[15],"{\"kind\":\"3D clear color / depth\"}"));for(uint32_t i=0;i<49152;i++)t.record(plane3D+i*4,g->rast.col[i].a?frag(g->rast.col[i]):0,4,m->Steps,m->ARM9->cpu->R[15],t.current);}
inline void publish(dsmachine_raster*r,int64_t i){auto&t=rrcapture::trace;if(t.active&&t.value(t.shadow,plane3D+i*4,4)!=r->frame[i])t.record(plane3D+i*4,r->frame[i],4,0,0,t.current);}
inline std::vector<uint8_t> memory(dsmachine_Machine*m){std::vector<uint8_t>b(total);for(int engine=0;engine<2;engine++){auto s=engine?m->gpu2d->b.out:m->gpu2d->a.out;for(int64_t i=0;i<s.n&&i<49152;i++){auto v=rgba(s[i]);std::memcpy(b.data()+engine*planeSize+i*4,&v,4);}}for(int64_t i=0;i<m->gpu2d->threeD.n&&i<49152;i++){auto v=m->gpu2d->threeD[i];std::memcpy(b.data()+plane3D+i*4,&v,4);}return b;}
}
