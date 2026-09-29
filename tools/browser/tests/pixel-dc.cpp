#include "../../platform/dc/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 machine=dc_NewMachine(nullptr);auto*m=machine;m->PVRRegs[0x44/4]=5;m->PVRRegs[0x48/4]=1;m->PVRRegs[0x5c/4]=15|(31<<10)|(1<<20);m->PVRRegs[0x60/4]=m->PVRRegs[0x50/4]=0;
 m->TAClosed=Slice<uint8_t>::make(128);auto put=[&](int a,uint32_t v){le_PutUint32(rrBorrow(m->TAClosed,a,a+4),v);};put(0,4u<<29);put(4,7u<<29);
 for(int i=0;i<3;i++){int a=32+i*32;put(a,7u<<29);put(a+4,std::bit_cast<uint32_t>(i==1?30.f:1.f));put(a+8,std::bit_cast<uint32_t>(i==2?30.f:1.f));put(a+12,std::bit_cast<uint32_t>(1.f));put(a+16,0xffff0000);}
 assert(rr_capture_begin());dc_Machine_renderFrame(m);assert(rr_capture_end());assert(rrcapture::trace.overflow==0);auto f=rrFrame(m);assert(f[(4*32+4)*4]==248);assert(f[(30*32+30)*4]==0);auto proof=std::string(rr_proof());rr_replay_begin();auto n=rrreplay::replay.steps.size();assert(n>=4);while(!rr_replay_seek(n)){}assert(!std::memcmp(f.data(),rr_replay_frame(),f.size()));assert(std::string(rr_pixel(4,4)).find("\"complete\":true")!=std::string::npos);while(!rr_replay_seek(0)){}assert(std::memcmp(f.data(),rr_replay_frame(),f.size()));assert(proof==rr_proof());
 auto size=rr_state_save();assert(size);std::vector<uint8_t>s(rr_state_data(),rr_state_data()+size);m->RAM[20]=7;std::memcpy(rr_state_input(size),s.data(),size);assert(rr_state_load(size));assert(proof==rr_proof());std::memcpy(rr_state_input(size-1),s.data(),size-1);assert(!rr_state_load(size-1));assert(proof==rr_proof());
 // RGB888 can cross a bank boundary inside a single pixel; retain the
 // individual byte addresses for both scanout and contributor queries.
 m=machine;
 for(unsigned depth=0;depth<4;depth++) {
  m->PVRRegs[0x44/4]=1|(depth<<2);m->PVRRegs[0x5c/4]=5|(2<<10)|(1<<20);m->PVRRegs[0x50/4]=3;
  assert(rr_capture_begin());
  for(unsigned i=0;i<80;i++){auto a=dc_vram32to64(i);rrcapture::trace.record(a,uint8_t(i*37),1,0,0);m->VRAM[a]=uint8_t(i*37);}
  assert(rr_capture_end());auto expected=rrFrame(m);rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}
  assert(!std::memcmp(expected.data(),rr_replay_frame(),expected.size()));
  auto pixel=std::string(rr_pixel(0,0));assert(pixel.find("\"complete\":true")!=std::string::npos);
  unsigned n=depth<2?2:3,v=0;for(unsigned i=0;i<n;i++)v|=unsigned(uint8_t((3+i)*37))<<(i*8);
  assert(pixel.find("\"reconstructed\":"+std::to_string(v)+",")!=std::string::npos);
 }
 m->PVRRegs[0x44/4]=0;assert(rr_capture_begin());assert(rr_capture_end());rr_replay_begin();auto black=rrFrame(m);assert(black.size()==640*480*4);assert(!std::memcmp(black.data(),rr_replay_frame(),black.size()));assert(std::string(rr_pixel(0,0)).find("\"blank\":true")!=std::string::npos);
 std::cout<<"Dreamcast TA primitives, clear, replay and state checks pass\n";
}
