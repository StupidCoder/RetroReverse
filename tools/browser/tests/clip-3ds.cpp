#include "../../platform/n3ds/browser/core/host.h"
#include <cassert>
int main(){
 auto m=arenaNew(n3ds_Machine{});m->CPU=arm_NewCPU(m);auto g=n3ds_newGPU(m);
 n3ds_fbState fb{};fb.width=fb.height=8;fb.vpHalfW=fb.vpHalfH=4;
 n3ds_vsOut a{},b{},c{};a.pos={-.5f,-.5f,-.5f,1};b.pos={.5f,-.5f,-.5f,1};c.pos={0,2,.5f,-1};
 auto tris=n3ds_GPU_clipTri(g,{},a,b,c,&fb);assert(tris.n>0);
 bool covered=false;
 for(auto t:tris){
  for(auto v:{t.v0,t.v1,t.v2})assert(std::isfinite(v.x)&&std::isfinite(v.y)&&v.x>=-.001f&&v.x<=8.001f&&v.y>=-.001f&&v.y<=8.001f&&v.z>=-1.001f&&v.z<=.001f&&v.iw>0);
  covered|=n3ds_edgeFn(t.v0.x,t.v0.y,t.v1.x,t.v1.y,4,3)>=0&&n3ds_edgeFn(t.v1.x,t.v1.y,t.v2.x,t.v2.y,4,3)>=0&&n3ds_edgeFn(t.v2.x,t.v2.y,t.v0.x,t.v0.y,4,3)>=0;
 }
 assert(covered);
 for(auto invalid:{std::numeric_limits<float>::quiet_NaN(),std::numeric_limits<float>::infinity()}){
  c.pos[0]=invalid;assert(n3ds_GPU_clipTri(g,{},a,b,c,&fb).n==0);
 }
}
