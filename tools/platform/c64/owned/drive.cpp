#include "drive.h"
namespace rr::c64 {
void Drive::power(){auto media=std::move(state.media);state={};state.media=std::move(media);Cpu cpu;cpu.reset();state.cpu=cpu.state;}
bool Drive::mount(std::span<const uint8_t> image,bool protect){if(!state.media.mount(image,protect))return false;state.mediaChange=200000;return true;}
void Drive::eject(){state.media={};state.mediaChange=200000;}
uint8_t Drive::peek(uint16_t address)const{
 if(address&0x8000)return rom[address&0x3fff];address&=0x1fff;
 if(address<0x800)return state.ram[address];if(address>=0x1800&&address<0x1c00)return state.serial.peek(address&15);if(address>=0x1c00)return state.disk.peek(address&15);return state.bus;
}
uint8_t Drive::read(uint16_t address){auto value=peek(address);if(!(address&0x8000)){address&=0x1fff;if(address>=0x1800&&address<0x1c00)value=state.serial.read(address&15);if(address>=0x1c00)value=state.disk.read(address&15);}state.bus=value;return value;}
void Drive::write(uint16_t address,uint8_t value){state.bus=value;if(address&0x8000)return;address&=0x1fff;if(address<0x800)state.ram[address]=value;else if(address>=0x1800&&address<0x1c00)state.serial.write(address&15,value);else if(address>=0x1c00)state.disk.write(address&15,value);}
bool Drive::clockPull()const{return state.serial.orb&state.serial.ddrb&8;}
bool Drive::dataPull(bool atn)const{const auto out=state.serial.orb&state.serial.ddrb;return (out&2)||(bool(out&16)!=!atn);}
void Drive::mechanics(){
 auto& s=state;const auto out=s.disk.orb&s.disk.ddrb;const unsigned phase=out&3;
 if(phase!=s.phase){const unsigned delta=(phase-s.phase)&3;if(delta==1&&s.halfTrack<85){++s.halfTrack;++s.steps;}if(delta==3&&s.halfTrack>2){--s.halfTrack;++s.steps;}s.phase=uint8_t(phase);}
 s.motor=out&4;s.led=out&8;s.byteReady=false;if(s.mediaChange)--s.mediaChange;
 const bool writing=!s.disk.cb2Out;if(writing!=s.writing){s.writing=writing;s.bits=0;s.ones=0;s.sync=false;}
 if(!s.motor)return;s.rotation=(s.rotation+1)%200000;
 const unsigned h=s.halfTrack;const auto& bytes=s.media.tracks[h-2];
 if(h!=s.lastHalf){const auto old=s.media.tracks[s.lastHalf-2].size();s.position=old?uint32_t(uint64_t(s.position)*bytes.size()/old):0;s.lastHalf=uint8_t(h);s.mediaPhase=0;}
 if(!bytes.empty())s.position%=uint32_t(bytes.size()*8);else s.position=0;
 // GCR images carry recovered bit cells, not flux transitions. Read timing uses
 // their recorded zone; analog PLL/weak-bit behavior is outside this model.
 const unsigned recorded=bytes.empty()?16:16-s.media.speeds[h-2][s.position/8];
 s.mediaPhase+=4;const bool advance=s.mediaPhase>=recorded;if(advance){s.mediaPhase=uint8_t(s.mediaPhase-recorded);if(!bytes.empty())s.position=(s.position+1)%uint32_t(bytes.size()*8);}
 s.bitPhase+=4;const unsigned period=16-((out>>5)&3);const bool writeClock=s.bitPhase>=period;if(writeClock)s.bitPhase=uint8_t(s.bitPhase-period);
 if(s.writing?!writeClock:!advance)return;
 if(s.writing){if(!s.bits)s.writeShift=s.disk.ora;const bool bit=s.writeShift&128;s.writeShift<<=1;
  if(s.media.present()&&!s.mediaChange&&!s.media.writeProtected){s.media.writeHalfBit(h,s.position,bit);++s.writtenBits;}
  if(++s.bits==8){s.bits=0;s.byteReady=s.disk.ca2Out;++s.bytes;}
 }else{
  const bool bit=s.media.present()&&!s.mediaChange&&s.media.halfBit(h,s.position);s.readShift=uint8_t((s.readShift<<1)|bit);
  s.ones=bit?uint8_t(s.ones<10?s.ones+1:10):0;s.sync=s.ones>=10;
  if(s.sync)s.bits=0;else if(++s.bits==8){s.bits=0;s.byteReady=s.disk.ca2Out;++s.bytes;}
 }
}
bool Drive::tick(IecLines lines){
 Cpu cpu;cpu.state=state.cpu;if(cpu.faulted())return false;++state.clocks;mechanics();
 const uint8_t serialInputs=uint8_t(0x1a|((state.device-8)&3)<<5|(!lines.data?1:0)|(!lines.clock?4:0)|(!lines.atn?128:0));
 state.serial.tick(255,serialInputs,!lines.atn);
 // Insertion/removal passes the jacket across the optical sensor, then
 // exposes it before the new notch settles. Even protected-to-protected
 // changes must produce edges, so DOS invalidates its cached BAM/directory.
 const bool light=state.mediaChange?(state.mediaChange<=100000):(!state.media.present()||!state.media.writeProtected);
 const uint8_t diskInputs=uint8_t(0x6f|(light?16:0)|(state.sync?0:128));
 state.disk.tick(state.readShift,diskInputs,!state.byteReady);
 const auto bus=cpu.bus();state.lastBus=bus;const auto data=bus.write?bus.data:read(bus.address);if(bus.write)write(bus.address,data);state.lastBus.data=data;
 cpu.tick(data,state.serial.irq()||state.disk.irq(),false,true,state.byteReady);state.cpu=cpu.state;return !cpu.faulted();
}
}
