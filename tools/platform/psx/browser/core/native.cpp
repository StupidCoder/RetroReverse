#include "proof.h"
#include <chrono>
#include <fstream>
int main(int argc, char **argv) {
  try {
    if (argc < 3) {
      fprintf(stderr, "usage: native DISC OUTPUT_PREFIX [MAX_STEPS]\n");
      return 2;
    }
    std::ifstream f(argv[1], std::ios::binary);
    std::vector<u8> bytes{std::istreambuf_iterator<char>(f), {}};
    auto d = std::make_shared<Disc>();
    d->open(bytes.data(), bytes.size());
    auto m = std::make_unique<Machine>();
    m->boot(d, 0);
    std::vector<std::pair<u64, u16>> inputs = {
        {380000000, 0xfff7}, {380380000, 65535},  {386000000, 0xbfff}, {386380000, 65535},
        {520000000, 0xbfff}, {550000000, 0xbf7f}, {560000000, 0xbfff}, {570000000, 65535}};
    std::vector<u64> targets = {0,         1000000,   10000000,  50000000,  100000000,
                                200000000, 300000000, 370000000, 380000000, 380380000,
                                386000000, 386380000, 430000000, 500000000, 520000000,
                                530000000, 550000000, 560000000, 570000000, 580000000};
    u64 limit = argc > 3 ? std::stoull(argv[3]) : 580000000;
    std::ofstream out(std::string(argv[2]) + ".jsonl");
    size_t cursor = 0;
    auto start = std::chrono::steady_clock::now();
    for (u64 target : targets) {
      if (target > limit)
        break;
      while (m->cpu.steps < target) {
        while (cursor < inputs.size() && inputs[cursor].first <= m->cpu.steps)
          m->buttons = inputs[cursor++].second;
        u64 next = cursor < inputs.size() ? std::min(target, inputs[cursor].first) : target;
        m->run(u32(std::min<u64>(20000, next - m->cpu.steps)));
      }
      out << proof(*m) << "\n";
      out.flush();
      fprintf(stderr, "steps %llu PC %08X fields %llu\n", (unsigned long long)m->cpu.steps,
              m->cpu.pc, (unsigned long long)m->fields);
      if (target >= 370000000) {
        auto frame = m->gpu.frame();
        std::ofstream ppm(std::string(argv[2]) + "-" + std::to_string(target) + ".ppm",
                          std::ios::binary);
        ppm << "P6\n" << m->gpu.dispW << " " << m->gpu.dispH << "\n255\n";
        for (auto px : frame) {
          char b[] = {char(px), char(px >> 8), char(px >> 16)};
          ppm.write(b, 3);
        }
      }
    }
    fprintf(stderr, "%.3f seconds; TTY %s\n",
            std::chrono::duration<double>(std::chrono::steady_clock::now() - start).count(),
            m->tty.c_str());
    for (auto &n : m->diagnostics)
      fprintf(stderr, "%s\n", n.c_str());
  } catch (const std::exception &e) {
    fprintf(stderr, "ERROR: %s\n", e.what());
    return 1;
  }
}
