#include "proof.h"
#include <memory>
#include <emscripten.h>
// clang-format off
EM_JS(int, browserRead, (uint32_t offset, uint8_t *dest, uint32_t n), {
  try {
    if (!Module.discFile) throw Error('No local disc mounted');
    const bytes = new Uint8Array(new FileReaderSync().readAsArrayBuffer(Module.discFile.slice(offset, offset+n)));
    if (bytes.length !== n) throw Error('Short local disc read');
    HEAPU8.set(bytes,dest); return 1;
  } catch(e) { Module.discError=String(e); return 0; }
});
// clang-format on
static std::vector<u8> input;
static std::unique_ptr<Machine> machine;
static std::unique_ptr<Machine> saved[4];
static std::vector<u32> frame;
static std::string text, error;
static u32 generation = 0;
struct Command {
  u32 pc;
  u64 step;
  std::vector<u32> words;
  u32 writes = 0;
};
struct PixelWrite {
  u32 command, previous;
  u16 value;
};
static std::vector<Command> commands;
static std::vector<PixelWrite> writes;
static std::vector<u32> last(1024 * 512);
static bool capturing = false, captured = false;
static u32 overflow = 0, originX = 0, originY = 0;
static constexpr size_t COMMAND_CAP = 16384, WRITE_CAP = 2 * 1024 * 1024;
static void clearCapture() {
  if (machine) {
    machine->gpu.onCommand = {};
    machine->gpu.onPixel = {};
  }
  capturing = captured = false;
}
static void requireMachine() {
  if (!machine)
    throw std::runtime_error("Load a disc first");
}
template <class F> static int guard(F fn) {
  try {
    error.clear();
    return fn();
  } catch (const std::exception &e) {
    error = e.what();
    return -1;
  }
}
extern "C" {
u8 *rr_input(int n) {
  if (n < 0 || n > 128 * 1024 * 1024) {
    error = "Disc exceeds the 128 MiB input limit";
    return nullptr;
  }
  input.resize(n);
  return input.data();
}
int rr_init(int n, u32 handler) {
  return guard([&]() {
    if (n < 0 || size_t(n) > input.size())
      throw std::runtime_error("Invalid input length");
    auto d = std::make_shared<Disc>();
    d->open(input.data(), n);
    auto candidate = std::make_unique<Machine>();
    candidate->boot(d, handler);
    clearCapture();
    machine = std::move(candidate);
    for (auto &s : saved)
      s.reset();
    generation++;
    return 1;
  });
}
int rr_init_file(uint32_t n) {
  return guard([&]() {
    auto d = std::make_shared<Disc>();
    d->openFile(n, [](size_t o,u8*p,size_t n){
      if (!browserRead(o,p,n)) throw std::runtime_error("Local disc read failed");
    });
    auto candidate = std::make_unique<Machine>();
    candidate->boot(d,0);
    clearCapture(); machine=std::move(candidate);
    for(auto &s:saved) s.reset();
    generation++; return 1;
  });
}
const char* rr_profile(){return rrprof::json();}
int rr_run(int n, int stopField) {
  return guard([&]() {
    requireMachine();
    if (n < 1 || n > 250000)
      throw std::runtime_error("Instruction budget must be 1–250,000");
    captured = false;
    rrprof::Scope clock(0,"R3000A / devices remainder");
    return int(machine->run(n, stopField));
  });
}
void rr_pad(int buttons) {
  if (machine)
    machine->buttons = u16(buttons);
}
const char *rr_error() {
  return error.c_str();
}
const char *rr_status() {
  if (!machine)
    return "{\"loaded\":false}";
  auto &m = *machine;
  std::ostringstream s;
  s << "{\"loaded\":true,\"generation\":" << generation << ",\"steps\":" << m.cpu.steps
    << ",\"fields\":" << m.fields << ",\"pc\":" << m.cpu.pc << ",\"width\":" << m.gpu.dispW
    << ",\"height\":" << m.gpu.dispH << ",\"displayX\":" << m.gpu.dispX
    << ",\"displayY\":" << m.gpu.dispY << ",\"drawX\":" << m.gpu.offX << ",\"drawY\":" << m.gpu.offY
    << ",\"commands\":" << m.gpu.commands << ",\"cdCommands\":" << m.cd.commands
    << ",\"pad\":" << m.buttons << ",\"padReady\":" << (m.padActive ? "true" : "false")
    << ",\"captured\":" << (captured ? "true" : "false") << ",\"disc\":" << quote(m.disc->name)
    << "}";
  text = s.str();
  return text.c_str();
}
const char *rr_proof() {
  if (!machine)
    return "{}";
  text = proof(*machine);
  return text.c_str();
}
u32 *rr_frame(int draw) {
  if (!machine)
    return nullptr;
  frame = machine->gpu.frame(draw);
  return frame.data();
}
u8 *rr_ram() {
  return machine ? machine->ram.data() : nullptr;
}
int rr_save(int slot) {
  return guard([&]() {
    requireMachine();
    if (slot < 0 || slot >= 4)
      throw std::runtime_error("Invalid checkpoint slot");
    auto s = std::make_unique<Machine>(*machine);
    s->gpu.onCommand = {};
    s->gpu.onPixel = {};
    saved[slot] = std::move(s);
    return 1;
  });
}
int rr_restore(int slot) {
  return guard([&]() {
    requireMachine();
    if (slot < 0 || slot >= 4 || !saved[slot])
      throw std::runtime_error("That checkpoint has not been saved");
    clearCapture();
    *machine = *saved[slot];
    return 1;
  });
}
int rr_capture_begin() {
  return guard([&]() {
    requireMachine();
    clearCapture();
    commands.clear();
    writes.clear();
    writes.push_back({0, 0, 0});
    std::fill(last.begin(), last.end(), 0);
    overflow = 0;
    capturing = true;
    machine->gpu.onCommand = [](const std::vector<u32> &w) {
      if (commands.size() == COMMAND_CAP) {
        overflow++;
        return;
      }
      commands.push_back({machine->cpu.cur, machine->cpu.steps, w, 0});
    };
    machine->gpu.onPixel = [](int x, int y, u16 value) {
      if (commands.empty() || overflow)
        return;
      auto id = u32(commands.size());
      commands.back().writes++;
      if (writes.size() == WRITE_CAP) {
        overflow++;
        return;
      }
      u32 at = y * 1024 + x;
      writes.push_back({id, last[at], value});
      last[at] = u32(writes.size() - 1);
    };
    return 1;
  });
}
int rr_capture_end() {
  return guard([&]() {
    requireMachine();
    if (!capturing)
      throw std::runtime_error("No capture is active");
    machine->gpu.onCommand = {};
    machine->gpu.onPixel = {};
    capturing = false;
    captured = true;
    originX = machine->gpu.offX;
    originY = machine->gpu.offY;
    return int(commands.size());
  });
}
const char *rr_pixel(int x, int y) {
  std::ostringstream s;
  if (!captured || x < 0 || y < 0 || x >= machine->gpu.dispW || y >= machine->gpu.dispH)
    return "{\"error\":\"Capture a drawing frame before selecting a pixel\"}";
  u32 vx = (x + originX) & 1023, vy = (y + originY) & 511;
  s << "{\"x\":" << x << ",\"y\":" << y << ",\"vramX\":" << vx << ",\"vramY\":" << vy
    << ",\"overflow\":" << overflow << ",\"commands\":" << commands.size() << ",\"history\":[";
  bool comma = false;
  std::vector<u32> chain;
  u32 cursor = last[vy * 1024 + vx];
  for (; cursor && chain.size() < 512; cursor = writes[cursor].previous)
    chain.push_back(cursor);
  std::reverse(chain.begin(), chain.end());
  for (u32 i : chain) {
    const auto &p = writes[i];
    const auto &c = commands[p.command - 1];
    if (comma)
      s << ",";
    comma = true;
    s << "{\"command\":" << p.command << ",\"pc\":" << c.pc << ",\"step\":" << c.step
      << ",\"value\":" << p.value << ",\"writes\":" << c.writes << ",\"words\":[";
    for (size_t j = 0; j < c.words.size(); j++) {
      if (j)
        s << ",";
      s << c.words[j];
    }
    s << "]}";
  }
  s << "],\"historyLimit\":512,\"truncated\":" << (cursor ? "true" : "false") << "}";
  text = s.str();
  return text.c_str();
}
const char *rr_diagnostics() {
  requireMachine();
  std::ostringstream s;
  s << "{\"tty\":" << quote(machine->tty) << ",\"messages\":[";
  for (size_t i = 0; i < machine->diagnostics.size(); i++) {
    if (i)
      s << ",";
    s << quote(machine->diagnostics[i]);
  }
  s << "]}";
  text = s.str();
  return text.c_str();
}
}
#include "state.h"
static void stateWrite(rrstate::Archive&a){requireMachine();a.header(2,1);a(*machine);}
static void stateRead(rrstate::Archive&a){requireMachine();a.header(2,1);auto next=std::make_unique<Machine>();a(*next);a.finish();validateState(*next);next->disc=machine->disc;clearCapture();machine=std::move(next);for(auto&s:saved)s.reset();}
#include "../../../../browser/state/api.inc"
