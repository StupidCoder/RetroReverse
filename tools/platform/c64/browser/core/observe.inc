// Generic observation and bounded provenance. Included after the board and Context.
// Physical spaces: 0 RAM, 1 color RAM, 2 I/O, 3 ROM (immutable, no guest writer).
namespace observation {
constexpr unsigned CELLS=65536+1024+4098, INSTRUCTIONS=2048;
struct Write {
 uint64_t cycle=0,instructionCycle=0,readCycle=0;
 uint16_t pc=0,address=0,readAddress=0;
 uint8_t space=0,oldValue=0,value=0,bank=0,a=0,x=0,y=0,s=0,p=0,code[3]={},readValue=0,readSpace=0,interrupt=0;
};
struct History {
 Write last[CELLS];
 uint64_t instructionCycles[INSTRUCTIONS]={},instructionCycle=0,readCycle=0;
 uint32_t instructionCount=0;
 uint16_t readAddress=0;
 uint8_t code[3]={},readValue=0,readSpace=0;
};
static History history;
static Write bankAtTick,bankDDRAtTick;
static std::unique_ptr<History> saved[16];
bool capturing=false;
static bool traceReads=false;
static std::string json;
static uint16_t canonical(int space,uint16_t a){
 if(space==1)return 0xd800|(a&1023);
 if(space==2){if(a>=0xd000&&a<0xd400)return 0xd000|(a&63);if(a>=0xd400&&a<0xd800)return 0xd400|(a&31);if(a>=0xdc00&&a<0xde00)return (a&0xff00)|(a&15);}
 return a;
}
static unsigned key(int space,uint16_t a){
 a=canonical(space,a);
 if(space==0)return a;
 if(space==1)return 65536+(a&1023);
 if(space==2)return 65536+1024+(a<2?4096+a:a&4095);
 return CELLS;
}
static void beginTick(){bankAtTick=history.last[key(2,0xdd00)];bankDDRAtTick=history.last[key(2,0xdd02)];}
static int cpuSpace(uint16_t a){
 if(a<2)return 2;
 if(a>=0xd000&&a<0xe000&&machine.io_mapped)return a>=0xd800&&a<0xdc00?1:2;
 const auto b=machine.cpu_port;
 if((a>=0xa000&&a<0xc000&&(b&3)==3)||(a>=0xe000&&(b&2))||(a>=0xd000&&a<0xe000&&(b&3)&&!(b&4)))return 3;
 return 0;
}
static uint8_t valueAt(int space,uint16_t a){
 if(space==0)return machine.ram[a];
 if(space==1)return machine.color_ram[a&1023]&15;
 if(space==2){auto k=key(space,a);return k<CELLS?history.last[k].value:0;}
 return mem_rd(&machine.mem_cpu,a);
}
static void reset(){
 // Construct in place: a temporary History consumes several MiB of WASM stack.
 std::destroy_at(&history);std::construct_at(&history);
 for(auto& p:saved)p.reset();capturing=false;traceReads=false;raster::reset();
}
static void instruction(){
 // SYNC marks the opcode which is about to execute; bytes are latched now,
 // before self-modifying writes can alter the displayed disassembly.
 history.instructionCycle=ctx.cycles;
 history.instructionCycles[history.instructionCount++%INSTRUCTIONS]=ctx.cycles;
 history.code[0]=M6502_GET_DATA(machine.pins);
 history.code[1]=mem_rd(&machine.mem_cpu,uint16_t(ctx.pc+1));
 history.code[2]=mem_rd(&machine.mem_cpu,uint16_t(ctx.pc+2));
 history.readCycle=0;
}
static void read(uint16_t a,uint8_t v){history.readCycle=ctx.cycles;history.readAddress=a;history.readValue=v;history.readSpace=cpuSpace(a);}
void write(int space,uint16_t address,uint8_t value){
 const unsigned k=key(space,address);if(k>=CELLS)return;
 Write w;w.cycle=ctx.cycles+1;w.instructionCycle=history.instructionCycle;w.pc=ctx.pc;w.address=address;w.space=space;
 w.interrupt=machine.cpu.brk_flags&(M6502_BRK_IRQ|M6502_BRK_NMI);
 w.oldValue=valueAt(space,address);w.value=value;w.bank=machine.cpu_port;
 w.a=machine.cpu.A;w.x=machine.cpu.X;w.y=machine.cpu.Y;w.s=machine.cpu.S;w.p=machine.cpu.P;
 memcpy(w.code,history.code,3);w.readCycle=history.readCycle;w.readAddress=history.readAddress;w.readValue=history.readValue;w.readSpace=history.readSpace;
 history.last[k]=w;
 if(capturing)raster::write(w);
}
static std::string writerJSON(const Write& w){
 if(!w.cycle)return "null";
 char b[900];snprintf(b,sizeof b,"{\"cycle\":%llu,\"instructionCycle\":%llu,\"pc\":%u,\"address\":%u,\"space\":%u,\"oldValue\":%u,\"oldMeaning\":\"%s\",\"value\":%u,\"bank\":%u,\"a\":%u,\"x\":%u,\"y\":%u,\"s\":%u,\"p\":%u,\"code\":[%u,%u,%u],\"nearbyRead\":{\"cycle\":%llu,\"address\":%u,\"value\":%u,\"space\":%u}}",
 (unsigned long long)w.cycle,(unsigned long long)w.instructionCycle,w.pc,w.address,w.space,w.oldValue,w.space==2?"previous CPU write; device may have changed internally":"previous stored byte",w.value,w.bank,w.a,w.x,w.y,w.s,w.p,w.code[0],w.code[1],w.code[2],(unsigned long long)w.readCycle,w.readAddress,w.readValue,w.readSpace);return std::string(b).substr(0,strlen(b)-1)+",\"interrupt\":"+std::to_string(w.interrupt)+"}";
}
struct Ref {Write writer;uint64_t cycle=0;uint16_t address=0;uint8_t space=0,value=0;uint32_t pointer=0;};
struct Graphics {uint32_t screen=0,color=0,data=0,pointers=0,bank=0,bankDDR=0;};
struct Control {uint32_t refs[47]={},bank=0,bankDDR=0;uint16_t vicBank=0;uint8_t regs[47]={};};
struct Pixel {uint64_t cycle=0;uint32_t graphics=0,control=0,sprites[8]={},pointers[8]={};uint16_t raster=0;uint8_t color=0,spriteMask=0,foreground=0,border=0;};
constexpr unsigned REF_CAP=100000, GRAPH_CAP=32768, CONTROL_CAP=8192;
static Ref refs[REF_CAP];static Graphics graphs[GRAPH_CAP];static Control controls[CONTROL_CAP];
static Pixel captured[392*272];
static uint32_t replayInitial[392*272],replayPixels[392*272];
static unsigned replayCursor=0;
static uint32_t refCount=1,graphCount=1,controlCount=1,overflow=0;
static uint32_t screenRefs[64],colorRefs[64],pointerRefs[8],spriteBits[8][32],spriteOut[8][2],pendingGraph=0,activeGraph=0,currentControl=0;
static uint64_t controlCycle=0,captureStart=0,captureEnd=0;
static uint32_t captureGeneration=0;
static uint32_t regCache[4098]={};
static unsigned vicSpace(uint16_t a){return (a&0x7000)==0x1000?3:0;}
static uint32_t ref(int space,uint16_t a,uint8_t value,bool cache=false,const Write* source=nullptr){
 unsigned k=key(space,a);const Write* w=source?source:k<CELLS?&history.last[k]:nullptr;
 unsigned ck=a<2?4096+a:a&4095;
 if(cache&&regCache[ck]){const auto& r=refs[regCache[ck]];if(r.value==value&&r.writer.cycle==(w?w->cycle:0))return regCache[ck];}
 if(refCount>=REF_CAP){overflow++;return 0;}
 auto& r=refs[refCount];r=Ref{};r.cycle=ctx.cycles+1;r.address=a;r.space=space;r.value=value;if(w)r.writer=*w;
 if(cache)regCache[ck]=refCount;return refCount++;
}
static uint32_t regRef(uint16_t a,uint8_t value){return ref(2,a,value,true);}
static uint32_t bankRef(){return ref(2,0xdd00,bankAtTick.value,true,&bankAtTick);}
static uint32_t bankDDRRef(){return ref(2,0xdd02,bankDDRAtTick.value,true,&bankDDRAtTick);}
void fetch(int kind,int slot,uint16_t a,uint16_t value){
 a|=machine.vic_bank_select;
 raster::used(kind,a);
 if(kind==0){screenRefs[slot]=ref(vicSpace(a),a,value);colorRefs[slot]=ref(1,0xd800|(a&1023),value>>8);}
 else if(kind==1){
  if(graphCount>=GRAPH_CAP){overflow++;pendingGraph=0;return;}
  auto& g=graphs[graphCount];g={screenRefs[slot],colorRefs[slot],ref(vicSpace(a),a,value),regRef(0xd018,machine.vic.reg.mem_ptrs),bankRef(),bankDDRRef()};pendingGraph=graphCount++;
 }else if(kind==2)pointerRefs[slot]=ref(vicSpace(a),a,value);
 else if(kind==3){
  const auto id=ref(vicSpace(a),a,value);if(id)refs[id].pointer=pointerRefs[slot];
  for(int bit=31;bit>=8;bit--)spriteBits[slot][bit]=spriteBits[slot][bit-8];
  for(int bit=8;bit<16;bit++)spriteBits[slot][bit]=id;
  for(int bit=0;bit<8;bit++)spriteBits[slot][bit]=0;
 }
}
void idle(){pendingGraph=0;}
void load(bool enabled){activeGraph=enabled?pendingGraph:0;}
void spriteOutput(int index,bool pair){spriteOut[index][pair]=spriteBits[index][31];}
void spriteShift(int index){for(int b=31;b>0;b--)spriteBits[index][b]=spriteBits[index][b-1];spriteBits[index][0]=0;}
void pixel(void* ptr,uint8_t* dst,uint16_t graphics,uint16_t sprites,bool border){
 auto& v=*static_cast<m6569_t*>(ptr);ptrdiff_t offset=dst-v.crt.fb;
 const unsigned y=unsigned(offset)/M6569_FRAMEBUFFER_WIDTH,x=unsigned(offset)%M6569_FRAMEBUFFER_WIDTH;
 if(x>=392||y>=272)return;
 if(x==0)raster::recordLine(y,ptr);
 if(controlCycle!=ctx.cycles+1){
  controlCycle=ctx.cycles+1;
  bool changed=!currentControl||memcmp(controls[currentControl].regs,v.reg.regs,47)||controls[currentControl].vicBank!=machine.vic_bank_select;
  if(!changed){const auto& old=controls[currentControl];changed=refs[old.bank].writer.cycle!=bankAtTick.cycle||refs[old.bankDDR].writer.cycle!=bankDDRAtTick.cycle;for(int i=0;i<47&&!changed;i++)changed=refs[old.refs[i]].writer.cycle!=history.last[key(2,0xd000+i)].cycle;}
  if(changed){
   if(controlCount>=CONTROL_CAP){overflow++;currentControl=0;}else{
    currentControl=controlCount++;auto& c=controls[currentControl];memcpy(c.regs,v.reg.regs,47);c.vicBank=machine.vic_bank_select;c.bank=bankRef();c.bankDDR=bankDDRRef();
    for(int i=0;i<47;i++)c.refs[i]=regRef(0xd000+i,c.regs[i]);
   }
  }
 }
 auto& p=captured[y*392+x];p=Pixel{};p.cycle=ctx.cycles+1;p.graphics=activeGraph;p.control=currentControl;p.raster=v.rs.v_count;p.color=*dst;p.spriteMask=sprites>>8;p.foreground=graphics>>8;p.border=border?(v.brd.main?2:1):0;
 for(int i=0;i<8;i++)if(p.spriteMask&(1<<i)){p.sprites[i]=spriteOut[i][(v.reg.mmc>>i)&1];p.pointers[i]=refs[p.sprites[i]].pointer;}
}
static std::string refJSON(unsigned id,const char* role){
 if(!id||id>=refCount)return "{\"role\":\""+std::string(role)+"\",\"missing\":true}";
 const auto& r=refs[id];char b[300];snprintf(b,sizeof b,"{\"role\":\"%s\",\"reference\":%u,\"address\":%u,\"space\":%u,\"value\":%u,\"fetchCycle\":%llu,\"writer\":",role,id,r.address,r.space,r.value,(unsigned long long)r.cycle);return std::string(b)+writerJSON(r.writer)+"}";
}
}
extern "C" {
void rr_trace_reads(int enabled){observation::traceReads=enabled!=0;}
double rr_cycle(){return double(ctx.cycles);}
double rr_previous(){using namespace observation;for(unsigned n=0;n<std::min(history.instructionCount,INSTRUCTIONS);n++){auto c=history.instructionCycles[(history.instructionCount-1-n)%INSTRUCTIONS];if(c<ctx.cycles)return double(c);}return -1;}
const char* rr_memory(int space,int address){
 using namespace observation;if(space<0||space>3||address<0||address>65535){json="{\"error\":\"Invalid memory selection\"}";return json.c_str();}
 if(space==3&&cpuSpace(address)!=3){json="{\"error\":\"The CPU does not currently map ROM at that address\"}";return json.c_str();}
 auto k=key(space,address);json="{\"address\":"+std::to_string(address)+",\"space\":"+std::to_string(space)+",\"value\":"+std::to_string(valueAt(space,address))+",\"writer\":"+writerJSON(k<CELLS?history.last[k]:Write{})+",\"page\":[";
 for(int i=0;i<16;i++){if(i)json+=",";json+=std::to_string(valueAt(space,(address&0xfff0)+i));}json+="]}";return json.c_str();
}
void rr_capture_begin(){
 using namespace observation;for(int y=0;y<272;y++)for(int x=0;x<392;x++)replayInitial[y*392+x]=m6569_color(machine.fb[y*M6569_FRAMEBUFFER_WIDTH+x]&15);capturing=true;refCount=graphCount=controlCount=1;overflow=0;pendingGraph=activeGraph=currentControl=0;controlCycle=0;captureStart=ctx.cycles;captureEnd=0;captureGeneration=ctx.generation;
 memset(captured,0,sizeof captured);memset(screenRefs,0,sizeof screenRefs);memset(colorRefs,0,sizeof colorRefs);memset(pointerRefs,0,sizeof pointerRefs);memset(spriteBits,0,sizeof spriteBits);memset(spriteOut,0,sizeof spriteOut);memset(regCache,0,sizeof regCache);
 raster::begin();
}
void rr_capture_end(){observation::capturing=false;observation::captureEnd=ctx.cycles;}
const char* rr_pixel(int x,int y){
 using namespace observation;
 if(x<0||x>=392||y<0||y>=272||!captureEnd||captureGeneration!=ctx.generation){json="{\"error\":\"Capture a frame before inspecting a pixel\"}";return json.c_str();}
 const auto& p=captured[y*392+x];const auto& c=controls[p.control];
 if(!p.control||!p.cycle){json="{\"error\":\"Pixel evidence was not captured. Capture a complete frame.\"}";return json.c_str();}
 json="{\"x\":"+std::to_string(x)+",\"y\":"+std::to_string(y)+",\"cycle\":"+std::to_string(p.cycle)+",\"raster\":"+std::to_string(p.raster)+",\"color\":"+std::to_string(p.color)+",\"rgba\":"+std::to_string(m6569_color(p.color&15))+",\"mode\":"+std::to_string(((c.regs[0x11]>>4)&6)|((c.regs[0x16]>>4)&1))+",\"border\":"+(p.border?"true":"false")+",\"spriteMask\":"+std::to_string(p.spriteMask)+",\"foreground\":"+(p.foreground?"true":"false")+",\"vicBank\":"+std::to_string(c.vicBank)+",\"captureStart\":"+std::to_string(captureStart)+",\"captureEnd\":"+std::to_string(captureEnd)+",\"overflow\":"+std::to_string(overflow)+",\"contributors\":[";
 bool comma=false;auto add=[&](unsigned r,const char* role){if(comma)json+=",";comma=true;json+=refJSON(r,role);};
 if(p.border)add(c.refs[p.border==2?0x20:0x21],"border/background color");
 else{
  if(p.graphics){const auto& g=graphs[p.graphics];add(g.screen,"screen byte");add(g.color,"color RAM");add(g.data,"graphics byte");add(g.pointers,"fetch-time memory pointers");add(g.bank,"fetch-time VIC bank control");add(g.bankDDR,"fetch-time VIC bank direction");}
  else add(0,"graphics shifter: idle or unavailable");
  for(int i=0;i<8;i++)if(p.spriteMask&(1<<i)){std::string role="sprite "+std::to_string(i)+" pixel byte";add(p.sprites[i],role.c_str());role="sprite "+std::to_string(i)+" pointer";add(p.pointers[i],role.c_str());add(c.refs[0x27+i],"sprite color");}
  for(int i=0x21;i<=0x26;i++)add(c.refs[i],"palette/multicolor register");
  for(int i:{0x15,0x17,0x1b,0x1c,0x1d})add(c.refs[i],"sprite enable/expansion/priority");
 }
 for(int i:{0x11,0x16,0x18})add(c.refs[i],"raster-time control");add(c.bank,"raster-time VIC bank control");add(c.bankDDR,"raster-time VIC bank direction");json+="]}";return json.c_str();
}
}

extern "C" const char*rr_replay_info(){using namespace observation;json="{\"count\":272,\"cursor\":"+std::to_string(replayCursor)+",\"cacheBytes\":0,\"complete\":"+(overflow?"false":"true")+",\"command\":{\"kind\":\"VIC visible raster line\"}}";return json.c_str();}
extern "C" const char*rr_replay_begin(){using namespace observation;replayCursor=0;memcpy(replayPixels,replayInitial,sizeof replayPixels);return rr_replay_info();}
extern "C" int rr_replay_seek(unsigned step){using namespace observation;step=std::min(step,272u);memcpy(replayPixels,replayInitial,sizeof replayPixels);for(unsigned y=0;y<step;y++)for(unsigned x=0;x<392;x++)replayPixels[y*392+x]=m6569_color(captured[y*392+x].color&15);replayCursor=step;return 1;}
extern "C" uint32_t*rr_replay_frame(){return observation::replayPixels;}
extern "C" uint32_t rr_replay_for_write(uint32_t y){return std::min(y+1,272u);}
