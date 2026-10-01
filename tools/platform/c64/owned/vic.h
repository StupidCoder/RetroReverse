#pragma once
#include <array>
#include <span>
#include <cstdint>
namespace rr::c64 {
enum class VicAccess : uint8_t { Idle,Refresh,Matrix,Graphics,Pointer,Sprite };
struct VicFetch {
 uint64_t cycle=0;
 uint16_t address=0;
 uint8_t value=0,color=0,slot=0,phase=0;
 VicAccess kind=VicAccess::Idle;
 bool rom=false;
};
struct VicCell {VicFetch matrix{},graphics{};uint8_t code=0,color=0;};
struct VicSprite {
 std::array<VicFetch,3> bytes{};
 uint32_t shift=0;
 uint8_t pointer=0,mc=0,base=0,pixel=24,repeat=0;
 bool dma=false,display=false,advance=true,active=false;
};
struct Vic {
 static constexpr unsigned Width=504,Height=312,CyclesPerLine=63;
 std::array<uint8_t,64> regs{};
 std::array<uint8_t,Width*Height> pixels{}; // Palette indices in VIC X/raster coordinates.
 std::array<VicCell,40> cells{};
 std::array<VicSprite,8> sprites{};
 std::array<VicFetch,2> fetches{}; // This clock's actual phi1/phi2 bus fetches.
 uint64_t clocks=0,frames=0;
 uint16_t raster=0,compare=0,vc=0,base=0;
 uint8_t cycle=0,row=0,index=0,flags=0,mask=0,collisionSprites=0,collisionGraphics=0,refresh=255,fetchCount=0;
 bool den=false,bad=false,display=false,verticalBorder=true,border=true,ba=true,aec=true,irqMatch=false;
 uint8_t peek(uint8_t reg)const;
 uint8_t read(uint8_t reg);
 void write(uint8_t reg,uint8_t value);
 // tick prepares phi1/phi2 VIC accesses and the BA/AEC state for the CPU bus.
 void tick(std::span<const uint8_t,65536> ram,std::span<const uint8_t,4096> chars,std::span<const uint8_t,1024> color,uint16_t bank);
 bool irq()const{return flags&mask;}
private:
 void rasterIrq();
 void draw();
};
}
