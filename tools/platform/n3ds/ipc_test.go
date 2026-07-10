package n3ds

import "testing"

func TestParseIPCHeader(t *testing.T) {
	// GetServiceHandle: command 5, 4 normal params, 0 translate.
	h := parseIPCHeader(0x00050100)
	if h.Command != 5 || h.Normal != 4 || h.Translate != 0 {
		t.Fatalf("got cmd=%d normal=%d translate=%d", h.Command, h.Normal, h.Translate)
	}
	// A reply with 1 normal + 2 translate (a moved handle).
	h = parseIPCHeader(uint32(5)<<16 | 1<<6 | 2)
	if h.Command != 5 || h.Normal != 1 || h.Translate != 2 {
		t.Fatalf("got cmd=%d normal=%d translate=%d", h.Command, h.Normal, h.Translate)
	}
}

func TestKnownServiceAndBase(t *testing.T) {
	cases := map[string]bool{
		"APT:U": true, "gsp::Gsp": true, "cfg:u": true, "ndm:u": true,
		"fs:USER": true, "T:APc": false, "m:nd": false, "": false,
	}
	for name, want := range cases {
		if got := knownService(name); got != want {
			t.Errorf("knownService(%q) = %v, want %v", name, got, want)
		}
	}
	if serviceBase("gsp::Gsp") != "gsp" {
		t.Errorf("serviceBase(gsp::Gsp) = %q", serviceBase("gsp::Gsp"))
	}
}
