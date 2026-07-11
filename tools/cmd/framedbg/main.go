// framedbg is a headless driver for the frame debugger's core: it advances the
// N64 oracle to a frame, then reports on how that frame was drawn — the same
// pipeline the ImGui GUI will present interactively, but verifiable from a script.
//
// It exists to prove (and regression-guard) the debug adapter end to end without a
// window: dump the captured RDP command stream, write the draw target as it stood
// after each of a range of commands (the command scrubber, as PNGs), and answer
// "which command drew this pixel?" for a chosen pixel (the pixel picker), with its
// full overdraw history.
//
// Usage:
//
//	framedbg -image ROM [-state FILE] [-skip N]
//	         [-list] [-listmax M]
//	         [-scrub STEP] [-o DIR]
//	         [-pixel X,Y]
//
// With no report flag it writes the finished frame and the VI scanout as PNGs.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n64adapter"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "framedbg:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		image_  = flag.String("image", "", "path to the N64 ROM (.z64) — required")
		state   = flag.String("state", "", "savestate file to load before stepping (skips the boot)")
		skip    = flag.Int("skip", -1, "advance this many frames before capturing; -1 = step until a drawn frame")
		list    = flag.Bool("list", false, "print the captured frame's RDP command stream")
		listmax = flag.Int("listmax", 0, "cap -list to the first M commands (0 = all)")
		scrub   = flag.Int("scrub", 0, "write the draw target after every Nth command as a PNG (0 = off)")
		out     = flag.String("o", ".", "output directory for PNGs")
		pixel   = flag.String("pixel", "", "report which command drew pixel X,Y (e.g. -pixel 160,120)")
	)
	flag.Parse()
	if *image_ == "" {
		flag.Usage()
		return fmt.Errorf("-image is required")
	}

	a, err := n64adapter.New(*image_)
	if err != nil {
		return err
	}
	if *state != "" {
		if err := a.LoadStateFile(*state); err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
	}

	withOverdraw := *pixel != ""
	fc, err := advance(a, *skip, withOverdraw)
	if err != nil {
		return err
	}
	fmt.Printf("captured frame: %d RDP commands, %dx%d draw target\n",
		len(fc.Commands), fc.Width, fc.Height)

	if *list {
		printCommands(fc, *listmax)
	}
	if *pixel != "" {
		if err := reportPixel(fc, *pixel); err != nil {
			return err
		}
	}
	if *scrub > 0 {
		if err := writeScrub(a, fc, *scrub, *out); err != nil {
			return err
		}
	}
	if !*list && *pixel == "" && *scrub == 0 {
		return writeFinalFrames(a, fc, *out)
	}
	return nil
}

// advance steps the machine to a drawn frame. It first advances skip video fields
// (skip<=0 advances none), then steps until a frame draws a real scene — so the
// captured frame is always one worth inspecting, however far in you jump.
func advance(a *n64adapter.Adapter, skip int, withOverdraw bool) (*debug.FrameCapture, error) {
	for i := 0; i < skip; i++ {
		if _, err := a.StepFrame(false); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 800; i++ {
		fc, err := a.StepFrame(withOverdraw)
		if err != nil {
			return nil, err
		}
		if len(fc.Commands) > 100 && fc.Prov != nil {
			return fc, nil
		}
	}
	return nil, fmt.Errorf("no drawn frame within the field budget (try a different -skip or -state)")
}

func printCommands(fc *debug.FrameCapture, max int) {
	n := len(fc.Commands)
	if max > 0 && max < n {
		n = max
	}
	for _, c := range fc.Commands[:n] {
		var w0 uint64
		if len(c.Words) > 0 {
			w0 = c.Words[0]
		}
		fmt.Printf("  %5d  %-20s op=%#02x  %016x\n", c.Index, c.Name, c.Op, w0)
	}
	if n < len(fc.Commands) {
		fmt.Printf("  ... %d more\n", len(fc.Commands)-n)
	}
}

func reportPixel(fc *debug.FrameCapture, spec string) error {
	x, y, err := parseXY(spec)
	if err != nil {
		return err
	}
	pc := fc.ProvAt(x, y)
	if pc < 0 {
		fmt.Printf("pixel (%d,%d): no command wrote it\n", x, y)
		return nil
	}
	fmt.Printf("pixel (%d,%d): last written by command %d %q\n", x, y, pc, fc.Commands[pc].Name)
	if fc.Overdraw != nil {
		if writes := fc.Overdraw[y*fc.Width+x]; len(writes) > 0 {
			fmt.Printf("  overdraw history (%d writes):\n", len(writes))
			for _, w := range writes {
				tag := "drawn"
				if w.Rejected {
					tag = "rejected"
				}
				fmt.Printf("    cmd %5d %-20s rgba=%02x%02x%02x%02x  %s\n",
					w.CmdIndex, fc.Commands[w.CmdIndex].Name, w.R, w.G, w.B, w.A, tag)
			}
		}
	}
	return nil
}

// writeScrub writes the draw target after commands 0, step, 2*step, ... and the
// last command — the command scrubber, frame by frame, as PNGs.
func writeScrub(a *n64adapter.Adapter, fc *debug.FrameCapture, step int, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ks := []int{}
	for k := 0; k < len(fc.Commands); k += step {
		ks = append(ks, k)
	}
	if last := len(fc.Commands) - 1; len(ks) == 0 || ks[len(ks)-1] != last {
		ks = append(ks, last)
	}
	for _, k := range ks {
		img, err := a.RenderAfter(fc, k)
		if err != nil {
			return fmt.Errorf("RenderAfter(%d): %w", k, err)
		}
		path := filepath.Join(dir, fmt.Sprintf("cmd-%05d.png", k))
		if err := writePNG(path, img); err != nil {
			return err
		}
	}
	fmt.Printf("wrote %d scrub frames to %s\n", len(ks), dir)
	return nil
}

func writeFinalFrames(a *n64adapter.Adapter, fc *debug.FrameCapture, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	final, err := a.RenderAfter(fc, len(fc.Commands)-1)
	if err != nil {
		return fmt.Errorf("rendering finished frame: %w", err)
	}
	if err := writePNG(filepath.Join(dir, "frame.png"), final); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", filepath.Join(dir, "frame.png"))
	if disp, err := a.Display(); err == nil {
		if err := writePNG(filepath.Join(dir, "scanout.png"), disp); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", filepath.Join(dir, "scanout.png"))
	}
	return nil
}

func parseXY(spec string) (int, int, error) {
	parts := strings.Split(spec, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad -pixel %q, want X,Y", spec)
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("bad -pixel X: %w", err)
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("bad -pixel Y: %w", err)
	}
	return x, y, nil
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
