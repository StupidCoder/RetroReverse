package main

import (
	"os"

	"retroreverse.com/tools/platform/threedo"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	vol, err := threedo.Open(data)
	if err != nil { panic(err) }
	raw, err := vol.ReadFile(os.Args[2])
	if err != nil { panic(err) }
	os.WriteFile(os.Args[3], raw, 0o644)
}
