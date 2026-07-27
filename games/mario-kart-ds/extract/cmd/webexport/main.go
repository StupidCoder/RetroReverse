// webexport builds Mario Kart DS's Retro-X game tree from the raw cartridge.
// It stages the filesystem in a temp dir, renders the SDAT music, exports
// every NSBMD model to GLB and writes each track as a scene3d level:
//
//	levels/<stem>.json           per course: the course model placed at the
//	                             origin (so its BTA0 texture animations play),
//	                             the "_V" skybox as a camera-attached layer,
//	                             the CPU drive line as a toggleable line
//	                             layer, and every OBJI map object at its
//	                             authored transform, movers on their NKM routes
//	levels/<sky|driveline>.glb   the level-only geometry
//	objects/<id>.json|.glb       characters, karts, per-course map objects,
//	                             the shared itembox, and the course scenes
//	music/seq_NN.mp3             every renderable SSEQ, via tools/nds/sdat
//
// Usage (from games/mario-kart-ds/):
//
//	go run ./extract/cmd/webexport -in "Mario Kart DS (Europe) (En,Fr,De,Es,It).nds"
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/nds"
)

func main() {
	cli.Main("mario-kart-ds", run)
}

func run(ctx *cli.Context) error {
	if ctx.In == "" {
		return fmt.Errorf("usage: webexport -in <rom.nds> [-o DIR] [-only levels,objects,music]")
	}
	img, err := os.ReadFile(ctx.In)
	if err != nil {
		return err
	}
	rom, err := nds.Open(img)
	if err != nil {
		return err
	}

	b := ctx.Builder
	b.SetTitle("Mario Kart DS")
	b.SetPlatform("Nintendo DS")
	b.SetYear(2005)
	b.SetDisplay(schema.Display{
		Native: schema.Size{W: 256, H: 192},
		TickHz: 60,
		// The DS renders unfiltered texels; keep them point-sampled.
		TexFilter: "nearest",
	})

	if ctx.Stage("music") {
		if err := runMusic(ctx, rom); err != nil {
			return err
		}
	}

	if ctx.Enabled("objects") || ctx.Enabled("levels") {
		// Stage the filesystem the NSBMD/NKM decoders read from disk.
		tmp, err := os.MkdirTemp("", "mkds-webexport-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		if err := extractFS(rom, tmp); err != nil {
			return err
		}

		// The object-ID → model-name bindings from the ARM9's map-object
		// descriptor table (Part V §2) drive the OBJI placements.
		arm9 := rom.ARM9()
		if nds.IsBLZ(arm9) {
			arm9 = nds.DecompressBLZ(arm9)
		}
		bindings = objectBindings(arm9)
		ctx.Logf("%d object-model bindings from ARM9", len(bindings))

		items, err := exportAllGLBs(ctx, filepath.Join(tmp, "files"))
		if err != nil {
			return err
		}
		var refs map[string]string
		if ctx.Stage("objects") {
			if refs, err = buildObjects(ctx, items); err != nil {
				return err
			}
		}
		if ctx.Stage("levels") {
			if err := buildLevels(ctx, items, refs); err != nil {
				return err
			}
		}
	}
	return nil
}

// extractFS writes the full filesystem under dir/files/, which the NSBMD/NKM
// decoders (mkds.LoadModels/LoadTextures/LoadNKM) read from disk.
func extractFS(rom *nds.ROM, dir string) error {
	for _, f := range rom.Files {
		p := filepath.Join(dir, "files", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, rom.File(f.ID), 0o644); err != nil {
			return err
		}
	}
	return nil
}
