#include "../board.h"
#include <cassert>
#include <cstdio>
#include <fstream>
#include <iostream>
#include <string>
using namespace rr::c64;
static void timerTests(){
 // Published CIA1TB123/CIA2TB123 timing table (Lorenz EDK 2.15).
 struct Row{uint8_t before,after;uint16_t samples[3];};
 const Row rows[]={{0,1,{10,10,9}},{0,16,{10,40,40}},{0,17,{10,40,40}},{1,17,{9,40,40}},{1,16,{9,40,40}},{1,0,{9,8,8}}};
 for(const auto& row:rows){Cia c;c.b.counter=10;c.b.latch=40;c.b.control=row.before;c.b.queue=row.before?3:0;c.write(15,row.after);for(const auto expected:row.samples){c.tick();assert(c.b.counter==expected);}}
 // Independent published cascaded timer/PB/ICR trace (CIA1TAB).
 Cia c;c.a.counter=c.a.latch=c.b.counter=c.b.latch=2;c.a.control=3;c.a.queue=3;c.b.control=0x47;c.b.toggle=true;c.ddrb=255;c.mask=2;
 const uint8_t expected[][4]={{1,2,0x80,0},{2,2,0xc0,1},{2,2,0x80,1},{1,1,0x80,1},{2,1,0xc0,1},{2,1,0x80,1},{1,0,0x80,1},{2,0,0xc0,1},{2,2,0,3},{1,2,0,0x83},{2,2,0x40,0x83},{2,2,0,0x83}};
 for(const auto& row:expected){c.tick();assert(c.a.counter==row[0]&&c.b.counter==row[1]&&c.peek(1)==row[2]&&c.peek(13)==row[3]);}
 const auto flags=c.peek(13);assert(flags==0x83&&c.peek(13)==flags);assert(c.read(13)==flags&&!c.irq&&c.peek(13)==0);
 // Same-cycle ICR read cancels the delayed interrupt, not the timer event readout.
 c={};c.a.counter=1;c.a.latch=2;c.a.control=1;c.a.queue=3;c.mask=1;c.tick();assert(c.read(13)==1);c.tick();assert(!c.irq);
 c={};c.write(13,0x90);c.tick(false);assert(c.peek(13)==16&&!c.irq);c.tick(false);assert(c.irq&&c.read(13)==0x90);c.tick(false);assert(c.peek(13)==0);c.tick(true);c.tick(false);assert(c.peek(13)==16);
 // Mask writes do not acknowledge an already latched IRQ.
 c.tick(false);c.write(13,16);assert(c.irq);c.read(13);assert(!c.irq);
 c={};c.a.counter=1;c.a.latch=2;c.a.control=9;c.a.queue=3;c.tick();assert(!(c.a.control&1));const auto stopped=c.a.counter;for(int i=0;i<20;i++)c.tick();assert(c.a.counter==stopped);
 // High latch writes reload only a stopped timer; low writes never reload.
 c={};c.write(4,7);assert(c.a.counter==0xffff);c.write(5,0x12);assert(c.a.counter==0x1207);c.write(14,1);c.write(5,0x34);assert(c.a.counter==0x1207&&c.a.latch==0x3407);
 // Edge-counted receive shift register and independent TOD latch/alarm.
 c={};for(unsigned bit=0;bit<8;bit++){const bool v=0xa6&(128>>bit);c.tick(true,false,v);c.tick(true,true,v);}assert(c.sdr==0xa6&&(c.peek(13)&8));
 c={};c.write(14,128);c.write(11,0x11);c.write(10,0x59);c.write(9,0x59);c.write(8,9);assert(c.read(11)==0x11);for(int i=0;i<5;i++)c.todEdge();assert(c.tod[3]==0x92&&c.tod[2]==0&&c.read(9)==0x59);assert(c.read(8)==9&&c.read(11)==0x92);c.read(8);
 c.write(15,128);c.write(11,0x92);c.write(10,0);c.write(9,0);c.write(8,1);c.write(15,0);for(int i=0;i<5;i++)c.todEdge();assert(c.flags&4);
 std::cout<<"PASS CIA: independent timer pipeline/cascade traces, ICR, FLAG, one-shot, latch, serial receive, TOD/alarm\n";
}
static void boardTests(){
 Board b;b.basic.fill(0xba);b.kernal.fill(0xce);b.chars.fill(0xcc);b.power();b.state.ram.fill(0x55);b.write(0,0x2f);
 for(uint8_t p=0;p<8;p++){b.write(1,p|0x30);assert(b.peek(0xa000)==((p&3)==3?0xba:0x55));assert(b.peek(0xe000)==((p&2)?0xce:0x55));if(p&3){if(!(p&4))assert(b.peek(0xd000)==0xcc);}else assert(b.peek(0xd000)==0x55);}
 b.write(1,0x37);b.write(0xa000,0x12);assert(b.peek(0xa000)==0xba&&b.state.ram[0xa000]==0x12);b.write(1,0x33);b.write(0xd000,0x23);assert(b.peek(0xd000)==0xcc&&b.state.ram[0xd000]==0x23);
 b.write(1,0x37);b.write(0xd800,0xfe);assert(b.state.color[0]==14&&(b.peek(0xd800)&15)==14);
 b.write(0xdc02,0xff);b.write(0xdc00,0xfd);b.key(1,2,true);assert(!(b.peek(0xdc01)&4));b.key(1,2,false);assert(b.peek(0xdc01)&4);
 b.state.joy1=16;assert(!(b.peek(0xdc01)&16));b.state.joy1=0;b.state.joy2=1;assert(!(b.peek(0xdc00)&1));b.state.joy2=0;
 // Reverse scan and ghosting propagate actual switch closures, not key codes.
 b.write(0xdc02,0);b.write(0xdc03,255);b.write(0xdc01,0xfb);b.key(1,2,true);assert(!(b.peek(0xdc00)&2));
 b.state.cia1.event(16);const auto event=b.state.cia1.flags;assert(b.peek(0xdc0d)==event&&b.state.cia1.flags==event);b.read(0xdc0d);assert(!b.state.cia1.flags);
 // Media-independent CPU loop for tape/timer progression and replay.
 b.kernal[0x1ffc]=0;b.kernal[0x1ffd]=2;b.power();b.state.ram[0x200]=0x4c;b.state.ram[0x201]=0;b.state.ram[0x202]=2;
 std::vector<uint8_t> tap(23);const char magic[]="C64-TAPE-RAW";for(unsigned i=0;i<12;i++)tap[i]=magic[i];tap[16]=3;tap[20]=1;tap[21]=2;tap[22]=3;
 assert(b.loadTape(tap));auto bad=tap;bad[16]=2;assert(!b.loadTape(bad)&&b.pulses.size()==3);
 b.write(0,0x2f);b.write(1,0x37);b.play(true);assert(!(b.port()&16));assert(b.run(100)&&b.state.tape.pulse==0);b.write(1,0x17);
 assert(b.run(7)&&b.state.tape.pulse==0);assert(b.run(1)&&b.state.tape.pulse==1&&(b.state.cia1.flags&16));b.state.cia1.read(13);
 const auto saved=b.state;assert(b.run(40)&&b.state.tape.pulse==3);const auto end=b.state;b.state=saved;assert(b.run(40));assert(b.state.cpu.pc==end.cpu.pc&&b.state.cpu.stage==end.cpu.stage&&b.state.tape.pulse==end.tape.pulse&&b.state.cia1.peek(13)==end.cia1.peek(13)&&b.state.cycles==end.cycles);
 assert(b.run(100)&&b.state.tape.pulse==3);b.resetCpu();assert(b.run(7)&&b.state.cpu.pc==0x200);
 // Consecutive one-cycle TAP v1 pulses are distinct falling edges.
 tap.resize(28);tap[12]=1;tap[16]=8;for(unsigned i=20;i<28;i++)tap[i]=0;tap[21]=tap[25]=1;
 assert(b.loadTape(tap));b.play(true);assert(b.run(1)&&b.state.tape.pulse==1);assert(b.state.cia1.read(13)&16);
 assert(b.run(1)&&b.state.tape.pulse==2);assert(b.state.cia1.read(13)&16);
 std::cout<<"PASS board: all banking combinations, RAM under ROM, color RAM, keyboard/joystick, safe peeks, motor/sense, TAP validation/end and replay\n";
}
static void load(const char* path,std::span<uint8_t> dst){std::ifstream f(path,std::ios::binary);assert(f);f.read(reinterpret_cast<char*>(dst.data()),dst.size());assert(f.gcount()==std::streamsize(dst.size()));assert(f.peek()==EOF);}
static std::string screen(const Board& b){std::string out;for(unsigned i=0;i<1000;i++){auto v=b.state.ram[0x400+i]&127;out+=char(v<32?v+64:v);}return out;}
static void type(Board& b,unsigned col,unsigned row){b.key(col,row,true);assert(b.run(100000));b.key(col,row,false);assert(b.run(100000));}
static void boot(const char* basic,const char* kernal,const char* chars){
 Board b;load(basic,b.basic);load(kernal,b.kernal);load(chars,b.chars);b.power();const bool running=b.run(3000000);
 if(!running||screen(b).find("READY.")==std::string::npos){std::fprintf(stderr,"Boot failed PC=%04x opcode=%02x cycles=%llu\n",b.state.cpu.pc,b.state.cpu.opcode,(unsigned long long)b.state.cycles);std::cerr<<screen(b)<<'\n';std::abort();}
 assert(screen(b).find("38911 BASIC BYTES FREE")!=std::string::npos);
 // Physical matrix: PRINT 2+2 followed by Return; no KERNAL hooks or buffer injection.
 for(const auto pair:std::array<std::array<unsigned,2>,10>{{{5,1},{2,1},{4,1},{4,7},{2,6},{7,4},{7,3},{5,0},{7,3},{0,1}}})type(b,pair[0],pair[1]);
 assert(b.run(300000));const auto output=screen(b);std::cout<<output.substr(0,400)<<'\n';assert(output.find("PRINT 2+2")!=std::string::npos);assert(output.find(" 4 ")!=std::string::npos);assert(!b.state.unimplementedSidRead);
 std::cout<<"PASS owned KERNAL/BASIC boot and physical-keyboard PRINT 2+2; PC="<<std::hex<<b.state.cpu.pc<<std::dec<<" cycles="<<b.state.cycles<<'\n';
}
int main(int argc,char** argv){timerTests();boardTests();if(argc==4)boot(argv[1],argv[2],argv[3]);}
