#pragma once
#include "../system.h"
#include <cassert>
#include <fstream>
#include <iostream>
#include <memory>
#include <string>
using namespace rr::c64;
inline void load(const char* path,std::span<uint8_t> out){std::ifstream f(path,std::ios::binary);assert(f);f.read(reinterpret_cast<char*>(out.data()),out.size());assert(f.gcount()==std::streamsize(out.size())&&f.peek()==EOF);}
inline void status(const System& s,const char* name){const auto& d=s.drive.state;std::cerr<<name<<" c64="<<std::hex<<s.board.state.cpu.pc<<" drive="<<d.cpu.pc<<" p="<<unsigned(d.cpu.p)<<" via="<<unsigned(d.serial.orb)<<'/'<<unsigned(d.serial.ddrb)<<" disk="<<unsigned(d.disk.orb)<<'/'<<unsigned(d.disk.pcr)<<" job="<<unsigned(d.ram[0])<<" track="<<unsigned(d.ram[0x22])<<std::dec<<" half="<<unsigned(d.halfTrack)<<" bytes="<<d.bytes<<" steps="<<d.steps<<" cycles="<<s.board.state.cycles<<" IEC="<<s.iec.lines.atn<<s.iec.lines.clock<<s.iec.lines.data<<'\n';}
inline void run(System& s,unsigned n){if(!s.run(n)){status(s,"FAULT");std::abort();}}
inline std::string screen(const System& s){std::string out;for(unsigned i=0;i<1000;i++){auto v=s.board.state.ram[0x400+i]&127;out+=char(v<32?v+64:v);}return out;}
inline void type(System& s,const char* text){
 const char* rows[]={"\x14\r\x1d\x88\x85\x86\x87\x11","3WA4ZSE\x01","5RD6CFTX","7YG8BHUV","9IJ0MKON","+PL-.:@,","\x1c*;\x13\x02=\x1e/",("1\x1f\x03" "2 \x04Q\x05")};
 // Explicit columns/rows for the subset used by acceptance commands.
 for(;*text;text++){
  unsigned col=0,row=0;bool shift=false;char ch=*text;if(ch=='"'){ch='2';shift=true;}if(ch=='$'){ch='4';shift=true;}
  bool found=false;for(unsigned c=0;c<8&&!found;c++)for(unsigned r=0;r<8;r++)if(rows[c][r]==ch){col=c;row=r;found=true;break;}assert(found);
  if(shift)s.board.key(1,7,true);s.board.key(col,row,true);run(s,100000);s.board.key(col,row,false);if(shift)s.board.key(1,7,false);run(s,100000);
 }
}
