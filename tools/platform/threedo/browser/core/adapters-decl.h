template <class... A> void arm60_CPU_Halt(arm60_CPU *c, std::string f, A... a) {
  c->Halted = true;
  c->HaltReason = go_fmt_Sprintf(f, a...);
}
std::tuple<threedo_Volume *, Error> threedo_Open(Slice<uint8_t>);
std::tuple<Slice<uint8_t>, Error> threedo_Volume_block(threedo_Volume *, int64_t);

uint32_t arm60_CPU_read32aligned(arm60_CPU *, uint32_t);
void arm60_CPU_write32(arm60_CPU *, uint32_t, uint32_t);
