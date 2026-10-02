#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>
int main() {
  machine=arenaNew(n3ds_Machine{});
  machine->CPU=arm_NewCPU(machine);
  machine->gpu=n3ds_newGPU(machine);
  n3ds_Machine_mapRegion(machine,"vram",0x1f000000,Slice<uint8_t>::make(4096));
  const uint32_t src=0x1f000800,dst=0x1f000000;
  for(uint32_t format=0;format<5;++format) {
    assert(rr_capture_begin());
    rrcapture::trace.event(0,0,"{\"kind\":\"draw red\"}");
    n3ds_Machine_WriteWord(machine,src,0xff0000ff);
    auto redWrite=rrcapture::trace.writes.size();
    rrcapture::trace.event(0,0,"{\"kind\":\"draw green\"}");
    n3ds_Machine_WriteWord(machine,src,0x00ff00ff);
    auto greenWrite=rrcapture::trace.writes.size();
    n3ds_Machine_gxDisplayTransfer(machine,src,dst,8|(8<<16),8|(8<<16),format<<12);
    // A later frame may reuse the source; the displayed copy must stay green.
    rrcapture::trace.event(0,0,"{\"kind\":\"reuse target\"}");
    n3ds_Machine_WriteWord(machine,src,0x0000ffff);
    assert(rr_capture_end());
    rr3ds::screens[0]={};rr3ds::screens[1]={};
    auto &screen=rr3ds::screens[0];screen.dst=dst;screen.w=screen.h=screen.stride=8;
    screen.format=format;screen.bpp=n3ds_fbBPP(format);
    rr_replay_begin();
    const auto offset=(7*400)*4;
    auto check=[&](uint32_t step,uint8_t red,uint8_t green,uint8_t blue){
      while(!rr_replay_seek(step)){}
      auto p=rr_replay_frame();
      assert(p[offset]==red&&p[offset+1]==green&&p[offset+2]==blue);
    };
    auto redStep=rr_replay_for_write(redWrite),greenStep=rr_replay_for_write(greenWrite);
    check(redStep,255,0,0);check(greenStep,0,255,0);
    check(rrreplay::replay.steps.size(),0,255,0);
    assert(pixels==rr3ds::display(rrcapture::trace.final));
    check(redStep,255,0,0);check(greenStep,0,255,0);
    assert(rrcapture::trace.overflow==0);
  }

  // The title's stencil-clear quad must never reach the colour buffer.
  auto g=machine->gpu;
  g->Regs[0x105]=0xff00ff01;g->Regs[0x106]=0x222;g->Regs[0x115]=1;
  n3ds_fbState fb{};fb.depthAddr=0x1f000000;
  n3ds_Machine_WriteWord(machine,fb.depthAddr,0xa5123456);
  assert(n3ds_GPU_stencilDepthTest(g,&fb,0,0)==1);
  assert(n3ds_Machine_ReadWord(machine,fb.depthAddr)==0x00123456);
  // Not-equal-zero shadows reject unmarked pixels and keep marked ones.
  g->Regs[0x105]=0xff00ff31;g->Regs[0x106]=0;
  assert(n3ds_GPU_stencilDepthTest(g,&fb,0,0)==1);
  n3ds_Machine_Write(machine,fb.depthAddr+3,1);
  assert(n3ds_GPU_stencilDepthTest(g,&fb,0,0)==0);
}
