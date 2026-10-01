#include "disk.h"
#include <algorithm>
namespace rr::c64 {
static constexpr uint8_t codes[]={0x0a,0x0b,0x12,0x13,0x0e,0x0f,0x16,0x17,0x09,0x19,0x1a,0x1b,0x0d,0x1d,0x1e,0x15};
unsigned Disk::sectors(unsigned t){return t<1||t>40?0:t<=17?21:t<=24?19:t<=30?18:17;}
unsigned Disk::offset(unsigned t,unsigned s){unsigned n=s;for(unsigned i=1;i<t;i++)n+=sectors(i);return n*256;}
unsigned Disk::density(unsigned t){return t<=17?3:t<=24?2:t<=30?1:0;}
static void encode(std::vector<uint8_t>& out,std::span<const uint8_t> input){uint32_t bits=0;unsigned count=0;for(auto v:input)for(auto nibble:{v>>4,v&15}){bits=(bits<<5)|codes[nibble];count+=5;if(count>=8){count-=8;out.push_back(uint8_t(bits>>count));}}}
bool Disk::mount(std::span<const uint8_t> image,bool protect){
 if(image.size()>=8&&std::equal(image.begin(),image.begin()+8,"GCR-1541"))return mountG64(image,protect);
 unsigned count=0;if(image.size()==174848)count=35;else if(image.size()==196608)count=40;else return false;
 Disk next;next.trackCount=count;next.writeProtected=protect;const auto bam=offset(18,0);const auto id1=image[bam+162],id2=image[bam+163];
 const unsigned sizes[]={6250,6666,7142,7692};
 for(unsigned t=1;t<=count;t++){
  auto& out=next.tracks[(t-1)*2];const auto length=sizes[density(t)],n=sectors(t);
  for(unsigned s=0;s<n;s++){
   const unsigned end=(s+1)*length/n;out.insert(out.end(),5,255);
   const uint8_t header[]={8,uint8_t(s^t^id1^id2),uint8_t(s),uint8_t(t),id2,id1,15,15};encode(out,header);out.insert(out.end(),9,0x55);out.insert(out.end(),5,255);
   std::array<uint8_t,260> block{};block[0]=7;const auto start=offset(t,s);for(unsigned i=0;i<256;i++){block[i+1]=image[start+i];block[257]^=image[start+i];}encode(out,block);out.resize(end,0x55);
  }
  next.speeds[(t-1)*2].assign(out.size(),uint8_t(density(t)));
 }*this=std::move(next);return true;
}
bool Disk::changed()const{return std::any_of(dirty.begin(),dirty.end(),[](bool v){return v;});}
bool Disk::bit(unsigned track,uint32_t position)const{return halfBit(track*2,position);}
void Disk::writeBit(unsigned track,uint32_t position,bool value){writeHalfBit(track*2,position,value);}
bool Disk::halfBit(unsigned h,uint32_t position)const{if(h<2||h>85)return false;const auto& data=tracks[h-2];if(data.empty())return false;position%=uint32_t(data.size()*8);return (data[position/8]>>(7-(position&7)))&1;}
void Disk::writeHalfBit(unsigned h,uint32_t position,bool value){if(writeProtected||h<2||h>85)return;auto& data=tracks[h-2];if(data.empty())return;position%=uint32_t(data.size()*8);const auto mask=uint8_t(128>>(position&7));auto& byte=data[position/8];const auto next=uint8_t(value?byte|mask:byte&uint8_t(~mask));if(byte!=next){byte=next;dirty[h-2]=true;}}
bool Disk::mountG64(std::span<const uint8_t> image,bool protect){
 if(image.size()<12||!std::equal(image.begin(),image.begin()+8,"GCR-1541")||image[8]!=0)return false;
 const unsigned n=image[9],max=image[10]|unsigned(image[11])<<8;
 if(!n||n>84||!max||image.size()<12+8*n)return false;
 auto word=[&](unsigned p){return uint32_t(image[p])|uint32_t(image[p+1])<<8|uint32_t(image[p+2])<<16|uint32_t(image[p+3])<<24;};
 Disk next;next.g64=true;next.trackCount=(n+1)/2;next.writeProtected=protect;
 for(unsigned i=0;i<n;i++){
  const auto off=word(12+4*i),speed=word(12+4*n+4*i);if(!off)continue;
  if(off<12+8*n||off>image.size()-2)return false;const unsigned length=image[off]|unsigned(image[off+1])<<8;
  if(!length||length>max||length>image.size()-off-2)return false;
  next.tracks[i].assign(image.begin()+off+2,image.begin()+off+2+length);auto& zones=next.speeds[i];zones.resize(length);
  if(speed<4)std::fill(zones.begin(),zones.end(),uint8_t(speed));
  else{if(speed<12+8*n||speed>image.size()||(length+3)/4>image.size()-speed)return false;for(unsigned j=0;j<length;j++)zones[j]=(image[speed+j/4]>>(6-2*(j%4)))&3;}
 }*this=std::move(next);return true;
}
bool Disk::exportG64(std::vector<uint8_t>& image)const{
 if(!present())return false;unsigned max=0;for(const auto& t:tracks)max=std::max(max,unsigned(t.size()));if(!max||max>65535)return false;
 std::vector<uint8_t> out(12+84*8);std::copy_n("GCR-1541",8,out.begin());out[9]=84;out[10]=uint8_t(max);out[11]=uint8_t(max>>8);
 auto word=[&](unsigned p,uint32_t v){for(unsigned i=0;i<4;i++)out[p+i]=uint8_t(v>>(8*i));};
 for(unsigned i=0;i<84;i++){const auto& t=tracks[i];const auto& z=speeds[i];if(t.empty())continue;if(z.size()!=t.size())return false;
  word(12+4*i,uint32_t(out.size()));out.push_back(uint8_t(t.size()));out.push_back(uint8_t(t.size()>>8));out.insert(out.end(),t.begin(),t.end());
  if(std::all_of(z.begin(),z.end(),[&](uint8_t v){return v==z[0];}))word(12+84*4+4*i,z[0]);
  else{word(12+84*4+4*i,uint32_t(out.size()));for(unsigned j=0;j<z.size();j+=4){uint8_t v=0;for(unsigned k=0;k<4&&j+k<z.size();k++)v|=uint8_t(z[j+k]<<(6-2*k));out.push_back(v);}}
 }image=std::move(out);return true;
}
bool Disk::exportD64(std::vector<uint8_t>& image)const{
 if(!present()||g64)return false;std::vector<uint8_t> result(offset(trackCount,sectors(trackCount)));int8_t decode[32];std::fill(std::begin(decode),std::end(decode),int8_t(-1));for(unsigned i=0;i<16;i++)decode[codes[i]]=int8_t(i);
 for(unsigned t=1;t<=trackCount;t++){
  std::array<bool,21> found{};int sector=-1;unsigned ones=0;const unsigned total=unsigned(tracks[(t-1)*2].size()*8);
  for(unsigned pos=0;pos<total+3000;pos++){
   if(bit(t,pos)){++ones;continue;}if(ones<10){ones=0;continue;}ones=0;
   const auto block=[&](unsigned bytes,std::vector<uint8_t>& out){out.clear();unsigned at=pos;for(unsigned i=0;i<bytes;i++){uint8_t v=0;for(unsigned half=0;half<2;half++){unsigned code=0;for(unsigned b=0;b<5;b++)code=(code<<1)|bit(t,at++);if(decode[code]<0)return false;v=uint8_t((v<<4)|decode[code]);}out.push_back(v);}return true;};
   std::vector<uint8_t> data;if(!block(1,data))continue;
   if(data[0]==8){sector=-1;if(block(8,data)&&data[3]==t&&data[2]<sectors(t)&&data[1]==uint8_t(data[2]^data[3]^data[4]^data[5]))sector=data[2];}
   else if(data[0]==7&&sector>=0){if(!block(260,data))continue;uint8_t sum=0;for(unsigned i=1;i<=256;i++)sum^=data[i];if(sum!=data[257])continue;const auto off=offset(t,unsigned(sector));std::copy(data.begin()+1,data.begin()+257,result.begin()+off);found[unsigned(sector)]=true;sector=-1;}
  }
  for(unsigned s=0;s<sectors(t);s++)if(!found[s])return false;
 }image=std::move(result);return true;
}
}
