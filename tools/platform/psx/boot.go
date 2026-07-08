package psx

import (
	"fmt"
	"strings"
)

// BootName reads SYSTEM.CNF and returns the boot executable's filename (the
// argument of the "BOOT = cdrom:NAME" line), stripped of the cdrom: prefix and
// any leading path separator.
func (v *Volume) BootName() (string, error) {
	cnf, err := v.ReadFile("SYSTEM.CNF")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(cnf), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "BOOT") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		name := strings.TrimSpace(line[eq+1:])
		name = strings.TrimPrefix(name, "cdrom:")
		name = strings.TrimPrefix(name, "cdrom0:")
		name = strings.TrimLeft(name, "\\/")
		return name, nil
	}
	return "", fmt.Errorf("psx: no BOOT line in SYSTEM.CNF")
}

// BootEXE resolves and parses the disc's boot executable.
func (v *Volume) BootEXE() (name string, exe *EXE, err error) {
	name, err = v.BootName()
	if err != nil {
		return "", nil, err
	}
	data, err := v.ReadFile(name)
	if err != nil {
		return name, nil, err
	}
	exe, err = ParseEXE(data)
	return name, exe, err
}
