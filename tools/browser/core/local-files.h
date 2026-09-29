#pragma once
#include <map>
#include <set>
#include <filesystem>
#include <fstream>
// Guest paths are virtual. Writes never reach the user-selected source files.
namespace rrfiles {
struct Source{uint64_t size{};std::function<size_t(uint64_t,uint8_t*,size_t)>read;};
inline std::map<std::string,Source>sources;
inline std::map<std::string,Slice<uint8_t>>overlay;
inline std::set<std::string>deleted,dirs;
inline std::string path(std::string p){for(auto&c:p){if(c=='\\')c='/';c=std::tolower(uint8_t(c));}std::string out;std::istringstream in(p);std::string v;while(std::getline(in,v,'/')){if(v.empty()||v==".")continue;if(v=="..")throw std::runtime_error("Path escapes selected game folder");if(!out.empty())out+='/';out+=v;}return out;}
inline bool exists(std::string p){p=path(p);return !deleted.contains(p)&&(overlay.contains(p)||sources.contains(p));}
inline int64_t size(std::string p){p=path(p);if(!exists(p))return -1;return overlay.contains(p)?overlay[p].n:sources[p].size;}
inline size_t read(std::string p,uint64_t at,uint8_t*out,size_t n){p=path(p);auto sz=size(p);if(sz<0||at>=uint64_t(sz))return 0;n=std::min(n,size_t(sz-at));if(overlay.contains(p))memcpy(out,overlay[p].p+at,n);else n=sources[p].read(at,out,n);return n;}
inline Slice<uint8_t>all(std::string p){auto n=size(p);if(n<0||n>256*1024*1024)throw std::runtime_error("Missing or oversized local file: "+p);auto b=Slice<uint8_t>::make(n);if(read(p,0,b.p,n)!=uint64_t(n))throw std::runtime_error("Short local file read");return b;}
inline bool directory(std::string p){p=path(p);if(p.empty()||dirs.contains(p))return true;auto pre=p+'/';for(auto&[k,v]:sources)if(k.starts_with(pre))return true;for(auto&[k,v]:overlay)if(k.starts_with(pre))return true;return false;}
inline void mountNative(const std::string&root){for(auto&e:std::filesystem::recursive_directory_iterator(root)){if(!e.is_regular_file())continue;auto p=e.path().string();auto name=path(std::filesystem::relative(e.path(),root).generic_string());sources[name]={e.file_size(),[p](uint64_t at,uint8_t*out,size_t n){std::ifstream f(p,std::ios::binary);f.seekg(at);f.read((char*)out,n);return size_t(f.gcount());}};}}
}
struct RRFileInfo {std::string name;int64_t size{};bool dir{};};
inline auto rrFileInfo_Name(RRFileInfo i){return i.name;}inline auto rrFileInfo_Size(RRFileInfo i){return i.size;}inline bool rrFileInfo_IsDir(RRFileInfo i){return i.dir;}
inline std::string go_filepath_Base(std::string p){p=rrfiles::path(p);auto i=p.rfind('/');return i==p.npos?p:p.substr(i+1);}
inline std::string go_filepath_Dir(std::string p){p=rrfiles::path(p);auto i=p.rfind('/');return i==p.npos?"":p.substr(0,i);}
inline std::string go_filepath_FromSlash(std::string p){return p;}
inline std::string go_filepath_Join(Slice<std::string>s){return rrfiles::path(go_strings_Join(s,"/"));}
inline std::tuple<RRFileInfo,Error>go_os_Stat(std::string p){if(rrfiles::exists(p))return {{go_filepath_Base(p),rrfiles::size(p),false},{}};if(rrfiles::directory(p))return {{go_filepath_Base(p),0,true},{}};return {{},{"File not found: "+p}};}
inline std::tuple<Slice<RRFileInfo>,Error>go_os_ReadDir(std::string p){p=rrfiles::path(p);if(!rrfiles::directory(p))return {{},{"Directory not found"}};auto pre=p.empty()?p:p+'/';std::map<std::string,RRFileInfo>items;auto add=[&](std::string k){if(!k.starts_with(pre)||rrfiles::deleted.contains(k))return;auto name=k.substr(pre.size());auto i=name.find('/');bool dir=i!=name.npos;name=name.substr(0,i);if(!name.empty())items[name]={name,dir?0:rrfiles::size(k),dir};};for(auto&[k,v]:rrfiles::sources)add(k);for(auto&[k,v]:rrfiles::overlay)add(k);for(auto&k:rrfiles::dirs)add(k+"/");Slice<RRFileInfo>out;for(auto&[k,v]:items)out=append(out,v);return {out,{}};}
inline std::tuple<os_File*,Error>go_os_OpenFile(std::string p,int64_t flags,uint32_t){p=rrfiles::path(p);if(!rrfiles::exists(p)&&!(flags&0x40))return {nullptr,{"File not found: "+p}};bool write=(flags&3)!=0;if(flags&0x200||!rrfiles::exists(p)){rrfiles::overlay[p]={};rrfiles::deleted.erase(p);}auto*f=arenaNew(os_File{p,0,write,false});return {f,{}};}
inline auto go_os_Open(std::string p){return go_os_OpenFile(p,0,0);}
inline auto go_os_Create(std::string p){return go_os_OpenFile(p,0x242,0);}
inline Error os_File_Close(os_File*f){if(f)f->closed=true;return {};}
inline std::tuple<int64_t,Error>os_File_Read(os_File*f,Slice<uint8_t>b){if(!f||f->closed)return {0,{"Closed file"}};auto n=rrfiles::read(f->name,f->offset,b.p,b.n);f->offset+=n;return {n,n==0&&b.n?go_io_EOF:Error{}};}
inline std::tuple<int64_t,Error>os_File_ReadAt(os_File*f,Slice<uint8_t>b,int64_t at){if(!f||f->closed||at<0)return {0,{"Invalid file read"}};auto n=rrfiles::read(f->name,at,b.p,b.n);return {n,n<uint64_t(b.n)?go_io_EOF:Error{}};}
inline std::tuple<int64_t,Error>os_File_Write(os_File*f,Slice<uint8_t>b){if(!f||f->closed||!f->writable)return {0,{"File is read only"}};if(f->offset<0||uint64_t(f->offset)+b.n>64*1024*1024)return {0,{"Virtual save file exceeds 64 MiB"}};if(!rrfiles::overlay.contains(f->name))rrfiles::overlay[f->name]=rrfiles::all(f->name);auto&out=rrfiles::overlay[f->name];if(f->offset+b.n>out.n){auto next=Slice<uint8_t>::make(f->offset+b.n);gcopy(next,out);out=next;}memcpy(out.p+f->offset,b.p,b.n);f->offset+=b.n;return {b.n,{}};}
inline std::tuple<int64_t,Error>os_File_Seek(os_File*f,int64_t off,int64_t whence){if(!f||f->closed||whence<0||whence>2)return {0,{"Invalid seek"}};auto at=off+(whence==1?f->offset:whence==2?rrfiles::size(f->name):0);if(at<0)return {0,{"Negative seek"}};f->offset=at;return {at,{}};}
inline std::tuple<Slice<uint8_t>,Error>go_os_ReadFile(std::string p){try{return {rrfiles::all(p),{}};}catch(const std::exception&e){return {{},{e.what()}};}}
inline Error go_os_WriteFile(std::string p,Slice<uint8_t>b,uint32_t){p=rrfiles::path(p);rrfiles::overlay[p]=Slice<uint8_t>::make(b.n);gcopy(rrfiles::overlay[p],b);rrfiles::deleted.erase(p);return {};}
inline Error go_os_MkdirAll(std::string p,uint32_t){rrfiles::dirs.insert(rrfiles::path(p));return {};}
inline std::tuple<std::string,Error>go_os_MkdirTemp(std::string,std::string){rrfiles::dirs.insert("__saves");return {"__saves",{}};}
inline Error go_os_Remove(std::string p){p=rrfiles::path(p);rrfiles::overlay.erase(p);rrfiles::deleted.insert(p);return {};}
