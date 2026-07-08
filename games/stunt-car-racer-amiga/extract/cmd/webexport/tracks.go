// tracks.go holds the circuit name table. The per-circuit geometry is baked straight into the
// GLBs by models.go (from package track's verified spine + $65BEC model); the old per-circuit
// tracks/<slug>.json export (consumed by the retired stunt-track wireframe viewer) is gone —
// every asset is now a GLB under manifest.models[]. cmd/trackjson still emits the standalone
// track JSON if the decoded geometry is wanted outside the site.
package main

// trackNames is the eight circuits in engine order (trackid 0..7). Source: the game's
// own circuit list (Part IV); names are unique so their slugs are stable file stems.
var trackNames = []string{
	"Little Ramp", "Stepping Stones", "Hump Back", "Big Ramp",
	"Ski Jump", "Draw Bridge", "High Jump", "Roller Coaster",
}
