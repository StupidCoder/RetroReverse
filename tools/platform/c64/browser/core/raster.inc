// Frozen-state layer previews use a separate VIC, never the paused board.
namespace raster {
constexpr int W=392,H=272;
struct Line {
 std::array<uint8_t,65536> ram{};
 std::array<uint8_t,1024> color{};
 std::array<uint8_t,47> regs{};
 uint64_t cycle=0;uint32_t before=0;
 uint16_t bank=0,raster=0;bool valid=false;
};
static std::array<Line,H> lines;
static std::array<uint8_t,4096> charROM;
static std::array<uint8_t,65536> usedMemory;
static std::unique_ptr<std::array<observation::Write,observation::CELLS>> initialWriters;
static std::vector<observation::Write> writes;
static std::array<uint32_t,observation::CELLS> lastWrite;
static std::array<std::array<uint32_t,W*H>,3> pictures;
static std::array<uint8_t,M6569_FRAMEBUFFER_SIZE_BYTES> scratch;
static int selected=-1;static bool journalOverflow=false;
bool previewing=false;
void reset(){selected=-1;previewing=false;for(auto&l:lines)l.valid=false;writes.clear();initialWriters.reset();}
void begin(){
 selected=-1;for(auto&l:lines)l.valid=false;writes.clear();journalOverflow=false;usedMemory.fill(0);
 if(!initialWriters)initialWriters=std::make_unique<std::array<observation::Write,observation::CELLS>>();
 std::copy_n(observation::history.last,observation::CELLS,initialWriters->begin());
 std::copy_n(machine.rom_char,4096,charROM.begin());
}
void write(const observation::Write&w){if(writes.size()<65536)writes.push_back(w);else journalOverflow=true;}
void used(int kind,uint16_t a){if(observation::vicSpace(a)==0)usedMemory[a]|=uint8_t(1<<kind);}
void recordLine(unsigned y,void*ptr){
 auto&v=*static_cast<m6569_t*>(ptr);auto&l=lines[y];
 std::copy_n(machine.ram,65536,l.ram.begin());std::copy_n(machine.color_ram,1024,l.color.begin());
 std::copy_n(v.reg.regs,47,l.regs.begin());l.bank=machine.vic_bank_select;l.raster=v.rs.v_count;
 l.cycle=ctx.cycles+1;l.before=writes.size();l.valid=true;
}
static bool valid(){return observation::captureEnd&&observation::captureGeneration==ctx.generation&&initialWriters&&!observation::capturing;}
static bool complete(){if(!valid()||observation::overflow||journalOverflow)return false;for(auto&l:lines)if(!l.valid)return false;return true;}
static bool videoRegister(uint16_t a){return (a>=0xd000&&a<=0xd018)||(a>=0xd01b&&a<=0xd02e)||a==0xdd00||a==0xdd02;}
static unsigned relevant(const observation::Write&w){
 auto a=observation::canonical(w.space,w.address);
 if(w.space==2)return videoRegister(a)?16:0;
 if(w.space==1)return 32;
 return w.space==0?usedMemory[a]:0;
}
static unsigned previous(int line){for(int y=line-1;y>=0;y--)if(lines[y].valid)return lines[y].before;return 0;}
// Timeline marks value changes; the side panel still exposes repeated writes.
static unsigned changes(int line){unsigned n=0;for(unsigned i=previous(line);i<lines[line].before;i++)if(relevant(writes[i])&&writes[i].oldValue!=writes[i].value)n++;return n;}
static std::string addressName(uint16_t a){char b[16];snprintf(b,sizeof b,"$%04X",a);return b;}
static std::string registerName(uint16_t a){
 switch(a){case 0xd011:return "D011";case 0xd016:return "D016";case 0xd018:return "D018";
 case 0xd020:return "Border";case 0xd021:return "BG 0";case 0xd022:return "BG 1";case 0xd023:return "BG 2";
 case 0xd015:return "Sprites";case 0xdd00:return "CIA2 port";case 0xdd02:return "CIA2 DDR";default:return addressName(a);}
}
static std::string info(){
 std::ostringstream o;o<<"{\"available\":"<<(valid()?"true":"false")<<",\"complete\":"<<(complete()?"true":"false")<<",\"width\":"<<W<<",\"height\":"<<H<<",\"lines\":[";bool comma=false;
 if(valid())for(int y=0;y<H;y++)if(lines[y].valid){if(comma)o<<',';comma=true;o<<"{\"line\":"<<y<<",\"raster\":"<<lines[y].raster<<",\"changes\":"<<changes(y)<<"}";}
 o<<"],\"bytes\":"<<sizeof(lines)+sizeof(charROM)+sizeof(usedMemory)+(initialWriters?sizeof(*initialWriters):0)+writes.size()*sizeof(observation::Write)<<"}";return o.str();
}
// Address/value tracking shadows only the preview VIC's shifters. Writers are
// resolved against the selected snapshot, not the current board's history.
struct Source {uint16_t address=0;uint8_t space=0,value=0;bool valid=false;};
struct Graphic {Source screen,color,data;};
struct SpriteByte {Source data,pointer;};
static Source screens[64],colors[64],pointers[8];
static SpriteByte spriteBits[8][32],spriteOut[8][2];
static Graphic pending,active;
static const Line* frozen=nullptr;
static int queryPanel=-1,queryX=-1,queryY=-1;
static std::string evidence;
static Source source(int space,uint16_t a,uint8_t v){return {a,uint8_t(space),v,true};}
void fetch(int kind,int slot,uint16_t a,uint16_t value){
 a|=frozen->bank;auto s=source(observation::vicSpace(a),a,value);
 if(kind==0){screens[slot]=s;colors[slot]=source(1,0xd800|(a&1023),value>>8);}
 else if(kind==1)pending={screens[slot],colors[slot],s};
 else if(kind==2)pointers[slot]=s;
 else if(kind==3){for(int bit=31;bit>=8;bit--)spriteBits[slot][bit]=spriteBits[slot][bit-8];for(int bit=8;bit<16;bit++)spriteBits[slot][bit]={s,pointers[slot]};for(int bit=0;bit<8;bit++)spriteBits[slot][bit]={};}
}
void idle(){pending={};}
void load(bool enabled){active=enabled?pending:Graphic{};}
void spriteOutput(int i,bool pair){spriteOut[i][pair]=spriteBits[i][31];}
void spriteShift(int i){for(int b=31;b>0;b--)spriteBits[i][b]=spriteBits[i][b-1];spriteBits[i][0]={};}
static std::string sourceJSON(const Source&s,const std::string&role){
 if(!s.valid)return "{\"role\":\""+role+"\",\"missing\":true}";
 auto k=observation::key(s.space,s.address);const observation::Write*w=nullptr;
 if(k<observation::CELLS)w=lastWrite[k]?&writes[lastWrite[k]-1]:&(*initialWriters)[k];
 return "{\"role\":\""+role+"\",\"address\":"+std::to_string(s.address)+",\"space\":"+std::to_string(s.space)+",\"value\":"+std::to_string(s.value)+",\"writer\":"+(w?observation::writerJSON(*w):"null")+"}";
}
void pixel(void*ptr,uint8_t*dst,uint16_t graphics,uint16_t sprites,bool border){
 auto&v=*static_cast<m6569_t*>(ptr);auto offset=dst-v.crt.fb;int y=offset/M6569_FRAMEBUFFER_WIDTH,x=offset%M6569_FRAMEBUFFER_WIDTH;
 if(x>=W||y>=H)return;
 pictures[0][y*W+x]=m6569_color((border?(v.brd.main?v.brd.bc:v.gunit.bg[0]):graphics)&15);
 pictures[1][y*W+x]=(sprites>>8)?m6569_color(sprites&15):0;
 if(x!=queryX||y!=queryY||queryPanel<0)return;
 std::ostringstream o;o<<"{\"panel\":"<<queryPanel<<",\"x\":"<<x<<",\"y\":"<<y<<",\"stateLine\":"<<selected<<",\"preview\":true,\"raster\":"<<frozen->raster<<",\"color\":"<<pictures[queryPanel][y*W+x]<<",\"complete\":"<<(complete()?"true":"false")<<",\"contributors\":[";
 bool comma=false;auto add=[&](Source s,const std::string&role){if(comma)o<<',';comma=true;o<<sourceJSON(s,role);};
 auto reg=[&](int i,const char*role){add(source(2,0xd000+i,frozen->regs[i]),role);};
 if(queryPanel==0){
  if(border)reg(v.brd.main?0x20:0x21,"border/background color");
  else {add(active.screen,"screen byte");add(active.color,"color RAM");add(active.data,(v.gunit.mode&2)?"bitmap byte":"character graphics byte");for(int i=0x21;i<=0x24;i++)reg(i,"background palette");}
 }else for(int i=0;i<8;i++)if((sprites>>8)&(1<<i)){
  const auto&s=spriteOut[i][(v.reg.mmc>>i)&1];add(s.data,"sprite "+std::to_string(i)+" pixel byte");add(s.pointer,"sprite "+std::to_string(i)+" pointer");reg(0x27+i,"sprite color");reg(i*2,"sprite X");reg(i*2+1,"sprite Y");
 }
 for(int i:{0x11,0x16,0x18})reg(i,"frozen VIC control");
 if(queryPanel==1)for(int i:{0x10,0x15,0x17,0x1b,0x1c,0x1d,0x25,0x26})reg(i,"sprite control/palette");
 for(int a:{0xdd00,0xdd02}){auto k=observation::key(2,a);const auto&w=lastWrite[k]?writes[lastWrite[k]-1]:(*initialWriters)[k];add(source(2,a,w.value),"VIC bank control");}
 o<<"]}";evidence=o.str();
}
static uint16_t previewFetch(uint16_t address,void*){
 address|=frozen->bank;
 uint8_t value=observation::vicSpace(address)==3?charROM[address&4095]:frozen->ram[address];
 return uint16_t(value)|(uint16_t(frozen->color[address&1023])<<8);
}
static void preview(int panel=-1,int x=-1,int y=-1){
 frozen=&lines[selected];queryPanel=panel;queryX=x;queryY=y;evidence="{\"error\":\"No preview pixel was produced\"}";
 pictures[0].fill(0);pictures[1].fill(0);pending=active={};
 std::fill_n(screens,64,Source{});std::fill_n(colors,64,Source{});std::fill_n(pointers,8,Source{});
 for(auto&row:spriteBits)for(auto&v:row)v={};for(auto&row:spriteOut)for(auto&v:row)v={};
 lastWrite.fill(0);for(unsigned i=0;i<frozen->before;i++){auto&w=writes[i];auto k=observation::key(w.space,w.address);if(k<lastWrite.size())lastWrite[k]=i+1;}
 m6569_t vic{};m6569_desc_t desc{};desc.fetch_cb=previewFetch;desc.framebuffer={scratch.data(),scratch.size()};desc.screen={_C64_SCREEN_X,_C64_SCREEN_Y,W,H};
 m6569_init(&vic,&desc);
 struct PreviewScope {PreviewScope(){previewing=true;}~PreviewScope(){previewing=false;}} scope;
 for(int i=0;i<47;i++){uint64_t pins=0xd000+i;M6569_SET_DATA(pins,frozen->regs[i]);_m6569_write(&vic,pins);}
 // Warm the fixed registers through one PAL frame, then collect a full frame.
 // This is a projection, not a replay of the guest's mid-frame register writes.
 for(int i=0;i<2*M6569_HTOTAL*M6569_VTOTAL;i++)m6569_tick(&vic,M6569_RW);
}
static std::string seek(int line){
 if(!valid()||line<0||line>=H||!lines[line].valid)return "{\"error\":\"Capture a complete display interval first\"}";
 selected=line;preview();pictures[2].fill(0);
 for(int y=0;y<=line;y++)for(int x=0;x<W;x++){auto&p=observation::captured[y*W+x];if(p.cycle)pictures[2][y*W+x]=m6569_color(p.color&15);}
 const auto&s=lines[line];unsigned screen=s.bank+((s.regs[0x18]&0xf0)<<6),chars=s.bank+((s.regs[0x18]&14)<<10),bitmap=s.bank+((s.regs[0x18]&8)<<10);
 std::ostringstream o;o<<"{\"line\":"<<line<<",\"raster\":"<<s.raster<<",\"clock\":"<<s.cycle<<",\"complete\":"<<(complete()?"true":"false")<<",\"vicBank\":"<<s.bank<<",\"screenBase\":"<<screen<<",\"charsetBase\":"<<chars<<",\"bitmapBase\":"<<bitmap<<",\"mode\":"<<(((s.regs[0x11]>>4)&6)|((s.regs[0x16]>>4)&1))<<",\"registers\":[";bool comma=false;
 for(int i:{0x11,0x16,0x18,0x15,0x20,0x21,0x22,0x23}){if(comma)o<<',';comma=true;o<<"{\"name\":\""<<registerName(0xd000+i)<<"\",\"address\":"<<0xd000+i<<",\"value\":"<<unsigned(s.regs[i])<<"}";}
 o<<"],\"changes\":[";comma=false;unsigned count=0,graphics=0,maps=0,sprites=0,colors=0;
 for(unsigned i=previous(line);i<s.before;i++){
  auto&w=writes[i];auto kind=relevant(w);if(!kind)continue;
  graphics+=!!(kind&2);maps+=!!(kind&1);sprites+=!!(kind&12);colors+=!!(kind&32);count++;
  if(count>64)continue;if(comma)o<<',';comma=true;auto a=observation::canonical(w.space,w.address);
  o<<"{\"name\":\""<<((kind&16)?registerName(a):addressName(a))<<"\",\"address\":"<<a<<",\"space\":"<<unsigned(w.space)<<",\"before\":"<<unsigned(w.oldValue)<<",\"after\":"<<unsigned(w.value)<<",\"pc\":"<<w.pc<<",\"clock\":"<<w.cycle<<",\"writer\":"<<observation::writerJSON(w)<<"}";
 }
 o<<"],\"changeCount\":"<<count<<",\"tileWrites\":"<<graphics<<",\"mapWrites\":"<<maps<<",\"objectWrites\":"<<sprites<<",\"colorWrites\":"<<colors<<"}";return o.str();
}
static std::string pixelQuery(int panel,int x,int y){
 if(!valid()||selected<0||panel<0||panel>2||x<0||x>=W||y<0||y>=H)return "{\"error\":\"Select a captured scanline first\"}";
 if(panel==2){
  if(y>selected||!observation::captured[y*W+x].cycle)return "{\"error\":\"This line has not been drawn at this point in the frame.\"}";
  std::string result=rr_pixel(x,y);if(result.find("\"error\"")!=std::string::npos)return result;
  result.pop_back();return result+",\"panel\":2,\"stateLine\":"+std::to_string(y)+",\"preview\":false,\"complete\":"+(complete()?"true":"false")+"}";
 }
 preview(panel,x,y);return evidence;
}
}
extern "C" {
// Registers [0,47), bases [64,68), RGBA palette [80,144), color RAM
// [144,1168), screen [1168,2192), VIC-visible character bytes [2192,4240).
int rr_tileset_size(){return raster::valid()&&raster::selected>=0&&raster::lines[raster::selected].valid?4240:0;}
uint8_t*rr_tileset_data(){
 static std::array<uint8_t,4240>b{};if(!rr_tileset_size())return nullptr;
 const auto&s=raster::lines[raster::selected];std::copy(s.regs.begin(),s.regs.end(),b.begin());
 unsigned chars=s.bank+((s.regs[0x18]&14)<<10),screen=s.bank+((s.regs[0x18]&0xf0)<<6);
 b[64]=chars;b[65]=chars>>8;b[66]=screen;b[67]=screen>>8;
 for(unsigned i=0;i<16;i++){auto c=m6569_color(i);for(unsigned k=0;k<4;k++)b[80+i*4+k]=c>>(k*8);}
 std::copy(s.color.begin(),s.color.end(),b.begin()+144);
 auto fetch=[&](unsigned a){return observation::vicSpace(a)==3?raster::charROM[a&4095]:s.ram[a];};
 for(unsigned i=0;i<1024;i++)b[1168+i]=fetch(screen+i);
 for(unsigned i=0;i<2048;i++)b[2192+i]=fetch(chars+i);
 return b.data();
}
const char*rr_raster_info(){observation::json=raster::info();return observation::json.c_str();}
const char*rr_raster_seek(int line){observation::json=raster::seek(line);return observation::json.c_str();}
uint32_t*rr_raster_frame(int panel){return panel>=0&&panel<3?raster::pictures[panel].data():nullptr;}
const char*rr_raster_pixel(int panel,int x,int y){observation::json=raster::pixelQuery(panel,x,y);return observation::json.c_str();}
}
