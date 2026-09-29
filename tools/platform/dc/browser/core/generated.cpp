#include "runtime.h"
struct arm_Inst;
struct arm_CPU;
struct arm_Banks;
struct arm_vfpState;
struct sh4_CPU;
struct sh4_onchip_137_gap;
struct sh4_Inst;
struct sh4_State;
struct sh4_TMUChannel;
struct sh4_TMUState;
struct iso9660_Geometry;
struct iso9660_Source;
struct iso9660_Volume;
struct iso9660_Entry;
struct dc_aicaTimers;
struct dc_AICASlot;
struct dc_armBus;
struct dc_gdRequest;
struct dc_biosState;
struct dc_biosHLE;
struct dc_Track;
struct dc_Holly;
struct dc_IPBin;
struct dc_Machine;
struct dc_WatchRange;
struct dc_PadState;
struct dc_mapleState;
struct dc_renderState;
struct dc_pvrVert;
struct dc_Result;
struct dc_RunConfig;
struct Anon0;
struct Anon1;
struct Anon10;
struct Anon11;
struct Anon12;
struct Anon13;
struct Anon14;
struct Anon15;
struct Anon16;
struct Anon17;
struct Anon18;
struct Anon19;
struct Anon2;
struct Anon20;
struct Anon21;
struct Anon22;
struct Anon23;
struct Anon24;
struct Anon25;
struct Anon26;
struct Anon27;
struct Anon28;
struct Anon29;
struct Anon3;
struct Anon30;
struct Anon31;
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
using arm_Flow=int64_t;
struct arm_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
arm_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool Thumb{};
bool TargetThumb{};
int64_t Cond{};
};
using arm_Variant=int64_t;
struct arm_vfpState{
std::array<uint32_t,32> S{};
uint32_t FPSCR{};
uint32_t FPEXC{};
};
struct arm_CPU{
std::array<uint32_t,16> R{};
bool N{};
bool Z{};
bool C{};
bool V{};
bool Q{};
uint32_t GE{};
bool Thumb{};
bool BigEndian{};
bool IRQDisable{};
bool FIQDisable{};
uint32_t Mode{};
arm_Variant Arch{};
arm_vfpState VFP{};
bool exclValid{};
uint32_t exclAddr{};
std::array<uint32_t,6> bankR13{};
std::array<uint32_t,6> bankR14{};
std::array<uint32_t,6> bankSPSR{};
std::array<uint32_t,5> fiqR8_12{};
std::array<uint32_t,5> usrR8_12{};
std::function<bool(arm_CPU*,uint32_t)> SWI{};
std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};
dc_armBus* bus{};
dc_armBus* wide{};
bool Halted{};
std::string HaltReason{};
uint64_t Instrs{};
uint32_t cur{};
bool branched{};
};
struct arm_Banks{
std::array<uint32_t,6> R13{};
std::array<uint32_t,6> R14{};
std::array<uint32_t,6> SPSR{};
std::array<uint32_t,5> FIQR8_12{};
std::array<uint32_t,5> USRR8_12{};
};
using arm_parMode=int64_t;
struct sh4_TMUChannel{
uint32_t TCOR{};
uint32_t TCNT{};
uint16_t TCR{};
uint32_t Frac{};
};
struct sh4_TMUState{
uint8_t TOCR{};
uint8_t TSTR{};
std::array<sh4_TMUChannel,3> Ch{};
};
struct sh4_CPU{
std::array<uint32_t,16> R{};
std::array<uint32_t,8> Rbank{};
uint32_t SR{};
uint32_t GBR{};
uint32_t VBR{};
uint32_t SSR{};
uint32_t SPC{};
uint32_t SGR{};
uint32_t DBR{};
uint32_t MACH{};
uint32_t MACL{};
uint32_t PR{};
uint32_t PC{};
std::array<std::array<uint32_t,16>,2> fpr{};
uint32_t FPSCR{};
uint32_t FPUL{};
uint32_t MMUCR{};
uint32_t CCR{};
uint32_t TRA{};
uint32_t EXPEVT{};
uint32_t INTEVT{};
uint32_t PTEH{};
uint32_t PTEL{};
uint32_t TTB{};
uint32_t TEA{};
uint32_t QACR0{};
uint32_t QACR1{};
uint32_t ICR{};
uint32_t IPRA{};
uint32_t IPRB{};
uint32_t IPRC{};
sh4_TMUState TMU{};
std::array<std::array<uint32_t,8>,2> SQ{};
uint32_t irlLevel{};
uint32_t irlCode{};
Slice<uint8_t> SerialTX{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
dc_Machine* bus{};
dc_Machine* fetcher{};
uint32_t nextPC{};
uint32_t curPC{};
bool delaySlot{};
bool pendingDelay{};
Map<uint32_t,uint32_t> onchipRaw{};
Map<uint32_t,int64_t> onchipGaps{};
};
struct sh4_onchip_137_gap{
uint32_t addr{};
int64_t reads{};
};
using sh4_Flow=int64_t;
struct sh4_Inst{
uint32_t Addr{};
uint16_t Word{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
sh4_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool HasDelay{};
uint32_t LitAddr{};
int64_t LitSize{};
};
struct sh4_State{
std::array<uint32_t,16> R{};
std::array<uint32_t,8> Rbank{};
uint32_t SR{};
uint32_t GBR{};
uint32_t VBR{};
uint32_t SSR{};
uint32_t SPC{};
uint32_t SGR{};
uint32_t DBR{};
uint32_t MACH{};
uint32_t MACL{};
uint32_t PR{};
uint32_t PC{};
uint32_t NextPC{};
std::array<std::array<uint32_t,16>,2> FPR{};
uint32_t FPSCR{};
uint32_t FPUL{};
uint32_t MMUCR{};
uint32_t CCR{};
uint32_t TRA{};
uint32_t EXPEVT{};
uint32_t INTEVT{};
uint32_t PTEH{};
uint32_t PTEL{};
uint32_t TTB{};
uint32_t TEA{};
uint32_t QACR0{};
uint32_t QACR1{};
uint32_t ICR{};
uint32_t IPRA{};
uint32_t IPRB{};
uint32_t IPRC{};
sh4_TMUState TMU{};
std::array<std::array<uint32_t,8>,2> SQ{};
uint32_t IRLLevel{};
uint32_t IRLCode{};
Slice<uint8_t> SerialTX{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
uint32_t CurPC{};
bool DelaySlot{};
bool PendingDelay{};
Map<uint32_t,uint32_t> OnchipRaw{};
Map<uint32_t,int64_t> OnchipGaps{};
};
struct iso9660_Geometry{
int64_t SectorSize{};
int64_t DataOffset{};
};
struct iso9660_Source{
LocalSource* ra{};
int64_t size{};
iso9660_Geometry geom{};
Slice<uint8_t> buf{};
};
struct iso9660_Volume{
dc_Disc* src{};
std::string System{};
std::string Name{};
int64_t Blocks{};
int64_t rootLBA{};
int64_t rootSize{};
};
struct iso9660_Entry{
std::string Name{};
std::string Path{};
bool IsDir{};
int64_t Size{};
int64_t Block{};
};
using iso9660_byteReaderAt=Slice<uint8_t>;
struct dc_aicaTimers{
uint32_t Frac{};
std::array<uint32_t,3> Sub{};
uint32_t SCIPD{};
uint32_t MCIPD{};
uint64_t SampleTick{};
};
struct dc_AICASlot{
bool Active{};
uint32_t Pos{};
uint32_t Frac{};
bool LP{};
uint8_t EGState{};
uint16_t EGLevel{};
int32_t Cur{};
int32_t Prev{};
uint32_t DecPos{};
int32_t AdHist{};
int32_t AdStep{};
int32_t LoopHist{};
int32_t LoopStep{};
bool LoopSeen{};
};
struct dc_armBus{
dc_Machine* m{};
};
struct dc_gdRequest{
uint32_t Cmd{};
std::array<uint32_t,4> Params{};
bool Done{};
std::array<uint32_t,4> Result{};
};
struct dc_biosState{
uint32_t NextID{};
Map<uint32_t,dc_gdRequest*> Requests{};
};
struct dc_biosHLE{
dc_Machine* m{};
dc_biosState state{};
Map<std::string,int64_t> calls{};
};
struct dc_Track{
int64_t Number{};
std::string Mode{};
int64_t FileOffset{};
int64_t Length{};
int64_t StartLBA{};
};
struct dc_Holly{
uint32_t ISTNRM{};
uint32_t ISTEXT{};
uint32_t ISTERR{};
uint32_t IML2NRM{};
uint32_t IML2EXT{};
uint32_t IML2ERR{};
uint32_t IML4NRM{};
uint32_t IML4EXT{};
uint32_t IML4ERR{};
uint32_t IML6NRM{};
uint32_t IML6EXT{};
uint32_t IML6ERR{};
};
struct dc_IPBin{
std::string HardwareID{};
std::string MakerID{};
std::string DeviceInfo{};
std::string AreaSyms{};
std::string Peripherals{};
std::string ProductNo{};
std::string Version{};
std::string ReleaseDate{};
std::string BootFile{};
std::string Company{};
std::string Title{};
};
struct dc_mapleState{
uint32_t MDSTAR{};
uint32_t MDTSEL{};
uint32_t MDEN{};
};
struct dc_PadState{
uint16_t Buttons{};
uint8_t LT{};
uint8_t RT{};
uint8_t JoyX{};
uint8_t JoyY{};
};
struct dc_WatchRange{
uint32_t Start{};
uint32_t Len{};
};
struct dc_Machine{
Slice<uint8_t> RAM{};
Slice<uint8_t> VRAM{};
Slice<uint8_t> AICARAM{};
Slice<uint8_t> Flash{};
sh4_CPU* CPU{};
dc_Disc* Disc{};
arm_CPU* ARM{};
bool ARMRunning{};
Map<uint32_t,uint32_t> AICARegs{};
dc_aicaTimers Timers{};
std::array<dc_AICASlot,64> Slots{};
int64_t armAcc{};
dc_Holly Holly{};
uint32_t C2DStat{};
uint32_t C2DLen{};
uint32_t RenderCountdown{};
uint32_t C2DCountdown{};
uint32_t C2DPendingBits{};
int32_t TAList{};
uint32_t TAVtx{};
uint32_t TANeed{};
Slice<uint8_t> TAFrame{};
Slice<uint8_t> TAClosed{};
Slice<uint8_t> TAFifo{};
uint32_t TAFifoCountdown{};
std::array<uint32_t,2048> PVRRegs{};
uint64_t TAWrites{};
dc_mapleState Maple{};
dc_PadState Pad{};
uint64_t Instrs{};
uint64_t Fields{};
uint32_t CurLine{};
uint32_t FieldNum{};
uint64_t instrInField{};
dc_biosHLE* bios{};
bool AudioCapture{};
Slice<int16_t> AudioPCM{};
std::function<void(uint32_t)> OnStep{};
std::function<void(uint64_t)> OnDisplay{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnGDRead{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnC2DTexture{};
LocalSource* Verbose{};
bool StopRequested{};
std::function<void(int64_t,int64_t)> OnPVRClear{};
std::function<void(Slice<uint8_t>)> OnPVRCmd{};
std::function<void(int64_t,int64_t,uint8_t,uint8_t,uint8_t,uint8_t,bool)> OnPVRPixel{};
std::function<void()> OnRender{};
int64_t RenderStopAfter{};
Slice<dc_WatchRange> WatchW{};
Slice<dc_WatchRange> WatchR{};
std::function<void(bool,uint32_t,uint32_t,int64_t,uint32_t)> OnWatch{};
Map<std::string,int64_t> gaps{};
uint32_t rrLastTotal{},rrLinePeriod{};
};
struct dc_pvrVert{
float x{};
float y{};
float z{};
float u{};
float v{};
uint32_t color{};
uint32_t offs{};
};
struct dc_renderState{
dc_Machine* m{};
uint32_t base{};
int64_t w{};
int64_t h{};
Slice<float> zbuf{};
bool sprite{};
int64_t strip{};
int64_t tris{};
int64_t px{};
std::array<dc_pvrVert,3> sv{};
bool texture{};
bool blend{};
bool gouraud{};
bool uv16{};
uint32_t colType{};
bool offsEn{};
uint32_t shade{};
uint32_t faceCol{};
uint32_t faceOffs{};
uint32_t depthCmp{};
bool zWrite{};
uint32_t texAddr{};
uint32_t texMipIx{};
uint32_t texFmt{};
bool texTwid{};
bool texVQ{};
uint32_t texBank{};
int64_t texUW{};
int64_t texVH{};
uint32_t baseCol{};
uint32_t baseOffs{};
};
struct dc_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
struct dc_RunConfig{
Map<uint32_t,bool> Breakpoints{};
bool NoSpin{};
};
struct Anon0{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};arm_Flow Flow{};uint32_t Target{};bool HasTarget{};bool Thumb{};bool TargetThumb{};int64_t Cond{};};
struct Anon1{std::array<uint32_t,16> R{};bool N{};bool Z{};bool C{};bool V{};bool Q{};uint32_t GE{};bool Thumb{};bool BigEndian{};bool IRQDisable{};bool FIQDisable{};uint32_t Mode{};arm_Variant Arch{};arm_vfpState VFP{};bool exclValid{};uint32_t exclAddr{};std::array<uint32_t,6> bankR13{};std::array<uint32_t,6> bankR14{};std::array<uint32_t,6> bankSPSR{};std::array<uint32_t,5> fiqR8_12{};std::array<uint32_t,5> usrR8_12{};std::function<bool(arm_CPU*,uint32_t)> SWI{};std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};dc_armBus* bus{};dc_armBus* wide{};bool Halted{};std::string HaltReason{};uint64_t Instrs{};uint32_t cur{};bool branched{};};
struct Anon10{uint8_t TOCR{};uint8_t TSTR{};std::array<sh4_TMUChannel,3> Ch{};};
struct Anon11{int64_t SectorSize{};int64_t DataOffset{};};
struct Anon12{LocalSource* ra{};int64_t size{};iso9660_Geometry geom{};Slice<uint8_t> buf{};};
struct Anon13{dc_Disc* src{};std::string System{};std::string Name{};int64_t Blocks{};int64_t rootLBA{};int64_t rootSize{};};
struct Anon14{std::string Name{};std::string Path{};bool IsDir{};int64_t Size{};int64_t Block{};};
struct Anon15{uint32_t Frac{};std::array<uint32_t,3> Sub{};uint32_t SCIPD{};uint32_t MCIPD{};uint64_t SampleTick{};};
struct Anon16{bool Active{};uint32_t Pos{};uint32_t Frac{};bool LP{};uint8_t EGState{};uint16_t EGLevel{};int32_t Cur{};int32_t Prev{};uint32_t DecPos{};int32_t AdHist{};int32_t AdStep{};int32_t LoopHist{};int32_t LoopStep{};bool LoopSeen{};};
struct Anon17{dc_Machine* m{};};
struct Anon18{uint32_t Cmd{};std::array<uint32_t,4> Params{};bool Done{};std::array<uint32_t,4> Result{};};
struct Anon19{uint32_t NextID{};Map<uint32_t,dc_gdRequest*> Requests{};};
struct Anon2{std::array<uint32_t,6> R13{};std::array<uint32_t,6> R14{};std::array<uint32_t,6> SPSR{};std::array<uint32_t,5> FIQR8_12{};std::array<uint32_t,5> USRR8_12{};};
struct Anon20{dc_Machine* m{};dc_biosState state{};Map<std::string,int64_t> calls{};};
struct Anon21{int64_t Number{};std::string Mode{};int64_t FileOffset{};int64_t Length{};int64_t StartLBA{};};
struct Anon22{uint32_t ISTNRM{};uint32_t ISTEXT{};uint32_t ISTERR{};uint32_t IML2NRM{};uint32_t IML2EXT{};uint32_t IML2ERR{};uint32_t IML4NRM{};uint32_t IML4EXT{};uint32_t IML4ERR{};uint32_t IML6NRM{};uint32_t IML6EXT{};uint32_t IML6ERR{};};
struct Anon23{std::string HardwareID{};std::string MakerID{};std::string DeviceInfo{};std::string AreaSyms{};std::string Peripherals{};std::string ProductNo{};std::string Version{};std::string ReleaseDate{};std::string BootFile{};std::string Company{};std::string Title{};};
struct Anon24{Slice<uint8_t> RAM{};Slice<uint8_t> VRAM{};Slice<uint8_t> AICARAM{};Slice<uint8_t> Flash{};sh4_CPU* CPU{};dc_Disc* Disc{};arm_CPU* ARM{};bool ARMRunning{};Map<uint32_t,uint32_t> AICARegs{};dc_aicaTimers Timers{};std::array<dc_AICASlot,64> Slots{};int64_t armAcc{};dc_Holly Holly{};uint32_t C2DStat{};uint32_t C2DLen{};uint32_t RenderCountdown{};uint32_t C2DCountdown{};uint32_t C2DPendingBits{};int32_t TAList{};uint32_t TAVtx{};uint32_t TANeed{};Slice<uint8_t> TAFrame{};Slice<uint8_t> TAClosed{};Slice<uint8_t> TAFifo{};uint32_t TAFifoCountdown{};std::array<uint32_t,2048> PVRRegs{};uint64_t TAWrites{};dc_mapleState Maple{};dc_PadState Pad{};uint64_t Instrs{};uint64_t Fields{};uint32_t CurLine{};uint32_t FieldNum{};uint64_t instrInField{};dc_biosHLE* bios{};bool AudioCapture{};Slice<int16_t> AudioPCM{};std::function<void(uint32_t)> OnStep{};std::function<void(uint64_t)> OnDisplay{};std::function<void(uint32_t,uint32_t,uint32_t)> OnGDRead{};std::function<void(uint32_t,uint32_t,uint32_t)> OnC2DTexture{};LocalSource* Verbose{};bool StopRequested{};std::function<void(int64_t,int64_t)> OnPVRClear{};std::function<void(Slice<uint8_t>)> OnPVRCmd{};std::function<void(int64_t,int64_t,uint8_t,uint8_t,uint8_t,uint8_t,bool)> OnPVRPixel{};std::function<void()> OnRender{};int64_t RenderStopAfter{};Slice<dc_WatchRange> WatchW{};Slice<dc_WatchRange> WatchR{};std::function<void(bool,uint32_t,uint32_t,int64_t,uint32_t)> OnWatch{};Map<std::string,int64_t> gaps{};};
struct Anon25{uint32_t Start{};uint32_t Len{};};
struct Anon26{uint16_t Buttons{};uint8_t LT{};uint8_t RT{};uint8_t JoyX{};uint8_t JoyY{};};
struct Anon27{uint32_t MDSTAR{};uint32_t MDTSEL{};uint32_t MDEN{};};
struct Anon28{dc_Machine* m{};uint32_t base{};int64_t w{};int64_t h{};Slice<float> zbuf{};bool sprite{};int64_t strip{};int64_t tris{};int64_t px{};std::array<dc_pvrVert,3> sv{};bool texture{};bool blend{};bool gouraud{};bool uv16{};uint32_t colType{};bool offsEn{};uint32_t shade{};uint32_t faceCol{};uint32_t faceOffs{};uint32_t depthCmp{};bool zWrite{};uint32_t texAddr{};uint32_t texMipIx{};uint32_t texFmt{};bool texTwid{};bool texVQ{};uint32_t texBank{};int64_t texUW{};int64_t texVH{};uint32_t baseCol{};uint32_t baseOffs{};};
struct Anon29{float x{};float y{};float z{};float u{};float v{};uint32_t color{};uint32_t offs{};};
struct Anon3{uint32_t bit{};std::string name{};};
struct Anon30{uint64_t Steps{};uint32_t PC{};std::string Reason{};};
struct Anon31{Map<uint32_t,bool> Breakpoints{};bool NoSpin{};};
struct Anon4{std::array<uint32_t,32> S{};uint32_t FPSCR{};uint32_t FPEXC{};};
struct Anon5{std::array<uint32_t,16> R{};std::array<uint32_t,8> Rbank{};uint32_t SR{};uint32_t GBR{};uint32_t VBR{};uint32_t SSR{};uint32_t SPC{};uint32_t SGR{};uint32_t DBR{};uint32_t MACH{};uint32_t MACL{};uint32_t PR{};uint32_t PC{};std::array<std::array<uint32_t,16>,2> fpr{};uint32_t FPSCR{};uint32_t FPUL{};uint32_t MMUCR{};uint32_t CCR{};uint32_t TRA{};uint32_t EXPEVT{};uint32_t INTEVT{};uint32_t PTEH{};uint32_t PTEL{};uint32_t TTB{};uint32_t TEA{};uint32_t QACR0{};uint32_t QACR1{};uint32_t ICR{};uint32_t IPRA{};uint32_t IPRB{};uint32_t IPRC{};sh4_TMUState TMU{};std::array<std::array<uint32_t,8>,2> SQ{};uint32_t irlLevel{};uint32_t irlCode{};Slice<uint8_t> SerialTX{};bool Halted{};std::string HaltReason{};uint64_t Steps{};dc_Machine* bus{};dc_Machine* fetcher{};uint32_t nextPC{};uint32_t curPC{};bool delaySlot{};bool pendingDelay{};Map<uint32_t,uint32_t> onchipRaw{};Map<uint32_t,int64_t> onchipGaps{};};
struct Anon6{uint32_t addr{};int64_t reads{};};
struct Anon7{uint32_t Addr{};uint16_t Word{};int64_t Len{};std::string Mnem{};std::string Text{};sh4_Flow Flow{};uint32_t Target{};bool HasTarget{};bool HasDelay{};uint32_t LitAddr{};int64_t LitSize{};};
struct Anon8{std::array<uint32_t,16> R{};std::array<uint32_t,8> Rbank{};uint32_t SR{};uint32_t GBR{};uint32_t VBR{};uint32_t SSR{};uint32_t SPC{};uint32_t SGR{};uint32_t DBR{};uint32_t MACH{};uint32_t MACL{};uint32_t PR{};uint32_t PC{};uint32_t NextPC{};std::array<std::array<uint32_t,16>,2> FPR{};uint32_t FPSCR{};uint32_t FPUL{};uint32_t MMUCR{};uint32_t CCR{};uint32_t TRA{};uint32_t EXPEVT{};uint32_t INTEVT{};uint32_t PTEH{};uint32_t PTEL{};uint32_t TTB{};uint32_t TEA{};uint32_t QACR0{};uint32_t QACR1{};uint32_t ICR{};uint32_t IPRA{};uint32_t IPRB{};uint32_t IPRC{};sh4_TMUState TMU{};std::array<std::array<uint32_t,8>,2> SQ{};uint32_t IRLLevel{};uint32_t IRLCode{};Slice<uint8_t> SerialTX{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint32_t CurPC{};bool DelaySlot{};bool PendingDelay{};Map<uint32_t,uint32_t> OnchipRaw{};Map<uint32_t,int64_t> OnchipGaps{};};
struct Anon9{uint32_t TCOR{};uint32_t TCNT{};uint16_t TCR{};uint32_t Frac{};};

#include "adapters-decl.h"
uint32_t arm_signExtend(uint32_t v,uint64_t n);
std::string arm_imm(uint32_t v);
arm_Inst arm_Decode(Slice<uint8_t> code,uint32_t addr,bool thumb);
std::tuple<uint32_t,bool> arm_word(Slice<uint8_t> code);
arm_Inst arm_DecodeARM(Slice<uint8_t> code,uint32_t addr);
std::string arm_cn(int64_t cond);
arm_Inst arm_undef(uint32_t w,arm_Inst in);
arm_Inst arm_decodeDataMisc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm);
arm_Inst arm_decodeDataProc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm);
void arm_classifyPCWrite(arm_Inst* in,uint32_t w,uint32_t op);
std::string arm_shiftOperand(uint32_t w);
uint32_t arm_ror32(uint32_t v,uint32_t n);
arm_Inst arm_decodeMisc(uint32_t w,arm_Inst in,bool immForm);
arm_Inst arm_decodeMSRreg(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMSRimm(uint32_t w,arm_Inst in);
std::string arm_msrFields(uint32_t w);
arm_Inst arm_decodeSignedMul(uint32_t w,arm_Inst in);
arm_Inst arm_decodeExtension(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMul(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMulLong(uint32_t w,arm_Inst in);
arm_Inst arm_decodeHalf(uint32_t w,arm_Inst in,uint32_t sh);
std::string arm_plusMinus(std::string s);
arm_Inst arm_decodeSingle(uint32_t w,arm_Inst in,bool regOff);
std::string arm_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb);
arm_Inst arm_decodeBlock(uint32_t w,arm_Inst in);
std::string arm_regList(uint32_t mask);
std::string arm_joinComma(Slice<std::string> parts);
arm_Inst arm_decodeBranch(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeUncond(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeCoproLS(uint32_t w,arm_Inst in);
arm_Inst arm_decodeCopro(uint32_t w,arm_Inst in);
int64_t arm_CPU_Step(arm_CPU* c);
uint32_t arm_CPU_boolToU(arm_CPU* c,bool b);
int64_t arm_CPU_stepARM(arm_CPU* c);
int64_t arm_CPU_execUncondARM(arm_CPU* c,uint32_t w);
void arm_CPU_execARM(arm_CPU* c,uint32_t w);
void arm_CPU_execDataMisc(arm_CPU* c,uint32_t w,bool immForm);
std::tuple<uint32_t,uint32_t> arm_CPU_dpOperand(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_execDataProc(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_branchTo(arm_CPU* c,uint32_t v);
void arm_CPU_bxTo(arm_CPU* c,uint32_t v);
void arm_CPU_execBranch(arm_CPU* c,uint32_t w);
void arm_CPU_execExtension(arm_CPU* c,uint32_t w);
void arm_CPU_execMul(arm_CPU* c,uint32_t w);
void arm_CPU_execMulLong(arm_CPU* c,uint32_t w);
void arm_CPU_execSwap(arm_CPU* c,uint32_t w);
void arm_CPU_execHalf(arm_CPU* c,uint32_t w,uint32_t sh);
void arm_CPU_execSingle(arm_CPU* c,uint32_t w,bool regOff);
void arm_CPU_execBlock(arm_CPU* c,uint32_t w);
bool arm_CPU_inUserBank(arm_CPU* c);
uint32_t arm_CPU_userReg(arm_CPU* c,uint32_t i);
void arm_CPU_setUserReg(arm_CPU* c,uint32_t i,uint32_t v);
void arm_CPU_Reset(arm_CPU* c);
uint32_t arm_CPU_PC(arm_CPU* c);
uint32_t arm_CPU_CPSR(arm_CPU* c);
void arm_CPU_SetCPSR(arm_CPU* c,uint32_t v);
int64_t arm_modeIndex(uint32_t mode);
void arm_CPU_switchMode(arm_CPU* c,uint32_t mode);
uint32_t arm_CPU_SPSR(arm_CPU* c);
void arm_CPU_SetSPSR(arm_CPU* c,uint32_t v);
uint8_t arm_CPU_read8(arm_CPU* c,uint32_t a);
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v);
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a);
void arm_CPU_write32(arm_CPU* c,uint32_t a,uint32_t v);
uint32_t arm_CPU_reg(arm_CPU* c,uint32_t i);
void arm_CPU_setReg(arm_CPU* c,uint32_t i,uint32_t v);
void arm_CPU_setNZ(arm_CPU* c,uint32_t v);
uint32_t arm_CPU_add(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
uint32_t arm_CPU_sub_(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
bool arm_CPU_cond(arm_CPU* c,int64_t cc);
std::tuple<uint32_t,uint32_t> arm_CPU_shift(arm_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin);
void arm_CPU_ClearExclusive(arm_CPU* c);
arm_Banks arm_CPU_SaveBanks(arm_CPU* c);
void arm_CPU_RestoreBanks(arm_CPU* c,arm_Banks b);
std::tuple<arm_Inst,bool> arm_decodeARMv6(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeSync(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMedia(uint32_t w,arm_Inst in);
std::string arm_parallelName(uint32_t class_,uint32_t op2);
arm_Inst arm_decodeMediaPack(uint32_t w,arm_Inst in);
arm_Inst arm_satText(uint32_t w,arm_Inst in,std::string name,uint32_t sat);
std::string arm_extendName(uint32_t op1);
arm_Inst arm_decodeMediaMul(uint32_t w,arm_Inst in);
std::tuple<arm_Inst,bool> arm_decodeV6Uncond(uint32_t w,arm_Inst in);
bool arm_CPU_execARMv6(arm_CPU* c,uint32_t w);
bool arm_CPU_execUncondARMv6(arm_CPU* c,uint32_t w);
void arm_CPU_execSync(arm_CPU* c,uint32_t w);
void arm_CPU_execUMAAL(arm_CPU* c,uint32_t w);
bool arm_CPU_execMedia(arm_CPU* c,uint32_t w);
bool arm_CPU_execParallel(arm_CPU* c,uint32_t w);
std::tuple<uint32_t,bool> arm_parLane(uint32_t av,uint32_t bv,uint64_t width,bool signed_,arm_parMode mode,bool sub_);
int32_t arm_int16v(uint32_t v);
int32_t arm_signExtendLane(uint32_t v,uint64_t width);
bool arm_CPU_execMediaPack(arm_CPU* c,uint32_t w);
uint32_t arm_CPU_doSat(arm_CPU* c,uint32_t w,bool signed_);
bool arm_CPU_execExtend(arm_CPU* c,uint32_t w);
bool arm_CPU_execMediaMul(arm_CPU* c,uint32_t w);
void arm_CPU_execUSAD(arm_CPU* c,uint32_t w);
void arm_CPU_execMisc(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_undefV4T(arm_CPU* c,uint32_t w,std::string name);
void arm_CPU_execMSR(arm_CPU* c,uint32_t w,bool immForm);
uint32_t arm_CPU_qsat(arm_CPU* c,int64_t v);
void arm_CPU_execSaturating(arm_CPU* c,uint32_t w);
int64_t arm_half16(uint32_t v,uint32_t t);
void arm_CPU_execSignedMul(arm_CPU* c,uint32_t w);
void arm_CPU_execSWI(arm_CPU* c,uint32_t w);
void arm_CPU_exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
void arm_CPU_Exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
bool arm_CPU_IRQ(arm_CPU* c);
void arm_CPU_execCopro(arm_CPU* c,uint32_t w);
arm_Inst arm_DecodeThumb(Slice<uint8_t> code,uint32_t addr);
std::string arm_lo(uint32_t v);
arm_Inst arm_thumbAddSub(uint32_t h,arm_Inst in);
arm_Inst arm_thumbALU(uint32_t h,arm_Inst in);
arm_Inst arm_thumbHiReg(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreReg(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreSExt(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreImm(uint32_t h,arm_Inst in);
arm_Inst arm_thumbPushPop(uint32_t h,arm_Inst in);
arm_Inst arm_thumbCondBranch(uint32_t h,uint32_t addr,arm_Inst in);
arm_Inst arm_thumbLongBranch(Slice<uint8_t> code,uint32_t h,uint32_t addr,arm_Inst in);
int64_t arm_CPU_stepThumb(arm_CPU* c);
void arm_CPU_execThumb(arm_CPU* c,uint32_t h);
void arm_CPU_thumbAddSub(arm_CPU* c,uint32_t h);
void arm_CPU_thumbImm(arm_CPU* c,uint32_t h);
void arm_CPU_thumbALUExec(arm_CPU* c,uint32_t h);
void arm_CPU_thumbHiRegExec(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreReg(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreSExt(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreImm(arm_CPU* c,uint32_t h);
void arm_CPU_thumbPushPop(arm_CPU* c,uint32_t h);
void arm_CPU_thumbBlock(arm_CPU* c,uint32_t h);
void arm_CPU_thumbCondBranch(arm_CPU* c,uint32_t h);
std::string arm_Variant_String(arm_Variant v);
bool arm_Variant_isV6(arm_Variant v);
bool arm_Variant_v5OrLater(arm_Variant v);
arm_Inst arm_DecodeVariant(Slice<uint8_t> code,uint32_t addr,bool thumb,arm_Variant v);
arm_Inst arm_DecodeARMVariant(Slice<uint8_t> code,uint32_t addr,arm_Variant v);
arm_Inst arm_v4tFilter(arm_Inst in);
float arm_CPU_sGet(arm_CPU* c,uint32_t n);
void arm_CPU_sSet(arm_CPU* c,uint32_t n,float v);
uint32_t arm_CPU_sBits(arm_CPU* c,uint32_t n);
void arm_CPU_sSetBits(arm_CPU* c,uint32_t n,uint32_t v);
double arm_CPU_dGet(arm_CPU* c,uint32_t n);
void arm_CPU_dSet(arm_CPU* c,uint32_t n,double v);
uint32_t arm_sReg(uint32_t v,uint32_t bit);
uint32_t arm_dReg(uint32_t v,uint32_t bit);
bool arm_isVFP(uint32_t w);
arm_Inst arm_decodeVFP(uint32_t w,arm_Inst in);
arm_Inst arm_decodeVFPMove(uint32_t w,arm_Inst in);
arm_Inst arm_decodeVFPData(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx);
arm_Inst arm_decodeVFPExt(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx,std::string vd,std::string vm);
std::string arm_vfpRegName(bool single,uint32_t v,uint32_t bit);
std::string arm_wbMark(uint32_t wb);
bool arm_CPU_execVFP(arm_CPU* c,uint32_t w);
void arm_CPU_execVFPLoadStore(arm_CPU* c,uint32_t w);
bool arm_CPU_execVFPData(arm_CPU* c,uint32_t w);
void arm_CPU_vfpNegate(arm_CPU* c,uint32_t vd,bool single);
void arm_CPU_vfpMulAcc(arm_CPU* c,uint32_t vd,uint32_t vn,uint32_t vm,bool single,bool sub_,bool negAcc);
bool arm_CPU_execVFPExt(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm);
void arm_CPU_vfpCopy(arm_CPU* c,uint32_t vd,uint32_t vm,bool single);
void arm_CPU_vfpUnary(arm_CPU* c,uint32_t vd,uint32_t vm,bool single,std::function<double(double)> f);
void arm_CPU_execVCMP(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm,bool withZero);
uint32_t arm_vdOf(uint32_t w,bool single);
uint32_t arm_vmOf(uint32_t w,bool single);
void arm_CPU_execVCVTPrec(arm_CPU* c,uint32_t w,bool srcSingle);
void arm_CPU_execVCVTFromInt(arm_CPU* c,uint32_t w,bool single);
void arm_CPU_execVCVTToInt(arm_CPU* c,uint32_t w,bool single);
bool arm_CPU_execVFPMove(arm_CPU* c,uint32_t w);
bool arm_CPU_execVFPMove64(arm_CPU* c,uint32_t w);
uint64_t arm_vfpExpandImm(uint32_t w,bool single);
uint32_t arm_repeat(uint32_t bit,uint32_t n);
uint64_t arm_repeat64(uint64_t bit,uint32_t n);
void sh4_CPU_Reset(sh4_CPU* c);
uint32_t sh4_CPU_CurPC(sh4_CPU* c);
void sh4_CPU_SetPC(sh4_CPU* c,uint32_t pc);
uint32_t sh4_CPU_Reg(sh4_CPU* c,uint32_t i);
void sh4_CPU_SetReg(sh4_CPU* c,uint32_t i,uint32_t v);
bool sh4_CPU_InDelaySlot(sh4_CPU* c);
bool sh4_CPU_NextIsDelaySlot(sh4_CPU* c);
uint32_t sh4_bankSelect(uint32_t sr);
void sh4_CPU_SetSR(sh4_CPU* c,uint32_t v);
void sh4_CPU_SetFPSCR(sh4_CPU* c,uint32_t v);
uint32_t sh4_CPU_T(sh4_CPU* c);
void sh4_CPU_setT(sh4_CPU* c,bool cond);
void sh4_CPU_SetIRL(sh4_CPU* c,uint32_t level,uint32_t code);
bool sh4_CPU_cacheArray(sh4_CPU* c,uint32_t addr,bool write);
uint8_t sh4_CPU_read8(sh4_CPU* c,uint32_t addr);
uint16_t sh4_CPU_read16(sh4_CPU* c,uint32_t addr);
uint32_t sh4_CPU_read32(sh4_CPU* c,uint32_t addr);
void sh4_CPU_write8(sh4_CPU* c,uint32_t addr,uint8_t v);
void sh4_CPU_write16(sh4_CPU* c,uint32_t addr,uint16_t v);
void sh4_CPU_write32(sh4_CPU* c,uint32_t addr,uint32_t v);
sh4_Inst sh4_Decode(Slice<uint8_t> code,uint32_t addr);
sh4_Inst sh4_DecodeHalfword(uint16_t h,uint32_t addr);
std::string sh4_r(uint32_t n);
std::string sh4_bank(uint32_t m);
uint32_t sh4_rn(uint16_t h);
uint32_t sh4_rm(uint16_t h);
int32_t sh4_s8(uint16_t h);
int32_t sh4_s12(uint16_t h);
void sh4_Inst_branch(sh4_Inst* in,sh4_Flow flow,int32_t disp,bool delayed);
void sh4_decode(sh4_Inst* in,uint16_t h,uint32_t addr);
void sh4_decode0(sh4_Inst* in,uint16_t h);
void sh4_decode2(sh4_Inst* in,uint16_t h);
void sh4_decode3(sh4_Inst* in,uint16_t h);
void sh4_decode4(sh4_Inst* in,uint16_t h);
void sh4_decode6(sh4_Inst* in,uint16_t h);
void sh4_decode8(sh4_Inst* in,uint16_t h);
void sh4_decodeC(sh4_Inst* in,uint16_t h,uint32_t addr);
std::string sh4_fr(uint32_t n);
std::string sh4_dr(uint32_t n);
std::string sh4_fv(uint32_t n);
void sh4_decodeFPU(sh4_Inst* in,uint16_t h);
void sh4_decodeFPUxD(sh4_Inst* in,uint16_t h,uint32_t n);
int64_t sh4_CPU_Step(sh4_CPU* c);
bool sh4_CPU_checkInterrupt(sh4_CPU* c);
void sh4_CPU_exception(sh4_CPU* c,uint32_t expevt,uint32_t spc);
void sh4_CPU_doBranch(sh4_CPU* c,bool taken,uint32_t target);
void sh4_CPU_doJumpNow(sh4_CPU* c,bool taken,uint32_t target);
void sh4_CPU_execute(sh4_CPU* c,uint16_t h);
void sh4_CPU_exec0(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_exec2(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_exec3(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_exec4(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_exec6(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_exec8(sh4_CPU* c,uint16_t h);
void sh4_CPU_execC(sh4_CPU* c,uint16_t h);
uint32_t sh4_CPU_bankReg(sh4_CPU* c,uint32_t m);
void sh4_CPU_setBankReg(sh4_CPU* c,uint32_t m,uint32_t v);
void sh4_CPU_div1(sh4_CPU* c,uint32_t n,uint32_t m);
void sh4_CPU_macL(sh4_CPU* c,uint32_t n,uint32_t m);
void sh4_CPU_macW(sh4_CPU* c,uint32_t n,uint32_t m);
void sh4_CPU_unknown(sh4_CPU* c,uint16_t h);
void sh4_CPU_execFPU(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m);
void sh4_CPU_execFPUxD(sh4_CPU* c,uint16_t h,uint32_t n,std::array<uint32_t,16>* fr,bool pr);
void sh4_CPU_fpuArith(sh4_CPU* c,std::array<uint32_t,16>* fr,bool pr,uint32_t n,uint32_t m,std::function<double(double,double)> op);
Slice<uint32_t> sh4_CPU_pair(sh4_CPU* c,uint32_t i);
void sh4_CPU_fmovLoad(sh4_CPU* c,uint32_t n,uint32_t m,uint32_t addr);
void sh4_CPU_fmovStore(sh4_CPU* c,uint32_t n,uint32_t m,uint32_t addr);
void sh4_CPU_ftrv(sh4_CPU* c,std::array<uint32_t,16>* fr,uint32_t v);
float sh4_getFR(std::array<uint32_t,16>* fr,uint32_t i);
double sh4_getDR(std::array<uint32_t,16>* fr,uint32_t i);
void sh4_setDR(std::array<uint32_t,16>* fr,uint32_t i,double v);
uint32_t sh4_CPU_onchipRead(sh4_CPU* c,uint32_t addr,int64_t size);
void sh4_CPU_onchipWrite(sh4_CPU* c,uint32_t addr,int64_t size,uint32_t v);
uint32_t sh4_CPU_OnchipReg(sh4_CPU* c,uint32_t addr);
Slice<std::string> sh4_CPU_Gaps(sh4_CPU* c);
std::string sh4_Flow_String(sh4_Flow f);
std::string sh4_Inst_String(sh4_Inst in);
uint32_t sh4_litW(uint32_t addr,uint32_t disp);
uint32_t sh4_litL(uint32_t addr,uint32_t disp);
void sh4_CPU_sqWrite(sh4_CPU* c,uint32_t addr,uint32_t size,uint32_t v);
uint32_t sh4_CPU_sqRead32(sh4_CPU* c,uint32_t addr);
void sh4_CPU_sqFlush(sh4_CPU* c,uint32_t addr);
sh4_State sh4_CPU_Snapshot(sh4_CPU* c);
void sh4_CPU_Restore(sh4_CPU* c,sh4_State s);
uint32_t sh4_tmuPeriod(uint16_t tcr);
void sh4_CPU_tickTMU(sh4_CPU* c);
std::tuple<uint32_t,uint32_t,bool> sh4_CPU_tmuPending(sh4_CPU* c);
std::tuple<uint32_t,bool> sh4_CPU_tmuRead(sh4_CPU* c,uint32_t addr);
bool sh4_CPU_tmuWrite(sh4_CPU* c,uint32_t addr,uint32_t v);
std::string iso9660_Geometry_String(iso9660_Geometry g);
iso9660_Geometry iso9660_Source_Geometry(iso9660_Source* s);
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadBlock(iso9660_Volume* v,int64_t n);
std::string iso9660_Entry_String(iso9660_Entry e);
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolume(dc_Disc* src);
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolumeAt(dc_Disc* src,int64_t pvd);
uint32_t iso9660_le32(Slice<uint8_t> b);
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_dirEntries(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath);
bool iso9660_nameEqual(std::string entry,std::string want);
Slice<std::string> iso9660_splitPath(std::string p);
std::tuple<iso9660_Entry,Error> iso9660_Volume_Resolve(iso9660_Volume* v,std::string path);
std::string iso9660_normalisePath(std::string p);
std::tuple<int64_t,int64_t,bool> iso9660_parseLbnPath(std::string path);
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_ReadDir(iso9660_Volume* v,std::string path);
Error iso9660_Volume_Walk(iso9660_Volume* v,std::function<Error(iso9660_Entry)> fn);
Error iso9660_Volume_walk(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(iso9660_Entry)> fn);
std::tuple<int64_t,Error> iso9660_Volume_ReadFileAt(iso9660_Volume* v,iso9660_Entry e,int64_t off,Slice<uint8_t> p);
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadFile(iso9660_Volume* v,std::string path);
std::tuple<iso9660_Entry,bool> iso9660_Volume_FileAt(iso9660_Volume* v,int64_t lba);
iso9660_byteReaderAt iso9660_newByteReaderAt(Slice<uint8_t> b);
std::tuple<int64_t,Error> iso9660_byteReaderAt_ReadAt(iso9660_byteReaderAt b,Slice<uint8_t> p,int64_t off);
uint32_t dc_timerReg(int64_t i);
uint32_t dc_Machine_slotW(dc_Machine* m,uint32_t slot,uint32_t off);
void dc_Machine_tickAICA(dc_Machine* m);
void dc_Machine_aicaMainIRQ(dc_Machine* m);
std::tuple<Slice<uint8_t>,uint32_t,bool> dc_armBus_region(dc_armBus b,uint32_t addr);
uint8_t dc_armBus_Read(dc_armBus b,uint32_t addr);
void dc_armBus_Write(dc_armBus b,uint32_t addr,uint8_t v);
uint16_t dc_armBus_Read16(dc_armBus b,uint32_t addr);
uint32_t dc_armBus_Read32(dc_armBus b,uint32_t addr);
uint32_t dc_armBus_Read32_reference(dc_armBus b,uint32_t addr);
void dc_armBus_Write16(dc_armBus b,uint32_t addr,uint16_t v);
void dc_armBus_Write32(dc_armBus b,uint32_t addr,uint32_t v);
uint32_t dc_Machine_aicaRead(dc_Machine* m,uint32_t off,int64_t size,bool sh4);
void dc_Machine_aicaWrite(dc_Machine* m,uint32_t off,int64_t size,uint32_t v,bool sh4);
std::string dc_side(bool sh4);
void dc_Machine_stepARM(dc_Machine* m);
uint32_t dc_Machine_rtcRead(dc_Machine* m,uint32_t addr);
void dc_init_aica_synth();
void dc_AICASlot_keyOn(dc_AICASlot* s);
void dc_AICASlot_keyOff(dc_AICASlot* s);
uint32_t dc_effRate(uint32_t r,uint32_t krs);
uint64_t dc_egPeriod(uint32_t er);
void dc_Machine_stepEG(dc_Machine* m,dc_AICASlot* s,uint32_t slot,uint64_t tick);
uint32_t dc_Machine_sampleAddr(dc_Machine* m,uint32_t slot);
int32_t dc_Machine_fetchPCM(dc_Machine* m,uint32_t sa,uint32_t n,bool eight);
void dc_Machine_decodeADPCMTo(dc_Machine* m,dc_AICASlot* s,uint32_t sa,uint32_t target,uint32_t lsa);
void dc_Machine_mixSample(dc_Machine* m);
int16_t dc_clamp16(int64_t v);
dc_biosHLE* dc_newBIOS(dc_Machine* m);
bool dc_Machine_trapPC(dc_Machine* m,uint32_t pc);
void dc_biosHLE_count(dc_biosHLE* b,std::string name);
Slice<std::string> dc_biosHLE_census(dc_biosHLE* b);
void dc_biosHLE_sysinfo(dc_biosHLE* b);
void dc_biosHLE_flashrom(dc_biosHLE* b);
void dc_biosHLE_gdrom(dc_biosHLE* b);
void dc_biosHLE_execGD(dc_biosHLE* b,dc_gdRequest* req);
Error dc_Machine_Boot(dc_Machine* m);
void dc_Machine_putRAM32(dc_Machine* m,int64_t off,uint32_t v);
uint32_t dc_Machine_ram32(dc_Machine* m,uint32_t addr);
bool dc_Track_IsData(dc_Track t);
std::tuple<std::string,Slice<dc_Track>,Error> dc_parseCue(std::string text);
std::tuple<std::string,std::string,bool> dc_quoted(std::string line);
uint32_t dc_Machine_hollyRead(dc_Machine* m,uint32_t addr);
void dc_Machine_hollyWrite(dc_Machine* m,uint32_t addr,uint32_t v);
void dc_Machine_taSubmitDMA(dc_Machine* m,uint32_t src,uint32_t byteLen);
void dc_Machine_taFifoWrite(dc_Machine* m,uint32_t v);
void dc_Machine_taFeed(dc_Machine* m,Slice<uint8_t> p);
void dc_Machine_taOpenList(dc_Machine* m,int32_t n);
void dc_Machine_tickCompletions(dc_Machine* m);
void dc_Machine_raiseNRM(dc_Machine* m,uint32_t bits);
void dc_Machine_updateIRL(dc_Machine* m);
uint32_t dc_Machine_spgStatus(dc_Machine* m);
std::tuple<dc_IPBin,Error> dc_parseIPBin(Slice<uint8_t> sector);
dc_Machine* dc_NewMachine(dc_Disc* disc);
Slice<std::string> dc_Machine_Census(dc_Machine* m);
std::tuple<Slice<uint8_t>,uint32_t> dc_Machine_backing(dc_Machine* m,uint32_t addr);
uint32_t dc_vram32to64(uint32_t off);
uint16_t dc_Machine_Fetch16(dc_Machine* m,uint32_t addr);
uint16_t dc_Machine_Fetch16_reference(dc_Machine* m,uint32_t addr);
bool dc_inRanges(Slice<dc_WatchRange> rs,uint32_t addr,int64_t size);
void dc_Machine_watch(dc_Machine* m,bool write,uint32_t addr,uint32_t v,int64_t size);
uint8_t dc_Machine_Read8(dc_Machine* m,uint32_t addr);
uint8_t dc_Machine_read8i(dc_Machine* m,uint32_t addr);
uint16_t dc_Machine_Read16(dc_Machine* m,uint32_t addr);
uint16_t dc_Machine_read16i(dc_Machine* m,uint32_t addr);
uint16_t dc_Machine_read16i_reference(dc_Machine* m,uint32_t addr);
uint32_t dc_Machine_Read32(dc_Machine* m,uint32_t addr);
uint32_t dc_Machine_read32i(dc_Machine* m,uint32_t addr);
uint32_t dc_Machine_read32i_reference(dc_Machine* m,uint32_t addr);
void dc_Machine_Write8(dc_Machine* m,uint32_t addr,uint8_t v);
void dc_Machine_Write16(dc_Machine* m,uint32_t addr,uint16_t v);
void dc_Machine_Write32(dc_Machine* m,uint32_t addr,uint32_t v);
uint32_t dc_Machine_ioRead(dc_Machine* m,uint32_t addr,int64_t size);
void dc_Machine_ioWrite(dc_Machine* m,uint32_t addr,int64_t size,uint32_t v);
uint32_t dc_Machine_pvrRead(dc_Machine* m,uint32_t addr);
void dc_Machine_pvrWrite(dc_Machine* m,uint32_t addr,uint32_t v);
std::tuple<uint16_t,bool> dc_PadButton(std::string name);
uint32_t dc_bswap(uint32_t v);
uint32_t dc_Machine_mapleRead(dc_Machine* m,uint32_t addr);
void dc_Machine_mapleWrite(dc_Machine* m,uint32_t addr,uint32_t v);
void dc_Machine_mapleDMA(dc_Machine* m);
void dc_Machine_mapleRespond(dc_Machine* m,uint32_t cmd,uint32_t dst,uint32_t payload,uint32_t recv);
void dc_Machine_renderFrame(dc_Machine* m);
void dc_Machine_clearFB(dc_Machine* m,uint32_t base,int64_t w,int64_t h);
void dc_renderState_skip(dc_renderState* st,std::string what,uint32_t v);
void dc_renderState_loadHeader(dc_renderState* st,uint32_t pcw,Slice<uint8_t> p);
uint32_t dc_mipTopOffset(int64_t side,uint32_t fmt);
uint32_t dc_vqMipIndexOffset(int64_t side);
uint32_t dc_packFloatCol(Slice<uint8_t> p);
float dc_f32(Slice<uint8_t> p);
void dc_renderState_drawSprite(dc_renderState* st,Slice<uint8_t> p);
void dc_renderState_pushStripVertex(dc_renderState* st,Slice<uint8_t> p,uint32_t pcw);
uint32_t dc_scaleCol(uint32_t col,float i);
uint32_t dc_lerp3(float w0,float w1,float w2,uint32_t a,uint32_t b,uint32_t c,uint64_t shift);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> dc_renderState_shadePixel(dc_renderState* st,float w0,float w1,float w2,dc_pvrVert a,dc_pvrVert b,dc_pvrVert c,uint32_t tr,uint32_t tg,uint32_t tb,uint32_t ta);
uint32_t dc_clamp255(uint32_t v);
bool dc_depthPass(uint32_t mode,float z,float old);
void dc_renderState_tri(dc_renderState* st,dc_pvrVert a,dc_pvrVert b,dc_pvrVert c);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> dc_renderState_sample(dc_renderState* st,float u,float v);
bool dc_renderState_plot(dc_renderState* st,int64_t x,int64_t y,uint32_t r,uint32_t g,uint32_t b,uint32_t a);
bool dc_renderState_plotStore(dc_renderState* st,int64_t x,int64_t y,uint32_t r,uint32_t g,uint32_t b,uint32_t a);
uint32_t dc_twiddle(uint32_t x,uint32_t y);
uint32_t dc_twiddle_reference(uint32_t x,uint32_t y);
std::tuple<float,float> dc_unpackUV16(uint32_t w);
float dc_float32frombits(uint32_t b);
float dc_min3(float a,float b,float c);
float dc_max3(float a,float b,float c);
int64_t dc_imin(int64_t a,int64_t b);
int64_t dc_imax(int64_t a,int64_t b);
std::string dc_Result_String(dc_Result r);
dc_Result dc_Machine_Run(dc_Machine* m,uint64_t maxSteps,dc_RunConfig cfg);
dc_Result dc_Machine_RunFields(dc_Machine* m,uint64_t n,uint64_t budget);
uint64_t dc_minU64(uint64_t a,uint64_t b);
void dc_Machine_tickField(dc_Machine* m);
void dc_Machine_tickField_reference(dc_Machine* m);
uint32_t dc_Machine_spgTotalLines(dc_Machine* m);
std::tuple<image_RGBA*,Error> dc_Machine_RenderFB(dc_Machine* m);
std::tuple<image_RGBA*,Error> dc_Machine_RenderDrawTarget(dc_Machine* m);
std::tuple<image_RGBA*,Error> dc_Machine_RenderVRAM(dc_Machine* m,uint32_t offset,int64_t w,int64_t h);
constexpr arm_Flow arm_FlowSeq=0ULL;
constexpr arm_Flow arm_FlowBranch=1ULL;
constexpr arm_Flow arm_FlowJump=2ULL;
constexpr arm_Flow arm_FlowCall=3ULL;
constexpr arm_Flow arm_FlowReturn=4ULL;
constexpr arm_Flow arm_FlowIndJump=5ULL;
constexpr arm_Flow arm_FlowStop=6ULL;
constexpr arm_Flow arm_FlowIndCall=7ULL;
constexpr int64_t arm_condEQ=0ULL;
constexpr int64_t arm_condNE=1ULL;
constexpr int64_t arm_condCS=2ULL;
constexpr int64_t arm_condCC=3ULL;
constexpr int64_t arm_condMI=4ULL;
constexpr int64_t arm_condPL=5ULL;
constexpr int64_t arm_condVS=6ULL;
constexpr int64_t arm_condVC=7ULL;
constexpr int64_t arm_condHI=8ULL;
constexpr int64_t arm_condLS=9ULL;
constexpr int64_t arm_condGE=10ULL;
constexpr int64_t arm_condLT=11ULL;
constexpr int64_t arm_condGT=12ULL;
constexpr int64_t arm_condLE=13ULL;
constexpr int64_t arm_condAL=14ULL;
constexpr int64_t arm_condNV=15ULL;
std::array<std::string,16> arm_condName=std::array<std::string,16>{std::string("EQ",2),std::string("NE",2),std::string("CS",2),std::string("CC",2),std::string("MI",2),std::string("PL",2),std::string("VS",2),std::string("VC",2),std::string("HI",2),std::string("LS",2),std::string("GE",2),std::string("LT",2),std::string("GT",2),std::string("LE",2),std::string("",0),std::string("NV",2)};
std::array<std::string,16> arm_regName=std::array<std::string,16>{std::string("r0",2),std::string("r1",2),std::string("r2",2),std::string("r3",2),std::string("r4",2),std::string("r5",2),std::string("r6",2),std::string("r7",2),std::string("r8",2),std::string("r9",2),std::string("r10",3),std::string("r11",3),std::string("r12",3),std::string("sp",2),std::string("lr",2),std::string("pc",2)};
std::array<std::string,16> arm_dpOps=std::array<std::string,16>{std::string("AND",3),std::string("EOR",3),std::string("SUB",3),std::string("RSB",3),std::string("ADD",3),std::string("ADC",3),std::string("SBC",3),std::string("RSC",3),std::string("TST",3),std::string("TEQ",3),std::string("CMP",3),std::string("CMN",3),std::string("ORR",3),std::string("MOV",3),std::string("BIC",3),std::string("MVN",3)};
std::array<std::string,4> arm_shiftName=std::array<std::string,4>{std::string("LSL",3),std::string("LSR",3),std::string("ASR",3),std::string("ROR",3)};
constexpr int64_t arm_ModeUSR=16ULL;
constexpr int64_t arm_ModeFIQ=17ULL;
constexpr int64_t arm_ModeIRQ=18ULL;
constexpr int64_t arm_ModeSVC=19ULL;
constexpr int64_t arm_ModeABT=23ULL;
constexpr int64_t arm_ModeUND=27ULL;
constexpr int64_t arm_ModeSYS=31ULL;
constexpr arm_parMode arm_parWrap=0ULL;
constexpr arm_parMode arm_parSat=1ULL;
constexpr arm_parMode arm_parHalve=2ULL;
std::array<std::string,16> arm_thumbALUOps=std::array<std::string,16>{std::string("AND",3),std::string("EOR",3),std::string("LSL",3),std::string("LSR",3),std::string("ASR",3),std::string("ADC",3),std::string("SBC",3),std::string("ROR",3),std::string("TST",3),std::string("NEG",3),std::string("CMP",3),std::string("CMN",3),std::string("ORR",3),std::string("MUL",3),std::string("BIC",3),std::string("MVN",3)};
constexpr arm_Variant arm_V5TE=0ULL;
constexpr arm_Variant arm_V6K=1ULL;
constexpr arm_Variant arm_V4T=2ULL;
Map<std::string,bool> arm_v4tMnem=Map<std::string,bool>{{std::string("BLX",3),true},{std::string("CLZ",3),true},{std::string("BKPT",4),true},{std::string("PLD",3),true},{std::string("LDRD",4),true},{std::string("STRD",4),true},{std::string("QADD",4),true},{std::string("QSUB",4),true},{std::string("QDADD",5),true},{std::string("QDSUB",5),true},{std::string("SMLABB",6),true},{std::string("SMLABT",6),true},{std::string("SMLATB",6),true},{std::string("SMLATT",6),true},{std::string("SMLAWB",6),true},{std::string("SMLAWT",6),true},{std::string("SMULWB",6),true},{std::string("SMULWT",6),true},{std::string("SMLALBB",7),true},{std::string("SMLALBT",7),true},{std::string("SMLALTB",7),true},{std::string("SMLALTT",7),true},{std::string("SMULBB",6),true},{std::string("SMULBT",6),true},{std::string("SMULTB",6),true},{std::string("SMULTT",6),true}};
constexpr int64_t sh4_SRT=1ULL;
constexpr int64_t sh4_SRS=2ULL;
constexpr int64_t sh4_SRQ=256ULL;
constexpr int64_t sh4_SRM=512ULL;
constexpr int64_t sh4_SRFD=32768ULL;
constexpr int64_t sh4_SRBL=268435456ULL;
constexpr int64_t sh4_SRRB=536870912ULL;
constexpr int64_t sh4_SRMD=1073741824ULL;
constexpr int64_t sh4_srIMASK=240ULL;
constexpr int64_t sh4_srMask=1879081971ULL;
constexpr int64_t sh4_FPSCRFR=2097152ULL;
constexpr int64_t sh4_FPSCRSZ=1048576ULL;
constexpr int64_t sh4_FPSCRPR=524288ULL;
constexpr int64_t sh4_FPSCRDN=262144ULL;
constexpr int64_t sh4_fpscrMask=4194303ULL;
constexpr int64_t sh4_scifSCFTDR=4293394444ULL;
constexpr int64_t sh4_scifSCFSR=4293394448ULL;
constexpr sh4_Flow sh4_FlowSeq=0ULL;
constexpr sh4_Flow sh4_FlowBranch=1ULL;
constexpr sh4_Flow sh4_FlowJump=2ULL;
constexpr sh4_Flow sh4_FlowCall=3ULL;
constexpr sh4_Flow sh4_FlowReturn=4ULL;
constexpr sh4_Flow sh4_FlowIndJump=5ULL;
constexpr sh4_Flow sh4_FlowIndCall=6ULL;
constexpr sh4_Flow sh4_FlowStop=7ULL;
constexpr int64_t sh4_tmuUNF=256ULL;
constexpr int64_t sh4_tmuUNIE=32ULL;
constexpr int64_t iso9660_BlockSize=2048ULL;
constexpr int64_t iso9660_pvdLBA=16ULL;
Slice<iso9660_Geometry> iso9660_geometries=Slice<iso9660_Geometry>{iso9660_Geometry{cast<int64_t>(2048ULL),cast<int64_t>(0ULL)},iso9660_Geometry{cast<int64_t>(2352ULL),cast<int64_t>(16ULL)},iso9660_Geometry{cast<int64_t>(2352ULL),cast<int64_t>(24ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(16ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(24ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(0ULL)},iso9660_Geometry{cast<int64_t>(2336ULL),cast<int64_t>(8ULL)},iso9660_Geometry{cast<int64_t>(2336ULL),cast<int64_t>(0ULL)}};
constexpr int64_t dc_armDivisor=9ULL;
constexpr int64_t dc_sampleInstrs=4535ULL;
constexpr int64_t dc_aicaIntTimerA=64ULL;
constexpr int64_t dc_aicaIntTimerB=128ULL;
constexpr int64_t dc_aicaIntTimerC=256ULL;
constexpr int64_t dc_rtcSeconds=1592000512ULL;
constexpr int64_t dc_egAttack=0ULL;
constexpr int64_t dc_egDecay1=1ULL;
constexpr int64_t dc_egDecay2=2ULL;
constexpr int64_t dc_egRelease=3ULL;
constexpr int64_t dc_egSilence=1023ULL;
std::array<int64_t,1024> dc_egGainQ16={};
std::array<int64_t,256> dc_tlGainQ16={};
std::array<int64_t,16> dc_sdlGainQ16={};
std::array<int64_t,16> dc_panGainQ16={};
std::array<int32_t,8> dc_adpcmScale=std::array<int32_t,8>{cast<int32_t>(230ULL),cast<int32_t>(230ULL),cast<int32_t>(230ULL),cast<int32_t>(230ULL),cast<int32_t>(307ULL),cast<int32_t>(409ULL),cast<int32_t>(512ULL),cast<int32_t>(614ULL)};
constexpr int64_t dc_adpcmInitStep=127ULL;
constexpr int64_t dc_gdCmdPIORead=16ULL;
constexpr int64_t dc_gdCmdDMARead=17ULL;
constexpr int64_t dc_gdCmdGetTOC2=19ULL;
constexpr int64_t dc_gdCmdInit=24ULL;
constexpr int64_t dc_trapSysinfo=2684354816ULL;
constexpr int64_t dc_trapRomfont=2684354820ULL;
constexpr int64_t dc_trapFlashrom=2684354824ULL;
constexpr int64_t dc_trapGdrom=2684354828ULL;
constexpr int64_t dc_trapMenu=2684354832ULL;
constexpr int64_t dc_istRenderDone=7ULL;
constexpr int64_t dc_istVBlankIn=8ULL;
constexpr int64_t dc_istVBlankOut=16ULL;
constexpr int64_t dc_istCh2DMA=524288ULL;
std::array<uint32_t,5> dc_taListDoneBit=std::array<uint32_t,5>{cast<uint32_t>(128ULL),cast<uint32_t>(256ULL),cast<uint32_t>(512ULL),cast<uint32_t>(1024ULL),cast<uint32_t>(2097152ULL)};
constexpr int64_t dc_RAMSize=16777216ULL;
constexpr int64_t dc_VRAMSize=8388608ULL;
constexpr int64_t dc_AICARAMSize=2097152ULL;
constexpr int64_t dc_FlashSize=262144ULL;
constexpr int64_t dc_ramBase=201326592ULL;
constexpr int64_t dc_vram64=67108864ULL;
constexpr int64_t dc_vram32=83886080ULL;
constexpr int64_t dc_aicaRAM=8388608ULL;
constexpr int64_t dc_flashBase=2097152ULL;
constexpr int64_t dc_sbBase=6252544ULL;
constexpr int64_t dc_gdBase=6254592ULL;
constexpr int64_t dc_pvrBase=6258688ULL;
constexpr int64_t dc_aicaBase=7340032ULL;
constexpr int64_t dc_taBase=268435456ULL;
constexpr int64_t dc_taEnd=335544320ULL;
constexpr int64_t dc_fieldInstructions=3333333ULL;
constexpr int64_t dc_sbMDSTAR=6253572ULL;
constexpr int64_t dc_sbMDTSEL=6253584ULL;
constexpr int64_t dc_sbMDEN=6253588ULL;
constexpr int64_t dc_sbMDST=6253592ULL;
constexpr int64_t dc_istMapleDMA=4096ULL;
constexpr int64_t dc_PadC=1ULL;
constexpr int64_t dc_PadB=2ULL;
constexpr int64_t dc_PadA=4ULL;
constexpr int64_t dc_PadStart=8ULL;
constexpr int64_t dc_PadUp=16ULL;
constexpr int64_t dc_PadDown=32ULL;
constexpr int64_t dc_PadLeft=64ULL;
constexpr int64_t dc_PadRight=128ULL;
constexpr int64_t dc_PadZ=256ULL;
constexpr int64_t dc_PadY=512ULL;
constexpr int64_t dc_PadX=1024ULL;
constexpr int64_t dc_PadD=2048ULL;
Map<std::string,uint16_t> dc_padButtonNames=Map<std::string,uint16_t>{{std::string("a",1),cast<uint16_t>(4ULL)},{std::string("b",1),cast<uint16_t>(2ULL)},{std::string("x",1),cast<uint16_t>(1024ULL)},{std::string("y",1),cast<uint16_t>(512ULL)},{std::string("c",1),cast<uint16_t>(1ULL)},{std::string("z",1),cast<uint16_t>(256ULL)},{std::string("d",1),cast<uint16_t>(2048ULL)},{std::string("start",5),cast<uint16_t>(8ULL)},{std::string("up",2),cast<uint16_t>(16ULL)},{std::string("down",4),cast<uint16_t>(32ULL)},{std::string("left",4),cast<uint16_t>(64ULL)},{std::string("right",5),cast<uint16_t>(128ULL)}};
int64_t dc_debugPixelX=cast<int64_t>(-1ULL);
int64_t dc_debugPixelY=cast<int64_t>(-1ULL);
constexpr int64_t dc_spinWindow=6666666ULL;
constexpr int64_t dc_fbRCtrl=17ULL;
constexpr int64_t dc_fbRSOF1=20ULL;
constexpr int64_t dc_fbRSize=23ULL;

#include "adapters.h"
// tools/cpu/arm/arm.go:109:1
uint32_t arm_signExtend(uint32_t v,uint64_t n){
{
uint32_t m = shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((n - cast<uint64_t>(1ULL)))));
return cast<uint32_t>(((cast<uint32_t>((v ^ m))) - m));
}
}
// tools/cpu/arm/arm.go:116:1
std::string arm_imm(uint32_t v){
{
if ((v < cast<uint32_t>(10ULL))) {
return go_fmt_Sprintf(std::string("#%d",3),v);
}
return go_fmt_Sprintf(std::string("#0x%X",5),v);
}
}
// tools/cpu/arm/arm.go:127:1
arm_Inst arm_Decode(Slice<uint8_t> code,uint32_t addr,bool thumb){
{
if (thumb) {
return arm_DecodeThumb(code,addr);
}
return arm_DecodeARM(code,addr);
}
}
// tools/cpu/arm/arm.go:136:1
std::tuple<uint32_t,bool> arm_word(Slice<uint8_t> code){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return {cast<uint32_t>(0ULL),false};
}
return {cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL)))),true};
}
}
// tools/cpu/arm/arm.go:144:1
arm_Inst arm_DecodeARM(Slice<uint8_t> code,uint32_t addr){
{
auto tmp1 = arm_word(code);
uint32_t w = std::get<0>(tmp1);
bool ok = std::get<1>(tmp1);
if ((!ok)) {
return arm_Inst{addr,len(code),std::string(".word",5),std::string(".word ; truncated",17),cast<arm_Flow>(6ULL),{},{},{},{},cast<int64_t>(14ULL)};
}
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
arm_Inst in = arm_Inst{addr,cast<int64_t>(4ULL),{},{},cast<arm_Flow>(0ULL),{},{},{},{},cond};
if ((cond == cast<int64_t>(15ULL))) {
return arm_decodeUncond(w,addr,in);
}
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return arm_decodeDataMisc(w,addr,in,false);
break;}
case cast<uint32_t>(1ULL):{
return arm_decodeDataMisc(w,addr,in,true);
break;}
case cast<uint32_t>(2ULL):{
return arm_decodeSingle(w,in,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_undef(w,in);
}
return arm_decodeSingle(w,in,true);
break;}
case cast<uint32_t>(4ULL):{
return arm_decodeBlock(w,in);
break;}
case cast<uint32_t>(5ULL):{
return arm_decodeBranch(w,addr,in);
break;}
case cast<uint32_t>(6ULL):{
return arm_decodeCoproLS(w,in);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("SWI",3) + arm_cn(cond));
in.Text = go_fmt_Sprintf(std::string("%s #0x%X",8),in.Mnem,cast<uint32_t>((w & cast<uint32_t>(16777215ULL))));
return in;
}
return arm_decodeCopro(w,in);
break;}
}}
}
}
// tools/cpu/arm/arm.go:186:1
std::string arm_cn(int64_t cond){
{
return arm_condName[cond];
}
}
// tools/cpu/arm/arm.go:188:1
arm_Inst arm_undef(uint32_t w,arm_Inst in){
{
auto tmp2 = std::make_tuple(std::string(".word",5),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp2);
in.Flow = std::get<1>(tmp2);
in.Text = go_fmt_Sprintf(std::string(".word 0x%08X",12),w);
return in;
}
}
// tools/cpu/arm/arm.go:199:1
arm_Inst arm_decodeDataMisc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return arm_decodeExtension(w,in);
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
return arm_decodeMisc(w,in,immForm);
}
return arm_decodeDataProc(w,addr,in,immForm);
}
}
// tools/cpu/arm/arm.go:215:1
arm_Inst arm_decodeDataProc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
std::string op2={};
if (immForm) {
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
op2 = arm_imm(v);
}
else {
op2 = arm_shiftOperand(w);
}
std::string sfx = std::string("",0);
if ((s == cast<uint32_t>(1ULL))) {
sfx = std::string("S",1);
}
std::string name = arm_dpOps[op];
in.Mnem = ((name + arm_cn(in.Cond)) + sfx);
{
switch(op){
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):{
in.Mnem = (arm_dpOps[op] + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[rn],op2);
break;}
case cast<uint32_t>(13ULL):case cast<uint32_t>(15ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[rd],op2);
break;}
default:{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rn],op2);
break;}
}}
if ((((((rd == cast<uint32_t>(15ULL)) && (op != cast<uint32_t>(8ULL))) && (op != cast<uint32_t>(9ULL))) && (op != cast<uint32_t>(10ULL))) && (op != cast<uint32_t>(11ULL)))) {
arm_classifyPCWrite((&in),w,op);
}
return in;
}
}
// tools/cpu/arm/arm.go:255:1
void arm_classifyPCWrite(arm_Inst* in,uint32_t w,uint32_t op){
{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
bool isReg = ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(0ULL)));
if ((((op == cast<uint32_t>(13ULL)) && isReg) && (rm == cast<uint32_t>(14ULL)))) {
in->Flow = cast<arm_Flow>(4ULL);
return ;
}
in->Flow = cast<arm_Flow>(5ULL);
}
}
// tools/cpu/arm/arm.go:266:1
std::string arm_shiftOperand(uint32_t w){
{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t styp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
return go_fmt_Sprintf(std::string("%s, %s %s",9),arm_regName[rm],arm_shiftName[styp],arm_regName[rs]);
}
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
{
switch(styp){
case cast<uint32_t>(0ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
return arm_regName[rm];
}
return go_fmt_Sprintf(std::string("%s, LSL #%d",11),arm_regName[rm],amt);
break;}
case cast<uint32_t>(1ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, LSR #%d",11),arm_regName[rm],amt);
break;}
case cast<uint32_t>(2ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, ASR #%d",11),arm_regName[rm],amt);
break;}
default:{
if ((amt == cast<uint32_t>(0ULL))) {
return go_fmt_Sprintf(std::string("%s, RRX",7),arm_regName[rm]);
}
return go_fmt_Sprintf(std::string("%s, ROR #%d",11),arm_regName[rm],amt);
break;}
}}
}
}
// tools/cpu/arm/arm.go:299:1
uint32_t arm_ror32(uint32_t v,uint32_t n){
{
n &= cast<uint32_t>(31ULL);
return cast<uint32_t>((shr<uint32_t>(v,n) | shl<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(32ULL) - n))))));
}
}
// tools/cpu/arm/arm.go:307:1
arm_Inst arm_decodeMisc(uint32_t w,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
return arm_decodeMSRimm(w,in);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_decodeSignedMul(w,in);
}
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)));
{
switch(op2){
case cast<uint32_t>(0ULL):{
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MRS",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],psr);
return in;
}
return arm_decodeMSRreg(w,in);
break;}
case cast<uint32_t>(1ULL):{
{
switch(op){
case cast<uint32_t>(9ULL):{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
in.Mnem = (std::string("BX",2) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,arm_regName[rm]);
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
return in;
break;}
case cast<uint32_t>(11ULL):{
in.Mnem = (std::string("CLZ",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
return in;
break;}
}}
break;}
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(9ULL))) {
in.Mnem = (std::string("BLX",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
in.Flow = cast<arm_Flow>(7ULL);
return in;
}
break;}
case cast<uint32_t>(5ULL):{
std::string sat = std::array<std::string,4>{std::string("QADD",4),std::string("QSUB",4),std::string("QDADD",5),std::string("QDSUB",5)}[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = (sat + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
break;}
case cast<uint32_t>(7ULL):{
if ((op == cast<uint32_t>(9ULL))) {
in.Mnem = std::string("BKPT",4);
uint32_t imm16 = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(65520ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
in.Text = go_fmt_Sprintf(std::string("BKPT #0x%X",10),imm16);
in.Flow = cast<arm_Flow>(6ULL);
return in;
}
break;}
}}
return arm_undef(w,in);
}
}
// tools/cpu/arm/arm.go:378:1
arm_Inst arm_decodeMSRreg(uint32_t w,arm_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MSR",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm_msrFields(w),arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:389:1
arm_Inst arm_decodeMSRimm(uint32_t w,arm_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
in.Mnem = (std::string("MSR",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm_msrFields(w),arm_imm(v));
return in;
}
}
// tools/cpu/arm/arm.go:401:1
std::string arm_msrFields(uint32_t w){
{
uint32_t m = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
if ((m == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
std::string s = std::string("_",1);
{auto&& tmp3 = Slice<std::string>{std::string("c",1),std::string("x",1),std::string("s",1),std::string("f",1)};
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto i=tmp4;auto f=tmp3[tmp4];if ((cast<uint32_t>((m & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
s += f;
}
}}
return s;
}
}
// tools/cpu/arm/arm.go:417:1
arm_Inst arm_decodeSignedMul(uint32_t w,arm_Inst in){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t x = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t y = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
auto bt = [&](uint32_t b)->uint8_t{
if ((b == cast<uint32_t>(0ULL))) {
return cast<uint8_t>(66ULL);
}
return cast<uint8_t>(84ULL);
}
;
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(8ULL):{
in.Mnem = go_fmt_Sprintf(std::string("SMLA%c%c%s",10),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
break;}
case cast<uint32_t>(9ULL):{
if ((x == cast<uint32_t>(0ULL))) {
in.Mnem = go_fmt_Sprintf(std::string("SMLAW%c%s",9),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
}
else {
in.Mnem = go_fmt_Sprintf(std::string("SMULW%c%s",9),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
}
break;}
case cast<uint32_t>(10ULL):{
in.Mnem = go_fmt_Sprintf(std::string("SMLAL%c%c%s",11),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rn],arm_regName[rd],arm_regName[rm],arm_regName[rs]);
break;}
default:{
in.Mnem = go_fmt_Sprintf(std::string("SMUL%c%c%s",10),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
break;}
}}
return in;
}
}
// tools/cpu/arm/arm.go:456:1
arm_Inst arm_decodeExtension(uint32_t w,arm_Inst in){
{
uint32_t sh = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((sh == cast<uint32_t>(0ULL))) {
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
std::string b = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
b = std::string("B",1);
}
in.Mnem = ((std::string("SWP",3) + arm_cn(in.Cond)) + b);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
return arm_decodeMulLong(w,in);
}
else {
return arm_decodeMul(w,in);
}
}
tmp5:;
}
return arm_decodeHalf(w,in,sh);
}
}
// tools/cpu/arm/arm.go:479:1
arm_Inst arm_decodeMul(uint32_t w,arm_Inst in){
{
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = ((std::string("MLA",3) + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
}
else {
in.Mnem = ((std::string("MUL",3) + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
}
return in;
}
}
// tools/cpu/arm/arm.go:499:1
arm_Inst arm_decodeMulLong(uint32_t w,arm_Inst in){
{
std::array<std::string,4> names = std::array<std::string,4>{std::string("UMULL",5),std::string("UMLAL",5),std::string("SMULL",5),std::string("SMLAL",5)};
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
in.Mnem = ((names[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))] + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:513:1
arm_Inst arm_decodeHalf(uint32_t w,arm_Inst in,uint32_t sh){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("",0);
{
if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(1ULL)))){
name = std::string("STRH",4);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(1ULL)))){
name = std::string("LDRH",4);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(2ULL)))){
name = std::string("LDRSB",5);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(3ULL)))){
name = std::string("LDRSH",5);
}
else {
if ((sh == cast<uint32_t>(2ULL))) {
name = std::string("LDRD",4);
}
else {
name = std::string("STRD",4);
}
}
}
tmp6:;
in.Mnem = (name + arm_cn(in.Cond));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string sign = std::string("+",1);
if ((u == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
std::string off={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t v = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(240ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
if ((v == cast<uint32_t>(0ULL))) {
off = std::string("",0);
}
else {
off = go_fmt_Sprintf(std::string(", #%s0x%X",9),arm_plusMinus(sign),v);
}
}
else {
off = go_fmt_Sprintf(std::string(", %s%s",6),arm_plusMinus(sign),arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
}
in.Text = arm_addrForm(in.Mnem,rd,rn,off,p,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))));
return in;
}
}
// tools/cpu/arm/arm.go:557:1
std::string arm_plusMinus(std::string s){
{
if ((s == std::string("-",1))) {
return std::string("-",1);
}
return std::string("",0);
}
}
// tools/cpu/arm/arm.go:565:1
arm_Inst arm_decodeSingle(uint32_t w,arm_Inst in,bool regOff){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
in.Mnem = (name + arm_cn(in.Cond));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string sign = std::string("",0);
if ((u == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
std::string off={};
if ((!regOff)) {
uint32_t v = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
if ((v == cast<uint32_t>(0ULL))) {
off = std::string("",0);
}
else {
off = go_fmt_Sprintf(std::string(", #%s0x%X",9),sign,v);
}
}
else {
off = ((std::string(", ",2) + sign) + arm_shiftOperand(w));
}
in.Text = arm_addrForm(in.Mnem,rd,rn,off,p,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))));
if (((l == cast<uint32_t>(1ULL)) && (rd == cast<uint32_t>(15ULL)))) {
in.Flow = cast<arm_Flow>(5ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:609:1
std::string arm_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb){
{
if ((p == cast<uint32_t>(1ULL))) {
std::string bang = std::string("",0);
if ((wb == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
return go_fmt_Sprintf(std::string("%s %s, [%s%s]%s",15),mnem,arm_regName[rd],arm_regName[rn],off,bang);
}
return go_fmt_Sprintf(std::string("%s %s, [%s]%s",13),mnem,arm_regName[rd],arm_regName[rn],off);
}
}
// tools/cpu/arm/arm.go:624:1
arm_Inst arm_decodeBlock(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string base = std::string("STM",3);
if ((l == cast<uint32_t>(1ULL))) {
base = std::string("LDM",3);
}
std::string mode = std::array<std::string,4>{std::string("DA",2),std::string("DB",2),std::string("IA",2),std::string("IB",2)}[cast<uint32_t>((shl<uint32_t>(u,cast<int64_t>(1ULL)) | p))];
in.Mnem = ((base + arm_cn(in.Cond)) + mode);
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
std::string bang = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
std::string usr = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
usr = std::string("^",1);
}
in.Text = go_fmt_Sprintf(std::string("%s %s%s, {%s}%s",15),in.Mnem,arm_regName[rn],bang,arm_regList(cast<uint32_t>((w & cast<uint32_t>(65535ULL)))),usr);
if (((l == cast<uint32_t>(1ULL)) && (cast<uint32_t>((w & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)))) {
in.Flow = cast<arm_Flow>(4ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:654:1
std::string arm_regList(uint32_t mask){
{
Slice<std::string> parts={};
int64_t i = cast<int64_t>(0ULL);
{;for (;(i < cast<int64_t>(16ULL));){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) == cast<uint32_t>(0ULL))) {
i++;
continue;
}
int64_t j = i;
{;for (;((cast<int64_t>((j + cast<int64_t>(1ULL))) < cast<int64_t>(16ULL)) && (cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(cast<int64_t>((j + cast<int64_t>(1ULL)))))))) != cast<uint32_t>(0ULL)));){
j++;
}
}{
if ((j == i)){
parts = append(parts,Slice<std::string>{arm_regName[i]});
}
else if ((j == cast<int64_t>((i + cast<int64_t>(1ULL))))){
parts = append(parts,Slice<std::string>{arm_regName[i],arm_regName[j]});
}
else {
parts = append(parts,Slice<std::string>{((arm_regName[i] + std::string("-",1)) + arm_regName[j])});
}
}
tmp7:;
i = cast<int64_t>((j + cast<int64_t>(1ULL)));
}
}return arm_joinComma(parts);
}
}
// tools/cpu/arm/arm.go:679:1
std::string arm_joinComma(Slice<std::string> parts){
{
std::string out = std::string("",0);
{auto&& tmp8 = parts;
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto i=tmp9;auto p=tmp8[tmp9];if ((i > cast<int64_t>(0ULL))) {
out += std::string(", ",2);
}
out += p;
}}
return out;
}
}
// tools/cpu/arm/arm.go:692:1
arm_Inst arm_decodeBranch(uint32_t w,uint32_t addr,arm_Inst in){
{
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + off));
auto tmp10 = std::make_tuple(target,true,false);
in.Target = std::get<0>(tmp10);
in.HasTarget = std::get<1>(tmp10);
in.TargetThumb = std::get<2>(tmp10);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("BL",2) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
in.Flow = cast<arm_Flow>(3ULL);
return in;
}
in.Mnem = (std::string("B",1) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
if ((in.Cond == cast<int64_t>(14ULL))) {
in.Flow = cast<arm_Flow>(2ULL);
}
else {
in.Flow = cast<arm_Flow>(1ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:715:1
arm_Inst arm_decodeUncond(uint32_t w,uint32_t addr,arm_Inst in){
{
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(5ULL))){
uint32_t h = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t off = cast<uint32_t>((shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL)) | shl<uint32_t>(h,cast<int64_t>(1ULL))));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + off));
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX 0x%08X",10),target);
auto tmp12 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,true);
in.Flow = std::get<0>(tmp12);
in.Target = std::get<1>(tmp12);
in.HasTarget = std::get<2>(tmp12);
in.TargetThumb = std::get<3>(tmp12);
return in;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(26ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(247ULL))) == cast<uint32_t>(85ULL)))){
in.Mnem = std::string("PLD",3);
in.Text = go_fmt_Sprintf(std::string("PLD [%s]",8),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
tmp11:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/arm.go:740:1
arm_Inst arm_decodeCoproLS(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STC",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDC",3);
}
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, c%d, [%s]",17),in.Mnem,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:754:1
arm_Inst arm_decodeCopro(uint32_t w,arm_Inst in){
{
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (std::string("CDP",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, c%d, c%d, c%d",26),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))));
return in;
}
std::string name = std::string("MCR",3);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
name = std::string("MRC",3);
}
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, %s, c%d, c%d, #%d",30),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL))));
return in;
}
}
// tools/cpu/arm/armexec.go:10:1
int64_t arm_CPU_Step(arm_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
c->cur = c->R[cast<int64_t>(15ULL)];
c->branched = false;
c->Instrs++;
if (c->Thumb) {
return arm_CPU_stepThumb(c);
}
return arm_CPU_stepARM(c);
}
}
// tools/cpu/arm/armexec.go:23:1
uint32_t arm_CPU_boolToU(arm_CPU* c,bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/arm/armexec.go:31:1
int64_t arm_CPU_stepARM(arm_CPU* c){
{
uint32_t w = arm_CPU_read32aligned(c,c->cur);
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
if ((cond == cast<int64_t>(15ULL))) {
if ((arm_Variant_isV6(c->Arch) && arm_CPU_execUncondARMv6(c,w))) {
return cast<int64_t>(1ULL);
}
return arm_CPU_execUncondARM(c,w);
}
if ((!arm_CPU_cond(c,cond))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return cast<int64_t>(1ULL);
}
if ((arm_Variant_isV6(c->Arch) && arm_CPU_execARMv6(c,w))) {
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<int64_t>(1ULL);
}
arm_CPU_execARM(c,w);
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/armexec.go:59:1
int64_t arm_CPU_execUncondARM(arm_CPU* c,uint32_t w){
{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("cond=0b1111 (BLX immediate / PLD)",33));
return cast<int64_t>(1ULL);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(5ULL))) {
uint32_t h = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t off = cast<uint32_t>((shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL)) | shl<uint32_t>(h,cast<int64_t>(1ULL))));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
c->Thumb = true;
c->R[cast<int64_t>(15ULL)] = ((cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(8ULL))) + off))) & ~(cast<uint32_t>(1ULL)));
c->branched = true;
return cast<int64_t>(3ULL);
}
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/armexec.go:82:1
void arm_CPU_execARM(arm_CPU* c,uint32_t w){
{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_execDataMisc(c,w,false);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_execDataMisc(c,w,true);
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_execSingle(c,w,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_Halt(c,std::string("undefined media instruction 0x%08X at 0x%08X",44),w,c->cur);
return ;
}
arm_CPU_execSingle(c,w,true);
break;}
case cast<uint32_t>(4ULL):{
arm_CPU_execBlock(c,w);
break;}
case cast<uint32_t>(5ULL):{
arm_CPU_execBranch(c,w);
break;}
case cast<uint32_t>(6ULL):{
arm_CPU_Halt(c,std::string("unimplemented coprocessor transfer 0x%08X at 0x%08X",51),w,c->cur);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_execSWI(c,w);
}
else {
arm_CPU_execCopro(c,w);
}
break;}
}}
}
}
// tools/cpu/arm/armexec.go:111:1
void arm_CPU_execDataMisc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
arm_CPU_execExtension(c,w);
return ;
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
arm_CPU_execMisc(c,w,immForm);
return ;
}
arm_CPU_execDataProc(c,w,immForm);
}
}
// tools/cpu/arm/armexec.go:127:1
std::tuple<uint32_t,uint32_t> arm_CPU_dpOperand(arm_CPU* c,uint32_t w,bool immForm){
{
if (immForm) {
uint32_t rot = cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL)));
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),rot);
if ((rot == cast<uint32_t>(0ULL))) {
return {v,arm_CPU_boolToU(c,c->C)};
}
return {v,cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t typ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_CPU_shift(c,typ,arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))),rm,true,arm_CPU_boolToU(c,c->C));
}
return arm_CPU_shift(c,typ,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),rm,false,arm_CPU_boolToU(c,c->C));
}
}
// tools/cpu/arm/armexec.go:144:1
void arm_CPU_execDataProc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
bool setFlags = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
auto tmp13 = arm_CPU_dpOperand(c,w,immForm);
uint32_t op2 = std::get<0>(tmp13);
uint32_t shc = std::get<1>(tmp13);
uint32_t a = arm_CPU_reg(c,rn);
auto tmp14 = std::make_tuple(c->N,c->Z,c->C,c->V);
bool sn = std::get<0>(tmp14);
bool sz = std::get<1>(tmp14);
bool sc = std::get<2>(tmp14);
bool sv = std::get<3>(tmp14);
uint32_t res={};
bool logical = true;
{
switch(op){
case cast<uint32_t>(0ULL):{
res = cast<uint32_t>((a & op2));
break;}
case cast<uint32_t>(1ULL):{
res = cast<uint32_t>((a ^ op2));
break;}
case cast<uint32_t>(2ULL):{
auto tmp15 = std::make_tuple(arm_CPU_sub_(c,a,op2,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp15);
logical = std::get<1>(tmp15);
break;}
case cast<uint32_t>(3ULL):{
auto tmp16 = std::make_tuple(arm_CPU_sub_(c,op2,a,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp16);
logical = std::get<1>(tmp16);
break;}
case cast<uint32_t>(4ULL):{
auto tmp17 = std::make_tuple(arm_CPU_add(c,a,op2,cast<uint32_t>(0ULL)),false);
res = std::get<0>(tmp17);
logical = std::get<1>(tmp17);
break;}
case cast<uint32_t>(5ULL):{
auto tmp18 = std::make_tuple(arm_CPU_add(c,a,op2,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp18);
logical = std::get<1>(tmp18);
break;}
case cast<uint32_t>(6ULL):{
auto tmp19 = std::make_tuple(arm_CPU_sub_(c,a,op2,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp19);
logical = std::get<1>(tmp19);
break;}
case cast<uint32_t>(7ULL):{
auto tmp20 = std::make_tuple(arm_CPU_sub_(c,op2,a,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp20);
logical = std::get<1>(tmp20);
break;}
case cast<uint32_t>(8ULL):{
res = cast<uint32_t>((a & op2));
break;}
case cast<uint32_t>(9ULL):{
res = cast<uint32_t>((a ^ op2));
break;}
case cast<uint32_t>(10ULL):{
arm_CPU_sub_(c,a,op2,cast<uint32_t>(1ULL));
logical = false;
break;}
case cast<uint32_t>(11ULL):{
arm_CPU_add(c,a,op2,cast<uint32_t>(0ULL));
logical = false;
break;}
case cast<uint32_t>(12ULL):{
res = cast<uint32_t>((a | op2));
break;}
case cast<uint32_t>(13ULL):{
res = op2;
break;}
case cast<uint32_t>(14ULL):{
res = (a & ~(op2));
break;}
default:{
res = cast<uint32_t>(~op2);
break;}
}}
bool isTest = ((op >= cast<uint32_t>(8ULL)) && (op <= cast<uint32_t>(11ULL)));
if (isTest) {
if (logical) {
arm_CPU_setNZ(c,res);
auto tmp21 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp21);
c->V = std::get<1>(tmp21);
}
return ;
}
if (setFlags) {
if (logical) {
arm_CPU_setNZ(c,res);
auto tmp22 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp22);
c->V = std::get<1>(tmp22);
}
}
else {
auto tmp23 = std::make_tuple(sn,sz,sc,sv);
c->N = std::get<0>(tmp23);
c->Z = std::get<1>(tmp23);
c->C = std::get<2>(tmp23);
c->V = std::get<3>(tmp23);
}
if ((rd == cast<uint32_t>(15ULL))) {
if (setFlags) {
arm_CPU_SetCPSR(c,arm_CPU_SPSR(c));
}
arm_CPU_branchTo(c,res);
return ;
}
arm_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm/armexec.go:223:1
void arm_CPU_branchTo(arm_CPU* c,uint32_t v){
{
if (c->Thumb) {
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(1ULL)));
}
else {
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(3ULL)));
}
c->branched = true;
}
}
// tools/cpu/arm/armexec.go:234:1
void arm_CPU_bxTo(arm_CPU* c,uint32_t v){
{
c->Thumb = (cast<uint32_t>((v & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
arm_CPU_branchTo(c,v);
}
}
// tools/cpu/arm/armexec.go:239:1
void arm_CPU_execBranch(arm_CPU* c,uint32_t w){
{
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(8ULL))) + off));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
arm_CPU_branchTo(c,target);
}
}
// tools/cpu/arm/armexec.go:249:1
void arm_CPU_execExtension(arm_CPU* c,uint32_t w){
{
uint32_t sh = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((sh == cast<uint32_t>(0ULL))) {
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm_CPU_execSwap(c,w);
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm_CPU_execMulLong(c,w);
}
else {
arm_CPU_execMul(c,w);
}
}
tmp24:;
return ;
}
arm_CPU_execHalf(c,w,sh);
}
}
// tools/cpu/arm/armexec.go:265:1
void arm_CPU_execMul(arm_CPU* c,uint32_t w){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t res = cast<uint32_t>((arm_CPU_reg(c,rm) * arm_CPU_reg(c,rs)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
res += arm_CPU_reg(c,rn);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_setNZ(c,res);
}
arm_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm/armexec.go:280:1
void arm_CPU_execMulLong(arm_CPU* c,uint32_t w){
{
uint32_t rdHi = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rdLo = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
bool accum = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint64_t result={};
if (signed_) {
result = cast<uint64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>(rm)) * cast<int64_t>(cast<int32_t>(rs)))));
}
else {
result = cast<uint64_t>((cast<uint64_t>(rm) * cast<uint64_t>(rs)));
}
if (accum) {
result += cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(arm_CPU_reg(c,rdHi)),cast<int64_t>(32ULL)) | cast<uint64_t>(arm_CPU_reg(c,rdLo))));
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->Z = (result == cast<uint64_t>(0ULL));
c->N = (cast<uint64_t>((result & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL));
}
arm_CPU_setReg(c,rdLo,cast<uint32_t>(result));
arm_CPU_setReg(c,rdHi,cast<uint32_t>(shr<uint64_t>(result,cast<int64_t>(32ULL))));
}
}
// tools/cpu/arm/armexec.go:305:1
void arm_CPU_execSwap(arm_CPU* c,uint32_t w){
{
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t old = cast<uint32_t>(arm_CPU_read8(c,rn));
arm_CPU_write8(c,rn,cast<uint8_t>(arm_CPU_reg(c,rm)));
arm_CPU_setReg(c,rd,old);
}
else {
uint32_t old = arm_CPU_read32(c,rn);
arm_CPU_write32aligned(c,rn,arm_CPU_reg(c,rm));
arm_CPU_setReg(c,rd,old);
}
}
}
// tools/cpu/arm/armexec.go:320:1
void arm_CPU_execHalf(arm_CPU* c,uint32_t w,uint32_t sh){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t base = arm_CPU_reg(c,rn);
uint32_t off={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
off = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(240ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
}
else {
off = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
}
uint32_t addr = base;
if ((p == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
addr += off;
}
else {
addr -= off;
}
}
auto writeback = [&]()->void{
if ((p == cast<uint32_t>(0ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
base += off;
}
else {
base -= off;
}
arm_CPU_setReg(c,rn,base);
}
else if ((wbit == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,addr);
}
}
;
{
if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(1ULL)))){
uint32_t v = arm_CPU_read16(c,addr);
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(2ULL)))){
uint32_t v = cast<uint32_t>(cast<int32_t>(cast<int8_t>(arm_CPU_read8(c,addr))));
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(3ULL)))){
uint32_t v = cast<uint32_t>(cast<int32_t>(cast<int16_t>(arm_CPU_read16(c,addr))));
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(1ULL)))){
writeback();
arm_CPU_write16(c,addr,arm_CPU_reg(c,rd));
}
else if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(2ULL)))){
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("LDRD",4));
return ;
}
uint32_t lo = arm_CPU_read32(c,addr);
uint32_t hi = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
writeback();
arm_CPU_setReg(c,rd,lo);
arm_CPU_setReg(c,cast<uint32_t>((rd + cast<uint32_t>(1ULL))),hi);
}
else {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("STRD",4));
return ;
}
uint32_t lo = arm_CPU_reg(c,rd);
uint32_t hi = arm_CPU_reg(c,cast<uint32_t>((rd + cast<uint32_t>(1ULL))));
writeback();
arm_CPU_write32(c,addr,lo);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),hi);
}
}
tmp25:;
}
}
// tools/cpu/arm/armexec.go:399:1
void arm_CPU_execSingle(arm_CPU* c,uint32_t w,bool regOff){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t base = arm_CPU_reg(c,rn);
uint32_t off={};
if ((!regOff)) {
off = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
}
else {
auto tmp26 = arm_CPU_shift(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))),false,arm_CPU_boolToU(c,c->C));
off = std::get<0>(tmp26);
}
uint32_t addr = base;
if ((p == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
addr += off;
}
else {
addr -= off;
}
}
auto writeback = [&]()->void{
if ((p == cast<uint32_t>(0ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
base += off;
}
else {
base -= off;
}
arm_CPU_setReg(c,rn,base);
}
else if ((wbit == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,addr);
}
}
;
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v={};
if ((b == cast<uint32_t>(1ULL))) {
v = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else {
v = arm_CPU_read32(c,addr);
}
writeback();
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_bxTo(c,v);
}
else {
arm_CPU_setReg(c,rd,v);
}
}
else {
uint32_t v = arm_CPU_reg(c,rd);
if ((rd == cast<uint32_t>(15ULL))) {
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
if ((b == cast<uint32_t>(1ULL))) {
arm_CPU_write8(c,addr,cast<uint8_t>(v));
}
else {
arm_CPU_write32(c,addr,v);
}
writeback();
}
}
}
// tools/cpu/arm/armexec.go:462:1
void arm_CPU_execBlock(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t mask = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
uint32_t base = c->R[rn];
uint32_t n = cast<uint32_t>(go_bits_OnesCount32(mask));
uint32_t start={};
uint32_t final={};
if ((u == cast<uint32_t>(1ULL))) {
final = cast<uint32_t>((base + cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
if ((p == cast<uint32_t>(1ULL))) {
start = cast<uint32_t>((base + cast<uint32_t>(4ULL)));
}
else {
start = base;
}
}
else {
final = cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
if ((p == cast<uint32_t>(1ULL))) {
start = cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
}
else {
start = cast<uint32_t>((cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL))))) + cast<uint32_t>(4ULL)));
}
}
bool userBank = ((s == cast<uint32_t>(1ULL)) && (cast<uint32_t>((mask & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL)));
uint32_t addr = start;
bool loadedRn = false;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v = arm_CPU_read32aligned(c,addr);
{
if ((i == cast<uint32_t>(15ULL))){
if ((s == cast<uint32_t>(1ULL))) {
arm_CPU_SetCPSR(c,arm_CPU_SPSR(c));
}
arm_CPU_bxTo(c,v);
}
else if (userBank){
arm_CPU_setUserReg(c,i,v);
}
else {
c->R[i] = v;
if ((i == rn)) {
loadedRn = true;
}
}
}
tmp27:;
}
else {
uint32_t v={};
{
if ((i == cast<uint32_t>(15ULL))){
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
else if (userBank){
v = arm_CPU_userReg(c,i);
}
else {
v = c->R[i];
}
}
tmp28:;
arm_CPU_write32aligned(c,addr,v);
}
addr += cast<uint32_t>(4ULL);
}
}if (((wbit == cast<uint32_t>(1ULL)) && (!(((l == cast<uint32_t>(1ULL)) && loadedRn))))) {
arm_CPU_setReg(c,rn,final);
}
}
}
// tools/cpu/arm/armexec.go:556:1
bool arm_CPU_inUserBank(arm_CPU* c){
{
return ((c->Mode == cast<uint32_t>(16ULL)) || (c->Mode == cast<uint32_t>(31ULL)));
}
}
// tools/cpu/arm/armexec.go:558:1
uint32_t arm_CPU_userReg(arm_CPU* c,uint32_t i){
{
{
if ((((i >= cast<uint32_t>(8ULL)) && (i <= cast<uint32_t>(12ULL))) && (c->Mode == cast<uint32_t>(17ULL)))){
return c->usrR8_12[cast<uint32_t>((i - cast<uint32_t>(8ULL)))];
}
else if (((i == cast<uint32_t>(13ULL)) && (!arm_CPU_inUserBank(c)))){
return c->bankR13[cast<int64_t>(0ULL)];
}
else if (((i == cast<uint32_t>(14ULL)) && (!arm_CPU_inUserBank(c)))){
return c->bankR14[cast<int64_t>(0ULL)];
}
}
tmp29:;
return c->R[i];
}
}
// tools/cpu/arm/armexec.go:570:1
void arm_CPU_setUserReg(arm_CPU* c,uint32_t i,uint32_t v){
{
{
if ((((i >= cast<uint32_t>(8ULL)) && (i <= cast<uint32_t>(12ULL))) && (c->Mode == cast<uint32_t>(17ULL)))){
c->usrR8_12[cast<uint32_t>((i - cast<uint32_t>(8ULL)))] = v;
}
else if (((i == cast<uint32_t>(13ULL)) && (!arm_CPU_inUserBank(c)))){
c->bankR13[cast<int64_t>(0ULL)] = v;
}
else if (((i == cast<uint32_t>(14ULL)) && (!arm_CPU_inUserBank(c)))){
c->bankR14[cast<int64_t>(0ULL)] = v;
}
else {
c->R[i] = v;
}
}
tmp30:;
}
}
// tools/cpu/arm/cpu.go:128:1
void arm_CPU_Reset(arm_CPU* c){
{
c->R = std::array<uint32_t,16>{};
auto tmp31 = std::make_tuple(false,false,false,false,false);
c->N = std::get<0>(tmp31);
c->Z = std::get<1>(tmp31);
c->C = std::get<2>(tmp31);
c->V = std::get<3>(tmp31);
c->Q = std::get<4>(tmp31);
c->Thumb = false;
auto tmp32 = std::make_tuple(true,true);
c->IRQDisable = std::get<0>(tmp32);
c->FIQDisable = std::get<1>(tmp32);
c->Mode = cast<uint32_t>(19ULL);
auto tmp33 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp33);
c->HaltReason = std::get<1>(tmp33);
}
}
// tools/cpu/arm/cpu.go:146:1
uint32_t arm_CPU_PC(arm_CPU* c){
{
return c->R[cast<int64_t>(15ULL)];
}
}
// tools/cpu/arm/cpu.go:151:1
uint32_t arm_CPU_CPSR(arm_CPU* c){
{
uint32_t v={};
if (c->N) {
v |= cast<uint32_t>(2147483648ULL);
}
if (c->Z) {
v |= cast<uint32_t>(1073741824ULL);
}
if (c->C) {
v |= cast<uint32_t>(536870912ULL);
}
if (c->V) {
v |= cast<uint32_t>(268435456ULL);
}
if (c->Q) {
v |= cast<uint32_t>(134217728ULL);
}
v |= shl<uint32_t>((cast<uint32_t>((c->GE & cast<uint32_t>(15ULL)))),cast<int64_t>(16ULL));
if (c->BigEndian) {
v |= cast<uint32_t>(512ULL);
}
if (c->IRQDisable) {
v |= cast<uint32_t>(128ULL);
}
if (c->FIQDisable) {
v |= cast<uint32_t>(64ULL);
}
if (c->Thumb) {
v |= cast<uint32_t>(32ULL);
}
return cast<uint32_t>((v | (cast<uint32_t>((c->Mode & cast<uint32_t>(31ULL))))));
}
}
// tools/cpu/arm/cpu.go:185:1
void arm_CPU_SetCPSR(arm_CPU* c,uint32_t v){
{
c->N = (cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
c->Z = (cast<uint32_t>((v & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL));
c->C = (cast<uint32_t>((v & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
c->V = (cast<uint32_t>((v & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
c->Q = (cast<uint32_t>((v & cast<uint32_t>(134217728ULL))) != cast<uint32_t>(0ULL));
c->GE = cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
c->BigEndian = (cast<uint32_t>((v & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
c->IRQDisable = (cast<uint32_t>((v & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL));
c->FIQDisable = (cast<uint32_t>((v & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
c->Thumb = (cast<uint32_t>((v & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL));
arm_CPU_switchMode(c,cast<uint32_t>((v & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/arm/cpu.go:200:1
int64_t arm_modeIndex(uint32_t mode){
{
{
switch(mode){
case cast<uint32_t>(17ULL):{
return cast<int64_t>(1ULL);
break;}
case cast<uint32_t>(18ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(19ULL):{
return cast<int64_t>(3ULL);
break;}
case cast<uint32_t>(23ULL):{
return cast<int64_t>(4ULL);
break;}
case cast<uint32_t>(27ULL):{
return cast<int64_t>(5ULL);
break;}
default:{
return cast<int64_t>(0ULL);
break;}
}}
}
}
// tools/cpu/arm/cpu.go:218:1
void arm_CPU_switchMode(arm_CPU* c,uint32_t mode){
{
mode &= cast<uint32_t>(31ULL);
if ((mode == c->Mode)) {
return ;
}
auto tmp34 = std::make_tuple(arm_modeIndex(c->Mode),arm_modeIndex(mode));
int64_t from = std::get<0>(tmp34);
int64_t to = std::get<1>(tmp34);
c->bankR13[from] = c->R[cast<int64_t>(13ULL)];
c->bankR14[from] = c->R[cast<int64_t>(14ULL)];
if ((c->Mode == cast<uint32_t>(17ULL))) {
gcopy(sub(c->fiqR8_12,0,len(c->fiqR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
else {
gcopy(sub(c->usrR8_12,0,len(c->usrR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
c->R[cast<int64_t>(13ULL)] = c->bankR13[to];
c->R[cast<int64_t>(14ULL)] = c->bankR14[to];
if ((mode == cast<uint32_t>(17ULL))) {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->fiqR8_12,0,len(c->fiqR8_12)));
}
else {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->usrR8_12,0,len(c->usrR8_12)));
}
c->Mode = mode;
}
}
// tools/cpu/arm/cpu.go:245:1
uint32_t arm_CPU_SPSR(arm_CPU* c){
{
return c->bankSPSR[arm_modeIndex(c->Mode)];
}
}
// tools/cpu/arm/cpu.go:248:1
void arm_CPU_SetSPSR(arm_CPU* c,uint32_t v){
{
c->bankSPSR[arm_modeIndex(c->Mode)] = v;
}
}
// tools/cpu/arm/cpu.go:252:1
uint8_t arm_CPU_read8(arm_CPU* c,uint32_t a){
{
return dc_armBus_Read(*(c->bus),a);
}
}
// tools/cpu/arm/cpu.go:253:1
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v){
{
dc_armBus_Write(*(c->bus),a,v);
}
}
// tools/cpu/arm/cpu.go:269:1
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a){
{
if (arm_Variant_isV6(c->Arch)) {
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
return dc_armBus_Read32(*(c->wide),a);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dc_armBus_Read(*(c->bus),a)) | shl<uint32_t>(cast<uint32_t>(dc_armBus_Read(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dc_armBus_Read(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dc_armBus_Read(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
uint32_t aligned = (a & ~(cast<uint32_t>(3ULL)));
uint32_t v = arm_CPU_read32aligned(c,aligned);
{
uint32_t r = cast<uint32_t>(((cast<uint32_t>((a & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
if ((r != cast<uint32_t>(0ULL))) {
v = arm_ror32(v,r);
}
}
return v;
}
}
// tools/cpu/arm/cpu.go:331:1
void arm_CPU_write32(arm_CPU* c,uint32_t a,uint32_t v){
{
if ((!arm_Variant_isV6(c->Arch))) {
a &= ~(cast<uint32_t>(3ULL));
}
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
dc_armBus_Write32(*(c->wide),a,v);
return ;
}
dc_armBus_Write(*(c->bus),a,cast<uint8_t>(v));
dc_armBus_Write(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
dc_armBus_Write(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
dc_armBus_Write(*(c->bus),cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/cpu/arm/cpu.go:349:1
uint32_t arm_CPU_reg(arm_CPU* c,uint32_t i){
{
if ((i == cast<uint32_t>(15ULL))) {
if (c->Thumb) {
return cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<uint32_t>((c->cur + cast<uint32_t>(8ULL)));
}
return c->R[i];
}
}
// tools/cpu/arm/cpu.go:360:1
void arm_CPU_setReg(arm_CPU* c,uint32_t i,uint32_t v){
{
c->R[i] = v;
if ((i == cast<uint32_t>(15ULL))) {
c->branched = true;
}
}
}
// tools/cpu/arm/cpu.go:369:1
void arm_CPU_setNZ(arm_CPU* c,uint32_t v){
{
c->Z = (v == cast<uint32_t>(0ULL));
c->N = (cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/arm/cpu.go:375:1
uint32_t arm_CPU_add(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
uint64_t r = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(a) + cast<uint64_t>(b))) + cast<uint64_t>(cin)));
uint32_t res = cast<uint32_t>(r);
arm_CPU_setNZ(c,res);
c->C = (r > cast<uint64_t>(4294967295ULL));
c->V = (cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ res))) & (cast<uint32_t>((b ^ res))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
return res;
}
}
// tools/cpu/arm/cpu.go:385:1
uint32_t arm_CPU_sub_(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
return arm_CPU_add(c,a,cast<uint32_t>(~b),cin);
}
}
// tools/cpu/arm/cpu.go:388:1
bool arm_CPU_cond(arm_CPU* c,int64_t cc){
{
{
switch(cc){
case cast<int64_t>(0ULL):{
return c->Z;
break;}
case cast<int64_t>(1ULL):{
return (!c->Z);
break;}
case cast<int64_t>(2ULL):{
return c->C;
break;}
case cast<int64_t>(3ULL):{
return (!c->C);
break;}
case cast<int64_t>(4ULL):{
return c->N;
break;}
case cast<int64_t>(5ULL):{
return (!c->N);
break;}
case cast<int64_t>(6ULL):{
return c->V;
break;}
case cast<int64_t>(7ULL):{
return (!c->V);
break;}
case cast<int64_t>(8ULL):{
return (c->C && (!c->Z));
break;}
case cast<int64_t>(9ULL):{
return ((!c->C) || c->Z);
break;}
case cast<int64_t>(10ULL):{
return (c->N == c->V);
break;}
case cast<int64_t>(11ULL):{
return (c->N != c->V);
break;}
case cast<int64_t>(12ULL):{
return ((!c->Z) && (c->N == c->V));
break;}
case cast<int64_t>(13ULL):{
return (c->Z || (c->N != c->V));
break;}
default:{
return true;
break;}
}}
}
}
// tools/cpu/arm/cpu.go:429:1
std::tuple<uint32_t,uint32_t> arm_CPU_shift(arm_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin){
{
if (regForm) {
amt &= cast<uint32_t>(255ULL);
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cin};
}
}
{
switch(typ){
case cast<uint32_t>(0ULL):{
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {shl<uint32_t>(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((cast<uint32_t>(32ULL) - amt))))) & cast<uint32_t>(1ULL)))};
}
else if ((amt == cast<uint32_t>(32ULL))){
return {cast<uint32_t>(0ULL),cast<uint32_t>((val & cast<uint32_t>(1ULL)))};
}
else {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp35:;
break;}
case cast<uint32_t>(1ULL):{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
amt = cast<uint32_t>(32ULL);
}
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {shr<uint32_t>(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
}
else if ((amt == cast<uint32_t>(32ULL))){
return {cast<uint32_t>(0ULL),cast<uint32_t>(((shr<uint32_t>(val,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
else {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp36:;
break;}
case cast<uint32_t>(2ULL):{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
amt = cast<uint32_t>(32ULL);
}
int32_t sv = cast<int32_t>(val);
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {cast<uint32_t>(shr<int32_t>(sv,amt)),cast<uint32_t>((cast<uint32_t>(shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
}
else {
if ((cast<uint32_t>((val & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(4294967295ULL),cast<uint32_t>(1ULL)};
}
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp37:;
break;}
default:{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
return {cast<uint32_t>((shl<uint32_t>(cin,cast<int64_t>(31ULL)) | shr<uint32_t>(val,cast<int64_t>(1ULL)))),cast<uint32_t>((val & cast<uint32_t>(1ULL)))};
}
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cin};
}
amt &= cast<uint32_t>(31ULL);
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cast<uint32_t>(((shr<uint32_t>(val,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
return {arm_ror32(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
break;}
}}
}
}
// tools/cpu/arm/cpu.go:497:1
void arm_CPU_ClearExclusive(arm_CPU* c){
{
c->exclValid = false;
}
}
// tools/cpu/arm/cpu.go:519:1
arm_Banks arm_CPU_SaveBanks(arm_CPU* c){
{
arm_Banks b = arm_Banks{c->bankR13,c->bankR14,c->bankSPSR,c->fiqR8_12,c->usrR8_12};
int64_t i = arm_modeIndex(c->Mode);
auto tmp38 = std::make_tuple(c->R[cast<int64_t>(13ULL)],c->R[cast<int64_t>(14ULL)]);
b.R13[i] = std::get<0>(tmp38);
b.R14[i] = std::get<1>(tmp38);
if ((c->Mode == cast<uint32_t>(17ULL))) {
gcopy(sub(b.FIQR8_12,0,len(b.FIQR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
else {
gcopy(sub(b.USRR8_12,0,len(b.USRR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
return b;
}
}
// tools/cpu/arm/cpu.go:534:1
void arm_CPU_RestoreBanks(arm_CPU* c,arm_Banks b){
{
auto tmp39 = std::make_tuple(b.R13,b.R14,b.SPSR);
c->bankR13 = std::get<0>(tmp39);
c->bankR14 = std::get<1>(tmp39);
c->bankSPSR = std::get<2>(tmp39);
auto tmp40 = std::make_tuple(b.FIQR8_12,b.USRR8_12);
c->fiqR8_12 = std::get<0>(tmp40);
c->usrR8_12 = std::get<1>(tmp40);
}
}
// tools/cpu/arm/decode_v6.go:10:1
std::tuple<arm_Inst,bool> arm_decodeARMv6(uint32_t w,uint32_t addr,arm_Inst in){
{
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
if ((arm_isVFP(w) && (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL)) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL)))))) {
return {arm_decodeVFP(w,in),true};
}
if ((cond == cast<int64_t>(15ULL))) {
{
auto tmp41 = arm_decodeV6Uncond(w,in);
arm_Inst out = std::get<0>(tmp41);
bool ok = std::get<1>(tmp41);
if (ok) {
return {out,true};
}
}
return {in,false};
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return {arm_decodeSync(w,in),true};
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(4ULL)))) {
in.Mnem = (std::string("UMAAL",5) + arm_cn(cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
return {in,true};
}
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return {arm_decodeMedia(w,in),true};
}
return {in,false};
}
}
// tools/cpu/arm/decode_v6.go:52:1
arm_Inst arm_decodeSync(uint32_t w,arm_Inst in){
{
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t sz = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)));
std::string suffix = std::array<std::string,4>{std::string("",0),std::string("D",1),std::string("B",1),std::string("H",1)}[sz];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
if (load) {
in.Mnem = ((std::string("LDREX",5) + suffix) + arm_cn(in.Cond));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((sz == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[rt],arm_regName[cast<uint32_t>(((cast<uint32_t>((rt + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))],rn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s]",11),in.Mnem,arm_regName[rt],rn);
}
}
else {
in.Mnem = ((std::string("STREX",5) + suffix) + arm_cn(in.Cond));
uint32_t rt = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((sz == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, [%s]",19),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[rt],arm_regName[cast<uint32_t>(((cast<uint32_t>((rt + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))],rn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[rt],rn);
}
}
return in;
}
}
// tools/cpu/arm/decode_v6.go:82:1
arm_Inst arm_decodeMedia(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
{
if (((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((op1 & cast<uint32_t>(7ULL))) != cast<uint32_t>(0ULL)))){
{
std::string name = arm_parallelName(cast<uint32_t>((op1 & cast<uint32_t>(7ULL))),op2);
if ((name != std::string("",0))) {
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
return in;
}
}
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(1ULL))){
return arm_decodeMediaPack(w,in);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(2ULL))){
return arm_decodeMediaMul(w,in);
}
else if ((op1 == cast<uint32_t>(24ULL))){
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((ra == cast<uint32_t>(15ULL))) {
in.Mnem = (std::string("USAD8",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rn,rm,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
}
else {
in.Mnem = (std::string("USADA8",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rn,rm,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))],arm_regName[ra]);
}
return in;
}
}
tmp42:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:116:1
std::string arm_parallelName(uint32_t class_,uint32_t op2){
{
std::string prefix = get(Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("S",1)},{cast<uint32_t>(2ULL),std::string("Q",1)},{cast<uint32_t>(3ULL),std::string("SH",2)},{cast<uint32_t>(5ULL),std::string("U",1)},{cast<uint32_t>(6ULL),std::string("UQ",2)},{cast<uint32_t>(7ULL),std::string("UH",2)}},class_);
std::string op = get(Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("ADD16",5)},{cast<uint32_t>(1ULL),std::string("ASX",3)},{cast<uint32_t>(2ULL),std::string("SAX",3)},{cast<uint32_t>(3ULL),std::string("SUB16",5)},{cast<uint32_t>(4ULL),std::string("ADD8",4)},{cast<uint32_t>(7ULL),std::string("SUB8",4)}},op2);
if (((prefix == std::string("",0)) || (op == std::string("",0)))) {
return std::string("",0);
}
return (prefix + op);
}
}
// tools/cpu/arm/decode_v6.go:127:1
arm_Inst arm_decodeMediaPack(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
{
if (((op1 == cast<uint32_t>(8ULL)) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (std::string("PKHBT",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, LSL #%d",22),in.Mnem,rd,rn,rm,imm);
}
else {
in.Mnem = (std::string("PKHTB",5) + arm_cn(in.Cond));
if ((imm == cast<uint32_t>(0ULL))) {
imm = cast<uint32_t>(32ULL);
}
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, ASR #%d",22),in.Mnem,rd,rn,rm,imm);
}
return in;
}
else if (((((op1 == cast<uint32_t>(10ULL)) || (op1 == cast<uint32_t>(11ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t sat = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))) + cast<uint32_t>(1ULL)));
return arm_satText(w,in,std::string("SSAT",4),sat);
}
else if (((((op1 == cast<uint32_t>(14ULL)) || (op1 == cast<uint32_t>(15ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t sat = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
return arm_satText(w,in,std::string("USAT",4),sat);
}
else if (((op1 == cast<uint32_t>(10ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("SSAT16",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))) + cast<uint32_t>(1ULL))),rm);
return in;
}
else if (((op1 == cast<uint32_t>(14ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("USAT16",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),rm);
return in;
}
else if ((op2 == cast<uint32_t>(3ULL))){
{
std::string name = arm_extendName(op1);
if ((name != std::string("",0))) {
uint32_t rot = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL))) * cast<uint32_t>(8ULL)));
bool acc = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))) != cast<uint32_t>(15ULL));
in.Mnem = (name + arm_cn(in.Cond));
if (acc) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
}
if ((rot != cast<uint32_t>(0ULL))) {
in.Text += go_fmt_Sprintf(std::string(", ROR #%d",9),rot);
}
return in;
}
}
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("REV",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("REV16",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(15ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("REVSH",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(8ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("SEL",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
return in;
}
}
tmp43:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:199:1
arm_Inst arm_satText(uint32_t w,arm_Inst in,std::string name,uint32_t sat){
{
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
std::string sh = std::string("LSL",3);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
sh = std::string("ASR",3);
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
}
in.Mnem = (name + arm_cn(in.Cond));
if ((amt == cast<uint32_t>(0ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,sat,rm);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s, %s #%d",22),in.Mnem,rd,sat,rm,sh,amt);
}
return in;
}
}
// tools/cpu/arm/decode_v6.go:221:1
std::string arm_extendName(uint32_t op1){
{
{
switch(op1){
case cast<uint32_t>(8ULL):{
return std::string("SXTB16",6);
break;}
case cast<uint32_t>(10ULL):{
return std::string("SXTB",4);
break;}
case cast<uint32_t>(11ULL):{
return std::string("SXTH",4);
break;}
case cast<uint32_t>(12ULL):{
return std::string("UXTB16",6);
break;}
case cast<uint32_t>(14ULL):{
return std::string("UXTB",4);
break;}
case cast<uint32_t>(15ULL):{
return std::string("UXTH",4);
break;}
}}
return std::string("",0);
}
}
// tools/cpu/arm/decode_v6.go:241:1
arm_Inst arm_decodeMediaMul(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
std::string rm = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
std::string x = std::string("X",1);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
x = std::string("",0);
}
{
switch(op1){
case cast<uint32_t>(16ULL):{
std::string base={};
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
base = std::string("SMLAD",5);
break;}
case cast<uint32_t>(1ULL):{
base = std::string("SMLSD",5);
break;}
default:{
return arm_undef(w,in);
break;}
}}
if ((ra == cast<uint32_t>(15ULL))) {
base = get(Map<std::string,std::string>{{std::string("SMLAD",5),std::string("SMUAD",5)},{std::string("SMLSD",5),std::string("SMUSD",5)}},base);
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
}
return in;
break;}
case cast<uint32_t>(20ULL):{
std::string base = std::string("SMLALD",6);
if ((shr<uint32_t>(op2,cast<int64_t>(1ULL)) == cast<uint32_t>(1ULL))) {
base = std::string("SMLSLD",6);
}
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[ra],rd,rn,rm);
return in;
break;}
case cast<uint32_t>(21ULL):{
std::string r = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
r = std::string("R",1);
}
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
if ((ra == cast<uint32_t>(15ULL))) {
in.Mnem = ((std::string("SMMUL",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Mnem = ((std::string("SMMLA",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
}
return in;
break;}
case cast<uint32_t>(3ULL):{
in.Mnem = ((std::string("SMMLS",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
return in;
break;}
}}
break;}
}}
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:308:1
std::tuple<arm_Inst,bool> arm_decodeV6Uncond(uint32_t w,arm_Inst in){
{
{
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(16ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imod = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(3ULL)));
in.Mnem = std::string("CPS",3);
{
switch(imod){
case cast<uint32_t>(2ULL):{
in.Mnem = std::string("CPSIE",5);
break;}
case cast<uint32_t>(3ULL):{
in.Mnem = std::string("CPSID",5);
break;}
}}
std::string flags = std::string("",0);
{auto&& tmp45 = Slice<Anon3>{Anon3{cast<uint32_t>(8ULL),std::string("a",1)},Anon3{cast<uint32_t>(7ULL),std::string("i",1)},Anon3{cast<uint32_t>(6ULL),std::string("f",1)}};
for(int64_t tmp46=0;tmp46<len(tmp45);++tmp46){
auto ch=tmp45[tmp46];if ((cast<uint32_t>(((shr<uint32_t>(w,ch.bit)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
flags += ch.name;
}
}}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, #0x%X",12),in.Mnem,flags,cast<uint32_t>((w & cast<uint32_t>(31ULL))));
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,flags);
}
return {in,true};
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))) == cast<uint32_t>(61697ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
std::string e = std::string("LE",2);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
e = std::string("BE",2);
}
in.Mnem = std::string("SETEND",6);
in.Text = (std::string("SETEND ",7) + e);
return {in,true};
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(87ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)))){
in.Mnem = std::string("CLREX",5);
in.Text = std::string("CLREX",5);
return {in,true};
}
}
tmp44:;
return {in,false};
}
}
// tools/cpu/arm/exec_v6.go:10:1
bool arm_CPU_execARMv6(arm_CPU* c,uint32_t w){
{
if ((arm_isVFP(w) && (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL)) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL)))))) {
return arm_CPU_execVFP(c,w);
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
arm_CPU_execSync(c,w);
return true;
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(4ULL)))) {
arm_CPU_execUMAAL(c,w);
return true;
}
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return arm_CPU_execMedia(c,w);
}
return false;
}
}
// tools/cpu/arm/exec_v6.go:34:1
bool arm_CPU_execUncondARMv6(arm_CPU* c,uint32_t w){
{
{
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(16ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_switchMode(c,cast<uint32_t>((w & cast<uint32_t>(31ULL))));
}
uint32_t imod = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(3ULL)));
if (((imod == cast<uint32_t>(2ULL)) || (imod == cast<uint32_t>(3ULL)))) {
bool disable = (imod == cast<uint32_t>(3ULL));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->IRQDisable = disable;
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->FIQDisable = disable;
}
}
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))) == cast<uint32_t>(61697ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
c->BigEndian = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(87ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)))){
c->exclValid = false;
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
}
tmp47:;
return false;
}
}
// tools/cpu/arm/exec_v6.go:66:1
void arm_CPU_execSync(arm_CPU* c,uint32_t w){
{
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t sz = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)));
uint32_t addr = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
if (load) {
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
{
switch(sz){
case cast<uint32_t>(0ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read32(c,addr));
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read32(c,addr));
arm_CPU_setReg(c,cast<uint32_t>((rt + cast<uint32_t>(1ULL))),arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL)))));
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_setReg(c,rt,cast<uint32_t>(arm_CPU_read8(c,addr)));
break;}
case cast<uint32_t>(3ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read16(c,addr));
break;}
}}
c->exclValid = true;
c->exclAddr = addr;
return ;
}
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rt = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
if (((!c->exclValid) || (c->exclAddr != addr))) {
arm_CPU_setReg(c,rd,cast<uint32_t>(1ULL));
return ;
}
{
switch(sz){
case cast<uint32_t>(0ULL):{
arm_CPU_write32(c,addr,rt);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_write32(c,addr,rt);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),arm_CPU_reg(c,cast<uint32_t>(((cast<uint32_t>((w & cast<uint32_t>(15ULL)))) + cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_write8(c,addr,cast<uint8_t>(rt));
break;}
case cast<uint32_t>(3ULL):{
arm_CPU_write16(c,addr,rt);
break;}
}}
arm_CPU_setReg(c,rd,cast<uint32_t>(0ULL));
c->exclValid = false;
}
}
// tools/cpu/arm/exec_v6.go:116:1
void arm_CPU_execUMAAL(arm_CPU* c,uint32_t w){
{
uint32_t rdHi = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rdLo = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint64_t res = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(rn) * cast<uint64_t>(rm))) + cast<uint64_t>(arm_CPU_reg(c,rdHi)))) + cast<uint64_t>(arm_CPU_reg(c,rdLo))));
arm_CPU_setReg(c,rdLo,cast<uint32_t>(res));
arm_CPU_setReg(c,rdHi,cast<uint32_t>(shr<uint64_t>(res,cast<int64_t>(32ULL))));
}
}
// tools/cpu/arm/exec_v6.go:129:1
bool arm_CPU_execMedia(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
{
if (((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((op1 & cast<uint32_t>(7ULL))) != cast<uint32_t>(0ULL)))){
return arm_CPU_execParallel(c,w);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(1ULL))){
return arm_CPU_execMediaPack(c,w);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(2ULL))){
return arm_CPU_execMediaMul(c,w);
}
else if ((op1 == cast<uint32_t>(24ULL))){
arm_CPU_execUSAD(c,w);
return true;
}
}
tmp48:;
arm_CPU_Halt(c,std::string("unimplemented media instruction 0x%08X at 0x%08X",48),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:159:1
bool arm_CPU_execParallel(arm_CPU* c,uint32_t w){
{
uint32_t class_ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(7ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
bool signed_={};
arm_parMode mode={};
{
switch(class_){
case cast<uint32_t>(1ULL):{
auto tmp49 = std::make_tuple(true,cast<arm_parMode>(0ULL));
signed_ = std::get<0>(tmp49);
mode = std::get<1>(tmp49);
break;}
case cast<uint32_t>(2ULL):{
auto tmp50 = std::make_tuple(true,cast<arm_parMode>(1ULL));
signed_ = std::get<0>(tmp50);
mode = std::get<1>(tmp50);
break;}
case cast<uint32_t>(3ULL):{
auto tmp51 = std::make_tuple(true,cast<arm_parMode>(2ULL));
signed_ = std::get<0>(tmp51);
mode = std::get<1>(tmp51);
break;}
case cast<uint32_t>(5ULL):{
auto tmp52 = std::make_tuple(false,cast<arm_parMode>(0ULL));
signed_ = std::get<0>(tmp52);
mode = std::get<1>(tmp52);
break;}
case cast<uint32_t>(6ULL):{
auto tmp53 = std::make_tuple(false,cast<arm_parMode>(1ULL));
signed_ = std::get<0>(tmp53);
mode = std::get<1>(tmp53);
break;}
case cast<uint32_t>(7ULL):{
auto tmp54 = std::make_tuple(false,cast<arm_parMode>(2ULL));
signed_ = std::get<0>(tmp54);
mode = std::get<1>(tmp54);
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented parallel arithmetic 0x%08X (class %d) at 0x%08X",61),w,class_,c->cur);
return true;
break;}
}}
auto half = [&](uint32_t v,uint64_t h)->uint32_t{
return cast<uint32_t>(((shr<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(16ULL) * h))))) & cast<uint32_t>(65535ULL)));
}
;
uint32_t res={};
uint32_t ge={};
{
switch(op2){
case cast<uint32_t>(0ULL):case cast<uint32_t>(3ULL):{
bool sub_ = (op2 == cast<uint32_t>(3ULL));
{uint64_t h = cast<uint64_t>(0ULL);for (;(h < cast<uint64_t>(2ULL));h++){
auto tmp55 = arm_parLane(half(rn,h),half(rm,h),cast<uint64_t>(16ULL),signed_,mode,sub_);
uint32_t v = std::get<0>(tmp55);
bool g = std::get<1>(tmp55);
res |= shl<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(16ULL) * h))));
if (g) {
ge |= shl<uint32_t>(cast<uint32_t>(3ULL),(cast<uint64_t>((cast<uint64_t>(2ULL) * h))));
}
}
}break;}
case cast<uint32_t>(1ULL):{
auto tmp56 = arm_parLane(half(rn,cast<uint64_t>(0ULL)),half(rm,cast<uint64_t>(1ULL)),cast<uint64_t>(16ULL),signed_,mode,true);
uint32_t lo = std::get<0>(tmp56);
bool glo = std::get<1>(tmp56);
auto tmp57 = arm_parLane(half(rn,cast<uint64_t>(1ULL)),half(rm,cast<uint64_t>(0ULL)),cast<uint64_t>(16ULL),signed_,mode,false);
uint32_t hi = std::get<0>(tmp57);
bool ghi = std::get<1>(tmp57);
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
if (glo) {
ge |= cast<uint32_t>(3ULL);
}
if (ghi) {
ge |= cast<uint32_t>(12ULL);
}
break;}
case cast<uint32_t>(2ULL):{
auto tmp58 = arm_parLane(half(rn,cast<uint64_t>(0ULL)),half(rm,cast<uint64_t>(1ULL)),cast<uint64_t>(16ULL),signed_,mode,false);
uint32_t lo = std::get<0>(tmp58);
bool glo = std::get<1>(tmp58);
auto tmp59 = arm_parLane(half(rn,cast<uint64_t>(1ULL)),half(rm,cast<uint64_t>(0ULL)),cast<uint64_t>(16ULL),signed_,mode,true);
uint32_t hi = std::get<0>(tmp59);
bool ghi = std::get<1>(tmp59);
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
if (glo) {
ge |= cast<uint32_t>(3ULL);
}
if (ghi) {
ge |= cast<uint32_t>(12ULL);
}
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(7ULL):{
bool sub_ = (op2 == cast<uint32_t>(7ULL));
{uint64_t b = cast<uint64_t>(0ULL);for (;(b < cast<uint64_t>(4ULL));b++){
uint32_t av = cast<uint32_t>(((shr<uint32_t>(rn,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))))) & cast<uint32_t>(255ULL)));
uint32_t bv = cast<uint32_t>(((shr<uint32_t>(rm,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))))) & cast<uint32_t>(255ULL)));
auto tmp60 = arm_parLane(av,bv,cast<uint64_t>(8ULL),signed_,mode,sub_);
uint32_t v = std::get<0>(tmp60);
bool g = std::get<1>(tmp60);
res |= shl<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))));
if (g) {
ge |= shl<uint32_t>(cast<uint32_t>(1ULL),b);
}
}
}break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented parallel op2=%d at 0x%08X",39),op2,c->cur);
return true;
break;}
}}
arm_CPU_setReg(c,rd,res);
if ((mode == cast<arm_parMode>(0ULL))) {
c->GE = ge;
}
return true;
}
}
// tools/cpu/arm/exec_v6.go:260:1
std::tuple<uint32_t,bool> arm_parLane(uint32_t av,uint32_t bv,uint64_t width,bool signed_,arm_parMode mode,bool sub_){
{
uint32_t mask = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(1ULL),width) - cast<uint32_t>(1ULL)));
int32_t val={};
if (signed_) {
auto tmp61 = std::make_tuple(arm_signExtendLane(av,width),arm_signExtendLane(bv,width));
int32_t a = std::get<0>(tmp61);
int32_t b = std::get<1>(tmp61);
if (sub_) {
val = cast<int32_t>((a - b));
}
else {
val = cast<int32_t>((a + b));
}
}
else if (sub_) {
val = cast<int32_t>((cast<int32_t>(av) - cast<int32_t>(bv)));
}
else {
val = cast<int32_t>((cast<int32_t>(av) + cast<int32_t>(bv)));
}
bool ge={};
{
if (signed_){
ge = (val >= cast<int32_t>(0ULL));
}
else if (sub_){
ge = (av >= bv);
}
else {
ge = (val > cast<int32_t>(mask));
}
}
tmp62:;
{
switch(mode){
case cast<arm_parMode>(1ULL):{
if (signed_) {
int32_t hi = cast<int32_t>((shl<int32_t>(cast<int32_t>(1ULL),(cast<uint64_t>((width - cast<uint64_t>(1ULL))))) - cast<int32_t>(1ULL)));
int32_t lo = cast<int32_t>(-(shl<int32_t>(cast<int32_t>(1ULL),(cast<uint64_t>((width - cast<uint64_t>(1ULL)))))));
if ((val > hi)) {
val = hi;
}
else if ((val < lo)) {
val = lo;
}
}
else {
if ((val < cast<int32_t>(0ULL))) {
val = cast<int32_t>(0ULL);
}
else if ((val > cast<int32_t>(mask))) {
val = cast<int32_t>(mask);
}
}
break;}
case cast<arm_parMode>(2ULL):{
val = shr<int32_t>(val,cast<int64_t>(1ULL));
break;}
}}
return {cast<uint32_t>((cast<uint32_t>(val) & mask)),ge};
}
}
// tools/cpu/arm/exec_v6.go:312:1
int32_t arm_int16v(uint32_t v){
{
return cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)));
}
}
// tools/cpu/arm/exec_v6.go:315:1
int32_t arm_signExtendLane(uint32_t v,uint64_t width){
{
uint64_t sh = cast<uint64_t>((cast<uint64_t>(32ULL) - width));
return shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,sh)),sh);
}
}
// tools/cpu/arm/exec_v6.go:320:1
bool arm_CPU_execMediaPack(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
{
if (((op1 == cast<uint32_t>(8ULL)) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
uint32_t res={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
res = cast<uint32_t>(((cast<uint32_t>((rnv & cast<uint32_t>(65535ULL)))) | (cast<uint32_t>(((shl<uint32_t>(rmv,imm)) & cast<uint32_t>(4294901760ULL))))));
}
else {
if ((imm == cast<uint32_t>(0ULL))) {
imm = cast<uint32_t>(32ULL);
}
res = cast<uint32_t>(((cast<uint32_t>((rnv & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((cast<uint32_t>(shr<int32_t>(cast<int32_t>(rmv),imm)) & cast<uint32_t>(65535ULL))))));
}
arm_CPU_setReg(c,rd,res);
return true;
}
else if (((((op1 == cast<uint32_t>(10ULL)) || (op1 == cast<uint32_t>(11ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_setReg(c,rd,arm_CPU_doSat(c,w,true));
return true;
}
else if (((((op1 == cast<uint32_t>(14ULL)) || (op1 == cast<uint32_t>(15ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_setReg(c,rd,arm_CPU_doSat(c,w,false));
return true;
}
else if ((op2 == cast<uint32_t>(3ULL))){
if (arm_CPU_execExtend(c,w)) {
return true;
}
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(1ULL)))){
arm_CPU_setReg(c,rd,go_bits_ReverseBytes32(rmv));
return true;
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(5ULL)))){
arm_CPU_setReg(c,rd,cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(16711935ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(4278255360ULL)))),cast<int64_t>(8ULL)))));
return true;
}
else if (((op1 == cast<uint32_t>(15ULL)) && (op2 == cast<uint32_t>(5ULL)))){
uint32_t v = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(255ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(65280ULL)))),cast<int64_t>(8ULL))));
arm_CPU_setReg(c,rd,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)))));
return true;
}
else if (((op1 == cast<uint32_t>(8ULL)) && (op2 == cast<uint32_t>(5ULL)))){
uint32_t res={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint64_t sh = cast<uint64_t>(cast<int64_t>((i * cast<int64_t>(8ULL))));
if ((cast<uint32_t>((c->GE & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
res |= cast<uint32_t>((rnv & (shl<uint32_t>(cast<uint32_t>(255ULL),sh))));
}
else {
res |= cast<uint32_t>((rmv & (shl<uint32_t>(cast<uint32_t>(255ULL),sh))));
}
}
}arm_CPU_setReg(c,rd,res);
return true;
}
}
tmp63:;
arm_CPU_Halt(c,std::string("unimplemented media pack 0x%08X at 0x%08X",41),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:381:1
uint32_t arm_CPU_doSat(arm_CPU* c,uint32_t w,bool signed_){
{
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
int32_t val = cast<int32_t>(rmv);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
val = shr<int32_t>(val,amt);
}
else if ((amt != cast<uint32_t>(0ULL))) {
val = cast<int32_t>(shl<uint32_t>(rmv,amt));
}
if (signed_) {
uint32_t n = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))) + cast<uint32_t>(1ULL)));
int32_t hi = cast<int32_t>((shl<int32_t>(cast<int32_t>(1ULL),(cast<uint32_t>((n - cast<uint32_t>(1ULL))))) - cast<int32_t>(1ULL)));
int32_t lo = shl<int32_t>(cast<int32_t>(-1ULL),(cast<uint32_t>((n - cast<uint32_t>(1ULL)))));
if ((val > hi)) {
c->Q = true;
return cast<uint32_t>(hi);
}
if ((val < lo)) {
c->Q = true;
return cast<uint32_t>(lo);
}
return cast<uint32_t>(val);
}
uint32_t n = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
int32_t hi = cast<int32_t>(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),n)) - cast<uint32_t>(1ULL))));
if ((val > hi)) {
c->Q = true;
return cast<uint32_t>(hi);
}
if ((val < cast<int32_t>(0ULL))) {
c->Q = true;
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>(val);
}
}
// tools/cpu/arm/exec_v6.go:422:1
bool arm_CPU_execExtend(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rot = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL))) * cast<uint32_t>(8ULL)));
uint32_t rotated = arm_ror32(rmv,rot);
bool acc = (rn != cast<uint32_t>(15ULL));
uint32_t rnv = arm_CPU_reg(c,rn);
uint32_t res={};
{
switch(op1){
case cast<uint32_t>(10ULL):{
res = cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(rotated))));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(11ULL):{
res = cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(rotated))));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(14ULL):{
res = cast<uint32_t>((rotated & cast<uint32_t>(255ULL)));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(15ULL):{
res = cast<uint32_t>((rotated & cast<uint32_t>(65535ULL)));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(8ULL):{
uint32_t lo = cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(rotated)))) & cast<uint32_t>(65535ULL)));
uint32_t hi = cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(shr<uint32_t>(rotated,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
if (acc) {
lo = cast<uint32_t>(((cast<uint32_t>((lo + rnv))) & cast<uint32_t>(65535ULL)));
hi = cast<uint32_t>(((cast<uint32_t>((hi + (shr<uint32_t>(rnv,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
}
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(12ULL):{
uint32_t lo = cast<uint32_t>((rotated & cast<uint32_t>(255ULL)));
uint32_t hi = cast<uint32_t>(((shr<uint32_t>(rotated,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
if (acc) {
lo = cast<uint32_t>(((cast<uint32_t>((lo + rnv))) & cast<uint32_t>(65535ULL)));
hi = cast<uint32_t>(((cast<uint32_t>((hi + (shr<uint32_t>(rnv,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
}
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
break;}
default:{
return false;
break;}
}}
arm_CPU_setReg(c,rd,res);
return true;
}
}
// tools/cpu/arm/exec_v6.go:479:1
bool arm_CPU_execMediaMul(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
bool swap = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t m = rmv;
if (swap) {
m = cast<uint32_t>(((shr<uint32_t>(rmv,cast<int64_t>(16ULL))) | (shl<uint32_t>(rmv,cast<int64_t>(16ULL)))));
}
{
switch(op1){
case cast<uint32_t>(16ULL):{
int32_t p0 = cast<int32_t>((arm_int16v(rnv) * arm_int16v(m)));
int32_t p1 = cast<int32_t>((arm_int16v(shr<uint32_t>(rnv,cast<int64_t>(16ULL))) * arm_int16v(shr<uint32_t>(m,cast<int64_t>(16ULL)))));
int32_t dual={};
if ((shr<uint32_t>(op2,cast<int64_t>(1ULL)) == cast<uint32_t>(0ULL))) {
dual = cast<int32_t>((p0 + p1));
}
else {
dual = cast<int32_t>((p0 - p1));
}
if ((ra != cast<uint32_t>(15ULL))) {
dual += cast<int32_t>(arm_CPU_reg(c,ra));
}
arm_CPU_setReg(c,rd,cast<uint32_t>(dual));
return true;
break;}
case cast<uint32_t>(21ULL):{
int64_t prod = cast<int64_t>((cast<int64_t>(cast<int32_t>(rnv)) * cast<int64_t>(cast<int32_t>(m))));
bool round = (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
if (round) {
prod += cast<int64_t>(2147483648ULL);
}
int32_t res={};
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
res = cast<int32_t>(shr<int64_t>(prod,cast<int64_t>(32ULL)));
if ((ra != cast<uint32_t>(15ULL))) {
res += cast<int32_t>(arm_CPU_reg(c,ra));
}
break;}
case cast<uint32_t>(3ULL):{
res = cast<int32_t>((cast<int32_t>(arm_CPU_reg(c,ra)) - cast<int32_t>(shr<int64_t>(prod,cast<int64_t>(32ULL)))));
break;}
default:{
return false;
break;}
}}
arm_CPU_setReg(c,rd,cast<uint32_t>(res));
return true;
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented media multiply 0x%08X at 0x%08X",45),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:534:1
void arm_CPU_execUSAD(arm_CPU* c,uint32_t w){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t sum={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint64_t sh = cast<uint64_t>(cast<int64_t>((i * cast<int64_t>(8ULL))));
int32_t d = cast<int32_t>((cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(rnv,sh)) & cast<uint32_t>(255ULL)))) - cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(rmv,sh)) & cast<uint32_t>(255ULL))))));
if ((d < cast<int32_t>(0ULL))) {
d = cast<int32_t>(-d);
}
sum += cast<uint32_t>(d);
}
}if ((ra != cast<uint32_t>(15ULL))) {
sum += arm_CPU_reg(c,ra);
}
arm_CPU_setReg(c,rd,sum);
}
}
// tools/cpu/arm/miscexec.go:7:1
void arm_CPU_execMisc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
arm_CPU_execMSR(c,w,true);
return ;
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("SMLAxy",6));
return ;
}
arm_CPU_execSignedMul(c,w);
return ;
}
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
uint32_t v = arm_CPU_CPSR(c);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = arm_CPU_SPSR(c);
}
arm_CPU_setReg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),v);
}
else {
arm_CPU_execMSR(c,w,false);
}
break;}
case cast<uint32_t>(1ULL):{
{
switch(op){
case cast<uint32_t>(9ULL):{
arm_CPU_bxTo(c,arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))));
break;}
case cast<uint32_t>(11ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("CLZ",3));
return ;
}
arm_CPU_setReg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(go_bits_LeadingZeros32(arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))))));
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
break;}
}}
break;}
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(9ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("BLX (register)",14));
return ;
}
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
arm_CPU_bxTo(c,arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))));
}
else {
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
}
break;}
case cast<uint32_t>(5ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("QADD/QSUB",9));
return ;
}
arm_CPU_execSaturating(c,w);
break;}
case cast<uint32_t>(7ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("BKPT",4));
return ;
}
arm_CPU_Halt(c,std::string("BKPT #0x%X at 0x%08X",20),cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(65520ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL)))))),c->cur);
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:78:1
void arm_CPU_undefV4T(arm_CPU* c,uint32_t w,std::string name){
{
arm_CPU_Halt(c,std::string("%s is undefined on ARMv4T: 0x%08X at 0x%08X",43),name,w,c->cur);
}
}
// tools/cpu/arm/miscexec.go:82:1
void arm_CPU_execMSR(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t val={};
if (immForm) {
val = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
}
else {
val = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
}
uint32_t fields = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t mask={};
if ((cast<uint32_t>((fields & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(255ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(65280ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(16711680ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(4278190080ULL);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_SetSPSR(c,cast<uint32_t>(((arm_CPU_SPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
return ;
}
if ((c->Mode == cast<uint32_t>(16ULL))) {
mask &= cast<uint32_t>(4278190080ULL);
}
arm_CPU_SetCPSR(c,cast<uint32_t>(((arm_CPU_CPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
}
}
// tools/cpu/arm/miscexec.go:117:1
uint32_t arm_CPU_qsat(arm_CPU* c,int64_t v){
{
if ((v > cast<int64_t>(2147483647ULL))) {
c->Q = true;
return cast<uint32_t>(2147483647ULL);
}
if ((v < cast<int64_t>(-2147483648ULL))) {
c->Q = true;
return cast<uint32_t>(2147483648ULL);
}
return cast<uint32_t>(cast<int32_t>(v));
}
}
// tools/cpu/arm/miscexec.go:129:1
void arm_CPU_execSaturating(arm_CPU* c,uint32_t w){
{
int64_t rn = cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))))));
int64_t rm = cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm + rn))));
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm - rn))));
break;}
case cast<uint32_t>(2ULL):{
uint32_t dbl = arm_CPU_qsat(c,cast<int64_t>((cast<int64_t>(2ULL) * rn)));
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm + cast<int64_t>(cast<int32_t>(dbl))))));
break;}
default:{
uint32_t dbl = arm_CPU_qsat(c,cast<int64_t>((cast<int64_t>(2ULL) * rn)));
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm - cast<int64_t>(cast<int32_t>(dbl))))));
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:148:1
int64_t arm_half16(uint32_t v,uint32_t t){
{
if ((t == cast<uint32_t>(1ULL))) {
return cast<int64_t>(cast<int16_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
}
return cast<int64_t>(cast<int16_t>(v));
}
}
// tools/cpu/arm/miscexec.go:155:1
void arm_CPU_execSignedMul(arm_CPU* c,uint32_t w){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t x = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t y = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(8ULL):{
int64_t prod = cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)));
int64_t acc = cast<int64_t>((prod + cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rn)))));
arm_CPU_setReg(c,rd,cast<uint32_t>(acc));
if ((acc != cast<int64_t>(cast<int32_t>(acc)))) {
c->Q = true;
}
break;}
case cast<uint32_t>(9ULL):{
int64_t wide = cast<int64_t>((cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rm))) * arm_half16(arm_CPU_reg(c,rs),y)));
int32_t res = cast<int32_t>(shr<int64_t>(wide,cast<int64_t>(16ULL)));
if ((x == cast<uint32_t>(0ULL))) {
int64_t acc = cast<int64_t>((cast<int64_t>(res) + cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rn)))));
arm_CPU_setReg(c,rd,cast<uint32_t>(acc));
if ((acc != cast<int64_t>(cast<int32_t>(acc)))) {
c->Q = true;
}
}
else {
arm_CPU_setReg(c,rd,cast<uint32_t>(res));
}
break;}
case cast<uint32_t>(10ULL):{
int64_t prod = cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)));
int64_t acc = cast<int64_t>((cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(arm_CPU_reg(c,rd)),cast<int64_t>(32ULL)) | cast<uint64_t>(arm_CPU_reg(c,rn))))) + prod));
arm_CPU_setReg(c,rn,cast<uint32_t>(acc));
arm_CPU_setReg(c,rd,cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(acc),cast<int64_t>(32ULL))));
break;}
default:{
arm_CPU_setReg(c,rd,cast<uint32_t>(cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)))));
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:193:1
void arm_CPU_execSWI(arm_CPU* c,uint32_t w){
{
if ((bool(c->SWI) && c->SWI(c,cast<uint32_t>((w & cast<uint32_t>(16777215ULL)))))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
c->branched = true;
return ;
}
arm_CPU_exception(c,cast<uint32_t>(19ULL),cast<uint32_t>(8ULL),cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))));
}
}
// tools/cpu/arm/miscexec.go:204:1
void arm_CPU_exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
uint32_t saved = arm_CPU_CPSR(c);
arm_CPU_switchMode(c,mode);
arm_CPU_SetSPSR(c,saved);
c->R[cast<int64_t>(14ULL)] = returnAddr;
c->IRQDisable = true;
c->Thumb = false;
c->R[cast<int64_t>(15ULL)] = vector;
c->branched = true;
}
}
// tools/cpu/arm/miscexec.go:221:1
void arm_CPU_Exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
arm_CPU_exception(c,mode,vector,returnAddr);
}
}
// tools/cpu/arm/miscexec.go:228:1
bool arm_CPU_IRQ(arm_CPU* c){
{
if (c->IRQDisable) {
return false;
}
uint32_t ret = c->R[cast<int64_t>(15ULL)];
if (c->Thumb) {
ret += cast<uint32_t>(2ULL);
}
arm_CPU_exception(c,cast<uint32_t>(18ULL),cast<uint32_t>(24ULL),cast<uint32_t>((cast<uint32_t>((ret + cast<uint32_t>(4ULL))) - cast<uint32_t>((arm_CPU_boolToU(c,c->Thumb) * cast<uint32_t>(2ULL))))));
return true;
}
}
// tools/cpu/arm/miscexec.go:244:1
void arm_CPU_execCopro(arm_CPU* c,uint32_t w){
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL)));
uint32_t crn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t crm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if (bool(c->Coproc)) {
uint32_t v = arm_CPU_reg(c,rd);
c->Coproc(c,load,cp,op1,crn,crm,op2,(&v));
if ((load && (rd != cast<uint32_t>(15ULL)))) {
arm_CPU_setReg(c,rd,v);
}
return ;
}
if ((load && (rd != cast<uint32_t>(15ULL)))) {
arm_CPU_setReg(c,rd,cast<uint32_t>(0ULL));
}
}
}
// tools/cpu/arm/thumb.go:12:1
arm_Inst arm_DecodeThumb(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(2ULL))) {
return arm_Inst{addr,len(code),std::string(".hword",6),std::string(".hword ; truncated",18),cast<arm_Flow>(6ULL),{},{},true,{},cast<int64_t>(14ULL)};
}
uint32_t h = cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
arm_Inst in = arm_Inst{addr,cast<int64_t>(2ULL),{},{},cast<arm_Flow>(0ULL),{},{},true,true,cast<int64_t>(14ULL)};
{
if (((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(3ULL)))){
std::string op = arm_shiftName[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, #%d",14),op,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(3ULL))){
return arm_thumbAddSub(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(1ULL))){
std::string op = std::array<std::string,4>{std::string("MOV",3),std::string("CMP",3),std::string("ADD",3),std::string("SUB",3)}[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, #0x%X",12),op,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>((h & cast<uint32_t>(255ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(16ULL))){
return arm_thumbALU(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(17ULL))){
return arm_thumbHiReg(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(9ULL))){
in.Mnem = std::string("LDR",3);
in.Text = go_fmt_Sprintf(std::string("LDR %s, [pc, #0x%X]",19),arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
return arm_thumbLoadStoreReg(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL))){
return arm_thumbLoadStoreSExt(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(3ULL))){
return arm_thumbLoadStoreImm(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(8ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STRH",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDRH",4);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #0x%X]",18),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)))) * cast<uint32_t>(2ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(9ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [sp, #0x%X]",18),name,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(10ULL))){
std::string base = std::string("pc",2);
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
base = std::string("sp",2);
}
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD %s, %s, #0x%X",17),arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),base,cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(8ULL)) == cast<uint32_t>(176ULL))){
std::string sign = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
sign = std::string("-",1);
}
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD sp, #%s0x%X",15),sign,cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(127ULL)))) * cast<uint32_t>(4ULL))));
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(11ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))){
return arm_thumbPushPop(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(12ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STMIA",5);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDMIA",5);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s!, {%s}",12),name,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),arm_regList(cast<uint32_t>((h & cast<uint32_t>(255ULL)))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(13ULL))){
return arm_thumbCondBranch(h,addr,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(28ULL))){
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + off));
in.Mnem = std::string("B",1);
in.Text = go_fmt_Sprintf(std::string("B 0x%08X",8),target);
auto tmp65 = std::make_tuple(cast<arm_Flow>(2ULL),target,true);
in.Flow = std::get<0>(tmp65);
in.Target = std::get<1>(tmp65);
in.HasTarget = std::get<2>(tmp65);
}
else if ((((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(30ULL)) || (shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(31ULL))) || (shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(29ULL)))){
return arm_thumbLongBranch(code,h,addr,in);
}
else {
auto tmp66 = std::make_tuple(std::string(".hword",6),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp66);
in.Flow = std::get<1>(tmp66);
in.Text = go_fmt_Sprintf(std::string(".hword 0x%04X",13),h);
}
}
tmp64:;
return in;
}
}
// tools/cpu/arm/thumb.go:101:1
std::string arm_lo(uint32_t v){
{
return arm_regName[cast<uint32_t>((v & cast<uint32_t>(7ULL)))];
}
}
// tools/cpu/arm/thumb.go:103:1
arm_Inst arm_thumbAddSub(uint32_t h,arm_Inst in){
{
uint32_t i = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL)));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL)));
uint32_t rnOff = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)));
std::string name = std::string("ADD",3);
if ((op == cast<uint32_t>(1ULL))) {
name = std::string("SUB",3);
}
in.Mnem = name;
if ((i == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, #%d",14),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),rnOff);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_regName[rnOff]);
}
return in;
}
}
// tools/cpu/arm/thumb.go:125:1
arm_Inst arm_thumbALU(uint32_t h,arm_Inst in){
{
std::string op = arm_thumbALUOps[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(15ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),op,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:135:1
arm_Inst arm_thumbHiReg(uint32_t h,arm_Inst in){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
uint32_t rd = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(7ULL)))) | cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(4ULL))) & cast<uint32_t>(8ULL)))));
uint32_t rm = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(0ULL):{
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD %s, %s",10),arm_regName[rd],arm_regName[rm]);
if ((rd == cast<uint32_t>(15ULL))) {
in.Flow = cast<arm_Flow>(5ULL);
}
break;}
case cast<uint32_t>(1ULL):{
in.Mnem = std::string("CMP",3);
in.Text = go_fmt_Sprintf(std::string("CMP %s, %s",10),arm_regName[rd],arm_regName[rm]);
break;}
case cast<uint32_t>(2ULL):{
in.Mnem = std::string("MOV",3);
in.Text = go_fmt_Sprintf(std::string("MOV %s, %s",10),arm_regName[rd],arm_regName[rm]);
if ((rd == cast<uint32_t>(15ULL))) {
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
}
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = std::string("BX",2);
in.Text = go_fmt_Sprintf(std::string("BX %s",5),arm_regName[rm]);
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
}
else {
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX %s",6),arm_regName[rm]);
in.Flow = cast<arm_Flow>(7ULL);
}
break;}
}}
return in;
}
}
// tools/cpu/arm/thumb.go:177:1
arm_Inst arm_thumbLoadStoreReg(uint32_t h,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, %s]",15),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_lo(shr<uint32_t>(h,cast<int64_t>(6ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:192:1
arm_Inst arm_thumbLoadStoreSExt(uint32_t h,arm_Inst in){
{
std::string name = std::array<std::string,4>{std::string("STRH",4),std::string("LDSB",4),std::string("LDRH",4),std::string("LDSH",4)}[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, %s]",15),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_lo(shr<uint32_t>(h,cast<int64_t>(6ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:199:1
arm_Inst arm_thumbLoadStoreImm(uint32_t h,arm_Inst in){
{
uint32_t b = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(12ULL))) & cast<uint32_t>(1ULL)));
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
uint32_t off = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
else {
off *= cast<uint32_t>(4ULL);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #0x%X]",18),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),off);
return in;
}
}
// tools/cpu/arm/thumb.go:217:1
arm_Inst arm_thumbPushPop(uint32_t h,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
uint32_t r = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL)));
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
if ((l == cast<uint32_t>(1ULL))) {
uint32_t extra = cast<uint32_t>(0ULL);
if ((r == cast<uint32_t>(1ULL))) {
extra = cast<uint32_t>(32768ULL);
}
in.Mnem = std::string("POP",3);
in.Text = go_fmt_Sprintf(std::string("POP {%s}",8),arm_regList(cast<uint32_t>((mask | extra))));
if ((r == cast<uint32_t>(1ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
}
else {
uint32_t extra = cast<uint32_t>(0ULL);
if ((r == cast<uint32_t>(1ULL))) {
extra = cast<uint32_t>(16384ULL);
}
in.Mnem = std::string("PUSH",4);
in.Text = go_fmt_Sprintf(std::string("PUSH {%s}",9),arm_regList(cast<uint32_t>((mask | extra))));
}
return in;
}
}
// tools/cpu/arm/thumb.go:242:1
arm_Inst arm_thumbCondBranch(uint32_t h,uint32_t addr,arm_Inst in){
{
int64_t cond = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
if ((cond == cast<int64_t>(15ULL))) {
in.Mnem = std::string("SWI",3);
in.Text = go_fmt_Sprintf(std::string("SWI #0x%X",9),cast<uint32_t>((h & cast<uint32_t>(255ULL))));
in.Cond = cast<int64_t>(14ULL);
return in;
}
if ((cond == cast<int64_t>(14ULL))) {
auto tmp67 = std::make_tuple(std::string(".hword",6),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp67);
in.Flow = std::get<1>(tmp67);
in.Text = go_fmt_Sprintf(std::string(".hword 0x%04X",13),h);
return in;
}
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(255ULL))),cast<uint64_t>(8ULL)),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + off));
in.Cond = cond;
in.Mnem = (std::string("B",1) + arm_condName[cond]);
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
auto tmp68 = std::make_tuple(cast<arm_Flow>(1ULL),target,true);
in.Flow = std::get<0>(tmp68);
in.Target = std::get<1>(tmp68);
in.HasTarget = std::get<2>(tmp68);
return in;
}
}
// tools/cpu/arm/thumb.go:268:1
arm_Inst arm_thumbLongBranch(Slice<uint8_t> code,uint32_t h,uint32_t addr,arm_Inst in){
{
uint32_t top = shr<uint32_t>(h,cast<int64_t>(11ULL));
if ((top != cast<uint32_t>(30ULL))) {
in.Mnem = std::string("BL(suffix)",10);
in.Text = go_fmt_Sprintf(std::string("; BL/BLX suffix (no prefix) 0x%04X",34),h);
return in;
}
if ((len(code) < cast<int64_t>(4ULL))) {
in.Mnem = std::string("BL(hi)",6);
in.Text = go_fmt_Sprintf(std::string("BL (hi) #0x%X",13),cast<uint32_t>((h & cast<uint32_t>(2047ULL))));
return in;
}
uint32_t h2 = cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(2ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(3ULL)]),cast<int64_t>(8ULL))));
uint32_t suf = shr<uint32_t>(h2,cast<int64_t>(11ULL));
if (((suf != cast<uint32_t>(31ULL)) && (suf != cast<uint32_t>(29ULL)))) {
in.Mnem = std::string("BL(hi)",6);
in.Text = go_fmt_Sprintf(std::string("BL (hi) #0x%X",13),cast<uint32_t>((h & cast<uint32_t>(2047ULL))));
return in;
}
in.Len = cast<int64_t>(4ULL);
uint32_t hiOff = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(12ULL));
uint32_t loOff = shl<uint32_t>((cast<uint32_t>((h2 & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + hiOff)) + loOff));
if ((suf == cast<uint32_t>(29ULL))) {
target &= ~(cast<uint32_t>(3ULL));
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX 0x%08X",10),target);
auto tmp69 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,false);
in.Flow = std::get<0>(tmp69);
in.Target = std::get<1>(tmp69);
in.HasTarget = std::get<2>(tmp69);
in.TargetThumb = std::get<3>(tmp69);
}
else {
in.Mnem = std::string("BL",2);
in.Text = go_fmt_Sprintf(std::string("BL 0x%08X",9),target);
auto tmp70 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,true);
in.Flow = std::get<0>(tmp70);
in.Target = std::get<1>(tmp70);
in.HasTarget = std::get<2>(tmp70);
in.TargetThumb = std::get<3>(tmp70);
}
return in;
}
}
// tools/cpu/arm/thumbexec.go:9:1
int64_t arm_CPU_stepThumb(arm_CPU* c){
{
uint32_t h = arm_CPU_read16(c,c->cur);
arm_CPU_execThumb(c,h);
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)));
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/thumbexec.go:18:1
void arm_CPU_execThumb(arm_CPU* c,uint32_t h){
{
{
if (((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(3ULL)))){
uint32_t typ = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)));
auto tmp72 = arm_CPU_shift(c,typ,cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))),c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))],false,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp72);
uint32_t carry = std::get<1>(tmp72);
c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(3ULL))){
arm_CPU_thumbAddSub(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(1ULL))){
arm_CPU_thumbImm(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(16ULL))){
arm_CPU_thumbALUExec(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(17ULL))){
arm_CPU_thumbHiRegExec(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(9ULL))){
uint32_t addr = cast<uint32_t>(((((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)))) & ~(cast<uint32_t>(3ULL)))) + cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)))));
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = arm_CPU_read32(c,addr);
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_thumbLoadStoreReg(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL))){
arm_CPU_thumbLoadStoreSExt(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(3ULL))){
arm_CPU_thumbLoadStoreImm(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(8ULL))){
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)))) * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))] = arm_CPU_read16(c,addr);
}
else {
arm_CPU_write16(c,addr,c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))]);
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(9ULL))){
uint32_t addr = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] + cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)))));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = arm_CPU_read32(c,addr);
}
else {
arm_CPU_write32(c,addr,c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))]);
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(10ULL))){
uint32_t off = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] + off));
}
else {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = cast<uint32_t>(((((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)))) & ~(cast<uint32_t>(3ULL)))) + off));
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(8ULL)) == cast<uint32_t>(176ULL))){
uint32_t off = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(127ULL)))) * cast<uint32_t>(4ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<int64_t>(13ULL)] -= off;
}
else {
c->R[cast<int64_t>(13ULL)] += off;
}
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(11ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))){
arm_CPU_thumbPushPop(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(12ULL))){
arm_CPU_thumbBlock(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(13ULL))){
arm_CPU_thumbCondBranch(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(28ULL))){
arm_CPU_branchTo(c,cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(1ULL)))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(30ULL))){
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(12ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(31ULL))){
uint32_t target = cast<uint32_t>((c->R[cast<int64_t>(14ULL)] + shl<uint32_t>((cast<uint32_t>((h & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL))));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
arm_CPU_branchTo(c,target);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(29ULL))){
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_Halt(c,std::string("Thumb BLX (immediate) is undefined on ARMv4T: 0x%04X at 0x%08X",62),h,c->cur);
return ;
}
uint32_t target = ((cast<uint32_t>((c->R[cast<int64_t>(14ULL)] + shl<uint32_t>((cast<uint32_t>((h & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL))))) & ~(cast<uint32_t>(3ULL)));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
c->Thumb = false;
c->R[cast<int64_t>(15ULL)] = target;
c->branched = true;
}
else {
arm_CPU_Halt(c,std::string("undefined Thumb 0x%04X at 0x%08X",32),h,c->cur);
}
}
tmp71:;
}
}
// tools/cpu/arm/thumbexec.go:100:1
void arm_CPU_thumbAddSub(arm_CPU* c,uint32_t h){
{
auto tmp73 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp73);
uint32_t rs = std::get<1>(tmp73);
uint32_t operand={};
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
operand = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)));
}
else {
operand = c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))];
}
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[rd] = arm_CPU_sub_(c,c->R[rs],operand,cast<uint32_t>(1ULL));
}
else {
c->R[rd] = arm_CPU_add(c,c->R[rs],operand,cast<uint32_t>(0ULL));
}
}
}
// tools/cpu/arm/thumbexec.go:115:1
void arm_CPU_thumbImm(arm_CPU* c,uint32_t h){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
uint32_t imm = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
c->R[rd] = imm;
arm_CPU_setNZ(c,imm);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_sub_(c,c->R[rd],imm,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(2ULL):{
c->R[rd] = arm_CPU_add(c,c->R[rd],imm,cast<uint32_t>(0ULL));
break;}
default:{
c->R[rd] = arm_CPU_sub_(c,c->R[rd],imm,cast<uint32_t>(1ULL));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:131:1
void arm_CPU_thumbALUExec(arm_CPU* c,uint32_t h){
{
auto tmp74 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp74);
uint32_t rs = std::get<1>(tmp74);
auto tmp75 = std::make_tuple(c->R[rd],c->R[rs]);
uint32_t a = std::get<0>(tmp75);
uint32_t b = std::get<1>(tmp75);
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
c->R[rd] = cast<uint32_t>((a & b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(1ULL):{
c->R[rd] = cast<uint32_t>((a ^ b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(2ULL):{
auto tmp76 = arm_CPU_shift(c,cast<uint32_t>(0ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp76);
uint32_t carry = std::get<1>(tmp76);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(3ULL):{
auto tmp77 = arm_CPU_shift(c,cast<uint32_t>(1ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp77);
uint32_t carry = std::get<1>(tmp77);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(4ULL):{
auto tmp78 = arm_CPU_shift(c,cast<uint32_t>(2ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp78);
uint32_t carry = std::get<1>(tmp78);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(5ULL):{
c->R[rd] = arm_CPU_add(c,a,b,arm_CPU_boolToU(c,c->C));
break;}
case cast<uint32_t>(6ULL):{
c->R[rd] = arm_CPU_sub_(c,a,b,arm_CPU_boolToU(c,c->C));
break;}
case cast<uint32_t>(7ULL):{
auto tmp79 = arm_CPU_shift(c,cast<uint32_t>(3ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp79);
uint32_t carry = std::get<1>(tmp79);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(8ULL):{
arm_CPU_setNZ(c,cast<uint32_t>((a & b)));
break;}
case cast<uint32_t>(9ULL):{
c->R[rd] = arm_CPU_sub_(c,cast<uint32_t>(0ULL),b,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(10ULL):{
arm_CPU_sub_(c,a,b,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(11ULL):{
arm_CPU_add(c,a,b,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(12ULL):{
c->R[rd] = cast<uint32_t>((a | b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(13ULL):{
c->R[rd] = cast<uint32_t>((a * b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(14ULL):{
c->R[rd] = (a & ~(b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
default:{
c->R[rd] = cast<uint32_t>(~b);
arm_CPU_setNZ(c,c->R[rd]);
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:188:1
void arm_CPU_thumbHiRegExec(arm_CPU* c,uint32_t h){
{
uint32_t rd = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(7ULL)))) | cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(4ULL))) & cast<uint32_t>(8ULL)))));
uint32_t rm = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(15ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
uint32_t v = cast<uint32_t>((arm_CPU_reg(c,rd) + arm_CPU_reg(c,rm)));
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_branchTo(c,v);
}
else {
c->R[rd] = v;
}
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_sub_(c,arm_CPU_reg(c,rd),arm_CPU_reg(c,rm),cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(2ULL):{
uint32_t v = arm_CPU_reg(c,rm);
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_branchTo(c,v);
}
else {
c->R[rd] = v;
}
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_Halt(c,std::string("Thumb BLX (register) is undefined on ARMv4T: 0x%04X at 0x%08X",61),h,c->cur);
return ;
}
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
}
arm_CPU_bxTo(c,arm_CPU_reg(c,rm));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:220:1
void arm_CPU_thumbLoadStoreReg(arm_CPU* c,uint32_t h){
{
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))]));
uint32_t rd = cast<uint32_t>((h & cast<uint32_t>(7ULL)));
auto tmp80 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL))));
uint32_t l = std::get<0>(tmp80);
uint32_t b = std::get<1>(tmp80);
{
if (((l == cast<uint32_t>(1ULL)) && (b == cast<uint32_t>(1ULL)))){
c->R[rd] = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else if ((l == cast<uint32_t>(1ULL))){
c->R[rd] = arm_CPU_read32(c,addr);
}
else if ((b == cast<uint32_t>(1ULL))){
arm_CPU_write8(c,addr,cast<uint8_t>(c->R[rd]));
}
else {
arm_CPU_write32(c,addr,c->R[rd]);
}
}
tmp81:;
}
}
// tools/cpu/arm/thumbexec.go:236:1
void arm_CPU_thumbLoadStoreSExt(arm_CPU* c,uint32_t h){
{
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))]));
uint32_t rd = cast<uint32_t>((h & cast<uint32_t>(7ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_write16(c,addr,c->R[rd]);
break;}
case cast<uint32_t>(1ULL):{
c->R[rd] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(arm_CPU_read8(c,addr))));
break;}
case cast<uint32_t>(2ULL):{
c->R[rd] = arm_CPU_read16(c,addr);
break;}
default:{
c->R[rd] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(arm_CPU_read16(c,addr))));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:251:1
void arm_CPU_thumbLoadStoreImm(arm_CPU* c,uint32_t h){
{
auto tmp82 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp82);
uint32_t rb = std::get<1>(tmp82);
auto tmp83 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(12ULL))) & cast<uint32_t>(1ULL))));
uint32_t l = std::get<0>(tmp83);
uint32_t b = std::get<1>(tmp83);
uint32_t off = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
if ((b == cast<uint32_t>(0ULL))) {
off *= cast<uint32_t>(4ULL);
}
uint32_t addr = cast<uint32_t>((c->R[rb] + off));
{
if (((l == cast<uint32_t>(1ULL)) && (b == cast<uint32_t>(1ULL)))){
c->R[rd] = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else if ((l == cast<uint32_t>(1ULL))){
c->R[rd] = arm_CPU_read32(c,addr);
}
else if ((b == cast<uint32_t>(1ULL))){
arm_CPU_write8(c,addr,cast<uint8_t>(c->R[rd]));
}
else {
arm_CPU_write32(c,addr,c->R[rd]);
}
}
tmp84:;
}
}
// tools/cpu/arm/thumbexec.go:271:1
void arm_CPU_thumbPushPop(arm_CPU* c,uint32_t h){
{
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
mask |= cast<uint32_t>(32768ULL);
}
uint32_t addr = c->R[cast<int64_t>(13ULL)];
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
uint32_t v = arm_CPU_read32aligned(c,addr);
if ((i == cast<uint32_t>(15ULL))) {
arm_CPU_bxTo(c,v);
}
else {
c->R[i] = v;
}
addr += cast<uint32_t>(4ULL);
}
}c->R[cast<int64_t>(13ULL)] = addr;
}
else {
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
mask |= cast<uint32_t>(16384ULL);
}
uint32_t addr = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] - cast<uint32_t>((cast<uint32_t>(4ULL) * cast<uint32_t>(go_bits_OnesCount32(mask))))));
c->R[cast<int64_t>(13ULL)] = addr;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
arm_CPU_write32aligned(c,addr,c->R[i]);
addr += cast<uint32_t>(4ULL);
}
}}
}
}
// tools/cpu/arm/thumbexec.go:307:1
void arm_CPU_thumbBlock(arm_CPU* c,uint32_t h){
{
uint32_t rb = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
uint32_t addr = c->R[rb];
bool load = (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(8ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
if (load) {
c->R[i] = arm_CPU_read32aligned(c,addr);
}
else {
arm_CPU_write32aligned(c,addr,c->R[i]);
}
addr += cast<uint32_t>(4ULL);
}
}if ((!((load && (cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),rb)))) != cast<uint32_t>(0ULL)))))) {
c->R[rb] = addr;
}
}
}
// tools/cpu/arm/thumbexec.go:329:1
void arm_CPU_thumbCondBranch(arm_CPU* c,uint32_t h){
{
int64_t cond = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
if ((cond == cast<int64_t>(15ULL))) {
if ((bool(c->SWI) && c->SWI(c,cast<uint32_t>((h & cast<uint32_t>(255ULL)))))) {
return ;
}
arm_CPU_exception(c,cast<uint32_t>(19ULL),cast<uint32_t>(8ULL),cast<uint32_t>((c->cur + cast<uint32_t>(2ULL))));
return ;
}
if ((cond == cast<int64_t>(14ULL))) {
arm_CPU_Halt(c,std::string("undefined Thumb 0x%04X at 0x%08X",32),h,c->cur);
return ;
}
if (arm_CPU_cond(c,cond)) {
arm_CPU_branchTo(c,cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(255ULL))),cast<uint64_t>(8ULL)),cast<int64_t>(1ULL)))));
}
}
}
// tools/cpu/arm/variant.go:52:1
std::string arm_Variant_String(arm_Variant v){
{
{
switch(v){
case cast<arm_Variant>(0ULL):{
return std::string("ARMv5TE",7);
break;}
case cast<arm_Variant>(1ULL):{
return std::string("ARMv6K",6);
break;}
case cast<arm_Variant>(2ULL):{
return std::string("ARMv4T",6);
break;}
}}
return std::string("ARM?",4);
}
}
// tools/cpu/arm/variant.go:66:1
bool arm_Variant_isV6(arm_Variant v){
{
return (v == cast<arm_Variant>(1ULL));
}
}
// tools/cpu/arm/variant.go:70:1
bool arm_Variant_v5OrLater(arm_Variant v){
{
return (v != cast<arm_Variant>(2ULL));
}
}
// tools/cpu/arm/variant.go:73:1
arm_Inst arm_DecodeVariant(Slice<uint8_t> code,uint32_t addr,bool thumb,arm_Variant v){
{
if (thumb) {
return arm_DecodeThumb(code,addr);
}
return arm_DecodeARMVariant(code,addr,v);
}
}
// tools/cpu/arm/variant.go:88:1
arm_Inst arm_DecodeARMVariant(Slice<uint8_t> code,uint32_t addr,arm_Variant v){
{
if ((v == cast<arm_Variant>(2ULL))) {
return arm_v4tFilter(arm_DecodeARM(code,addr));
}
if ((!arm_Variant_isV6(v))) {
return arm_DecodeARM(code,addr);
}
auto tmp85 = arm_word(code);
uint32_t w = std::get<0>(tmp85);
bool ok = std::get<1>(tmp85);
if ((!ok)) {
return arm_Inst{addr,len(code),std::string(".word",5),std::string(".word ; truncated",17),cast<arm_Flow>(6ULL),{},{},{},{},cast<int64_t>(14ULL)};
}
arm_Inst in = arm_Inst{addr,cast<int64_t>(4ULL),{},{},cast<arm_Flow>(0ULL),{},{},{},{},cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)))};
{
auto tmp86 = arm_decodeARMv6(w,addr,in);
arm_Inst out = std::get<0>(tmp86);
bool handled = std::get<1>(tmp86);
if (handled) {
return out;
}
}
return arm_DecodeARM(code,addr);
}
}
// tools/cpu/arm/variant.go:126:1
arm_Inst arm_v4tFilter(arm_Inst in){
{
std::string mnem = in.Mnem;
{auto&& tmp87 = arm_condName;
for(int64_t tmp88=0;tmp88<len(tmp87);++tmp88){
auto c=tmp87[tmp88];if ((((c != std::string("",0)) && (len(mnem) > len(c))) && (sub(mnem,cast<int64_t>((len(mnem) - len(c))),len(mnem)) == c))) {
mnem = sub(mnem,0,cast<int64_t>((len(mnem) - len(c))));
break;
}
}}
if ((!get(arm_v4tMnem,mnem))) {
return in;
}
return arm_Inst{in.Addr,in.Len,std::string(".word",5),((std::string(".word ; ",8) + in.Text) + std::string(" \342\200\224 undefined on ARMv4T",24)),cast<arm_Flow>(6ULL),{},{},{},{},in.Cond};
}
}
// tools/cpu/arm/vfp.go:40:1
float arm_CPU_sGet(arm_CPU* c,uint32_t n){
{
return go_math_Float32frombits(c->VFP.S[n]);
}
}
// tools/cpu/arm/vfp.go:41:1
void arm_CPU_sSet(arm_CPU* c,uint32_t n,float v){
{
c->VFP.S[n] = go_math_Float32bits(v);
}
}
// tools/cpu/arm/vfp.go:42:1
uint32_t arm_CPU_sBits(arm_CPU* c,uint32_t n){
{
return c->VFP.S[n];
}
}
// tools/cpu/arm/vfp.go:43:1
void arm_CPU_sSetBits(arm_CPU* c,uint32_t n,uint32_t v){
{
c->VFP.S[n] = v;
}
}
// tools/cpu/arm/vfp.go:45:1
double arm_CPU_dGet(arm_CPU* c,uint32_t n){
{
return go_math_Float64frombits(cast<uint64_t>((cast<uint64_t>(c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * n))]) | shl<uint64_t>(cast<uint64_t>(c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * n)) + cast<uint32_t>(1ULL)))]),cast<int64_t>(32ULL)))));
}
}
// tools/cpu/arm/vfp.go:48:1
void arm_CPU_dSet(arm_CPU* c,uint32_t n,double v){
{
uint64_t b = go_math_Float64bits(v);
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * n))] = cast<uint32_t>(b);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * n)) + cast<uint32_t>(1ULL)))] = cast<uint32_t>(shr<uint64_t>(b,cast<int64_t>(32ULL)));
}
}
// tools/cpu/arm/vfp.go:57:1
uint32_t arm_sReg(uint32_t v,uint32_t bit){
{
return cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(1ULL)) | bit));
}
}
// tools/cpu/arm/vfp.go:58:1
uint32_t arm_dReg(uint32_t v,uint32_t bit){
{
return cast<uint32_t>((shl<uint32_t>(bit,cast<int64_t>(4ULL)) | v));
}
}
// tools/cpu/arm/vfp.go:61:1
bool arm_isVFP(uint32_t w){
{
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
return ((cp == cast<uint32_t>(10ULL)) || (cp == cast<uint32_t>(11ULL)));
}
}
// tools/cpu/arm/vfp.go:72:1
arm_Inst arm_decodeVFP(uint32_t w,arm_Inst in){
{
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
std::string prec = std::string("F64",3);
if (single) {
prec = std::string("F32",3);
}
std::string sfx = arm_cn(in.Cond);
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL))){
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(127ULL))) == cast<uint32_t>(98ULL)) || (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(31ULL))) == cast<uint32_t>(24ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))))) {
in.Mnem = (std::string("VMOV",4) + sfx);
in.Text = go_fmt_Sprintf(std::string("%s (64-bit move) ; 0x%08X",25),in.Mnem,w);
return in;
}
auto tmp90 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))));
uint32_t p = std::get<0>(tmp90);
uint32_t wb = std::get<1>(tmp90);
uint32_t l = std::get<2>(tmp90);
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string vd = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
if (((p == cast<uint32_t>(1ULL)) && (wb == cast<uint32_t>(0ULL)))) {
std::string name = std::string("VSTR",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("VLDR",4);
}
uint32_t off = shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(255ULL)))),cast<int64_t>(2ULL));
std::string sign = std::string("+",1);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
in.Mnem = (name + sfx);
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #%s0x%X]",20),in.Mnem,vd,rn,sign,off);
return in;
}
std::string name = std::string("VSTM",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("VLDM",4);
}
in.Mnem = (name + sfx);
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s{%d regs}",20),in.Mnem,rn,arm_wbMark(wb),std::string("",0),cast<uint32_t>((w & cast<uint32_t>(255ULL))));
return in;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_decodeVFPMove(w,in);
}
return arm_decodeVFPData(w,in,single,prec,sfx);
}
}
tmp89:;
in.Mnem = std::string("VFP?",4);
in.Text = go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w);
return in;
}
}
// tools/cpu/arm/vfp.go:123:1
arm_Inst arm_decodeVFPMove(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string rt = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(7ULL))) {
if ((l == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("VMRS",4) + arm_cn(in.Cond));
std::string dst = rt;
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(15ULL))) {
dst = std::string("APSR_nzcv",9);
}
in.Text = go_fmt_Sprintf(std::string("%s %s, FPSCR",12),in.Mnem,dst);
}
else {
in.Mnem = (std::string("VMSR",4) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s FPSCR, %s",12),in.Mnem,rt);
}
return in;
}
std::string sn = arm_vfpRegName(true,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
in.Mnem = (std::string("VMOV",4) + arm_cn(in.Cond));
if ((l == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rt,sn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,sn,rt);
}
return in;
}
}
// tools/cpu/arm/vfp.go:150:1
arm_Inst arm_decodeVFPData(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx){
{
uint32_t pqr = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(2ULL)) | shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(1ULL)))) | cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)))));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
std::string vd = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
std::string vn = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
std::string vm = arm_vfpRegName(single,cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
std::string name = std::string("",0);
{
switch(pqr){
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VADD",4);
}
else {
name = std::string("VSUB",4);
}
break;}
case cast<uint32_t>(2ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VMUL",4);
}
else {
name = std::string("VNMUL",5);
}
break;}
case cast<uint32_t>(4ULL):{
name = std::string("VDIV",4);
break;}
case cast<uint32_t>(0ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VMLA",4);
}
else {
name = std::string("VMLS",4);
}
break;}
case cast<uint32_t>(1ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VNMLS",5);
}
else {
name = std::string("VNMLA",5);
}
break;}
case cast<uint32_t>(7ULL):{
return arm_decodeVFPExt(w,in,single,prec,sfx,vd,vm);
break;}
}}
if ((name == std::string("",0))) {
auto tmp91 = std::make_tuple(std::string("VFP?",4),go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w));
in.Mnem = std::get<0>(tmp91);
in.Text = std::get<1>(tmp91);
return in;
}
in.Mnem = (((name + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,vd,vn,vm);
return in;
}
}
// tools/cpu/arm/vfp.go:197:1
arm_Inst arm_decodeVFPExt(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx,std::string vd,std::string vm){
{
uint32_t opc2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t opc3 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((opc3 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (((std::string("VMOV",4) + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, #imm",11),in.Mnem,vd);
return in;
}
std::string name = std::string("",0);
{
switch(opc2){
case cast<uint32_t>(0ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
name = std::string("VMOV",4);
}
else {
name = std::string("VABS",4);
}
break;}
case cast<uint32_t>(1ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
name = std::string("VNEG",4);
}
else {
name = std::string("VSQRT",5);
}
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):{
name = std::string("VCMP",4);
break;}
case cast<uint32_t>(7ULL):{
name = std::string("VCVT",4);
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(12ULL):case cast<uint32_t>(13ULL):{
name = std::string("VCVT",4);
break;}
}}
if ((name == std::string("",0))) {
auto tmp92 = std::make_tuple(std::string("VFP?",4),go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w));
in.Mnem = std::get<0>(tmp92);
in.Text = std::get<1>(tmp92);
return in;
}
in.Mnem = (((name + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,vd,vm);
return in;
}
}
// tools/cpu/arm/vfp.go:235:1
std::string arm_vfpRegName(bool single,uint32_t v,uint32_t bit){
{
if (single) {
return go_fmt_Sprintf(std::string("s%d",3),arm_sReg(v,bit));
}
return go_fmt_Sprintf(std::string("d%d",3),arm_dReg(v,bit));
}
}
// tools/cpu/arm/vfp.go:242:1
std::string arm_wbMark(uint32_t wb){
{
if ((wb == cast<uint32_t>(1ULL))) {
return std::string("!",1);
}
return std::string("",0);
}
}
// tools/cpu/arm/vfp.go:253:1
bool arm_CPU_execVFP(arm_CPU* c,uint32_t w){
{
if ((!arm_isVFP(w))) {
return false;
}
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL))){
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(127ULL))) == cast<uint32_t>(98ULL)) || ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(31ULL))) == cast<uint32_t>(24ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))))) {
return arm_CPU_execVFPMove64(c,w);
}
arm_CPU_execVFPLoadStore(c,w);
return true;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_CPU_execVFPMove(c,w);
}
return arm_CPU_execVFPData(c,w);
}
}
tmp93:;
return false;
}
}
// tools/cpu/arm/vfp.go:275:1
void arm_CPU_execVFPLoadStore(arm_CPU* c,uint32_t w){
{
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t d = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t wb = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t vd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t imm8 = cast<uint32_t>((w & cast<uint32_t>(255ULL)));
uint32_t base = arm_CPU_reg(c,rn);
if (((p == cast<uint32_t>(1ULL)) && (wb == cast<uint32_t>(0ULL)))) {
uint32_t offset = shl<uint32_t>(imm8,cast<int64_t>(2ULL));
uint32_t addr = cast<uint32_t>((base + offset));
if ((u == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>((base - offset));
}
if (single) {
uint32_t sd = arm_sReg(vd,d);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_sSetBits(c,sd,arm_CPU_read32(c,addr));
}
else {
arm_CPU_write32(c,addr,arm_CPU_sBits(c,sd));
}
}
else {
uint32_t dd = arm_dReg(vd,d);
if ((l == cast<uint32_t>(1ULL))) {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))] = arm_CPU_read32(c,addr);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))] = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
}
else {
arm_CPU_write32(c,addr,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))]);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))]);
}
}
return ;
}
uint32_t count = imm8;
uint32_t regs = count;
if ((!single)) {
regs = divi<uint32_t>(count,cast<uint32_t>(2ULL));
}
uint32_t addr = base;
if ((u == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>((base - cast<uint32_t>((count * cast<uint32_t>(4ULL)))));
}
uint32_t start = addr;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < regs);i++){
if (single) {
uint32_t sd = cast<uint32_t>((arm_sReg(vd,d) + i));
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_sSetBits(c,sd,arm_CPU_read32(c,addr));
}
else {
arm_CPU_write32(c,addr,arm_CPU_sBits(c,sd));
}
addr += cast<uint32_t>(4ULL);
}
else {
uint32_t dd = cast<uint32_t>((arm_dReg(vd,d) + i));
if ((l == cast<uint32_t>(1ULL))) {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))] = arm_CPU_read32(c,addr);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))] = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
}
else {
arm_CPU_write32(c,addr,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))]);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))]);
}
addr += cast<uint32_t>(8ULL);
}
}
}if ((wb == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,cast<uint32_t>((base + cast<uint32_t>((count * cast<uint32_t>(4ULL))))));
}
else {
arm_CPU_setReg(c,rn,start);
}
}
}
}
// tools/cpu/arm/vfp.go:363:1
bool arm_CPU_execVFPData(arm_CPU* c,uint32_t w){
{
uint32_t pqr = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(2ULL)) | shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(1ULL)))) | cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)))));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
uint32_t dn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t dd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t dm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t nBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL)));
uint32_t mBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t dBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t vd={};
uint32_t vn={};
uint32_t vm={};
if (single) {
auto tmp94 = std::make_tuple(arm_sReg(dd,dBit),arm_sReg(dn,nBit),arm_sReg(dm,mBit));
vd = std::get<0>(tmp94);
vn = std::get<1>(tmp94);
vm = std::get<2>(tmp94);
}
else {
auto tmp95 = std::make_tuple(arm_dReg(dd,dBit),arm_dReg(dn,nBit),arm_dReg(dm,mBit));
vd = std::get<0>(tmp95);
vn = std::get<1>(tmp95);
vm = std::get<2>(tmp95);
}
auto binop = [&](std::function<float(float,float)> f32,std::function<double(double,double)> f64)->void{
if (single) {
arm_CPU_sSet(c,vd,f32(arm_CPU_sGet(c,vn),arm_CPU_sGet(c,vm)));
}
else {
arm_CPU_dSet(c,vd,f64(arm_CPU_dGet(c,vn),arm_CPU_dGet(c,vm)));
}
}
;
{
switch(pqr){
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(0ULL))) {
binop([&](float a,float b)->float{
return (a + b);
}
,[&](double a,double b)->double{
return (a + b);
}
);
}
else {
binop([&](float a,float b)->float{
return (a - b);
}
,[&](double a,double b)->double{
return (a - b);
}
);
}
return true;
break;}
case cast<uint32_t>(2ULL):{
binop([&](float a,float b)->float{
return (a * b);
}
,[&](double a,double b)->double{
return (a * b);
}
);
if ((op == cast<uint32_t>(1ULL))) {
arm_CPU_vfpNegate(c,vd,single);
}
return true;
break;}
case cast<uint32_t>(4ULL):{
binop([&](float a,float b)->float{
return (a / b);
}
,[&](double a,double b)->double{
return (a / b);
}
);
return true;
break;}
case cast<uint32_t>(0ULL):{
arm_CPU_vfpMulAcc(c,vd,vn,vm,single,(op == cast<uint32_t>(1ULL)),false);
return true;
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_vfpMulAcc(c,vd,vn,vm,single,(op == cast<uint32_t>(1ULL)),true);
return true;
break;}
case cast<uint32_t>(7ULL):{
return arm_CPU_execVFPExt(c,w,single,vd,vm);
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented VFP data op %03b (0x%08X) at 0x%08X",49),pqr,w,c->cur);
return true;
}
}
// tools/cpu/arm/vfp.go:426:1
void arm_CPU_vfpNegate(arm_CPU* c,uint32_t vd,bool single){
{
if (single) {
arm_CPU_sSet(c,vd,cast<float>(-arm_CPU_sGet(c,vd)));
}
else {
arm_CPU_dSet(c,vd,cast<double>(-arm_CPU_dGet(c,vd)));
}
}
}
// tools/cpu/arm/vfp.go:436:1
void arm_CPU_vfpMulAcc(arm_CPU* c,uint32_t vd,uint32_t vn,uint32_t vm,bool single,bool sub_,bool negAcc){
{
if (single) {
float acc = arm_CPU_sGet(c,vd);
if (negAcc) {
acc = cast<float>(-acc);
}
float p = (arm_CPU_sGet(c,vn) * arm_CPU_sGet(c,vm));
if (sub_) {
p = cast<float>(-p);
}
arm_CPU_sSet(c,vd,(acc + p));
}
else {
double acc = arm_CPU_dGet(c,vd);
if (negAcc) {
acc = cast<double>(-acc);
}
double p = (arm_CPU_dGet(c,vn) * arm_CPU_dGet(c,vm));
if (sub_) {
p = cast<double>(-p);
}
arm_CPU_dSet(c,vd,(acc + p));
}
}
}
// tools/cpu/arm/vfp.go:462:1
bool arm_CPU_execVFPExt(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm){
{
uint32_t opc2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t opc3 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((opc3 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
uint64_t imm = arm_vfpExpandImm(w,single);
if (single) {
arm_CPU_sSetBits(c,vd,cast<uint32_t>(imm));
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vd))] = cast<uint32_t>(imm);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vd)) + cast<uint32_t>(1ULL)))] = cast<uint32_t>(shr<uint64_t>(imm,cast<int64_t>(32ULL)));
}
return true;
}
{
switch(opc2){
case cast<uint32_t>(0ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
arm_CPU_vfpCopy(c,vd,vm,single);
}
else {
arm_CPU_vfpUnary(c,vd,vm,single,go_math_Abs);
}
return true;
break;}
case cast<uint32_t>(1ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
arm_CPU_vfpUnary(c,vd,vm,single,[&](double f)->double{
return cast<double>(-f);
}
);
}
else {
arm_CPU_vfpUnary(c,vd,vm,single,go_math_Sqrt);
}
return true;
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):{
arm_CPU_execVCMP(c,w,single,vd,vm,(opc2 == cast<uint32_t>(5ULL)));
return true;
break;}
case cast<uint32_t>(7ULL):{
arm_CPU_execVCVTPrec(c,w,single);
return true;
break;}
case cast<uint32_t>(8ULL):{
arm_CPU_execVCVTFromInt(c,w,single);
return true;
break;}
case cast<uint32_t>(12ULL):case cast<uint32_t>(13ULL):{
arm_CPU_execVCVTToInt(c,w,single);
return true;
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented VFP ext op2=0x%X (0x%08X) at 0x%08X",49),opc2,w,c->cur);
return true;
}
}
// tools/cpu/arm/vfp.go:510:1
void arm_CPU_vfpCopy(arm_CPU* c,uint32_t vd,uint32_t vm,bool single){
{
if (single) {
arm_CPU_sSetBits(c,vd,arm_CPU_sBits(c,vm));
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vd))] = c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vm))];
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vd)) + cast<uint32_t>(1ULL)))] = c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vm)) + cast<uint32_t>(1ULL)))];
}
}
}
// tools/cpu/arm/vfp.go:519:1
void arm_CPU_vfpUnary(arm_CPU* c,uint32_t vd,uint32_t vm,bool single,std::function<double(double)> f){
{
if (single) {
arm_CPU_sSet(c,vd,cast<float>(f(cast<double>(arm_CPU_sGet(c,vm)))));
}
else {
arm_CPU_dSet(c,vd,f(arm_CPU_dGet(c,vm)));
}
}
}
// tools/cpu/arm/vfp.go:530:1
void arm_CPU_execVCMP(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm,bool withZero){
{
double a={};
double b={};
if (single) {
a = cast<double>(arm_CPU_sGet(c,vd));
if ((!withZero)) {
b = cast<double>(arm_CPU_sGet(c,vm));
}
}
else {
a = arm_CPU_dGet(c,vd);
if ((!withZero)) {
b = arm_CPU_dGet(c,vm);
}
}
bool n={};
bool z={};
bool cc={};
bool v={};
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
auto tmp97 = std::make_tuple(true,true);
cc = std::get<0>(tmp97);
v = std::get<1>(tmp97);
}
else if ((a == b)){
auto tmp98 = std::make_tuple(true,true);
z = std::get<0>(tmp98);
cc = std::get<1>(tmp98);
}
else if ((a < b)){
n = true;
}
else {
cc = true;
}
}
tmp96:;
uint32_t f = (c->VFP.FPSCR & ~(cast<uint32_t>(4026531840ULL)));
if (n) {
f |= cast<uint32_t>(2147483648ULL);
}
if (z) {
f |= cast<uint32_t>(1073741824ULL);
}
if (cc) {
f |= cast<uint32_t>(536870912ULL);
}
if (v) {
f |= cast<uint32_t>(268435456ULL);
}
c->VFP.FPSCR = f;
}
}
// tools/cpu/arm/vfp.go:575:1
uint32_t arm_vdOf(uint32_t w,bool single){
{
if (single) {
return arm_sReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
}
return arm_dReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
}
}
// tools/cpu/arm/vfp.go:581:1
uint32_t arm_vmOf(uint32_t w,bool single){
{
if (single) {
return arm_sReg(cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
}
return arm_dReg(cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
}
}
// tools/cpu/arm/vfp.go:591:1
void arm_CPU_execVCVTPrec(arm_CPU* c,uint32_t w,bool srcSingle){
{
if (srcSingle) {
arm_CPU_dSet(c,arm_vdOf(w,false),cast<double>(arm_CPU_sGet(c,arm_vmOf(w,true))));
}
else {
arm_CPU_sSet(c,arm_vdOf(w,true),cast<float>(arm_CPU_dGet(c,arm_vmOf(w,false))));
}
}
}
// tools/cpu/arm/vfp.go:601:1
void arm_CPU_execVCVTFromInt(arm_CPU* c,uint32_t w,bool single){
{
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t bits = arm_CPU_sBits(c,arm_vmOf(w,true));
double f={};
if (signed_) {
f = cast<double>(cast<int32_t>(bits));
}
else {
f = cast<double>(bits);
}
if (single) {
arm_CPU_sSet(c,arm_vdOf(w,true),cast<float>(f));
}
else {
arm_CPU_dSet(c,arm_vdOf(w,false),f);
}
}
}
// tools/cpu/arm/vfp.go:620:1
void arm_CPU_execVCVTToInt(arm_CPU* c,uint32_t w,bool single){
{
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
double f={};
if (single) {
f = cast<double>(arm_CPU_sGet(c,arm_vmOf(w,true)));
}
else {
f = arm_CPU_dGet(c,arm_vmOf(w,false));
}
uint32_t out={};
if (signed_) {
out = cast<uint32_t>(cast<int32_t>(f));
}
else {
if ((f < cast<double>(0ULL))) {
f = cast<double>(0ULL);
}
out = cast<uint32_t>(f);
}
arm_CPU_sSetBits(c,arm_vdOf(w,true),out);
}
}
// tools/cpu/arm/vfp.go:641:1
bool arm_CPU_execVFPMove(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(7ULL))) {
uint32_t sysreg = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
if ((l == cast<uint32_t>(1ULL))) {
if ((sysreg == cast<uint32_t>(1ULL))) {
if ((rt == cast<uint32_t>(15ULL))) {
c->N = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
c->Z = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL));
c->C = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
c->V = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
}
else {
arm_CPU_setReg(c,rt,c->VFP.FPSCR);
}
}
else if ((sysreg == cast<uint32_t>(0ULL))) {
arm_CPU_setReg(c,rt,cast<uint32_t>(1090592948ULL));
}
else {
arm_CPU_setReg(c,rt,cast<uint32_t>(0ULL));
}
}
else {
if ((sysreg == cast<uint32_t>(1ULL))) {
c->VFP.FPSCR = arm_CPU_reg(c,rt);
}
}
return true;
}
uint32_t vn = arm_sReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,arm_CPU_sBits(c,vn));
}
else {
arm_CPU_sSetBits(c,vn,arm_CPU_reg(c,rt));
}
return true;
}
}
// tools/cpu/arm/vfp.go:683:1
bool arm_CPU_execVFPMove64(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rt2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t mBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t vm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if (single) {
uint32_t sm = arm_sReg(vm,mBit);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,arm_CPU_sBits(c,sm));
arm_CPU_setReg(c,rt2,arm_CPU_sBits(c,cast<uint32_t>((sm + cast<uint32_t>(1ULL)))));
}
else {
arm_CPU_sSetBits(c,sm,arm_CPU_reg(c,rt));
arm_CPU_sSetBits(c,cast<uint32_t>((sm + cast<uint32_t>(1ULL))),arm_CPU_reg(c,rt2));
}
}
else {
uint32_t dm = arm_dReg(vm,mBit);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dm))]);
arm_CPU_setReg(c,rt2,c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dm)) + cast<uint32_t>(1ULL)))]);
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dm))] = arm_CPU_reg(c,rt);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dm)) + cast<uint32_t>(1ULL)))] = arm_CPU_reg(c,rt2);
}
}
return true;
}
}
// tools/cpu/arm/vfp.go:723:1
uint64_t arm_vfpExpandImm(uint32_t w,bool single){
{
uint32_t imm8 = cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))),cast<int64_t>(4ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
uint32_t sign = cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
if (single) {
uint32_t exp = cast<uint32_t>((cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((cast<uint32_t>(~b) & cast<uint32_t>(1ULL)))),cast<int64_t>(7ULL))) | (shl<uint32_t>(arm_repeat(b,cast<uint32_t>(5ULL)),cast<int64_t>(2ULL))))) | (cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL))))));
uint32_t frac = cast<uint32_t>((imm8 & cast<uint32_t>(15ULL)));
return cast<uint64_t>(cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(sign,cast<int64_t>(31ULL)) | shl<uint32_t>(exp,cast<int64_t>(23ULL)))) | shl<uint32_t>(frac,cast<int64_t>(19ULL)))));
}
uint64_t exp = (cast<uint64_t>((cast<uint64_t>(((shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((cast<uint32_t>(~b) & cast<uint32_t>(1ULL)))),cast<int64_t>(10ULL))) | (shl<uint64_t>(arm_repeat64(cast<uint64_t>(b),cast<uint32_t>(8ULL)),cast<int64_t>(2ULL))))) | cast<uint64_t>(cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL)))))));
uint64_t frac = cast<uint64_t>(cast<uint32_t>((imm8 & cast<uint32_t>(15ULL))));
return cast<uint64_t>((cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(sign),cast<int64_t>(63ULL)) | shl<uint64_t>(exp,cast<int64_t>(52ULL)))) | shl<uint64_t>(frac,cast<int64_t>(48ULL))));
}
}
// tools/cpu/arm/vfp.go:738:1
uint32_t arm_repeat(uint32_t bit,uint32_t n){
{
uint32_t v={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v = cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(1ULL)) | (cast<uint32_t>((bit & cast<uint32_t>(1ULL))))));
}
}return v;
}
}
// tools/cpu/arm/vfp.go:745:1
uint64_t arm_repeat64(uint64_t bit,uint32_t n){
{
uint64_t v={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v = cast<uint64_t>((shl<uint64_t>(v,cast<int64_t>(1ULL)) | (cast<uint64_t>((bit & cast<uint64_t>(1ULL))))));
}
}return v;
}
}
// tools/cpu/sh4/cpu.go:149:1
void sh4_CPU_Reset(sh4_CPU* c){
{
(*c) = sh4_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},c->bus,c->fetcher,{},{},{},{},{},{}};
c->SR = cast<uint32_t>(1879048432ULL);
c->FPSCR = cast<uint32_t>(262145ULL);
c->PC = cast<uint32_t>(2684354560ULL);
c->nextPC = cast<uint32_t>((c->PC + cast<uint32_t>(2ULL)));
c->EXPEVT = cast<uint32_t>(0ULL);
c->onchipRaw = Map<uint32_t,uint32_t>{};
c->onchipGaps = Map<uint32_t,int64_t>{};
}
}
// tools/cpu/sh4/cpu.go:172:1
uint32_t sh4_CPU_CurPC(sh4_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/sh4/cpu.go:175:1
void sh4_CPU_SetPC(sh4_CPU* c,uint32_t pc){
{
c->PC = pc;
c->nextPC = cast<uint32_t>((pc + cast<uint32_t>(2ULL)));
auto tmp1 = std::make_tuple(false,false);
c->pendingDelay = std::get<0>(tmp1);
c->delaySlot = std::get<1>(tmp1);
}
}
// tools/cpu/sh4/cpu.go:182:1
uint32_t sh4_CPU_Reg(sh4_CPU* c,uint32_t i){
{
return c->R[cast<uint32_t>((i & cast<uint32_t>(15ULL)))];
}
}
// tools/cpu/sh4/cpu.go:183:1
void sh4_CPU_SetReg(sh4_CPU* c,uint32_t i,uint32_t v){
{
c->R[cast<uint32_t>((i & cast<uint32_t>(15ULL)))] = v;
}
}
// tools/cpu/sh4/cpu.go:186:1
bool sh4_CPU_InDelaySlot(sh4_CPU* c){
{
return c->delaySlot;
}
}
// tools/cpu/sh4/cpu.go:191:1
bool sh4_CPU_NextIsDelaySlot(sh4_CPU* c){
{
return c->pendingDelay;
}
}
// tools/cpu/sh4/cpu.go:195:1
uint32_t sh4_bankSelect(uint32_t sr){
{
if (((cast<uint32_t>((sr & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((sr & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL)))) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/sh4/cpu.go:204:1
void sh4_CPU_SetSR(sh4_CPU* c,uint32_t v){
{
v &= cast<uint32_t>(1879081971ULL);
if ((sh4_bankSelect(v) != sh4_bankSelect(c->SR))) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp2 = std::make_tuple(c->Rbank[i],c->R[i]);
c->R[i] = std::get<0>(tmp2);
c->Rbank[i] = std::get<1>(tmp2);
}
}}
c->SR = v;
}
}
// tools/cpu/sh4/cpu.go:216:1
void sh4_CPU_SetFPSCR(sh4_CPU* c,uint32_t v){
{
c->FPSCR = cast<uint32_t>((v & cast<uint32_t>(4194303ULL)));
}
}
// tools/cpu/sh4/cpu.go:219:1
uint32_t sh4_CPU_T(sh4_CPU* c){
{
return cast<uint32_t>((c->SR & cast<uint32_t>(1ULL)));
}
}
// tools/cpu/sh4/cpu.go:222:1
void sh4_CPU_setT(sh4_CPU* c,bool cond){
{
if (cond) {
c->SR |= cast<uint32_t>(1ULL);
}
else {
c->SR &= ~(cast<uint32_t>(1ULL));
}
}
}
// tools/cpu/sh4/cpu.go:234:1
void sh4_CPU_SetIRL(sh4_CPU* c,uint32_t level,uint32_t code){
{
auto tmp3 = std::make_tuple(cast<uint32_t>((level & cast<uint32_t>(15ULL))),code);
c->irlLevel = std::get<0>(tmp3);
c->irlCode = std::get<1>(tmp3);
}
}
// tools/cpu/sh4/cpu.go:260:1
bool sh4_CPU_cacheArray(sh4_CPU* c,uint32_t addr,bool write){
{
if (((addr < cast<uint32_t>(4026531840ULL)) || (addr >= cast<uint32_t>(4227858432ULL)))) {
return false;
}
c->onchipGaps[cast<uint32_t>((addr & cast<uint32_t>(4278190080ULL)))]++;
return true;
}
}
// tools/cpu/sh4/cpu.go:268:1
uint8_t sh4_CPU_read8(sh4_CPU* c,uint32_t addr){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
return dc_Machine_Read8(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))));
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
return cast<uint8_t>(shr<uint32_t>(sh4_CPU_sqRead32(c,addr),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((addr & cast<uint32_t>(3ULL)))))))));
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
return cast<uint8_t>(sh4_CPU_onchipRead(c,addr,cast<int64_t>(1ULL)));
}
else if (sh4_CPU_cacheArray(c,addr,false)){
return cast<uint8_t>(0ULL);
}
}
tmp4:;
sh4_CPU_Halt(c,std::string("read8 from unmapped P4 at %08X (PC %08X)",40),addr,c->curPC);
return cast<uint8_t>(0ULL);
}
}
// tools/cpu/sh4/cpu.go:283:1
uint16_t sh4_CPU_read16(sh4_CPU* c,uint32_t addr){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
return dc_Machine_Read16(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))));
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
return cast<uint16_t>(shr<uint32_t>(sh4_CPU_sqRead32(c,addr),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((addr & cast<uint32_t>(2ULL)))))))));
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
return cast<uint16_t>(sh4_CPU_onchipRead(c,addr,cast<int64_t>(2ULL)));
}
else if (sh4_CPU_cacheArray(c,addr,false)){
return cast<uint16_t>(0ULL);
}
}
tmp5:;
sh4_CPU_Halt(c,std::string("read16 from unmapped P4 at %08X (PC %08X)",41),addr,c->curPC);
return cast<uint16_t>(0ULL);
}
}
// tools/cpu/sh4/cpu.go:298:1
uint32_t sh4_CPU_read32(sh4_CPU* c,uint32_t addr){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
return dc_Machine_Read32(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))));
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
return sh4_CPU_sqRead32(c,addr);
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
return sh4_CPU_onchipRead(c,addr,cast<int64_t>(4ULL));
}
else if (sh4_CPU_cacheArray(c,addr,false)){
return cast<uint32_t>(0ULL);
}
}
tmp6:;
sh4_CPU_Halt(c,std::string("read32 from unmapped P4 at %08X (PC %08X)",41),addr,c->curPC);
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/sh4/cpu.go:313:1
void sh4_CPU_write8(sh4_CPU* c,uint32_t addr,uint8_t v){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
dc_Machine_Write8(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))),v);
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
sh4_CPU_sqWrite(c,addr,cast<uint32_t>(1ULL),cast<uint32_t>(v));
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
sh4_CPU_onchipWrite(c,addr,cast<int64_t>(1ULL),cast<uint32_t>(v));
}
else if (sh4_CPU_cacheArray(c,addr,true)){
}
else {
sh4_CPU_Halt(c,std::string("write8 to unmapped P4 at %08X (PC %08X)",39),addr,c->curPC);
}
}
tmp7:;
}
}
// tools/cpu/sh4/cpu.go:327:1
void sh4_CPU_write16(sh4_CPU* c,uint32_t addr,uint16_t v){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
dc_Machine_Write16(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))),v);
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
sh4_CPU_sqWrite(c,addr,cast<uint32_t>(2ULL),cast<uint32_t>(v));
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
sh4_CPU_onchipWrite(c,addr,cast<int64_t>(2ULL),cast<uint32_t>(v));
}
else if (sh4_CPU_cacheArray(c,addr,true)){
}
else {
sh4_CPU_Halt(c,std::string("write16 to unmapped P4 at %08X (PC %08X)",40),addr,c->curPC);
}
}
tmp8:;
}
}
// tools/cpu/sh4/cpu.go:341:1
void sh4_CPU_write32(sh4_CPU* c,uint32_t addr,uint32_t v){
{
{
if ((addr < cast<uint32_t>(3758096384ULL))){
dc_Machine_Write32(c->bus,cast<uint32_t>((addr & cast<uint32_t>(536870911ULL))),v);
}
else if ((addr < cast<uint32_t>(3825205248ULL))){
sh4_CPU_sqWrite(c,addr,cast<uint32_t>(4ULL),v);
}
else if ((addr >= cast<uint32_t>(4227858432ULL))){
sh4_CPU_onchipWrite(c,addr,cast<int64_t>(4ULL),v);
}
else if (sh4_CPU_cacheArray(c,addr,true)){
}
else {
sh4_CPU_Halt(c,std::string("write32 to unmapped P4 at %08X (PC %08X)",40),addr,c->curPC);
}
}
tmp9:;
}
}
// tools/cpu/sh4/decode.go:27:1
sh4_Inst sh4_Decode(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(2ULL))) {
return sh4_Inst{addr,{},cast<int64_t>(0ULL),std::string(".hword",6),std::string(".hword ; truncated",18),cast<sh4_Flow>(7ULL),{},{},{},{},{}};
}
return sh4_DecodeHalfword(le_Uint16(code),addr);
}
}
// tools/cpu/sh4/decode.go:35:1
sh4_Inst sh4_DecodeHalfword(uint16_t h,uint32_t addr){
{
sh4_Inst in = sh4_Inst{addr,h,cast<int64_t>(2ULL),{},{},cast<sh4_Flow>(0ULL),{},{},{},{},{}};
sh4_decode((&in),h,addr);
if ((in.Mnem == std::string("",0))) {
in.Mnem = std::string(".hword",6);
in.Text = go_fmt_Sprintf(std::string(".hword 0x%04X",13),h);
in.Flow = cast<sh4_Flow>(7ULL);
}
return in;
}
}
// tools/cpu/sh4/decode.go:56:1
std::string sh4_r(uint32_t n){
{
return go_fmt_Sprintf(std::string("r%d",3),n);
}
}
// tools/cpu/sh4/decode.go:60:1
std::string sh4_bank(uint32_t m){
{
return go_fmt_Sprintf(std::string("r%d_bank",8),cast<uint32_t>((m & cast<uint32_t>(7ULL))));
}
}
// tools/cpu/sh4/decode.go:63:1
uint32_t sh4_rn(uint16_t h){
{
return cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
}
}
// tools/cpu/sh4/decode.go:64:1
uint32_t sh4_rm(uint16_t h){
{
return cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(h,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)));
}
}
// tools/cpu/sh4/decode.go:67:1
int32_t sh4_s8(uint16_t h){
{
return cast<int32_t>(cast<int8_t>(h));
}
}
// tools/cpu/sh4/decode.go:68:1
int32_t sh4_s12(uint16_t h){
{
return shr<int32_t>(cast<int32_t>(cast<int16_t>(shl<uint16_t>(h,cast<int64_t>(4ULL)))),cast<int64_t>(4ULL));
}
}
// tools/cpu/sh4/decode.go:72:1
void sh4_Inst_branch(sh4_Inst* in,sh4_Flow flow,int32_t disp,bool delayed){
{
in->Flow = flow;
in->Target = cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((in->Addr + cast<uint32_t>(4ULL)))) + disp)));
in->HasTarget = true;
in->HasDelay = delayed;
}
}
// tools/cpu/sh4/decode.go:79:1
void sh4_decode(sh4_Inst* in,uint16_t h,uint32_t addr){
{
{
switch(shr<uint16_t>(h,cast<int64_t>(12ULL))){
case cast<uint16_t>(0ULL):{
sh4_decode0(in,h);
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("%s, @(0x%X,%s)",14),sh4_r(sh4_rm(h)),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(4ULL))),sh4_r(sh4_rn(h)));
break;}
case cast<uint16_t>(2ULL):{
sh4_decode2(in,h);
break;}
case cast<uint16_t>(3ULL):{
sh4_decode3(in,h);
break;}
case cast<uint16_t>(4ULL):{
sh4_decode4(in,h);
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@(0x%X,%s), %s",14),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(4ULL))),sh4_r(sh4_rm(h)),sh4_r(sh4_rn(h)));
break;}
case cast<uint16_t>(6ULL):{
sh4_decode6(in,h);
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("add",3),std::string("#%d, %s",7),sh4_s8(h),sh4_r(sh4_rn(h)));
break;}
case cast<uint16_t>(8ULL):{
sh4_decode8(in,h);
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@(0x%X,pc), %s",14),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL)))) * cast<uint32_t>(2ULL))),sh4_r(sh4_rn(h)));
auto tmp10 = std::make_tuple(sh4_litW(addr,cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL))))),cast<int64_t>(2ULL));
in->LitAddr = std::get<0>(tmp10);
in->LitSize = std::get<1>(tmp10);
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("bra",3),std::string("",0));
sh4_Inst_branch(in,cast<sh4_Flow>(2ULL),cast<int32_t>((sh4_s12(h) * cast<int32_t>(2ULL))),true);
in->Text = go_fmt_Sprintf(std::string("bra 0x%08X",10),in->Target);
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("bsr",3),std::string("",0));
sh4_Inst_branch(in,cast<sh4_Flow>(3ULL),cast<int32_t>((sh4_s12(h) * cast<int32_t>(2ULL))),true);
in->Text = go_fmt_Sprintf(std::string("bsr 0x%08X",10),in->Target);
break;}
case cast<uint16_t>(12ULL):{
sh4_decodeC(in,h,addr);
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@(0x%X,pc), %s",14),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL)))) * cast<uint32_t>(4ULL))),sh4_r(sh4_rn(h)));
auto tmp11 = std::make_tuple(sh4_litL(addr,cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL))))),cast<int64_t>(4ULL));
in->LitAddr = std::get<0>(tmp11);
in->LitSize = std::get<1>(tmp11);
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("mov",3),std::string("#%d, %s",7),sh4_s8(h),sh4_r(sh4_rn(h)));
break;}
case cast<uint16_t>(15ULL):{
sh4_decodeFPU(in,h);
break;}
}}
}
}
// tools/cpu/sh4/decode.go:125:1
void sh4_decode0(sh4_Inst* in,uint16_t h){
{
auto tmp12 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp12);
uint32_t m = std::get<1>(tmp12);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(2ULL):{
{
if ((m == cast<uint32_t>(0ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("sr, %s",6),sh4_r(n));
}
else if ((m == cast<uint32_t>(1ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("gbr, %s",7),sh4_r(n));
}
else if ((m == cast<uint32_t>(2ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("vbr, %s",7),sh4_r(n));
}
else if ((m == cast<uint32_t>(3ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("ssr, %s",7),sh4_r(n));
}
else if ((m == cast<uint32_t>(4ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("spc, %s",7),sh4_r(n));
}
else if ((m >= cast<uint32_t>(8ULL))){
sh4_Inst_set(in,std::string("stc",3),std::string("%s, %s",6),sh4_bank(m),sh4_r(n));
}
}
tmp13:;
break;}
case cast<uint16_t>(3ULL):{
{
switch(m){
case cast<uint32_t>(0ULL):{
sh4_Inst_set(in,std::string("bsrf",4),std::string("%s",2),sh4_r(n));
auto tmp14 = std::make_tuple(cast<sh4_Flow>(6ULL),true);
in->Flow = std::get<0>(tmp14);
in->HasDelay = std::get<1>(tmp14);
break;}
case cast<uint32_t>(2ULL):{
sh4_Inst_set(in,std::string("braf",4),std::string("%s",2),sh4_r(n));
auto tmp15 = std::make_tuple(cast<sh4_Flow>(5ULL),true);
in->Flow = std::get<0>(tmp15);
in->HasDelay = std::get<1>(tmp15);
break;}
case cast<uint32_t>(8ULL):{
sh4_Inst_set(in,std::string("pref",4),std::string("@%s",3),sh4_r(n));
break;}
case cast<uint32_t>(9ULL):{
sh4_Inst_set(in,std::string("ocbi",4),std::string("@%s",3),sh4_r(n));
break;}
case cast<uint32_t>(10ULL):{
sh4_Inst_set(in,std::string("ocbp",4),std::string("@%s",3),sh4_r(n));
break;}
case cast<uint32_t>(11ULL):{
sh4_Inst_set(in,std::string("ocbwb",5),std::string("@%s",3),sh4_r(n));
break;}
case cast<uint32_t>(12ULL):{
sh4_Inst_set(in,std::string("movca.l",7),std::string("r0, @%s",7),sh4_r(n));
break;}
}}
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("%s, @(r0,%s)",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("%s, @(r0,%s)",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("%s, @(r0,%s)",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("mul.l",5),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
if ((n != cast<uint32_t>(0ULL))) {
return ;
}
{
switch(m){
case cast<uint32_t>(0ULL):{
sh4_Inst_set(in,std::string("clrt",4),std::string("",0));
break;}
case cast<uint32_t>(1ULL):{
sh4_Inst_set(in,std::string("sett",4),std::string("",0));
break;}
case cast<uint32_t>(2ULL):{
sh4_Inst_set(in,std::string("clrmac",6),std::string("",0));
break;}
case cast<uint32_t>(3ULL):{
sh4_Inst_set(in,std::string("ldtlb",5),std::string("",0));
break;}
case cast<uint32_t>(4ULL):{
sh4_Inst_set(in,std::string("clrs",4),std::string("",0));
break;}
case cast<uint32_t>(5ULL):{
sh4_Inst_set(in,std::string("sets",4),std::string("",0));
break;}
}}
break;}
case cast<uint16_t>(9ULL):{
{
if ((h == cast<uint16_t>(9ULL))){
sh4_Inst_set(in,std::string("nop",3),std::string("",0));
}
else if ((h == cast<uint16_t>(25ULL))){
sh4_Inst_set(in,std::string("div0u",5),std::string("",0));
}
else if ((m == cast<uint32_t>(2ULL))){
sh4_Inst_set(in,std::string("movt",4),std::string("%s",2),sh4_r(n));
}
}
tmp16:;
break;}
case cast<uint16_t>(10ULL):{
{
switch(m){
case cast<uint32_t>(0ULL):{
sh4_Inst_set(in,std::string("sts",3),std::string("mach, %s",8),sh4_r(n));
break;}
case cast<uint32_t>(1ULL):{
sh4_Inst_set(in,std::string("sts",3),std::string("macl, %s",8),sh4_r(n));
break;}
case cast<uint32_t>(2ULL):{
sh4_Inst_set(in,std::string("sts",3),std::string("pr, %s",6),sh4_r(n));
break;}
case cast<uint32_t>(3ULL):{
sh4_Inst_set(in,std::string("stc",3),std::string("sgr, %s",7),sh4_r(n));
break;}
case cast<uint32_t>(5ULL):{
sh4_Inst_set(in,std::string("sts",3),std::string("fpul, %s",8),sh4_r(n));
break;}
case cast<uint32_t>(6ULL):{
sh4_Inst_set(in,std::string("sts",3),std::string("fpscr, %s",9),sh4_r(n));
break;}
case cast<uint32_t>(15ULL):{
sh4_Inst_set(in,std::string("stc",3),std::string("dbr, %s",7),sh4_r(n));
break;}
}}
break;}
case cast<uint16_t>(11ULL):{
{
switch(h){
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("rts",3),std::string("",0));
auto tmp17 = std::make_tuple(cast<sh4_Flow>(4ULL),true);
in->Flow = std::get<0>(tmp17);
in->HasDelay = std::get<1>(tmp17);
break;}
case cast<uint16_t>(27ULL):{
sh4_Inst_set(in,std::string("sleep",5),std::string("",0));
in->Flow = cast<sh4_Flow>(7ULL);
break;}
case cast<uint16_t>(43ULL):{
sh4_Inst_set(in,std::string("rte",3),std::string("",0));
auto tmp18 = std::make_tuple(cast<sh4_Flow>(7ULL),true);
in->Flow = std::get<0>(tmp18);
in->HasDelay = std::get<1>(tmp18);
break;}
}}
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("@(r0,%s), %s",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@(r0,%s), %s",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@(r0,%s), %s",12),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("mac.l",5),std::string("@%s+, @%s+",10),sh4_r(m),sh4_r(n));
break;}
}}
}
}
// tools/cpu/sh4/decode.go:239:1
void sh4_decode2(sh4_Inst* in,uint16_t h){
{
auto tmp19 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp19);
uint32_t m = std::get<1>(tmp19);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("%s, @%s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("%s, @%s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("%s, @%s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("%s, @-%s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("%s, @-%s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("%s, @-%s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("div0s",5),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("tst",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("and",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("xor",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("or",2),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("cmp/str",7),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("xtrct",5),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("mulu.w",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("muls.w",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
}}
}
}
// tools/cpu/sh4/decode.go:277:1
void sh4_decode3(sh4_Inst* in,uint16_t h){
{
auto tmp20 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp20);
uint32_t m = std::get<1>(tmp20);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("cmp/eq",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("cmp/hs",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(3ULL):{
sh4_Inst_set(in,std::string("cmp/ge",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("div1",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("dmulu.l",7),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("cmp/hi",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("cmp/gt",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("sub",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("subc",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("subv",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("add",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("dmuls.l",7),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("addc",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("addv",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
}}
}
}
// tools/cpu/sh4/decode.go:314:1
void sh4_decode4(sh4_Inst* in,uint16_t h){
{
auto tmp21 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp21);
uint32_t m = std::get<1>(tmp21);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("shad",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
return ;
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("shld",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
return ;
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("mac.w",5),std::string("@%s+, @%s+",10),sh4_r(m),sh4_r(n));
return ;
break;}
case cast<uint16_t>(3ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
sh4_Inst_set(in,std::string("stc.l",5),std::string("%s, @-%s",8),sh4_bank(m),sh4_r(n));
return ;
}
break;}
case cast<uint16_t>(7ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, %s",8),sh4_r(n),sh4_bank(m));
return ;
}
break;}
case cast<uint16_t>(14ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, %s",6),sh4_r(n),sh4_bank(m));
return ;
}
break;}
}}
{
switch(cast<uint16_t>((h & cast<uint16_t>(255ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("shll",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("shlr",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("sts.l",5),std::string("mach, @-%s",10),sh4_r(n));
break;}
case cast<uint16_t>(3ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("sr, @-%s",8),sh4_r(n));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("rotl",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("rotr",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("lds.l",5),std::string("@%s+, mach",10),sh4_r(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, sr",8),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("shll2",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("shlr2",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("lds",3),std::string("%s, mach",8),sh4_r(n));
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("jsr",3),std::string("@%s",3),sh4_r(n));
auto tmp22 = std::make_tuple(cast<sh4_Flow>(6ULL),true);
in->Flow = std::get<0>(tmp22);
in->HasDelay = std::get<1>(tmp22);
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, sr",6),sh4_r(n));
break;}
case cast<uint16_t>(16ULL):{
sh4_Inst_set(in,std::string("dt",2),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(17ULL):{
sh4_Inst_set(in,std::string("cmp/pz",6),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(18ULL):{
sh4_Inst_set(in,std::string("sts.l",5),std::string("macl, @-%s",10),sh4_r(n));
break;}
case cast<uint16_t>(19ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("gbr, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(21ULL):{
sh4_Inst_set(in,std::string("cmp/pl",6),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(22ULL):{
sh4_Inst_set(in,std::string("lds.l",5),std::string("@%s+, macl",10),sh4_r(n));
break;}
case cast<uint16_t>(23ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, gbr",9),sh4_r(n));
break;}
case cast<uint16_t>(24ULL):{
sh4_Inst_set(in,std::string("shll8",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(25ULL):{
sh4_Inst_set(in,std::string("shlr8",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(26ULL):{
sh4_Inst_set(in,std::string("lds",3),std::string("%s, macl",8),sh4_r(n));
break;}
case cast<uint16_t>(27ULL):{
sh4_Inst_set(in,std::string("tas.b",5),std::string("@%s",3),sh4_r(n));
break;}
case cast<uint16_t>(30ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, gbr",7),sh4_r(n));
break;}
case cast<uint16_t>(32ULL):{
sh4_Inst_set(in,std::string("shal",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(33ULL):{
sh4_Inst_set(in,std::string("shar",4),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(34ULL):{
sh4_Inst_set(in,std::string("sts.l",5),std::string("pr, @-%s",8),sh4_r(n));
break;}
case cast<uint16_t>(35ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("vbr, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(36ULL):{
sh4_Inst_set(in,std::string("rotcl",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(37ULL):{
sh4_Inst_set(in,std::string("rotcr",5),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(38ULL):{
sh4_Inst_set(in,std::string("lds.l",5),std::string("@%s+, pr",8),sh4_r(n));
break;}
case cast<uint16_t>(39ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, vbr",9),sh4_r(n));
break;}
case cast<uint16_t>(40ULL):{
sh4_Inst_set(in,std::string("shll16",6),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(41ULL):{
sh4_Inst_set(in,std::string("shlr16",6),std::string("%s",2),sh4_r(n));
break;}
case cast<uint16_t>(42ULL):{
sh4_Inst_set(in,std::string("lds",3),std::string("%s, pr",6),sh4_r(n));
break;}
case cast<uint16_t>(43ULL):{
sh4_Inst_set(in,std::string("jmp",3),std::string("@%s",3),sh4_r(n));
auto tmp23 = std::make_tuple(cast<sh4_Flow>(5ULL),true);
in->Flow = std::get<0>(tmp23);
in->HasDelay = std::get<1>(tmp23);
break;}
case cast<uint16_t>(46ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, vbr",7),sh4_r(n));
break;}
case cast<uint16_t>(50ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("sgr, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(51ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("ssr, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(55ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, ssr",9),sh4_r(n));
break;}
case cast<uint16_t>(62ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, ssr",7),sh4_r(n));
break;}
case cast<uint16_t>(67ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("spc, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(71ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, spc",9),sh4_r(n));
break;}
case cast<uint16_t>(78ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, spc",7),sh4_r(n));
break;}
case cast<uint16_t>(82ULL):{
sh4_Inst_set(in,std::string("sts.l",5),std::string("fpul, @-%s",10),sh4_r(n));
break;}
case cast<uint16_t>(86ULL):{
sh4_Inst_set(in,std::string("lds.l",5),std::string("@%s+, fpul",10),sh4_r(n));
break;}
case cast<uint16_t>(90ULL):{
sh4_Inst_set(in,std::string("lds",3),std::string("%s, fpul",8),sh4_r(n));
break;}
case cast<uint16_t>(98ULL):{
sh4_Inst_set(in,std::string("sts.l",5),std::string("fpscr, @-%s",11),sh4_r(n));
break;}
case cast<uint16_t>(102ULL):{
sh4_Inst_set(in,std::string("lds.l",5),std::string("@%s+, fpscr",11),sh4_r(n));
break;}
case cast<uint16_t>(106ULL):{
sh4_Inst_set(in,std::string("lds",3),std::string("%s, fpscr",9),sh4_r(n));
break;}
case cast<uint16_t>(242ULL):{
sh4_Inst_set(in,std::string("stc.l",5),std::string("dbr, @-%s",9),sh4_r(n));
break;}
case cast<uint16_t>(246ULL):{
sh4_Inst_set(in,std::string("ldc.l",5),std::string("@%s+, dbr",9),sh4_r(n));
break;}
case cast<uint16_t>(250ULL):{
sh4_Inst_set(in,std::string("ldc",3),std::string("%s, dbr",7),sh4_r(n));
break;}
}}
}
}
// tools/cpu/sh4/decode.go:458:1
void sh4_decode6(sh4_Inst* in,uint16_t h){
{
auto tmp24 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp24);
uint32_t m = std::get<1>(tmp24);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("@%s, %s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@%s, %s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@%s, %s",7),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(3ULL):{
sh4_Inst_set(in,std::string("mov",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("@%s+, %s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@%s+, %s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@%s+, %s",8),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("not",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("swap.b",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("swap.w",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("negc",4),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("neg",3),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("extu.b",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("extu.w",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("exts.b",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("exts.w",6),std::string("%s, %s",6),sh4_r(m),sh4_r(n));
break;}
}}
}
}
// tools/cpu/sh4/decode.go:499:1
void sh4_decode8(sh4_Inst* in,uint16_t h){
{
uint32_t m = sh4_rm(h);
{
switch(cast<uint16_t>(((shr<uint16_t>(h,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("r0, @(0x%X,%s)",14),cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))),sh4_r(m));
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("r0, @(0x%X,%s)",14),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(2ULL))),sh4_r(m));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("@(0x%X,%s), r0",14),cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))),sh4_r(m));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@(0x%X,%s), r0",14),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(2ULL))),sh4_r(m));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("cmp/eq",6),std::string("#%d, r0",7),sh4_s8(h));
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_branch(in,cast<sh4_Flow>(1ULL),cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL))),false);
sh4_Inst_set(in,std::string("bt",2),std::string("0x%08X",6),in->Target);
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_branch(in,cast<sh4_Flow>(1ULL),cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL))),false);
sh4_Inst_set(in,std::string("bf",2),std::string("0x%08X",6),in->Target);
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_branch(in,cast<sh4_Flow>(1ULL),cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL))),true);
sh4_Inst_set(in,std::string("bt/s",4),std::string("0x%08X",6),in->Target);
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_branch(in,cast<sh4_Flow>(1ULL),cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL))),true);
sh4_Inst_set(in,std::string("bf/s",4),std::string("0x%08X",6),in->Target);
break;}
}}
}
}
// tools/cpu/sh4/decode.go:529:1
void sh4_decodeC(sh4_Inst* in,uint16_t h,uint32_t addr){
{
uint32_t d = cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL))));
{
switch(cast<uint16_t>(((shr<uint16_t>(h,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("r0, @(0x%X,gbr)",15),d);
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("r0, @(0x%X,gbr)",15),cast<uint32_t>((d * cast<uint32_t>(2ULL))));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("r0, @(0x%X,gbr)",15),cast<uint32_t>((d * cast<uint32_t>(4ULL))));
break;}
case cast<uint16_t>(3ULL):{
sh4_Inst_set(in,std::string("trapa",5),std::string("#%d",3),d);
in->Flow = cast<sh4_Flow>(7ULL);
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("mov.b",5),std::string("@(0x%X,gbr), r0",15),d);
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("mov.w",5),std::string("@(0x%X,gbr), r0",15),cast<uint32_t>((d * cast<uint32_t>(2ULL))));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("mov.l",5),std::string("@(0x%X,gbr), r0",15),cast<uint32_t>((d * cast<uint32_t>(4ULL))));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("mova",4),std::string("@(0x%X,pc), r0",14),cast<uint32_t>((d * cast<uint32_t>(4ULL))));
auto tmp25 = std::make_tuple(sh4_litL(addr,d),cast<int64_t>(4ULL));
in->LitAddr = std::get<0>(tmp25);
in->LitSize = std::get<1>(tmp25);
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("tst",3),std::string("#%d, r0",7),d);
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("and",3),std::string("#%d, r0",7),d);
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("xor",3),std::string("#%d, r0",7),d);
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("or",2),std::string("#%d, r0",7),d);
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("tst.b",5),std::string("#%d, @(r0,gbr)",14),d);
break;}
case cast<uint16_t>(13ULL):{
sh4_Inst_set(in,std::string("and.b",5),std::string("#%d, @(r0,gbr)",14),d);
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("xor.b",5),std::string("#%d, @(r0,gbr)",14),d);
break;}
case cast<uint16_t>(15ULL):{
sh4_Inst_set(in,std::string("or.b",4),std::string("#%d, @(r0,gbr)",14),d);
break;}
}}
}
}
// tools/cpu/sh4/decode_fpu.go:15:1
std::string sh4_fr(uint32_t n){
{
return go_fmt_Sprintf(std::string("fr%d",4),n);
}
}
// tools/cpu/sh4/decode_fpu.go:16:1
std::string sh4_dr(uint32_t n){
{
return go_fmt_Sprintf(std::string("dr%d",4),n);
}
}
// tools/cpu/sh4/decode_fpu.go:17:1
std::string sh4_fv(uint32_t n){
{
return go_fmt_Sprintf(std::string("fv%d",4),cast<uint32_t>((n * cast<uint32_t>(4ULL))));
}
}
// tools/cpu/sh4/decode_fpu.go:19:1
void sh4_decodeFPU(sh4_Inst* in,uint16_t h){
{
auto tmp26 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp26);
uint32_t m = std::get<1>(tmp26);
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_Inst_set(in,std::string("fadd",4),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(1ULL):{
sh4_Inst_set(in,std::string("fsub",4),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(2ULL):{
sh4_Inst_set(in,std::string("fmul",4),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(3ULL):{
sh4_Inst_set(in,std::string("fdiv",4),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(4ULL):{
sh4_Inst_set(in,std::string("fcmp/eq",7),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(5ULL):{
sh4_Inst_set(in,std::string("fcmp/gt",7),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(6ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("@(r0,%s), %s",12),sh4_r(m),sh4_fr(n));
break;}
case cast<uint16_t>(7ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("%s, @(r0,%s)",12),sh4_fr(m),sh4_r(n));
break;}
case cast<uint16_t>(8ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("@%s, %s",7),sh4_r(m),sh4_fr(n));
break;}
case cast<uint16_t>(9ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("@%s+, %s",8),sh4_r(m),sh4_fr(n));
break;}
case cast<uint16_t>(10ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("%s, @%s",7),sh4_fr(m),sh4_r(n));
break;}
case cast<uint16_t>(11ULL):{
sh4_Inst_set(in,std::string("fmov.s",6),std::string("%s, @-%s",8),sh4_fr(m),sh4_r(n));
break;}
case cast<uint16_t>(12ULL):{
sh4_Inst_set(in,std::string("fmov",4),std::string("%s, %s",6),sh4_fr(m),sh4_fr(n));
break;}
case cast<uint16_t>(13ULL):{
sh4_decodeFPUxD(in,h,n);
break;}
case cast<uint16_t>(14ULL):{
sh4_Inst_set(in,std::string("fmac",4),std::string("fr0, %s, %s",11),sh4_fr(m),sh4_fr(n));
break;}
}}
}
}
// tools/cpu/sh4/decode_fpu.go:58:1
void sh4_decodeFPUxD(sh4_Inst* in,uint16_t h,uint32_t n){
{
{
switch(sh4_rm(h)){
case cast<uint32_t>(0ULL):{
sh4_Inst_set(in,std::string("fsts",4),std::string("fpul, %s",8),sh4_fr(n));
break;}
case cast<uint32_t>(1ULL):{
sh4_Inst_set(in,std::string("flds",4),std::string("%s, fpul",8),sh4_fr(n));
break;}
case cast<uint32_t>(2ULL):{
sh4_Inst_set(in,std::string("float",5),std::string("fpul, %s",8),sh4_fr(n));
break;}
case cast<uint32_t>(3ULL):{
sh4_Inst_set(in,std::string("ftrc",4),std::string("%s, fpul",8),sh4_fr(n));
break;}
case cast<uint32_t>(4ULL):{
sh4_Inst_set(in,std::string("fneg",4),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(5ULL):{
sh4_Inst_set(in,std::string("fabs",4),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(6ULL):{
sh4_Inst_set(in,std::string("fsqrt",5),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(7ULL):{
sh4_Inst_set(in,std::string("fsrra",5),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(8ULL):{
sh4_Inst_set(in,std::string("fldi0",5),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(9ULL):{
sh4_Inst_set(in,std::string("fldi1",5),std::string("%s",2),sh4_fr(n));
break;}
case cast<uint32_t>(10ULL):{
sh4_Inst_set(in,std::string("fcnvsd",6),std::string("fpul, %s",8),sh4_dr(n));
break;}
case cast<uint32_t>(11ULL):{
sh4_Inst_set(in,std::string("fcnvds",6),std::string("%s, fpul",8),sh4_dr(n));
break;}
case cast<uint32_t>(14ULL):{
sh4_Inst_set(in,std::string("fipr",4),std::string("%s, %s",6),sh4_fv(cast<uint32_t>((n & cast<uint32_t>(3ULL)))),sh4_fv(shr<uint32_t>(n,cast<int64_t>(2ULL))));
break;}
case cast<uint32_t>(15ULL):{
{
if ((h == cast<uint16_t>(64509ULL))){
sh4_Inst_set(in,std::string("frchg",5),std::string("",0));
}
else if ((h == cast<uint16_t>(62461ULL))){
sh4_Inst_set(in,std::string("fschg",5),std::string("",0));
}
else if ((cast<uint32_t>((n & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL))){
sh4_Inst_set(in,std::string("ftrv",4),std::string("xmtrx, %s",9),sh4_fv(shr<uint32_t>(n,cast<int64_t>(2ULL))));
}
else if ((cast<uint32_t>((n & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))){
sh4_Inst_set(in,std::string("fsca",4),std::string("fpul, %s",8),sh4_dr(n));
}
}
tmp27:;
break;}
}}
}
}
// tools/cpu/sh4/exec.go:17:1
int64_t sh4_CPU_Step(sh4_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
sh4_CPU_tickTMU(c);
if (sh4_CPU_checkInterrupt(c)) {
return cast<int64_t>(1ULL);
}
c->curPC = c->PC;
c->delaySlot = c->pendingDelay;
c->pendingDelay = false;
if ((cast<uint32_t>((c->PC & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
sh4_CPU_Halt(c,std::string("misaligned PC %08X",18),c->PC);
return cast<int64_t>(1ULL);
}
uint16_t h = sh4_CPU_fetchInstr(c,c->PC);
c->PC = c->nextPC;
c->nextPC += cast<uint32_t>(2ULL);
sh4_CPU_execute(c,h);
c->Steps++;
return cast<int64_t>(1ULL);
}
}
// tools/cpu/sh4/exec.go:49:1
bool sh4_CPU_checkInterrupt(sh4_CPU* c){
{
if ((c->pendingDelay || (cast<uint32_t>((c->SR & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL)))) {
return false;
}
auto tmp28 = std::make_tuple(c->irlLevel,c->irlCode);
uint32_t level = std::get<0>(tmp28);
uint32_t code = std::get<1>(tmp28);
{
auto tmp29 = sh4_CPU_tmuPending(c);
uint32_t l = std::get<0>(tmp29);
uint32_t cd = std::get<1>(tmp29);
bool ok = std::get<2>(tmp29);
if ((ok && (l > level))) {
auto tmp30 = std::make_tuple(l,cd);
level = std::get<0>(tmp30);
code = std::get<1>(tmp30);
}
}
if (((level == cast<uint32_t>(0ULL)) || (level <= cast<uint32_t>(((shr<uint32_t>(c->SR,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)))))) {
return false;
}
c->INTEVT = code;
c->SSR = c->SR;
c->SPC = c->PC;
c->SGR = c->R[cast<int64_t>(15ULL)];
sh4_CPU_SetSR(c,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((c->SR | cast<uint32_t>(1073741824ULL))) | cast<uint32_t>(536870912ULL))) | cast<uint32_t>(268435456ULL))));
c->PC = cast<uint32_t>((c->VBR + cast<uint32_t>(1536ULL)));
c->nextPC = cast<uint32_t>((c->PC + cast<uint32_t>(2ULL)));
return true;
}
}
// tools/cpu/sh4/exec.go:72:1
void sh4_CPU_exception(sh4_CPU* c,uint32_t expevt,uint32_t spc){
{
c->EXPEVT = expevt;
c->SSR = c->SR;
c->SPC = spc;
c->SGR = c->R[cast<int64_t>(15ULL)];
sh4_CPU_SetSR(c,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((c->SR | cast<uint32_t>(1073741824ULL))) | cast<uint32_t>(536870912ULL))) | cast<uint32_t>(268435456ULL))));
c->PC = cast<uint32_t>((c->VBR + cast<uint32_t>(256ULL)));
c->nextPC = cast<uint32_t>((c->PC + cast<uint32_t>(2ULL)));
}
}
// tools/cpu/sh4/exec.go:86:1
void sh4_CPU_doBranch(sh4_CPU* c,bool taken,uint32_t target){
{
if (c->delaySlot) {
sh4_CPU_Halt(c,std::string("branch in delay slot at %08X",28),c->curPC);
return ;
}
c->pendingDelay = true;
if (taken) {
c->nextPC = target;
}
}
}
// tools/cpu/sh4/exec.go:99:1
void sh4_CPU_doJumpNow(sh4_CPU* c,bool taken,uint32_t target){
{
if (c->delaySlot) {
sh4_CPU_Halt(c,std::string("branch in delay slot at %08X",28),c->curPC);
return ;
}
if (taken) {
c->PC = target;
c->nextPC = cast<uint32_t>((target + cast<uint32_t>(2ULL)));
}
}
}
// tools/cpu/sh4/exec.go:110:1
void sh4_CPU_execute(sh4_CPU* c,uint16_t h){
{
auto tmp31 = std::make_tuple(sh4_rn(h),sh4_rm(h));
uint32_t n = std::get<0>(tmp31);
uint32_t m = std::get<1>(tmp31);
{
switch(shr<uint16_t>(h,cast<int64_t>(12ULL))){
case cast<uint16_t>(0ULL):{
sh4_CPU_exec0(c,h,n,m);
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_write32(c,cast<uint32_t>((c->R[n] + cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(4ULL))))),c->R[m]);
break;}
case cast<uint16_t>(2ULL):{
sh4_CPU_exec2(c,h,n,m);
break;}
case cast<uint16_t>(3ULL):{
sh4_CPU_exec3(c,h,n,m);
break;}
case cast<uint16_t>(4ULL):{
sh4_CPU_exec4(c,h,n,m);
break;}
case cast<uint16_t>(5ULL):{
c->R[n] = sh4_CPU_read32(c,cast<uint32_t>((c->R[m] + cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(4ULL))))));
break;}
case cast<uint16_t>(6ULL):{
sh4_CPU_exec6(c,h,n,m);
break;}
case cast<uint16_t>(7ULL):{
c->R[n] += cast<uint32_t>(sh4_s8(h));
break;}
case cast<uint16_t>(8ULL):{
sh4_CPU_exec8(c,h);
break;}
case cast<uint16_t>(9ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,sh4_litW(c->curPC,cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL)))))))));
break;}
case cast<uint16_t>(10ULL):{
sh4_CPU_doBranch(c,true,cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s12(h) * cast<int32_t>(2ULL)))))));
break;}
case cast<uint16_t>(11ULL):{
c->PR = cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)));
sh4_CPU_doBranch(c,true,cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s12(h) * cast<int32_t>(2ULL)))))));
break;}
case cast<uint16_t>(12ULL):{
sh4_CPU_execC(c,h);
break;}
case cast<uint16_t>(13ULL):{
c->R[n] = sh4_CPU_read32(c,sh4_litL(c->curPC,cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL))))));
break;}
case cast<uint16_t>(14ULL):{
c->R[n] = cast<uint32_t>(sh4_s8(h));
break;}
case cast<uint16_t>(15ULL):{
sh4_CPU_execFPU(c,h,n,m);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:149:1
void sh4_CPU_exec0(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(2ULL):{
{
if ((m == cast<uint32_t>(0ULL))){
c->R[n] = c->SR;
}
else if ((m == cast<uint32_t>(1ULL))){
c->R[n] = c->GBR;
}
else if ((m == cast<uint32_t>(2ULL))){
c->R[n] = c->VBR;
}
else if ((m == cast<uint32_t>(3ULL))){
c->R[n] = c->SSR;
}
else if ((m == cast<uint32_t>(4ULL))){
c->R[n] = c->SPC;
}
else if ((m >= cast<uint32_t>(8ULL))){
c->R[n] = sh4_CPU_bankReg(c,m);
}
else {
sh4_CPU_unknown(c,h);
}
}
tmp32:;
break;}
case cast<uint16_t>(3ULL):{
{
switch(m){
case cast<uint32_t>(0ULL):{
c->PR = cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)));
sh4_CPU_doBranch(c,true,cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + c->R[n])));
break;}
case cast<uint32_t>(2ULL):{
sh4_CPU_doBranch(c,true,cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + c->R[n])));
break;}
case cast<uint32_t>(8ULL):{
{
uint32_t a = c->R[n];
if (((a >= cast<uint32_t>(3758096384ULL)) && (a < cast<uint32_t>(3825205248ULL)))) {
sh4_CPU_sqFlush(c,a);
}
}
break;}
case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):{
break;}
case cast<uint32_t>(12ULL):{
sh4_CPU_write32(c,c->R[n],c->R[cast<int64_t>(0ULL)]);
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
break;}
case cast<uint16_t>(4ULL):{
sh4_CPU_write8(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[n])),cast<uint8_t>(c->R[m]));
break;}
case cast<uint16_t>(5ULL):{
sh4_CPU_write16(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[n])),cast<uint16_t>(c->R[m]));
break;}
case cast<uint16_t>(6ULL):{
sh4_CPU_write32(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[n])),c->R[m]);
break;}
case cast<uint16_t>(7ULL):{
c->MACL = cast<uint32_t>((c->R[n] * c->R[m]));
break;}
case cast<uint16_t>(8ULL):{
if ((n != cast<uint32_t>(0ULL))) {
sh4_CPU_unknown(c,h);
return ;
}
{
switch(m){
case cast<uint32_t>(0ULL):{
sh4_CPU_setT(c,false);
break;}
case cast<uint32_t>(1ULL):{
sh4_CPU_setT(c,true);
break;}
case cast<uint32_t>(2ULL):{
auto tmp33 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->MACH = std::get<0>(tmp33);
c->MACL = std::get<1>(tmp33);
break;}
case cast<uint32_t>(3ULL):{
sh4_CPU_Halt(c,std::string("ldtlb with the MMU unmodelled (PC %08X)",39),c->curPC);
break;}
case cast<uint32_t>(4ULL):{
c->SR &= ~(cast<uint32_t>(2ULL));
break;}
case cast<uint32_t>(5ULL):{
c->SR |= cast<uint32_t>(2ULL);
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
break;}
case cast<uint16_t>(9ULL):{
{
if ((h == cast<uint16_t>(9ULL))){
}
else if ((h == cast<uint16_t>(25ULL))){
c->SR &= ~(cast<uint32_t>(769ULL));
}
else if ((m == cast<uint32_t>(2ULL))){
c->R[n] = sh4_CPU_T(c);
}
else {
sh4_CPU_unknown(c,h);
}
}
tmp34:;
break;}
case cast<uint16_t>(10ULL):{
{
switch(m){
case cast<uint32_t>(0ULL):{
c->R[n] = c->MACH;
break;}
case cast<uint32_t>(1ULL):{
c->R[n] = c->MACL;
break;}
case cast<uint32_t>(2ULL):{
c->R[n] = c->PR;
break;}
case cast<uint32_t>(3ULL):{
c->R[n] = c->SGR;
break;}
case cast<uint32_t>(5ULL):{
c->R[n] = c->FPUL;
break;}
case cast<uint32_t>(6ULL):{
c->R[n] = c->FPSCR;
break;}
case cast<uint32_t>(15ULL):{
c->R[n] = c->DBR;
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
break;}
case cast<uint16_t>(11ULL):{
{
switch(h){
case cast<uint16_t>(11ULL):{
sh4_CPU_doBranch(c,true,c->PR);
break;}
case cast<uint16_t>(27ULL):{
c->PC = c->curPC;
c->nextPC = cast<uint32_t>((c->curPC + cast<uint32_t>(2ULL)));
break;}
case cast<uint16_t>(43ULL):{
uint32_t target = c->SPC;
sh4_CPU_SetSR(c,c->SSR);
sh4_CPU_doBranch(c,true,target);
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
break;}
case cast<uint16_t>(12ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(sh4_CPU_read8(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[m]))))));
break;}
case cast<uint16_t>(13ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[m]))))));
break;}
case cast<uint16_t>(14ULL):{
c->R[n] = sh4_CPU_read32(c,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[m])));
break;}
case cast<uint16_t>(15ULL):{
sh4_CPU_macL(c,n,m);
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:270:1
void sh4_CPU_exec2(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_write8(c,c->R[n],cast<uint8_t>(c->R[m]));
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_write16(c,c->R[n],cast<uint16_t>(c->R[m]));
break;}
case cast<uint16_t>(2ULL):{
sh4_CPU_write32(c,c->R[n],c->R[m]);
break;}
case cast<uint16_t>(4ULL):{
c->R[n] -= cast<uint32_t>(1ULL);
sh4_CPU_write8(c,c->R[n],cast<uint8_t>(c->R[m]));
break;}
case cast<uint16_t>(5ULL):{
c->R[n] -= cast<uint32_t>(2ULL);
sh4_CPU_write16(c,c->R[n],cast<uint16_t>(c->R[m]));
break;}
case cast<uint16_t>(6ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->R[m]);
break;}
case cast<uint16_t>(7ULL):{
c->SR &= ~(cast<uint32_t>(768ULL));
if ((cast<uint32_t>((c->R[n] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
c->SR |= cast<uint32_t>(256ULL);
}
if ((cast<uint32_t>((c->R[m] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
c->SR |= cast<uint32_t>(512ULL);
}
sh4_CPU_setT(c,(((cast<uint32_t>((c->SR & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) != ((cast<uint32_t>((c->SR & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL)))));
break;}
case cast<uint16_t>(8ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & c->R[m])) == cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(9ULL):{
c->R[n] &= c->R[m];
break;}
case cast<uint16_t>(10ULL):{
c->R[n] ^= c->R[m];
break;}
case cast<uint16_t>(11ULL):{
c->R[n] |= c->R[m];
break;}
case cast<uint16_t>(12ULL):{
uint32_t d = cast<uint32_t>((c->R[n] ^ c->R[m]));
sh4_CPU_setT(c,((((cast<uint32_t>((d & cast<uint32_t>(4278190080ULL))) == cast<uint32_t>(0ULL)) || (cast<uint32_t>((d & cast<uint32_t>(16711680ULL))) == cast<uint32_t>(0ULL))) || (cast<uint32_t>((d & cast<uint32_t>(65280ULL))) == cast<uint32_t>(0ULL))) || (cast<uint32_t>((d & cast<uint32_t>(255ULL))) == cast<uint32_t>(0ULL))));
break;}
case cast<uint16_t>(13ULL):{
c->R[n] = cast<uint32_t>((shl<uint32_t>(c->R[m],cast<int64_t>(16ULL)) | shr<uint32_t>(c->R[n],cast<int64_t>(16ULL))));
break;}
case cast<uint16_t>(14ULL):{
c->MACL = cast<uint32_t>((cast<uint32_t>(cast<uint32_t>((c->R[n] & cast<uint32_t>(65535ULL)))) * cast<uint32_t>(cast<uint32_t>((c->R[m] & cast<uint32_t>(65535ULL))))));
break;}
case cast<uint16_t>(15ULL):{
c->MACL = cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<int16_t>(c->R[n])) * cast<int32_t>(cast<int16_t>(c->R[m])))));
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:318:1
void sh4_CPU_exec3(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_setT(c,(c->R[n] == c->R[m]));
break;}
case cast<uint16_t>(2ULL):{
sh4_CPU_setT(c,(c->R[n] >= c->R[m]));
break;}
case cast<uint16_t>(3ULL):{
sh4_CPU_setT(c,(cast<int32_t>(c->R[n]) >= cast<int32_t>(c->R[m])));
break;}
case cast<uint16_t>(4ULL):{
sh4_CPU_div1(c,n,m);
break;}
case cast<uint16_t>(5ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(c->R[n]) * cast<uint64_t>(c->R[m])));
auto tmp35 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL))),cast<uint32_t>(p));
c->MACH = std::get<0>(tmp35);
c->MACL = std::get<1>(tmp35);
break;}
case cast<uint16_t>(6ULL):{
sh4_CPU_setT(c,(c->R[n] > c->R[m]));
break;}
case cast<uint16_t>(7ULL):{
sh4_CPU_setT(c,(cast<int32_t>(c->R[n]) > cast<int32_t>(c->R[m])));
break;}
case cast<uint16_t>(8ULL):{
c->R[n] -= c->R[m];
break;}
case cast<uint16_t>(10ULL):{
uint32_t tmp = cast<uint32_t>((c->R[n] - c->R[m]));
uint32_t res = cast<uint32_t>((tmp - sh4_CPU_T(c)));
sh4_CPU_setT(c,((c->R[n] < tmp) || (tmp < res)));
c->R[n] = res;
break;}
case cast<uint16_t>(11ULL):{
uint32_t res = cast<uint32_t>((c->R[n] - c->R[m]));
sh4_CPU_setT(c,(cast<int32_t>(cast<uint32_t>(((cast<uint32_t>((c->R[n] ^ c->R[m]))) & (cast<uint32_t>((c->R[n] ^ res)))))) < cast<int32_t>(0ULL)));
c->R[n] = res;
break;}
case cast<uint16_t>(12ULL):{
c->R[n] += c->R[m];
break;}
case cast<uint16_t>(13ULL):{
uint64_t p = cast<uint64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>(c->R[n])) * cast<int64_t>(cast<int32_t>(c->R[m])))));
auto tmp36 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL))),cast<uint32_t>(p));
c->MACH = std::get<0>(tmp36);
c->MACL = std::get<1>(tmp36);
break;}
case cast<uint16_t>(14ULL):{
uint32_t tmp = cast<uint32_t>((c->R[n] + c->R[m]));
uint32_t res = cast<uint32_t>((tmp + sh4_CPU_T(c)));
sh4_CPU_setT(c,((tmp < c->R[n]) || (res < tmp)));
c->R[n] = res;
break;}
case cast<uint16_t>(15ULL):{
uint32_t res = cast<uint32_t>((c->R[n] + c->R[m]));
sh4_CPU_setT(c,(cast<int32_t>(cast<uint32_t>((cast<uint32_t>(~(cast<uint32_t>((c->R[n] ^ c->R[m])))) & (cast<uint32_t>((c->R[n] ^ res)))))) < cast<int32_t>(0ULL)));
c->R[n] = res;
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:365:1
void sh4_CPU_exec4(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(12ULL):{
{
uint32_t sh = c->R[m];
if ((cast<int32_t>(sh) >= cast<int32_t>(0ULL))) {
c->R[n] = shl<uint32_t>(c->R[n],cast<uint32_t>((sh & cast<uint32_t>(31ULL))));
}
else if ((cast<uint32_t>((sh & cast<uint32_t>(31ULL))) == cast<uint32_t>(0ULL))) {
c->R[n] = cast<uint32_t>(shr<int32_t>(cast<int32_t>(c->R[n]),cast<int64_t>(31ULL)));
}
else {
c->R[n] = cast<uint32_t>(shr<int32_t>(cast<int32_t>(c->R[n]),(cast<uint32_t>((cast<uint32_t>(32ULL) - cast<uint32_t>((sh & cast<uint32_t>(31ULL))))))));
}
}
return ;
break;}
case cast<uint16_t>(13ULL):{
{
uint32_t sh = c->R[m];
if ((cast<int32_t>(sh) >= cast<int32_t>(0ULL))) {
c->R[n] = shl<uint32_t>(c->R[n],cast<uint32_t>((sh & cast<uint32_t>(31ULL))));
}
else if ((cast<uint32_t>((sh & cast<uint32_t>(31ULL))) == cast<uint32_t>(0ULL))) {
c->R[n] = cast<uint32_t>(0ULL);
}
else {
c->R[n] = shr<uint32_t>(c->R[n],cast<uint32_t>((cast<uint32_t>(32ULL) - cast<uint32_t>((sh & cast<uint32_t>(31ULL))))));
}
}
return ;
break;}
case cast<uint16_t>(15ULL):{
sh4_CPU_macW(c,n,m);
return ;
break;}
case cast<uint16_t>(3ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],sh4_CPU_bankReg(c,m));
return ;
}
break;}
case cast<uint16_t>(7ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
sh4_CPU_setBankReg(c,m,sh4_CPU_read32(c,c->R[n]));
c->R[n] += cast<uint32_t>(4ULL);
return ;
}
break;}
case cast<uint16_t>(14ULL):{
if ((cast<uint16_t>((h & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
sh4_CPU_setBankReg(c,m,c->R[n]);
return ;
}
break;}
}}
{
switch(cast<uint16_t>((h & cast<uint16_t>(255ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = shl<uint32_t>(c->R[n],cast<int64_t>(1ULL));
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = shr<uint32_t>(c->R[n],cast<int64_t>(1ULL));
break;}
case cast<uint16_t>(2ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->MACH);
break;}
case cast<uint16_t>(3ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->SR);
break;}
case cast<uint16_t>(4ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = cast<uint32_t>((shl<uint32_t>(c->R[n],cast<int64_t>(1ULL)) | shr<uint32_t>(c->R[n],cast<int64_t>(31ULL))));
break;}
case cast<uint16_t>(5ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = cast<uint32_t>((shr<uint32_t>(c->R[n],cast<int64_t>(1ULL)) | shl<uint32_t>(c->R[n],cast<int64_t>(31ULL))));
break;}
case cast<uint16_t>(6ULL):{
c->MACH = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(7ULL):{
sh4_CPU_SetSR(c,sh4_CPU_read32(c,c->R[n]));
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(8ULL):{
c->R[n] = shl<uint32_t>(c->R[n],cast<int64_t>(2ULL));
break;}
case cast<uint16_t>(9ULL):{
c->R[n] = shr<uint32_t>(c->R[n],cast<int64_t>(2ULL));
break;}
case cast<uint16_t>(10ULL):{
c->MACH = c->R[n];
break;}
case cast<uint16_t>(11ULL):{
c->PR = cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)));
sh4_CPU_doBranch(c,true,c->R[n]);
break;}
case cast<uint16_t>(14ULL):{
sh4_CPU_SetSR(c,c->R[n]);
break;}
case cast<uint16_t>(16ULL):{
c->R[n]--;
sh4_CPU_setT(c,(c->R[n] == cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(17ULL):{
sh4_CPU_setT(c,(cast<int32_t>(c->R[n]) >= cast<int32_t>(0ULL)));
break;}
case cast<uint16_t>(18ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->MACL);
break;}
case cast<uint16_t>(19ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->GBR);
break;}
case cast<uint16_t>(21ULL):{
sh4_CPU_setT(c,(cast<int32_t>(c->R[n]) > cast<int32_t>(0ULL)));
break;}
case cast<uint16_t>(22ULL):{
c->MACL = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(23ULL):{
c->GBR = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(24ULL):{
c->R[n] = shl<uint32_t>(c->R[n],cast<int64_t>(8ULL));
break;}
case cast<uint16_t>(25ULL):{
c->R[n] = shr<uint32_t>(c->R[n],cast<int64_t>(8ULL));
break;}
case cast<uint16_t>(26ULL):{
c->MACL = c->R[n];
break;}
case cast<uint16_t>(27ULL):{
uint8_t b = sh4_CPU_read8(c,c->R[n]);
sh4_CPU_setT(c,(b == cast<uint8_t>(0ULL)));
sh4_CPU_write8(c,c->R[n],cast<uint8_t>((b | cast<uint8_t>(128ULL))));
break;}
case cast<uint16_t>(30ULL):{
c->GBR = c->R[n];
break;}
case cast<uint16_t>(32ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = shl<uint32_t>(c->R[n],cast<int64_t>(1ULL));
break;}
case cast<uint16_t>(33ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[n] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
c->R[n] = cast<uint32_t>(shr<int32_t>(cast<int32_t>(c->R[n]),cast<int64_t>(1ULL)));
break;}
case cast<uint16_t>(34ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->PR);
break;}
case cast<uint16_t>(35ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->VBR);
break;}
case cast<uint16_t>(36ULL):{
uint32_t t = shr<uint32_t>(c->R[n],cast<int64_t>(31ULL));
c->R[n] = cast<uint32_t>((shl<uint32_t>(c->R[n],cast<int64_t>(1ULL)) | sh4_CPU_T(c)));
sh4_CPU_setT(c,(t != cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(37ULL):{
uint32_t t = cast<uint32_t>((c->R[n] & cast<uint32_t>(1ULL)));
c->R[n] = cast<uint32_t>((shr<uint32_t>(c->R[n],cast<int64_t>(1ULL)) | shl<uint32_t>(sh4_CPU_T(c),cast<int64_t>(31ULL))));
sh4_CPU_setT(c,(t != cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(38ULL):{
c->PR = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(39ULL):{
c->VBR = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(40ULL):{
c->R[n] = shl<uint32_t>(c->R[n],cast<int64_t>(16ULL));
break;}
case cast<uint16_t>(41ULL):{
c->R[n] = shr<uint32_t>(c->R[n],cast<int64_t>(16ULL));
break;}
case cast<uint16_t>(42ULL):{
c->PR = c->R[n];
break;}
case cast<uint16_t>(43ULL):{
sh4_CPU_doBranch(c,true,c->R[n]);
break;}
case cast<uint16_t>(46ULL):{
c->VBR = c->R[n];
break;}
case cast<uint16_t>(50ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->SGR);
break;}
case cast<uint16_t>(51ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->SSR);
break;}
case cast<uint16_t>(55ULL):{
c->SSR = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(62ULL):{
c->SSR = c->R[n];
break;}
case cast<uint16_t>(67ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->SPC);
break;}
case cast<uint16_t>(71ULL):{
c->SPC = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(78ULL):{
c->SPC = c->R[n];
break;}
case cast<uint16_t>(82ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->FPUL);
break;}
case cast<uint16_t>(86ULL):{
c->FPUL = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(90ULL):{
c->FPUL = c->R[n];
break;}
case cast<uint16_t>(98ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->FPSCR);
break;}
case cast<uint16_t>(102ULL):{
sh4_CPU_SetFPSCR(c,sh4_CPU_read32(c,c->R[n]));
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(106ULL):{
sh4_CPU_SetFPSCR(c,c->R[n]);
break;}
case cast<uint16_t>(242ULL):{
c->R[n] -= cast<uint32_t>(4ULL);
sh4_CPU_write32(c,c->R[n],c->DBR);
break;}
case cast<uint16_t>(246ULL):{
c->DBR = sh4_CPU_read32(c,c->R[n]);
c->R[n] += cast<uint32_t>(4ULL);
break;}
case cast<uint16_t>(250ULL):{
c->DBR = c->R[n];
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:557:1
void sh4_CPU_exec6(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(sh4_CPU_read8(c,c->R[m]))));
break;}
case cast<uint16_t>(1ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,c->R[m]))));
break;}
case cast<uint16_t>(2ULL):{
c->R[n] = sh4_CPU_read32(c,c->R[m]);
break;}
case cast<uint16_t>(3ULL):{
c->R[n] = c->R[m];
break;}
case cast<uint16_t>(4ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(sh4_CPU_read8(c,c->R[m]))));
if ((n != m)) {
c->R[m]++;
}
break;}
case cast<uint16_t>(5ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,c->R[m]))));
if ((n != m)) {
c->R[m] += cast<uint32_t>(2ULL);
}
break;}
case cast<uint16_t>(6ULL):{
c->R[n] = sh4_CPU_read32(c,c->R[m]);
if ((n != m)) {
c->R[m] += cast<uint32_t>(4ULL);
}
break;}
case cast<uint16_t>(7ULL):{
c->R[n] = cast<uint32_t>(~c->R[m]);
break;}
case cast<uint16_t>(8ULL):{
uint32_t v = c->R[m];
c->R[n] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(65280ULL))))) | cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))));
break;}
case cast<uint16_t>(9ULL):{
c->R[n] = cast<uint32_t>((shl<uint32_t>(c->R[m],cast<int64_t>(16ULL)) | shr<uint32_t>(c->R[m],cast<int64_t>(16ULL))));
break;}
case cast<uint16_t>(10ULL):{
uint32_t res = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(0ULL) - c->R[m])) - sh4_CPU_T(c)));
sh4_CPU_setT(c,((c->R[m] != cast<uint32_t>(0ULL)) || (sh4_CPU_T(c) != cast<uint32_t>(0ULL))));
c->R[n] = res;
break;}
case cast<uint16_t>(11ULL):{
c->R[n] = cast<uint32_t>(-c->R[m]);
break;}
case cast<uint16_t>(12ULL):{
c->R[n] = cast<uint32_t>((c->R[m] & cast<uint32_t>(255ULL)));
break;}
case cast<uint16_t>(13ULL):{
c->R[n] = cast<uint32_t>((c->R[m] & cast<uint32_t>(65535ULL)));
break;}
case cast<uint16_t>(14ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(c->R[m])));
break;}
case cast<uint16_t>(15ULL):{
c->R[n] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(c->R[m])));
break;}
}}
}
}
// tools/cpu/sh4/exec.go:606:1
void sh4_CPU_exec8(sh4_CPU* c,uint16_t h){
{
uint32_t m = sh4_rm(h);
{
switch(cast<uint16_t>(((shr<uint16_t>(h,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_write8(c,cast<uint32_t>((c->R[m] + cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))))),cast<uint8_t>(c->R[cast<int64_t>(0ULL)]));
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_write16(c,cast<uint32_t>((c->R[m] + cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(2ULL))))),cast<uint16_t>(c->R[cast<int64_t>(0ULL)]));
break;}
case cast<uint16_t>(4ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(sh4_CPU_read8(c,cast<uint32_t>((c->R[m] + cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL))))))))));
break;}
case cast<uint16_t>(5ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,cast<uint32_t>((c->R[m] + cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(15ULL)))) * cast<uint32_t>(2ULL)))))))));
break;}
case cast<uint16_t>(8ULL):{
sh4_CPU_setT(c,(cast<int32_t>(c->R[cast<int64_t>(0ULL)]) == sh4_s8(h)));
break;}
case cast<uint16_t>(9ULL):{
sh4_CPU_doJumpNow(c,(sh4_CPU_T(c) != cast<uint32_t>(0ULL)),cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL)))))));
break;}
case cast<uint16_t>(11ULL):{
sh4_CPU_doJumpNow(c,(sh4_CPU_T(c) == cast<uint32_t>(0ULL)),cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL)))))));
break;}
case cast<uint16_t>(13ULL):{
sh4_CPU_doBranch(c,(sh4_CPU_T(c) != cast<uint32_t>(0ULL)),cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL)))))));
break;}
case cast<uint16_t>(15ULL):{
sh4_CPU_doBranch(c,(sh4_CPU_T(c) == cast<uint32_t>(0ULL)),cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) + cast<int32_t>((sh4_s8(h) * cast<int32_t>(2ULL)))))));
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec.go:632:1
void sh4_CPU_execC(sh4_CPU* c,uint16_t h){
{
uint32_t d = cast<uint32_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL))));
{
switch(cast<uint16_t>(((shr<uint16_t>(h,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_write8(c,cast<uint32_t>((c->GBR + d)),cast<uint8_t>(c->R[cast<int64_t>(0ULL)]));
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_write16(c,cast<uint32_t>((c->GBR + cast<uint32_t>((d * cast<uint32_t>(2ULL))))),cast<uint16_t>(c->R[cast<int64_t>(0ULL)]));
break;}
case cast<uint16_t>(2ULL):{
sh4_CPU_write32(c,cast<uint32_t>((c->GBR + cast<uint32_t>((d * cast<uint32_t>(4ULL))))),c->R[cast<int64_t>(0ULL)]);
break;}
case cast<uint16_t>(3ULL):{
c->TRA = shl<uint32_t>(d,cast<int64_t>(2ULL));
sh4_CPU_exception(c,cast<uint32_t>(352ULL),c->PC);
break;}
case cast<uint16_t>(4ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(sh4_CPU_read8(c,cast<uint32_t>((c->GBR + d))))));
break;}
case cast<uint16_t>(5ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(sh4_CPU_read16(c,cast<uint32_t>((c->GBR + cast<uint32_t>((d * cast<uint32_t>(2ULL)))))))));
break;}
case cast<uint16_t>(6ULL):{
c->R[cast<int64_t>(0ULL)] = sh4_CPU_read32(c,cast<uint32_t>((c->GBR + cast<uint32_t>((d * cast<uint32_t>(4ULL))))));
break;}
case cast<uint16_t>(7ULL):{
c->R[cast<int64_t>(0ULL)] = sh4_litL(c->curPC,d);
break;}
case cast<uint16_t>(8ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((c->R[cast<int64_t>(0ULL)] & d)) == cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(9ULL):{
c->R[cast<int64_t>(0ULL)] &= d;
break;}
case cast<uint16_t>(10ULL):{
c->R[cast<int64_t>(0ULL)] ^= d;
break;}
case cast<uint16_t>(11ULL):{
c->R[cast<int64_t>(0ULL)] |= d;
break;}
case cast<uint16_t>(12ULL):{
sh4_CPU_setT(c,(cast<uint32_t>((cast<uint32_t>(sh4_CPU_read8(c,cast<uint32_t>((c->GBR + c->R[cast<int64_t>(0ULL)])))) & d)) == cast<uint32_t>(0ULL)));
break;}
case cast<uint16_t>(13ULL):{
uint32_t a = cast<uint32_t>((c->GBR + c->R[cast<int64_t>(0ULL)]));
sh4_CPU_write8(c,a,cast<uint8_t>((sh4_CPU_read8(c,a) & cast<uint8_t>(d))));
break;}
case cast<uint16_t>(14ULL):{
uint32_t a = cast<uint32_t>((c->GBR + c->R[cast<int64_t>(0ULL)]));
sh4_CPU_write8(c,a,cast<uint8_t>((sh4_CPU_read8(c,a) ^ cast<uint8_t>(d))));
break;}
case cast<uint16_t>(15ULL):{
uint32_t a = cast<uint32_t>((c->GBR + c->R[cast<int64_t>(0ULL)]));
sh4_CPU_write8(c,a,cast<uint8_t>((sh4_CPU_read8(c,a) | cast<uint8_t>(d))));
break;}
}}
}
}
// tools/cpu/sh4/exec.go:675:1
uint32_t sh4_CPU_bankReg(sh4_CPU* c,uint32_t m){
{
return c->Rbank[cast<uint32_t>((m & cast<uint32_t>(7ULL)))];
}
}
// tools/cpu/sh4/exec.go:676:1
void sh4_CPU_setBankReg(sh4_CPU* c,uint32_t m,uint32_t v){
{
c->Rbank[cast<uint32_t>((m & cast<uint32_t>(7ULL)))] = v;
}
}
// tools/cpu/sh4/exec.go:681:1
void sh4_CPU_div1(sh4_CPU* c,uint32_t n,uint32_t m){
{
bool oldQ = (cast<uint32_t>((c->SR & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL));
bool q = (cast<uint32_t>((c->R[n] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
bool mf = (cast<uint32_t>((c->SR & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
uint32_t tmp2 = c->R[m];
c->R[n] = cast<uint32_t>((shl<uint32_t>(c->R[n],cast<int64_t>(1ULL)) | sh4_CPU_T(c)));
uint32_t tmp0 = c->R[n];
bool tmp1={};
{
if (((!oldQ) && (!mf))){
c->R[n] -= tmp2;
tmp1 = (c->R[n] > tmp0);
q = (q != tmp1);
}
else if (((!oldQ) && mf)){
c->R[n] += tmp2;
tmp1 = (c->R[n] < tmp0);
q = (q == tmp1);
}
else if ((oldQ && (!mf))){
c->R[n] += tmp2;
tmp1 = (c->R[n] < tmp0);
q = (q != tmp1);
}
else {
c->R[n] -= tmp2;
tmp1 = (c->R[n] > tmp0);
q = (q == tmp1);
}
}
tmp37:;
if (q) {
c->SR |= cast<uint32_t>(256ULL);
}
else {
c->SR &= ~(cast<uint32_t>(256ULL));
}
sh4_CPU_setT(c,(q == mf));
}
}
// tools/cpu/sh4/exec.go:717:1
void sh4_CPU_macL(sh4_CPU* c,uint32_t n,uint32_t m){
{
int64_t a = cast<int64_t>(cast<int32_t>(sh4_CPU_read32(c,c->R[n])));
c->R[n] += cast<uint32_t>(4ULL);
int64_t b = cast<int64_t>(cast<int32_t>(sh4_CPU_read32(c,c->R[m])));
c->R[m] += cast<uint32_t>(4ULL);
int64_t mac = cast<int64_t>((cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(c->MACH),cast<int64_t>(32ULL)) | cast<uint64_t>(c->MACL)))) + cast<int64_t>((a * b))));
if ((cast<uint32_t>((c->SR & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
constexpr int64_t hi=140737488355327ULL;
constexpr int64_t lo=-140737488355328ULL;
if ((mac > cast<int64_t>(140737488355327ULL))) {
mac = cast<int64_t>(140737488355327ULL);
}
else if ((mac < cast<int64_t>(-140737488355328ULL))) {
mac = cast<int64_t>(-140737488355328ULL);
}
}
auto tmp38 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(mac),cast<int64_t>(32ULL))),cast<uint32_t>(cast<uint64_t>(mac)));
c->MACH = std::get<0>(tmp38);
c->MACL = std::get<1>(tmp38);
}
}
// tools/cpu/sh4/exec.go:737:1
void sh4_CPU_macW(sh4_CPU* c,uint32_t n,uint32_t m){
{
int64_t a = cast<int64_t>(cast<int16_t>(sh4_CPU_read16(c,c->R[n])));
c->R[n] += cast<uint32_t>(2ULL);
int64_t b = cast<int64_t>(cast<int16_t>(sh4_CPU_read16(c,c->R[m])));
c->R[m] += cast<uint32_t>(2ULL);
if ((cast<uint32_t>((c->SR & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
int64_t sum = cast<int64_t>((cast<int64_t>(cast<int32_t>(c->MACL)) + cast<int64_t>((a * b))));
if ((sum > cast<int64_t>(2147483647ULL))) {
sum = cast<int64_t>(2147483647ULL);
c->MACH |= cast<uint32_t>(1ULL);
}
else if ((sum < cast<int64_t>(-2147483648ULL))) {
sum = cast<int64_t>(-2147483648ULL);
c->MACH |= cast<uint32_t>(1ULL);
}
c->MACL = cast<uint32_t>(sum);
return ;
}
int64_t mac = cast<int64_t>((cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(c->MACH),cast<int64_t>(32ULL)) | cast<uint64_t>(c->MACL)))) + cast<int64_t>((a * b))));
auto tmp39 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(mac),cast<int64_t>(32ULL))),cast<uint32_t>(cast<uint64_t>(mac)));
c->MACH = std::get<0>(tmp39);
c->MACL = std::get<1>(tmp39);
}
}
// tools/cpu/sh4/exec.go:758:1
void sh4_CPU_unknown(sh4_CPU* c,uint16_t h){
{
sh4_CPU_Halt(c,std::string("unimplemented instruction %04X (%s) at %08X",43),h,sh4_DecodeHalfword(h,c->curPC).Text,c->curPC);
}
}
// tools/cpu/sh4/exec_fpu.go:18:1
void sh4_CPU_execFPU(sh4_CPU* c,uint16_t h,uint32_t n,uint32_t m){
{
if ((cast<uint32_t>((c->SR & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
sh4_CPU_Halt(c,std::string("FPU instruction %04X with SR.FD set at %08X",43),h,c->curPC);
return ;
}
std::array<uint32_t,16>* fr = (&c->fpr[cast<uint32_t>(((shr<uint32_t>(c->FPSCR,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)))]);
bool pr = (cast<uint32_t>((c->FPSCR & cast<uint32_t>(524288ULL))) != cast<uint32_t>(0ULL));
{
switch(cast<uint16_t>((h & cast<uint16_t>(15ULL)))){
case cast<uint16_t>(0ULL):{
sh4_CPU_fpuArith(c,fr,pr,n,m,[&](double a,double b)->double{
return (a + b);
}
);
break;}
case cast<uint16_t>(1ULL):{
sh4_CPU_fpuArith(c,fr,pr,n,m,[&](double a,double b)->double{
return (a - b);
}
);
break;}
case cast<uint16_t>(2ULL):{
sh4_CPU_fpuArith(c,fr,pr,n,m,[&](double a,double b)->double{
return (a * b);
}
);
break;}
case cast<uint16_t>(3ULL):{
sh4_CPU_fpuArith(c,fr,pr,n,m,[&](double a,double b)->double{
return (a / b);
}
);
break;}
case cast<uint16_t>(4ULL):{
if (pr) {
sh4_CPU_setT(c,(sh4_getDR(fr,n) == sh4_getDR(fr,m)));
}
else {
sh4_CPU_setT(c,(sh4_getFR(fr,n) == sh4_getFR(fr,m)));
}
break;}
case cast<uint16_t>(5ULL):{
if (pr) {
sh4_CPU_setT(c,(sh4_getDR(fr,n) > sh4_getDR(fr,m)));
}
else {
sh4_CPU_setT(c,(sh4_getFR(fr,n) > sh4_getFR(fr,m)));
}
break;}
case cast<uint16_t>(6ULL):{
sh4_CPU_fmovLoad(c,n,m,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[m])));
break;}
case cast<uint16_t>(7ULL):{
sh4_CPU_fmovStore(c,n,m,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + c->R[n])));
break;}
case cast<uint16_t>(8ULL):{
sh4_CPU_fmovLoad(c,n,m,c->R[m]);
break;}
case cast<uint16_t>(9ULL):{
sh4_CPU_fmovLoad(c,n,m,c->R[m]);
if ((cast<uint32_t>((c->FPSCR & cast<uint32_t>(1048576ULL))) != cast<uint32_t>(0ULL))) {
c->R[m] += cast<uint32_t>(8ULL);
}
else {
c->R[m] += cast<uint32_t>(4ULL);
}
break;}
case cast<uint16_t>(10ULL):{
sh4_CPU_fmovStore(c,n,m,c->R[n]);
break;}
case cast<uint16_t>(11ULL):{
uint32_t sz = cast<uint32_t>(4ULL);
if ((cast<uint32_t>((c->FPSCR & cast<uint32_t>(1048576ULL))) != cast<uint32_t>(0ULL))) {
sz = cast<uint32_t>(8ULL);
}
c->R[n] -= sz;
sh4_CPU_fmovStore(c,n,m,c->R[n]);
break;}
case cast<uint16_t>(12ULL):{
if ((cast<uint32_t>((c->FPSCR & cast<uint32_t>(1048576ULL))) != cast<uint32_t>(0ULL))) {
Slice<uint32_t> src = sh4_CPU_pair(c,m);
Slice<uint32_t> dst = sh4_CPU_pair(c,n);
auto tmp40 = std::make_tuple(src[cast<int64_t>(0ULL)],src[cast<int64_t>(1ULL)]);
dst[cast<int64_t>(0ULL)] = std::get<0>(tmp40);
dst[cast<int64_t>(1ULL)] = std::get<1>(tmp40);
}
else {
(*fr)[n] = (*fr)[m];
}
break;}
case cast<uint16_t>(13ULL):{
sh4_CPU_execFPUxD(c,h,n,fr,pr);
break;}
case cast<uint16_t>(14ULL):{
if (pr) {
sh4_CPU_Halt(c,std::string("fmac with FPSCR.PR set at %08X",30),c->curPC);
return ;
}
(*fr)[n] = go_math_Float32bits(cast<float>(((cast<double>(sh4_getFR(fr,cast<uint32_t>(0ULL))) * cast<double>(sh4_getFR(fr,m))) + cast<double>(sh4_getFR(fr,n)))));
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec_fpu.go:89:1
void sh4_CPU_execFPUxD(sh4_CPU* c,uint16_t h,uint32_t n,std::array<uint32_t,16>* fr,bool pr){
{
{
switch(sh4_rm(h)){
case cast<uint32_t>(0ULL):{
(*fr)[n] = c->FPUL;
break;}
case cast<uint32_t>(1ULL):{
c->FPUL = (*fr)[n];
break;}
case cast<uint32_t>(2ULL):{
if (pr) {
sh4_setDR(fr,n,cast<double>(cast<int32_t>(c->FPUL)));
}
else {
(*fr)[n] = go_math_Float32bits(cast<float>(cast<int32_t>(c->FPUL)));
}
break;}
case cast<uint32_t>(3ULL):{
double v={};
if (pr) {
v = sh4_getDR(fr,n);
}
else {
v = cast<double>(sh4_getFR(fr,n));
}
{
if (go_math_IsNaN(v)){
c->FPUL = cast<uint32_t>(2147483648ULL);
}
else if ((v >= cast<double>(2147483647ULL))){
c->FPUL = cast<uint32_t>(2147483647ULL);
}
else if ((v <= cast<double>(-cast<int64_t>(2147483648ULL)))){
c->FPUL = cast<uint32_t>(2147483648ULL);
}
else {
c->FPUL = cast<uint32_t>(cast<int32_t>(v));
}
}
tmp41:;
break;}
case cast<uint32_t>(4ULL):{
(*fr)[n] ^= cast<uint32_t>(2147483648ULL);
break;}
case cast<uint32_t>(5ULL):{
(*fr)[n] &= ~(cast<uint32_t>(2147483648ULL));
break;}
case cast<uint32_t>(6ULL):{
if (pr) {
sh4_setDR(fr,n,go_math_Sqrt(sh4_getDR(fr,n)));
}
else {
(*fr)[n] = go_math_Float32bits(cast<float>(go_math_Sqrt(cast<double>(sh4_getFR(fr,n)))));
}
break;}
case cast<uint32_t>(7ULL):{
(*fr)[n] = go_math_Float32bits(cast<float>((cast<double>(1ULL) / go_math_Sqrt(cast<double>(sh4_getFR(fr,n))))));
break;}
case cast<uint32_t>(8ULL):{
(*fr)[n] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(9ULL):{
(*fr)[n] = go_math_Float32bits(cast<float>(1ULL));
break;}
case cast<uint32_t>(10ULL):{
sh4_setDR(fr,(n & ~(cast<uint32_t>(1ULL))),cast<double>(go_math_Float32frombits(c->FPUL)));
break;}
case cast<uint32_t>(11ULL):{
c->FPUL = go_math_Float32bits(cast<float>(sh4_getDR(fr,(n & ~(cast<uint32_t>(1ULL))))));
break;}
case cast<uint32_t>(14ULL):{
auto tmp42 = std::make_tuple(cast<uint32_t>(((cast<uint32_t>((n & cast<uint32_t>(3ULL)))) * cast<uint32_t>(4ULL))),cast<uint32_t>(((shr<uint32_t>(n,cast<int64_t>(2ULL))) * cast<uint32_t>(4ULL))));
uint32_t vm = std::get<0>(tmp42);
uint32_t vn = std::get<1>(tmp42);
double sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
sum += (cast<double>(sh4_getFR(fr,cast<uint32_t>((vm + i)))) * cast<double>(sh4_getFR(fr,cast<uint32_t>((vn + i)))));
}
}(*fr)[cast<uint32_t>((vn + cast<uint32_t>(3ULL)))] = go_math_Float32bits(cast<float>(sum));
break;}
case cast<uint32_t>(15ULL):{
{
if ((h == cast<uint16_t>(64509ULL))){
c->FPSCR ^= cast<uint32_t>(2097152ULL);
}
else if ((h == cast<uint16_t>(62461ULL))){
c->FPSCR ^= cast<uint32_t>(1048576ULL);
}
else if ((cast<uint32_t>((n & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL))){
sh4_CPU_ftrv(c,fr,cast<uint32_t>(((shr<uint32_t>(n,cast<int64_t>(2ULL))) * cast<uint32_t>(4ULL))));
}
else if ((cast<uint32_t>((n & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))){
auto tmp44 = go_math_Sincos((((cast<double>(2ULL) * go_math_Pi) * cast<double>(cast<uint32_t>((c->FPUL & cast<uint32_t>(65535ULL))))) / cast<double>(65536ULL)));
double s = std::get<0>(tmp44);
double co = std::get<1>(tmp44);
(*fr)[n] = go_math_Float32bits(cast<float>(s));
(*fr)[cast<uint32_t>((n + cast<uint32_t>(1ULL)))] = go_math_Float32bits(cast<float>(co));
}
else {
sh4_CPU_unknown(c,h);
}
}
tmp43:;
break;}
default:{
sh4_CPU_unknown(c,h);
break;}
}}
}
}
// tools/cpu/sh4/exec_fpu.go:167:1
void sh4_CPU_fpuArith(sh4_CPU* c,std::array<uint32_t,16>* fr,bool pr,uint32_t n,uint32_t m,std::function<double(double,double)> op){
{
if (pr) {
sh4_setDR(fr,n,op(sh4_getDR(fr,n),sh4_getDR(fr,m)));
return ;
}
(*fr)[n] = go_math_Float32bits(cast<float>(op(cast<double>(sh4_getFR(fr,n)),cast<double>(sh4_getFR(fr,m)))));
}
}
// tools/cpu/sh4/exec_fpu.go:178:1
Slice<uint32_t> sh4_CPU_pair(sh4_CPU* c,uint32_t i){
{
uint32_t b = cast<uint32_t>(((shr<uint32_t>(c->FPSCR,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
if ((cast<uint32_t>((i & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
b = cast<uint32_t>((cast<uint32_t>(1ULL) - b));
}
uint32_t base = (i & ~(cast<uint32_t>(1ULL)));
return sub(c->fpr[b],base,cast<uint32_t>((base + cast<uint32_t>(2ULL))));
}
}
// tools/cpu/sh4/exec_fpu.go:190:1
void sh4_CPU_fmovLoad(sh4_CPU* c,uint32_t n,uint32_t m,uint32_t addr){
{
if ((cast<uint32_t>((c->FPSCR & cast<uint32_t>(1048576ULL))) != cast<uint32_t>(0ULL))) {
Slice<uint32_t> dst = sh4_CPU_pair(c,n);
dst[cast<int64_t>(0ULL)] = sh4_CPU_read32(c,addr);
dst[cast<int64_t>(1ULL)] = sh4_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
return ;
}
c->fpr[cast<uint32_t>(((shr<uint32_t>(c->FPSCR,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)))][n] = sh4_CPU_read32(c,addr);
}
}
// tools/cpu/sh4/exec_fpu.go:200:1
void sh4_CPU_fmovStore(sh4_CPU* c,uint32_t n,uint32_t m,uint32_t addr){
{
if ((cast<uint32_t>((c->FPSCR & cast<uint32_t>(1048576ULL))) != cast<uint32_t>(0ULL))) {
Slice<uint32_t> src = sh4_CPU_pair(c,m);
sh4_CPU_write32(c,addr,src[cast<int64_t>(0ULL)]);
sh4_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),src[cast<int64_t>(1ULL)]);
return ;
}
sh4_CPU_write32(c,addr,c->fpr[cast<uint32_t>(((shr<uint32_t>(c->FPSCR,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)))][m]);
}
}
// tools/cpu/sh4/exec_fpu.go:212:1
void sh4_CPU_ftrv(sh4_CPU* c,std::array<uint32_t,16>* fr,uint32_t v){
{
std::array<uint32_t,16>* xf = (&c->fpr[cast<uint32_t>((cast<uint32_t>(1ULL) - cast<uint32_t>(((shr<uint32_t>(c->FPSCR,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)))))]);
std::array<double,4> in={};
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < cast<uint32_t>(4ULL));j++){
in[j] = cast<double>(sh4_getFR(fr,cast<uint32_t>((v + j))));
}
}{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
double sum={};
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < cast<uint32_t>(4ULL));j++){
sum += (cast<double>(go_math_Float32frombits((*xf)[cast<uint32_t>((i + cast<uint32_t>((cast<uint32_t>(4ULL) * j))))])) * in[j]);
}
}(*fr)[cast<uint32_t>((v + i))] = go_math_Float32bits(cast<float>(sum));
}
}}
}
// tools/cpu/sh4/exec_fpu.go:227:1
float sh4_getFR(std::array<uint32_t,16>* fr,uint32_t i){
{
return go_math_Float32frombits((*fr)[cast<uint32_t>((i & cast<uint32_t>(15ULL)))]);
}
}
// tools/cpu/sh4/exec_fpu.go:230:1
double sh4_getDR(std::array<uint32_t,16>* fr,uint32_t i){
{
i &= cast<uint32_t>(14ULL);
return go_math_Float64frombits(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>((*fr)[i]),cast<int64_t>(32ULL)) | cast<uint64_t>((*fr)[cast<uint32_t>((i + cast<uint32_t>(1ULL)))]))));
}
}
// tools/cpu/sh4/exec_fpu.go:235:1
void sh4_setDR(std::array<uint32_t,16>* fr,uint32_t i,double v){
{
i &= cast<uint32_t>(14ULL);
uint64_t b = go_math_Float64bits(v);
auto tmp45 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(b,cast<int64_t>(32ULL))),cast<uint32_t>(b));
(*fr)[i] = std::get<0>(tmp45);
(*fr)[cast<uint32_t>((i + cast<uint32_t>(1ULL)))] = std::get<1>(tmp45);
}
}
// tools/cpu/sh4/onchip.go:31:1
uint32_t sh4_CPU_onchipRead(sh4_CPU* c,uint32_t addr,int64_t size){
{
{
auto tmp46 = sh4_CPU_tmuRead(c,addr);
uint32_t v = std::get<0>(tmp46);
bool ok = std::get<1>(tmp46);
if (ok) {
return v;
}
}
{
switch(addr){
case cast<uint32_t>(4293394448ULL):{
return cast<uint32_t>(96ULL);
break;}
case cast<uint32_t>(4278190080ULL):{
return c->PTEH;
break;}
case cast<uint32_t>(4278190084ULL):{
return c->PTEL;
break;}
case cast<uint32_t>(4278190088ULL):{
return c->TTB;
break;}
case cast<uint32_t>(4278190092ULL):{
return c->TEA;
break;}
case cast<uint32_t>(4278190096ULL):{
return c->MMUCR;
break;}
case cast<uint32_t>(4278190108ULL):{
return c->CCR;
break;}
case cast<uint32_t>(4278190112ULL):{
return c->TRA;
break;}
case cast<uint32_t>(4278190116ULL):{
return c->EXPEVT;
break;}
case cast<uint32_t>(4278190120ULL):{
return c->INTEVT;
break;}
case cast<uint32_t>(4278190128ULL):{
return cast<uint32_t>(67241409ULL);
break;}
case cast<uint32_t>(4278190136ULL):{
return c->QACR0;
break;}
case cast<uint32_t>(4278190140ULL):{
return c->QACR1;
break;}
case cast<uint32_t>(4291821568ULL):{
return c->ICR;
break;}
case cast<uint32_t>(4291821572ULL):{
return c->IPRA;
break;}
case cast<uint32_t>(4291821576ULL):{
return c->IPRB;
break;}
case cast<uint32_t>(4291821580ULL):{
return c->IPRC;
break;}
}}
{
auto tmp47 = lookup(c->onchipRaw,addr);
uint32_t v = std::get<0>(tmp47);
bool ok = std::get<1>(tmp47);
if (ok) {
return v;
}
}
c->onchipGaps[addr]++;
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/sh4/onchip.go:78:1
void sh4_CPU_onchipWrite(sh4_CPU* c,uint32_t addr,int64_t size,uint32_t v){
{
if (sh4_CPU_tmuWrite(c,addr,v)) {
return ;
}
{
switch(addr){
case cast<uint32_t>(4293394444ULL):{
c->SerialTX = append(c->SerialTX,Slice<uint8_t>{cast<uint8_t>(v)});
return ;
break;}
case cast<uint32_t>(4293394448ULL):{
return ;
break;}
case cast<uint32_t>(4278190080ULL):{
c->PTEH = v;
break;}
case cast<uint32_t>(4278190084ULL):{
c->PTEL = v;
break;}
case cast<uint32_t>(4278190088ULL):{
c->TTB = v;
break;}
case cast<uint32_t>(4278190092ULL):{
c->TEA = v;
break;}
case cast<uint32_t>(4278190096ULL):{
c->MMUCR = v;
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
sh4_CPU_Halt(c,std::string("MMUCR.AT enabled at PC %08X \342\200\224 the MMU is unmodelled",53),c->curPC);
}
break;}
case cast<uint32_t>(4278190108ULL):{
c->CCR = (v & ~(cast<uint32_t>(2056ULL)));
break;}
case cast<uint32_t>(4278190112ULL):{
c->TRA = v;
break;}
case cast<uint32_t>(4278190116ULL):{
c->EXPEVT = v;
break;}
case cast<uint32_t>(4278190120ULL):{
c->INTEVT = v;
break;}
case cast<uint32_t>(4278190136ULL):{
c->QACR0 = v;
break;}
case cast<uint32_t>(4278190140ULL):{
c->QACR1 = v;
break;}
case cast<uint32_t>(4291821568ULL):{
c->ICR = v;
break;}
case cast<uint32_t>(4291821572ULL):{
c->IPRA = v;
break;}
case cast<uint32_t>(4291821576ULL):{
c->IPRB = v;
break;}
case cast<uint32_t>(4291821580ULL):{
c->IPRC = v;
break;}
default:{
c->onchipRaw[addr] = v;
break;}
}}
}
}
// tools/cpu/sh4/onchip.go:132:1
uint32_t sh4_CPU_OnchipReg(sh4_CPU* c,uint32_t addr){
{
return get(c->onchipRaw,addr);
}
}
// tools/cpu/sh4/onchip.go:136:1
Slice<std::string> sh4_CPU_Gaps(sh4_CPU* c){
{
Slice<sh4_onchip_137_gap> gs={};
{auto&& tmp48 = c->onchipGaps;
for(auto [tmp49,tmp50]:tmp48){
auto a=tmp49;auto n=tmp50;gs = append(gs,Slice<sh4_onchip_137_gap>{sh4_onchip_137_gap{a,n}});
}}
go_sort_Slice(gs,[&](int64_t i,int64_t j)->bool{
return (gs[i].reads > gs[j].reads);
}
);
Slice<std::string> out={};
{auto&& tmp51 = gs;
for(int64_t tmp52=0;tmp52<len(tmp51);++tmp52){
auto g=tmp51[tmp52];out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("onchip read %08X x%d (never written)",36),g.addr,g.reads)});
}}
return out;
}
}
// tools/cpu/sh4/sh4.go:69:1
std::string sh4_Flow_String(sh4_Flow f){
{
{
switch(f){
case cast<sh4_Flow>(0ULL):{
return std::string("seq",3);
break;}
case cast<sh4_Flow>(1ULL):{
return std::string("branch",6);
break;}
case cast<sh4_Flow>(2ULL):{
return std::string("jump",4);
break;}
case cast<sh4_Flow>(3ULL):{
return std::string("call",4);
break;}
case cast<sh4_Flow>(4ULL):{
return std::string("return",6);
break;}
case cast<sh4_Flow>(5ULL):{
return std::string("indjump",7);
break;}
case cast<sh4_Flow>(6ULL):{
return std::string("indcall",7);
break;}
case cast<sh4_Flow>(7ULL):{
return std::string("stop",4);
break;}
}}
return std::string("?",1);
}
}
// tools/cpu/sh4/sh4.go:112:1
std::string sh4_Inst_String(sh4_Inst in){
{
return go_fmt_Sprintf(std::string("$%08X: %s",9),in.Addr,in.Text);
}
}
// tools/cpu/sh4/sh4.go:121:1
uint32_t sh4_litW(uint32_t addr,uint32_t disp){
{
return cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + cast<uint32_t>((disp * cast<uint32_t>(2ULL)))));
}
}
// tools/cpu/sh4/sh4.go:122:1
uint32_t sh4_litL(uint32_t addr,uint32_t disp){
{
return cast<uint32_t>((((cast<uint32_t>((addr + cast<uint32_t>(4ULL)))) & ~(cast<uint32_t>(3ULL))) + cast<uint32_t>((disp * cast<uint32_t>(4ULL)))));
}
}
// tools/cpu/sh4/sq.go:13:1
void sh4_CPU_sqWrite(sh4_CPU* c,uint32_t addr,uint32_t size,uint32_t v){
{
uint32_t* w = (&c->SQ[cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))][cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)))]);
{
switch(size){
case cast<uint32_t>(1ULL):{
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((addr & cast<uint32_t>(3ULL))))));
(*w) = cast<uint32_t>((((*w) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(255ULL)))),sh)));
break;}
case cast<uint32_t>(2ULL):{
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((addr & cast<uint32_t>(2ULL))))));
(*w) = cast<uint32_t>((((*w) & ~((shl<uint32_t>(cast<uint32_t>(65535ULL),sh)))) | shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(65535ULL)))),sh)));
break;}
default:{
(*w) = v;
break;}
}}
}
}
// tools/cpu/sh4/sq.go:28:1
uint32_t sh4_CPU_sqRead32(sh4_CPU* c,uint32_t addr){
{
return c->SQ[cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))][cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)))];
}
}
// tools/cpu/sh4/sq.go:35:1
void sh4_CPU_sqFlush(sh4_CPU* c,uint32_t addr){
{
uint32_t q = cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t qacr = c->QACR0;
if ((q == cast<uint32_t>(1ULL))) {
qacr = c->QACR1;
}
uint32_t ext = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((qacr & cast<uint32_t>(28ULL)))),cast<int64_t>(24ULL)) | cast<uint32_t>((addr & cast<uint32_t>(67108832ULL)))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(8ULL));i++){
dc_Machine_Write32(c->bus,cast<uint32_t>((ext + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))),c->SQ[q][i]);
}
}}
}
// tools/cpu/sh4/state.go:51:1
sh4_State sh4_CPU_Snapshot(sh4_CPU* c){
{
sh4_State s = sh4_State{c->R,c->Rbank,c->SR,c->GBR,c->VBR,c->SSR,c->SPC,c->SGR,c->DBR,c->MACH,c->MACL,c->PR,c->PC,c->nextPC,c->fpr,c->FPSCR,c->FPUL,c->MMUCR,c->CCR,c->TRA,c->EXPEVT,c->INTEVT,c->PTEH,c->PTEL,c->TTB,c->TEA,c->QACR0,c->QACR1,c->ICR,c->IPRA,c->IPRB,c->IPRC,c->TMU,c->SQ,c->irlLevel,c->irlCode,append(Slice<uint8_t>{},c->SerialTX),c->Halted,c->HaltReason,c->Steps,c->curPC,c->delaySlot,c->pendingDelay,Map<uint32_t,uint32_t>{},Map<uint32_t,int64_t>{}};
{auto&& tmp53 = c->onchipRaw;
for(auto [tmp54,tmp55]:tmp53){
auto k=tmp54;auto v=tmp55;s.OnchipRaw[k] = v;
}}
{auto&& tmp56 = c->onchipGaps;
for(auto [tmp57,tmp58]:tmp56){
auto k=tmp57;auto v=tmp58;s.OnchipGaps[k] = v;
}}
return s;
}
}
// tools/cpu/sh4/state.go:79:1
void sh4_CPU_Restore(sh4_CPU* c,sh4_State s){
{
auto tmp59 = std::make_tuple(s.R,s.Rbank,s.SR);
c->R = std::get<0>(tmp59);
c->Rbank = std::get<1>(tmp59);
c->SR = std::get<2>(tmp59);
auto tmp60 = std::make_tuple(s.GBR,s.VBR,s.SSR,s.SPC,s.SGR,s.DBR);
c->GBR = std::get<0>(tmp60);
c->VBR = std::get<1>(tmp60);
c->SSR = std::get<2>(tmp60);
c->SPC = std::get<3>(tmp60);
c->SGR = std::get<4>(tmp60);
c->DBR = std::get<5>(tmp60);
auto tmp61 = std::make_tuple(s.MACH,s.MACL,s.PR,s.PC,s.NextPC);
c->MACH = std::get<0>(tmp61);
c->MACL = std::get<1>(tmp61);
c->PR = std::get<2>(tmp61);
c->PC = std::get<3>(tmp61);
c->nextPC = std::get<4>(tmp61);
auto tmp62 = std::make_tuple(s.FPR,s.FPSCR,s.FPUL);
c->fpr = std::get<0>(tmp62);
c->FPSCR = std::get<1>(tmp62);
c->FPUL = std::get<2>(tmp62);
auto tmp63 = std::make_tuple(s.MMUCR,s.CCR,s.TRA,s.EXPEVT,s.INTEVT);
c->MMUCR = std::get<0>(tmp63);
c->CCR = std::get<1>(tmp63);
c->TRA = std::get<2>(tmp63);
c->EXPEVT = std::get<3>(tmp63);
c->INTEVT = std::get<4>(tmp63);
auto tmp64 = std::make_tuple(s.PTEH,s.PTEL,s.TTB,s.TEA);
c->PTEH = std::get<0>(tmp64);
c->PTEL = std::get<1>(tmp64);
c->TTB = std::get<2>(tmp64);
c->TEA = std::get<3>(tmp64);
auto tmp65 = std::make_tuple(s.QACR0,s.QACR1);
c->QACR0 = std::get<0>(tmp65);
c->QACR1 = std::get<1>(tmp65);
auto tmp66 = std::make_tuple(s.ICR,s.IPRA,s.IPRB,s.IPRC);
c->ICR = std::get<0>(tmp66);
c->IPRA = std::get<1>(tmp66);
c->IPRB = std::get<2>(tmp66);
c->IPRC = std::get<3>(tmp66);
auto tmp67 = std::make_tuple(s.TMU,s.SQ);
c->TMU = std::get<0>(tmp67);
c->SQ = std::get<1>(tmp67);
auto tmp68 = std::make_tuple(s.IRLLevel,s.IRLCode);
c->irlLevel = std::get<0>(tmp68);
c->irlCode = std::get<1>(tmp68);
c->SerialTX = append(Slice<uint8_t>{},s.SerialTX);
auto tmp69 = std::make_tuple(s.Halted,s.HaltReason,s.Steps);
c->Halted = std::get<0>(tmp69);
c->HaltReason = std::get<1>(tmp69);
c->Steps = std::get<2>(tmp69);
auto tmp70 = std::make_tuple(s.CurPC,s.DelaySlot,s.PendingDelay);
c->curPC = std::get<0>(tmp70);
c->delaySlot = std::get<1>(tmp70);
c->pendingDelay = std::get<2>(tmp70);
c->onchipRaw = Map<uint32_t,uint32_t>{};
{auto&& tmp71 = s.OnchipRaw;
for(auto [tmp72,tmp73]:tmp71){
auto k=tmp72;auto v=tmp73;c->onchipRaw[k] = v;
}}
c->onchipGaps = Map<uint32_t,int64_t>{};
{auto&& tmp74 = s.OnchipGaps;
for(auto [tmp75,tmp76]:tmp74){
auto k=tmp75;auto v=tmp76;c->onchipGaps[k] = v;
}}
}
}
// tools/cpu/sh4/tmu.go:39:1
uint32_t sh4_tmuPeriod(uint16_t tcr){
{
{
switch(cast<uint16_t>((tcr & cast<uint16_t>(7ULL)))){
case cast<uint16_t>(0ULL):{
return cast<uint32_t>(16ULL);
break;}
case cast<uint16_t>(1ULL):{
return cast<uint32_t>(64ULL);
break;}
case cast<uint16_t>(2ULL):{
return cast<uint32_t>(256ULL);
break;}
case cast<uint16_t>(3ULL):{
return cast<uint32_t>(1024ULL);
break;}
default:{
return cast<uint32_t>(4096ULL);
break;}
}}
}
}
// tools/cpu/sh4/tmu.go:55:1
void sh4_CPU_tickTMU(sh4_CPU* c){
{
uint8_t ts = cast<uint8_t>((c->TMU.TSTR & cast<uint8_t>(7ULL)));
if ((ts == cast<uint8_t>(0ULL))) {
return ;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
if ((cast<uint8_t>((ts & (shl<uint8_t>(cast<uint8_t>(1ULL),i)))) == cast<uint8_t>(0ULL))) {
continue;
}
sh4_TMUChannel* ch = (&c->TMU.Ch[i]);
ch->Frac++;
if ((ch->Frac < sh4_tmuPeriod(ch->TCR))) {
continue;
}
ch->Frac = cast<uint32_t>(0ULL);
if ((ch->TCNT == cast<uint32_t>(0ULL))) {
ch->TCNT = ch->TCOR;
ch->TCR |= cast<uint16_t>(256ULL);
}
else {
ch->TCNT--;
}
}
}}
}
// tools/cpu/sh4/tmu.go:81:1
std::tuple<uint32_t,uint32_t,bool> sh4_CPU_tmuPending(sh4_CPU* c){
uint32_t level{};
uint32_t code{};
bool ok{};
{
std::array<uint32_t,3> shift = std::array<uint32_t,3>{cast<uint32_t>(12ULL),cast<uint32_t>(8ULL),cast<uint32_t>(4ULL)};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
sh4_TMUChannel* ch = (&c->TMU.Ch[i]);
if (((cast<uint16_t>((ch->TCR & cast<uint16_t>(256ULL))) == cast<uint16_t>(0ULL)) || (cast<uint16_t>((ch->TCR & cast<uint16_t>(32ULL))) == cast<uint16_t>(0ULL)))) {
continue;
}
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(c->IPRA,shift[i])) & cast<uint32_t>(15ULL)));
if ((l > level)) {
auto tmp77 = std::make_tuple(l,cast<uint32_t>((cast<uint32_t>(1024ULL) + cast<uint32_t>((cast<uint32_t>(32ULL) * cast<uint32_t>(i))))),true);
level = std::get<0>(tmp77);
code = std::get<1>(tmp77);
ok = std::get<2>(tmp77);
}
}
}
}return {level,code,ok};
}
}
// tools/cpu/sh4/tmu.go:96:1
std::tuple<uint32_t,bool> sh4_CPU_tmuRead(sh4_CPU* c,uint32_t addr){
{
{
switch(addr){
case cast<uint32_t>(4292345856ULL):{
return {cast<uint32_t>(c->TMU.TOCR),true};
break;}
case cast<uint32_t>(4292345860ULL):{
return {cast<uint32_t>(c->TMU.TSTR),true};
break;}
}}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(3ULL));i++){
uint32_t base = cast<uint32_t>((cast<uint32_t>(4292345864ULL) + cast<uint32_t>((i * cast<uint32_t>(12ULL)))));
{
auto tmp79=addr;
if (tmp79==(base)){
return {c->TMU.Ch[i].TCOR,true};
}
else if (tmp79==(cast<uint32_t>((base + cast<uint32_t>(4ULL))))){
return {c->TMU.Ch[i].TCNT,true};
}
else if (tmp79==(cast<uint32_t>((base + cast<uint32_t>(8ULL))))){
return {cast<uint32_t>(c->TMU.Ch[i].TCR),true};
}
}
tmp78:;
}
}return {cast<uint32_t>(0ULL),false};
}
}
// tools/cpu/sh4/tmu.go:117:1
bool sh4_CPU_tmuWrite(sh4_CPU* c,uint32_t addr,uint32_t v){
{
{
switch(addr){
case cast<uint32_t>(4292345856ULL):{
c->TMU.TOCR = cast<uint8_t>(v);
return true;
break;}
case cast<uint32_t>(4292345860ULL):{
c->TMU.TSTR = cast<uint8_t>(v);
return true;
break;}
}}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(3ULL));i++){
uint32_t base = cast<uint32_t>((cast<uint32_t>(4292345864ULL) + cast<uint32_t>((i * cast<uint32_t>(12ULL)))));
{
auto tmp81=addr;
if (tmp81==(base)){
c->TMU.Ch[i].TCOR = v;
return true;
}
else if (tmp81==(cast<uint32_t>((base + cast<uint32_t>(4ULL))))){
c->TMU.Ch[i].TCNT = v;
return true;
}
else if (tmp81==(cast<uint32_t>((base + cast<uint32_t>(8ULL))))){
uint16_t old = c->TMU.Ch[i].TCR;
c->TMU.Ch[i].TCR = cast<uint16_t>(((cast<uint16_t>(v) & ~(cast<uint16_t>(256ULL))) | cast<uint16_t>((cast<uint16_t>((old & cast<uint16_t>(v))) & cast<uint16_t>(256ULL)))));
return true;
}
}
tmp80:;
}
}return false;
}
}
// tools/lib/iso9660/iso9660.go:58:1
std::string iso9660_Geometry_String(iso9660_Geometry g){
{
return go_fmt_Sprintf(std::string("%d-byte sectors, data at +%d",28),g.SectorSize,g.DataOffset);
}
}
// tools/lib/iso9660/iso9660.go:81:1
iso9660_Geometry iso9660_Source_Geometry(iso9660_Source* s){
{
return s->geom;
}
}
// tools/lib/iso9660/iso9660.go:162:1
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadBlock(iso9660_Volume* v,int64_t n){
{
return dc_Disc_ReadBlock(v->src,n);
}
}
// tools/lib/iso9660/iso9660.go:173:1
std::string iso9660_Entry_String(iso9660_Entry e){
{
if (e.IsDir) {
return go_fmt_Sprintf(std::string("%-32s  <dir>   (lba %d)",23),e.Path,e.Block);
}
return go_fmt_Sprintf(std::string("%-32s  %10d  (lba %d)",21),e.Path,e.Size,e.Block);
}
}
// tools/lib/iso9660/iso9660.go:197:1
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolume(dc_Disc* src){
{
return iso9660_OpenVolumeAt(src,cast<int64_t>(16ULL));
}
}
// tools/lib/iso9660/iso9660.go:205:1
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolumeAt(dc_Disc* src,int64_t pvd){
{
auto tmp1 = dc_Disc_ReadBlock(src,pvd);
Slice<uint8_t> b = std::get<0>(tmp1);
Error err = std::get<1>(tmp1);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("iso9660: reading PVD: %w",24),err)};
}
if (((b[cast<int64_t>(0ULL)] != cast<uint8_t>(1ULL)) || (cast<std::string>(sub(b,cast<int64_t>(1ULL),cast<int64_t>(6ULL))) != std::string("CD001",5)))) {
return {{},go_fmt_Errorf(std::string("iso9660: no Primary Volume Descriptor at LBA %d",47),pvd)};
}
iso9660_Volume* v = arenaNew(iso9660_Volume{src,go_strings_TrimRight(cast<std::string>(sub(b,cast<int64_t>(8ULL),cast<int64_t>(40ULL))),std::string(" \000",2)),go_strings_TrimRight(cast<std::string>(sub(b,cast<int64_t>(40ULL),cast<int64_t>(72ULL))),std::string(" \000",2)),cast<int64_t>(iso9660_le32(rrBorrow(b,cast<int64_t>(80ULL),len(b)))),{},{}});
Slice<uint8_t> root = sub(b,cast<int64_t>(156ULL),cast<int64_t>(190ULL));
v->rootLBA = cast<int64_t>(iso9660_le32(rrBorrow(root,cast<int64_t>(2ULL),len(root))));
v->rootSize = cast<int64_t>(iso9660_le32(rrBorrow(root,cast<int64_t>(10ULL),len(root))));
return {v,{}};
}
}
// tools/lib/iso9660/iso9660.go:227:1
uint32_t iso9660_le32(Slice<uint8_t> b){
{
return le_Uint32(rrBorrow(b,0,cast<int64_t>(4ULL)));
}
}
// tools/lib/iso9660/iso9660.go:231:1
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_dirEntries(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath){
{
Slice<iso9660_Entry> out={};
int64_t remaining = size;
{int64_t sect = cast<int64_t>(0ULL);for (;(remaining > cast<int64_t>(0ULL));sect++){
auto tmp2 = dc_Disc_ReadBlock(v->src,cast<int64_t>((lba + sect)));
Slice<uint8_t> b = std::get<0>(tmp2);
Error err = std::get<1>(tmp2);
if (bool(err)) {
return {{},err};
}
int64_t n = cast<int64_t>(2048ULL);
if ((remaining < n)) {
n = remaining;
}
remaining -= cast<int64_t>(2048ULL);
{int64_t p = cast<int64_t>(0ULL);for (;(p < n);){
int64_t recLen = cast<int64_t>(b[p]);
if ((recLen == cast<int64_t>(0ULL))) {
break;
}
if ((cast<int64_t>((p + recLen)) > cast<int64_t>(2048ULL))) {
return {{},go_fmt_Errorf(std::string("iso9660: directory record overruns block at lba %d",50),cast<int64_t>((lba + sect)))};
}
Slice<uint8_t> rec = sub(b,p,cast<int64_t>((p + recLen)));
p += recLen;
int64_t idLen = cast<int64_t>(rec[cast<int64_t>(32ULL)]);
Slice<uint8_t> id = sub(rec,cast<int64_t>(33ULL),cast<int64_t>((cast<int64_t>(33ULL) + idLen)));
if (((idLen == cast<int64_t>(1ULL)) && (((id[cast<int64_t>(0ULL)] == cast<uint8_t>(0ULL)) || (id[cast<int64_t>(0ULL)] == cast<uint8_t>(1ULL)))))) {
continue;
}
iso9660_Entry e = iso9660_Entry{cast<std::string>(id),{},(cast<uint8_t>((rec[cast<int64_t>(25ULL)] & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL)),cast<int64_t>(iso9660_le32(rrBorrow(rec,cast<int64_t>(10ULL),len(rec)))),cast<int64_t>(iso9660_le32(rrBorrow(rec,cast<int64_t>(2ULL),len(rec))))};
if ((dirPath == std::string("",0))) {
e.Path = e.Name;
}
else {
e.Path = ((dirPath + std::string("/",1)) + e.Name);
}
out = append(out,Slice<iso9660_Entry>{e});
}
}}
}return {out,{}};
}
}
// tools/lib/iso9660/iso9660.go:280:1
bool iso9660_nameEqual(std::string entry,std::string want){
{
if (go_strings_EqualFold(entry,want)) {
return true;
}
if ((!go_strings_Contains(want,std::string(";",1)))) {
{
int64_t i = go_strings_IndexByte(entry,cast<uint8_t>(59ULL));
if ((i >= cast<int64_t>(0ULL))) {
return go_strings_EqualFold(sub(entry,0,i),want);
}
}
}
return false;
}
}
// tools/lib/iso9660/iso9660.go:292:1
Slice<std::string> iso9660_splitPath(std::string p){
{
Slice<std::string> parts={};
{auto&& tmp3 = go_strings_Split(go_strings_Trim(p,std::string("/",1)),std::string("/",1));
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto c=tmp3[tmp4];if ((c != std::string("",0))) {
parts = append(parts,Slice<std::string>{c});
}
}}
return parts;
}
}
// tools/lib/iso9660/iso9660.go:309:1
std::tuple<iso9660_Entry,Error> iso9660_Volume_Resolve(iso9660_Volume* v,std::string path){
{
path = iso9660_normalisePath(path);
{
auto tmp5 = iso9660_parseLbnPath(path);
int64_t lbn = std::get<0>(tmp5);
int64_t size = std::get<1>(tmp5);
bool ok = std::get<2>(tmp5);
if (ok) {
return {iso9660_Entry{path,path,{},size,lbn},{}};
}
}
iso9660_Entry cur = iso9660_Entry{v->Name,std::string("",0),true,v->rootSize,v->rootLBA};
{auto&& tmp6 = iso9660_splitPath(path);
for(int64_t tmp7=0;tmp7<len(tmp6);++tmp7){
auto want=tmp6[tmp7];if ((!cur.IsDir)) {
return {iso9660_Entry{},go_fmt_Errorf(std::string("iso9660: %q is not a directory",30),cur.Path)};
}
auto tmp8 = iso9660_Volume_dirEntries(v,cur.Block,cur.Size,cur.Path);
Slice<iso9660_Entry> entries = std::get<0>(tmp8);
Error err = std::get<1>(tmp8);
if (bool(err)) {
return {iso9660_Entry{},err};
}
bool found = false;
{auto&& tmp9 = entries;
for(int64_t tmp10=0;tmp10<len(tmp9);++tmp10){
auto e=tmp9[tmp10];if (iso9660_nameEqual(e.Name,want)) {
auto tmp11 = std::make_tuple(e,true);
cur = std::get<0>(tmp11);
found = std::get<1>(tmp11);
break;
}
}}
if ((!found)) {
return {iso9660_Entry{},go_fmt_Errorf(std::string("iso9660: %q not found",21),path)};
}
}}
return {cur,{}};
}
}
// tools/lib/iso9660/iso9660.go:338:1
std::string iso9660_normalisePath(std::string p){
{
{
int64_t i = go_strings_IndexByte(p,cast<uint8_t>(58ULL));
if (((i >= cast<int64_t>(0ULL)) && (!go_strings_ContainsAny(sub(p,0,i),std::string("/\\",2))))) {
p = sub(p,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p));
}
}
return go_strings_ReplaceAll(p,std::string("\\",1),std::string("/",1));
}
}
// tools/lib/iso9660/iso9660.go:347:1
std::tuple<int64_t,int64_t,bool> iso9660_parseLbnPath(std::string path){
int64_t lbn{};
int64_t size{};
bool ok{};
{
std::string p = go_strings_ToLower(go_strings_Trim(path,std::string("/",1)));
if ((!go_strings_HasPrefix(p,std::string("sce_lbn",7)))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
std::string rest = sub(p,cast<int64_t>(7ULL),len(p));
int64_t i = go_strings_Index(rest,std::string("_size",5));
if ((i < cast<int64_t>(0ULL))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
auto tmp12 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,0,i),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t l = std::get<0>(tmp12);
Error err1 = std::get<1>(tmp12);
auto tmp13 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,cast<int64_t>((i + cast<int64_t>(5ULL))),len(rest)),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t s = std::get<0>(tmp13);
Error err2 = std::get<1>(tmp13);
if ((bool(err1) || bool(err2))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
return {cast<int64_t>(l),cast<int64_t>(s),true};
}
}
// tools/lib/iso9660/iso9660.go:366:1
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_ReadDir(iso9660_Volume* v,std::string path){
{
auto tmp14 = iso9660_Volume_Resolve(v,path);
iso9660_Entry e = std::get<0>(tmp14);
Error err = std::get<1>(tmp14);
if (bool(err)) {
return {{},err};
}
if ((!e.IsDir)) {
return {{},go_fmt_Errorf(std::string("iso9660: %q is not a directory",30),path)};
}
return iso9660_Volume_dirEntries(v,e.Block,e.Size,e.Path);
}
}
// tools/lib/iso9660/iso9660.go:378:1
Error iso9660_Volume_Walk(iso9660_Volume* v,std::function<Error(iso9660_Entry)> fn){
{
return iso9660_Volume_walk(v,v->rootLBA,v->rootSize,std::string("",0),fn);
}
}
// tools/lib/iso9660/iso9660.go:382:1
Error iso9660_Volume_walk(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(iso9660_Entry)> fn){
{
auto tmp15 = iso9660_Volume_dirEntries(v,lba,size,dirPath);
Slice<iso9660_Entry> entries = std::get<0>(tmp15);
Error err = std::get<1>(tmp15);
if (bool(err)) {
return err;
}
{auto&& tmp16 = entries;
for(int64_t tmp17=0;tmp17<len(tmp16);++tmp17){
auto e=tmp16[tmp17];{
Error err = fn(e);
if (bool(err)) {
return err;
}
}
if (e.IsDir) {
{
Error err = iso9660_Volume_walk(v,e.Block,e.Size,e.Path,fn);
if (bool(err)) {
return err;
}
}
}
}}
return {};
}
}
// tools/lib/iso9660/iso9660.go:402:1
std::tuple<int64_t,Error> iso9660_Volume_ReadFileAt(iso9660_Volume* v,iso9660_Entry e,int64_t off,Slice<uint8_t> p){
{
if ((off >= e.Size)) {
return {cast<int64_t>(0ULL),{}};
}
if ((cast<int64_t>((off + len(p))) > e.Size)) {
p = sub(p,0,cast<int64_t>((e.Size - off)));
}
int64_t got = cast<int64_t>(0ULL);
{;for (;(got < len(p));){
auto tmp18 = dc_Disc_ReadBlock(v->src,cast<int64_t>((e.Block + divi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL)))));
Slice<uint8_t> b = std::get<0>(tmp18);
Error err = std::get<1>(tmp18);
if (bool(err)) {
return {got,err};
}
got += gcopy(sub(p,got,len(p)),sub(b,modi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL)),len(b)));
}
}return {got,{}};
}
}
// tools/lib/iso9660/iso9660.go:421:1
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadFile(iso9660_Volume* v,std::string path){
{
auto tmp19 = iso9660_Volume_Resolve(v,path);
iso9660_Entry e = std::get<0>(tmp19);
Error err = std::get<1>(tmp19);
if (bool(err)) {
return {{},err};
}
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("iso9660: %q is a directory",26),path)};
}
Slice<uint8_t> out = Slice<uint8_t>::make(e.Size);
{
auto tmp20 = iso9660_Volume_ReadFileAt(v,e,cast<int64_t>(0ULL),out);
Error err = std::get<1>(tmp20);
if (bool(err)) {
return {{},err};
}
}
return {out,{}};
}
}
// tools/lib/iso9660/iso9660.go:439:1
std::tuple<iso9660_Entry,bool> iso9660_Volume_FileAt(iso9660_Volume* v,int64_t lba){
{
iso9660_Entry hit={};
bool found = false;
iso9660_Volume_Walk(v,[&](iso9660_Entry e)->Error{
if (e.IsDir) {
return {};
}
int64_t blocks = divi<int64_t>((cast<int64_t>((cast<int64_t>((e.Size + cast<int64_t>(2048ULL))) - cast<int64_t>(1ULL)))),cast<int64_t>(2048ULL));
if (((lba >= e.Block) && (lba < cast<int64_t>((e.Block + blocks))))) {
auto tmp21 = std::make_tuple(e,true);
hit = std::get<0>(tmp21);
found = std::get<1>(tmp21);
}
return {};
}
);
return {hit,found};
}
}
// tools/lib/iso9660/iso9660.go:458:1
iso9660_byteReaderAt iso9660_newByteReaderAt(Slice<uint8_t> b){
{
return cast<iso9660_byteReaderAt>(b);
}
}
// tools/lib/iso9660/iso9660.go:460:1
std::tuple<int64_t,Error> iso9660_byteReaderAt_ReadAt(iso9660_byteReaderAt b,Slice<uint8_t> p,int64_t off){
{
if (((off < cast<int64_t>(0ULL)) || (off >= cast<int64_t>(len(b))))) {
return {cast<int64_t>(0ULL),go_io_EOF};
}
int64_t n = gcopy(p,sub(b,off,len(b)));
if ((n < len(p))) {
return {n,go_io_EOF};
}
return {n,{}};
}
}
// tools/platform/dc/aica.go:48:1
uint32_t dc_timerReg(int64_t i){
{
return cast<uint32_t>((cast<uint32_t>(10384ULL) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL)))));
}
}
// tools/platform/dc/aica.go:76:1
uint32_t dc_Machine_slotW(dc_Machine* m,uint32_t slot,uint32_t off){
{
return get(m->AICARegs,cast<uint32_t>((cast<uint32_t>((slot * cast<uint32_t>(128ULL))) + off)));
}
}
// tools/platform/dc/aica.go:81:1
void dc_Machine_tickAICA(dc_Machine* m){
{
dc_aicaTimers* t = (&m->Timers);
t->Frac++;
if ((t->Frac < cast<uint32_t>(4535ULL))) {
return ;
}
t->Frac = cast<uint32_t>(0ULL);
dc_Machine_mixSample(m);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
uint32_t reg = get(m->AICARegs,dc_timerReg(i));
uint32_t pre = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
t->Sub[i]++;
if ((t->Sub[i] < shl<uint32_t>(cast<uint32_t>(1ULL),pre))) {
continue;
}
t->Sub[i] = cast<uint32_t>(0ULL);
uint32_t count = cast<uint32_t>((cast<uint32_t>((reg & cast<uint32_t>(255ULL))) + cast<uint32_t>(1ULL)));
if ((count > cast<uint32_t>(255ULL))) {
count = cast<uint32_t>(0ULL);
uint32_t bit = shl<uint32_t>(cast<uint32_t>(64ULL),i);
t->SCIPD |= bit;
t->MCIPD |= bit;
dc_Machine_aicaMainIRQ(m);
}
m->AICARegs[dc_timerReg(i)] = cast<uint32_t>(((reg & ~(cast<uint32_t>(255ULL))) | count));
}
}}
}
// tools/platform/dc/aica.go:113:1
void dc_Machine_aicaMainIRQ(dc_Machine* m){
{
if ((cast<uint32_t>((m->Timers.MCIPD & get(m->AICARegs,cast<uint32_t>(10420ULL)))) != cast<uint32_t>(0ULL))) {
m->Holly.ISTEXT |= cast<uint32_t>(2ULL);
}
else {
m->Holly.ISTEXT &= ~(cast<uint32_t>(2ULL));
}
dc_Machine_updateIRL(m);
}
}
// tools/platform/dc/aica.go:126:1
std::tuple<Slice<uint8_t>,uint32_t,bool> dc_armBus_region(dc_armBus b,uint32_t addr){
Slice<uint8_t> mem{};
uint32_t off{};
bool reg{};
{
uint32_t a = cast<uint32_t>((addr & cast<uint32_t>(16777215ULL)));
if ((a < cast<uint32_t>(8388608ULL))) {
return {b.m->AICARAM,cast<uint32_t>((a & cast<uint32_t>(2097151ULL))),false};
}
return {{},cast<uint32_t>((a - cast<uint32_t>(8388608ULL))),true};
}
}
// tools/platform/dc/aica.go:134:1
uint8_t dc_armBus_Read(dc_armBus b,uint32_t addr){
{
{
auto tmp1 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp1);
uint32_t off = std::get<1>(tmp1);
bool reg = std::get<2>(tmp1);
if ((!reg)) {
return mem[off];
}
else {
return cast<uint8_t>(shr<uint32_t>(dc_Machine_aicaRead(b.m,(off & ~(cast<uint32_t>(3ULL))),cast<int64_t>(4ULL),false),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL)))))))));
}
}
}
}
// tools/platform/dc/aica.go:142:1
void dc_armBus_Write(dc_armBus b,uint32_t addr,uint8_t v){
{
{
auto tmp2 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp2);
uint32_t off = std::get<1>(tmp2);
bool reg = std::get<2>(tmp2);
if ((!reg)) {
mem[off] = v;
}
else {
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL))))));
uint32_t old = dc_Machine_aicaRead(b.m,(off & ~(cast<uint32_t>(3ULL))),cast<int64_t>(4ULL),false);
dc_Machine_aicaWrite(b.m,(off & ~(cast<uint32_t>(3ULL))),cast<int64_t>(4ULL),cast<uint32_t>(((old & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | shl<uint32_t>(cast<uint32_t>(v),sh))),false);
}
}
}
}
// tools/platform/dc/aica.go:152:1
uint16_t dc_armBus_Read16(dc_armBus b,uint32_t addr){
{
{
auto tmp3 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp3);
uint32_t off = std::get<1>(tmp3);
bool reg = std::get<2>(tmp3);
if ((!reg)) {
return cast<uint16_t>((cast<uint16_t>(mem[off]) | shl<uint16_t>(cast<uint16_t>(mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
else {
return cast<uint16_t>(shr<uint32_t>(dc_Machine_aicaRead(b.m,(off & ~(cast<uint32_t>(3ULL))),cast<int64_t>(2ULL),false),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(2ULL)))))))));
}
}
}
}
// tools/platform/dc/aica.go:160:1
uint32_t dc_armBus_Read32_reference(dc_armBus b,uint32_t addr){
{
{
auto tmp4 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp4);
uint32_t off = std::get<1>(tmp4);
bool reg = std::get<2>(tmp4);
if ((!reg)) {
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(mem[off]) | shl<uint32_t>(cast<uint32_t>(mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(mem[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(mem[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
else {
return dc_Machine_aicaRead(b.m,off,cast<int64_t>(4ULL),false);
}
}
}
}
// tools/platform/dc/aica.go:168:1
void dc_armBus_Write16(dc_armBus b,uint32_t addr,uint16_t v){
{
{
auto tmp5 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp5);
uint32_t off = std::get<1>(tmp5);
bool reg = std::get<2>(tmp5);
if ((!reg)) {
auto tmp6 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
mem[off] = std::get<0>(tmp6);
mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp6);
}
else {
dc_Machine_aicaWrite(b.m,off,cast<int64_t>(2ULL),cast<uint32_t>(v),false);
}
}
}
}
// tools/platform/dc/aica.go:176:1
void dc_armBus_Write32(dc_armBus b,uint32_t addr,uint32_t v){
{
{
auto tmp7 = dc_armBus_region(b,addr);
Slice<uint8_t> mem = std::get<0>(tmp7);
uint32_t off = std::get<1>(tmp7);
bool reg = std::get<2>(tmp7);
if ((!reg)) {
auto tmp8 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
mem[off] = std::get<0>(tmp8);
mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp8);
mem[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = std::get<2>(tmp8);
mem[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = std::get<3>(tmp8);
}
else {
dc_Machine_aicaWrite(b.m,off,cast<int64_t>(4ULL),v,false);
}
}
}
}
// tools/platform/dc/aica.go:186:1
uint32_t dc_Machine_aicaRead(dc_Machine* m,uint32_t off,int64_t size,bool sh4){
{
{
switch(off){
case cast<uint32_t>(11264ULL):{
if (m->ARMRunning) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>(1ULL);
break;}
case cast<uint32_t>(10400ULL):{
return m->Timers.SCIPD;
break;}
case cast<uint32_t>(10424ULL):{
return m->Timers.MCIPD;
break;}
case cast<uint32_t>(10256ULL):{
dc_AICASlot* s = (&m->Slots[cast<uint32_t>((shr<uint32_t>(get(m->AICARegs,cast<uint32_t>(10252ULL)),cast<int64_t>(8ULL)) & cast<uint32_t>(63ULL)))]);
uint32_t v = cast<uint32_t>(0ULL);
if (s->LP) {
v |= cast<uint32_t>(32768ULL);
s->LP = false;
}
if (s->Active) {
v |= cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(s->EGState),cast<int64_t>(13ULL)) | shl<uint32_t>(cast<uint32_t>(s->EGLevel),cast<int64_t>(3ULL))));
}
else {
v |= cast<uint32_t>(32767ULL);
}
return v;
break;}
case cast<uint32_t>(10260ULL):{
return cast<uint32_t>((m->Slots[cast<uint32_t>((shr<uint32_t>(get(m->AICARegs,cast<uint32_t>(10252ULL)),cast<int64_t>(8ULL)) & cast<uint32_t>(63ULL)))].Pos & cast<uint32_t>(65535ULL)));
break;}
}}
{
auto tmp9 = lookup(m->AICARegs,(off & ~(cast<uint32_t>(3ULL))));
uint32_t v = std::get<0>(tmp9);
bool ok = std::get<1>(tmp9);
if (ok) {
return v;
}
}
dc_Machine_logf(m,std::string("AICA read %04X by %s (never written)",36),off,dc_side(sh4));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/dc/aica.go:222:1
void dc_Machine_aicaWrite(dc_Machine* m,uint32_t off,int64_t size,uint32_t v,bool sh4){
{
{
switch(off){
case cast<uint32_t>(11264ULL):{
if ((!sh4)) {
dc_Machine_logf(m,std::string("AICA ARM wrote its own reset register",37));
return ;
}
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
m->ARMRunning = false;
return ;
}
if ((!m->ARMRunning)) {
arm_CPU_Reset(m->ARM);
m->ARM->R[cast<int64_t>(15ULL)] = cast<uint32_t>(0ULL);
m->ARMRunning = true;
}
return ;
break;}
case cast<uint32_t>(10404ULL):{
m->Timers.SCIPD &= ~(v);
return ;
break;}
case cast<uint32_t>(10428ULL):{
m->Timers.MCIPD &= ~(v);
dc_Machine_aicaMainIRQ(m);
return ;
break;}
case cast<uint32_t>(10420ULL):{
m->AICARegs[off] = v;
dc_Machine_aicaMainIRQ(m);
return ;
break;}
}}
{
uint32_t a = (off & ~(cast<uint32_t>(3ULL)));
if ((((a < cast<uint32_t>(8192ULL)) && (cast<uint32_t>((a & cast<uint32_t>(127ULL))) == cast<uint32_t>(0ULL))) && (cast<uint32_t>((off & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL)))) {
m->AICARegs[a] = (v & ~(cast<uint32_t>(32768ULL)));
if ((cast<uint32_t>((v & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
{auto&& tmp10 = m->Slots;
for(int64_t tmp11=0;tmp11<len(tmp10);++tmp11){
auto i=tmp11;dc_AICASlot* s = (&m->Slots[i]);
bool on = (cast<uint32_t>((shr<uint32_t>(get(m->AICARegs,cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(128ULL)))),cast<int64_t>(14ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
{
if ((on && (((!s->Active) || (s->EGState == cast<uint8_t>(3ULL)))))){
dc_AICASlot_keyOn(s);
}
else if (((!on) && s->Active)){
dc_AICASlot_keyOff(s);
}
}
tmp12:;
}}
}
return ;
}
}
m->AICARegs[(off & ~(cast<uint32_t>(3ULL)))] = v;
}
}
// tools/platform/dc/aica.go:277:1
std::string dc_side(bool sh4){
{
if (sh4) {
return std::string("sh4",3);
}
return std::string("arm",3);
}
}
// tools/platform/dc/aica.go:287:1
void dc_Machine_stepARM(dc_Machine* m){
{
if (((!m->ARMRunning) || m->ARM->Halted)) {
return ;
}
arm_CPU_Step(m->ARM);
if (m->ARM->Halted) {
dc_Machine_logf(m,std::string("AICA ARM halted: %s",19),m->ARM->HaltReason);
}
}
}
// tools/platform/dc/aica.go:303:1
uint32_t dc_Machine_rtcRead(dc_Machine* m,uint32_t addr){
{
{
switch(addr){
case cast<uint32_t>(7405568ULL):{
return cast<uint32_t>(24292ULL);
break;}
case cast<uint32_t>(7405572ULL):{
return cast<uint32_t>(0ULL);
break;}
}}
dc_Machine_logf(m,std::string("RTC read %08X",13),addr);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/dc/aica_synth.go:47:1
void dc_init_aica_synth(){
{
auto pow = [&](double db)->int64_t{
double g = 1.0;
{double i = 0.0;for (;(i < db);i++){
g *= 0.8912509381337456;
}
}return cast<int64_t>((g * cast<double>(65536ULL)));
}
;
{auto&& tmp13 = dc_egGainQ16;
for(int64_t tmp14=0;tmp14<len(tmp13);++tmp14){
auto i=tmp14;dc_egGainQ16[i] = pow(((96.0 * cast<double>(i)) / cast<double>(cast<double>(1023ULL))));
}}
{auto&& tmp15 = dc_tlGainQ16;
for(int64_t tmp16=0;tmp16<len(tmp15);++tmp16){
auto i=tmp16;dc_tlGainQ16[i] = pow((0.375 * cast<double>(i)));
}}
{auto&& tmp17 = dc_sdlGainQ16;
for(int64_t tmp18=0;tmp18<len(tmp17);++tmp18){
auto i=tmp18;if ((i == cast<int64_t>(0ULL))) {
continue;
}
dc_sdlGainQ16[i] = pow((3.0 * cast<double>(cast<int64_t>((cast<int64_t>(15ULL) - i)))));
}}
{auto&& tmp19 = dc_panGainQ16;
for(int64_t tmp20=0;tmp20<len(tmp19);++tmp20){
auto i=tmp20;if ((i == cast<int64_t>(15ULL))) {
continue;
}
dc_panGainQ16[i] = pow((3.0 * cast<double>(i)));
}}
}
}
// tools/platform/dc/aica_synth.go:82:1
void dc_AICASlot_keyOn(dc_AICASlot* s){
{
(*s) = dc_AICASlot{true,{},{},{},cast<uint8_t>(0ULL),cast<uint16_t>(1023ULL),{},{},cast<uint32_t>(4294967295ULL),{},cast<int32_t>(127ULL),{},{},{}};
}
}
// tools/platform/dc/aica_synth.go:89:1
void dc_AICASlot_keyOff(dc_AICASlot* s){
{
if (s->Active) {
s->EGState = cast<uint8_t>(3ULL);
}
}
}
// tools/platform/dc/aica_synth.go:98:1
uint32_t dc_effRate(uint32_t r,uint32_t krs){
{
if ((r == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
uint32_t er = cast<uint32_t>((cast<uint32_t>(2ULL) * r));
if ((krs != cast<uint32_t>(15ULL))) {
er += krs;
}
if ((er > cast<uint32_t>(63ULL))) {
er = cast<uint32_t>(63ULL);
}
return er;
}
}
// tools/platform/dc/aica_synth.go:114:1
uint64_t dc_egPeriod(uint32_t er){
{
uint64_t p = shr<uint64_t>(cast<uint64_t>(4096ULL),(shr<uint32_t>(er,cast<int64_t>(2ULL))));
if ((p == cast<uint64_t>(0ULL))) {
return cast<uint64_t>(1ULL);
}
return p;
}
}
// tools/platform/dc/aica_synth.go:123:1
void dc_Machine_stepEG(dc_Machine* m,dc_AICASlot* s,uint32_t slot,uint64_t tick){
{
uint32_t env = dc_Machine_slotW(m,slot,cast<uint32_t>(16ULL));
uint32_t env2 = dc_Machine_slotW(m,slot,cast<uint32_t>(20ULL));
uint32_t krs = cast<uint32_t>((shr<uint32_t>(env2,cast<int64_t>(10ULL)) & cast<uint32_t>(15ULL)));
uint32_t r={};
{
switch(s->EGState){
case cast<uint8_t>(0ULL):{
r = cast<uint32_t>((env & cast<uint32_t>(31ULL)));
break;}
case cast<uint8_t>(1ULL):{
r = cast<uint32_t>((shr<uint32_t>(env,cast<int64_t>(6ULL)) & cast<uint32_t>(31ULL)));
break;}
case cast<uint8_t>(2ULL):{
r = cast<uint32_t>((shr<uint32_t>(env,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL)));
break;}
case cast<uint8_t>(3ULL):{
r = cast<uint32_t>((env2 & cast<uint32_t>(31ULL)));
break;}
}}
uint32_t er = dc_effRate(r,krs);
if ((er == cast<uint32_t>(0ULL))) {
return ;
}
if (((s->EGState == cast<uint8_t>(0ULL)) && (er >= cast<uint32_t>(60ULL)))) {
auto tmp21 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint8_t>(1ULL));
s->EGLevel = std::get<0>(tmp21);
s->EGState = std::get<1>(tmp21);
return ;
}
if ((modi<uint64_t>(tick,dc_egPeriod(er)) != cast<uint64_t>(0ULL))) {
return ;
}
{
switch(s->EGState){
case cast<uint8_t>(0ULL):{
uint16_t step = cast<uint16_t>((cast<uint16_t>(shr<uint16_t>(s->EGLevel,cast<int64_t>(4ULL))) + cast<uint16_t>(1ULL)));
if ((s->EGLevel <= step)) {
auto tmp22 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint8_t>(1ULL));
s->EGLevel = std::get<0>(tmp22);
s->EGState = std::get<1>(tmp22);
}
else {
s->EGLevel -= step;
}
break;}
case cast<uint8_t>(1ULL):{
if ((s->EGLevel < cast<uint16_t>(1023ULL))) {
s->EGLevel++;
}
{
uint16_t dl = cast<uint16_t>(shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(env2,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(5ULL)));
if ((s->EGLevel >= dl)) {
s->EGState = cast<uint8_t>(2ULL);
}
}
break;}
case cast<uint8_t>(2ULL):{
if ((s->EGLevel < cast<uint16_t>(1023ULL))) {
s->EGLevel++;
}
break;}
case cast<uint8_t>(3ULL):{
if ((s->EGLevel < cast<uint16_t>(1023ULL))) {
s->EGLevel++;
}
else {
s->Active = false;
}
break;}
}}
}
}
// tools/platform/dc/aica_synth.go:181:1
uint32_t dc_Machine_sampleAddr(dc_Machine* m,uint32_t slot){
{
return cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((dc_Machine_slotW(m,slot,cast<uint32_t>(0ULL)) & cast<uint32_t>(127ULL)))),cast<int64_t>(16ULL)) | cast<uint32_t>((dc_Machine_slotW(m,slot,cast<uint32_t>(4ULL)) & cast<uint32_t>(65535ULL)))));
}
}
// tools/platform/dc/aica_synth.go:186:1
int32_t dc_Machine_fetchPCM(dc_Machine* m,uint32_t sa,uint32_t n,bool eight){
{
if (eight) {
return shl<int32_t>(cast<int32_t>(cast<int8_t>(m->AICARAM[cast<uint32_t>(((cast<uint32_t>((sa + n))) & cast<uint32_t>(2097151ULL)))])),cast<int64_t>(8ULL));
}
uint32_t a = cast<uint32_t>(((cast<uint32_t>((sa + cast<uint32_t>((cast<uint32_t>(2ULL) * n))))) & cast<uint32_t>(2097151ULL)));
return cast<int32_t>(cast<int16_t>(cast<uint16_t>((cast<uint16_t>(m->AICARAM[a]) | shl<uint16_t>(cast<uint16_t>(m->AICARAM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(1ULL)))) & cast<uint32_t>(2097151ULL)))]),cast<int64_t>(8ULL))))));
}
}
// tools/platform/dc/aica_synth.go:198:1
void dc_Machine_decodeADPCMTo(dc_Machine* m,dc_AICASlot* s,uint32_t sa,uint32_t target,uint32_t lsa){
{
if ((s->AdStep == cast<int32_t>(0ULL))) {
s->AdStep = cast<int32_t>(127ULL);
}
{;for (;(s->DecPos != target);){
uint32_t n = cast<uint32_t>((s->DecPos + cast<uint32_t>(1ULL)));
uint8_t b = m->AICARAM[cast<uint32_t>(((cast<uint32_t>((sa + divi<uint32_t>(n,cast<uint32_t>(2ULL))))) & cast<uint32_t>(2097151ULL)))];
uint8_t nib = cast<uint8_t>((shr<uint8_t>(b,(cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((n & cast<uint32_t>(1ULL)))))))) & cast<uint8_t>(15ULL)));
int32_t delta = divi<int32_t>(cast<int32_t>((s->AdStep * cast<int32_t>(cast<uint8_t>((cast<uint8_t>((cast<uint8_t>(2ULL) * (cast<uint8_t>((nib & cast<uint8_t>(7ULL)))))) + cast<uint8_t>(1ULL)))))),cast<int32_t>(8ULL));
if ((cast<uint8_t>((nib & cast<uint8_t>(8ULL))) != cast<uint8_t>(0ULL))) {
delta = cast<int32_t>(-delta);
}
int32_t h = cast<int32_t>((s->AdHist + delta));
if ((h > cast<int32_t>(32767ULL))) {
h = cast<int32_t>(32767ULL);
}
else if ((h < cast<int32_t>(-32768ULL))) {
h = cast<int32_t>(-32768ULL);
}
s->AdHist = h;
s->AdStep = divi<int32_t>(cast<int32_t>((s->AdStep * dc_adpcmScale[cast<uint8_t>((nib & cast<uint8_t>(7ULL)))])),cast<int32_t>(256ULL));
if ((s->AdStep < cast<int32_t>(127ULL))) {
s->AdStep = cast<int32_t>(127ULL);
}
else if ((s->AdStep > cast<int32_t>(24576ULL))) {
s->AdStep = cast<int32_t>(24576ULL);
}
auto tmp23 = std::make_tuple(s->Cur,h);
s->Prev = std::get<0>(tmp23);
s->Cur = std::get<1>(tmp23);
s->DecPos = n;
if (((n == lsa) && (!s->LoopSeen))) {
auto tmp24 = std::make_tuple(true,s->AdHist,s->AdStep);
s->LoopSeen = std::get<0>(tmp24);
s->LoopHist = std::get<1>(tmp24);
s->LoopStep = std::get<2>(tmp24);
}
}
}}
}
// tools/platform/dc/aica_synth.go:235:1
void dc_Machine_mixSample(dc_Machine* m){
{rrprof::Scope timing(3,"AICA synthesis");
{
uint64_t tick = m->Timers.SampleTick;
m->Timers.SampleTick++;
int64_t accL={};
int64_t accR={};
{auto&& tmp25 = m->Slots;
for(int64_t tmp26=0;tmp26<len(tmp25);++tmp26){
auto i=tmp26;dc_AICASlot* s = (&m->Slots[i]);
if ((!s->Active)) {
continue;
}
uint32_t slot = cast<uint32_t>(i);
uint32_t ctl = dc_Machine_slotW(m,slot,cast<uint32_t>(0ULL));
uint32_t pitch = dc_Machine_slotW(m,slot,cast<uint32_t>(24ULL));
int32_t oct = cast<int32_t>((cast<int32_t>(shr<uint32_t>(pitch,cast<int64_t>(11ULL))) & cast<int32_t>(15ULL)));
if ((oct >= cast<int32_t>(8ULL))) {
oct -= cast<int32_t>(16ULL);
}
uint64_t incr = cast<uint64_t>(cast<uint32_t>((cast<uint32_t>(1024ULL) + cast<uint32_t>((pitch & cast<uint32_t>(1023ULL))))));
if ((oct >= cast<int32_t>(0ULL))) {
incr = shl<uint64_t>(incr,cast<uint64_t>(oct));
}
else {
incr = shr<uint64_t>(incr,cast<uint64_t>(cast<int32_t>(-oct)));
}
uint64_t acc = cast<uint64_t>((cast<uint64_t>(s->Frac) + incr));
s->Frac = cast<uint32_t>(cast<uint64_t>((acc & cast<uint64_t>(1023ULL))));
s->Pos += cast<uint32_t>(shr<uint64_t>(acc,cast<int64_t>(10ULL)));
uint32_t lea = cast<uint32_t>((dc_Machine_slotW(m,slot,cast<uint32_t>(12ULL)) & cast<uint32_t>(65535ULL)));
uint32_t lsa = cast<uint32_t>((dc_Machine_slotW(m,slot,cast<uint32_t>(8ULL)) & cast<uint32_t>(65535ULL)));
bool wrapped = false;
if ((s->Pos >= lea)) {
if ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(9ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
if ((lea > lsa)) {
s->Pos = cast<uint32_t>((lsa + modi<uint32_t>((cast<uint32_t>((s->Pos - lea))),(cast<uint32_t>((lea - lsa))))));
}
else {
s->Pos = lsa;
}
s->LP = true;
wrapped = true;
}
else {
s->Pos = lea;
s->Active = false;
continue;
}
}
uint32_t sa = dc_Machine_sampleAddr(m,slot);
if ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(10ULL)) & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
dc_Machine_logf(m,std::string("AICA slot SSCTL %d (noise source) unimplemented",47),cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(10ULL)) & cast<uint32_t>(3ULL))));
}
else {
{
uint32_t fmtBits = cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(7ULL)) & cast<uint32_t>(3ULL)));
switch(fmtBits){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
if ((s->DecPos != s->Pos)) {
if ((s->Pos == cast<uint32_t>((s->DecPos + cast<uint32_t>(1ULL))))) {
s->Prev = s->Cur;
}
else if ((s->Pos > cast<uint32_t>(0ULL))) {
s->Prev = dc_Machine_fetchPCM(m,sa,cast<uint32_t>((s->Pos - cast<uint32_t>(1ULL))),(fmtBits == cast<uint32_t>(1ULL)));
}
s->Cur = dc_Machine_fetchPCM(m,sa,s->Pos,(fmtBits == cast<uint32_t>(1ULL)));
s->DecPos = s->Pos;
}
break;}
default:{
if ((wrapped || ((s->Pos < s->DecPos) && (s->DecPos != cast<uint32_t>(4294967295ULL))))) {
if (((fmtBits == cast<uint32_t>(2ULL)) && s->LoopSeen)) {
auto tmp27 = std::make_tuple(s->LoopHist,s->LoopStep);
s->AdHist = std::get<0>(tmp27);
s->AdStep = std::get<1>(tmp27);
}
s->DecPos = cast<uint32_t>(4294967295ULL);
if ((lsa > cast<uint32_t>(0ULL))) {
s->DecPos = cast<uint32_t>((lsa - cast<uint32_t>(1ULL)));
}
if ((s->Pos < lsa)) {
s->DecPos = cast<uint32_t>(4294967295ULL);
}
}
dc_Machine_decodeADPCMTo(m,s,sa,s->Pos,lsa);
break;}
}}
}
dc_Machine_stepEG(m,s,slot,tick);
int64_t out = cast<int64_t>((cast<int64_t>(s->Prev) + divi<int64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>((s->Cur - s->Prev))) * cast<int64_t>(s->Frac))),cast<int64_t>(1024ULL))));
out = shr<int64_t>(cast<int64_t>((out * dc_egGainQ16[s->EGLevel])),cast<int64_t>(16ULL));
out = shr<int64_t>(cast<int64_t>((out * dc_tlGainQ16[cast<uint32_t>((shr<uint32_t>(dc_Machine_slotW(m,slot,cast<uint32_t>(40ULL)),cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))])),cast<int64_t>(16ULL));
uint32_t send = dc_Machine_slotW(m,slot,cast<uint32_t>(36ULL));
uint32_t disdl = cast<uint32_t>((shr<uint32_t>(send,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL)));
if ((disdl == cast<uint32_t>(0ULL))) {
continue;
}
out = shr<int64_t>(cast<int64_t>((out * dc_sdlGainQ16[disdl])),cast<int64_t>(16ULL));
uint32_t pan = cast<uint32_t>((send & cast<uint32_t>(31ULL)));
auto tmp28 = std::make_tuple(out,out);
int64_t l = std::get<0>(tmp28);
int64_t r = std::get<1>(tmp28);
if ((cast<uint32_t>((pan & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
r = shr<int64_t>(cast<int64_t>((r * dc_panGainQ16[cast<uint32_t>((pan & cast<uint32_t>(15ULL)))])),cast<int64_t>(16ULL));
}
else {
l = shr<int64_t>(cast<int64_t>((l * dc_panGainQ16[cast<uint32_t>((pan & cast<uint32_t>(15ULL)))])),cast<int64_t>(16ULL));
}
accL += l;
accR += r;
}}
if ((!m->AudioCapture)) {
return ;
}
int64_t mvol = dc_sdlGainQ16[cast<uint32_t>((get(m->AICARegs,cast<uint32_t>(10240ULL)) & cast<uint32_t>(15ULL)))];
accL = shr<int64_t>(cast<int64_t>((accL * mvol)),cast<int64_t>(16ULL));
accR = shr<int64_t>(cast<int64_t>((accR * mvol)),cast<int64_t>(16ULL));
m->AudioPCM = append(m->AudioPCM,Slice<int16_t>{dc_clamp16(accL),dc_clamp16(accR)});
}
}
}
// tools/platform/dc/aica_synth.go:346:1
int16_t dc_clamp16(int64_t v){
{
if ((v > cast<int64_t>(32767ULL))) {
return cast<int16_t>(32767ULL);
}
if ((v < cast<int64_t>(-32768ULL))) {
return cast<int16_t>(-32768ULL);
}
return cast<int16_t>(v);
}
}
// tools/platform/dc/bios.go:50:1
dc_biosHLE* dc_newBIOS(dc_Machine* m){
{
return arenaNew(dc_biosHLE{m,dc_biosState{cast<uint32_t>(1ULL),Map<uint32_t,dc_gdRequest*>{}},Map<std::string,int64_t>{}});
}
}
// tools/platform/dc/bios.go:60:1
bool dc_Machine_trapPC(dc_Machine* m,uint32_t pc){
{
if ((((!m->bios) || (pc < cast<uint32_t>(2684354816ULL))) || (pc > cast<uint32_t>(2684354832ULL)))) {
return false;
}
sh4_CPU* c = m->CPU;
if (sh4_CPU_NextIsDelaySlot(c)) {
return false;
}
dc_biosHLE* b = m->bios;
{
switch(pc){
case cast<uint32_t>(2684354816ULL):{
dc_biosHLE_sysinfo(b);
break;}
case cast<uint32_t>(2684354820ULL):{
dc_biosHLE_count(b,std::string("romfont",7));
dc_Machine_logf(m,std::string("romfont syscall (r1=%d) unimplemented; returning 0",50),c->R[cast<int64_t>(1ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(2684354824ULL):{
dc_biosHLE_flashrom(b);
break;}
case cast<uint32_t>(2684354828ULL):{
dc_biosHLE_gdrom(b);
break;}
case cast<uint32_t>(2684354832ULL):{
sh4_CPU_Halt(c,std::string("game called the BIOS menu vector (8C0000E0) \342\200\224 exit to BIOS",60));
return true;
break;}
}}
sh4_CPU_SetPC(c,c->PR);
return true;
}
}
// tools/platform/dc/bios.go:88:1
void dc_biosHLE_count(dc_biosHLE* b,std::string name){
{
b->calls[name]++;
}
}
// tools/platform/dc/bios.go:91:1
Slice<std::string> dc_biosHLE_census(dc_biosHLE* b){
{
Slice<std::string> names={};
{auto&& tmp29 = b->calls;
for(auto [tmp30,tmp31]:tmp29){
auto k=tmp30;names = append(names,Slice<std::string>{k});
}}
go_sort_Strings(names);
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(names));
{auto&& tmp32 = names;
for(int64_t tmp33=0;tmp33<len(tmp32);++tmp33){
auto n=tmp32[tmp33];out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("syscall %s x%d",14),n,get(b->calls,n))});
}}
return out;
}
}
// tools/platform/dc/bios.go:105:1
void dc_biosHLE_sysinfo(dc_biosHLE* b){
{
sh4_CPU* c = b->m->CPU;
{
switch(c->R[cast<int64_t>(7ULL)]){
case cast<uint32_t>(0ULL):{
dc_biosHLE_count(b,std::string("sysinfo.init",12));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(3ULL):{
dc_biosHLE_count(b,std::string("sysinfo.id",10));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(2348810344ULL);
break;}
default:{
dc_biosHLE_count(b,std::string("sysinfo.?",9));
dc_Machine_logf(b->m,std::string("sysinfo syscall r7=%d unimplemented (r4=%08X r5=%08X)",53),c->R[cast<int64_t>(7ULL)],c->R[cast<int64_t>(4ULL)],c->R[cast<int64_t>(5ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
}}
}
}
// tools/platform/dc/bios.go:123:1
void dc_biosHLE_flashrom(dc_biosHLE* b){
{
sh4_CPU* c = b->m->CPU;
{
switch(c->R[cast<int64_t>(7ULL)]){
case cast<uint32_t>(0ULL):{
dc_biosHLE_count(b,std::string("flashrom.info",13));
std::array<std::array<uint32_t,2>,5> parts = std::array<std::array<uint32_t,2>,5>{std::array<uint32_t,2>{cast<uint32_t>(106496ULL),cast<uint32_t>(8192ULL)},std::array<uint32_t,2>{cast<uint32_t>(98304ULL),cast<uint32_t>(8192ULL)},std::array<uint32_t,2>{cast<uint32_t>(114688ULL),cast<uint32_t>(16384ULL)},std::array<uint32_t,2>{cast<uint32_t>(65536ULL),cast<uint32_t>(32768ULL)},std::array<uint32_t,2>{cast<uint32_t>(0ULL),cast<uint32_t>(65536ULL)}};
{
uint32_t p = c->R[cast<int64_t>(4ULL)];
if ((p < cast<uint32_t>(5ULL))) {
dc_Machine_Write32(b->m,cast<uint32_t>((c->R[cast<int64_t>(5ULL)] & cast<uint32_t>(536870911ULL))),parts[p][cast<int64_t>(0ULL)]);
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((c->R[cast<int64_t>(5ULL)] + cast<uint32_t>(4ULL)))) & cast<uint32_t>(536870911ULL))),parts[p][cast<int64_t>(1ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
else {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(4294967295ULL);
}
}
break;}
case cast<uint32_t>(1ULL):{
dc_biosHLE_count(b,std::string("flashrom.read",13));
auto tmp34 = std::make_tuple(c->R[cast<int64_t>(4ULL)],c->R[cast<int64_t>(5ULL)],c->R[cast<int64_t>(6ULL)]);
uint32_t off = std::get<0>(tmp34);
uint32_t buf = std::get<1>(tmp34);
uint32_t n = std::get<2>(tmp34);
if ((off >= cast<uint32_t>(262144ULL))) {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(4294967295ULL);
return ;
}
if ((cast<uint32_t>((off + n)) > cast<uint32_t>(262144ULL))) {
n = cast<uint32_t>((cast<uint32_t>(262144ULL) - off));
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
dc_Machine_Write8(b->m,cast<uint32_t>(((cast<uint32_t>((buf + i))) & cast<uint32_t>(536870911ULL))),b->m->Flash[cast<uint32_t>((off + i))]);
}
}c->R[cast<int64_t>(0ULL)] = n;
break;}
default:{
dc_biosHLE_count(b,std::string("flashrom.?",10));
dc_Machine_logf(b->m,std::string("flashrom syscall r7=%d unimplemented (r4=%08X r5=%08X r6=%08X)",62),c->R[cast<int64_t>(7ULL)],c->R[cast<int64_t>(4ULL)],c->R[cast<int64_t>(5ULL)],c->R[cast<int64_t>(6ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(4294967295ULL);
break;}
}}
}
}
// tools/platform/dc/bios.go:164:1
void dc_biosHLE_gdrom(dc_biosHLE* b){
{
sh4_CPU* c = b->m->CPU;
if ((c->R[cast<int64_t>(6ULL)] == cast<uint32_t>(4294967295ULL))) {
dc_biosHLE_count(b,std::string("gdrom.misc",10));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
return ;
}
if ((c->R[cast<int64_t>(6ULL)] != cast<uint32_t>(0ULL))) {
dc_biosHLE_count(b,std::string("gdrom.?",7));
dc_Machine_logf(b->m,std::string("vector BC with r6=%08X unimplemented",36),c->R[cast<int64_t>(6ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
return ;
}
{
switch(c->R[cast<int64_t>(7ULL)]){
case cast<uint32_t>(0ULL):{
dc_biosHLE_count(b,go_fmt_Sprintf(std::string("gdrom.send.%d",13),c->R[cast<int64_t>(4ULL)]));
dc_gdRequest* req = arenaNew(dc_gdRequest{c->R[cast<int64_t>(4ULL)],{},{},{}});
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
req->Params[i] = dc_Machine_ram32(b->m,cast<uint32_t>((c->R[cast<int64_t>(5ULL)] + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))));
}
}uint32_t id = b->state.NextID;
b->state.NextID++;
b->state.Requests[id] = req;
c->R[cast<int64_t>(0ULL)] = id;
break;}
case cast<uint32_t>(1ULL):{
dc_biosHLE_count(b,std::string("gdrom.check",11));
auto tmp35 = lookup(b->state.Requests,c->R[cast<int64_t>(4ULL)]);
dc_gdRequest* req = std::get<0>(tmp35);
bool ok = std::get<1>(tmp35);
{
if ((!ok)){
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
else if ((!req->Done)){
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(1ULL);
}
else {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((c->R[cast<int64_t>(5ULL)] + cast<uint32_t>((cast<uint32_t>(4ULL) * i))))) & cast<uint32_t>(536870911ULL))),req->Result[i]);
}
}removeKey(b->state.Requests,c->R[cast<int64_t>(4ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(2ULL);
}
}
tmp36:;
break;}
case cast<uint32_t>(2ULL):{
dc_biosHLE_count(b,std::string("gdrom.exec",10));
{auto&& tmp37 = b->state.Requests;
for(auto [tmp38,tmp39]:tmp37){
auto req=tmp39;if ((!req->Done)) {
dc_biosHLE_execGD(b,req);
}
}}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(3ULL):{
dc_biosHLE_count(b,std::string("gdrom.init",10));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(4ULL):{
dc_biosHLE_count(b,std::string("gdrom.checkdrive",16));
dc_Machine_Write32(b->m,cast<uint32_t>((c->R[cast<int64_t>(4ULL)] & cast<uint32_t>(536870911ULL))),cast<uint32_t>(1ULL));
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((c->R[cast<int64_t>(4ULL)] + cast<uint32_t>(4ULL)))) & cast<uint32_t>(536870911ULL))),cast<uint32_t>(128ULL));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
default:{
dc_biosHLE_count(b,go_fmt_Sprintf(std::string("gdrom.?%d",9),c->R[cast<int64_t>(7ULL)]));
dc_Machine_logf(b->m,std::string("gdrom syscall r7=%d unimplemented (r4=%08X r5=%08X)",51),c->R[cast<int64_t>(7ULL)],c->R[cast<int64_t>(4ULL)],c->R[cast<int64_t>(5ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
}}
}
}
// tools/platform/dc/bios.go:229:1
void dc_biosHLE_execGD(dc_biosHLE* b,dc_gdRequest* req){
{rrprof::Scope timing(2,"GD-ROM HLE");
{
{
switch(req->Cmd){
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):{
auto tmp40 = std::make_tuple(req->Params[cast<int64_t>(0ULL)],req->Params[cast<int64_t>(1ULL)],req->Params[cast<int64_t>(2ULL)]);
uint32_t fad = std::get<0>(tmp40);
uint32_t count = std::get<1>(tmp40);
uint32_t buf = std::get<2>(tmp40);
if (bool(b->m->OnGDRead)) {
b->m->OnGDRead(fad,count,buf);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
auto tmp41 = dc_Disc_ReadSector(b->m->Disc,cast<int64_t>((cast<int64_t>(cast<uint32_t>((fad - cast<uint32_t>(150ULL)))) + cast<int64_t>(i))));
Slice<uint8_t> sec = std::get<0>(tmp41);
Error err = std::get<1>(tmp41);
if (bool(err)) {
dc_Machine_logf(b->m,std::string("gdrom read FAD %d failed: %v",28),cast<uint32_t>((fad + i)),err);
req->Result[cast<int64_t>(0ULL)] = cast<uint32_t>(4294967295ULL);
break;
}
{auto&& tmp42 = sec;
for(int64_t tmp43=0;tmp43<len(tmp42);++tmp43){
auto j=tmp43;auto by=tmp42[tmp43];dc_Machine_Write8(b->m,cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((buf + cast<uint32_t>((i * cast<uint32_t>(2048ULL))))) + cast<uint32_t>(j)))) & cast<uint32_t>(536870911ULL))),by);
}}
}
}break;}
case cast<uint32_t>(19ULL):{
uint32_t buf = req->Params[cast<int64_t>(1ULL)];
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(99ULL));i++){
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((buf + cast<uint32_t>((cast<uint32_t>(4ULL) * i))))) & cast<uint32_t>(536870911ULL))),cast<uint32_t>(4294967295ULL));
}
}dc_Disc* d = b->m->Disc;
int64_t first={};
int64_t last={};
{auto&& tmp44 = d->Tracks;
for(int64_t tmp45=0;tmp45<len(tmp44);++tmp45){
auto t=tmp44[tmp45];if (((!dc_Track_IsData(t)) || (t.StartLBA < cast<int64_t>(0ULL)))) {
continue;
}
if ((first == cast<int64_t>(0ULL))) {
first = t.Number;
}
last = t.Number;
uint32_t entry = cast<uint32_t>((cast<uint32_t>(1090519040ULL) | cast<uint32_t>((cast<uint32_t>(cast<int64_t>((t.StartLBA + cast<int64_t>(150ULL)))) & cast<uint32_t>(16777215ULL)))));
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((buf + cast<uint32_t>((cast<uint32_t>(4ULL) * cast<uint32_t>(cast<int64_t>((t.Number - cast<int64_t>(1ULL))))))))) & cast<uint32_t>(536870911ULL))),entry);
}}
int64_t end = cast<int64_t>((cast<int64_t>((d->data.StartLBA + cast<int64_t>(divi<int64_t>(d->data.Length,cast<int64_t>(2352ULL))))) + cast<int64_t>(150ULL)));
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((buf + cast<uint32_t>(396ULL)))) & cast<uint32_t>(536870911ULL))),cast<uint32_t>((cast<uint32_t>(1073741824ULL) | shl<uint32_t>(cast<uint32_t>(first),cast<int64_t>(16ULL)))));
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((buf + cast<uint32_t>(400ULL)))) & cast<uint32_t>(536870911ULL))),cast<uint32_t>((cast<uint32_t>(1073741824ULL) | shl<uint32_t>(cast<uint32_t>(last),cast<int64_t>(16ULL)))));
dc_Machine_Write32(b->m,cast<uint32_t>(((cast<uint32_t>((buf + cast<uint32_t>(404ULL)))) & cast<uint32_t>(536870911ULL))),cast<uint32_t>((cast<uint32_t>(end) & cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(24ULL):{
break;}
default:{
dc_Machine_logf(b->m,std::string("gdrom command %d unimplemented (params %08X %08X %08X %08X)",59),req->Cmd,req->Params[cast<int64_t>(0ULL)],req->Params[cast<int64_t>(1ULL)],req->Params[cast<int64_t>(2ULL)],req->Params[cast<int64_t>(3ULL)]);
break;}
}}
req->Done = true;
}
}
}
// tools/platform/dc/boot.go:28:1
Error dc_Machine_Boot(dc_Machine* m){
{
if ((!m->Disc)) {
return go_fmt_Errorf(std::string("boot: no disc mounted",21));
}
std::string name = dc_Disc_BootFilePath(m->Disc);
auto tmp46 = iso9660_Volume_ReadFile(m->Disc->Vol,name);
Slice<uint8_t> bin = std::get<0>(tmp46);
Error err = std::get<1>(tmp46);
if (bool(err)) {
return go_fmt_Errorf(std::string("boot: reading %s: %w",20),name,err);
}
constexpr int64_t load=65536ULL;
if ((cast<int64_t>((cast<int64_t>(65536ULL) + len(bin))) > len(m->RAM))) {
return go_fmt_Errorf(std::string("boot: %s (%d bytes) does not fit at 8C010000",44),name,len(bin));
}
{int64_t off = cast<int64_t>(0ULL);for (;(off < cast<int64_t>(65536ULL));off += cast<int64_t>(4ULL)){
uint32_t v = cast<uint32_t>((cast<uint32_t>(4027383808ULL) | cast<uint32_t>(off)));
auto tmp47 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
m->RAM[off] = std::get<0>(tmp47);
m->RAM[cast<int64_t>((off + cast<int64_t>(1ULL)))] = std::get<1>(tmp47);
m->RAM[cast<int64_t>((off + cast<int64_t>(2ULL)))] = std::get<2>(tmp47);
m->RAM[cast<int64_t>((off + cast<int64_t>(3ULL)))] = std::get<3>(tmp47);
}
}dc_Machine_putRAM32(m,cast<int64_t>(176ULL),cast<uint32_t>(2684354816ULL));
dc_Machine_putRAM32(m,cast<int64_t>(180ULL),cast<uint32_t>(2684354820ULL));
dc_Machine_putRAM32(m,cast<int64_t>(184ULL),cast<uint32_t>(2684354824ULL));
dc_Machine_putRAM32(m,cast<int64_t>(188ULL),cast<uint32_t>(2684354828ULL));
dc_Machine_putRAM32(m,cast<int64_t>(224ULL),cast<uint32_t>(2684354832ULL));
auto tmp48 = std::make_tuple(cast<uint8_t>(43ULL),cast<uint8_t>(0ULL));
m->RAM[cast<int64_t>(16ULL)] = std::get<0>(tmp48);
m->RAM[cast<int64_t>(17ULL)] = std::get<1>(tmp48);
auto tmp49 = std::make_tuple(cast<uint8_t>(9ULL),cast<uint8_t>(0ULL));
m->RAM[cast<int64_t>(18ULL)] = std::get<0>(tmp49);
m->RAM[cast<int64_t>(19ULL)] = std::get<1>(tmp49);
gcopy(sub(m->RAM,cast<int64_t>(104ULL),cast<int64_t>(112ULL)),Slice<uint8_t>{cast<uint8_t>(82ULL),cast<uint8_t>(69ULL),cast<uint8_t>(84ULL),cast<uint8_t>(82ULL),cast<uint8_t>(79ULL),cast<uint8_t>(82ULL),cast<uint8_t>(86ULL),cast<uint8_t>(0ULL)});
dc_Machine_putRAM32(m,cast<int64_t>(65520ULL),cast<uint32_t>(0ULL));
dc_Machine_putRAM32(m,cast<int64_t>(65524ULL),cast<uint32_t>(0ULL));
gcopy(sub(m->RAM,cast<int64_t>(65536ULL),len(m->RAM)),bin);
m->bios = dc_newBIOS(m);
sh4_CPU* c = m->CPU;
sh4_CPU_Reset(c);
sh4_CPU_SetSR(c,cast<uint32_t>(1073742064ULL));
sh4_CPU_SetFPSCR(c,cast<uint32_t>(262145ULL));
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>(2348872704ULL);
sh4_CPU_SetPC(c,cast<uint32_t>(2348875776ULL));
return {};
}
}
// tools/platform/dc/boot.go:91:1
void dc_Machine_putRAM32(dc_Machine* m,int64_t off,uint32_t v){
{
auto tmp50 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
m->RAM[off] = std::get<0>(tmp50);
m->RAM[cast<int64_t>((off + cast<int64_t>(1ULL)))] = std::get<1>(tmp50);
m->RAM[cast<int64_t>((off + cast<int64_t>(2ULL)))] = std::get<2>(tmp50);
m->RAM[cast<int64_t>((off + cast<int64_t>(3ULL)))] = std::get<3>(tmp50);
}
}
// tools/platform/dc/boot.go:95:1
uint32_t dc_Machine_ram32(dc_Machine* m,uint32_t addr){
{
uint32_t off = cast<uint32_t>((addr & cast<uint32_t>(16777215ULL)));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->RAM[off]) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}
// tools/platform/dc/cue.go:42:1
bool dc_Track_IsData(dc_Track t){
{
return go_strings_HasPrefix(t.Mode,std::string("MODE1",5));
}
}
// tools/platform/dc/cue.go:47:1
std::tuple<std::string,Slice<dc_Track>,Error> dc_parseCue(std::string text){
std::string bin{};
Slice<dc_Track> tracks{};
Error err{};
{
{auto&& tmp51 = go_strings_Split(text,std::string("\012",1));
for(int64_t tmp52=0;tmp52<len(tmp51);++tmp52){
auto ln=tmp52;auto line=tmp51[tmp52];line = go_strings_TrimSpace(line);
{
int64_t i = go_strings_Index(line,std::string("//",2));
if ((i >= cast<int64_t>(0ULL))) {
line = go_strings_TrimSpace(sub(line,0,i));
}
}
if ((line == std::string("",0))) {
continue;
}
Slice<std::string> f = go_strings_Fields(line);
{
auto tmp54=f[cast<int64_t>(0ULL)];
if (tmp54==(std::string("TRACK",5))){
if ((len(f) < cast<int64_t>(2ULL))) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue line %d: TRACK without a mode",33),cast<int64_t>((ln + cast<int64_t>(1ULL))))};
}
tracks = append(tracks,Slice<dc_Track>{dc_Track{cast<int64_t>((len(tracks) + cast<int64_t>(1ULL))),f[cast<int64_t>(1ULL)],{},{},cast<int64_t>(-1ULL)}});
}
else if (tmp54==(std::string("DATAFILE",8)) || tmp54==(std::string("FILE",4)) || tmp54==(std::string("AUDIOFILE",9))){
if ((len(tracks) == cast<int64_t>(0ULL))) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue line %d: %s before any TRACK",32),cast<int64_t>((ln + cast<int64_t>(1ULL))),f[cast<int64_t>(0ULL)])};
}
auto tmp55 = dc_quoted(line);
std::string name = std::get<0>(tmp55);
std::string rest = std::get<1>(tmp55);
bool ok = std::get<2>(tmp55);
if ((!ok)) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue line %d: no quoted filename",31),cast<int64_t>((ln + cast<int64_t>(1ULL))))};
}
if ((bin == std::string("",0))) {
bin = name;
}
else if ((bin != name)) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue line %d: second data file %q (only single-file rips are handled)",68),cast<int64_t>((ln + cast<int64_t>(1ULL))),name)};
}
dc_Track* t = (&tracks[cast<int64_t>((len(tracks) - cast<int64_t>(1ULL)))]);
{auto&& tmp56 = go_strings_Fields(rest);
for(int64_t tmp57=0;tmp57<len(tmp56);++tmp57){
auto tok=tmp56[tmp57];if (go_strings_HasPrefix(tok,std::string("#",1))) {
auto tmp58 = go_strconv_ParseInt(sub(tok,cast<int64_t>(1ULL),len(tok)),cast<int64_t>(10ULL),cast<int64_t>(64ULL));
int64_t off = std::get<0>(tmp58);
Error err = std::get<1>(tmp58);
if (bool(err)) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue line %d: bad offset %q",26),cast<int64_t>((ln + cast<int64_t>(1ULL))),tok)};
}
t->FileOffset = off;
}
}}
}
}
tmp53:;
}}
if (((bin == std::string("",0)) || (len(tracks) == cast<int64_t>(0ULL)))) {
return {std::string("",0),{},go_fmt_Errorf(std::string("cue: no tracks found",20))};
}
return {bin,tracks,{}};
}
}
// tools/platform/dc/cue.go:97:1
std::tuple<std::string,std::string,bool> dc_quoted(std::string line){
std::string s{};
std::string rest{};
bool ok{};
{
int64_t i = go_strings_IndexByte(line,cast<uint8_t>(34ULL));
if ((i < cast<int64_t>(0ULL))) {
return {std::string("",0),std::string("",0),false};
}
int64_t j = go_strings_IndexByte(sub(line,cast<int64_t>((i + cast<int64_t>(1ULL))),len(line)),cast<uint8_t>(34ULL));
if ((j < cast<int64_t>(0ULL))) {
return {std::string("",0),std::string("",0),false};
}
return {sub(line,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((cast<int64_t>((i + cast<int64_t>(1ULL))) + j))),sub(line,cast<int64_t>((cast<int64_t>((i + cast<int64_t>(2ULL))) + j)),len(line)),true};
}
}
// tools/platform/dc/holly.go:34:1
uint32_t dc_Machine_hollyRead(dc_Machine* m,uint32_t addr){
{
dc_Holly* h = (&m->Holly);
{
switch(addr){
case cast<uint32_t>(6252800ULL):{
uint32_t v = h->ISTNRM;
if ((h->ISTEXT != cast<uint32_t>(0ULL))) {
v |= cast<uint32_t>(1073741824ULL);
}
if ((h->ISTERR != cast<uint32_t>(0ULL))) {
v |= cast<uint32_t>(2147483648ULL);
}
return v;
break;}
case cast<uint32_t>(6252804ULL):{
return h->ISTEXT;
break;}
case cast<uint32_t>(6252808ULL):{
return h->ISTERR;
break;}
case cast<uint32_t>(6252816ULL):{
return h->IML2NRM;
break;}
case cast<uint32_t>(6252820ULL):{
return h->IML2EXT;
break;}
case cast<uint32_t>(6252824ULL):{
return h->IML2ERR;
break;}
case cast<uint32_t>(6252832ULL):{
return h->IML4NRM;
break;}
case cast<uint32_t>(6252836ULL):{
return h->IML4EXT;
break;}
case cast<uint32_t>(6252840ULL):{
return h->IML4ERR;
break;}
case cast<uint32_t>(6252848ULL):{
return h->IML6NRM;
break;}
case cast<uint32_t>(6252852ULL):{
return h->IML6EXT;
break;}
case cast<uint32_t>(6252856ULL):{
return h->IML6ERR;
break;}
case cast<uint32_t>(6252684ULL):{
return cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(6252700ULL):{
return cast<uint32_t>(11ULL);
break;}
case cast<uint32_t>(6252544ULL):{
return m->C2DStat;
break;}
case cast<uint32_t>(6252548ULL):{
return m->C2DLen;
break;}
case cast<uint32_t>(6252552ULL):{
return cast<uint32_t>(0ULL);
break;}
}}
dc_Machine_logf(m,std::string("SB read %08X (PC %08X)",22),addr,sh4_CPU_CurPC(m->CPU));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/dc/holly.go:85:1
void dc_Machine_hollyWrite(dc_Machine* m,uint32_t addr,uint32_t v){
{
dc_Holly* h = (&m->Holly);
{
switch(addr){
case cast<uint32_t>(6252800ULL):{
h->ISTNRM &= ~(v);
break;}
case cast<uint32_t>(6252804ULL):{
break;}
case cast<uint32_t>(6252808ULL):{
h->ISTERR &= ~(v);
break;}
case cast<uint32_t>(6252816ULL):{
h->IML2NRM = v;
break;}
case cast<uint32_t>(6252820ULL):{
h->IML2EXT = v;
break;}
case cast<uint32_t>(6252824ULL):{
h->IML2ERR = v;
break;}
case cast<uint32_t>(6252832ULL):{
h->IML4NRM = v;
break;}
case cast<uint32_t>(6252836ULL):{
h->IML4EXT = v;
break;}
case cast<uint32_t>(6252840ULL):{
h->IML4ERR = v;
break;}
case cast<uint32_t>(6252848ULL):{
h->IML6NRM = v;
break;}
case cast<uint32_t>(6252852ULL):{
h->IML6EXT = v;
break;}
case cast<uint32_t>(6252856ULL):{
h->IML6ERR = v;
break;}
case cast<uint32_t>(6252544ULL):{
m->C2DStat = v;
break;}
case cast<uint32_t>(6252548ULL):{
m->C2DLen = v;
break;}
case cast<uint32_t>(6252552ULL):{
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
uint32_t src = sh4_CPU_OnchipReg(m->CPU,cast<uint32_t>(4288675872ULL));
if ((cast<uint32_t>((m->C2DStat & cast<uint32_t>(16777216ULL))) == cast<uint32_t>(0ULL))) {
dc_Machine_taSubmitDMA(m,src,m->C2DLen);
}
else {
uint32_t dst = cast<uint32_t>((m->C2DStat & cast<uint32_t>(16777215ULL)));
if (bool(m->OnC2DTexture)) {
m->OnC2DTexture(src,dst,m->C2DLen);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;((cast<uint32_t>((i + cast<uint32_t>(4ULL))) <= m->C2DLen) && (cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(4ULL))) <= cast<uint32_t>(8388608ULL)));i += cast<uint32_t>(4ULL)){
uint32_t v = dc_Machine_ram32(m,cast<uint32_t>((src + i)));
m->VRAM[cast<uint32_t>((dst + i))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((dst + i)));
m->VRAM[cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(1ULL))));
m->VRAM[cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(2ULL))));
m->VRAM[cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(3ULL))));
}
}}
m->TAWrites += cast<uint64_t>(divi<uint32_t>(m->C2DLen,cast<uint32_t>(4ULL)));
m->C2DCountdown = cast<uint32_t>((cast<uint32_t>(2000ULL) + divi<uint32_t>(m->C2DLen,cast<uint32_t>(16ULL))));
m->C2DLen = cast<uint32_t>(0ULL);
}
break;}
default:{
dc_Machine_logf(m,std::string("SB write %08X = %08X (PC %08X)",30),addr,v,sh4_CPU_CurPC(m->CPU));
return ;
break;}
}}
dc_Machine_updateIRL(m);
}
}
// tools/platform/dc/holly.go:157:1
void dc_Machine_taSubmitDMA(dc_Machine* m,uint32_t src,uint32_t byteLen){
{
std::array<uint8_t,32> chunk={};
{uint32_t off = cast<uint32_t>(0ULL);for (;(cast<uint32_t>((off + cast<uint32_t>(32ULL))) <= byteLen);off += cast<uint32_t>(32ULL)){
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(32ULL));i += cast<uint32_t>(4ULL)){
le_PutUint32(rrBorrow(chunk,i,len(chunk)),dc_Machine_ram32(m,cast<uint32_t>((cast<uint32_t>((src + off)) + i))));
}
}m->TAFrame = append(m->TAFrame,sub(chunk,0,len(chunk)));
dc_Machine_taFeed(m,sub(chunk,0,len(chunk)));
}
}}
}
// tools/platform/dc/holly.go:174:1
void dc_Machine_taFifoWrite(dc_Machine* m,uint32_t v){
{
std::array<uint8_t,4> b={};
le_PutUint32(rrBorrow(b,0,len(b)),v);
m->TAFifo = append(m->TAFifo,sub(b,0,len(b)));
if ((len(m->TAFifo) < cast<int64_t>(32ULL))) {
return ;
}
m->TAFrame = append(m->TAFrame,m->TAFifo);
dc_Machine_taFeed(m,m->TAFifo);
m->TAFifo = sub(m->TAFifo,0,cast<int64_t>(0ULL));
if ((((m->C2DPendingBits != cast<uint32_t>(0ULL)) && (m->C2DCountdown == cast<uint32_t>(0ULL))) && (m->TAFifoCountdown == cast<uint32_t>(0ULL)))) {
m->TAFifoCountdown = cast<uint32_t>(2000ULL);
}
}
}
// tools/platform/dc/holly.go:197:1
void dc_Machine_taFeed(dc_Machine* m,Slice<uint8_t> p){
{
if ((m->TANeed > cast<uint32_t>(0ULL))) {
m->TANeed -= cast<uint32_t>(32ULL);
return ;
}
uint32_t pcw = le_Uint32(p);
{
switch(shr<uint32_t>(pcw,cast<int64_t>(29ULL))){
case cast<uint32_t>(0ULL):{
if (((m->TAList >= cast<int32_t>(0ULL)) && (m->TAList < cast<int32_t>(5ULL)))) {
m->C2DPendingBits |= dc_taListDoneBit[m->TAList];
}
m->TAList = cast<int32_t>(-1ULL);
break;}
case cast<uint32_t>(4ULL):{
dc_Machine_taOpenList(m,cast<int32_t>(cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL)))));
bool textured = (cast<uint32_t>((pcw & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
uint32_t colType = cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(4ULL)) & cast<uint32_t>(3ULL)));
bool volume = (cast<uint32_t>((pcw & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
m->TAVtx = cast<uint32_t>(32ULL);
if ((((volume || ((textured && (colType == cast<uint32_t>(1ULL))))) || (m->TAList == cast<int32_t>(1ULL))) || (m->TAList == cast<int32_t>(3ULL)))) {
m->TAVtx = cast<uint32_t>(64ULL);
}
if ((volume && textured)) {
m->TANeed = cast<uint32_t>(32ULL);
}
else if ((((colType == cast<uint32_t>(2ULL)) && (cast<uint32_t>((pcw & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) && textured)) {
m->TANeed = cast<uint32_t>(32ULL);
}
break;}
case cast<uint32_t>(5ULL):{
dc_Machine_taOpenList(m,cast<int32_t>(cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL)))));
m->TAVtx = cast<uint32_t>(64ULL);
break;}
case cast<uint32_t>(7ULL):{
if ((m->TAVtx == cast<uint32_t>(64ULL))) {
m->TANeed = cast<uint32_t>(32ULL);
}
break;}
}}
}
}
// tools/platform/dc/holly.go:242:1
void dc_Machine_taOpenList(dc_Machine* m,int32_t n){
{
if (((m->TAList >= cast<int32_t>(0ULL)) && (m->TAList != n))) {
if ((m->TAList < cast<int32_t>(5ULL))) {
m->C2DPendingBits |= dc_taListDoneBit[m->TAList];
}
}
m->TAList = n;
}
}
// tools/platform/dc/holly.go:252:1
void dc_Machine_tickCompletions(dc_Machine* m){
{
if ((m->RenderCountdown > cast<uint32_t>(0ULL))) {
{
m->RenderCountdown--;
if ((m->RenderCountdown == cast<uint32_t>(0ULL))) {
dc_Machine_renderFrame(m);
dc_Machine_raiseNRM(m,cast<uint32_t>(7ULL));
if (bool(m->OnRender)) {
m->OnRender();
}
}
}
}
if ((m->C2DCountdown > cast<uint32_t>(0ULL))) {
{
m->C2DCountdown--;
if ((m->C2DCountdown == cast<uint32_t>(0ULL))) {
dc_Machine_raiseNRM(m,cast<uint32_t>((cast<uint32_t>(524288ULL) | m->C2DPendingBits)));
m->C2DPendingBits = cast<uint32_t>(0ULL);
}
}
}
if ((m->TAFifoCountdown > cast<uint32_t>(0ULL))) {
{
m->TAFifoCountdown--;
if ((m->TAFifoCountdown == cast<uint32_t>(0ULL))) {
dc_Machine_raiseNRM(m,m->C2DPendingBits);
m->C2DPendingBits = cast<uint32_t>(0ULL);
}
}
}
}
}
// tools/platform/dc/holly.go:280:1
void dc_Machine_raiseNRM(dc_Machine* m,uint32_t bits){
{
m->Holly.ISTNRM |= bits;
dc_Machine_updateIRL(m);
}
}
// tools/platform/dc/holly.go:293:1
void dc_Machine_updateIRL(dc_Machine* m){
{
dc_Holly* h = (&m->Holly);
{
if ((((cast<uint32_t>((h->ISTNRM & h->IML6NRM)) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((h->ISTEXT & h->IML6EXT)) != cast<uint32_t>(0ULL))) || (cast<uint32_t>((h->ISTERR & h->IML6ERR)) != cast<uint32_t>(0ULL)))){
sh4_CPU_SetIRL(m->CPU,cast<uint32_t>(6ULL),cast<uint32_t>(800ULL));
}
else if ((((cast<uint32_t>((h->ISTNRM & h->IML4NRM)) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((h->ISTEXT & h->IML4EXT)) != cast<uint32_t>(0ULL))) || (cast<uint32_t>((h->ISTERR & h->IML4ERR)) != cast<uint32_t>(0ULL)))){
sh4_CPU_SetIRL(m->CPU,cast<uint32_t>(4ULL),cast<uint32_t>(864ULL));
}
else if ((((cast<uint32_t>((h->ISTNRM & h->IML2NRM)) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((h->ISTEXT & h->IML2EXT)) != cast<uint32_t>(0ULL))) || (cast<uint32_t>((h->ISTERR & h->IML2ERR)) != cast<uint32_t>(0ULL)))){
sh4_CPU_SetIRL(m->CPU,cast<uint32_t>(2ULL),cast<uint32_t>(928ULL));
}
else {
sh4_CPU_SetIRL(m->CPU,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
}
}
tmp59:;
}
}
// tools/platform/dc/holly.go:310:1
uint32_t dc_Machine_spgStatus(dc_Machine* m){
{
uint32_t vbl = m->PVRRegs[cast<int64_t>(51ULL)];
auto tmp60 = std::make_tuple(cast<uint32_t>((vbl & cast<uint32_t>(1023ULL))),cast<uint32_t>((shr<uint32_t>(vbl,cast<int64_t>(16ULL)) & cast<uint32_t>(1023ULL))));
uint32_t vbIn = std::get<0>(tmp60);
uint32_t vbOut = std::get<1>(tmp60);
uint32_t line = m->CurLine;
uint32_t v={};
bool inBlank = false;
if ((vbIn > vbOut)) {
inBlank = ((line >= vbIn) || (line < vbOut));
}
else if ((vbIn != vbOut)) {
inBlank = ((line >= vbIn) && (line < vbOut));
}
if (inBlank) {
v |= cast<uint32_t>(8192ULL);
}
v |= shl<uint32_t>(m->FieldNum,cast<int64_t>(10ULL));
return cast<uint32_t>((v | cast<uint32_t>((line & cast<uint32_t>(1023ULL)))));
}
}
// tools/platform/dc/ipbin.go:33:1
std::tuple<dc_IPBin,Error> dc_parseIPBin(Slice<uint8_t> sector){
{
if ((len(sector) < cast<int64_t>(256ULL))) {
return {dc_IPBin{},go_fmt_Errorf(std::string("ipbin: sector too short (%d bytes)",34),len(sector))};
}
auto field = [&](int64_t off,int64_t n)->std::string{
return go_strings_TrimRight(cast<std::string>(sub(sector,off,cast<int64_t>((off + n)))),std::string(" ",1));
}
;
dc_IPBin ip = dc_IPBin{field(cast<int64_t>(0ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(16ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(32ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(48ULL),cast<int64_t>(8ULL)),field(cast<int64_t>(56ULL),cast<int64_t>(8ULL)),field(cast<int64_t>(64ULL),cast<int64_t>(10ULL)),field(cast<int64_t>(74ULL),cast<int64_t>(6ULL)),field(cast<int64_t>(80ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(96ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(112ULL),cast<int64_t>(16ULL)),field(cast<int64_t>(128ULL),cast<int64_t>(128ULL))};
if ((ip.HardwareID != std::string("SEGA SEGAKATANA",15))) {
return {dc_IPBin{},go_fmt_Errorf(std::string("ipbin: hardware ID %q, want SEGA SEGAKATANA \342\200\224 not a Dreamcast boot sector",75),ip.HardwareID)};
}
return {ip,{}};
}
}
// tools/platform/dc/machine.go:170:1
dc_Machine* dc_NewMachine(dc_Disc* disc){
{
dc_Machine* m = arenaNew(dc_Machine{Slice<uint8_t>::make(cast<int64_t>(16777216ULL)),Slice<uint8_t>::make(cast<int64_t>(8388608ULL)),Slice<uint8_t>::make(cast<int64_t>(2097152ULL)),Slice<uint8_t>::make(cast<int64_t>(262144ULL)),{},disc,{},{},{},{},{},{},{},{},{},{},{},{},cast<int32_t>(-1ULL),cast<uint32_t>(32ULL),{},{},{},{},{},{},{},{},dc_PadState{cast<uint16_t>(65535ULL),{},{},cast<uint8_t>(128ULL),cast<uint8_t>(128ULL)},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,int64_t>{}});
{auto&& tmp61 = m->Flash;
for(int64_t tmp62=0;tmp62<len(tmp61);++tmp62){
auto i=tmp62;m->Flash[i] = cast<uint8_t>(255ULL);
}}
m->CPU = sh4_NewCPU(m);
m->ARM = arm_NewCPU(dc_armBus{m});
m->AICARegs = Map<uint32_t,uint32_t>{};
return m;
}
}
// tools/platform/dc/machine.go:206:1
Slice<std::string> dc_Machine_Census(dc_Machine* m){
{
Slice<std::string> out={};
{auto&& tmp63 = m->gaps;
for(auto [tmp64,tmp65]:tmp63){
auto k=tmp64;auto n=tmp65;out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("%s x%d",6),k,n)});
}}
out = append(out,sh4_CPU_Gaps(m->CPU));
if (bool(m->bios)) {
out = append(out,dc_biosHLE_census(m->bios));
}
return out;
}
}
// tools/platform/dc/machine.go:222:1
std::tuple<Slice<uint8_t>,uint32_t> dc_Machine_backing(dc_Machine* m,uint32_t addr){
{
{
if ((shr<uint32_t>(addr,cast<int64_t>(26ULL)) == cast<uint32_t>(3ULL))){
return {m->RAM,cast<uint32_t>((addr & cast<uint32_t>(16777215ULL)))};
}
else if (((addr >= cast<uint32_t>(83886080ULL)) && (addr < cast<uint32_t>(92274688ULL)))){
return {m->VRAM,dc_vram32to64(cast<uint32_t>((addr - cast<uint32_t>(83886080ULL))))};
}
else if (((addr >= cast<uint32_t>(67108864ULL)) && (addr < cast<uint32_t>(75497472ULL)))){
return {m->VRAM,cast<uint32_t>((addr - cast<uint32_t>(67108864ULL)))};
}
else if (((addr >= cast<uint32_t>(8388608ULL)) && (addr < cast<uint32_t>(10485760ULL)))){
return {m->AICARAM,cast<uint32_t>((addr - cast<uint32_t>(8388608ULL)))};
}
}
tmp66:;
return {{},cast<uint32_t>(0ULL)};
}
}
// tools/platform/dc/machine.go:242:1
uint32_t dc_vram32to64(uint32_t off){
{
uint32_t bank = cast<uint32_t>((shr<uint32_t>(off,cast<int64_t>(22ULL)) & cast<uint32_t>(1ULL)));
uint32_t i = cast<uint32_t>((off & cast<uint32_t>(4194303ULL)));
return cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(shr<uint32_t>(i,cast<int64_t>(2ULL)),cast<int64_t>(3ULL)) | shl<uint32_t>(bank,cast<int64_t>(2ULL)))) | cast<uint32_t>((i & cast<uint32_t>(3ULL)))));
}
}
// tools/platform/dc/machine.go:248:1
uint16_t dc_Machine_Fetch16_reference(dc_Machine* m,uint32_t addr){
{
{
auto tmp67 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp67);
uint32_t off = std::get<1>(tmp67);
if (bool(b)) {
return cast<uint16_t>((cast<uint16_t>(b[off]) | shl<uint16_t>(cast<uint16_t>(b[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
return dc_Machine_Read16(m,addr);
}
}
// tools/platform/dc/machine.go:262:1
bool dc_inRanges(Slice<dc_WatchRange> rs,uint32_t addr,int64_t size){
{
{auto&& tmp68 = rs;
for(int64_t tmp69=0;tmp69<len(tmp68);++tmp69){
auto r=tmp68[tmp69];if (((cast<uint32_t>((addr + cast<uint32_t>(size))) > r.Start) && (addr < cast<uint32_t>((r.Start + r.Len))))) {
return true;
}
}}
return false;
}
}
// tools/platform/dc/machine.go:271:1
void dc_Machine_watch(dc_Machine* m,bool write,uint32_t addr,uint32_t v,int64_t size){
{
Slice<dc_WatchRange> rs = m->WatchR;
if (write) {
rs = m->WatchW;
}
if (dc_inRanges(rs,addr,size)) {
m->OnWatch(write,addr,v,size,sh4_CPU_CurPC(m->CPU));
}
}
}
// tools/platform/dc/machine.go:281:1
uint8_t dc_Machine_Read8(dc_Machine* m,uint32_t addr){
{
uint8_t v = dc_Machine_read8i(m,addr);
if (bool(m->OnWatch)) {
dc_Machine_watch(m,false,addr,cast<uint32_t>(v),cast<int64_t>(1ULL));
}
return v;
}
}
// tools/platform/dc/machine.go:289:1
uint8_t dc_Machine_read8i(dc_Machine* m,uint32_t addr){
{
{
auto tmp70 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp70);
uint32_t off = std::get<1>(tmp70);
if (bool(b)) {
return b[off];
}
}
if (((addr >= cast<uint32_t>(2097152ULL)) && (addr < cast<uint32_t>(2359296ULL)))) {
return m->Flash[cast<uint32_t>((addr - cast<uint32_t>(2097152ULL)))];
}
return cast<uint8_t>(dc_Machine_ioRead(m,addr,cast<int64_t>(1ULL)));
}
}
// tools/platform/dc/machine.go:299:1
uint16_t dc_Machine_Read16(dc_Machine* m,uint32_t addr){
{
uint16_t v = dc_Machine_read16i(m,addr);
if (bool(m->OnWatch)) {
dc_Machine_watch(m,false,addr,cast<uint32_t>(v),cast<int64_t>(2ULL));
}
return v;
}
}
// tools/platform/dc/machine.go:307:1
uint16_t dc_Machine_read16i_reference(dc_Machine* m,uint32_t addr){
{
{
auto tmp71 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp71);
uint32_t off = std::get<1>(tmp71);
if (bool(b)) {
return cast<uint16_t>((cast<uint16_t>(b[off]) | shl<uint16_t>(cast<uint16_t>(b[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
if (((addr >= cast<uint32_t>(2097152ULL)) && (addr < cast<uint32_t>(2359296ULL)))) {
uint32_t off = cast<uint32_t>((addr - cast<uint32_t>(2097152ULL)));
return cast<uint16_t>((cast<uint16_t>(m->Flash[off]) | shl<uint16_t>(cast<uint16_t>(m->Flash[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
return cast<uint16_t>(dc_Machine_ioRead(m,addr,cast<int64_t>(2ULL)));
}
}
// tools/platform/dc/machine.go:318:1
uint32_t dc_Machine_Read32(dc_Machine* m,uint32_t addr){
{
uint32_t v = dc_Machine_read32i(m,addr);
if (bool(m->OnWatch)) {
dc_Machine_watch(m,false,addr,v,cast<int64_t>(4ULL));
}
return v;
}
}
// tools/platform/dc/machine.go:326:1
uint32_t dc_Machine_read32i_reference(dc_Machine* m,uint32_t addr){
{
{
auto tmp72 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp72);
uint32_t off = std::get<1>(tmp72);
if (bool(b)) {
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(b[off]) | shl<uint32_t>(cast<uint32_t>(b[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}
if (((addr >= cast<uint32_t>(2097152ULL)) && (addr < cast<uint32_t>(2359296ULL)))) {
uint32_t off = cast<uint32_t>((addr - cast<uint32_t>(2097152ULL)));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->Flash[off]) | shl<uint32_t>(cast<uint32_t>(m->Flash[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(m->Flash[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->Flash[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
return dc_Machine_ioRead(m,addr,cast<int64_t>(4ULL));
}
}
// tools/platform/dc/machine.go:337:1
void dc_Machine_Write8(dc_Machine* m,uint32_t addr,uint8_t v){
{
if (bool(m->OnWatch)) {
dc_Machine_watch(m,true,addr,cast<uint32_t>(v),cast<int64_t>(1ULL));
}
{
auto tmp73 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp73);
uint32_t off = std::get<1>(tmp73);
if (bool(b)) {
b[off] = v;
if(rrcapture::trace.active)rrDCWrite(m,b,off);
return ;
}
}
dc_Machine_ioWrite(m,addr,cast<int64_t>(1ULL),cast<uint32_t>(v));
}
}
// tools/platform/dc/machine.go:348:1
void dc_Machine_Write16(dc_Machine* m,uint32_t addr,uint16_t v){
{
if (bool(m->OnWatch)) {
dc_Machine_watch(m,true,addr,cast<uint32_t>(v),cast<int64_t>(2ULL));
}
{
auto tmp74 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp74);
uint32_t off = std::get<1>(tmp74);
if (bool(b)) {
auto tmp75 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
b[off] = std::get<0>(tmp75);
b[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp75);
if(rrcapture::trace.active)rrDCWrite(m,b,off);
if(rrcapture::trace.active)rrDCWrite(m,b,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
return ;
}
}
dc_Machine_ioWrite(m,addr,cast<int64_t>(2ULL),cast<uint32_t>(v));
}
}
// tools/platform/dc/machine.go:359:1
void dc_Machine_Write32(dc_Machine* m,uint32_t addr,uint32_t v){
{
if (bool(m->OnWatch)) {
dc_Machine_watch(m,true,addr,v,cast<int64_t>(4ULL));
}
{
auto tmp76 = dc_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp76);
uint32_t off = std::get<1>(tmp76);
if (bool(b)) {
auto tmp77 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
b[off] = std::get<0>(tmp77);
b[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp77);
b[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = std::get<2>(tmp77);
b[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = std::get<3>(tmp77);
if(rrcapture::trace.active)rrDCWrite(m,b,off);
if(rrcapture::trace.active)rrDCWrite(m,b,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
if(rrcapture::trace.active)rrDCWrite(m,b,cast<uint32_t>((off + cast<uint32_t>(2ULL))));
if(rrcapture::trace.active)rrDCWrite(m,b,cast<uint32_t>((off + cast<uint32_t>(3ULL))));
return ;
}
}
dc_Machine_ioWrite(m,addr,cast<int64_t>(4ULL),v);
}
}
// tools/platform/dc/machine.go:371:1
uint32_t dc_Machine_ioRead(dc_Machine* m,uint32_t addr,int64_t size){
{
{
if (((addr >= cast<uint32_t>(6253568ULL)) && (addr < cast<uint32_t>(6253824ULL)))){
return dc_Machine_mapleRead(m,addr);
}
else if (((addr >= cast<uint32_t>(6252544ULL)) && (addr < cast<uint32_t>(6254592ULL)))){
return dc_Machine_hollyRead(m,addr);
}
else if (((addr >= cast<uint32_t>(6254592ULL)) && (addr < cast<uint32_t>(6254848ULL)))){
dc_Machine_logf(m,std::string("GD-ROM ATA read %08X (unmodelled; the syscall HLE is the supported path)",72),addr);
return cast<uint32_t>(0ULL);
}
else if (((addr >= cast<uint32_t>(6258688ULL)) && (addr < cast<uint32_t>(6266880ULL)))){
return dc_Machine_pvrRead(m,addr);
}
else if (((addr >= cast<uint32_t>(7340032ULL)) && (addr < cast<uint32_t>(7405568ULL)))){
return dc_Machine_aicaRead(m,cast<uint32_t>((addr - cast<uint32_t>(7340032ULL))),size,true);
}
else if (((addr >= cast<uint32_t>(7405568ULL)) && (addr < cast<uint32_t>(7405584ULL)))){
return dc_Machine_rtcRead(m,addr);
}
else if ((addr < cast<uint32_t>(2097152ULL))){
dc_Machine_logf(m,std::string("boot ROM read %08X (no BIOS image; syscalls are HLE'd)",54),addr);
return cast<uint32_t>(0ULL);
}
else if (((addr >= cast<uint32_t>(16777216ULL)) && (addr < cast<uint32_t>(67108864ULL)))){
dc_Machine_logf(m,std::string("G2 expansion read %08X (empty socket, reads FF)",47),addr);
{
switch(size){
case cast<int64_t>(1ULL):{
return cast<uint32_t>(255ULL);
break;}
case cast<int64_t>(2ULL):{
return cast<uint32_t>(65535ULL);
break;}
}}
return cast<uint32_t>(4294967295ULL);
}
else if (((addr >= cast<uint32_t>(268435456ULL)) && (addr < cast<uint32_t>(335544320ULL)))){
dc_Machine_logf(m,std::string("TA FIFO read %08X",17),addr);
return cast<uint32_t>(0ULL);
}
}
tmp78:;
dc_Machine_logf(m,std::string("read%d unmodelled %08X (PC %08X)",32),cast<int64_t>((size * cast<int64_t>(8ULL))),addr,sh4_CPU_CurPC(m->CPU));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/dc/machine.go:410:1
void dc_Machine_ioWrite(dc_Machine* m,uint32_t addr,int64_t size,uint32_t v){
{
{
if (((addr >= cast<uint32_t>(6253568ULL)) && (addr < cast<uint32_t>(6253824ULL)))){
dc_Machine_mapleWrite(m,addr,v);
return ;
}
else if (((addr >= cast<uint32_t>(6252544ULL)) && (addr < cast<uint32_t>(6254592ULL)))){
dc_Machine_hollyWrite(m,addr,v);
return ;
}
else if (((addr >= cast<uint32_t>(6254592ULL)) && (addr < cast<uint32_t>(6254848ULL)))){
dc_Machine_logf(m,std::string("GD-ROM ATA write %08X (unmodelled; the syscall HLE is the supported path)",73),addr);
return ;
}
else if (((addr >= cast<uint32_t>(6258688ULL)) && (addr < cast<uint32_t>(6266880ULL)))){
dc_Machine_pvrWrite(m,addr,v);
return ;
}
else if (((addr >= cast<uint32_t>(7340032ULL)) && (addr < cast<uint32_t>(7405568ULL)))){
dc_Machine_aicaWrite(m,cast<uint32_t>((addr - cast<uint32_t>(7340032ULL))),size,v,true);
return ;
}
else if (((addr >= cast<uint32_t>(7405568ULL)) && (addr < cast<uint32_t>(7405584ULL)))){
dc_Machine_logf(m,std::string("RTC write %08X = %08X",21),addr,v);
return ;
}
else if (((addr >= cast<uint32_t>(268435456ULL)) && (addr < cast<uint32_t>(335544320ULL)))){
m->TAWrites++;
{
if ((size != cast<int64_t>(4ULL))){
dc_Machine_logf(m,std::string("TA FIFO write size %d unimplemented",35),cast<int64_t>((size * cast<int64_t>(8ULL))));
}
else if ((addr < cast<uint32_t>(276824064ULL))){
dc_Machine_taFifoWrite(m,v);
}
else if ((addr < cast<uint32_t>(285212672ULL))){
dc_Machine_logf(m,std::string("TA YUV-converter FIFO write unimplemented",41));
}
else {
dc_Machine_logf(m,std::string("TA texture-path FIFO write unimplemented",40));
}
}
tmp80:;
return ;
}
else if (((addr >= cast<uint32_t>(2097152ULL)) && (addr < cast<uint32_t>(2359296ULL)))){
dc_Machine_logf(m,std::string("flash write %08X (read-only stub)",33),addr);
return ;
}
}
tmp79:;
dc_Machine_logf(m,std::string("write%d unmodelled %08X = %08X (PC %08X)",40),cast<int64_t>((size * cast<int64_t>(8ULL))),addr,v,sh4_CPU_CurPC(m->CPU));
}
}
// tools/platform/dc/machine.go:453:1
uint32_t dc_Machine_pvrRead(dc_Machine* m,uint32_t addr){
{
{
switch(addr){
case cast<uint32_t>(6258688ULL):{
return cast<uint32_t>(402461147ULL);
break;}
case cast<uint32_t>(6258692ULL):{
return cast<uint32_t>(17ULL);
break;}
case cast<uint32_t>(6258956ULL):{
return dc_Machine_spgStatus(m);
break;}
}}
return m->PVRRegs[divi<uint32_t>((cast<uint32_t>((addr - cast<uint32_t>(6258688ULL)))),cast<uint32_t>(4ULL))];
}
}
// tools/platform/dc/machine.go:465:1
void dc_Machine_pvrWrite(dc_Machine* m,uint32_t addr,uint32_t v){
{
m->PVRRegs[divi<uint32_t>((cast<uint32_t>((addr - cast<uint32_t>(6258688ULL)))),cast<uint32_t>(4ULL))] = v;
if (((addr == cast<uint32_t>(6259012ULL)) && (cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
auto tmp81 = std::make_tuple(cast<int32_t>(0ULL),cast<uint32_t>(32ULL),cast<uint32_t>(0ULL));
m->TAList = std::get<0>(tmp81);
m->TAVtx = std::get<1>(tmp81);
m->TANeed = std::get<2>(tmp81);
m->TAFifo = sub(m->TAFifo,0,cast<int64_t>(0ULL));
m->TAClosed = append(sub(m->TAClosed,0,cast<int64_t>(0ULL)),m->TAFrame);
m->TAFrame = sub(m->TAFrame,0,cast<int64_t>(0ULL));
}
if (((addr == cast<uint32_t>(6258696ULL)) && (cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
auto tmp82 = std::make_tuple(cast<int32_t>(0ULL),cast<uint32_t>(32ULL),cast<uint32_t>(0ULL));
m->TAList = std::get<0>(tmp82);
m->TAVtx = std::get<1>(tmp82);
m->TANeed = std::get<2>(tmp82);
m->TAFifo = sub(m->TAFifo,0,cast<int64_t>(0ULL));
}
if ((addr == cast<uint32_t>(6258708ULL))) {
m->RenderCountdown = cast<uint32_t>(416666ULL);
}
}
}
// tools/platform/dc/maple.go:63:1
std::tuple<uint16_t,bool> dc_PadButton(std::string name){
{
auto tmp83 = lookup(dc_padButtonNames,name);
uint16_t b = std::get<0>(tmp83);
bool ok = std::get<1>(tmp83);
return {b,ok};
}
}
// tools/platform/dc/maple.go:69:1
uint32_t dc_bswap(uint32_t v){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(24ULL)) | shr<uint32_t>(v,cast<int64_t>(24ULL)))) | cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(16711680ULL))))) | cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(65280ULL)))));
}
}
// tools/platform/dc/maple.go:79:1
uint32_t dc_Machine_mapleRead(dc_Machine* m,uint32_t addr){
{
{
switch(addr){
case cast<uint32_t>(6253572ULL):{
return m->Maple.MDSTAR;
break;}
case cast<uint32_t>(6253584ULL):{
return m->Maple.MDTSEL;
break;}
case cast<uint32_t>(6253588ULL):{
return m->Maple.MDEN;
break;}
case cast<uint32_t>(6253592ULL):{
return cast<uint32_t>(0ULL);
break;}
}}
dc_Machine_logf(m,std::string("maple read %08X (PC %08X)",25),addr,sh4_CPU_CurPC(m->CPU));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/dc/maple.go:94:1
void dc_Machine_mapleWrite(dc_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(addr){
case cast<uint32_t>(6253572ULL):{
m->Maple.MDSTAR = cast<uint32_t>((v & cast<uint32_t>(536870880ULL)));
break;}
case cast<uint32_t>(6253584ULL):{
m->Maple.MDTSEL = v;
break;}
case cast<uint32_t>(6253588ULL):{
m->Maple.MDEN = cast<uint32_t>((v & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(6253592ULL):{
if (((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (m->Maple.MDEN != cast<uint32_t>(0ULL)))) {
dc_Machine_mapleDMA(m);
}
break;}
default:{
dc_Machine_logf(m,std::string("maple write %08X = %08X (PC %08X)",33),addr,v,sh4_CPU_CurPC(m->CPU));
break;}
}}
}
}
// tools/platform/dc/maple.go:112:1
void dc_Machine_mapleDMA(dc_Machine* m){
{
uint32_t addr = m->Maple.MDSTAR;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(64ULL));i++){
uint32_t ctrl = dc_Machine_ram32(m,addr);
uint32_t recv = cast<uint32_t>((dc_Machine_ram32(m,cast<uint32_t>((addr + cast<uint32_t>(4ULL)))) & cast<uint32_t>(536870911ULL)));
uint32_t frame = dc_Machine_ram32(m,cast<uint32_t>((addr + cast<uint32_t>(8ULL))));
uint32_t words = cast<uint32_t>((shr<uint32_t>(frame,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)));
uint32_t payload = cast<uint32_t>((addr + cast<uint32_t>(12ULL)));
addr = cast<uint32_t>((payload + cast<uint32_t>((cast<uint32_t>(4ULL) * words))));
uint32_t cmd = cast<uint32_t>((frame & cast<uint32_t>(255ULL)));
uint32_t dst = cast<uint32_t>((shr<uint32_t>(frame,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)));
dc_Machine_mapleRespond(m,cmd,dst,payload,recv);
if (((cmd != cast<uint32_t>(1ULL)) && (cmd != cast<uint32_t>(4ULL)))) {
dc_Machine_logf(m,std::string("maple frame cmd %d dst %02X fn %08X arg %08X recv %08X",54),cmd,dst,dc_Machine_ram32(m,payload),dc_Machine_ram32(m,cast<uint32_t>((payload + cast<uint32_t>(4ULL)))),recv);
}
if ((cast<uint32_t>((ctrl & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
break;
}
}
}dc_Machine_raiseNRM(m,cast<uint32_t>(4096ULL));
}
}
// tools/platform/dc/maple.go:138:1
void dc_Machine_mapleRespond(dc_Machine* m,uint32_t cmd,uint32_t dst,uint32_t payload,uint32_t recv){
{
if ((dst != cast<uint32_t>(32ULL))) {
dc_Machine_Write32(m,recv,cast<uint32_t>(4294967295ULL));
return ;
}
uint32_t src = cast<uint32_t>(32ULL);
uint32_t host = cast<uint32_t>(0ULL);
{
switch(cmd){
case cast<uint32_t>(1ULL):{
std::array<uint8_t,112> blk={};
le_PutUint32(rrBorrow(blk,cast<int64_t>(0ULL),len(blk)),cast<uint32_t>(16777216ULL));
le_PutUint32(rrBorrow(blk,cast<int64_t>(4ULL),len(blk)),dc_bswap(cast<uint32_t>(984830ULL)));
blk[cast<int64_t>(16ULL)] = cast<uint8_t>(255ULL);
blk[cast<int64_t>(17ULL)] = cast<uint8_t>(0ULL);
gcopy(sub(blk,cast<int64_t>(18ULL),len(blk)),std::string("Dreamcast Controller          ",30));
gcopy(sub(blk,cast<int64_t>(48ULL),len(blk)),std::string("Produced By or Under License From SEGA ENTERPRISES,LTD.    ",59));
le_PutUint16(sub(blk,cast<int64_t>(108ULL),len(blk)),cast<uint16_t>(430ULL));
le_PutUint16(sub(blk,cast<int64_t>(110ULL),len(blk)),cast<uint16_t>(500ULL));
dc_Machine_Write32(m,recv,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(5ULL) | shl<uint32_t>(host,cast<int64_t>(8ULL)))) | shl<uint32_t>(src,cast<int64_t>(16ULL)))) | cast<uint32_t>(469762048ULL))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(112ULL));i += cast<uint32_t>(4ULL)){
dc_Machine_Write32(m,cast<uint32_t>((cast<uint32_t>((recv + cast<uint32_t>(4ULL))) + i)),le_Uint32(rrBorrow(blk,i,len(blk))));
}
}break;}
case cast<uint32_t>(9ULL):{
dc_Machine_Write32(m,recv,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(8ULL) | shl<uint32_t>(host,cast<int64_t>(8ULL)))) | shl<uint32_t>(src,cast<int64_t>(16ULL)))) | cast<uint32_t>(50331648ULL))));
dc_Machine_Write32(m,cast<uint32_t>((recv + cast<uint32_t>(4ULL))),dc_bswap(cast<uint32_t>(16777216ULL)));
dc_Machine_Write32(m,cast<uint32_t>((recv + cast<uint32_t>(8ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->Pad.Buttons) | shl<uint32_t>(cast<uint32_t>(m->Pad.RT),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->Pad.LT),cast<int64_t>(24ULL)))));
dc_Machine_Write32(m,cast<uint32_t>((recv + cast<uint32_t>(12ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->Pad.JoyX) | shl<uint32_t>(cast<uint32_t>(m->Pad.JoyY),cast<int64_t>(8ULL)))) | cast<uint32_t>(8388608ULL))) | cast<uint32_t>(2147483648ULL))));
break;}
default:{
dc_Machine_logf(m,std::string("maple command %d answered as unsupported",40),cmd);
dc_Machine_Write32(m,recv,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(254ULL) | shl<uint32_t>(host,cast<int64_t>(8ULL)))) | shl<uint32_t>(src,cast<int64_t>(16ULL)))));
break;}
}}
}
}
// tools/platform/dc/pvr.go:39:1
void dc_Machine_renderFrame(dc_Machine* m){
{rrprof::Scope timing(1,"PowerVR software rasterizer");
{rrconsole::EventScope restore(rrcapture::trace.current);
{
uint32_t fbW = m->PVRRegs[cast<int64_t>(18ULL)];
uint32_t base = cast<uint32_t>((m->PVRRegs[cast<int64_t>(24ULL)] & cast<uint32_t>(16777215ULL)));
if ((cast<uint32_t>((fbW & cast<uint32_t>(7ULL))) != cast<uint32_t>(1ULL))) {
dc_Machine_logf(m,std::string("render: FB_W_CTRL packmode %d unimplemented (only 565)",54),cast<uint32_t>((fbW & cast<uint32_t>(7ULL))));
return ;
}
uint32_t size = m->PVRRegs[cast<int64_t>(23ULL)];
int64_t w = cast<int64_t>(((cast<int64_t>((cast<int64_t>(cast<uint32_t>((size & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL)));
int64_t h = cast<int64_t>((cast<int64_t>(cast<uint32_t>((shr<uint32_t>(size,cast<int64_t>(10ULL)) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
if (bool(m->OnPVRClear)) {
m->OnPVRClear(w,h);
}
dc_Machine_clearFB(m,base,w,h);
int64_t cmds = cast<int64_t>(1ULL);
dc_renderState st = dc_renderState{m,base,w,h,Slice<float>::make(cast<int64_t>((w * h))),{},{},{},{},{},{},{},{},{},{},{},{},cast<uint32_t>(4294967295ULL),{},{},{},{},{},{},{},{},{},{},{},{},{}};
Slice<uint8_t> stream = m->TAClosed;
uint32_t vtx = cast<uint32_t>(32ULL);
{int64_t off = cast<int64_t>(0ULL);for (;(cast<int64_t>((off + cast<int64_t>(32ULL))) <= len(stream));){
if (((m->RenderStopAfter > cast<int64_t>(0ULL)) && (cmds >= m->RenderStopAfter))) {
break;
}
uint32_t pcw = le_Uint32(rrBorrow(stream,off,len(stream)));
uint32_t typ = shr<uint32_t>(pcw,cast<int64_t>(29ULL));
int64_t sz = cast<int64_t>(32ULL);
{
switch(typ){
case cast<uint32_t>(4ULL):{
bool textured = (cast<uint32_t>((pcw & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
if (((((cast<uint32_t>((pcw & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL)) && textured)) || ((((cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(4ULL)) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)) && (cast<uint32_t>((pcw & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) && textured)))) {
sz = cast<int64_t>(64ULL);
}
break;}
case cast<uint32_t>(7ULL):{
if (st.sprite) {
sz = cast<int64_t>(64ULL);
}
else {
sz = cast<int64_t>(vtx);
}
break;}
}}
if (bool(m->OnPVRCmd)) {
int64_t end = cast<int64_t>((off + sz));
if ((end > len(stream))) {
end = len(stream);
}
m->OnPVRCmd(sub(stream,off,end));
}
cmds++;
{
switch(typ){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):{
break;}
case cast<uint32_t>(4ULL):{
dc_renderState_loadHeader(&(st),pcw,sub(stream,off,len(stream)));
st.sprite = false;
st.strip = cast<int64_t>(0ULL);
bool textured = (cast<uint32_t>((pcw & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
uint32_t colType = cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(4ULL)) & cast<uint32_t>(3ULL)));
bool offsCol = (cast<uint32_t>((pcw & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL));
bool volume = (cast<uint32_t>((pcw & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
st.colType = colType;
st.uv16 = (textured && (cast<uint32_t>((pcw & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
vtx = cast<uint32_t>(32ULL);
if ((((volume || ((textured && (colType == cast<uint32_t>(1ULL))))) || (cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL))) == cast<uint32_t>(1ULL))) || (cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)))) {
vtx = cast<uint32_t>(64ULL);
}
{
if ((volume && textured)){
}
else if ((((colType == cast<uint32_t>(2ULL)) && offsCol) && textured)){
if ((cast<int64_t>((off + cast<int64_t>(64ULL))) <= len(stream))) {
st.faceCol = dc_packFloatCol(sub(stream,cast<int64_t>((off + cast<int64_t>(32ULL))),len(stream)));
st.faceOffs = dc_packFloatCol(sub(stream,cast<int64_t>((off + cast<int64_t>(48ULL))),len(stream)));
}
}
else if ((colType == cast<uint32_t>(2ULL))){
st.faceCol = dc_packFloatCol(sub(stream,cast<int64_t>((off + cast<int64_t>(16ULL))),len(stream)));
}
}
tmp84:;
break;}
case cast<uint32_t>(5ULL):{
dc_renderState_loadHeader(&(st),pcw,sub(stream,off,len(stream)));
st.sprite = true;
vtx = cast<uint32_t>(64ULL);
break;}
case cast<uint32_t>(7ULL):{
if (st.sprite) {
if ((cast<int64_t>((off + cast<int64_t>(64ULL))) <= len(stream))) {
dc_renderState_drawSprite(&(st),sub(stream,off,len(stream)));
}
}
else {
dc_renderState_pushStripVertex(&(st),sub(stream,off,len(stream)),pcw);
}
break;}
}}
off += sz;
}
}dc_Machine_logf(m,std::string("render: stream %dB, %d tris, %d px -> %06X %dx%d",48),len(stream),st.tris,st.px,base,w,h);
}
}
}
}
// tools/platform/dc/pvr.go:149:1
void dc_Machine_clearFB(dc_Machine* m,uint32_t base,int64_t w,int64_t h){
{
uint32_t n = cast<uint32_t>(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(2ULL))));
if ((cast<uint32_t>((base + n)) > cast<uint32_t>(8388608ULL))) {
n = cast<uint32_t>((cast<uint32_t>(8388608ULL) - base));
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(cast<uint32_t>((i + cast<uint32_t>(4ULL))) <= n);i += cast<uint32_t>(4ULL)){
uint32_t off = dc_vram32to64(cast<uint32_t>((base + i)));
m->VRAM[off] = cast<uint8_t>(0ULL);
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,off);
m->VRAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(0ULL);
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
m->VRAM[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = cast<uint8_t>(0ULL);
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((off + cast<uint32_t>(2ULL))));
m->VRAM[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = cast<uint8_t>(0ULL);
if(rrcapture::trace.active)rrDCWrite(m,m->VRAM,cast<uint32_t>((off + cast<uint32_t>(3ULL))));
if ((bool(m->OnPVRPixel) && (w > cast<int64_t>(0ULL)))) {
int64_t p = cast<int64_t>(divi<uint32_t>(i,cast<uint32_t>(2ULL)));
m->OnPVRPixel(modi<int64_t>(p,w),divi<int64_t>(p,w),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(255ULL),true);
m->OnPVRPixel(modi<int64_t>((cast<int64_t>((p + cast<int64_t>(1ULL)))),w),divi<int64_t>((cast<int64_t>((p + cast<int64_t>(1ULL)))),w),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(255ULL),true);
}
}
}}
}
// tools/platform/dc/pvr.go:209:1
void dc_renderState_skip(dc_renderState* st,std::string what,uint32_t v){
{
dc_Machine_logf(st->m,std::string("render: %s %d unimplemented",27),what,v);
}
}
// tools/platform/dc/pvr.go:214:1
void dc_renderState_loadHeader(dc_renderState* st,uint32_t pcw,Slice<uint8_t> p){
{
uint32_t isp = le_Uint32(rrBorrow(p,cast<int64_t>(4ULL),len(p)));
uint32_t tsp = le_Uint32(rrBorrow(p,cast<int64_t>(8ULL),len(p)));
uint32_t tcw = le_Uint32(rrBorrow(p,cast<int64_t>(12ULL),len(p)));
st->texture = (cast<uint32_t>((pcw & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
st->gouraud = (cast<uint32_t>((pcw & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
uint32_t list = cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL)));
st->blend = (list == cast<uint32_t>(2ULL));
st->offsEn = ((cast<uint32_t>((pcw & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && st->texture);
st->shade = cast<uint32_t>((shr<uint32_t>(tsp,cast<int64_t>(6ULL)) & cast<uint32_t>(3ULL)));
st->depthCmp = cast<uint32_t>((shr<uint32_t>(isp,cast<int64_t>(29ULL)) & cast<uint32_t>(7ULL)));
st->zWrite = (cast<uint32_t>((shr<uint32_t>(isp,cast<int64_t>(26ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL));
st->texUW = shl<int64_t>(cast<int64_t>(8ULL),(cast<uint32_t>((shr<uint32_t>(tsp,cast<int64_t>(3ULL)) & cast<uint32_t>(7ULL)))));
st->texVH = shl<int64_t>(cast<int64_t>(8ULL),(cast<uint32_t>((tsp & cast<uint32_t>(7ULL)))));
st->texAddr = shl<uint32_t>(cast<uint32_t>((tcw & cast<uint32_t>(2097151ULL))),cast<int64_t>(3ULL));
st->texMipIx = cast<uint32_t>(0ULL);
st->texFmt = cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(27ULL)) & cast<uint32_t>(7ULL)));
st->texTwid = (cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(26ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL));
st->texVQ = (cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((st->texFmt == cast<uint32_t>(6ULL))) {
st->texBank = cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(25ULL)) & cast<uint32_t>(3ULL)));
}
else {
st->texBank = cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(21ULL)) & cast<uint32_t>(63ULL)));
}
if ((cast<uint32_t>((shr<uint32_t>(tcw,cast<int64_t>(31ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
if (st->texVQ) {
st->texMipIx = dc_vqMipIndexOffset(st->texUW);
}
else {
st->texAddr += dc_mipTopOffset(st->texUW,st->texFmt);
}
st->texVH = st->texUW;
}
if ((len(p) >= cast<int64_t>(24ULL))) {
st->baseCol = le_Uint32(rrBorrow(p,cast<int64_t>(16ULL),len(p)));
st->baseOffs = le_Uint32(rrBorrow(p,cast<int64_t>(20ULL),len(p)));
}
}
}
// tools/platform/dc/pvr.go:259:1
uint32_t dc_mipTopOffset(int64_t side,uint32_t fmt){
{
uint32_t texels = cast<uint32_t>(cast<int64_t>((divi<int64_t>((cast<int64_t>((cast<int64_t>((side * side)) - cast<int64_t>(1ULL)))),cast<int64_t>(3ULL)) + cast<int64_t>(3ULL))));
{
switch(fmt){
case cast<uint32_t>(5ULL):{
return divi<uint32_t>(texels,cast<uint32_t>(2ULL));
break;}
case cast<uint32_t>(6ULL):{
return texels;
break;}
default:{
return cast<uint32_t>((texels * cast<uint32_t>(2ULL)));
break;}
}}
}
}
// tools/platform/dc/pvr.go:274:1
uint32_t dc_vqMipIndexOffset(int64_t side){
{
uint32_t off={};
{int64_t s = cast<int64_t>(1ULL);for (;(s < side);s *= cast<int64_t>(2ULL)){
int64_t n = divi<int64_t>(cast<int64_t>((s * s)),cast<int64_t>(4ULL));
if ((n < cast<int64_t>(1ULL))) {
n = cast<int64_t>(1ULL);
}
off += cast<uint32_t>(n);
}
}return off;
}
}
// tools/platform/dc/pvr.go:287:1
uint32_t dc_packFloatCol(Slice<uint8_t> p){
{
auto cl = [&](float f)->uint32_t{
if ((f <= cast<float>(0ULL))) {
return cast<uint32_t>(0ULL);
}
if ((f >= cast<float>(1ULL))) {
return cast<uint32_t>(255ULL);
}
return cast<uint32_t>((f * cast<float>(255ULL)));
}
;
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cl(dc_f32(p)),cast<int64_t>(24ULL)) | shl<uint32_t>(cl(dc_f32(sub(p,cast<int64_t>(4ULL),len(p)))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cl(dc_f32(sub(p,cast<int64_t>(8ULL),len(p)))),cast<int64_t>(8ULL)))) | cl(dc_f32(sub(p,cast<int64_t>(12ULL),len(p))))));
}
}
// tools/platform/dc/pvr.go:300:1
float dc_f32(Slice<uint8_t> p){
{
return dc_float32frombits(le_Uint32(p));
}
}
// tools/platform/dc/pvr.go:307:1
void dc_renderState_drawSprite(dc_renderState* st,Slice<uint8_t> p){
{
auto tmp85 = std::make_tuple(dc_f32(sub(p,cast<int64_t>(4ULL),len(p))),dc_f32(sub(p,cast<int64_t>(8ULL),len(p))),dc_f32(sub(p,cast<int64_t>(12ULL),len(p))));
float ax = std::get<0>(tmp85);
float ay = std::get<1>(tmp85);
float az = std::get<2>(tmp85);
auto tmp86 = std::make_tuple(dc_f32(sub(p,cast<int64_t>(16ULL),len(p))),dc_f32(sub(p,cast<int64_t>(20ULL),len(p))),dc_f32(sub(p,cast<int64_t>(24ULL),len(p))));
float bx = std::get<0>(tmp86);
float by = std::get<1>(tmp86);
float bz = std::get<2>(tmp86);
auto tmp87 = std::make_tuple(dc_f32(sub(p,cast<int64_t>(28ULL),len(p))),dc_f32(sub(p,cast<int64_t>(32ULL),len(p))),dc_f32(sub(p,cast<int64_t>(36ULL),len(p))));
float cx = std::get<0>(tmp87);
float cy = std::get<1>(tmp87);
float cz = std::get<2>(tmp87);
auto tmp88 = std::make_tuple(dc_f32(sub(p,cast<int64_t>(40ULL),len(p))),dc_f32(sub(p,cast<int64_t>(44ULL),len(p))));
float dx = std::get<0>(tmp88);
float dy = std::get<1>(tmp88);
auto tmp89 = dc_unpackUV16(le_Uint32(rrBorrow(p,cast<int64_t>(52ULL),len(p))));
float au = std::get<0>(tmp89);
float av = std::get<1>(tmp89);
auto tmp90 = dc_unpackUV16(le_Uint32(rrBorrow(p,cast<int64_t>(56ULL),len(p))));
float bu = std::get<0>(tmp90);
float bv = std::get<1>(tmp90);
auto tmp91 = dc_unpackUV16(le_Uint32(rrBorrow(p,cast<int64_t>(60ULL),len(p))));
float cu = std::get<0>(tmp91);
float cv = std::get<1>(tmp91);
auto tmp92 = std::make_tuple(((au + cu) - bu),((av + cv) - bv));
float du = std::get<0>(tmp92);
float dv = std::get<1>(tmp92);
float dz = ((az + cz) - bz);
uint32_t offs = cast<uint32_t>(0ULL);
if (st->offsEn) {
offs = st->baseOffs;
}
dc_pvrVert a = dc_pvrVert{ax,ay,az,au,av,st->baseCol,offs};
dc_pvrVert b = dc_pvrVert{bx,by,bz,bu,bv,st->baseCol,offs};
dc_pvrVert c = dc_pvrVert{cx,cy,cz,cu,cv,st->baseCol,offs};
dc_pvrVert d = dc_pvrVert{dx,dy,dz,du,dv,st->baseCol,offs};
dc_renderState_tri(st,a,b,c);
dc_renderState_tri(st,a,c,d);
}
}
// tools/platform/dc/pvr.go:331:1
void dc_renderState_pushStripVertex(dc_renderState* st,Slice<uint8_t> p,uint32_t pcw){
{
dc_pvrVert v = dc_pvrVert{cast<float>(dc_f32(sub(p,cast<int64_t>(4ULL),len(p)))),cast<float>(dc_f32(sub(p,cast<int64_t>(8ULL),len(p)))),cast<float>(dc_f32(sub(p,cast<int64_t>(12ULL),len(p)))),{},{},{},{}};
if (st->texture) {
if (st->uv16) {
auto tmp93 = dc_unpackUV16(le_Uint32(rrBorrow(p,cast<int64_t>(16ULL),len(p))));
v.u = std::get<0>(tmp93);
v.v = std::get<1>(tmp93);
}
else {
auto tmp94 = std::make_tuple(dc_f32(sub(p,cast<int64_t>(16ULL),len(p))),dc_f32(sub(p,cast<int64_t>(20ULL),len(p))));
v.u = std::get<0>(tmp94);
v.v = std::get<1>(tmp94);
}
}
{
switch(st->colType){
case cast<uint32_t>(1ULL):{
if (st->texture) {
v.color = dc_packFloatCol(sub(p,cast<int64_t>(32ULL),len(p)));
if (st->offsEn) {
v.offs = dc_packFloatCol(sub(p,cast<int64_t>(48ULL),len(p)));
}
}
else {
v.color = dc_packFloatCol(sub(p,cast<int64_t>(16ULL),len(p)));
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
v.color = dc_scaleCol(st->faceCol,dc_f32(sub(p,cast<int64_t>(24ULL),len(p))));
if (st->offsEn) {
v.offs = dc_scaleCol(st->faceOffs,dc_f32(sub(p,cast<int64_t>(28ULL),len(p))));
}
break;}
default:{
v.color = le_Uint32(rrBorrow(p,cast<int64_t>(24ULL),len(p)));
if (st->offsEn) {
v.offs = le_Uint32(rrBorrow(p,cast<int64_t>(28ULL),len(p)));
}
if ((!st->texture)) {
v.color = le_Uint32(rrBorrow(p,cast<int64_t>(16ULL),len(p)));
}
break;}
}}
if ((st->strip < cast<int64_t>(3ULL))) {
st->sv[st->strip] = v;
st->strip++;
}
else {
auto tmp95 = std::make_tuple(st->sv[cast<int64_t>(1ULL)],st->sv[cast<int64_t>(2ULL)]);
st->sv[cast<int64_t>(0ULL)] = std::get<0>(tmp95);
st->sv[cast<int64_t>(1ULL)] = std::get<1>(tmp95);
st->sv[cast<int64_t>(2ULL)] = v;
}
if ((st->strip == cast<int64_t>(3ULL))) {
dc_renderState_tri(st,st->sv[cast<int64_t>(0ULL)],st->sv[cast<int64_t>(1ULL)],st->sv[cast<int64_t>(2ULL)]);
}
if ((cast<uint32_t>((shr<uint32_t>(pcw,cast<int64_t>(28ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
st->strip = cast<int64_t>(0ULL);
}
}
}
// tools/platform/dc/pvr.go:384:1
uint32_t dc_scaleCol(uint32_t col,float i){
{
if ((i <= cast<float>(0ULL))) {
return cast<uint32_t>((col & cast<uint32_t>(4278190080ULL)));
}
if ((i >= cast<float>(1ULL))) {
return col;
}
uint32_t r = cast<uint32_t>((cast<float>(cast<uint32_t>((shr<uint32_t>(col,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))) * i));
uint32_t g = cast<uint32_t>((cast<float>(cast<uint32_t>((shr<uint32_t>(col,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))) * i));
uint32_t b = cast<uint32_t>((cast<float>(cast<uint32_t>((col & cast<uint32_t>(255ULL)))) * i));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((col & cast<uint32_t>(4278190080ULL))) | shl<uint32_t>(r,cast<int64_t>(16ULL)))) | shl<uint32_t>(g,cast<int64_t>(8ULL)))) | b));
}
}
// tools/platform/dc/pvr.go:398:1
uint32_t dc_lerp3(float w0,float w1,float w2,uint32_t a,uint32_t b,uint32_t c,uint64_t shift){
{
return cast<uint32_t>((((w0 * cast<float>(cast<uint32_t>((shr<uint32_t>(a,shift) & cast<uint32_t>(255ULL))))) + (w1 * cast<float>(cast<uint32_t>((shr<uint32_t>(b,shift) & cast<uint32_t>(255ULL)))))) + (w2 * cast<float>(cast<uint32_t>((shr<uint32_t>(c,shift) & cast<uint32_t>(255ULL)))))));
}
}
// tools/platform/dc/pvr.go:406:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> dc_renderState_shadePixel(dc_renderState* st,float w0,float w1,float w2,dc_pvrVert a,dc_pvrVert b,dc_pvrVert c,uint32_t tr,uint32_t tg,uint32_t tb,uint32_t ta){
uint32_t cr{};
uint32_t cg{};
uint32_t cb{};
uint32_t ca{};
{
uint32_t br={};
uint32_t bg={};
uint32_t bb={};
uint32_t ba={};
uint32_t or_={};
uint32_t og={};
uint32_t ob={};
if (st->gouraud) {
br = dc_lerp3(w0,w1,w2,a.color,b.color,c.color,cast<uint64_t>(16ULL));
bg = dc_lerp3(w0,w1,w2,a.color,b.color,c.color,cast<uint64_t>(8ULL));
bb = dc_lerp3(w0,w1,w2,a.color,b.color,c.color,cast<uint64_t>(0ULL));
ba = dc_lerp3(w0,w1,w2,a.color,b.color,c.color,cast<uint64_t>(24ULL));
if (st->offsEn) {
or_ = dc_lerp3(w0,w1,w2,a.offs,b.offs,c.offs,cast<uint64_t>(16ULL));
og = dc_lerp3(w0,w1,w2,a.offs,b.offs,c.offs,cast<uint64_t>(8ULL));
ob = dc_lerp3(w0,w1,w2,a.offs,b.offs,c.offs,cast<uint64_t>(0ULL));
}
}
else {
auto tmp96 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((a.color & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))));
br = std::get<0>(tmp96);
bg = std::get<1>(tmp96);
bb = std::get<2>(tmp96);
ba = std::get<3>(tmp96);
if (st->offsEn) {
auto tmp97 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(a.offs,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(a.offs,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((a.offs & cast<uint32_t>(255ULL))));
or_ = std::get<0>(tmp97);
og = std::get<1>(tmp97);
ob = std::get<2>(tmp97);
}
}
{
switch(st->shade){
case cast<uint32_t>(0ULL):{
auto tmp98 = std::make_tuple(tr,tg,tb,ta);
cr = std::get<0>(tmp98);
cg = std::get<1>(tmp98);
cb = std::get<2>(tmp98);
ca = std::get<3>(tmp98);
break;}
case cast<uint32_t>(1ULL):{
auto tmp99 = std::make_tuple(divi<uint32_t>(cast<uint32_t>((tr * br)),cast<uint32_t>(255ULL)),divi<uint32_t>(cast<uint32_t>((tg * bg)),cast<uint32_t>(255ULL)),divi<uint32_t>(cast<uint32_t>((tb * bb)),cast<uint32_t>(255ULL)),ta);
cr = std::get<0>(tmp99);
cg = std::get<1>(tmp99);
cb = std::get<2>(tmp99);
ca = std::get<3>(tmp99);
break;}
case cast<uint32_t>(2ULL):{
cr = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((tr * ta)) + cast<uint32_t>((br * (cast<uint32_t>((cast<uint32_t>(255ULL) - ta)))))))),cast<uint32_t>(255ULL));
cg = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((tg * ta)) + cast<uint32_t>((bg * (cast<uint32_t>((cast<uint32_t>(255ULL) - ta)))))))),cast<uint32_t>(255ULL));
cb = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((tb * ta)) + cast<uint32_t>((bb * (cast<uint32_t>((cast<uint32_t>(255ULL) - ta)))))))),cast<uint32_t>(255ULL));
ca = ba;
break;}
default:{
auto tmp100 = std::make_tuple(divi<uint32_t>(cast<uint32_t>((tr * br)),cast<uint32_t>(255ULL)),divi<uint32_t>(cast<uint32_t>((tg * bg)),cast<uint32_t>(255ULL)),divi<uint32_t>(cast<uint32_t>((tb * bb)),cast<uint32_t>(255ULL)),divi<uint32_t>(cast<uint32_t>((ta * ba)),cast<uint32_t>(255ULL)));
cr = std::get<0>(tmp100);
cg = std::get<1>(tmp100);
cb = std::get<2>(tmp100);
ca = std::get<3>(tmp100);
break;}
}}
auto tmp101 = std::make_tuple(dc_clamp255(cast<uint32_t>((cr + or_))),dc_clamp255(cast<uint32_t>((cg + og))),dc_clamp255(cast<uint32_t>((cb + ob))));
cr = std::get<0>(tmp101);
cg = std::get<1>(tmp101);
cb = std::get<2>(tmp101);
return {cr,cg,cb,ca};
}
}
// tools/platform/dc/pvr.go:442:1
uint32_t dc_clamp255(uint32_t v){
{
if ((v > cast<uint32_t>(255ULL))) {
return cast<uint32_t>(255ULL);
}
return v;
}
}
// tools/platform/dc/pvr.go:450:1
bool dc_depthPass(uint32_t mode,float z,float old){
{
{
switch(mode){
case cast<uint32_t>(0ULL):{
return false;
break;}
case cast<uint32_t>(1ULL):{
return (z < old);
break;}
case cast<uint32_t>(2ULL):{
return (z == old);
break;}
case cast<uint32_t>(3ULL):{
return (z <= old);
break;}
case cast<uint32_t>(4ULL):{
return (z > old);
break;}
case cast<uint32_t>(5ULL):{
return (z != old);
break;}
case cast<uint32_t>(6ULL):{
return (z >= old);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/dc/pvr.go:474:1
void dc_renderState_tri(dc_renderState* st,dc_pvrVert a,dc_pvrVert b,dc_pvrVert c){
{
int64_t minX = dc_imax(cast<int64_t>(0ULL),cast<int64_t>(dc_min3(a.x,b.x,c.x)));
int64_t maxX = dc_imin(cast<int64_t>((st->w - cast<int64_t>(1ULL))),cast<int64_t>(dc_max3(a.x,b.x,c.x)));
int64_t minY = dc_imax(cast<int64_t>(0ULL),cast<int64_t>(dc_min3(a.y,b.y,c.y)));
int64_t maxY = dc_imin(cast<int64_t>((st->h - cast<int64_t>(1ULL))),cast<int64_t>(dc_max3(a.y,b.y,c.y)));
if (((minX > maxX) || (minY > maxY))) {
return ;
}
float area = ((((b.x - a.x)) * ((c.y - a.y))) - (((c.x - a.x)) * ((b.y - a.y))));
if ((area == cast<float>(0ULL))) {
return ;
}
if (((((dc_debugPixelX >= minX) && (dc_debugPixelX <= maxX)) && (dc_debugPixelY >= minY)) && (dc_debugPixelY <= maxY))) {
dc_Machine_logf(st->m,std::string("debug-pixel tri tex=%v addr=%06X fmt=%d shade=%d offsEn=%v colType=%d gouraud=%v %dx%d blend=%v a.color=%08X a.offs=%08X",120),st->texture,st->texAddr,st->texFmt,st->shade,st->offsEn,st->colType,st->gouraud,st->texUW,st->texVH,st->blend,a.color,a.offs);
}
st->tris++;
float inv = (cast<float>(1ULL) / area);
auto tmp102 = std::make_tuple((a.u * a.z),(a.v * a.z));
float au = std::get<0>(tmp102);
float av = std::get<1>(tmp102);
auto tmp103 = std::make_tuple((b.u * b.z),(b.v * b.z));
float bu = std::get<0>(tmp103);
float bv = std::get<1>(tmp103);
auto tmp104 = std::make_tuple((c.u * c.z),(c.v * c.z));
float cu = std::get<0>(tmp104);
float cv = std::get<1>(tmp104);
{int64_t y = minY;for (;(y <= maxY);y++){
{int64_t x = minX;for (;(x <= maxX);x++){
auto tmp105 = std::make_tuple((cast<float>(x) + 0.5),(cast<float>(y) + 0.5));
float px = std::get<0>(tmp105);
float py = std::get<1>(tmp105);
float w0 = ((((((b.x - px)) * ((c.y - py))) - (((c.x - px)) * ((b.y - py))))) * inv);
float w1 = ((((((c.x - px)) * ((a.y - py))) - (((a.x - px)) * ((c.y - py))))) * inv);
float w2 = ((cast<float>(1ULL) - w0) - w1);
if ((((w0 < cast<float>(0ULL)) || (w1 < cast<float>(0ULL))) || (w2 < cast<float>(0ULL)))) {
continue;
}
float z = (((w0 * a.z) + (w1 * b.z)) + (w2 * c.z));
int64_t zi = cast<int64_t>((cast<int64_t>((y * st->w)) + x));
if ((!dc_depthPass(st->depthCmp,z,st->zbuf[zi]))) {
continue;
}
uint32_t cr={};
uint32_t cg={};
uint32_t cb={};
uint32_t ca={};
if (st->texture) {
float u = (((w0 * au) + (w1 * bu)) + (w2 * cu));
float v = (((w0 * av) + (w1 * bv)) + (w2 * cv));
if ((z != cast<float>(0ULL))) {
auto tmp106 = std::make_tuple((u / z),(v / z));
u = std::get<0>(tmp106);
v = std::get<1>(tmp106);
}
auto tmp107 = dc_renderState_sample(st,u,v);
uint32_t tr = std::get<0>(tmp107);
uint32_t tg = std::get<1>(tmp107);
uint32_t tb = std::get<2>(tmp107);
uint32_t ta = std::get<3>(tmp107);
auto tmp108 = dc_renderState_shadePixel(st,w0,w1,w2,a,b,c,tr,tg,tb,ta);
cr = std::get<0>(tmp108);
cg = std::get<1>(tmp108);
cb = std::get<2>(tmp108);
ca = std::get<3>(tmp108);
}
else if (st->gouraud) {
cr = cast<uint32_t>((((w0 * cast<float>(cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))))) + (w1 * cast<float>(cast<uint32_t>((shr<uint32_t>(b.color,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))))) + (w2 * cast<float>(cast<uint32_t>((shr<uint32_t>(c.color,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))))));
cg = cast<uint32_t>((((w0 * cast<float>(cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))))) + (w1 * cast<float>(cast<uint32_t>((shr<uint32_t>(b.color,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))))) + (w2 * cast<float>(cast<uint32_t>((shr<uint32_t>(c.color,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))))));
cb = cast<uint32_t>((((w0 * cast<float>(cast<uint32_t>((a.color & cast<uint32_t>(255ULL))))) + (w1 * cast<float>(cast<uint32_t>((b.color & cast<uint32_t>(255ULL)))))) + (w2 * cast<float>(cast<uint32_t>((c.color & cast<uint32_t>(255ULL)))))));
ca = cast<uint32_t>((((w0 * cast<float>(cast<uint32_t>((shr<uint32_t>(a.color,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))))) + (w1 * cast<float>(cast<uint32_t>((shr<uint32_t>(b.color,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))))) + (w2 * cast<float>(cast<uint32_t>((shr<uint32_t>(c.color,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))))));
}
else {
uint32_t col = a.color;
auto tmp109 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(col,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(col,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((col & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(col,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))));
cr = std::get<0>(tmp109);
cg = std::get<1>(tmp109);
cb = std::get<2>(tmp109);
ca = std::get<3>(tmp109);
}
if (dc_renderState_plot(st,x,y,cr,cg,cb,ca)) {
if (((x == dc_debugPixelX) && (y == dc_debugPixelY))) {
dc_Machine_logf(st->m,std::string("debug-pixel plot tex=%v addr=%06X rgba=%d,%d,%d,%d z=%v",55),st->texture,st->texAddr,cr,cg,cb,ca,z);
}
if (st->zWrite) {
st->zbuf[zi] = z;
}
}
}
}}
}}
}
// tools/platform/dc/pvr.go:546:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> dc_renderState_sample(dc_renderState* st,float u,float v){
uint32_t r{};
uint32_t g{};
uint32_t b{};
uint32_t a{};
{
int64_t tx = cast<int64_t>((cast<int64_t>((u * cast<float>(st->texUW))) & (cast<int64_t>((st->texUW - cast<int64_t>(1ULL))))));
int64_t ty = cast<int64_t>((cast<int64_t>((v * cast<float>(st->texVH))) & (cast<int64_t>((st->texVH - cast<int64_t>(1ULL))))));
if (((st->texFmt == cast<uint32_t>(5ULL)) || (st->texFmt == cast<uint32_t>(6ULL)))) {
int64_t side = st->texUW;
if ((st->texVH < side)) {
side = st->texVH;
}
int64_t block = cast<int64_t>((cast<int64_t>(((divi<int64_t>(ty,side)) * (divi<int64_t>(st->texUW,side)))) + divi<int64_t>(tx,side)));
uint32_t idx = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((block * side)) * side))) + dc_twiddle(cast<uint32_t>(modi<int64_t>(tx,side)),cast<uint32_t>(modi<int64_t>(ty,side)))));
uint32_t entry={};
if ((st->texFmt == cast<uint32_t>(6ULL))) {
uint32_t off = cast<uint32_t>((st->texAddr + idx));
if ((off >= cast<uint32_t>(8388608ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
entry = cast<uint32_t>((shl<uint32_t>(st->texBank,cast<int64_t>(8ULL)) | cast<uint32_t>(st->m->VRAM[off])));
}
else {
uint32_t off = cast<uint32_t>((st->texAddr + divi<uint32_t>(idx,cast<uint32_t>(2ULL))));
if ((off >= cast<uint32_t>(8388608ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
entry = cast<uint32_t>((shl<uint32_t>(st->texBank,cast<int64_t>(4ULL)) | cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(st->m->VRAM[off],(cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((idx & cast<uint32_t>(1ULL))))))))) & cast<uint32_t>(15ULL)))));
}
uint32_t pal = st->m->PVRRegs[cast<uint32_t>((cast<uint32_t>(1024ULL) + cast<uint32_t>((entry & cast<uint32_t>(1023ULL)))))];
{
switch(cast<uint32_t>((st->m->PVRRegs[cast<int64_t>(66ULL)] & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return {shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(10ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL)),shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL)),shl<uint32_t>(cast<uint32_t>((pal & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL)),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(15ULL)) & cast<uint32_t>(1ULL))) * cast<uint32_t>(255ULL)))};
break;}
case cast<uint32_t>(1ULL):{
return {shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL)),shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(5ULL)) & cast<uint32_t>(63ULL))),cast<int64_t>(2ULL)),shl<uint32_t>(cast<uint32_t>((pal & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL)),cast<uint32_t>(255ULL)};
break;}
case cast<uint32_t>(2ULL):{
return {cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL))),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL))),cast<uint32_t>((cast<uint32_t>((pal & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL))),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(12ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL)))};
break;}
default:{
return {cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((pal & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(pal,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))};
break;}
}}
}
uint32_t px={};
if (st->texVQ) {
int64_t bw = divi<int64_t>(st->texUW,cast<int64_t>(2ULL));
int64_t bh = divi<int64_t>(st->texVH,cast<int64_t>(2ULL));
int64_t side = bw;
if ((bh < side)) {
side = bh;
}
auto tmp110 = std::make_tuple(divi<int64_t>(tx,cast<int64_t>(2ULL)),divi<int64_t>(ty,cast<int64_t>(2ULL)));
int64_t bx = std::get<0>(tmp110);
int64_t by = std::get<1>(tmp110);
uint32_t block = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>(((divi<int64_t>(by,side)) * (divi<int64_t>(bw,side)))) + divi<int64_t>(bx,side)))) * cast<uint32_t>(cast<int64_t>((side * side))))) + dc_twiddle(cast<uint32_t>(modi<int64_t>(bx,side)),cast<uint32_t>(modi<int64_t>(by,side)))));
uint32_t iOff = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((st->texAddr + cast<uint32_t>(2048ULL))) + st->texMipIx)) + block));
if ((iOff >= cast<uint32_t>(8388608ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
uint32_t entry = cast<uint32_t>((cast<uint32_t>((st->texAddr + cast<uint32_t>((cast<uint32_t>(st->m->VRAM[iOff]) * cast<uint32_t>(8ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>(((cast<int64_t>((tx & cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL))) + (cast<int64_t>((ty & cast<int64_t>(1ULL))))))) * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>((entry + cast<uint32_t>(2ULL))) > cast<uint32_t>(8388608ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
px = cast<uint32_t>((cast<uint32_t>(st->m->VRAM[entry]) | shl<uint32_t>(cast<uint32_t>(st->m->VRAM[cast<uint32_t>((entry + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
else {
uint32_t idx={};
if (st->texTwid) {
int64_t side = st->texUW;
if ((st->texVH < side)) {
side = st->texVH;
}
int64_t block = cast<int64_t>((cast<int64_t>(((divi<int64_t>(ty,side)) * (divi<int64_t>(st->texUW,side)))) + divi<int64_t>(tx,side)));
idx = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((block * side)) * side))) + dc_twiddle(cast<uint32_t>(modi<int64_t>(tx,side)),cast<uint32_t>(modi<int64_t>(ty,side)))));
}
else {
idx = cast<uint32_t>(cast<int64_t>((cast<int64_t>((ty * st->texUW)) + tx)));
}
uint32_t off = cast<uint32_t>((st->texAddr + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>((off + cast<uint32_t>(2ULL))) > cast<uint32_t>(8388608ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
px = cast<uint32_t>((cast<uint32_t>(st->m->VRAM[off]) | shl<uint32_t>(cast<uint32_t>(st->m->VRAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
{
switch(st->texFmt){
case cast<uint32_t>(0ULL):{
a = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(15ULL)) & cast<uint32_t>(1ULL))) * cast<uint32_t>(255ULL)));
r = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(10ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
g = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
b = shl<uint32_t>(cast<uint32_t>((px & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
break;}
case cast<uint32_t>(1ULL):{
a = cast<uint32_t>(255ULL);
r = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
g = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(5ULL)) & cast<uint32_t>(63ULL))),cast<int64_t>(2ULL));
b = shl<uint32_t>(cast<uint32_t>((px & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
break;}
case cast<uint32_t>(2ULL):{
a = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(12ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL)));
r = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL)));
g = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL)));
b = cast<uint32_t>((cast<uint32_t>((px & cast<uint32_t>(15ULL))) * cast<uint32_t>(17ULL)));
break;}
default:{
dc_renderState_skip(st,std::string("texture format",14),st->texFmt);
return {cast<uint32_t>(255ULL),cast<uint32_t>(0ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL)};
break;}
}}
return {r,g,b,a};
}
}
// tools/platform/dc/pvr.go:657:1
bool dc_renderState_plot(dc_renderState* st,int64_t x,int64_t y,uint32_t r,uint32_t g,uint32_t b,uint32_t a){
{
bool drawn = dc_renderState_plotStore(st,x,y,r,g,b,a);
if (bool(st->m->OnPVRPixel)) {
st->m->OnPVRPixel(x,y,cast<uint8_t>(r),cast<uint8_t>(g),cast<uint8_t>(b),cast<uint8_t>(a),drawn);
}
return drawn;
}
}
// tools/platform/dc/pvr.go:665:1
bool dc_renderState_plotStore(dc_renderState* st,int64_t x,int64_t y,uint32_t r,uint32_t g,uint32_t b,uint32_t a){
{
uint32_t off = cast<uint32_t>((st->base + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((y * st->w)) + x))) * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>((off + cast<uint32_t>(2ULL))) > cast<uint32_t>(8388608ULL))) {
return false;
}
off = dc_vram32to64(off);
if ((a == cast<uint32_t>(0ULL))) {
return false;
}
if (st->blend) {
uint32_t old = cast<uint32_t>((cast<uint32_t>(st->m->VRAM[off]) | shl<uint32_t>(cast<uint32_t>(st->m->VRAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
uint32_t or_ = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(old,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
uint32_t og = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(old,cast<int64_t>(5ULL)) & cast<uint32_t>(63ULL))),cast<int64_t>(2ULL));
uint32_t ob = shl<uint32_t>(cast<uint32_t>((old & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
r = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((r * a)) + cast<uint32_t>((or_ * (cast<uint32_t>((cast<uint32_t>(255ULL) - a)))))))),cast<uint32_t>(255ULL));
g = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((g * a)) + cast<uint32_t>((og * (cast<uint32_t>((cast<uint32_t>(255ULL) - a)))))))),cast<uint32_t>(255ULL));
b = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((b * a)) + cast<uint32_t>((ob * (cast<uint32_t>((cast<uint32_t>(255ULL) - a)))))))),cast<uint32_t>(255ULL));
}
st->px++;
uint16_t px = cast<uint16_t>(cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(shr<uint32_t>(r,cast<int64_t>(3ULL)),cast<int64_t>(11ULL)) | shl<uint32_t>(shr<uint32_t>(g,cast<int64_t>(2ULL)),cast<int64_t>(5ULL)))) | shr<uint32_t>(b,cast<int64_t>(3ULL)))));
st->m->VRAM[off] = cast<uint8_t>(px);
if(rrcapture::trace.active)rrDCWrite(st->m,st->m->VRAM,off);
st->m->VRAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint16_t>(px,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrDCWrite(st->m,st->m->VRAM,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
return true;
}
}
// tools/platform/dc/pvr.go:694:1
uint32_t dc_twiddle_reference(uint32_t x,uint32_t y){
{
uint32_t out={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
out |= shl<uint32_t>((cast<uint32_t>((shr<uint32_t>(y,i) & cast<uint32_t>(1ULL)))),(cast<uint32_t>((cast<uint32_t>(2ULL) * i))));
out |= shl<uint32_t>((cast<uint32_t>((shr<uint32_t>(x,i) & cast<uint32_t>(1ULL)))),(cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * i)) + cast<uint32_t>(1ULL)))));
}
}return out;
}
}
// tools/platform/dc/pvr.go:703:1
std::tuple<float,float> dc_unpackUV16(uint32_t w){
float u{};
float v{};
{
return {dc_float32frombits(cast<uint32_t>((w & cast<uint32_t>(4294901760ULL)))),dc_float32frombits(shl<uint32_t>(w,cast<int64_t>(16ULL)))};
}
}
// tools/platform/dc/pvr.go:707:1
float dc_float32frombits(uint32_t b){
{
return go_math_Float32frombits(b);
}
}
// tools/platform/dc/pvr.go:709:1
float dc_min3(float a,float b,float c){
{
if ((b < a)) {
a = b;
}
if ((c < a)) {
a = c;
}
return a;
}
}
// tools/platform/dc/pvr.go:719:1
float dc_max3(float a,float b,float c){
{
if ((b > a)) {
a = b;
}
if ((c > a)) {
a = c;
}
return a;
}
}
// tools/platform/dc/pvr.go:729:1
int64_t dc_imin(int64_t a,int64_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/dc/pvr.go:736:1
int64_t dc_imax(int64_t a,int64_t b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/dc/run.go:17:1
std::string dc_Result_String(dc_Result r){
{
return go_fmt_Sprintf(std::string("stopped after %d steps at PC %08X: %s",37),r.Steps,r.PC,r.Reason);
}
}
// tools/platform/dc/run.go:33:1
dc_Result dc_Machine_Run(dc_Machine* m,uint64_t maxSteps,dc_RunConfig cfg){
{rrprof::Scope timing(0,"SH-4 / ARM7 and devices");
{
std::array<uint32_t,4> spinPCs={};
int64_t spinCount={};
uint64_t spinSteps={};
{uint64_t start = m->Instrs;for (;(cast<uint64_t>((m->Instrs - start)) < maxSteps);){
if (m->StopRequested) {
m->StopRequested = false;
return dc_Result{m->Instrs,m->CPU->PC,std::string("stop requested",14)};
}
uint32_t pc = m->CPU->PC;
if (get(cfg.Breakpoints,pc)) {
return dc_Result{m->Instrs,pc,std::string("breakpoint",10)};
}
if (bool(m->OnStep)) {
m->OnStep(pc);
}
if (dc_Machine_trapPC(m,pc)) {
continue;
}
sh4_CPU_Step(m->CPU);
m->Instrs++;
if (m->CPU->Halted) {
return dc_Result{m->Instrs,sh4_CPU_CurPC(m->CPU),(std::string("halt: ",6) + m->CPU->HaltReason)};
}
{
m->armAcc++;
if ((m->armAcc >= cast<int64_t>(9ULL))) {
m->armAcc = cast<int64_t>(0ULL);
dc_Machine_stepARM(m);
}
}
dc_Machine_tickAICA(m);
dc_Machine_tickCompletions(m);
dc_Machine_tickField(m);
if ((!cfg.NoSpin)) {
bool known = false;
{int64_t i = cast<int64_t>(0ULL);for (;(i < spinCount);i++){
if ((spinPCs[i] == pc)) {
known = true;
break;
}
}
}if ((!known)) {
if ((spinCount < cast<int64_t>(4ULL))) {
spinPCs[spinCount] = pc;
spinCount++;
}
else {
auto tmp111 = std::make_tuple(cast<int64_t>(0ULL),cast<uint64_t>(0ULL));
spinCount = std::get<0>(tmp111);
spinSteps = std::get<1>(tmp111);
auto tmp112 = std::make_tuple(pc,cast<int64_t>(1ULL));
spinPCs[cast<int64_t>(0ULL)] = std::get<0>(tmp112);
spinCount = std::get<1>(tmp112);
}
}
spinSteps++;
if (((spinSteps >= cast<uint64_t>(6666666ULL)) && (spinCount <= cast<int64_t>(4ULL)))) {
return dc_Result{m->Instrs,pc,go_fmt_Sprintf(std::string("spin: %d distinct PCs over %d instructions",42),spinCount,spinSteps)};
}
}
}
}return dc_Result{m->Instrs,m->CPU->PC,std::string("steps",5)};
}
}
}
// tools/platform/dc/run.go:97:1
dc_Result dc_Machine_RunFields(dc_Machine* m,uint64_t n,uint64_t budget){
{
uint64_t target = cast<uint64_t>((m->Fields + n));
{;for (;(m->Fields < target);){
dc_Result r = dc_Machine_Run(m,dc_minU64(budget,cast<uint64_t>(3333334ULL)),dc_RunConfig{});
budget -= dc_minU64(budget,cast<uint64_t>(3333334ULL));
if ((r.Reason != std::string("steps",5))) {
return r;
}
if ((budget == cast<uint64_t>(0ULL))) {
return dc_Result{m->Instrs,m->CPU->PC,std::string("field budget exhausted",22)};
}
}
}return dc_Result{m->Instrs,m->CPU->PC,std::string("fields",6)};
}
}
// tools/platform/dc/run.go:112:1
uint64_t dc_minU64(uint64_t a,uint64_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/dc/run.go:124:1
void dc_Machine_tickField_reference(dc_Machine* m){
{
m->instrInField++;
uint32_t total = dc_Machine_spgTotalLines(m);
if ((m->instrInField < cast<uint64_t>(divi<uint32_t>(cast<uint32_t>(3333333ULL),total)))) {
return ;
}
m->instrInField = cast<uint64_t>(0ULL);
m->CurLine++;
if ((m->CurLine >= total)) {
m->CurLine = cast<uint32_t>(0ULL);
m->FieldNum ^= cast<uint32_t>(1ULL);
}
uint32_t vbl = m->PVRRegs[cast<int64_t>(51ULL)];
{
auto tmp114=m->CurLine;
if (tmp114==(cast<uint32_t>((vbl & cast<uint32_t>(1023ULL))))){
m->Fields++;
dc_Machine_raiseNRM(m,cast<uint32_t>(8ULL));
if (bool(m->OnDisplay)) {
m->OnDisplay(m->Fields);
}
}
else if (tmp114==(cast<uint32_t>((shr<uint32_t>(vbl,cast<int64_t>(16ULL)) & cast<uint32_t>(1023ULL))))){
dc_Machine_raiseNRM(m,cast<uint32_t>(16ULL));
}
}
tmp113:;
}
}
// tools/platform/dc/run.go:151:1
uint32_t dc_Machine_spgTotalLines(dc_Machine* m){
{
{
uint32_t t = cast<uint32_t>((shr<uint32_t>(m->PVRRegs[cast<int64_t>(62ULL)],cast<int64_t>(16ULL)) & cast<uint32_t>(1023ULL)));
if ((t != cast<uint32_t>(0ULL))) {
return cast<uint32_t>((t + cast<uint32_t>(1ULL)));
}
}
return cast<uint32_t>(525ULL);
}
}
// tools/platform/dc/video.go:27:1
std::tuple<image_RGBA*,Error> dc_Machine_RenderFB(dc_Machine* m){
{
uint32_t ctrl = m->PVRRegs[cast<int64_t>(17ULL)];
if ((cast<uint32_t>((ctrl & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("no framebuffer: FB_R_CTRL enable is clear (nothing has configured video yet)",76))};
}
uint32_t depth = cast<uint32_t>(((shr<uint32_t>(ctrl,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
uint32_t size = m->PVRRegs[cast<int64_t>(23ULL)];
int64_t xUnits = cast<int64_t>((cast<int64_t>(cast<uint32_t>((size & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
int64_t lines = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(size,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
int64_t modulus = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(size,cast<int64_t>(20ULL))) & cast<uint32_t>(1023ULL)))) - cast<int64_t>(1ULL)));
uint32_t base = cast<uint32_t>((m->PVRRegs[cast<int64_t>(20ULL)] & cast<uint32_t>(16777215ULL)));
int64_t bytesPerLine = cast<int64_t>((xUnits * cast<int64_t>(4ULL)));
int64_t bpp={};
{
switch(depth){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
bpp = cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(2ULL):{
bpp = cast<int64_t>(3ULL);
break;}
default:{
bpp = cast<int64_t>(4ULL);
break;}
}}
int64_t w = divi<int64_t>(bytesPerLine,bpp);
if (((w <= cast<int64_t>(0ULL)) || (lines <= cast<int64_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("degenerate framebuffer: FB_R_SIZE=%08X",38),size)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,lines));
int64_t off = cast<int64_t>(base);
{int64_t y = cast<int64_t>(0ULL);for (;(y < lines);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
int64_t p = cast<int64_t>((off + cast<int64_t>((x * bpp))));
if ((cast<int64_t>((p + bpp)) > len(m->VRAM))) {
return {{},go_fmt_Errorf(std::string("framebuffer overruns VRAM at line %d",36),y)};
}
uint8_t b0 = m->VRAM[dc_vram32to64(cast<uint32_t>(p))];
uint8_t b1 = m->VRAM[dc_vram32to64(cast<uint32_t>(cast<int64_t>((p + cast<int64_t>(1ULL)))))];
uint8_t r={};
uint8_t g={};
uint8_t b={};
{
switch(depth){
case cast<uint32_t>(0ULL):{
uint16_t v = cast<uint16_t>((cast<uint16_t>(b0) | shl<uint16_t>(cast<uint16_t>(b1),cast<int64_t>(8ULL))));
r = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(10ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
g = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
b = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
break;}
case cast<uint32_t>(1ULL):{
uint16_t v = cast<uint16_t>((cast<uint16_t>(b0) | shl<uint16_t>(cast<uint16_t>(b1),cast<int64_t>(8ULL))));
r = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
g = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL)))),cast<int64_t>(2ULL));
b = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
break;}
default:{
auto tmp115 = std::make_tuple(b0,b1,m->VRAM[dc_vram32to64(cast<uint32_t>(cast<int64_t>((p + cast<int64_t>(2ULL)))))]);
b = std::get<0>(tmp115);
g = std::get<1>(tmp115);
r = std::get<2>(tmp115);
break;}
}}
int64_t i = image_RGBA_PixOffset(img,x,y);
auto tmp116 = std::make_tuple(r,g,b,cast<uint8_t>(255ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = std::get<0>(tmp116);
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = std::get<1>(tmp116);
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = std::get<2>(tmp116);
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = std::get<3>(tmp116);
}
}off += cast<int64_t>((bytesPerLine + cast<int64_t>((modulus * cast<int64_t>(4ULL)))));
}
}return {img,{}};
}
}
// tools/platform/dc/video.go:97:1
std::tuple<image_RGBA*,Error> dc_Machine_RenderDrawTarget(dc_Machine* m){
{
if ((cast<uint32_t>((m->PVRRegs[cast<int64_t>(18ULL)] & cast<uint32_t>(7ULL))) != cast<uint32_t>(1ULL))) {
return {{},go_fmt_Errorf(std::string("draw target: FB_W_CTRL packmode %d unimplemented (only 565)",59),cast<uint32_t>((m->PVRRegs[cast<int64_t>(18ULL)] & cast<uint32_t>(7ULL))))};
}
uint32_t size = m->PVRRegs[cast<int64_t>(23ULL)];
int64_t w = cast<int64_t>(((cast<int64_t>((cast<int64_t>(cast<uint32_t>((size & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL)));
int64_t h = cast<int64_t>((cast<int64_t>(cast<uint32_t>((shr<uint32_t>(size,cast<int64_t>(10ULL)) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
uint32_t base = cast<uint32_t>((m->PVRRegs[cast<int64_t>(24ULL)] & cast<uint32_t>(16777215ULL)));
if ((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (size == cast<uint32_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("draw target: degenerate FB_R_SIZE=%08X",38),size)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t off = cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>((off + cast<uint32_t>(2ULL))) > cast<uint32_t>(8388608ULL))) {
continue;
}
off = dc_vram32to64(off);
uint16_t v = cast<uint16_t>((cast<uint16_t>(m->VRAM[off]) | shl<uint16_t>(cast<uint16_t>(m->VRAM[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
int64_t i = image_RGBA_PixOffset(img,x,y);
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL)))),cast<int64_t>(2ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}}
}return {img,{}};
}
}
// tools/platform/dc/video.go:129:1
std::tuple<image_RGBA*,Error> dc_Machine_RenderVRAM(dc_Machine* m,uint32_t offset,int64_t w,int64_t h){
{
if ((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (cast<int64_t>((cast<int64_t>(offset) + cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(2ULL))))) > len(m->VRAM)))) {
return {{},go_fmt_Errorf(std::string("VRAM window %08X %dx%d out of range",35),offset,w,h)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
int64_t p = cast<int64_t>((cast<int64_t>(offset) + cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(2ULL)))));
uint16_t v = cast<uint16_t>((cast<uint16_t>(m->VRAM[p]) | shl<uint16_t>(cast<uint16_t>(m->VRAM[cast<int64_t>((p + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
int64_t i = image_RGBA_PixOffset(img,x,y);
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL)))),cast<int64_t>(2ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = shl<uint8_t>(cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL));
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}}
}return {img,{}};
}
}

#include "fast.h"
