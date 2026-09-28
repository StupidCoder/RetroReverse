#include "host.h"
int main(int argc, char **argv) {
  try {
    if (argc < 3)
      throw std::runtime_error("usage: n64-native ROM PREFIX [steps]");
    std::ifstream f(argv[1], std::ios::binary);
    std::vector<uint8_t> d((std::istreambuf_iterator<char>(f)), {});
    auto rom = Slice<uint8_t>::make(d.size());
    std::copy(d.begin(), d.end(), rom.p);
    auto m = boot(rom);
    std::ofstream out(std::string(argv[2]) + ".jsonl");
    uint64_t end = argc > 3 ? std::stoull(argv[3]) : 300000000, steps = 0;
    std::unordered_map<uint64_t, std::array<int, 3>> inputs;
    std::vector<uint64_t> targets = {0,         1,         1000,      100000,    1000000,
                                     10000000,  50000000,  100000000, 150000000, 200000000,
                                     225000000, 250000000, 275000000, 300000000, 400000000,
                                     500000000, 600000000, 750000000, 1000000000};
    if (argc > 4) {
      targets.clear();
      std::ifstream plan(argv[4]);
      uint64_t t;
      int b, x, y;
      while (plan >> t >> b >> x >> y) {
        targets.push_back(t);
        inputs[t] = {b, x, y};
      }
    }
    auto start = std::chrono::steady_clock::now();
    for (uint64_t target : targets) {
      if (target > end)
        break;
      if (target > steps) {
        auto res = n64_Machine_Run(m, target - steps);
        steps += res.Steps;
        if (m->CPU->Halted)
          throw std::runtime_error(m->CPU->HaltReason);
      }
      if (inputs.count(target)) {
        auto a = inputs[target];
        m->Controllers[0].Buttons = a[0];
        m->Controllers[0].StickX = a[1];
        m->Controllers[0].StickY = a[2];
      }
      auto s = proof(m, steps);
      out << s << std::endl;
      std::cerr << s << std::endl;
      auto pixels = frame(m);
      auto w = n64_Machine_Width(m);
      if (w > 0 && w <= 1024) {
        std::ofstream ppm(std::string(argv[2]) + "-" + std::to_string(steps) + ".ppm",
                          std::ios::binary);
        ppm << "P6\n" << w << " " << height(m) << "\n255\n";
        for (size_t i = 0; i < pixels.size(); i += 4)
          ppm.write((char *)&pixels[i], 3);
      }
    }
    std::cerr << "seconds "
              << std::chrono::duration<double>(std::chrono::steady_clock::now() - start).count()
              << "\n";
    destroy(m);
  } catch (const std::exception &e) {
    std::cerr << "ERROR " << e.what() << "\n";
    return 1;
  }
}
