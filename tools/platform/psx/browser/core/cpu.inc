// Included after Machine. Memory operations preserve the reference's byte-bus I/O effects.
void CPU::step(Machine &m) {
  if (halted)
    return;
  cur = pc;
  delay = pendingDelay;
  pendingDelay = false;
  u32 w = m.read32(pc);
  pc = next;
  next += 4;
  if (ldReg)
    out[ldReg] = ldVal;
  ldReg = ldVal = 0;
  execute(m, w);
  r = out;
  steps++;
}
void CPU::execute(Machine &m, u32 w) {
  u32 op = w >> 26, rs = (w >> 21) & 31, rt = (w >> 16) & 31, rd = (w >> 11) & 31,
      sh = (w >> 6) & 31, imm = w & 65535, simm = u32(i32(i16(imm)));
  u32 a = r[rs], b = r[rt], addr = a + simm, bt = cur + 4 + (simm << 2),
      jt = ((cur + 4) & 0xf0000000) | ((w & 0x3ffffff) << 2);
  auto add = [&](u32 dest, u32 x, u32 y) {
    u32 z = x + y;
    if ((x ^ z) & (y ^ z) & 0x80000000)
      exception(12);
    else
      set(dest, z);
  };
  auto sub = [&](u32 dest, u32 x, u32 y) {
    u32 z = x - y;
    if ((x ^ y) & (x ^ z) & 0x80000000)
      exception(12);
    else
      set(dest, z);
  };
  switch (op) {
  case 0:
    switch (w & 63) {
    case 0:
      set(rd, b << sh);
      break;
    case 2:
      set(rd, b >> sh);
      break;
    case 3:
      set(rd, u32(i32(b) >> sh));
      break;
    case 4:
      set(rd, b << (a & 31));
      break;
    case 6:
      set(rd, b >> (a & 31));
      break;
    case 7:
      set(rd, u32(i32(b) >> (a & 31)));
      break;
    case 8:
      branch(true, a);
      break;
    case 9:
      set(rd, cur + 8);
      branch(true, a);
      break;
    case 12:
      exception(8);
      break;
    case 13:
      exception(9);
      break;
    case 16:
      set(rd, hi);
      break;
    case 17:
      hi = a;
      break;
    case 18:
      set(rd, lo);
      break;
    case 19:
      lo = a;
      break;
    case 24: {
      i64 p = i64(i32(a)) * i32(b);
      lo = u32(p);
      hi = u32(u64(p) >> 32);
      break;
    }
    case 25: {
      u64 p = u64(a) * b;
      lo = u32(p);
      hi = u32(p >> 32);
      break;
    }
    case 26:
      if (!b) {
        hi = a;
        lo = i32(a) >= 0 ? 0xffffffff : 1;
      } else if (a == 0x80000000 && b == 0xffffffff) {
        lo = a;
        hi = 0;
      } else {
        lo = u32(i32(a) / i32(b));
        hi = u32(i32(a) % i32(b));
      }
      break;
    case 27:
      lo = b ? a / b : 0xffffffff;
      hi = b ? a % b : a;
      break;
    case 32:
      add(rd, a, b);
      break;
    case 33:
      set(rd, a + b);
      break;
    case 34:
      sub(rd, a, b);
      break;
    case 35:
      set(rd, a - b);
      break;
    case 36:
      set(rd, a & b);
      break;
    case 37:
      set(rd, a | b);
      break;
    case 38:
      set(rd, a ^ b);
      break;
    case 39:
      set(rd, ~(a | b));
      break;
    case 42:
      set(rd, i32(a) < i32(b));
      break;
    case 43:
      set(rd, a < b);
      break;
    default:
      halt(w);
      break;
    }
    break;
  case 1:
    if ((rt & 0x1e) == 0x10)
      set(31, cur + 8);
    branch((rt & 1) ? i32(a) >= 0 : i32(a) < 0, bt);
    break;
  case 2:
    branch(true, jt);
    break;
  case 3:
    set(31, cur + 8);
    branch(true, jt);
    break;
  case 4:
    branch(a == b, bt);
    break;
  case 5:
    branch(a != b, bt);
    break;
  case 6:
    branch(i32(a) <= 0, bt);
    break;
  case 7:
    branch(i32(a) > 0, bt);
    break;
  case 8:
    add(rt, a, simm);
    break;
  case 9:
    set(rt, a + simm);
    break;
  case 10:
    set(rt, i32(a) < i32(simm));
    break;
  case 11:
    set(rt, a < simm);
    break;
  case 12:
    set(rt, a & imm);
    break;
  case 13:
    set(rt, a | imm);
    break;
  case 14:
    set(rt, a ^ imm);
    break;
  case 15:
    set(rt, imm << 16);
    break;
  case 16:
    if (w & (1 << 25)) {
      if ((w & 63) == 16)
        rfe();
      else
        halt(w);
    } else if (rs == 0)
      load(rt, cop0[rd]);
    else if (rs == 4)
      cop0[rd] = b;
    else
      halt(w);
    break;
  case 18:
    if (w & (1 << 25))
      m.gte.command(w & 0x1ffffff);
    else
      switch (rs) {
      case 0:
        load(rt, m.gte.read(rd));
        break;
      case 2:
        load(rt, m.gte.ctrl[rd]);
        break;
      case 4:
        m.gte.write(rd, b);
        break;
      case 6:
        m.gte.ctrl[rd] = b;
        break;
      default:
        halt(w);
      }
    break;
  case 32:
    load(rt, u32(i32(i8(m.read(addr)))));
    break;
  case 36:
    load(rt, m.read(addr));
    break;
  case 33:
  case 37:
    if (addr & 1)
      addrError(4, addr);
    else {
      u32 v = m.read16(addr);
      load(rt, op == 33 ? u32(i32(i16(v))) : v);
    }
    break;
  case 35:
    if (addr & 3)
      addrError(4, addr);
    else
      load(rt, m.read32(addr));
    break;
  case 34: {
    u32 v = m.read32(addr & ~3u), s = (addr & 3) * 8;
    load(rt, (out[rt] & (0x00ffffff >> s)) | (v << (24 - s)));
    break;
  }
  case 38: {
    u32 v = m.read32(addr & ~3u), s = (addr & 3) * 8;
    load(rt, (out[rt] & (0xffffff00u << (24 - s))) | (v >> s));
    break;
  }
  case 40:
    m.write(addr, u8(b));
    break;
  case 41:
    if (addr & 1)
      addrError(5, addr);
    else
      m.write16(addr, b);
    break;
  case 43:
    if (addr & 3)
      addrError(5, addr);
    else
      m.write32(addr, b);
    break;
  case 42: {
    u32 v = m.read32(addr & ~3u), s = (addr & 3) * 8;
    m.write32(addr & ~3u, (v & (0xffffff00u << s)) | (b >> (24 - s)));
    break;
  }
  case 46: {
    u32 v = m.read32(addr & ~3u), s = (addr & 3) * 8;
    m.write32(addr & ~3u, (v & (0x00ffffff >> (24 - s))) | (b << s));
    break;
  }
  case 50:
    if (addr & 3)
      addrError(4, addr);
    else
      m.gte.write(rt, m.read32(addr));
    break;
  case 58:
    if (addr & 3)
      addrError(5, addr);
    else
      m.write32(addr, m.gte.read(rt));
    break;
  default:
    halt(w);
    break;
  }
}
