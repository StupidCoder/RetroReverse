#!/usr/bin/env python3
# Capture c/zzz's key_init key-array: boot full-speed, poll for the launcher
# (Initial CLI) task, then arm the AllocMem(220) breakpoint and read key_init's
# args (keyarray ptr + count) + the protection inputs.
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

# phase 1: poll (full speed between polls) until the launcher CLI appears
print("polling for launcher (Initial CLI)...", flush=True)
launcher_seen = False
for i in range(400):
    d.interrupt()
    tl = tasklist()
    if any(t in ('Painter', 'Framer', 'SfxTask', 'Scroller') for t in tl):
        print("game already running -- missed c/zzz (tasks=%s)" % tl, flush=True)
        d.cont(); sys.exit(2)
    if 'Initial CLI' in tl:
        launcher_seen = True
        print("launcher detected (tasks=%s) -- arming AllocMem bp" % tl, flush=True)
        break
    d.cont(); time.sleep(0.25)
if not launcher_seen:
    print("launcher never appeared", flush=True); sys.exit(1)

# phase 2: arm AllocMem bp and catch key_init's AllocMem(220)
d.bp(0xC001B0)
found = None
for i in range(200000):
    try:
        pkt = d.cont_wait(60)
    except Exception as e:
        print("timeout #%d: %s" % (i, e), flush=True); break
    sr = d.stop_regs(pkt)
    d0, d1, a7 = sr.get(0, 0), sr.get(1, 0), sr.get(15, 0)
    if d0 != 0xDC or d1 != 0x10002:
        continue
    ret = d.rl(a7)
    try:
        if d.rl(ret) != ALLOC_RET:
            continue
        base = ret - OFF_ALLOC_RET
        if d.rl(base + OFF_PROT) != PROT_LINK:
            continue
    except Exception:
        continue
    found = (a7, ret, base)
    print("FOUND c/zzz key_init AllocMem(220): hunk0_base=%08X (#%d)" % (base, i), flush=True)
    break

if not found:
    print("NOT FOUND", flush=True); sys.exit(1)

a7, ret, base = found
sb = d.sysbase(); tt = d.thistask()
keyinit_a6 = d.rl(a7 + 4)
keyarray_ptr = d.rl(keyinit_a6 + 8)
count = d.rl(keyinit_a6 + 0xC)
cap = dict(hunk0_base=base, thistask=tt, task_name=d.amiga_str(d.rl(tt + 10)),
           keyarray_ptr=keyarray_ptr, count=count,
           tc_ExceptCode_2A=d.rl(tt + 0x2A), tc_TrapCode_32=d.rl(tt + 0x32))
vec = d.rd(0, 0xC0)
cap["vec_pages"] = sorted(set("%02X" % vec[a+1] for a in range(8, 0xC0, 4)))
if 0 < count <= 0x40:
    ka = d.rd(keyarray_ptr, count * 4)
    cap["keyarray"] = ka.hex()
    cap["keyarray_words"] = ["%08X" % int.from_bytes(ka[i:i+4], "big") for i in range(0, len(ka), 4)]
else:
    cap["keyarray"] = "BAD COUNT %d" % count
json.dump(cap, open("/tmp/keyarray.json", "w"), indent=2)
print("count=%d keyarray_ptr=%08X tc_Except=%08X tc_Trap=%08X" %
      (count, keyarray_ptr, cap["tc_ExceptCode_2A"], cap["tc_TrapCode_32"]), flush=True)
print("keyarray:", cap.get("keyarray_words"), flush=True)
print("vec pages:", cap["vec_pages"], flush=True)
