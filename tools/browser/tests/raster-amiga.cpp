#include "../../platform/amiga/browser/core/api.cpp"
#include <cassert>
#include <iostream>
using namespace rramiga;
static void boot(){
 std::vector<uint8_t>rom(262144),disk(901120);rom[1]=7;rom[5]=0xfc;rom[7]=8;rom[8]=0x60;rom[9]=0xfe;
 std::memcpy(rr_firmware(rom.size()),rom.data(),rom.size());std::memcpy(rr_input(disk.size()),disk.data(),disk.size());assert(rr_init(disk.size()));
}
static bool matched(const Source&s){return std::string(rr_source(s.address,s.size,s.before,s.value)).find("\"complete\":true")!=std::string::npos;}
int main(){
 boot();auto&m=*machine;m.overlay=false;m.line=44;m.dma=0x300;
 m.reg[0x8e/2]=0x2c81;m.reg[0x90/2]=0x2cc1;m.reg[0x92/2]=0x38;m.reg[0x94/2]=0xd0;m.reg[0x100/2]=0x1000;m.bplPointer[0]=0x10000;
 m.reg[0x180/2]=0xf00;m.reg[0x182/2]=0x0f0;m.beginPalette();
 assert(rr_capture_begin());m.beginPalette();m.beam=200;m.customWrite(0x180,0x00f);m.beam=440;m.renderLine();m.screen=m.draw;assert(rr_capture_end());
 assert(m.draw[0]==0xff0000ff);assert(m.draw[141]==0xff0000ff);assert(m.draw[142]==0xffff0000);
 auto state=save();rr_raster_seek(0);assert(inspect::pictures[2][0]==m.draw[0]);assert(inspect::pictures[2][W]==0);
 assert(std::string(rr_raster_pixel(2,141,0)).find("\"value\":15")!=std::string::npos);
 auto&l=inspect::lines[0];assert(matched({"red",registerBase+0x180,swapped(0xf00),inspect::paletteCutoff(l,0,141)}));assert(matched({"blue",registerBase+0x180,swapped(0x00f),inspect::paletteCutoff(l,0,142)}));
 rr_raster_seek(0);rr_raster_pixel(0,100,20);assert(save()==state);
 // Real Copper WAIT/MOVE changes a palette at its beam location; elapsed clocks
 // must advance within a multi-line host tick, not all share its ending time.
 boot();auto&c=*machine;c.overlay=false;c.dma=0x280;c.line=44;c.reg[0x180/2]=0xf00;c.copperPC=0x8000;c.copperStopped=false;
 uint16_t list[]={0x2c65,0xfffe,0x0180,0x00f,0xffff,0xfffe};for(unsigned i=0;i<6;i++)c.chipWrite(0x8000+i*2,list[i]);
 assert(rr_capture_begin());c.tick(454);c.screen=c.draw;assert(rr_capture_end());
 assert(inspect::changes.size()==1);auto change=inspect::changes[0];assert(change.line==44&&change.beam>200&&change.beam<230);assert(change.clock==change.beam);assert(change.copperPC==0x8004);assert(c.draw[0]==0xff0000ff&&c.draw[639]==0xffff0000);
 // Actual shifted cookie-cut inputs, edge masks, overlapping source histories.
 boot();auto&b=*machine;b.overlay=false;b.line=44;b.dma=0x300;b.reg[0x8e/2]=0x2c81;b.reg[0x90/2]=0x2cc1;b.reg[0x92/2]=0x38;b.reg[0x94/2]=0xd0;b.reg[0x100/2]=0x1000;b.bplPointer[0]=0x6000;b.reg[0x182/2]=0xfff;
 b.reg[0x44/2]=0x0fff;b.reg[0x46/2]=0xf000;b.reg[0x40/2]=0x4fca;b.reg[0x42/2]=0x4000;b.setPtr(0x50,0x4000);b.setPtr(0x4c,0x5000);b.setPtr(0x48,0x6000);b.setPtr(0x54,0x6000);
 b.chipWrite(0x4000,0xffff);b.chipWrite(0x4002,0xffff);b.chipWrite(0x5000,0xaaaa);b.chipWrite(0x5002,0x5555);b.chipWrite(0x6000,0x3333);b.chipWrite(0x6002,0xcccc);
 assert(rr_capture_begin());b.pc=0x1234;b.customWrite(0x58,66);b.renderLine();b.screen=b.draw;assert(rr_capture_end());
 assert(inspect::blits.size()==1&&inspect::words.size()==2);auto w=inspect::words[0];assert(w.a==0xff&&w.b==0xaaa&&w.c==0x3333&&w.d==0x33aa);assert(w.rawA==0xffff&&w.oldD==0x3333);assert(inspect::words[1].a==0xff00);assert(b.chip16(0x6002)==0xa5cc);
 assert(std::string(rr_source(0x6000,2,UINT32_MAX,0)).find("\"blit\":0")!=std::string::npos);
 assert(matched({"A",0x4000,swapped(0xffff),w.before}));assert(matched({"C",0x6000,swapped(0x3333),w.before}));assert(matched({"D",0x6000,swapped(0x33aa),w.write}));
 state=save();rr_blit_seek(0);rr_blit_pixel(20,0);rr_raster_seek(0);rr_raster_pixel(2,0,0);assert(save()==state);assert(inspect::blitPictures[4]!=inspect::blitPictures[5]);
 // Hardware sprites remain separate from the playfield preview, including
 // transparent sprite pixels and pixels hidden by playfield priority.
 boot();auto&sp=*machine;sp.overlay=false;sp.line=44;sp.dma=0x320;sp.reg[0x8e/2]=0x2c81;sp.reg[0x90/2]=0x2cc1;sp.reg[0x92/2]=0x38;sp.reg[0x94/2]=0xd0;sp.reg[0x100/2]=0x1000;sp.bplPointer[0]=0x6000;
 sp.reg[0x180/2]=0xf00;sp.reg[0x182/2]=0x00f;sp.reg[(0x180+17*2)/2]=0x0f0;sp.chipWrite(0x6000,0xffff);
 sp.sprites[0].pos=0x2c41;sp.sprites[0].ctl=0x2d01;sp.sprites[0].pointer=0x7004;sp.sprites[0].active=true;sp.sprites[0].a=0x8000;sp.chipWrite(0x7000,0x8000);
 assert(rr_capture_begin());sp.renderLine();sp.screen=sp.draw;assert(rr_capture_end());rr_raster_seek(0);
 assert(inspect::pictures[0][4]==0xffff0000);assert(inspect::pictures[1][4]==0xff00ff00);assert(inspect::pictures[2][4]==0xffff0000);assert(inspect::pictures[1][6]==0);
 assert(std::string(rr_raster_pixel(1,4,0)).find("Hardware sprite")!=std::string::npos);
 // In line mode A/B are register-derived, not reads from their DMA pointers.
 boot();auto&ln=*machine;ln.overlay=false;ln.reg[0x40/2]=0x0fca;ln.reg[0x42/2]=1;ln.reg[0x74/2]=0x8000;ln.reg[0x72/2]=0xffff;ln.setPtr(0x48,0x6000);ln.setPtr(0x54,0x6000);
 assert(rr_capture_begin());ln.customWrite(0x58,66);assert(rr_capture_end());rr_blit_seek(0);auto linePixel=std::string(rr_blit_pixel(0,0));
 assert(linePixel.find("A data register")!=std::string::npos&&linePixel.find("B data register")!=std::string::npos&&linePixel.find("A fetched word")==std::string::npos);
 assert(matched({"line A",registerBase+0x74,swapped(0x8000),inspect::words[0].before}));assert(matched({"line B",registerBase+0x72,swapped(0xffff),inspect::words[0].before}));
 // Exact pending-palette save and restore, including horizontal progress.
 boot();auto&s=*machine;s.line=44;s.beam=180;s.beginPalette();s.customWrite(0x180,0x0af);auto initial=save();s.tick(274);auto expected=save();restore(initial.data(),initial.size());s.tick(274);assert(save()==expected);
 assert(std::string(rr_raster_info()).find("\"lines\":[]")!=std::string::npos);
 std::cout<<"Amiga scanline/Copper palette positions, planar previews, cookie-cut inputs, source history and immutable inspection passed\n";
}
