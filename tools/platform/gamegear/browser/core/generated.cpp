#include "runtime.h"
struct z80_CPU;
struct gamegear_VDP;
struct gamegear_CapEntry;
struct gamegear_Machine;
struct gamegear_PSG;
struct Anon0;
struct Anon1;
struct Anon2;
struct Anon3;
struct Anon4;
struct z80_CPU{
uint8_t A{};
uint8_t F{};
uint8_t B{};
uint8_t C{};
uint8_t D{};
uint8_t E{};
uint8_t H{};
uint8_t L{};
uint8_t A2{};
uint8_t F2{};
uint8_t B2{};
uint8_t C2{};
uint8_t D2{};
uint8_t E2{};
uint8_t H2{};
uint8_t L2{};
uint16_t IX{};
uint16_t IY{};
uint16_t SP{};
uint16_t PC{};
uint8_t I{};
uint8_t R{};
bool IFF1{};
bool IFF2{};
uint8_t IM{};
bool Halted{};
std::string HaltReason{};
bool waiting{};
uint64_t Instrs{};
gamegear_Machine* bus{};
bool intReq{};
bool eiPending{};
int64_t idx{};
bool usesHL{};
uint16_t disp{};
bool dispSet{};
};
struct gamegear_VDP{
std::array<uint8_t,16384> VRAM{};
std::array<uint8_t,64> CRAM{};
std::array<uint8_t,16> Regs{};
uint16_t addr{};
uint8_t code{};
uint8_t latch{};
bool latched{};
uint8_t readBuf{};
uint8_t status{};
uint8_t line{};
std::array<uint32_t,16> Writes{};
uint32_t CRAMWrites{};
Slice<std::array<uint8_t,2>>* CRAMLineLog{};
};
struct gamegear_CapEntry{
uint16_t HL{};
uint16_t BC{};
uint16_t DE{};
int64_t Slot1{};
};
struct gamegear_PSG{
std::array<uint16_t,8> Reg{};
int64_t latch{};
};
struct gamegear_Machine{
z80_CPU* CPU{};
gamegear_VDP VDP{};
gamegear_PSG PSG{};
Slice<uint8_t> rom{};
int64_t nbanks{};
std::array<uint8_t,8192> ram{};
std::array<int64_t,3> slot{};
uint8_t Pad00{};
uint8_t PadDC{};
bool Sample{};
Map<uint16_t,int64_t> PCHist{};
uint16_t RAMWatchLo{};
uint16_t RAMWatchHi{};
Map<uint16_t,int64_t> RAMWatchPCs{};
std::function<void(uint16_t,uint16_t,uint8_t)> WriteHook{};
uint16_t stepPC{};
uint16_t CapturePC{};
uint16_t CapHL{};
uint16_t CapBC{};
int64_t CapSlot1{};
bool Captured{};
Slice<gamegear_CapEntry> CapLog{};
uint16_t CapLo{};
uint16_t CapHi{};
uint16_t CapOutBase{};
std::array<uint8_t,4096> CapOut{};
bool CapOutDone{};
uint16_t WatchLo{};
uint16_t WatchHi{};
Map<uint16_t,int64_t> WatchPCs{};
};
struct Anon0{uint8_t A{};uint8_t F{};uint8_t B{};uint8_t C{};uint8_t D{};uint8_t E{};uint8_t H{};uint8_t L{};uint8_t A2{};uint8_t F2{};uint8_t B2{};uint8_t C2{};uint8_t D2{};uint8_t E2{};uint8_t H2{};uint8_t L2{};uint16_t IX{};uint16_t IY{};uint16_t SP{};uint16_t PC{};uint8_t I{};uint8_t R{};bool IFF1{};bool IFF2{};uint8_t IM{};bool Halted{};std::string HaltReason{};bool waiting{};uint64_t Instrs{};gamegear_Machine* bus{};bool intReq{};bool eiPending{};int64_t idx{};bool usesHL{};uint16_t disp{};bool dispSet{};};
struct Anon1{std::array<uint8_t,16384> VRAM{};std::array<uint8_t,64> CRAM{};std::array<uint8_t,16> Regs{};uint16_t addr{};uint8_t code{};uint8_t latch{};bool latched{};uint8_t readBuf{};uint8_t status{};uint8_t line{};std::array<uint32_t,16> Writes{};uint32_t CRAMWrites{};Slice<std::array<uint8_t,2>>* CRAMLineLog{};};
struct Anon2{uint16_t HL{};uint16_t BC{};uint16_t DE{};int64_t Slot1{};};
struct Anon3{z80_CPU* CPU{};gamegear_VDP VDP{};gamegear_PSG PSG{};Slice<uint8_t> rom{};int64_t nbanks{};std::array<uint8_t,8192> ram{};std::array<int64_t,3> slot{};uint8_t Pad00{};uint8_t PadDC{};bool Sample{};Map<uint16_t,int64_t> PCHist{};uint16_t RAMWatchLo{};uint16_t RAMWatchHi{};Map<uint16_t,int64_t> RAMWatchPCs{};std::function<void(uint16_t,uint16_t,uint8_t)> WriteHook{};uint16_t stepPC{};uint16_t CapturePC{};uint16_t CapHL{};uint16_t CapBC{};int64_t CapSlot1{};bool Captured{};Slice<gamegear_CapEntry> CapLog{};uint16_t CapLo{};uint16_t CapHi{};uint16_t CapOutBase{};std::array<uint8_t,4096> CapOut{};bool CapOutDone{};uint16_t WatchLo{};uint16_t WatchHi{};Map<uint16_t,int64_t> WatchPCs{};};
struct Anon4{std::array<uint16_t,8> Reg{};int64_t latch{};};

#include "adapters-decl.h"
z80_CPU* z80_NewCPU(gamegear_Machine* bus);
void z80_CPU_RequestIRQ(z80_CPU* c,bool b);
uint8_t z80_CPU_read(z80_CPU* c,uint16_t a);
void z80_CPU_write(z80_CPU* c,uint16_t a,uint8_t v);
uint16_t z80_CPU_read16(z80_CPU* c,uint16_t a);
void z80_CPU_write16(z80_CPU* c,uint16_t a,uint16_t v);
uint8_t z80_CPU_fetch(z80_CPU* c);
uint16_t z80_CPU_fetch16(z80_CPU* c);
void z80_CPU_push16(z80_CPU* c,uint16_t v);
uint16_t z80_CPU_pop16(z80_CPU* c);
uint16_t z80_CPU_bc(z80_CPU* c);
uint16_t z80_CPU_de(z80_CPU* c);
uint16_t z80_CPU_hl(z80_CPU* c);
uint16_t z80_CPU_af(z80_CPU* c);
void z80_CPU_setBC(z80_CPU* c,uint16_t v);
void z80_CPU_setDE(z80_CPU* c,uint16_t v);
void z80_CPU_setHL(z80_CPU* c,uint16_t v);
void z80_CPU_setAF(z80_CPU* c,uint16_t v);
uint16_t z80_CPU_idxReg(z80_CPU* c);
void z80_CPU_setIdxReg(z80_CPU* c,uint16_t v);
uint16_t z80_CPU_getRP(z80_CPU* c,int64_t p,bool af);
void z80_CPU_setRP(z80_CPU* c,int64_t p,uint16_t v,bool af);
uint16_t z80_CPU_memAddr(z80_CPU* c);
uint8_t z80_CPU_getR(z80_CPU* c,int64_t i);
void z80_CPU_setR(z80_CPU* c,int64_t i,uint8_t v);
bool z80_CPU_getf(z80_CPU* c,uint8_t b);
void z80_CPU_setf(z80_CPU* c,uint8_t b,bool on);
bool z80_parity(uint8_t v);
void z80_CPU_szxy(z80_CPU* c,uint8_t v);
void z80_CPU_add8(z80_CPU* c,uint8_t n,bool carry);
uint8_t z80_CPU_sub8(z80_CPU* c,uint8_t n,bool carry,bool store);
void z80_CPU_and8(z80_CPU* c,uint8_t n);
void z80_CPU_or8(z80_CPU* c,uint8_t n);
void z80_CPU_xor8(z80_CPU* c,uint8_t n);
void z80_CPU_alu(z80_CPU* c,int64_t y,uint8_t n);
uint8_t z80_CPU_inc8(z80_CPU* c,uint8_t v);
uint8_t z80_CPU_dec8(z80_CPU* c,uint8_t v);
uint16_t z80_CPU_add16(z80_CPU* c,uint16_t a,uint16_t b);
uint16_t z80_CPU_adc16(z80_CPU* c,uint16_t a,uint16_t b);
uint16_t z80_CPU_sbc16(z80_CPU* c,uint16_t a,uint16_t b);
uint8_t z80_CPU_rot(z80_CPU* c,int64_t y,uint8_t v);
void z80_CPU_Step(z80_CPU* c);
void z80_CPU_Step_Reference(z80_CPU* c);
void z80_CPU_exec(z80_CPU* c,uint8_t op);
void z80_CPU_execMain(z80_CPU* c,uint8_t op);
void z80_CPU_execX0(z80_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q);
void z80_CPU_accOp(z80_CPU* c,int64_t y);
void z80_CPU_daa(z80_CPU* c);
void z80_CPU_execX3(z80_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q);
bool z80_CPU_cond(z80_CPU* c,int64_t y);
void z80_CPU_execCB(z80_CPU* c,uint8_t op);
void z80_CPU_setRPlain(z80_CPU* c,int64_t i,uint8_t v);
void z80_CPU_execED(z80_CPU* c,uint8_t op);
uint8_t z80_CPU_getRplain(z80_CPU* c,int64_t i);
void z80_CPU_ldAir(z80_CPU* c);
void z80_CPU_rrd(z80_CPU* c);
void z80_CPU_rld(z80_CPU* c);
void z80_CPU_block(z80_CPU* c,int64_t kind,int64_t mode);
bool z80_mainUsesHL(uint8_t x,uint8_t y,uint8_t z);
std::tuple<int64_t,bool> gamegear_FileOffset(std::array<int64_t,3> slots,uint16_t addr);
Slice<uint8_t> gamegear_BankView(Slice<uint8_t> rom,std::array<int64_t,3> slots);
void gamegear_VDP_ResetWrites(gamegear_VDP* v);
int64_t gamegear_VDP_ActiveSprites(gamegear_VDP* v);
void gamegear_VDP_writeControl(gamegear_VDP* v,uint8_t b);
void gamegear_VDP_writeData(gamegear_VDP* v,uint8_t b);
uint8_t gamegear_VDP_readData(gamegear_VDP* v);
uint8_t gamegear_VDP_readStatus(gamegear_VDP* v);
void gamegear_Machine_Watch(gamegear_Machine* m,uint16_t lo,uint16_t hi);
gamegear_Machine* gamegear_NewMachine(Slice<uint8_t> rom);
uint8_t gamegear_Machine_Read(gamegear_Machine* m,uint16_t a);
void gamegear_Machine_Write(gamegear_Machine* m,uint16_t a,uint8_t v);
uint8_t gamegear_Machine_In(gamegear_Machine* m,uint16_t port);
uint8_t gamegear_Machine_In_Reference(gamegear_Machine* m,uint16_t port);
void gamegear_Machine_Out(gamegear_Machine* m,uint16_t port,uint8_t v);
void gamegear_Machine_Out_Reference(gamegear_Machine* m,uint16_t port,uint8_t v);
bool gamegear_Machine_RunFrame(gamegear_Machine* m);
bool gamegear_Machine_step(gamegear_Machine* m);
void gamegear_PSG_Write(gamegear_PSG* p,uint8_t v);
constexpr int64_t z80_flagC=1ULL;
constexpr int64_t z80_flagN=2ULL;
constexpr int64_t z80_flagP=4ULL;
constexpr int64_t z80_flagX=8ULL;
constexpr int64_t z80_flagH=16ULL;
constexpr int64_t z80_flagY=32ULL;
constexpr int64_t z80_flagZ=64ULL;
constexpr int64_t z80_flagS=128ULL;
constexpr double gamegear_PSGClock=3.57954500000000000e+06;
std::array<double,16> gamegear_psgVolume=[]()->std::array<double,16>{
std::array<double,16> t={};
double a = 1.0;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(15ULL));i++){
t[i] = a;
a *= 0.79432823;
}
}t[cast<int64_t>(15ULL)] = cast<double>(0ULL);
return t;
}
();

#include "adapters.h"
// tools/cpu/z80/cpu.go:57:1
z80_CPU* z80_NewCPU(gamegear_Machine* bus){
{
return arenaNew(z80_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},bus,{},{},{},{},{},{}});
}
}
// tools/cpu/z80/cpu.go:66:1
void z80_CPU_RequestIRQ(z80_CPU* c,bool b){
{
c->intReq = b;
}
}
// tools/cpu/z80/cpu.go:68:1
uint8_t z80_CPU_read(z80_CPU* c,uint16_t a){
{auto value = [&]()->uint8_t {
return gamegear_Machine_Read(c->bus,a);
}
(); rrAfterRead(c->bus,a,value); return value;}
}
// tools/cpu/z80/cpu.go:69:1
void z80_CPU_write(z80_CPU* c,uint16_t a,uint8_t v){
{
gamegear_Machine_Write(c->bus,a,v);
}
}
// tools/cpu/z80/cpu.go:70:1
uint16_t z80_CPU_read16(z80_CPU* c,uint16_t a){
{
return cast<uint16_t>((cast<uint16_t>(z80_CPU_read(c,a)) | shl<uint16_t>(cast<uint16_t>(z80_CPU_read(c,cast<uint16_t>((a + cast<uint16_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/cpu/z80/cpu.go:71:1
void z80_CPU_write16(z80_CPU* c,uint16_t a,uint16_t v){
{
z80_CPU_write(c,a,cast<uint8_t>(v));
z80_CPU_write(c,cast<uint16_t>((a + cast<uint16_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/cpu/z80/cpu.go:72:1
uint8_t z80_CPU_fetch(z80_CPU* c){
{
uint8_t v = z80_CPU_read(c,c->PC);
c->PC++;
return v;
}
}
// tools/cpu/z80/cpu.go:73:1
uint16_t z80_CPU_fetch16(z80_CPU* c){
{
uint16_t v = z80_CPU_read16(c,c->PC);
c->PC += cast<uint16_t>(2ULL);
return v;
}
}
// tools/cpu/z80/cpu.go:75:1
void z80_CPU_push16(z80_CPU* c,uint16_t v){
{
c->SP -= cast<uint16_t>(2ULL);
z80_CPU_write16(c,c->SP,v);
}
}
// tools/cpu/z80/cpu.go:76:1
uint16_t z80_CPU_pop16(z80_CPU* c){
{
uint16_t v = z80_CPU_read16(c,c->SP);
c->SP += cast<uint16_t>(2ULL);
return v;
}
}
// tools/cpu/z80/cpu.go:80:1
uint16_t z80_CPU_bc(z80_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->B),cast<int64_t>(8ULL)) | cast<uint16_t>(c->C)));
}
}
// tools/cpu/z80/cpu.go:81:1
uint16_t z80_CPU_de(z80_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->D),cast<int64_t>(8ULL)) | cast<uint16_t>(c->E)));
}
}
// tools/cpu/z80/cpu.go:82:1
uint16_t z80_CPU_hl(z80_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->H),cast<int64_t>(8ULL)) | cast<uint16_t>(c->L)));
}
}
// tools/cpu/z80/cpu.go:83:1
uint16_t z80_CPU_af(z80_CPU* c){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->A),cast<int64_t>(8ULL)) | cast<uint16_t>(c->F)));
}
}
// tools/cpu/z80/cpu.go:84:1
void z80_CPU_setBC(z80_CPU* c,uint16_t v){
{
auto tmp1 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->B = std::get<0>(tmp1);
c->C = std::get<1>(tmp1);
}
}
// tools/cpu/z80/cpu.go:85:1
void z80_CPU_setDE(z80_CPU* c,uint16_t v){
{
auto tmp2 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->D = std::get<0>(tmp2);
c->E = std::get<1>(tmp2);
}
}
// tools/cpu/z80/cpu.go:86:1
void z80_CPU_setHL(z80_CPU* c,uint16_t v){
{
auto tmp3 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->H = std::get<0>(tmp3);
c->L = std::get<1>(tmp3);
}
}
// tools/cpu/z80/cpu.go:87:1
void z80_CPU_setAF(z80_CPU* c,uint16_t v){
{
auto tmp4 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
c->A = std::get<0>(tmp4);
c->F = std::get<1>(tmp4);
}
}
// tools/cpu/z80/cpu.go:90:1
uint16_t z80_CPU_idxReg(z80_CPU* c){
{
{
switch(c->idx){
case cast<int64_t>(1ULL):{
return c->IX;
break;}
case cast<int64_t>(2ULL):{
return c->IY;
break;}
}}
return z80_CPU_hl(c);
}
}
// tools/cpu/z80/cpu.go:99:1
void z80_CPU_setIdxReg(z80_CPU* c,uint16_t v){
{
{
switch(c->idx){
case cast<int64_t>(1ULL):{
c->IX = v;
break;}
case cast<int64_t>(2ULL):{
c->IY = v;
break;}
default:{
z80_CPU_setHL(c,v);
break;}
}}
}
}
// tools/cpu/z80/cpu.go:111:1
uint16_t z80_CPU_getRP(z80_CPU* c,int64_t p,bool af){
{
{
switch(p){
case cast<int64_t>(0ULL):{
return z80_CPU_bc(c);
break;}
case cast<int64_t>(1ULL):{
return z80_CPU_de(c);
break;}
case cast<int64_t>(2ULL):{
return z80_CPU_idxReg(c);
break;}
default:{
if (af) {
return z80_CPU_af(c);
}
return c->SP;
break;}
}}
}
}
// tools/cpu/z80/cpu.go:126:1
void z80_CPU_setRP(z80_CPU* c,int64_t p,uint16_t v,bool af){
{
{
switch(p){
case cast<int64_t>(0ULL):{
z80_CPU_setBC(c,v);
break;}
case cast<int64_t>(1ULL):{
z80_CPU_setDE(c,v);
break;}
case cast<int64_t>(2ULL):{
z80_CPU_setIdxReg(c,v);
break;}
default:{
if (af) {
z80_CPU_setAF(c,v);
}
else {
c->SP = v;
}
break;}
}}
}
}
// tools/cpu/z80/cpu.go:146:1
uint16_t z80_CPU_memAddr(z80_CPU* c){
{
if ((c->idx == cast<int64_t>(0ULL))) {
return z80_CPU_hl(c);
}
if ((!c->dispSet)) {
c->disp = cast<uint16_t>((z80_CPU_idxReg(c) + cast<uint16_t>(cast<int8_t>(z80_CPU_fetch(c)))));
c->dispSet = true;
}
return c->disp;
}
}
// tools/cpu/z80/cpu.go:157:1
uint8_t z80_CPU_getR(z80_CPU* c,int64_t i){
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
if (((c->idx != cast<int64_t>(0ULL)) && (!c->usesHL))) {
return cast<uint8_t>(shr<uint16_t>(z80_CPU_idxReg(c),cast<int64_t>(8ULL)));
}
return c->H;
break;}
case cast<int64_t>(5ULL):{
if (((c->idx != cast<int64_t>(0ULL)) && (!c->usesHL))) {
return cast<uint8_t>(z80_CPU_idxReg(c));
}
return c->L;
break;}
case cast<int64_t>(6ULL):{
return z80_CPU_read(c,z80_CPU_memAddr(c));
break;}
default:{
return c->A;
break;}
}}
}
}
// tools/cpu/z80/cpu.go:184:1
void z80_CPU_setR(z80_CPU* c,int64_t i,uint8_t v){
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
if (((c->idx != cast<int64_t>(0ULL)) && (!c->usesHL))) {
z80_CPU_setIdxReg(c,cast<uint16_t>((cast<uint16_t>((z80_CPU_idxReg(c) & cast<uint16_t>(255ULL))) | shl<uint16_t>(cast<uint16_t>(v),cast<int64_t>(8ULL)))));
return ;
}
c->H = v;
break;}
case cast<int64_t>(5ULL):{
if (((c->idx != cast<int64_t>(0ULL)) && (!c->usesHL))) {
uint16_t r = cast<uint16_t>((cast<uint16_t>((z80_CPU_idxReg(c) & cast<uint16_t>(65280ULL))) | cast<uint16_t>(v)));
z80_CPU_setIdxReg(c,r);
return ;
}
c->L = v;
break;}
case cast<int64_t>(6ULL):{
z80_CPU_write(c,z80_CPU_memAddr(c),v);
break;}
default:{
c->A = v;
break;}
}}
}
}
// tools/cpu/z80/cpu.go:216:1
bool z80_CPU_getf(z80_CPU* c,uint8_t b){
{
return (cast<uint8_t>((c->F & b)) != cast<uint8_t>(0ULL));
}
}
// tools/cpu/z80/cpu.go:217:1
void z80_CPU_setf(z80_CPU* c,uint8_t b,bool on){
{
if (on) {
c->F |= b;
}
else {
c->F &= ~(b);
}
}
}
// tools/cpu/z80/cpu.go:225:1
bool z80_parity(uint8_t v){
{
v ^= shr<uint8_t>(v,cast<int64_t>(4ULL));
v ^= shr<uint8_t>(v,cast<int64_t>(2ULL));
v ^= shr<uint8_t>(v,cast<int64_t>(1ULL));
return (cast<uint8_t>((v & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL));
}
}
// tools/cpu/z80/cpu.go:228:1
void z80_CPU_szxy(z80_CPU* c,uint8_t v){
{
c->F &= ~(cast<uint8_t>(232ULL));
c->F |= cast<uint8_t>((v & cast<uint8_t>(168ULL)));
if ((v == cast<uint8_t>(0ULL))) {
c->F |= cast<uint8_t>(64ULL);
}
}
}
// tools/cpu/z80/cpu.go:238:1
void z80_CPU_add8(z80_CPU* c,uint8_t n,bool carry){
{
uint8_t cin = cast<uint8_t>(0ULL);
if ((carry && z80_CPU_getf(c,cast<uint8_t>(1ULL)))) {
cin = cast<uint8_t>(1ULL);
}
uint16_t r = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(c->A) + cast<uint16_t>(n))) + cast<uint16_t>(cin)));
uint8_t res = cast<uint8_t>(r);
c->F = cast<uint8_t>(0ULL);
z80_CPU_szxy(c,res);
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint8_t>((cast<uint8_t>(((cast<uint8_t>((c->A & cast<uint8_t>(15ULL)))) + (cast<uint8_t>((n & cast<uint8_t>(15ULL)))))) + cin)) > cast<uint8_t>(15ULL)));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(r > cast<uint16_t>(255ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),((cast<uint8_t>(((cast<uint8_t>((c->A ^ n))) & cast<uint8_t>(128ULL))) == cast<uint8_t>(0ULL)) && (cast<uint8_t>(((cast<uint8_t>((c->A ^ res))) & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))));
c->A = res;
}
}
// tools/cpu/z80/cpu.go:253:1
uint8_t z80_CPU_sub8(z80_CPU* c,uint8_t n,bool carry,bool store){
{
uint8_t cin = cast<uint8_t>(0ULL);
if ((carry && z80_CPU_getf(c,cast<uint8_t>(1ULL)))) {
cin = cast<uint8_t>(1ULL);
}
int64_t r = cast<int64_t>((cast<int64_t>((cast<int64_t>(c->A) - cast<int64_t>(n))) - cast<int64_t>(cin)));
uint8_t res = cast<uint8_t>(r);
c->F = cast<uint8_t>(2ULL);
z80_CPU_szxy(c,res);
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint8_t>((c->A & cast<uint8_t>(15ULL)))) - cast<int64_t>(cast<uint8_t>((n & cast<uint8_t>(15ULL)))))) - cast<int64_t>(cin))) < cast<int64_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(r < cast<int64_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),((cast<uint8_t>(((cast<uint8_t>((c->A ^ n))) & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)) && (cast<uint8_t>(((cast<uint8_t>((c->A ^ res))) & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))));
if ((!store)) {
c->F &= ~(cast<uint8_t>(40ULL));
c->F |= cast<uint8_t>((n & cast<uint8_t>(40ULL)));
}
return res;
}
}
// tools/cpu/z80/cpu.go:272:1
void z80_CPU_and8(z80_CPU* c,uint8_t n){
{
c->A &= n;
c->F = cast<uint8_t>(16ULL);
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(c->A));
}
}
// tools/cpu/z80/cpu.go:273:1
void z80_CPU_or8(z80_CPU* c,uint8_t n){
{
c->A |= n;
c->F = cast<uint8_t>(0ULL);
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(c->A));
}
}
// tools/cpu/z80/cpu.go:274:1
void z80_CPU_xor8(z80_CPU* c,uint8_t n){
{
c->A ^= n;
c->F = cast<uint8_t>(0ULL);
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(c->A));
}
}
// tools/cpu/z80/cpu.go:277:1
void z80_CPU_alu(z80_CPU* c,int64_t y,uint8_t n){
{
{
switch(y){
case cast<int64_t>(0ULL):{
z80_CPU_add8(c,n,false);
break;}
case cast<int64_t>(1ULL):{
z80_CPU_add8(c,n,true);
break;}
case cast<int64_t>(2ULL):{
c->A = z80_CPU_sub8(c,n,false,true);
break;}
case cast<int64_t>(3ULL):{
c->A = z80_CPU_sub8(c,n,true,true);
break;}
case cast<int64_t>(4ULL):{
z80_CPU_and8(c,n);
break;}
case cast<int64_t>(5ULL):{
z80_CPU_xor8(c,n);
break;}
case cast<int64_t>(6ULL):{
z80_CPU_or8(c,n);
break;}
default:{
z80_CPU_sub8(c,n,false,false);
break;}
}}
}
}
// tools/cpu/z80/cpu.go:298:1
uint8_t z80_CPU_inc8(z80_CPU* c,uint8_t v){
{
uint8_t r = cast<uint8_t>((v + cast<uint8_t>(1ULL)));
c->F &= ~(cast<uint8_t>(254ULL));
z80_CPU_szxy(c,r);
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint8_t>((v & cast<uint8_t>(15ULL))) == cast<uint8_t>(15ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),(v == cast<uint8_t>(127ULL)));
return r;
}
}
// tools/cpu/z80/cpu.go:306:1
uint8_t z80_CPU_dec8(z80_CPU* c,uint8_t v){
{
uint8_t r = cast<uint8_t>((v - cast<uint8_t>(1ULL)));
c->F = cast<uint8_t>((((c->F & ~(cast<uint8_t>(252ULL)))) | cast<uint8_t>(2ULL)));
z80_CPU_szxy(c,r);
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint8_t>((v & cast<uint8_t>(15ULL))) == cast<uint8_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),(v == cast<uint8_t>(128ULL)));
return r;
}
}
// tools/cpu/z80/cpu.go:317:1
uint16_t z80_CPU_add16(z80_CPU* c,uint16_t a,uint16_t b){
{
uint32_t r = cast<uint32_t>((cast<uint32_t>(a) + cast<uint32_t>(b)));
c->F &= ~(cast<uint8_t>(59ULL));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(r > cast<uint32_t>(65535ULL)));
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint16_t>(((cast<uint16_t>((a & cast<uint16_t>(4095ULL)))) + (cast<uint16_t>((b & cast<uint16_t>(4095ULL)))))) > cast<uint16_t>(4095ULL)));
c->F |= cast<uint8_t>((cast<uint8_t>(shr<uint32_t>(r,cast<int64_t>(8ULL))) & cast<uint8_t>(40ULL)));
return cast<uint16_t>(r);
}
}
// tools/cpu/z80/cpu.go:325:1
uint16_t z80_CPU_adc16(z80_CPU* c,uint16_t a,uint16_t b){
{
uint32_t cin = cast<uint32_t>(0ULL);
if (z80_CPU_getf(c,cast<uint8_t>(1ULL))) {
cin = cast<uint32_t>(1ULL);
}
uint32_t r = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(a) + cast<uint32_t>(b))) + cin));
uint16_t res = cast<uint16_t>(r);
c->F = cast<uint8_t>(0ULL);
z80_CPU_setf(c,cast<uint8_t>(128ULL),(cast<uint16_t>((res & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(64ULL),(res == cast<uint16_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(r > cast<uint32_t>(65535ULL)));
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint16_t>((cast<uint16_t>(((cast<uint16_t>((a & cast<uint16_t>(4095ULL)))) + (cast<uint16_t>((b & cast<uint16_t>(4095ULL)))))) + cast<uint16_t>(cin))) > cast<uint16_t>(4095ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),((cast<uint16_t>(((cast<uint16_t>((a ^ b))) & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL)) && (cast<uint16_t>(((cast<uint16_t>((a ^ res))) & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))));
c->F |= cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(res,cast<int64_t>(8ULL))) & cast<uint8_t>(40ULL)));
return res;
}
}
// tools/cpu/z80/cpu.go:341:1
uint16_t z80_CPU_sbc16(z80_CPU* c,uint16_t a,uint16_t b){
{
uint32_t cin = cast<uint32_t>(0ULL);
if (z80_CPU_getf(c,cast<uint8_t>(1ULL))) {
cin = cast<uint32_t>(1ULL);
}
int32_t r = cast<int32_t>((cast<int32_t>((cast<int32_t>(a) - cast<int32_t>(b))) - cast<int32_t>(cin)));
uint16_t res = cast<uint16_t>(r);
c->F = cast<uint8_t>(2ULL);
z80_CPU_setf(c,cast<uint8_t>(128ULL),(cast<uint16_t>((res & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(64ULL),(res == cast<uint16_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(r < cast<int32_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<int32_t>((cast<int32_t>((cast<int32_t>(cast<uint16_t>((a & cast<uint16_t>(4095ULL)))) - cast<int32_t>(cast<uint16_t>((b & cast<uint16_t>(4095ULL)))))) - cast<int32_t>(cin))) < cast<int32_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),((cast<uint16_t>(((cast<uint16_t>((a ^ b))) & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>(((cast<uint16_t>((a ^ res))) & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))));
c->F |= cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(res,cast<int64_t>(8ULL))) & cast<uint8_t>(40ULL)));
return res;
}
}
// tools/cpu/z80/cpu.go:360:1
uint8_t z80_CPU_rot(z80_CPU* c,int64_t y,uint8_t v){
{
uint8_t r={};
bool cf = z80_CPU_getf(c,cast<uint8_t>(1ULL));
{
switch(y){
case cast<int64_t>(0ULL):{
r = cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(1ULL)) | shr<uint8_t>(v,cast<int64_t>(7ULL))));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(1ULL):{
r = cast<uint8_t>((shr<uint8_t>(v,cast<int64_t>(1ULL)) | shl<uint8_t>(v,cast<int64_t>(7ULL))));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(2ULL):{
r = shl<uint8_t>(v,cast<int64_t>(1ULL));
if (cf) {
r |= cast<uint8_t>(1ULL);
}
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(3ULL):{
r = shr<uint8_t>(v,cast<int64_t>(1ULL));
if (cf) {
r |= cast<uint8_t>(128ULL);
}
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(4ULL):{
r = shl<uint8_t>(v,cast<int64_t>(1ULL));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(5ULL):{
r = cast<uint8_t>((shr<uint8_t>(v,cast<int64_t>(1ULL)) | cast<uint8_t>((v & cast<uint8_t>(128ULL)))));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(6ULL):{
r = cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(1ULL)) | cast<uint8_t>(1ULL)));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
break;}
default:{
r = shr<uint8_t>(v,cast<int64_t>(1ULL));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((v & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
break;}
}}
c->F &= ~(cast<uint8_t>(254ULL));
z80_CPU_szxy(c,r);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(r));
return r;
}
}
// tools/cpu/z80/cpu.go:404:1
void z80_CPU_Step_Reference(z80_CPU* c){
{
if (c->Halted) {
return ;
}
bool servicedEI = c->eiPending;
if (((c->intReq && c->IFF1) && (!c->eiPending))) {
c->waiting = false;
auto tmp5 = std::make_tuple(false,false);
c->IFF1 = std::get<0>(tmp5);
c->IFF2 = std::get<1>(tmp5);
z80_CPU_push16(c,c->PC);
if ((c->IM == cast<uint8_t>(2ULL))) {
c->PC = z80_CPU_read16(c,cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->I),cast<int64_t>(8ULL)) | cast<uint16_t>(255ULL))));
}
else {
c->PC = cast<uint16_t>(56ULL);
}
c->Instrs++;
return ;
}
if (c->waiting) {
c->Instrs++;
return ;
}
c->Instrs++;
auto tmp6 = std::make_tuple(cast<int64_t>(0ULL),false,false);
c->idx = std::get<0>(tmp6);
c->usesHL = std::get<1>(tmp6);
c->dispSet = std::get<2>(tmp6);
z80_CPU_exec(c,z80_CPU_fetch(c));
if (servicedEI) {
c->eiPending = false;
}
}
}
// tools/cpu/z80/cpu.go:435:1
void z80_CPU_exec(z80_CPU* c,uint8_t op){
{
{
switch(op){
case cast<uint8_t>(203ULL):{
z80_CPU_execCB(c,z80_CPU_fetch(c));
break;}
case cast<uint8_t>(237ULL):{
z80_CPU_execED(c,z80_CPU_fetch(c));
break;}
case cast<uint8_t>(221ULL):case cast<uint8_t>(253ULL):{
c->idx = cast<int64_t>(1ULL);
if ((op == cast<uint8_t>(253ULL))) {
c->idx = cast<int64_t>(2ULL);
}
uint8_t nb = z80_CPU_fetch(c);
if ((nb == cast<uint8_t>(203ULL))) {
c->disp = cast<uint16_t>((z80_CPU_idxReg(c) + cast<uint16_t>(cast<int8_t>(z80_CPU_fetch(c)))));
auto tmp7 = std::make_tuple(true,true);
c->dispSet = std::get<0>(tmp7);
c->usesHL = std::get<1>(tmp7);
z80_CPU_execCB(c,z80_CPU_fetch(c));
return ;
}
z80_CPU_execMain(c,nb);
break;}
default:{
z80_CPU_execMain(c,op);
break;}
}}
}
}
// tools/cpu/z80/cpu.go:460:1
void z80_CPU_execMain(z80_CPU* c,uint8_t op){
{
auto tmp8 = std::make_tuple(shr<uint8_t>(op,cast<int64_t>(6ULL)),cast<int64_t>((cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))) & cast<int64_t>(7ULL))),cast<int64_t>((cast<int64_t>(op) & cast<int64_t>(7ULL))));
uint8_t x = std::get<0>(tmp8);
int64_t y = std::get<1>(tmp8);
int64_t z = std::get<2>(tmp8);
auto tmp9 = std::make_tuple(shr<int64_t>(y,cast<int64_t>(1ULL)),cast<int64_t>((y & cast<int64_t>(1ULL))));
int64_t p = std::get<0>(tmp9);
int64_t q = std::get<1>(tmp9);
c->usesHL = ((c->idx != cast<int64_t>(0ULL)) && z80_mainUsesHL(cast<uint8_t>(x),cast<uint8_t>(y),cast<uint8_t>(z)));
{
switch(x){
case cast<uint8_t>(0ULL):{
z80_CPU_execX0(c,y,z,p,q);
break;}
case cast<uint8_t>(1ULL):{
if (((z == cast<int64_t>(6ULL)) && (y == cast<int64_t>(6ULL)))) {
c->waiting = true;
return ;
}
z80_CPU_setR(c,y,z80_CPU_getR(c,z));
break;}
case cast<uint8_t>(2ULL):{
z80_CPU_alu(c,y,z80_CPU_getR(c,z));
break;}
default:{
z80_CPU_execX3(c,y,z,p,q);
break;}
}}
}
}
// tools/cpu/z80/cpu.go:481:1
void z80_CPU_execX0(z80_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q){
{
{
switch(z){
case cast<int64_t>(0ULL):{
{
switch(y){
case cast<int64_t>(0ULL):{
break;}
case cast<int64_t>(1ULL):{
auto tmp10 = std::make_tuple(c->A2,c->A);
c->A = std::get<0>(tmp10);
c->A2 = std::get<1>(tmp10);
auto tmp11 = std::make_tuple(c->F2,c->F);
c->F = std::get<0>(tmp11);
c->F2 = std::get<1>(tmp11);
break;}
case cast<int64_t>(2ULL):{
int8_t off = cast<int8_t>(z80_CPU_fetch(c));
c->B--;
if ((c->B != cast<uint8_t>(0ULL))) {
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(c->PC) + cast<int64_t>(off))));
}
break;}
case cast<int64_t>(3ULL):{
int8_t off = cast<int8_t>(z80_CPU_fetch(c));
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(c->PC) + cast<int64_t>(off))));
break;}
default:{
int8_t off = cast<int8_t>(z80_CPU_fetch(c));
if (z80_CPU_cond(c,cast<int64_t>((y - cast<int64_t>(4ULL))))) {
c->PC = cast<uint16_t>(cast<int64_t>((cast<int64_t>(c->PC) + cast<int64_t>(off))));
}
break;}
}}
break;}
case cast<int64_t>(1ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_setRP(c,p,z80_CPU_fetch16(c),false);
}
else {
z80_CPU_setIdxReg(c,z80_CPU_add16(c,z80_CPU_idxReg(c),z80_CPU_getRP(c,p,false)));
}
break;}
case cast<int64_t>(2ULL):{
{
if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(0ULL)))){
z80_CPU_write(c,z80_CPU_bc(c),c->A);
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(1ULL)))){
z80_CPU_write(c,z80_CPU_de(c),c->A);
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(2ULL)))){
z80_CPU_write16(c,z80_CPU_fetch16(c),z80_CPU_idxReg(c));
}
else if (((q == cast<int64_t>(0ULL)) && (p == cast<int64_t>(3ULL)))){
z80_CPU_write(c,z80_CPU_fetch16(c),c->A);
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(0ULL)))){
c->A = z80_CPU_read(c,z80_CPU_bc(c));
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(1ULL)))){
c->A = z80_CPU_read(c,z80_CPU_de(c));
}
else if (((q == cast<int64_t>(1ULL)) && (p == cast<int64_t>(2ULL)))){
z80_CPU_setIdxReg(c,z80_CPU_read16(c,z80_CPU_fetch16(c)));
}
else {
c->A = z80_CPU_read(c,z80_CPU_fetch16(c));
}
}
tmp12:;
break;}
case cast<int64_t>(3ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_setRP(c,p,cast<uint16_t>((z80_CPU_getRP(c,p,false) + cast<uint16_t>(1ULL))),false);
}
else {
z80_CPU_setRP(c,p,cast<uint16_t>((z80_CPU_getRP(c,p,false) - cast<uint16_t>(1ULL))),false);
}
break;}
case cast<int64_t>(4ULL):{
z80_CPU_setR(c,y,z80_CPU_inc8(c,z80_CPU_getR(c,y)));
break;}
case cast<int64_t>(5ULL):{
z80_CPU_setR(c,y,z80_CPU_dec8(c,z80_CPU_getR(c,y)));
break;}
case cast<int64_t>(6ULL):{
if (((y == cast<int64_t>(6ULL)) && (c->idx != cast<int64_t>(0ULL)))) {
z80_CPU_memAddr(c);
}
z80_CPU_setR(c,y,z80_CPU_fetch(c));
break;}
default:{
z80_CPU_accOp(c,y);
break;}
}}
}
}
// tools/cpu/z80/cpu.go:552:1
void z80_CPU_accOp(z80_CPU* c,int64_t y){
{
{
switch(y){
case cast<int64_t>(0ULL):{
c->A = cast<uint8_t>((shl<uint8_t>(c->A,cast<int64_t>(1ULL)) | shr<uint8_t>(c->A,cast<int64_t>(7ULL))));
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(59ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((c->A & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
break;}
case cast<int64_t>(1ULL):{
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((c->A & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
c->A = cast<uint8_t>((shr<uint8_t>(c->A,cast<int64_t>(1ULL)) | shl<uint8_t>(c->A,cast<int64_t>(7ULL))));
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(58ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
case cast<int64_t>(2ULL):{
bool cf = z80_CPU_getf(c,cast<uint8_t>(1ULL));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((c->A & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
c->A = shl<uint8_t>(c->A,cast<int64_t>(1ULL));
if (cf) {
c->A |= cast<uint8_t>(1ULL);
}
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(58ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
case cast<int64_t>(3ULL):{
bool cf = z80_CPU_getf(c,cast<uint8_t>(1ULL));
z80_CPU_setf(c,cast<uint8_t>(1ULL),(cast<uint8_t>((c->A & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)));
c->A = shr<uint8_t>(c->A,cast<int64_t>(1ULL));
if (cf) {
c->A |= cast<uint8_t>(128ULL);
}
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(58ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
case cast<int64_t>(4ULL):{
z80_CPU_daa(c);
break;}
case cast<int64_t>(5ULL):{
c->A = cast<uint8_t>(~c->A);
c->F |= cast<uint8_t>(18ULL);
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(40ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
case cast<int64_t>(6ULL):{
c->F &= ~(cast<uint8_t>(18ULL));
c->F |= cast<uint8_t>(1ULL);
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(40ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
default:{
z80_CPU_setf(c,cast<uint8_t>(16ULL),z80_CPU_getf(c,cast<uint8_t>(1ULL)));
c->F ^= cast<uint8_t>(1ULL);
c->F &= ~(cast<uint8_t>(2ULL));
c->F = cast<uint8_t>(((c->F & ~(cast<uint8_t>(40ULL))) | cast<uint8_t>((c->A & cast<uint8_t>(40ULL)))));
break;}
}}
}
}
// tools/cpu/z80/cpu.go:596:1
void z80_CPU_daa(z80_CPU* c){
{
uint8_t a = c->A;
uint8_t add={};
bool carry = z80_CPU_getf(c,cast<uint8_t>(1ULL));
if ((z80_CPU_getf(c,cast<uint8_t>(16ULL)) || (cast<uint8_t>((a & cast<uint8_t>(15ULL))) > cast<uint8_t>(9ULL)))) {
add |= cast<uint8_t>(6ULL);
}
if ((carry || (a > cast<uint8_t>(153ULL)))) {
add |= cast<uint8_t>(96ULL);
carry = true;
}
if (z80_CPU_getf(c,cast<uint8_t>(2ULL))) {
z80_CPU_setf(c,cast<uint8_t>(16ULL),(z80_CPU_getf(c,cast<uint8_t>(16ULL)) && (cast<uint8_t>((a & cast<uint8_t>(15ULL))) < cast<uint8_t>(6ULL))));
a -= add;
}
else {
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<uint8_t>((a & cast<uint8_t>(15ULL))) > cast<uint8_t>(9ULL)));
a += add;
}
c->A = a;
z80_CPU_szxy(c,a);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(a));
z80_CPU_setf(c,cast<uint8_t>(1ULL),carry);
}
}
// tools/cpu/z80/cpu.go:620:1
void z80_CPU_execX3(z80_CPU* c,int64_t y,int64_t z,int64_t p,int64_t q){
{
{
switch(z){
case cast<int64_t>(0ULL):{
if (z80_CPU_cond(c,y)) {
c->PC = z80_CPU_pop16(c);
}
break;}
case cast<int64_t>(1ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_setRP(c,p,z80_CPU_pop16(c),true);
}
else {
{
switch(p){
case cast<int64_t>(0ULL):{
c->PC = z80_CPU_pop16(c);
break;}
case cast<int64_t>(1ULL):{
auto tmp13 = std::make_tuple(c->B2,c->B);
c->B = std::get<0>(tmp13);
c->B2 = std::get<1>(tmp13);
auto tmp14 = std::make_tuple(c->C2,c->C);
c->C = std::get<0>(tmp14);
c->C2 = std::get<1>(tmp14);
auto tmp15 = std::make_tuple(c->D2,c->D);
c->D = std::get<0>(tmp15);
c->D2 = std::get<1>(tmp15);
auto tmp16 = std::make_tuple(c->E2,c->E);
c->E = std::get<0>(tmp16);
c->E2 = std::get<1>(tmp16);
auto tmp17 = std::make_tuple(c->H2,c->H);
c->H = std::get<0>(tmp17);
c->H2 = std::get<1>(tmp17);
auto tmp18 = std::make_tuple(c->L2,c->L);
c->L = std::get<0>(tmp18);
c->L2 = std::get<1>(tmp18);
break;}
case cast<int64_t>(2ULL):{
c->PC = z80_CPU_idxReg(c);
break;}
default:{
c->SP = z80_CPU_idxReg(c);
break;}
}}
}
break;}
case cast<int64_t>(2ULL):{
uint16_t a = z80_CPU_fetch16(c);
if (z80_CPU_cond(c,y)) {
c->PC = a;
}
break;}
case cast<int64_t>(3ULL):{
{
switch(y){
case cast<int64_t>(0ULL):{
c->PC = z80_CPU_fetch16(c);
break;}
case cast<int64_t>(2ULL):{
gamegear_Machine_Out(c->bus,cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->A),cast<int64_t>(8ULL)) | cast<uint16_t>(z80_CPU_fetch(c)))),c->A);
break;}
case cast<int64_t>(3ULL):{
c->A = gamegear_Machine_In(c->bus,cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(c->A),cast<int64_t>(8ULL)) | cast<uint16_t>(z80_CPU_fetch(c)))));
break;}
case cast<int64_t>(4ULL):{
uint16_t t = z80_CPU_read16(c,c->SP);
z80_CPU_write16(c,c->SP,z80_CPU_idxReg(c));
z80_CPU_setIdxReg(c,t);
break;}
case cast<int64_t>(5ULL):{
auto tmp19 = std::make_tuple(z80_CPU_de(c),z80_CPU_hl(c));
uint16_t d = std::get<0>(tmp19);
uint16_t h = std::get<1>(tmp19);
z80_CPU_setDE(c,h);
z80_CPU_setHL(c,d);
break;}
case cast<int64_t>(6ULL):{
auto tmp20 = std::make_tuple(false,false);
c->IFF1 = std::get<0>(tmp20);
c->IFF2 = std::get<1>(tmp20);
break;}
default:{
auto tmp21 = std::make_tuple(true,true);
c->IFF1 = std::get<0>(tmp21);
c->IFF2 = std::get<1>(tmp21);
c->eiPending = true;
break;}
}}
break;}
case cast<int64_t>(4ULL):{
uint16_t a = z80_CPU_fetch16(c);
if (z80_CPU_cond(c,y)) {
z80_CPU_push16(c,c->PC);
c->PC = a;
}
break;}
case cast<int64_t>(5ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_push16(c,z80_CPU_getRP(c,p,true));
}
else if ((p == cast<int64_t>(0ULL))) {
uint16_t a = z80_CPU_fetch16(c);
z80_CPU_push16(c,c->PC);
c->PC = a;
}
break;}
case cast<int64_t>(6ULL):{
z80_CPU_alu(c,y,z80_CPU_fetch(c));
break;}
default:{
z80_CPU_push16(c,c->PC);
c->PC = cast<uint16_t>((cast<uint16_t>(y) * cast<uint16_t>(8ULL)));
break;}
}}
}
}
// tools/cpu/z80/cpu.go:696:1
bool z80_CPU_cond(z80_CPU* c,int64_t y){
{
{
switch(y){
case cast<int64_t>(0ULL):{
return (!z80_CPU_getf(c,cast<uint8_t>(64ULL)));
break;}
case cast<int64_t>(1ULL):{
return z80_CPU_getf(c,cast<uint8_t>(64ULL));
break;}
case cast<int64_t>(2ULL):{
return (!z80_CPU_getf(c,cast<uint8_t>(1ULL)));
break;}
case cast<int64_t>(3ULL):{
return z80_CPU_getf(c,cast<uint8_t>(1ULL));
break;}
case cast<int64_t>(4ULL):{
return (!z80_CPU_getf(c,cast<uint8_t>(4ULL)));
break;}
case cast<int64_t>(5ULL):{
return z80_CPU_getf(c,cast<uint8_t>(4ULL));
break;}
case cast<int64_t>(6ULL):{
return (!z80_CPU_getf(c,cast<uint8_t>(128ULL)));
break;}
default:{
return z80_CPU_getf(c,cast<uint8_t>(128ULL));
break;}
}}
}
}
// tools/cpu/z80/cpu.go:717:1
void z80_CPU_execCB(z80_CPU* c,uint8_t op){
{
auto tmp22 = std::make_tuple(shr<uint8_t>(op,cast<int64_t>(6ULL)),cast<int64_t>((cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))) & cast<int64_t>(7ULL))),cast<int64_t>((cast<int64_t>(op) & cast<int64_t>(7ULL))));
uint8_t x = std::get<0>(tmp22);
int64_t y = std::get<1>(tmp22);
int64_t z = std::get<2>(tmp22);
auto get = [&]()->uint8_t{
if ((c->idx != cast<int64_t>(0ULL))) {
return z80_CPU_read(c,c->disp);
}
return z80_CPU_getR(c,z);
}
;
auto put = [&](uint8_t v)->void{
if ((c->idx != cast<int64_t>(0ULL))) {
z80_CPU_write(c,c->disp,v);
if ((z != cast<int64_t>(6ULL))) {
z80_CPU_setRPlain(c,z,v);
}
return ;
}
z80_CPU_setR(c,z,v);
}
;
{
switch(x){
case cast<uint8_t>(0ULL):{
put(z80_CPU_rot(c,y,get()));
break;}
case cast<uint8_t>(1ULL):{
uint8_t v = get();
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(1ULL))) | cast<uint8_t>(16ULL)));
z80_CPU_setf(c,cast<uint8_t>(64ULL),(cast<uint8_t>((v & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y))))) == cast<uint8_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),(cast<uint8_t>((v & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y))))) == cast<uint8_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(128ULL),((y == cast<int64_t>(7ULL)) && (cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))));
c->F |= cast<uint8_t>((v & cast<uint8_t>(40ULL)));
break;}
case cast<uint8_t>(2ULL):{
put((get() & ~((shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y))))));
break;}
default:{
put(cast<uint8_t>((get() | shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(y)))));
break;}
}}
}
}
// tools/cpu/z80/cpu.go:756:1
void z80_CPU_setRPlain(z80_CPU* c,int64_t i,uint8_t v){
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
case cast<int64_t>(7ULL):{
c->A = v;
break;}
}}
}
}
// tools/cpu/z80/cpu.go:775:1
void z80_CPU_execED(z80_CPU* c,uint8_t op){
{
auto tmp23 = std::make_tuple(shr<uint8_t>(op,cast<int64_t>(6ULL)),cast<int64_t>((cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))) & cast<int64_t>(7ULL))),cast<int64_t>((cast<int64_t>(op) & cast<int64_t>(7ULL))));
uint8_t x = std::get<0>(tmp23);
int64_t y = std::get<1>(tmp23);
int64_t z = std::get<2>(tmp23);
auto tmp24 = std::make_tuple(shr<int64_t>(y,cast<int64_t>(1ULL)),cast<int64_t>((y & cast<int64_t>(1ULL))));
int64_t p = std::get<0>(tmp24);
int64_t q = std::get<1>(tmp24);
if ((x == cast<uint8_t>(2ULL))) {
if (((z <= cast<int64_t>(3ULL)) && (y >= cast<int64_t>(4ULL)))) {
z80_CPU_block(c,z,cast<int64_t>((y - cast<int64_t>(4ULL))));
}
return ;
}
if ((x != cast<uint8_t>(1ULL))) {
return ;
}
{
switch(z){
case cast<int64_t>(0ULL):{
uint8_t v = gamegear_Machine_In(c->bus,z80_CPU_bc(c));
if ((y != cast<int64_t>(6ULL))) {
z80_CPU_setRPlain(c,y,v);
}
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(1ULL))) | cast<uint8_t>(0ULL)));
z80_CPU_szxy(c,v);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(v));
break;}
case cast<int64_t>(1ULL):{
if ((y == cast<int64_t>(6ULL))) {
gamegear_Machine_Out(c->bus,z80_CPU_bc(c),cast<uint8_t>(0ULL));
}
else {
gamegear_Machine_Out(c->bus,z80_CPU_bc(c),z80_CPU_getRplain(c,y));
}
break;}
case cast<int64_t>(2ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_setHL(c,z80_CPU_sbc16(c,z80_CPU_hl(c),z80_CPU_getRP(c,cast<int64_t>(p),false)));
}
else {
z80_CPU_setHL(c,z80_CPU_adc16(c,z80_CPU_hl(c),z80_CPU_getRP(c,cast<int64_t>(p),false)));
}
break;}
case cast<int64_t>(3ULL):{
if ((q == cast<int64_t>(0ULL))) {
z80_CPU_write16(c,z80_CPU_fetch16(c),z80_CPU_getRP(c,cast<int64_t>(p),false));
}
else {
z80_CPU_setRP(c,cast<int64_t>(p),z80_CPU_read16(c,z80_CPU_fetch16(c)),false);
}
break;}
case cast<int64_t>(4ULL):{
uint8_t a = c->A;
c->A = cast<uint8_t>(0ULL);
c->A = z80_CPU_sub8(c,a,false,true);
break;}
case cast<int64_t>(5ULL):{
c->PC = z80_CPU_pop16(c);
c->IFF1 = c->IFF2;
break;}
case cast<int64_t>(6ULL):{
c->IM = Slice<uint8_t>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(2ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(2ULL)}[y];
break;}
default:{
{
switch(y){
case cast<int64_t>(0ULL):{
c->I = c->A;
break;}
case cast<int64_t>(1ULL):{
c->R = c->A;
break;}
case cast<int64_t>(2ULL):{
c->A = c->I;
z80_CPU_ldAir(c);
break;}
case cast<int64_t>(3ULL):{
c->A = c->R;
z80_CPU_ldAir(c);
break;}
case cast<int64_t>(4ULL):{
z80_CPU_rrd(c);
break;}
case cast<int64_t>(5ULL):{
z80_CPU_rld(c);
break;}
}}
break;}
}}
}
}
// tools/cpu/z80/cpu.go:845:1
uint8_t z80_CPU_getRplain(z80_CPU* c,int64_t i){
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
default:{
return c->A;
break;}
}}
}
}
// tools/cpu/z80/cpu.go:864:1
void z80_CPU_ldAir(z80_CPU* c){
{
c->F = cast<uint8_t>((c->F & cast<uint8_t>(1ULL)));
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),c->IFF2);
}
}
// tools/cpu/z80/cpu.go:870:1
void z80_CPU_rrd(z80_CPU* c){
{
uint8_t m = z80_CPU_read(c,z80_CPU_hl(c));
z80_CPU_write(c,z80_CPU_hl(c),cast<uint8_t>((shl<uint8_t>(c->A,cast<int64_t>(4ULL)) | shr<uint8_t>(m,cast<int64_t>(4ULL)))));
c->A = cast<uint8_t>((cast<uint8_t>((c->A & cast<uint8_t>(240ULL))) | cast<uint8_t>((m & cast<uint8_t>(15ULL)))));
c->F = cast<uint8_t>((c->F & cast<uint8_t>(1ULL)));
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(c->A));
}
}
// tools/cpu/z80/cpu.go:878:1
void z80_CPU_rld(z80_CPU* c){
{
uint8_t m = z80_CPU_read(c,z80_CPU_hl(c));
z80_CPU_write(c,z80_CPU_hl(c),cast<uint8_t>((shl<uint8_t>(m,cast<int64_t>(4ULL)) | cast<uint8_t>((c->A & cast<uint8_t>(15ULL))))));
c->A = cast<uint8_t>((cast<uint8_t>((c->A & cast<uint8_t>(240ULL))) | shr<uint8_t>(m,cast<int64_t>(4ULL))));
c->F = cast<uint8_t>((c->F & cast<uint8_t>(1ULL)));
z80_CPU_szxy(c,c->A);
z80_CPU_setf(c,cast<uint8_t>(4ULL),z80_parity(c->A));
}
}
// tools/cpu/z80/cpu.go:888:1
void z80_CPU_block(z80_CPU* c,int64_t kind,int64_t mode){
{
bool inc = ((mode == cast<int64_t>(0ULL)) || (mode == cast<int64_t>(2ULL)));
bool repeat = (mode >= cast<int64_t>(2ULL));
{
switch(kind){
case cast<int64_t>(0ULL):{
uint8_t v = z80_CPU_read(c,z80_CPU_hl(c));
z80_CPU_write(c,z80_CPU_de(c),v);
if (inc) {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) + cast<uint16_t>(1ULL))));
z80_CPU_setDE(c,cast<uint16_t>((z80_CPU_de(c) + cast<uint16_t>(1ULL))));
}
else {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) - cast<uint16_t>(1ULL))));
z80_CPU_setDE(c,cast<uint16_t>((z80_CPU_de(c) - cast<uint16_t>(1ULL))));
}
z80_CPU_setBC(c,cast<uint16_t>((z80_CPU_bc(c) - cast<uint16_t>(1ULL))));
c->F &= ~(cast<uint8_t>(22ULL));
z80_CPU_setf(c,cast<uint8_t>(4ULL),(z80_CPU_bc(c) != cast<uint16_t>(0ULL)));
uint8_t n = cast<uint8_t>((v + c->A));
c->F = cast<uint8_t>((cast<uint8_t>(((c->F & ~(cast<uint8_t>(40ULL))) | shl<uint8_t>((cast<uint8_t>((n & cast<uint8_t>(2ULL)))),cast<int64_t>(4ULL)))) | cast<uint8_t>((n & cast<uint8_t>(8ULL)))));
if ((repeat && (z80_CPU_bc(c) != cast<uint16_t>(0ULL)))) {
c->PC -= cast<uint16_t>(2ULL);
}
break;}
case cast<int64_t>(1ULL):{
uint8_t v = z80_CPU_read(c,z80_CPU_hl(c));
uint8_t res = cast<uint8_t>((c->A - v));
if (inc) {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) + cast<uint16_t>(1ULL))));
}
else {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) - cast<uint16_t>(1ULL))));
}
z80_CPU_setBC(c,cast<uint16_t>((z80_CPU_bc(c) - cast<uint16_t>(1ULL))));
c->F = cast<uint8_t>((cast<uint8_t>((c->F & cast<uint8_t>(1ULL))) | cast<uint8_t>(2ULL)));
z80_CPU_setf(c,cast<uint8_t>(128ULL),(cast<uint8_t>((res & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(64ULL),(res == cast<uint8_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(16ULL),(cast<int64_t>((cast<int64_t>(cast<uint8_t>((c->A & cast<uint8_t>(15ULL)))) - cast<int64_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL)))))) < cast<int64_t>(0ULL)));
z80_CPU_setf(c,cast<uint8_t>(4ULL),(z80_CPU_bc(c) != cast<uint16_t>(0ULL)));
if (((repeat && (z80_CPU_bc(c) != cast<uint16_t>(0ULL))) && (res != cast<uint8_t>(0ULL)))) {
c->PC -= cast<uint16_t>(2ULL);
}
break;}
case cast<int64_t>(2ULL):{
uint8_t v = gamegear_Machine_In(c->bus,z80_CPU_bc(c));
z80_CPU_write(c,z80_CPU_hl(c),v);
c->B--;
if (inc) {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) + cast<uint16_t>(1ULL))));
}
else {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) - cast<uint16_t>(1ULL))));
}
z80_CPU_setf(c,cast<uint8_t>(2ULL),true);
z80_CPU_setf(c,cast<uint8_t>(64ULL),(c->B == cast<uint8_t>(0ULL)));
if ((repeat && (c->B != cast<uint8_t>(0ULL)))) {
c->PC -= cast<uint16_t>(2ULL);
}
break;}
default:{
uint8_t v = z80_CPU_read(c,z80_CPU_hl(c));
c->B--;
gamegear_Machine_Out(c->bus,z80_CPU_bc(c),v);
if (inc) {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) + cast<uint16_t>(1ULL))));
}
else {
z80_CPU_setHL(c,cast<uint16_t>((z80_CPU_hl(c) - cast<uint16_t>(1ULL))));
}
z80_CPU_setf(c,cast<uint8_t>(2ULL),true);
z80_CPU_setf(c,cast<uint8_t>(64ULL),(c->B == cast<uint8_t>(0ULL)));
if ((repeat && (c->B != cast<uint8_t>(0ULL)))) {
c->PC -= cast<uint16_t>(2ULL);
}
break;}
}}
}
}
// tools/cpu/z80/decode.go:460:1
bool z80_mainUsesHL(uint8_t x,uint8_t y,uint8_t z){
{
{
switch(x){
case cast<uint8_t>(0ULL):{
return (((((z == cast<uint8_t>(4ULL)) || (z == cast<uint8_t>(5ULL))) || (z == cast<uint8_t>(6ULL)))) && (y == cast<uint8_t>(6ULL)));
break;}
case cast<uint8_t>(1ULL):{
return ((z == cast<uint8_t>(6ULL)) || (y == cast<uint8_t>(6ULL)));
break;}
case cast<uint8_t>(2ULL):{
return (z == cast<uint8_t>(6ULL));
break;}
}}
return false;
}
}
// tools/platform/gamegear/addressing.go:13:1
std::tuple<int64_t,bool> gamegear_FileOffset(std::array<int64_t,3> slots,uint16_t addr){
int64_t off{};
bool inRAM{};
{
{
if ((addr < cast<uint16_t>(1024ULL))){
return {cast<int64_t>(addr),false};
}
else if ((addr < cast<uint16_t>(16384ULL))){
return {cast<int64_t>((cast<int64_t>((slots[cast<int64_t>(0ULL)] * cast<int64_t>(16384ULL))) + cast<int64_t>(addr))),false};
}
else if ((addr < cast<uint16_t>(32768ULL))){
return {cast<int64_t>((cast<int64_t>((slots[cast<int64_t>(1ULL)] * cast<int64_t>(16384ULL))) + cast<int64_t>(cast<uint16_t>((addr - cast<uint16_t>(16384ULL)))))),false};
}
else if ((addr < cast<uint16_t>(49152ULL))){
return {cast<int64_t>((cast<int64_t>((slots[cast<int64_t>(2ULL)] * cast<int64_t>(16384ULL))) + cast<int64_t>(cast<uint16_t>((addr - cast<uint16_t>(32768ULL)))))),false};
}
else {
return {cast<int64_t>(0ULL),true};
}
}
tmp1:;
}
}
// tools/platform/gamegear/addressing.go:32:1
Slice<uint8_t> gamegear_BankView(Slice<uint8_t> rom,std::array<int64_t,3> slots){
{
Slice<uint8_t> view = Slice<uint8_t>::make(cast<int64_t>(49152ULL));
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(49152ULL));a++){
auto tmp2 = gamegear_FileOffset(slots,cast<uint16_t>(a));
int64_t off = std::get<0>(tmp2);
if ((off < len(rom))) {
view[a] = rom[off];
}
else {
view[a] = cast<uint8_t>(255ULL);
}
}
}return view;
}
}
// tools/platform/gamegear/machine.go:44:1
void gamegear_VDP_ResetWrites(gamegear_VDP* v){
{
v->Writes = std::array<uint32_t,16>{};
v->CRAMWrites = cast<uint32_t>(0ULL);
}
}
// tools/platform/gamegear/machine.go:51:1
int64_t gamegear_VDP_ActiveSprites(gamegear_VDP* v){
{
int64_t n = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(64ULL));i++){
{
uint8_t y = v->VRAM[cast<int64_t>((cast<int64_t>(16128ULL) + i))];
if (((y != cast<uint8_t>(208ULL)) && (y != cast<uint8_t>(224ULL)))) {
n++;
}
}
}
}return n;
}
}
// tools/platform/gamegear/machine.go:62:1
void gamegear_VDP_writeControl(gamegear_VDP* v,uint8_t b){
{
if ((!v->latched)) {
v->latch = b;
v->latched = true;
return ;
}
v->latched = false;
v->code = shr<uint8_t>(b,cast<int64_t>(6ULL));
v->addr = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(cast<uint8_t>((b & cast<uint8_t>(63ULL)))),cast<int64_t>(8ULL)) | cast<uint16_t>(v->latch)));
{
switch(v->code){
case cast<uint8_t>(2ULL):{
v->Regs[cast<uint8_t>((b & cast<uint8_t>(15ULL)))] = v->latch;
break;}
case cast<uint8_t>(0ULL):{
v->readBuf = v->VRAM[cast<uint16_t>((v->addr & cast<uint16_t>(16383ULL)))];
v->addr++;
break;}
}}
}
}
// tools/platform/gamegear/machine.go:81:1
void gamegear_VDP_writeData(gamegear_VDP* v,uint8_t b){
{
v->latched = false;
if ((v->code == cast<uint8_t>(3ULL))) {
if (bool(v->CRAMLineLog)) {
(*v->CRAMLineLog) = append((*v->CRAMLineLog),std::array<uint8_t,2>{cast<uint8_t>(cast<uint16_t>((v->addr & cast<uint16_t>(63ULL)))),v->line});
}
v->CRAM[cast<uint16_t>((v->addr & cast<uint16_t>(63ULL)))] = b;
v->CRAMWrites++;
}
else {
uint16_t a = cast<uint16_t>((v->addr & cast<uint16_t>(16383ULL)));
v->VRAM[a] = b;
v->Writes[shr<uint16_t>(a,cast<int64_t>(10ULL))]++;
}
v->addr++;
v->readBuf = b;
}
}
// tools/platform/gamegear/machine.go:100:1
uint8_t gamegear_VDP_readData(gamegear_VDP* v){
{
v->latched = false;
uint8_t r = v->readBuf;
v->readBuf = v->VRAM[cast<uint16_t>((v->addr & cast<uint16_t>(16383ULL)))];
v->addr++;
return r;
}
}
// tools/platform/gamegear/machine.go:110:1
uint8_t gamegear_VDP_readStatus(gamegear_VDP* v){
{
v->latched = false;
uint8_t s = v->status;
v->status &= cast<uint8_t>(31ULL);
return s;
}
}
// tools/platform/gamegear/machine.go:186:1
void gamegear_Machine_Watch(gamegear_Machine* m,uint16_t lo,uint16_t hi){
{
auto tmp3 = std::make_tuple(lo,hi);
m->WatchLo = std::get<0>(tmp3);
m->WatchHi = std::get<1>(tmp3);
m->WatchPCs = Map<uint16_t,int64_t>{};
}
}
// tools/platform/gamegear/machine.go:193:1
gamegear_Machine* gamegear_NewMachine(Slice<uint8_t> rom){
{
gamegear_Machine* m = arenaNew(gamegear_Machine{{},{},{},rom,divi<int64_t>(len(rom),cast<int64_t>(16384ULL)),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
m->slot = std::array<int64_t,3>{cast<int64_t>(0ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL)};
auto tmp4 = std::make_tuple(cast<uint8_t>(255ULL),cast<uint8_t>(255ULL));
m->Pad00 = std::get<0>(tmp4);
m->PadDC = std::get<1>(tmp4);
m->CPU = z80_NewCPU(m);
return m;
}
}
// tools/platform/gamegear/machine.go:204:1
uint8_t gamegear_Machine_Read(gamegear_Machine* m,uint16_t a){
{
auto tmp5 = gamegear_FileOffset(m->slot,a);
int64_t off = std::get<0>(tmp5);
bool inRAM = std::get<1>(tmp5);
if (inRAM) {
return m->ram[cast<uint16_t>((a & cast<uint16_t>(8191ULL)))];
}
if ((off < len(m->rom))) {
return m->rom[off];
}
return cast<uint8_t>(255ULL);
}
}
// tools/platform/gamegear/machine.go:217:1
void gamegear_Machine_Write(gamegear_Machine* m,uint16_t a,uint8_t v){
{auto writtenAddress=a; auto writtenValue=v;
{
if ((a < cast<uint16_t>(49152ULL))) {
return ;
}
if (((bool(m->RAMWatchPCs) && (a >= m->RAMWatchLo)) && (a < m->RAMWatchHi))) {
m->RAMWatchPCs[m->CPU->PC]++;
}
if (bool(m->WriteHook)) {
m->WriteHook(m->stepPC,a,v);
}
m->ram[cast<uint16_t>((a & cast<uint16_t>(8191ULL)))] = v;
{
switch(a){
case cast<uint16_t>(65533ULL):{
m->slot[cast<int64_t>(0ULL)] = modi<int64_t>(cast<int64_t>(v),m->nbanks);
break;}
case cast<uint16_t>(65534ULL):{
m->slot[cast<int64_t>(1ULL)] = modi<int64_t>(cast<int64_t>(v),m->nbanks);
break;}
case cast<uint16_t>(65535ULL):{
m->slot[cast<int64_t>(2ULL)] = modi<int64_t>(cast<int64_t>(v),m->nbanks);
break;}
}}
}
rrAfterWrite(m,writtenAddress,writtenValue);
}
}
// tools/platform/gamegear/machine.go:240:1
uint8_t gamegear_Machine_In_Reference(gamegear_Machine* m,uint16_t port){
{
{
switch(cast<uint8_t>(port)){
case cast<uint8_t>(190ULL):{
return gamegear_VDP_readData(&(m->VDP));
break;}
case cast<uint8_t>(191ULL):{
z80_CPU_RequestIRQ(m->CPU,false);
return gamegear_VDP_readStatus(&(m->VDP));
break;}
case cast<uint8_t>(126ULL):{
return m->VDP.line;
break;}
case cast<uint8_t>(127ULL):{
return cast<uint8_t>(0ULL);
break;}
case cast<uint8_t>(0ULL):{
return m->Pad00;
break;}
case cast<uint8_t>(220ULL):{
return m->PadDC;
break;}
default:{
return cast<uint8_t>(255ULL);
break;}
}}
}
}
// tools/platform/gamegear/machine.go:262:1
void gamegear_Machine_Out_Reference(gamegear_Machine* m,uint16_t port,uint8_t v){
{
{
{
switch(cast<uint8_t>(port)){
case cast<uint8_t>(190ULL):{
if ((bool(m->WatchPCs) && (m->VDP.code != cast<uint8_t>(3ULL)))) {
{
uint16_t a = cast<uint16_t>((m->VDP.addr & cast<uint16_t>(16383ULL)));
if (((a >= m->WatchLo) && (a < m->WatchHi))) {
m->WatchPCs[m->CPU->PC]++;
}
}
}
gamegear_VDP_writeData(&(m->VDP),v);
break;}
case cast<uint8_t>(191ULL):{
gamegear_VDP_writeControl(&(m->VDP),v);
break;}
case cast<uint8_t>(127ULL):{
gamegear_PSG_Write(&(m->PSG),v);
break;}
}}
}
rrAfterPort(m,port,v);
}
}
// tools/platform/gamegear/machine.go:288:1
bool gamegear_Machine_RunFrame(gamegear_Machine* m){
{
constexpr int64_t budget=20000ULL;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(20000ULL));i++){
m->VDP.line = cast<uint8_t>(cast<int64_t>(((divi<int64_t>(cast<int64_t>((i * cast<int64_t>(262ULL))),cast<int64_t>(20000ULL))) & cast<int64_t>(255ULL))));
if ((m->VDP.line >= cast<uint8_t>(192ULL))) {
m->VDP.status |= cast<uint8_t>(128ULL);
}
if ((!gamegear_Machine_step(m))) {
return false;
}
}
}m->VDP.line = cast<uint8_t>(192ULL);
m->VDP.status |= cast<uint8_t>(128ULL);
if ((cast<uint8_t>((m->VDP.Regs[cast<int64_t>(1ULL)] & cast<uint8_t>(32ULL))) != cast<uint8_t>(0ULL))) {
z80_CPU_RequestIRQ(m->CPU,true);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(10000ULL));i++){
if ((!gamegear_Machine_step(m))) {
return false;
}
}
}z80_CPU_RequestIRQ(m->CPU,false);
return true;
}
}
// tools/platform/gamegear/machine.go:320:1
bool gamegear_Machine_step(gamegear_Machine* m){
{
m->stepPC = m->CPU->PC;
z80_CPU_Step(m->CPU);
if (m->Sample) {
m->PCHist[m->CPU->PC]++;
}
if (((m->CapturePC != cast<uint16_t>(0ULL)) && (m->CPU->PC == m->CapturePC))) {
uint16_t hl = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->CPU->H),cast<int64_t>(8ULL)) | cast<uint16_t>(m->CPU->L)));
uint16_t bc = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->CPU->B),cast<int64_t>(8ULL)) | cast<uint16_t>(m->CPU->C)));
uint16_t de = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->CPU->D),cast<int64_t>(8ULL)) | cast<uint16_t>(m->CPU->E)));
m->CapLog = append(m->CapLog,gamegear_CapEntry{hl,bc,de,m->slot[cast<int64_t>(1ULL)]});
if ((!m->Captured)) {
auto tmp6 = std::make_tuple(hl,bc,m->slot[cast<int64_t>(1ULL)],true);
m->CapHL = std::get<0>(tmp6);
m->CapBC = std::get<1>(tmp6);
m->CapSlot1 = std::get<2>(tmp6);
m->Captured = std::get<3>(tmp6);
}
}
if (((((m->CapHi != cast<uint16_t>(0ULL)) && m->Captured) && (!m->CapOutDone)) && (((m->CPU->PC < m->CapLo) || (m->CPU->PC > m->CapHi))))) {
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(4096ULL));j++){
m->CapOut[j] = m->ram[cast<uint16_t>(((cast<uint16_t>((m->CapOutBase + cast<uint16_t>(j)))) & cast<uint16_t>(8191ULL)))];
}
}m->CapOutDone = true;
}
return (!m->CPU->Halted);
}
}
// tools/platform/gamegear/psg.go:18:1
void gamegear_PSG_Write(gamegear_PSG* p,uint8_t v){
{
if ((cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
p->latch = cast<int64_t>(cast<uint8_t>(((shr<uint8_t>(v,cast<int64_t>(4ULL))) & cast<uint8_t>(7ULL))));
if (((cast<int64_t>((p->latch & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL)) && (p->latch != cast<int64_t>(6ULL)))) {
p->Reg[p->latch] = cast<uint16_t>(((cast<uint16_t>((p->Reg[p->latch] & cast<uint16_t>(1008ULL)))) | cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))))));
}
else {
p->Reg[p->latch] = cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
return ;
}
if (((cast<int64_t>((p->latch & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL)) && (p->latch != cast<int64_t>(6ULL)))) {
p->Reg[p->latch] = cast<uint16_t>(((shl<uint16_t>(cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(63ULL)))),cast<int64_t>(4ULL))) | (cast<uint16_t>((p->Reg[p->latch] & cast<uint16_t>(15ULL))))));
}
else {
p->Reg[p->latch] = cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
}
}

#include "fast.h"
