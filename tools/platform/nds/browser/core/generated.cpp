#include "runtime.h"
struct arm_Inst;
struct arm_CPU;
struct arm_Banks;
struct arm_vfpState;
struct nds_Header;
struct nds_FATEntry;
struct nds_ROM;
struct nds_FileInfo;
struct nds_Overlay;
struct dsmachine_bus;
struct dsmachine_card;
struct dsmachine_PixelEvent;
struct dsmachine_gxHalt;
struct dsmachine_MemRegion;
struct dsmachine_dmaChan;
struct dsmachine_gpu2d;
struct dsmachine_engine;
struct dsmachine_cand;
struct dsmachine_gpu3d;
struct dsmachine_mtx;
struct dsmachine_gxVertex;
struct dsmachine_gxPolygon;
struct dsmachine_geom;
struct dsmachine_rfrag;
struct dsmachine_raster;
struct dsmachine_rvert;
struct dsmachine_polyState;
struct dsmachine_texState;
struct dsmachine_Machine;
struct dsmachine_core;
struct dsmachine_ipc;
struct dsmachine_divider;
struct dsmachine_sqrter;
struct dsmachine_Profile;
struct dsmachine_profiler;
struct dsmachine_Result;
struct dsmachine_spibus;
struct dsmachine_timer;
struct dsmachine_video;
struct dsmachine_vram;
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
struct Anon32;
struct Anon33;
struct Anon34;
struct Anon35;
struct Anon36;
struct Anon37;
struct Anon38;
struct Anon39;
struct Anon4;
struct Anon40;
struct Anon41;
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
dsmachine_bus* bus{};
dsmachine_bus* wide{};
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
struct nds_Header{
std::string Title{};
std::string GameCode{};
std::string MakerCode{};
uint8_t UnitCode{};
uint8_t SeedSelect{};
uint8_t DeviceCap{};
uint8_t ROMVersion{};
uint8_t AutoStart{};
uint32_t ARM9ROMOff{};
uint32_t ARM9Entry{};
uint32_t ARM9RAMAddr{};
uint32_t ARM9Size{};
uint32_t ARM7ROMOff{};
uint32_t ARM7Entry{};
uint32_t ARM7RAMAddr{};
uint32_t ARM7Size{};
uint32_t FNTOff{};
uint32_t FNTSize{};
uint32_t FATOff{};
uint32_t FATSize{};
uint32_t ARM9OverlayOff{};
uint32_t ARM9OverlaySize{};
uint32_t ARM7OverlayOff{};
uint32_t ARM7OverlaySize{};
uint32_t IconOff{};
uint16_t SecureCRC{};
uint32_t ARM9HookRAM{};
uint32_t ARM7HookRAM{};
uint32_t TotalUsedSize{};
uint32_t HeaderSize{};
uint16_t LogoCRC{};
uint16_t HeaderCRC{};
};
struct nds_FATEntry{
uint32_t Start{};
uint32_t End{};
};
struct nds_FileInfo{
int64_t ID{};
std::string Path{};
};
struct nds_ROM{
Slice<uint8_t> Data{};
nds_Header Header{};
Slice<nds_FATEntry> FAT{};
Slice<nds_FileInfo> Files{};
Map<std::string,int64_t> byPath{};
};
struct nds_Overlay{
uint32_t ID{};
uint32_t RAMAddr{};
uint32_t RAMSize{};
uint32_t BSSSize{};
uint32_t StaticInitStart{};
uint32_t StaticInitEnd{};
uint32_t FileID{};
uint32_t CompressedSize{};
bool Compressed{};
};
struct dsmachine_bus{
dsmachine_core* c{};
};
struct dsmachine_card{
Slice<uint8_t> rom{};
std::array<uint8_t,8> cmd{};
uint32_t ctrl{};
Slice<uint8_t> buf{};
int64_t pos{};
bool owner9{};
std::function<void(std::array<uint8_t,8>,uint32_t,int64_t)> OnXfer{};
};
struct dsmachine_PixelEvent{
bool Drawn{};
bool ZReject{};
bool AlphaReject{};
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
};
struct dsmachine_gxHalt{
};
struct dsmachine_MemRegion{
std::string Name{};
uint32_t Base{};
uint32_t Size{};
};
struct dsmachine_dmaChan{
uint32_t src{};
uint32_t dst{};
uint32_t ctrl{};
uint32_t count{};
uint32_t csrc{};
uint32_t cdst{};
uint32_t crem{};
bool active{};
};
struct dsmachine_engine{
dsmachine_Machine* m{};
bool isB{};
uint32_t ioBase{};
uint32_t palBG{};
uint32_t palOBJ{};
uint32_t oamBase{};
int64_t spBG{};
int64_t spOBJ{};
int64_t spBGExt{};
int64_t spOBJExt{};
Slice<uint32_t> out{};
Slice<uint32_t> threeD{};
std::array<std::array<uint16_t,256>,4> bg{};
std::array<std::array<bool,256>,4> bgOK{};
std::array<uint8_t,256> a3D{};
bool is3D{};
std::array<uint8_t,4> prio{};
std::array<bool,4> shown{};
std::array<uint16_t,256> obj{};
std::array<bool,256> objOK{};
std::array<uint8_t,256> objPrio{};
std::array<bool,256> objSemi{};
std::array<uint8_t,256> objEVA{};
std::array<bool,256> objWin{};
std::array<uint8_t,256> win{};
std::array<uint16_t,256> line{};
std::array<bool,49152> vis3D{};
};
struct dsmachine_gpu2d{
dsmachine_engine a{};
dsmachine_engine b{};
Slice<uint32_t> threeD{};
bool swap{};
};
struct dsmachine_cand{
uint16_t c{};
int64_t layer{};
bool semi{};
uint8_t eva{};
uint8_t a3d{};
};
struct dsmachine_mtx{
std::array<int32_t,16> m{};
};
struct dsmachine_gxVertex{
int64_t x{};
int64_t y{};
int64_t z{};
int64_t w{};
int32_t r{};
int32_t g{};
int32_t b{};
int32_t s{};
int32_t t{};
};
struct dsmachine_gxPolygon{
Slice<dsmachine_gxVertex> verts{};
uint32_t attr{};
uint32_t texParam{};
uint32_t pltt{};
bool wbuffer{};
int64_t cmd{};
};
struct dsmachine_geom{
int64_t mode{};
dsmachine_mtx proj{};
dsmachine_mtx pos{};
dsmachine_mtx vec{};
dsmachine_mtx tex{};
std::array<dsmachine_mtx,1> projStack{};
std::array<dsmachine_mtx,1> texStack{};
std::array<dsmachine_mtx,32> posStack{};
std::array<dsmachine_mtx,32> vecStack{};
int64_t sp{};
int64_t projSP{};
int64_t texSP{};
bool clipDirty{};
dsmachine_mtx clipMtx{};
bool begun{};
int64_t primMode{};
Slice<dsmachine_gxVertex> strip{};
int64_t stripLen{};
std::array<int32_t,3> lastVtx{};
std::array<int32_t,3> color{};
int32_t texS{};
int32_t texT{};
uint32_t attr{};
uint32_t attrNext{};
uint32_t texParam{};
uint32_t pltt{};
std::array<std::array<int32_t,3>,4> lightVec{};
std::array<std::array<int32_t,3>,4> lightColor{};
std::array<int32_t,3> diffuse{};
std::array<int32_t,3> ambient{};
std::array<int32_t,3> specular{};
std::array<int32_t,3> emission{};
std::array<uint8_t,128> shininess{};
int32_t viewX1{};
int32_t viewY1{};
int32_t viewX2{};
int32_t viewY2{};
Slice<dsmachine_gxVertex> verts{};
Slice<dsmachine_gxPolygon> polys{};
std::array<int32_t,4> posResult{};
std::array<int32_t,3> vecResult{};
bool wbuffer{};
int64_t nEmit{};
int64_t nClipped{};
};
struct dsmachine_rfrag{
uint8_t r{};
uint8_t g{};
uint8_t b{};
uint8_t a{};
};
struct dsmachine_raster{
std::array<dsmachine_rfrag,49152> col{};
std::array<uint32_t,49152> depth{};
Slice<uint32_t> frame{};
};
struct dsmachine_gpu3d{
Slice<uint32_t> fifo{};
Slice<uint8_t> packed{};
uint8_t cmd{};
Slice<uint32_t> params{};
int64_t need{};
dsmachine_geom geom{};
dsmachine_raster rast{};
Map<uint32_t,uint32_t> regs{};
bool swapPending{};
uint32_t swapMode{};
int64_t lastPolys{};
int64_t swaps{};
std::array<int64_t,256> cmdHist{};
std::function<void(uint8_t,Slice<uint32_t>)> OnCmd{};
int64_t count{};
int64_t limit{};
int64_t cur{};
};
struct dsmachine_rvert{
double x{};
double y{};
double depth{};
double iw{};
double w{};
double r{};
double g{};
double b{};
double s{};
double t{};
};
struct dsmachine_texState{
uint32_t base{};
uint32_t pal{};
int64_t sizeS{};
int64_t sizeT{};
uint32_t format{};
bool repeatS{};
bool repeatT{};
bool flipS{};
bool flipT{};
bool color0{};
};
struct dsmachine_polyState{
dsmachine_texState tex{};
bool hasTex{};
uint32_t mode{};
uint8_t alpha{};
bool trans{};
bool depthEq{};
bool depthWr{};
bool wbuf{};
bool blend{};
bool toonHi{};
bool alphaTe{};
uint8_t alphaRf{};
std::array<std::array<uint8_t,3>,32> toon{};
bool wireframe{};
};
struct dsmachine_video{
int64_t line{};
bool hblank{};
uint64_t frames{};
};
struct dsmachine_divider{
uint32_t cnt{};
uint64_t numer{};
uint64_t denom{};
uint64_t result{};
uint64_t rem{};
};
struct dsmachine_sqrter{
uint32_t cnt{};
uint64_t param{};
uint32_t result{};
};
struct dsmachine_profiler{
bool on{};
time_Time frameStart{};
time_Duration geom{};
time_Duration raster{};
time_Duration compose{};
time_Duration dma{};
int64_t polys{};
int64_t frags{};
int64_t xfers{};
int64_t cmdStart{};
};
struct dsmachine_ipc{
uint8_t sync9{};
uint8_t sync7{};
Slice<uint32_t> to7{};
Slice<uint32_t> to9{};
};
struct dsmachine_Machine{
Slice<uint8_t> ram{};
Slice<uint8_t> swram{};
Slice<uint8_t> pal{};
Slice<uint8_t> oam{};
dsmachine_vram* vram{};
dsmachine_card* cd{};
dsmachine_spibus* spi{};
dsmachine_gpu2d* gpu2d{};
dsmachine_gpu3d* gpu3d{};
dsmachine_video vid{};
dsmachine_divider div{};
dsmachine_sqrter sqrt{};
dsmachine_profiler prof{};
uint32_t powcnt{};
uint32_t keys{};
uint8_t wramcnt{};
dsmachine_ipc ipc{};
dsmachine_core* ARM9{};
dsmachine_core* ARM7{};
uint64_t Steps{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
Map<uint32_t,bool> visited{};
std::function<void(std::string,uint8_t,uint32_t)> SyncTrace{};
std::function<void(bool,uint32_t)> OnStep{};
std::function<void(bool,bool,uint32_t,uint32_t,uint32_t)> OnIO{};
std::function<void(bool,uint32_t,uint8_t,uint32_t)> OnWrite{};
std::function<void(bool,uint32_t,uint32_t,uint32_t)> OnIRQ{};
std::function<void(bool,uint32_t,uint8_t,uint32_t)> OnRead{};
std::function<void(int64_t,int64_t,dsmachine_PixelEvent)> OnPixel{};
std::function<void(int64_t)> OnPoly{};
Map<uint32_t,bool> bps{};
bool stop{};
bool stopped{};
uint32_t stoppedPC{};
std::function<void()> OnFrame{};
};
struct dsmachine_timer{
uint16_t counter{};
uint16_t reload{};
uint16_t ctrl{};
int64_t frac{};
};
struct dsmachine_core{
dsmachine_Machine* m{};
arm_CPU* cpu{};
std::string name{};
bool arm9{};
Slice<uint8_t> itcm{};
uint32_t itcmBase{};
Slice<uint8_t> dtcm{};
uint32_t dtcmBase{};
Slice<uint8_t> low{};
Slice<uint8_t> wram7{};
std::array<dsmachine_dmaChan,4> dma{};
std::array<dsmachine_timer,4> timers{};
bool ime{};
uint32_t ie{};
uint32_t if_{};
bool waiting{};
uint32_t waitMask{};
bool waitAny{};
uint32_t handlerBase{};
Map<uint32_t,uint32_t> io{};
uint32_t lastRecv{};
int64_t sleep{};
bool wfi{};
};
struct dsmachine_Profile{
double TotalMs{};
double GeometryMs{};
double RasterMs{};
double ComposeMs{};
double DMAMs{};
double CPUMs{};
int64_t Commands{};
int64_t Polygons{};
int64_t Fragments{};
int64_t DMAXfers{};
uint64_t Frames{};
};
struct dsmachine_Result{
uint64_t Steps{};
uint64_t Frames{};
std::string Reason{};
Map<uint32_t,uint64_t> ARM9Milest{};
};
struct dsmachine_spibus{
Slice<uint8_t> firmware{};
int64_t dev{};
int64_t phase{};
uint8_t cmd{};
uint32_t addr{};
uint8_t out{};
int64_t chanSel{};
int64_t resultIdx{};
int64_t touchX{};
int64_t touchY{};
bool touchDown{};
};
struct dsmachine_vram{
std::array<Slice<uint8_t>,9> bank{};
std::array<uint8_t,9> cnt{};
std::array<Slice<Slice<Slice<uint8_t>>>,11> pages{};
};
struct Anon0{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};arm_Flow Flow{};uint32_t Target{};bool HasTarget{};bool Thumb{};bool TargetThumb{};int64_t Cond{};};
struct Anon1{std::array<uint32_t,16> R{};bool N{};bool Z{};bool C{};bool V{};bool Q{};uint32_t GE{};bool Thumb{};bool BigEndian{};bool IRQDisable{};bool FIQDisable{};uint32_t Mode{};arm_Variant Arch{};arm_vfpState VFP{};bool exclValid{};uint32_t exclAddr{};std::array<uint32_t,6> bankR13{};std::array<uint32_t,6> bankR14{};std::array<uint32_t,6> bankSPSR{};std::array<uint32_t,5> fiqR8_12{};std::array<uint32_t,5> usrR8_12{};std::function<bool(arm_CPU*,uint32_t)> SWI{};std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};dsmachine_bus* bus{};dsmachine_bus* wide{};bool Halted{};std::string HaltReason{};uint64_t Instrs{};uint32_t cur{};bool branched{};};
struct Anon10{dsmachine_core* c{};};
struct Anon11{Slice<uint8_t> rom{};std::array<uint8_t,8> cmd{};uint32_t ctrl{};Slice<uint8_t> buf{};int64_t pos{};bool owner9{};std::function<void(std::array<uint8_t,8>,uint32_t,int64_t)> OnXfer{};};
struct Anon12{bool Drawn{};bool ZReject{};bool AlphaReject{};uint8_t R{};uint8_t G{};uint8_t B{};uint8_t A{};};
struct Anon13{};
struct Anon14{std::string Name{};uint32_t Base{};uint32_t Size{};};
struct Anon15{uint32_t src{};uint32_t dst{};uint32_t ctrl{};uint32_t count{};uint32_t csrc{};uint32_t cdst{};uint32_t crem{};bool active{};};
struct Anon16{dsmachine_engine a{};dsmachine_engine b{};Slice<uint32_t> threeD{};bool swap{};};
struct Anon17{dsmachine_Machine* m{};bool isB{};uint32_t ioBase{};uint32_t palBG{};uint32_t palOBJ{};uint32_t oamBase{};int64_t spBG{};int64_t spOBJ{};int64_t spBGExt{};int64_t spOBJExt{};Slice<uint32_t> out{};Slice<uint32_t> threeD{};std::array<std::array<uint16_t,256>,4> bg{};std::array<std::array<bool,256>,4> bgOK{};std::array<uint8_t,256> a3D{};bool is3D{};std::array<uint8_t,4> prio{};std::array<bool,4> shown{};std::array<uint16_t,256> obj{};std::array<bool,256> objOK{};std::array<uint8_t,256> objPrio{};std::array<bool,256> objSemi{};std::array<uint8_t,256> objEVA{};std::array<bool,256> objWin{};std::array<uint8_t,256> win{};std::array<uint16_t,256> line{};std::array<bool,49152> vis3D{};};
struct Anon18{uint16_t c{};int64_t layer{};bool semi{};uint8_t eva{};uint8_t a3d{};};
struct Anon19{Slice<uint32_t> fifo{};Slice<uint8_t> packed{};uint8_t cmd{};Slice<uint32_t> params{};int64_t need{};dsmachine_geom geom{};dsmachine_raster rast{};Map<uint32_t,uint32_t> regs{};bool swapPending{};uint32_t swapMode{};int64_t lastPolys{};int64_t swaps{};std::array<int64_t,256> cmdHist{};std::function<void(uint8_t,Slice<uint32_t>)> OnCmd{};int64_t count{};int64_t limit{};int64_t cur{};};
struct Anon2{std::array<uint32_t,6> R13{};std::array<uint32_t,6> R14{};std::array<uint32_t,6> SPSR{};std::array<uint32_t,5> FIQR8_12{};std::array<uint32_t,5> USRR8_12{};};
struct Anon20{std::array<int32_t,16> m{};};
struct Anon21{int64_t x{};int64_t y{};int64_t z{};int64_t w{};int32_t r{};int32_t g{};int32_t b{};int32_t s{};int32_t t{};};
struct Anon22{Slice<dsmachine_gxVertex> verts{};uint32_t attr{};uint32_t texParam{};uint32_t pltt{};bool wbuffer{};int64_t cmd{};};
struct Anon23{int64_t mode{};dsmachine_mtx proj{};dsmachine_mtx pos{};dsmachine_mtx vec{};dsmachine_mtx tex{};std::array<dsmachine_mtx,1> projStack{};std::array<dsmachine_mtx,1> texStack{};std::array<dsmachine_mtx,32> posStack{};std::array<dsmachine_mtx,32> vecStack{};int64_t sp{};int64_t projSP{};int64_t texSP{};bool clipDirty{};dsmachine_mtx clipMtx{};bool begun{};int64_t primMode{};Slice<dsmachine_gxVertex> strip{};int64_t stripLen{};std::array<int32_t,3> lastVtx{};std::array<int32_t,3> color{};int32_t texS{};int32_t texT{};uint32_t attr{};uint32_t attrNext{};uint32_t texParam{};uint32_t pltt{};std::array<std::array<int32_t,3>,4> lightVec{};std::array<std::array<int32_t,3>,4> lightColor{};std::array<int32_t,3> diffuse{};std::array<int32_t,3> ambient{};std::array<int32_t,3> specular{};std::array<int32_t,3> emission{};std::array<uint8_t,128> shininess{};int32_t viewX1{};int32_t viewY1{};int32_t viewX2{};int32_t viewY2{};Slice<dsmachine_gxVertex> verts{};Slice<dsmachine_gxPolygon> polys{};std::array<int32_t,4> posResult{};std::array<int32_t,3> vecResult{};bool wbuffer{};int64_t nEmit{};int64_t nClipped{};};
struct Anon24{uint8_t r{};uint8_t g{};uint8_t b{};uint8_t a{};};
struct Anon25{std::array<dsmachine_rfrag,49152> col{};std::array<uint32_t,49152> depth{};Slice<uint32_t> frame{};};
struct Anon26{double x{};double y{};double depth{};double iw{};double w{};double r{};double g{};double b{};double s{};double t{};};
struct Anon27{dsmachine_texState tex{};bool hasTex{};uint32_t mode{};uint8_t alpha{};bool trans{};bool depthEq{};bool depthWr{};bool wbuf{};bool blend{};bool toonHi{};bool alphaTe{};uint8_t alphaRf{};std::array<std::array<uint8_t,3>,32> toon{};bool wireframe{};};
struct Anon28{uint32_t base{};uint32_t pal{};int64_t sizeS{};int64_t sizeT{};uint32_t format{};bool repeatS{};bool repeatT{};bool flipS{};bool flipT{};bool color0{};};
struct Anon29{Slice<uint8_t> ram{};Slice<uint8_t> swram{};Slice<uint8_t> pal{};Slice<uint8_t> oam{};dsmachine_vram* vram{};dsmachine_card* cd{};dsmachine_spibus* spi{};dsmachine_gpu2d* gpu2d{};dsmachine_gpu3d* gpu3d{};dsmachine_video vid{};dsmachine_divider div{};dsmachine_sqrter sqrt{};dsmachine_profiler prof{};uint32_t powcnt{};uint32_t keys{};uint8_t wramcnt{};dsmachine_ipc ipc{};dsmachine_core* ARM9{};dsmachine_core* ARM7{};uint64_t Steps{};Slice<std::string> Log{};Map<std::string,bool> logSeen{};Map<uint32_t,bool> visited{};std::function<void(std::string,uint8_t,uint32_t)> SyncTrace{};std::function<void(bool,uint32_t)> OnStep{};std::function<void(bool,bool,uint32_t,uint32_t,uint32_t)> OnIO{};std::function<void(bool,uint32_t,uint8_t,uint32_t)> OnWrite{};std::function<void(bool,uint32_t,uint32_t,uint32_t)> OnIRQ{};std::function<void(bool,uint32_t,uint8_t,uint32_t)> OnRead{};std::function<void(int64_t,int64_t,dsmachine_PixelEvent)> OnPixel{};std::function<void(int64_t)> OnPoly{};Map<uint32_t,bool> bps{};bool stop{};bool stopped{};uint32_t stoppedPC{};std::function<void()> OnFrame{};};
struct Anon3{uint32_t bit{};std::string name{};};
struct Anon30{dsmachine_Machine* m{};arm_CPU* cpu{};std::string name{};bool arm9{};Slice<uint8_t> itcm{};uint32_t itcmBase{};Slice<uint8_t> dtcm{};uint32_t dtcmBase{};Slice<uint8_t> low{};Slice<uint8_t> wram7{};std::array<dsmachine_dmaChan,4> dma{};std::array<dsmachine_timer,4> timers{};bool ime{};uint32_t ie{};uint32_t if_{};bool waiting{};uint32_t waitMask{};bool waitAny{};uint32_t handlerBase{};Map<uint32_t,uint32_t> io{};uint32_t lastRecv{};int64_t sleep{};bool wfi{};};
struct Anon31{uint8_t sync9{};uint8_t sync7{};Slice<uint32_t> to7{};Slice<uint32_t> to9{};};
struct Anon32{uint32_t cnt{};uint64_t numer{};uint64_t denom{};uint64_t result{};uint64_t rem{};};
struct Anon33{uint32_t cnt{};uint64_t param{};uint32_t result{};};
struct Anon34{double TotalMs{};double GeometryMs{};double RasterMs{};double ComposeMs{};double DMAMs{};double CPUMs{};int64_t Commands{};int64_t Polygons{};int64_t Fragments{};int64_t DMAXfers{};uint64_t Frames{};};
struct Anon35{bool on{};time_Time frameStart{};time_Duration geom{};time_Duration raster{};time_Duration compose{};time_Duration dma{};int64_t polys{};int64_t frags{};int64_t xfers{};int64_t cmdStart{};};
struct Anon36{uint64_t Steps{};uint64_t Frames{};std::string Reason{};Map<uint32_t,uint64_t> ARM9Milest{};};
struct Anon37{std::string name{};image_RGBA* img{};};
struct Anon38{Slice<uint8_t> firmware{};int64_t dev{};int64_t phase{};uint8_t cmd{};uint32_t addr{};uint8_t out{};int64_t chanSel{};int64_t resultIdx{};int64_t touchX{};int64_t touchY{};bool touchDown{};};
struct Anon39{uint16_t counter{};uint16_t reload{};uint16_t ctrl{};int64_t frac{};};
struct Anon4{std::array<uint32_t,32> S{};uint32_t FPSCR{};uint32_t FPEXC{};};
struct Anon40{int64_t line{};bool hblank{};uint64_t frames{};};
struct Anon41{std::array<Slice<uint8_t>,9> bank{};std::array<uint8_t,9> cnt{};std::array<Slice<Slice<Slice<uint8_t>>>,11> pages{};};
struct Anon5{std::string Title{};std::string GameCode{};std::string MakerCode{};uint8_t UnitCode{};uint8_t SeedSelect{};uint8_t DeviceCap{};uint8_t ROMVersion{};uint8_t AutoStart{};uint32_t ARM9ROMOff{};uint32_t ARM9Entry{};uint32_t ARM9RAMAddr{};uint32_t ARM9Size{};uint32_t ARM7ROMOff{};uint32_t ARM7Entry{};uint32_t ARM7RAMAddr{};uint32_t ARM7Size{};uint32_t FNTOff{};uint32_t FNTSize{};uint32_t FATOff{};uint32_t FATSize{};uint32_t ARM9OverlayOff{};uint32_t ARM9OverlaySize{};uint32_t ARM7OverlayOff{};uint32_t ARM7OverlaySize{};uint32_t IconOff{};uint16_t SecureCRC{};uint32_t ARM9HookRAM{};uint32_t ARM7HookRAM{};uint32_t TotalUsedSize{};uint32_t HeaderSize{};uint16_t LogoCRC{};uint16_t HeaderCRC{};};
struct Anon6{uint32_t Start{};uint32_t End{};};
struct Anon7{Slice<uint8_t> Data{};nds_Header Header{};Slice<nds_FATEntry> FAT{};Slice<nds_FileInfo> Files{};Map<std::string,int64_t> byPath{};};
struct Anon8{int64_t ID{};std::string Path{};};
struct Anon9{uint32_t ID{};uint32_t RAMAddr{};uint32_t RAMSize{};uint32_t BSSSize{};uint32_t StaticInitStart{};uint32_t StaticInitEnd{};uint32_t FileID{};uint32_t CompressedSize{};bool Compressed{};};

#include "adapters-decl.h"
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
constexpr int64_t dsmachine_bootHeaderCopy=41942528ULL;
constexpr int64_t dsmachine_bootUserSettings=41942144ULL;
constexpr int64_t dsmachine_bootChipID=41940992ULL;
constexpr int64_t dsmachine_bootChipIDAlt=41942016ULL;
constexpr int64_t dsmachine_regAUXSPICNT=67109280ULL;
constexpr int64_t dsmachine_regAUXSPIDATA=67109282ULL;
constexpr int64_t dsmachine_regROMCTRL=67109284ULL;
constexpr int64_t dsmachine_regCARDCMD=67109288ULL;
constexpr int64_t dsmachine_regCARDDATA=68157456ULL;
constexpr int64_t dsmachine_regEXMEMCNT=67109380ULL;
constexpr int64_t dsmachine_romctrlWordReady=8388608ULL;
constexpr int64_t dsmachine_romctrlBusy=2147483648ULL;
Slice<std::string> dsmachine_textureFormats=Slice<std::string>{std::string("none",4),std::string("a3i5",4),std::string("4-colour",8),std::string("16-colour",9),std::string("256-colour",10),std::string("4x4-compressed",14),std::string("a5i3",4),std::string("direct16",8)};
constexpr int64_t dsmachine_dmaImmediate=0ULL;
constexpr int64_t dsmachine_dmaVBlank=1ULL;
constexpr int64_t dsmachine_dmaHBlank=2ULL;
constexpr int64_t dsmachine_dmaDisplaySync=3ULL;
constexpr int64_t dsmachine_dmaMainMemDisplay=4ULL;
constexpr int64_t dsmachine_dmaCard=5ULL;
constexpr int64_t dsmachine_dmaGBACart=6ULL;
constexpr int64_t dsmachine_dmaGXFIFO=7ULL;
constexpr int64_t dsmachine_screenW=256ULL;
constexpr int64_t dsmachine_screenH=192ULL;
constexpr int64_t dsmachine_rDISPCNT=0ULL;
constexpr int64_t dsmachine_rBG0CNT=8ULL;
constexpr int64_t dsmachine_rBG0HOFS=16ULL;
constexpr int64_t dsmachine_rBG2PA=32ULL;
constexpr int64_t dsmachine_rWIN0H=64ULL;
constexpr int64_t dsmachine_rWIN0V=68ULL;
constexpr int64_t dsmachine_rWININ=72ULL;
constexpr int64_t dsmachine_rWINOUT=74ULL;
constexpr int64_t dsmachine_rMOSAIC=76ULL;
constexpr int64_t dsmachine_rBLDCNT=80ULL;
constexpr int64_t dsmachine_rBLDALPHA=82ULL;
constexpr int64_t dsmachine_rBLDY=84ULL;
constexpr int64_t dsmachine_rDISPCAP=100ULL;
constexpr int64_t dsmachine_rMASTERBR=108ULL;
constexpr int64_t dsmachine_lyBG0=0ULL;
constexpr int64_t dsmachine_lyBG1=1ULL;
constexpr int64_t dsmachine_lyBG2=2ULL;
constexpr int64_t dsmachine_lyBG3=3ULL;
constexpr int64_t dsmachine_lyOBJ=4ULL;
constexpr int64_t dsmachine_lyBackdrop=5ULL;
constexpr int64_t dsmachine_kNone=0ULL;
constexpr int64_t dsmachine_kText=1ULL;
constexpr int64_t dsmachine_kAffine=2ULL;
constexpr int64_t dsmachine_kExtended=3ULL;
constexpr int64_t dsmachine_kLarge=4ULL;
std::array<std::array<int64_t,4>,8> dsmachine_bgKind=[](){std::array<std::array<int64_t,4>,8> v{};v[0]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL)};v[1]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL)};v[2]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<int64_t>(2ULL)};v[3]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(3ULL)};v[4]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<int64_t>(3ULL)};v[5]=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(3ULL),cast<int64_t>(3ULL)};v[6]=std::array<int64_t,4>{cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(4ULL),cast<int64_t>(0ULL)};v[7]=std::array<int64_t,4>{cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL)};return v;}();
std::array<std::array<int64_t,2>,4> dsmachine_extBmpSize=std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(128ULL),cast<int64_t>(128ULL)},std::array<int64_t,2>{cast<int64_t>(256ULL),cast<int64_t>(256ULL)},std::array<int64_t,2>{cast<int64_t>(512ULL),cast<int64_t>(256ULL)},std::array<int64_t,2>{cast<int64_t>(512ULL),cast<int64_t>(512ULL)}};
std::array<std::array<std::array<int64_t,2>,4>,3> dsmachine_objSize=std::array<std::array<std::array<int64_t,2>,4>,3>{std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(64ULL),cast<int64_t>(64ULL)}},std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(64ULL),cast<int64_t>(32ULL)}},std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(64ULL)}}};
constexpr int64_t dsmachine_regGXFIFO=67109888ULL;
constexpr int64_t dsmachine_regGXCMDPORT=67109952ULL;
constexpr int64_t dsmachine_regGXCMDPORTEnd=67110344ULL;
constexpr int64_t dsmachine_regGXSTAT=67110400ULL;
constexpr int64_t dsmachine_regRAMCOUNT=67110404ULL;
constexpr int64_t dsmachine_regDISP3DCNT=67108960ULL;
constexpr int64_t dsmachine_regCLEARCOLOR=67109712ULL;
constexpr int64_t dsmachine_regCLEARDEPTH=67109716ULL;
constexpr int64_t dsmachine_regCLIPMTXRESULT=67110464ULL;
constexpr int64_t dsmachine_regVECMTXRESULT=67110528ULL;
constexpr int64_t dsmachine_regPOSRESULT=67110432ULL;
constexpr int64_t dsmachine_regVECRESULT=67110448ULL;
Map<uint8_t,int64_t> dsmachine_gxParams=Map<uint8_t,int64_t>{{cast<uint8_t>(16ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(17ULL),cast<int64_t>(0ULL)},{cast<uint8_t>(18ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(19ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(20ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(21ULL),cast<int64_t>(0ULL)},{cast<uint8_t>(22ULL),cast<int64_t>(16ULL)},{cast<uint8_t>(23ULL),cast<int64_t>(12ULL)},{cast<uint8_t>(24ULL),cast<int64_t>(16ULL)},{cast<uint8_t>(25ULL),cast<int64_t>(12ULL)},{cast<uint8_t>(26ULL),cast<int64_t>(9ULL)},{cast<uint8_t>(27ULL),cast<int64_t>(3ULL)},{cast<uint8_t>(28ULL),cast<int64_t>(3ULL)},{cast<uint8_t>(32ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(33ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(34ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(35ULL),cast<int64_t>(2ULL)},{cast<uint8_t>(36ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(37ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(38ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(39ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(40ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(41ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(42ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(43ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(48ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(49ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(50ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(51ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(52ULL),cast<int64_t>(32ULL)},{cast<uint8_t>(64ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(65ULL),cast<int64_t>(0ULL)},{cast<uint8_t>(80ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(96ULL),cast<int64_t>(1ULL)},{cast<uint8_t>(112ULL),cast<int64_t>(3ULL)},{cast<uint8_t>(113ULL),cast<int64_t>(2ULL)},{cast<uint8_t>(114ULL),cast<int64_t>(1ULL)}};
constexpr int64_t dsmachine_mtxProjection=0ULL;
constexpr int64_t dsmachine_mtxPosition=1ULL;
constexpr int64_t dsmachine_mtxPosVec=2ULL;
constexpr int64_t dsmachine_mtxTexture=3ULL;
constexpr int64_t dsmachine_rastW=256ULL;
constexpr int64_t dsmachine_rastH=192ULL;
constexpr int64_t dsmachine_regKEYINPUT=67109168ULL;
constexpr int64_t dsmachine_regKEYCNT=67109170ULL;
constexpr int64_t dsmachine_regRCNT=67109172ULL;
constexpr int64_t dsmachine_regEXTKEYIN=67109174ULL;
constexpr int64_t dsmachine_KeyA=1ULL;
constexpr int64_t dsmachine_KeyB=2ULL;
constexpr int64_t dsmachine_KeySelect=4ULL;
constexpr int64_t dsmachine_KeyStart=8ULL;
constexpr int64_t dsmachine_KeyRight=16ULL;
constexpr int64_t dsmachine_KeyLeft=32ULL;
constexpr int64_t dsmachine_KeyUp=64ULL;
constexpr int64_t dsmachine_KeyDown=128ULL;
constexpr int64_t dsmachine_KeyR=256ULL;
constexpr int64_t dsmachine_KeyL=512ULL;
constexpr int64_t dsmachine_KeyX=1024ULL;
constexpr int64_t dsmachine_KeyY=2048ULL;
Map<std::string,uint32_t> dsmachine_keyNames=Map<std::string,uint32_t>{{std::string("a",1),cast<uint32_t>(1ULL)},{std::string("b",1),cast<uint32_t>(2ULL)},{std::string("x",1),cast<uint32_t>(1024ULL)},{std::string("y",1),cast<uint32_t>(2048ULL)},{std::string("l",1),cast<uint32_t>(512ULL)},{std::string("r",1),cast<uint32_t>(256ULL)},{std::string("start",5),cast<uint32_t>(8ULL)},{std::string("select",6),cast<uint32_t>(4ULL)},{std::string("up",2),cast<uint32_t>(64ULL)},{std::string("down",4),cast<uint32_t>(128ULL)},{std::string("left",4),cast<uint32_t>(32ULL)},{std::string("right",5),cast<uint32_t>(16ULL)}};
constexpr int64_t dsmachine_regDISPSTAT=67108868ULL;
constexpr int64_t dsmachine_regVCOUNT=67108870ULL;
constexpr int64_t dsmachine_regDMA0SAD=67109040ULL;
constexpr int64_t dsmachine_regDMAFILL=67109088ULL;
constexpr int64_t dsmachine_regTM0CNT=67109120ULL;
constexpr int64_t dsmachine_regIPCSYNC=67109248ULL;
constexpr int64_t dsmachine_regIPCFIFOCNT=67109252ULL;
constexpr int64_t dsmachine_regIPCFIFOSND=67109256ULL;
constexpr int64_t dsmachine_regIME=67109384ULL;
constexpr int64_t dsmachine_regIE=67109392ULL;
constexpr int64_t dsmachine_regIF=67109396ULL;
constexpr int64_t dsmachine_regVRAMCNT=67109440ULL;
constexpr int64_t dsmachine_regWRAMCNT=67109447ULL;
constexpr int64_t dsmachine_regPOSTFLG=67109632ULL;
constexpr int64_t dsmachine_regPOWCNT=67109636ULL;
constexpr int64_t dsmachine_regIPCFIFORCV=68157440ULL;
constexpr int64_t dsmachine_regRTC=67109176ULL;
Map<uint32_t,bool> dsmachine_ioKnown=Map<uint32_t,bool>{{cast<uint32_t>(67109168ULL),true},{cast<uint32_t>(67109280ULL),true},{cast<uint32_t>(67109380ULL),true},{cast<uint32_t>(67109632ULL),true},{cast<uint32_t>(67109636ULL),true},{cast<uint32_t>(67108868ULL),true},{cast<uint32_t>(67109176ULL),true},{cast<uint32_t>(67109088ULL),true},{cast<uint32_t>(67109092ULL),true},{cast<uint32_t>(67109096ULL),true},{cast<uint32_t>(67109100ULL),true}};
constexpr int64_t dsmachine_mainBase=33554432ULL;
constexpr int64_t dsmachine_mainSize=4194304ULL;
constexpr int64_t dsmachine_mainEnd=37748736ULL;
constexpr int64_t dsmachine_mainMirrorEnd=50331648ULL;
constexpr int64_t dsmachine_swramBase=50331648ULL;
constexpr int64_t dsmachine_swramEnd=58720256ULL;
constexpr int64_t dsmachine_swramSize=32768ULL;
constexpr int64_t dsmachine_wram7Base=58720256ULL;
constexpr int64_t dsmachine_wram7Size=65536ULL;
constexpr int64_t dsmachine_wram7End=58785792ULL;
constexpr int64_t dsmachine_itcmDefault=33521664ULL;
constexpr int64_t dsmachine_itcmSize=32768ULL;
constexpr int64_t dsmachine_palBase=83886080ULL;
constexpr int64_t dsmachine_palSize=2048ULL;
constexpr int64_t dsmachine_oamBase=117440512ULL;
constexpr int64_t dsmachine_oamSize=2048ULL;
constexpr int64_t dsmachine_regDIVCNT=67109504ULL;
constexpr int64_t dsmachine_regDIV_NUMER=67109520ULL;
constexpr int64_t dsmachine_regDIV_DENOM=67109528ULL;
constexpr int64_t dsmachine_regDIV_RESULT=67109536ULL;
constexpr int64_t dsmachine_regDIVREM=67109544ULL;
constexpr int64_t dsmachine_regSQRTCNT=67109552ULL;
constexpr int64_t dsmachine_regSQRT_RESULT=67109556ULL;
constexpr int64_t dsmachine_regSQRT_PARAM=67109560ULL;
constexpr int64_t dsmachine_biosIRQReturn=4294905856ULL;
constexpr int64_t dsmachine_ScreenW=256ULL;
constexpr int64_t dsmachine_ScreenH=192ULL;
constexpr int64_t dsmachine_regSOUNDxCNT=67109888ULL;
constexpr int64_t dsmachine_regSOUNDCNT=67110144ULL;
constexpr int64_t dsmachine_regSOUNDBIAS=67110148ULL;
constexpr int64_t dsmachine_regSPICNT=67109312ULL;
constexpr int64_t dsmachine_regSPIDATA=67109314ULL;
constexpr int64_t dsmachine_firmwareSize=262144ULL;
constexpr int64_t dsmachine_userSettings1=261632ULL;
constexpr int64_t dsmachine_userSettings2=261888ULL;
constexpr int64_t dsmachine_spiPower=0ULL;
constexpr int64_t dsmachine_spiFirm=1ULL;
constexpr int64_t dsmachine_spiTouch=2ULL;
constexpr int64_t dsmachine_calADCX1=735ULL;
constexpr int64_t dsmachine_calADCY1=812ULL;
constexpr int64_t dsmachine_calScrX1=32ULL;
constexpr int64_t dsmachine_calScrY1=32ULL;
constexpr int64_t dsmachine_calADCX2=3387ULL;
constexpr int64_t dsmachine_calADCY2=3303ULL;
constexpr int64_t dsmachine_calScrX2=224ULL;
constexpr int64_t dsmachine_calScrY2=158ULL;
std::array<int64_t,4> dsmachine_timerPrescale=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(64ULL),cast<int64_t>(256ULL),cast<int64_t>(1024ULL)};
constexpr int64_t dsmachine_dotsPerLine=355ULL;
constexpr int64_t dsmachine_linesPerFrame=263ULL;
constexpr int64_t dsmachine_visibleLines=192ULL;
constexpr int64_t dsmachine_visibleDots=256ULL;
constexpr int64_t dsmachine_cyclesPerLine9=4260ULL;
constexpr int64_t dsmachine_cyclesPerLine7=2130ULL;
constexpr int64_t dsmachine_dispstatVBlankFlag=1ULL;
constexpr int64_t dsmachine_dispstatHBlankFlag=2ULL;
constexpr int64_t dsmachine_dispstatVMatchFlag=4ULL;
constexpr int64_t dsmachine_dispstatVBlankIRQ=8ULL;
constexpr int64_t dsmachine_dispstatHBlankIRQ=16ULL;
constexpr int64_t dsmachine_dispstatVMatchIRQ=32ULL;
constexpr int64_t dsmachine_irqVBlank=1ULL;
constexpr int64_t dsmachine_irqHBlank=2ULL;
constexpr int64_t dsmachine_irqVMatch=4ULL;
constexpr int64_t dsmachine_irqTimer0=8ULL;
constexpr int64_t dsmachine_irqDMA0=256ULL;
constexpr int64_t dsmachine_irqCard=524288ULL;
constexpr int64_t dsmachine_irqIPCSync=65536ULL;
constexpr int64_t dsmachine_irqIPCSend=131072ULL;
constexpr int64_t dsmachine_irqIPCRecv=262144ULL;
constexpr int64_t dsmachine_irqGXFIFO=2097152ULL;
constexpr int64_t dsmachine_irqSPI=8388608ULL;
constexpr int64_t dsmachine_vramPage=8192ULL;
std::array<int64_t,9> dsmachine_bankSizes=std::array<int64_t,9>{cast<int64_t>(131072ULL),cast<int64_t>(131072ULL),cast<int64_t>(131072ULL),cast<int64_t>(131072ULL),cast<int64_t>(65536ULL),cast<int64_t>(16384ULL),cast<int64_t>(16384ULL),cast<int64_t>(32768ULL),cast<int64_t>(16384ULL)};
std::array<uint32_t,9> dsmachine_lcdcBase=std::array<uint32_t,9>{cast<uint32_t>(109051904ULL),cast<uint32_t>(109182976ULL),cast<uint32_t>(109314048ULL),cast<uint32_t>(109445120ULL),cast<uint32_t>(109576192ULL),cast<uint32_t>(109641728ULL),cast<uint32_t>(109658112ULL),cast<uint32_t>(109674496ULL),cast<uint32_t>(109707264ULL)};
constexpr int64_t dsmachine_spBGA=0ULL;
constexpr int64_t dsmachine_spBGB=1ULL;
constexpr int64_t dsmachine_spOBJA=2ULL;
constexpr int64_t dsmachine_spOBJB=3ULL;
constexpr int64_t dsmachine_spTex=4ULL;
constexpr int64_t dsmachine_spTexPal=5ULL;
constexpr int64_t dsmachine_spBGExtA=6ULL;
constexpr int64_t dsmachine_spBGExtB=7ULL;
constexpr int64_t dsmachine_spOBJExtA=8ULL;
constexpr int64_t dsmachine_spOBJExtB=9ULL;
constexpr int64_t dsmachine_spARM7=10ULL;
constexpr int64_t dsmachine_numSpaces=11ULL;
std::array<int64_t,11> dsmachine_spaceSize=std::array<int64_t,11>{cast<int64_t>(524288ULL),cast<int64_t>(131072ULL),cast<int64_t>(262144ULL),cast<int64_t>(131072ULL),cast<int64_t>(524288ULL),cast<int64_t>(131072ULL),cast<int64_t>(32768ULL),cast<int64_t>(32768ULL),cast<int64_t>(8192ULL),cast<int64_t>(8192ULL),cast<int64_t>(262144ULL)};
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
bool nds_IsBLZ(Slice<uint8_t> data);
Slice<uint8_t> nds_DecompressBLZ(Slice<uint8_t> data);
void nds_reverse(Slice<uint8_t> b);
void nds_reverseRange(Slice<uint8_t> b,int64_t lo,int64_t hi);
uint16_t nds_CRC16(Slice<uint8_t> data);
std::tuple<uint16_t,bool> nds_ROM_VerifyHeaderCRC(nds_ROM* r);
bool nds_IsLZ77(Slice<uint8_t> data);
std::tuple<Slice<uint8_t>,Error> nds_DecompressLZ77(Slice<uint8_t> data);
Slice<uint8_t> nds_Decompress(Slice<uint8_t> data);
int64_t nds_Header_ChipBytes(nds_Header h);
std::tuple<nds_Header,Error> nds_ParseHeader(Slice<uint8_t> data);
uint32_t nds_FATEntry_Size(nds_FATEntry e);
std::tuple<nds_ROM*,Error> nds_Open(Slice<uint8_t> data);
Error nds_ROM_parseFAT(nds_ROM* r);
Error nds_ROM_parseFNT(nds_ROM* r);
Slice<uint8_t> nds_ROM_File(nds_ROM* r,int64_t id);
Slice<uint8_t> nds_ROM_FileByPath(nds_ROM* r,std::string path);
Slice<nds_Overlay> nds_ROM_ARM9Overlays(nds_ROM* r);
Slice<nds_Overlay> nds_ROM_overlays(nds_ROM* r,uint32_t off,uint32_t size);
Slice<uint8_t> nds_ROM_ARM9(nds_ROM* r);
Slice<uint8_t> nds_ROM_ARM7(nds_ROM* r);
Slice<uint8_t> nds_ROM_slice(nds_ROM* r,uint32_t off,uint32_t size);
std::function<bool(arm_CPU*,uint32_t)> dsmachine_biosSWI(dsmachine_core* c);
void dsmachine_core_park(dsmachine_core* c,arm_CPU* cpu,uint32_t mask,bool any);
void dsmachine_cpuSet(dsmachine_bus* b,arm_CPU* c,bool fast);
void dsmachine_lz77UnComp(dsmachine_bus* b,uint32_t src,uint32_t dst);
void dsmachine_rlUnComp(dsmachine_bus* b,uint32_t src,uint32_t dst);
void dsmachine_diffUnFilter(dsmachine_bus* b,uint32_t src,uint32_t dst,bool wide);
uint32_t dsmachine_isqrt(uint64_t v);
std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> dsmachine_cp15(dsmachine_core* c);
void dsmachine_Machine_directBoot(dsmachine_Machine* m,nds_ROM* rom);
Slice<uint8_t> dsmachine_spibus_currentUserSettings(dsmachine_spibus* s);
std::tuple<Slice<uint8_t>,uint32_t> dsmachine_bus_slot(dsmachine_bus* b,uint32_t a);
std::tuple<Slice<uint8_t>,uint32_t> dsmachine_core_wramSlot(dsmachine_core* c,uint32_t a);
uint8_t dsmachine_bus_Read(dsmachine_bus* b,uint32_t a);
uint8_t dsmachine_bus_read(dsmachine_bus* b,uint32_t a);
void dsmachine_bus_Write(dsmachine_bus* b,uint32_t a,uint8_t v);
uint8_t dsmachine_core_vramRead(dsmachine_core* c,uint32_t a);
void dsmachine_core_vramWrite(dsmachine_core* c,uint32_t a,uint8_t b);
uint16_t dsmachine_bus_Read16(dsmachine_bus* b,uint32_t a);
uint32_t dsmachine_bus_Read32(dsmachine_bus* b,uint32_t a);
void dsmachine_bus_watchRead(dsmachine_bus* b,uint32_t a,uint8_t v);
void dsmachine_bus_Write16(dsmachine_bus* b,uint32_t a,uint16_t v);
void dsmachine_bus_Write32(dsmachine_bus* b,uint32_t a,uint32_t v);
uint32_t dsmachine_bus_r32(dsmachine_bus* b,uint32_t a);
void dsmachine_bus_w32(dsmachine_bus* b,uint32_t a,uint32_t v);
int64_t dsmachine_blockBytes(uint32_t ctrl);
void dsmachine_card_start(dsmachine_card* cd,uint32_t ctrl);
uint8_t dsmachine_card_romByte(dsmachine_card* cd,uint32_t a);
uint32_t dsmachine_core_cardReadData(dsmachine_core* c);
dsmachine_core* dsmachine_Machine_cardCore(dsmachine_Machine* m);
int64_t dsmachine_Machine_GXCommandsRun(dsmachine_Machine* m);
void dsmachine_Machine_ForceRender(dsmachine_Machine* m);
void dsmachine_Machine_StopRequested(dsmachine_Machine* m);
void dsmachine_Machine_AddBreakpoint(dsmachine_Machine* m,uint32_t pc);
void dsmachine_Machine_ClearBreakpoint(dsmachine_Machine* m,uint32_t pc);
void dsmachine_Machine_ClearBreakpoints(dsmachine_Machine* m);
std::tuple<bool,uint32_t> dsmachine_Machine_Stopped(dsmachine_Machine* m);
int64_t dsmachine_Machine_StepInstructions(dsmachine_Machine* m,int64_t n);
Slice<dsmachine_MemRegion> dsmachine_Machine_MemRegions(dsmachine_Machine* m);
Slice<std::string> dsmachine_TextureFormats();
std::tuple<uint32_t,bool> dsmachine_TextureFormat(std::string name);
std::tuple<image_RGBA*,Error> dsmachine_Machine_RenderTexture(dsmachine_Machine* m,uint32_t offset,uint32_t palBase,uint32_t format,uint32_t w,uint32_t h);
image_RGBA* dsmachine_Machine_RenderDepth(dsmachine_Machine* m);
Slice<uint8_t> dsmachine_Machine_VRAMBank(dsmachine_Machine* m,int64_t i);
std::tuple<std::string,int64_t,uint8_t> dsmachine_Machine_VRAMBankInfo(dsmachine_Machine* m,int64_t i);
uint8_t dsmachine_expand6(uint8_t v);
uint8_t dsmachine_expand5(uint8_t v);
Slice<bool> dsmachine_Machine_ThreeDVisible(dsmachine_Machine* m);
bool dsmachine_Machine_EngineAOnTop(dsmachine_Machine* m);
Error dsmachine_Machine_Draw3DInto(dsmachine_Machine* m,image_RGBA* img);
int64_t dsmachine_core_dmaMode(dsmachine_core* c,dsmachine_dmaChan* ch);
uint32_t dsmachine_core_maxCount(dsmachine_core* c);
void dsmachine_core_writeDMACtrl(dsmachine_core* c,int64_t n);
void dsmachine_Machine_runDMA(dsmachine_Machine* m,int64_t mode);
void dsmachine_core_runDMAChan(dsmachine_core* c,int64_t n);
void dsmachine_Machine_gxfifoDrained(dsmachine_Machine* m);
dsmachine_gpu2d* dsmachine_newGPU2D();
void dsmachine_gpu2d_beginFrame(dsmachine_gpu2d* g,dsmachine_Machine* m);
void dsmachine_gpu2d_render(dsmachine_gpu2d* g,dsmachine_Machine* m);
std::tuple<Slice<uint32_t>,Slice<uint32_t>> dsmachine_gpu2d_screens(dsmachine_gpu2d* g);
uint32_t dsmachine_engine_reg32(dsmachine_engine* e,uint32_t off);
uint16_t dsmachine_engine_reg16(dsmachine_engine* e,uint32_t off);
uint32_t dsmachine_bgr555RGBA(uint16_t c);
uint16_t dsmachine_rgb555(uint32_t r,uint32_t g,uint32_t b);
uint32_t dsmachine_chan5(uint16_t c,uint64_t n);
uint16_t dsmachine_engine_palBGColor(dsmachine_engine* e,int64_t i);
uint16_t dsmachine_engine_palOBJColor(dsmachine_engine* e,int64_t i);
uint16_t dsmachine_palAt(Slice<uint8_t> pal,uint32_t off);
uint16_t dsmachine_engine_bgExtColor(dsmachine_engine* e,int64_t slot,int64_t pal,int64_t idx);
uint16_t dsmachine_engine_objExtColor(dsmachine_engine* e,int64_t pal,int64_t idx);
void dsmachine_engine_frame(dsmachine_engine* e,dsmachine_Machine* m);
std::string dsmachine_engine_name(dsmachine_engine* e);
void dsmachine_engine_fillWhite(dsmachine_engine* e);
void dsmachine_engine_fillBlack(dsmachine_engine* e);
void dsmachine_engine_vramDisplay(dsmachine_engine* e,uint32_t dispcnt);
void dsmachine_engine_graphics(dsmachine_engine* e,uint32_t dispcnt);
void dsmachine_engine_clearLayers(dsmachine_engine* e);
void dsmachine_engine_emit(dsmachine_engine* e,int64_t y);
void dsmachine_engine_masterBright(dsmachine_engine* e);
void dsmachine_engine_background(dsmachine_engine* e,uint32_t dispcnt,int64_t mode,int64_t n,int64_t y);
void dsmachine_engine_threeDLine(dsmachine_engine* e,int64_t y);
std::tuple<int64_t,int64_t> dsmachine_engine_mosaic(dsmachine_engine* e);
std::tuple<uint32_t,uint32_t> dsmachine_engine_bgBases(dsmachine_engine* e,uint32_t dispcnt,uint16_t cnt);
void dsmachine_engine_textBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y);
std::tuple<int32_t,int32_t,int32_t,int32_t,int32_t,int32_t> dsmachine_engine_affineParams(dsmachine_engine* e,int64_t n);
std::tuple<int32_t,int32_t> dsmachine_affineOrigin(int32_t pb,int32_t pd,int32_t x0,int32_t y0,int64_t y);
void dsmachine_engine_affineBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y);
std::tuple<int64_t,int64_t,bool> dsmachine_wrapAffine(int64_t x,int64_t y,int64_t w,int64_t h,bool wrap);
void dsmachine_engine_extendedBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y);
void dsmachine_engine_extTiledBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y,uint16_t cnt);
void dsmachine_engine_extBitmapBG(dsmachine_engine* e,int64_t n,int64_t y,uint16_t cnt);
void dsmachine_engine_largeBG(dsmachine_engine* e,int64_t n,int64_t y);
void dsmachine_engine_sprites(dsmachine_engine* e,uint32_t dispcnt,int64_t y);
uint16_t dsmachine_engine_oam16(dsmachine_engine* e,uint32_t off);
std::tuple<int32_t,int32_t,int32_t,int32_t> dsmachine_engine_oamAffine(dsmachine_engine* e,int64_t i);
std::tuple<uint16_t,bool> dsmachine_engine_objTilePixel(dsmachine_engine* e,uint32_t dispcnt,uint32_t tile,bool bpp8,int64_t pal,int64_t w,int64_t px,int64_t py);
std::tuple<uint16_t,bool> dsmachine_engine_objBitmapPixel(dsmachine_engine* e,uint32_t dispcnt,uint32_t tile,int64_t w,int64_t px,int64_t py);
void dsmachine_engine_windowMask(dsmachine_engine* e,uint32_t dispcnt,int64_t y);
bool dsmachine_engine_inWindowY(dsmachine_engine* e,int64_t n,int64_t y);
bool dsmachine_engine_inWindowX(dsmachine_engine* e,int64_t n,int64_t x);
void dsmachine_engine_composite(dsmachine_engine* e,int64_t y);
int64_t dsmachine_clamp16(int64_t v);
uint16_t dsmachine_blend16(uint16_t a,uint16_t b,int64_t eva,int64_t evb);
uint16_t dsmachine_blend32(uint16_t a,uint16_t b,int64_t ca,int64_t cb);
uint16_t dsmachine_blendN(uint16_t a,uint16_t b,int64_t ca,int64_t cb,int64_t div);
uint16_t dsmachine_brighten(uint16_t c,int64_t evy);
uint16_t dsmachine_darken(uint16_t c,int64_t evy);
dsmachine_gpu3d* dsmachine_newGPU3D();
bool dsmachine_gpu3d_fifoBelowHalf(dsmachine_gpu3d* g);
void dsmachine_gpu3d_writeReg(dsmachine_gpu3d* g,uint32_t a,uint32_t w);
uint32_t dsmachine_gpu3d_readReg(dsmachine_gpu3d* g,uint32_t a);
uint32_t dsmachine_gpu3d_gxstat(dsmachine_gpu3d* g);
void dsmachine_gpu3d_pushPacked(dsmachine_gpu3d* g,uint32_t w);
void dsmachine_gpu3d_nextPacked(dsmachine_gpu3d* g);
void dsmachine_gpu3d_feed(dsmachine_gpu3d* g,uint32_t w);
void dsmachine_gpu3d_pushDirect(dsmachine_gpu3d* g,uint8_t cmd,uint32_t w);
void dsmachine_gpu3d_vblank(dsmachine_gpu3d* g,dsmachine_Machine* m);
dsmachine_mtx dsmachine_identity();
dsmachine_mtx dsmachine_mtx_mul(dsmachine_mtx a,dsmachine_mtx b);
std::tuple<int64_t,int64_t,int64_t,int64_t> dsmachine_mtx_apply(dsmachine_mtx a,int64_t x,int64_t y,int64_t z,int64_t w);
void dsmachine_geom_reset(dsmachine_geom* g);
void dsmachine_geom_beginFrame(dsmachine_geom* g);
int64_t dsmachine_geom_stackDepth(dsmachine_geom* g);
dsmachine_mtx dsmachine_geom_clip(dsmachine_geom* g);
int32_t dsmachine_geom_vecMtx3x3(dsmachine_geom* g,uint32_t i);
Slice<dsmachine_mtx*> dsmachine_geom_current(dsmachine_geom* g);
void dsmachine_gpu3d_exec(dsmachine_gpu3d* g,uint8_t cmd,Slice<uint32_t> p);
void dsmachine_geom_applyMatrix(dsmachine_geom* g,bool load,dsmachine_mtx n);
int32_t dsmachine_sext(int32_t v,uint64_t n);
std::array<int32_t,3> dsmachine_unpackRGB5(uint32_t v);
void dsmachine_geom_normal(dsmachine_geom* g,uint32_t v);
std::tuple<int32_t,int32_t> dsmachine_geom_texTransformNormal(dsmachine_geom* g,int64_t nx,int64_t ny,int64_t nz);
std::tuple<int32_t,int32_t> dsmachine_geom_texCoordFor(dsmachine_geom* g,int32_t x,int32_t y,int32_t z);
void dsmachine_geom_vertex(dsmachine_geom* g,dsmachine_gpu3d* gp,int32_t x,int32_t y,int32_t z);
void dsmachine_geom_emitIfComplete(dsmachine_geom* g,dsmachine_gpu3d* gp);
void dsmachine_geom_emit(dsmachine_geom* g,dsmachine_gpu3d* gp,Slice<dsmachine_gxVertex> vs);
Slice<dsmachine_gxVertex> dsmachine_clipPolygon(Slice<dsmachine_gxVertex> vs);
int64_t dsmachine_planeDist(dsmachine_gxVertex v,int64_t plane);
dsmachine_gxVertex dsmachine_lerpVertex(dsmachine_gxVertex a,dsmachine_gxVertex b,int64_t da,int64_t db);
void dsmachine_raster_reset(dsmachine_raster* r);
dsmachine_rvert dsmachine_lerpRV(dsmachine_rvert a,dsmachine_rvert b,double u);
void dsmachine_gpu3d_render(dsmachine_gpu3d* g,dsmachine_Machine* m);
void dsmachine_raster_publish(dsmachine_raster* r);
uint8_t dsmachine_c6to8(uint8_t v);
uint8_t dsmachine_a5to8(uint8_t v);
void dsmachine_gpu3d_clear(dsmachine_gpu3d* g,dsmachine_Machine* m,uint32_t disp3d);
void dsmachine_gpu3d_drawPoly(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_gxPolygon* p,uint32_t disp3d);
dsmachine_polyState dsmachine_gpu3d_polyState(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_gxPolygon* p,uint32_t disp3d);
void dsmachine_gpu3d_shade(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_polyState* st,dsmachine_gxPolygon* p,dsmachine_rvert f,int64_t idx);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> dsmachine_polyState_toonShade(dsmachine_polyState* st,uint8_t tr,uint8_t tg,uint8_t tb,uint8_t ta,uint8_t vr);
uint8_t dsmachine_rastClamp63(double v);
double dsmachine_rastCeil(double v);
std::tuple<dsmachine_texState,bool> dsmachine_texStateOf(dsmachine_gxPolygon* p);
int64_t dsmachine_wrapTex(int64_t v,int64_t size,bool repeat,bool flip);
std::tuple<uint8_t,uint8_t,uint8_t> dsmachine_rastBGR555(uint16_t c);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> dsmachine_gpu3d_sampleTex(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_texState* st,int64_t s,int64_t t);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> dsmachine_gpu3d_sampleCompressed(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_texState* st,int64_t s,int64_t t);
std::tuple<uint8_t,uint8_t,uint8_t> dsmachine_rastMix555(uint16_t a,uint16_t b,int64_t wa,int64_t wb,int64_t shift);
std::tuple<uint32_t,bool> dsmachine_ParseKeys(std::string s);
void dsmachine_Machine_SetKeys(dsmachine_Machine* m,uint32_t mask);
uint32_t dsmachine_Machine_keyinput(dsmachine_Machine* m);
uint32_t dsmachine_Machine_extkeyin(dsmachine_Machine* m);
dsmachine_core* dsmachine_core_other(dsmachine_core* c);
uint32_t dsmachine_core_ioRead(dsmachine_core* c,uint32_t a);
std::tuple<uint32_t,bool> dsmachine_core_ioReadReg(dsmachine_core* c,uint32_t a);
void dsmachine_core_ioWrite(dsmachine_core* c,uint32_t a,uint8_t v);
int64_t dsmachine_vramBankIndex(uint32_t a);
uint32_t dsmachine_half64(uint64_t v,uint32_t off);
uint64_t dsmachine_setHalf64(uint64_t v,uint32_t off,uint32_t w);
uint32_t dsmachine_Machine_divRead(dsmachine_Machine* m,uint32_t a);
void dsmachine_Machine_divWrite(dsmachine_Machine* m,uint32_t a,uint32_t w);
uint32_t dsmachine_core_dmaRead(dsmachine_core* c,uint32_t a);
void dsmachine_core_dmaWrite(dsmachine_core* c,uint32_t a,uint32_t w,uint32_t lane);
Slice<uint32_t>* dsmachine_core_sendQ(dsmachine_core* c);
Slice<uint32_t>* dsmachine_core_recvQ(dsmachine_core* c);
void dsmachine_core_fifoSend(dsmachine_core* c,uint32_t v);
uint32_t dsmachine_core_fifoRecv(dsmachine_core* c);
uint16_t dsmachine_core_fifoCnt(dsmachine_core* c);
void dsmachine_core_fifoCntWrite(dsmachine_core* c,uint16_t v);
dsmachine_Machine* dsmachine_New(nds_ROM* rom,uint32_t dtcm9Base);
void dsmachine_copyInto(dsmachine_Machine* m,dsmachine_core* c,uint32_t addr,Slice<uint8_t> data);
uint32_t dsmachine_Machine_ARM9PC(dsmachine_Machine* m);
uint32_t dsmachine_Machine_ARM7PC(dsmachine_Machine* m);
std::tuple<uint8_t,uint8_t> dsmachine_Machine_SyncNibbles(dsmachine_Machine* m);
std::tuple<int64_t,int64_t> dsmachine_Machine_FifoLens(dsmachine_Machine* m);
uint64_t dsmachine_Machine_Frame(dsmachine_Machine* m);
int64_t dsmachine_Machine_Line(dsmachine_Machine* m);
Slice<uint8_t> dsmachine_Machine_Snapshot(dsmachine_Machine* m,bool arm9,uint32_t addr,uint32_t n);
void dsmachine_Machine_Poke(dsmachine_Machine* m,bool arm9,uint32_t addr,Slice<uint8_t> data);
std::array<uint32_t,16> dsmachine_Machine_Regs(dsmachine_Machine* m,bool arm9);
bool dsmachine_Machine_Thumb(dsmachine_Machine* m,bool arm9);
std::tuple<uint32_t,uint32_t,bool> dsmachine_Machine_IRQState(dsmachine_Machine* m,bool arm9);
bool dsmachine_Machine_Parked(dsmachine_Machine* m,bool arm9);
void dsmachine_Machine_onFrame(dsmachine_Machine* m);
bool dsmachine_Machine_IRQDisabled(dsmachine_Machine* m,bool arm9);
uint32_t dsmachine_Machine_Reg(dsmachine_Machine* m,uint32_t a);
int64_t dsmachine_Machine_Sleep(dsmachine_Machine* m,bool arm9);
void dsmachine_Machine_OnCardXfer(dsmachine_Machine* m,std::function<void(std::array<uint8_t,8>,uint32_t,int64_t)> f);
std::tuple<int64_t,int64_t> dsmachine_Machine_GX(dsmachine_Machine* m);
std::array<int64_t,256> dsmachine_Machine_GXHist(dsmachine_Machine* m);
std::tuple<int64_t,int64_t> dsmachine_Machine_GXClip(dsmachine_Machine* m);
uint32_t dsmachine_Machine_Reg7(dsmachine_Machine* m,uint32_t a);
Map<uint32_t,uint32_t> dsmachine_Machine_GXRegs(dsmachine_Machine* m);
void dsmachine_Machine_OnGXCmd(dsmachine_Machine* m,std::function<void(uint8_t,Slice<uint32_t>)> f);
uint16_t dsmachine_Machine_VRAMTexPal(dsmachine_Machine* m,uint32_t off);
void dsmachine_divider_run(dsmachine_divider* d);
void dsmachine_sqrter_run(dsmachine_sqrter* s);
void dsmachine_Machine_SetProfile(dsmachine_Machine* m,bool on);
void dsmachine_profiler_reset(dsmachine_profiler* p,dsmachine_Machine* m);
dsmachine_Profile dsmachine_Machine_FrameProfile(dsmachine_Machine* m);
dsmachine_Result dsmachine_Machine_Run(dsmachine_Machine* m,uint64_t budget,int64_t quantum,Map<uint32_t,std::string> milestones);
dsmachine_Result dsmachine_Machine_RunFrames(dsmachine_Machine* m,uint64_t n,uint64_t budget,int64_t quantum);
dsmachine_Result dsmachine_Machine_run(dsmachine_Machine* m,uint64_t budget,int64_t quantum,Map<uint32_t,std::string> milestones,uint64_t untilFrame);
void dsmachine_Machine_runQuantum(dsmachine_Machine* m,dsmachine_core* c,int64_t n,const Map<uint32_t,std::string>& milestones,const Map<uint32_t,uint64_t>& hit);
void dsmachine_Machine_deliver(dsmachine_Machine* m,dsmachine_core* c);
void dsmachine_core_biosIRQExit(dsmachine_core* c);
uint64_t dsmachine_Machine_progressSig(dsmachine_Machine* m);
std::string dsmachine_parkState(dsmachine_core* c);
int64_t dsmachine_Machine_runInstrs(dsmachine_Machine* m,int64_t n);
std::tuple<image_RGBA*,image_RGBA*> dsmachine_Machine_Screens(dsmachine_Machine* m);
image_RGBA* dsmachine_toImage(Slice<uint32_t> px);
std::tuple<int64_t,int64_t> dsmachine_Machine_EngineStats(dsmachine_Machine* m);
void dsmachine_Machine_soundKeyed(dsmachine_Machine* m);
dsmachine_spibus* dsmachine_newSPI();
void dsmachine_spibus_buildFirmware(dsmachine_spibus* s);
void dsmachine_Machine_SetTouch(dsmachine_Machine* m,int64_t x,int64_t y,bool down);
std::tuple<int64_t,int64_t,bool> dsmachine_Machine_Touch(dsmachine_Machine* m);
void dsmachine_core_spiTransfer(dsmachine_core* c,uint8_t v);
uint8_t dsmachine_spibus_firmwareByte(dsmachine_spibus* s,uint8_t v);
uint8_t dsmachine_spibus_touchByte(dsmachine_spibus* s,uint8_t v);
uint16_t dsmachine_spibus_touchSample(dsmachine_spibus* s,int64_t ch);
bool dsmachine_timer_enabled(dsmachine_timer* t);
bool dsmachine_timer_cascade(dsmachine_timer* t);
bool dsmachine_timer_irqOn(dsmachine_timer* t);
void dsmachine_core_tickTimers(dsmachine_core* c,int64_t cycles);
void dsmachine_core_writeTimerCtrl(dsmachine_core* c,int64_t n,uint16_t v);
int64_t dsmachine_core_vcountMatch(dsmachine_core* c);
uint32_t dsmachine_core_dispstat(dsmachine_core* c);
void dsmachine_Machine_startLine(dsmachine_Machine* m);
void dsmachine_Machine_hblankNow(dsmachine_Machine* m);
std::array<dsmachine_core*,2> dsmachine_Machine_cores(dsmachine_Machine* m);
void dsmachine_core_raise(dsmachine_core* c,uint32_t src);
dsmachine_vram* dsmachine_newVRAM();
void dsmachine_vram_setCNT(dsmachine_vram* v,int64_t bank,uint8_t val);
void dsmachine_vram_remap(dsmachine_vram* v);
void dsmachine_vram_attach(dsmachine_vram* v,int64_t space,int64_t off,Slice<uint8_t> data);
std::tuple<int64_t,int64_t,bool> dsmachine_bankTarget(int64_t b,int64_t mst,int64_t ofs);
uint8_t dsmachine_vram_read8(dsmachine_vram* v,int64_t space,uint32_t off);
uint16_t dsmachine_vram_read16(dsmachine_vram* v,int64_t space,uint32_t off);
void dsmachine_vram_write8(dsmachine_vram* v,int64_t space,uint32_t off,uint8_t b);
std::tuple<Slice<uint8_t>,uint32_t,bool> dsmachine_vram_lcdcSlot(dsmachine_vram* v,uint32_t a);
std::tuple<int64_t,uint32_t,bool> dsmachine_vram_cpuAccess(dsmachine_vram* v,bool arm9,uint32_t a);

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
parts = append(parts,arm_regName[i]);
}
else if ((j == cast<int64_t>((i + cast<int64_t>(1ULL))))){
parts = append(parts,arm_regName[i],arm_regName[j]);
}
else {
parts = append(parts,((arm_regName[i] + std::string("-",1)) + arm_regName[j]));
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
return dsmachine_bus_Read(c->bus,a);
}
}
// tools/cpu/arm/cpu.go:253:1
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v){
{
dsmachine_bus_Write(c->bus,a,v);
}
}
// tools/cpu/arm/cpu.go:269:1
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a){
{
if (arm_Variant_isV6(c->Arch)) {
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
return dsmachine_bus_Read32(c->wide,a);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dsmachine_bus_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
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
dsmachine_bus_Write32(c->wide,a,v);
return ;
}
dsmachine_bus_Write(c->bus,a,cast<uint8_t>(v));
dsmachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
dsmachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
dsmachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
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
// tools/platform/nds/blz.go:23:1
bool nds_IsBLZ(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(8ULL))) {
return false;
}
uint32_t inc = le_Uint32(sub(data,cast<int64_t>((len(data) - cast<int64_t>(4ULL))),len(data)));
if ((inc == cast<uint32_t>(0ULL))) {
return false;
}
int64_t hdr = cast<int64_t>(data[cast<int64_t>((len(data) - cast<int64_t>(5ULL)))]);
int64_t enc = cast<int64_t>(cast<uint32_t>((le_Uint32(sub(data,cast<int64_t>((len(data) - cast<int64_t>(8ULL))),len(data))) & cast<uint32_t>(16777215ULL))));
return ((((hdr >= cast<int64_t>(8ULL)) && (hdr <= cast<int64_t>(12ULL))) && (enc >= hdr)) && (enc <= len(data)));
}
}
// tools/platform/nds/blz.go:38:1
Slice<uint8_t> nds_DecompressBLZ(Slice<uint8_t> data){
{
int64_t n = len(data);
binary_littleEndian le = go_binary_LittleEndian;
uint32_t incLen = le_Uint32(sub(data,cast<int64_t>((n - cast<int64_t>(4ULL))),len(data)));
if ((incLen == cast<uint32_t>(0ULL))) {
return append(Slice<uint8_t>{},data);
}
int64_t hdrLen = cast<int64_t>(data[cast<int64_t>((n - cast<int64_t>(5ULL)))]);
int64_t encLen = cast<int64_t>(cast<uint32_t>((le_Uint32(sub(data,cast<int64_t>((n - cast<int64_t>(8ULL))),len(data))) & cast<uint32_t>(16777215ULL))));
int64_t decLen = cast<int64_t>((n - encLen));
int64_t pakLen = cast<int64_t>((encLen - hdrLen));
int64_t rawLen = cast<int64_t>((n + cast<int64_t>(incLen)));
Slice<uint8_t> out = Slice<uint8_t>::make(rawLen);
gcopy(sub(out,0,decLen),sub(data,0,decLen));
Slice<uint8_t> comp = Slice<uint8_t>::make(pakLen);
gcopy(comp,sub(data,decLen,cast<int64_t>((decLen + pakLen))));
nds_reverse(comp);
auto tmp1 = std::make_tuple(cast<int64_t>(0ULL),decLen);
int64_t ip = std::get<0>(tmp1);
int64_t op = std::get<1>(tmp1);
uint8_t flags={};
uint8_t mask={};
{;for (;((op < rawLen) && (ip < len(comp)));){
if ((mask == cast<uint8_t>(0ULL))) {
flags = comp[ip];
ip++;
mask = cast<uint8_t>(128ULL);
}
if ((cast<uint8_t>((flags & mask)) == cast<uint8_t>(0ULL))) {
out[op] = comp[ip];
ip++;
op++;
}
else {
uint8_t b1 = comp[ip];
uint8_t b2 = comp[cast<int64_t>((ip + cast<int64_t>(1ULL)))];
ip += cast<int64_t>(2ULL);
int64_t length = cast<int64_t>((cast<int64_t>(shr<uint8_t>(b1,cast<int64_t>(4ULL))) + cast<int64_t>(3ULL)));
int64_t disp = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b1 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b2)))) + cast<int64_t>(3ULL)));
{int64_t k = cast<int64_t>(0ULL);for (;((k < length) && (op < rawLen));k++){
out[op] = out[cast<int64_t>((op - disp))];
op++;
}
}}
mask = shr<uint8_t>(mask,cast<int64_t>(1ULL));
}
}nds_reverseRange(out,decLen,rawLen);
return out;
}
}
// tools/platform/nds/blz.go:89:1
void nds_reverse(Slice<uint8_t> b){
{
nds_reverseRange(b,cast<int64_t>(0ULL),len(b));
}
}
// tools/platform/nds/blz.go:91:1
void nds_reverseRange(Slice<uint8_t> b,int64_t lo,int64_t hi){
{
{auto tmp2 = std::make_tuple(lo,cast<int64_t>((hi - cast<int64_t>(1ULL))));
int64_t i = std::get<0>(tmp2);
int64_t j = std::get<1>(tmp2);for (;(i < j);[&](){auto tmp3 = std::make_tuple(cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((j - cast<int64_t>(1ULL))));
i = std::get<0>(tmp3);
j = std::get<1>(tmp3);}()){
auto tmp4 = std::make_tuple(b[j],b[i]);
b[i] = std::get<0>(tmp4);
b[j] = std::get<1>(tmp4);
}
}}
}
// tools/platform/nds/crc.go:6:1
uint16_t nds_CRC16(Slice<uint8_t> data){
{
uint16_t crc = cast<uint16_t>(65535ULL);
{auto&& tmp5 = data;
for(int64_t tmp6=0;tmp6<len(tmp5);++tmp6){
auto b=tmp5[tmp6];crc ^= cast<uint16_t>(b);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
if ((cast<uint16_t>((crc & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL))) {
crc = cast<uint16_t>((shr<uint16_t>(crc,cast<int64_t>(1ULL)) ^ cast<uint16_t>(40961ULL)));
}
else {
crc = shr<uint16_t>(crc,cast<int64_t>(1ULL));
}
}
}}}
return crc;
}
}
// tools/platform/nds/crc.go:23:1
std::tuple<uint16_t,bool> nds_ROM_VerifyHeaderCRC(nds_ROM* r){
uint16_t computed{};
bool ok{};
{
if ((len(r->Data) < cast<int64_t>(352ULL))) {
return {cast<uint16_t>(0ULL),false};
}
computed = nds_CRC16(sub(r->Data,cast<int64_t>(0ULL),cast<int64_t>(350ULL)));
return {computed,(computed == r->Header.HeaderCRC)};
}
}
// tools/platform/nds/lz77.go:23:1
bool nds_IsLZ77(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(4ULL))) {
return false;
}
if (((data[cast<int64_t>(0ULL)] != cast<uint8_t>(16ULL)) && (data[cast<int64_t>(0ULL)] != cast<uint8_t>(17ULL)))) {
return false;
}
uint32_t size = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(data[cast<int64_t>(1ULL)]) | shl<uint32_t>(cast<uint32_t>(data[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(data[cast<int64_t>(3ULL)]),cast<int64_t>(16ULL))));
return ((size > cast<uint32_t>(0ULL)) && (size < cast<uint32_t>(268435456ULL)));
}
}
// tools/platform/nds/lz77.go:35:1
std::tuple<Slice<uint8_t>,Error> nds_DecompressLZ77(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(4ULL))) {
return {{},go_fmt_Errorf(std::string("nds: LZ77 data too short",24))};
}
uint8_t typ = data[cast<int64_t>(0ULL)];
if (((typ != cast<uint8_t>(16ULL)) && (typ != cast<uint8_t>(17ULL)))) {
return {{},go_fmt_Errorf(std::string("nds: not an LZ77 stream (type 0x%02X)",37),typ)};
}
int64_t size = cast<int64_t>(shr<uint32_t>(le_Uint32(data),cast<int64_t>(8ULL)));
Slice<uint8_t> out = Slice<uint8_t>::make(cast<int64_t>(0ULL),size);
int64_t p = cast<int64_t>(4ULL);
auto next = [&]()->std::tuple<uint8_t,Error>{
if ((p >= len(data))) {
return {cast<uint8_t>(0ULL),go_fmt_Errorf(std::string("nds: LZ77 input truncated",25))};
}
uint8_t b = data[p];
p++;
return {b,{}};
}
;
{;for (;(len(out) < size);){
auto tmp7 = next();
uint8_t flags = std::get<0>(tmp7);
Error err = std::get<1>(tmp7);
if (bool(err)) {
return {{},err};
}
{int64_t bit = cast<int64_t>(0ULL);for (;((bit < cast<int64_t>(8ULL)) && (len(out) < size));bit++){
if ((cast<uint8_t>((flags & cast<uint8_t>(128ULL))) == cast<uint8_t>(0ULL))) {
auto tmp8 = next();
uint8_t b = std::get<0>(tmp8);
Error err = std::get<1>(tmp8);
if (bool(err)) {
return {{},err};
}
out = append(out,b);
}
else {
auto tmp9 = next();
uint8_t b0 = std::get<0>(tmp9);
Error err = std::get<1>(tmp9);
if (bool(err)) {
return {{},err};
}
auto tmp10 = next();
uint8_t b1 = std::get<0>(tmp10);
err = std::get<1>(tmp10);
if (bool(err)) {
return {{},err};
}
int64_t length={};
int64_t disp={};
if ((typ == cast<uint8_t>(16ULL))) {
length = cast<int64_t>((cast<int64_t>(shr<uint8_t>(b0,cast<int64_t>(4ULL))) + cast<int64_t>(3ULL)));
disp = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b0 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b1)))) + cast<int64_t>(1ULL)));
}
else {
{
switch(shr<uint8_t>(b0,cast<int64_t>(4ULL))){
case cast<uint8_t>(0ULL):{
auto tmp11 = next();
uint8_t b2 = std::get<0>(tmp11);
Error err = std::get<1>(tmp11);
if (bool(err)) {
return {{},err};
}
length = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b0 & cast<uint8_t>(15ULL)))),cast<int64_t>(4ULL)) | cast<int64_t>(shr<uint8_t>(b1,cast<int64_t>(4ULL)))))) + cast<int64_t>(17ULL)));
disp = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b1 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b2)))) + cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(1ULL):{
auto tmp12 = next();
uint8_t b2 = std::get<0>(tmp12);
Error err = std::get<1>(tmp12);
if (bool(err)) {
return {{},err};
}
auto tmp13 = next();
uint8_t b3 = std::get<0>(tmp13);
err = std::get<1>(tmp13);
if (bool(err)) {
return {{},err};
}
length = cast<int64_t>(((cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b0 & cast<uint8_t>(15ULL)))),cast<int64_t>(12ULL)) | shl<int64_t>(cast<int64_t>(b1),cast<int64_t>(4ULL)))) | cast<int64_t>(shr<uint8_t>(b2,cast<int64_t>(4ULL)))))) + cast<int64_t>(273ULL)));
disp = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b2 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b3)))) + cast<int64_t>(1ULL)));
break;}
default:{
length = cast<int64_t>((cast<int64_t>(shr<uint8_t>(b0,cast<int64_t>(4ULL))) + cast<int64_t>(1ULL)));
disp = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b0 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b1)))) + cast<int64_t>(1ULL)));
break;}
}}
}
int64_t start = cast<int64_t>((len(out) - disp));
if ((start < cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("nds: LZ77 back-reference before start",37))};
}
{int64_t i = cast<int64_t>(0ULL);for (;((i < length) && (len(out) < size));i++){
out = append(out,out[cast<int64_t>((start + i))]);
}
}}
flags = shl<uint8_t>(flags,cast<int64_t>(1ULL));
}
}}
}return {out,{}};
}
}
// tools/platform/nds/lz77.go:121:1
Slice<uint8_t> nds_Decompress(Slice<uint8_t> data){
{
if (nds_IsLZ77(data)) {
{
auto tmp14 = nds_DecompressLZ77(data);
Slice<uint8_t> out = std::get<0>(tmp14);
Error err = std::get<1>(tmp14);
if ((!err)) {
return out;
}
}
}
return data;
}
}
// tools/platform/nds/ndsrom.go:71:1
int64_t nds_Header_ChipBytes(nds_Header h){
{
return shl<int64_t>(cast<int64_t>(131072ULL),h.DeviceCap);
}
}
// tools/platform/nds/ndsrom.go:74:1
std::tuple<nds_Header,Error> nds_ParseHeader(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(512ULL))) {
return {nds_Header{},go_errors_New(std::string("nds: image shorter than a 0x200 header",38))};
}
binary_littleEndian le = go_binary_LittleEndian;
auto str = [&](int64_t off,int64_t n)->std::string{
return go_strings_TrimRight(cast<std::string>(sub(data,off,cast<int64_t>((off + n)))),std::string("\000",1));
}
;
return {nds_Header{str(cast<int64_t>(0ULL),cast<int64_t>(12ULL)),str(cast<int64_t>(12ULL),cast<int64_t>(4ULL)),str(cast<int64_t>(16ULL),cast<int64_t>(2ULL)),data[cast<int64_t>(18ULL)],data[cast<int64_t>(19ULL)],data[cast<int64_t>(20ULL)],data[cast<int64_t>(30ULL)],data[cast<int64_t>(31ULL)],le_Uint32(sub(data,cast<int64_t>(32ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(36ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(40ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(44ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(48ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(52ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(56ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(60ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(64ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(68ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(72ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(76ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(80ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(84ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(88ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(92ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(104ULL),len(data))),le_Uint16(sub(data,cast<int64_t>(108ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(112ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(116ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(128ULL),len(data))),le_Uint32(sub(data,cast<int64_t>(132ULL),len(data))),le_Uint16(sub(data,cast<int64_t>(348ULL),len(data))),le_Uint16(sub(data,cast<int64_t>(350ULL),len(data)))},{}};
}
}
// tools/platform/nds/ndsrom.go:130:1
uint32_t nds_FATEntry_Size(nds_FATEntry e){
{
return cast<uint32_t>((e.End - e.Start));
}
}
// tools/platform/nds/ndsrom.go:149:1
std::tuple<nds_ROM*,Error> nds_Open(Slice<uint8_t> data){
{
auto tmp15 = nds_ParseHeader(data);
nds_Header h = std::get<0>(tmp15);
Error err = std::get<1>(tmp15);
if (bool(err)) {
return {{},err};
}
nds_ROM* r = arenaNew(nds_ROM{data,h,{},{},Map<std::string,int64_t>{}});
{
Error err = nds_ROM_parseFAT(r);
if (bool(err)) {
return {{},err};
}
}
{
Error err = nds_ROM_parseFNT(r);
if (bool(err)) {
return {{},err};
}
}
return {r,{}};
}
}
// tools/platform/nds/ndsrom.go:164:1
Error nds_ROM_parseFAT(nds_ROM* r){
{
auto tmp16 = std::make_tuple(cast<int64_t>(r->Header.FATOff),cast<int64_t>(r->Header.FATSize));
int64_t off = std::get<0>(tmp16);
int64_t size = std::get<1>(tmp16);
if ((cast<int64_t>((off + size)) > len(r->Data))) {
return go_errors_New(std::string("nds: FAT out of range",21));
}
binary_littleEndian le = go_binary_LittleEndian;
int64_t n = divi<int64_t>(size,cast<int64_t>(8ULL));
r->FAT = Slice<nds_FATEntry>::make(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
Slice<uint8_t> b = sub(r->Data,cast<int64_t>((off + cast<int64_t>((i * cast<int64_t>(8ULL))))),len(r->Data));
r->FAT[i] = nds_FATEntry{le_Uint32(b),le_Uint32(sub(b,cast<int64_t>(4ULL),len(b)))};
}
}return {};
}
}
// tools/platform/nds/ndsrom.go:182:1
Error nds_ROM_parseFNT(nds_ROM* r){
{
binary_littleEndian le = go_binary_LittleEndian;
int64_t base = cast<int64_t>(r->Header.FNTOff);
if ((cast<int64_t>((base + cast<int64_t>(8ULL))) > len(r->Data))) {
return go_errors_New(std::string("nds: FNT out of range",21));
}
int64_t dirCount = cast<int64_t>(le_Uint16(sub(r->Data,cast<int64_t>((base + cast<int64_t>(6ULL))),len(r->Data))));
if (((dirCount == cast<int64_t>(0ULL)) || (dirCount > cast<int64_t>(4096ULL)))) {
dirCount = cast<int64_t>(1ULL);
}
std::function<Error(uint16_t,std::string)> walk={};
walk = [&](uint16_t dirID,std::string prefix)->Error{
int64_t idx = cast<int64_t>(cast<uint16_t>((dirID & cast<uint16_t>(4095ULL))));
int64_t rec = cast<int64_t>((base + cast<int64_t>((idx * cast<int64_t>(8ULL)))));
if ((cast<int64_t>((rec + cast<int64_t>(8ULL))) > len(r->Data))) {
return go_errors_New(std::string("nds: FNT directory record out of range",38));
}
int64_t sub_ = cast<int64_t>((base + cast<int64_t>(le_Uint32(sub(r->Data,rec,len(r->Data))))));
int64_t fileID = cast<int64_t>(le_Uint16(sub(r->Data,cast<int64_t>((rec + cast<int64_t>(4ULL))),len(r->Data))));
int64_t p = sub_;
{;for (;;){
if ((p >= len(r->Data))) {
return go_errors_New(std::string("nds: FNT sub-table overran image",32));
}
uint8_t ctrl = r->Data[p];
p++;
if ((ctrl == cast<uint8_t>(0ULL))) {
break;
}
int64_t nameLen = cast<int64_t>(cast<uint8_t>((ctrl & cast<uint8_t>(127ULL))));
if ((cast<int64_t>((p + nameLen)) > len(r->Data))) {
return go_errors_New(std::string("nds: FNT name overran image",27));
}
std::string name = cast<std::string>(sub(r->Data,p,cast<int64_t>((p + nameLen))));
p += nameLen;
std::string full = name;
if ((prefix != std::string("",0))) {
full = ((prefix + std::string("/",1)) + name);
}
if ((cast<uint8_t>((ctrl & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
uint16_t childID = le_Uint16(sub(r->Data,p,len(r->Data)));
p += cast<int64_t>(2ULL);
{
Error err = walk(childID,full);
if (bool(err)) {
return err;
}
}
}
else {
if ((fileID < len(r->FAT))) {
r->Files = append(r->Files,nds_FileInfo{fileID,full});
r->byPath[full] = fileID;
}
fileID++;
}
}
}return {};
}
;
return walk(cast<uint16_t>(61440ULL),std::string("",0));
}
}
// tools/platform/nds/ndsrom.go:243:1
Slice<uint8_t> nds_ROM_File(nds_ROM* r,int64_t id){
{
if (((id < cast<int64_t>(0ULL)) || (id >= len(r->FAT)))) {
return {};
}
nds_FATEntry e = r->FAT[id];
if (((cast<int64_t>(e.End) > len(r->Data)) || (e.Start > e.End))) {
return {};
}
return sub(r->Data,e.Start,e.End);
}
}
// tools/platform/nds/ndsrom.go:256:1
Slice<uint8_t> nds_ROM_FileByPath(nds_ROM* r,std::string path){
{
auto tmp17 = lookup(r->byPath,go_strings_TrimPrefix(path,std::string("/",1)));
int64_t id = std::get<0>(tmp17);
bool ok = std::get<1>(tmp17);
if ((!ok)) {
return {};
}
return nds_ROM_File(r,id);
}
}
// tools/platform/nds/ndsrom.go:279:1
Slice<nds_Overlay> nds_ROM_ARM9Overlays(nds_ROM* r){
{
return nds_ROM_overlays(r,r->Header.ARM9OverlayOff,r->Header.ARM9OverlaySize);
}
}
// tools/platform/nds/ndsrom.go:283:1
Slice<nds_Overlay> nds_ROM_overlays(nds_ROM* r,uint32_t off,uint32_t size){
{
if (((size == cast<uint32_t>(0ULL)) || (cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(size))) > len(r->Data)))) {
return {};
}
binary_littleEndian le = go_binary_LittleEndian;
int64_t n = divi<int64_t>(cast<int64_t>(size),cast<int64_t>(32ULL));
Slice<nds_Overlay> out = Slice<nds_Overlay>::make(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
Slice<uint8_t> b = sub(r->Data,cast<int64_t>((cast<int64_t>(off) + cast<int64_t>((i * cast<int64_t>(32ULL))))),len(r->Data));
uint32_t flags = le_Uint32(sub(b,cast<int64_t>(28ULL),len(b)));
out[i] = nds_Overlay{le_Uint32(sub(b,cast<int64_t>(0ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(4ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(8ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(12ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(16ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(20ULL),len(b))),le_Uint32(sub(b,cast<int64_t>(24ULL),len(b))),cast<uint32_t>((flags & cast<uint32_t>(16777215ULL))),(cast<uint32_t>((flags & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL))};
}
}return out;
}
}
// tools/platform/nds/ndsrom.go:309:1
Slice<uint8_t> nds_ROM_ARM9(nds_ROM* r){
{
return nds_ROM_slice(r,r->Header.ARM9ROMOff,r->Header.ARM9Size);
}
}
// tools/platform/nds/ndsrom.go:312:1
Slice<uint8_t> nds_ROM_ARM7(nds_ROM* r){
{
return nds_ROM_slice(r,r->Header.ARM7ROMOff,r->Header.ARM7Size);
}
}
// tools/platform/nds/ndsrom.go:314:1
Slice<uint8_t> nds_ROM_slice(nds_ROM* r,uint32_t off,uint32_t size){
{
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(size))) > len(r->Data))) {
return {};
}
return sub(r->Data,off,cast<uint32_t>((off + size)));
}
}
// tools/platform/nds/dsmachine/bios.go:23:1
std::function<bool(arm_CPU*,uint32_t)> dsmachine_biosSWI(dsmachine_core* c){
{
dsmachine_bus* b = arenaNew(dsmachine_bus{c});
return [=](arm_CPU* cpu,uint32_t comment)->bool{
uint32_t n = cast<uint32_t>((comment & cast<uint32_t>(255ULL)));
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(((shr<uint32_t>(comment,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
}
{
switch(n){
case cast<uint32_t>(3ULL):{
int64_t d = cast<int64_t>(cpu->R[cast<int64_t>(0ULL)]);
if ((d > cast<int64_t>(1048576ULL))) {
d = cast<int64_t>(1048576ULL);
}
c->sleep += d;
break;}
case cast<uint32_t>(4ULL):{
dsmachine_core_park(c,cpu,cpu->R[cast<int64_t>(1ULL)],false);
break;}
case cast<uint32_t>(5ULL):{
dsmachine_core_park(c,cpu,cast<uint32_t>(1ULL),false);
break;}
case cast<uint32_t>(6ULL):{
dsmachine_core_park(c,cpu,cast<uint32_t>(0ULL),true);
break;}
case cast<uint32_t>(31ULL):{
dsmachine_core_park(c,cpu,cast<uint32_t>(0ULL),true);
break;}
case cast<uint32_t>(7ULL):{
arm_CPU_Halt(cpu,std::string("%s: SWI Stop (deep sleep)",25),c->name);
break;}
case cast<uint32_t>(9ULL):{
auto tmp1 = std::make_tuple(cast<int32_t>(cpu->R[cast<int64_t>(0ULL)]),cast<int32_t>(cpu->R[cast<int64_t>(1ULL)]));
int32_t num = std::get<0>(tmp1);
int32_t den = std::get<1>(tmp1);
if ((den == cast<int32_t>(0ULL))) {
break;
}
auto tmp2 = std::make_tuple(divi<int32_t>(num,den),modi<int32_t>(num,den));
int32_t q = std::get<0>(tmp2);
int32_t r = std::get<1>(tmp2);
auto tmp3 = std::make_tuple(cast<uint32_t>(q),cast<uint32_t>(r));
cpu->R[cast<int64_t>(0ULL)] = std::get<0>(tmp3);
cpu->R[cast<int64_t>(1ULL)] = std::get<1>(tmp3);
if ((q < cast<int32_t>(0ULL))) {
q = cast<int32_t>(-q);
}
cpu->R[cast<int64_t>(3ULL)] = cast<uint32_t>(q);
break;}
case cast<uint32_t>(13ULL):{
cpu->R[cast<int64_t>(0ULL)] = dsmachine_isqrt(cast<uint64_t>(cpu->R[cast<int64_t>(0ULL)]));
break;}
case cast<uint32_t>(14ULL):{
Slice<uint8_t> data = Slice<uint8_t>::make(cpu->R[cast<int64_t>(2ULL)]);
{auto&& tmp4 = data;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;data[i] = dsmachine_bus_Read(b,cast<uint32_t>((cpu->R[cast<int64_t>(1ULL)] + cast<uint32_t>(i))));
}}
cpu->R[cast<int64_t>(0ULL)] = cast<uint32_t>(nds_CRC16(data));
break;}
case cast<uint32_t>(11ULL):{
dsmachine_cpuSet(b,cpu,false);
break;}
case cast<uint32_t>(12ULL):{
dsmachine_cpuSet(b,cpu,true);
break;}
case cast<uint32_t>(15ULL):{
cpu->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):{
dsmachine_lz77UnComp(b,cpu->R[cast<int64_t>(0ULL)],cpu->R[cast<int64_t>(1ULL)]);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):{
dsmachine_rlUnComp(b,cpu->R[cast<int64_t>(0ULL)],cpu->R[cast<int64_t>(1ULL)]);
break;}
case cast<uint32_t>(22ULL):case cast<uint32_t>(24ULL):{
dsmachine_diffUnFilter(b,cpu->R[cast<int64_t>(0ULL)],cpu->R[cast<int64_t>(1ULL)],(n == cast<uint32_t>(24ULL)));
break;}
default:{
dsmachine_Machine_note(c->m,std::string("%s: SWI 0x%02X not implemented (ignored)",40),c->name,n);
break;}
}}
return true;
}
;
}
}
// tools/platform/nds/dsmachine/bios.go:106:1
void dsmachine_core_park(dsmachine_core* c,arm_CPU* cpu,uint32_t mask,bool any){
{
c->waiting = true;
c->waitMask = mask;
c->waitAny = any;
}
}
// tools/platform/nds/dsmachine/bios.go:114:1
void dsmachine_cpuSet(dsmachine_bus* b,arm_CPU* c,bool fast){
{
auto tmp6 = std::make_tuple(c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)]);
uint32_t src = std::get<0>(tmp6);
uint32_t dst = std::get<1>(tmp6);
uint32_t ctrl = std::get<2>(tmp6);
bool fill = (cast<uint32_t>((ctrl & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL));
uint32_t n = cast<uint32_t>((ctrl & cast<uint32_t>(2097151ULL)));
bool word = (fast || (cast<uint32_t>((ctrl & cast<uint32_t>(67108864ULL))) != cast<uint32_t>(0ULL)));
if (word) {
uint32_t v={};
if (fill) {
v = dsmachine_bus_r32(b,src);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((!fill)) {
v = dsmachine_bus_r32(b,cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
dsmachine_bus_w32(b,cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(4ULL))))),v);
}
}return ;
}
uint32_t v={};
if (fill) {
v = cast<uint32_t>((cast<uint32_t>(dsmachine_bus_Read(b,src)) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((!fill)) {
v = cast<uint32_t>((cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(2ULL))))))) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(2ULL))))) + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
dsmachine_bus_Write(b,cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(2ULL))))),cast<uint8_t>(v));
dsmachine_bus_Write(b,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(2ULL))))) + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
}
}}
}
// tools/platform/nds/dsmachine/bios.go:149:1
void dsmachine_lz77UnComp(dsmachine_bus* b,uint32_t src,uint32_t dst){
{
uint32_t hdr = dsmachine_bus_r32(b,src);
uint32_t size = shr<uint32_t>(hdr,cast<int64_t>(8ULL));
src += cast<uint32_t>(4ULL);
uint32_t written={};
{;for (;(written < size);){
uint8_t flags = dsmachine_bus_Read(b,src);
src++;
{int64_t i = cast<int64_t>(0ULL);for (;((i < cast<int64_t>(8ULL)) && (written < size));i++){
if ((cast<uint8_t>((flags & (shr<uint8_t>(cast<uint8_t>(128ULL),cast<uint64_t>(i))))) == cast<uint8_t>(0ULL))) {
dsmachine_bus_Write(b,cast<uint32_t>((dst + written)),dsmachine_bus_Read(b,src));
src++;
written++;
continue;
}
auto tmp7 = std::make_tuple(dsmachine_bus_Read(b,src),dsmachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>(1ULL)))));
uint8_t hi = std::get<0>(tmp7);
uint8_t lo = std::get<1>(tmp7);
src += cast<uint32_t>(2ULL);
uint32_t length = cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(hi,cast<int64_t>(4ULL))) + cast<uint32_t>(3ULL)));
uint32_t disp = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(cast<uint8_t>((hi & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<uint32_t>(lo))) + cast<uint32_t>(1ULL)));
{uint32_t j = cast<uint32_t>(0ULL);for (;((j < length) && (written < size));j++){
dsmachine_bus_Write(b,cast<uint32_t>((dst + written)),dsmachine_bus_Read(b,cast<uint32_t>((cast<uint32_t>((dst + written)) - disp))));
written++;
}
}}
}}
}}
}
// tools/platform/nds/dsmachine/bios.go:178:1
void dsmachine_rlUnComp(dsmachine_bus* b,uint32_t src,uint32_t dst){
{
uint32_t hdr = dsmachine_bus_r32(b,src);
uint32_t size = shr<uint32_t>(hdr,cast<int64_t>(8ULL));
src += cast<uint32_t>(4ULL);
uint32_t written={};
{;for (;(written < size);){
uint8_t f = dsmachine_bus_Read(b,src);
src++;
if ((cast<uint8_t>((f & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
uint32_t n = cast<uint32_t>((cast<uint32_t>(cast<uint8_t>((f & cast<uint8_t>(127ULL)))) + cast<uint32_t>(3ULL)));
uint8_t v = dsmachine_bus_Read(b,src);
src++;
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (written < size));i++){
dsmachine_bus_Write(b,cast<uint32_t>((dst + written)),v);
written++;
}
}}
else {
uint32_t n = cast<uint32_t>((cast<uint32_t>(cast<uint8_t>((f & cast<uint8_t>(127ULL)))) + cast<uint32_t>(1ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (written < size));i++){
dsmachine_bus_Write(b,cast<uint32_t>((dst + written)),dsmachine_bus_Read(b,src));
src++;
written++;
}
}}
}
}}
}
// tools/platform/nds/dsmachine/bios.go:208:1
void dsmachine_diffUnFilter(dsmachine_bus* b,uint32_t src,uint32_t dst,bool wide){
{
uint32_t hdr = dsmachine_bus_r32(b,src);
uint32_t size = shr<uint32_t>(hdr,cast<int64_t>(8ULL));
src += cast<uint32_t>(4ULL);
if (wide) {
uint16_t sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i += cast<uint32_t>(2ULL)){
uint16_t d = cast<uint16_t>((cast<uint16_t>(dsmachine_bus_Read(b,src)) | shl<uint16_t>(cast<uint16_t>(dsmachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
src += cast<uint32_t>(2ULL);
sum += d;
dsmachine_bus_Write(b,cast<uint32_t>((dst + i)),cast<uint8_t>(sum));
dsmachine_bus_Write(b,cast<uint32_t>((cast<uint32_t>((dst + i)) + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(sum,cast<int64_t>(8ULL))));
}
}return ;
}
uint8_t sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
sum += dsmachine_bus_Read(b,src);
src++;
dsmachine_bus_Write(b,cast<uint32_t>((dst + i)),sum);
}
}}
}
// tools/platform/nds/dsmachine/bios.go:233:1
uint32_t dsmachine_isqrt(uint64_t v){
{
uint64_t res=cast<uint64_t>(0ULL);
uint64_t bit=cast<uint64_t>(4611686018427387904ULL);
{;for (;(bit > v);){
bit = shr<uint64_t>(bit,cast<int64_t>(2ULL));
}
}{;for (;(bit != cast<uint64_t>(0ULL));){
if ((v >= cast<uint64_t>((res + bit)))) {
v -= cast<uint64_t>((res + bit));
res = cast<uint64_t>((shr<uint64_t>(res,cast<int64_t>(1ULL)) + bit));
}
else {
res = shr<uint64_t>(res,cast<int64_t>(1ULL));
}
bit = shr<uint64_t>(bit,cast<int64_t>(2ULL));
}
}return cast<uint32_t>(res);
}
}
// tools/platform/nds/dsmachine/bios.go:255:1
std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> dsmachine_cp15(dsmachine_core* c){
{
return [=](arm_CPU* cpu,bool load,uint32_t cp,uint32_t op1,uint32_t crn,uint32_t crm,uint32_t op2,uint32_t* rd)->void{
if (((((((!load) && (cp == cast<uint32_t>(15ULL))) && (op1 == cast<uint32_t>(0ULL))) && (crn == cast<uint32_t>(7ULL))) && (crm == cast<uint32_t>(0ULL))) && (op2 == cast<uint32_t>(4ULL)))) {
c->wfi = true;
return ;
}
if (((load || (crn != cast<uint32_t>(9ULL))) || (crm != cast<uint32_t>(1ULL)))) {
return ;
}
uint32_t v = (*rd);
uint32_t base = (v & ~(cast<uint32_t>(4095ULL)));
if ((op2 == cast<uint32_t>(0ULL))) {
if ((!c->dtcm)) {
c->dtcm = Slice<uint8_t>::make(cast<int64_t>(16384ULL));
}
c->dtcmBase = base;
c->handlerBase = cast<uint32_t>((base + cast<uint32_t>(len(c->dtcm))));
dsmachine_Machine_note(c->m,std::string("ARM9: CP15 DTCM base 0x%08X",27),base);
}
else {
dsmachine_Machine_note(c->m,std::string("ARM9: CP15 ITCM control 0x%08X",30),v);
}
}
;
}
}
// tools/platform/nds/dsmachine/boot.go:37:1
void dsmachine_Machine_directBoot(dsmachine_Machine* m,nds_ROM* rom){
{
dsmachine_bus bStorage = dsmachine_bus{m->ARM9};
dsmachine_bus* b = &bStorage;
binary_littleEndian le = go_binary_LittleEndian;
int64_t n = cast<int64_t>(368ULL);
if ((len(rom->Data) < n)) {
n = len(rom->Data);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
dsmachine_bus_Write(b,cast<uint32_t>((cast<uint32_t>(41942528ULL) + cast<uint32_t>(i))),rom->Data[i]);
}
}constexpr int64_t chipID=16322ULL;
{auto&& tmp8 = Slice<uint32_t>{cast<uint32_t>(41940992ULL),cast<uint32_t>(41942016ULL)};
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto base=tmp8[tmp9];dsmachine_bus_w32(b,cast<uint32_t>((base + cast<uint32_t>(0ULL))),cast<uint32_t>(16322ULL));
dsmachine_bus_w32(b,cast<uint32_t>((base + cast<uint32_t>(4ULL))),cast<uint32_t>(16322ULL));
dsmachine_bus_Write(b,cast<uint32_t>((base + cast<uint32_t>(8ULL))),cast<uint8_t>(rom->Header.HeaderCRC));
dsmachine_bus_Write(b,cast<uint32_t>((base + cast<uint32_t>(9ULL))),cast<uint8_t>(shr<uint16_t>(rom->Header.HeaderCRC,cast<int64_t>(8ULL))));
dsmachine_bus_Write(b,cast<uint32_t>((base + cast<uint32_t>(10ULL))),cast<uint8_t>(rom->Header.SecureCRC));
dsmachine_bus_Write(b,cast<uint32_t>((base + cast<uint32_t>(11ULL))),cast<uint8_t>(shr<uint16_t>(rom->Header.SecureCRC,cast<int64_t>(8ULL))));
}}
dsmachine_bus_Write(b,cast<uint32_t>(41942080ULL),cast<uint8_t>(1ULL));
Slice<uint8_t> us = dsmachine_spibus_currentUserSettings(m->spi);
{auto&& tmp10 = us;
for(int64_t tmp11=0;tmp11<len(tmp10);++tmp11){
auto i=tmp11;auto v=tmp10[tmp11];dsmachine_bus_Write(b,cast<uint32_t>((cast<uint32_t>(41942144ULL) + cast<uint32_t>(i))),v);
}}
(void)(le);
}
}
// tools/platform/nds/dsmachine/boot.go:73:1
Slice<uint8_t> dsmachine_spibus_currentUserSettings(dsmachine_spibus* s){
{
binary_littleEndian le = go_binary_LittleEndian;
Slice<uint8_t> best = Slice<uint8_t>{};
int64_t bestCount = cast<int64_t>(-1ULL);
{auto&& tmp12 = Slice<int64_t>{cast<int64_t>(261632ULL),cast<int64_t>(261888ULL)};
for(int64_t tmp13=0;tmp13<len(tmp12);++tmp13){
auto off=tmp12[tmp13];Slice<uint8_t> blk = sub(s->firmware,off,cast<int64_t>((off + cast<int64_t>(256ULL))));
if ((nds_CRC16(sub(blk,cast<int64_t>(0ULL),cast<int64_t>(112ULL))) != le_Uint16(sub(blk,cast<int64_t>(114ULL),len(blk))))) {
continue;
}
{
int64_t c = cast<int64_t>(le_Uint16(sub(blk,cast<int64_t>(112ULL),len(blk))));
if ((c > bestCount)) {
auto tmp14 = std::make_tuple(blk,c);
best = std::get<0>(tmp14);
bestCount = std::get<1>(tmp14);
}
}
}}
if ((!best)) {
return Slice<uint8_t>::make(cast<int64_t>(112ULL));
}
return sub(best,0,cast<int64_t>(112ULL));
}
}
// tools/platform/nds/dsmachine/bus.go:12:1
std::tuple<Slice<uint8_t>,uint32_t> dsmachine_bus_slot(dsmachine_bus* b,uint32_t a){
{
dsmachine_core* c = b->c;
{
if (((a >= cast<uint32_t>(33554432ULL)) && (a < cast<uint32_t>(50331648ULL)))){
if (((bool(c->dtcm) && (a >= c->dtcmBase)) && (a < cast<uint32_t>((c->dtcmBase + cast<uint32_t>(len(c->dtcm))))))) {
return {c->dtcm,cast<uint32_t>((a - c->dtcmBase))};
}
return {c->m->ram,cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(33554432ULL)))) & cast<uint32_t>(4194303ULL)))};
}
else if (((a >= cast<uint32_t>(50331648ULL)) && (a < cast<uint32_t>(58720256ULL)))){
return dsmachine_core_wramSlot(c,a);
}
}
tmp15:;
if (c->arm9) {
if (((a >= c->itcmBase) && (a < cast<uint32_t>((c->itcmBase + cast<uint32_t>(len(c->itcm))))))) {
return {c->itcm,cast<uint32_t>((a - c->itcmBase))};
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(5ULL):{
return {c->m->pal,cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(83886080ULL)))) & cast<uint32_t>(2047ULL)))};
break;}
case cast<uint32_t>(7ULL):{
return {c->m->oam,cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(117440512ULL)))) & cast<uint32_t>(2047ULL)))};
break;}
}}
}
else {
if ((a < cast<uint32_t>(len(c->low)))) {
return {c->low,a};
}
if (((a >= cast<uint32_t>(58720256ULL)) && (a < cast<uint32_t>(67108864ULL)))) {
return {c->wram7,cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(58720256ULL)))) & cast<uint32_t>(65535ULL)))};
}
}
return {{},cast<uint32_t>(0ULL)};
}
}
// tools/platform/nds/dsmachine/bus.go:70:1
std::tuple<Slice<uint8_t>,uint32_t> dsmachine_core_wramSlot(dsmachine_core* c,uint32_t a){
{
uint32_t off = cast<uint32_t>((a - cast<uint32_t>(50331648ULL)));
constexpr int64_t bank=16384ULL;
uint8_t mode = cast<uint8_t>((c->m->wramcnt & cast<uint8_t>(3ULL)));
if (c->arm9) {
{
switch(mode){
case cast<uint8_t>(0ULL):{
return {c->m->swram,cast<uint32_t>((off & cast<uint32_t>(32767ULL)))};
break;}
case cast<uint8_t>(1ULL):{
return {c->m->swram,cast<uint32_t>((cast<uint32_t>(16384ULL) + cast<uint32_t>((off & cast<uint32_t>(16383ULL)))))};
break;}
case cast<uint8_t>(2ULL):{
return {c->m->swram,cast<uint32_t>((off & cast<uint32_t>(16383ULL)))};
break;}
default:{
return {{},cast<uint32_t>(0ULL)};
break;}
}}
}
{
switch(mode){
case cast<uint8_t>(1ULL):{
return {c->m->swram,cast<uint32_t>((off & cast<uint32_t>(16383ULL)))};
break;}
case cast<uint8_t>(2ULL):{
return {c->m->swram,cast<uint32_t>((cast<uint32_t>(16384ULL) + cast<uint32_t>((off & cast<uint32_t>(16383ULL)))))};
break;}
case cast<uint8_t>(3ULL):{
return {c->m->swram,cast<uint32_t>((off & cast<uint32_t>(32767ULL)))};
break;}
default:{
return {c->wram7,cast<uint32_t>((a & cast<uint32_t>(65535ULL)))};
break;}
}}
}
}
// tools/platform/nds/dsmachine/bus.go:98:1
uint8_t dsmachine_bus_ReadReference(dsmachine_bus* b,uint32_t a){
{
dsmachine_core* c = b->c;
uint8_t v = dsmachine_bus_read(b,a);
if (bool(c->m->OnRead)) {
c->m->OnRead(c->arm9,a,v,c->cpu->R[cast<int64_t>(15ULL)]);
}
return v;
}
}
// tools/platform/nds/dsmachine/bus.go:107:1
uint8_t dsmachine_bus_read(dsmachine_bus* b,uint32_t a){
{
dsmachine_core* c = b->c;
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(4ULL):{
return cast<uint8_t>(shr<uint32_t>(dsmachine_core_ioRead(c,(a & ~(cast<uint32_t>(3ULL)))),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(3ULL)))))))));
break;}
case cast<uint32_t>(6ULL):{
return dsmachine_core_vramRead(c,a);
break;}
}}
{
auto tmp16 = dsmachine_bus_slot(b,a);
Slice<uint8_t> s = std::get<0>(tmp16);
uint32_t i = std::get<1>(tmp16);
if (bool(s)) {
return s[i];
}
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/nds/dsmachine/bus.go:121:1
void dsmachine_bus_WriteReference(dsmachine_bus* b,uint32_t a,uint8_t v){
{
dsmachine_core* c = b->c;
if (bool(c->m->OnWrite)) {
c->m->OnWrite(c->arm9,a,v,c->cpu->R[cast<int64_t>(15ULL)]);
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(4ULL):{
dsmachine_core_ioWrite(c,a,v);
return ;
break;}
case cast<uint32_t>(6ULL):{
dsmachine_core_vramWrite(c,a,v);
return ;
break;}
}}
{
auto tmp17 = dsmachine_bus_slot(b,a);
Slice<uint8_t> s = std::get<0>(tmp17);
uint32_t i = std::get<1>(tmp17);
if (bool(s)) {
s[i] = v;
}
}
}
}
// tools/platform/nds/dsmachine/bus.go:143:1
uint8_t dsmachine_core_vramRead(dsmachine_core* c,uint32_t a){
{
dsmachine_vram* v = c->m->vram;
if (c->arm9) {
{
auto tmp18 = dsmachine_vram_lcdcSlot(v,a);
Slice<uint8_t> data = std::get<0>(tmp18);
uint32_t off = std::get<1>(tmp18);
bool ok = std::get<2>(tmp18);
if (ok) {
return data[off];
}
}
}
{
auto tmp19 = dsmachine_vram_cpuAccess(v,c->arm9,a);
int64_t space = std::get<0>(tmp19);
uint32_t off = std::get<1>(tmp19);
bool ok = std::get<2>(tmp19);
if (ok) {
return dsmachine_vram_read8(v,space,off);
}
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/nds/dsmachine/bus.go:156:1
void dsmachine_core_vramWrite(dsmachine_core* c,uint32_t a,uint8_t b){
{
dsmachine_vram* v = c->m->vram;
if (c->arm9) {
{
auto tmp20 = dsmachine_vram_lcdcSlot(v,a);
Slice<uint8_t> data = std::get<0>(tmp20);
uint32_t off = std::get<1>(tmp20);
bool ok = std::get<2>(tmp20);
if (ok) {
data[off] = b;
return ;
}
}
}
{
auto tmp21 = dsmachine_vram_cpuAccess(v,c->arm9,a);
int64_t space = std::get<0>(tmp21);
uint32_t off = std::get<1>(tmp21);
bool ok = std::get<2>(tmp21);
if (ok) {
dsmachine_vram_write8(v,space,off,b);
}
}
}
}
// tools/platform/nds/dsmachine/bus.go:185:1
uint16_t dsmachine_bus_Read16Reference(dsmachine_bus* b,uint32_t a){
{
if ((shr<uint32_t>(a,cast<int64_t>(24ULL)) == cast<uint32_t>(4ULL))) {
uint16_t v = cast<uint16_t>(shr<uint32_t>(dsmachine_core_ioRead(b->c,(a & ~(cast<uint32_t>(3ULL)))),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(2ULL)))))))));
dsmachine_bus_watchRead(b,a,cast<uint8_t>(v));
return v;
}
return cast<uint16_t>((cast<uint16_t>(dsmachine_bus_Read(b,a)) | shl<uint16_t>(cast<uint16_t>(dsmachine_bus_Read(b,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/platform/nds/dsmachine/bus.go:194:1
uint32_t dsmachine_bus_Read32Reference(dsmachine_bus* b,uint32_t a){
{
if ((shr<uint32_t>(a,cast<int64_t>(24ULL)) == cast<uint32_t>(4ULL))) {
uint32_t v = dsmachine_core_ioRead(b->c,a);
dsmachine_bus_watchRead(b,a,cast<uint8_t>(v));
return v;
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dsmachine_bus_Read(b,a)) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/platform/nds/dsmachine/bus.go:206:1
void dsmachine_bus_watchRead(dsmachine_bus* b,uint32_t a,uint8_t v){
{
if (bool(b->c->m->OnRead)) {
b->c->m->OnRead(b->c->arm9,a,v,b->c->cpu->R[cast<int64_t>(15ULL)]);
}
}
}
// tools/platform/nds/dsmachine/bus.go:212:1
void dsmachine_bus_Write16Reference(dsmachine_bus* b,uint32_t a,uint16_t v){
{
dsmachine_bus_Write(b,a,cast<uint8_t>(v));
dsmachine_bus_Write(b,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/platform/nds/dsmachine/bus.go:217:1
void dsmachine_bus_Write32Reference(dsmachine_bus* b,uint32_t a,uint32_t v){
{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
dsmachine_bus_Write(b,cast<uint32_t>((a + i)),cast<uint8_t>(shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(8ULL) * i))))));
}
}}
}
// tools/platform/nds/dsmachine/bus.go:225:1
uint32_t dsmachine_bus_r32(dsmachine_bus* b,uint32_t a){
{
return dsmachine_bus_Read32(b,a);
}
}
// tools/platform/nds/dsmachine/bus.go:226:1
void dsmachine_bus_w32(dsmachine_bus* b,uint32_t a,uint32_t v){
{
dsmachine_bus_Write32(b,a,v);
}
}
// tools/platform/nds/dsmachine/card.go:63:1
int64_t dsmachine_blockBytes(uint32_t ctrl){
{
{
uint32_t n = cast<uint32_t>(((shr<uint32_t>(ctrl,cast<int64_t>(24ULL))) & cast<uint32_t>(7ULL)));
switch(n){
case cast<uint32_t>(0ULL):{
return cast<int64_t>(0ULL);
break;}
case cast<uint32_t>(7ULL):{
return cast<int64_t>(4ULL);
break;}
default:{
return shl<int64_t>(cast<int64_t>(256ULL),n);
break;}
}}
}
}
// tools/platform/nds/dsmachine/card.go:75:1
void dsmachine_card_start(dsmachine_card* cd,uint32_t ctrl){
{
cd->ctrl = ctrl;
int64_t n = dsmachine_blockBytes(ctrl);
cd->pos = cast<int64_t>(0ULL);
cd->buf = sub(cd->buf,0,cast<int64_t>(0ULL));
if ((n == cast<int64_t>(0ULL))) {
cd->ctrl &= ~(cast<uint32_t>(2155872256ULL));
return ;
}
{
switch(cd->cmd[cast<int64_t>(0ULL)]){
case cast<uint8_t>(183ULL):{
uint32_t addr = be_Uint32(sub(cd->cmd,cast<int64_t>(1ULL),cast<int64_t>(5ULL)));
if (bool(cd->OnXfer)) {
cd->OnXfer(cd->cmd,addr,n);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
uint32_t a = cast<uint32_t>(((addr & ~(cast<uint32_t>(4095ULL))) | cast<uint32_t>(((cast<uint32_t>((addr + cast<uint32_t>(i)))) & cast<uint32_t>(4095ULL)))));
cd->buf = append(cd->buf,dsmachine_card_romByte(cd,a));
}
}break;}
case cast<uint8_t>(184ULL):{
Slice<uint8_t> id = Slice<uint8_t>{cast<uint8_t>(194ULL),cast<uint8_t>(127ULL),cast<uint8_t>(63ULL),cast<uint8_t>(0ULL)};
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
cd->buf = append(cd->buf,id[cast<int64_t>((i & cast<int64_t>(3ULL)))]);
}
}break;}
default:{
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
cd->buf = append(cd->buf,cast<uint8_t>(255ULL));
}
}break;}
}}
cd->ctrl |= cast<uint32_t>(2155872256ULL);
}
}
// tools/platform/nds/dsmachine/card.go:114:1
uint8_t dsmachine_card_romByte(dsmachine_card* cd,uint32_t a){
{
if ((cast<int64_t>(a) < len(cd->rom))) {
return cd->rom[a];
}
return cast<uint8_t>(255ULL);
}
}
// tools/platform/nds/dsmachine/card.go:123:1
uint32_t dsmachine_core_cardReadData(dsmachine_core* c){
{
dsmachine_card* cd = c->m->cd;
if ((cd->pos >= len(cd->buf))) {
return cast<uint32_t>(4294967295ULL);
}
uint32_t v = le_Uint32(sub(cd->buf,cd->pos,len(cd->buf)));
cd->pos += cast<int64_t>(4ULL);
if ((cd->pos >= len(cd->buf))) {
cd->ctrl &= ~(cast<uint32_t>(2155872256ULL));
if ((cast<uint32_t>((get(c->io,cast<uint32_t>(67109280ULL)) & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(c,cast<uint32_t>(524288ULL));
}
}
return v;
}
}
// tools/platform/nds/dsmachine/card.go:144:1
dsmachine_core* dsmachine_Machine_cardCore(dsmachine_Machine* m){
{
if ((cast<uint32_t>((get(m->ARM9->io,cast<uint32_t>(67109380ULL)) & cast<uint32_t>(2048ULL))) == cast<uint32_t>(0ULL))) {
return m->ARM9;
}
return m->ARM7;
}
}
// tools/platform/nds/dsmachine/debug.go:81:1
int64_t dsmachine_Machine_GXCommandsRun(dsmachine_Machine* m){
{
return m->gpu3d->count;
}
}
// tools/platform/nds/dsmachine/debug.go:89:1
void dsmachine_Machine_ForceRender(dsmachine_Machine* m){
{
dsmachine_gpu3d_render(m->gpu3d,m);
dsmachine_gpu2d_render(m->gpu2d,m);
}
}
// tools/platform/nds/dsmachine/debug.go:96:1
void dsmachine_Machine_StopRequested(dsmachine_Machine* m){
{
m->stop = true;
}
}
// tools/platform/nds/dsmachine/debug.go:100:1
void dsmachine_Machine_AddBreakpoint(dsmachine_Machine* m,uint32_t pc){
{
if ((!m->bps)) {
m->bps = Map<uint32_t,bool>{};
}
m->bps[pc] = true;
}
}
// tools/platform/nds/dsmachine/debug.go:107:1
void dsmachine_Machine_ClearBreakpoint(dsmachine_Machine* m,uint32_t pc){
{
removeKey(m->bps,pc);
}
}
// tools/platform/nds/dsmachine/debug.go:108:1
void dsmachine_Machine_ClearBreakpoints(dsmachine_Machine* m){
{
m->bps = {};
}
}
// tools/platform/nds/dsmachine/debug.go:109:1
std::tuple<bool,uint32_t> dsmachine_Machine_Stopped(dsmachine_Machine* m){
{
return {m->stopped,m->stoppedPC};
}
}
// tools/platform/nds/dsmachine/debug.go:110:1
int64_t dsmachine_Machine_StepInstructions(dsmachine_Machine* m,int64_t n){
{
return dsmachine_Machine_runInstrs(m,n);
}
}
// tools/platform/nds/dsmachine/debug.go:123:1
Slice<dsmachine_MemRegion> dsmachine_Machine_MemRegions(dsmachine_Machine* m){
{
Slice<dsmachine_MemRegion> r = Slice<dsmachine_MemRegion>{dsmachine_MemRegion{std::string("main RAM",8),cast<uint32_t>(33554432ULL),cast<uint32_t>(4194304ULL)},dsmachine_MemRegion{std::string("shared WRAM",11),cast<uint32_t>(50331648ULL),cast<uint32_t>(32768ULL)},dsmachine_MemRegion{std::string("ARM7 WRAM",9),cast<uint32_t>(58720256ULL),cast<uint32_t>(65536ULL)},dsmachine_MemRegion{std::string("I/O",3),cast<uint32_t>(67108864ULL),cast<uint32_t>(1048576ULL)},dsmachine_MemRegion{std::string("palette",7),cast<uint32_t>(83886080ULL),cast<uint32_t>(2048ULL)},dsmachine_MemRegion{std::string("VRAM (engine A BG)",18),cast<uint32_t>(100663296ULL),cast<uint32_t>(524288ULL)},dsmachine_MemRegion{std::string("VRAM (engine B BG)",18),cast<uint32_t>(102760448ULL),cast<uint32_t>(131072ULL)},dsmachine_MemRegion{std::string("VRAM (engine A OBJ)",19),cast<uint32_t>(104857600ULL),cast<uint32_t>(262144ULL)},dsmachine_MemRegion{std::string("VRAM (engine B OBJ)",19),cast<uint32_t>(106954752ULL),cast<uint32_t>(131072ULL)},dsmachine_MemRegion{std::string("VRAM (LCDC window)",18),cast<uint32_t>(109051904ULL),cast<uint32_t>(671744ULL)},dsmachine_MemRegion{std::string("OAM",3),cast<uint32_t>(117440512ULL),cast<uint32_t>(2048ULL)}};
if (bool(m->ARM9->itcm)) {
r = append(r,dsmachine_MemRegion{std::string("ITCM",4),m->ARM9->itcmBase,cast<uint32_t>(len(m->ARM9->itcm))});
}
if (bool(m->ARM9->dtcm)) {
r = append(r,dsmachine_MemRegion{std::string("DTCM",4),m->ARM9->dtcmBase,cast<uint32_t>(len(m->ARM9->dtcm))});
}
return r;
}
}
// tools/platform/nds/dsmachine/debug.go:155:1
Slice<std::string> dsmachine_TextureFormats(){
{
return append(Slice<std::string>{},dsmachine_textureFormats);
}
}
// tools/platform/nds/dsmachine/debug.go:158:1
std::tuple<uint32_t,bool> dsmachine_TextureFormat(std::string name){
{
{auto&& tmp22 = dsmachine_textureFormats;
for(int64_t tmp23=0;tmp23<len(tmp22);++tmp23){
auto i=tmp23;auto f=tmp22[tmp23];if ((f == name)) {
return {cast<uint32_t>(i),true};
}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/nds/dsmachine/debug.go:176:1
std::tuple<image_RGBA*,Error> dsmachine_Machine_RenderTexture(dsmachine_Machine* m,uint32_t offset,uint32_t palBase,uint32_t format,uint32_t w,uint32_t h){
{
if ((cast<int64_t>(format) >= len(dsmachine_textureFormats))) {
return {{},dsmachine_errf(std::string("dsmachine: texture format %d is not a DS format",47),format)};
}
if (((((w == cast<uint32_t>(0ULL)) || (h == cast<uint32_t>(0ULL))) || (w > cast<uint32_t>(1024ULL))) || (h > cast<uint32_t>(1024ULL)))) {
return {{},dsmachine_errf(std::string("dsmachine: texture size %dx%d out of range",42),w,h)};
}
dsmachine_texState st = dsmachine_texState{offset,palBase,cast<int64_t>(w),cast<int64_t>(h),format,true,true,{},{},{}};
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(w),cast<int64_t>(h)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(h));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(w));x++){
auto tmp24 = dsmachine_gpu3d_sampleTex(m->gpu3d,m,(&st),x,y);
uint8_t r = std::get<0>(tmp24);
uint8_t g = std::get<1>(tmp24);
uint8_t b = std::get<2>(tmp24);
uint8_t a = std::get<3>(tmp24);
bool ok = std::get<4>(tmp24);
color_RGBA c = color_RGBA{};
if (ok) {
c = color_RGBA{dsmachine_expand6(r),dsmachine_expand6(g),dsmachine_expand6(b),dsmachine_expand5(a)};
}
image_RGBA_Set(img,x,y,c);
}
}}
}return {img,{}};
}
}
// tools/platform/nds/dsmachine/debug.go:212:1
image_RGBA* dsmachine_Machine_RenderDepth(dsmachine_Machine* m){
{
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(256ULL),cast<int64_t>(192ULL)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(192ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint8_t v = cast<uint8_t>(shr<uint32_t>(m->gpu3d->rast.depth[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))],cast<int64_t>(16ULL)));
image_RGBA_Set(img,x,y,color_RGBA{v,v,v,cast<uint8_t>(255ULL)});
}
}}
}return img;
}
}
// tools/platform/nds/dsmachine/debug.go:225:1
Slice<uint8_t> dsmachine_Machine_VRAMBank(dsmachine_Machine* m,int64_t i){
{
if (((i < cast<int64_t>(0ULL)) || (i >= cast<int64_t>(9ULL)))) {
return {};
}
return m->vram->bank[i];
}
}
// tools/platform/nds/dsmachine/debug.go:233:1
std::tuple<std::string,int64_t,uint8_t> dsmachine_Machine_VRAMBankInfo(dsmachine_Machine* m,int64_t i){
std::string name{};
int64_t size{};
uint8_t cnt{};
{
if (((i < cast<int64_t>(0ULL)) || (i >= cast<int64_t>(9ULL)))) {
return {std::string("",0),cast<int64_t>(0ULL),cast<uint8_t>(0ULL)};
}
return {cast<std::string>(cast<int32_t>(cast<int64_t>((cast<int64_t>(65ULL) + i)))),dsmachine_bankSizes[i],m->vram->cnt[i]};
}
}
// tools/platform/nds/dsmachine/debug.go:243:1
uint8_t dsmachine_expand6(uint8_t v){
{
if ((v > cast<uint8_t>(63ULL))) {
v = cast<uint8_t>(63ULL);
}
return cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(2ULL)) | shr<uint8_t>(v,cast<int64_t>(4ULL))));
}
}
// tools/platform/nds/dsmachine/debug.go:250:1
uint8_t dsmachine_expand5(uint8_t v){
{
if ((v > cast<uint8_t>(31ULL))) {
v = cast<uint8_t>(31ULL);
}
return cast<uint8_t>((shl<uint8_t>(v,cast<int64_t>(3ULL)) | shr<uint8_t>(v,cast<int64_t>(2ULL))));
}
}
// tools/platform/nds/dsmachine/debug.go:265:1
Slice<bool> dsmachine_Machine_ThreeDVisible(dsmachine_Machine* m){
{
return sub(m->gpu2d->a.vis3D,0,len(m->gpu2d->a.vis3D));
}
}
// tools/platform/nds/dsmachine/debug.go:272:1
bool dsmachine_Machine_EngineAOnTop(dsmachine_Machine* m){
{
return (cast<uint32_t>((m->powcnt & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/debug.go:281:1
Error dsmachine_Machine_Draw3DInto(dsmachine_Machine* m,image_RGBA* img){
{
Slice<uint32_t> px = m->gpu2d->threeD;
if ((len(px) < cast<int64_t>(49152ULL))) {
return dsmachine_errf(std::string("dsmachine: the 3D engine has not drawn a frame",46));
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(192ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint32_t v = px[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))];
image_RGBA_Set(img,x,y,color_RGBA{cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v)});
}
}}
}return {};
}
}
// tools/platform/nds/dsmachine/dma.go:41:1
int64_t dsmachine_core_dmaMode(dsmachine_core* c,dsmachine_dmaChan* ch){
{
if (c->arm9) {
return cast<int64_t>((cast<int64_t>(shr<uint32_t>(ch->ctrl,cast<int64_t>(11ULL))) & cast<int64_t>(7ULL)));
}
{
switch(cast<uint32_t>(((shr<uint32_t>(ch->ctrl,cast<int64_t>(12ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return cast<int64_t>(0ULL);
break;}
case cast<uint32_t>(1ULL):{
return cast<int64_t>(1ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<int64_t>(5ULL);
break;}
default:{
return cast<int64_t>(6ULL);
break;}
}}
}
}
// tools/platform/nds/dsmachine/dma.go:61:1
uint32_t dsmachine_core_maxCount(dsmachine_core* c){
{
if (c->arm9) {
return cast<uint32_t>(2097151ULL);
}
return cast<uint32_t>(65535ULL);
}
}
// tools/platform/nds/dsmachine/dma.go:70:1
void dsmachine_core_writeDMACtrl(dsmachine_core* c,int64_t n){
{
dsmachine_dmaChan* ch = (&c->dma[n]);
if ((cast<uint32_t>((ch->ctrl & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL))) {
ch->active = false;
return ;
}
if (ch->active) {
return ;
}
ch->active = true;
auto tmp25 = std::make_tuple(ch->src,ch->dst);
ch->csrc = std::get<0>(tmp25);
ch->cdst = std::get<1>(tmp25);
ch->crem = ch->count;
if ((ch->crem == cast<uint32_t>(0ULL))) {
ch->crem = dsmachine_core_maxCount(c);
}
if ((dsmachine_core_dmaMode(c,ch) == cast<int64_t>(0ULL))) {
dsmachine_core_runDMAChan(c,n);
}
}
}
// tools/platform/nds/dsmachine/dma.go:94:1
void dsmachine_Machine_runDMA(dsmachine_Machine* m,int64_t mode){
{
{auto&& tmp26 = dsmachine_Machine_cores(m);
for(int64_t tmp27=0;tmp27<len(tmp26);++tmp27){
auto c=tmp26[tmp27];{auto&& tmp28 = c->dma;
for(int64_t tmp29=0;tmp29<len(tmp28);++tmp29){
auto n=tmp29;if ((c->dma[n].active && (dsmachine_core_dmaMode(c,(&c->dma[n])) == mode))) {
dsmachine_core_runDMAChan(c,n);
}
}}
}}
}
}
// tools/platform/nds/dsmachine/dma.go:110:1
void dsmachine_core_runDMAChan(dsmachine_core* c,int64_t n){
rrprof::Scope measured(4,"DMA");
{
if (c->m->prof.on) {
time_Time t0 = go_time_Now();
auto tmp30=defer([&](){[&]()->void{
c->m->prof.dma += go_time_Since(t0);
}
();});
}
c->m->prof.xfers++;
dsmachine_dmaChan* ch = (&c->dma[n]);
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
bool word = (cast<uint32_t>((ch->ctrl & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL));
uint32_t dstMode = cast<uint32_t>(((shr<uint32_t>(ch->ctrl,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
uint32_t srcMode = cast<uint32_t>(((shr<uint32_t>(ch->ctrl,cast<int64_t>(7ULL))) & cast<uint32_t>(3ULL)));
uint32_t unit = cast<uint32_t>(2ULL);
if (word) {
unit = cast<uint32_t>(4ULL);
}
uint32_t todo = ch->crem;
if ((c->arm9 && (dsmachine_core_dmaMode(c,ch) == cast<int64_t>(7ULL)))) {
if ((!dsmachine_gpu3d_fifoBelowHalf(c->m->gpu3d))) {
return ;
}
if ((todo > cast<uint32_t>(112ULL))) {
todo = cast<uint32_t>(112ULL);
}
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < todo);i++){
if (word) {
dsmachine_bus_w32(b,ch->cdst,dsmachine_bus_r32(b,ch->csrc));
}
else {
uint32_t v = cast<uint32_t>((cast<uint32_t>(dsmachine_bus_Read(b,ch->csrc)) | shl<uint32_t>(cast<uint32_t>(dsmachine_bus_Read(b,cast<uint32_t>((ch->csrc + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
dsmachine_bus_Write(b,ch->cdst,cast<uint8_t>(v));
dsmachine_bus_Write(b,cast<uint32_t>((ch->cdst + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
}
{
switch(srcMode){
case cast<uint32_t>(0ULL):{
ch->csrc += unit;
break;}
case cast<uint32_t>(1ULL):{
ch->csrc -= unit;
break;}
}}
{
switch(dstMode){
case cast<uint32_t>(0ULL):case cast<uint32_t>(3ULL):{
ch->cdst += unit;
break;}
case cast<uint32_t>(1ULL):{
ch->cdst -= unit;
break;}
}}
}
}ch->crem -= todo;
if ((ch->crem > cast<uint32_t>(0ULL))) {
return ;
}
if ((cast<uint32_t>((ch->ctrl & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(c,shl<uint32_t>(cast<uint32_t>(256ULL),cast<uint64_t>(n)));
}
bool repeat = (cast<uint32_t>((ch->ctrl & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
if ((repeat && (dsmachine_core_dmaMode(c,ch) != cast<int64_t>(0ULL)))) {
ch->crem = ch->count;
if ((ch->crem == cast<uint32_t>(0ULL))) {
ch->crem = dsmachine_core_maxCount(c);
}
if ((dstMode == cast<uint32_t>(3ULL))) {
ch->cdst = ch->dst;
}
return ;
}
ch->active = false;
ch->ctrl &= ~(cast<uint32_t>(32768ULL));
}
}
// tools/platform/nds/dsmachine/dma.go:189:1
void dsmachine_Machine_gxfifoDrained(dsmachine_Machine* m){
{
dsmachine_core* c = m->ARM9;
{auto&& tmp31 = c->dma;
for(int64_t tmp32=0;tmp32<len(tmp31);++tmp32){
auto n=tmp32;if ((c->dma[n].active && (dsmachine_core_dmaMode(c,(&c->dma[n])) == cast<int64_t>(7ULL)))) {
dsmachine_core_runDMAChan(c,n);
}
}}
}
}
// tools/platform/nds/dsmachine/gpu2d.go:157:1
dsmachine_gpu2d* dsmachine_newGPU2D(){
{
dsmachine_gpu2d* g = arenaNew(dsmachine_gpu2d{});
g->a = dsmachine_engine{{},false,cast<uint32_t>(67108864ULL),cast<uint32_t>(0ULL),cast<uint32_t>(512ULL),cast<uint32_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(2ULL),cast<int64_t>(6ULL),cast<int64_t>(8ULL),Slice<uint32_t>::make(cast<int64_t>(49152ULL)),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
g->b = dsmachine_engine{{},true,cast<uint32_t>(67112960ULL),cast<uint32_t>(1024ULL),cast<uint32_t>(1536ULL),cast<uint32_t>(1024ULL),cast<int64_t>(1ULL),cast<int64_t>(3ULL),cast<int64_t>(7ULL),cast<int64_t>(9ULL),Slice<uint32_t>::make(cast<int64_t>(49152ULL)),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
return g;
}
}
// tools/platform/nds/dsmachine/gpu2d.go:178:1
void dsmachine_gpu2d_beginFrame(dsmachine_gpu2d* g,dsmachine_Machine* m){
{
dsmachine_gpu2d_render(g,m);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:181:1
void dsmachine_gpu2d_render(dsmachine_gpu2d* g,dsmachine_Machine* m){
rrprof::Scope measured(3,"2D composition");
{
if (m->prof.on) {
time_Time t0 = go_time_Now();
auto tmp33=defer([&](){[&]()->void{
m->prof.compose += go_time_Since(t0);
}
();});
}
g->swap = (cast<uint32_t>((m->powcnt & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
g->a.threeD = g->threeD;
dsmachine_engine_frame(&(g->a),m);
dsmachine_engine_frame(&(g->b),m);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:196:1
std::tuple<Slice<uint32_t>,Slice<uint32_t>> dsmachine_gpu2d_screens(dsmachine_gpu2d* g){
Slice<uint32_t> top{};
Slice<uint32_t> bottom{};
{
if (g->swap) {
return {g->a.out,g->b.out};
}
return {g->b.out,g->a.out};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:209:1
uint32_t dsmachine_engine_reg32(dsmachine_engine* e,uint32_t off){
{
return get(e->m->ARM9->io,cast<uint32_t>((e->ioBase + off)));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:211:1
uint16_t dsmachine_engine_reg16(dsmachine_engine* e,uint32_t off){
{
uint32_t w = get(e->m->ARM9->io,cast<uint32_t>((e->ioBase + ((off & ~(cast<uint32_t>(3ULL)))))));
if ((cast<uint32_t>((off & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint16_t>(shr<uint32_t>(w,cast<int64_t>(16ULL)));
}
return cast<uint16_t>(w);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:224:1
uint32_t dsmachine_bgr555RGBA(uint16_t c){
{
uint32_t r = cast<uint32_t>(cast<uint16_t>((c & cast<uint16_t>(31ULL))));
uint32_t g = cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(c,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)));
uint32_t b = cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(c,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)));
r = cast<uint32_t>((shl<uint32_t>(r,cast<int64_t>(3ULL)) | shr<uint32_t>(r,cast<int64_t>(2ULL))));
g = cast<uint32_t>((shl<uint32_t>(g,cast<int64_t>(3ULL)) | shr<uint32_t>(g,cast<int64_t>(2ULL))));
b = cast<uint32_t>((shl<uint32_t>(b,cast<int64_t>(3ULL)) | shr<uint32_t>(b,cast<int64_t>(2ULL))));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(r,cast<int64_t>(24ULL)) | shl<uint32_t>(g,cast<int64_t>(16ULL)))) | shl<uint32_t>(b,cast<int64_t>(8ULL)))) | cast<uint32_t>(255ULL)));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:234:1
uint16_t dsmachine_rgb555(uint32_t r,uint32_t g,uint32_t b){
{
return cast<uint16_t>(cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((r & cast<uint32_t>(31ULL))) | shl<uint32_t>((cast<uint32_t>((g & cast<uint32_t>(31ULL)))),cast<int64_t>(5ULL)))) | shl<uint32_t>((cast<uint32_t>((b & cast<uint32_t>(31ULL)))),cast<int64_t>(10ULL)))));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:236:1
uint32_t dsmachine_chan5(uint16_t c,uint64_t n){
{
return cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(c,(cast<uint64_t>((cast<uint64_t>(5ULL) * n))))) & cast<uint32_t>(31ULL)));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:241:1
uint16_t dsmachine_engine_palBGColor(dsmachine_engine* e,int64_t i){
{
return dsmachine_palAt(e->m->pal,cast<uint32_t>((e->palBG + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(2ULL))))));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:242:1
uint16_t dsmachine_engine_palOBJColor(dsmachine_engine* e,int64_t i){
{
return dsmachine_palAt(e->m->pal,cast<uint32_t>((e->palOBJ + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(2ULL))))));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:244:1
uint16_t dsmachine_palAt(Slice<uint8_t> pal,uint32_t off){
{
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(1ULL))) >= len(pal))) {
return cast<uint16_t>(0ULL);
}
return cast<uint16_t>((cast<uint16_t>(pal[off]) | shl<uint16_t>(cast<uint16_t>(pal[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:255:1
uint16_t dsmachine_engine_bgExtColor(dsmachine_engine* e,int64_t slot,int64_t pal,int64_t idx){
{
uint32_t off = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(slot) * cast<uint32_t>(8192ULL))) + cast<uint32_t>((cast<uint32_t>(pal) * cast<uint32_t>(512ULL))))) + cast<uint32_t>((cast<uint32_t>(idx) * cast<uint32_t>(2ULL)))));
return dsmachine_vram_read16(e->m->vram,e->spBGExt,off);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:262:1
uint16_t dsmachine_engine_objExtColor(dsmachine_engine* e,int64_t pal,int64_t idx){
{
return dsmachine_vram_read16(e->m->vram,e->spOBJExt,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(pal) * cast<uint32_t>(512ULL))) + cast<uint32_t>((cast<uint32_t>(idx) * cast<uint32_t>(2ULL))))));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:268:1
void dsmachine_engine_frame(dsmachine_engine* e,dsmachine_Machine* m){
{
e->m = m;
uint32_t powerBit = cast<uint32_t>(2ULL);
if (e->isB) {
powerBit = cast<uint32_t>(512ULL);
}
if ((cast<uint32_t>((m->powcnt & powerBit)) == cast<uint32_t>(0ULL))) {
dsmachine_engine_fillWhite(e);
return ;
}
uint32_t dispcnt = dsmachine_engine_reg32(e,cast<uint32_t>(0ULL));
if ((cast<uint32_t>((dispcnt & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_engine_fillWhite(e);
return ;
}
{
switch(cast<uint32_t>(((shr<uint32_t>(dispcnt,cast<int64_t>(16ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
dsmachine_engine_fillWhite(e);
return ;
break;}
case cast<uint32_t>(1ULL):{
dsmachine_engine_graphics(e,dispcnt);
break;}
case cast<uint32_t>(2ULL):{
if (e->isB) {
dsmachine_Machine_note(m,std::string("2D engine B: display mode 2 (VRAM display) does not exist on engine B",69));
dsmachine_engine_fillWhite(e);
return ;
}
dsmachine_engine_vramDisplay(e,dispcnt);
break;}
case cast<uint32_t>(3ULL):{
dsmachine_Machine_note(m,std::string("2D engine %s: display mode 3 (main-memory display) not implemented",66),dsmachine_engine_name(e));
dsmachine_engine_fillBlack(e);
return ;
break;}
}}
if ((cast<uint32_t>((dsmachine_engine_reg32(e,cast<uint32_t>(100ULL)) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_Machine_note(m,std::string("2D engine %s: display capture (DISPCAPCNT) not implemented",58),dsmachine_engine_name(e));
}
dsmachine_engine_masterBright(e);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:323:1
std::string dsmachine_engine_name(dsmachine_engine* e){
{
if (e->isB) {
return std::string("B",1);
}
return std::string("A",1);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:330:1
void dsmachine_engine_fillWhite(dsmachine_engine* e){
if(rrcapture::trace.active)rrds::event(e,"fillWhite");
{
{auto&& tmp34 = e->out;
for(int64_t tmp35=0;tmp35<len(tmp34);++tmp35){
auto i=tmp35;e->out[i] = cast<uint32_t>(4294967295ULL);
rrds::write(e,i, e->out[i]);
}}
}
}
// tools/platform/nds/dsmachine/gpu2d.go:336:1
void dsmachine_engine_fillBlack(dsmachine_engine* e){
if(rrcapture::trace.active)rrds::event(e,"fillBlack");
{
{auto&& tmp36 = e->out;
for(int64_t tmp37=0;tmp37<len(tmp36);++tmp37){
auto i=tmp37;e->out[i] = cast<uint32_t>(255ULL);
rrds::write(e,i, e->out[i]);
}}
}
}
// tools/platform/nds/dsmachine/gpu2d.go:347:1
void dsmachine_engine_vramDisplay(dsmachine_engine* e,uint32_t dispcnt){
{
Slice<uint8_t> bank = e->m->vram->bank[cast<uint32_t>(((shr<uint32_t>(dispcnt,cast<int64_t>(18ULL))) & cast<uint32_t>(3ULL)))];
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(192ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t off = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))) * cast<int64_t>(2ULL)));
uint16_t c={};
if ((cast<int64_t>((off + cast<int64_t>(1ULL))) < len(bank))) {
c = cast<uint16_t>((cast<uint16_t>(bank[off]) | shl<uint16_t>(cast<uint16_t>(bank[cast<int64_t>((off + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
e->line[x] = c;
}
}dsmachine_engine_emit(e,y);
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:364:1
void dsmachine_engine_graphics(dsmachine_engine* e,uint32_t dispcnt){
{
int64_t mode = cast<int64_t>(cast<uint32_t>((dispcnt & cast<uint32_t>(7ULL))));
if ((mode == cast<int64_t>(7ULL))) {
dsmachine_Machine_note(e->m,std::string("2D engine %s: BG mode 7 does not exist",38),dsmachine_engine_name(e));
}
if (((mode == cast<int64_t>(6ULL)) && e->isB)) {
dsmachine_Machine_note(e->m,std::string("2D engine B: BG mode 6 (large bitmap) does not exist on engine B",64));
}
e->is3D = ((!e->isB) && (cast<uint32_t>((dispcnt & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)));
if ((e->is3D && (!e->threeD))) {
dsmachine_Machine_note(e->m,std::string("2D engine A: BG0 is the 3D layer but the 3D engine has not drawn a frame",72));
}
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(4ULL));n++){
uint16_t cnt = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
e->prio[n] = cast<uint8_t>(cast<uint16_t>((cnt & cast<uint16_t>(3ULL))));
e->shown[n] = (cast<uint32_t>((dispcnt & (shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((cast<uint64_t>(8ULL) + cast<uint64_t>(n)))))))) != cast<uint32_t>(0ULL));
}
}{auto&& tmp38 = e->vis3D;
for(int64_t tmp39=0;tmp39<len(tmp38);++tmp39){
auto i=tmp39;e->vis3D[i] = false;
}}
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(192ULL));y++){
dsmachine_engine_clearLayers(e);
if (((cast<uint32_t>((dispcnt & cast<uint32_t>(4096ULL))) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((dispcnt & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)))) {
dsmachine_engine_sprites(e,dispcnt,y);
}
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(4ULL));n++){
if ((!e->shown[n])) {
continue;
}
dsmachine_engine_background(e,dispcnt,mode,n,y);
}
}dsmachine_engine_windowMask(e,dispcnt,y);
dsmachine_engine_composite(e,y);
dsmachine_engine_emit(e,y);
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:408:1
void dsmachine_engine_clearLayers(dsmachine_engine* e){
{
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(4ULL));n++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
e->bgOK[n][x] = false;
}
}}
}{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
e->a3D[x] = cast<uint8_t>(0ULL);
e->objOK[x] = false;
e->objSemi[x] = false;
e->objEVA[x] = cast<uint8_t>(0ULL);
e->objWin[x] = false;
e->objPrio[x] = cast<uint8_t>(3ULL);
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:426:1
void dsmachine_engine_emit(dsmachine_engine* e,int64_t y){
if(rrcapture::trace.active)rrds::event(e,"emit");
{
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
e->out[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))] = dsmachine_bgr555RGBA(e->line[x]);
rrds::write(e,cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x)), e->out[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))]);
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:438:1
void dsmachine_engine_masterBright(dsmachine_engine* e){
if(rrcapture::trace.active)rrds::event(e,"masterBright");
{
uint16_t v = dsmachine_engine_reg16(e,cast<uint32_t>(108ULL));
uint16_t mode = cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(14ULL))) & cast<uint16_t>(3ULL)));
if (((mode == cast<uint16_t>(0ULL)) || (mode == cast<uint16_t>(3ULL)))) {
return ;
}
uint32_t f = cast<uint32_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL))));
if ((f > cast<uint32_t>(16ULL))) {
f = cast<uint32_t>(16ULL);
}
if ((f == cast<uint32_t>(0ULL))) {
return ;
}
{auto&& tmp40 = e->out;
for(int64_t tmp41=0;tmp41<len(tmp40);++tmp41){
auto i=tmp41;auto p=tmp40[tmp41];auto tmp42 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(p,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(p,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(p,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))));
uint32_t r = std::get<0>(tmp42);
uint32_t g = std::get<1>(tmp42);
uint32_t b = std::get<2>(tmp42);
if ((mode == cast<uint16_t>(1ULL))) {
r += divi<uint32_t>(cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(255ULL) - r))) * f)),cast<uint32_t>(16ULL));
g += divi<uint32_t>(cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(255ULL) - g))) * f)),cast<uint32_t>(16ULL));
b += divi<uint32_t>(cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(255ULL) - b))) * f)),cast<uint32_t>(16ULL));
}
else {
r -= divi<uint32_t>(cast<uint32_t>((r * f)),cast<uint32_t>(16ULL));
g -= divi<uint32_t>(cast<uint32_t>((g * f)),cast<uint32_t>(16ULL));
b -= divi<uint32_t>(cast<uint32_t>((b * f)),cast<uint32_t>(16ULL));
}
e->out[i] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(r,cast<int64_t>(24ULL)) | shl<uint32_t>(g,cast<int64_t>(16ULL)))) | shl<uint32_t>(b,cast<int64_t>(8ULL)))) | cast<uint32_t>(255ULL)));
rrds::write(e,i, e->out[i]);
}}
}
}
// tools/platform/nds/dsmachine/gpu2d.go:468:1
void dsmachine_engine_background(dsmachine_engine* e,uint32_t dispcnt,int64_t mode,int64_t n,int64_t y){
{
if (((n == cast<int64_t>(0ULL)) && e->is3D)) {
dsmachine_engine_threeDLine(e,y);
return ;
}
if (((n == cast<int64_t>(0ULL)) && (mode == cast<int64_t>(6ULL)))) {
return ;
}
{
switch(dsmachine_bgKind[mode][n]){
case cast<int64_t>(1ULL):{
dsmachine_engine_textBG(e,dispcnt,n,y);
break;}
case cast<int64_t>(2ULL):{
dsmachine_engine_affineBG(e,dispcnt,n,y);
break;}
case cast<int64_t>(3ULL):{
dsmachine_engine_extendedBG(e,dispcnt,n,y);
break;}
case cast<int64_t>(4ULL):{
if ((!e->isB)) {
dsmachine_engine_largeBG(e,n,y);
}
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu2d.go:495:1
void dsmachine_engine_threeDLine(dsmachine_engine* e,int64_t y){
{
if ((!e->threeD)) {
return ;
}
int64_t hofs = cast<int64_t>((cast<int64_t>(dsmachine_engine_reg16(e,cast<uint32_t>(16ULL))) & cast<int64_t>(511ULL)));
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint32_t p = e->threeD[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + cast<int64_t>(((cast<int64_t>((x + hofs))) & cast<int64_t>(255ULL)))))];
uint8_t a = cast<uint8_t>(cast<uint32_t>((p & cast<uint32_t>(255ULL))));
if ((a == cast<uint8_t>(0ULL))) {
continue;
}
auto tmp43 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(p,cast<int64_t>(24ULL))) & cast<uint32_t>(255ULL))),cast<uint32_t>(((shr<uint32_t>(p,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL))),cast<uint32_t>(((shr<uint32_t>(p,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
uint32_t r = std::get<0>(tmp43);
uint32_t g = std::get<1>(tmp43);
uint32_t b = std::get<2>(tmp43);
e->bg[cast<int64_t>(0ULL)][x] = dsmachine_rgb555(shr<uint32_t>(r,cast<int64_t>(3ULL)),shr<uint32_t>(g,cast<int64_t>(3ULL)),shr<uint32_t>(b,cast<int64_t>(3ULL)));
e->bgOK[cast<int64_t>(0ULL)][x] = true;
e->a3D[x] = cast<uint8_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(a) * cast<uint32_t>(31ULL))),cast<uint32_t>(255ULL)));
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:516:1
std::tuple<int64_t,int64_t> dsmachine_engine_mosaic(dsmachine_engine* e){
int64_t h{};
int64_t v{};
{
uint16_t m = dsmachine_engine_reg16(e,cast<uint32_t>(76ULL));
return {cast<int64_t>((cast<int64_t>(cast<uint16_t>((m & cast<uint16_t>(15ULL)))) + cast<int64_t>(1ULL))),cast<int64_t>((cast<int64_t>(cast<uint16_t>((shr<uint16_t>(m,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL)))) + cast<int64_t>(1ULL)))};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:527:1
std::tuple<uint32_t,uint32_t> dsmachine_engine_bgBases(dsmachine_engine* e,uint32_t dispcnt,uint16_t cnt){
uint32_t charBase{};
uint32_t screenBase{};
{
charBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(2ULL)) & cast<uint16_t>(15ULL)))) * cast<uint32_t>(16384ULL)));
screenBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(8ULL)) & cast<uint16_t>(31ULL)))) * cast<uint32_t>(2048ULL)));
if ((!e->isB)) {
charBase += cast<uint32_t>(((cast<uint32_t>((shr<uint32_t>(dispcnt,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL)))) * cast<uint32_t>(65536ULL)));
screenBase += cast<uint32_t>(((cast<uint32_t>((shr<uint32_t>(dispcnt,cast<int64_t>(27ULL)) & cast<uint32_t>(7ULL)))) * cast<uint32_t>(65536ULL)));
}
return {charBase,screenBase};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:539:1
void dsmachine_engine_textBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y){
{
uint16_t cnt = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
auto tmp44 = dsmachine_engine_bgBases(e,dispcnt,cnt);
uint32_t charBase = std::get<0>(tmp44);
uint32_t screenBase = std::get<1>(tmp44);
int64_t hofs = cast<int64_t>((cast<int64_t>(dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(16ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))))))) & cast<int64_t>(511ULL)));
int64_t vofs = cast<int64_t>((cast<int64_t>(dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(16ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(2ULL))))) & cast<int64_t>(511ULL)));
auto tmp45 = std::make_tuple(cast<int64_t>(256ULL),cast<int64_t>(256ULL));
int64_t w = std::get<0>(tmp45);
int64_t h = std::get<1>(tmp45);
if ((cast<uint16_t>((cnt & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL))) {
w = cast<int64_t>(512ULL);
}
if ((cast<uint16_t>((cnt & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
h = cast<int64_t>(512ULL);
}
bool bpp8 = (cast<uint16_t>((cnt & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
bool extPal = (bpp8 && (cast<uint32_t>((dispcnt & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL)));
int64_t slot = n;
if (((n < cast<int64_t>(2ULL)) && (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL)))) {
slot = cast<int64_t>((n + cast<int64_t>(2ULL)));
}
auto tmp46 = std::make_tuple(cast<int64_t>(1ULL),cast<int64_t>(1ULL));
int64_t mx = std::get<0>(tmp46);
int64_t my = std::get<1>(tmp46);
if ((cast<uint16_t>((cnt & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL))) {
auto tmp47 = dsmachine_engine_mosaic(e);
mx = std::get<0>(tmp47);
my = std::get<1>(tmp47);
}
int64_t sy = y;
if ((my > cast<int64_t>(1ULL))) {
sy -= modi<int64_t>(sy,my);
}
sy = cast<int64_t>(((cast<int64_t>((sy + vofs))) & (cast<int64_t>((h - cast<int64_t>(1ULL))))));
auto tmp48 = std::make_tuple(modi<int64_t>(divi<int64_t>(sy,cast<int64_t>(8ULL)),cast<int64_t>(32ULL)),modi<int64_t>(sy,cast<int64_t>(8ULL)));
int64_t ty = std::get<0>(tmp48);
int64_t py = std::get<1>(tmp48);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t sx = x;
if ((mx > cast<int64_t>(1ULL))) {
sx -= modi<int64_t>(sx,mx);
}
sx = cast<int64_t>(((cast<int64_t>((sx + hofs))) & (cast<int64_t>((w - cast<int64_t>(1ULL))))));
int64_t block = cast<int64_t>((divi<int64_t>(sx,cast<int64_t>(256ULL)) + cast<int64_t>(((divi<int64_t>(sy,cast<int64_t>(256ULL))) * (divi<int64_t>(w,cast<int64_t>(256ULL)))))));
uint32_t off = cast<uint32_t>((cast<uint32_t>((screenBase + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(2048ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((ty * cast<int64_t>(32ULL))) + modi<int64_t>(divi<int64_t>(sx,cast<int64_t>(8ULL)),cast<int64_t>(32ULL))))) * cast<uint32_t>(2ULL)))));
uint16_t ent = dsmachine_vram_read16(e->m->vram,e->spBG,off);
uint32_t tile = cast<uint32_t>(cast<uint16_t>((ent & cast<uint16_t>(1023ULL))));
auto tmp49 = std::make_tuple(modi<int64_t>(sx,cast<int64_t>(8ULL)),py);
int64_t fx = std::get<0>(tmp49);
int64_t fy = std::get<1>(tmp49);
if ((cast<uint16_t>((ent & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL))) {
fx = cast<int64_t>((cast<int64_t>(7ULL) - fx));
}
if ((cast<uint16_t>((ent & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
fy = cast<int64_t>((cast<int64_t>(7ULL) - fy));
}
int64_t idx={};
if (bpp8) {
idx = cast<int64_t>(dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(64ULL))))) + cast<uint32_t>(cast<int64_t>((cast<int64_t>((fy * cast<int64_t>(8ULL))) + fx)))))));
}
else {
uint8_t b = dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(32ULL))))) + cast<uint32_t>(cast<int64_t>((cast<int64_t>((fy * cast<int64_t>(4ULL))) + divi<int64_t>(fx,cast<int64_t>(2ULL))))))));
if ((cast<int64_t>((fx & cast<int64_t>(1ULL))) != cast<int64_t>(0ULL))) {
idx = cast<int64_t>(shr<uint8_t>(b,cast<int64_t>(4ULL)));
}
else {
idx = cast<int64_t>(cast<uint8_t>((b & cast<uint8_t>(15ULL))));
}
}
if ((idx == cast<int64_t>(0ULL))) {
continue;
}
int64_t pal = cast<int64_t>(shr<uint16_t>(ent,cast<int64_t>(12ULL)));
{
if (extPal){
e->bg[n][x] = dsmachine_engine_bgExtColor(e,slot,pal,idx);
}
else if (bpp8){
e->bg[n][x] = dsmachine_engine_palBGColor(e,idx);
}
else {
e->bg[n][x] = dsmachine_engine_palBGColor(e,cast<int64_t>((cast<int64_t>((pal * cast<int64_t>(16ULL))) + idx)));
}
}
tmp50:;
e->bgOK[n][x] = true;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:630:1
std::tuple<int32_t,int32_t,int32_t,int32_t,int32_t,int32_t> dsmachine_engine_affineParams(dsmachine_engine* e,int64_t n){
int32_t pa{};
int32_t pb{};
int32_t pc{};
int32_t pd{};
int32_t x0{};
int32_t y0{};
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(32ULL) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((n - cast<int64_t>(2ULL)))) * cast<uint32_t>(16ULL)))));
pa = cast<int32_t>(cast<int16_t>(dsmachine_engine_reg16(e,base)));
pb = cast<int32_t>(cast<int16_t>(dsmachine_engine_reg16(e,cast<uint32_t>((base + cast<uint32_t>(2ULL))))));
pc = cast<int32_t>(cast<int16_t>(dsmachine_engine_reg16(e,cast<uint32_t>((base + cast<uint32_t>(4ULL))))));
pd = cast<int32_t>(cast<int16_t>(dsmachine_engine_reg16(e,cast<uint32_t>((base + cast<uint32_t>(6ULL))))));
x0 = shr<int32_t>(cast<int32_t>(shl<uint32_t>(dsmachine_engine_reg32(e,cast<uint32_t>((base + cast<uint32_t>(8ULL)))),cast<int64_t>(4ULL))),cast<int64_t>(4ULL));
y0 = shr<int32_t>(cast<int32_t>(shl<uint32_t>(dsmachine_engine_reg32(e,cast<uint32_t>((base + cast<uint32_t>(12ULL)))),cast<int64_t>(4ULL))),cast<int64_t>(4ULL));
return {pa,pb,pc,pd,x0,y0};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:646:1
std::tuple<int32_t,int32_t> dsmachine_affineOrigin(int32_t pb,int32_t pd,int32_t x0,int32_t y0,int64_t y){
{
return {cast<int32_t>((x0 + cast<int32_t>((pb * cast<int32_t>(y))))),cast<int32_t>((y0 + cast<int32_t>((pd * cast<int32_t>(y)))))};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:653:1
void dsmachine_engine_affineBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y){
{
uint16_t cnt = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
auto tmp51 = dsmachine_engine_bgBases(e,dispcnt,cnt);
uint32_t charBase = std::get<0>(tmp51);
uint32_t screenBase = std::get<1>(tmp51);
int64_t size = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL))));
int64_t dim = shl<int64_t>(cast<int64_t>(128ULL),cast<uint64_t>(size));
int64_t tiles = divi<int64_t>(dim,cast<int64_t>(8ULL));
bool wrap = (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
auto tmp52 = dsmachine_engine_affineParams(e,n);
int32_t pa = std::get<0>(tmp52);
int32_t pb = std::get<1>(tmp52);
int32_t pc = std::get<2>(tmp52);
int32_t pd = std::get<3>(tmp52);
int32_t x0 = std::get<4>(tmp52);
int32_t y0 = std::get<5>(tmp52);
auto tmp53 = dsmachine_affineOrigin(pb,pd,x0,y0,y);
int32_t ox = std::get<0>(tmp53);
int32_t oy = std::get<1>(tmp53);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t px = cast<int64_t>(shr<int32_t>((cast<int32_t>((ox + cast<int32_t>((pa * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
int64_t py = cast<int64_t>(shr<int32_t>((cast<int32_t>((oy + cast<int32_t>((pc * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
bool ok={};
auto tmp54 = dsmachine_wrapAffine(px,py,dim,dim,wrap);
px = std::get<0>(tmp54);
py = std::get<1>(tmp54);
ok = std::get<2>(tmp54);
if ((!ok)) {
continue;
}
uint32_t tile = cast<uint32_t>(dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((screenBase + cast<uint32_t>(cast<int64_t>((cast<int64_t>(((divi<int64_t>(py,cast<int64_t>(8ULL))) * tiles)) + divi<int64_t>(px,cast<int64_t>(8ULL)))))))));
uint8_t idx = dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(64ULL))))) + cast<uint32_t>(cast<int64_t>((cast<int64_t>(((modi<int64_t>(py,cast<int64_t>(8ULL))) * cast<int64_t>(8ULL))) + modi<int64_t>(px,cast<int64_t>(8ULL))))))));
if ((idx == cast<uint8_t>(0ULL))) {
continue;
}
e->bg[n][x] = dsmachine_engine_palBGColor(e,cast<int64_t>(idx));
e->bgOK[n][x] = true;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:685:1
std::tuple<int64_t,int64_t,bool> dsmachine_wrapAffine(int64_t x,int64_t y,int64_t w,int64_t h,bool wrap){
{
if (wrap) {
return {modi<int64_t>((cast<int64_t>(((modi<int64_t>(x,w)) + w))),w),modi<int64_t>((cast<int64_t>(((modi<int64_t>(y,h)) + h))),h),true};
}
if (((((x < cast<int64_t>(0ULL)) || (x >= w)) || (y < cast<int64_t>(0ULL))) || (y >= h))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
return {x,y,true};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:699:1
void dsmachine_engine_extendedBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y){
{
uint16_t cnt = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
if ((cast<uint16_t>((cnt & cast<uint16_t>(128ULL))) == cast<uint16_t>(0ULL))) {
dsmachine_engine_extTiledBG(e,dispcnt,n,y,cnt);
return ;
}
dsmachine_engine_extBitmapBG(e,n,y,cnt);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:712:1
void dsmachine_engine_extTiledBG(dsmachine_engine* e,uint32_t dispcnt,int64_t n,int64_t y,uint16_t cnt){
{
auto tmp55 = dsmachine_engine_bgBases(e,dispcnt,cnt);
uint32_t charBase = std::get<0>(tmp55);
uint32_t screenBase = std::get<1>(tmp55);
int64_t size = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL))));
int64_t dim = shl<int64_t>(cast<int64_t>(128ULL),cast<uint64_t>(size));
int64_t tiles = divi<int64_t>(dim,cast<int64_t>(8ULL));
bool wrap = (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
bool extPal = (cast<uint32_t>((dispcnt & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL));
auto tmp56 = dsmachine_engine_affineParams(e,n);
int32_t pa = std::get<0>(tmp56);
int32_t pb = std::get<1>(tmp56);
int32_t pc = std::get<2>(tmp56);
int32_t pd = std::get<3>(tmp56);
int32_t x0 = std::get<4>(tmp56);
int32_t y0 = std::get<5>(tmp56);
auto tmp57 = dsmachine_affineOrigin(pb,pd,x0,y0,y);
int32_t ox = std::get<0>(tmp57);
int32_t oy = std::get<1>(tmp57);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t px = cast<int64_t>(shr<int32_t>((cast<int32_t>((ox + cast<int32_t>((pa * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
int64_t py = cast<int64_t>(shr<int32_t>((cast<int32_t>((oy + cast<int32_t>((pc * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
bool ok={};
auto tmp58 = dsmachine_wrapAffine(px,py,dim,dim,wrap);
px = std::get<0>(tmp58);
py = std::get<1>(tmp58);
ok = std::get<2>(tmp58);
if ((!ok)) {
continue;
}
uint16_t ent = dsmachine_vram_read16(e->m->vram,e->spBG,cast<uint32_t>((screenBase + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>(((divi<int64_t>(py,cast<int64_t>(8ULL))) * tiles)) + divi<int64_t>(px,cast<int64_t>(8ULL))))) * cast<uint32_t>(2ULL))))));
uint32_t tile = cast<uint32_t>(cast<uint16_t>((ent & cast<uint16_t>(1023ULL))));
auto tmp59 = std::make_tuple(modi<int64_t>(px,cast<int64_t>(8ULL)),modi<int64_t>(py,cast<int64_t>(8ULL)));
int64_t fx = std::get<0>(tmp59);
int64_t fy = std::get<1>(tmp59);
if ((cast<uint16_t>((ent & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL))) {
fx = cast<int64_t>((cast<int64_t>(7ULL) - fx));
}
if ((cast<uint16_t>((ent & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
fy = cast<int64_t>((cast<int64_t>(7ULL) - fy));
}
int64_t idx = cast<int64_t>(dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(64ULL))))) + cast<uint32_t>(cast<int64_t>((cast<int64_t>((fy * cast<int64_t>(8ULL))) + fx)))))));
if ((idx == cast<int64_t>(0ULL))) {
continue;
}
if (extPal) {
e->bg[n][x] = dsmachine_engine_bgExtColor(e,n,cast<int64_t>(shr<uint16_t>(ent,cast<int64_t>(12ULL))),idx);
}
else {
e->bg[n][x] = dsmachine_engine_palBGColor(e,idx);
}
e->bgOK[n][x] = true;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:762:1
void dsmachine_engine_extBitmapBG(dsmachine_engine* e,int64_t n,int64_t y,uint16_t cnt){
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(8ULL)) & cast<uint16_t>(31ULL)))) * cast<uint32_t>(16384ULL)));
bool direct = (cast<uint16_t>((cnt & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL));
int64_t size = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL))));
auto tmp60 = std::make_tuple(dsmachine_extBmpSize[size][cast<int64_t>(0ULL)],dsmachine_extBmpSize[size][cast<int64_t>(1ULL)]);
int64_t w = std::get<0>(tmp60);
int64_t h = std::get<1>(tmp60);
bool wrap = (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
auto tmp61 = dsmachine_engine_affineParams(e,n);
int32_t pa = std::get<0>(tmp61);
int32_t pb = std::get<1>(tmp61);
int32_t pc = std::get<2>(tmp61);
int32_t pd = std::get<3>(tmp61);
int32_t x0 = std::get<4>(tmp61);
int32_t y0 = std::get<5>(tmp61);
auto tmp62 = dsmachine_affineOrigin(pb,pd,x0,y0,y);
int32_t ox = std::get<0>(tmp62);
int32_t oy = std::get<1>(tmp62);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t px = cast<int64_t>(shr<int32_t>((cast<int32_t>((ox + cast<int32_t>((pa * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
int64_t py = cast<int64_t>(shr<int32_t>((cast<int32_t>((oy + cast<int32_t>((pc * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
bool ok={};
auto tmp63 = dsmachine_wrapAffine(px,py,w,h,wrap);
px = std::get<0>(tmp63);
py = std::get<1>(tmp63);
ok = std::get<2>(tmp63);
if ((!ok)) {
continue;
}
if (direct) {
uint16_t c = dsmachine_vram_read16(e->m->vram,e->spBG,cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((py * w)) + px))) * cast<uint32_t>(2ULL))))));
if ((cast<uint16_t>((c & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL))) {
continue;
}
e->bg[n][x] = cast<uint16_t>((c & cast<uint16_t>(32767ULL)));
}
else {
uint8_t idx = dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((cast<int64_t>((py * w)) + px))))));
if ((idx == cast<uint8_t>(0ULL))) {
continue;
}
e->bg[n][x] = dsmachine_engine_palBGColor(e,cast<int64_t>(idx));
}
e->bgOK[n][x] = true;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:802:1
void dsmachine_engine_largeBG(dsmachine_engine* e,int64_t n,int64_t y){
{
uint16_t cnt = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
auto tmp64 = std::make_tuple(cast<int64_t>(512ULL),cast<int64_t>(1024ULL));
int64_t w = std::get<0>(tmp64);
int64_t h = std::get<1>(tmp64);
if ((cast<uint16_t>((cnt & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL))) {
auto tmp65 = std::make_tuple(cast<int64_t>(1024ULL),cast<int64_t>(512ULL));
w = std::get<0>(tmp65);
h = std::get<1>(tmp65);
}
bool wrap = (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
auto tmp66 = dsmachine_engine_affineParams(e,n);
int32_t pa = std::get<0>(tmp66);
int32_t pb = std::get<1>(tmp66);
int32_t pc = std::get<2>(tmp66);
int32_t pd = std::get<3>(tmp66);
int32_t x0 = std::get<4>(tmp66);
int32_t y0 = std::get<5>(tmp66);
auto tmp67 = dsmachine_affineOrigin(pb,pd,x0,y0,y);
int32_t ox = std::get<0>(tmp67);
int32_t oy = std::get<1>(tmp67);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
int64_t px = cast<int64_t>(shr<int32_t>((cast<int32_t>((ox + cast<int32_t>((pa * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
int64_t py = cast<int64_t>(shr<int32_t>((cast<int32_t>((oy + cast<int32_t>((pc * cast<int32_t>(x)))))),cast<int64_t>(8ULL)));
bool ok={};
auto tmp68 = dsmachine_wrapAffine(px,py,w,h,wrap);
px = std::get<0>(tmp68);
py = std::get<1>(tmp68);
ok = std::get<2>(tmp68);
if ((!ok)) {
continue;
}
uint8_t idx = dsmachine_vram_read8(e->m->vram,e->spBG,cast<uint32_t>(cast<int64_t>((cast<int64_t>((py * w)) + px))));
if ((idx == cast<uint8_t>(0ULL))) {
continue;
}
e->bg[n][x] = dsmachine_engine_palBGColor(e,cast<int64_t>(idx));
e->bgOK[n][x] = true;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:843:1
void dsmachine_engine_sprites(dsmachine_engine* e,uint32_t dispcnt,int64_t y){
{
{int64_t i = cast<int64_t>(127ULL);for (;(i >= cast<int64_t>(0ULL));i--){
uint32_t o = cast<uint32_t>((e->oamBase + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(8ULL)))));
uint16_t a0 = dsmachine_engine_oam16(e,o);
uint16_t a1 = dsmachine_engine_oam16(e,cast<uint32_t>((o + cast<uint32_t>(2ULL))));
uint16_t a2 = dsmachine_engine_oam16(e,cast<uint32_t>((o + cast<uint32_t>(4ULL))));
uint16_t mode = cast<uint16_t>(((shr<uint16_t>(a0,cast<int64_t>(8ULL))) & cast<uint16_t>(3ULL)));
if ((mode == cast<uint16_t>(2ULL))) {
continue;
}
bool affine = ((mode == cast<uint16_t>(1ULL)) || (mode == cast<uint16_t>(3ULL)));
bool double_ = (mode == cast<uint16_t>(3ULL));
int64_t shape = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a0,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL))));
if ((shape == cast<int64_t>(3ULL))) {
dsmachine_Machine_note(e->m,std::string("2D engine %s: OAM entry %d uses the reserved sprite shape 3",59),dsmachine_engine_name(e),i);
continue;
}
int64_t size = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a1,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL))));
auto tmp69 = std::make_tuple(dsmachine_objSize[shape][size][cast<int64_t>(0ULL)],dsmachine_objSize[shape][size][cast<int64_t>(1ULL)]);
int64_t w = std::get<0>(tmp69);
int64_t h = std::get<1>(tmp69);
auto tmp70 = std::make_tuple(w,h);
int64_t bw = std::get<0>(tmp70);
int64_t bh = std::get<1>(tmp70);
if (double_) {
auto tmp71 = std::make_tuple(cast<int64_t>((w * cast<int64_t>(2ULL))),cast<int64_t>((h * cast<int64_t>(2ULL))));
bw = std::get<0>(tmp71);
bh = std::get<1>(tmp71);
}
int64_t sy = cast<int64_t>(cast<uint16_t>((a0 & cast<uint16_t>(255ULL))));
int64_t dy = cast<int64_t>(((cast<int64_t>((y - sy))) & cast<int64_t>(255ULL)));
if ((dy >= bh)) {
continue;
}
int64_t sx = cast<int64_t>(cast<uint16_t>((a1 & cast<uint16_t>(511ULL))));
if ((sx >= cast<int64_t>(256ULL))) {
sx -= cast<int64_t>(512ULL);
}
if ((cast<uint16_t>((a0 & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL))) {
dsmachine_Machine_note(e->m,std::string("2D engine %s: sprite mosaic (OAM attr0 bit 12) not implemented",62),dsmachine_engine_name(e));
}
uint16_t gfx = cast<uint16_t>(((shr<uint16_t>(a0,cast<int64_t>(10ULL))) & cast<uint16_t>(3ULL)));
bool bpp8 = (cast<uint16_t>((a0 & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
uint8_t prio = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(a2,cast<int64_t>(10ULL)) & cast<uint16_t>(3ULL))));
int64_t pal = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a2,cast<int64_t>(12ULL)) & cast<uint16_t>(15ULL))));
int64_t alpha = pal;
if (((gfx == cast<uint16_t>(3ULL)) && (alpha == cast<int64_t>(0ULL)))) {
continue;
}
int32_t pa={};
int32_t pb={};
int32_t pc={};
int32_t pd={};
if (affine) {
auto tmp72 = dsmachine_engine_oamAffine(e,cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a1,cast<int64_t>(9ULL)) & cast<uint16_t>(31ULL)))));
pa = std::get<0>(tmp72);
pb = std::get<1>(tmp72);
pc = std::get<2>(tmp72);
pd = std::get<3>(tmp72);
}
bool hflip = ((!affine) && (cast<uint16_t>((a1 & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL)));
bool vflip = ((!affine) && (cast<uint16_t>((a1 & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL)));
{int64_t bx = cast<int64_t>(0ULL);for (;(bx < bw);bx++){
int64_t x = cast<int64_t>((sx + bx));
if (((x < cast<int64_t>(0ULL)) || (x >= cast<int64_t>(256ULL)))) {
continue;
}
int64_t px={};
int64_t py={};
if (affine) {
int32_t rx = cast<int32_t>(cast<int64_t>((bx - divi<int64_t>(bw,cast<int64_t>(2ULL)))));
int32_t ry = cast<int32_t>(cast<int64_t>((dy - divi<int64_t>(bh,cast<int64_t>(2ULL)))));
px = cast<int64_t>((cast<int64_t>(shr<int32_t>((cast<int32_t>((cast<int32_t>((pa * rx)) + cast<int32_t>((pb * ry))))),cast<int64_t>(8ULL))) + divi<int64_t>(w,cast<int64_t>(2ULL))));
py = cast<int64_t>((cast<int64_t>(shr<int32_t>((cast<int32_t>((cast<int32_t>((pc * rx)) + cast<int32_t>((pd * ry))))),cast<int64_t>(8ULL))) + divi<int64_t>(h,cast<int64_t>(2ULL))));
if (((((px < cast<int64_t>(0ULL)) || (px >= w)) || (py < cast<int64_t>(0ULL))) || (py >= h))) {
continue;
}
}
else {
auto tmp73 = std::make_tuple(bx,dy);
px = std::get<0>(tmp73);
py = std::get<1>(tmp73);
if (hflip) {
px = cast<int64_t>((cast<int64_t>((w - cast<int64_t>(1ULL))) - px));
}
if (vflip) {
py = cast<int64_t>((cast<int64_t>((h - cast<int64_t>(1ULL))) - py));
}
}
if ((e->objOK[x] && (e->objPrio[x] < prio))) {
continue;
}
uint16_t c={};
bool ok={};
uint8_t eva={};
if ((gfx == cast<uint16_t>(3ULL))) {
auto tmp74 = dsmachine_engine_objBitmapPixel(e,dispcnt,cast<uint32_t>(cast<uint16_t>((a2 & cast<uint16_t>(1023ULL)))),w,px,py);
c = std::get<0>(tmp74);
ok = std::get<1>(tmp74);
eva = cast<uint8_t>((cast<uint8_t>(alpha) + cast<uint8_t>(1ULL)));
}
else {
auto tmp75 = dsmachine_engine_objTilePixel(e,dispcnt,cast<uint32_t>(cast<uint16_t>((a2 & cast<uint16_t>(1023ULL)))),bpp8,pal,w,px,py);
c = std::get<0>(tmp75);
ok = std::get<1>(tmp75);
}
if ((!ok)) {
continue;
}
if ((gfx == cast<uint16_t>(2ULL))) {
e->objWin[x] = true;
continue;
}
if ((cast<uint32_t>((dispcnt & cast<uint32_t>(4096ULL))) == cast<uint32_t>(0ULL))) {
continue;
}
e->obj[x] = c;
e->objOK[x] = true;
e->objPrio[x] = prio;
e->objSemi[x] = ((gfx == cast<uint16_t>(1ULL)) || (gfx == cast<uint16_t>(3ULL)));
e->objEVA[x] = eva;
}
}}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:970:1
uint16_t dsmachine_engine_oam16(dsmachine_engine* e,uint32_t off){
{
return dsmachine_palAt(e->m->oam,off);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:976:1
std::tuple<int32_t,int32_t,int32_t,int32_t> dsmachine_engine_oamAffine(dsmachine_engine* e,int64_t i){
int32_t pa{};
int32_t pb{};
int32_t pc{};
int32_t pd{};
{
uint32_t b = cast<uint32_t>((e->oamBase + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(32ULL)))));
pa = cast<int32_t>(cast<int16_t>(dsmachine_engine_oam16(e,cast<uint32_t>((b + cast<uint32_t>(6ULL))))));
pb = cast<int32_t>(cast<int16_t>(dsmachine_engine_oam16(e,cast<uint32_t>((b + cast<uint32_t>(14ULL))))));
pc = cast<int32_t>(cast<int16_t>(dsmachine_engine_oam16(e,cast<uint32_t>((b + cast<uint32_t>(22ULL))))));
pd = cast<int32_t>(cast<int16_t>(dsmachine_engine_oam16(e,cast<uint32_t>((b + cast<uint32_t>(30ULL))))));
return {pa,pb,pc,pd};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:993:1
std::tuple<uint16_t,bool> dsmachine_engine_objTilePixel(dsmachine_engine* e,uint32_t dispcnt,uint32_t tile,bool bpp8,int64_t pal,int64_t w,int64_t px,int64_t py){
{
bool oneD = (cast<uint32_t>((dispcnt & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL));
int64_t tw = divi<int64_t>(w,cast<int64_t>(8ULL));
uint32_t addr={};
if (oneD) {
uint32_t bound = shl<uint32_t>(cast<uint32_t>(32ULL),(cast<uint32_t>((shr<uint32_t>(dispcnt,cast<int64_t>(20ULL)) & cast<uint32_t>(3ULL)))));
uint32_t sz = cast<uint32_t>(32ULL);
if (bpp8) {
sz = cast<uint32_t>(64ULL);
}
addr = cast<uint32_t>((cast<uint32_t>((tile * bound)) + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(divi<int64_t>(py,cast<int64_t>(8ULL))) * cast<uint32_t>(tw))) + cast<uint32_t>(divi<int64_t>(px,cast<int64_t>(8ULL)))))) * sz))));
}
else {
addr = cast<uint32_t>((tile * cast<uint32_t>(32ULL)));
addr += cast<uint32_t>((cast<uint32_t>(divi<int64_t>(py,cast<int64_t>(8ULL))) * cast<uint32_t>(1024ULL)));
if (bpp8) {
addr += cast<uint32_t>((cast<uint32_t>(divi<int64_t>(px,cast<int64_t>(8ULL))) * cast<uint32_t>(64ULL)));
}
else {
addr += cast<uint32_t>((cast<uint32_t>(divi<int64_t>(px,cast<int64_t>(8ULL))) * cast<uint32_t>(32ULL)));
}
}
auto tmp76 = std::make_tuple(cast<uint32_t>(modi<int64_t>(px,cast<int64_t>(8ULL))),cast<uint32_t>(modi<int64_t>(py,cast<int64_t>(8ULL))));
uint32_t fx = std::get<0>(tmp76);
uint32_t fy = std::get<1>(tmp76);
int64_t idx={};
if (bpp8) {
idx = cast<int64_t>(dsmachine_vram_read8(e->m->vram,e->spOBJ,cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>((fy * cast<uint32_t>(8ULL))))) + fx))));
}
else {
uint8_t b = dsmachine_vram_read8(e->m->vram,e->spOBJ,cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>((fy * cast<uint32_t>(4ULL))))) + divi<uint32_t>(fx,cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>((fx & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
idx = cast<int64_t>(shr<uint8_t>(b,cast<int64_t>(4ULL)));
}
else {
idx = cast<int64_t>(cast<uint8_t>((b & cast<uint8_t>(15ULL))));
}
}
if ((idx == cast<int64_t>(0ULL))) {
return {cast<uint16_t>(0ULL),false};
}
{
if ((bpp8 && (cast<uint32_t>((dispcnt & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))){
return {dsmachine_engine_objExtColor(e,pal,idx),true};
}
else if (bpp8){
return {dsmachine_engine_palOBJColor(e,idx),true};
}
else {
return {dsmachine_engine_palOBJColor(e,cast<int64_t>((cast<int64_t>((pal * cast<int64_t>(16ULL))) + idx))),true};
}
}
tmp77:;
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1045:1
std::tuple<uint16_t,bool> dsmachine_engine_objBitmapPixel(dsmachine_engine* e,uint32_t dispcnt,uint32_t tile,int64_t w,int64_t px,int64_t py){
{
uint32_t base={};
uint32_t stride={};
if ((cast<uint32_t>((dispcnt & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
uint32_t bound = cast<uint32_t>(128ULL);
if ((cast<uint32_t>((dispcnt & cast<uint32_t>(4194304ULL))) != cast<uint32_t>(0ULL))) {
bound = cast<uint32_t>(256ULL);
}
base = cast<uint32_t>((tile * bound));
stride = cast<uint32_t>((cast<uint32_t>(w) * cast<uint32_t>(2ULL)));
}
else if ((cast<uint32_t>((dispcnt & cast<uint32_t>(32ULL))) == cast<uint32_t>(0ULL))) {
base = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((tile & cast<uint32_t>(31ULL)))) * cast<uint32_t>(16ULL))) + cast<uint32_t>(((cast<uint32_t>((tile & cast<uint32_t>(992ULL)))) * cast<uint32_t>(128ULL)))));
stride = cast<uint32_t>(256ULL);
}
else {
base = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((tile & cast<uint32_t>(15ULL)))) * cast<uint32_t>(16ULL))) + cast<uint32_t>(((cast<uint32_t>((tile & cast<uint32_t>(1008ULL)))) * cast<uint32_t>(128ULL)))));
stride = cast<uint32_t>(512ULL);
}
uint16_t c = dsmachine_vram_read16(e->m->vram,e->spOBJ,cast<uint32_t>((cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(py) * stride)))) + cast<uint32_t>((cast<uint32_t>(px) * cast<uint32_t>(2ULL))))));
if ((cast<uint16_t>((c & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL))) {
return {cast<uint16_t>(0ULL),false};
}
return {cast<uint16_t>((c & cast<uint16_t>(32767ULL))),true};
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1074:1
void dsmachine_engine_windowMask(dsmachine_engine* e,uint32_t dispcnt,int64_t y){
{
if ((cast<uint32_t>((dispcnt & cast<uint32_t>(57344ULL))) == cast<uint32_t>(0ULL))) {
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
e->win[x] = cast<uint8_t>(63ULL);
}
}return ;
}
uint16_t winin = dsmachine_engine_reg16(e,cast<uint32_t>(72ULL));
uint16_t winout = dsmachine_engine_reg16(e,cast<uint32_t>(74ULL));
uint8_t out = cast<uint8_t>(cast<uint16_t>((winout & cast<uint16_t>(63ULL))));
uint8_t objw = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(winout,cast<int64_t>(8ULL)) & cast<uint16_t>(63ULL))));
uint8_t w0 = cast<uint8_t>(cast<uint16_t>((winin & cast<uint16_t>(63ULL))));
uint8_t w1 = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(winin,cast<int64_t>(8ULL)) & cast<uint16_t>(63ULL))));
bool in0 = ((cast<uint32_t>((dispcnt & cast<uint32_t>(8192ULL))) != cast<uint32_t>(0ULL)) && dsmachine_engine_inWindowY(e,cast<int64_t>(0ULL),y));
bool in1 = ((cast<uint32_t>((dispcnt & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL)) && dsmachine_engine_inWindowY(e,cast<int64_t>(1ULL),y));
bool useObj = (cast<uint32_t>((dispcnt & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint8_t m = out;
if ((useObj && e->objWin[x])) {
m = objw;
}
if ((in1 && dsmachine_engine_inWindowX(e,cast<int64_t>(1ULL),x))) {
m = w1;
}
if ((in0 && dsmachine_engine_inWindowX(e,cast<int64_t>(0ULL),x))) {
m = w0;
}
e->win[x] = m;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:1112:1
bool dsmachine_engine_inWindowY(dsmachine_engine* e,int64_t n,int64_t y){
{
uint16_t v = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(68ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
auto tmp78 = std::make_tuple(cast<int64_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL)))));
int64_t y1 = std::get<0>(tmp78);
int64_t y2 = std::get<1>(tmp78);
if (((y2 <= y1) || (y2 > cast<int64_t>(192ULL)))) {
y2 = cast<int64_t>(192ULL);
}
return ((y >= y1) && (y < y2));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1121:1
bool dsmachine_engine_inWindowX(dsmachine_engine* e,int64_t n,int64_t x){
{
uint16_t v = dsmachine_engine_reg16(e,cast<uint32_t>((cast<uint32_t>(64ULL) + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(2ULL))))));
auto tmp79 = std::make_tuple(cast<int64_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL)))));
int64_t x1 = std::get<0>(tmp79);
int64_t x2 = std::get<1>(tmp79);
if ((x2 <= x1)) {
x2 = cast<int64_t>(256ULL);
}
return ((x >= x1) && (x < x2));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1144:1
void dsmachine_engine_composite(dsmachine_engine* e,int64_t y){
{
uint16_t bldcnt = dsmachine_engine_reg16(e,cast<uint32_t>(80ULL));
uint16_t bldalpha = dsmachine_engine_reg16(e,cast<uint32_t>(82ULL));
uint16_t bldy = dsmachine_engine_reg16(e,cast<uint32_t>(84ULL));
int64_t effect = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(bldcnt,cast<int64_t>(6ULL)) & cast<uint16_t>(3ULL))));
uint8_t first = cast<uint8_t>(cast<uint16_t>((bldcnt & cast<uint16_t>(63ULL))));
uint8_t second = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(bldcnt,cast<int64_t>(8ULL)) & cast<uint16_t>(63ULL))));
int64_t eva = dsmachine_clamp16(cast<int64_t>(cast<uint16_t>((bldalpha & cast<uint16_t>(31ULL)))));
int64_t evb = dsmachine_clamp16(cast<int64_t>(cast<uint16_t>((shr<uint16_t>(bldalpha,cast<int64_t>(8ULL)) & cast<uint16_t>(31ULL)))));
int64_t evy = dsmachine_clamp16(cast<int64_t>(cast<uint16_t>((bldy & cast<uint16_t>(31ULL)))));
uint16_t backdrop = dsmachine_engine_palBGColor(e,cast<int64_t>(0ULL));
std::array<dsmachine_cand,6> cs={};
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint8_t mask = e->win[x];
int64_t n = cast<int64_t>(0ULL);
{uint8_t p = cast<uint8_t>(0ULL);for (;(p < cast<uint8_t>(4ULL));p++){
if (((e->objOK[x] && (e->objPrio[x] == p)) && (cast<uint8_t>((mask & cast<uint8_t>(16ULL))) != cast<uint8_t>(0ULL)))) {
cs[n] = dsmachine_cand{e->obj[x],cast<int64_t>(4ULL),e->objSemi[x],e->objEVA[x],{}};
n++;
}
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(4ULL));b++){
if (((((!e->shown[b]) || (!e->bgOK[b][x])) || (e->prio[b] != p)) || (cast<uint8_t>((mask & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(b))))) == cast<uint8_t>(0ULL)))) {
continue;
}
cs[n] = dsmachine_cand{e->bg[b][x],b,{},{},{}};
if (((b == cast<int64_t>(0ULL)) && e->is3D)) {
cs[n].a3d = e->a3D[x];
}
n++;
}
}}
}cs[n] = dsmachine_cand{backdrop,cast<int64_t>(5ULL),{},{},cast<uint8_t>(31ULL)};
n++;
dsmachine_cand top = cs[cast<int64_t>(0ULL)];
if (((top.layer == cast<int64_t>(0ULL)) && e->is3D)) {
e->vis3D[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))] = true;
}
dsmachine_cand below = cs[cast<int64_t>(1ULL)];
if ((n == cast<int64_t>(1ULL))) {
below = dsmachine_cand{backdrop,cast<int64_t>(5ULL),{},{},{}};
}
bool fx = (cast<uint8_t>((mask & cast<uint8_t>(32ULL))) != cast<uint8_t>(0ULL));
{
if ((((top.layer == cast<int64_t>(0ULL)) && e->is3D) && (top.a3d < cast<uint8_t>(31ULL)))){
e->line[x] = dsmachine_blend32(top.c,below.c,cast<int64_t>((cast<int64_t>(top.a3d) + cast<int64_t>(1ULL))),cast<int64_t>((cast<int64_t>(31ULL) - cast<int64_t>(top.a3d))));
rrds::composite(e,x,y,top,below,effect,fx);
}
else if (((top.semi && fx) && (cast<uint8_t>((second & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(below.layer))))) != cast<uint8_t>(0ULL)))){
auto tmp81 = std::make_tuple(eva,evb);
int64_t a = std::get<0>(tmp81);
int64_t b = std::get<1>(tmp81);
if ((top.eva != cast<uint8_t>(0ULL))) {
a = dsmachine_clamp16(cast<int64_t>(top.eva));
b = cast<int64_t>((cast<int64_t>(16ULL) - a));
}
e->line[x] = dsmachine_blend16(top.c,below.c,a,b);
rrds::composite(e,x,y,top,below,effect,fx);
}
else if ((fx && (cast<uint8_t>((first & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(top.layer))))) != cast<uint8_t>(0ULL)))){
{
switch(effect){
case cast<int64_t>(1ULL):{
if ((cast<uint8_t>((second & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(below.layer))))) != cast<uint8_t>(0ULL))) {
e->line[x] = dsmachine_blend16(top.c,below.c,eva,evb);
rrds::composite(e,x,y,top,below,effect,fx);
}
else {
e->line[x] = top.c;
rrds::composite(e,x,y,top,below,effect,fx);
}
break;}
case cast<int64_t>(2ULL):{
e->line[x] = dsmachine_brighten(top.c,evy);
rrds::composite(e,x,y,top,below,effect,fx);
break;}
case cast<int64_t>(3ULL):{
e->line[x] = dsmachine_darken(top.c,evy);
rrds::composite(e,x,y,top,below,effect,fx);
break;}
default:{
e->line[x] = top.c;
rrds::composite(e,x,y,top,below,effect,fx);
break;}
}}
}
else {
e->line[x] = top.c;
rrds::composite(e,x,y,top,below,effect,fx);
}
}
tmp80:;
}
}}
}
// tools/platform/nds/dsmachine/gpu2d.go:1233:1
int64_t dsmachine_clamp16(int64_t v){
{
if ((v > cast<int64_t>(16ULL))) {
return cast<int64_t>(16ULL);
}
return v;
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1243:1
uint16_t dsmachine_blend16(uint16_t a,uint16_t b,int64_t eva,int64_t evb){
{
return dsmachine_blendN(a,b,eva,evb,cast<int64_t>(16ULL));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1249:1
uint16_t dsmachine_blend32(uint16_t a,uint16_t b,int64_t ca,int64_t cb){
{
return dsmachine_blendN(a,b,ca,cb,cast<int64_t>(32ULL));
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1253:1
uint16_t dsmachine_blendN(uint16_t a,uint16_t b,int64_t ca,int64_t cb,int64_t div){
{
std::array<uint32_t,3> out={};
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < cast<uint64_t>(3ULL));i++){
uint32_t v = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((dsmachine_chan5(a,i) * cast<uint32_t>(ca))) + cast<uint32_t>((dsmachine_chan5(b,i) * cast<uint32_t>(cb)))))),cast<uint32_t>(div));
if ((v > cast<uint32_t>(31ULL))) {
v = cast<uint32_t>(31ULL);
}
out[i] = v;
}
}return dsmachine_rgb555(out[cast<int64_t>(0ULL)],out[cast<int64_t>(1ULL)],out[cast<int64_t>(2ULL)]);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1265:1
uint16_t dsmachine_brighten(uint16_t c,int64_t evy){
{
std::array<uint32_t,3> out={};
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < cast<uint64_t>(3ULL));i++){
uint32_t v = dsmachine_chan5(c,i);
out[i] = cast<uint32_t>((v + divi<uint32_t>(cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(31ULL) - v))) * cast<uint32_t>(evy))),cast<uint32_t>(16ULL))));
}
}return dsmachine_rgb555(out[cast<int64_t>(0ULL)],out[cast<int64_t>(1ULL)],out[cast<int64_t>(2ULL)]);
}
}
// tools/platform/nds/dsmachine/gpu2d.go:1274:1
uint16_t dsmachine_darken(uint16_t c,int64_t evy){
{
std::array<uint32_t,3> out={};
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < cast<uint64_t>(3ULL));i++){
uint32_t v = dsmachine_chan5(c,i);
out[i] = cast<uint32_t>((v - divi<uint32_t>(cast<uint32_t>((v * cast<uint32_t>(evy))),cast<uint32_t>(16ULL))));
}
}return dsmachine_rgb555(out[cast<int64_t>(0ULL)],out[cast<int64_t>(1ULL)],out[cast<int64_t>(2ULL)]);
}
}
// tools/platform/nds/dsmachine/gpu3d.go:76:1
dsmachine_gpu3d* dsmachine_newGPU3D(){
{
dsmachine_gpu3d* g = arenaNew(dsmachine_gpu3d{{},{},{},{},{},{},{},Map<uint32_t,uint32_t>{},{},{},{},{},{},{},{},cast<int64_t>(-1ULL),{}});
dsmachine_geom_reset(&(g->geom));
dsmachine_raster_reset(&(g->rast));
return g;
}
}
// tools/platform/nds/dsmachine/gpu3d.go:128:1
bool dsmachine_gpu3d_fifoBelowHalf(dsmachine_gpu3d* g){
{
return (len(g->fifo) < cast<int64_t>(128ULL));
}
}
// tools/platform/nds/dsmachine/gpu3d.go:130:1
void dsmachine_gpu3d_writeReg(dsmachine_gpu3d* g,uint32_t a,uint32_t w){
{
{
if (((a >= cast<uint32_t>(67109888ULL)) && (a < cast<uint32_t>(67109952ULL)))){
dsmachine_gpu3d_pushPacked(g,w);
return ;
}
else if (((a >= cast<uint32_t>(67109952ULL)) && (a <= cast<uint32_t>(67110344ULL)))){
dsmachine_gpu3d_pushDirect(g,cast<uint8_t>((cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67109952ULL)))),cast<uint32_t>(4ULL))) + cast<uint8_t>(16ULL))),w);
return ;
}
else if ((a == cast<uint32_t>(67110400ULL))){
return ;
}
}
tmp82:;
g->regs[a] = w;
}
}
// tools/platform/nds/dsmachine/gpu3d.go:163:1
uint32_t dsmachine_gpu3d_readReg(dsmachine_gpu3d* g,uint32_t a){
{
{
if ((a == cast<uint32_t>(67110400ULL))){
return dsmachine_gpu3d_gxstat(g);
}
else if ((a == cast<uint32_t>(67110404ULL))){
return cast<uint32_t>((cast<uint32_t>(len(g->geom.polys)) | shl<uint32_t>(cast<uint32_t>(len(g->geom.verts)),cast<int64_t>(16ULL))));
}
else if (((a >= cast<uint32_t>(67110464ULL)) && (a < cast<uint32_t>(67110528ULL)))){
return cast<uint32_t>(dsmachine_geom_clip(&(g->geom)).m[divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67110464ULL)))),cast<uint32_t>(4ULL))]);
}
else if (((a >= cast<uint32_t>(67110528ULL)) && (a < cast<uint32_t>(67110564ULL)))){
return cast<uint32_t>(dsmachine_geom_vecMtx3x3(&(g->geom),divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67110528ULL)))),cast<uint32_t>(4ULL))));
}
else if (((a >= cast<uint32_t>(67110432ULL)) && (a < cast<uint32_t>(67110448ULL)))){
return cast<uint32_t>(g->geom.posResult[divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67110432ULL)))),cast<uint32_t>(4ULL))]);
}
else if (((a >= cast<uint32_t>(67110448ULL)) && (a < cast<uint32_t>(67110456ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67110448ULL)))),cast<uint32_t>(2ULL));
uint32_t v = cast<uint32_t>(cast<uint16_t>(g->geom.vecResult[i]));
if ((cast<uint32_t>((i + cast<uint32_t>(1ULL))) < cast<uint32_t>(3ULL))) {
v |= shl<uint32_t>(cast<uint32_t>(cast<uint16_t>(g->geom.vecResult[cast<uint32_t>((i + cast<uint32_t>(1ULL)))])),cast<int64_t>(16ULL));
}
return v;
}
}
tmp83:;
return get(g->regs,a);
}
}
// tools/platform/nds/dsmachine/gpu3d.go:193:1
uint32_t dsmachine_gpu3d_gxstat(dsmachine_gpu3d* g){
{
uint32_t n = cast<uint32_t>(len(g->fifo));
if ((n > cast<uint32_t>(256ULL))) {
n = cast<uint32_t>(256ULL);
}
uint32_t v = shl<uint32_t>(n,cast<int64_t>(16ULL));
if ((n == cast<uint32_t>(0ULL))) {
v |= cast<uint32_t>(67108864ULL);
}
if ((n < cast<uint32_t>(128ULL))) {
v |= cast<uint32_t>(33554432ULL);
}
v |= shl<uint32_t>(cast<uint32_t>(dsmachine_geom_stackDepth(&(g->geom))),cast<int64_t>(8ULL));
return v;
}
}
// tools/platform/nds/dsmachine/gpu3d.go:212:1
void dsmachine_gpu3d_pushPacked(dsmachine_gpu3d* g,uint32_t w){
{
if (((g->cmd == cast<uint8_t>(0ULL)) && (len(g->packed) == cast<int64_t>(0ULL)))) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
{
uint8_t b = cast<uint8_t>(shr<uint32_t>(w,(cast<int64_t>((cast<int64_t>(8ULL) * i)))));
if ((b != cast<uint8_t>(0ULL))) {
g->packed = append(g->packed,b);
}
}
}
}dsmachine_gpu3d_nextPacked(g);
return ;
}
dsmachine_gpu3d_feed(g,w);
}
}
// tools/platform/nds/dsmachine/gpu3d.go:227:1
void dsmachine_gpu3d_nextPacked(dsmachine_gpu3d* g){
{
{;for (;(len(g->packed) > cast<int64_t>(0ULL));){
uint8_t c = g->packed[cast<int64_t>(0ULL)];
int64_t n = get(dsmachine_gxParams,c);
if ((n > cast<int64_t>(0ULL))) {
g->packed = sub(g->packed,cast<int64_t>(1ULL),len(g->packed));
auto tmp84 = std::make_tuple(c,n,sub(g->params,0,cast<int64_t>(0ULL)));
g->cmd = std::get<0>(tmp84);
g->need = std::get<1>(tmp84);
g->params = std::get<2>(tmp84);
return ;
}
g->packed = sub(g->packed,cast<int64_t>(1ULL),len(g->packed));
dsmachine_gpu3d_exec(g,c,{});
}
}auto tmp85 = std::make_tuple(cast<uint8_t>(0ULL),cast<int64_t>(0ULL));
g->cmd = std::get<0>(tmp85);
g->need = std::get<1>(tmp85);
}
}
// tools/platform/nds/dsmachine/gpu3d.go:243:1
void dsmachine_gpu3d_feed(dsmachine_gpu3d* g,uint32_t w){
{
if ((g->cmd == cast<uint8_t>(0ULL))) {
return ;
}
g->params = append(g->params,w);
if ((len(g->params) < g->need)) {
return ;
}
dsmachine_gpu3d_exec(g,g->cmd,g->params);
auto tmp86 = std::make_tuple(cast<uint8_t>(0ULL),cast<int64_t>(0ULL));
g->cmd = std::get<0>(tmp86);
g->need = std::get<1>(tmp86);
dsmachine_gpu3d_nextPacked(g);
}
}
// tools/platform/nds/dsmachine/gpu3d.go:257:1
void dsmachine_gpu3d_pushDirect(dsmachine_gpu3d* g,uint8_t cmd,uint32_t w){
{
int64_t n = get(dsmachine_gxParams,cmd);
if ((n == cast<int64_t>(0ULL))) {
dsmachine_gpu3d_exec(g,cmd,{});
return ;
}
if ((g->cmd != cmd)) {
auto tmp87 = std::make_tuple(cmd,n,sub(g->params,0,cast<int64_t>(0ULL)));
g->cmd = std::get<0>(tmp87);
g->need = std::get<1>(tmp87);
g->params = std::get<2>(tmp87);
}
g->params = append(g->params,w);
if ((len(g->params) >= g->need)) {
dsmachine_gpu3d_exec(g,g->cmd,g->params);
auto tmp88 = std::make_tuple(cast<uint8_t>(0ULL),cast<int64_t>(0ULL));
g->cmd = std::get<0>(tmp88);
g->need = std::get<1>(tmp88);
}
}
}
// tools/platform/nds/dsmachine/gpu3d.go:276:1
void dsmachine_gpu3d_vblank(dsmachine_gpu3d* g,dsmachine_Machine* m){
{
if ((!g->swapPending)) {
return ;
}
g->swapPending = false;
g->lastPolys = len(g->geom.polys);
g->swaps++;
dsmachine_gpu3d_render(g,m);
dsmachine_geom_beginFrame(&(g->geom));
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:23:1
dsmachine_mtx dsmachine_identity(){
{
dsmachine_mtx r={};
auto tmp89 = std::make_tuple(cast<int32_t>(4096ULL),cast<int32_t>(4096ULL),cast<int32_t>(4096ULL),cast<int32_t>(4096ULL));
r.m[cast<int64_t>(0ULL)] = std::get<0>(tmp89);
r.m[cast<int64_t>(5ULL)] = std::get<1>(tmp89);
r.m[cast<int64_t>(10ULL)] = std::get<2>(tmp89);
r.m[cast<int64_t>(15ULL)] = std::get<3>(tmp89);
return r;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:30:1
dsmachine_mtx dsmachine_mtx_mul(dsmachine_mtx a,dsmachine_mtx b){
{
dsmachine_mtx r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(4ULL));j++){
int64_t acc={};
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(4ULL));k++){
acc += cast<int64_t>((cast<int64_t>(a.m[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + k))]) * cast<int64_t>(b.m[cast<int64_t>((cast<int64_t>((k * cast<int64_t>(4ULL))) + j))])));
}
}r.m[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + j))] = cast<int32_t>(shr<int64_t>(acc,cast<int64_t>(12ULL)));
}
}}
}return r;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:45:1
std::tuple<int64_t,int64_t,int64_t,int64_t> dsmachine_mtx_apply(dsmachine_mtx a,int64_t x,int64_t y,int64_t z,int64_t w){
{
auto c = [&](int64_t j)->int64_t{
return shr<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((x * cast<int64_t>(a.m[cast<int64_t>((cast<int64_t>(0ULL) + j))]))) + cast<int64_t>((y * cast<int64_t>(a.m[cast<int64_t>((cast<int64_t>(4ULL) + j))]))))) + cast<int64_t>((z * cast<int64_t>(a.m[cast<int64_t>((cast<int64_t>(8ULL) + j))]))))) + cast<int64_t>((w * cast<int64_t>(a.m[cast<int64_t>((cast<int64_t>(12ULL) + j))])))))),cast<int64_t>(12ULL));
}
;
return {c(cast<int64_t>(0ULL)),c(cast<int64_t>(1ULL)),c(cast<int64_t>(2ULL)),c(cast<int64_t>(3ULL))};
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:143:1
void dsmachine_geom_reset(dsmachine_geom* g){
{
auto tmp90 = std::make_tuple(dsmachine_identity(),dsmachine_identity(),dsmachine_identity(),dsmachine_identity());
g->proj = std::get<0>(tmp90);
g->pos = std::get<1>(tmp90);
g->vec = std::get<2>(tmp90);
g->tex = std::get<3>(tmp90);
g->clipDirty = true;
auto tmp91 = std::make_tuple(cast<int32_t>(255ULL),cast<int32_t>(191ULL));
g->viewX2 = std::get<0>(tmp91);
g->viewY2 = std::get<1>(tmp91);
g->attrNext = cast<uint32_t>(0ULL);
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:150:1
void dsmachine_geom_beginFrame(dsmachine_geom* g){
{
g->verts = sub(g->verts,0,cast<int64_t>(0ULL));
g->polys = sub(g->polys,0,cast<int64_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:155:1
int64_t dsmachine_geom_stackDepth(dsmachine_geom* g){
{
return g->sp;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:160:1
dsmachine_mtx dsmachine_geom_clip(dsmachine_geom* g){
{
if (g->clipDirty) {
g->clipMtx = dsmachine_mtx_mul(g->pos,g->proj);
g->clipDirty = false;
}
return g->clipMtx;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:169:1
int32_t dsmachine_geom_vecMtx3x3(dsmachine_geom* g,uint32_t i){
{
auto tmp92 = std::make_tuple(divi<uint32_t>(i,cast<uint32_t>(3ULL)),modi<uint32_t>(i,cast<uint32_t>(3ULL)));
uint32_t row = std::get<0>(tmp92);
uint32_t col = std::get<1>(tmp92);
return g->vec.m[cast<uint32_t>((cast<uint32_t>((row * cast<uint32_t>(4ULL))) + col))];
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:177:1
Slice<dsmachine_mtx*> dsmachine_geom_current(dsmachine_geom* g){
{
{
switch(g->mode){
case cast<int64_t>(0ULL):{
return Slice<dsmachine_mtx*>{(&g->proj)};
break;}
case cast<int64_t>(1ULL):{
return Slice<dsmachine_mtx*>{(&g->pos)};
break;}
case cast<int64_t>(2ULL):{
return Slice<dsmachine_mtx*>{(&g->pos),(&g->vec)};
break;}
default:{
return Slice<dsmachine_mtx*>{(&g->tex)};
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:191:1
void dsmachine_gpu3d_exec(dsmachine_gpu3d* g,uint8_t cmd,Slice<uint32_t> p){
rrprof::Scope measured(1,"GX geometry");
{
if (((g->limit >= cast<int64_t>(0ULL)) && (g->count >= g->limit))) {
throw dsmachine_gxHalt{};
}
g->cur = g->count;
g->count++;
g->cmdHist[cmd]++;
if (bool(g->OnCmd)) {
g->OnCmd(cmd,p);
}
dsmachine_geom* ge = (&g->geom);
{
switch(cmd){
case cast<uint8_t>(16ULL):{
ge->mode = cast<int64_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(3ULL))));
break;}
case cast<uint8_t>(17ULL):{
{
switch(ge->mode){
case cast<int64_t>(0ULL):{
ge->projStack[cast<int64_t>(0ULL)] = ge->proj;
ge->projSP = cast<int64_t>(1ULL);
break;}
case cast<int64_t>(3ULL):{
ge->texStack[cast<int64_t>(0ULL)] = ge->tex;
ge->texSP = cast<int64_t>(1ULL);
break;}
default:{
if ((ge->sp < cast<int64_t>(31ULL))) {
ge->posStack[ge->sp] = ge->pos;
ge->vecStack[ge->sp] = ge->vec;
ge->sp++;
}
break;}
}}
break;}
case cast<uint8_t>(18ULL):{
{
switch(ge->mode){
case cast<int64_t>(0ULL):{
ge->proj = ge->projStack[cast<int64_t>(0ULL)];
ge->projSP = cast<int64_t>(0ULL);
ge->clipDirty = true;
break;}
case cast<int64_t>(3ULL):{
ge->tex = ge->texStack[cast<int64_t>(0ULL)];
ge->texSP = cast<int64_t>(0ULL);
break;}
default:{
int64_t n = cast<int64_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(63ULL))));
if ((n >= cast<int64_t>(32ULL))) {
n -= cast<int64_t>(64ULL);
}
ge->sp -= n;
if ((ge->sp < cast<int64_t>(0ULL))) {
ge->sp = cast<int64_t>(0ULL);
}
if ((ge->sp > cast<int64_t>(31ULL))) {
ge->sp = cast<int64_t>(31ULL);
}
ge->pos = ge->posStack[ge->sp];
ge->vec = ge->vecStack[ge->sp];
ge->clipDirty = true;
break;}
}}
break;}
case cast<uint8_t>(19ULL):{
int64_t i = cast<int64_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL))));
{
switch(ge->mode){
case cast<int64_t>(0ULL):{
ge->projStack[cast<int64_t>(0ULL)] = ge->proj;
break;}
case cast<int64_t>(3ULL):{
ge->texStack[cast<int64_t>(0ULL)] = ge->tex;
break;}
default:{
ge->posStack[i] = ge->pos;
ge->vecStack[i] = ge->vec;
break;}
}}
break;}
case cast<uint8_t>(20ULL):{
int64_t i = cast<int64_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL))));
{
switch(ge->mode){
case cast<int64_t>(0ULL):{
ge->proj = ge->projStack[cast<int64_t>(0ULL)];
ge->clipDirty = true;
break;}
case cast<int64_t>(3ULL):{
ge->tex = ge->texStack[cast<int64_t>(0ULL)];
break;}
default:{
ge->pos = ge->posStack[i];
ge->vec = ge->vecStack[i];
ge->clipDirty = true;
break;}
}}
break;}
case cast<uint8_t>(21ULL):{
{auto&& tmp93 = dsmachine_geom_current(ge);
for(int64_t tmp94=0;tmp94<len(tmp93);++tmp94){
auto m=tmp93[tmp94];(*m) = dsmachine_identity();
}}
ge->clipDirty = true;
break;}
case cast<uint8_t>(22ULL):case cast<uint8_t>(24ULL):{
dsmachine_mtx n={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(16ULL));i++){
n.m[i] = cast<int32_t>(p[i]);
}
}dsmachine_geom_applyMatrix(ge,(cmd == cast<uint8_t>(22ULL)),n);
break;}
case cast<uint8_t>(23ULL):case cast<uint8_t>(25ULL):{
dsmachine_mtx n = dsmachine_identity();
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(3ULL));j++){
n.m[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + j))] = cast<int32_t>(p[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(3ULL))) + j))]);
}
}n.m[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + cast<int64_t>(3ULL)))] = cast<int32_t>(0ULL);
}
}n.m[cast<int64_t>(15ULL)] = cast<int32_t>(4096ULL);
dsmachine_geom_applyMatrix(ge,(cmd == cast<uint8_t>(23ULL)),n);
break;}
case cast<uint8_t>(26ULL):{
dsmachine_mtx n = dsmachine_identity();
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(3ULL));j++){
n.m[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + j))] = cast<int32_t>(p[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(3ULL))) + j))]);
}
}}
}dsmachine_geom_applyMatrix(ge,false,n);
break;}
case cast<uint8_t>(27ULL):{
dsmachine_mtx n = dsmachine_identity();
auto tmp95 = std::make_tuple(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<int32_t>(p[cast<int64_t>(1ULL)]),cast<int32_t>(p[cast<int64_t>(2ULL)]));
n.m[cast<int64_t>(0ULL)] = std::get<0>(tmp95);
n.m[cast<int64_t>(5ULL)] = std::get<1>(tmp95);
n.m[cast<int64_t>(10ULL)] = std::get<2>(tmp95);
if ((ge->mode == cast<int64_t>(0ULL))) {
ge->proj = dsmachine_mtx_mul(n,ge->proj);
}
else if ((ge->mode == cast<int64_t>(3ULL))) {
ge->tex = dsmachine_mtx_mul(n,ge->tex);
}
else {
ge->pos = dsmachine_mtx_mul(n,ge->pos);
}
ge->clipDirty = true;
break;}
case cast<uint8_t>(28ULL):{
dsmachine_mtx n = dsmachine_identity();
auto tmp96 = std::make_tuple(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<int32_t>(p[cast<int64_t>(1ULL)]),cast<int32_t>(p[cast<int64_t>(2ULL)]));
n.m[cast<int64_t>(12ULL)] = std::get<0>(tmp96);
n.m[cast<int64_t>(13ULL)] = std::get<1>(tmp96);
n.m[cast<int64_t>(14ULL)] = std::get<2>(tmp96);
dsmachine_geom_applyMatrix(ge,false,n);
break;}
case cast<uint8_t>(32ULL):{
ge->color = dsmachine_unpackRGB5(p[cast<int64_t>(0ULL)]);
break;}
case cast<uint8_t>(33ULL):{
dsmachine_geom_normal(ge,p[cast<int64_t>(0ULL)]);
break;}
case cast<uint8_t>(34ULL):{
ge->texS = cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)])));
ge->texT = cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)))));
break;}
case cast<uint8_t>(35ULL):{
int32_t x = cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)])));
int32_t y = cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)))));
int32_t z = cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(1ULL)])));
dsmachine_geom_vertex(ge,g,x,y,z);
break;}
case cast<uint8_t>(36ULL):{
int32_t x = shl<int32_t>(dsmachine_sext(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<uint64_t>(10ULL)),cast<int64_t>(6ULL));
int32_t y = shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(6ULL));
int32_t z = shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(20ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(6ULL));
dsmachine_geom_vertex(ge,g,x,y,z);
break;}
case cast<uint8_t>(37ULL):{
dsmachine_geom_vertex(ge,g,cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)]))),cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL))))),ge->lastVtx[cast<int64_t>(2ULL)]);
break;}
case cast<uint8_t>(38ULL):{
dsmachine_geom_vertex(ge,g,cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)]))),ge->lastVtx[cast<int64_t>(1ULL)],cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL))))));
break;}
case cast<uint8_t>(39ULL):{
dsmachine_geom_vertex(ge,g,ge->lastVtx[cast<int64_t>(0ULL)],cast<int32_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)]))),cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL))))));
break;}
case cast<uint8_t>(40ULL):{
int32_t dx = dsmachine_sext(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<uint64_t>(10ULL));
int32_t dy = dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))),cast<uint64_t>(10ULL));
int32_t dz = dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(20ULL))),cast<uint64_t>(10ULL));
dsmachine_geom_vertex(ge,g,cast<int32_t>((ge->lastVtx[cast<int64_t>(0ULL)] + dx)),cast<int32_t>((ge->lastVtx[cast<int64_t>(1ULL)] + dy)),cast<int32_t>((ge->lastVtx[cast<int64_t>(2ULL)] + dz)));
break;}
case cast<uint8_t>(41ULL):{
ge->attrNext = p[cast<int64_t>(0ULL)];
break;}
case cast<uint8_t>(42ULL):{
ge->texParam = p[cast<int64_t>(0ULL)];
break;}
case cast<uint8_t>(43ULL):{
ge->pltt = p[cast<int64_t>(0ULL)];
break;}
case cast<uint8_t>(48ULL):{
ge->diffuse = dsmachine_unpackRGB5(p[cast<int64_t>(0ULL)]);
ge->ambient = dsmachine_unpackRGB5(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)));
if ((cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
ge->color = ge->diffuse;
}
break;}
case cast<uint8_t>(49ULL):{
ge->specular = dsmachine_unpackRGB5(p[cast<int64_t>(0ULL)]);
ge->emission = dsmachine_unpackRGB5(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)));
break;}
case cast<uint8_t>(50ULL):{
int64_t n = cast<int64_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(30ULL)));
int64_t x = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t y = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t z = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(20ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
auto tmp97 = dsmachine_mtx_apply(ge->vec,x,y,z,cast<int64_t>(0ULL));
int64_t tx = std::get<0>(tmp97);
int64_t ty = std::get<1>(tmp97);
int64_t tz = std::get<2>(tmp97);
ge->lightVec[n] = std::array<int32_t,3>{cast<int32_t>(tx),cast<int32_t>(ty),cast<int32_t>(tz)};
break;}
case cast<uint8_t>(51ULL):{
ge->lightColor[shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(30ULL))] = dsmachine_unpackRGB5(p[cast<int64_t>(0ULL)]);
break;}
case cast<uint8_t>(52ULL):{
{auto&& tmp98 = p;
for(int64_t tmp99=0;tmp99<len(tmp98);++tmp99){
auto i=tmp99;auto w=tmp98[tmp99];{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(4ULL));j++){
ge->shininess[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + j))] = cast<uint8_t>(shr<uint32_t>(w,(cast<int64_t>((cast<int64_t>(8ULL) * j)))));
}
}}}
break;}
case cast<uint8_t>(64ULL):{
ge->begun = true;
ge->primMode = cast<int64_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(3ULL))));
ge->strip = sub(ge->strip,0,cast<int64_t>(0ULL));
ge->stripLen = cast<int64_t>(0ULL);
ge->attr = ge->attrNext;
break;}
case cast<uint8_t>(65ULL):{
ge->begun = false;
break;}
case cast<uint8_t>(80ULL):{
g->swapPending = true;
g->swapMode = p[cast<int64_t>(0ULL)];
ge->wbuffer = (cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint8_t>(96ULL):{
ge->viewX1 = cast<int32_t>(cast<uint32_t>((p[cast<int64_t>(0ULL)] & cast<uint32_t>(255ULL))));
ge->viewY1 = cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
ge->viewX2 = cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL))));
ge->viewY2 = cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(24ULL))) & cast<uint32_t>(255ULL))));
break;}
case cast<uint8_t>(112ULL):{
g->regs[cast<uint32_t>(67110400ULL)] |= cast<uint32_t>(2ULL);
break;}
case cast<uint8_t>(113ULL):{
int64_t x = cast<int64_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(0ULL)])));
int64_t y = cast<int64_t>(cast<int16_t>(cast<uint16_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)))));
int64_t z = cast<int64_t>(cast<int16_t>(cast<uint16_t>(p[cast<int64_t>(1ULL)])));
auto tmp100 = dsmachine_mtx_apply(dsmachine_geom_clip(ge),x,y,z,cast<int64_t>(4096ULL));
int64_t cx = std::get<0>(tmp100);
int64_t cy = std::get<1>(tmp100);
int64_t cz = std::get<2>(tmp100);
int64_t cw = std::get<3>(tmp100);
ge->posResult = std::array<int32_t,4>{cast<int32_t>(cx),cast<int32_t>(cy),cast<int32_t>(cz),cast<int32_t>(cw)};
break;}
case cast<uint8_t>(114ULL):{
int64_t x = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(p[cast<int64_t>(0ULL)]),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t y = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t z = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(p[cast<int64_t>(0ULL)],cast<int64_t>(20ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
auto tmp101 = dsmachine_mtx_apply(ge->vec,x,y,z,cast<int64_t>(0ULL));
int64_t vx = std::get<0>(tmp101);
int64_t vy = std::get<1>(tmp101);
int64_t vz = std::get<2>(tmp101);
ge->vecResult = std::array<int32_t,3>{cast<int32_t>(vx),cast<int32_t>(vy),cast<int32_t>(vz)};
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:452:1
void dsmachine_geom_applyMatrix(dsmachine_geom* g,bool load,dsmachine_mtx n){
{
{auto&& tmp102 = dsmachine_geom_current(g);
for(int64_t tmp103=0;tmp103<len(tmp102);++tmp103){
auto m=tmp102[tmp103];if (load) {
(*m) = n;
}
else {
(*m) = dsmachine_mtx_mul(n,(*m));
}
}}
g->clipDirty = true;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:464:1
int32_t dsmachine_sext(int32_t v,uint64_t n){
{
uint64_t sh = cast<uint64_t>((cast<uint64_t>(32ULL) - n));
return shr<int32_t>(shl<int32_t>(v,sh),sh);
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:472:1
std::array<int32_t,3> dsmachine_unpackRGB5(uint32_t v){
{
int32_t r = cast<int32_t>((cast<int32_t>(cast<uint32_t>((v & cast<uint32_t>(31ULL)))) * cast<int32_t>(2ULL)));
int32_t g = cast<int32_t>((cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)))) * cast<int32_t>(2ULL)));
int32_t b = cast<int32_t>((cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)))) * cast<int32_t>(2ULL)));
return std::array<int32_t,3>{r,g,b};
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:483:1
void dsmachine_geom_normal(dsmachine_geom* g,uint32_t v){
{
int64_t nx = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(v),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t ny = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(v,cast<int64_t>(10ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
int64_t nz = cast<int64_t>(shl<int32_t>(dsmachine_sext(cast<int32_t>(shr<uint32_t>(v,cast<int64_t>(20ULL))),cast<uint64_t>(10ULL)),cast<int64_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(g->texParam,cast<int64_t>(30ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))) {
auto tmp104 = dsmachine_geom_texTransformNormal(g,nx,ny,nz);
int32_t s = std::get<0>(tmp104);
int32_t t = std::get<1>(tmp104);
auto tmp105 = std::make_tuple(s,t);
g->texS = std::get<0>(tmp105);
g->texT = std::get<1>(tmp105);
}
auto tmp106 = dsmachine_mtx_apply(g->vec,nx,ny,nz,cast<int64_t>(0ULL));
int64_t tx = std::get<0>(tmp106);
int64_t ty = std::get<1>(tmp106);
int64_t tz = std::get<2>(tmp106);
std::array<int32_t,3> col = g->emission;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
if ((cast<uint32_t>((g->attrNext & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) == cast<uint32_t>(0ULL))) {
continue;
}
std::array<int32_t,3> lv = g->lightVec[i];
std::array<int32_t,3> lc = g->lightColor[i];
int64_t dot = shr<int64_t>(cast<int64_t>(-(cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(0ULL)]) * tx)) + cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(1ULL)]) * ty)))) + cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(2ULL)]) * tz)))))),cast<int64_t>(12ULL));
if ((dot < cast<int64_t>(0ULL))) {
dot = cast<int64_t>(0ULL);
}
if ((dot > cast<int64_t>(4096ULL))) {
dot = cast<int64_t>(4096ULL);
}
int64_t shine={};
int64_t hz = cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(2ULL)]) - cast<int64_t>(4096ULL)));
int64_t hlen = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(0ULL)]) * cast<int64_t>(lv[cast<int64_t>(0ULL)]))) + cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(1ULL)]) * cast<int64_t>(lv[cast<int64_t>(1ULL)]))))) + cast<int64_t>((hz * hz))));
if ((hlen > cast<int64_t>(0ULL))) {
int64_t hdot = shr<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(0ULL)]) * tx)) + cast<int64_t>((cast<int64_t>(lv[cast<int64_t>(1ULL)]) * ty)))) + cast<int64_t>((hz * tz))))),cast<int64_t>(12ULL));
if ((hdot > cast<int64_t>(0ULL))) {
int64_t sh = divi<int64_t>((shl<int64_t>(cast<int64_t>((hdot * hdot)),cast<int64_t>(12ULL))),hlen);
if ((sh > cast<int64_t>(4096ULL))) {
sh = cast<int64_t>(4096ULL);
}
int64_t idx = shr<int64_t>(sh,cast<int64_t>(5ULL));
if ((idx > cast<int64_t>(127ULL))) {
idx = cast<int64_t>(127ULL);
}
if ((cast<uint32_t>((g->attrNext & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
shine = shl<int64_t>(cast<int64_t>(g->shininess[idx]),cast<int64_t>(4ULL));
}
else {
shine = sh;
}
}
}
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(3ULL));c++){
int64_t d = shr<int64_t>(cast<int64_t>((cast<int64_t>((cast<int64_t>(g->diffuse[c]) * cast<int64_t>(lc[c]))) * dot)),cast<int64_t>(18ULL));
int64_t a = shr<int64_t>(cast<int64_t>((cast<int64_t>(g->ambient[c]) * cast<int64_t>(lc[c]))),cast<int64_t>(6ULL));
int64_t s = shr<int64_t>(cast<int64_t>((cast<int64_t>((cast<int64_t>(g->specular[c]) * cast<int64_t>(lc[c]))) * shine)),cast<int64_t>(18ULL));
col[c] += cast<int32_t>(cast<int64_t>((cast<int64_t>((d + a)) + s)));
}
}}
}{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(3ULL));c++){
if ((col[c] > cast<int32_t>(63ULL))) {
col[c] = cast<int32_t>(63ULL);
}
if ((col[c] < cast<int32_t>(0ULL))) {
col[c] = cast<int32_t>(0ULL);
}
}
}g->color = col;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:563:1
std::tuple<int32_t,int32_t> dsmachine_geom_texTransformNormal(dsmachine_geom* g,int64_t nx,int64_t ny,int64_t nz){
{
auto tmp107 = dsmachine_mtx_apply(g->tex,nx,ny,nz,cast<int64_t>(4096ULL));
int64_t s = std::get<0>(tmp107);
int64_t t = std::get<1>(tmp107);
return {cast<int32_t>(shr<int64_t>(s,cast<int64_t>(8ULL))),cast<int32_t>(shr<int64_t>(t,cast<int64_t>(8ULL)))};
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:573:1
std::tuple<int32_t,int32_t> dsmachine_geom_texCoordFor(dsmachine_geom* g,int32_t x,int32_t y,int32_t z){
{
{
switch(cast<uint32_t>(((shr<uint32_t>(g->texParam,cast<int64_t>(30ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(1ULL):{
auto tmp108 = dsmachine_mtx_apply(g->tex,shl<int64_t>(cast<int64_t>(g->texS),cast<int64_t>(8ULL)),shl<int64_t>(cast<int64_t>(g->texT),cast<int64_t>(8ULL)),cast<int64_t>(4096ULL),cast<int64_t>(4096ULL));
int64_t s = std::get<0>(tmp108);
int64_t t = std::get<1>(tmp108);
return {cast<int32_t>(shr<int64_t>(s,cast<int64_t>(8ULL))),cast<int32_t>(shr<int64_t>(t,cast<int64_t>(8ULL)))};
break;}
case cast<uint32_t>(3ULL):{
auto tmp109 = dsmachine_mtx_apply(g->tex,cast<int64_t>(x),cast<int64_t>(y),cast<int64_t>(z),cast<int64_t>(4096ULL));
int64_t s = std::get<0>(tmp109);
int64_t t = std::get<1>(tmp109);
return {cast<int32_t>(shr<int64_t>(s,cast<int64_t>(8ULL))),cast<int32_t>(shr<int64_t>(t,cast<int64_t>(8ULL)))};
break;}
}}
return {g->texS,g->texT};
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:587:1
void dsmachine_geom_vertex(dsmachine_geom* g,dsmachine_gpu3d* gp,int32_t x,int32_t y,int32_t z){
{
g->lastVtx = std::array<int32_t,3>{x,y,z};
if ((!g->begun)) {
return ;
}
auto tmp110 = dsmachine_mtx_apply(dsmachine_geom_clip(g),cast<int64_t>(x),cast<int64_t>(y),cast<int64_t>(z),cast<int64_t>(4096ULL));
int64_t cx = std::get<0>(tmp110);
int64_t cy = std::get<1>(tmp110);
int64_t cz = std::get<2>(tmp110);
int64_t cw = std::get<3>(tmp110);
auto tmp111 = dsmachine_geom_texCoordFor(g,x,y,z);
int32_t s = std::get<0>(tmp111);
int32_t t = std::get<1>(tmp111);
dsmachine_gxVertex v = dsmachine_gxVertex{cx,cy,cz,cw,g->color[cast<int64_t>(0ULL)],g->color[cast<int64_t>(1ULL)],g->color[cast<int64_t>(2ULL)],s,t};
g->strip = append(g->strip,v);
g->stripLen++;
dsmachine_geom_emitIfComplete(g,gp);
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:607:1
void dsmachine_geom_emitIfComplete(dsmachine_geom* g,dsmachine_gpu3d* gp){
{
{
switch(g->primMode){
case cast<int64_t>(0ULL):{
if ((len(g->strip) == cast<int64_t>(3ULL))) {
dsmachine_geom_emit(g,gp,g->strip[cast<int64_t>(0ULL)],g->strip[cast<int64_t>(1ULL)],g->strip[cast<int64_t>(2ULL)]);
g->strip = sub(g->strip,0,cast<int64_t>(0ULL));
}
break;}
case cast<int64_t>(1ULL):{
if ((len(g->strip) == cast<int64_t>(4ULL))) {
dsmachine_geom_emit(g,gp,g->strip[cast<int64_t>(0ULL)],g->strip[cast<int64_t>(1ULL)],g->strip[cast<int64_t>(2ULL)],g->strip[cast<int64_t>(3ULL)]);
g->strip = sub(g->strip,0,cast<int64_t>(0ULL));
}
break;}
case cast<int64_t>(2ULL):{
if ((len(g->strip) >= cast<int64_t>(3ULL))) {
int64_t n = len(g->strip);
auto tmp112 = std::make_tuple(g->strip[cast<int64_t>((n - cast<int64_t>(3ULL)))],g->strip[cast<int64_t>((n - cast<int64_t>(2ULL)))],g->strip[cast<int64_t>((n - cast<int64_t>(1ULL)))]);
dsmachine_gxVertex a = std::get<0>(tmp112);
dsmachine_gxVertex b = std::get<1>(tmp112);
dsmachine_gxVertex c = std::get<2>(tmp112);
if ((modi<int64_t>((cast<int64_t>((g->stripLen - cast<int64_t>(3ULL)))),cast<int64_t>(2ULL)) == cast<int64_t>(1ULL))) {
auto tmp113 = std::make_tuple(b,a);
a = std::get<0>(tmp113);
b = std::get<1>(tmp113);
}
dsmachine_geom_emit(g,gp,a,b,c);
if ((n > cast<int64_t>(3ULL))) {
g->strip = sub(g->strip,cast<int64_t>((n - cast<int64_t>(2ULL))),len(g->strip));
g->strip = append(Slice<dsmachine_gxVertex>{},g->strip);
}
}
break;}
case cast<int64_t>(3ULL):{
if (((len(g->strip) >= cast<int64_t>(4ULL)) && (modi<int64_t>(len(g->strip),cast<int64_t>(2ULL)) == cast<int64_t>(0ULL)))) {
int64_t n = len(g->strip);
dsmachine_geom_emit(g,gp,g->strip[cast<int64_t>((n - cast<int64_t>(4ULL)))],g->strip[cast<int64_t>((n - cast<int64_t>(3ULL)))],g->strip[cast<int64_t>((n - cast<int64_t>(1ULL)))],g->strip[cast<int64_t>((n - cast<int64_t>(2ULL)))]);
g->strip = append(Slice<dsmachine_gxVertex>{},sub(g->strip,cast<int64_t>((n - cast<int64_t>(2ULL))),len(g->strip)));
}
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:644:1
void dsmachine_geom_emit(dsmachine_geom* g,dsmachine_gpu3d* gp,Slice<dsmachine_gxVertex> vs){
{
g->nEmit++;
Slice<dsmachine_gxVertex> poly = dsmachine_clipPolygon(vs);
if ((len(poly) < cast<int64_t>(3ULL))) {
g->nClipped++;
return ;
}
if ((len(g->polys) >= cast<int64_t>(2048ULL))) {
return ;
}
g->polys = append(g->polys,dsmachine_gxPolygon{poly,g->attr,g->texParam,g->pltt,g->wbuffer,gp->cur});
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:666:1
Slice<dsmachine_gxVertex> dsmachine_clipPolygon(Slice<dsmachine_gxVertex> vs){
{
Slice<dsmachine_gxVertex> out = append(Slice<dsmachine_gxVertex>{},vs);
{int64_t plane = cast<int64_t>(0ULL);for (;(plane < cast<int64_t>(6ULL));plane++){
if ((len(out) == cast<int64_t>(0ULL))) {
return {};
}
Slice<dsmachine_gxVertex> in = out;
out = fullSub(out,0,cast<int64_t>(0ULL),cast<int64_t>(0ULL));
{auto&& tmp114 = in;
for(int64_t tmp115=0;tmp115<len(tmp114);++tmp115){
auto i=tmp115;auto tmp116 = std::make_tuple(in[i],in[modi<int64_t>((cast<int64_t>((cast<int64_t>((i + len(in))) - cast<int64_t>(1ULL)))),len(in))]);
dsmachine_gxVertex cur = std::get<0>(tmp116);
dsmachine_gxVertex prev = std::get<1>(tmp116);
auto tmp117 = std::make_tuple(dsmachine_planeDist(cur,plane),dsmachine_planeDist(prev,plane));
int64_t dc = std::get<0>(tmp117);
int64_t dp = std::get<1>(tmp117);
if ((dc >= cast<int64_t>(0ULL))) {
if ((dp < cast<int64_t>(0ULL))) {
out = append(out,dsmachine_lerpVertex(prev,cur,dp,dc));
}
out = append(out,cur);
}
else if ((dp >= cast<int64_t>(0ULL))) {
out = append(out,dsmachine_lerpVertex(prev,cur,dp,dc));
}
}}
}
}return out;
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:691:1
int64_t dsmachine_planeDist(dsmachine_gxVertex v,int64_t plane){
{
{
switch(plane){
case cast<int64_t>(0ULL):{
return cast<int64_t>((v.w - v.x));
break;}
case cast<int64_t>(1ULL):{
return cast<int64_t>((v.w + v.x));
break;}
case cast<int64_t>(2ULL):{
return cast<int64_t>((v.w - v.y));
break;}
case cast<int64_t>(3ULL):{
return cast<int64_t>((v.w + v.y));
break;}
case cast<int64_t>(4ULL):{
return cast<int64_t>((v.w - v.z));
break;}
default:{
return cast<int64_t>((v.w + v.z));
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_geom.go:710:1
dsmachine_gxVertex dsmachine_lerpVertex(dsmachine_gxVertex a,dsmachine_gxVertex b,int64_t da,int64_t db){
{
int64_t den = cast<int64_t>((da - db));
if ((den == cast<int64_t>(0ULL))) {
return a;
}
int64_t t = divi<int64_t>((shl<int64_t>(da,cast<int64_t>(12ULL))),den);
auto li = [&](int64_t x,int64_t y)->int64_t{
return cast<int64_t>((x + shr<int64_t>(cast<int64_t>(((cast<int64_t>((y - x))) * t)),cast<int64_t>(12ULL))));
}
;
auto li32 = [&](int32_t x,int32_t y)->int32_t{
return cast<int32_t>(cast<int64_t>((cast<int64_t>(x) + shr<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(y) - cast<int64_t>(x)))) * t)),cast<int64_t>(12ULL)))));
}
;
return dsmachine_gxVertex{li(a.x,b.x),li(a.y,b.y),li(a.z,b.z),li(a.w,b.w),li32(a.r,b.r),li32(a.g,b.g),li32(a.b,b.b),li32(a.s,b.s),li32(a.t,b.t)};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:61:1
void dsmachine_raster_reset(dsmachine_raster* r){
{
{auto&& tmp118 = r->col;
for(int64_t tmp119=0;tmp119<len(tmp118);++tmp119){
auto i=tmp119;r->col[i] = dsmachine_rfrag{};
r->depth[i] = cast<uint32_t>(16777215ULL);
}}
if ((!r->frame)) {
r->frame = Slice<uint32_t>::make(cast<int64_t>(49152ULL));
}
{auto&& tmp120 = r->frame;
for(int64_t tmp121=0;tmp121<len(tmp120);++tmp121){
auto i=tmp121;r->frame[i] = cast<uint32_t>(0ULL);
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:89:1
dsmachine_rvert dsmachine_lerpRV(dsmachine_rvert a,dsmachine_rvert b,double u){
{
auto l = [&](double x,double y)->double{
return (x + (((y - x)) * u));
}
;
return dsmachine_rvert{l(a.x,b.x),l(a.y,b.y),l(a.depth,b.depth),l(a.iw,b.iw),l(a.w,b.w),l(a.r,b.r),l(a.g,b.g),l(a.b,b.b),l(a.s,b.s),l(a.t,b.t)};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:100:1
void dsmachine_gpu3d_render(dsmachine_gpu3d* g,dsmachine_Machine* m){
rrprof::Scope measured(2,"3D software rasterizer");
{
if (m->prof.on) {
time_Time t0 = go_time_Now();
auto tmp122=defer([&](){[&]()->void{
m->prof.raster += go_time_Since(t0);
}
();});
}
m->prof.polys += len(g->geom.polys);
uint32_t disp3d = get(m->ARM9->io,cast<uint32_t>(67108960ULL));
dsmachine_gpu3d_clear(g,m,disp3d);
{auto&& tmp123 = g->geom.polys;
for(int64_t tmp124=0;tmp124<len(tmp123);++tmp124){
auto i=tmp124;if (bool(m->OnPoly)) {
m->OnPoly(g->geom.polys[i].cmd);
}
dsmachine_gpu3d_drawPoly(g,m,(&g->geom.polys[i]),disp3d);
}}
if ((cast<uint32_t>((disp3d & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_Machine_note(m,std::string("3D: fog enabled (DISP3DCNT bit 7) \342\200\224 not implemented, the image is unfogged",76));
}
if ((cast<uint32_t>((disp3d & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_Machine_note(m,std::string("3D: edge marking enabled (DISP3DCNT bit 5) \342\200\224 not implemented",62));
}
if ((cast<uint32_t>((disp3d & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_Machine_note(m,std::string("3D: anti-aliasing enabled (DISP3DCNT bit 4) \342\200\224 not implemented, edges are hard",79));
}
dsmachine_raster_publish(&(g->rast));
if (bool(m->gpu2d)) {
if ((!m->gpu2d->threeD)) {
m->gpu2d->threeD = Slice<uint32_t>::make(cast<int64_t>(49152ULL));
}
gcopy(m->gpu2d->threeD,g->rast.frame);
}
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:142:1
void dsmachine_raster_publish(dsmachine_raster* r){
{
if ((!r->frame)) {
r->frame = Slice<uint32_t>::make(cast<int64_t>(49152ULL));
}
{auto&& tmp125 = r->col;
for(int64_t tmp126=0;tmp126<len(tmp125);++tmp126){
auto i=tmp126;auto c=tmp125[tmp126];if ((c.a == cast<uint8_t>(0ULL))) {
r->frame[i] = cast<uint32_t>(0ULL);
rrds::publish(r,i);
continue;
}
r->frame[i] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dsmachine_c6to8(c.r)) | shl<uint32_t>(cast<uint32_t>(dsmachine_c6to8(c.g)),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_c6to8(c.b)),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dsmachine_a5to8(c.a)),cast<int64_t>(24ULL))));
rrds::publish(r,i);
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:156:1
uint8_t dsmachine_c6to8(uint8_t v){
{
return cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(v) * cast<int64_t>(255ULL))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL)));
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:157:1
uint8_t dsmachine_a5to8(uint8_t v){
{
return cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(v) * cast<int64_t>(255ULL))) + cast<int64_t>(15ULL)))),cast<int64_t>(31ULL)));
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:165:1
void dsmachine_gpu3d_clear(dsmachine_gpu3d* g,dsmachine_Machine* m,uint32_t disp3d){
{
if ((cast<uint32_t>((disp3d & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_Machine_note(m,std::string("3D: clear-image / rear-plane bitmap enabled (DISP3DCNT bit 14) \342\200\224 not implemented, clearing to CLEAR_COLOR instead",115));
}
uint32_t cc = get(g->regs,cast<uint32_t>(67109712ULL));
auto tmp127 = std::make_tuple(cast<int32_t>(cast<uint32_t>((cc & cast<uint32_t>(31ULL)))),cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)))),cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)))));
int32_t r5 = std::get<0>(tmp127);
int32_t g5 = std::get<1>(tmp127);
int32_t b5 = std::get<2>(tmp127);
uint8_t ca = cast<uint8_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))));
dsmachine_rfrag c = dsmachine_rfrag{cast<uint8_t>(cast<int32_t>((r5 * cast<int32_t>(2ULL)))),cast<uint8_t>(cast<int32_t>((g5 * cast<int32_t>(2ULL)))),cast<uint8_t>(cast<int32_t>((b5 * cast<int32_t>(2ULL)))),ca};
uint32_t d = cast<uint32_t>(((cast<uint32_t>((get(g->regs,cast<uint32_t>(67109716ULL)) & cast<uint32_t>(32767ULL)))) * cast<uint32_t>(512ULL)));
d += cast<uint32_t>(511ULL);
{auto&& tmp128 = g->rast.col;
for(int64_t tmp129=0;tmp129<len(tmp128);++tmp129){
auto i=tmp129;g->rast.col[i] = c;
g->rast.depth[i] = d;
}}
rrds::clear3D(g,m);
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:185:1
void dsmachine_gpu3d_drawPoly(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_gxPolygon* p,uint32_t disp3d){
{
if ((len(p->verts) < cast<int64_t>(3ULL))) {
return ;
}
uint32_t mode = cast<uint32_t>(((shr<uint32_t>(p->attr,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL)));
if ((mode == cast<uint32_t>(3ULL))) {
dsmachine_Machine_note(m,std::string("3D: shadow polygon (POLYGON_ATTR mode 3) skipped \342\200\224 shadow volumes not implemented",83));
return ;
}
auto tmp130 = std::make_tuple(cast<int64_t>(g->geom.viewX1),cast<int64_t>(g->geom.viewY1));
int64_t vx1 = std::get<0>(tmp130);
int64_t vy1 = std::get<1>(tmp130);
auto tmp131 = std::make_tuple(cast<int64_t>(g->geom.viewX2),cast<int64_t>(g->geom.viewY2));
int64_t vx2 = std::get<0>(tmp131);
int64_t vy2 = std::get<1>(tmp131);
int64_t vw = cast<int64_t>((cast<int64_t>((vx2 - vx1)) + cast<int64_t>(1ULL)));
int64_t vh = cast<int64_t>((cast<int64_t>((vy2 - vy1)) + cast<int64_t>(1ULL)));
if (((vw <= cast<int64_t>(0ULL)) || (vh <= cast<int64_t>(0ULL)))) {
return ;
}
Slice<dsmachine_rvert> verts = Slice<dsmachine_rvert>::make(cast<int64_t>(0ULL),len(p->verts));
{auto&& tmp132 = p->verts;
for(int64_t tmp133=0;tmp133<len(tmp132);++tmp133){
auto v=tmp132[tmp133];double w = cast<double>(v.w);
if ((v.w == cast<int64_t>(0ULL))) {
return ;
}
double sx = (((((cast<double>(v.x) + w)) * cast<double>(vw)) / ((cast<double>(2ULL) * w))) + cast<double>(vx1));
double sy = (((((cast<double>(v.y) + w)) * cast<double>(vh)) / ((cast<double>(2ULL) * w))) + cast<double>(vy1));
sy = (cast<double>(cast<double>(191ULL)) - sy);
double iw = (1.0 / w);
double depth={};
if ((!p->wbuffer)) {
double d = (((((cast<double>(v.z) * 16384.0) / w) + 16383.0)) * 512.0);
if ((d < cast<double>(0ULL))) {
d = cast<double>(0ULL);
}
if ((d > cast<double>(16777215ULL))) {
d = cast<double>(16777215ULL);
}
depth = d;
}
verts = append(verts,dsmachine_rvert{sx,sy,depth,iw,w,(cast<double>(v.r) * iw),(cast<double>(v.g) * iw),(cast<double>(v.b) * iw),(cast<double>(v.s) * iw),(cast<double>(v.t) * iw)});
}}
bool drawBack = (cast<uint32_t>((p->attr & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
bool drawFront = (cast<uint32_t>((p->attr & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL));
if (((!drawBack) && (!drawFront))) {
return ;
}
double area={};
{auto&& tmp134 = verts;
for(int64_t tmp135=0;tmp135<len(tmp134);++tmp135){
auto i=tmp135;int64_t j = modi<int64_t>((cast<int64_t>((i + cast<int64_t>(1ULL)))),len(verts));
area += ((verts[i].x * verts[j].y) - (verts[j].x * verts[i].y));
}}
bool front = (area < cast<double>(0ULL));
if ((area == cast<double>(0ULL))) {
return ;
}
if ((front && (!drawFront))) {
return ;
}
if (((!front) && (!drawBack))) {
return ;
}
auto tmp136 = std::make_tuple(vx1,vx2);
int64_t clipX0 = std::get<0>(tmp136);
int64_t clipX1 = std::get<1>(tmp136);
int64_t clipY0 = cast<int64_t>((cast<int64_t>(191ULL) - vy2));
int64_t clipY1 = cast<int64_t>((cast<int64_t>(191ULL) - vy1));
if ((clipX0 < cast<int64_t>(0ULL))) {
clipX0 = cast<int64_t>(0ULL);
}
if ((clipX1 > cast<int64_t>(255ULL))) {
clipX1 = cast<int64_t>(255ULL);
}
if ((clipY0 < cast<int64_t>(0ULL))) {
clipY0 = cast<int64_t>(0ULL);
}
if ((clipY1 > cast<int64_t>(191ULL))) {
clipY1 = cast<int64_t>(191ULL);
}
auto tmp137 = std::make_tuple(verts[cast<int64_t>(0ULL)].y,verts[cast<int64_t>(0ULL)].y);
double minY = std::get<0>(tmp137);
double maxY = std::get<1>(tmp137);
{auto&& tmp138 = sub(verts,cast<int64_t>(1ULL),len(verts));
for(int64_t tmp139=0;tmp139<len(tmp138);++tmp139){
auto v=tmp138[tmp139];if ((v.y < minY)) {
minY = v.y;
}
if ((v.y > maxY)) {
maxY = v.y;
}
}}
int64_t y0 = cast<int64_t>(dsmachine_rastCeil((minY - 0.5)));
int64_t y1 = cast<int64_t>((cast<int64_t>(dsmachine_rastCeil((maxY - 0.5))) - cast<int64_t>(1ULL)));
if ((y0 < clipY0)) {
y0 = clipY0;
}
if ((y1 > clipY1)) {
y1 = clipY1;
}
dsmachine_polyState st = dsmachine_gpu3d_polyState(g,m,p,disp3d);
{int64_t y = y0;for (;(y <= y1);y++){
double yc = (cast<double>(y) + 0.5);
bool have={};
dsmachine_rvert left={};
dsmachine_rvert right={};
{auto&& tmp140 = verts;
for(int64_t tmp141=0;tmp141<len(tmp140);++tmp141){
auto i=tmp141;auto tmp142 = std::make_tuple(verts[i],verts[modi<int64_t>((cast<int64_t>((i + cast<int64_t>(1ULL)))),len(verts))]);
dsmachine_rvert a = std::get<0>(tmp142);
dsmachine_rvert b = std::get<1>(tmp142);
if ((a.y == b.y)) {
continue;
}
auto tmp143 = std::make_tuple(a,b);
dsmachine_rvert lo = std::get<0>(tmp143);
dsmachine_rvert hi = std::get<1>(tmp143);
if ((lo.y > hi.y)) {
auto tmp144 = std::make_tuple(hi,lo);
lo = std::get<0>(tmp144);
hi = std::get<1>(tmp144);
}
if (((yc < lo.y) || (yc >= hi.y))) {
continue;
}
double u = (((yc - lo.y)) / ((hi.y - lo.y)));
dsmachine_rvert s = dsmachine_lerpRV(lo,hi,u);
if ((!have)) {
auto tmp145 = std::make_tuple(s,s,true);
left = std::get<0>(tmp145);
right = std::get<1>(tmp145);
have = std::get<2>(tmp145);
}
else if ((s.x < left.x)) {
left = s;
}
else if ((s.x > right.x)) {
right = s;
}
}}
if ((!have)) {
continue;
}
int64_t x0 = cast<int64_t>(dsmachine_rastCeil((left.x - 0.5)));
int64_t x1 = cast<int64_t>((cast<int64_t>(dsmachine_rastCeil((right.x - 0.5))) - cast<int64_t>(1ULL)));
if ((x0 == cast<int64_t>((x1 + cast<int64_t>(1ULL))))) {
x1 = x0;
}
if ((x0 < clipX0)) {
x0 = clipX0;
}
if ((x1 > clipX1)) {
x1 = clipX1;
}
if ((x0 > x1)) {
continue;
}
double dx = (right.x - left.x);
{int64_t x = x0;for (;(x <= x1);x++){
if (((((st.wireframe && (x != x0)) && (x != x1)) && (y != y0)) && (y != y1))) {
continue;
}
double u={};
if ((dx > cast<double>(0ULL))) {
u = ((((cast<double>(x) + 0.5) - left.x)) / dx);
if ((u < cast<double>(0ULL))) {
u = cast<double>(0ULL);
}
if ((u > cast<double>(1ULL))) {
u = cast<double>(1ULL);
}
}
dsmachine_rvert f = dsmachine_lerpRV(left,right,u);
dsmachine_gpu3d_shade(g,m,(&st),p,f,cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x)));
}
}}
}}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:421:1
dsmachine_polyState dsmachine_gpu3d_polyState(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_gxPolygon* p,uint32_t disp3d){
{
dsmachine_polyState st={};
st.mode = cast<uint32_t>(((shr<uint32_t>(p->attr,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL)));
st.alpha = cast<uint8_t>(cast<uint32_t>(((shr<uint32_t>(p->attr,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))));
st.depthEq = (cast<uint32_t>((p->attr & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL));
st.wbuf = p->wbuffer;
st.blend = (cast<uint32_t>((disp3d & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
st.toonHi = (cast<uint32_t>((disp3d & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
st.alphaTe = (cast<uint32_t>((disp3d & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL));
st.alphaRf = cast<uint8_t>(cast<uint32_t>((get(g->regs,cast<uint32_t>(67109696ULL)) & cast<uint32_t>(31ULL))));
if ((st.alpha == cast<uint8_t>(0ULL))) {
st.wireframe = true;
st.alpha = cast<uint8_t>(31ULL);
dsmachine_Machine_note(m,std::string("3D: wireframe polygons (POLYGON_ATTR alpha 0) approximated by span-edge pixels",78));
}
st.trans = (st.alpha < cast<uint8_t>(31ULL));
st.depthWr = ((!st.trans) || (cast<uint32_t>((p->attr & cast<uint32_t>(2048ULL))) != cast<uint32_t>(0ULL)));
if ((cast<uint32_t>((disp3d & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
auto tmp146 = dsmachine_texStateOf(p);
st.tex = std::get<0>(tmp146);
st.hasTex = std::get<1>(tmp146);
}
if ((st.mode == cast<uint32_t>(2ULL))) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(32ULL));i++){
uint32_t w = get(g->regs,cast<uint32_t>((cast<uint32_t>(67109760ULL) + cast<uint32_t>((cast<uint32_t>(divi<int64_t>(i,cast<int64_t>(2ULL))) * cast<uint32_t>(4ULL))))));
uint16_t c = cast<uint16_t>(shr<uint32_t>(w,(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(cast<int64_t>((i & cast<int64_t>(1ULL)))))))));
st.toon[i] = std::array<uint8_t,3>{cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>((c & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL)))),cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(c,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL)))),cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(c,cast<int64_t>(10ULL))) & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL))))};
}
}}
return st;
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:466:1
void dsmachine_gpu3d_shade(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_polyState* st,dsmachine_gxPolygon* p,dsmachine_rvert f,int64_t idx){
{
auto reject = [&](bool z,bool a)->void{
if (bool(m->OnPixel)) {
m->OnPixel(modi<int64_t>(idx,cast<int64_t>(256ULL)),divi<int64_t>(idx,cast<int64_t>(256ULL)),dsmachine_PixelEvent{{},z,a,{},{},{},{}});
}
}
;
if ((f.iw == cast<double>(0ULL))) {
return ;
}
double w = (1.0 / f.iw);
uint32_t depth={};
if (st->wbuf) {
double d = w;
if ((d < cast<double>(0ULL))) {
d = cast<double>(0ULL);
}
if ((d > cast<double>(16777215ULL))) {
d = cast<double>(16777215ULL);
}
depth = cast<uint32_t>(d);
}
else {
double d = f.depth;
if ((d < cast<double>(0ULL))) {
d = cast<double>(0ULL);
}
if ((d > cast<double>(16777215ULL))) {
d = cast<double>(16777215ULL);
}
depth = cast<uint32_t>(d);
}
uint32_t old = g->rast.depth[idx];
if (st->depthEq) {
int64_t diff = cast<int64_t>((cast<int64_t>(depth) - cast<int64_t>(old)));
if ((diff < cast<int64_t>(0ULL))) {
diff = cast<int64_t>(-diff);
}
if ((diff > cast<int64_t>(512ULL))) {
reject(true,false);
return ;
}
}
else if ((depth >= old)) {
reject(true,false);
return ;
}
uint8_t vr = dsmachine_rastClamp63((f.r * w));
uint8_t vg = dsmachine_rastClamp63((f.g * w));
uint8_t vb = dsmachine_rastClamp63((f.b * w));
auto tmp147 = std::make_tuple(vr,vg,vb);
uint8_t cr = std::get<0>(tmp147);
uint8_t cg = std::get<1>(tmp147);
uint8_t cb = std::get<2>(tmp147);
uint8_t ca = st->alpha;
if (st->hasTex) {
int64_t s = cast<int64_t>((f.s * w));
int64_t t = cast<int64_t>((f.t * w));
auto tmp148 = dsmachine_gpu3d_sampleTex(g,m,(&st->tex),shr<int64_t>(s,cast<int64_t>(4ULL)),shr<int64_t>(t,cast<int64_t>(4ULL)));
uint8_t tr = std::get<0>(tmp148);
uint8_t tg = std::get<1>(tmp148);
uint8_t tb = std::get<2>(tmp148);
uint8_t ta = std::get<3>(tmp148);
bool ok = std::get<4>(tmp148);
if ((!ok)) {
reject(false,true);
return ;
}
{
switch(st->mode){
case cast<uint32_t>(0ULL):{
cr = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tr) * cast<int64_t>(vr))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL)));
cg = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tg) * cast<int64_t>(vg))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL)));
cb = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tb) * cast<int64_t>(vb))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL)));
ca = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(ta) * cast<int64_t>(st->alpha))) + cast<int64_t>(15ULL)))),cast<int64_t>(31ULL)));
break;}
case cast<uint32_t>(1ULL):{
cr = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tr) * cast<int64_t>(ta))) + cast<int64_t>((cast<int64_t>(vr) * (cast<int64_t>((cast<int64_t>(31ULL) - cast<int64_t>(ta))))))))),cast<int64_t>(31ULL)));
cg = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tg) * cast<int64_t>(ta))) + cast<int64_t>((cast<int64_t>(vg) * (cast<int64_t>((cast<int64_t>(31ULL) - cast<int64_t>(ta))))))))),cast<int64_t>(31ULL)));
cb = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tb) * cast<int64_t>(ta))) + cast<int64_t>((cast<int64_t>(vb) * (cast<int64_t>((cast<int64_t>(31ULL) - cast<int64_t>(ta))))))))),cast<int64_t>(31ULL)));
ca = st->alpha;
break;}
case cast<uint32_t>(2ULL):{
auto tmp149 = dsmachine_polyState_toonShade(st,tr,tg,tb,ta,vr);
cr = std::get<0>(tmp149);
cg = std::get<1>(tmp149);
cb = std::get<2>(tmp149);
ca = std::get<3>(tmp149);
break;}
}}
}
else if ((st->mode == cast<uint32_t>(2ULL))) {
auto tmp150 = dsmachine_polyState_toonShade(st,cast<uint8_t>(63ULL),cast<uint8_t>(63ULL),cast<uint8_t>(63ULL),cast<uint8_t>(31ULL),vr);
cr = std::get<0>(tmp150);
cg = std::get<1>(tmp150);
cb = std::get<2>(tmp150);
ca = std::get<3>(tmp150);
}
if ((ca == cast<uint8_t>(0ULL))) {
reject(false,true);
return ;
}
if ((st->alphaTe && (ca <= st->alphaRf))) {
reject(false,true);
return ;
}
dsmachine_rfrag dst = g->rast.col[idx];
dsmachine_rfrag out = dsmachine_rfrag{cr,cg,cb,ca};
if (((st->trans && st->blend) && (dst.a != cast<uint8_t>(0ULL)))) {
int64_t a = cast<int64_t>(ca);
auto mix = [&](uint8_t s,uint8_t d)->uint8_t{
return cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(s) * a)) + cast<int64_t>((cast<int64_t>(d) * (cast<int64_t>((cast<int64_t>(31ULL) - a)))))))),cast<int64_t>(31ULL)));
}
;
out.r = mix(cr,dst.r);
out.g = mix(cg,dst.g);
out.b = mix(cb,dst.b);
if ((cast<int64_t>(dst.a) > a)) {
out.a = dst.a;
}
}
m->prof.frags++;
g->rast.col[idx] = out;
if (st->depthWr) {
g->rast.depth[idx] = depth;
}
if (bool(m->OnPixel)) {
m->OnPixel(modi<int64_t>(idx,cast<int64_t>(256ULL)),divi<int64_t>(idx,cast<int64_t>(256ULL)),dsmachine_PixelEvent{true,{},{},out.r,out.g,out.b,out.a});
}
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:619:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> dsmachine_polyState_toonShade(dsmachine_polyState* st,uint8_t tr,uint8_t tg,uint8_t tb,uint8_t ta,uint8_t vr){
{
std::array<uint8_t,3> tc = st->toon[shr<uint8_t>(vr,cast<int64_t>(1ULL))];
uint8_t a = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(ta) * cast<int64_t>(st->alpha))) + cast<int64_t>(15ULL)))),cast<int64_t>(31ULL)));
if ((!st->toonHi)) {
return {cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tr) * cast<int64_t>(tc[cast<int64_t>(0ULL)]))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tg) * cast<int64_t>(tc[cast<int64_t>(1ULL)]))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(tb) * cast<int64_t>(tc[cast<int64_t>(2ULL)]))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL))),a};
}
auto add = [&](uint8_t t,uint8_t v,uint8_t h)->uint8_t{
int64_t c = cast<int64_t>((divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(t) * cast<int64_t>(v))) + cast<int64_t>(31ULL)))),cast<int64_t>(63ULL)) + cast<int64_t>(h)));
if ((c > cast<int64_t>(63ULL))) {
c = cast<int64_t>(63ULL);
}
return cast<uint8_t>(c);
}
;
return {add(tr,vr,tc[cast<int64_t>(0ULL)]),add(tg,vr,tc[cast<int64_t>(1ULL)]),add(tb,vr,tc[cast<int64_t>(2ULL)]),a};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:637:1
uint8_t dsmachine_rastClamp63(double v){
{
if ((v < cast<double>(0ULL))) {
return cast<uint8_t>(0ULL);
}
if ((v > cast<double>(63ULL))) {
return cast<uint8_t>(63ULL);
}
return cast<uint8_t>((v + 0.5));
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:647:1
double dsmachine_rastCeil(double v){
{
double i = cast<double>(cast<int64_t>(v));
if ((v > i)) {
i++;
}
return i;
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:676:1
std::tuple<dsmachine_texState,bool> dsmachine_texStateOf(dsmachine_gxPolygon* p){
{
dsmachine_texState st={};
st.format = cast<uint32_t>(((shr<uint32_t>(p->texParam,cast<int64_t>(26ULL))) & cast<uint32_t>(7ULL)));
if ((st.format == cast<uint32_t>(0ULL))) {
return {st,false};
}
st.base = shl<uint32_t>((cast<uint32_t>((p->texParam & cast<uint32_t>(65535ULL)))),cast<int64_t>(3ULL));
st.repeatS = (cast<uint32_t>((p->texParam & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL));
st.repeatT = (cast<uint32_t>((p->texParam & cast<uint32_t>(131072ULL))) != cast<uint32_t>(0ULL));
st.flipS = (cast<uint32_t>((p->texParam & cast<uint32_t>(262144ULL))) != cast<uint32_t>(0ULL));
st.flipT = (cast<uint32_t>((p->texParam & cast<uint32_t>(524288ULL))) != cast<uint32_t>(0ULL));
st.sizeS = shl<int64_t>(cast<int64_t>(8ULL),(cast<uint32_t>(((shr<uint32_t>(p->texParam,cast<int64_t>(20ULL))) & cast<uint32_t>(7ULL)))));
st.sizeT = shl<int64_t>(cast<int64_t>(8ULL),(cast<uint32_t>(((shr<uint32_t>(p->texParam,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)))));
st.color0 = (cast<uint32_t>((p->texParam & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
uint32_t base = (cast<uint32_t>((p->pltt & cast<uint32_t>(8191ULL))));
if ((st.format == cast<uint32_t>(2ULL))) {
st.pal = shl<uint32_t>(base,cast<int64_t>(3ULL));
}
else {
st.pal = shl<uint32_t>(base,cast<int64_t>(4ULL));
}
return {st,true};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:709:1
int64_t dsmachine_wrapTex(int64_t v,int64_t size,bool repeat,bool flip){
{
if ((!repeat)) {
if ((v < cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
if ((v >= size)) {
return cast<int64_t>((size - cast<int64_t>(1ULL)));
}
return v;
}
int64_t mask = cast<int64_t>((size - cast<int64_t>(1ULL)));
if ((flip && (cast<int64_t>((v & size)) != cast<int64_t>(0ULL)))) {
return cast<int64_t>((mask - (cast<int64_t>((v & mask)))));
}
return cast<int64_t>((v & mask));
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:729:1
std::tuple<uint8_t,uint8_t,uint8_t> dsmachine_rastBGR555(uint16_t c){
{
return {cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>((c & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL)))),cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(c,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL)))),cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(c,cast<int64_t>(10ULL))) & cast<uint16_t>(31ULL)))) * cast<uint16_t>(2ULL))))};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:736:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> dsmachine_gpu3d_sampleTex(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_texState* st,int64_t s,int64_t t){
{
s = dsmachine_wrapTex(s,st->sizeS,st->repeatS,st->flipS);
t = dsmachine_wrapTex(t,st->sizeT,st->repeatT,st->flipT);
uint32_t i = cast<uint32_t>(cast<int64_t>((cast<int64_t>((t * st->sizeS)) + s)));
{
switch(st->format){
case cast<uint32_t>(1ULL):{
uint8_t b = dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + i)));
uint32_t idx = cast<uint32_t>(cast<uint8_t>((b & cast<uint8_t>(31ULL))));
uint32_t a3 = cast<uint32_t>(shr<uint8_t>(b,cast<int64_t>(5ULL)));
uint8_t a = cast<uint8_t>(cast<uint32_t>(((shl<uint32_t>(a3,cast<int64_t>(2ULL))) | (shr<uint32_t>(a3,cast<int64_t>(1ULL))))));
auto tmp151 = dsmachine_rastBGR555(dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((st->pal + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))))));
uint8_t r = std::get<0>(tmp151);
uint8_t gg = std::get<1>(tmp151);
uint8_t bb = std::get<2>(tmp151);
return {r,gg,bb,a,(a != cast<uint8_t>(0ULL))};
break;}
case cast<uint32_t>(2ULL):{
uint8_t b = dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + divi<uint32_t>(i,cast<uint32_t>(4ULL)))));
uint32_t idx = cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(b,(cast<uint32_t>((cast<uint32_t>(2ULL) * (cast<uint32_t>((i & cast<uint32_t>(3ULL))))))))) & cast<uint32_t>(3ULL)));
if (((idx == cast<uint32_t>(0ULL)) && st->color0)) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
auto tmp152 = dsmachine_rastBGR555(dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((st->pal + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))))));
uint8_t r = std::get<0>(tmp152);
uint8_t gg = std::get<1>(tmp152);
uint8_t bb = std::get<2>(tmp152);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint32_t>(3ULL):{
uint8_t b = dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + divi<uint32_t>(i,cast<uint32_t>(2ULL)))));
uint32_t idx = cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(b,(cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((i & cast<uint32_t>(1ULL))))))))) & cast<uint32_t>(15ULL)));
if (((idx == cast<uint32_t>(0ULL)) && st->color0)) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
auto tmp153 = dsmachine_rastBGR555(dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((st->pal + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))))));
uint8_t r = std::get<0>(tmp153);
uint8_t gg = std::get<1>(tmp153);
uint8_t bb = std::get<2>(tmp153);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint32_t>(4ULL):{
uint32_t idx = cast<uint32_t>(dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + i))));
if (((idx == cast<uint32_t>(0ULL)) && st->color0)) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
auto tmp154 = dsmachine_rastBGR555(dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((st->pal + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))))));
uint8_t r = std::get<0>(tmp154);
uint8_t gg = std::get<1>(tmp154);
uint8_t bb = std::get<2>(tmp154);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint32_t>(5ULL):{
return dsmachine_gpu3d_sampleCompressed(g,m,st,s,t);
break;}
case cast<uint32_t>(6ULL):{
uint8_t b = dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + i)));
uint32_t idx = cast<uint32_t>(cast<uint8_t>((b & cast<uint8_t>(7ULL))));
uint8_t a = cast<uint8_t>(shr<uint8_t>(b,cast<int64_t>(3ULL)));
auto tmp155 = dsmachine_rastBGR555(dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((st->pal + cast<uint32_t>((idx * cast<uint32_t>(2ULL)))))));
uint8_t r = std::get<0>(tmp155);
uint8_t gg = std::get<1>(tmp155);
uint8_t bb = std::get<2>(tmp155);
return {r,gg,bb,a,(a != cast<uint8_t>(0ULL))};
break;}
case cast<uint32_t>(7ULL):{
uint16_t c = dsmachine_vram_read16(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((st->base + cast<uint32_t>((i * cast<uint32_t>(2ULL))))));
if ((cast<uint16_t>((c & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL))) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
auto tmp156 = dsmachine_rastBGR555(c);
uint8_t r = std::get<0>(tmp156);
uint8_t gg = std::get<1>(tmp156);
uint8_t bb = std::get<2>(tmp156);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
}}
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:820:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> dsmachine_gpu3d_sampleCompressed(dsmachine_gpu3d* g,dsmachine_Machine* m,dsmachine_texState* st,int64_t s,int64_t t){
{
int64_t blocksW = divi<int64_t>(st->sizeS,cast<int64_t>(4ULL));
uint32_t blk = cast<uint32_t>(cast<int64_t>((cast<int64_t>(((divi<int64_t>(t,cast<int64_t>(4ULL))) * blocksW)) + (divi<int64_t>(s,cast<int64_t>(4ULL))))));
uint32_t addr = cast<uint32_t>((st->base + cast<uint32_t>((blk * cast<uint32_t>(4ULL)))));
uint8_t row = dsmachine_vram_read8(m->vram,cast<int64_t>(4ULL),cast<uint32_t>((addr + cast<uint32_t>(cast<int64_t>((t & cast<int64_t>(3ULL)))))));
uint32_t idx = cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(row,(cast<uint64_t>((cast<uint64_t>(2ULL) * cast<uint64_t>(cast<int64_t>((s & cast<int64_t>(3ULL))))))))) & cast<uint32_t>(3ULL)));
uint32_t slot = shr<uint32_t>(st->base,cast<int64_t>(17ULL));
uint32_t palIdxAddr = cast<uint32_t>((cast<uint32_t>(131072ULL) + divi<uint32_t>((cast<uint32_t>((addr & cast<uint32_t>(131071ULL)))),cast<uint32_t>(2ULL))));
if ((slot == cast<uint32_t>(2ULL))) {
palIdxAddr += cast<uint32_t>(65536ULL);
}
uint16_t info = dsmachine_vram_read16(m->vram,cast<int64_t>(4ULL),palIdxAddr);
uint32_t palAddr = cast<uint32_t>((st->pal + cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((info & cast<uint16_t>(16383ULL)))) * cast<uint32_t>(4ULL)))));
uint16_t mode = shr<uint16_t>(info,cast<int64_t>(14ULL));
auto raw = [&](uint32_t n)->uint16_t{
return dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),cast<uint32_t>((palAddr + cast<uint32_t>((n * cast<uint32_t>(2ULL))))));
}
;
{
switch(mode){
case cast<uint16_t>(0ULL):{
if ((idx == cast<uint32_t>(3ULL))) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
auto tmp157 = dsmachine_rastBGR555(raw(idx));
uint8_t r = std::get<0>(tmp157);
uint8_t gg = std::get<1>(tmp157);
uint8_t bb = std::get<2>(tmp157);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint16_t>(2ULL):{
auto tmp158 = dsmachine_rastBGR555(raw(idx));
uint8_t r = std::get<0>(tmp158);
uint8_t gg = std::get<1>(tmp158);
uint8_t bb = std::get<2>(tmp158);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint16_t>(1ULL):{
{
switch(idx){
case cast<uint32_t>(3ULL):{
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
break;}
case cast<uint32_t>(2ULL):{
auto tmp159 = dsmachine_rastMix555(raw(cast<uint32_t>(0ULL)),raw(cast<uint32_t>(1ULL)),cast<int64_t>(1ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL));
uint8_t r = std::get<0>(tmp159);
uint8_t gg = std::get<1>(tmp159);
uint8_t bb = std::get<2>(tmp159);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
default:{
auto tmp160 = dsmachine_rastBGR555(raw(idx));
uint8_t r = std::get<0>(tmp160);
uint8_t gg = std::get<1>(tmp160);
uint8_t bb = std::get<2>(tmp160);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
}}
break;}
default:{
{
switch(idx){
case cast<uint32_t>(2ULL):{
auto tmp161 = dsmachine_rastMix555(raw(cast<uint32_t>(0ULL)),raw(cast<uint32_t>(1ULL)),cast<int64_t>(5ULL),cast<int64_t>(3ULL),cast<int64_t>(3ULL));
uint8_t r = std::get<0>(tmp161);
uint8_t gg = std::get<1>(tmp161);
uint8_t bb = std::get<2>(tmp161);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
case cast<uint32_t>(3ULL):{
auto tmp162 = dsmachine_rastMix555(raw(cast<uint32_t>(0ULL)),raw(cast<uint32_t>(1ULL)),cast<int64_t>(3ULL),cast<int64_t>(5ULL),cast<int64_t>(3ULL));
uint8_t r = std::get<0>(tmp162);
uint8_t gg = std::get<1>(tmp162);
uint8_t bb = std::get<2>(tmp162);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
default:{
auto tmp163 = dsmachine_rastBGR555(raw(idx));
uint8_t r = std::get<0>(tmp163);
uint8_t gg = std::get<1>(tmp163);
uint8_t bb = std::get<2>(tmp163);
return {r,gg,bb,cast<uint8_t>(31ULL),true};
break;}
}}
break;}
}}
}
}
// tools/platform/nds/dsmachine/gpu3d_raster.go:881:1
std::tuple<uint8_t,uint8_t,uint8_t> dsmachine_rastMix555(uint16_t a,uint16_t b,int64_t wa,int64_t wb,int64_t shift){
{
auto ch = [&](uint64_t n)->uint8_t{
int64_t ca = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(a,(cast<uint64_t>((cast<uint64_t>(5ULL) * n))))) & cast<uint16_t>(31ULL))));
int64_t cb = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(b,(cast<uint64_t>((cast<uint64_t>(5ULL) * n))))) & cast<uint16_t>(31ULL))));
int64_t v = shr<int64_t>((cast<int64_t>((cast<int64_t>((ca * wa)) + cast<int64_t>((cb * wb))))),cast<uint64_t>(shift));
if ((v > cast<int64_t>(31ULL))) {
v = cast<int64_t>(31ULL);
}
return cast<uint8_t>(cast<int64_t>((v * cast<int64_t>(2ULL))));
}
;
return {ch(cast<uint64_t>(0ULL)),ch(cast<uint64_t>(1ULL)),ch(cast<uint64_t>(2ULL))};
}
}
// tools/platform/nds/dsmachine/input.go:46:1
std::tuple<uint32_t,bool> dsmachine_ParseKeys(std::string s){
{
uint32_t mask={};
{auto&& tmp164 = go_strings_Split(s,std::string(",",1));
for(int64_t tmp165=0;tmp165<len(tmp164);++tmp165){
auto f=tmp164[tmp165];f = go_strings_ToLower(go_strings_TrimSpace(f));
if ((f == std::string("",0))) {
continue;
}
auto tmp166 = lookup(dsmachine_keyNames,f);
uint32_t k = std::get<0>(tmp166);
bool ok = std::get<1>(tmp166);
if ((!ok)) {
return {cast<uint32_t>(0ULL),false};
}
mask |= k;
}}
return {mask,true};
}
}
// tools/platform/nds/dsmachine/input.go:63:1
void dsmachine_Machine_SetKeys(dsmachine_Machine* m,uint32_t mask){
{
m->keys = mask;
}
}
// tools/platform/nds/dsmachine/input.go:66:1
uint32_t dsmachine_Machine_keyinput(dsmachine_Machine* m){
{
return cast<uint32_t>((cast<uint32_t>(~m->keys) & cast<uint32_t>(1023ULL)));
}
}
// tools/platform/nds/dsmachine/input.go:79:1
uint32_t dsmachine_Machine_extkeyin(dsmachine_Machine* m){
{
uint32_t v = cast<uint32_t>(127ULL);
if ((cast<uint32_t>((m->keys & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
v &= ~(cast<uint32_t>(1ULL));
}
if ((cast<uint32_t>((m->keys & cast<uint32_t>(2048ULL))) != cast<uint32_t>(0ULL))) {
v &= ~(cast<uint32_t>(2ULL));
}
if (m->spi->touchDown) {
v &= ~(cast<uint32_t>(64ULL));
}
return v;
}
}
// tools/platform/nds/dsmachine/io.go:37:1
dsmachine_core* dsmachine_core_other(dsmachine_core* c){
{
if (c->arm9) {
return c->m->ARM7;
}
return c->m->ARM9;
}
}
// tools/platform/nds/dsmachine/io.go:45:1
uint32_t dsmachine_core_ioRead(dsmachine_core* c,uint32_t a){
{
dsmachine_Machine* m = c->m;
auto tmp167 = dsmachine_core_ioReadReg(c,a);
uint32_t v = std::get<0>(tmp167);
bool ok = std::get<1>(tmp167);
if (bool(m->OnIO)) {
m->OnIO(c->arm9,false,a,v,c->cpu->R[cast<int64_t>(15ULL)]);
}
if ((!ok)) {
dsmachine_Machine_note(m,std::string("%s: read unmodelled I/O 0x%08X",30),c->name,a);
}
return v;
}
}
// tools/platform/nds/dsmachine/io.go:57:1
std::tuple<uint32_t,bool> dsmachine_core_ioReadReg(dsmachine_core* c,uint32_t a){
{
dsmachine_Machine* m = c->m;
{
switch(a){
case cast<uint32_t>(67109384ULL):{
if (c->ime) {
return {cast<uint32_t>(1ULL),true};
}
return {cast<uint32_t>(0ULL),true};
break;}
case cast<uint32_t>(67109392ULL):{
return {c->ie,true};
break;}
case cast<uint32_t>(67109396ULL):{
return {c->if_,true};
break;}
case cast<uint32_t>(67108868ULL):{
return {cast<uint32_t>((dsmachine_core_dispstat(c) | shl<uint32_t>(cast<uint32_t>(cast<int64_t>((m->vid.line & cast<int64_t>(255ULL)))),cast<int64_t>(16ULL)))),true};
break;}
case cast<uint32_t>(67108870ULL):{
return {cast<uint32_t>(m->vid.line),true};
break;}
case cast<uint32_t>(67109168ULL):{
return {cast<uint32_t>((dsmachine_Machine_keyinput(m) | shl<uint32_t>(get(c->io,cast<uint32_t>(67109170ULL)),cast<int64_t>(16ULL)))),true};
break;}
case cast<uint32_t>(67109172ULL):{
if ((!c->arm9)) {
return {cast<uint32_t>((cast<uint32_t>((get(c->io,cast<uint32_t>(67109172ULL)) & cast<uint32_t>(65535ULL))) | shl<uint32_t>(dsmachine_Machine_extkeyin(m),cast<int64_t>(16ULL)))),true};
}
break;}
case cast<uint32_t>(67109248ULL):{
uint32_t in={};
uint32_t out={};
if (c->arm9) {
auto tmp168 = std::make_tuple(cast<uint32_t>(m->ipc.sync7),cast<uint32_t>(m->ipc.sync9));
in = std::get<0>(tmp168);
out = std::get<1>(tmp168);
}
else {
auto tmp169 = std::make_tuple(cast<uint32_t>(m->ipc.sync9),cast<uint32_t>(m->ipc.sync7));
in = std::get<0>(tmp169);
out = std::get<1>(tmp169);
}
return {cast<uint32_t>((cast<uint32_t>((in | shl<uint32_t>(out,cast<int64_t>(8ULL)))) | cast<uint32_t>((get(c->io,cast<uint32_t>(67109248ULL)) & cast<uint32_t>(24576ULL))))),true};
break;}
case cast<uint32_t>(67109252ULL):{
return {cast<uint32_t>(dsmachine_core_fifoCnt(c)),true};
break;}
case cast<uint32_t>(68157440ULL):{
return {dsmachine_core_fifoRecv(c),true};
break;}
case cast<uint32_t>(67109284ULL):{
if ((c == dsmachine_Machine_cardCore(m))) {
return {m->cd->ctrl,true};
}
return {cast<uint32_t>(0ULL),true};
break;}
case cast<uint32_t>(68157456ULL):{
if ((c == dsmachine_Machine_cardCore(m))) {
return {dsmachine_core_cardReadData(c),true};
}
return {cast<uint32_t>(4294967295ULL),true};
break;}
case cast<uint32_t>(67109280ULL):{
return {get(c->io,cast<uint32_t>(67109280ULL)),true};
break;}
case cast<uint32_t>(67109312ULL):{
if ((!c->arm9)) {
return {cast<uint32_t>(((get(c->io,cast<uint32_t>(67109312ULL)) & ~(cast<uint32_t>(128ULL))) | shl<uint32_t>(cast<uint32_t>(m->spi->out),cast<int64_t>(16ULL)))),true};
}
break;}
case cast<uint32_t>(67109176ULL):{
if ((!c->arm9)) {
return {cast<uint32_t>(0ULL),true};
}
break;}
case cast<uint32_t>(67109504ULL):{
return {m->div.cnt,true};
break;}
case cast<uint32_t>(67109552ULL):{
return {m->sqrt.cnt,true};
break;}
case cast<uint32_t>(67109556ULL):{
return {m->sqrt.result,true};
break;}
case cast<uint32_t>(67109632ULL):{
return {cast<uint32_t>(1ULL),true};
break;}
case cast<uint32_t>(67109636ULL):{
return {m->powcnt,true};
break;}
case cast<uint32_t>(67109380ULL):{
return {get(c->io,cast<uint32_t>(67109380ULL)),true};
break;}
case cast<uint32_t>(67109170ULL):{
return {get(c->io,cast<uint32_t>(67109170ULL)),true};
break;}
}}
{
if (((a >= cast<uint32_t>(67109520ULL)) && (a < cast<uint32_t>(67109552ULL)))){
return {dsmachine_Machine_divRead(m,a),true};
}
else if (((a >= cast<uint32_t>(67109560ULL)) && (a < cast<uint32_t>(67109568ULL)))){
return {dsmachine_half64(m->sqrt.param,cast<uint32_t>((a - cast<uint32_t>(67109560ULL)))),true};
}
else if (((a >= cast<uint32_t>(67109040ULL)) && (a < cast<uint32_t>(67109088ULL)))){
return {dsmachine_core_dmaRead(c,a),true};
}
else if (((a >= cast<uint32_t>(67109120ULL)) && (a < cast<uint32_t>(67109136ULL)))){
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(67109120ULL)))),cast<int64_t>(4ULL));
return {cast<uint32_t>((cast<uint32_t>(c->timers[n].counter) | shl<uint32_t>(cast<uint32_t>(c->timers[n].ctrl),cast<int64_t>(16ULL)))),true};
}
else if ((((a >= cast<uint32_t>(67109664ULL)) && (a < cast<uint32_t>(67110564ULL))) && c->arm9)){
return {dsmachine_gpu3d_readReg(m->gpu3d,a),true};
}
else if (((a >= cast<uint32_t>(67108864ULL)) && (a < cast<uint32_t>(67108976ULL))) || ((a >= cast<uint32_t>(67112960ULL)) && (a < cast<uint32_t>(67113072ULL)))){
return {get(c->io,a),true};
}
else if ((((a >= cast<uint32_t>(67109888ULL)) && (a < cast<uint32_t>(67110176ULL))) && (!c->arm9))){
return {get(c->io,a),true};
}
}
tmp170:;
return {get(c->io,a),false};
}
}
// tools/platform/nds/dsmachine/io.go:158:1
void dsmachine_core_ioWrite(dsmachine_core* c,uint32_t a,uint8_t v){
{
dsmachine_Machine* m = c->m;
if (bool(m->OnIO)) {
m->OnIO(c->arm9,true,a,cast<uint32_t>(v),c->cpu->R[cast<int64_t>(15ULL)]);
}
if ((((a >= cast<uint32_t>(67109440ULL)) && (a <= cast<uint32_t>(67109449ULL))) && c->arm9)) {
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(3ULL))))));
c->io[(a & ~(cast<uint32_t>(3ULL)))] = cast<uint32_t>(((get(c->io,(a & ~(cast<uint32_t>(3ULL)))) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | shl<uint32_t>(cast<uint32_t>(v),sh)));
if ((a == cast<uint32_t>(67109447ULL))) {
m->wramcnt = v;
}
else {
dsmachine_vram_setCNT(m->vram,dsmachine_vramBankIndex(a),v);
}
return ;
}
uint32_t base = (a & ~(cast<uint32_t>(3ULL)));
uint32_t shift = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(3ULL))))));
uint32_t lane = cast<uint32_t>((a & cast<uint32_t>(3ULL)));
c->io[base] = cast<uint32_t>(((get(c->io,base) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),shift)))) | shl<uint32_t>(cast<uint32_t>(v),shift)));
uint32_t w = get(c->io,base);
bool handled = true;
{
switch(base){
case cast<uint32_t>(67109384ULL):{
c->ime = (cast<uint32_t>((w & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(67109392ULL):{
c->ie = w;
break;}
case cast<uint32_t>(67109396ULL):{
c->if_ &= ~(shl<uint32_t>(cast<uint32_t>(v),shift));
c->io[base] = c->if_;
break;}
case cast<uint32_t>(67108868ULL):{
c->io[cast<uint32_t>(67108868ULL)] = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(67109248ULL):{
if ((lane != cast<uint32_t>(1ULL))) {
return ;
}
uint8_t out = cast<uint8_t>((cast<uint8_t>(shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint8_t>(15ULL)));
if (c->arm9) {
m->ipc.sync9 = out;
}
else {
m->ipc.sync7 = out;
}
if (bool(m->SyncTrace)) {
m->SyncTrace(c->name,out,c->cpu->R[cast<int64_t>(15ULL)]);
}
if ((cast<uint32_t>((w & cast<uint32_t>(8192ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core* o = dsmachine_core_other(c);
if ((cast<uint32_t>((get(o->io,cast<uint32_t>(67109248ULL)) & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(o,cast<uint32_t>(65536ULL));
}
}
break;}
case cast<uint32_t>(67109256ULL):{
if ((lane == cast<uint32_t>(3ULL))) {
dsmachine_core_fifoSend(c,w);
}
break;}
case cast<uint32_t>(67109252ULL):{
if ((lane == cast<uint32_t>(1ULL))) {
dsmachine_core_fifoCntWrite(c,cast<uint16_t>(w));
}
break;}
case cast<uint32_t>(67109284ULL):{
if ((((lane == cast<uint32_t>(3ULL)) && (c == dsmachine_Machine_cardCore(m))) && (cast<uint32_t>((w & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
dsmachine_card_start(m->cd,w);
dsmachine_Machine_runDMA(m,cast<int64_t>(5ULL));
}
break;}
case cast<uint32_t>(67109312ULL):{
if (((lane == cast<uint32_t>(2ULL)) && (!c->arm9))) {
dsmachine_core_spiTransfer(c,v);
}
break;}
case cast<uint32_t>(67109504ULL):{
m->div.cnt = w;
dsmachine_divider_run(&(m->div));
break;}
case cast<uint32_t>(67109552ULL):{
m->sqrt.cnt = w;
dsmachine_sqrter_run(&(m->sqrt));
break;}
case cast<uint32_t>(67109636ULL):{
m->powcnt = w;
break;}
default:{
handled = false;
break;}
}}
if (handled) {
return ;
}
{
if (((base >= cast<uint32_t>(67109288ULL)) && (base < cast<uint32_t>(67109296ULL)))){
m->cd->cmd[cast<uint32_t>((a - cast<uint32_t>(67109288ULL)))] = v;
}
else if (((base >= cast<uint32_t>(67109520ULL)) && (base < cast<uint32_t>(67109552ULL)))){
dsmachine_Machine_divWrite(m,base,w);
}
else if (((base >= cast<uint32_t>(67109560ULL)) && (base < cast<uint32_t>(67109568ULL)))){
m->sqrt.param = dsmachine_setHalf64(m->sqrt.param,cast<uint32_t>((base - cast<uint32_t>(67109560ULL))),w);
dsmachine_sqrter_run(&(m->sqrt));
}
else if (((base >= cast<uint32_t>(67109040ULL)) && (base < cast<uint32_t>(67109088ULL)))){
dsmachine_core_dmaWrite(c,base,w,lane);
}
else if (((base >= cast<uint32_t>(67109120ULL)) && (base < cast<uint32_t>(67109136ULL)))){
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((base - cast<uint32_t>(67109120ULL)))),cast<int64_t>(4ULL));
if ((lane <= cast<uint32_t>(1ULL))) {
c->timers[n].reload = cast<uint16_t>(w);
}
if ((lane == cast<uint32_t>(3ULL))) {
dsmachine_core_writeTimerCtrl(c,n,cast<uint16_t>(shr<uint32_t>(w,cast<int64_t>(16ULL))));
}
}
else if ((((base >= cast<uint32_t>(67109664ULL)) && (base < cast<uint32_t>(67110564ULL))) && c->arm9)){
if ((lane == cast<uint32_t>(3ULL))) {
dsmachine_gpu3d_writeReg(m->gpu3d,base,w);
}
}
else if (((base >= cast<uint32_t>(67108864ULL)) && (base < cast<uint32_t>(67108976ULL))) || ((base >= cast<uint32_t>(67112960ULL)) && (base < cast<uint32_t>(67113072ULL)))){
}
else if ((((base >= cast<uint32_t>(67109888ULL)) && (base < cast<uint32_t>(67110176ULL))) && (!c->arm9))){
if (((((lane == cast<uint32_t>(3ULL)) && (base < cast<uint32_t>(67110144ULL))) && (modi<uint32_t>((cast<uint32_t>((base - cast<uint32_t>(67109888ULL)))),cast<uint32_t>(16ULL)) == cast<uint32_t>(0ULL))) && (cast<uint32_t>((w & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
dsmachine_Machine_soundKeyed(m);
}
}
else {
if ((!get(dsmachine_ioKnown,base))) {
dsmachine_Machine_note(m,std::string("%s: wrote unmodelled I/O 0x%08X = 0x%08X",40),c->name,base,w);
}
}
}
tmp171:;
}
}
// tools/platform/nds/dsmachine/io.go:293:1
int64_t dsmachine_vramBankIndex(uint32_t a){
{
if ((a <= cast<uint32_t>(67109446ULL))) {
return cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(67109440ULL))));
}
return cast<int64_t>((cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(67109448ULL)))) + cast<int64_t>(7ULL)));
}
}
// tools/platform/nds/dsmachine/io.go:310:1
uint32_t dsmachine_half64(uint64_t v,uint32_t off){
{
if ((off == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(v);
}
return cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL)));
}
}
// tools/platform/nds/dsmachine/io.go:317:1
uint64_t dsmachine_setHalf64(uint64_t v,uint32_t off,uint32_t w){
{
if ((off == cast<uint32_t>(0ULL))) {
return cast<uint64_t>(((v & ~(cast<uint64_t>(4294967295ULL))) | cast<uint64_t>(w)));
}
return cast<uint64_t>((cast<uint64_t>((v & cast<uint64_t>(4294967295ULL))) | shl<uint64_t>(cast<uint64_t>(w),cast<int64_t>(32ULL))));
}
}
// tools/platform/nds/dsmachine/io.go:324:1
uint32_t dsmachine_Machine_divRead(dsmachine_Machine* m,uint32_t a){
{
{
if ((a < cast<uint32_t>(67109528ULL))){
return dsmachine_half64(m->div.numer,cast<uint32_t>((a - cast<uint32_t>(67109520ULL))));
}
else if ((a < cast<uint32_t>(67109536ULL))){
return dsmachine_half64(m->div.denom,cast<uint32_t>((a - cast<uint32_t>(67109528ULL))));
}
else if ((a < cast<uint32_t>(67109544ULL))){
return dsmachine_half64(m->div.result,cast<uint32_t>((a - cast<uint32_t>(67109536ULL))));
}
else {
return dsmachine_half64(m->div.rem,cast<uint32_t>((a - cast<uint32_t>(67109544ULL))));
}
}
tmp172:;
}
}
// tools/platform/nds/dsmachine/io.go:337:1
void dsmachine_Machine_divWrite(dsmachine_Machine* m,uint32_t a,uint32_t w){
{
{
if ((a < cast<uint32_t>(67109528ULL))){
m->div.numer = dsmachine_setHalf64(m->div.numer,cast<uint32_t>((a - cast<uint32_t>(67109520ULL))),w);
}
else if ((a < cast<uint32_t>(67109536ULL))){
m->div.denom = dsmachine_setHalf64(m->div.denom,cast<uint32_t>((a - cast<uint32_t>(67109528ULL))),w);
}
else {
return ;
}
}
tmp173:;
dsmachine_divider_run(&(m->div));
}
}
// tools/platform/nds/dsmachine/io.go:351:1
uint32_t dsmachine_core_dmaRead(dsmachine_core* c,uint32_t a){
{
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(67109040ULL)))),cast<int64_t>(12ULL));
if ((n > cast<int64_t>(3ULL))) {
return cast<uint32_t>(0ULL);
}
{
switch(modi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67109040ULL)))),cast<uint32_t>(12ULL))){
case cast<uint32_t>(0ULL):{
return c->dma[n].src;
break;}
case cast<uint32_t>(4ULL):{
return c->dma[n].dst;
break;}
default:{
return cast<uint32_t>((c->dma[n].count | shl<uint32_t>(c->dma[n].ctrl,cast<int64_t>(16ULL))));
break;}
}}
}
}
// tools/platform/nds/dsmachine/io.go:366:1
void dsmachine_core_dmaWrite(dsmachine_core* c,uint32_t a,uint32_t w,uint32_t lane){
{
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(67109040ULL)))),cast<int64_t>(12ULL));
if ((n > cast<int64_t>(3ULL))) {
return ;
}
{
switch(modi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(67109040ULL)))),cast<uint32_t>(12ULL))){
case cast<uint32_t>(0ULL):{
c->dma[n].src = w;
break;}
case cast<uint32_t>(4ULL):{
c->dma[n].dst = w;
break;}
default:{
c->dma[n].count = cast<uint32_t>((w & dsmachine_core_maxCount(c)));
c->dma[n].ctrl = shr<uint32_t>(w,cast<int64_t>(16ULL));
if ((lane == cast<uint32_t>(3ULL))) {
dsmachine_core_writeDMACtrl(c,n);
}
break;}
}}
}
}
// tools/platform/nds/dsmachine/io.go:387:1
Slice<uint32_t>* dsmachine_core_sendQ(dsmachine_core* c){
{
if (c->arm9) {
return (&c->m->ipc.to7);
}
return (&c->m->ipc.to9);
}
}
// tools/platform/nds/dsmachine/io.go:394:1
Slice<uint32_t>* dsmachine_core_recvQ(dsmachine_core* c){
{
if (c->arm9) {
return (&c->m->ipc.to9);
}
return (&c->m->ipc.to7);
}
}
// tools/platform/nds/dsmachine/io.go:401:1
void dsmachine_core_fifoSend(dsmachine_core* c,uint32_t v){
{
if ((cast<uint32_t>((get(c->io,cast<uint32_t>(67109252ULL)) & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
Slice<uint32_t>* q = dsmachine_core_sendQ(c);
if ((len((*q)) >= cast<int64_t>(16ULL))) {
return ;
}
(*q) = append((*q),v);
dsmachine_core* o = dsmachine_core_other(c);
if ((cast<uint32_t>((get(o->io,cast<uint32_t>(67109252ULL)) & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(o,cast<uint32_t>(262144ULL));
}
}
}
// tools/platform/nds/dsmachine/io.go:416:1
uint32_t dsmachine_core_fifoRecv(dsmachine_core* c){
{
Slice<uint32_t>* q = dsmachine_core_recvQ(c);
if ((len((*q)) == cast<int64_t>(0ULL))) {
return c->lastRecv;
}
uint32_t v = ((*q))[cast<int64_t>(0ULL)];
(*q) = sub(((*q)),cast<int64_t>(1ULL),len(((*q))));
c->lastRecv = v;
if ((len((*q)) == cast<int64_t>(0ULL))) {
dsmachine_core* o = dsmachine_core_other(c);
if ((cast<uint32_t>((get(o->io,cast<uint32_t>(67109252ULL)) & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(o,cast<uint32_t>(131072ULL));
}
}
return v;
}
}
// tools/platform/nds/dsmachine/io.go:433:1
uint16_t dsmachine_core_fifoCnt(dsmachine_core* c){
{
auto tmp174 = std::make_tuple((*dsmachine_core_sendQ(c)),(*dsmachine_core_recvQ(c)));
Slice<uint32_t> send = std::get<0>(tmp174);
Slice<uint32_t> recv = std::get<1>(tmp174);
uint16_t v = cast<uint16_t>((cast<uint16_t>(get(c->io,cast<uint32_t>(67109252ULL))) & cast<uint16_t>(48132ULL)));
if ((len(send) == cast<int64_t>(0ULL))) {
v |= cast<uint16_t>(1ULL);
}
if ((len(send) >= cast<int64_t>(16ULL))) {
v |= cast<uint16_t>(2ULL);
}
if ((len(recv) == cast<int64_t>(0ULL))) {
v |= cast<uint16_t>(256ULL);
}
if ((len(recv) >= cast<int64_t>(16ULL))) {
v |= cast<uint16_t>(512ULL);
}
return v;
}
}
// tools/platform/nds/dsmachine/io.go:451:1
void dsmachine_core_fifoCntWrite(dsmachine_core* c,uint16_t v){
{
if ((cast<uint16_t>((v & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL))) {
(*dsmachine_core_sendQ(c)) = sub(((*dsmachine_core_sendQ(c))),0,cast<int64_t>(0ULL));
}
c->io[cast<uint32_t>(67109252ULL)] = cast<uint32_t>(v);
}
}
// tools/platform/nds/dsmachine/machine.go:181:1
dsmachine_Machine* dsmachine_New(nds_ROM* rom,uint32_t dtcm9Base){
{
dsmachine_Machine* m = arenaNew(dsmachine_Machine{Slice<uint8_t>::make(cast<int64_t>(4194304ULL)),Slice<uint8_t>::make(cast<int64_t>(32768ULL)),Slice<uint8_t>::make(cast<int64_t>(2048ULL)),Slice<uint8_t>::make(cast<int64_t>(2048ULL)),dsmachine_newVRAM(),arenaNew(dsmachine_card{rom->Data,{},{},{},{},{},{}}),dsmachine_newSPI(),{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
m->gpu2d = dsmachine_newGPU2D();
m->gpu3d = dsmachine_newGPU3D();
m->vid.line = cast<int64_t>(262ULL);
m->wramcnt = cast<uint8_t>(3ULL);
dsmachine_core* a9 = arenaNew(dsmachine_core{m,{},std::string("ARM9",4),true,Slice<uint8_t>::make(cast<int64_t>(32768ULL)),cast<uint32_t>(33521664ULL),{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,uint32_t>{},{},{},{}});
if ((dtcm9Base != cast<uint32_t>(0ULL))) {
a9->dtcm = Slice<uint8_t>::make(cast<int64_t>(16384ULL));
a9->dtcmBase = dtcm9Base;
a9->handlerBase = cast<uint32_t>((dtcm9Base + cast<uint32_t>(16384ULL)));
}
a9->cpu = arm_NewCPU(arenaNew(dsmachine_bus{a9}));
a9->cpu->Mode = cast<uint32_t>(19ULL);
a9->cpu->R[cast<int64_t>(15ULL)] = rom->Header.ARM9Entry;
a9->cpu->SWI = dsmachine_biosSWI(a9);
a9->cpu->Coproc = dsmachine_cp15(a9);
dsmachine_copyInto(m,a9,rom->Header.ARM9RAMAddr,nds_ROM_ARM9(rom));
m->ARM9 = a9;
dsmachine_core* a7 = arenaNew(dsmachine_core{m,{},std::string("ARM7",4),false,{},{},{},{},Slice<uint8_t>::make(cast<int64_t>(16384ULL)),Slice<uint8_t>::make(cast<int64_t>(65536ULL)),{},{},{},{},{},{},{},{},cast<uint32_t>(58785792ULL),Map<uint32_t,uint32_t>{},{},{},{}});
a7->cpu = arm_NewCPU(arenaNew(dsmachine_bus{a7}));
a7->cpu->Mode = cast<uint32_t>(19ULL);
a7->cpu->R[cast<int64_t>(15ULL)] = rom->Header.ARM7Entry;
a7->cpu->SWI = dsmachine_biosSWI(a7);
dsmachine_copyInto(m,a7,rom->Header.ARM7RAMAddr,nds_ROM_ARM7(rom));
m->ARM7 = a7;
dsmachine_Machine_directBoot(m,rom);
return m;
}
}
// tools/platform/nds/dsmachine/machine.go:243:1
void dsmachine_copyInto(dsmachine_Machine* m,dsmachine_core* c,uint32_t addr,Slice<uint8_t> data){
{
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
{auto&& tmp175 = data;
for(int64_t tmp176=0;tmp176<len(tmp175);++tmp176){
auto i=tmp176;auto v=tmp175[tmp176];dsmachine_bus_Write(b,cast<uint32_t>((addr + cast<uint32_t>(i))),v);
}}
}
}
// tools/platform/nds/dsmachine/machine.go:251:1
uint32_t dsmachine_Machine_ARM9PC(dsmachine_Machine* m){
{
return m->ARM9->cpu->R[cast<int64_t>(15ULL)];
}
}
// tools/platform/nds/dsmachine/machine.go:252:1
uint32_t dsmachine_Machine_ARM7PC(dsmachine_Machine* m){
{
return m->ARM7->cpu->R[cast<int64_t>(15ULL)];
}
}
// tools/platform/nds/dsmachine/machine.go:255:1
std::tuple<uint8_t,uint8_t> dsmachine_Machine_SyncNibbles(dsmachine_Machine* m){
uint8_t arm9{};
uint8_t arm7{};
{
return {m->ipc.sync9,m->ipc.sync7};
}
}
// tools/platform/nds/dsmachine/machine.go:258:1
std::tuple<int64_t,int64_t> dsmachine_Machine_FifoLens(dsmachine_Machine* m){
int64_t to7{};
int64_t to9{};
{
return {len(m->ipc.to7),len(m->ipc.to9)};
}
}
// tools/platform/nds/dsmachine/machine.go:262:1
uint64_t dsmachine_Machine_Frame(dsmachine_Machine* m){
{
return m->vid.frames;
}
}
// tools/platform/nds/dsmachine/machine.go:263:1
int64_t dsmachine_Machine_Line(dsmachine_Machine* m){
{
return m->vid.line;
}
}
// tools/platform/nds/dsmachine/machine.go:267:1
Slice<uint8_t> dsmachine_Machine_Snapshot(dsmachine_Machine* m,bool arm9,uint32_t addr,uint32_t n){
{
dsmachine_core* c = m->ARM7;
if (arm9) {
c = m->ARM9;
}
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
Slice<uint8_t> out = Slice<uint8_t>::make(n);
{auto&& tmp177 = out;
for(int64_t tmp178=0;tmp178<len(tmp177);++tmp178){
auto i=tmp178;out[i] = dsmachine_bus_Read(b,cast<uint32_t>((addr + cast<uint32_t>(i))));
}}
return out;
}
}
// tools/platform/nds/dsmachine/machine.go:282:1
void dsmachine_Machine_Poke(dsmachine_Machine* m,bool arm9,uint32_t addr,Slice<uint8_t> data){
{
dsmachine_core* c = m->ARM7;
if (arm9) {
c = m->ARM9;
}
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
{auto&& tmp179 = data;
for(int64_t tmp180=0;tmp180<len(tmp179);++tmp180){
auto i=tmp180;auto v=tmp179[tmp180];dsmachine_bus_Write(b,cast<uint32_t>((addr + cast<uint32_t>(i))),v);
}}
}
}
// tools/platform/nds/dsmachine/machine.go:294:1
std::array<uint32_t,16> dsmachine_Machine_Regs(dsmachine_Machine* m,bool arm9){
{
if (arm9) {
return m->ARM9->cpu->R;
}
return m->ARM7->cpu->R;
}
}
// tools/platform/nds/dsmachine/machine.go:303:1
bool dsmachine_Machine_Thumb(dsmachine_Machine* m,bool arm9){
{
if (arm9) {
return m->ARM9->cpu->Thumb;
}
return m->ARM7->cpu->Thumb;
}
}
// tools/platform/nds/dsmachine/machine.go:313:1
std::tuple<uint32_t,uint32_t,bool> dsmachine_Machine_IRQState(dsmachine_Machine* m,bool arm9){
uint32_t ie{};
uint32_t if_{};
bool ime{};
{
dsmachine_core* c = m->ARM7;
if (arm9) {
c = m->ARM9;
}
return {c->ie,c->if_,c->ime};
}
}
// tools/platform/nds/dsmachine/machine.go:322:1
bool dsmachine_Machine_Parked(dsmachine_Machine* m,bool arm9){
{
if (arm9) {
return m->ARM9->waiting;
}
return m->ARM7->waiting;
}
}
// tools/platform/nds/dsmachine/machine.go:329:1
void dsmachine_Machine_onFrame(dsmachine_Machine* m){
{
if (bool(m->OnFrame)) {
m->OnFrame();
}
}
}
// tools/platform/nds/dsmachine/machine.go:346:1
bool dsmachine_Machine_IRQDisabled(dsmachine_Machine* m,bool arm9){
{
if (arm9) {
return m->ARM9->cpu->IRQDisable;
}
return m->ARM7->cpu->IRQDisable;
}
}
// tools/platform/nds/dsmachine/machine.go:355:1
uint32_t dsmachine_Machine_Reg(dsmachine_Machine* m,uint32_t a){
{
return get(m->ARM9->io,(a & ~(cast<uint32_t>(3ULL))));
}
}
// tools/platform/nds/dsmachine/machine.go:359:1
int64_t dsmachine_Machine_Sleep(dsmachine_Machine* m,bool arm9){
{
if (arm9) {
return m->ARM9->sleep;
}
return m->ARM7->sleep;
}
}
// tools/platform/nds/dsmachine/machine.go:369:1
void dsmachine_Machine_OnCardXfer(dsmachine_Machine* m,std::function<void(std::array<uint8_t,8>,uint32_t,int64_t)> f){
{
m->cd->OnXfer = f;
}
}
// tools/platform/nds/dsmachine/machine.go:376:1
std::tuple<int64_t,int64_t> dsmachine_Machine_GX(dsmachine_Machine* m){
int64_t polys{};
int64_t swaps{};
{
return {m->gpu3d->lastPolys,m->gpu3d->swaps};
}
}
// tools/platform/nds/dsmachine/machine.go:382:1
std::array<int64_t,256> dsmachine_Machine_GXHist(dsmachine_Machine* m){
{
return m->gpu3d->cmdHist;
}
}
// tools/platform/nds/dsmachine/machine.go:387:1
std::tuple<int64_t,int64_t> dsmachine_Machine_GXClip(dsmachine_Machine* m){
int64_t emitted{};
int64_t clipped{};
{
return {m->gpu3d->geom.nEmit,m->gpu3d->geom.nClipped};
}
}
// tools/platform/nds/dsmachine/machine.go:392:1
uint32_t dsmachine_Machine_Reg7(dsmachine_Machine* m,uint32_t a){
{
return get(m->ARM7->io,(a & ~(cast<uint32_t>(3ULL))));
}
}
// tools/platform/nds/dsmachine/machine.go:395:1
Map<uint32_t,uint32_t> dsmachine_Machine_GXRegs(dsmachine_Machine* m){
{
return m->gpu3d->regs;
}
}
// tools/platform/nds/dsmachine/machine.go:399:1
void dsmachine_Machine_OnGXCmd(dsmachine_Machine* m,std::function<void(uint8_t,Slice<uint32_t>)> f){
{
m->gpu3d->OnCmd = f;
}
}
// tools/platform/nds/dsmachine/machine.go:403:1
uint16_t dsmachine_Machine_VRAMTexPal(dsmachine_Machine* m,uint32_t off){
{
return dsmachine_vram_read16(m->vram,cast<int64_t>(5ULL),off);
}
}
// tools/platform/nds/dsmachine/math.go:44:1
void dsmachine_divider_run(dsmachine_divider* d){
{
uint32_t mode = cast<uint32_t>((d->cnt & cast<uint32_t>(3ULL)));
d->cnt &= ~(cast<uint32_t>(16384ULL));
int64_t num={};
int64_t den={};
{
switch(mode){
case cast<uint32_t>(0ULL):{
auto tmp181 = std::make_tuple(cast<int64_t>(cast<int32_t>(cast<uint32_t>(d->numer))),cast<int64_t>(cast<int32_t>(cast<uint32_t>(d->denom))));
num = std::get<0>(tmp181);
den = std::get<1>(tmp181);
break;}
case cast<uint32_t>(1ULL):{
auto tmp182 = std::make_tuple(cast<int64_t>(d->numer),cast<int64_t>(cast<int32_t>(cast<uint32_t>(d->denom))));
num = std::get<0>(tmp182);
den = std::get<1>(tmp182);
break;}
default:{
auto tmp183 = std::make_tuple(cast<int64_t>(d->numer),cast<int64_t>(d->denom));
num = std::get<0>(tmp183);
den = std::get<1>(tmp183);
break;}
}}
if ((den == cast<int64_t>(0ULL))) {
d->cnt |= cast<uint32_t>(16384ULL);
d->rem = cast<uint64_t>(num);
if ((num < cast<int64_t>(0ULL))) {
d->result = cast<uint64_t>(1ULL);
}
else {
d->result = cast<uint64_t>(18446744073709551615ULL);
}
if ((mode == cast<uint32_t>(0ULL))) {
if ((num < cast<int64_t>(0ULL))) {
d->result = cast<uint64_t>(1ULL);
}
else {
d->result = cast<uint64_t>(18446744073709551615ULL);
}
}
return ;
}
if (((num == cast<int64_t>(-9223372036854775808ULL)) && (den == cast<int64_t>(-1ULL)))) {
auto tmp184 = std::make_tuple(cast<uint64_t>(num),cast<uint64_t>(0ULL));
d->result = std::get<0>(tmp184);
d->rem = std::get<1>(tmp184);
return ;
}
d->result = cast<uint64_t>(divi<int64_t>(num,den));
d->rem = cast<uint64_t>(modi<int64_t>(num,den));
}
}
// tools/platform/nds/dsmachine/math.go:98:1
void dsmachine_sqrter_run(dsmachine_sqrter* s){
{
uint64_t v = s->param;
if ((cast<uint32_t>((s->cnt & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
v &= cast<uint64_t>(4294967295ULL);
}
uint64_t res=cast<uint64_t>(0ULL);
uint64_t bit=cast<uint64_t>(4611686018427387904ULL);
{;for (;(bit > v);){
bit = shr<uint64_t>(bit,cast<int64_t>(2ULL));
}
}{;for (;(bit != cast<uint64_t>(0ULL));){
if ((v >= cast<uint64_t>((res + bit)))) {
v -= cast<uint64_t>((res + bit));
res = cast<uint64_t>((shr<uint64_t>(res,cast<int64_t>(1ULL)) + bit));
}
else {
res = shr<uint64_t>(res,cast<int64_t>(1ULL));
}
bit = shr<uint64_t>(bit,cast<int64_t>(2ULL));
}
}s->result = cast<uint32_t>(res);
}
}
// tools/platform/nds/dsmachine/profile.go:57:1
void dsmachine_Machine_SetProfile(dsmachine_Machine* m,bool on){
{
m->prof.on = on;
dsmachine_profiler_reset(&(m->prof),m);
}
}
// tools/platform/nds/dsmachine/profile.go:62:1
void dsmachine_profiler_reset(dsmachine_profiler* p,dsmachine_Machine* m){
{
p->frameStart = go_time_Now();
auto tmp185 = std::make_tuple(cast<time_Duration>(0ULL),cast<time_Duration>(0ULL),cast<time_Duration>(0ULL),cast<time_Duration>(0ULL));
p->geom = std::get<0>(tmp185);
p->raster = std::get<1>(tmp185);
p->compose = std::get<2>(tmp185);
p->dma = std::get<3>(tmp185);
auto tmp186 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL));
p->polys = std::get<0>(tmp186);
p->frags = std::get<1>(tmp186);
p->xfers = std::get<2>(tmp186);
p->cmdStart = m->gpu3d->count;
}
}
// tools/platform/nds/dsmachine/profile.go:72:1
dsmachine_Profile dsmachine_Machine_FrameProfile(dsmachine_Machine* m){
{
dsmachine_profiler* p = (&m->prof);
time_Duration total = go_time_Since(p->frameStart);
auto ms = [&](time_Duration d)->double{
return (cast<double>(time_Duration_Nanoseconds(d)) / 1e6);
}
;
dsmachine_Profile pr = dsmachine_Profile{ms(total),ms(p->geom),ms(p->raster),ms(p->compose),ms(p->dma),{},cast<int64_t>((m->gpu3d->count - p->cmdStart)),p->polys,p->frags,p->xfers,m->vid.frames};
pr.CPUMs = ((((pr.TotalMs - pr.GeometryMs) - pr.RasterMs) - pr.ComposeMs) - pr.DMAMs);
if ((pr.CPUMs < cast<double>(0ULL))) {
pr.CPUMs = cast<double>(0ULL);
}
return pr;
}
}
// tools/platform/nds/dsmachine/run.go:39:1
dsmachine_Result dsmachine_Machine_Run(dsmachine_Machine* m,uint64_t budget,int64_t quantum,Map<uint32_t,std::string> milestones){
{
return dsmachine_Machine_run(m,budget,quantum,milestones,cast<uint64_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/run.go:47:1
dsmachine_Result dsmachine_Machine_RunFrames(dsmachine_Machine* m,uint64_t n,uint64_t budget,int64_t quantum){
{
return dsmachine_Machine_run(m,budget,quantum,{},cast<uint64_t>((m->vid.frames + n)));
}
}
// tools/platform/nds/dsmachine/run.go:51:1
dsmachine_Result dsmachine_Machine_run(dsmachine_Machine* m,uint64_t budget,int64_t quantum,Map<uint32_t,std::string> milestones,uint64_t untilFrame){
static Map<uint32_t,std::string> noMilestones; static Map<uint32_t,uint64_t> noHits;
{
dsmachine_Result res = dsmachine_Result{{},{},{},Map<uint32_t,uint64_t>{}};
if ((quantum <= cast<int64_t>(0ULL))) {
quantum = cast<int64_t>(64ULL);
}
if ((!m->visited)) {
m->visited = Map<uint32_t,bool>{};
}
uint64_t lastProgress = m->Steps;
uint64_t prevSig = dsmachine_Machine_progressSig(m);
int64_t prevPages = len(m->visited);
auto tmp187 = std::make_tuple(false,false,cast<uint32_t>(0ULL));
m->stop = std::get<0>(tmp187);
m->stopped = std::get<1>(tmp187);
m->stoppedPC = std::get<2>(tmp187);
uint64_t end = cast<uint64_t>((m->Steps + budget));
{;for (;(m->Steps < end);){
dsmachine_Machine_startLine(m);
if (m->stop) {
res.Reason = std::string("stopped",7);
break;
}
if (((untilFrame != cast<uint64_t>(0ULL)) && (m->vid.frames >= untilFrame))) {
res.Reason = go_fmt_Sprintf(std::string("reached frame %d",16),m->vid.frames);
break;
}
bool hb = false;
{int64_t spent = cast<int64_t>(0ULL);for (;(spent < cast<int64_t>(4260ULL));spent += quantum){
if (((!hb) && (divi<int64_t>(cast<int64_t>((spent * cast<int64_t>(355ULL))),cast<int64_t>(4260ULL)) >= cast<int64_t>(256ULL)))) {
dsmachine_Machine_hblankNow(m);
hb = true;
}
dsmachine_Machine_deliver(m,m->ARM9);
dsmachine_Machine_deliver(m,m->ARM7);
dsmachine_Machine_runQuantum(m,m->ARM9,quantum,milestones,res.ARM9Milest);
dsmachine_Machine_runQuantum(m,m->ARM7,divi<int64_t>(quantum,cast<int64_t>(2ULL)),noMilestones,noHits);
m->Steps += cast<uint64_t>(quantum);
if (((m->ARM9->cpu->Halted || m->ARM7->cpu->Halted) || m->stop)) {
break;
}
}
}if (m->stop) {
res.Reason = std::string("stopped",7);
if (m->stopped) {
res.Reason = go_fmt_Sprintf(std::string("breakpoint at 0x%08X",20),m->stoppedPC);
}
break;
}
if ((!hb)) {
dsmachine_Machine_hblankNow(m);
}
{auto&& tmp188 = dsmachine_Machine_cores(m);
for(int64_t tmp189=0;tmp189<len(tmp188);++tmp189){
auto c=tmp188[tmp189];dsmachine_core_tickTimers(c,cast<int64_t>(2130ULL));
}}
if (m->ARM9->cpu->Halted) {
res.Reason = (std::string("ARM9 halted: ",13) + m->ARM9->cpu->HaltReason);
break;
}
if (m->ARM7->cpu->Halted) {
res.Reason = (std::string("ARM7 halted: ",13) + m->ARM7->cpu->HaltReason);
break;
}
uint64_t sig = dsmachine_Machine_progressSig(m);
if (((sig != prevSig) || (len(m->visited) != prevPages))) {
auto tmp190 = std::make_tuple(sig,len(m->visited));
prevSig = std::get<0>(tmp190);
prevPages = std::get<1>(tmp190);
lastProgress = m->Steps;
}
else if ((cast<uint64_t>((m->Steps - lastProgress)) > cast<uint64_t>(24000000ULL))) {
res.Reason = go_fmt_Sprintf(std::string("settled \342\200\224 ARM9 spinning at 0x%08X (%s), ARM7 at 0x%08X (%s); no new code or IPC traffic",89),m->ARM9->cpu->R[cast<int64_t>(15ULL)],dsmachine_parkState(m->ARM9),m->ARM7->cpu->R[cast<int64_t>(15ULL)],dsmachine_parkState(m->ARM7));
break;
}
}
}if ((res.Reason == std::string("",0))) {
res.Reason = go_fmt_Sprintf(std::string("step budget (%d) reached",24),budget);
}
auto tmp191 = std::make_tuple(m->Steps,m->vid.frames);
res.Steps = std::get<0>(tmp191);
res.Frames = std::get<1>(tmp191);
return res;
}
}
// tools/platform/nds/dsmachine/run.go:148:1
void dsmachine_Machine_runQuantum(dsmachine_Machine* m,dsmachine_core* c,int64_t n,const Map<uint32_t,std::string>& milestones,const Map<uint32_t,uint64_t>& hit){
uint32_t lastPage = UINT32_MAX;
{
if ((c->waiting || c->wfi)) {
return ;
}
if ((c->sleep > cast<int64_t>(0ULL))) {
c->sleep -= n;
return ;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
if (((((c->waiting || c->wfi) || (c->sleep > cast<int64_t>(0ULL))) || c->cpu->Halted) || m->stop)) {
return ;
}
if ((c->cpu->R[cast<int64_t>(15ULL)] == cast<uint32_t>(4294905856ULL))) {
dsmachine_core_biosIRQExit(c);
continue;
}
uint32_t pc = c->cpu->R[cast<int64_t>(15ULL)];
if (bool(m->OnStep)) {
m->OnStep(c->arm9,pc);
}
if (c->arm9) {
if (m->bps.size() && get(m->bps,pc)) {
auto tmp192 = std::make_tuple(true,true,pc);
m->stop = std::get<0>(tmp192);
m->stopped = std::get<1>(tmp192);
m->stoppedPC = std::get<2>(tmp192);
return ;
}
if ((pc >> 8) != lastPage) { lastPage=pc >> 8; m->visited[lastPage]=true; }
if (milestones.size()) {
{
auto tmp193 = lookup(milestones,pc);
bool ok = std::get<1>(tmp193);
if (ok) {
{
auto tmp194 = lookup(hit,pc);
bool seen = std::get<1>(tmp194);
if ((!seen)) {
(*hit.p)[pc] = cast<uint64_t>((m->Steps + cast<uint64_t>(i)));
}
}
}
}
}
}
arm_CPU_Step(c->cpu);
}
}}
}
// tools/platform/nds/dsmachine/run.go:217:1
void dsmachine_Machine_deliver(dsmachine_Machine* m,dsmachine_core* c){
{
uint32_t pending = cast<uint32_t>((c->ie & c->if_));
if (((pending == cast<uint32_t>(0ULL)) || (!c->ime))) {
return ;
}
if (c->wfi) {
c->wfi = false;
if (c->cpu->IRQDisable) {
return ;
}
}
if (c->waiting) {
if ((((!c->waitAny) && (c->waitMask != cast<uint32_t>(0ULL))) && (cast<uint32_t>((pending & c->waitMask)) == cast<uint32_t>(0ULL)))) {
return ;
}
}
else if (c->cpu->IRQDisable) {
return ;
}
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
uint32_t flag = dsmachine_bus_r32(b,cast<uint32_t>((c->handlerBase - cast<uint32_t>(8ULL))));
dsmachine_bus_w32(b,cast<uint32_t>((c->handlerBase - cast<uint32_t>(8ULL))),cast<uint32_t>((flag | pending)));
uint32_t handler = dsmachine_bus_r32(b,cast<uint32_t>((c->handlerBase - cast<uint32_t>(4ULL))));
if ((handler == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t ret = c->cpu->R[cast<int64_t>(15ULL)];
c->waiting = false;
if (bool(m->OnIRQ)) {
m->OnIRQ(c->arm9,pending,handler,ret);
}
arm_CPU_Exception(c->cpu,cast<uint32_t>(18ULL),handler,cast<uint32_t>((ret + cast<uint32_t>(4ULL))));
uint32_t sp = cast<uint32_t>((c->cpu->R[cast<int64_t>(13ULL)] - cast<uint32_t>(24ULL)));
c->cpu->R[cast<int64_t>(13ULL)] = sp;
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(0ULL))),c->cpu->R[cast<int64_t>(0ULL)]);
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(4ULL))),c->cpu->R[cast<int64_t>(1ULL)]);
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(8ULL))),c->cpu->R[cast<int64_t>(2ULL)]);
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(12ULL))),c->cpu->R[cast<int64_t>(3ULL)]);
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(16ULL))),c->cpu->R[cast<int64_t>(12ULL)]);
dsmachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(20ULL))),c->cpu->R[cast<int64_t>(14ULL)]);
c->cpu->R[cast<int64_t>(14ULL)] = cast<uint32_t>(4294905856ULL);
}
}
// tools/platform/nds/dsmachine/run.go:283:1
void dsmachine_core_biosIRQExit(dsmachine_core* c){
{
dsmachine_bus bStorage = dsmachine_bus{c};
dsmachine_bus* b = &bStorage;
uint32_t sp = c->cpu->R[cast<int64_t>(13ULL)];
c->cpu->R[cast<int64_t>(0ULL)] = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(0ULL))));
c->cpu->R[cast<int64_t>(1ULL)] = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(4ULL))));
c->cpu->R[cast<int64_t>(2ULL)] = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(8ULL))));
c->cpu->R[cast<int64_t>(3ULL)] = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(12ULL))));
c->cpu->R[cast<int64_t>(12ULL)] = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(16ULL))));
uint32_t lr = dsmachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(20ULL))));
c->cpu->R[cast<int64_t>(13ULL)] = cast<uint32_t>((sp + cast<uint32_t>(24ULL)));
uint32_t spsr = arm_CPU_SPSR(c->cpu);
arm_CPU_SetCPSR(c->cpu,spsr);
c->cpu->R[cast<int64_t>(15ULL)] = cast<uint32_t>((lr - cast<uint32_t>(4ULL)));
}
}
// tools/platform/nds/dsmachine/run.go:310:1
uint64_t dsmachine_Machine_progressSig(dsmachine_Machine* m){
{
return cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(m->ipc.sync9),cast<int64_t>(1ULL)) ^ shl<uint64_t>(cast<uint64_t>(m->ipc.sync7),cast<int64_t>(5ULL)))) ^ shl<uint64_t>(cast<uint64_t>(len(m->ipc.to7)),cast<int64_t>(8ULL)))) ^ shl<uint64_t>(cast<uint64_t>(len(m->ipc.to9)),cast<int64_t>(12ULL)))) ^ shl<uint64_t>(cast<uint64_t>(m->ARM9->if_),cast<int64_t>(16ULL)))) ^ shl<uint64_t>(cast<uint64_t>(m->ARM7->if_),cast<int64_t>(32ULL))));
}
}
// tools/platform/nds/dsmachine/run.go:316:1
std::string dsmachine_parkState(dsmachine_core* c){
{
if (c->waiting) {
if (c->waitAny) {
return std::string("halted for IRQ",14);
}
return go_fmt_Sprintf(std::string("IntrWait 0x%X",13),c->waitMask);
}
return std::string("running",7);
}
}
// tools/platform/nds/dsmachine/run.go:331:1
int64_t dsmachine_Machine_runInstrs(dsmachine_Machine* m,int64_t n){
{
auto tmp195 = std::make_tuple(false,false,cast<uint32_t>(0ULL));
m->stop = std::get<0>(tmp195);
m->stopped = std::get<1>(tmp195);
m->stoppedPC = std::get<2>(tmp195);
if ((!m->visited)) {
m->visited = Map<uint32_t,bool>{};
}
int64_t ran = cast<int64_t>(0ULL);
std::function<void(bool,uint32_t)> prev = m->OnStep;
m->OnStep = [&](bool arm9,uint32_t pc)->void{
if (arm9) {
ran++;
if ((ran > n)) {
m->stop = true;
}
}
if (bool(prev)) {
prev(arm9,pc);
}
}
;
auto tmp196=defer([&](){[&]()->void{
m->OnStep = prev;
}
();});
dsmachine_Machine_run(m,cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(n) * cast<uint64_t>(64ULL))) + cast<uint64_t>(4000000ULL))),cast<int64_t>(64ULL),{},cast<uint64_t>(0ULL));
if ((ran > n)) {
ran = n;
}
return ran;
}
}
// tools/platform/nds/dsmachine/screen.go:27:1
std::tuple<image_RGBA*,image_RGBA*> dsmachine_Machine_Screens(dsmachine_Machine* m){
image_RGBA* top{};
image_RGBA* bottom{};
{
auto tmp197 = dsmachine_gpu2d_screens(m->gpu2d);
Slice<uint32_t> t = std::get<0>(tmp197);
Slice<uint32_t> b = std::get<1>(tmp197);
return {dsmachine_toImage(t),dsmachine_toImage(b)};
}
}
// tools/platform/nds/dsmachine/screen.go:32:1
image_RGBA* dsmachine_toImage(Slice<uint32_t> px){
{
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(256ULL),cast<int64_t>(192ULL)));
if ((!px)) {
return img;
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(192ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(256ULL));x++){
uint32_t v = px[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(256ULL))) + x))];
image_RGBA_Set(img,x,y,color_RGBA{cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(255ULL)});
}
}}
}return img;
}
}
// tools/platform/nds/dsmachine/screen.go:111:1
std::tuple<int64_t,int64_t> dsmachine_Machine_EngineStats(dsmachine_Machine* m){
int64_t a{};
int64_t b{};
{
auto count = [&](Slice<uint32_t> px)->int64_t{
int64_t n = cast<int64_t>(0ULL);
{auto&& tmp198 = px;
for(int64_t tmp199=0;tmp199<len(tmp198);++tmp199){
auto v=tmp198[tmp199];if ((cast<uint32_t>((v & cast<uint32_t>(4294967040ULL))) != cast<uint32_t>(0ULL))) {
n++;
}
}}
return n;
}
;
return {count(m->gpu2d->a.out),count(m->gpu2d->b.out)};
}
}
// tools/platform/nds/dsmachine/sound.go:31:1
void dsmachine_Machine_soundKeyed(dsmachine_Machine* m){
{
dsmachine_Machine_note(m,std::string("ARM7: sound channels are being keyed on; the register file is modelled but the mixer is not (sound.go)",102));
}
}
// tools/platform/nds/dsmachine/spi.go:74:1
dsmachine_spibus* dsmachine_newSPI(){
{
dsmachine_spibus* s = arenaNew(dsmachine_spibus{Slice<uint8_t>::make(cast<int64_t>(262144ULL)),{},{},{},{},{},{},{},{},{},{}});
dsmachine_spibus_buildFirmware(s);
return s;
}
}
// tools/platform/nds/dsmachine/spi.go:82:1
void dsmachine_spibus_buildFirmware(dsmachine_spibus* s){
{
Slice<uint8_t> f = s->firmware;
binary_littleEndian le = go_binary_LittleEndian;
le_PutUint16(sub(f,cast<int64_t>(32ULL),len(f)),cast<uint16_t>(32704ULL));
gcopy(sub(f,cast<int64_t>(8ULL),len(f)),cast<Slice<uint8_t>>(std::string("MACP",4)));
f[cast<int64_t>(29ULL)] = cast<uint8_t>(255ULL);
std::array<uint8_t,256> us={};
us[cast<int64_t>(0ULL)] = cast<uint8_t>(5ULL);
us[cast<int64_t>(2ULL)] = cast<uint8_t>(4ULL);
us[cast<int64_t>(3ULL)] = cast<uint8_t>(7ULL);
us[cast<int64_t>(4ULL)] = cast<uint8_t>(4ULL);
Slice<uint16_t> name = go_utf16_Encode(cast<Slice<int32_t>>(std::string("RETRO",5)));
{auto&& tmp200 = name;
for(int64_t tmp201=0;tmp201<len(tmp200);++tmp201){
auto i=tmp201;auto r=tmp200[tmp201];le_PutUint16(sub(us,cast<int64_t>((cast<int64_t>(6ULL) + cast<int64_t>((i * cast<int64_t>(2ULL))))),len(us)),r);
}}
le_PutUint16(sub(us,cast<int64_t>(26ULL),len(us)),cast<uint16_t>(len(name)));
le_PutUint16(sub(us,cast<int64_t>(80ULL),len(us)),cast<uint16_t>(0ULL));
le_PutUint16(sub(us,cast<int64_t>(88ULL),len(us)),cast<uint16_t>(735ULL));
le_PutUint16(sub(us,cast<int64_t>(90ULL),len(us)),cast<uint16_t>(812ULL));
us[cast<int64_t>(92ULL)] = cast<uint8_t>(32ULL);
us[cast<int64_t>(93ULL)] = cast<uint8_t>(32ULL);
le_PutUint16(sub(us,cast<int64_t>(94ULL),len(us)),cast<uint16_t>(3387ULL));
le_PutUint16(sub(us,cast<int64_t>(96ULL),len(us)),cast<uint16_t>(3303ULL));
us[cast<int64_t>(98ULL)] = cast<uint8_t>(224ULL);
us[cast<int64_t>(99ULL)] = cast<uint8_t>(158ULL);
le_PutUint16(sub(us,cast<int64_t>(100ULL),len(us)),cast<uint16_t>(32841ULL));
{auto&& tmp202 = Slice<int64_t>{cast<int64_t>(261632ULL),cast<int64_t>(261888ULL)};
for(int64_t tmp203=0;tmp203<len(tmp202);++tmp203){
auto i=tmp203;auto off=tmp202[tmp203];std::array<uint8_t,256> blk = us;
le_PutUint16(sub(blk,cast<int64_t>(112ULL),len(blk)),cast<uint16_t>(i));
le_PutUint16(sub(blk,cast<int64_t>(114ULL),len(blk)),nds_CRC16(sub(blk,cast<int64_t>(0ULL),cast<int64_t>(112ULL))));
gcopy(sub(f,off,len(f)),sub(blk,0,len(blk)));
}}
}
}
// tools/platform/nds/dsmachine/spi.go:130:1
void dsmachine_Machine_SetTouch(dsmachine_Machine* m,int64_t x,int64_t y,bool down){
{
auto tmp204 = std::make_tuple(x,y,down);
m->spi->touchX = std::get<0>(tmp204);
m->spi->touchY = std::get<1>(tmp204);
m->spi->touchDown = std::get<2>(tmp204);
}
}
// tools/platform/nds/dsmachine/spi.go:135:1
std::tuple<int64_t,int64_t,bool> dsmachine_Machine_Touch(dsmachine_Machine* m){
int64_t x{};
int64_t y{};
bool down{};
{
return {m->spi->touchX,m->spi->touchY,m->spi->touchDown};
}
}
// tools/platform/nds/dsmachine/spi.go:140:1
void dsmachine_core_spiTransfer(dsmachine_core* c,uint8_t v){
{
dsmachine_spibus* s = c->m->spi;
uint32_t cnt = get(c->io,cast<uint32_t>(67109312ULL));
if ((cast<uint32_t>((cnt & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
int64_t dev = cast<int64_t>((cast<int64_t>(shr<uint32_t>(cnt,cast<int64_t>(8ULL))) & cast<int64_t>(3ULL)));
if ((dev != s->dev)) {
auto tmp205 = std::make_tuple(dev,cast<int64_t>(0ULL));
s->dev = std::get<0>(tmp205);
s->phase = std::get<1>(tmp205);
}
if ((s->phase == cast<int64_t>(0ULL))) {
s->cmd = v;
}
{
switch(dev){
case cast<int64_t>(1ULL):{
s->out = dsmachine_spibus_firmwareByte(s,v);
break;}
case cast<int64_t>(2ULL):{
s->out = dsmachine_spibus_touchByte(s,v);
break;}
default:{
s->out = cast<uint8_t>(0ULL);
break;}
}}
s->phase++;
if ((cast<uint32_t>((cnt & cast<uint32_t>(2048ULL))) == cast<uint32_t>(0ULL))) {
s->phase = cast<int64_t>(0ULL);
s->dev = cast<int64_t>(-1ULL);
}
if ((cast<uint32_t>((cnt & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(c,cast<uint32_t>(8388608ULL));
}
}
}
// tools/platform/nds/dsmachine/spi.go:178:1
uint8_t dsmachine_spibus_firmwareByte(dsmachine_spibus* s,uint8_t v){
{
{
switch(s->cmd){
case cast<uint8_t>(3ULL):{
{
if ((s->phase == cast<int64_t>(0ULL))){
s->addr = cast<uint32_t>(0ULL);
return cast<uint8_t>(0ULL);
}
else if ((s->phase <= cast<int64_t>(3ULL))){
s->addr = cast<uint32_t>((shl<uint32_t>(s->addr,cast<int64_t>(8ULL)) | cast<uint32_t>(v)));
return cast<uint8_t>(0ULL);
}
else {
uint8_t b = cast<uint8_t>(0ULL);
if ((cast<int64_t>(s->addr) < len(s->firmware))) {
b = s->firmware[s->addr];
}
s->addr++;
return b;
}
}
tmp206:;
break;}
case cast<uint8_t>(5ULL):{
return cast<uint8_t>(0ULL);
break;}
case cast<uint8_t>(159ULL):{
Slice<uint8_t> ids = Slice<uint8_t>{cast<uint8_t>(32ULL),cast<uint8_t>(64ULL),cast<uint8_t>(18ULL)};
if (((s->phase >= cast<int64_t>(1ULL)) && (s->phase <= cast<int64_t>(3ULL)))) {
return ids[cast<int64_t>((s->phase - cast<int64_t>(1ULL)))];
}
return cast<uint8_t>(0ULL);
break;}
}}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/nds/dsmachine/spi.go:221:1
uint8_t dsmachine_spibus_touchByte(dsmachine_spibus* s,uint8_t v){
{
if ((cast<uint8_t>((v & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
s->chanSel = cast<int64_t>((cast<int64_t>(shr<uint8_t>(v,cast<int64_t>(4ULL))) & cast<int64_t>(7ULL)));
s->resultIdx = cast<int64_t>(0ULL);
return cast<uint8_t>(0ULL);
}
uint16_t val = dsmachine_spibus_touchSample(s,s->chanSel);
s->resultIdx++;
if ((s->resultIdx == cast<int64_t>(1ULL))) {
return cast<uint8_t>(shr<uint16_t>(val,cast<int64_t>(5ULL)));
}
return cast<uint8_t>((cast<uint8_t>(shl<uint16_t>(val,cast<int64_t>(3ULL))) & cast<uint8_t>(248ULL)));
}
}
// tools/platform/nds/dsmachine/spi.go:239:1
uint16_t dsmachine_spibus_touchSample(dsmachine_spibus* s,int64_t ch){
{
if ((!s->touchDown)) {
return cast<uint16_t>(0ULL);
}
{
switch(ch){
case cast<int64_t>(1ULL):{
return cast<uint16_t>(cast<int64_t>((cast<int64_t>(812ULL) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((s->touchY - cast<int64_t>(32ULL)))) * cast<int64_t>(2491ULL))),cast<int64_t>(126ULL)))));
break;}
case cast<int64_t>(5ULL):{
return cast<uint16_t>(cast<int64_t>((cast<int64_t>(735ULL) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((s->touchX - cast<int64_t>(32ULL)))) * cast<int64_t>(2652ULL))),cast<int64_t>(192ULL)))));
break;}
case cast<int64_t>(3ULL):case cast<int64_t>(4ULL):{
return cast<uint16_t>(512ULL);
break;}
}}
return cast<uint16_t>(0ULL);
}
}
// tools/platform/nds/dsmachine/timer.go:24:1
bool dsmachine_timer_enabled(dsmachine_timer* t){
{
return (cast<uint16_t>((t->ctrl & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/timer.go:25:1
bool dsmachine_timer_cascade(dsmachine_timer* t){
{
return (cast<uint16_t>((t->ctrl & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/timer.go:26:1
bool dsmachine_timer_irqOn(dsmachine_timer* t){
{
return (cast<uint16_t>((t->ctrl & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL));
}
}
// tools/platform/nds/dsmachine/timer.go:30:1
void dsmachine_core_tickTimers(dsmachine_core* c,int64_t cycles){
{
int64_t carry={};
{auto&& tmp207 = c->timers;
for(int64_t tmp208=0;tmp208<len(tmp207);++tmp208){
auto n=tmp208;dsmachine_timer* t = (&c->timers[n]);
if ((!dsmachine_timer_enabled(t))) {
carry = cast<int64_t>(0ULL);
continue;
}
int64_t ticks={};
if (((n > cast<int64_t>(0ULL)) && dsmachine_timer_cascade(t))) {
ticks = carry;
}
else {
int64_t p = dsmachine_timerPrescale[cast<uint16_t>((t->ctrl & cast<uint16_t>(3ULL)))];
t->frac += cycles;
ticks = divi<int64_t>(t->frac,p);
t->frac -= cast<int64_t>((ticks * p));
}
carry = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < ticks);i++){
if ((t->counter == cast<uint16_t>(65535ULL))) {
t->counter = t->reload;
carry++;
if (dsmachine_timer_irqOn(t)) {
dsmachine_core_raise(c,shl<uint32_t>(cast<uint32_t>(8ULL),cast<uint64_t>(n)));
}
}
else {
t->counter++;
}
}
}}}
}
}
// tools/platform/nds/dsmachine/timer.go:68:1
void dsmachine_core_writeTimerCtrl(dsmachine_core* c,int64_t n,uint16_t v){
{
dsmachine_timer* t = (&c->timers[n]);
bool wasOn = dsmachine_timer_enabled(t);
t->ctrl = v;
if (((!wasOn) && dsmachine_timer_enabled(t))) {
t->counter = t->reload;
t->frac = cast<int64_t>(0ULL);
}
}
}
// tools/platform/nds/dsmachine/video.go:73:1
int64_t dsmachine_core_vcountMatch(dsmachine_core* c){
{
uint32_t d = get(c->io,cast<uint32_t>(67108868ULL));
return cast<int64_t>(cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(d,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))) | shl<uint32_t>((cast<uint32_t>((d & cast<uint32_t>(128ULL)))),cast<int64_t>(1ULL)))));
}
}
// tools/platform/nds/dsmachine/video.go:80:1
uint32_t dsmachine_core_dispstat(dsmachine_core* c){
{
uint32_t v = cast<uint32_t>((get(c->io,cast<uint32_t>(67108868ULL)) & cast<uint32_t>(65464ULL)));
int64_t ln = c->m->vid.line;
if (((ln >= cast<int64_t>(192ULL)) && (ln < cast<int64_t>(262ULL)))) {
v |= cast<uint32_t>(1ULL);
}
if (c->m->vid.hblank) {
v |= cast<uint32_t>(2ULL);
}
if ((ln == dsmachine_core_vcountMatch(c))) {
v |= cast<uint32_t>(4ULL);
}
return v;
}
}
// tools/platform/nds/dsmachine/video.go:102:1
void dsmachine_Machine_startLine(dsmachine_Machine* m){
{
dsmachine_video* v = (&m->vid);
v->line++;
if ((v->line >= cast<int64_t>(263ULL))) {
v->line = cast<int64_t>(0ULL);
v->frames++;
}
v->hblank = false;
{auto&& tmp209 = dsmachine_Machine_cores(m);
for(int64_t tmp210=0;tmp210<len(tmp209);++tmp210){
auto c=tmp209[tmp210];if (((v->line == dsmachine_core_vcountMatch(c)) && (cast<uint32_t>((get(c->io,cast<uint32_t>(67108868ULL)) & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL)))) {
dsmachine_core_raise(c,cast<uint32_t>(4ULL));
}
}}
{
switch(v->line){
case cast<int64_t>(192ULL):{
{auto&& tmp211 = dsmachine_Machine_cores(m);
for(int64_t tmp212=0;tmp212<len(tmp211);++tmp212){
auto c=tmp211[tmp212];if ((cast<uint32_t>((get(c->io,cast<uint32_t>(67108868ULL)) & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(c,cast<uint32_t>(1ULL));
}
}}
dsmachine_Machine_runDMA(m,cast<int64_t>(1ULL));
dsmachine_gpu3d_vblank(m->gpu3d,m);
dsmachine_Machine_onFrame(m);
dsmachine_profiler_reset(&(m->prof),m);
break;}
case cast<int64_t>(0ULL):{
dsmachine_gpu2d_beginFrame(m->gpu2d,m);
break;}
}}
}
}
// tools/platform/nds/dsmachine/video.go:136:1
void dsmachine_Machine_hblankNow(dsmachine_Machine* m){
{
m->vid.hblank = true;
{auto&& tmp213 = dsmachine_Machine_cores(m);
for(int64_t tmp214=0;tmp214<len(tmp213);++tmp214){
auto c=tmp213[tmp214];if ((cast<uint32_t>((get(c->io,cast<uint32_t>(67108868ULL)) & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
dsmachine_core_raise(c,cast<uint32_t>(2ULL));
}
}}
if ((m->vid.line < cast<int64_t>(192ULL))) {
dsmachine_Machine_runDMA(m,cast<int64_t>(2ULL));
}
}
}
// tools/platform/nds/dsmachine/video.go:148:1
std::array<dsmachine_core*,2> dsmachine_Machine_cores(dsmachine_Machine* m){
{
return std::array<dsmachine_core*,2>{m->ARM9,m->ARM7};
}
}
// tools/platform/nds/dsmachine/video.go:152:1
void dsmachine_core_raise(dsmachine_core* c,uint32_t src){
{
c->if_ |= src;
}
}
// tools/platform/nds/dsmachine/vram.go:76:1
dsmachine_vram* dsmachine_newVRAM(){
{
dsmachine_vram* v = arenaNew(dsmachine_vram{});
{auto&& tmp215 = v->bank;
for(int64_t tmp216=0;tmp216<len(tmp215);++tmp216){
auto i=tmp216;v->bank[i] = Slice<uint8_t>::make(dsmachine_bankSizes[i]);
}}
{int64_t s = cast<int64_t>(0ULL);for (;(s < cast<int64_t>(11ULL));s++){
v->pages[s] = Slice<Slice<Slice<uint8_t>>>::make(divi<int64_t>(dsmachine_spaceSize[s],cast<int64_t>(8192ULL)));
}
}return v;
}
}
// tools/platform/nds/dsmachine/vram.go:88:1
void dsmachine_vram_setCNT(dsmachine_vram* v,int64_t bank,uint8_t val){
{
if ((v->cnt[bank] == val)) {
return ;
}
v->cnt[bank] = val;
dsmachine_vram_remap(v);
}
}
// tools/platform/nds/dsmachine/vram.go:97:1
void dsmachine_vram_remap(dsmachine_vram* v){
{
{int64_t s = cast<int64_t>(0ULL);for (;(s < cast<int64_t>(11ULL));s++){
{auto&& tmp217 = v->pages[s];
for(int64_t tmp218=0;tmp218<len(tmp217);++tmp218){
auto p=tmp218;v->pages[s][p] = sub(v->pages[s][p],0,cast<int64_t>(0ULL));
}}
}
}{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(9ULL));b++){
uint8_t cnt = v->cnt[b];
if ((cast<uint8_t>((cnt & cast<uint8_t>(128ULL))) == cast<uint8_t>(0ULL))) {
continue;
}
int64_t mst = cast<int64_t>(cast<uint8_t>((cnt & cast<uint8_t>(7ULL))));
int64_t ofs = cast<int64_t>((cast<int64_t>(shr<uint8_t>(cnt,cast<int64_t>(3ULL))) & cast<int64_t>(3ULL)));
auto tmp219 = dsmachine_bankTarget(b,mst,ofs);
int64_t space = std::get<0>(tmp219);
int64_t off = std::get<1>(tmp219);
bool ok = std::get<2>(tmp219);
if ((!ok)) {
continue;
}
dsmachine_vram_attach(v,space,off,v->bank[b]);
}
}}
}
// tools/platform/nds/dsmachine/vram.go:119:1
void dsmachine_vram_attach(dsmachine_vram* v,int64_t space,int64_t off,Slice<uint8_t> data){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(data));i += cast<int64_t>(8192ULL)){
int64_t p = divi<int64_t>((cast<int64_t>((off + i))),cast<int64_t>(8192ULL));
if (((p < cast<int64_t>(0ULL)) || (p >= len(v->pages[space])))) {
continue;
}
int64_t end = cast<int64_t>((i + cast<int64_t>(8192ULL)));
if ((end > len(data))) {
end = len(data);
}
v->pages[space][p] = append(v->pages[space][p],sub(data,i,end));
}
}}
}
// tools/platform/nds/dsmachine/vram.go:138:1
std::tuple<int64_t,int64_t,bool> dsmachine_bankTarget(int64_t b,int64_t mst,int64_t ofs){
int64_t space{};
int64_t off{};
bool ok{};
{
{
switch(b){
case cast<int64_t>(0ULL):case cast<int64_t>(1ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(0ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(2ULL),cast<int64_t>((cast<int64_t>(131072ULL) * (cast<int64_t>((ofs & cast<int64_t>(1ULL)))))),true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(4ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
}}
break;}
case cast<int64_t>(2ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(0ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(10ULL),cast<int64_t>((cast<int64_t>(131072ULL) * (cast<int64_t>((ofs & cast<int64_t>(1ULL)))))),true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(4ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
case cast<int64_t>(4ULL):{
return {cast<int64_t>(1ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
case cast<int64_t>(3ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(0ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(10ULL),cast<int64_t>((cast<int64_t>(131072ULL) * (cast<int64_t>((ofs & cast<int64_t>(1ULL)))))),true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(4ULL),cast<int64_t>((cast<int64_t>(131072ULL) * ofs)),true};
break;}
case cast<int64_t>(4ULL):{
return {cast<int64_t>(3ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
case cast<int64_t>(4ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(2ULL),cast<int64_t>(0ULL),true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(5ULL),cast<int64_t>(0ULL),true};
break;}
case cast<int64_t>(4ULL):{
return {cast<int64_t>(6ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
case cast<int64_t>(5ULL):case cast<int64_t>(6ULL):{
int64_t step = cast<int64_t>((cast<int64_t>((cast<int64_t>(16384ULL) * (cast<int64_t>((ofs & cast<int64_t>(1ULL)))))) + cast<int64_t>((cast<int64_t>(65536ULL) * (shr<int64_t>(ofs,cast<int64_t>(1ULL)))))));
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(0ULL),step,true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(2ULL),step,true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(5ULL),step,true};
break;}
case cast<int64_t>(4ULL):{
return {cast<int64_t>(6ULL),cast<int64_t>((cast<int64_t>(16384ULL) * (cast<int64_t>((ofs & cast<int64_t>(1ULL)))))),true};
break;}
case cast<int64_t>(5ULL):{
return {cast<int64_t>(8ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
case cast<int64_t>(7ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(1ULL),cast<int64_t>(0ULL),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(7ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
case cast<int64_t>(8ULL):{
{
switch(mst){
case cast<int64_t>(1ULL):{
return {cast<int64_t>(1ULL),cast<int64_t>(32768ULL),true};
break;}
case cast<int64_t>(2ULL):{
return {cast<int64_t>(3ULL),cast<int64_t>(0ULL),true};
break;}
case cast<int64_t>(3ULL):{
return {cast<int64_t>(9ULL),cast<int64_t>(0ULL),true};
break;}
}}
break;}
}}
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
}
// tools/platform/nds/dsmachine/vram.go:221:1
uint8_t dsmachine_vram_read8(dsmachine_vram* v,int64_t space,uint32_t off){
{
uint32_t p = divi<uint32_t>(off,cast<uint32_t>(8192ULL));
if ((cast<int64_t>(p) >= len(v->pages[space]))) {
return cast<uint8_t>(0ULL);
}
const auto& refs = v->pages[space][p];
if ((len(refs) == cast<int64_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
uint32_t i = modi<uint32_t>(off,cast<uint32_t>(8192ULL));
if ((cast<int64_t>(i) >= len(refs[cast<int64_t>(0ULL)]))) {
return cast<uint8_t>(0ULL);
}
return refs[cast<int64_t>(0ULL)][i];
}
}
// tools/platform/nds/dsmachine/vram.go:237:1
uint16_t dsmachine_vram_read16(dsmachine_vram* v,int64_t space,uint32_t off){
{
return cast<uint16_t>((cast<uint16_t>(dsmachine_vram_read8(v,space,off)) | shl<uint16_t>(cast<uint16_t>(dsmachine_vram_read8(v,space,cast<uint32_t>((off + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/platform/nds/dsmachine/vram.go:242:1
void dsmachine_vram_write8(dsmachine_vram* v,int64_t space,uint32_t off,uint8_t b){
{
uint32_t p = divi<uint32_t>(off,cast<uint32_t>(8192ULL));
if ((cast<int64_t>(p) >= len(v->pages[space]))) {
return ;
}
uint32_t i = modi<uint32_t>(off,cast<uint32_t>(8192ULL));
{auto&& tmp220 = v->pages[space][p];
for(int64_t tmp221=0;tmp221<len(tmp220);++tmp221){
auto r=tmp220[tmp221];if ((cast<int64_t>(i) < len(r))) {
r[i] = b;
}
}}
}
}
// tools/platform/nds/dsmachine/vram.go:259:1
std::tuple<Slice<uint8_t>,uint32_t,bool> dsmachine_vram_lcdcSlot(dsmachine_vram* v,uint32_t a){
{
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(9ULL));b++){
uint32_t base = dsmachine_lcdcBase[b];
if (((a >= base) && (a < cast<uint32_t>((base + cast<uint32_t>(dsmachine_bankSizes[b])))))) {
return {v->bank[b],cast<uint32_t>((a - base)),true};
}
}
}return {{},cast<uint32_t>(0ULL),false};
}
}
// tools/platform/nds/dsmachine/vram.go:273:1
std::tuple<int64_t,uint32_t,bool> dsmachine_vram_cpuAccess(dsmachine_vram* v,bool arm9,uint32_t a){
int64_t space{};
uint32_t off{};
bool ok{};
{
if ((!arm9)) {
if (((a >= cast<uint32_t>(100663296ULL)) && (a < cast<uint32_t>(117440512ULL)))) {
return {cast<int64_t>(10ULL),cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(100663296ULL)))) & cast<uint32_t>(262143ULL))),true};
}
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
{
if (((a >= cast<uint32_t>(109051904ULL)) && (a < cast<uint32_t>(109723648ULL)))){
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
else if (((a >= cast<uint32_t>(100663296ULL)) && (a < cast<uint32_t>(102760448ULL)))){
return {cast<int64_t>(0ULL),cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(100663296ULL)))) & cast<uint32_t>(524287ULL))),true};
}
else if (((a >= cast<uint32_t>(102760448ULL)) && (a < cast<uint32_t>(104857600ULL)))){
return {cast<int64_t>(1ULL),cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(102760448ULL)))) & cast<uint32_t>(131071ULL))),true};
}
else if (((a >= cast<uint32_t>(104857600ULL)) && (a < cast<uint32_t>(106954752ULL)))){
return {cast<int64_t>(2ULL),cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(104857600ULL)))) & cast<uint32_t>(262143ULL))),true};
}
else if (((a >= cast<uint32_t>(106954752ULL)) && (a < cast<uint32_t>(109051904ULL)))){
return {cast<int64_t>(3ULL),cast<uint32_t>(((cast<uint32_t>((a - cast<uint32_t>(106954752ULL)))) & cast<uint32_t>(131071ULL))),true};
}
}
tmp222:;
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
}

#include "fast.h"
