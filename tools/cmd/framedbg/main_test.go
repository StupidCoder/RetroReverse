package main

import (
	"testing"

	"retroreverse.com/tools/debug"
)

// TestEveryAdapterHasAPlatformName belongs here rather than in the debug package, for the
// same reason the registry does: debug cannot import the adapters (they import it), so its
// own tests only ever see the fakes they register. THIS binary is the one place where every
// adapter is actually linked in, so it is the only place that can ask whether the set is
// complete.
//
// The library menu groups games by platform and labels each group with the human name. A
// platform with no name falls back to its tag — which renders fine, and renders exactly the
// "gc" that the names exist to replace, so nothing would look broken.
func TestEveryAdapterHasAPlatformName(t *testing.T) {
	ps := debug.Platforms()
	if len(ps) < 5 {
		t.Fatalf("only %d adapters are linked into framedbg (%v); an import was dropped", len(ps), ps)
	}
	for _, p := range ps {
		if n := debug.PlatformName(p); n == p {
			t.Errorf("platform %q has an adapter but no human name — add it to platformNames "+
				"in tools/debug/registry.go", p)
		}
	}
}
