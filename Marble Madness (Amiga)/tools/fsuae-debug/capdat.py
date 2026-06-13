#!/usr/bin/env python3
# Capture the .dat key_init: catch each c/zzz AllocMem(220) and read key_init's
# count; the c/xxx pass is count=0, the .dat pass is count=20 -> read its 20-long
# key array. Boots auto-run full speed, polls for the launcher, then arms the bp.
import dbg, json, sys, time

ALLOC_RET = 0x2C5F4E75
PROT_LINK = 0x4E56FFEC
OFF_ALLOC_RET = 0xF42
OFF_PROT = 0xDAA

d = dbg.Dbg(timeout=130)

def tasklist():
    sb = d.sysbase(); out = []
    for off in (0x196, 0x1a4):
        node = d.rl(sb + off)
        for _ in range(50):
            succ = d.rl(node)
            if succ == 0 or succ == node:
                break
            out.append(d.amiga_str(d.rl(node + 10))); node = succ
    return out

print("polling for launcher...", flush=True)
seen = False
for i in range(600):
    d.interrupt()
    tl = tasklist()
    if 'Initial CLI' in tl:
        seen = True; print("launcher up:", tl, flush=True); break
    if any(t in ('Painter','Framer','SfxTask','Scroller') for t in tl):
        print("game already running -- missed it", flush=True); d.cont(); sys.exit(2)
    d.cont(); time.sleep(0.2)
if not seen:
    print("no launcher", flush=True); sys.exit(1)

d.bp(0xC001B0)
print("bp armed, hunting key_init passes...", flush=True)
passes = []
for i in range(500000):
    try:
        pkt = d.cont_wait(200)
    except Exception as e:
        print("timeout #%d: %s" % (i, e), flush=True); break
    sr = d.stop_regs(pkt)
    if sr.get(0, 0) != 0xDC or sr.get(1, 0) != 0x10002:
        continue
    a7 = sr.get(15, 0)
    ret = d.rl(a7)
    try:
        if d.rl(ret) != ALLOC_RET:
            continue
        base = ret - OFF_ALLOC_RET
        if d.rl(base + OFF_PROT) != PROT_LINK:
            continue
    except Exception:
        continue
    keyinit_a6 = d.rl(a7 + 4)
    cnt = d.rl(keyinit_a6 + 0xC)
    kp = d.rl(keyinit_a6 + 8)
    print("key_init pass: count=%d keyarray_ptr=%08X base=%08X" % (cnt, kp, base), flush=True)
    passes.append(cnt)
    if cnt == 20:
        ka = d.rd(kp, cnt * 4)
        tt = d.thistask()
        cap = dict(count=cnt, keyarray_ptr=kp,
                   keyarray_words=["%08X" % int.from_bytes(ka[j:j+4], "big") for j in range(0, len(ka), 4)],
                   tc_ExceptCode_2A="%08X" % d.rl(tt + 0x2A),
                   tc_TrapCode_32="%08X" % d.rl(tt + 0x32),
                   task_name=d.amiga_str(d.rl(tt + 10)))
        json.dump(cap, open("/tmp/datkey.json", "w"), indent=2)
        print("CAPTURED .dat keyarray:", cap["keyarray_words"], flush=True)
        print("task=%r tc_Except=%s tc_Trap=%s" % (cap["task_name"], cap["tc_ExceptCode_2A"], cap["tc_TrapCode_32"]), flush=True)
        sys.exit(0)
print("passes seen:", passes, flush=True)
print("did not catch count=20 pass", flush=True)
