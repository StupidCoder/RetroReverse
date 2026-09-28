u8 CD::read(u32 p) {
  switch (p) {
  case 0:
    return index | (params.empty() ? 8 : 0) | 16 | (!response.empty() ? 32 : 0) |
           (dataPos < data.size() ? 64 : 0);
  case 1:
    if (!response.empty()) {
      u8 v = response.front();
      response.pop_front();
      return v;
    }
    return 0;
  case 2:
    return dataPos < data.size() ? data[dataPos++] : 0;
  case 3:
    return ((index & 1) ? irqFlags : irqEnable) | 0xe0;
  }
  return 0;
}
void CD::write(Machine &m, u32 p, u8 v) {
  switch (p) {
  case 0:
    index = v & 3;
    break;
  case 1:
    if (!index)
      command(m, v);
    break;
  case 2:
    if (!index) {
      if (params.size() >= 256)
        throw std::runtime_error("CD parameter FIFO exceeded capacity");
      params.push_back(v);
    } else if (index == 1)
      irqEnable = v & 31;
    break;
  case 3:
    if (!index) {
      if (v & 128)
        loadData(m);
      else {
        data.clear();
        dataPos = 0;
      }
    } else if (index == 1) {
      irqFlags &= ~(v & 31);
      if (v & 64)
        params.clear();
      if (!irqFlags)
        deliver(m);
    }
    break;
  }
}
void CD::command(Machine &m, u8 cmd) {
  commands++;
  auto fromBCD = [](u8 v) { return (v >> 4) * 10 + (v & 15); };
  auto bcd = [](int n) { return u8((n / 10) * 16 + n % 10); };
  switch (cmd) {
  case 2:
    if (params.size() >= 3)
      loc = std::max(0, (fromBCD(params[0]) * 60 + fromBCD(params[1])) * 75 + fromBCD(params[2]) -
                            150);
    ack();
    break;
  case 6:
  case 0x1b:
    stat |= 32;
    reading = true;
    readLBA = loc;
    ack();
    queueRead();
    break;
  case 9:
    stat &= ~32;
    reading = false;
    ack();
    second();
    break;
  case 10:
    stat = 2;
    mode = 0;
    reading = false;
    ack();
    second();
    break;
  case 14:
    if (!params.empty())
      mode = params[0];
    ack();
    break;
  case 0x15:
  case 0x16:
    stat |= 64;
    ack();
    stat &= ~64;
    second();
    break;
  case 0x13:
    enqueue(1000, 3, {stat, 1, 1});
    break;
  case 0x14: {
    int track = params.empty() ? 0 : fromBCD(params[0]),
        lba = (track == 0 && m.disc) ? m.disc->sectors : 0, t = lba + 150;
    enqueue(1000, 3, {stat, bcd(t / (75 * 60)), bcd((t / 75) % 60)});
    break;
  }
  case 0x19:
    if (!params.empty() && params[0] == 0x20)
      enqueue(1000, 3, {0x94, 9, 0x19, 0xc0});
    else
      ack();
    break;
  default:
    ack();
    break;
  }
  params.clear();
}
void CD::tick(Machine &m) {
  if (queue.empty() || irqFlags)
    return;
  if (queue.front().delay > 0) {
    queue.front().delay--;
    return;
  }
  deliver(m);
}
void CD::deliver(Machine &m) {
  if (queue.empty() || irqFlags || queue.front().delay > 0)
    return;
  auto p = std::move(queue.front());
  queue.pop_front();
  response.assign(p.response.begin(), p.response.end());
  irqFlags = p.cause;
  if (p.read && reading)
    queueRead();
  m.irq(2);
}
void CD::loadData(Machine &m) {
  if (!m.disc || readLBA < 0 || readLBA >= m.disc->sectors)
    return;
  const u8 *b = m.disc->block(readLBA);
  data.assign(b, b + 2048);
  dataPos = 0;
  readLBA++;
}
void CD::dma(Machine &m, u32 addr, u32 bcr) {
  u32 words = bcr & 65535;
  if ((bcr >> 16) > 1)
    words *= bcr >> 16;
  if (words > 4 * 1024 * 1024)
    throw std::runtime_error("CD DMA word budget exceeded");
  for (u32 i = 0; i < words; i++) {
    if (dataPos + 4 > data.size()) {
      if (!reading)
        break;
      loadData(m);
    }
    if (dataPos + 4 > data.size())
      break;
    u32 w = le32(data.data() + dataPos);
    dataPos += 4;
    m.write32(addr, w);
    addr += 4;
  }
}
