#include "host.h"
#include <emscripten.h>
// clang-format off
EM_JS(int, browserRead, (uint32_t offset, uint8_t *dest, uint32_t n), {
  try {
    let bytes;
    if (Module.discFile)
      bytes = new Uint8Array(
          new FileReaderSync().readAsArrayBuffer(Module.discFile.slice(offset, offset + n)));
    else throw Error('No local disc mounted');
    if (bytes.length !== n)
      throw Error('Short disc read');
    HEAPU8.set(bytes, dest);
    return 1;
  } catch (e) {
    Module.discError = String(e);
    return 0;
  }
});
// clang-format on
static threedo_Machine *machine = nullptr;
static std::string errorText, reply;
static std::vector<uint8_t> pixels;
extern "C" {
int rr_init_config(uint32_t size, int nfsProfile) {
  try {
    machine = nullptr;
    discSize = size;
    discRead = [](uint64_t o, uint8_t *p, size_t n) {
      if (!browserRead(o, p, n))
        throw std::runtime_error("Disc read failed; check the file or server");
    };
    machine = boot(nfsProfile != 0);
    errorText.clear();
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_init(uint32_t size) { return rr_init_config(size, 1); }
int rr_run_slice(uint32_t n) {
  try { if(!machine || n>50000) throw std::runtime_error("Invalid execution slice");
    runSlice(machine,n); return 1;
  } catch(const std::exception&e){errorText=e.what();return 0;}
}
const char* rr_profile(){return rrprof::json();}
int rr_run() {
  try {
    if (!machine)
      throw std::runtime_error("No disc loaded");
    nextFrame(machine);
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
void rr_pad(uint32_t b) {
  if (machine)
    threedo_Machine_SendPadEvent(machine, b);
}
const char *rr_error() {
  return errorText.c_str();
}
const char *rr_proof() {
  reply = proof(machine);
  return reply.c_str();
}
const char *rr_status() {
  std::ostringstream s;
  s << "{\"frame\":" << machine->frame << ",\"steps\":" << totalSteps
    << ",\"pc\":" << machine->CPU->R[15] << ",\"discBytes\":" << discBytesRead
    << ",\"discReads\":" << discReads << "}";
  reply = s.str();
  return reply.c_str();
}
uint8_t *rr_frame() {
  pixels = frame(machine);
  return pixels.data();
}
}
#include "state.h"
static std::vector<std::shared_ptr<void>> stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(4,1);a(machine,totalSteps,runContext);}
static void stateRead(rrstate::Archive&a){a.header(4,1);threedo_Machine*next=nullptr;uint64_t ticks=0;RunContext context;a(next,ticks,context);a.finish();if(!next)throw std::runtime_error("Missing machine");rebindState(next);stateOwners=std::move(a.owned);machine=next;totalSteps=ticks;runContext=std::move(context);fileEntries.clear();discCache.clear();}
#include "../../../../browser/state/api.inc"
static uint32_t capBuffer=0;static threedo_CelDraw capCel{};
extern "C" {
int rr_capture_begin(){try{auto&t=rrcapture::trace;t.begin(machine->vram.p,machine->vram.n,0x200000);
 machine->WatchLo=0x200000;machine->WatchHi=0x300000;
 machine->OnWrite=[](uint32_t a,uint32_t v,uint32_t pc){auto&t=rrcapture::trace;if(!t.rendering){t.source=t.palette=t.texel=0;t.record(a,v,1,machine->CPU->Instrs,pc);}};
 machine->OnCel=[](threedo_CelDraw c){capCel=c;auto&t=rrcapture::trace;std::ostringstream o;o<<"{\"kind\":\"Cel / CCB\",\"ccb\":"<<c.CCB<<",\"source\":"<<c.Src<<",\"plut\":"<<c.PLUT<<",\"pixc\":"<<c.PIXC<<",\"flags\":"<<c.Flags<<",\"bpp\":"<<c.BPP<<",\"width\":"<<c.Width<<",\"height\":"<<c.Height<<",\"packed\":"<<(c.Packed?"true":"false")<<",\"lrform\":"<<(c.LRForm?"true":"false")<<",\"target\":"<<c.Bitmap<<",\"plutValues\":[";if(c.Coded&&c.PLUT)for(int i=0;i<32;i++){if(i)o<<',';o<<((uint32_t(threedo_Machine_rawRead(machine,c.PLUT+i*2))<<8)|threedo_Machine_rawRead(machine,c.PLUT+i*2+1));}o<<"]}";t.event(machine->CPU->Instrs,machine->CPU->cur,o.str());};
 machine->OnPixel=[](uint32_t x,uint32_t y,threedo_PixelEvent e){auto&t=rrcapture::trace;uint32_t a=capCel.Bitmap+(y/2)*capCel.BitmapW*4+x*4+(y&1)*2;if(a<0x200000||a+2>0x300000)return;t.record(a,rrcapture::little(machine->vram.p+a-0x200000,2),2,machine->CPU->Instrs,machine->CPU->cur,t.current,e.Drawn?1:4);};return 1;}catch(...){return 0;}}
int rr_capture_end(){machine->OnWrite={};machine->OnCel={};machine->OnPixel={};machine->WatchLo=machine->WatchHi=0;capBuffer=machine->displayBuf;rrcapture::trace.end(machine->vram.p,machine->vram.n);return 1;}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){if(!capBuffer)return "{\"blank\":true,\"contributors\":[],\"complete\":true}";if(x<0||y<0||x>=320||y>=240)return "{\"error\":\"Outside captured display\"}";reply=rrcapture::trace.pixel(capBuffer+(y/2)*320*4+x*4+(y&1)*2,2);return reply.c_str();}
}

extern "C" const char*rr_source(uint32_t address,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(address,size,before,expected);return reply.c_str();}
extern "C" const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
