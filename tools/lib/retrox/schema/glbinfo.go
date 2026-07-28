package schema

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

// GLBInfo is the slice of a GLB's glTF document that validation cares about.
type GLBInfo struct {
	Animations []string // animation names, in order
	Materials  []string // material names, in order
	Scenes     []string // scene names, in order (variants are scenes)
}

// ReadGLBInfo parses a GLB (glTF-Binary container) far enough to list its
// animation and material names: 12-byte header, then the first chunk, which
// the glTF 2.0 spec requires to be the JSON chunk.
func ReadGLBInfo(r io.Reader) (*GLBInfo, error) {
	var head struct {
		Magic, Version, Length uint32
	}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return nil, fmt.Errorf("glb header: %w", err)
	}
	if head.Magic != 0x46546C67 { // "glTF"
		return nil, fmt.Errorf("not a GLB (magic %08x)", head.Magic)
	}
	if head.Version != 2 {
		return nil, fmt.Errorf("unsupported GLB version %d", head.Version)
	}
	var chunk struct {
		Length, Type uint32
	}
	if err := binary.Read(r, binary.LittleEndian, &chunk); err != nil {
		return nil, fmt.Errorf("glb chunk header: %w", err)
	}
	if chunk.Type != 0x4E4F534A { // "JSON"
		return nil, fmt.Errorf("first GLB chunk is %08x, not JSON", chunk.Type)
	}
	buf := make([]byte, chunk.Length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("glb JSON chunk: %w", err)
	}
	var doc struct {
		Animations []struct {
			Name string `json:"name"`
		} `json:"animations"`
		Scenes []struct {
			Name string `json:"name"`
		} `json:"scenes"`
		Materials []struct {
			Name string `json:"name"`
		} `json:"materials"`
	}
	if err := json.Unmarshal(buf, &doc); err != nil {
		return nil, fmt.Errorf("glb JSON chunk: %w", err)
	}
	info := &GLBInfo{}
	for _, a := range doc.Animations {
		info.Animations = append(info.Animations, a.Name)
	}
	for _, m := range doc.Materials {
		info.Materials = append(info.Materials, m.Name)
	}
	for _, sc := range doc.Scenes {
		info.Scenes = append(info.Scenes, sc.Name)
	}
	return info, nil
}
