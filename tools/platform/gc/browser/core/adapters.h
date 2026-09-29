#pragma once
inline gekko_CPU*gekko_NewCPU(gc_Machine*m){auto*c=arenaNew(gekko_CPU{});c->bus=c->fetcher=m;gekko_CPU_Reset(c);return c;}
inline uint32_t gekko_CPU_fetch(gekko_CPU*c,uint32_t a){auto[p,ok]=gekko_CPU_Translate(c,a,false,true);return ok?gc_Machine_Fetch32(c->bus,p):0;}
inline void gc_gpu_fill(gc_gpu*g,gc_Machine*m,gc_tevState*tev,Slice<gc_rasterTri>tris){rrprof::Scope timing(3,"TEV and software rasterizer");gc_gpu_ensureRaster(g);gc_rstats stats{};g->profSerFills++;for(auto&t:tris)gc_gpu_fillTri(g,m,tev,&t,t.minY,t.maxY,&stats);gc_gpu_mergeStats(g,&stats);}
