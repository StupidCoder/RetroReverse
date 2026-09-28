#include "amiga.h"
#include "state.h"
#include <chrono>
#include <cstdlib>
#include <fstream>
#include <memory>
static std::vector<uint8_t> readFile(const char *p) {
  std::ifstream f(p, std::ios::binary);
  if (!f)
    throw std::runtime_error(std::string("Cannot open ") + p);
  return {std::istreambuf_iterator<char>(f), {}};
}
int main(int argc, char **argv) {
  try {
    if (argc < 4) {
      std::fprintf(
          stderr,
          "usage: amiga-native kick.rom game.adf frames [screen.ppm]\n");
      return 2;
    }
    auto m = std::make_unique<rramiga::Machine>();
    m->reset(readFile(argv[1]), readFile(argv[2]));
    m->logging = std::getenv("RR_LOG");
    unsigned frames = std::stoul(argv[3]);
    auto begin = std::chrono::steady_clock::now();
    if (auto p = std::getenv("RR_LOAD")) {
      auto b = readFile(p);
      rramiga::restore(b.data(), b.size());
    }
    struct Event {
      unsigned frame;
      std::string type;
      int a, b, c;
    };
    std::vector<Event> events;
    if (auto p = std::getenv("RR_EVENTS")) {
      std::ifstream f(p);
      Event e;
      while (f >> e.frame >> e.type >> e.a >> e.b >> e.c)
        events.push_back(e);
    }
    while (m->frames < frames) {
      auto f = m->frames;
      while (m->frames == f)
        m->run(10000);
      if (std::getenv("RR_FIRE"))
        m->buttons =
            m->frames > unsigned(std::atoi(std::getenv("RR_FIRE"))) ? 16 : 0;
      if (auto p = std::getenv("RR_MOUSE")) {
        unsigned f = std::atoi(p);
        m->mouseLeft = m->frames > f && m->frames < f + 10;
      }
      for (auto &e : events)
        if (e.frame == m->frames) {
          if (e.type == "mouse") {
            m->mouseX += e.a;
            m->mouseY += e.b;
            m->mouseLeft = e.c & 1;
            m->mouseRight = e.c & 2;
          } else if (e.type == "pad")
            m->buttons = e.a;
          else if (e.type == "key")
            m->key(e.a, e.b);
        }
      if (m->frames % 100 == 0)
        std::fprintf(
            stderr,
            "frame %llu PC=%06x SR=%04x disk=%llu track=%d DMA=%04x "
            "INT=%04x/%04x copper=%06x blits=%llu\n",
            (unsigned long long)m->frames, m68k_get_reg(nullptr, M68K_REG_PC),
            m68k_get_reg(nullptr, M68K_REG_SR),
            (unsigned long long)m->diskReads, m->cylinder * 2 + m->side, m->dma,
            m->intena, m->intreq, m->copperPC, (unsigned long long)m->blits);
    }
    if (argc > 4) {
      std::ofstream f(argv[4], std::ios::binary);
      f << "P6\n" << rramiga::W << ' ' << rramiga::H << "\n255\n";
      for (auto c : m->screen) {
        f.put(c);
        f.put(c >> 8);
        f.put(c >> 16);
      }
    }
    if (auto p = std::getenv("RR_SAVE")) {
      auto b = rramiga::save();
      std::ofstream f(p, std::ios::binary);
      f.write((char *)b.data(), b.size());
    }
    if (auto p = std::getenv("RR_RAM")) {
      std::ofstream f(p, std::ios::binary);
      f.write((char *)m->ram.data(), m->ram.size());
      std::ofstream g(std::string(p) + ".regs");
      for (unsigned i = 0; i < 256; i++)
        g << std::hex << i * 2 << ' ' << m->reg[i] << '\n';
    }
    std::printf(
        "frames=%llu steps=%llu seconds=%.3f PC=%06x diskReads=%llu "
        "copper=%llu blits=%llu\n",
        (unsigned long long)m->frames, (unsigned long long)m->steps,
        std::chrono::duration<double>(std::chrono::steady_clock::now() - begin)
            .count(),
        m68k_get_reg(nullptr, M68K_REG_PC), (unsigned long long)m->diskReads,
        (unsigned long long)m->copperMoves, (unsigned long long)m->blits);
  } catch (const std::exception &e) {
    std::fprintf(stderr, "%s\n", e.what());
    return 1;
  }
}
