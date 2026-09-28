#include "host.h"
#include <emscripten.h>
// clang-format off
EM_JS(int, browserRead, (uint32_t offset, uint8_t *dest, uint32_t n), {
  try {
    let bytes;
    if (Module.discFile)
      bytes = new Uint8Array(
          new FileReaderSync().readAsArrayBuffer(Module.discFile.slice(offset, offset + n)));
    else {
      const x = new XMLHttpRequest();
      x.open('GET', '/fixtures/need-for-speed.bin', false);
      x.responseType = 'arraybuffer';
      x.setRequestHeader('Range', 'bytes=' + offset + '-' + (offset + n - 1));
      x.send();
      if (x.status !== 206)
        throw Error('Disc server must support byte ranges (HTTP ' + x.status + ')');
      bytes = new Uint8Array(x.response);
    }
    if (bytes.length !== n)
      throw Error('Short disc read');
    HEAPU8.set(bytes, dest);
    return 1;
  } catch (e) {
    Module.discError = String(e);
    return 0;
  }
});
// clang-format on
static threedo_Machine *machine = nullptr;
static std::string errorText, reply;
static std::vector<uint8_t> pixels;
extern "C" {
int rr_init_config(uint32_t size, int nfsProfile) {
  try {
    machine = nullptr;
    discSize = size;
    discRead = [](uint64_t o, uint8_t *p, size_t n) {
      if (!browserRead(o, p, n))
        throw std::runtime_error("Disc read failed; check the file or server");
    };
    machine = boot(nfsProfile != 0);
    errorText.clear();
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_init(uint32_t size) { return rr_init_config(size, 1); }
int rr_run_slice(uint32_t n) {
  try { if(!machine || n>50000) throw std::runtime_error("Invalid execution slice");
    runSlice(machine,n); return 1;
  } catch(const std::exception&e){errorText=e.what();return 0;}
}
const char* rr_profile(){return rrprof::json();}
int rr_run() {
  try {
    if (!machine)
      throw std::runtime_error("No disc loaded");
    nextFrame(machine);
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
void rr_pad(uint32_t b) {
  if (machine)
    threedo_Machine_SendPadEvent(machine, b);
}
const char *rr_error() {
  return errorText.c_str();
}
const char *rr_proof() {
  reply = proof(machine);
  return reply.c_str();
}
const char *rr_status() {
  std::ostringstream s;
  s << "{\"frame\":" << machine->frame << ",\"steps\":" << totalSteps
    << ",\"pc\":" << machine->CPU->R[15] << ",\"discBytes\":" << discBytesRead
    << ",\"discReads\":" << discReads << "}";
  reply = s.str();
  return reply.c_str();
}
uint8_t *rr_frame() {
  pixels = frame(machine);
  return pixels.data();
}
}
#include "state.h"
static std::vector<std::shared_ptr<void>> stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(4,1);a(machine,totalSteps,runContext);}
static void stateRead(rrstate::Archive&a){a.header(4,1);threedo_Machine*next=nullptr;uint64_t ticks=0;RunContext context;a(next,ticks,context);a.finish();if(!next)throw std::runtime_error("Missing machine");rebindState(next);stateOwners=std::move(a.owned);machine=next;totalSteps=ticks;runContext=std::move(context);fileEntries.clear();discCache.clear();}
#include "../../../../browser/state/api.inc"
