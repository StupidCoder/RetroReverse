# Inspecting a real Giana disk transfer

C5 supplies the isolated core's dual-machine debugger and a repeatable loader
investigation. Browser panels and portable knowledge/tour integration are C6.
The example below uses the exact G64 and firmware identities in
[the compatibility ledger](COMPATIBILITY.md), not arbitrary releases of Giana.

## Reproduce the investigation

Run the optional disk acceptance command from [README.md](README.md). It now
builds and runs `giana-native`, `giana-ubsan` and `giana-wasm.mjs`, in addition to
the ordinary DOS and media-free checks. The Giana test cold-boots both machines
and types `LOAD"LADER",8,1` using the keyboard matrix. Neither executable bytes
nor expected payloads are injected into either machine.

`tests/giana.cpp` performs the following sequence:

1. Stop the drive before `$046C` executes. DOS has received the game's `M-W`
   commands and uploaded 512 bytes from C64 `$CDEE–$CFED` to drive `$0400–$05FF`.
   Compare the live source and destination; their FNV-1a is `440266918` in this
   pinned run. The source includes the game-prepared tail beyond file `F`.
2. Inspect `LDA #$C8` at `$046C`. While paused, request execution and verify that
   neither CPU nor disk position moves. Step the drive once: A becomes `$C8`,
   one instruction retires, and the C64 advances by the corresponding clocks.
3. Stop at drive `$04DF`, `TAX`, immediately before a byte sender. `$04E0` is
   `BIT $1800`, polling the serial VIA. Capture a whole-system checkpoint here.
4. Run 1,000 clock edges and inspect the timestamped IEC transitions. Restore,
   run the same edges again and require identical transitions and both RAMs.
   The harness prints the first eight transitions for direct inspection.
5. Restore again and follow the transfer. At drive `$04DF`, A contains the next
   byte to send. At C64 `$CDED`, A contains the byte reconstructed by the receiver.
   Check all 20,168 protocol bytes in order, including framing and load address.
6. The payload stores at C64 `$CD06`/`$CD73` produce 20,086 bytes at `$0801`.
   Require FNV-1a `1795265874`, independently derived from file `2` by the
   checksum-validating GCR oracle. Finish at the real unpacker entry `$0810`.

These addresses describe the observed code in this release. They are test
observations and breakpoint locations, never emulator traps or dispatch rules.
A longer exploratory native run reaches the release's animated Time Warp intro;
full gameplay and later disk accesses are not C5's automated acceptance claim.

## What the transfer demonstrates

The drive sends two bits at a time using CLK and DATA. Its code writes serial
VIA `$1800` four times per data byte, with eight drive clocks between those
writes. The C64's receiver samples CIA2 `$DD00` four times and combines lookup
results at `$CF00`, `$CF08`, `$CF10` and `$CF18`. The two CPUs run at different
frequencies, so their relative phase changes continuously.

C4 read the latest wire levels at CPU-cycle completion. This caused intermittent
corruption in the unmodified custom loader even though normal DOS reads passed.
C5 records each port's input at its PHI2 rising read phase and preserves that
sample until consumption. The MOS [6526 read timing diagram, page 4](https://www.emuverse.ru/downloads/datasheets/peripherals/PIA/6526/mos_6526_cia.pdf)
and [6522 Figure 21, catalog page 2-64](https://www.retrodocs.fr/wp-content/uploads/pdf/MOS-6522.pdf)
place peripheral-input setup relative to the rising clock. Focused tests exercise
changes from both machines between that sample and the later CPU read.
This is a digital phase model, not a simulation of cable rise times, gate delays
or every chip revision's setup/hold behavior.

## Core API

```cpp
System machine; // Supply firmware/media and boot as described in README.md.
Debugger debug(machine, 4096);
debug.breakpoint(Processor::Drive, 0x046c);
debug.resume();
auto result = debug.run(40000000); // Budget is combined CPU clock edges.
// Check result.reason; a budget expiry is not evidence that a breakpoint hit.
auto code = debug.disassemble(Processor::Drive, result.address);
auto registers = debug.cpu(Processor::Drive);
auto byte = debug.peek(Processor::Drive, 0x1800); // No acknowledgement side effect.
const auto& state = debug.drive(); // Both VIAs, RAM, head/bit, motor/LED, media.
auto checkpoint = machine.save();
auto step = debug.step(Processor::Drive);
debug.restore(checkpoint); // Pauses and clears old trace/breakpoint bypass state.
```

Use a persistent debugger to control all execution. Calling `System::run` or
`tickEdge` directly bypasses its pause/breakpoint/trace policy; those are low-level
hardware APIs for hosts and tests. `Debugger` is synchronous and single-threaded;
callers yield between bounded runs to process user pause/cancel requests.

Breakpoints identify a processor and pending opcode-fetch address. They are
checked before that processor's next clock, including the second of consecutive
drive clocks. A fetch held by VIC RDY can be a breakpoint; it has not executed.
`resume()` bypasses the current breakpoint for one visit, then re-arms it.
`step()` completes the selected processor's next retirement (or its current
partial instruction), advances all other hardware in time order, and pauses.
A peer breakpoint or fault can interrupt a step. Reset/interrupt entry does not
count as an instruction retirement. `Budget` during a step leaves the machine
paused mid-instruction; normal `run` budget exhaustion leaves it runnable.

Disassembly reads current memory on every request, including self-modified code,
and reports bytes, length, opcode, addressing mode and text. Peeking at VIA
registers must not clear IFR bits or trigger handshakes. `DriveState::lastBus`
records the latest serviced drive CPU address, value, direction and SYNC.

IEC events contain integer time, source clock, changed/level bit masks, each
participant's pull-low mask, both post-edge PCs, half-track and bit position.
Bits 0/1/2 mean ATN/CLK/DATA; actor 0 is C64 and actor 1 is drive. Time units are
`1 / (985248 * 1000000)` seconds. The bounded queue reports overwritten events
through `droppedEvents()`; zero capacity disables capture. Consume or clear it
between runs. Snapshots include both port samples and their phase markers;
restoring clears external trace history but retains configured breakpoints.
