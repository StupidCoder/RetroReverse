#pragma once
#include <array>
#include <cstdint>
#include <span>
#include <vector>
namespace rr::c64 {
// Mutable encoded media is value-owned snapshot state. The host's source file
// is never modified. D64 export validates every normal header/data checksum.
struct Disk {
 std::array<std::vector<uint8_t>,84> tracks;
 std::array<bool,84> dirty{};
 std::array<std::vector<uint8_t>,84> speeds; // One recorded speed zone per byte.
 unsigned trackCount=0;
 bool g64=false;
 bool writeProtected=true;
 static unsigned sectors(unsigned track);
 static unsigned offset(unsigned track,unsigned sector);
 static unsigned density(unsigned track);
 bool mount(std::span<const uint8_t> image,bool protect=true);
 bool mountG64(std::span<const uint8_t> image,bool protect=true);
 bool exportG64(std::vector<uint8_t>& image)const;
 bool halfBit(unsigned halfTrack,uint32_t position)const;
 void writeHalfBit(unsigned halfTrack,uint32_t position,bool value);
 bool exportD64(std::vector<uint8_t>& image)const;
 bool present()const{return trackCount!=0;}
 bool changed()const;
 bool bit(unsigned track,uint32_t position)const;
 void writeBit(unsigned track,uint32_t position,bool value);
};
}
