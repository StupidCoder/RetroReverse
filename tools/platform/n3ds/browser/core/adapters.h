#pragma once
#include "arm-memory.h"
arm_CPU*arm_NewCPU(n3ds_Machine*m){auto c=arenaNew(arm_CPU{});c->bus=m;c->wide=nullptr;arm_CPU_Reset(c);return c;}
uint32_t arm_CPU_read16(arm_CPU*c,uint32_t a){if(auto p=rrarm::ordinary(c->bus,a,2))return rrarm::load<uint16_t>(p);return n3ds_Machine_Read16(c->bus,a);}
uint32_t arm_CPU_read32aligned(arm_CPU*c,uint32_t a){a&=~3u;if(auto p=rrarm::ordinary(c->bus,a,4))return rrarm::load<uint32_t>(p);return n3ds_Machine_Read32(c->bus,a);}
void arm_CPU_write16(arm_CPU*c,uint32_t a,uint32_t v){if(auto p=rrarm::ordinary(c->bus,a,2)){rrarm::store(p,uint16_t(v));return;}n3ds_Machine_Write16(c->bus,a,v);}
void arm_CPU_write32aligned(arm_CPU*c,uint32_t a,uint32_t v){a&=~3u;if(auto p=rrarm::ordinary(c->bus,a,4)){rrarm::store(p,v);return;}n3ds_Machine_Write32(c->bus,a,v);}
n3ds_workPool* n3ds_GPU_pool(n3ds_GPU*g){if(!g->workers)g->workers=arenaNew(n3ds_workPool{});return g->workers;}
void n3ds_workPool_run(n3ds_workPool*,int64_t n,std::function<void(int64_t)>f){for(int64_t i=0;i<n;i++)f(i);}
void n3ds_Machine_Close(n3ds_Machine*){}

uint32_t n3ds_Machine_Read16(n3ds_Machine*m,uint32_t a){return n3ds_Machine_Read(m,a)|uint32_t(n3ds_Machine_Read(m,a+1))<<8;}
uint32_t n3ds_Machine_Read32(n3ds_Machine*m,uint32_t a){return n3ds_Machine_ReadWord(m,a);}
void n3ds_Machine_Write16(n3ds_Machine*m,uint32_t a,uint32_t v){n3ds_Machine_Write(m,a,v);n3ds_Machine_Write(m,a+1,v>>8);}
void n3ds_Machine_Write32(n3ds_Machine*m,uint32_t a,uint32_t v){n3ds_Machine_WriteWord(m,a,v);}

// Texture cache entries and temporary decoder images have bounded C++ ownership.
inline std::unordered_map<n3ds_texImage*,std::unique_ptr<n3ds_texImage>> rrTextureOwners;
inline n3ds_texImage*rrNewTexture(n3ds_texImage v){auto p=std::make_unique<n3ds_texImage>(std::move(v));auto raw=p.get();rrTextureOwners.emplace(raw,std::move(p));return raw;}
inline void rrCollectTextures(n3ds_GPU*g){if(rrTextureOwners.size()<=g->texCache.size()+16)return;std::unordered_map<n3ds_texImage*,bool>live;for(auto[k,p]:g->texCache)live[p]=true;for(auto it=rrTextureOwners.begin();it!=rrTextureOwners.end();)if(!live.contains(it->first))it=rrTextureOwners.erase(it);else ++it;}
inline Slice<uint8_t>rrDecodeETC(Slice<uint8_t>d,int64_t w,int64_t h,bool alpha){auto at=arena.size();auto image=alpha?n3ds_decodeETC1A4(d,w,h):n3ds_decodeETC1(d,w,h);auto p=image->Pix;while(arena.size()>at){arena.back()();arena.pop_back();}return p;}
