#include "../system.h"
#include <algorithm>
#include <cassert>
#include <iostream>
#include <memory>
using namespace rr::c64;
static void viaTests(){
 Via v;v.write(3,0xf0);v.write(15,0x50);v.tick(0xa6);assert(v.peek(15)==0x56);
 v.write(14,0xc2);v.write(4,2);v.write(5,0);for(int i=0;i<3;i++){v.tick();assert(!v.irq());}v.tick();assert(v.peek(13)==0xc0);assert(v.read(4)==255&&!v.irq());
 for(unsigned i=0;i<65540;i++)v.tick();assert(!(v.ifr&64)); // One-shot does not retrigger.
 v={};v.write(11,0xc0);v.write(4,1);v.write(5,0);v.tick();v.tick();v.tick();assert(v.t1Out&&(v.ifr&64));v.read(4);v.tick();v.tick();v.tick();assert(!v.t1Out&&(v.ifr&64));
 v={};v.write(11,32);v.write(8,1);v.write(9,0);v.tick();for(int i=0;i<10;i++)v.tick();assert(v.t2==1);v.tick(255,0xbf);assert(v.t2==0&&!(v.ifr&32));v.tick();v.tick(255,0xbf);assert(v.ifr&32);v.read(8);assert(!(v.ifr&32));
 v={};v.write(11,1);v.write(14,0x82);v.tick(0x36,255,false);assert(v.irq());v.tick(0x92,255,false);assert(v.peek(1)==0x36);assert(v.read(15)==0x36&&v.irq());v.read(1);assert(!v.irq());
 v={};v.write(12,2);v.tick(255,255,true,false);assert(v.ifr&1);v.read(1);assert(v.ifr&1);v.write(13,1);assert(!v.ifr); // Independent CA2 flag.
 v.write(12,8);v.write(1,0);assert(!v.ca2Out);v.tick(255,255,false);assert(v.ca2Out);
 v.write(12,10);v.read(1);assert(!v.ca2Out);v.tick();assert(v.ca2Out);
 v.write(12,0xcc);assert(!v.ca2Out&&!v.cb2Out);v.write(12,0xee);assert(v.ca2Out&&v.cb2Out);
 v={};v.write(11,12);v.write(10,0);for(unsigned i=0;i<8;i++){bool b=0xa6&(128>>i);v.tick(255,255,true,true,false,b);v.tick(255,255,true,true,true,b);}assert(v.sr==0xa6&&(v.ifr&4));v.read(10);assert(!(v.ifr&4));
 v.write(14,0x84);v.ifr|=4;assert(v.irq());v.write(14,4);assert(!v.irq()&&(v.ifr&4)&&v.peek(14)==128);
 Cpu c;c.start(0x200);c.state.p&=uint8_t(~64);c.tick(0xea,false,false,false,true);assert(c.state.p&64);c.state.p&=uint8_t(~64);c.tick(0xea,false,false,false,true);assert(!(c.state.p&64));c.tick(0xea,false,false,false,false);c.tick(0xea,false,false,false,true);assert(c.state.p&64);
 std::cout<<"PASS VIA timers, PB7, pulse counting, latches, IRQ acknowledgement, handshakes, external shift and CPU SO edges\n";
}
static void diskTests(){
 for(unsigned length:{174848u,196608u}){std::vector<uint8_t> image(length);for(unsigned i=0;i<length;i++)image[i]=uint8_t(i*23+(i>>8));Disk d;assert(d.mount(image));std::vector<uint8_t> out;assert(d.exportD64(out)&&out==image);
  const bool old=d.bit(18,19);d.writeBit(18,19,!old);assert(d.bit(18,19)==old&&!d.changed());d.writeProtected=false;d.writeBit(18,19,!old);assert(d.bit(18,19)!=old&&d.changed());d.writeBit(18,19,old);assert(d.exportD64(out)&&out==image);
  auto before=d.tracks;assert(!d.mount(std::span(image).first(100))&&d.tracks==before);
  assert(d.exportG64(out));Disk g;assert(g.mount(out)&&g.g64&&g.tracks==d.tracks&&g.speeds==d.speeds);assert(!g.exportD64(out));
 }
 Disk d;assert(d.mount(std::vector<uint8_t>(174848),false));d.tracks[1]={0x81,0x42,0x24,0x18,0xff};d.speeds[1]={0,1,2,3,2};std::vector<uint8_t> encoded;assert(d.exportG64(encoded));Disk g;assert(g.mount(encoded,false)&&g.tracks[1]==d.tracks[1]&&g.speeds[1]==d.speeds[1]);assert(g.halfBit(3,0));g.writeHalfBit(3,0,false);assert(!g.halfBit(3,0)&&g.dirty[1]&&!g.dirty[0]);
 for(unsigned cut:{0u,11u,683u,unsigned(encoded.size()-1)})assert(!g.mount(std::span(encoded).first(cut)));
 auto invalid=encoded;invalid[8]=1;assert(!g.mount(invalid));invalid=encoded;invalid[12]=255;invalid[13]=255;invalid[14]=255;invalid[15]=255;assert(!g.mount(invalid));assert(!g.halfBit(85,0));g.writeHalfBit(85,0,true); // Empty track is safe.
 std::cout<<"PASS D64 35/40-track roundtrip, G64 raw half-tracks and speed maps, malformed mounts, protection and local dirty media\n";
}
static void systemTests(){
 auto p=std::make_unique<System>();auto& s=*p;s.board.kernal[0x1ffc]=0;s.board.kernal[0x1ffd]=2;s.drive.rom[0x3ffc]=0;s.drive.rom[0x3ffd]=2;s.power();
 for(auto* ram:{s.board.state.ram.data(),s.drive.state.ram.data()}){ram[0x200]=0x4c;ram[0x201]=0;ram[0x202]=2;}
 assert(s.run(985248));assert(s.drive.state.clocks==1000000);s.board.state.cia2.ddra=0x38;s.board.state.cia2.pra=8;assert(s.tick());assert(!s.iec.lines.atn&&!s.iec.lines.data&&s.iec.lines.clock); // Hardware ATN automatic acknowledgement.
 s.drive.state.serial.ddrb=0x1a;s.drive.state.serial.orb=0x10;assert(s.tick());assert(s.iec.lines.data);s.board.state.cia2.pra=0x28;assert(s.tick());assert(!s.iec.lines.data);s.drive.state.serial.orb=0x18;assert(s.tick());assert(!s.iec.lines.clock);assert(s.tick());assert(!(s.board.peek(0xdd00)&0xc0));
 assert(s.drive.mount(std::vector<uint8_t>(174848),false));s.drive.state.disk.ddrb=0x6f;s.drive.state.disk.orb=0x64;s.drive.state.disk.write(12,0xee);assert(s.run(300000));assert(s.drive.state.bytes&&s.drive.state.motor);
 auto saved=std::make_unique<SystemSnapshot>(s.save());assert(s.run(20000));const auto expected=s.drive.state;const auto iec=s.iec;s.restore(*saved);assert(s.run(20000));assert(s.drive.state.cpu.pc==expected.cpu.pc&&s.drive.state.position==expected.position&&s.drive.state.bits==expected.bits&&s.drive.state.bytes==expected.bytes&&s.iec.transitions==iec.transitions);
 s.drive.state.disk.write(12,0xce);s.drive.state.disk.ora=0;assert(s.run(2000));assert(s.drive.state.media.changed());*saved=s.save();const auto written=s.drive.state.media.tracks;s.drive.state.media.writeHalfBit(36,0,!s.drive.state.media.halfBit(36,0));s.restore(*saved);assert(s.drive.state.media.tracks==written&&s.drive.state.media.changed());
 s.drive.eject();assert(!s.drive.state.media.present());s.restore(*saved);assert(s.drive.state.media.present());
 std::cout<<"PASS independent PAL/1MHz clocks, IEC wired-AND and ATN acknowledgement, rotating media, in-flight replay and dirty-media snapshots\n";
}
int main(){viaTests();diskTests();systemTests();}
