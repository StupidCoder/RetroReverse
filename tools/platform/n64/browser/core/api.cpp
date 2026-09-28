#include "host.h"
static n64_Machine *machine = nullptr;
static Slice<uint8_t> upload;
static uint64_t steps = 0;
static std::string errorText, reply;
static std::vector<uint8_t> pixels;
extern "C" {
uint8_t *rr_input(uint32_t n) {
  try {
    if (n > 64 * 1024 * 1024)
      throw std::runtime_error("ROM exceeds 64 MiB");
    upload = Slice<uint8_t>::make(n);
    return upload.p;
  } catch (const std::exception &e) {
    errorText = e.what();
    return nullptr;
  }
}
int rr_init(uint32_t n) {
  try {
    if (upload.n != n)
      throw std::runtime_error("ROM upload size mismatch");
    auto candidate = boot(upload);
    destroy(machine);
    machine = candidate;
    steps = 0;
    errorText.clear();
    return 1;
  } catch (const std::exception &e) {
    errorText = e.what();
    return 0;
  }
}
int rr_run(uint32_t n) {
  try {
    if (!machine)
      throw std::runtime_error("No cartridge loaded");
    if (n > 50000)
      throw std::runtime_error("Run slice too large");
    auto r = n64_Machine_Run(machine, n);
    steps += r.Steps;
    if (machine->CPU->Halted)
      throw std::runtime_error(machine->CPU->HaltReason);
    return int(r.Steps);
  } catch (const std::exception &e) {
    errorText = e.what();
    return -1;
  } catch (...) {
    errorText = "Unexpected core exception";
    return -1;
  }
}
void rr_pad(uint32_t buttons, int x, int y) {
  if (machine) {
    machine->Controllers[0].Buttons = buttons;
    machine->Controllers[0].StickX = std::clamp(x, -80, 80);
    machine->Controllers[0].StickY = std::clamp(y, -80, 80);
  }
}
const char *rr_error() {
  return errorText.c_str();
}
const char *rr_proof() {
  reply = proof(machine, steps);
  return reply.c_str();
}
const char *rr_status() {
  auto w = n64_Machine_Width(machine);
  if (w == 0 || w > 1024)
    w = 320;
  std::ostringstream s;
  s << "{\"steps\":" << steps << ",\"pc\":" << uint32_t(machine->CPU->PC) << ",\"width\":" << w
    << ",\"height\":" << height(machine) << ",\"rspSteps\":" << machine->rspSteps
    << ",\"rdpWords\":" << machine->rdpWords << ",\"polls\":" << machine->ContPolls << "}";
  reply = s.str();
  return reply.c_str();
}
uint8_t *rr_frame() {
  pixels = frame(machine);
  return pixels.data();
}
}
