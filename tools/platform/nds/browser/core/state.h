#pragma once
#include "../../../../browser/state/translated.h"
inline void stateFields(rrstate::Archive&a,image_Point&v){a(v.X,v.Y);}
inline void stateFields(rrstate::Archive&a,image_Rectangle&v){a(v.Min,v.Max);}
inline void stateFields(rrstate::Archive&a,image_RGBA&v){a(v.Pix,v.Stride,v.Rect);}
#include "state-fields.h"
inline void rebindState(dsmachine_Machine*m,Slice<uint8_t>rom){
 if(!m||!m->ARM9||!m->ARM7||!m->ARM9->cpu||!m->ARM7->cpu||!m->cd||!m->spi||!m->vram||!m->gpu2d||!m->gpu3d||m->ram.n!=4194304||m->swram.n!=32768||m->pal.n!=2048||m->oam.n!=2048||m->spi->firmware.n!=262144||m->vid.line<0||m->vid.line>=263)throw std::runtime_error("Invalid DS machine state");
 if(m->gpu2d->a.out.n!=49152||m->gpu2d->b.out.n!=49152||(m->gpu3d->rast.frame.n!=0&&m->gpu3d->rast.frame.n!=49152)||(m->gpu2d->threeD.n!=0&&m->gpu2d->threeD.n!=49152))throw std::runtime_error("Invalid display surface");
 for(int i=0;i<2;i++){auto c=i?m->ARM7:m->ARM9;c->m=m;auto bus=arenaNew(dsmachine_bus{c});c->cpu->bus=c->cpu->wide=bus;c->cpu->SWI=dsmachine_biosSWI(c);if(!i)c->cpu->Coproc=dsmachine_cp15(c);}
 m->cd->rom=rom;
 for(int i=0;i<9;i++)if(m->vram->bank[i].n!=dsmachine_bankSizes[i])throw std::runtime_error("Invalid VRAM bank");
 for(int i=0;i<11;i++)m->vram->pages[i]=Slice<Slice<Slice<uint8_t>>>::make(dsmachine_spaceSize[i]/8192);
 dsmachine_vram_remap(m->vram);
 m->gpu2d->a.m=m;m->gpu2d->b.m=m;
 m->prof.on=false;
}
