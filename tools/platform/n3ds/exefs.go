package n3ds

import (
	"encoding/binary"
	"fmt"
)

// The ExeFS header is a 0x200-byte block: ten 16-byte file entries, 0x20 bytes
// reserved, then ten SHA-256 hashes in *reverse* file order. File offsets are
// relative to the end of this header.
const (
	exefsMaxFiles   = 10
	exefsHeaderSize = 0x200
)

// ExeFSFile is one entry of the ExeFS file table.
type ExeFSFile struct {
	Name   string
	Offset int64 // relative to the end of the ExeFS header
	Size   int64
	Hash   [32]byte // SHA-256 of the file's bytes, from the header's hash table
}

// ExeFS is the parsed executable filesystem.
type ExeFS struct {
	raw   []byte
	Files []ExeFSFile
}

// ParseExeFS parses the ExeFS header and validates each entry's extent.
func ParseExeFS(b []byte) (*ExeFS, error) {
	if len(b) < exefsHeaderSize {
		return nil, fmt.Errorf("n3ds: ExeFS too short (%d bytes)", len(b))
	}
	fs := &ExeFS{raw: b}
	for i := 0; i < exefsMaxFiles; i++ {
		e := b[i*16:]
		name := string(trimNul(e[0:8]))
		if name == "" {
			continue
		}
		f := ExeFSFile{
			Name:   name,
			Offset: int64(binary.LittleEndian.Uint32(e[8:])),
			Size:   int64(binary.LittleEndian.Uint32(e[12:])),
		}
		end := exefsHeaderSize + f.Offset + f.Size
		if f.Offset < 0 || f.Size < 0 || end > int64(len(b)) {
			return nil, fmt.Errorf("n3ds: ExeFS file %q [0x%x+0x%x] runs past the region end (0x%x)", name, f.Offset, f.Size, len(b))
		}
		// The hash table runs backwards: the last file's hash comes first.
		copy(f.Hash[:], b[exefsHeaderSize-32*(i+1):])
		fs.Files = append(fs.Files, f)
	}
	if len(fs.Files) == 0 {
		return nil, fmt.Errorf("n3ds: ExeFS has no files")
	}
	return fs, nil
}

// File returns the bytes of the named ExeFS file.
func (fs *ExeFS) File(name string) ([]byte, error) {
	for _, f := range fs.Files {
		if f.Name == name {
			start := exefsHeaderSize + f.Offset
			return fs.raw[start : start+f.Size], nil
		}
	}
	return nil, fmt.Errorf("n3ds: ExeFS has no file %q", name)
}

// Code returns ExeFS/.code, decompressing it when the ExHeader says it is
// BLZ-compressed, and verifies the result against the ExHeader's own segment
// arithmetic. Passing a nil ExHeader skips both the decompression decision and
// the check, returning the stored bytes.
func (fs *ExeFS) Code(ex *ExHeader) ([]byte, error) {
	raw, err := fs.File(".code")
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return raw, nil
	}
	code := raw
	if ex.CompressedExeFSCode() {
		if code, err = DecompressBLZ(raw); err != nil {
			return nil, fmt.Errorf("n3ds: decompressing .code: %w", err)
		}
	}
	if want := ex.CodeSize(); uint32(len(code)) != want {
		return nil, fmt.Errorf("n3ds: .code is 0x%x bytes but the ExHeader describes 0x%x "+
			"(text extent 0x%x + rodata extent 0x%x + data 0x%x)",
			len(code), want, ex.Text.Extent(), ex.ROData.Extent(), ex.Data.Size)
	}
	return code, nil
}
