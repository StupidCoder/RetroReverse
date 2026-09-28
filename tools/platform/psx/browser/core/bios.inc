void Machine::bios(char table) {
  u32 fn = cpu.r[9] & 255;
  if (table == 'B' && fn == 0x17) { returnISR(); return; }
  if (table == 'A' && (fn == 0x13 || fn == 0x14)) {
    const int regs[] = {31,29,30,16,17,18,19,20,21,22,23,28};
    u32 a=cpu.r[4], result=cpu.r[5];
    for(int i=0;i<12;i++) {
      if(fn==0x13) write32(a+i*4,cpu.r[regs[i]]);
      else cpu.seed(regs[i],read32(a+i*4));
    }
    cpu.seed(2,fn==0x13?0:result); cpu.setPC(cpu.r[31]); return;
  }
  u32 ret = service(table, fn);
  cpu.seed(2, ret);
  cpu.setPC(cpu.r[31]);
}
u32 Machine::service(char table, u32 fn) {
  u32 a = cpu.r[4], b = cpu.r[5], c = cpu.r[6];
  if (table == 'A')
    switch (fn) {
    case 0x2a:
      if (c > 32 * 1024 * 1024)
        throw std::runtime_error("BIOS memcpy budget exceeded");
      for (u32 i = 0; i < c; i++)
        write(a + i, read(b + i));
      return a;
    case 0x28:
      c = b;
      b = 0;
      [[fallthrough]];
    case 0x2b:
      if (c > 32 * 1024 * 1024)
        throw std::runtime_error("BIOS memset budget exceeded");
      for (u32 i = 0; i < c; i++)
        write(a + i, u8(b));
      return fn == 0x28 ? 0 : a;
    case 0x33:
      return alloc(a);
    case 0x34:
      return 0;
    case 0x39:
      heapPtr = a;
      heapEnd = a + b;
      return 0;
    case 0x3f:
      printfGuest(a);
      return 0;
    case 0x44:
    case 0x49:
    case 0x72:
    case 0x96:
    case 0x97:
    case 0x99:
    case 0xa3:
    case 0x9f:
    case 0x70:
      return fn == 0x9f ? a : 0;
    case 0x17:
      c = 0xffffffff;
      [[fallthrough]];
    case 0x18:
      for (u32 i = 0; i < std::min(c, 32u * 1024 * 1024); i++) {
        u8 ca = read(a + i), cb = read(b + i);
        if (ca != cb)
          return u32(i32(ca) - i32(cb));
        if (!ca)
          break;
      }
      return 0;
    case 0x19:
      for (u32 i = 0; i < 32 * 1024 * 1024; i++) {
        u8 v = read(b + i);
        write(a + i, v);
        if (!v)
          return a;
      }
      throw std::runtime_error("BIOS strcpy budget exceeded");
    case 0x1b:
      for (u32 n = 0; n < 32 * 1024 * 1024; n++)
        if (!read(a + n))
          return n;
      throw std::runtime_error("BIOS strlen budget exceeded");
    case 0x2f:
      randSeed = randSeed * 0x41c64e6d + 0x3039;
      return (randSeed >> 16) & 32767;
    case 0x30:
      randSeed = a;
      return 0;
    }
  if (table == 'B')
    switch (fn) {
    case 0x3d:
      appendTTY(std::string(1, char(a)));
      return 0;
    case 0x3f:
      appendTTY(cstr(a));
      return 0;
    case 0:
      return alloc(a);
    case 7:
      return 0;
    case 8:
      return 0xf1000000 | ++nextEvent;
    case 9:
    case 11:
    case 12:
    case 13:
      return 1;
    case 0x12:
      padBuf = a;
      return 1;
    case 0x13:
      padActive = true;
      return 1;
    case 0x14:
    case 0x17:
    case 0x18:
    case 0x47:
    case 0x4a:
    case 0x4b:
    case 0x56:
    case 0x5b:
      return 0;
    case 0x19:
      isrChain = a;
      return 0;
    }
  if (table == 'C')
    switch (fn) {
    case 0:
    case 1:
    case 2:
    case 3:
    case 8:
    case 10:
    case 12:
    case 28:
      return 0;
    }
  char text[64];
  snprintf(text, sizeof text, "Unhandled BIOS %c0(%02X)", table, fn);
  note(text);
  return 0;
}
void Machine::exception() {
  u32 cause = cpu.cop0[13], code = (cause >> 2) & 31, epc = cpu.cop0[14];
  if (!code) {
    if (dispatch(epc))
      return;
    irqStat &= ~irqMask;
    rfeTo(epc);
    return;
  }
  if (code == 8) {
    if (cpu.r[4] == 1) {
      cpu.seed(2, (cpu.cop0[12] >> 2) & 1);
      cpu.cop0[12] &= ~4u;
    } else if (cpu.r[4] == 2)
      cpu.cop0[12] |= 4;
  } else {
    char t[80];
    snprintf(t, sizeof t, "Exception %u at %08X", code, epc);
    note(t);
  }
  rfeTo(epc + ((cause & 0x80000000) ? 8 : 4));
}
bool Machine::dispatch(u32 epc) {
  u32 handler = isrHandler;
  if (!handler && isrChain)
    handler = read32(isrChain);
  if (!handler || !read32(handler))
    return false;
  isr = {true, epc, cpu.r, cpu.hi, cpu.lo};
  if (!isrHandler && isrChain) {
    const int regs[] = {31,29,30,16,17,18,19,20,21,22,23,28};
    for(int i=0;i<12;i++) cpu.seed(regs[i],read32(isrChain+i*4));
    cpu.seed(2,1);
  } else {
    cpu.seed(31,0xe0); cpu.seed(29,0x8000d000); cpu.seed(30,0x8000d000);
  }
  cpu.setPC(handler);
  return true;
}
void Machine::returnISR() {
  if (!isr.active) {
    note("ISR return without active interrupt");
    rfeTo(isr.pc);
    return;
  }
  isr.active = false;
  for (int i = 0; i < 32; i++)
    cpu.seed(i, isr.r[i]);
  cpu.hi = isr.hi;
  cpu.lo = isr.lo;
  rfeTo(isr.pc);
}
void Machine::printfGuest(u32 addr) {
  std::string fmt = cstr(addr), result;
  int argN = 0;
  auto arg = [&]() {
    int n = argN++;
    return n < 3 ? cpu.r[5 + n] : read32(cpu.r[29] + 4 * (n + 1));
  };
  for (size_t i = 0; i < fmt.size(); i++) {
    if (fmt[i] != '%' || i + 1 == fmt.size()) {
      result += fmt[i];
      continue;
    }
    std::string spec = "%";
    i++;
    while (i < fmt.size() && std::string("-+ #0123456789.lh").find(fmt[i]) != std::string::npos) {
      if (fmt[i] != 'l' && fmt[i] != 'h')
        spec += fmt[i];
      i++;
    }
    if (i == fmt.size())
      break;
    char kind = fmt[i];
    spec += kind;
    char tmp[4096] = {};
    if (kind == '%')
      result += '%';
    else if (kind == 's') {
      auto s = cstr(arg());
      snprintf(tmp, sizeof tmp, spec.c_str(), s.c_str());
      result += tmp;
    } else if (kind == 'c')
      result += char(arg());
    else if (kind == 'd' || kind == 'i') {
      snprintf(tmp, sizeof tmp, spec.c_str(), i32(arg()));
      result += tmp;
    } else if (kind == 'u' || kind == 'x' || kind == 'X' || kind == 'o') {
      snprintf(tmp, sizeof tmp, spec.c_str(), arg());
      result += tmp;
    } else
      result += spec;
  }
  appendTTY(result);
}
