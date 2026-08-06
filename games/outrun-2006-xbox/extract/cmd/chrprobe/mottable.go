package main

import (
	"fmt"
	"os"
	"strings"
)

// motTableDump reads /Common/motdata_table.bin: records {nameOff u32,
// listOff u32, count u32, 0}; list = count u32 offsets of clip-name
// c-strings. Maps every mot_*.gz to its ordered clip names.
func motTableDump(path, filter string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal("%v", err)
	}
	for o := 0; o+16 <= len(b); o += 16 {
		nameOff, listOff, cnt := int(u32(b, o)), int(u32(b, o+4)), int(u32(b, o+8))
		if nameOff == 0 || nameOff >= len(b) || listOff >= len(b) {
			break
		}
		name := cstr(b, nameOff)
		if !strings.HasPrefix(name, "mot_") {
			break
		}
		if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
			continue
		}
		fmt.Printf("%s: %d clips\n", name, cnt)
		for i := 0; i < cnt; i++ {
			no := int(u32(b, listOff+4*i))
			if no >= len(b) {
				fmt.Printf("  %2d: <bad %#x>\n", i, no)
				continue
			}
			fmt.Printf("  %2d: %s\n", i, cstr(b, no))
		}
	}
}
