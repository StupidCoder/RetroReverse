// pspinfo is the Part I inspector for a PSP UMD image (a .cso CISO container or a
// flat .iso). It lists the ISO 9660 filesystem, dumps PARAM.SFO title metadata,
// decrypts and describes the boot executable, and extracts files — all directly
// from the raw image with no manual staging.
//
// Usage:
//
//	pspinfo -image LocoRoco.cso -ls
//	pspinfo -image LocoRoco.cso -sfo
//	pspinfo -image LocoRoco.cso -exe PSP_GAME/SYSDIR/EBOOT.BIN
//	pspinfo -image LocoRoco.cso -extract PSP_GAME/SYSDIR/PARAM.SFO -o param.sfo
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/psp"
)

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "pspinfo: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	image := flag.String("image", "", "UMD image (.cso or .iso)")
	ls := flag.Bool("ls", false, "list the ISO 9660 filesystem")
	sfo := flag.Bool("sfo", false, "dump PSP_GAME/PARAM.SFO title metadata")
	exe := flag.String("exe", "", "decrypt and describe an executable (path on the disc)")
	extract := flag.String("extract", "", "extract a file (path on the disc)")
	out := flag.String("o", "", "with -extract, output file")
	flag.Parse()

	if *image == "" {
		die("need -image")
	}
	im, err := psp.OpenImage(*image)
	if err != nil {
		die("%v", err)
	}
	defer im.Close()

	fmt.Fprintf(os.Stderr, "volume: system=%q name=%q\n", im.System, im.Name)

	switch {
	case *ls:
		if err := im.Walk(func(e psp.Entry) error {
			fmt.Println(e)
			return nil
		}); err != nil {
			die("%v", err)
		}

	case *sfo:
		dumpSFO(im)

	case *exe != "":
		raw, err := im.ReadFile(*exe)
		if err != nil {
			die("%v", err)
		}
		if len(raw) >= 4 && string(raw[0:4]) == "~PSP" && *out != "" {
			plain, tag, err := psp.DecryptPRX(raw)
			if err != nil {
				die("decrypt: %v", err)
			}
			if err := os.WriteFile(*out, plain, 0644); err != nil {
				die("%v", err)
			}
			fmt.Fprintf(os.Stderr, "decrypted tag 0x%08X -> %d bytes to %s\n", tag, len(plain), *out)
		}
		mod, err := psp.LoadModuleImage(raw)
		if err != nil {
			die("%v", err)
		}
		fmt.Print(mod.Describe())

	case *extract != "":
		if *out == "" {
			die("need -o with -extract")
		}
		data, err := im.ReadFile(*extract)
		if err != nil {
			die("%v", err)
		}
		if err := os.WriteFile(*out, data, 0644); err != nil {
			die("%v", err)
		}
		fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(data), *out)

	default:
		die("nothing to do: pass -ls, -sfo, -exe PATH, or -extract PATH -o FILE")
	}
}

// dumpSFO reads and prints PARAM.SFO, trying the standard UMD location first.
func dumpSFO(im *psp.Image) {
	for _, path := range []string{"PSP_GAME/PARAM.SFO", "PSP_GAME/SYSDIR/PARAM.SFO", "PARAM.SFO"} {
		data, err := im.ReadFile(path)
		if err != nil {
			continue
		}
		sfo, err := psp.ParseSFO(data)
		if err != nil {
			die("parsing %s: %v", path, err)
		}
		fmt.Fprintf(os.Stderr, "%s:\n", path)
		fmt.Print(sfo.Describe())
		return
	}
	die("no PARAM.SFO found")
}
