#pragma once
#include "../../../../browser/state/translated.h"
inline void stateFields(rrstate::Archive&a,image_Point&v){a(v.X,v.Y);}
inline void stateFields(rrstate::Archive&a,image_Rectangle&v){a(v.Min,v.Max);}
inline void stateFields(rrstate::Archive&a,image_RGBA&v){a(v.Pix,v.Stride,v.Rect);}
#include "state-fields.h"
inline void stateFields(rrstate::Archive&a,RunContext&v){a(v.steps,v.sinceNew,v.seen,v.ring,v.ri,v.unstickTries,v.stallSwitches);}
inline void rebindState(threedo_Machine*m){
 if(!m->CPU||!m->vol||m->dram.n!=2*1024*1024||m->vram.n!=1024*1024||m->tasks.n>65536||m->cur<0||m->cur>=m->tasks.n)throw std::runtime_error("Invalid 3DO machine state");
 for(auto&[id,stream]:*m->streams.p){if(!stream)throw std::runtime_error("Missing disc stream");auto [bytes,e]=threedo_Volume_ReadFile(m->vol,stream->name);if(e||stream->pos<0||stream->pos>bytes.n)throw std::runtime_error("Invalid disc stream");stream->data=bytes;}
 for(int64_t i=m->moviePos;i<m->movieQueue.n;i++){
  auto movie=m->movieQueue[i];if(!movie||!movie->mv)throw std::runtime_error("Invalid movie state");
  auto [bytes,error]=threedo_Volume_ReadFile(m->vol,movie->name);if(error)throw std::runtime_error("Missing movie stream");
  auto [decoded,err]=threedo_DemuxStream(bytes);
  if(err||decoded->Width!=movie->mv->Width||decoded->Height!=movie->mv->Height||decoded->Times.n!=movie->mv->Times.n)throw std::runtime_error("Movie identity mismatch");
  movie->mv->Frames=decoded->Frames;decoded->Frames={};
 }
 m->CPU->bus=m;m->CPU->SWI=[m](auto...args){return threedo_Machine_swi(m,args...);};
 m->OnDisplay=[](threedo_Machine*m,uint64_t,uint32_t){m->StopRequested=true;};
}
