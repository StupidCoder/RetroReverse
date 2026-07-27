package main

// camera.go keeps lmtool's -camera flag: bake a demo shot's .scd/.sco pair
// and write it as JSON (the same shape as Retro-X camera-track documents).

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/export"
)

func cameraExport(image, spec, out string) error {
	parts := strings.Split(spec, ":")
	if len(parts) != 3 {
		return fmt.Errorf("want /disc/file.szp:shot.scd:shot.sco, got %q", spec)
	}
	src, err := export.Open(image)
	if err != nil {
		return err
	}
	defer src.Close()
	files, err := src.Archive(parts[0])
	if err != nil {
		return err
	}
	cam, err := export.BakeCamera(files, parts[1], parts[2])
	if err != nil {
		return err
	}
	type frame struct {
		Pos    [3]float32 `json:"pos"`
		Target [3]float32 `json:"target"`
		Roll   float32    `json:"roll"`
		Fov    float32    `json:"fov"`
	}
	doc := struct {
		Frames int     `json:"frames"`
		Fps    float64 `json:"fps"`
		Near   float32 `json:"near"`
		Far    float32 `json:"far"`
		Track  []frame `json:"track"`
	}{Frames: cam.Frames, Fps: cam.FPS, Near: cam.Near, Far: cam.Far}
	for _, c := range cam.Track {
		doc.Track = append(doc.Track, frame{c.Pos, c.Target, c.Roll, c.Fov})
	}
	j, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(out, j, 0o644)
}
