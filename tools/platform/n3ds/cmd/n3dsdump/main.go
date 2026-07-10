// n3dsdump lists the contents of a Nintendo 3DS cartridge image (CCI/.3ds) and
// can extract them. This is the first step when reverse engineering a 3DS game:
// it shows the NCSD partition table, each NCCH container's header, the ExHeader's
// code-set layout, the ExeFS file table and the RomFS tree.
//
// It refuses to guess at an encrypted image: the AES-CTR keys are console state,
// not cartridge data, so a retail dump must be decrypted before it is readable.
//
// Usage:
//
//	n3dsdump game.cci                 list partitions, ExHeader, ExeFS, RomFS summary
//	n3dsdump -romfs game.cci          also list every RomFS file
//	n3dsdump -code out.bin game.cci   write the decompressed ARM11 .code
//	n3dsdump -x outdir game.cci       extract the ExeFS files and the whole RomFS
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/platform/n3ds"
)

func main() {
	listRomFS := flag.Bool("romfs", false, "list every file in the RomFS")
	codeOut := flag.String("code", "", "write the decompressed ExeFS/.code to this file")
	extract := flag.String("x", "", "extract the ExeFS files and the RomFS tree into this directory")
	part := flag.Int("p", 0, "NCSD partition to inspect")
	verify := flag.Bool("verify", false, "verify the RomFS IVFC hash tree end to end (reads the whole region)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: n3dsdump [-romfs] [-verify] [-code FILE] [-x DIR] [-p N] game.cci")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), *part, *listRomFS, *verify, *codeOut, *extract); err != nil {
		fmt.Fprintln(os.Stderr, "n3dsdump:", err)
		os.Exit(1)
	}
}

func run(path string, part int, listRomFS, verify bool, codeOut, extract string) error {
	img, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ncsd, err := n3ds.ParseNCSD(img)
	if err != nil {
		return err
	}

	fmt.Printf("NCSD  %s\n", filepath.Base(path))
	fmt.Printf("  file size       %d bytes (0x%x)\n", len(img), len(img))
	fmt.Printf("  header size     %d bytes (0x%x)\n", ncsd.ImageSize, ncsd.ImageSize)
	fmt.Printf("  media ID        %016x\n", ncsd.MediaID)
	fmt.Printf("  media unit      %d bytes\n", ncsd.MediaUnitSize)
	fmt.Println("  partitions:")
	for _, p := range ncsd.Partitions {
		if p.Empty() {
			continue
		}
		fmt.Printf("    %d  offset 0x%08x  size 0x%08x  fstype %d  crypt %d  id %016x\n",
			p.Index, p.Offset, p.Size, p.FSType, p.Crypt, p.ID)
	}

	c, err := ncsd.Partition(part)
	if err != nil {
		return err
	}
	fmt.Printf("\nNCCH  partition %d\n", part)
	fmt.Printf("  product code    %s\n", c.ProductCode)
	fmt.Printf("  maker code      %s\n", c.MakerCode)
	fmt.Printf("  version         %d\n", c.Version)
	fmt.Printf("  program ID      %016x\n", c.ProgramID)
	fmt.Printf("  partition ID    %016x\n", c.PartitionID)
	fmt.Printf("  content type    %s\n", c.ContentType())
	fmt.Printf("  crypto          %s\n", cryptoDesc(c))
	fmt.Printf("  exheader        0x%x bytes\n", c.ExHeaderSize)
	fmt.Printf("  exefs           offset 0x%08x  size 0x%08x\n", c.ExeFSRegion.Offset, c.ExeFSRegion.Size)
	fmt.Printf("  romfs           offset 0x%08x  size 0x%08x\n", c.RomFSRegion.Offset, c.RomFSRegion.Size)

	if c.Encrypted() {
		return fmt.Errorf("partition %d is encrypted (%s); nothing further can be read", part, c.CryptoMethod())
	}

	ex, err := c.ExHeader()
	if err != nil {
		return err
	}
	fmt.Printf("\nExHeader\n")
	fmt.Printf("  title           %s\n", ex.Title)
	fmt.Printf("  remaster ver    %d\n", ex.RemasterVer)
	fmt.Printf("  flags           0x%02x (exefs .code compressed: %v, SD application: %v)\n",
		ex.Flag, ex.CompressedExeFSCode(), ex.SDApplication())
	fmt.Printf("  text            addr 0x%08x  %3d pages (0x%06x)  size 0x%06x\n", ex.Text.Address, ex.Text.NumPages, ex.Text.Extent(), ex.Text.Size)
	fmt.Printf("  rodata          addr 0x%08x  %3d pages (0x%06x)  size 0x%06x\n", ex.ROData.Address, ex.ROData.NumPages, ex.ROData.Extent(), ex.ROData.Size)
	fmt.Printf("  data            addr 0x%08x  %3d pages (0x%06x)  size 0x%06x\n", ex.Data.Address, ex.Data.NumPages, ex.Data.Extent(), ex.Data.Size)
	fmt.Printf("  bss             addr 0x%08x  size 0x%06x\n", ex.BSSAddress(), ex.BSSSize)
	fmt.Printf("  stack           0x%x\n", ex.StackSize)
	fmt.Printf("  save data       0x%x\n", ex.SaveDataSize)
	fmt.Printf("  .code size      0x%x (expected, from the segment arithmetic)\n", ex.CodeSize())
	fmt.Printf("  dependencies    %d sysmodules\n", len(ex.Dependencies))
	for _, d := range ex.Dependencies {
		fmt.Printf("                  %016x\n", d)
	}

	efs, err := c.ExeFS()
	if err != nil {
		return err
	}
	fmt.Printf("\nExeFS  %d files\n", len(efs.Files))
	for _, f := range efs.Files {
		fmt.Printf("  %-10s offset 0x%08x  size 0x%08x\n", f.Name, f.Offset, f.Size)
	}

	code, err := efs.Code(ex)
	if err != nil {
		return fmt.Errorf("reading .code: %w", err)
	}
	stored, _ := efs.File(".code")
	fmt.Printf("\n.code  0x%x stored -> 0x%x decompressed (%.1f%%), matches the ExHeader\n",
		len(stored), len(code), 100*float64(len(stored))/float64(len(code)))

	if codeOut != "" {
		if err := os.WriteFile(codeOut, code, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", codeOut)
	}

	rfs, err := c.RomFS()
	if err != nil {
		fmt.Printf("\nRomFS  %v\n", err)
	} else {
		var total int64
		for _, f := range rfs.Files {
			total += f.Size
		}
		fmt.Printf("\nRomFS  %d files, %d directories, %d bytes of file data\n", len(rfs.Files), len(rfs.Dirs), total)
		if verify {
			if err := rfs.VerifyIVFC(); err != nil {
				return err
			}
			fmt.Println("  IVFC hash tree verifies (master -> level 1 -> level 2 -> level 3)")
		}
		if listRomFS {
			for _, f := range rfs.Files {
				fmt.Printf("  %-60s 0x%09x  %d\n", f.Path, f.Offset, f.Size)
			}
		} else {
			fmt.Println("  top level:")
			for _, d := range topLevel(rfs) {
				fmt.Printf("    %s\n", d)
			}
		}
	}

	if extract != "" {
		return extractAll(extract, efs, code, rfs)
	}
	return nil
}

func cryptoDesc(c *n3ds.NCCH) string {
	if c.Encrypted() {
		return "AES-CTR, " + c.CryptoMethod() + " (ENCRYPTED)"
	}
	return "none (NoCrypto set — decrypted dump)"
}

// topLevel summarises the RomFS root: its immediate directories, plus a count of
// files sitting directly at the root.
func topLevel(fs *n3ds.RomFS) []string {
	seen := map[string]bool{}
	var out []string
	for _, d := range fs.Dirs {
		if strings.Count(d, "/") == 1 && !seen[d] {
			seen[d] = true
			out = append(out, d+"/")
		}
	}
	for _, f := range fs.Files {
		if strings.Count(f.Path, "/") == 1 {
			out = append(out, f.Path)
		}
	}
	sort.Strings(out)
	return out
}

func extractAll(dir string, efs *n3ds.ExeFS, code []byte, rfs *n3ds.RomFS) error {
	exeDir := filepath.Join(dir, "exefs")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		return err
	}
	for _, f := range efs.Files {
		b, err := efs.File(f.Name)
		if err != nil {
			return err
		}
		if f.Name == ".code" {
			b = code // write the decompressed form
		}
		if err := os.WriteFile(filepath.Join(exeDir, strings.TrimPrefix(f.Name, ".")), b, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("extracted %d ExeFS files into %s\n", len(efs.Files), exeDir)

	if rfs == nil {
		return nil
	}
	romDir := filepath.Join(dir, "romfs")
	for _, f := range rfs.Files {
		p := filepath.Join(romDir, filepath.FromSlash(strings.TrimPrefix(f.Path, "/")))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, rfs.Data(f), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("extracted %d RomFS files into %s\n", len(rfs.Files), romDir)
	return nil
}
