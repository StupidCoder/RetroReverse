package psp

// These tests exercise the Part I image pipeline (CISO codec → ISO 9660 → PARAM.SFO)
// against a real UMD image. They skip when no image is available; point PSP_IMAGE at
// a .cso/.iso, or drop the test image at one of the default locations below.

import (
	"os"
	"sort"
	"testing"
)

func testImagePath(t *testing.T) string {
	if p := os.Getenv("PSP_IMAGE"); p != "" {
		return p
	}
	for _, p := range []string{
		"../../../games/loco-roco-psp/image/LocoRoco.cso",
		"../../../games/loco-roco-psp/LocoRoco.cso",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("no PSP image (set PSP_IMAGE)")
	return ""
}

func TestCSOAndISO(t *testing.T) {
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer im.Close()

	if im.System != "PSP GAME" {
		t.Errorf("system = %q, want %q", im.System, "PSP GAME")
	}
	// The boot files must be present.
	for _, path := range []string{
		"PSP_GAME/SYSDIR/EBOOT.BIN",
		"PSP_GAME/SYSDIR/BOOT.BIN",
		"PSP_GAME/PARAM.SFO",
	} {
		if _, err := im.resolve(path); err != nil {
			t.Errorf("resolve %s: %v", path, err)
		}
	}
	// EBOOT.BIN is a ~PSP encrypted container.
	eboot, err := im.ReadFile("PSP_GAME/SYSDIR/EBOOT.BIN")
	if err != nil {
		t.Fatal(err)
	}
	if string(eboot[0:4]) != "~PSP" {
		t.Errorf("EBOOT magic = % X, want ~PSP", eboot[0:4])
	}
}

func TestDecryptEBOOT(t *testing.T) {
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer im.Close()

	mod, err := im.LoadExecutable("PSP_GAME/SYSDIR/EBOOT.BIN")
	if err != nil {
		t.Fatal(err)
	}
	if !mod.Encrypted || mod.Tag != 0xC0CB167C {
		t.Errorf("tag = 0x%08X encrypted=%v, want 0xC0CB167C true", mod.Tag, mod.Encrypted)
	}
	if mod.Name != "LocoRoco" {
		t.Errorf("module name = %q, want LocoRoco", mod.Name)
	}
	if mod.EntryPC != 0x0003C500 {
		t.Errorf("entry = 0x%08X, want 0x0003C500", mod.EntryPC)
	}
	// The import list drives the kernel HLE; a correct decrypt yields the real one.
	if len(mod.Imports) != 29 {
		t.Errorf("imports = %d, want 29", len(mod.Imports))
	}
	haveThreadMan := false
	for _, imp := range mod.Imports {
		if imp.Library == "ThreadManForUser" {
			haveThreadMan = true
		}
	}
	if !haveThreadMan {
		t.Errorf("ThreadManForUser import missing (decrypt/parse wrong)")
	}
}

func TestSFO(t *testing.T) {
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer im.Close()

	data, err := im.ReadFile("PSP_GAME/PARAM.SFO")
	if err != nil {
		t.Fatal(err)
	}
	sfo, err := ParseSFO(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := sfo.String("TITLE"); got == "" {
		t.Errorf("TITLE missing")
	}
	if got := sfo.String("DISC_ID"); got == "" {
		t.Errorf("DISC_ID missing")
	}
}

func TestBoot(t *testing.T) {
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer im.Close()
	mod, err := im.LoadExecutable("PSP_GAME/SYSDIR/EBOOT.BIN")
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine()
	if err := m.LoadModule(mod); err != nil {
		t.Fatal(err)
	}
	res := m.Run(20_000_000)
	t.Logf("boot: %s", res)
	t.Logf("syscalls reached: %d distinct", len(m.SyscallCalls))
	type kv struct {
		name string
		n    int
	}
	var top []kv
	for k, v := range m.SyscallCalls {
		top = append(top, kv{k, v})
	}
	sort.Slice(top, func(i, j int) bool { return top[i].n > top[j].n })
	for i, e := range top {
		if i >= 20 {
			break
		}
		t.Logf("  %5d  %s", e.n, e.name)
	}
	for _, need := range []string{"sceKernelCreateThread", "sceKernelStartThread"} {
		if m.SyscallCalls[need] == 0 {
			t.Errorf("boot did not reach %s", need)
		}
	}
}

func TestSaveStateRoundTrip(t *testing.T) {
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer im.Close()
	mod, err := im.LoadExecutable("PSP_GAME/SYSDIR/EBOOT.BIN")
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine()
	m.SetImageHash("test-hash")
	m.LoadModule(mod)
	m.Run(500_000)

	path := t.TempDir() + "/s.state"
	if err := m.SaveStateFile(path); err != nil {
		t.Fatal(err)
	}
	// Restore into a fresh machine; the decoded state must match (the serialized bytes
	// need not — gob map ordering is nondeterministic).
	m2 := NewMachine()
	m2.SetImageHash("test-hash")
	if err := m2.LoadStateFile(path); err != nil {
		t.Fatal(err)
	}
	s1, s2 := m.SaveState(), m2.SaveState()
	if s1.CPU.PC != s2.CPU.PC || s1.CPU.Steps != s2.CPU.Steps {
		t.Errorf("restored PC/steps mismatch: %08X/%d vs %08X/%d", s2.CPU.PC, s2.CPU.Steps, s1.CPU.PC, s1.CPU.Steps)
	}
	if !bytesEqual(s1.RAM, s2.RAM) {
		t.Errorf("restored RAM differs")
	}
	if s1.CPU.F != s2.CPU.F || s1.CPU.V != s2.CPU.V {
		t.Errorf("restored FPU/VFPU register files differ")
	}
	// A mismatched image hash must be rejected.
	m3 := NewMachine()
	m3.SetImageHash("other-hash")
	if err := m3.LoadStateFile(path); err == nil {
		t.Errorf("load accepted a mismatched image hash")
	}
}
