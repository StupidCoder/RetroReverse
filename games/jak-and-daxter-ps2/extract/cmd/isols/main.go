package main

import (
	"fmt"
	"os"

	"retroreverse.com/tools/lib/iso9660"
)

func main() {
	f, _ := os.Open("image/Jak and Daxter - The Precursor Legacy.iso")
	st, _ := f.Stat()
	vol, err := iso9660.Open(f, st.Size())
	if err != nil { panic(err) }
	vol.Walk(func(e iso9660.Entry) error {
		if !e.IsDir {
			fmt.Printf("%10d  %s\n", e.Size, e.Path)
		}
		return nil
	})
}
