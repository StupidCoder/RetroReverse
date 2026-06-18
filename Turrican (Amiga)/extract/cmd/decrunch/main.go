// Command decrunch unpacks the Turrican (Amiga) main part straight from the disk
// image, using the pure-Go reimplementation of the $50008 decruncher.
//
//	decrunch [-o image.bin] [Turrican.adf]
//
// The crunched blob is the boot loader's main read: it lives at disk offset
// $2C00 and its length is the big-endian long at its head ($22C98). The decoded
// image loads at $43880 with the game entered at $5F500.
package main

import (
	"crypto/md5"
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"turrican/extract/decrunch"
)

const blobDiskOffset = 0x2C00

func main() {
	out := flag.String("o", "", "write the decrunched image to this file")
	flag.Parse()

	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican.adf"
	}

	adf, err := os.ReadFile(adfPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read adf:", err)
		os.Exit(1)
	}
	if len(adf) < blobDiskOffset+4 {
		fmt.Fprintln(os.Stderr, "image too small")
		os.Exit(1)
	}

	blob := adf[blobDiskOffset:]
	packedLen := int(binary.BigEndian.Uint32(blob))
	if packedLen > len(blob) {
		fmt.Fprintf(os.Stderr, "packedLen $%X exceeds available data $%X\n", packedLen, len(blob))
		os.Exit(1)
	}
	blob = blob[:packedLen]

	res, err := decrunch.Decrunch(blob)
	if err != nil {
		fmt.Fprintln(os.Stderr, "decrunch:", err)
		os.Exit(1)
	}

	fmt.Printf("base  = $%05X\n", res.Base)
	fmt.Printf("entry = $%05X (offset $%05X into image)\n", res.Entry, res.Entry-res.Base)
	fmt.Printf("size  = $%05X (%d bytes), ends at $%05X\n", len(res.Data), len(res.Data), res.Base+uint32(len(res.Data)))
	fmt.Printf("md5   = %x\n", md5.Sum(res.Data))

	if *out != "" {
		if err := os.WriteFile(*out, res.Data, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", *out)
	}
}
