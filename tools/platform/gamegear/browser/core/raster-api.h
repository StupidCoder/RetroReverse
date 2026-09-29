#pragma once
namespace rrgg {
inline std::array<std::array<uint32_t,23040>,3> pictures{};
struct Sample {uint32_t color,source,bits,palette,paletteValue,texel;int object,u,v;uint8_t flags;GGPixelSource extra;};
struct Paint {
 int object=-1,queryX=-1,queryY=-1;std::vector<Sample> samples;
 std::array<uint32_t,23040>*background=nullptr,*sprites=nullptr;
 void event(const char*,int,bool,int index,uint32_t){object=index;}
 void dot(int x,int y,uint32_t color,uint32_t source,uint32_t bits,uint32_t pal,uint32_t palette,unsigned texel,int u,int v,uint8_t flags,const GGPixelSource&extra){
  if(x<0||x>=160||y<0||y>=144)return;
  if(background&&object<0)(*background)[y*160+x]=color;
  if(sprites&&object>=0&&(flags&3))(*sprites)[y*160+x]=color;
  if(x==queryX&&y==queryY)samples.push_back({color,source,bits,pal,palette,texel,object,u,v,flags,extra});
 }
};
struct Events {Paint&p;void operator()(const char*k,int y,bool source=true,int i=-1,uint32_t desc=0)const{p.event(k,y,source,i,desc);}};
struct Dots {Paint&p;void operator()(int x,int y,uint32_t color,uint32_t source,uint32_t bits,uint32_t pal,uint32_t palette,unsigned texel,int u,int v,uint8_t flags,const GGPixelSource&extra)const{p.dot(x,y,color,source,bits,pal,palette,texel,u,v,flags,extra);}};
inline void restore(gamegear_Machine&m,const Line&s){m.VDP.VRAM=s.vram;m.VDP.CRAM=s.cram;m.VDP.Regs=s.regs;}
inline constexpr const char*registers[]={"Mode 1","Mode 2","Tilemap","R3","R4","Sprites","Sprite tiles","Backdrop","Scroll X","Scroll Y","Line reload"};
inline unsigned changesStart(int line){for(int y=line-1;y>=0;y--)if(lines[y].valid)return lines[y].before;return 0;}
inline bool videoAddress(uint32_t a){return a>=0x10000&&a<0x1404b;}
inline unsigned changes(int line){unsigned n=0;const auto&t=rrcapture::trace;for(unsigned i=changesStart(line);i<lines[line].before;i++)if(videoAddress(t.writes[i].address))n++;return n;}
inline bool complete(){if(!rrcapture::trace.valid||rrcapture::trace.overflow)return false;for(const auto&line:lines)if(!line.valid)return false;return true;}
inline std::string info(){
 const auto&t=rrcapture::trace;std::ostringstream o;o<<"{\"available\":"<<(t.valid?"true":"false")<<",\"complete\":"<<(complete()?"true":"false")<<",\"lines\":[";bool comma=false;
 if(t.valid)for(int y=0;y<144;y++)if(lines[y].valid){if(comma)o<<',';comma=true;o<<"{\"line\":"<<y<<",\"raster\":"<<y+24<<",\"changes\":"<<changes(y)<<"}";}
 o<<"],\"bytes\":"<<sizeof(lines)<<"}";return o.str();
}
inline std::string seek(int line){
 if(!rrcapture::trace.valid||line<0||line>=144||!lines[line].valid)return "{\"error\":\"Capture a complete display interval first.\"}";
 selected=line;const auto&s=lines[line];gamegear_Machine preview{};restore(preview,s);for(auto&p:pictures)p.fill(0);
 Paint p;p.background=&pictures[0];p.sprites=&pictures[1];
 for(int y=0;y<144;y++)rrGGRenderLine(&preview,y+24,s.scrollX,s.scrollY,Events{p},Dots{p});
 for(int y=0;y<=line;y++)if(lines[y].valid)std::copy(lines[y].output.begin(),lines[y].output.end(),pictures[2].begin()+y*160);
 unsigned nt=(s.regs[2]&14)<<10,sat=(s.regs[5]&0x7e)<<7;
 std::ostringstream o;o<<"{\"line\":"<<line<<",\"raster\":"<<line+24<<",\"clock\":"<<s.clock<<",\"scrollX\":"<<unsigned(s.scrollX)<<",\"scrollY\":"<<unsigned(s.scrollY)<<",\"lineCounter\":"<<unsigned(s.lineCounter)<<",\"linePending\":"<<(s.linePending?"true":"false")<<",\"framePending\":"<<(s.framePending?"true":"false")<<",\"tilemapBase\":"<<nt<<",\"spriteBase\":"<<sat<<",\"complete\":"<<(complete()?"true":"false")<<",\"registers\":[";bool comma=false;
 for(unsigned i:{0,1,2,5,6,7,8,9,10}){if(comma)o<<',';comma=true;o<<"{\"name\":\""<<registers[i]<<"\",\"address\":"<<0x14040+i<<",\"value\":"<<unsigned(s.regs[i])<<"}";}
 o<<"],\"changes\":[";comma=false;unsigned count=0,tiles=0,maps=0,objects=0,colors=0;
 for(unsigned i=changesStart(line);i<s.before;i++){
  const auto&w=rrcapture::trace.writes[i];if(!videoAddress(w.address))continue;
  if(w.address<0x14000){unsigned a=w.address-0x10000;tiles++;if(a>=nt&&a<nt+0x700)maps++;if(a>=sat&&a<sat+256)objects++;}
  else if(w.address<0x14040)colors++;
  else {count++;if(count>64)continue;if(comma)o<<',';comma=true;o<<"{\"name\":\""<<registers[w.address-0x14040]<<"\",\"address\":"<<w.address<<",\"before\":"<<w.before<<",\"after\":"<<w.after<<",\"pc\":"<<w.pc<<",\"clock\":"<<w.clock<<"}";}
 }
 o<<"],\"changeCount\":"<<count<<",\"tileWrites\":"<<tiles<<",\"mapWrites\":"<<maps<<",\"objectWrites\":"<<objects<<",\"colorWrites\":"<<colors<<",\"sourceBefore\":"<<s.before<<"}";return o.str();
}
inline std::string pixel(int panel,int x,int y){
 if(selected<0||!rrcapture::trace.valid||!lines[selected].valid||panel<0||panel>2||x<0||x>=160||y<0||y>=144)return "{\"error\":\"Select a captured scanline first.\"}";
 if(panel==2&&(y>selected||!lines[y].valid))return "{\"error\":\"This line has not been drawn at this point in the frame.\"}";
 int anchor=panel==2?y:selected;const auto&s=lines[anchor];gamegear_Machine preview{};restore(preview,s);
 Paint p;p.queryX=x;p.queryY=y;rrGGRenderLine(&preview,y+24,s.scrollX,s.scrollY,Events{p},Dots{p});
 uint32_t composed=0;for(const auto&v:p.samples)if(v.flags&1)composed=v.color;
 std::ostringstream o;o<<"{\"panel\":"<<panel<<",\"x\":"<<x<<",\"y\":"<<y<<",\"stateLine\":"<<anchor<<",\"sourceBefore\":"<<s.before<<",\"color\":"<<pictures[panel][y*160+x]<<",\"complete\":"<<(!rrcapture::trace.overflow&&(panel!=2||composed==s.output[x])?"true":"false")<<",\"candidates\":[";bool comma=false;
 for(const auto&v:p.samples){if((panel==0&&v.object>=0)||(panel==1&&v.object<0))continue;if(comma)o<<',';comma=true;
  o<<"{\"kind\":\""<<(v.object>=0?"Sprite":v.source?"Background":"Backdrop")<<"\",\"object\":"<<v.object<<",\"flags\":"<<unsigned(v.flags)<<",\"color\":"<<v.color<<",\"texel\":"<<v.texel<<",\"sources\":[";bool sep=false;
  auto source=[&](const char*role,uint32_t a,int size,uint32_t value){if(!a)return;if(sep)o<<',';sep=true;o<<"{\"role\":\""<<role<<"\",\"address\":"<<a<<",\"size\":"<<size<<",\"value\":"<<value<<"}";};
  source("Tile row",v.source,4,v.bits);source("Tilemap entry",v.extra.map,2,v.extra.mapValue);source("Sprite Y",v.extra.spriteY,1,v.extra.spriteYValue);source("Sprite X / tile",v.extra.spriteXT,2,v.extra.spriteXTValue);source("Palette color",v.palette,2,v.paletteValue);o<<"]}";
 }
 o<<"]}";return o.str();
}
}
extern "C" {
// Inspection-only VRAM/palette snapshot; never reads the live machine.
int rr_tileset_size(){return rrcapture::trace.valid&&rrgg::selected>=0&&rrgg::lines[rrgg::selected].valid?16464:0;}
uint8_t*rr_tileset_data(){static std::array<uint8_t,16464>b; if(!rr_tileset_size())return nullptr;const auto&s=rrgg::lines[rrgg::selected];std::copy(s.vram.begin(),s.vram.end(),b.begin());std::copy(s.cram.begin(),s.cram.end(),b.begin()+16384);std::copy(s.regs.begin(),s.regs.end(),b.begin()+16448);return b.data();}
const char*rr_raster_info(){reply=rrgg::info();return reply.c_str();}
const char*rr_raster_seek(int line){reply=rrgg::seek(line);return reply.c_str();}
uint8_t*rr_raster_frame(int panel){return panel>=0&&panel<3?reinterpret_cast<uint8_t*>(rrgg::pictures[panel].data()):nullptr;}
const char*rr_raster_pixel(int panel,int x,int y){reply=rrgg::pixel(panel,x,y);return reply.c_str();}
}
