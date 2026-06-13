#!/usr/bin/env python3
"""Capable GDB-RSP client for the FS-UAE m68k stub + Amiga exec helpers."""
import socket, time, re

class Dbg:
    def __init__(self, host="127.0.0.1", port=6860, timeout=5):
        self.s = socket.create_connection((host, port), timeout=timeout)
        self.s.settimeout(timeout); self.buf = b""

    def _rb(self):
        while not self.buf:
            d = self.s.recv(8192)
            if not d: raise EOFError("closed")
            self.buf += d
        b, self.buf = self.buf[:1], self.buf[1:]; return b

    def _send(self, data):
        cks = sum(data.encode()) & 0xFF
        self.s.sendall(b"$" + data.encode() + b"#" + ("%02x" % cks).encode())
        while True:                      # wait for +/-
            b = self._rb()
            if b == b"+": return
            if b == b"-": self.s.sendall(b"$" + data.encode() + b"#" + ("%02x" % cks).encode())

    def _recv_pkt(self):
        while True:
            b = self._rb()
            if b == b"$": break
            if b == b"%":                # async notification: read like a packet
                break
        data = b""
        while True:
            b = self._rb()
            if b == b"#": break
            data += b
        self._rb(); self._rb()           # checksum
        self.s.sendall(b"+")
        return data.decode(errors="replace")

    def cmd(self, data):
        self._send(data); return self._recv_pkt()

    def cont(self):
        self._send("c")                  # 'c' replies nothing until a stop

    def wait_stop(self, timeout=10):
        """Read packets until a stop reply (T/S/W/X or a '%Stop' notification)."""
        old = self.s.gettimeout(); self.s.settimeout(timeout)
        try:
            while True:
                p = self._recv_pkt()
                if p[:1] in ("T","S","W","X") or p.startswith("Stop"):
                    return p
        finally: self.s.settimeout(old)

    def cont_wait(self, timeout=10):
        self.cont(); return self.wait_stop(timeout)

    def interrupt(self):
        self.s.sendall(b"\x03"); return self._recv_pkt()

    def drain(self, t=0.3):
        old = self.s.gettimeout(); self.s.settimeout(t); out=b""
        try:
            while True: out += self.s.recv(8192)
        except Exception: pass
        finally: self.s.settimeout(old)
        self.buf = b""
        return out

    # ---- registers ----
    def regs(self):
        g = self.cmd("g"); v=[int(g[i:i+8],16) for i in range(0,18*8,8)]
        return {"D":v[0:8],"A":v[8:16],"SR":v[16],"PC":v[17]}

    # ---- memory ----
    def rd(self, addr, n):
        m = self.cmd("m%x,%x" % (addr, n))
        if m.startswith("E"): raise IOError("read err %s @%x" % (m, addr))
        return bytes(int(m[i:i+2],16) for i in range(0,len(m),2))
    def rl(self, addr):
        b = self.rd(addr,4); return int.from_bytes(b,"big")
    def rw(self, addr):
        b = self.rd(addr,2); return int.from_bytes(b,"big")

    # ---- breakpoints (absolute) ----
    def bp(self, addr):   return self.cmd("Z0,%x" % addr)
    def rmbp(self, addr): return self.cmd("z0,%x" % addr)

    @staticmethod
    def stop_regs(pkt):
        """Parse 'NN:value;' register pairs from a T-stop packet -> dict idx->int."""
        out = {}
        for m in re.finditer(r"([0-9a-f]{2}):([0-9a-f]{8})", pkt):
            out[int(m.group(1),16)] = int(m.group(2),16)
        return out   # 0-7=D0-D7, 8-15=A0-A7, 16=SR? 17=PC (stub order)

    # ---- amiga exec helpers ----
    def sysbase(self): return self.rl(4)
    def thistask(self): return self.rl(self.sysbase()+0x114)
    def amiga_str(self, ptr, maxlen=40):
        if not ptr: return ""
        b=self.rd(ptr,maxlen);
        z=b.find(b"\x00"); return b[:z if z>=0 else maxlen].decode("latin1")
    def liblist(self):
        sb=self.sysbase(); head=self.rl(sb+0x17a)   # ExecBase.LibList.lh_Head
        out=[]; node=head
        for _ in range(60):
            succ=self.rl(node)
            if succ==0: break
            name=self.amiga_str(self.rl(node+10))
            out.append((node,name)); node=succ
        return out
    def find_lib(self, want):
        for base,name in self.liblist():
            if name==want: return base
        return None

if __name__=="__main__":
    d=Dbg()
    d.interrupt()
    r=d.regs(); print("PC=%08X" % r["PC"])
    sb=d.sysbase(); print("SysBase=%08X ThisTask=%08X" % (sb, d.thistask()))
    print("--- libraries ---")
    for base,name in d.liblist():
        print("  %08X  %s" % (base, name))
    dos=d.find_lib("dos.library")
    print("dos.library base=%s" % ("%08X"%dos if dos else "NOT FOUND"))
    if dos:
        # LoadSeg LVO -150: read the JMP at base-150
        jmp=d.rd(dos-150,6)
        print("  base-150 bytes:", jmp.hex(), "(4ef9=JMP abs)")
        if jmp[0]==0x4e and jmp[1]==0xf9:
            print("  LoadSeg ROM code @ %08X" % int.from_bytes(jmp[2:6],"big"))
    d.cont()
