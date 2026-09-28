#pragma once
#include "../../../../browser/state/translated.h"
// A single 3DS heap region can exceed the generic collection element bound.
inline void stateFields(rrstate::Archive&a,Slice<uint8_t>&s){uint32_t n=s.n;a(n);if(n>rrstate::limit)throw std::runtime_error("Invalid memory region size");if(a.reading){a.charge(n);s=Slice<uint8_t>::make(n);}a.raw(s.p,n);}
inline void stateFields(rrstate::Archive&a,image_Point&v){a(v.X,v.Y);}
inline void stateFields(rrstate::Archive&a,image_Rectangle&v){a(v.Min,v.Max);}
inline void stateFields(rrstate::Archive&a,image_RGBA&v){a(v.Pix,v.Stride,v.Rect);}
#include "state-fields.h"
inline void rebindState(n3ds_Machine*m,n3ds_RomFS*romfs,Slice<uint8_t>raw){
 if(!m||!m->CPU||!m->gpu||m->regions.n>128||m->threads.n>256||!m->curThread)throw std::runtime_error("Invalid 3DS machine state");
 m->romfs=romfs;m->romfsRaw=raw;m->pages={};m->Profile=false;m->SingleThreaded=true;
 for(auto r:m->regions){if(!r||uint64_t(r->base)+r->data.n>0x100000000ULL)throw std::runtime_error("Invalid memory region");n3ds_Machine_indexRegion(m,r);}
 auto bind=[m](arm_CPU*c){c->bus=m;c->wide=nullptr;c->SWI=[m](auto...v){return n3ds_Machine_handleSVC(m,v...);};c->Coproc=[m](auto...v){return n3ds_Machine_handleCP15(m,v...);};};bind(m->CPU);for(auto t:m->threads){if(!t)throw std::runtime_error("Invalid thread");bind(&t->ctx);}
 m->gpu->m=m;n3ds_GPU_invalidateShaders(m->gpu);m->gpu->texCache={};
 for(auto[h,f]:m->fsFiles){if(!f)throw std::runtime_error("Invalid file session");if(!f->save.empty())continue;if(f->path=="<romfs-l3>"){auto off=romfs?romfs->Levels[2].Offset:0;f->data=sub(raw,off,raw.n);}else if(romfs){auto[d,e]=n3ds_RomFS_File(romfs,f->path);if(!e)f->data=d;}}
}
