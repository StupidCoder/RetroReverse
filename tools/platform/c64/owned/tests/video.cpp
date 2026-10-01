#include "../board.h"
#include <cassert>
#include <iostream>
#include <memory>
using namespace rr::c64;
static void at(Board& b,unsigned line,unsigned cycle){for(unsigned i=0;i<20000;i++){if(b.state.vic.raster==line&&b.state.vic.cycle==cycle)return;assert(b.tick());}assert(false);}
static void setup(Board& b){
 b.power();b.kernal[8190]=0;b.kernal[8191]=2;b.state.ram[0x200]=0x4c;b.state.ram[0x201]=0;b.state.ram[0x202]=2;
 Cpu cpu;cpu.start(0x200);b.state.cpu=cpu.state;b.write(0,0x2f);b.write(1,0x37);
 b.write(0xdd02,3);b.write(0xdd00,3);b.write(0xd011,0x1b);b.write(0xd016,8);b.write(0xd018,0x18);
 b.write(0xd020,6);b.write(0xd021,0);b.write(0xd022,2);b.write(0xd023,3);b.write(0xd024,4);
}
static void video(){auto ptr=std::make_unique<Board>();auto& b=*ptr;setup(b);auto& v=b.state.vic;
 b.state.ram[0x400]=1;b.state.color[0]=5;b.state.ram[0x2008]=0xa5;
 at(b,51,11);assert(v.ba&&v.aec);const auto clocks=b.state.cpu.retired;
 for(unsigned c=12;c<=54;c++){assert(b.tick());assert(!v.ba);assert(v.aec==(c<15));if(c==15){assert(v.fetchCount==2&&v.fetches[1].kind==VicAccess::Matrix&&v.fetches[1].address==0x400&&v.fetches[1].value==1&&v.fetches[1].color==5);}if(c==16){assert(v.cells[0].graphics.address==0x2008&&v.cells[0].graphics.value==0xa5);b.state.ram[0x2008]=0;}}
 assert(b.state.cpu.retired==clocks);assert(b.tick()&&v.ba&&v.aec);
 for(unsigned x=0;x<8;x++)assert(v.pixels[51*504+24+x]==((0xa5&(128>>x))?5:0));assert(v.pixels[51*504+23]==6);
 assert(v.cells[0].graphics.value==0xa5); // RAM edit after fetch cannot change emitted pixels.
 // Read-only inspection does not acknowledge collision latches; real reads do.
 v.collisionSprites=3;assert(b.peek(0xd01e)==3&&v.collisionSprites==3);assert(b.read(0xd01e)==3&&!v.collisionSprites);
 v.flags=0;b.write(0xd01a,1);b.write(0xd012,52);at(b,52,1);assert(v.irq());b.write(0xd019,1);assert(!v.irq());b.write(0xd012,53);at(b,53,1);assert(v.irq());
 // Pure register comparison edges, including the delayed raster-zero compare.
 setup(b);b.write(0xd012,1);b.write(0xd01a,1);at(b,1,1);assert(v.irq());b.write(0xd019,1);b.write(0xd012,0);at(b,0,1);assert(!v.irq());assert(b.tick()&&v.irq());
 // Character-ROM mapping follows VIC bank, independent of CPU banking.
 setup(b);b.chars[8]=0x80;b.write(0xd018,0x14);b.state.ram[0x400]=1;b.state.color[0]=7;at(b,51,20);assert(v.cells[0].graphics.rom&&v.cells[0].graphics.address==0x1008&&v.pixels[51*504+24]==7);
 setup(b);b.write(0xdd00,1);b.write(0xd018,0x14);b.state.ram[0x8400]=1;b.state.color[0]=7;at(b,51,20);assert(v.cells[0].graphics.rom&&v.cells[0].graphics.address==0x9008);
 // All five valid modes, multicolor foreground classification and scroll.
 for(unsigned mode=0;mode<5;mode++){
  setup(b);const bool bitmap=mode==2||mode==3,multi=mode==1||mode==3,ecm=mode==4;
  b.write(0xd011,uint8_t(0x1b|(bitmap?32:0)|(ecm?64:0)));b.write(0xd016,uint8_t(8|(multi?16:0)));b.write(0xd018,0x18);
  b.state.ram[0x400]=bitmap?0x56:ecm?0x81:1;b.state.color[0]=multi?13:5;b.state.ram[bitmap?0x2000:0x2008]=0x1b;
  at(b,51,22);const uint8_t wanted[5][8]={{0,0,0,5,5,0,5,5},{0,0,2,2,3,3,5,5},{6,6,6,5,5,6,5,5},{0,0,5,5,6,6,13,13},{3,3,3,5,5,3,5,5}};
  for(unsigned x=0;x<8;x++)assert(v.pixels[51*504+24+x]==wanted[mode][x]);
 }
 setup(b);b.state.ram[0x400]=1;b.state.color[0]=5;b.state.ram[0x2008]=255;b.write(0xd016,11);at(b,51,23);assert(v.pixels[51*504+26]==0&&v.pixels[51*504+27]==5);
 // Sprite 0 and 1 overlap a text foreground; lower-numbered sprite priority,
 // behind-foreground and both collision IRQs. Data fetched before display.
 setup(b);b.state.ram[0x400]=1;b.state.color[0]=5;b.state.ram[0x2008]=255;
 b.write(0xd015,3);b.write(0xd000,24);b.write(0xd002,24);b.write(0xd001,50);b.write(0xd003,50);b.write(0xd027,2);b.write(0xd028,3);
 b.state.ram[0x7f8]=0xc0;b.state.ram[0x7f9]=0xc1;b.state.ram[0x3000]=b.state.ram[0x3040]=255;
 at(b,50,55);assert(!v.ba&&v.aec);at(b,50,58);assert(!v.aec&&v.fetches[0].kind==VicAccess::Pointer&&v.fetches[1].address==0x3000);
 at(b,51,22);assert(v.pixels[51*504+24]==2&&v.collisionSprites==3&&v.collisionGraphics==3&&(v.flags&6)==6);
 // Next row, behind-foreground sprite 0 also occludes higher-numbered sprite 1.
 b.write(0xd01b,1);b.state.ram[0x2009]=255;b.state.ram[0x3003]=b.state.ram[0x3043]=255;at(b,52,22);assert(v.pixels[52*504+24]==5);
 // X/Y expanded multicolor sprite and re-use of its pointer each raster.
 setup(b);b.write(0xd015,1);b.write(0xd000,40);b.write(0xd001,50);b.write(0xd017,1);b.write(0xd01d,1);b.write(0xd01c,1);b.write(0xd025,2);b.write(0xd026,3);b.write(0xd027,4);b.state.ram[0x7f8]=0xc0;b.state.ram[0x3000]=0x6c;
 at(b,51,25);const uint8_t expanded[]={2,2,2,2,4,4,4,4,3,3,3,3,0,0,0,0};for(unsigned x=0;x<16;x++)assert(v.pixels[51*504+40+x]==expanded[x]);at(b,52,25);for(unsigned x=0;x<16;x++)assert(v.pixels[52*504+40+x]==expanded[x]);
 // Vertical border suppresses graphics but must not itself cover sprites when
 // the main border has been opened (the two flip-flops are independent).
 setup(b);v.raster=20;v.cycle=30;v.border=false;v.verticalBorder=true;v.sprites[0].display=true;v.sprites[0].pixel=0;v.sprites[0].shift=0xffffff;b.write(0xd000,132);b.write(0xd027,3);assert(b.tick());assert(v.pixels[20*504+132]==3);
 // A late CPU write drains during the BA warning, before AEC takes the bus.
 setup(b);at(b,51,11);Cpu cpu;cpu.start(0x300);cpu.state.a=0xa7;cpu.tick(0x8d);cpu.tick(0x00);cpu.tick(0x30);b.state.cpu=cpu.state;assert(b.tick());assert(!v.ba&&v.aec&&b.state.ram[0x3000]==0xa7&&b.state.lastBus.write&&b.state.lastBus.valid);
 at(b,51,15);assert(!b.state.lastBus.valid);
 // PLA BA gating prevents the early read from clearing CIA ICR while the
 // CPU is stalled. This is device selection, not a general RDY read cache.
 for(uint16_t address:{0xdc0d,0xdd0d}){
  setup(b);at(b,51,11);auto& cia=address==0xdc0d?b.state.cia1:b.state.cia2;cia.flags=1;
  b.state.ram[address]=0x5a;cpu.start(0x300);cpu.tick(0xad);cpu.tick(uint8_t(address));cpu.tick(uint8_t(address>>8));b.state.cpu=cpu.state;
  for(unsigned i=0;i<3;i++){assert(b.tick());assert(v.aec&&!v.ba&&cia.flags==1&&b.state.lastBus.value==0x5a&&b.state.cpu.stage==Stage::Read);}
  for(unsigned i=0;i<40;i++){assert(b.tick());assert(!v.aec&&cia.flags==1);}
  assert(b.tick());assert(v.ba&&cia.flags==0&&b.state.cpu.a==1&&b.state.cpu.stage==Stage::Fetch);
 }
 // Writes remain selected with BA low, while an unrelated external RDY
 // stall with BA high still performs real device reads.
 setup(b);at(b,51,11);cpu.start(0x300);cpu.state.a=0x1a;cpu.tick(0x8d);cpu.tick(0x00);cpu.tick(0xdd);b.state.cpu=cpu.state;
 assert(b.tick());assert(!v.ba&&v.aec&&b.state.cia2.pra==0x1a);
 setup(b);b.state.cia1.flags=1;cpu.start(0x300);cpu.tick(0xad);cpu.tick(0x0d);cpu.tick(0xdc);b.state.cpu=cpu.state;
 assert(b.tick(false));assert(v.ba&&b.state.cia1.flags==0&&b.state.cpu.stage==Stage::Read);
 // D018 writes between g-accesses change the next fetch, not earlier pixels.
 setup(b);b.state.ram[0x400]=b.state.ram[0x401]=1;b.state.color[0]=b.state.color[1]=5;b.state.ram[0x2008]=255;b.state.ram[0x2808]=0;at(b,51,16);b.write(0xd018,0x1a);at(b,51,22);assert(v.cells[0].graphics.address==0x2008&&v.cells[1].graphics.address==0x2808);assert(v.pixels[51*504+24]==5&&v.pixels[51*504+32]==0);
 // Board checkpoint includes the partial scanline and captured bytes.
 auto saved=std::make_unique<BoardState>(b.state);assert(b.run(500));auto expected=std::make_unique<BoardState>(b.state);b.state=*saved;assert(b.run(500));assert(b.state.vic.pixels==expected->vic.pixels&&b.state.cpu.pc==expected->cpu.pc&&b.state.vic.clocks==expected->vic.clocks);
 std::cout<<"PASS VIC: PAL bus schedule, badlines, five display modes, bank/ROM fetches, latched bytes, scrolling, raster IRQ, sprite DMA/priority/collisions/expansion, replay\n";
}
static void sid(){Sid s;s.write(14,0);s.write(15,0x80);s.write(18,0x20);for(unsigned i=0;i<4;i++)s.tick();assert(s.peek(27)==2);s.write(18,0x18);assert(s.voice[2].phase==0);for(unsigned i=0;i<100;i++)s.tick();assert(s.peek(27)==0);s.write(18,0x10);s.tick();assert(s.peek(27)==1);
 // Alstrup's measured post-TEST noise prefix: FE FC FC FC F8. Check a
 // fixed vector after the first shift, not another LFSR implementation.
 // The initial TEST discharge/shift timing remains a selected profile.
 s={};s.write(15,0x80);s.write(18,0x88);for(unsigned i=0;i<32768;i++)s.tick();s.write(18,0x80);assert(s.peek(27)==0xfe);
 const uint8_t prefix[]={0xfe,0xfc,0xfc,0xfc,0xf8};for(unsigned i=0;i<5;i++){for(unsigned c=0;c<(i?32:18);c++)s.tick();assert(s.peek(27)==prefix[i]);}
 // Silent oscillators still advance; pulse threshold, ring and sync readback.
 s={};s.write(15,0x80);for(unsigned i=0;i<10;i++)s.tick();assert(s.peek(27)==0);s.write(18,0x10);assert(s.peek(27)==10);
 s.write(17,8);s.write(18,0x40);assert(s.peek(27)==0);for(unsigned i=0;i<246;i++)s.tick();assert(s.peek(27)==255);
 s={};s.voice[1].phase=0x800000;s.write(18,0x14);assert(s.peek(27)==255);
 s={};s.voice[0].phase=0x12345;s.voice[2].phase=0x7f8000;s.write(15,0x80);s.write(4,2);s.tick();assert(s.voice[0].phase==0&&s.voice[2].phase==0x800000);
 // Attack, sustain, release and the rate-change delay bug.
 s={};s.write(19,0);s.write(20,0xf0);s.write(18,1);for(unsigned i=0;i<9*255;i++)s.tick();assert(s.peek(28)==255);for(unsigned i=0;i<100;i++)s.tick();assert(s.peek(28)==255);s.write(18,0);for(unsigned i=0;i<18;i++)s.tick();assert(s.peek(28)<255);
 s={};s.write(19,0xf0);s.write(18,1);for(unsigned i=0;i<100;i++)s.tick();s.write(19,0);for(unsigned i=0;i<100;i++)s.tick();assert(!s.peek(28));for(unsigned i=0;i<32767;i++)s.tick();assert(s.peek(28)>0);
 s={};s.write(15,0x80);s.write(18,0x81);for(unsigned i=0;i<117;i++)s.tick();const auto saved=s;uint32_t digest=0;for(unsigned i=0;i<1000;i++){s.tick();digest=digest*33+s.read(27)+s.read(28);}s=saved;uint32_t replay=0;for(unsigned i=0;i<1000;i++){s.tick();replay=replay*33+s.read(27)+s.read(28);}assert(digest==replay);
 std::cout<<"PASS SID: oscillator phase/TEST, noise reference prefix, pulse/ring/sync, ADSR and rate delay, readback and replay; digest="<<digest<<'\n';
}
int main(){video();sid();}
