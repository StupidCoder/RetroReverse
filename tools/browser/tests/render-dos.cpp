// Exercise the real CPU store/copy paths, not a synthetic list of trace records.
#define RR_DOS_RAW_CAP 128
#include "../../platform/dos/browser/core/api.cpp"
#include <cassert>
#include <iostream>

struct Machine {
  dos_Machine real{}; dos_PM pm{}; dos_vgaState vga{}; dos_ioState io{};
  x86_CPU *cpu; bool protectedMode;
  Machine(bool pmode = false) : protectedMode(pmode) {
    rrcapture::trace.active = rrdos::collecting = false;
    realMachine = rrDOSReal = nullptr; protectedMachine = rrDOSProtected = nullptr;
    if (pmode) {
      pm.Mem = Slice<uint8_t>::make(65 * 1048576); cpu = pm.CPU = x86_NewCPU(&pm); cpu->Mode = 1;
      protectedMachine = rrDOSProtected = &pm;
      for (int i=0;i<256;i++) pm.Pal[i*3]=i&63;
    } else {
      real.Mem = Slice<uint8_t>::make(1048576); real.vga = &vga; real.io = &io;
      cpu = real.CPU = x86_NewCPU(&real); realMachine = rrDOSReal = &real;
      dos_Machine_vgaInit13h(&real, true);
      for (int i=0;i<256;i++) io.Pal[i*3]=i&63;
    }
  }
  auto &ram() { return protectedMode ? pm.Mem : real.Mem; }
  void instruction(uint32_t pc, std::initializer_list<uint8_t> bytes) {
    std::copy(bytes.begin(), bytes.end(), ram().p + pc);
    cpu->Seg[1] = protectedMode ? 0 : pc >> 4; cpu->SegBase[1] = 0;
    cpu->IP = protectedMode ? pc : pc & 15;
    x86_CPU_Step(cpu); assert(!cpu->Halted);
  }
  void store(uint32_t a, uint8_t value, uint32_t pc=0x12345) {
    cpu->Seg[3] = a >> 4; cpu->SegBase[3] = 0;
    if (protectedMode) instruction(pc, {0xc6,0x05,uint8_t(a),uint8_t(a>>8),uint8_t(a>>16),uint8_t(a>>24),value});
    else instruction(pc, {0xc6,0x06,uint8_t(a&15),0,value});
  }
  void copy(uint32_t dst, uint32_t src, int bytes=1, int count=1, bool reverse=false) {
    cpu->Seg[3] = src >> 4; cpu->Seg[0] = dst >> 4; cpu->SegBase[3] = cpu->SegBase[0] = 0;
    cpu->Regs[6] = protectedMode ? src : src & 15; cpu->Regs[7] = protectedMode ? dst : dst & 15;
    cpu->Regs[1] = count;
    cpu->DF = reverse;
    if (bytes==1) instruction(0x23000, {0xf3,0xa4});
    else if ((bytes==4)==protectedMode) instruction(0x23000, {0xf3,0xa5});
    else instruction(0x23000, {0x66,0xf3,0xa5});
  }
  void finish() { assert(rr_capture_end()); rr_replay_begin(); }
  void seek(uint32_t step) { while(!rr_replay_seek(step)){} }
  uint32_t writer(uint32_t guest, uint32_t pc) {
    const auto a=rrdos::packed(guest);
    for(uint32_t i=0;i<rrcapture::trace.writes.size();i++) {
      const auto&w=rrcapture::trace.writes[i];
      if(w.address<=a&&a<w.address+w.size&&w.pc==pc) return i+1;
    }
    assert(false); return 0;
  }
  ~Machine(){realMachine=rrDOSReal=nullptr;protectedMachine=rrDOSProtected=nullptr;}
};

int main() {
  for (bool pmode : {false,true}) {
    Machine m(pmode);
    assert(rr_capture_begin());
    m.store(0x40000,4); m.store(0x40001,7,0x12456);
    m.copy(0xa0000,0x40000,2);
    // The same buffer is reused before capture ends: the display and its query
    // must stop at the copy, not read the newer RAM value.
    m.store(0x40000,9,0x12567); m.finish();
    auto &p=rrdos::pixels[0]; assert(p.source==0x40000&&rrdos::pixels[1].source==0x40001);
    auto evidence=std::string(rr_pixel(0,0));
    assert(evidence.find("\"complete\":true")!=std::string::npos);
    assert(evidence.find("\"pc\":74565")!=std::string::npos); // 0x12345, including CS
    assert(evidence.find("\"pc\":75111")==std::string::npos); // 0x12567, after copy
    m.seek(rr_replay_for_write(m.writer(0x40000,0x12345)));
    assert(rr_replay_frame()[0]==16&&rr_replay_frame()[4]==0);
    rr_replay_view(1); assert(rr_replay_frame()[0]==16); // reveal keeps written pixel
    auto proof=std::string(rr_proof()); auto final=rrFrame();
    m.seek(rrreplay::replay.steps.size()); assert(!memcmp(rr_replay_frame(),final.data(),final.size()));
    m.seek(0); assert(std::string(rr_proof())==proof);
    rr_replay_view(0);
  }
  { // Chunky RAM -> planar VGA, including independent source buffers.
    Machine m; m.vga.seq[4]&=~8; assert(rr_capture_begin());
    for(int i=0;i<8;i++)m.store(0x40000+i,i+1);
    m.store(0x50000,20);
    for(int plane=0;plane<4;plane++) {
      m.vga.seq[2]=1<<plane;
      m.copy(0xa0000,0x40000+plane);
      m.copy(0xa0001,plane==3?0x50000:0x40004+plane);
    }
    m.finish();
    for(int i=0;i<8;i++)assert(rrdos::pixels[i].source==(i==7?0x50000:0x40000+i));
    m.seek(rrreplay::replay.steps.size());auto final=rrFrame();assert(!memcmp(rr_replay_frame(),final.data(),final.size()));
  }
  { // Multiple back buffers: preserve each MOVS edge and historical cutoff.
    Machine m(true); assert(rr_capture_begin());
    m.store(0x40000,12); m.copy(0x50000,0x40000);
    m.store(0x40000,24); m.copy(0xa0000,0x50000); m.finish();
    auto id=m.writer(0x50000,0x23000);auto w=rrcapture::trace.writes[id-1];
    assert(w.source==rrdos::packed(0x40000)&&w.sourceValue==12);
    auto source=std::string(rr_source(w.source,w.size,w.sourceBefore,w.sourceValue));
    assert(source.find("\"reconstructed\":12")!=std::string::npos&&source.find("\"complete\":true")!=std::string::npos);
  }
  { // REP with DF, and a later direct VGA store replacing one copied byte.
    Machine m(true); assert(rr_capture_begin());
    for(int i=0;i<4;i++)m.store(0x40000+i,10+i);
    m.copy(0xa0003,0x40003,1,4,true); m.store(0xa0002,42); m.finish();
    assert(rrdos::pixels[0].source==0x40000&&rrdos::pixels[3].source==0x40003);
    assert(rrdos::pixels[2].source==rrdos::NONE);
    assert(std::string(rr_pixel(2,0)).find("\"complete\":true")!=std::string::npos);
  }
  { // Real-mode source wraps at both a 16-bit offset and the 20-bit bus.
    Machine m; assert(rr_capture_begin());
    m.store(0x4ffff,10); m.store(0x40000,11);
    m.cpu->Seg[3]=0x4000;m.cpu->Seg[0]=0xa000;m.cpu->Regs[6]=0xffff;m.cpu->Regs[7]=0;
    m.instruction(0x23000,{0xa5});
    m.store(0xfffff,12);m.store(0,13);m.copy(0xa0002,0xfffff,2);m.finish();
    assert(rrdos::pixels[0].source==0x4ffff&&rrdos::pixels[1].source==0x40000);
    assert(rrdos::pixels[2].source==0xfffff&&rrdos::pixels[3].source==0);
    for(int i=0;i<4;i++)assert(std::string(rr_pixel(i,0)).find("\"complete\":true")!=std::string::npos);
  }
  { // A masked VGA write is not a literal copy, even if values happen to match.
    Machine m; m.vga.seq[4]&=~8;m.vga.seq[2]=1;m.vga.gc[8]=0x0f;
    assert(rr_capture_begin());m.store(0x40000,15);m.copy(0xa0000,0x40000);m.finish();
    assert(rrdos::producerPixels==0&&rrdos::regions.empty());
  }
  { // Identical-value writes remain visible with the explicit reveal view.
    Machine m(true);m.store(0x40000,40);assert(rr_capture_begin());
    m.store(0x40000,40);m.copy(0xa0000,0x40000);m.finish();
    m.seek(0);rr_replay_view(0);assert(rr_replay_frame()[0]==160);
    rr_replay_view(1);assert(rr_replay_frame()[0]==32);
    m.seek(rr_replay_for_write(m.writer(0x40000,0x12345)));assert(rr_replay_frame()[0]==160);
  }
  { // Mode 3 retains the existing emulator's mode-0 approximation.
    Machine m;m.vga.seq[4]&=~8;m.vga.seq[2]=1;m.vga.gc[5]=3;
    assert(rr_capture_begin());m.store(0x40000,15);m.copy(0xa0000,0x40000);m.finish();
    assert(rrdos::pixels[0].source==0x40000);
  }
  { // Quiet gaps delimit two VGA bursts; no-draw execution has a bounded timeout.
    Machine m;assert(rr_capture_begin());m.store(0xa0000,1);m.cpu->Steps+=20000;
    rr_capture_info();assert(rrdos::bursts==1);
    m.store(0xa0001,2);m.cpu->Steps+=20000;
    assert(std::string(rr_capture_info()).find("\"ready\":true")!=std::string::npos);m.finish();
    assert(rr_capture_begin());m.cpu->Steps+=rrFramePeriod()*140;
    rr_capture_info();assert(rrdos::timedOut);m.finish();
  }
  { // Bounded raw tracing must report loss instead of inventing complete history.
    Machine m;assert(rr_capture_begin());
    for(unsigned i=0;i<rrdos::RAW_CAP+4;i++)m.store(0x40000+i,1);
    m.copy(0xa0000,0x40000);m.finish();assert(rrcapture::trace.overflow>0);
    assert(std::string(rr_pixel(0,0)).find("\"complete\":false")!=std::string::npos);
  }
  std::cout<<"DOS render buffers: CPU writer PCs, MOVS ancestry, Mode X, reverse/wrapped copies, RAM reuse and reveal replay pass\n";
}
