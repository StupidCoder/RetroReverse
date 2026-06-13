// modadf patches a copy of the Marble Madness ADF so its s/startup-sequence
// auto-runs the launcher (MarbleMadness!) instead of LoadWb -- lets the game
// boot straight to the c/zzz decode with no Workbench/mouse interaction.
// The replacement is the same byte length as the original, so only the one OFS
// data block changes; we recompute that block's checksum (the file header and
// bootblock are untouched).
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: modadf in.adf out.adf")
		os.Exit(2)
	}
	img, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	orig := []byte("LoadWb\nendcli > nil:\n") // 21 bytes
	idx := bytes.Index(img, orig)
	if idx < 0 {
		fmt.Fprintln(os.Stderr, "startup-sequence content not found")
		os.Exit(1)
	}
	repl := []byte("MarbleMadness!\n     \n") // exactly 21 bytes
	if len(repl) != len(orig) {
		panic(fmt.Sprintf("length mismatch %d != %d", len(repl), len(orig)))
	}
	fmt.Printf("startup-sequence data at file offset %d (block %d, data@blk+%d)\n",
		idx, idx/512, idx%512)
	copy(img[idx:], repl)

	// OFS data block: 512 bytes, checksum longword at block+20; block sum == 0.
	blk := (idx / 512) * 512
	binary.BigEndian.PutUint32(img[blk+20:], 0) // zero old checksum
	var sum uint32
	for i := 0; i < 512; i += 4 {
		sum += binary.BigEndian.Uint32(img[blk+i:])
	}
	binary.BigEndian.PutUint32(img[blk+20:], -sum) // checksum = -sum (block sums to 0)

	if err := os.WriteFile(os.Args[2], img, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s: startup-sequence now runs MarbleMadness!\n", os.Args[2])
}
