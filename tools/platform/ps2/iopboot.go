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
func (m *Machine) RebootIOP() error {
	if m.vol == nil {
		return fmt.Errorf("ps2: no disc is mounted, so the IOP has nothing to boot from")
	}
	raw, err := m.vol.ReadFile(iopBootImage)
	if err != nil {
		return fmt.Errorf("ps2: reading the IOP's boot image: %w", err)
	}
	entries, err := ROMDIRModules(raw)
	if err != nil {
		return fmt.Errorf("ps2: %s is not a ROMDIR archive: %w", iopBootImage, err)
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
			return fmt.Errorf("ps2: %s holds no module called %s", iopBootImage, name)
		}
		if err := m.IOP.LoadAndStart(name, raw); err != nil {
			return err
		}
	}
	m.note("IOP: booted on %s — %d modules", iopBootImage, len(m.IOP.modules))
	return nil
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
	res, err := p.Start(mod)
	if err != nil {
		return err
	}
	p.ps2.note("IOP: %s loaded at 0x%08X (%d KiB), started -> %d", name, mod.Base, mod.Size/1024, int32(res))
	return nil
}
