// packetinfo inspects a 3DO "wwww" resource container — a Need for Speed track
// packet (DriveData/DriveArt/*_PKT_*) or a car .WrapFam — listing the cels, 3D
// models and shapes it holds.
//
//	packetinfo -image disc.bin -path DriveData/DriveArt/Al1_PKT_000
//	packetinfo -f Al1_PKT_000
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	imagePath := flag.String("image", "", "3DO disc image")
	path := flag.String("path", "", "packet path within the disc")
	file := flag.String("f", "", "standalone packet file")
	list := flag.Bool("l", false, "list every resource (offset, kind)")
	flag.Parse()

	var data []byte
	var err error
	switch {
	case *file != "":
		data, err = os.ReadFile(*file)
	case *imagePath != "" && *path != "":
		var img []byte
		if img, err = os.ReadFile(*imagePath); err == nil {
			var vol *threedo.Volume
			if vol, err = threedo.Open(img); err == nil {
				data, err = vol.ReadFile(*path)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: packetinfo -image DISC -path P | -f FILE [-l]")
		os.Exit(2)
	}
	if err != nil {
		die(err)
	}

	res, err := threedo.ParseWrap(data)
	if err != nil {
		die(err)
	}
	inv := threedo.Inventory(res)
	fmt.Printf("wwww container: %d resources — %d cels, %d models, %d shapes, %d unknown\n",
		len(res), inv["cel"], inv["model"], inv["shape"], inv["unknown"])
	if *list {
		for _, r := range res {
			fmt.Printf("  depth %d  0x%06X  %s\n", r.Depth, r.Offset, r.Kind)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "packetinfo:", err)
	os.Exit(1)
}
