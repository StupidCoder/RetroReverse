#pragma once
#include "gte.h"
#include <cstdio>
#include <string>
struct Machine;
// R3000A: delayed branches and loads, fixed-width wraparound, COP0/COP2.
struct CPU {
  std::array<u32, 32> r{}, out{}, cop0{};
  u32 hi = 0, lo = 0, pc = 0xbfc00000, next = 0xbfc00004, cur = 0, ldReg = 0, ldVal = 0,
      branchAddr = 0;
  bool delay = false, pendingDelay = false, halted = false;
  u64 steps = 0;
  std::string error;
  CPU() { cop0[15] = 2; }
  void set(u32 i, u32 v) {
    if (i)
      out[i] = v;
  }
  void seed(u32 i, u32 v) {
    if (i)
      r[i] = out[i] = v;
  }
  void setPC(u32 v) {
    pc = v;
    next = v + 4;
    pendingDelay = false;
  }
  void branch(bool taken, u32 to) {
    pendingDelay = true;
    branchAddr = cur;
    if (taken)
      next = to;
  }
  void load(u32 i, u32 v) {
    ldReg = i;
    ldVal = v;
  }
  void exception(u32 code) {
    u32 epc = cur, cause = code << 2;
    if (delay) {
      epc = branchAddr;
      cause |= 0x80000000;
    }
    cop0[14] = epc;
    cop0[13] = (cop0[13] & 0xff00) | (cause & ~0xff00u);
    u32 sr = cop0[12];
    cop0[12] = (sr & ~63u) | ((sr << 2) & 63);
    setPC(sr & (1 << 22) ? 0xbfc00180 : 0x80000080);
  }
  void addrError(u32 code, u32 addr) {
    cop0[8] = addr;
    exception(code);
  }
  void rfe() {
    u32 sr = cop0[12];
    cop0[12] = (sr & ~15u) | ((sr >> 2) & 15);
  }
  void interrupt(bool pending) {
    if (pending)
      cop0[13] |= 1 << 10;
    else
      cop0[13] &= ~(1 << 10);
    if (!(cop0[12] & 1) || !(cop0[13] & cop0[12] & 0xff00) || pendingDelay)
      return;
    if (ldReg) {
      r[ldReg] = out[ldReg] = ldVal;
      ldReg = ldVal = 0;
    }
    cur = pc;
    delay = false;
    exception(0);
  }
  void halt(u32 w) {
    halted = true;
    char b[100];
    snprintf(b, sizeof b, "Unsupported instruction %08X at %08X", w, cur);
    error = b;
  }
  void step(Machine &);
  void execute(Machine &, u32);
};
