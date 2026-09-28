#include "../../../../browser/core/capture.h"
#include "proof.h"
#include <memory>
#include <emscripten.h>
// clang-format off
EM_JS(int, browserRead, (uint32_t offset, uint8_t *dest, uint32_t n), {
  try {
    if (!Module.discFile) throw Error('No local disc mounted');
    const bytes = new Uint8Array(new FileReaderSync().readAsArrayBuffer(Module.discFile.slice(offset, offset+n)));
    if (bytes.length !== n) throw Error('Short local disc read');
    HEAPU8.set(bytes,dest); return 1;
  } catch(e) { Module.discError=String(e); return 0; }
});
// clang-format on
static std::vector<u8> input;
static std::unique_ptr<Machine> machine;
static std::unique_ptr<Machine> saved[4];
static std::vector<u32> frame;
static std::string text, error;
static u32 generation = 0;
static bool capturing = false, captured = false;
static u32 originX = 0, originY = 0;
static void clearCapture() {
  if (machine) {
    machine->gpu.onCommand = {};
    machine->gpu.onPixel = {};
  }
  capturing = captured = false;rrcapture::trace.active=rrcapture::trace.valid=false;
}
static void requireMachine() {
  if (!machine)
    throw std::runtime_error("Load a disc first");
}
template <class F> static int guard(F fn) {
  try {
    error.clear();
    return fn();
  } catch (const std::exception &e) {
    error = e.what();
    return -1;
  }
}
extern "C" {
u8 *rr_input(int n) {
  if (n < 0 || n > 128 * 1024 * 1024) {
    error = "Disc exceeds the 128 MiB input limit";
    return nullptr;
  }
  input.resize(n);
  return input.data();
}
int rr_init(int n, u32 handler) {
  return guard([&]() {
    if (n < 0 || size_t(n) > input.size())
      throw std::runtime_error("Invalid input length");
    auto d = std::make_shared<Disc>();
    d->open(input.data(), n);
    auto candidate = std::make_unique<Machine>();
    candidate->boot(d, handler);
    clearCapture();
    machine = std::move(candidate);
    for (auto &s : saved)
      s.reset();
    generation++;
    return 1;
  });
}
int rr_init_file(uint32_t n) {
  return guard([&]() {
    auto d = std::make_shared<Disc>();
    d->openFile(n, [](size_t o,u8*p,size_t n){
      if (!browserRead(o,p,n)) throw std::runtime_error("Local disc read failed");
    });
    auto candidate = std::make_unique<Machine>();
    candidate->boot(d,0);
    clearCapture(); machine=std::move(candidate);
    for(auto &s:saved) s.reset();
    generation++; return 1;
  });
}
const char* rr_profile(){return rrprof::json();}
int rr_run(int n, int stopField) {
  return guard([&]() {
    requireMachine();
    if (n < 1 || n > 250000)
      throw std::runtime_error("Instruction budget must be 1–250,000");
    captured = false;
    rrprof::Scope clock(0,"R3000A / devices remainder");
    return int(machine->run(n, stopField));
  });
}
void rr_pad(int buttons) {
  if (machine)
    machine->buttons = u16(buttons);
}
const char *rr_error() {
  return error.c_str();
}
const char *rr_status() {
  if (!machine)
    return "{\"loaded\":false}";
  auto &m = *machine;
  std::ostringstream s;
  s << "{\"loaded\":true,\"generation\":" << generation << ",\"steps\":" << m.cpu.steps
    << ",\"fields\":" << m.fields << ",\"pc\":" << m.cpu.pc << ",\"width\":" << m.gpu.dispW
    << ",\"height\":" << m.gpu.dispH << ",\"displayX\":" << m.gpu.dispX
    << ",\"displayY\":" << m.gpu.dispY << ",\"drawX\":" << m.gpu.offX << ",\"drawY\":" << m.gpu.offY
    << ",\"commands\":" << m.gpu.commands << ",\"cdCommands\":" << m.cd.commands
    << ",\"pad\":" << m.buttons << ",\"padReady\":" << (m.padActive ? "true" : "false")
    << ",\"captured\":" << (captured ? "true" : "false") << ",\"disc\":" << quote(m.disc->name)
    << "}";
  text = s.str();
  return text.c_str();
}
const char *rr_proof() {
  if (!machine)
    return "{}";
  text = proof(*machine);
  return text.c_str();
}
u32 *rr_frame(int draw) {
  if (!machine)
    return nullptr;
  frame = machine->gpu.frame(draw);
  return frame.data();
}
u8 *rr_ram() {
  return machine ? machine->ram.data() : nullptr;
}
int rr_save(int slot) {
  return guard([&]() {
    requireMachine();
    if (slot < 0 || slot >= 4)
      throw std::runtime_error("Invalid checkpoint slot");
    auto s = std::make_unique<Machine>(*machine);
    s->gpu.onCommand = {};
    s->gpu.onPixel = {};
    saved[slot] = std::move(s);
    return 1;
  });
}
int rr_restore(int slot) {
  return guard([&]() {
    requireMachine();
    if (slot < 0 || slot >= 4 || !saved[slot])
      throw std::runtime_error("That checkpoint has not been saved");
    clearCapture();
    *machine = *saved[slot];
    return 1;
  });
}
int rr_capture_begin() {
 return guard([&](){requireMachine();clearCapture();auto&t=rrcapture::trace;t.begin((uint8_t*)machine->gpu.vram.data(),1024*512*2);capturing=true;
 machine->gpu.onCommand=[](const std::vector<u32>&w){auto&t=rrcapture::trace;t.source=t.palette=t.texel=0;t.u=t.v=0;auto op=w.empty()?0:w[0]>>24;bool hasSource=(op>=0x20&&op<0x40&&(op&4))||(op>=0x60&&op<0x80&&(op&4))||(op>=0x80&&op<0xa0);const char*kind=op==2?"GPU fill":op>=0x20&&op<0x40?"GPU polygon":op>=0x60&&op<0x80?"GPU rectangle":op>=0x80&&op<0xa0?"VRAM copy":op>=0xa0&&op<0xc0?"CPU/DMA VRAM upload":"GPU command";std::ostringstream o;o<<"{\"kind\":\""<<kind<<"\",\"hasSource\":"<<(hasSource?"true":"false")<<",\"words\":[";for(size_t i=0;i<w.size();i++){if(i)o<<',';o<<w[i];}o<<"],\"textureDepth\":"<<machine->gpu.texDepth<<",\"drawX\":"<<machine->gpu.offX<<",\"drawY\":"<<machine->gpu.offY<<"}";t.event(machine->cpu.steps,machine->cpu.cur,o.str());};
 machine->gpu.onPixel=[](int x,int y,u16 value){auto&t=rrcapture::trace;t.record((y*1024+x)*2,value,2,machine->cpu.steps,machine->cpu.cur,t.current);};
 if(machine->gpu.imgPx)t.event(machine->cpu.steps,machine->cpu.cur,"{\"kind\":\"GPU transfer already in progress at capture start\"}");
 return 1;});
}
int rr_capture_end(){return guard([&](){requireMachine();machine->gpu.onCommand={};machine->gpu.onPixel={};capturing=false;captured=true;originX=machine->gpu.dispX;originY=machine->gpu.dispY;rrcapture::trace.end((uint8_t*)machine->gpu.vram.data(),1024*512*2);return int(rrcapture::trace.events.size());});}
const char* rr_capture_info(){text=rrcapture::trace.info();return text.c_str();}
const char* rr_pixel(int x,int y){
 if(!captured||x<0||y<0||x>=machine->gpu.dispW||y>=machine->gpu.dispH)return "{\"error\":\"Select a captured display pixel\"}";
 text=rrcapture::trace.pixel((((y+originY)&511)*1024+((x+originX)&1023))*2,2);return text.c_str();
}
const char *rr_diagnostics() {
  requireMachine();
  std::ostringstream s;
  s << "{\"tty\":" << quote(machine->tty) << ",\"messages\":[";
  for (size_t i = 0; i < machine->diagnostics.size(); i++) {
    if (i)
      s << ",";
    s << quote(machine->diagnostics[i]);
  }
  s << "]}";
  text = s.str();
  return text.c_str();
}
}
#include "state.h"
static void stateWrite(rrstate::Archive&a){requireMachine();a.header(2,1);a(*machine);}
static void stateRead(rrstate::Archive&a){requireMachine();a.header(2,1);auto next=std::make_unique<Machine>();a(*next);a.finish();validateState(*next);next->disc=machine->disc;clearCapture();machine=std::move(next);for(auto&s:saved)s.reset();}
#include "../../../../browser/state/api.inc"

extern "C" const char*rr_source(uint32_t address,int size,uint32_t before,uint32_t expected){text=rrcapture::trace.pixel(address,size,before,expected);return text.c_str();}
extern "C" const char*rr_resource(uint32_t id,uint32_t offset){text=rrcapture::trace.resourceJSON(id,offset);return text.c_str();}
