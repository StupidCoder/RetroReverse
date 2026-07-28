// statediff loads two oracle savestates and diffs guest DRAM word-wise —
// the cheap way to surface which globals a window of gameplay touched
// (used to hunt the crash system's counters and flags).
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func dump(path, image string) []byte {
	data, err := os.ReadFile(image)
	if err != nil {
		panic(err)
	}
	vol, err := threedo.Open(data)
	if err != nil {
		panic(err)
	}
	prog, err := vol.ReadFile("LaunchMe")
	if err != nil {
		panic(err)
	}
	aif, err := threedo.ParseAIF(prog)
	if err != nil {
		panic(err)
	}
	m := threedo.NewMachine()
	m.SetVolume(vol)
	m.LoadAIF(aif)
	if err := m.LoadStateFile(path); err != nil {
		panic(err)
	}
	buf := make([]byte, 3*1024*1024)
	for i := range buf {
		buf[i] = m.Read(uint32(i))
	}
	return buf
}

func main() {
	image := flag.String("image", "", "disc image")
	lo := flag.Uint64("lo", 0, "range start")
	hi := flag.Uint64("hi", 0x300000, "range end")
	dec := flag.Bool("dec", false, "only words where post == pre-1")
	inc := flag.Bool("inc", false, "only words where post == pre+1")
	flag.Parse()
	a := dump(flag.Arg(0), *image)
	b := dump(flag.Arg(1), *image)
	n := 0
	for off := *lo &^ 3; off < *hi; off += 4 {
		wa := binary.BigEndian.Uint32(a[off:])
		wb := binary.BigEndian.Uint32(b[off:])
		if wa == wb {
			continue
		}
		if *dec && wb != wa-1 {
			continue
		}
		if *inc && wb != wa+1 {
			continue
		}
		fmt.Printf("%08X  %08X -> %08X\n", off, wa, wb)
		if n++; n > 400 {
			fmt.Println("... (truncated)")
			break
		}
	}
}
