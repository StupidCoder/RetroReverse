#pragma once
#include <fstream>
#include <vector>
#include <cstring>
#include <stdexcept>
#include "profile.h"
#ifdef __EMSCRIPTEN__
#include <emscripten.h>
EM_JS(int, rrReadLocalDisc, (uint8_t* dst, double offset, int size), {
  try {
    const file = Module.discFile;
    if (!file) return 0;
    const bytes = new Uint8Array(new FileReaderSync().readAsArrayBuffer(file.slice(offset, offset + size)));
    if (bytes.length !== size) return 0;
    HEAPU8.set(bytes, dst);
    return 1;
  } catch (e) { Module.discError = String(e); return 0; }
});
#else
inline std::ifstream rrDiscFile;
inline int rrReadLocalDisc(uint8_t* dst, double offset, int size) {
  rrDiscFile.clear(); rrDiscFile.seekg(uint64_t(offset));
  rrDiscFile.read(reinterpret_cast<char*>(dst), size);
  return rrDiscFile.gcount() == size;
}
#endif
struct LocalDisc {
  uint64_t size=0, windowAt=UINT64_MAX, reads=0, bytesRead=0;
  std::vector<uint8_t> window;
  void mount(uint64_t bytes) {
    if (bytes < 32768 || bytes > 16ull*1024*1024*1024) throw std::runtime_error("Unsupported disc image size");
    size=bytes; windowAt=UINT64_MAX; window.clear(); reads=bytesRead=0;
  }
  void read(uint64_t offset, uint8_t* dst, size_t count) {
    if (offset>size || count>size-offset) throw std::runtime_error("Disc read outside selected image");
    while(count) {
      if(windowAt==UINT64_MAX || offset<windowAt || offset>=windowAt+window.size()) {
        windowAt=offset&~uint64_t(65535);
        window.resize(std::min<uint64_t>(262144,size-windowAt));
        rrprof::Scope timing(6,"Disc file reads");
        if(!rrReadLocalDisc(window.data(),double(windowAt),int(window.size()))) throw std::runtime_error("Cannot read selected disc image");
        reads++; bytesRead+=window.size();
      }
      auto n=std::min<uint64_t>(count,windowAt+window.size()-offset);
      std::memcpy(dst,window.data()+offset-windowAt,n); dst+=n; offset+=n; count-=n;
    }
  }
};
