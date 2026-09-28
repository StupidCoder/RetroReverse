r4300_CPU *r4300_NewCPU(n64_Machine *);
n64_Machine *n64_NewMachine(n64_ROM *);
template <class... A> void r4300_CPU_Halt(r4300_CPU *c, std::string f, A... a) {
  c->Halted = true;
  c->HaltReason = go_fmt_Sprintf(f, a...);
}
template <class... A> void rsp_CPU_Halt(rsp_CPU *c, std::string f, A... a) {
  c->Halted = true;
  c->HaltReason = go_fmt_Sprintf(f, a...);
}
template <class... A> void n64_Machine_note(n64_Machine *m, std::string f, A... a) {
  auto s = go_fmt_Sprintf(f, a...);
  if (!m->logSeen[s]) {
    m->logSeen[s] = true;
    m->Log = append(m->Log, s);
  }
}
template <class... A> void n64_regFile_init(n64_regFile r, A... args) {
  uint32_t a[] = {uint32_t(args)...};
  for (size_t i = 0; i + 1 < sizeof...(args); i += 2)
    r[a[i]] = a[i + 1];
}
constexpr std::array<uint32_t, 32> rsp_memShift = {0, 1, 2, 3, 4, 4, 3, 3, 4, 4, 4, 4};
uint32_t n64_Machine_Fetch32(n64_Machine *, uint32_t);
uint32_t n64_Machine_Read32(n64_Machine *, uint32_t);
void n64_Machine_Write32(n64_Machine *, uint32_t, uint32_t);
