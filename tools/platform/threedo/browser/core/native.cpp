#include "host.h"
int main(int argc, char **argv) {
  try {
    if (argc < 3)
      throw std::runtime_error("usage: threedo-native DISC PREFIX [frames]");
    nativeDisc(argv[1]);
    auto m = boot();
    auto end = argc > 3 ? std::stoull(argv[3]) : 300;
    std::ofstream out(std::string(argv[2]) + ".jsonl");
    std::vector<std::pair<uint64_t, uint32_t>> inputs;
    if (argc > 4) {
      std::ifstream in(argv[4]);
      uint64_t f;
      uint32_t b;
      while (in >> f >> b)
        inputs.push_back({f, b});
    }
    auto start = std::chrono::steady_clock::now();
    for (uint64_t f = 0; f <= end; f++) {
      if (f > 0) {
        for (auto [at, b] : inputs)
          if (at == f - 1)
            threedo_Machine_SendPadEvent(m, b);
        nextFrame(m);
      }
      if (f == 0 || f == 1 || f % 30 == 0) {
        auto s = proof(m);
        out << s << std::endl;
        std::cerr << s << std::endl;
        auto pixels = frame(m);
        std::ofstream ppm(std::string(argv[2]) + "-" + std::to_string(f) + ".ppm",
                          std::ios::binary);
        ppm << "P6\n320 240\n255\n";
        for (size_t i = 0; i < pixels.size(); i += 4)
          ppm.write((char *)&pixels[i], 3);
      }
    }
    std::cerr << "seconds "
              << std::chrono::duration<double>(std::chrono::steady_clock::now() - start).count()
              << " discBytes " << discBytesRead << "\n";
    arenaClear();
  } catch (const std::exception &e) {
    std::cerr << "ERROR " << e.what() << "\n";
    return 1;
  }
}
