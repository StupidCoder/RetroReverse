// Owned VIC inspection. Fetches and pixels come from the live renderer; frozen
// previews run an isolated VIC against captured RAM. No vendored renderer is used.
namespace render {
using namespace bridge;
constexpr unsigned W=392,H=272,Cells=65536+1024+4096+2,MaxRefs=524288,MaxWrites=65536;
constexpr std::array<uint32_t,16> Palette={0xff000000,0xffffffff,0xff383381,0xffc8ce75,0xff973c8e,0xff4dac56,0xff9b2c2e,0xff71f1ed,0xff29508e,0xff003855,0xff716cc4,0xff4a4a4a,0xff7b7b7b,0xff9fffa9,0xffeb6d70,0xffb2b2b2};
struct Write {uint64_t cycle=0;uint16_t pc=0,address=0;uint8_t space=0,value=0,oldValue=0;std::array<uint8_t,3> code{};bool interrupt=false;};
struct Ref {uint64_t cycle=0;uint16_t address=0;uint8_t space=0,value=0;Write writer;uint32_t pointer=0,bank=0,bankDDR=0,memoryPointers=0;};
struct Control {std::array<uint32_t,49> refs{};std::array<uint8_t,47> regs{};uint16_t bank=0;};
struct Pixel {uint64_t cycle=0;uint32_t screen=0,colorRam=0,data=0,control=0;std::array<uint32_t,8> sprites{};uint8_t color=0,hits=0;bool border=false,foreground=false;};
struct Line {std::array<uint8_t,65536> ram{};std::array<uint8_t,1024> color{};std::array<uint8_t,47> regs{};uint16_t bank=0;uint64_t cycle=0;unsigned before=0;};
std::array<Write,Cells> history{},initial{};
std::array<Line,H> lines;
std::array<Pixel,W*H> captured;
std::array<std::array<uint32_t,W*H>,3> pictures;
std::array<uint32_t,W*H> initialPixels{},replayPixels{};
std::array<uint8_t,4096> chars{};
std::array<uint8_t,65536> used{};
std::vector<Ref> refs(1);std::vector<Control> controls(1);std::vector<Write> writes;
using FetchKey=std::tuple<uint64_t,uint16_t,uint8_t,uint8_t>;
std::map<FetchKey,uint32_t> fetched;
std::array<uint32_t,8> pointers{};

uint64_t start=0,end=0,controlVersion=0,seenVersion=0;uint32_t currentControl=0,overflow=0,captureGeneration=0,replayCursor=0;
bool active=false;int selected=-1;std::string json;
uint16_t canonical(unsigned sp,uint16_t a){if(sp==1)return 0xd800|(a&1023);if(sp==2){if(a>=0xd000&&a<0xd400)return 0xd000|(a&63);if(a>=0xdc00&&a<0xde00)return (a&0xff00)|(a&15);}return a;}
unsigned key(unsigned sp,uint16_t a){a=canonical(sp,a);return sp==0?a:sp==1?65536+(a&1023):sp==2?(a<2?Cells-2+a:65536+1024+(a&4095)):Cells;}
uint32_t add(Ref r){if(refs.size()>=MaxRefs){++overflow;return 0;}refs.push_back(r);return unsigned(refs.size()-1);}
Ref source(unsigned sp,uint16_t a,uint8_t value,uint64_t cycle){Ref r{cycle,canonical(sp,a),uint8_t(sp),value,{}};const auto k=key(sp,a);if(k<Cells){const auto& w=history[k];if(w.cycle<cycle&&((sp==1?w.value&15:w.value)==value))r.writer=w;}return r;}
uint32_t fetchRef(const VicFetch& f,bool color=false){if(!f.cycle)return 0;const unsigned sp=color?1:f.rom?3:0;const uint16_t a=color?uint16_t(0xd800|(f.address&1023)):f.address;const FetchKey k={f.cycle,a,uint8_t(sp),f.phase};auto i=fetched.find(k);if(i!=fetched.end())return i->second;auto r=source(sp,a,color?f.color:f.value,f.cycle);if(f.kind==VicAccess::Sprite)r.pointer=pointers[f.slot];if(active&&f.cycle>start){r.bank=add(source(2,0xdd00,board.state.cia2.pra,f.cycle));r.bankDDR=add(source(2,0xdd02,board.state.cia2.ddra,f.cycle));r.memoryPointers=add(source(2,0xd018,board.state.vic.regs[0x18],f.cycle));}const auto id=add(r);if(fetched.size()<MaxRefs)fetched.emplace(k,id);else ++overflow;return id;}
bool valid(){return end&&captureGeneration==generation&&!active;}
bool complete(){if(!valid()||overflow)return false;return std::all_of(captured.begin(),captured.end(),[](const auto& p){return p.cycle!=0;});}
struct Recorder:VicObserver {
 void fetch(const VicFetch& f)override{if(f.kind==VicAccess::Pointer)pointers[f.slot]=fetchRef(f);else if(f.kind==VicAccess::Matrix){fetchRef(f);fetchRef(f,true);}else if(f.kind==VicAccess::Graphics||f.kind==VicAccess::Sprite)fetchRef(f);if(!f.rom)used[f.address]|=f.kind==VicAccess::Matrix?1:f.kind==VicAccess::Graphics?2:(f.kind==VicAccess::Sprite||f.kind==VicAccess::Pointer)?4:0;}
 void pixel(const Vic& v,const VicPixel& dot)override{
  if(v.raster<16||v.raster>=16+H)return;const unsigned x=(dot.x+8)%504,y=v.raster-16;if(x>=W)return;
  if(x==0){auto& l=lines[y];l.ram=board.state.ram;l.color=board.state.color;std::copy_n(v.regs.begin(),47,l.regs.begin());l.bank=uint16_t((~board.state.cia2.outputA()&3)<<14);l.cycle=v.clocks;l.before=writes.size();}
  if(!currentControl||seenVersion!=controlVersion){
   Control c;std::copy_n(v.regs.begin(),47,c.regs.begin());c.bank=uint16_t((~board.state.cia2.outputA()&3)<<14);
   for(unsigned r=0;r<47;r++)c.refs[r]=add(source(2,0xd000+r,v.regs[r],v.clocks));
   c.refs[47]=add(source(2,0xdd00,board.state.cia2.pra,v.clocks));c.refs[48]=add(source(2,0xdd02,board.state.cia2.ddra,v.clocks));
   if(controls.size()<MaxWrites){controls.push_back(c);currentControl=unsigned(controls.size()-1);}else{++overflow;currentControl=0;}seenVersion=controlVersion;
  }
  auto& p=captured[y*W+x];p={};p.cycle=v.clocks;p.control=currentControl;p.color=dot.color;p.hits=dot.hits;p.border=dot.border;p.foreground=dot.foreground;
  const int gx=int(dot.x)-24-(v.regs[0x16]&7);if(gx>=0&&gx<320&&!v.verticalBorder){const auto& cell=v.cells[unsigned(gx)/8];p.screen=fetchRef(cell.matrix);p.colorRam=fetchRef(cell.matrix,true);p.data=fetchRef(cell.graphics);}
  for(unsigned i=0;i<8;i++)if(dot.hits&(1<<i))p.sprites[i]=fetchRef(v.sprites[i].bytes[dot.positions[i]/8]);
 }
} recorder;
void reset(){board.observer=nullptr;active=false;start=end=0;selected=-1;history.fill({});initial.fill({});fetched.clear();refs.resize(1);controls.resize(1);writes.clear();}
void edited(uint16_t address){history[key(0,address)]={};}
void beforeTick(){}
void afterTick(const CpuState& before,int mapped,uint8_t old){
 const auto& b=board.state.lastBus;if(!b.valid||(!b.write&&!b.ready))return;
 if(before.stage==Stage::Fetch&&b.ready){if(before.interruptPending)++ctx.interruptDepth;else {ctx.instruction={b.value,0,0};if(b.value==0)++ctx.interruptDepth;}}
 if(!b.write){for(unsigned i=1;i<3;i++)if(b.address==uint16_t(ctx.pc+i))ctx.instruction[i]=b.value;}
 if(before.retired!=board.state.cpu.retired&&ctx.instruction[0]==0x40&&ctx.interruptDepth)--ctx.interruptDepth;
 if(!b.write)return;
 const unsigned sp=unsigned(mapped);const auto a=canonical(sp,b.address);const auto k=key(sp,a);if(k>=Cells)return;
 if(sp==2&&a>=0xd020&&a<=0xd02e)old&=15;if(sp==1)old&=15;
 Write w{board.state.cycles,ctx.pc,a,uint8_t(sp),uint8_t(sp==1||(sp==2&&a>=0xd020&&a<=0xd02e)?b.value&15:b.value),old,ctx.instruction,ctx.interruptDepth!=0};
 history[k]=w;if(sp==2)++controlVersion;
 if(active){if(writes.size()<MaxWrites)writes.push_back(w);else ++overflow;}
}
std::string writerJSON(const Write& w){if(!w.cycle)return "null";std::ostringstream o;o<<"{\"cycle\":"<<w.cycle<<",\"pc\":"<<w.pc<<",\"value\":"<<unsigned(w.value)<<",\"oldValue\":"<<unsigned(w.oldValue)<<",\"interrupt\":"<<(w.interrupt?"true":"false")<<",\"code\":["<<unsigned(w.code[0])<<','<<unsigned(w.code[1])<<','<<unsigned(w.code[2])<<"]}";return o.str();}
std::string refJSON(const Ref& r,const std::string& role){return "{\"role\":\""+role+"\",\"address\":"+std::to_string(r.address)+",\"space\":"+std::to_string(r.space)+",\"value\":"+std::to_string(r.value)+",\"fetchCycle\":"+std::to_string(r.cycle)+",\"writer\":"+writerJSON(r.writer)+"}";}
std::string refJSON(unsigned id,const std::string& role){return id&&id<refs.size()?refJSON(refs[id],role):"{\"role\":\""+role+"\",\"missing\":true}";}
std::string pixelJSON(int x,int y){if(!valid()||x<0||x>=int(W)||y<0||y>=int(H))return "{\"error\":\"Capture a display first\"}";const auto& p=captured[y*W+x];if(!p.cycle||!p.control)return "{\"error\":\"Pixel evidence is incomplete\"}";const auto& c=controls[p.control];
 std::ostringstream o;o<<"{\"x\":"<<x<<",\"y\":"<<y<<",\"cycle\":"<<p.cycle<<",\"raster\":"<<y+16<<",\"color\":"<<unsigned(p.color)<<",\"rgba\":"<<Palette[p.color&15]<<",\"border\":"<<(p.border?"true":"false")<<",\"foreground\":"<<(p.foreground?"true":"false")<<",\"spriteMask\":"<<unsigned(p.hits)<<",\"overflow\":"<<overflow<<",\"complete\":"<<(complete()?"true":"false")<<",\"contributors\":[";
 bool comma=false;auto addRef=[&](unsigned id,const std::string& role){if(comma)o<<',';comma=true;o<<refJSON(id,role);};
 if(p.border)addRef(c.refs[0x20],"border color");else{addRef(p.screen,"screen byte");addRef(p.colorRam,"color RAM");addRef(p.data,"graphics byte");if(p.data){const auto& f=refs[p.data];addRef(f.bank,"fetch-time VIC bank control");addRef(f.bankDDR,"fetch-time VIC bank direction");addRef(f.memoryPointers,"fetch-time memory pointers");}for(unsigned i=0;i<8;i++)if(p.hits&(1<<i)){addRef(p.sprites[i],"sprite "+std::to_string(i)+" pixel byte");if(p.sprites[i])addRef(refs[p.sprites[i]].pointer,"sprite "+std::to_string(i)+" pointer");}}
 for(unsigned r=0;r<47;r++)if(r<=0x18||r>=0x1b)addRef(c.refs[r],"VIC control/palette");addRef(c.refs[47],"VIC bank control");addRef(c.refs[48],"VIC bank direction");o<<"]}";return o.str();
}
// Previews use a fresh observer and VIC. Source attribution is reconstructed at
// the selected snapshot, while actual pixels retain their fetch-time versions.
struct Preview:VicObserver {
 const Line& line;std::array<Write,Cells> writers;std::array<VicFetch,8> pointer{};std::string evidence;int panel,qx,qy;
 Preview(const Line& l,int p=-1,int x=-1,int y=-1):line(l),writers(initial),panel(p),qx(x),qy(y){for(unsigned i=0;i<l.before;i++)writers[key(writes[i].space,writes[i].address)]=writes[i];}
 void fetch(const VicFetch& f)override{if(f.kind==VicAccess::Pointer)pointer[f.slot]=f;}
 Ref ref(unsigned sp,uint16_t a,uint8_t value){Ref r{0,canonical(sp,a),uint8_t(sp),value,{}};auto k=key(sp,a);if(k<Cells&&writers[k].value==value)r.writer=writers[k];return r;}
 void pixel(const Vic& v,const VicPixel& dot)override{if(v.frames<1||v.raster<16||v.raster>=16+H)return;unsigned x=(dot.x+8)%504,y=v.raster-16;if(x>=W)return;
  pictures[0][y*W+x]=Palette[dot.background&15];pictures[1][y*W+x]=dot.hits?Palette[dot.spriteColor&15]:0;
  if(int(x)!=qx||int(y)!=qy)return;std::ostringstream o;o<<"{\"panel\":"<<panel<<",\"stateLine\":"<<selected<<",\"x\":"<<x<<",\"y\":"<<y<<",\"preview\":true,\"complete\":"<<(complete()?"true":"false")<<",\"contributors\":[";bool comma=false;
  auto addRef=[&](Ref r,const std::string& role){if(comma)o<<',';comma=true;o<<refJSON(r,role);};auto addFetch=[&](const VicFetch& f,const std::string& role){addRef(ref(f.rom?3:0,f.address,f.value),role);};
  if(panel==0){if(dot.border)addRef(ref(2,0xd020,v.regs[0x20]),"border color");else {int gx=int(dot.x)-24-(v.regs[0x16]&7);if(gx>=0&&gx<320){const auto& c=v.cells[unsigned(gx)/8];addFetch(c.graphics,"graphics byte");addFetch(c.matrix,"screen byte");addRef(ref(1,0xd800|(c.matrix.address&1023),c.matrix.color),"color RAM");}}}
  else for(unsigned i=0;i<8;i++)if(dot.hits&(1<<i)){addFetch(v.sprites[i].bytes[dot.positions[i]/8],"sprite "+std::to_string(i)+" pixel byte");addFetch(pointer[i],"sprite "+std::to_string(i)+" pointer");}
  for(unsigned r=0;r<47;r++)if(r<=0x18||r>=0x1b)addRef(ref(2,0xd000+r,line.regs[r]),"frozen VIC control/palette");o<<"]}";evidence=o.str();
 }
};
std::string preview(int panel=-1,int x=-1,int y=-1){auto observer=std::make_unique<Preview>(lines[selected],panel,x,y);auto v=std::make_unique<Vic>();std::copy(lines[selected].regs.begin(),lines[selected].regs.end(),v->regs.begin());for(unsigned i=0;i<2*19656+1;i++)v->tick(lines[selected].ram,chars,lines[selected].color,lines[selected].bank,observer.get());return observer->evidence;}
unsigned relevant(const Write& w){return w.space==2?((w.address<0xd02f||w.address==0xdd00||w.address==0xdd02)?16:0):w.space==1?32:used[w.address];}
unsigned previous(int line){return line?lines[line-1].before:0;}
std::string info(){std::ostringstream o;o<<"{\"available\":"<<(valid()?"true":"false")<<",\"complete\":"<<(complete()?"true":"false")<<",\"width\":392,\"height\":272,\"lines\":[";bool comma=false;if(valid())for(unsigned y=0;y<H;y++)if(lines[y].cycle){unsigned changes=0;for(unsigned i=previous(y);i<lines[y].before;i++)changes+=relevant(writes[i])&&writes[i].oldValue!=writes[i].value;if(comma)o<<',';comma=true;o<<"{\"line\":"<<y<<",\"raster\":"<<y+16<<",\"changes\":"<<changes<<"}";}o<<"]}";return o.str();}
std::string seek(int line){if(!valid()||line<0||line>=int(H)||!lines[line].cycle)return "{\"error\":\"Select a recorded line\"}";selected=line;preview();pictures[2].fill(0);for(int y=0;y<=line;y++)for(unsigned x=0;x<W;x++){const auto& p=captured[y*W+x];if(p.cycle)pictures[2][y*W+x]=Palette[p.color&15];}const auto& l=lines[line];std::ostringstream o;
 o<<"{\"line\":"<<line<<",\"raster\":"<<line+16<<",\"clock\":"<<l.cycle<<",\"complete\":"<<(complete()?"true":"false")<<",\"vicBank\":"<<l.bank<<",\"screenBase\":"<<l.bank+((l.regs[0x18]&0xf0)<<6)<<",\"charsetBase\":"<<l.bank+((l.regs[0x18]&14)<<10)<<",\"bitmapBase\":"<<l.bank+((l.regs[0x18]&8)<<10)<<",\"mode\":"<<(((l.regs[0x11]>>4)&6)|((l.regs[0x16]>>4)&1))<<",\"registers\":[";
 for(unsigned i=0;i<47;i++){if(i)o<<',';o<<"{\"name\":\"$"<<std::hex<<0xd000+i<<std::dec<<"\",\"value\":"<<unsigned(l.regs[i])<<"}";}o<<"],\"changes\":[";unsigned count=0,graphics=0,maps=0,objects=0,colors=0;
 for(unsigned i=previous(line);i<l.before;i++){const auto& w=writes[i];const auto k=relevant(w);if(!k)continue;graphics+=!!(k&2);maps+=!!(k&1);objects+=!!(k&4);colors+=!!(k&32);if(count++>=64)continue;if(count>1)o<<',';o<<"{\"name\":\"$"<<std::hex<<w.address<<std::dec<<"\",\"address\":"<<w.address<<",\"space\":"<<unsigned(w.space)<<",\"before\":"<<unsigned(w.oldValue)<<",\"after\":"<<unsigned(w.value)<<",\"pc\":"<<w.pc<<",\"writer\":"<<writerJSON(w)<<"}";}
 o<<"],\"changeCount\":"<<count<<",\"tileWrites\":"<<graphics<<",\"mapWrites\":"<<maps<<",\"objectWrites\":"<<objects<<",\"colorWrites\":"<<colors<<"}";return o.str();}
}
extern "C" {
int rr_capture_begin(){using namespace render;if(!initialized)return 0;active=true;start=board.state.cycles;end=0;overflow=currentControl=0;captureGeneration=generation;selected=-1;initial=history;chars=board.chars;for(auto& l:lines)l.cycle=0;captured.fill({});refs.resize(1);controls.resize(1);fetched.clear();pointers.fill(0);writes.clear();used.fill(0);std::copy_n(rr_frame(),W*H,initialPixels.begin());board.observer=&recorder;return 1;}
int rr_capture_end(){render::active=false;board.observer=nullptr;render::end=board.state.cycles;return 1;}
const char* rr_capture_info(){using namespace render;json="{\"events\":"+std::to_string(refs.size())+",\"writes\":"+std::to_string(writes.size())+",\"overflow\":"+std::to_string(overflow)+",\"bytes\":"+std::to_string(sizeof(captured)+refs.size()*sizeof(Ref)+controls.size()*sizeof(Control))+",\"rasterBytes\":"+std::to_string(sizeof(lines)+sizeof(initial)+writes.size()*sizeof(Write))+"}";return json.c_str();}
const char* rr_pixel(int x,int y){render::json=render::pixelJSON(x,y);return render::json.c_str();}
const char* rr_raster_info(){render::json=render::info();return render::json.c_str();}
const char* rr_raster_seek(int line){render::json=render::seek(line);return render::json.c_str();}
uint32_t* rr_raster_frame(int panel){return panel>=0&&panel<3?render::pictures[panel].data():nullptr;}
const char* rr_raster_pixel(int panel,int x,int y){using namespace render;if(!valid()||selected<0||panel<0||panel>2||x<0||x>=int(W)||y<0||y>=int(H))json="{\"error\":\"Select a recorded pixel\"}";else if(panel==2){if(y>selected)json="{\"error\":\"This line has not been drawn at this point\"}";else{json=pixelJSON(x,y);if(json.find("error")==std::string::npos){json.pop_back();json+=",\"panel\":2,\"stateLine\":"+std::to_string(y)+",\"preview\":false}";}}}else json=preview(panel,x,y);return json.c_str();}
const char* rr_replay_info(){using namespace render;json="{\"count\":272,\"cursor\":"+std::to_string(replayCursor)+",\"cacheBytes\":0,\"complete\":"+(complete()?"true":"false")+",\"command\":{\"kind\":\"VIC visible raster line\"}}";return json.c_str();}
const char* rr_replay_begin(){render::replayCursor=0;render::replayPixels=render::initialPixels;return rr_replay_info();}
int rr_replay_seek(unsigned count){using namespace render;count=std::min(count,H);replayPixels=initialPixels;for(unsigned y=0;y<count;y++)for(unsigned x=0;x<W;x++){const auto& p=captured[y*W+x];if(p.cycle)replayPixels[y*W+x]=Palette[p.color&15];}replayCursor=count;return 1;}
uint32_t* rr_replay_frame(){return render::replayPixels.data();}
uint32_t rr_replay_for_write(uint32_t y){return std::min(y+1,272u);}
int rr_tileset_size(){return render::valid()&&render::selected>=0?4240:0;}
uint8_t* rr_tileset_data(){using namespace render;static std::array<uint8_t,4240>b{};if(!rr_tileset_size())return nullptr;const auto& l=lines[selected];std::copy(l.regs.begin(),l.regs.end(),b.begin());unsigned charset=l.bank+((l.regs[0x18]&14)<<10),screen=l.bank+((l.regs[0x18]&0xf0)<<6);b[64]=charset;b[65]=charset>>8;b[66]=screen;b[67]=screen>>8;for(unsigned i=0;i<16;i++)for(unsigned k=0;k<4;k++)b[80+i*4+k]=Palette[i]>>(k*8);std::copy(l.color.begin(),l.color.end(),b.begin()+144);auto get=[&](unsigned a){return (a&0x7000)==0x1000?chars[a&4095]:l.ram[a];};for(unsigned i=0;i<1024;i++)b[1168+i]=get(screen+i);for(unsigned i=0;i<2048;i++)b[2192+i]=get(charset+i);return b.data();}
const char* rr_memory(int sp,int address){using namespace render;if(sp<0||sp>3||address<0||address>65535){json="{\"error\":\"Invalid memory selection\"}";return json.c_str();}const auto k=key(sp,address);json="{\"address\":"+std::to_string(address)+",\"space\":"+std::to_string(sp)+",\"writer\":"+writerJSON(k<Cells?history[k]:Write{})+"}";return json.c_str();}
}
