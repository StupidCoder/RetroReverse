package psp

// image.go opens a UMD image file — either a CISO ("CSO") compressed container or
// a flat "cooked" .iso — and mounts its ISO 9660 volume. Callers use the returned
// Volume for listing and extraction and Close it when done.

import (
	"fmt"
	"io"
	"os"
)

// Image is a mounted UMD: the ISO 9660 volume plus the resource that backs it.
type Image struct {
	*Volume
	closer io.Closer
}

// Close releases the underlying file handle (a no-op for an in-memory .iso).
func (im *Image) Close() error {
	if im.closer != nil {
		return im.closer.Close()
	}
	return nil
}

// OpenImage opens the UMD image at path, auto-detecting CISO vs flat ISO by magic.
// A CISO is read through the streaming block cache so the image is never fully
// resident; a flat .iso is read into memory.
func OpenImage(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		f.Close()
		return nil, fmt.Errorf("psp: reading %s: %w", path, err)
	}

	if string(magic[:]) == "CISO" {
		st, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, err
		}
		cso, err := newCSO(f, st.Size())
		if err != nil {
			f.Close()
			return nil, err
		}
		cso.closer = f
		v, err := OpenVolume(cso)
		if err != nil {
			cso.Close()
			return nil, err
		}
		return &Image{Volume: v, closer: cso}, nil
	}

	// Flat .iso: read it all in.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, err
	}
	data, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return nil, err
	}
	v, err := OpenISO(data)
	if err != nil {
		return nil, err
	}
	return &Image{Volume: v}, nil
}
