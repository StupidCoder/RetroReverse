#pragma once
// Browser disc files stay sector-backed. The legacy NVRAM/write path retains
// its original semantics; an ordinary disc READ no longer materializes a movie.
inline std::unordered_map<std::string, threedo_Entry> fileEntries;
std::tuple<int32_t, int32_t> threedo_Machine_fileDeviceIO(threedo_Machine *m, std::string name,
                                                          uint32_t cmd, uint32_t offset,
                                                          uint32_t sendBuf, uint32_t sendLen,
                                                          uint32_t buf, uint32_t length) {
  if (!m->vol || m->NoStreams || m->MovieHLE || std::get<1>(threedo_nvramPath(name)) ||
      (cmd != threedo_cmdRead && cmd != threedo_cmdStatus))
    return threedo_Machine_fileDeviceIOLegacy(m, name, cmd, offset, sendBuf, sendLen, buf, length);
  auto at = fileEntries.find(name);
  if (at == fileEntries.end()) {
    auto [entry, error] = threedo_Volume_resolve(m->vol, name);
    if (error || entry.IsDir)
      return threedo_Machine_fileDeviceIOLegacy(m, name, cmd, offset, sendBuf, sendLen, buf,
                                                length);
    if (entry.Size < 0 || entry.Size > 0x7fffffff)
      throw std::runtime_error("Invalid file extent");
    at = fileEntries.emplace(name, entry).first;
  }
  const auto &e = at->second;
  if (cmd == threedo_cmdStatus) {
    const uint32_t words[] = {
        0, 40, 2048, uint32_t((e.Size + 2047) / 2048), 0, 0, 0, 0, 0, uint32_t(e.Size)};
    for (unsigned i = 0; i < 10; i++)
      if (i * 4 < length)
        threedo_Machine_write32(m, buf + i * 4, words[i]);
    return {int32_t(std::min(length, 40u)), 0};
  }
  const uint32_t byteOff = offset * 2048u;
  if (!buf || byteOff >= e.Size)
    return {0, 0};
  const uint32_t n = std::min<uint64_t>(length, e.Size - byteOff);
  for (uint32_t done = 0; done < n;) {
    const auto position = uint64_t(byteOff) + done;
    auto [sector, error] = threedo_Volume_block(m->vol, e.Block + position / 2048);
    if (error)
      throw std::runtime_error(error.text);
    const auto within = uint32_t(position % 2048), take = std::min(n - done, 2048 - within);
    uint32_t dest = buf + done;
    uint8_t *p = nullptr;
    if (!m->OnWrite) {
      if (dest < 0x200000 && take <= 0x200000 - dest)
        p = m->dram.p + dest;
      else if (dest >= 0x200000 && dest < 0x300000 && take <= 0x300000 - dest)
        p = m->vram.p + dest - 0x200000;
      else if (dest >= 0x400000 && dest < 0x800000 && take <= 0x800000 - dest)
        p = m->imem.p + dest - 0x400000;
    }
    if (p)
      std::memcpy(p, sector.p + within, take);
    else
      for (uint32_t i = 0; i < take; i++)
        threedo_Machine_Write(m, dest + i, sector[within + i]);
    done += take;
  }
  return {int32_t(n), 0};
}
