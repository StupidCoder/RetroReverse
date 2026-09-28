#include "amiga.h"
#include "state.h"
#include <memory>
namespace {
std::unique_ptr<rramiga::Machine> machine;
std::vector<uint8_t> input, firmware, stateInput, stateOutput;
std::string errorText, reply;
uint32_t hash(const uint8_t *p, size_t n) {
  uint32_t h = 2166136261;
  for (size_t i = 0; i < n; i++)
    h = (h ^ p[i]) * 16777619;
  return h;
}
std::vector<uint8_t> memoryImage() {
  using namespace rramiga;
  std::vector<uint8_t> b(captureSize);
  std::memcpy(b.data(), machine->ram.data(), CHIP);
  std::memcpy(b.data() + CHIP, machine->slow.data(), SLOW);
  for (unsigned i = 0; i < 256; i++) {
    b[registerBase + i * 2] = machine->reg[i] >> 8;
    b[registerBase + i * 2 + 1] = machine->reg[i];
  }
  std::memcpy(b.data() + frameBase, machine->draw.data(), W * H * 4);
  return b;
}
} // namespace
extern "C" {
uint8_t *rr_input(uint32_t n) {
  try {
    if (n > 901120)
      throw std::runtime_error("ADF exceeds 880 KiB");
    input.resize(n);
    return input.data();
  } catch (const std::exception &e) {
    errorText = e.what();
    return nullptr;
  }
}
uint8_t *rr_firmware(uint32_t n) {
  try {
    if (n != 262144 && n != 524288)
      throw std::runtime_error("Kickstart must be 256 or 512 KiB");
    firmware.resize(n);
    return firmware.data();
  } catch (const std::exception &e) {
    errorText = e.what();
    return nullptr;
  }
}
int rr_init(uint32_t n) {
  try {
    if (n != input.size() || firmware.empty())
      throw std::runtime_error("Missing disk or Kickstart firmware");
    machine = std::make_unique<rramiga::Machine>();
    machine->reset(firmware, input);
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_run(uint32_t n) {
  try {
    if (!machine || !n || n > 1000000)
      throw std::runtime_error("Invalid execution slice");
    return machine->run(n);
  } catch (const std::exception &e) {
    errorText = e.what();
    return -1;
  }
}
void rr_pad(uint32_t mask) {
  if (machine) {
    machine->buttons = mask & 31;
    machine->mouseLeft = mask & 32;
    machine->mouseRight = mask & 64;
  }
}
void rr_mouse(int x, int y) {
  if (machine) {
    machine->mouseX += std::clamp(x, -127, 127);
    machine->mouseY += std::clamp(y, -127, 127);
  }
}
void rr_key(unsigned code, unsigned down) {
  if (machine && code < 128 && machine->keys.size() < 1024)
    machine->key(code, down);
}
const char *rr_error() { return errorText.c_str(); }
const char *rr_status() {
  std::ostringstream s;
  s << "{\"steps\":" << machine->steps << ",\"cycles\":" << machine->cycles
    << ",\"frames\":" << machine->frames
    << ",\"pc\":" << m68k_get_reg(nullptr, M68K_REG_PC)
    << ",\"seconds\":" << double(machine->cycles) / 7093790
    << ",\"inputSeconds\":" << double(machine->cycles) / 7093790
    << ",\"width\":640,\"height\":256,\"diskReads\":" << machine->diskReads
    << ",\"track\":" << machine->cylinder * 2 + machine->side << "}";
  reply = s.str();
  return reply.c_str();
}
const char *rr_profile() { return rrprof::json(); }
uint8_t *rr_frame() {
  return reinterpret_cast<uint8_t *>(machine->screen.data());
}
const char *rr_proof() {
  auto b = memoryImage();
  std::ostringstream s;
  s << "{\"steps\":" << machine->steps << ",\"frames\":" << machine->frames
    << ",\"pc\":" << m68k_get_reg(nullptr, M68K_REG_PC)
    << ",\"memory\":" << hash(b.data(), b.size())
    << ",\"rgba\":" << hash(rr_frame(), rramiga::W * rramiga::H * 4) << "}";
  reply = s.str();
  return reply.c_str();
}
uint8_t *rr_state_input(uint32_t n) {
  try {
    if (n > rrstate::limit)
      throw std::runtime_error("State too large");
    stateInput.resize(n);
    return stateInput.data();
  } catch (const std::exception &e) {
    errorText = e.what();
    return nullptr;
  }
}
uint32_t rr_state_save() {
  try {
    stateOutput = rramiga::save();
    return stateOutput.size();
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
uint8_t *rr_state_data() { return stateOutput.data(); }
int rr_state_load(uint32_t n) {
  try {
    if (n != stateInput.size())
      throw std::runtime_error("State length mismatch");
    rramiga::restore(stateInput.data(), n);
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_capture_begin() {
  try {
    auto b = memoryImage();
    rrcapture::trace.begin(b.data(), b.size());
    rramiga::hardwareEvent = rramiga::diskEvent = 0;
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_capture_end() {
  try {
    auto b = memoryImage();
    rrcapture::trace.end(b.data(), b.size());
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
const char *rr_capture_info() {
  reply = rrcapture::trace.info();
  return reply.c_str();
}
const char *rr_pixel(int x, int y) {
  uint32_t a = x >= 0 && x < 640 && y >= 0 && y < 256
                   ? rramiga::frameBase + (y * 640 + x) * 4
                   : UINT32_MAX;
  reply = rrcapture::trace.pixel(a, 4);
  return reply.c_str();
}
const char *rr_source(uint32_t a, int size, uint32_t before,
                      uint32_t expected) {
  reply = rrcapture::trace.pixel(a, size, before, expected);
  return reply.c_str();
}
const char *rr_resource(uint32_t id, uint32_t offset) {
  reply = rrcapture::trace.resourceJSON(id, offset);
  return reply.c_str();
}
const char *rr_replay_begin() {
  rrreplay::replay.begin();
  reply = rrreplay::replay.info();
  return reply.c_str();
}
int rr_replay_seek(uint32_t n) { return rrreplay::replay.seek(n); }
const char *rr_replay_info() {
  reply = rrreplay::replay.info();
  return reply.c_str();
}
uint32_t rr_replay_for_write(uint32_t id) {
  return rrreplay::replay.forWrite(id);
}
uint8_t *rr_replay_frame() {
  auto &b = rrreplay::replay.memory;
  return b.size() >= rramiga::captureSize ? b.data() + rramiga::frameBase
                                          : nullptr;
}
}
