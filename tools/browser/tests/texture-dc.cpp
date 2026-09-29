#include "../../platform/dc/browser/core/api.cpp"
#include <cassert>
#include <fstream>
#include <iostream>
static Slice<uint8_t> packet(uint32_t pcw,uint32_t tsp=0,uint32_t tcw=0){auto p=Slice<uint8_t>::make(64);le_PutUint32(p,pcw);le_PutUint32(rrBorrow(p,4,8),7u<<29);le_PutUint32(rrBorrow(p,8,12),tsp);le_PutUint32(rrBorrow(p,12,16),tcw);return p;}
static void put(Slice<uint8_t>p,int at,uint32_t v){le_PutUint32(rrBorrow(p,at,at+4),v);}
static void seek(unsigned n){while(!rr_replay_seek(n)){};}
int main(int argc,char**argv){
 machine=dc_NewMachine(nullptr);auto*m=machine;
 for(unsigned i=0;i<8388608;i++)m->VRAM[i]=uint8_t((i*37+(i>>8)*13)^((i>>4)&255));
 for(unsigned i=0;i<1024;i++)m->PVRRegs[0x400+i]=0x80204060u+i*0x01020408u;
 std::ofstream fixtures;if(argc>1)fixtures.open(argv[1]);unsigned cases=0;
 for(unsigned fmt:{0,1,2,5,6})for(unsigned layout=0;layout<(fmt<3?3:1);layout++)for(unsigned mip=0;mip<2;mip++)for(unsigned pal=0;pal<(fmt>=5?4:1);pal++){
  m->PVRRegs[0x108/4]=pal;
  auto p=packet((4u<<29)|8,2u<<3|1,0x2000/8|(fmt<<27)|(layout==1?1u<<26:layout==2?1u<<30:0)|(mip<<31)|(fmt==5?3u<<21:fmt==6?1u<<25:0));
  auto before=std::string(rr_proof());assert(rr_capture_begin());assert(rr_vram_size()==0);rrdc::command(m,"test header",p);assert(rr_capture_end());rr_replay_begin();seek(1);
  auto state=rrdc::textures::selected();assert(state.known&&state.tcw==le_Uint32(rrBorrow(p,12,16)));assert(rr_vram_size()==8388608);
  dc_renderState st{};st.m=m;dc_renderState_loadHeader(&st,le_Uint32(p),p);
  if(fixtures){auto info=std::string(rr_vram_info());info.pop_back();fixtures<<info<<",\"expected\":[";bool first=true;for(int y=0;y<st.texVH;y+=3)for(int x=0;x<st.texUW;x+=5){auto[r,g,b,a]=dc_renderState_sample(&st,float(x)/st.texUW,float(y)/st.texVH);if(!first)fixtures<<',';first=false;fixtures<<'['<<x<<','<<y<<','<<r<<','<<g<<','<<b<<','<<a<<']';}fixtures<<"]}\n";}
  assert(before==rr_proof());cases++;
 }
 // Historical palette and VRAM must rewind independently of the live machine.
 auto p=packet((4u<<29)|9,0,0x2000/8|(5u<<27));m->PVRRegs[0x400]=0xff112233;m->PVRRegs[0x108/4]=3;
 assert(rr_capture_begin());rrdc::command(m,"A",p);const auto firstEvent=rrcapture::trace.current;
 auto v=packet(7u<<29);put(v,16,0x3f003e80);for(int i=0;i<3;i++)rrdc::command(m,"vertex",v);assert(rrdc::textures::last.draw&&rrdc::textures::last.count==3);assert(rrdc::textures::last.uv[0][0]==.5f&&rrdc::textures::last.uv[0][1]==.25f);
 const auto old=m->VRAM[0x2000];m->VRAM[0x2000]=123;rrcapture::trace.record(0x2000,123,1,0,0);
 m->PVRRegs[0x400]=0xffaabbcc;rrdc::command(m,"B",p);auto secondEvent=rrcapture::trace.current;
 // Strip reset on end-of-strip and changing headers; sprites derive corner D.
 put(v,0,(7u<<29)|(1u<<28));rrdc::command(m,"end strip",v);rrdc::command(m,"new vertex",v);assert(rrdc::textures::last.count==1&&!rrdc::textures::last.draw);
 p=packet((5u<<29)|8);rrdc::command(m,"sprite header",p);put(v,52,0);put(v,56,0x3f800000);put(v,60,0x3f803f80);rrdc::command(m,"sprite",v);assert(rrdc::textures::last.count==4&&rrdc::textures::last.uv[3][0]==0&&rrdc::textures::last.uv[3][1]==1);
 assert(rr_capture_end());auto proof=std::string(rr_proof());rr_replay_begin();unsigned aStep=0,bStep=0;for(unsigned i=0;i<rrreplay::replay.steps.size();i++){if(rrreplay::replay.steps[i].event==firstEvent)aStep=i+1;if(rrreplay::replay.steps[i].event==secondEvent)bStep=i+1;}
 seek(bStep);assert(std::string(rr_vram_info()).find("4289379276")!=std::string::npos);assert(rr_vram_data()[0x2000]==123);
 seek(aStep);assert(std::string(rr_vram_info()).find("4279312947")!=std::string::npos);assert(rr_vram_data()[0x2000]==old);
 seek(0);assert(!rrdc::textures::selected().known);assert(std::string(rr_vram_info()).find("\"known\":false")!=std::string::npos);assert(proof==rr_proof());
 assert(rr_capture_begin());assert(rrdc::textures::states.empty());assert(rr_vram_size()==0);assert(rr_capture_end());
 std::cout<<"Dreamcast historical bindings, palette/VRAM rewind, strips/sprites and "<<cases<<" sampler fixtures pass\n";
}
