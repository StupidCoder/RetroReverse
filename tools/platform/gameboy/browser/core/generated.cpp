#include "runtime.h"
struct sm83_CPU;
struct gameboy_Machine;
struct Anon0;
struct Anon1;
struct sm83_CPU{
uint8_t A{};
uint8_t F{};
uint8_t B{};
uint8_t C{};
uint8_t D{};
uint8_t E{};
uint8_t H{};
uint8_t L{};
uint16_t SP{};
uint16_t PC{};
bool IME{};
bool imeEnable{};
bool Halted{};
std::string HaltReason{};
bool halt{};
bool Stopped{};
uint64_t Instrs{};
gameboy_Machine* bus{};
int64_t takenExtra{};
};
struct gameboy_Machine{
sm83_CPU* CPU{};
Slice<uint8_t> rom{};
int64_t nbanks{};
int64_t romBank{};
int64_t ramBank{};
bool ramEnable{};
uint8_t mode{};
std::array<uint8_t,8192> vram{};
std::array<uint8_t,8192> wram{};
std::array<uint8_t,160> oam{};
std::array<uint8_t,127> hram{};
std::array<uint8_t,128> io{};
std::array<uint8_t,32768> extram{};
uint8_t ie{};
int64_t divCounter{};
int64_t timaCounter{};
int64_t lcdDot{};
int64_t Cycles{};
uint8_t Buttons{};
bool Sample{};
Map<uint16_t,int64_t> PCHist{};
uint16_t WatchLo{};
uint16_t WatchHi{};
Map<uint16_t,int64_t> WatchPCs{};
std::function<void(uint16_t,uint16_t,uint8_t)> OnWrite{};
std::function<void(uint16_t,int64_t,uint16_t)> OnROMRead{};
};
struct Anon0{uint8_t A{};uint8_t F{};uint8_t B{};uint8_t C{};uint8_t D{};uint8_t E{};uint8_t H{};uint8_t L{};uint16_t SP{};uint16_t PC{};bool IME{};bool imeEnable{};bool Halted{};std::string HaltReason{};bool halt{};bool Stopped{};uint64_t Instrs{};gameboy_Machine* bus{};int64_t takenExtra{};};
struct Anon1{sm83_CPU* CPU{};Slice<uint8_t> rom{};int64_t nbanks{};int64_t romBank{};int64_t ramBank{};bool ramEnable{};uint8_t mode{};std::array<uint8_t,8192> vram{};std::array<uint8_t,8192> wram{};std::array<uint8_t,160> oam{};std::array<uint8_t,127> hram{};std::array<uint8_t,128> io{};std::array<uint8_t,32768> extram{};uint8_t ie{};int64_t divCounter{};int64_t timaCounter{};int64_t lcdDot{};int64_t Cycles{};uint8_t Buttons{};bool Sample{};Map<uint16_t,int64_t> PCHist{};uint16_t WatchLo{};uint16_t WatchHi{};Map<uint16_t,int64_t> WatchPCs{};std::function<void(uint16_t,uint16_t,uint8_t)> OnWrite{};std::function<void(uint16_t,int64_t,uint16_t)> OnROMRead{};};

#include "adapters-decl.h"
sm83_CPU* sm83_NewCPU(gameboy_Machine* bus);
void sm83_CPU_Reset(sm83_CPU* c);
uint8_t sm83_CPU_read(sm83_CPU* c,uint16_t a);
void sm83_CPU_write(sm83_CPU* c,uint16_t a,uint8_t v);
uint16_t sm83_CPU_read16(sm83_CPU* c,uint16_t a);
void sm83_CPU_write16(sm83_CPU* c,uint16_t a,uint16_t v);
uint8_t sm83_CPU_fetch(sm83_CPU* c);
uint16_t sm83_CPU_fetch16(sm83_CPU* c);
void sm83_CPU_push16(sm83_CPU* c,uint16_t v);
uint16_t sm83_CPU_pop16(sm83_CPU* c);
uint16_t sm83_CPU_bc(sm83_CPU* c);
uint16_t sm83_CPU_de(sm83_CPU* c);
uint16_t sm83_CPU_hl(sm83_CPU* c);
uint16_t sm83_CPU_af(sm83_CPU* c);
void sm83_CPU_setBC(sm83_CPU* c,uint16_t v);
void sm83_CPU_setDE(sm83_CPU* c,uint16_t v);
void sm83_CPU_setHL(sm83_CPU* c,uint16_t v);
void sm83_CPU_setAF(sm83_CPU* c,uint16_t v);
uint16_t sm83_CPU_getRP(sm83_CPU* c,int64_t p,bool af);
void sm83_CPU_setRP(sm83_CPU* c,int64_t p,uint16_t v,bool af);
uint8_t sm83_CPU_getR(sm83_CPU* c,int64_t i);
void sm83_CPU_setR(sm83_CPU* c,int64_t i,uint8_t v);
bool sm83_CPU_getf(sm83_CPU* c,uint8_t b);
void sm83_CPU_setFlags(sm83_CPU* c,bool z,bool n,bool h,bool cy);
void sm83_CPU_add8(sm83_CPU* c,uint8_t n,bool carry);
void sm83_CPU_sub8(sm83_CPU* c,uint8_t n,bool carry,bool store);
void sm83_CPU_and8(sm83_CPU* c,uint8_t n);
void sm83_CPU_or8(sm83_CPU* c,uint8_t n);
void sm83_CPU_xor8(sm83_CPU* c,uint8_t n);
void sm83_CPU_alu(sm83_CPU* c,int64_t y,uint8_t n);
uint8_t sm83_CPU_inc8(sm83_CPU* c,uint8_t v);
uint8_t sm83_CPU_dec8(sm83_CPU* c,uint8_t v);
uint16_t sm83_CPU_add16(sm83_CPU* c,uint16_t a,uint16_t b);
uint16_t sm83_CPU_addSP(sm83_CPU* c,uint8_t e);
uint8_t sm83_CPU_rot(sm83_CPU* c,int64_t y,uint8_t v);
int64_t sm83_CPU_Step(sm83_CPU* c);
void sm83_CPU_execMain(sm83_CPU* c,uint8_t op);
void sm83_CPU_execX0(sm83_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q);
void sm83_CPU_accOp(sm83_CPU* c,int64_t y);
void sm83_CPU_daa(sm83_CPU* c);
void sm83_CPU_execX3(sm83_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q);
bool sm83_CPU_cond(sm83_CPU* c,int64_t y);
void sm83_CPU_execCB(sm83_CPU* c,uint8_t op);
int64_t sm83_cbCycles(uint8_t op);
gameboy_Machine* gameboy_NewMachine(Slice<uint8_t> rom);
void gameboy_Machine_Watch(gameboy_Machine* m,uint16_t lo,uint16_t hi);
Slice<uint8_t> gameboy_Machine_VRAM(gameboy_Machine* m);
Slice<uint8_t> gameboy_Machine_OAM(gameboy_Machine* m);
Slice<uint8_t> gameboy_Machine_WRAM(gameboy_Machine* m);
int64_t gameboy_Machine_ROMBank(gameboy_Machine* m);
uint8_t gameboy_Machine_Read(gameboy_Machine* m,uint16_t a);
uint8_t gameboy_Machine_Read_Reference(gameboy_Machine* m,uint16_t a);
void gameboy_Machine_Write(gameboy_Machine* m,uint16_t a,uint8_t v);
int64_t gameboy_Machine_ramOff(gameboy_Machine* m,uint16_t a);
void gameboy_Machine_mbcWrite(gameboy_Machine* m,uint16_t a,uint8_t v);
void gameboy_Machine_mbcWrite_Reference(gameboy_Machine* m,uint16_t a,uint8_t v);
uint8_t gameboy_Machine_readIO(gameboy_Machine* m,uint16_t a);
void gameboy_Machine_writeIO(gameboy_Machine* m,uint16_t a,uint8_t v);
void gameboy_Machine_writeIO_Reference(gameboy_Machine* m,uint16_t a,uint8_t v);
uint8_t gameboy_Machine_joyp(gameboy_Machine* m);
void gameboy_Machine_tick(gameboy_Machine* m,int64_t cyc);
void gameboy_Machine_tick_Reference(gameboy_Machine* m,int64_t cyc);
void gameboy_Machine_reqInt(gameboy_Machine* m,uint8_t bit);
int64_t gameboy_Machine_Step(gameboy_Machine* m);
bool gameboy_Machine_RunFrame(gameboy_Machine* m);
int64_t gameboy_Machine_RunFrames(gameboy_Machine* m,int64_t n);
constexpr int64_t sm83_flagC=16ULL;
constexpr int64_t sm83_flagH=32ULL;
constexpr int64_t sm83_flagN=64ULL;
constexpr int64_t sm83_flagZ=128ULL;
constexpr int64_t sm83_regIF=65295ULL;
constexpr int64_t sm83_regIE=65535ULL;
std::array<uint8_t,256> sm83_mainCycles=std::array<uint8_t,256>{cast<uint8_t>(4ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(20ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(8ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(4ULL),cast<uint8_t>(12ULL),cast<uint8_t>(24ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(0ULL),cast<uint8_t>(12ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(0ULL),cast<uint8_t>(12ULL),cast<uint8_t>(0ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(16ULL),cast<uint8_t>(4ULL),cast<uint8_t>(16ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(4ULL),cast<uint8_t>(0ULL),cast<uint8_t>(16ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(12ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL),cast<uint8_t>(4ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(8ULL),cast<uint8_t>(16ULL)};
constexpr int64_t gameboy_regJOYP=65280ULL;
constexpr int64_t gameboy_regDIV=65284ULL;
constexpr int64_t gameboy_regTIMA=65285ULL;
constexpr int64_t gameboy_regTMA=65286ULL;
constexpr int64_t gameboy_regTAC=65287ULL;
constexpr int64_t gameboy_regIF=65295ULL;
constexpr int64_t gameboy_regLCDC=65344ULL;
constexpr int64_t gameboy_regSTAT=65345ULL;
constexpr int64_t gameboy_regLY=65348ULL;
constexpr int64_t gameboy_regLYC=65349ULL;
constexpr int64_t gameboy_regDMA=65350ULL;
constexpr int64_t gameboy_regIE=65535ULL;
constexpr int64_t gameboy_BtnRight=1ULL;
constexpr int64_t gameboy_BtnLeft=2ULL;
constexpr int64_t gameboy_BtnUp=4ULL;
constexpr int64_t gameboy_BtnDown=8ULL;
constexpr int64_t gameboy_BtnA=16ULL;
constexpr int64_t gameboy_BtnB=32ULL;
constexpr int64_t gameboy_BtnSelect=64ULL;
constexpr int64_t gameboy_BtnStart=128ULL;
constexpr int64_t gameboy_dotsPerLine=456ULL;
constexpr int64_t gameboy_linesPerFrame=154ULL;
constexpr int64_t gameboy_cyclesPerFrame=70224ULL;

#include "adapters.h"
// tools/cpu/sm83/cpu.go:56:1
sm83_CPU* sm83_NewCPU(gameboy_Machine* bus){
{
sm83_CPU* c = arenaNew(sm83_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},bus,{}});
sm83_CPU_Reset(c);
return c;
}
}
// tools/cpu/sm83/cpu.go:63:1
void sm83_CPU_Reset(sm83_CPU* c){
{
auto tmp1 = std::make_tuple(cast<uint8_t>(1ULL),cast<uint8_t>(176ULL));
c->A = std::get<0>(tmp1);
c->F = std::get<1>(tmp1);
auto tmp2 = std::make_tuple(cast<uint8_t>(0ULL),cast<uint8_t>(19ULL));
c->B = std::get<0>(tmp2);
c->C = std::get<1>(tmp2);
auto tmp3 = std::make_tuple(cast<uint8_t>(0ULL),cast<uint8_t>(216ULL));
c->D = std::get<0>(tmp3);
c->E = std::get<1>(tmp3);
auto tmp4 = std::make_tuple(cast<uint8_t>(1ULL),cast<uint8_t>(77ULL));
c->H = std::get<0>(tmp4);
c->L = std::get<1>(tmp4);
auto tmp5 = std::make_tuple(cast<uint16_t>(65534ULL),cast<uint16_t>(256ULL));
c->SP = std::get<0>(tmp5);
c->PC = std::get<1>(tmp5);
auto tmp6 = std::make_tuple(false,false);
c->IME = std::get<0>(tmp6);
c->imeEnable = std::get<1>(tmp6);
auto tmp7 = std::make_tuple(false,false,false);
c->halt = std::get<0>(tmp7);
c->Stopped = std::get<1>(tmp7);
c->Halted = std::get<2>(tmp7);
}
}
// tools/cpu/sm83/cpu.go:79:1
uint8_t sm83_CPU_read(sm83_CPU* c,uint16_t a){
{
return gameboy_Machine_Read(c->bus,a);
}
}
// tools/cpu/sm83/cpu.go:80:1
void sm83_CPU_write(sm83_CPU* c,uint16_t a,uint8_t v){
{
gameboy_Machine_Write(c->bus,a,v);
}
}
// tools/cpu/sm83/cpu.go:81:1
uint16_t sm83_CPU_read16(sm83_CPU* c,uint16_t a){
{
return cast<uint16_t>((cast<uint16_t>(sm83_CPU_read(c,a)) | shl<uint16_t>(cast<uint16_t>(sm83_CPU_read(c,cast<uint16_t>((a + cast<uint16_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/cpu/sm83/cpu.go:82:1
void sm83_CPU_write16(sm83_CPU* c,uint16_t a,uint16_t v){
{
sm83_CPU_write(c,a,cast<uint8_t>(v));
sm83_CPU_write(c,cast<uint16_t>((a + cast<uint16_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/cpu/sm83/cpu.go:83:1
uint8_t sm83_CPU_fetch(sm83_CPU* c){
{
uint8_t v = sm83_CPU_read(c,c->PC);
c->PC++;
return v;
}
}
// tools/cpu/sm83/cpu.go:84:1
uint16_t sm83_CPU_fetch16(sm83_CPU* c){
{
uint16_t v = sm83_CPU_read16(c,c->PC);
c->PC += cast<uint16_t>(2ULL);
return v;
}
}
// tools/cpu/sm83/cpu.go:85:1
void sm83_CPU_push16(sm83_CPU* c,uint16_t v){
{
c->SP -= cast<uint16_t>(2ULL);
sm83_CPU_write16(c,c->SP,v);
}
}
// tools/cpu/sm83/cpu.go:86:1
uint16_t sm83_CPU_pop16(sm83_CPU* c){
{
uint16_t v = sm83_CPU_read16(c,c->SP);
c->SP += cast<uint16_t>(2ULL);
return v;
}
}
// tools/cpu/sm83/cpu.go:90:1
uint16_t sm83_CPU_bc(sm83_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->B),cast<int64_t>(8ULL)) | cast<uint16_t>(c->C)));
}
}
// tools/cpu/sm83/cpu.go:91:1
uint16_t sm83_CPU_de(sm83_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->D),cast<int64_t>(8ULL)) | cast<uint16_t>(c->E)));
}
}
// tools/cpu/sm83/cpu.go:92:1
uint16_t sm83_CPU_hl(sm83_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->H),cast<int64_t>(8ULL)) | cast<uint16_t>(c->L)));
}
}
// tools/cpu/sm83/cpu.go:93:1
uint16_t sm83_CPU_af(sm83_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->A),cast<int64_t>(8ULL)) | cast<uint16_t>(c->F)));
}
}
// tools/cpu/sm83/cpu.go:94:1
void sm83_CPU_setBC(sm83_CPU* c,uint16_t v){
{
auto tmp8 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->B = std::get<0>(tmp8);
c->C = std::get<1>(tmp8);
}
}
// tools/cpu/sm83/cpu.go:95:1
void sm83_CPU_setDE(sm83_CPU* c,uint16_t v){
{
auto tmp9 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->D = std::get<0>(tmp9);
c->E = std::get<1>(tmp9);
}
}
// tools/cpu/sm83/cpu.go:96:1
void sm83_CPU_setHL(sm83_CPU* c,uint16_t v){
{
auto tmp10 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->H = std::get<0>(tmp10);
c->L = std::get<1>(tmp10);
}
}
// tools/cpu/sm83/cpu.go:97:1
void sm83_CPU_setAF(sm83_CPU* c,uint16_t v){
{
auto tmp11 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>((cast<uint8_t>(v) & cast<uint8_t>(240ULL))));
c->A = std::get<0>(tmp11);
c->F = std::get<1>(tmp11);
}
}
// tools/cpu/sm83/cpu.go:100:1
uint16_t sm83_CPU_getRP(sm83_CPU* c,int64_t p,bool af){
{
{
switch(p){
case cast<int64_t>(0ULL):{
return sm83_CPU_bc(c);
break;}
case cast<int64_t>(1ULL):{
return sm83_CPU_de(c);
break;}
case cast<int64_t>(2ULL):{
return sm83_CPU_hl(c);
break;}
default:{
if (af) {
return sm83_CPU_af(c);
}
return c->SP;
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:115:1
void sm83_CPU_setRP(sm83_CPU* c,int64_t p,uint16_t v,bool af){
{
{
switch(p){
case cast<int64_t>(0ULL):{
sm83_CPU_setBC(c,v);
break;}
case cast<int64_t>(1ULL):{
sm83_CPU_setDE(c,v);
break;}
case cast<int64_t>(2ULL):{
sm83_CPU_setHL(c,v);
break;}
default:{
if (af) {
sm83_CPU_setAF(c,v);
}
else {
c->SP = v;
}
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:133:1
uint8_t sm83_CPU_getR(sm83_CPU* c,int64_t i){
{
{
switch(i){
case cast<int64_t>(0ULL):{
return c->B;
break;}
case cast<int64_t>(1ULL):{
return c->C;
break;}
case cast<int64_t>(2ULL):{
return c->D;
break;}
case cast<int64_t>(3ULL):{
return c->E;
break;}
case cast<int64_t>(4ULL):{
return c->H;
break;}
case cast<int64_t>(5ULL):{
return c->L;
break;}
case cast<int64_t>(6ULL):{
return sm83_CPU_read(c,sm83_CPU_hl(c));
break;}
default:{
return c->A;
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:153:1
void sm83_CPU_setR(sm83_CPU* c,int64_t i,uint8_t v){
{
{
switch(i){
case cast<int64_t>(0ULL):{
c->B = v;
break;}
case cast<int64_t>(1ULL):{
c->C = v;
break;}
case cast<int64_t>(2ULL):{
c->D = v;
break;}
case cast<int64_t>(3ULL):{
c->E = v;
break;}
case cast<int64_t>(4ULL):{
c->H = v;
break;}
case cast<int64_t>(5ULL):{
c->L = v;
break;}
case cast<int64_t>(6ULL):{
sm83_CPU_write(c,sm83_CPU_hl(c),v);
break;}
default:{
c->A = v;
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:176:1
bool sm83_CPU_getf(sm83_CPU* c,uint8_t b){
{
return (cast<uint8_t>((c->F & b)) != cast<uint8_t>(0ULL));
}
}
// tools/cpu/sm83/cpu.go:177:1
void sm83_CPU_setFlags(sm83_CPU* c,bool z,bool n,bool h,bool cy){
{
c->F = cast<uint8_t>(0ULL);
if (z) {
c->F |= cast<uint8_t>(128ULL);
}
if (n) {
c->F |= cast<uint8_t>(64ULL);
}
if (h) {
c->F |= cast<uint8_t>(32ULL);
}
if (cy) {
c->F |= cast<uint8_t>(16ULL);
}
}
}
// tools/cpu/sm83/cpu.go:195:1
void sm83_CPU_add8(sm83_CPU* c,uint8_t n,bool carry){
{
uint8_t cin = cast<uint8_t>(0ULL);
if ((carry && sm83_CPU_getf(c,cast<uint8_t>(16ULL)))) {
cin = cast<uint8_t>(1ULL);
}
uint16_t r = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(c->A) + cast<uint16_t>(n))) + cast<uint16_t>(cin)));
uint8_t res = cast<uint8_t>(r);
sm83_CPU_setFlags(c,(res == cast<uint8_t>(0ULL)),false,(cast<uint8_t>((cast<uint8_t>(((cast<uint8_t>((c->A & cast<uint8_t>(15ULL)))) + (cast<uint8_t>((n & cast<uint8_t>(15ULL)))))) + cin)) > cast<uint8_t>(15ULL)),(r > cast<uint16_t>(255ULL)));
c->A = res;
}
}
// tools/cpu/sm83/cpu.go:207:1
void sm83_CPU_sub8(sm83_CPU* c,uint8_t n,bool carry,bool store){
{
uint8_t cin = cast<uint8_t>(0ULL);
if ((carry && sm83_CPU_getf(c,cast<uint8_t>(16ULL)))) {
cin = cast<uint8_t>(1ULL);
}
int64_t r = cast<int64_t>((cast<int64_t>((cast<int64_t>(c->A) - cast<int64_t>(n))) - cast<int64_t>(cin)));
uint8_t res = cast<uint8_t>(r);
sm83_CPU_setFlags(c,(res == cast<uint8_t>(0ULL)),true,(cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint8_t>((c->A & cast<uint8_t>(15ULL)))) - cast<int64_t>(cast<uint8_t>((n & cast<uint8_t>(15ULL)))))) - cast<int64_t>(cin))) < cast<int64_t>(0ULL)),(r < cast<int64_t>(0ULL)));
if (store) {
c->A = res;
}
}
}
// tools/cpu/sm83/cpu.go:220:1
void sm83_CPU_and8(sm83_CPU* c,uint8_t n){
{
c->A &= n;
sm83_CPU_setFlags(c,(c->A == cast<uint8_t>(0ULL)),false,true,false);
}
}
// tools/cpu/sm83/cpu.go:221:1
void sm83_CPU_or8(sm83_CPU* c,uint8_t n){
{
c->A |= n;
sm83_CPU_setFlags(c,(c->A == cast<uint8_t>(0ULL)),false,false,false);
}
}
// tools/cpu/sm83/cpu.go:222:1
void sm83_CPU_xor8(sm83_CPU* c,uint8_t n){
{
c->A ^= n;
sm83_CPU_setFlags(c,(c->A == cast<uint8_t>(0ULL)),false,false,false);
}
}
// tools/cpu/sm83/cpu.go:224:1
void sm83_CPU_alu(sm83_CPU* c,int64_t y,uint8_t n){
{
{
switch(y){
case cast<int64_t>(0ULL):{
sm83_CPU_add8(c,n,false);
break;}
case cast<int64_t>(1ULL):{
sm83_CPU_add8(c,n,true);
break;}
case cast<int64_t>(2ULL):{
sm83_CPU_sub8(c,n,false,true);
break;}
case cast<int64_t>(3ULL):{
sm83_CPU_sub8(c,n,true,true);
break;}
case cast<int64_t>(4ULL):{
sm83_CPU_and8(c,n);
break;}
case cast<int64_t>(5ULL):{
sm83_CPU_xor8(c,n);
break;}
case cast<int64_t>(6ULL):{
sm83_CPU_or8(c,n);
break;}
default:{
sm83_CPU_sub8(c,n,false,false);
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:245:1
uint8_t sm83_CPU_inc8(sm83_CPU* c,uint8_t v){
{
uint8_t r = cast<uint8_t>((v + cast<uint8_t>(1ULL)));
sm83_CPU_setFlags(c,(r == cast<uint8_t>(0ULL)),false,(cast<uint8_t>((v & cast<uint8_t>(15ULL))) == cast<uint8_t>(15ULL)),sm83_CPU_getf(c,cast<uint8_t>(16ULL)));
return r;
}
}
// tools/cpu/sm83/cpu.go:250:1
uint8_t sm83_CPU_dec8(sm83_CPU* c,uint8_t v){
{
uint8_t r = cast<uint8_t>((v - cast<uint8_t>(1ULL)));
sm83_CPU_setFlags(c,(r == cast<uint8_t>(0ULL)),true,(cast<uint8_t>((v & cast<uint8_t>(15ULL))) == cast<uint8_t>(0ULL)),sm83_CPU_getf(c,cast<uint8_t>(16ULL)));
return r;
}
}
// tools/cpu/sm83/cpu.go:257:1
uint16_t sm83_CPU_add16(sm83_CPU* c,uint16_t a,uint16_t b){
{
uint32_t r = cast<uint32_t>((cast<uint32_t>(a) + cast<uint32_t>(b)));
bool z = sm83_CPU_getf(c,cast<uint8_t>(128ULL));
sm83_CPU_setFlags(c,z,false,(cast<uint16_t>(((cast<uint16_t>((a & cast<uint16_t>(4095ULL)))) + (cast<uint16_t>((b & cast<uint16_t>(4095ULL)))))) > cast<uint16_t>(4095ULL)),(r > cast<uint32_t>(65535ULL)));
return cast<uint16_t>(r);
}
}
// tools/cpu/sm83/cpu.go:265:1
uint16_t sm83_CPU_addSP(sm83_CPU* c,uint8_t e){
{
uint16_t res = cast<uint16_t>((c->SP + cast<uint16_t>(cast<int8_t>(e))));
sm83_CPU_setFlags(c,false,false,(cast<uint16_t>(((cast<uint16_t>((c->SP & cast<uint16_t>(15ULL)))) + (cast<uint16_t>((cast<uint16_t>(e) & cast<uint16_t>(15ULL)))))) > cast<uint16_t>(15ULL)),(cast<uint16_t>(((cast<uint16_t>((c->SP & cast<uint16_t>(255ULL)))) + (cast<uint16_t>((cast<uint16_t>(e) & cast<uint16_t>(255ULL)))))) > cast<uint16_t>(255ULL)));
return res;
}
}
// tools/cpu/sm83/cpu.go:272:1
uint8_t sm83_CPU_rot(sm83_CPU* c,int64_t y,uint8_t v){
{
uint8_t r={};
bool cf = sm83_CPU_getf(c,cast<uint8_t>(16ULL));
bool newC={};
{
switch(y){
case cast<int64_t>(0ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL));
r = cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(1ULL)) | shr<uint8_t>(v,cast<int64_t>(7ULL))));
break;}
case cast<int64_t>(1ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
r = cast<uint8_t>((shr<uint8_t>(v,cast<int64_t>(1ULL)) | shl<uint8_t>(v,cast<int64_t>(7ULL))));
break;}
case cast<int64_t>(2ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL));
r = shl<uint8_t>(v,cast<int64_t>(1ULL));
if (cf) {
r |= cast<uint8_t>(1ULL);
}
break;}
case cast<int64_t>(3ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
r = shr<uint8_t>(v,cast<int64_t>(1ULL));
if (cf) {
r |= cast<uint8_t>(128ULL);
}
break;}
case cast<int64_t>(4ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL));
r = shl<uint8_t>(v,cast<int64_t>(1ULL));
break;}
case cast<int64_t>(5ULL):{
newC = (cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
r = cast<uint8_t>((shr<uint8_t>(v,cast<int64_t>(1ULL)) | cast<uint8_t>((v & cast<uint8_t>(128ULL)))));
break;}
case cast<int64_t>(6ULL):{
r = cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(4ULL)) | shr<uint8_t>(v,cast<int64_t>(4ULL))));
newC = false;
break;}
default:{
newC = (cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
r = shr<uint8_t>(v,cast<int64_t>(1ULL));
break;}
}}
sm83_CPU_setFlags(c,(r == cast<uint8_t>(0ULL)),false,false,newC);
return r;
}
}
// tools/cpu/sm83/cpu.go:316:1
int64_t sm83_CPU_Step(sm83_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
uint8_t iff = sm83_CPU_read(c,cast<uint16_t>(65295ULL));
uint8_t pend = cast<uint8_t>((cast<uint8_t>((sm83_CPU_read(c,cast<uint16_t>(65535ULL)) & iff)) & cast<uint8_t>(31ULL)));
if ((pend != cast<uint8_t>(0ULL))) {
c->halt = false;
}
if ((c->IME && (pend != cast<uint8_t>(0ULL)))) {
auto tmp12 = std::make_tuple(false,false);
c->IME = std::get<0>(tmp12);
c->imeEnable = std::get<1>(tmp12);
int64_t bit = cast<int64_t>(0ULL);
{;for (;(bit < cast<int64_t>(5ULL));bit++){
if ((cast<uint8_t>((pend & (shl<uint8_t>(cast<uint8_t>(1ULL),bit)))) != cast<uint8_t>(0ULL))) {
break;
}
}
}sm83_CPU_write(c,cast<uint16_t>(65295ULL),(iff & ~((shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(bit))))));
sm83_CPU_push16(c,c->PC);
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(64ULL) + cast<int64_t>((bit * cast<int64_t>(8ULL))))));
c->Instrs++;
return cast<int64_t>(20ULL);
}
if ((c->halt || c->Stopped)) {
c->Instrs++;
return cast<int64_t>(4ULL);
}
bool pendingEI = c->imeEnable;
c->takenExtra = cast<int64_t>(0ULL);
c->Instrs++;
uint8_t op = sm83_CPU_fetch(c);
int64_t cyc = cast<int64_t>(0ULL);
if ((op == cast<uint8_t>(203ULL))) {
uint8_t cb = sm83_CPU_fetch(c);
sm83_CPU_execCB(c,cb);
cyc = sm83_cbCycles(cb);
}
else {
sm83_CPU_execMain(c,op);
cyc = cast<int64_t>((cast<int64_t>(sm83_mainCycles[op]) + c->takenExtra));
}
if ((pendingEI && c->imeEnable)) {
auto tmp13 = std::make_tuple(true,false);
c->IME = std::get<0>(tmp13);
c->imeEnable = std::get<1>(tmp13);
}
return cyc;
}
}
// tools/cpu/sm83/cpu.go:367:1
void sm83_CPU_execMain(sm83_CPU* c,uint8_t op){
{
auto tmp14 = std::make_tuple(shr<uint8_t>(op,cast<int64_t>(6ULL)),cast<int64_t>((cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))) & cast<int64_t>(7ULL))),cast<int64_t>((cast<int64_t>(op) & cast<int64_t>(7ULL))));
uint8_t x = std::get<0>(tmp14);
int64_t y = std::get<1>(tmp14);
int64_t z = std::get<2>(tmp14);
auto tmp15 = std::make_tuple(shr<int64_t>(y,cast<int64_t>(1ULL)),cast<int64_t>((y & cast<int64_t>(1ULL))));
int64_t p = std::get<0>(tmp15);
int64_t q = std::get<1>(tmp15);
{
switch(x){
case cast<uint8_t>(0ULL):{
sm83_CPU_execX0(c,y,z,p,q);
break;}
case cast<uint8_t>(1ULL):{
if (((z == cast<int64_t>(6ULL)) && (y == cast<int64_t>(6ULL)))) {
c->halt = true;
return ;
}
sm83_CPU_setR(c,y,sm83_CPU_getR(c,z));
break;}
case cast<uint8_t>(2ULL):{
sm83_CPU_alu(c,y,sm83_CPU_getR(c,z));
break;}
default:{
sm83_CPU_execX3(c,y,z,p,q);
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:386:1
void sm83_CPU_execX0(sm83_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q){
{
{
switch(z){
case cast<int64_t>(0ULL):{
{
switch(y){
case cast<int64_t>(0ULL):{
break;}
case cast<int64_t>(1ULL):{
sm83_CPU_write16(c,sm83_CPU_fetch16(c),c->SP);
break;}
case cast<int64_t>(2ULL):{
sm83_CPU_fetch(c);
c->Stopped = true;
break;}
case cast<int64_t>(3ULL):{
int8_t off = cast<int8_t>(sm83_CPU_fetch(c));
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(c->PC) + cast<int64_t>(off))));
break;}
default:{
int8_t off = cast<int8_t>(sm83_CPU_fetch(c));
if (sm83_CPU_cond(c,cast<int64_t>((y - cast<int64_t>(4ULL))))) {
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(c->PC) + cast<int64_t>(off))));
c->takenExtra = cast<int64_t>(4ULL);
}
break;}
}}
break;}
case cast<int64_t>(1ULL):{
if ((q == cast<int64_t>(0ULL))) {
sm83_CPU_setRP(c,p,sm83_CPU_fetch16(c),false);
}
else {
sm83_CPU_setHL(c,sm83_CPU_add16(c,sm83_CPU_hl(c),sm83_CPU_getRP(c,p,false)));
}
break;}
case cast<int64_t>(2ULL):{
{
if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(0ULL)))){
sm83_CPU_write(c,sm83_CPU_bc(c),c->A);
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(1ULL)))){
sm83_CPU_write(c,sm83_CPU_de(c),c->A);
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(2ULL)))){
sm83_CPU_write(c,sm83_CPU_hl(c),c->A);
sm83_CPU_setHL(c,cast<uint16_t>((sm83_CPU_hl(c) + cast<uint16_t>(1ULL))));
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(3ULL)))){
sm83_CPU_write(c,sm83_CPU_hl(c),c->A);
sm83_CPU_setHL(c,cast<uint16_t>((sm83_CPU_hl(c) - cast<uint16_t>(1ULL))));
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(0ULL)))){
c->A = sm83_CPU_read(c,sm83_CPU_bc(c));
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(1ULL)))){
c->A = sm83_CPU_read(c,sm83_CPU_de(c));
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(2ULL)))){
c->A = sm83_CPU_read(c,sm83_CPU_hl(c));
sm83_CPU_setHL(c,cast<uint16_t>((sm83_CPU_hl(c) + cast<uint16_t>(1ULL))));
}
else {
c->A = sm83_CPU_read(c,sm83_CPU_hl(c));
sm83_CPU_setHL(c,cast<uint16_t>((sm83_CPU_hl(c) - cast<uint16_t>(1ULL))));
}
}
tmp16:;
break;}
case cast<int64_t>(3ULL):{
if ((q == cast<int64_t>(0ULL))) {
sm83_CPU_setRP(c,p,cast<uint16_t>((sm83_CPU_getRP(c,p,false) + cast<uint16_t>(1ULL))),false);
}
else {
sm83_CPU_setRP(c,p,cast<uint16_t>((sm83_CPU_getRP(c,p,false) - cast<uint16_t>(1ULL))),false);
}
break;}
case cast<int64_t>(4ULL):{
sm83_CPU_setR(c,y,sm83_CPU_inc8(c,sm83_CPU_getR(c,y)));
break;}
case cast<int64_t>(5ULL):{
sm83_CPU_setR(c,y,sm83_CPU_dec8(c,sm83_CPU_getR(c,y)));
break;}
case cast<int64_t>(6ULL):{
sm83_CPU_setR(c,y,sm83_CPU_fetch(c));
break;}
default:{
sm83_CPU_accOp(c,y);
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:452:1
void sm83_CPU_accOp(sm83_CPU* c,int64_t y){
{
{
switch(y){
case cast<int64_t>(0ULL):{
bool cy = (cast<uint8_t>((c->A & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL));
c->A = cast<uint8_t>((shl<uint8_t>(c->A,cast<int64_t>(1ULL)) | shr<uint8_t>(c->A,cast<int64_t>(7ULL))));
sm83_CPU_setFlags(c,false,false,false,cy);
break;}
case cast<int64_t>(1ULL):{
bool cy = (cast<uint8_t>((c->A & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
c->A = cast<uint8_t>((shr<uint8_t>(c->A,cast<int64_t>(1ULL)) | shl<uint8_t>(c->A,cast<int64_t>(7ULL))));
sm83_CPU_setFlags(c,false,false,false,cy);
break;}
case cast<int64_t>(2ULL):{
bool old = sm83_CPU_getf(c,cast<uint8_t>(16ULL));
bool cy = (cast<uint8_t>((c->A & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL));
c->A = shl<uint8_t>(c->A,cast<int64_t>(1ULL));
if (old) {
c->A |= cast<uint8_t>(1ULL);
}
sm83_CPU_setFlags(c,false,false,false,cy);
break;}
case cast<int64_t>(3ULL):{
bool old = sm83_CPU_getf(c,cast<uint8_t>(16ULL));
bool cy = (cast<uint8_t>((c->A & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
c->A = shr<uint8_t>(c->A,cast<int64_t>(1ULL));
if (old) {
c->A |= cast<uint8_t>(128ULL);
}
sm83_CPU_setFlags(c,false,false,false,cy);
break;}
case cast<int64_t>(4ULL):{
sm83_CPU_daa(c);
break;}
case cast<int64_t>(5ULL):{
c->A = cast<uint8_t>(~c->A);
c->F |= cast<uint8_t>(96ULL);
break;}
case cast<int64_t>(6ULL):{
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(128ULL))) | cast<uint8_t>(16ULL)));
break;}
default:{
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(128ULL))) | cast<uint8_t>(((cast<uint8_t>((c->F ^ cast<uint8_t>(16ULL)))) & cast<uint8_t>(16ULL)))));
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:490:1
void sm83_CPU_daa(sm83_CPU* c){
{
uint8_t corr={};
bool carry = sm83_CPU_getf(c,cast<uint8_t>(16ULL));
if ((sm83_CPU_getf(c,cast<uint8_t>(32ULL)) || (((!sm83_CPU_getf(c,cast<uint8_t>(64ULL))) && (cast<uint8_t>((c->A & cast<uint8_t>(15ULL))) > cast<uint8_t>(9ULL)))))) {
corr |= cast<uint8_t>(6ULL);
}
if ((sm83_CPU_getf(c,cast<uint8_t>(16ULL)) || (((!sm83_CPU_getf(c,cast<uint8_t>(64ULL))) && (c->A > cast<uint8_t>(153ULL)))))) {
corr |= cast<uint8_t>(96ULL);
carry = true;
}
if (sm83_CPU_getf(c,cast<uint8_t>(64ULL))) {
c->A -= corr;
}
else {
c->A += corr;
}
c->F &= ~(cast<uint8_t>(176ULL));
if ((c->A == cast<uint8_t>(0ULL))) {
c->F |= cast<uint8_t>(128ULL);
}
if (carry) {
c->F |= cast<uint8_t>(16ULL);
}
}
}
// tools/cpu/sm83/cpu.go:514:1
void sm83_CPU_execX3(sm83_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q){
{
{
switch(z){
case cast<int64_t>(0ULL):{
{
switch(y){
case cast<int64_t>(0ULL):case cast<int64_t>(1ULL):case cast<int64_t>(2ULL):case cast<int64_t>(3ULL):{
if (sm83_CPU_cond(c,y)) {
c->PC = sm83_CPU_pop16(c);
c->takenExtra = cast<int64_t>(12ULL);
}
break;}
case cast<int64_t>(4ULL):{
sm83_CPU_write(c,cast<uint16_t>((cast<uint16_t>(65280ULL) | cast<uint16_t>(sm83_CPU_fetch(c)))),c->A);
break;}
case cast<int64_t>(5ULL):{
c->SP = sm83_CPU_addSP(c,sm83_CPU_fetch(c));
break;}
case cast<int64_t>(6ULL):{
c->A = sm83_CPU_read(c,cast<uint16_t>((cast<uint16_t>(65280ULL) | cast<uint16_t>(sm83_CPU_fetch(c)))));
break;}
default:{
sm83_CPU_setHL(c,sm83_CPU_addSP(c,sm83_CPU_fetch(c)));
break;}
}}
break;}
case cast<int64_t>(1ULL):{
if ((q == cast<int64_t>(0ULL))) {
sm83_CPU_setRP(c,p,sm83_CPU_pop16(c),true);
}
else {
{
switch(p){
case cast<int64_t>(0ULL):{
c->PC = sm83_CPU_pop16(c);
break;}
case cast<int64_t>(1ULL):{
c->PC = sm83_CPU_pop16(c);
auto tmp17 = std::make_tuple(true,false);
c->IME = std::get<0>(tmp17);
c->imeEnable = std::get<1>(tmp17);
break;}
case cast<int64_t>(2ULL):{
c->PC = sm83_CPU_hl(c);
break;}
default:{
c->SP = sm83_CPU_hl(c);
break;}
}}
}
break;}
case cast<int64_t>(2ULL):{
{
switch(y){
case cast<int64_t>(0ULL):case cast<int64_t>(1ULL):case cast<int64_t>(2ULL):case cast<int64_t>(3ULL):{
uint16_t a = sm83_CPU_fetch16(c);
if (sm83_CPU_cond(c,y)) {
c->PC = a;
c->takenExtra = cast<int64_t>(4ULL);
}
break;}
case cast<int64_t>(4ULL):{
sm83_CPU_write(c,cast<uint16_t>((cast<uint16_t>(65280ULL) | cast<uint16_t>(c->C))),c->A);
break;}
case cast<int64_t>(5ULL):{
sm83_CPU_write(c,sm83_CPU_fetch16(c),c->A);
break;}
case cast<int64_t>(6ULL):{
c->A = sm83_CPU_read(c,cast<uint16_t>((cast<uint16_t>(65280ULL) | cast<uint16_t>(c->C))));
break;}
default:{
c->A = sm83_CPU_read(c,sm83_CPU_fetch16(c));
break;}
}}
break;}
case cast<int64_t>(3ULL):{
{
switch(y){
case cast<int64_t>(0ULL):{
c->PC = sm83_CPU_fetch16(c);
break;}
case cast<int64_t>(6ULL):{
auto tmp18 = std::make_tuple(false,false);
c->IME = std::get<0>(tmp18);
c->imeEnable = std::get<1>(tmp18);
break;}
case cast<int64_t>(7ULL):{
c->imeEnable = true;
break;}
default:{
sm83_CPU_Halt(c,std::string("illegal opcode at $%04X",23),cast<uint16_t>((c->PC - cast<uint16_t>(1ULL))));
break;}
}}
break;}
case cast<int64_t>(4ULL):{
if ((y <= cast<int64_t>(3ULL))) {
uint16_t a = sm83_CPU_fetch16(c);
if (sm83_CPU_cond(c,y)) {
sm83_CPU_push16(c,c->PC);
c->PC = a;
c->takenExtra = cast<int64_t>(12ULL);
}
}
else {
sm83_CPU_Halt(c,std::string("illegal opcode at $%04X",23),cast<uint16_t>((c->PC - cast<uint16_t>(1ULL))));
}
break;}
case cast<int64_t>(5ULL):{
if ((q == cast<int64_t>(0ULL))) {
sm83_CPU_push16(c,sm83_CPU_getRP(c,p,true));
}
else if ((p == cast<int64_t>(0ULL))) {
uint16_t a = sm83_CPU_fetch16(c);
sm83_CPU_push16(c,c->PC);
c->PC = a;
}
else {
sm83_CPU_Halt(c,std::string("illegal opcode at $%04X",23),cast<uint16_t>((c->PC - cast<uint16_t>(1ULL))));
}
break;}
case cast<int64_t>(6ULL):{
sm83_CPU_alu(c,y,sm83_CPU_fetch(c));
break;}
default:{
sm83_CPU_push16(c,c->PC);
c->PC = cast<uint16_t>((cast<uint16_t>(y) * cast<uint16_t>(8ULL)));
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:606:1
bool sm83_CPU_cond(sm83_CPU* c,int64_t y){
{
{
switch(y){
case cast<int64_t>(0ULL):{
return (!sm83_CPU_getf(c,cast<uint8_t>(128ULL)));
break;}
case cast<int64_t>(1ULL):{
return sm83_CPU_getf(c,cast<uint8_t>(128ULL));
break;}
case cast<int64_t>(2ULL):{
return (!sm83_CPU_getf(c,cast<uint8_t>(16ULL)));
break;}
default:{
return sm83_CPU_getf(c,cast<uint8_t>(16ULL));
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:619:1
void sm83_CPU_execCB(sm83_CPU* c,uint8_t op){
{
auto tmp19 = std::make_tuple(shr<uint8_t>(op,cast<int64_t>(6ULL)),cast<int64_t>((cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))) & cast<int64_t>(7ULL))),cast<int64_t>((cast<int64_t>(op) & cast<int64_t>(7ULL))));
uint8_t x = std::get<0>(tmp19);
int64_t y = std::get<1>(tmp19);
int64_t z = std::get<2>(tmp19);
{
switch(x){
case cast<uint8_t>(0ULL):{
sm83_CPU_setR(c,z,sm83_CPU_rot(c,y,sm83_CPU_getR(c,z)));
break;}
case cast<uint8_t>(1ULL):{
uint8_t v = sm83_CPU_getR(c,z);
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(16ULL))) | cast<uint8_t>(32ULL)));
if ((cast<uint8_t>((v & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y))))) == cast<uint8_t>(0ULL))) {
c->F |= cast<uint8_t>(128ULL);
}
break;}
case cast<uint8_t>(2ULL):{
sm83_CPU_setR(c,z,(sm83_CPU_getR(c,z) & ~((shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y))))));
break;}
default:{
sm83_CPU_setR(c,z,cast<uint8_t>((sm83_CPU_getR(c,z) | shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y)))));
break;}
}}
}
}
// tools/cpu/sm83/cpu.go:639:1
int64_t sm83_cbCycles(uint8_t op){
{
if ((cast<uint8_t>((op & cast<uint8_t>(7ULL))) != cast<uint8_t>(6ULL))) {
return cast<int64_t>(8ULL);
}
if ((shr<uint8_t>(op,cast<int64_t>(6ULL)) == cast<uint8_t>(1ULL))) {
return cast<int64_t>(12ULL);
}
return cast<int64_t>(16ULL);
}
}
// tools/platform/gameboy/machine.go:102:1
gameboy_Machine* gameboy_NewMachine(Slice<uint8_t> rom){
{
gameboy_Machine* m = arenaNew(gameboy_Machine{{},rom,divi<int64_t>(len(rom),cast<int64_t>(16384ULL)),cast<int64_t>(1ULL),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
if ((m->nbanks == cast<int64_t>(0ULL))) {
m->nbanks = cast<int64_t>(1ULL);
}
m->io[cast<int64_t>(64ULL)] = cast<uint8_t>(145ULL);
m->io[cast<int64_t>(0ULL)] = cast<uint8_t>(63ULL);
m->CPU = sm83_NewCPU(m);
return m;
}
}
// tools/platform/gameboy/machine.go:114:1
void gameboy_Machine_Watch(gameboy_Machine* m,uint16_t lo,uint16_t hi){
{
auto tmp1 = std::make_tuple(lo,hi);
m->WatchLo = std::get<0>(tmp1);
m->WatchHi = std::get<1>(tmp1);
m->WatchPCs = Map<uint16_t,int64_t>{};
}
}
// tools/platform/gameboy/machine.go:120:1
Slice<uint8_t> gameboy_Machine_VRAM(gameboy_Machine* m){
{
return sub(m->vram,0,len(m->vram));
}
}
// tools/platform/gameboy/machine.go:121:1
Slice<uint8_t> gameboy_Machine_OAM(gameboy_Machine* m){
{
return sub(m->oam,0,len(m->oam));
}
}
// tools/platform/gameboy/machine.go:122:1
Slice<uint8_t> gameboy_Machine_WRAM(gameboy_Machine* m){
{
return sub(m->wram,0,len(m->wram));
}
}
// tools/platform/gameboy/machine.go:125:1
int64_t gameboy_Machine_ROMBank(gameboy_Machine* m){
{
return m->romBank;
}
}
// tools/platform/gameboy/machine.go:129:1
uint8_t gameboy_Machine_Read_Reference(gameboy_Machine* m,uint16_t a){
{
{
if ((a < cast<uint16_t>(16384ULL))){
return m->rom[cast<int64_t>(a)];
}
else if ((a < cast<uint16_t>(32768ULL))){
if (bool(m->OnROMRead)) {
m->OnROMRead(m->CPU->PC,m->romBank,a);
}
int64_t off = cast<int64_t>((cast<int64_t>((m->romBank * cast<int64_t>(16384ULL))) + cast<int64_t>(cast<uint16_t>((a - cast<uint16_t>(16384ULL))))));
if ((off < len(m->rom))) {
return m->rom[off];
}
return cast<uint8_t>(255ULL);
}
else if ((a < cast<uint16_t>(40960ULL))){
return m->vram[cast<uint16_t>((a - cast<uint16_t>(32768ULL)))];
}
else if ((a < cast<uint16_t>(49152ULL))){
if (m->ramEnable) {
return m->extram[gameboy_Machine_ramOff(m,a)];
}
return cast<uint8_t>(255ULL);
}
else if ((a < cast<uint16_t>(57344ULL))){
return m->wram[cast<uint16_t>((a - cast<uint16_t>(49152ULL)))];
}
else if ((a < cast<uint16_t>(65024ULL))){
return m->wram[cast<uint16_t>((a - cast<uint16_t>(57344ULL)))];
}
else if ((a < cast<uint16_t>(65184ULL))){
return m->oam[cast<uint16_t>((a - cast<uint16_t>(65024ULL)))];
}
else if ((a < cast<uint16_t>(65280ULL))){
return cast<uint8_t>(255ULL);
}
else if ((a < cast<uint16_t>(65408ULL))){
return gameboy_Machine_readIO(m,a);
}
else if ((a < cast<uint16_t>(65535ULL))){
return m->hram[cast<uint16_t>((a - cast<uint16_t>(65408ULL)))];
}
else {
return m->ie;
}
}
tmp2:;
}
}
// tools/platform/gameboy/machine.go:166:1
void gameboy_Machine_Write(gameboy_Machine* m,uint16_t a,uint8_t v){
{auto writtenAddress=a; auto writtenValue=v;
{
if (bool(m->OnWrite)) {
m->OnWrite(m->CPU->PC,a,v);
}
{
if ((a < cast<uint16_t>(32768ULL))){
gameboy_Machine_mbcWrite(m,a,v);
}
else if ((a < cast<uint16_t>(40960ULL))){
if (((bool(m->WatchPCs) && (a >= m->WatchLo)) && (a < m->WatchHi))) {
m->WatchPCs[m->CPU->PC]++;
}
m->vram[cast<uint16_t>((a - cast<uint16_t>(32768ULL)))] = v;
}
else if ((a < cast<uint16_t>(49152ULL))){
if (m->ramEnable) {
m->extram[gameboy_Machine_ramOff(m,a)] = v;
}
}
else if ((a < cast<uint16_t>(57344ULL))){
m->wram[cast<uint16_t>((a - cast<uint16_t>(49152ULL)))] = v;
}
else if ((a < cast<uint16_t>(65024ULL))){
m->wram[cast<uint16_t>((a - cast<uint16_t>(57344ULL)))] = v;
}
else if ((a < cast<uint16_t>(65184ULL))){
m->oam[cast<uint16_t>((a - cast<uint16_t>(65024ULL)))] = v;
}
else if ((a < cast<uint16_t>(65280ULL))){
}
else if ((a < cast<uint16_t>(65408ULL))){
gameboy_Machine_writeIO(m,a,v);
}
else if ((a < cast<uint16_t>(65535ULL))){
m->hram[cast<uint16_t>((a - cast<uint16_t>(65408ULL)))] = v;
}
else {
m->ie = v;
}
}
tmp3:;
}
rrAfterWrite(m,writtenAddress,writtenValue);
}
}
// tools/platform/gameboy/machine.go:201:1
int64_t gameboy_Machine_ramOff(gameboy_Machine* m,uint16_t a){
{
int64_t bank = cast<int64_t>(0ULL);
if ((m->mode == cast<uint8_t>(1ULL))) {
bank = m->ramBank;
}
return cast<int64_t>((cast<int64_t>((bank * cast<int64_t>(8192ULL))) + cast<int64_t>(cast<uint16_t>((a - cast<uint16_t>(40960ULL))))));
}
}
// tools/platform/gameboy/machine.go:210:1
void gameboy_Machine_mbcWrite_Reference(gameboy_Machine* m,uint16_t a,uint8_t v){
{
{
if ((a < cast<uint16_t>(8192ULL))){
m->ramEnable = (cast<uint8_t>((v & cast<uint8_t>(15ULL))) == cast<uint8_t>(10ULL));
}
else if ((a < cast<uint16_t>(16384ULL))){
int64_t lo = cast<int64_t>(cast<uint8_t>((v & cast<uint8_t>(31ULL))));
if ((lo == cast<int64_t>(0ULL))) {
lo = cast<int64_t>(1ULL);
}
m->romBank = modi<int64_t>((cast<int64_t>((cast<int64_t>((m->romBank & cast<int64_t>(96ULL))) | lo))),m->nbanks);
}
else if ((a < cast<uint16_t>(24576ULL))){
m->ramBank = cast<int64_t>(cast<uint8_t>((v & cast<uint8_t>(3ULL))));
m->romBank = modi<int64_t>((cast<int64_t>((cast<int64_t>((m->romBank & cast<int64_t>(31ULL))) | shl<int64_t>(cast<int64_t>(cast<uint8_t>((v & cast<uint8_t>(3ULL)))),cast<int64_t>(5ULL))))),m->nbanks);
if ((m->romBank == cast<int64_t>(0ULL))) {
m->romBank = cast<int64_t>(1ULL);
}
}
else {
m->mode = cast<uint8_t>((v & cast<uint8_t>(1ULL)));
}
}
tmp4:;
}
}
// tools/platform/gameboy/machine.go:231:1
uint8_t gameboy_Machine_readIO(gameboy_Machine* m,uint16_t a){
{
{
switch(a){
case cast<uint16_t>(65280ULL):{
return gameboy_Machine_joyp(m);
break;}
default:{
return m->io[cast<uint16_t>((a - cast<uint16_t>(65280ULL)))];
break;}
}}
}
}
// tools/platform/gameboy/machine.go:240:1
void gameboy_Machine_writeIO_Reference(gameboy_Machine* m,uint16_t a,uint8_t v){
{
{
switch(a){
case cast<uint16_t>(65284ULL):{
m->divCounter = cast<int64_t>(0ULL);
m->io[cast<int64_t>(4ULL)] = cast<uint8_t>(0ULL);
break;}
case cast<uint16_t>(65350ULL):{
uint16_t src = shl<uint16_t>(cast<uint16_t>(v),cast<int64_t>(8ULL));
{uint16_t i = cast<uint16_t>(0ULL);for (;(i < cast<uint16_t>(160ULL));i++){
m->oam[i] = gameboy_Machine_Read(m,cast<uint16_t>((src + i)));
}
}m->io[cast<int64_t>(70ULL)] = v;
break;}
case cast<uint16_t>(65348ULL):{
break;}
default:{
m->io[cast<uint16_t>((a - cast<uint16_t>(65280ULL)))] = v;
break;}
}}
}
}
// tools/platform/gameboy/machine.go:258:1
uint8_t gameboy_Machine_joyp(gameboy_Machine* m){
{
uint8_t sel = cast<uint8_t>((m->io[cast<int64_t>(0ULL)] & cast<uint8_t>(48ULL)));
uint8_t out = cast<uint8_t>(15ULL);
if ((cast<uint8_t>((sel & cast<uint8_t>(16ULL))) == cast<uint8_t>(0ULL))) {
out &= cast<uint8_t>(~(cast<uint8_t>((m->Buttons & cast<uint8_t>(15ULL)))));
}
if ((cast<uint8_t>((sel & cast<uint8_t>(32ULL))) == cast<uint8_t>(0ULL))) {
out &= cast<uint8_t>(~(shr<uint8_t>(m->Buttons,cast<int64_t>(4ULL))));
}
return cast<uint8_t>((cast<uint8_t>((cast<uint8_t>(192ULL) | sel)) | (cast<uint8_t>((out & cast<uint8_t>(15ULL))))));
}
}
// tools/platform/gameboy/machine.go:274:1
void gameboy_Machine_tick_Reference(gameboy_Machine* m,int64_t cyc){
{
m->Cycles += cast<int64_t>(cyc);
m->divCounter = cast<int64_t>(((cast<int64_t>((m->divCounter + cyc))) & cast<int64_t>(65535ULL)));
m->io[cast<int64_t>(4ULL)] = cast<uint8_t>(shr<int64_t>(m->divCounter,cast<int64_t>(8ULL)));
{
uint8_t tac = m->io[cast<int64_t>(7ULL)];
if ((cast<uint8_t>((tac & cast<uint8_t>(4ULL))) != cast<uint8_t>(0ULL))) {
int64_t period = std::array<int64_t,4>{cast<int64_t>(1024ULL),cast<int64_t>(16ULL),cast<int64_t>(64ULL),cast<int64_t>(256ULL)}[cast<uint8_t>((tac & cast<uint8_t>(3ULL)))];
m->timaCounter += cyc;
{;for (;(m->timaCounter >= period);){
m->timaCounter -= period;
uint8_t t = cast<uint8_t>((m->io[cast<int64_t>(5ULL)] + cast<uint8_t>(1ULL)));
if ((t == cast<uint8_t>(0ULL))) {
t = m->io[cast<int64_t>(6ULL)];
gameboy_Machine_reqInt(m,cast<uint8_t>(4ULL));
}
m->io[cast<int64_t>(5ULL)] = t;
}
}}
}
if ((cast<uint8_t>((m->io[cast<int64_t>(64ULL)] & cast<uint8_t>(128ULL))) == cast<uint8_t>(0ULL))) {
auto tmp5 = std::make_tuple(cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
m->lcdDot = std::get<0>(tmp5);
m->io[cast<int64_t>(68ULL)] = std::get<1>(tmp5);
return ;
}
m->lcdDot += cyc;
{;for (;(m->lcdDot >= cast<int64_t>(456ULL));){
m->lcdDot -= cast<int64_t>(456ULL);
uint8_t ly = cast<uint8_t>((m->io[cast<int64_t>(68ULL)] + cast<uint8_t>(1ULL)));
if ((ly >= cast<uint8_t>(154ULL))) {
ly = cast<uint8_t>(0ULL);
}
m->io[cast<int64_t>(68ULL)] = ly;
uint8_t stat = m->io[cast<int64_t>(65ULL)];
if ((ly == cast<uint8_t>(144ULL))) {
gameboy_Machine_reqInt(m,cast<uint8_t>(1ULL));
if ((cast<uint8_t>((stat & cast<uint8_t>(16ULL))) != cast<uint8_t>(0ULL))) {
gameboy_Machine_reqInt(m,cast<uint8_t>(2ULL));
}
}
if ((ly == m->io[cast<int64_t>(69ULL)])) {
m->io[cast<int64_t>(65ULL)] |= cast<uint8_t>(4ULL);
if ((cast<uint8_t>((stat & cast<uint8_t>(64ULL))) != cast<uint8_t>(0ULL))) {
gameboy_Machine_reqInt(m,cast<uint8_t>(2ULL));
}
}
else {
m->io[cast<int64_t>(65ULL)] &= ~(cast<uint8_t>(4ULL));
}
}
}}
}
// tools/platform/gameboy/machine.go:327:1
void gameboy_Machine_reqInt(gameboy_Machine* m,uint8_t bit){
{
m->io[cast<int64_t>(15ULL)] |= bit;
}
}
// tools/platform/gameboy/machine.go:333:1
int64_t gameboy_Machine_Step(gameboy_Machine* m){
{
if (m->CPU->Halted) {
return cast<int64_t>(0ULL);
}
uint16_t pc = m->CPU->PC;
int64_t cyc = sm83_CPU_Step(m->CPU);
gameboy_Machine_tick(m,cyc);
if (m->Sample) {
if ((!m->PCHist)) {
m->PCHist = Map<uint16_t,int64_t>{};
}
m->PCHist[pc]++;
}
return cyc;
}
}
// tools/platform/gameboy/machine.go:351:1
bool gameboy_Machine_RunFrame(gameboy_Machine* m){
{
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(70224ULL));){
int64_t cyc = gameboy_Machine_Step(m);
if (((cyc == cast<int64_t>(0ULL)) || m->CPU->Halted)) {
return false;
}
c += cyc;
}
}return true;
}
}
// tools/platform/gameboy/machine.go:363:1
int64_t gameboy_Machine_RunFrames(gameboy_Machine* m,int64_t n){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
if ((!gameboy_Machine_RunFrame(m))) {
return i;
}
}
}return n;
}
}

#include "fast.h"
