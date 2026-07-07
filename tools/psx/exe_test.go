package psx

import "testing"

func TestParseEXE(t *testing.T) {
	v := loadDisc(t)
	data, err := v.ReadFile("SCUS-943.00;1")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	e, err := ParseEXE(data)
	if err != nil {
		t.Fatalf("ParseEXE: %v", err)
	}
	if e.PC0 != 0x80040004 {
		t.Errorf("PC0 = 0x%08X, want 0x80040004", e.PC0)
	}
	if e.TAddr != 0x80010000 {
		t.Errorf("TAddr = 0x%08X, want 0x80010000", e.TAddr)
	}
	if int(e.TSize) != len(e.Text) {
		t.Errorf("TSize %d != len(Text) %d", e.TSize, len(e.Text))
	}
	if e.PC0 < e.TAddr || e.PC0 >= e.TAddr+e.TSize {
		t.Errorf("entry PC 0x%08X outside text [0x%08X,0x%08X)", e.PC0, e.TAddr, e.TAddr+e.TSize)
	}
}

func TestParseEXERejectsGarbage(t *testing.T) {
	if _, err := ParseEXE(make([]byte, 0x800)); err == nil {
		t.Error("expected error for missing magic")
	}
	if _, err := ParseEXE([]byte("PS-X EXE")); err == nil {
		t.Error("expected error for truncated header")
	}
}
