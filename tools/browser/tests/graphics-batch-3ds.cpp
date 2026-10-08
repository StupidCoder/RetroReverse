#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>
int main(int argc,char**argv){
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);machine->SingleThreaded=true;
 auto region=n3ds_Machine_mapRegion(machine,"batch",0x14000000,Slice<uint8_t>::make(1024*1024));
 n3ds_Machine_mapRegion(machine,"alias",0x15000000,region->data);
 for(size_t i=0;i<size_t(region->data.n);i++)region->data[i]=uint8_t(i*17+i/256);
 auto g=machine->gpu;g->Regs[0x116]=3;g->Regs[0x115]=1;g->Regs[0x100]=0;g->Regs[0x102]=3;
 g->Regs[0x82]=8|(8<<16);g->Regs[0x85]=(region->base+0x80000)>>3;
 n3ds_fbState fb{};fb.width=fb.height=64;fb.colorAddr=region->base+0x20000;fb.depthAddr=region->base+0x30000;fb.depthScale=1;fb.depthZBuffer=true;
 n3ds_rasterTri tri{};tri.v0.x=.5f;tri.v0.y=.5f;tri.v1.x=63.5f;tri.v1.y=.5f;tri.v2.x=.5f;tri.v2.y=63.5f;tri.maxX=tri.maxY=64;tri.area=63*63;
 for(auto*v:{&tri.v0,&tri.v1,&tri.v2}){v->iw=1;v->z=.5f;v->col={.8f,.2f,.4f,.7f};v->quat={0,0,0,1};v->view={0,0,1};}
 const auto initial=std::vector<uint8_t>(region->data.begin(),region->data.end());
 auto reset=[&](){std::copy(initial.begin(),initial.end(),region->data.begin());g->PixelsDrawn=g->DepthKilled=0;};
 auto draw=[&](uint32_t i){
  // Retain older decoded images after the live cache changes, so a deferred
  // Reference replay cannot accidentally sample the newest texture version.
  for(int j=0;j<256;j++)region->data[0x80000+j]=uint8_t(i*13+j*7);
  n3ds_GPU_invalidateTextures(g,region->base+0x80000,256);
  assert(std::get<1>(n3ds_GPU_texture(g,region->base+0x80000,0,8,8)));
  fb.depthTest=i<24;fb.depthWr=i%3!=0;fb.depthFunc=i%8;fb.colorMask=(i*7)%16;
  g->Regs[0x105]=i<24&&i%4>=2?0xff800f01|((i%8)<<4):0;g->Regs[0x106]=(i%8)|(((i+2)%8)<<4)|(((i+4)%8)<<8);
  g->Regs[0x102]=i%16;
  n3ds_tevState tv{};tv.texEnable=1;tv.alphaTest=true;tv.alphaFunc=(i/2)%8;tv.alphaRef=127;
  for(auto&s:tv.stages)s.colr[0]=s.alph[0]={15,0};
  tv.stages[0].colr[0]={uint8_t(i<24&&i%2?1:3),0};
  n3ds_lightState ls{};ls.enabled=i<24&&i%2;ls.count=1;ls.noD1=ls.noFR=ls.noRR=ls.noRG=ls.noRB=true;
  ls.bumpMode=1;
  ls.lights[0].directional=true;ls.lights[0].pos={0,0,1};ls.lights[0].diffuse={.25f,.5f,.75f};ls.lights[0].specular0={.8f,.7f,.6f};
  for(int j=0;j<256;j++){g->LUT[0][j]=float(i+1)/40;g->LUTDiff[0][j]=-.1f;}
  tri.v0.col[3]=tri.v1.col[3]=tri.v2.col[3]=float(i%5)/4;
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});
 };
 std::vector<uint32_t> observed;
 rr_graphics_begin();for(uint32_t i=0;i<32;i++){draw(i);observed.push_back(n3ds_Machine_ReadWord(machine,fb.colorAddr));}rr_graphics_end();
 if(argc>1){std::ofstream file(argv[1],std::ios::binary);file.write((char*)rrgpu::stream.data(),rrgpu::stream.size());}
 const auto expected=std::vector<uint8_t>(region->data.begin(),region->data.end());const auto drawn=g->PixelsDrawn,killed=g->DepthKilled;
 for(int dependency=0;dependency<4;dependency++){
  reset();size_t submits=0,maxBatch=0;
  rrbatch::testSubmit=[&](const auto&entries,uint32_t*){submits++;maxBatch=std::max(maxBatch,entries.size());assert(entries.size()<=8);return false;};
  rrgpu::enabled=true;
  {rrbatch::Scope scope(g);assert(scope.acquired);
   for(uint32_t i=0;i<32;i++){
    draw(i);assert(rrbatch::pending.size()<=8);
    if(i%5==3&&dependency){
     const uint32_t alias=0x15000000+(fb.colorAddr-region->base);
     if(dependency==1)assert(rrbatch::readWord(machine,alias)==observed[i]);
     if(dependency==2){auto p=rrvertex::ordinary(machine,alias,4);assert(p&&p[0]==uint8_t(observed[i]));}
     if(dependency==3){auto[p,off]=rrbatch::readRange(machine,alias,16);assert(p&&p[off]==uint8_t(observed[i]));}
     assert(rrbatch::pending.empty());
    }
   }
  }
  assert(!rrbatch::owner&&rrbatch::pending.empty());assert(submits>0&&maxBatch>1);
  assert(g->PixelsDrawn==drawn&&g->DepthKilled==killed);assert(std::equal(expected.begin(),expected.end(),region->data.begin()));
 }
 // A dependent read of an unknown range and a software draw both flush first.
 reset();{rrbatch::Scope scope(g);draw(0);assert(!rrbatch::pending.empty());rrbatch::beforeRead(machine,0xfffffff0,32);assert(rrbatch::pending.empty());draw(1);rrgpu::rasterEnabled=false;draw(2);rrgpu::rasterEnabled=true;assert(rrbatch::pending.empty());}
 // A read starts in ordinary RAM but crosses into an overriding alias page.
 n3ds_Machine_mapRegion(machine,"outer",0x16000000,Slice<uint8_t>::make(8192));
 n3ds_Machine_mapRegion(machine,"inner alias",0x16001000,sub(region->data,0x20000,0x21000));
 {rrbatch::Scope scope(g);draw(0);assert(!rrbatch::pending.empty());rrbatch::beforeRead(machine,0x16000ffc,8);assert(rrbatch::pending.empty());}
 machine->OnRead=[](uint32_t,uint32_t,uint32_t){};{rrbatch::Scope scope(g);assert(!scope.acquired);}machine->OnRead={};
 machine->picaLimit=1;{rrbatch::Scope scope(g);assert(!scope.acquired);}machine->picaLimit=0;
 rrgpu::recording=true;{rrbatch::Scope scope(g);assert(!scope.acquired);}rrgpu::recording=false;
 rrbatch::testSubmit={};rrgpu::enabled=false;
 std::cout<<"PASS: bounded GPU batches, immutable Reference replay, register/LUT changes, physical aliases, byte/direct reads, software fallback and observation guards\n";
}
