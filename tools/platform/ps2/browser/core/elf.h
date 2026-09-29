#pragma once
// Bounded ELF32 reader replacing Go's debug/elf, with the same EE/IRX semantics.
struct RRELF {
 Slice<uint8_t> b;
 struct Section{uint32_t name,type,addr,off,size,link,entsize;};
 std::vector<Section> sections;uint32_t entry;uint16_t names;
 Slice<uint8_t> span(uint64_t at,uint64_t n)const{if(at>uint64_t(b.n)||n>uint64_t(b.n)-at)throw std::runtime_error("ELF range outside file");return sub(b,at,at+n);}
 uint32_t u32(uint64_t at)const{return le_Uint32(span(at,4));}
 uint16_t u16(uint64_t at)const{return le_Uint16(span(at,2));}
 RRELF(Slice<uint8_t>raw):b(raw){
  if(b.n<52||std::memcmp(b.p,"\177ELF",4)||b[4]!=1||b[5]!=1||u16(18)!=8)throw std::runtime_error("Expected little-endian MIPS ELF32");
  entry=u32(24);names=u16(50);auto sh=u32(32);auto count=u16(48),stride=u16(46);
  if(count>4096||(count&&stride<40))throw std::runtime_error("Invalid ELF sections");
  for(unsigned i=0;i<count;i++){uint64_t a=uint64_t(sh)+i*stride;span(a,40);sections.push_back({u32(a),u32(a+4),u32(a+12),u32(a+16),u32(a+20),u32(a+24),u32(a+36)});}
 }
 std::string stringAt(const Section&s,uint32_t off)const{if(off>=s.size)return {};auto v=span(uint64_t(s.off)+off,s.size-off);return ps2_cstring(v);}
 template<class F>void segments(F use){auto n=u16(44),stride=u16(42);auto off=u32(28);if(n>1024||(n&&stride<32))throw std::runtime_error("Invalid ELF programs");for(unsigned i=0;i<n;i++){uint64_t p=uint64_t(off)+i*stride;span(p,32);if(u32(p)!=1)continue;auto addr=u32(p+8),filesz=u32(p+16),memsz=u32(p+20);if(memsz<filesz||memsz>32*1024*1024)throw std::runtime_error("Invalid ELF segment size");use(addr,span(u32(p+4),filesz),memsz);}}
 Slice<ps2_Symbol>symbols(){Slice<ps2_Symbol>out;for(auto&s:sections)if(s.type==2){if(s.link>=sections.size()||s.entsize<16)throw std::runtime_error("Invalid ELF symbol table");for(uint64_t o=0;o+16<=s.size;o+=s.entsize){auto a=uint64_t(s.off)+o;span(a,16);auto name=stringAt(sections[s.link],u32(a));if(!name.empty())out=append(out,ps2_Symbol{name,u32(a+4),u32(a+8),(b[a+12]&15)==2});}}std::stable_sort(out.begin(),out.end(),[](auto&a,auto&b){return a.Addr<b.Addr;});return out;}
};
inline std::tuple<ps2_Executable*,Error>ps2_LoadELF(Slice<uint8_t>raw){try{RRELF f(raw);auto*e=arenaNew(ps2_Executable{});e->Entry=f.entry;f.segments([&](uint32_t a,Slice<uint8_t>b,uint32_t n){if(n)e->Segments=append(e->Segments,ps2_Segment{a,b,n});});if(!e->Segments.n)throw std::runtime_error("ELF has no loadable segments");for(auto s:f.symbols())if(s.Addr){e->Symbols=append(e->Symbols,s);if(s.Func)e->byAddr=append(e->byAddr,s);}return {e,{}};}catch(const std::exception&e){return {nullptr,{e.what()}};}}
inline std::tuple<ps2_IRX*,Error>ps2_ReadIRX(Slice<uint8_t>raw){try{RRELF f(raw);auto*m=arenaNew(ps2_IRX{});f.segments([&](uint32_t a,Slice<uint8_t>b,uint32_t n){if(a)throw std::runtime_error("IRX linked at nonzero address");m->Image=b;m->MemSz=n;});if(!m->Image.n)throw std::runtime_error("IRX has no image");bool mod=false;for(auto&s:f.sections){if(f.names<f.sections.size()&&f.stringAt(f.sections[f.names],s.name)==".iopmod"){auto b=f.span(s.off,s.size);if(b.n<26)throw std::runtime_error("Short IOP module header");m->Entry=le_Uint32(sub(b,4,8));m->GP=le_Uint32(sub(b,8,12));m->Version=le_Uint16(sub(b,24,26));m->Name=ps2_cstring(sub(b,26,b.n));if(m->Name.empty())m->Name="?";mod=true;}if(s.type==9){auto b=f.span(s.off,s.size);if(b.n%8)throw std::runtime_error("Invalid IRX relocations");for(int64_t i=0;i<b.n;i+=8)m->Relocs=append(m->Relocs,ps2_IRXReloc{le_Uint32(sub(b,i,i+4)),uint8_t(le_Uint32(sub(b,i+4,i+8)))});}}if(!mod)throw std::runtime_error("IRX has no .iopmod section");m->Symbols=f.symbols();ps2_IRX_scanLibraries(m);return {m,{}};}catch(const std::exception&e){return {nullptr,{e.what()}};}}
