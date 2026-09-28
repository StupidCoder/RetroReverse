#include "host.h"
static n64_Machine *machine = nullptr;
static Slice<uint8_t> upload;
static uint64_t steps = 0;
static std::string errorText, reply;
static std::vector<uint8_t> pixels;
extern "C" {
uint8_t *rr_input(uint32_t n) {
  try {
    if (n > 64 * 1024 * 1024)
      throw std::runtime_error("ROM exceeds 64 MiB");
    upload = Slice<uint8_t>::make(n);
    return upload.p;
  } catch (const std::exception &e) {
    errorText = e.what();
    return nullptr;
  }
}
int rr_init(uint32_t n) {
  try {
    if (upload.n != n)
      throw std::runtime_error("ROM upload size mismatch");
    auto candidate = boot(upload);
    destroy(machine);
    machine = candidate;
    steps = 0;
    errorText.clear();
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
const char* rr_profile(){return rrprof::json();}
int rr_run(uint32_t n) {
  try {
    if (!machine)
      throw std::runtime_error("No cartridge loaded");
    if (n > 50000)
      throw std::runtime_error("Run slice too large");
    rrprof::Scope clock(0,"VR4300 / scheduler remainder");
    auto r = n64_Machine_Run(machine, n);
    steps += r.Steps;
    if (machine->CPU->Halted)
      throw std::runtime_error(machine->CPU->HaltReason);
    return int(r.Steps);
  } catch (const std::exception &e) {
    errorText = e.what();
    return -1;
  } catch (...) {
    errorText = "Unexpected core exception";
    return -1;
  }
}
void rr_pad(uint32_t buttons, int x, int y) {
  if (machine) {
    machine->Controllers[0].Buttons = buttons;
    machine->Controllers[0].StickX = std::clamp(x, -80, 80);
    machine->Controllers[0].StickY = std::clamp(y, -80, 80);
  }
}
const char *rr_error() {
  return errorText.c_str();
}
const char *rr_proof() {
  reply = proof(machine, steps);
  return reply.c_str();
}
const char *rr_status() {
  auto w = n64_Machine_Width(machine);
  if (w == 0 || w > 1024)
    w = 320;
  std::ostringstream s;
  s << "{\"steps\":" << steps << ",\"pc\":" << uint32_t(machine->CPU->PC) << ",\"width\":" << w
    << ",\"height\":" << height(machine) << ",\"rspSteps\":" << machine->rspSteps
    << ",\"rdpWords\":" << machine->rdpWords << ",\"polls\":" << machine->ContPolls << "}";
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
static void stateWrite(rrstate::Archive&a){a.header(3,1);a(machine,steps);}
static void stateRead(rrstate::Archive&a){a.header(3,1);n64_Machine*next=nullptr;uint64_t ticks=0;a(next,ticks);a.finish();if(!next)throw std::runtime_error("Missing machine");rebindState(next,machine->ROM);if(stateOwners.empty())destroy(machine);stateOwners=std::move(a.owned);machine=next;steps=ticks;}
#include "../../../../browser/state/api.inc"
#include "../../../../browser/core/capture.h"
static uint32_t capOrigin,capWidth,capHeight,capType;
extern "C" {
int rr_capture_begin(){try{auto&t=rrcapture::trace;t.begin(machine->RDRAM.p,machine->RDRAM.n);
 machine->WatchLo=0;machine->WatchHi=0xffffffff;
 machine->OnWrite=[](uint32_t a,uint32_t v,uint32_t pc){auto&t=rrcapture::trace;t.source=t.palette=t.texel=0;t.u=t.v=0;t.record(a&0x1fffffff,v,1,machine->CPU->Steps,pc);};
 machine->OnRDPCmd=[](n64_Machine*m,uint32_t op,Slice<uint64_t>w){auto&t=rrcapture::trace;t.source=t.palette=t.texel=0;t.u=t.v=0;std::ostringstream o;o<<"{\"kind\":\""<<(op>=8&&op<=15?"RDP triangle":op==0x36?"RDP fill":op==0x24||op==0x25?"RDP texture rectangle":"RDP command")<<"\",\"hasSource\":"<<((op==0x24||op==0x25||(op>=8&&op<=15&&(op&2)))?"true":"false")<<",\"opcode\":"<<op<<",\"words\":[";for(int64_t i=0;i<w.n;i++){if(i)o<<',';o<<'"'<<std::hex<<w[i]<<'"';}o<<std::dec<<"],\"colorBuffer\":"<<m->rdp.Color.Addr<<",\"textureAddress\":"<<m->rdp.Texture.Addr<<",\"depthBuffer\":"<<m->rdp.Mask<<",\"otherModes\":\""<<std::hex<<m->rdp.OtherModes<<"\",\"combine\":\""<<m->rdp.Combine<<std::dec<<"\",\"rspPC\":"<<(m->RSP?m->RSP->PC:0)<<",\"rspSteps\":"<<m->rspSteps<<"}";auto resource=t.resource(m->rdp.TMem.data(),m->rdp.TMem.size());t.event(m->CPU->Steps,uint32_t(m->CPU->PC),o.str(),resource);};
 machine->OnPixel=[](uint32_t x,uint32_t y,n64_PixelEvent e){auto&t=rrcapture::trace;auto&r=machine->rdp;uint32_t a=n64_rdp_pixelAddr(&r,x,y);int n=r.Color.Size==2?2:4;if(uint64_t(a)+n>uint64_t(machine->RDRAM.n))return;t.texel=e.TexR|(e.TexG<<8)|(e.TexB<<16)|(e.TexA<<24);t.u=e.TexS;t.v=e.TexT;t.record(a,rrcapture::little(machine->RDRAM.p+a,n),n,machine->CPU->Steps,uint32_t(machine->CPU->PC),t.current,(e.Drawn?1:0)|(e.ZReject?2:0)|(e.AlphaReject?4:0),e.Z);};return 1;}catch(...){return 0;}}
int rr_capture_end(){machine->OnWrite={};machine->OnRDPCmd={};machine->OnPixel={};machine->WatchLo=machine->WatchHi=0;capOrigin=n64_Machine_Origin(machine);capWidth=n64_Machine_Width(machine);if(!capWidth||capWidth>1024)capWidth=320;capHeight=height(machine);capType=n64_Machine_PixelType(machine);rrcapture::trace.end(machine->RDRAM.p,machine->RDRAM.n);return 1;}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){if(!capOrigin||(capType!=2&&capType!=3))return "{\"blank\":true,\"contributors\":[],\"complete\":true}";if(x<0||y<0||uint32_t(x)>=capWidth||uint32_t(y)>=capHeight)return "{\"error\":\"Outside captured display\"}";int n=capType==2?2:4;reply=rrcapture::trace.pixel(capOrigin+(y*capWidth+x)*n,n);return reply.c_str();}
}

extern "C" const char*rr_source(uint32_t address,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(address,size,before,expected);return reply.c_str();}
extern "C" const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
