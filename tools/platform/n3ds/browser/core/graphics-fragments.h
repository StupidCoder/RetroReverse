#pragma once
// Exact hybrid path: reference float32 coverage/interpolation and texture
// sampling; integer TEV, alpha test, blend/logic and color masks on the GPU.
// Fragment links preserve primitive order independently for each target pixel.
namespace rrgpu {
inline uint32_t rgba(n3ds_rgba c){return uint32_t(c.r)|(uint32_t(c.g)<<8)|(uint32_t(c.b)<<16)|(uint32_t(c.a)<<24);}
inline Operation Operation::fragments(n3ds_GPU*g,n3ds_fbState*fb,n3ds_lightState*ls,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris){
 Operation o;auto m=g->m;
 if((!enabled&&!recording)||m->OnRead||m->OnWrite||m->OnPixel||m->HidTrace||m->Profile||ls->enabled||fb->shadowMode||fb->depthTest||fb->depthWr||(g->Regs[0x105]&1)||!tris.n||tris.n>1024||!fb->width||!fb->height||fb->width>1024||fb->height>1024||fb->width%8||fb->height%8)return o;
 uint32_t blend=g->Regs[0x101];if(g->Regs[0x100]&256){if((blend&7)>4||((blend>>8)&7)>4)return o;for(int s=16;s<=28;s+=4)if(((blend>>s)&15)>14)return o;}
 if(!std::get<1>(n3ds_tevState_run(tv,{},{},{},{})))return o;
 // Shadow sampling changes canonical counters; it remains in Reference.
 if((tv->texEnable&1)&&((g->Regs[0x83]>>28)&7)!=0)return o;
 for(int u=0;u<3;u++)if(tv->texEnable&(1u<<u)){auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);if((g->Regs[fmt]&15)>13)return o;}
 const uint32_t pixels=fb->width*fb->height;
 if(!range(m,fb->colorAddr,uint64_t(pixels)*4))return o;
 uint64_t work=0;
 for(auto&t:tris){
  if(t.minX<0||t.minY<0||t.maxX>fb->width||t.maxY>fb->height||t.minX>t.maxX||t.minY>t.maxY||!std::isfinite(t.area)||!(t.area>0))return o;
  work+=uint64_t(t.maxX-t.minX)*(t.maxY-t.minY);
  for(auto*v:{&t.v0,&t.v1,&t.v2}){
   if(!std::isfinite(v->x)||!std::isfinite(v->y)||!(v->iw>=0x1p-20f&&v->iw<=0x1p20f))return o;
   for(auto c:v->col)if(!std::isfinite(c)||std::abs(c)>1e6f)return o;
   for(auto uv:v->uv)for(auto c:uv)if(!std::isfinite(c)||std::abs(c)>1024)return o;
   if(!std::isfinite(v->uv0w)||std::abs(v->uv0w)>1024)return o;
  }
 }
 // Preparation is bounded and speculative. A fallback has committed no writes.
 if(work<4096||uint64_t(pixels)*4+work*20>16u*1024*1024)return o;
 std::vector<uint32_t> data(pixels,UINT32_MAX),tails(pixels,UINT32_MAX);data.reserve(pixels+work*5);
 for(auto&t:tris){auto&a=t.v0;auto&b=t.v1;auto&c=t.v2;
  for(int64_t y=t.minY;y<t.maxY;y++)for(int64_t x=t.minX;x<t.maxX;x++){
   float px=float(x)+.5f,py=float(y)+.5f;
   float w0=n3ds_edgeFn(b.x,b.y,c.x,c.y,px,py),w1=n3ds_edgeFn(c.x,c.y,a.x,a.y,px,py),w2=n3ds_edgeFn(a.x,a.y,b.x,b.y,px,py);
   if(w0<0||w1<0||w2<0)continue;
   float l0=w0/t.area,l1=w1/t.area,l2=w2/t.area,iw=(l0*a.iw+l1*b.iw)+l2*c.iw;
   if(!(iw>0)||!std::isfinite(iw))return o;
   auto pc=[&](float x,float y,float z){return (((l0*x)*a.iw+(l1*y)*b.iw)+(l2*z)*c.iw)/iw;};
   n3ds_rgba color{n3ds_clamp255(pc(a.col[0],b.col[0],c.col[0])*255.f),n3ds_clamp255(pc(a.col[1],b.col[1],c.col[1])*255.f),n3ds_clamp255(pc(a.col[2],b.col[2],c.col[2])*255.f),n3ds_clamp255(pc(a.col[3],b.col[3],c.col[3])*255.f)};
   std::array<uint32_t,3> tex{};n3ds_rstats st{};float q=pc(a.uv0w,b.uv0w,c.uv0w);
   for(int u=0;u<3;u++)if(tv->texEnable&(1u<<u)){
    float s=pc(a.uv[u][0],b.uv[u][0],c.uv[u][0]),t=pc(a.uv[u][1],b.uv[u][1],c.uv[u][1]);
    auto[value,ok]=n3ds_GPU_sampleTextureSt(g,u,s,t,q,&st);if(!ok)return o;tex[u]=rgba(value);
   }
   uint32_t pixel=n3ds_tiledOffset(x,fb->height-1-y,fb->width)/4,at=data.size();
   if(tails[pixel]==UINT32_MAX)data[pixel]=at;else data[tails[pixel]]=at;tails[pixel]=at;
   data.insert(data.end(),{UINT32_MAX,rgba(color),tex[0],tex[1],tex[2]});
  }
 }
 std::vector<uint32_t> params={fb->width,fb->height,fb->colorMask,g->Regs[0x100],blend,g->Regs[0x102],g->Regs[0x103],rgba(tv->bufColor),uint32_t(tv->alphaTest)|(uint32_t(tv->alphaFunc)<<4)|(uint32_t(tv->alphaRef)<<8),uint32_t((data.size()-pixels)/5)};
 for(auto&s:tv->stages){uint32_t col=0,alpha=0;for(int j=0;j<3;j++){col|=(s.colr[j].src|(s.colr[j].op<<4))<<(j*8);alpha|=(s.alph[j].src|(s.alph[j].op<<4))<<(j*8);}params.insert(params.end(),{col,alpha,uint32_t(s.combC)|(uint32_t(s.combA)<<8)|(uint32_t(s.scaleC)<<16)|(uint32_t(s.scaleA)<<20)|(uint32_t(s.updC)<<24)|(uint32_t(s.updA)<<25),rgba(s.konst)});}
 params.push_back(UINT32_MAX); // Expected drawn count supplied only by recording.
 if(recording&&stream.size()+data.size()*4+uint64_t(pixels)*8+params.size()*4+256>limit){dropped++;return o;}
 o.start(m,5,0,0,fb->colorAddr,uint64_t(pixels)*4,std::move(params));if(!o.target)return o;
 o.fragmentWords=std::move(data);o.source=reinterpret_cast<uint8_t*>(o.fragmentWords.data());o.inputSize=o.fragmentWords.size()*4;if(recording)o.input.assign(o.source,o.source+o.inputSize);o.statsOwner=g;o.statsBefore=g->PixelsDrawn;
 return o;
}
inline Operation Operation::draw(n3ds_GPU*g,n3ds_fbState*fb,n3ds_lightState*ls,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris){auto o=stencil(g,fb,ls,tv,tris);if(o.target)return o;return fragments(g,fb,ls,tv,tris);}
}
