#pragma once
// Upload immutable triangles, coarse ordered bins and decoded textures. Coverage,
// perspective interpolation and sampling execute on the GPU; no per-pixel CPU
// lists are built. The old hybrid path remains a fallback for larger inputs.
namespace rrgpu {
inline bool rasterEnabled=true;
inline uint32_t floatWord(float f){uint32_t bits;std::memcpy(&bits,&f,4);return bits;}
inline std::vector<uint32_t> fragmentParams(n3ds_GPU*g,n3ds_fbState*fb,n3ds_tevState*tv,uint32_t count){
 std::vector<uint32_t> p={fb->width,fb->height,fb->colorMask,g->Regs[0x100],g->Regs[0x101],g->Regs[0x102],g->Regs[0x103],rgba(tv->bufColor),uint32_t(tv->alphaTest)|(uint32_t(tv->alphaFunc)<<4)|(uint32_t(tv->alphaRef)<<8),count};
 for(auto&s:tv->stages){uint32_t col=0,alpha=0;for(int j=0;j<3;j++){col|=(s.colr[j].src|(s.colr[j].op<<4))<<(j*8);alpha|=(s.alph[j].src|(s.alph[j].op<<4))<<(j*8);}p.insert(p.end(),{col,alpha,uint32_t(s.combC)|(uint32_t(s.combA)<<8)|(uint32_t(s.scaleC)<<16)|(uint32_t(s.scaleA)<<20)|(uint32_t(s.updC)<<24)|(uint32_t(s.updA)<<25),rgba(s.konst)});}
 p.push_back(UINT32_MAX);return p;
}
inline Operation raster(n3ds_GPU*g,n3ds_fbState*fb,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris,uint64_t work){
 Operation o;if(!rasterEnabled)return o;
 const bool depth=fb->depthTest;const uint64_t bytes=uint64_t(fb->width)*fb->height*4;
 auto depthTarget=depth?range(g->m,fb->depthAddr,bytes):nullptr;
 if(depth&&(!depthTarget||overlap(range(g->m,fb->colorAddr,bytes),bytes,depthTarget,bytes)||fb->depthFunc>7||!std::isfinite(fb->depthScale)||!std::isfinite(fb->depthOff)||std::abs(fb->depthScale)>0x1p20f||std::abs(fb->depthOff)>0x1p20f))return o;
 if(depth)for(auto&t:tris)for(auto*v:{&t.v0,&t.v1,&t.v2})if(!std::isfinite(v->z)||std::abs(v->z)>0x1p20f)return o;
 for(auto&t:tris){if(t.area<0x1p-20f||t.area>0x1p28f)return o;for(auto*v:{&t.v0,&t.v1,&t.v2})if(std::abs(v->x)>4096||std::abs(v->y)>4096)return o;}
 rrprof::Scope preparation(6,"PICA coverage / sampling / GPU inputs");
 const uint32_t nx=(fb->width+15)/16,ny=(fb->height+15)/16,header=nx*ny*2;
 std::vector<uint32_t> data(header);std::vector<std::vector<uint32_t>> bins(nx*ny);
 for(auto&t:tris){
  uint32_t at=data.size();data.insert(data.end(),{uint32_t(t.minX),uint32_t(t.maxX),uint32_t(t.minY),uint32_t(t.maxY),floatWord(t.area)});
  for(auto*v:{&t.v0,&t.v1,&t.v2}){data.insert(data.end(),{floatWord(v->x),floatWord(v->y),floatWord(v->iw)});for(auto c:v->col)data.push_back(floatWord(c));for(auto uv:v->uv)for(auto c:uv)data.push_back(floatWord(c));if(depth)data.push_back(floatWord(v->z));}
  if(t.minX==t.maxX||t.minY==t.maxY)continue;
  for(uint32_t y=t.minY/16;y<uint32_t(t.maxY+15)/16;y++)for(uint32_t x=t.minX/16;x<uint32_t(t.maxX+15)/16;x++)bins[y*nx+x].push_back(at);
 }
 for(uint32_t i=0;i<bins.size();i++){data[i*2]=data.size();data[i*2+1]=bins[i].size();data.insert(data.end(),bins[i].begin(),bins[i].end());}
 auto params=fragmentParams(g,fb,tv,work);params.insert(params.end(),{uint32_t(tris.n),nx,ny,uint32_t(tv->texEnable&7)});
 for(int u=0;u<3;u++){
  if(!(tv->texEnable&(1u<<u))){params.insert(params.end(),{0,0,0,0});continue;}
  auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);uint32_t w=(g->Regs[dim]>>16)&2047,h=g->Regs[dim]&2047;
  params.insert(params.end(),{uint32_t(data.size()),w,h,g->Regs[param]});if(!w||!h)continue;
  if((data.size()+uint64_t(w)*h)*4>16u*1024*1024)return o;
  const uint32_t address=n3ds_Machine_gpuAddrToVirt(g->m,g->Regs[addr]<<3),format=g->Regs[fmt]&15;
  n3ds_texImage*img=nullptr;
  if(depth){
   // Early depth rejection can avoid every Reference texture fetch. Preparing
   // such a draw must not populate a cold cache with a snapshot that a later
   // draw would observe after intervening depth writes. Reuse existing entries;
   // Reference warms missing textures only when it actually samples them.
   auto[cached,ok]=lookup(g->texCache,n3ds_texKey{address,format,w,h});if(!ok)return o;img=cached;
  }else{auto[decoded,ok]=n3ds_GPU_texture(g,address,format,w,h);if(!ok)return o;img=decoded;}
  size_t at=data.size();data.resize(at+uint64_t(w)*h);std::memcpy(data.data()+at,img->pix.p,uint64_t(w)*h*4);
 }
 if(depth)params.insert(params.end(),{1u|(uint32_t(fb->depthWr)<<1)|(uint32_t(fb->depthZBuffer)<<2)|(fb->depthFunc<<4),floatWord(fb->depthScale),floatWord(fb->depthOff),0,UINT32_MAX});
 if(recording&&stream.size()+data.size()*4+bytes*(depth?4:2)+params.size()*4+256>limit){dropped++;return o;}
 o.start(g->m,depth?7:6,0,0,fb->colorAddr,bytes,std::move(params));if(!o.target)return o;
 if(depth){o.depthTarget=depthTarget;o.killedBefore=g->DepthKilled;if(recording)o.before.insert(o.before.end(),depthTarget,depthTarget+bytes);}
 o.fragmentWords=std::move(data);o.source=reinterpret_cast<uint8_t*>(o.fragmentWords.data());o.inputSize=o.fragmentWords.size()*4;if(recording)o.input.assign(o.source,o.source+o.inputSize);o.statsOwner=g;o.statsBefore=g->PixelsDrawn;return o;
}
}
