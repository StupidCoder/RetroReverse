// Synthetic 6510 fixture: no game bytes. Tests the existing ABI, not the future debugger.
export function probeC64(core) {
  const check = (v, text) => { if (!v) throw Error(text); };
  const status = () => JSON.parse(core.UTF8ToString(core._rr_status()));
  const input = core._rr_input();
  core.HEAPU8.fill(0xea, input, input + 20480);
  check(core._rr_init(8192, 8192, 4096), 'init');
  const tape = new Uint8Array(21);
  tape.set(new TextEncoder().encode('C64-TAPE-RAW'));
  tape[16] = 1; tape[20] = 48;
  core.HEAPU8.set(tape, core._rr_input());
  check(core._rr_tape(tape.length), 'tape');
  const ram = new Uint8Array(65536);
  // INC $0200; JMP $0800. Initial opcode is harmless until actually executed.
  ram.set([0xee, 0x00, 0x02, 0x4c, 0x00, 0x08], 0x0800);
  core.HEAPU8.set(ram, core._rr_input());
  check(core._rr_prepare(0x0800, 0), 'prepare');
  const before = status();
  const ran = core._rr_run(100, 6, 0);
  const after = status(), nextBusPC = core._rr_bus() & 65535;
  check(ran > 0 && core._rr_stop_reason() === 6, 'next opcode stop');
  check(core.HEAPU8[core._rr_ram() + 0x200] === 1, 'exactly one INC');
  check(nextBusPC === 0x0803, 'next opcode bus address');
  const boundary = {beforePC: before.pc, statusPC: after.pc, nextBusPC,
    flags: core._rr_bus_flags(), cycles: after.cycle - before.cycle,
    statusIsNextPC: after.pc === nextBusPC};
  // Check PC stop does not execute target INC.
  const value = core.HEAPU8[core._rr_ram() + 0x200];
  check(core._rr_run(100, 2, 0x0800) > 0, 'PC run');
  check(core._rr_stop_reason() === 2 && (core._rr_bus() & 65535) === 0x0800, 'PC stop');
  check(core.HEAPU8[core._rr_ram() + 0x200] === value, 'stop before target writes');
  const n = core._rr_state_save(); check(n > 0, 'save');
  const saved = core.HEAPU8.slice(core._rr_state_data(), core._rr_state_data() + n);
  const restore = () => {core.HEAPU8.set(saved, core._rr_state_input(n));check(core._rr_state_load(n), 'restore');};
  const samples = [];
  for (let k = 0; k < 4; k++) {
    restore(); const t = performance.now();
    for (let i = 0; i < 100; i++) check(core._rr_run(10000, 0, 0) === 10000, 'run budget');
    const ms = performance.now() - t;
    if (k) samples.push({cycles: 1000000, ms, cyclesPerSecond: 1e9 / ms});
  }
  restore(); const t = performance.now();
  for (let i = 0; i < 1000; i++) {check(core._rr_run(100, 6, 0) > 0, 'step');status();}
  return {fixture: 'synthetic INC/JMP at 0800, dummy ROM, display disabled',
    boundary, checkpointBytes: n, wasmHeapBytes: core.HEAPU8.length,
    throughput: samples, instructionStepAndStatus: {count: 1000, totalMs: performance.now() - t},
    limitations: 'Core-only synthetic workload. Status hashes RAM/framebuffer. No disassembly, DOM or real gameplay in this measurement.'};
}
