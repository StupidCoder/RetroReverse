#include "board.h"
#include <cstring>
namespace rr::c64 {
void Board::power(){state={};Cpu cpu;cpu.reset();state.cpu=cpu.state;if(!pulses.empty())state.tape.remaining=pulses[0];}
void Board::resetCpu(){Cpu cpu;cpu.state=state.cpu;cpu.reset();state.cpu=cpu.state;}
uint8_t Board::port()const{const uint8_t input=state.tape.play?7:0x17;return (state.data&state.ddr)|(input&uint8_t(~state.ddr));}
bool Board::ioVisible()const{const auto p=port();return (p&3)&&(p&4);}
std::array<uint8_t,2> Board::keyboard()const{
 uint8_t pa=state.cia1.outputA()&uint8_t(~state.joy2),pb=state.cia1.outputB()&uint8_t(~state.joy1);
 // Closed switches propagate a low level in either direction (including ghosting).
 for(unsigned iteration=0;iteration<8;iteration++){const auto oldA=pa,oldB=pb;for(unsigned column=0;column<8;column++){
  if(!(pa&(1<<column)))pb&=uint8_t(~state.keys[column]);if(state.keys[column]&uint8_t(~pb))pa&=uint8_t(~(1<<column));
 }if(oldA==pa&&oldB==pb)break;}return {pa,pb};
}
uint8_t Board::io(uint16_t address,bool effects){
 if(address<0xd400)return effects?state.vic.read(address&63):state.vic.peek(address&63);
 if(address<0xd800)return effects?state.sid.read(address&31):state.sid.peek(address&31);
 if(address<0xdc00)return (state.bus&0xf0)|state.color[address&1023];
 if(address<0xdd00){const auto pins=keyboard();return effects?state.cia1.read(address&15,pins[0],pins[1]):state.cia1.peek(address&15,pins[0],pins[1]);}
 if(address<0xde00)return effects?state.cia2.read(address&15):state.cia2.peek(address&15);
 return state.bus;
}
uint8_t Board::peek(uint16_t address)const{
 if(address==0)return state.ddr;if(address==1)return port();const auto p=port();
 if(address>=0xe000&&(p&2))return kernal[address&8191];
 if(address>=0xa000&&address<0xc000&&(p&3)==3)return basic[address&8191];
 if(address>=0xd000&&address<0xe000&&(p&3)){if(p&4)return const_cast<Board*>(this)->io(address,false);return chars[address&4095];}
 return state.ram[address];
}
uint8_t Board::read(uint16_t address){const auto value=address>=0xd000&&address<0xe000&&ioVisible()?io(address,true):peek(address);state.bus=value;return value;}
void Board::write(uint16_t address,uint8_t value){
 state.bus=value;if(address==0){state.ddr=value;return;}if(address==1){state.data=value;return;}
 if(address>=0xd000&&address<0xe000&&ioVisible()){
  if(address<0xd400){state.vic.write(address&63,value);return;}
  if(address<0xd800){state.sid.write(address&31,value);return;}
  if(address<0xdc00){state.color[address&1023]=value&15;return;}
  if(address<0xdd00){state.cia1.write(address&15,value);return;}
  if(address<0xde00){state.cia2.write(address&15,value);return;}return;
 }
 state.ram[address]=value;
}
bool Board::tick(bool cpuReady){
 Cpu cpu;cpu.state=state.cpu;if(cpu.faulted())return false;++state.cycles;state.vic.tick(state.ram,chars,state.color,uint16_t((~state.cia2.outputA()&3)<<14));state.sid.tick();
 state.tape.flag=true;
 if(state.tape.play&&motor()&&state.tape.pulse<pulses.size()){
  if(state.tape.remaining&&!--state.tape.remaining){state.tape.flag=false;if(++state.tape.pulse<pulses.size())state.tape.remaining=pulses[state.tape.pulse];}
 }
 state.cia1.tick();if(!state.tape.flag)state.cia1.event(16);state.cia2.tick();state.todPhase+=50;if(state.todPhase>=985248){state.todPhase-=985248;state.cia1.todEdge();state.cia2.todEdge();}
 const auto bus=cpu.bus();uint8_t data=state.bus;
 state.lastBus={state.cycles,bus.address,cpu.state.pc,0,state.vic.aec,bus.write,cpuReady&&state.vic.ba&&state.vic.aec,bus.sync};
 // AEC disconnects the CPU. BA gives it three clocks to finish writes first.
 for(unsigned i=0;i<state.vic.fetchCount;i++)state.bus=state.vic.fetches[i].value;
 if(state.vic.aec){data=bus.write?bus.data:read(bus.address);if(bus.write)write(bus.address,data);}

 state.lastBus.value=data;
 cpu.tick(data,state.cia1.irq||state.vic.irq(),state.cia2.irq||state.restore,cpuReady&&state.vic.ba&&state.vic.aec);state.cpu=cpu.state;return !cpu.faulted();
}
bool Board::run(uint64_t cycles){while(cycles--)if(!tick())return false;return true;}
void Board::key(unsigned column,unsigned row,bool down){if(column>=8||row>=8)return;if(down)state.keys[column]|=uint8_t(1<<row);else state.keys[column]&=uint8_t(~(1<<row));}
bool Board::loadTape(std::span<const uint8_t> bytes){
 if(bytes.size()<20||bytes.size()>2*1024*1024||std::memcmp(bytes.data(),"C64-TAPE-RAW",12)||bytes[12]>1||bytes[13]||bytes[14])return false;
 const uint32_t size=uint32_t(bytes[16])|(uint32_t(bytes[17])<<8)|(uint32_t(bytes[18])<<16)|(uint32_t(bytes[19])<<24);if(size!=bytes.size()-20||!size)return false;
 std::vector<uint32_t> next;for(size_t at=20;at<bytes.size();){uint32_t duration=bytes[at++]*8;if(!duration){if(!bytes[12])duration=2048;else{if(at+3>bytes.size())return false;duration=uint32_t(bytes[at])|(uint32_t(bytes[at+1])<<8)|(uint32_t(bytes[at+2])<<16);at+=3;}}if(!duration)return false;next.push_back(duration);}
 pulses=std::move(next);state.tape={};state.tape.remaining=pulses[0];return true;
}
}
