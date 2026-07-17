package ps2

// iopboot.go brings the second processor up on the modules the disc carries.
//
// A retail PS2 boots the IOP twice. The BIOS starts it on the modules in its own ROM,
// and then the game — wanting a different set, and its own — reboots it, handing it a
// boot image to come up on. Jak's is IOPRP221.IMG, and it is on the disc, so the
// second boot is one we can do properly even though the first is one we cannot do at
// all. That is the whole trick: we skip the boot we have no ROM for and perform the
// one the game actually cares about.
//
// The order below is not the order the archive lists the modules in. The archive is a
// directory, not a schedule; a module has to be loaded after everything it imports,
// because the loader wires an import straight to the exporting module's code and
// cannot wire it to a module that is not there yet. So the order is a topological sort
// of the import graph — derived from the modules' own import tables, which irxinfo
// prints — and if it is ever wrong, LoadIRX says which module wanted what, rather than
// leaving a stub that will fail three million instructions later.

import (
	"fmt"
	"strings"
)

// iopBootOrder is the order the IOP kernel modules are started in — the subset of the
// modules we load for real (the base libraries beneath them are HLE'd; see iopkernel.go).
//
// This is the CONSOLE ROM's own order, read from its IOPBTCONF and filtered to our set.
// It was previously a topological sort of the import graph, derived by hand — a valid
// order, and it booted Jak, but a *different* order from the one the machine actually
// uses (it loaded SIFMAN before THREADMAN, and CDVDMAN before LOADFILE, where the ROM
// does the reverse). Both are valid topological sorts of the same dependency DAG, so
// nothing forced a choice and the hand-derived one drifted. Two ROM revisions four
// years apart list this exact order, so it is the machine's authority and not a guess;
// TestBootOrderMatchesBIOS pins ours to it. The dependencies each module still relies on:
//
//	TIMEMANI             imports nothing but intrman, which is ours
//	THREADMAN            wants timrman and heaplib
//	IOMAN                wants only ours
//	MODLOAD, ROMDRV      want ioman
//	SIFMAN               imports nothing but intrman
//	SIFCMD               wants sifman and threadman's thbase/thevent
//	LOADFILE             wants modload and sifcmd
//	CDVDMAN              wants ioman and threadman
//	CDVDFSV              wants cdvdman and sifcmd
//	FILEIO               wants ioman and sifcmd
//	EESYNC               wants sifman and ioman
var iopBootOrder = []string{
	"TIMEMANI",
	"THREADMAN",
	"IOMAN",
	"MODLOAD",
	"ROMDRV",
	"SIFMAN",
	"SIFCMD",
	"LOADFILE",
	"CDVDMAN",
	"CDVDFSV",
	"FILEIO",
	"EESYNC",
}

// iopBootImage is the boot image the game reboots the IOP onto.
const iopBootImage = "/DRIVERS/IOPRP221.IMG"

// RebootIOP brings the second processor up at power-on, before the EE runs.
//
// On real hardware this first boot comes from the console ROM alone: the BIOS starts
// the IOP on rom0's own modules, and only later does the game reboot it onto a disc
// image. So when a ROM is supplied, that is exactly what happens — the disc image is
// not read yet; the game's own `rom0:UDNL <image>` reboot brings it in and overlays
// it (iopRebootImage reads the path out of the game's packet).
//
// When no ROM is supplied we have no rom0 to boot from, so we fall back to the one
// boot the disc can do on its own: a self-contained IOPRP image. Jak's carries all
// twelve kernel modules, which is the whole reason this machine ran without a ROM.
// A module that fails to load stops the boot then and there and says why.
func (m *Machine) RebootIOP() error {
	if m.bios != nil {
		return m.RebootIOPFrom("") // power-on from rom0; the game's UDNL reboot overlays the disc
	}
	return m.RebootIOPFrom(iopBootImage)
}

// RebootIOPFrom boots the second processor on a nominated image.
//
// The image is nominated by the *game*, and that is the point of this being a parameter. The
// EE asks for the reboot by sending the IOP a command packet, and the packet carries the
// request as a string — "rom0:UDNL cdrom0:\DRIVERS\IOPRP221.IMG;1". So the boot image is not
// something this machine has to know in advance; it is something the game says out loud, and
// reading it out of the request rather than out of a constant is the difference between
// running the game's boot and reproducing it from memory.
func (m *Machine) RebootIOPFrom(image string) error {
	if m.vol == nil {
		return fmt.Errorf("ps2: no disc is mounted, so the IOP has nothing to boot from")
	}
	// The modules come from two places, in this order: the console ROM's base set,
	// then the game's boot image on top. The image is an *update* — it carries only
	// the modules the game replaces, and `rom0:UDNL <image>` is a request to boot the
	// ROM's set and update it with the image's, so a name in the image wins. When no
	// ROM was supplied the image must be self-contained (Jak's is); when it is not,
	// the missing-module error below says which name had no source.
	byName := map[string][]byte{}
	if m.bios != nil {
		entries, err := ROMDIRModules(m.bios)
		if err != nil {
			return fmt.Errorf("ps2: the supplied BIOS is not a ROMDIR image: %w", err)
		}
		for _, e := range entries {
			byName[e.Name] = e.Data
		}
	}
	// An empty image name is a ROM-only boot (the power-on boot when a ROM is present):
	// the disc is not read at all. Only the ROM can serve that, so it is an error to ask
	// for it without one.
	if image == "" {
		if m.bios == nil {
			return fmt.Errorf("ps2: a ROM-only boot was requested but no BIOS is mounted")
		}
	} else {
		raw, err := m.vol.ReadFile(image)
		if err != nil {
			return fmt.Errorf("ps2: reading the IOP's boot image %s: %w", image, err)
		}
		entries, err := ROMDIRModules(raw)
		if err != nil {
			return fmt.Errorf("ps2: %s is not a ROMDIR archive: %w", image, err)
		}
		for _, e := range entries {
			byName[e.Name] = e.Data
		}
	}

	// The EE's half of the SIF is up before the IOP's is: a game does not reboot the
	// second processor until its own side can talk to it. On a retail machine the EE's
	// BIOS kernel raises this bit; here the EE's kernel is Go, so this is where it gets
	// raised — and *which* bit was read out of SIFMAN, which is the module that waits
	// for it (sifbus.go).
	//
	// Without it the IOP boots as far as SIFCMD and stops there forever, in a loop four
	// instructions long.
	m.sbusFlagSet(sbusMSFLG, sifEESIFReady)

	m.StartIOP()
	for _, name := range iopBootOrder {
		raw, ok := byName[name]
		if !ok {
			// A module the boot order calls for that neither the ROM nor the image
			// supplies. Without a ROM this is the usual "the image is not
			// self-contained" case; with one it means the ROM is missing a base
			// module, which is a different and louder problem.
			if m.bios == nil {
				return fmt.Errorf("ps2: %s holds no module called %s, and no BIOS was supplied to fall back on", image, name)
			}
			return fmt.Errorf("ps2: neither %s nor the BIOS holds a module called %s", image, name)
		}
		if err := m.IOP.LoadAndStart(name, raw); err != nil {
			return err
		}
	}
	switch {
	case image == "":
		m.note("IOP: power-on boot from ROM (%d modules)", len(m.IOP.modules))
	case m.bios != nil:
		m.note("IOP: booted on ROM + %s (%d modules)", image, len(m.IOP.modules))
	default:
		m.note("IOP: booted on %s (%d modules, image only)", image, len(m.IOP.modules))
	}

	// Every module is loaded, so the boot is over — and the modules that asked to be told
	// that are now told (loadcore#20, iopkernel.go). It is EESYNC that cares: its callback
	// raises the flag that says the reboot has finished, and the EE is spinning on it.
	m.IOP.runBootCallbacks()

	// Anything the harness wants written into the second processor's memory once its modules
	// are in place. Sony's modules carry their own tracing, behind a verbosity level that a
	// retail boot leaves at zero — CDVDMAN's is a word it tests before every message it has —
	// and turning one on is the most direct instrument there is for a module we cannot read the
	// source of: it stops being a black box and starts narrating. It has to happen here, after
	// the modules have landed and before any of their threads run.
	for addr, val := range m.IOPPokes {
		m.IOP.Write32(addr, val)
		m.note("IOP: poked 0x%08X = 0x%08X (%s)", addr, val, m.IOP.Sym(addr))
	}

	// And now hand the processor over to its own scheduler.
	//
	// This is the last act of the loader, and without it the second processor never runs a
	// single one of its threads. Every module's entry point has been called and has returned;
	// each one created the threads it wanted and started them, and they are all sitting ready.
	// But nothing has *asked* THREADMAN to run them. The machine is parked in the idle loop the
	// kernel HLE keeps at 0x200, which is not a thread and which THREADMAN has never heard of,
	// and THREADMAN goes on believing that the thread it last saw running is still running —
	// so its predicate says no switch is wanted, and on the way out of every interrupt it
	// declines to schedule anybody. The profile is unambiguous about it: ninety-five per cent
	// of the IOP's life at 0x00000204, and zero thread switches in a boot with twelve modules
	// and a dozen threads in it.
	//
	// On the board there is no such gap, because the code that loads the modules is itself a
	// thread: when it has finished, it blocks or it exits, and the scheduler picks up the next
	// one as a matter of course. Here the loader is Go, and Go cannot block on a semaphore. So
	// the loader does what a thread does when it has finished: it exits.
	m.IOP.exitLoaderThread()
	return nil
}

// iopRebootImage reads the boot image out of the EE's reset request.
//
// The request is one string, and it names two things: the loader in the IOP's own ROM that is
// to perform the reboot, and the image it is to boot. `rom0:UDNL` is the first; the rest is
// the second, and it is a path on this disc in the game's own notation —
// `cdrom0:\DRIVERS\IOPRP221.IMG;1`, which is `/DRIVERS/IOPRP221.IMG` written the way the CD
// filesystem writes it. Nothing here is guessed at: the drive, the separators and the version
// suffix are all conventions of the disc, and the disc is what will be asked for the file.
//
// An unrecognised request is answered with an error rather than the image we happen to expect.
// A machine that boots the right file for the wrong reason is one that will boot the wrong
// file the moment the game asks for a different one — and the whole reason to read the name is
// that the game is the authority on it.
func iopRebootImage(cmd string) (string, error) {
	arg := cmd
	if i := strings.IndexByte(arg, ' '); i >= 0 {
		arg = arg[i+1:] // drop the ROM loader that is to do the rebooting
	}
	if !strings.HasPrefix(arg, "cdrom0:") {
		return "", fmt.Errorf("ps2: the EE asked the IOP to boot %q, which is not on the disc", cmd)
	}
	path := strings.TrimPrefix(arg, "cdrom0:")
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.IndexByte(path, ';'); i >= 0 {
		path = path[:i] // the ISO version suffix
	}
	return path, nil
}

// LoadModuleFromDisc loads an IRX off the disc by path and starts it. This is what the
// game asks for, once it is running: OVERLORD, 989SND, PADMAN and the rest all arrive
// this way.
func (p *IOP) LoadModuleFromDisc(path string) error {
	raw, err := p.ps2.vol.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ps2: reading %s: %w", path, err)
	}
	name := path
	if i := strings.LastIndexAny(name, "/\\"); i >= 0 {
		name = name[i+1:]
	}
	return p.LoadAndStart(strings.TrimSuffix(name, ";1"), raw)
}

// LoadAndStart places a module, links it and runs its entry point.
// runBootCallbacks calls the routines the modules registered with loadcore#20, in the order
// they registered them.
//
// It is the loader's last duty before it stands down, and the loader is the only thing that
// can do it: a boot callback exists precisely because the module registering it cannot see the
// end of the boot from where it stands. See loadcoreRegisterBootCallback for how the identity
// was earned, and what EESYNC's callback says to the EE.
func (p *IOP) runBootCallbacks() {
	for _, fn := range p.bootCallbacks {
		if _, err := p.callGuest(fn); err != nil {
			p.ps2.note("IOP: the boot callback at %s did not return: %v", p.Sym(fn), err)
			return
		}
		p.ps2.note("IOP: the boot is over, and %s was told so", p.Sym(fn))
	}
}

func (p *IOP) LoadAndStart(name string, raw []byte) error {
	mod, err := p.LoadIRX(name, raw)
	if err != nil {
		return err
	}
	if p.ps2.OnIOPModule != nil {
		p.ps2.OnIOPModule(p, name)
	}
	res, err := p.Start(mod)
	if err != nil {
		return err
	}
	p.ps2.note("IOP: %s loaded at 0x%08X (%d KiB), started -> %d", name, mod.Base, mod.Size/1024, int32(res))
	return nil
}
