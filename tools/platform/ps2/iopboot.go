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

// iopBootOrder is the dependency order of the IOP kernel modules in IOPRP221.IMG.
//
//	TIMEMANI, SIFMAN     import nothing but intrman, which is ours
//	THREADMAN            wants timrman and heaplib
//	SIFCMD               wants sifman and threadman's thbase/thevent
//	IOMAN                wants only ours
//	ROMDRV, MODLOAD      want ioman
//	CDVDMAN              wants ioman and threadman
//	CDVDFSV              wants cdvdman and sifcmd
//	LOADFILE             wants modload and sifcmd
//	FILEIO               wants ioman and sifcmd
//	EESYNC               wants sifman and ioman
var iopBootOrder = []string{
	"TIMEMANI",
	"SIFMAN",
	"THREADMAN",
	"SIFCMD",
	"IOMAN",
	"ROMDRV",
	"MODLOAD",
	"CDVDMAN",
	"CDVDFSV",
	"LOADFILE",
	"FILEIO",
	"EESYNC",
}

// iopBootImage is the boot image the game reboots the IOP onto.
const iopBootImage = "/DRIVERS/IOPRP221.IMG"

// RebootIOP builds the second processor and starts it on the disc's boot image.
//
// It returns the modules it started, in the order it started them. A module that fails
// to load stops the boot then and there and says why: the alternative is an IOP that
// runs with a hole in it, which looks exactly like an IOP that is merely slow.
func (m *Machine) RebootIOP() error { return m.RebootIOPFrom(iopBootImage) }

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
	raw, err := m.vol.ReadFile(image)
	if err != nil {
		return fmt.Errorf("ps2: reading the IOP's boot image %s: %w", image, err)
	}
	entries, err := ROMDIRModules(raw)
	if err != nil {
		return fmt.Errorf("ps2: %s is not a ROMDIR archive: %w", image, err)
	}

	byName := map[string][]byte{}
	for _, e := range entries {
		byName[e.Name] = e.Data
	}

	// The EE's half of the SIF is up before the IOP's is: a game does not reboot the
	// second processor until its own side can talk to it. On a retail machine the EE's
	// BIOS kernel raises this bit; here the EE's kernel is Go, so this is where it gets
	// raised — and *which* bit was read out of SIFMAN, which is the module that waits
	// for it (sifbus.go).
	//
	// Without it the IOP boots as far as SIFCMD and stops there forever, in a loop four
	// instructions long.
	m.sbusSetFlag(sbusMSFLG, sifEESIFReady)

	m.StartIOP()
	for _, name := range iopBootOrder {
		raw, ok := byName[name]
		if !ok {
			return fmt.Errorf("ps2: %s holds no module called %s", image, name)
		}
		if err := m.IOP.LoadAndStart(name, raw); err != nil {
			return err
		}
	}
	m.note("IOP: booted on %s — %d modules", image, len(m.IOP.modules))

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
