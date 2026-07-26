package main

// camera.go exports a demo shot's camera track — the .scd channels plus the
// .sco cut bases — as JSON: one sample per frame at the native 30 fps, in the
// same model space as the shot's exported GLB set.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

func cameraExport(image, spec, out string) error {
	parts := strings.Split(spec, ":")
	if len(parts) != 3 {
		return fmt.Errorf("want /disc/file.szp:shot.scd:shot.sco, got %q", spec)
	}
	b, err := discFile(image, parts[0])
	if err != nil {
		return err
	}
	if len(b) >= 4 && string(b[:4]) == "Yay0" {
		if b, err = lm.Yay0(b); err != nil {
			return err
		}
	}
	files, err := lm.RARC(b)
	if err != nil {
		return err
	}
	var scdData, scoData []byte
	for _, f := range files {
		if strings.EqualFold(f.Name, parts[1]) {
			scdData = f.Data
		}
		if strings.EqualFold(f.Name, parts[2]) {
			scoData = f.Data
		}
	}
	if scdData == nil || scoData == nil {
		return fmt.Errorf("members %q/%q not found in %s", parts[1], parts[2], parts[0])
	}
	scd, err := lm.ParseSCD(scdData)
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
		Fps    int     `json:"fps"`
		Near   float32 `json:"near"`
		Far    float32 `json:"far"`
		Track  []frame `json:"track"`
	}{Frames: scd.FrameCount, Fps: 30}
	for f := 0; f <= scd.FrameCount; f++ {
		c := scd.EvalCamera(scoData, float32(f))
		doc.Near, doc.Far = c.Near, c.Far
		doc.Track = append(doc.Track, frame{Pos: c.Pos, Target: c.Target, Roll: c.Roll, Fov: c.Fov})
	}
	j, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(out, j, 0o644)
}
