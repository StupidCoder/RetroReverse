#pragma once
namespace rrgb {
inline std::array<std::array<uint32_t,23040>,3> pictures{};
struct Sample {uint32_t color=0,source=0,bits=0,palette=0,paletteValue=0,texel=0,descriptor=0;int object=-1,u=0,v=0;uint8_t flags=0;bool sprite=false;};
struct Paint {
 bool sprite=false;int object=-1;uint32_t descriptor=0;
 int queryX=-1,queryY=-1;std::vector<Sample> samples;
 std::array<uint32_t,23040>*background=nullptr,*sprites=nullptr;
 void event(const char*,int,bool=true,int index=-1,uint32_t desc=0){sprite=index>=0;object=index;descriptor=desc;}
 void dot(int x,int y,uint32_t color,uint32_t source,uint32_t bits,uint32_t pal,uint32_t palette,unsigned texel,int u,int v,uint8_t flags=1){
  if(x<0||x>=160||y<0||y>=144)return;
  if(background&&!sprite)(*background)[y*160+x]=color;
  // The object which wins against other objects remains visible in the sprite
  // preview even if background priority hides it in the composited screen.
  if(sprites&&sprite&&(flags&3))(*sprites)[y*160+x]=color;
  if(x==queryX&&y==queryY)samples.push_back({color,source,bits,pal,palette,texel,descriptor,object,u,v,flags,sprite});
 }
};
struct Events {Paint&p;void operator()(const char*k,int y,bool source=true,int index=-1,uint32_t descriptor=0)const{p.event(k,y,source,index,descriptor);}};
struct Dots {Paint&p;void operator()(int x,int y,uint32_t color,uint32_t source,uint32_t bits,uint32_t pal,uint32_t palette,unsigned texel,int u,int v,uint8_t flags=1)const{p.dot(x,y,color,source,bits,pal,palette,texel,u,v,flags);}};
inline void restore(gameboy_Machine&m,const Line&s){m.vram=s.vram;m.oam=s.oam;m.io=s.io;}
inline int previewWindow(const Line&s,int anchor,int y){
 bool active=(s.io[0x40]&0x21)==0x21&&anchor>=s.io[0x4a]&&s.io[0x4b]<=166;
 return std::max(0,active?s.windowLine+y-anchor:y-int(s.io[0x4a]));
}
struct Register {uint16_t address;const char*name;};
inline constexpr Register registers[]={{0xff40,"LCDC"},{0xff42,"SCY"},{0xff43,"SCX"},{0xff45,"LYC"},{0xff46,"DMA"},{0xff47,"BGP"},{0xff48,"OBP0"},{0xff49,"OBP1"},{0xff4a,"WY"},{0xff4b,"WX"}};
inline const char*regName(uint32_t a){for(auto r:registers)if(r.address==a)return r.name;return nullptr;}
inline uint32_t changesStart(int line){for(int y=line-1;y>=0;y--)if(lines[y].valid)return lines[y].before;return 0;}
inline bool videoAddress(uint32_t a){return (a>=0x8000&&a<0xa000)||(a>=0xfe00&&a<0xfea0)||regName(a);}
inline unsigned changes(int line){unsigned n=0;auto&t=rrcapture::trace;for(unsigned i=changesStart(line);i<lines[line].before;i++)if(videoAddress(t.writes[i].address))n++;return n;}
inline std::string info(){
 const auto&t=rrcapture::trace;bool full=t.valid&&!t.overflow;
 for(const auto&line:lines)full=full&&line.valid;
 std::ostringstream o;o<<"{\"available\":"<<(t.valid?"true":"false")<<",\"complete\":"<<(full?"true":"false")<<",\"lines\":[";bool comma=false;
 if(rrcapture::trace.valid)for(int y=0;y<144;y++)if(lines[y].valid){if(comma)o<<',';comma=true;o<<"{\"line\":"<<y<<",\"changes\":"<<changes(y)<<"}";}
 o<<"],\"bytes\":"<<sizeof(lines)<<"}";return o.str();
}
inline std::string seek(int line){
 if(!rrcapture::trace.valid||line<0||line>=144||!lines[line].valid)return "{\"error\":\"No rendered scanline was recorded here. The LCD may have been disabled.\"}";
 selected=line;const auto&s=lines[line];gameboy_Machine preview{};restore(preview,s);
 for(auto&p:pictures)p.fill(0);
 Paint p;p.background=&pictures[0];p.sprites=&pictures[1];
 for(int y=0;y<144;y++)rrGBRenderLine(&preview,y,previewWindow(s,line,y),Events{p},Dots{p});
 for(int y=0;y<=line;y++)if(lines[y].valid)std::copy(lines[y].output.begin(),lines[y].output.end(),pictures[2].begin()+y*160);
 auto&t=rrcapture::trace;std::ostringstream o;
 o<<"{\"line\":"<<line<<",\"clock\":"<<s.clock<<",\"windowLine\":"<<s.windowLine<<",\"complete\":"<<(!t.overflow?"true":"false")<<",\"registers\":[";bool comma=false;
 for(auto r:registers){if(comma)o<<',';comma=true;o<<"{\"name\":\""<<r.name<<"\",\"address\":"<<r.address<<",\"value\":"<<unsigned(s.io[r.address&127])<<"}";}
 o<<"],\"changes\":[";comma=false;unsigned count=0,tiles=0,maps=0,objects=0;
 for(unsigned i=changesStart(line);i<s.before;i++){
  const auto&w=t.writes[i];if(w.address>=0x8000&&w.address<0x9800)tiles++;else if(w.address>=0x9800&&w.address<0xa000)maps++;else if(w.address>=0xfe00&&w.address<0xfea0)objects++;
  auto name=regName(w.address);if(!name)continue;count++;if(count>64)continue;if(comma)o<<',';comma=true;
  o<<"{\"name\":\""<<name<<"\",\"address\":"<<w.address<<",\"before\":"<<w.before<<",\"after\":"<<w.after<<",\"pc\":"<<w.pc<<",\"clock\":"<<w.clock<<"}";
 }
 o<<"],\"changeCount\":"<<count<<",\"tileWrites\":"<<tiles<<",\"mapWrites\":"<<maps<<",\"objectWrites\":"<<objects<<",\"sourceBefore\":"<<s.before<<"}";return o.str();
}
inline std::string pixel(int panel,int x,int y){
 if(selected<0||!lines[selected].valid||panel<0||panel>2||x<0||x>=160||y<0||y>=144||!rrcapture::trace.valid)return "{\"error\":\"Select a captured scanline first.\"}";
 if(panel==2&&(y>selected||!lines[y].valid))return "{\"error\":\"This line has not been drawn at this point in the frame.\"}";
 int anchor=panel==2?y:selected;const auto&s=lines[anchor];gameboy_Machine preview{};restore(preview,s);
 Paint p;p.queryX=x;p.queryY=y;int window=previewWindow(s,anchor,y);
 rrGBRenderLine(&preview,y,window,Events{p},Dots{p});
 uint32_t composed=0xffffffff;for(const auto&v:p.samples)if(v.flags&1)composed=v.color;
 std::ostringstream o;o<<"{\"panel\":"<<panel<<",\"x\":"<<x<<",\"y\":"<<y<<",\"stateLine\":"<<anchor<<",\"sourceBefore\":"<<s.before<<",\"color\":"<<pictures[panel][y*160+x]<<",\"complete\":"<<(!rrcapture::trace.overflow&&(panel!=2||composed==s.output[x])?"true":"false")<<",\"candidates\":[";bool comma=false;
 for(auto v:p.samples){if((panel==0&&v.sprite)||(panel==1&&!v.sprite))continue;if(comma)o<<',';comma=true;
  bool win=!v.sprite&&(s.io[0x40]&0x21)==0x21&&y>=s.io[0x4a]&&s.io[0x4b]<=166&&x>=int(s.io[0x4b])-7;
  o<<"{\"kind\":\""<<(v.sprite?"Sprite":win?"Window":"Background")<<"\",\"object\":"<<v.object<<",\"flags\":"<<unsigned(v.flags)<<",\"color\":"<<v.color<<",\"texel\":"<<v.texel<<",\"sourceAddress\":"<<v.source<<",\"sourceValue\":"<<v.bits<<",\"paletteAddress\":"<<v.palette<<",\"paletteValue\":"<<v.paletteValue;
  if(v.sprite){uint32_t descriptor=0;for(int k=0;k<4;k++)descriptor|=uint32_t(s.oam[v.object*4+k])<<(k*8);o<<",\"descriptor\":"<<v.descriptor<<",\"descriptorValue\":"<<descriptor;}
  else if(v.source){int bx=win?x-(int(s.io[0x4b])-7):(x+s.io[0x43])&255,by=win?window:(y+s.io[0x42])&255;uint32_t nt=(s.io[0x40]&(win?64:8))?0x1c00:0x1800;uint32_t map=nt+(by/8)*32+bx/8;o<<",\"mapAddress\":"<<0x8000+map<<",\"mapValue\":"<<unsigned(s.vram[map]);}
  o<<"}";
 }
 o<<"]}";return o.str();
}
}
extern "C" {
const char*rr_raster_info(){reply=rrgb::info();return reply.c_str();}
const char*rr_raster_seek(int line){reply=rrgb::seek(line);return reply.c_str();}
uint8_t*rr_raster_frame(int panel){return panel>=0&&panel<3?reinterpret_cast<uint8_t*>(rrgb::pictures[panel].data()):nullptr;}
const char*rr_raster_pixel(int panel,int x,int y){reply=rrgb::pixel(panel,x,y);return reply.c_str();}
}
