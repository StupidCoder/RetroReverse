// Stunt Car Racer — the car physics, hand-ported from the verified Go reimplementation
// (Stunt Car Racer (Amiga)/extract/physics, Part V) one routine at a time. The Go is
// checked coordinate-exact against the original 68000 on the m68k core; this JS is kept
// exact by self-checking against a Go-generated golden trace (selfTest below). It runs
// directly on the game's memory image (the $1B000..$30000 slice the physics reads, loaded
// from <id>.bin, plus a few baked code-region constants), exactly as the original does.
//
// Fixed-point: i16/u16/i8 wrap to width; 16x16 products use Math.imul; >> is arithmetic,
// >>> logical, exactly mirroring the m68k word/long ops.

const i16 = (x) => (x << 16) >> 16;
const u16 = (x) => x & 0xFFFF;
const i8 = (x) => (x << 24) >> 24;
const ror16 = (v, n) => { v &= 0xFFFF; return ((v >>> n) | (v << (16 - n))) & 0xFFFF; };

// addresses (same as package physics)
const A = {
  PosX: 0x1BCD8, PosY: 0x1BCDC, PosZ: 0x1BCE0,
  Roll: 0x1BCE4, Yaw: 0x1BCE6, Pit: 0x1BCE8,
  VelX: 0x1BCEA, VelY: 0x1BCEC, VelZ: 0x1BCEE,
  AmR: 0x1BCF0, AmP: 0x1BCF2, AmY: 0x1BCF4,
  FrcX: 0x1BCF6, FrcY: 0x1BCF8, FrcZ: 0x1BCFA,
  TqR: 0x1BCFC, TqP: 0x1BCFE, TqY: 0x1BD00,
  WAmR: 0x1BD3A, WAmY: 0x1BD3C, WAmP: 0x1BD3E,
  Mtx: 0x1C230, Tmpl: 0x1EC46, Hdg: 0x1BD5A, angLimits: 0x61AD4,
  BVelL: 0x1BD2C, BVelM: 0x1BD2E, BVelV: 0x1BD30,
  GrvA: 0x1BD0E, GrvB: 0x1BD10, GrvC: 0x1BD12,
  BFrcA: 0x1BD32, BFrcB: 0x1BD34, BFrcC: 0x1BD36,
  Rest: 0x1BCA0, NetLift: 0x1BD38, RollTq: 0x1BD28, PitchTq: 0x1BD26,
  OnGround: 0x1BB7E, Bottom: 0x1BB7D, DmgEvt: 0x1BB54,
  Drive: 0x1BD2A, LoadA: 0x1BD40, LoadB: 0x1BD42, LoadC: 0x1BD44, Slip: 0x1BBC1,
  TqAppR: 0x1BCFC, TqAppY: 0x1BD00,
};
const DAMP = 0xEE, GRAV = 0x13D, DMGLIMIT = 0x63CE2;
const MAGIC = 0x9CEDCD02; // genuine-disk protection value

export class Physics {
  constructor() { this.B = new Uint8Array(0x66000); } // +scratch room for $65EC0 ($5FF94)

  // load the per-track init ($1B000..$1C800) and the shared static data ($1C800..$30000),
  // and bake the code-region constants the physics reads.
  loadTrack(perTrackBuf, staticBuf) {
    this.B.fill(0);
    this.B.set(new Uint8Array(staticBuf), 0x1C900);
    const u = new Uint8Array(perTrackBuf);
    const split = 0x1C900 - 0x1B000;
    this.B.set(u.subarray(0, split), 0x1B000);
    this.B.set(u.subarray(split), 0x1CA1A); // control bytes (override static's track-0 values)
    this.B.set([0, 0, 0, 217, 255, 39], 0x6125A);            // $6125A table
    this.B.set([44, 0, 10, 0, 211, 0, 245, 0, 48, 57, 0, 1], 0x61AD4); // $61AD4 limits
    this.B.set([0x9C, 0xED, 0xCD, 0x02], 0x64AEC);           // protection: genuine disk
    this.B.set([0x00, 0xD4, 0x80, 0xD4, 0x00, 0x00, 0xAB, 0xAB, 0x40, 0x40, 0x00, 0x00], 0x5C6B8); // $5C6B8 radius coefs
  }

  // --- memory access (big-endian, like the 68000) ---
  u8(a) { return this.B[a]; }
  w(a) { return i16((this.B[a] << 8) | this.B[a + 1]); }
  setW(a, v) { this.B[a] = (v >> 8) & 0xFF; this.B[a + 1] = v & 0xFF; }
  l(a) { return ((this.B[a] << 24) | (this.B[a + 1] << 16) | (this.B[a + 2] << 8) | this.B[a + 3]) | 0; }
  setL(a, v) { this.B[a] = (v >>> 24) & 0xFF; this.B[a + 1] = (v >>> 16) & 0xFF; this.B[a + 2] = (v >>> 8) & 0xFF; this.B[a + 3] = v & 0xFF; }

  mul0_93(v) { return i16((Math.imul(i16(v), DAMP)) >> 8); }

  // --- trig: $64D08 is SINE (selector 0), $64D10 is COSINE (selector $4000) ---
  sin(a) { return this.sinSel(a, 0x0000); }
  cos(a) { return this.sinSel(a, 0x4000); }
  sinSel(a, d5) {
    let d0 = a & 0xFFFF;
    let d3 = d0 & 0x3FFF;
    if (((d0 & 0x4000) ^ d5) === 0) d3 = ((d3 ^ 0x3FFF) + 1) & 0xFFFF;
    d3 = ror16(d3, 5);
    const d4 = d3 & 0x3FE;
    const s0 = this.w(0x1CA42 + d4), s1 = this.w(0x1CA42 + d4 + 2);
    const d6 = (s0 - s1) & 0xFFFF;
    const frac = ror16(d3, 1) & 0xFC00;
    const hi = ((frac * d6) >>> 16) & 0xFFFF;
    let d7 = i16((u16(s0) - hi) & 0xFFFF);
    d7 = i16(u16(d7) >>> 1);
    const sg = (d0 & d5) << 1;
    if (i16((d0 ^ sg) & 0xFFFF) < 0) d7 = -d7;
    return i16(d7);
  }

  // value * matrix[$1C230+idx*2] >> 15 ($61344)
  mtxMul(value, idx) { return i16((Math.imul(i16(value), this.w(A.Mtx + idx * 2)) << 1) >> 16); }

  // --- integrator ---
  force61ADC() {
    this.setW(A.VelX, i16(this.w(A.VelX) + this.mul0_93(this.w(A.FrcX))));
    this.setW(A.VelY, i16(this.w(A.VelY) + this.mul0_93(this.w(A.FrcY))));
    this.setW(A.VelZ, i16(this.w(A.VelZ) + this.mul0_93(this.w(A.FrcZ))));
  }
  torque61B26() {
    this.setW(A.AmR, i16(this.w(A.AmR) + this.mul0_93(this.w(A.TqR))));
    this.setW(A.AmP, i16(this.w(A.AmP) + this.mul0_93(this.w(A.TqP))));
    this.setW(A.AmY, i16(this.w(A.AmY) + this.mul0_93(this.w(A.TqY))));
  }
  integrate61950() {
    this.setL(A.PosX, (this.l(A.PosX) + (this.mul0_93(this.w(A.VelX)) << 6)) | 0);
    this.setL(A.PosY, (this.l(A.PosY) + (this.mul0_93(this.w(A.VelY)) << 7)) | 0);
    this.setL(A.PosZ, (this.l(A.PosZ) + (this.mul0_93(this.w(A.VelZ)) << 6)) | 0);
    if (this.w(A.PosY) >= 0x3E8) this.setW(A.PosY, 0x3E8);
    this.setW(A.Roll, i16(this.w(A.Roll) + this.mul0_93(this.w(A.WAmR))));
    this.clamp619E4(this.mul0_93(this.w(A.WAmY)));
  }
  clamp619E4(d0) {
    this.setW(A.Yaw, i16(this.w(A.Yaw) + d0));
    this.setW(A.Pit, i16(this.w(A.Pit) + this.mul0_93(this.w(A.WAmP))));
    let d2 = 0;
    if (i8(this.u8(0x1BB75)) < 0 && this.u8(0x1BB9A) === 0xE0) d2 = 2;
    this.clampAngle(A.Roll, A.AmR, A.angLimits, d2);
    this.clampAngle(A.Pit, A.AmY, A.angLimits, d2);
  }
  clampAngle(ang, mom, a0, d2) {
    const d3 = this.w(ang);
    let lim;
    if (d3 >= 0) { lim = this.w(a0 + d2); if (u16(lim) >= u16(d3)) return; }
    else { lim = this.w(a0 + d2 + 4); if (u16(lim) < u16(d3)) return; }
    this.setW(ang, lim);
    if ((i16(lim ^ this.w(mom))) >= 0) this.setW(mom, 0);
  }

  // --- orientation matrix + transforms ---
  mt(off) { return this.w(A.Mtx + off); }
  smt(off, v) { this.setW(A.Mtx + off, v); }
  prod(off, d5) { this.smt(off, i16((Math.imul(this.mt(off), d5) << 1) >> 16)); }
  matrix61368() {
    const sy = this.sin(this.w(A.Yaw));
    for (const o of [0x4, 0xC, 0xE, 0x14, 0x16]) this.smt(o, sy);
    const cy = this.cos(this.w(A.Yaw));
    for (const o of [0x6, 0x10, 0x12, 0x18, 0x1A]) this.smt(o, cy);
    const yh = i16(this.w(A.Yaw) - this.w(A.Hdg));
    const sh = this.sin(yh); for (const o of [0x34, 0x42, 0x44]) this.smt(o, sh);
    const ch = this.cos(yh); for (const o of [0x38, 0x3E, 0x46]) this.smt(o, ch);
    this.smt(0x8, this.sin(this.w(A.Roll)));
    const cr = this.cos(this.w(A.Roll)); for (const o of [0xA, 0x1C, 0x1E]) this.smt(o, cr);
    this.smt(0x22, this.cos(this.w(A.Pit)));
    this.smt(0x20, this.sin(this.w(A.Pit)));
    let d5 = this.mt(0x8);
    for (let o = 0xC; o <= 0x12; o += 2) this.prod(o, d5);
    for (let o = 0x34; o <= 0x38; o += 4) this.prod(o, d5);
    this.smt(0x0, this.mt(0xC)); this.smt(0x2, this.mt(0x10));
    d5 = this.mt(0xA);
    for (let o = 0x4; o <= 0x6; o += 2) this.prod(o, d5);
    for (let o = 0x44; o <= 0x46; o += 2) this.prod(o, d5);
    d5 = this.mt(0x20);
    for (let o = 0xC; o <= 0x1C; o += 4) this.prod(o, d5);
    for (let o = 0x34; o <= 0x38; o += 4) this.prod(o, d5);
    d5 = this.mt(0x22);
    for (let o = 0xE; o <= 0x1E; o += 4) this.prod(o, d5);
    for (let o = 0x3E; o <= 0x42; o += 4) this.prod(o, d5);
    this.smt(0x28, i16(this.mt(0x18) - this.mt(0xE)));
    this.smt(0x2A, i16(-this.mt(0x12) - this.mt(0x14)));
    this.smt(0x2C, i16(this.mt(0x1A) + this.mt(0xC)));
    this.smt(0x2E, i16(this.mt(0x10) - this.mt(0x16)));
    this.smt(0x30, i16(-this.mt(0x1C)));
    this.smt(0x24, i16(-this.mt(0x20)));
  }
  tmpl(off) { return this.u8(A.Tmpl + off); }
  velToBody6158C() {
    for (let d2 = 2; ; d2 -= 2) {
      let d5 = 0;
      d5 = i16(d5 + this.mtxMul(this.w(A.VelX), this.tmpl(d2 + 0)));
      d5 = i16(d5 + this.mtxMul(this.w(A.VelY), this.tmpl(d2 + 3)));
      d5 = i16(d5 + this.mtxMul(this.w(A.VelZ), this.tmpl(d2 + 6)));
      this.setW(A.BVelL + (d2 << 1), d5);
      if (d2 === 0) break;
    }
  }
  gravToBody615E6() {
    this.setW(A.GrvB, this.mtxMul(-GRAV, 0xF));
    this.setW(A.GrvC, this.mtxMul(-GRAV, 0x4));
    this.setW(A.GrvA, this.mtxMul(GRAV, 0xE));
  }
  forceToWorld61618() {
    for (let d2 = 2; ; d2 -= 1) {
      let d5 = 0;
      d5 = i16(d5 + this.mtxMul(this.w(A.BFrcA), this.tmpl(d2 + 0x9)));
      d5 = i16(d5 + this.mtxMul(this.w(A.BFrcB), this.tmpl(d2 + 0xC)));
      d5 = i16(d5 + this.mtxMul(this.w(A.BFrcC), this.tmpl(d2 + 0xF)));
      this.setW(A.FrcX + (d2 << 1), d5);
      if (d2 === 0) break;
    }
  }
  torqueToWorld61672() {
    for (let d2 = 1; ; d2 -= 1) {
      let d5 = 0;
      d5 = i16(d5 + this.mtxMul(this.w(A.AmR), this.tmpl(d2 + 0x12)));
      d5 = i16(d5 + this.mtxMul(this.w(A.AmP), this.tmpl(d2 + 0x14)));
      this.setW(A.WAmR + (d2 << 1), d5);
      if (d2 === 0) break;
    }
    this.setW(A.WAmP, i16(this.mtxMul(this.w(A.WAmY), 0x4) + this.w(A.AmY)));
  }

  // --- track-surface sample ---
  interp5C554() {
    const along = this.u8(0x1BC4D);
    const left = (Math.imul(i16(this.w(0x1BC04) - this.w(0x1BC02)), along) + (this.w(0x1BC02) << 8)) | 0;
    let d0 = (Math.imul(i16(this.w(0x1BC08) - this.w(0x1BC06)), along) + (this.w(0x1BC06) << 8)) | 0;
    const across = this.u8(0x1BC41);
    d0 = (d0 - left) | 0;
    let d4 = d0 < 0 ? -d0 : d0;
    const big = (d4 >>> 0) >= 0x8000;
    if (big) d0 = d0 >> 3;
    let p;
    if (i16(d0) < 0) { const wv = u16(-i16(d0)); p = (Math.imul(wv, across) & ~0xFF) | 0; p = (-p) | 0; }
    else { p = Math.imul(u16(i16(d0)), across) | 0; }
    if (big) p = p << 3;
    p = p >> 8;
    p = (p + left) | 0;
    this.setL(0x1BB18, p);
  }
  corners618CE() {
    let d4 = i16((this.w(0x1C26E) >> 1) - (this.w(0x1C264) >> 1));
    let d5 = i16((this.w(0x1C268) >> 1) - (this.w(0x1C272) >> 1));
    const d0 = this.w(0x1C274) >> 5, d3 = this.w(0x1C276) >> 5;
    d4 = d4 >> 5; d5 = d5 >> 5;
    this.setW(0x1BD06, i16(-d0)); this.setW(0x1BD0C, i16(-d3));
    this.setW(0x1BD02, i16(d0 - d4)); this.setW(0x1BD04, i16(d0 + d4));
    this.setW(0x1BD08, i16(d3 - d5)); this.setW(0x1BD0A, i16(d3 + d5));
  }

  // --- suspension ---
  spring6180E(delta, travel) { return i16(u16((Math.imul(i16(delta), 0x114) >> 8)) + u16(travel)); }
  spring(surf, car, comp, travel, prev, force, dmg) {
    let d0 = (this.l(surf) - this.l(car) - this.l(A.Rest)) | 0;
    this.setL(comp, d0);
    if (d0 >= 0) { if (d0 >= 0x1400) d0 = 0x1400; } else if (d0 < -0x300) d0 = -0x300;
    this.setW(travel, i16(d0));
    const d6 = i16(d0);
    const f = this.spring6180E(i16(i16(d0) - this.w(prev)), d6);
    if (f < 0) { this.setW(force, 0); this.B[0x1BB56] = 0; this.setW(prev, this.w(travel)); return; }
    const d4 = this.w(force);
    this.setW(force, f);
    if (f >= 0x400 && d4 < 0x200) this.B[A.Bottom] = (this.B[A.Bottom] + 1) & 0xFF;
    let d = i16(f - (this.u8(0x1BB01) << 8));
    if (d < 0 || d < 0x700) { this.B[0x1BB56] = 0; this.setW(prev, this.w(travel)); return; }
    if (d >= this.w(0x1BC3A)) this.setW(0x1BC3A, d);
    d = i16(d - 0x600);
    if (i8(this.u8(0x1BBCD)) >= 0) {
      this.B[0x1BB56] = (this.u8(0x1BB56) + 1) & 0xFF;
      if (i8(this.u8(0x1BB56)) < i8(this.u8(DMGLIMIT))) {
        let sev = u16(d) >>> 8;
        sev = (sev + (sev >> 1)) & 0xFF;
        let n = sev + this.u8(dmg); if (n > 0xFF) n = 0xFF;
        this.B[dmg] = n & 0xFF; this.B[A.DmgEvt] = 0x80;
      }
    }
    if (this.w(force) >= 0x1200) this.setW(force, 0x11FF);
    this.setW(prev, this.w(travel));
  }
  contactHeights61B70() {
    const sr = this.sin(this.w(A.Roll));
    this.setW(0x1BBF6, sr);
    const d0 = (this.sin(this.w(A.Pit)) << 3) | 0;
    const d3 = (sr << 4) | 0;
    this.setL(0x1BC9C, (this.l(A.PosY) - d3) >> 8);
    const d4 = (this.l(A.PosY) + d3) | 0;
    this.setL(0x1BC98, (d4 - d0) >> 8);
    this.setL(0x1BC94, (d4 + d0) >> 8);
  }
  suspension61BCC() {
    this.B[A.Bottom] = 0; this.B[0x1BC3A] = 0;
    this.spring(0x1BCA4, 0x1BC94, 0x1BCB0, 0x1BD14, 0x1BD1A, 0x1BD20, 0x1BB4F);
    this.spring(0x1BCA8, 0x1BC98, 0x1BCB4, 0x1BD16, 0x1BD1C, 0x1BD22, 0x1BB50);
    this.spring(0x1BCAC, 0x1BC9C, 0x1BCB8, 0x1BD18, 0x1BD1E, 0x1BD24, 0x1BB51);
    let d0 = i16((this.w(0x1BD20) + this.w(0x1BD22)) >> 1);
    this.setW(0x1BBF6, d0);
    d0 = i16((d0 + this.w(0x1BD24)) >> 1);
    this.setW(A.NetLift, d0);
    let dd = i16(this.w(0x1BD20) - this.w(0x1BD22));
    let t = i16((dd << 1) + dd);
    if (t < 0) t = -t; if (t >= 0x1000) t = 0x1000; if (dd < 0) t = -t;
    this.setW(A.RollTq, i16(t));
    this.setW(A.PitchTq, i16(this.w(0x1BBF6) - this.w(0x1BD24)));
    const og = this.u8(A.NetLift) | this.u8(A.NetLift + 1);
    this.B[A.OnGround] = og & 0xFF;
    if (og !== 0) return;
    if (this.u8(0x1BBDF) !== 0) return;
    let d3 = i16(-0x80);
    const roll = this.w(A.Roll);
    if (roll >= 0) { if (roll >= 0x1000) d3 = i16(-0x100); }
    else {
      switch (this.u8(0x1CA33)) {
        case 7: d3 = i16(-0x80); break;
        case 4: d3 = i16(-0x8); break;
        default: return;
      }
    }
    d3 = i16(d3 - this.w(A.PitchTq));
    if (d3 >= 0) return;
    const c = this.u8(0x1BCF0);
    if (i8(c) >= 0 || c === 0xFF) this.setW(A.PitchTq, d3);
  }

  // --- drive / tyre ---
  grip621DA() { return this.u8(A.OnGround) === 0 ? 0 : i16(this.w(A.LoadB) << 1); }
  lateralTire6217A() {
    const d4 = i16(this.w(A.GrvA) + this.w(A.LoadA));
    let d3 = i16(d4 - this.w(A.BVelL)); if (d3 < 0) d3 = -d3;
    const g = this.grip621DA();
    if (u16(d3) < u16(g)) { this.setW(A.BFrcA, i16(this.w(A.LoadA) - this.w(A.BVelL))); this.B[A.Slip] = 0; return; }
    let gg = g; if (this.w(A.BVelL) < 0) gg = -g;
    this.setW(A.BFrcA, i16(d4 - gg)); this.B[A.Slip] = 0x80;
  }
  drive620B8() {
    this.setW(A.BFrcB, i16(this.w(A.GrvB) + this.w(A.LoadB)));
    const d0b = this.u8(A.Drive) | this.u8(A.BVelV);
    if (i8(d0b) >= 0 && this.u8(0x1BD2B) !== 0) this.setW(A.Drive, i16(this.w(A.Drive) - (d0b & 0xFF)));
    let d3 = this.w(A.Drive); if (d3 < 0) d3 = -d3;
    const g = this.grip621DA();
    if (u16(d3) >= u16(g)) { this.setW(A.Drive, this.w(A.Drive) >= 0 ? g : i16(-g)); }
    this.setW(A.BFrcC, i16(this.w(A.Drive) + this.w(A.LoadC) + this.w(A.GrvC)));
    this.lateralTire6217A();
  }
  torqueApply62138() {
    let d0 = i16(this.w(A.PitchTq) - (this.w(A.AmR) >> 4));
    if (this.u8(A.OnGround) !== 0) d0 = i16(d0 + (this.w(A.BFrcC) >> 2));
    this.setW(A.TqAppR, d0);
    this.setW(A.TqAppY, i16(this.w(A.RollTq) - (this.w(A.AmY) >> 4)));
  }
  drag621F4() {
    const absw = (v) => v < 0 ? i16(-v) : v;
    let d7 = 1, d0 = 0x6000, handled = false;
    if (this.u8(A.OnGround) !== 0) {
      let s = this.u8(0x1BD46); if (i8(s) < 0) s ^= 0xFF;
      if ((s & 0xFF) >= 3 || i8(this.u8(0x1BB9C)) < 0) handled = true;
      else if (this.u8(0x1BCA2) !== 0) { d7 = 3; handled = true; }
    }
    let low = false;
    if (!handled) { if (this.u8(0x1BBDF) === 0) low = true; else d7 = 3; }
    if (low) {
      d0 = absw(this.w(0x1BD2C));
      let v = absw(this.w(0x1BD2E)); if (v > d0) d0 = v;
      v = absw(this.w(0x1BD30)); if (v > d0) d0 = v;
      d7 = 5;
      if (i8(this.u8(0x1BBC7)) < 0 && i8(this.u8(0x1BBB8)) >= 0) {
        if (u16(d0) >= 0xA00) d0 = i16(u16(d0) - 0xA00); else d0 = 0;
      }
    }
    const apply = (va, fa) => {
      const hi = i16((Math.imul(this.w(va), i16(d0)) >> 16)) >> d7;
      this.setW(fa, i16(this.w(fa) - hi));
    };
    apply(A.VelX, A.FrcX); apply(A.VelY, A.FrcY); apply(A.VelZ, A.FrcZ);
  }

  // --- load projection ($622DC) ---
  slope62424(d0w) {
    let d0 = i16(d0w); if (d0 < 0) d0 = -d0;
    let d1 = 0xFF; if (d0 < 0x100) d1 = d0 & 0xFF;
    this.B[0x1BB2B] = d1;
    this.B[0x1BB2D] = this.B[0x1EECA + (d1 >> 1)];
  }
  frac612A2(d0) {
    const d3 = ((this.u8(0x1BB1A) & 0xFF) * (d0 & 0xFF)) | 0;
    this.B[0x1BB1B] = d3 & 0xFF;
    return (d3 >> 8) & 0xFFFF;
  }
  loadProject622DC() {
    this.setW(0x1BD4A, 0);
    let d0 = ((((this.l(0x1BCB0) + this.l(0x1BCB4)) >> 1) - this.l(0x1BCB8)) | 0) >> 4;
    this.setW(0x1BD4C, i16(u16(i16(d0)) ^ 0x8000));
    this.slope62424(i16(d0));
    this.B[0x1BB2C] = this.B[0x1BB2D];
    this.B[0x1BD52] = this.B[0x1BB2B];
    d0 = (this.l(0x1BCB0) - this.l(0x1BCB4)) >> 3;
    this.setW(0x1BD48, i16(d0));
    this.slope62424(i16(d0));
    this.B[0x1BB1A] = this.B[0x1BB2C];
    this.B[0x1BD50] = this.frac612A2(this.u8(0x1BB2D)) & 0xFF;
    this.B[0x1BD4E] = this.frac612A2(this.u8(0x1BB2B)) & 0xFF;
    const proj = () => {
      let d3 = this.u8(0x1BB1A) & 0xFF;
      if (i8(this.u8(0x1BBBB)) < 0) d3 = (0 - d3) & 0xFFFF;
      d3 = (d3 << 7) & 0xFFFF;
      return i16((Math.imul(this.w(A.NetLift), i16(d3)) << 1) >> 16);
    };
    this.B[0x1BB1A] = this.B[0x1BD4E]; this.B[0x1BBBB] = this.B[0x1BD48]; this.setW(A.LoadA, proj());
    this.B[0x1BB1A] = this.B[0x1BD50]; this.B[0x1BBBB] = this.B[0x1BD4A]; this.setW(A.LoadB, proj());
    this.B[0x1BB1A] = this.B[0x1BD52]; this.B[0x1BBBB] = this.B[0x1BD4C]; this.setW(A.LoadC, proj());
  }

  // --- sound + tail housekeeping ---
  sound60FBE() {
    let d0 = this.w(0x1BD30); if (d0 < 0) d0 = -d0;
    this.setW(0x1BD5C, d0);
    if (this.u8(A.OnGround) === 0) { this.setW(0x1BC62, i16(this.w(0x1BC62) - i16(u16(this.w(0x1BC62)) >>> 2))); return; }
    if (d0 >= 0x800) {
      let s = (u16(d0) << 1) + 0x3000;
      if (s > 0xFFFF) s = 0xFF00;
      this.setW(0x1BC62, i16(u16(s)));
    } else this.setW(0x1BC62, i16(d0 << 3));
  }
  tail63E2E() {
    if (this.u8(0x63EE0) !== 0) this.B[0x63EE0] = (this.B[0x63EE0] - 1) & 0xFF;
    if (this.u8(0x1BB46) === 0) return;
    this.B[0x1BB46] = 0;
    let d0 = this.w(0x1BBEE) - this.w(0x1BD58); if (d0 < 0) d0 = 0;
    this.setW(0x1BBEE, i16(d0));
    const imp = this.w(0x1BD56) >> 4;
    this.setW(0x1BD76, i16(this.w(0x1BD76) - imp));
    this.setW(0x1BD78, i16(this.w(0x1BD78) - imp));
    this.setW(0x1BD7A, i16(this.w(0x1BD7A) - imp));
    this.setW(0x1BD40, i16(this.w(0x1BD40) + this.w(0x1BD54)));
    this.setW(0x1BD42, i16(this.w(0x1BD42) + this.w(0x1BD56)));
    this.setW(0x1BD44, i16(this.w(0x1BD44) + this.w(0x1BD58)));
    this.setW(0x1BD54, 0); this.setW(0x1BD56, 0); this.setW(0x1BD58, 0);
    if (this.u8(0x63EE0) === 0) this.B[0x63EE0] = 5;
  }

  // --- track helpers (ports of physics/track.go) ---
  handlePhys(wv) { return ((((wv << 8 | wv >>> 8) & 0xFFFF) - 0xB100) & 0xFFFF) + 0x1EF82; }
  setup5FE56(sec) {
    const p2 = this.u8(0x1C4C0 + sec);
    this.B[0x1BB79] = p2;
    this.setW(0x1BC8C, this.w(0x1EFA2 + ((p2 << 1) & 0xFF)));
    const attr = this.u8(0x1C524 + sec);
    const d2 = (attr << 1) & 0xFF;
    this.B[0x1BBDC] = (((attr >> 7) & 1) << 1) & 0xFF;
    this.setW(0x1BC90, this.w(0x1EFA2 + d2));
    this.setW(0x1BC0E, this.w(0x1C650 + sec * 2));
    this.setW(0x1BC10, this.w(0x1C718 + sec * 2));
    const typ = this.u8(0x1C5EC + sec);
    this.B[0x1BC4A] = typ & 0xC0;
    this.B[0x1BC32] = ((typ & 0x10) << 3) & 0xFF;
    const nib = typ & 0x0F;
    this.B[0x1BB86] = nib;
    this.setW(0x1BCBC, this.w(0x1EF82 + nib * 2));
    const a0 = this.handlePhys(u16(this.w(0x1BCBC)));
    this.B[0x1BB4D] = this.u8(a0 + 1);
    const off = this.u8(a0);
    const cnt = this.u8(a0 + off);
    let d2i = off + 1;
    this.B[0x1BB97] = cnt;
    this.B[0x1BB59] = (cnt << 1) & 0xFF;
    const cm2 = (cnt - 2) & 0xFF;
    this.B[0x1BB98] = cm2;
    this.B[0x1BB5A] = (cm2 << 1) & 0xFF;
    this.B[0x1BB6A] = (((cnt >> 1) - 1)) & 0xFF;
    let v = this.u8(a0 + d2i); d2i++;
    v = ((v >> 1) | ((v & 1) << 7)) & 0xFF;
    this.B[0x1BC44] = v & 0x80;
    this.B[0x1BB7B] = this.u8(a0 + d2i); d2i++;
    this.B[0x1BBD9] = this.u8(a0 + d2i); d2i += 3;
    this.B[0x1BBD4] = this.u8(a0 + d2i);
  }
  railHeight5C0AA(a4, a5, d1) {
    const p2 = this.u8(0x1BB79);
    const b650 = this.w(0x1BC0E), b718 = this.w(0x1BC10);
    let d2 = d1;
    if (i8(p2) >= 0) {
      const odd = d2 & 1; d2 >>= 1;
      let d0;
      if (odd) { const v = this.u8(a5 + d2); d0 = (((v << 1) & 0xE0) | ((v & 0xF) << 8)) + b718; }
      else { const v = this.u8(a4 + d2); d0 = (((v << 1) & 0xE0) | ((v & 0xF) << 8)) + b650; }
      return i16(d0 & 0xFFFF) >> 5;
    }
    d2 &= ~1;
    if (d1 & 1) { const d3 = this.u8(a5 + d2 + 1); const d0 = ((this.u8(a5 + d2) & 0x7F) << 8 | d3) + b718; return i16(d0 & 0xFFFF) >> 5; }
    const d3 = this.u8(a4 + d2 + 1); const d0 = ((this.u8(a4 + d2) & 0x7F) << 8 | d3) + b650; return i16(d0 & 0xFFFF) >> 5;
  }
  secAdvance5C51A() { let d1 = this.u8(0x1BB85) + 1; if (d1 >= this.u8(0x1CA1A)) d1 = 0; this.B[0x1BB85] = d1 & 0xFF; }
  secRetreat5C538() { let d1 = i8(this.u8(0x1BB85)) - 1; if (d1 < 0) d1 = i8(this.u8(0x1CA1A)) - 1; this.B[0x1BB85] = d1 & 0xFF; }
  edgeCross5C484() {
    let dc;
    if (i8((this.u8(0x1BC40) ^ this.u8(0x1BC32)) & 0xFF) >= 0) {
      this.secAdvance5C51A(); this.setup5FE56(this.u8(0x1BB85)); dc = i8(this.u8(0x1BC32)) < 0;
    } else {
      this.secRetreat5C538(); this.setup5FE56(this.u8(0x1BB85)); dc = i8(this.u8(0x1BC32)) >= 0;
    }
    if (dc) { this.B[0x1BBA3] = (this.u8(0x1BB97) - 4) & 0xFF; if (i8(this.u8(0x1BC40)) < 0) return; }
    else { this.B[0x1BBA3] = 0; if (i8(this.u8(0x1BC40)) >= 0) return; }
    let vv = (256 - this.u8(0x1BC41)) & 0xFF; this.B[0x1BC41] = vv !== 0 ? vv : 0xFF;
    vv = (256 - this.u8(0x1BC4D)) & 0xFF; this.B[0x1BC4D] = vv !== 0 ? vv : 0xFF;
  }
  offTrack5C872() { this.setL(0x1BB18, 0x1000); this.B[0x1BB9A] = ((this.u8(0x1BB9A) >> 1) | 0x80) & 0xFF; }
  offEdge5C808() {
    const v = this.w(0x1BC22);
    let d0;
    if (v < 0) d0 = -v; else { d0 = 0x180 - v; if (d0 < 0) d0 = -d0; }
    if (d0 > 0x30) { this.offTrack5C872(); return; }
    d0 = (d0 & 0xFF) << 4;
    const d3 = (this.l(0x1BB18) - d0 - 0x100) | 0;
    if (d3 < 0x1000) { this.offTrack5C872(); return; }
    this.setL(0x1BB18, d3);
    if (((this.u8(0x1BC22) ^ this.u8(0x1BC32)) & 0x80) !== 0) this.B[0x1BBDA] = 0x80; else this.B[0x1BBDA] = 0x40;
  }
  store5C5F2(d1) {
    this.interp5C554();
    const off = d1 << 1;
    if ((this.u8(0x1BB65) & 0x80) !== 0) { this.B[0x1BB65] &= ~0x80; this.offEdge5C808(); }
    else this.B[0x1BB65] &= ~0x80;
    if (i8(this.u8(0x1BD5C)) >= 0x0A) { this.setL(0x1BCA4 + off, this.l(0x1BB18)); return; }
    let r = this.u8(0x1BCE4); if (i8(r) < 0) r = (256 - r) & 0xFF;
    if (i8(r & 0xFF) > 5) { this.setL(0x1BCA4 + off, this.l(0x1BB18)); return; }
    const a = this.l(0x1BCA4 + off) >>> 0, b = this.l(0x1BB18) >>> 0;
    const sum = (a + b) >>> 0;
    const carry = (a + b) > 0xFFFFFFFF ? 1 : 0;
    this.setL(0x1BCA4 + off, ((carry << 31) | (sum >>> 1)) | 0);
  }
  mulSwap(d0, d3) { return i16((Math.imul(i16(d0), i16(d3)) << 1) >> 16); }
  surface5C1D0() {
    let sec = this.u8(0x1BB1C);
    this.B[0x1BB85] = sec;
    this.setup5FE56(sec);
    this.B[0x1BB9A] = 0;
    let d1 = 4;
    for (;;) {
      this.B[0x1BBF9] = d1 & 0xFF;
      if (this.u8(0x1BB1C) !== this.u8(0x1BB85)) {
        const s = this.u8(0x1BB1C); this.B[0x1BB85] = s; this.setup5FE56(s); d1 = this.u8(0x1BBF9);
      }
      this.B[0x1BB1A] = this.u8(0x1BB7B);
      const pos = ((this.w(0x1BD02 + d1) >> 4) + this.w(0x1BC5E)) | 0;
      let d0;
      if ((u16(pos) >>> 0) >= 0x180) {
        this.B[0x1BB65] |= 0x80; this.setW(0x1BC22, i16(pos));
        d0 = i16(pos) < 0 ? 0 : 0xFF;
      } else {
        let a = i16(pos); if (a < 0) a = -a;
        d0 = this.mulSwap(a, (this.u8(0x1BB1A) << 7) & 0x7FFF);
        if (d0 >= 0x100) d0 = this.u8(0xFF);
      }
      this.B[0x1BC4D] = d0 & 0xFF;
      if (i8(this.u8(0x1BC32)) < 0) d0 = (d0 ^ 0xFF) & 0xFF;
      if (d1 === 4) this.B[0x1BBA1] = d0 & 0xFF;
      this.B[0x1BB1A] = this.u8(0x1BBD9);
      const ac = (this.mulSwap(this.w(0x1BD08 + d1) >> 3, (this.u8(0x1BB1A) << 7) & 0x7FFF) + this.w(0x1BB10)) | 0;
      this.setW(0x1BC40, i16(ac));
      const e = (this.u8(0x1BC40) << 1) & 0xFF;
      this.B[0x1BBA3] = e;
      if (i8(e) < 0 || i8(e) >= i8(this.u8(0x1BB98))) this.edgeCross5C484();
      const a4 = this.handlePhys(u16(this.w(0x1BC8C)));
      const a5 = this.handlePhys(u16(this.w(0x1BC90)));
      if (i8(this.u8(0x1BC32)) < 0) {
        const vi = (this.u8(0x1BB97) - this.u8(0x1BBA3) - 4) & 0xFF;
        this.setW(0x1BC08, this.railHeight5C0AA(a4, a5, vi));
        this.setW(0x1BC06, this.railHeight5C0AA(a4, a5, (vi + 1) & 0xFF));
        this.setW(0x1BC04, this.railHeight5C0AA(a4, a5, (vi + 2) & 0xFF));
        this.setW(0x1BC02, this.railHeight5C0AA(a4, a5, (vi + 3) & 0xFF));
      } else {
        const vi = this.u8(0x1BBA3);
        this.setW(0x1BC02, this.railHeight5C0AA(a4, a5, vi));
        this.setW(0x1BC04, this.railHeight5C0AA(a4, a5, (vi + 1) & 0xFF));
        this.setW(0x1BC06, this.railHeight5C0AA(a4, a5, (vi + 2) & 0xFF));
        this.setW(0x1BC08, this.railHeight5C0AA(a4, a5, (vi + 3) & 0xFF));
      }
      this.store5C5F2(this.u8(0x1BBF9));
      d1 = i8(this.u8(0x1BBF9)) - 2;
      if (d1 < 0) break;
    }
  }
  sectionLocate61012() {
    let d0 = 0, d4 = 0, dd0 = 0;
    const sec = this.u8(0x1BB1C);
    this.B[0x1BB85] = sec; this.setup5FE56(sec);
    const d3w = this.w(0x1BC32);
    d4 = i16(u16(this.w(0x1BD5A) - this.w(A.Yaw)) ^ u16(d3w));
    let d2 = 0;
    if (i8(this.u8(0x1BB4D)) < 0) { d2 += 2; if (i16(u16(this.w(0x1BC44)) ^ u16(d3w)) < 0) d2 += 2; }
    d4 = i16(d4 + this.w(0x6125A + d2));
    d0 = d4; if (i16(d0) < 0) d0 = -i16(d0);
    this.setW(0x1BC2A, i16(d0)); this.setW(0x1BBF6, i16(d4));
    if (u16(d0) >= 0x800) d0 = 0x7FFF; else d0 = i16(i16(d0) << 4);
    this.setW(0x1BC3C, i16(d0));
    if (((this.u8(0x1BB6A) - this.u8(0x1BB0A)) & 0xFF) < 2) { this.secAdvance5C51A(); this.setup5FE56(this.u8(0x1BB85)); }
    this.B[0x1BB5D] = (this.u8(0x1BC44) ^ this.u8(0x1BC32)) & 0xFF;
    let label;
    if (this.u8(0x1BBC6) === 0) label = 'branch3C';
    else {
      this.B[0x1BB1B] = (this.u8(0x1BBC6) ^ this.u8(0x1BBF6)) & 0xFF;
      if (i8(this.u8(0x1BB4D)) < 0) {
        if (i8((this.u8(0x1BBC6) ^ this.u8(0x1BB5D)) & 0xFF) < 0) {
          this.B[0x1BBC6] = this.u8(0x1BB5D); dd0 = (this.u8(0x1BBD4) - 0x23) & 0xFF; label = 'label21C';
        } else { dd0 = (this.u8(0x1BBD4) + 0x2D) & 0xFF; label = 'label126'; }
      } else { dd0 = this.u8(0x1BBD4); label = 'label126'; }
    }
    for (;;) {
      if (label === 'label126') {
        if (i8(this.u8(0x1BB1B)) >= 0) dd0 = (dd0 + this.u8(0x1BC3C)) & 0xFF;
        label = 'label21C';
      } else if (label === 'branch3C') {
        d4 = 0;
        if (i8(this.u8(0x1BB4D)) < 0) { this.B[0x1BBC6] = this.u8(0x1BB5D); dd0 = this.u8(0x1BBD4); label = 'label21C'; }
        else label = 'label160';
      } else if (label === 'label160') {
        this.B[0x1BBC6] = this.u8(0x1BBF6);
        let d2b = this.w(0x1BC2A) & 0xFF;
        let dv = this.w(0x1BC2A);
        let toLabel1C2 = false;
        if (this.u8(0x1BC2A) !== 0) {
          dv = i16(dv - 0x1E00);
          if (i16(dv) >= 0) { dd0 = dv; label = 'label1C2'; toLabel1C2 = true; }
          else d2b = 0xFF;
        }
        if (!toLabel1C2) {
          this.B[0x1BB1A] = d2b & 0xFF;
          let v = this.w(0x1BD30); if (i16(v) < 0) v = -i16(v);
          v = i16(v + 0xA00); if (i16(v) < 0) v = 0x7F00;
          const d3b = (this.u8(0x1BB1A) << 7) & 0x7FFF;
          v = i16((Math.imul(i16(v), i16(d3b)) << 1) >> 16);
          v = u16(v) >>> 7; if ((v & 0xFF) === 0) v++;
          dd0 = v; label = 'label1C2';
        }
      } else if (label === 'label1C2') {
        if (i8(this.u8(0x1BBF6)) < 0) dd0 = i16(-i16(dd0));
        this.setW(A.Yaw, i16(this.w(A.Yaw) + i16(dd0)));
        d4 = i16(d4 - this.w(A.AmP));
        if ((this.l(0x64AEC) >>> 0) !== (MAGIC >>> 0) || this.u8(A.OnGround) === 0) d4 = 0;
        this.setW(A.TqP, i16(d4)); return;
      } else if (label === 'label21C') {
        this.B[0x1BB1A] = dd0 & 0xFF;
        let v = i16((Math.imul(this.w(0x1BD30), i16((this.u8(0x1BB1A) << 7) & 0x7FFF)) << 1) >> 16);
        if (i8(this.u8(0x1BBC6)) < 0) v = i16(-i16(v));
        d4 = i16(v) >> 3;
        if ((this.u8(0x1BC2A) & 0xFF) < 0x1E) {
          d4 = i16(d4 - this.w(A.AmP));
          if ((this.l(0x64AEC) >>> 0) !== (MAGIC >>> 0) || this.u8(A.OnGround) === 0) d4 = 0;
          this.setW(A.TqP, i16(d4)); return;
        }
        label = 'label160';
      }
    }
  }

  // --- the full frame ($6185C) ---
  frame6185C() {
    this.matrix61368();
    this.corners618CE();
    this.surface5C1D0();
    this.contactHeights61B70();
    this.velToBody6158C();
    this.sound60FBE();
    this.gravToBody615E6();
    this.suspension61BCC();
    this.loadProject622DC();
    this.setW(0x1BD46, this.w(0x1BD44));
    this.tail63E2E();
    if (this.u8(0x620B6) !== 0) this.B[0x620B6] = (this.B[0x620B6] - 1) & 0xFF;
    if (this.u8(0x1BB72) !== 0) {
      this.drive620B8();
      this.sectionLocate61012();
      this.forceToWorld61618();
      this.drag621F4();
      this.torqueApply62138();
      this.torque61B26();
      this.torqueToWorld61672();
    }
    this.force61ADC();
    this.integrate61950();
  }

  // --- render coupling (Part V §7): per-frame placement of the car's section on the track,
  // ported from package physics and verified exact against the engine (cmd/physverify). ---

  le16(a) { return i16(this.u8(a) | (this.u8(a + 1) << 8)); }

  dir5FF94(d0) {
    const cell = this.u8(0x1C588 + (d0 & 0xFF));
    let x = cell & 0x0F, y = (cell >> 4) & 0x0F;
    x = (x - this.u8(0x1BB04)) & 0xFF;
    y = (y - this.u8(0x1BB06)) & 0xFF;
    const q = this.u8(0x1BC30);
    if ((q & 0x80) === 0 && (q & 0x40) !== 0) { [x, y] = [y, x]; x = (-x) & 0xFF; }
    else if ((q & 0x80) !== 0 && (q & 0x40) === 0) { x = (-x) & 0xFF; y = (-y) & 0xFF; }
    else if ((q & 0x80) !== 0 && (q & 0x40) !== 0) { [x, y] = [y, x]; y = (-y) & 0xFF; }
    this.B[0x65EC0] = x; this.B[0x65EC2] = y;
    this.B[0x1BB22] = ((x << 3) + this.u8(0x1BB2E)) & 0xFF;
    this.B[0x1BB26] = ((y << 3) + this.u8(0x1BB32)) & 0xFF;
  }

  planPoint5C3DA() {
    const q = this.u8(0x1BBF2), bx = this.w(0x1BB22), by = this.w(0x1BB26);
    if ((q & 0x80) === 0 && (q & 0x40) === 0) { this.setW(0x1BBF6, -bx); this.setW(0x1BBF8, -by); }
    else if ((q & 0x80) === 0 && (q & 0x40) !== 0) { this.setW(0x1BBF6, -by); this.setW(0x1BBF8, 0x800 + bx); }
    else if ((q & 0x80) !== 0 && (q & 0x40) === 0) { this.setW(0x1BBF6, 0x800 + bx); this.setW(0x1BBF8, 0x800 + by); }
    else { this.setW(0x1BBF6, 0x800 + by); this.setW(0x1BBF8, -bx); }
  }

  couple5BE44() {
    const sec = this.u8(0x1BB85);
    this.setup5FE56(sec);
    const a5 = this.handlePhys(u16(this.w(0x1BCBC)));
    this.dir5FF94(sec);
    this.setW(0x1BBF2, this.w(0x1BC30) - this.w(0x1BC4A));
    const flags = this.u8(0x1BB4D);
    if (flags & 0x80) { this.couple5BF50(a5); }
    else if (flags & 0x40) {
      this.planPoint5C3DA();
      this.B[0x1BB1A] = 0xB5;
      let p = (Math.imul(i16(this.w(0x1BBF6) - this.w(0x1BBF8)), i16((0xB5 << 7) & 0x7FFF)) << 1) >> 16;
      this.setW(0x1BC5E, i16(p) - this.le16(a5 + 2));
      this.B[0x1BB1A] = this.u8(a5 + 7);
      let p2 = (Math.imul(i16(this.w(0x1BBF6) + this.w(0x1BBF8)), i16(((this.u8(0x1BB1A) << 7) & 0x7FFF))) << 1) >> 16;
      this.setW(0x1BB10, i16(p2));
      this.setW(0x1BD5A, this.le16(a5 + 4) + this.w(0x1BC4A));
    } else {
      this.planPoint5C3DA();
      this.setW(0x1BC5E, this.w(0x1BBF6) - this.le16(a5 + 2));
      this.setW(0x1BB10, this.w(0x1BBF8));
      this.setW(0x1BD5A, this.w(0x1BC4A));
    }
  }

  // planPoint5C6C4 ($5C6C4): piece-shape plan vertex at byte offset d2 (two LE words),
  // added to the plan base $1BB22/$1BB26 and rotated by the section quadrant $1BBF2.
  planPoint5C6C4(d2) {
    const a5 = this.handlePhys(u16(this.w(0x1BCBC)));
    const v0 = this.le16(a5 + d2), v1 = this.le16(a5 + d2 + 2);
    const bx = this.w(0x1BB22), by = this.w(0x1BB26);
    const q = this.u8(0x1BBF2);
    if ((q & 0x80) === 0 && (q & 0x40) === 0) { this.setW(0x1BBF6, bx + v0); this.setW(0x1BBF8, by + v1); }
    else if ((q & 0x80) === 0) { this.setW(0x1BBF8, by + v0); this.setW(0x1BBF6, bx - v1 + 0x800); }
    else if ((q & 0x40) === 0) { this.setW(0x1BBF6, bx - v0 + 0x800); this.setW(0x1BBF8, by - v1 + 0x800); }
    else { this.setW(0x1BBF8, by - v0 + 0x800); this.setW(0x1BBF6, bx + v1); }
  }

  // atan264D66 ($64D66): quarter-table ATAN2 (table $1CC46, full circle $10000, 0 = +y).
  // Returns [angle, ratio]; the ratio is the d7 quotient hypot64DE8 reuses. On DIVU
  // overflow the destination's cleared low word stays 0, so the ratio reads as 0.
  atan264D66(x, y) {
    const ax = x < 0 ? i16(-x) : x, ay = y < 0 ? i16(-y) : y;
    let d7, t;
    if (ay === ax) { d7 = 0xFFFF; t = 0x2000; }
    else if (ay > ax) {
      let q = Math.floor((u16(ax) * 0x10000) / u16(ay)); if (q > 0xFFFF) q = 0;
      d7 = q; t = this.w(0x1CC46 + ((d7 >> 4) & ~1));
    } else {
      let q = Math.floor((u16(ay) * 0x10000) / u16(ax)); if (q > 0xFFFF) q = 0;
      d7 = q; t = this.w(0x1CC46 + ((d7 >> 4) & ~1));
      if (i16(x ^ y) >= 0) t = i16(-t); // same signs -> negate (BMI skips the NEG)
      return [i16((x < 0 ? 0xC000 : 0x4000) + t), d7];
    }
    if (i16(x ^ y) < 0) t = i16(-t); // opposite signs -> negate (BPL skips the NEG)
    if (y < 0) t = i16(t + 0x8000);
    return [i16(t), d7];
  }

  // hypot64DE8 ($64DE8): table hypot (table $1DC46) of the original x/y, indexed by the
  // atan2 ratio: max + (table * min >> 16).
  hypot64DE8(x, y, d7) {
    let ax = x < 0 ? i16(-x) : x, ay = y < 0 ? i16(-y) : y;
    if (ay < ax) { const tv = ax; ax = ay; ay = tv; }
    const t = this.w(0x1DC46 + ((d7 >> 4) & ~1));
    const p = Math.floor((u16(t) * u16(ax)) / 0x10000);
    return i16(p + u16(ay));
  }

  // radius5C65A ($5C65A): refine the curve radius $1BC36 by a step proportional to
  // (x^2 + z^2 - r^2), scaled by the per-type-nibble coefficient from $5C6B8.
  radius5C65A() {
    const sq = (v) => { let p = Math.imul(i16(v), i16(v)); if (p < 0) p = -p; return p; };
    let d4 = (sq(this.w(0x1BBF6)) + sq(this.w(0x1BBF8))) | 0;
    const rsq = sq(this.w(0x1BC36));
    this.B[0x1BB1A] = this.u8(0x5C6B8 + this.u8(0x1BB86));
    d4 = (d4 - rsq) | 0;
    const d0 = i16(d4 >>> 8); // LSR.l #8, low word
    const coef = (this.u8(0x1BB1A) << 7) & 0x7FFF;
    const p = Math.imul(d0, coef) << 1;
    const r = i16(p >> 16) >> 4; // SWAP -> high word, ASR.w #4
    this.setW(0x1BC36, i16(this.w(0x1BC36) + r));
  }

  // couple5BF50 ($5BF50, ramp-type-2 = curved ramp pieces, $1BB4D bit 7): place the car
  // on the piece's ARC. atan2/hypot of the plan point give the polar angle/radius around
  // the curve centre; heading $1BC1C / orientation ref $1BD5A from angle + camera quadrant;
  // along-progress $1BB10 = swept angle x arc coefficient (header byte 8) scaled by the
  // exponent in the ramp-flag low bits; past the piece end step the section and, for a
  // type-nibble-4 neighbour, re-enter the whole coupling; then refine the radius and set
  // $1BC5E = piece radius (header word 9) - car radius, sign-flipped by $1BC44.
  couple5BF50(a5) {
    this.planPoint5C6C4(2);
    const x = this.w(0x1BBF6), y = this.w(0x1BBF8);
    const [ang, d7] = this.atan264D66(x, y);
    this.setW(0x1BC36, this.hypot64DE8(x, y, d7));
    let d0 = i16(ang + this.w(0x1BC30));
    d0 = i16(d0 + 0x8000); // the BPL/BMI paths both flip by $8000 (word wrap)
    this.setW(0x1BC1C, d0);
    this.setW(0x1BD5A, i16(d0 + 0x4000 - this.w(0x1BC44)));
    const d4 = i8(1 - (this.u8(0x1BB4D) & 3)); // shift exponent from the ramp-flag low bits
    const arc0 = i16(this.le16(a5 + 6) << 6);
    let sweep = i16(this.w(0x1BC1C) - arc0 - this.w(0x1BC4A));
    if (sweep < 0) sweep = i16(-sweep);
    this.B[0x1BB1A] = this.u8(a5 + 8);
    const coef = (this.u8(0x1BB1A) << 7) & 0x7FFF;
    let along = i16((Math.imul(sweep, coef) << 1) >> 16);
    if (d4 < 0) along = i16(u16(along) << ((-d4) & 7)); // ASL.w
    else along = i16(u16(along) >>> (d4 & 7));          // LSR.w
    this.setW(0x1BB10, along);
    const e = ((u16(along) >>> 7) + 2) & 0xFF;
    if (i8(e) >= i8(this.u8(0x1BB97))) {
      const nib = this.u8(0x1BB86);
      if (nib === 1 || nib === 3) {
        this.B[0x1BB19] = this.u8(0x1BB85);
        if (this.u8(0x1BC32) !== 0) this.secRetreat5C538(); else this.secAdvance5C51A();
        if ((this.u8(0x1C5EC + this.u8(0x1BB85)) & 0x0F) === 4) { this.couple5BE44(); return; }
        this.B[0x1BB85] = this.u8(0x1BB19);
      }
    }
    this.radius5C65A();
    let d3 = i16((this.u8(a5 + 10) << 8) | this.u8(a5 + 9)); // le16(a5+9)
    d3 = i16(d3 - this.w(0x1BC36));
    if (i8(this.u8(0x1BC44)) < 0) d3 = i16(-d3);
    this.setW(0x1BC5E, d3);
  }

  // placeCar605B6 reproduces $605B6's placement path: seat the car at the start section's
  // grid cell (x/z = cell*128 + 64), facing along the section, run the coupling + one
  // physics tick, then force posY to the local resting height (16.0). Mirrors the original.
  placeCar605B6(section) {
    this.B[0x1BB85] = section; this.B[0x1BB1C] = section;
    this.setup5FE56(section);
    const cell = this.u8(0x1C588 + section);
    this.B[0x1BBD5] = cell & 0x0F;
    this.B[0x1BBD6] = (cell >> 4) & 0x0F;
    this.setW(0x1BCDA, 0); this.setW(0x1BCE2, 0);
    this.setW(0x1BCD8, ((this.u8(0x1BBD5) << 7) + 0x40)); // posX high word
    this.setW(0x1BCE0, ((this.u8(0x1BBD6) << 7) + 0x40)); // posZ high word
    const t = this.u8(0x1BB86); const d1 = (t === 4 || t === 0x0A) ? 0x20 : 0;
    this.B[0x1BCE6] = ((this.u8(0x1BC4A) ^ this.u8(0x1BC32)) + d1) & 0xFF; // yaw high byte
    this.B[0x1BCE7] = 0;
    this.setW(0x1BCDC, 0x0400); this.setW(0x1BCDE, 0); // posY = 1024 (pre-settle)
    this.camera60190();
    this.couple5BE44();
    this.setW(A.Drive, 0); this.frame6185C();
    this.setW(0x1BCDE, 0); this.setW(0x1BCDC, 0x10); // posY = 16.0 (local resting frame)
    for (const a of [A.VelX, A.VelY, A.VelZ, A.AmR, A.AmP, A.AmY]) this.setW(a, 0);
  }

  // driveTickCoupled runs the real per-frame render coupling (camera-follow + section +
  // $5BE44 placement) so the suspension samples the REAL track surface under the wheels,
  // then the verified physics frame. The grounded drive block is prone to a vertical
  // launch on a near-rest car (an artefact of running the sim outside the original's full
  // render state -- see Part V Sec.7), so we keep the car seated vertically (clamp posY to
  // the local resting band) and bound the attitude, while the throttle/grip/drag and the
  // surface-following are the exact model. Returns the world speed for the viewer.
  driveTickCoupled(throttle, section) {
    if (section !== undefined) { this.B[0x1BB1C] = section & 0xFF; this.B[0x1BB85] = section & 0xFF; }
    this.camera60190();
    this.couple5BE44();
    this.setW(A.Drive, throttle);
    this.frame6185C();
    // Keep the car seated: the grounded block tends to launch a near-rest car (an artefact
    // of running the sim outside the original's full render state). Re-seat posY to the
    // local resting band each frame so the car tracks the surface instead of flying off,
    // and bleed off any upward velocity, while throttle/grip/drag and surface-following run.
    this.setW(A.PosY, 0x10);
    if (this.w(A.VelY) > 0) this.setW(A.VelY, 0);
    for (const a of [A.Roll, A.Pit]) { const v = this.w(a); if (v > 0x400) this.setW(a, 0x400); else if (v < -0x400) this.setW(a, -0x400); }
    for (const a of [A.AmR, A.AmP, A.AmY]) { const v = this.w(a); if (v > 0x80) this.setW(a, 0x80); else if (v < -0x80) this.setW(a, -0x80); }
    const vx = this.w(A.VelX), vz = this.w(A.VelZ);
    return Math.round(Math.sqrt(vx * vx + vz * vz));
  }

  refGrid600A6(d0q) {
    this.B[0x1BB19] = (d0q >> 8) & 0xFF;
    if (this.w(0x1BCDA) === 0) this.setW(0x1BCDA, 1);
    const hx = ((this.l(0x1BCD8) << 1) >>> 16) & 0xFFFF;
    this.B[0x1BB07] = hx & 0xFF; this.B[0x1BB04] = (hx >> 8) & 0xFF;
    if (this.w(0x1BCE2) === 0) this.setW(0x1BCE2, 1);
    const hz = ((this.l(0x1BCE0) << 1) >>> 16) & 0xFFFF;
    this.B[0x1BB09] = hz & 0xFF; this.B[0x1BB06] = (hz >> 8) & 0xFF;
    const q = this.u8(0x1BB19), refl = 0x8000000;
    if ((q & 0x80) === 0 && (q & 0x40) === 0) { this.setL(0x1BCCC, this.l(0x1BCD8)); this.setL(0x1BCD4, this.l(0x1BCE0)); }
    else if ((q & 0x80) === 0 && (q & 0x40) !== 0) { this.setL(0x1BCD4, this.l(0x1BCD8)); this.setL(0x1BCCC, (refl - this.l(0x1BCE0)) | 0); }
    else if ((q & 0x80) !== 0 && (q & 0x40) === 0) { this.setL(0x1BCCC, (refl - this.l(0x1BCD8)) | 0); this.setL(0x1BCD4, (refl - this.l(0x1BCE0)) | 0); }
    else { this.setL(0x1BCD4, (refl - this.l(0x1BCD8)) | 0); this.setL(0x1BCCC, this.l(0x1BCE0)); }
  }

  camera60190() {
    const q = i16((u16(this.w(0x1BCE6)) + 0x2000) & 0xC000);
    this.setW(0x1BC30, q);
    this.refGrid600A6(q & 0xFFFF);
    this.setW(0x1BCD0, i16((this.l(0x1BCDC) >>> 11) & 0xFFFF));
    let d3 = 0x780, d0 = this.w(0x1BD38);
    if (u16(d0) >= 0x500) { d0 = i16(d0 << 1); d3 = 0x280; }
    d0 = i16(d0 + d3);
    const roll = this.w(0x1BCE4);
    if (roll < 0) d0 = i16(d0 - (roll >> 1));
    d0 = i16((d0 >> 4) + this.w(0x1BCD0));
    this.setW(0x1BBFA, d0);
    const cx = i16(-(((this.l(0x1BCCC) >>> 12) & 0x7FF)));
    this.B[0x1BB23] = cx & 0xFF; this.B[0x1BB2E] = (u16(cx) >> 8) & 0xFF;
    const cz = i16(-(((this.l(0x1BCD4) >>> 12) & 0x7FF)));
    this.B[0x1BB27] = cz & 0xFF; this.B[0x1BB32] = (u16(cz) >> 8) & 0xFF;
    this.setW(0x1BC2E, i16(((u16(this.w(0x1BCE6)) + 0x2000) & 0x3FFE) - 0x2000));
  }

  section5FE04() {
    const gy = (this.u8(0x1BB06) + this.u8(0x1BBD6)) & 0xFF;
    if (gy >= 0x10) return { sec: 0, off: true };
    this.B[0x1BB1B] = (gy << 4) & 0xFF;
    const gx = (this.u8(0x1BB04) + this.u8(0x1BBD5)) & 0xFF;
    if (gx >= 0x10) return { sec: 0, off: true };
    return { sec: this.u8(0x1C280 + (((gx & 0x0F) | this.B[0x1BB1B]) & 0xFF)), off: false };
  }

  // driveTick is a PROVISIONAL drive coupling (not the verified frame6185C). The exact
  // physics needs per-frame state the original computes in its render pass -- where the
  // car sits on the track, the surface under each wheel, the section heading -- which
  // isn't reimplemented yet (the remaining piece of the Part V gameplay). Without it the
  // isolated physics floats in a
  // mismatched frame and tumbles. So here we (1) inject a flat ground just under the
  // wheels so the verified suspension/grip/drive engage -- the throttle/acceleration/drag
  // are the real model -- and (2) keep the chassis roughly upright, since the render
  // coupling that would balance the drive->roll torque against the banked track is the
  // missing piece. Returns the body-frame world speed magnitude for the viewer.
  driveTick(throttle) {
    this.matrix61368();
    this.corners618CE();
    this.contactHeights61B70();
    const off = 0x800 << 8; // mid-range compression so the springs hold the car
    this.setL(0x1BCA4, (this.l(0x1BC94) + this.l(A.Rest) + off) | 0);
    this.setL(0x1BCA8, (this.l(0x1BC98) + this.l(A.Rest) + off) | 0);
    this.setL(0x1BCAC, (this.l(0x1BC9C) + this.l(A.Rest) + off) | 0);
    this.velToBody6158C();
    this.sound60FBE();
    this.gravToBody615E6();
    this.suspension61BCC();
    this.loadProject622DC();
    this.setW(0x1BD46, this.w(0x1BD44));
    this.tail63E2E();
    if (this.u8(0x1BB72) !== 0) {
      this.setW(A.Drive, throttle);
      this.drive620B8();   // exact drive force + grip clamp
      this.forceToWorld61618();
      this.drag621F4();    // exact velocity drag
      this.torqueApply62138();
      this.torque61B26();
      this.torqueToWorld61672();
    }
    this.force61ADC();
    this.integrate61950();
    for (const a of [A.Roll, A.Pit]) { const v = this.w(a); if (v > 0x400) this.setW(a, 0x400); else if (v < -0x400) this.setW(a, -0x400); }
    for (const a of [A.AmR, A.AmP, A.AmY]) { const v = this.w(a); if (v > 0x80) this.setW(a, 0x80); else if (v < -0x80) this.setW(a, -0x80); }
    const vx = this.w(A.VelX), vz = this.w(A.VelZ);
    return Math.round(Math.sqrt(vx * vx + vz * vz));
  }

  // self-check against the Go golden trace; returns the first mismatching frame or -1.
  selfTest(trace) {
    const L = ['Drive', 'PosX', 'PosY', 'PosZ', 'Roll', 'Yaw', 'Pit', 'VelX', 'VelY', 'VelZ'];
    for (let f = 0; f < trace.frames.length; f++) {
      const fr = trace.frames[f];
      this.setW(A.Drive, fr[0]);
      this.frame6185C();
      const got = [fr[0], this.l(A.PosX), this.l(A.PosY), this.l(A.PosZ),
        this.w(A.Roll), this.w(A.Yaw), this.w(A.Pit), this.w(A.VelX), this.w(A.VelY), this.w(A.VelZ)];
      for (let k = 1; k < got.length; k++) {
        if ((got[k] | 0) !== (fr[k] | 0)) {
          console.warn(`physics selfTest: frame ${f} ${L[k]} js=${got[k]} go=${fr[k]}`);
          return f;
        }
      }
    }
    return -1;
  }
}
