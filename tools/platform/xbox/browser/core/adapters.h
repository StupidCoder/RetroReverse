#pragma once
uint8_t rrRead(xbox_Machine*m,uint32_t a){return xbox_Machine_Read(m,a);}void rrWrite(xbox_Machine*m,uint32_t a,uint8_t v){xbox_Machine_Write(m,a,v);}
void xbox_pgraph_rasterParallel(xbox_pgraph*g,Slice<xbox_kelvinVtx>v,Slice<std::array<int64_t,3>>tris){xbox_rstats stats{};for(auto t:tris)xbox_pgraph_rasterTriBand(g,&v[t[0]],&v[t[1]],&v[t[2]],0,1,&stats);xbox_pgraph_mergeStats(g,&stats);}
SHA1*xbox_Machine_shaLoad(xbox_Machine*m,uint32_t ctx){auto*h=go_sha1_New();auto b=get(m->shaCtx,ctx);if(b.n==96){for(int i=0;i<5;i++)h->h[i]=be_Uint32(sub(b,4+i*4,b.n));memcpy(h->buffer.data(),b.p+24,64);h->length=be_Uint64(sub(b,88,96));}return h;}
void xbox_Machine_shaStore(xbox_Machine*m,uint32_t ctx,SHA1*h){auto b=Slice<uint8_t>::make(96);memcpy(b.p,"sha\x01",4);for(int i=0;i<5;i++)be_PutUint32(sub(b,4+i*4,96),h->h[i]);memcpy(b.p+24,h->buffer.data(),h->length%64);be_PutUint64(sub(b,88,96),h->length);m->shaCtx[ctx]=b;}
