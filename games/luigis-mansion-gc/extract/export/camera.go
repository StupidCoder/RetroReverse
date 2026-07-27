package export

// camera.go bakes a demo shot's camera track — the .scd channels plus the
// .sco cut bases — to one sample per frame at the native 30 fps, in the same
// model space as the shot's exported GLB set.

import (
	"fmt"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
)

// CamSample is one baked camera frame.
type CamSample struct {
	Pos    [3]float32
	Target [3]float32
	Roll   float32
	Fov    float32
}

// CamTrack is a fully baked camera path. Frames == len(Track): the .scd
// declares FrameCount and the channels are sampled 0..FrameCount inclusive,
// so a 337-frame shot yields 338 samples and Frames says 338.
type CamTrack struct {
	Frames int
	FPS    float64
	Near   float32
	Far    float32
	Track  []CamSample
}

// BakeCamera evaluates an archive's .scd/.sco pair per frame.
func BakeCamera(files []lm.RARCFile, scdName, scoName string) (*CamTrack, error) {
	scdF, scoF := Member(files, scdName), Member(files, scoName)
	if scdF == nil || scoF == nil {
		return nil, fmt.Errorf("members %q/%q not found", scdName, scoName)
	}
	scd, err := lm.ParseSCD(scdF.Data)
	if err != nil {
		return nil, err
	}
	out := &CamTrack{FPS: 30}
	for f := 0; f <= scd.FrameCount; f++ {
		c := scd.EvalCamera(scoF.Data, float32(f))
		out.Near, out.Far = c.Near, c.Far
		out.Track = append(out.Track, CamSample{Pos: c.Pos, Target: c.Target, Roll: c.Roll, Fov: c.Fov})
	}
	out.Frames = len(out.Track)
	return out, nil
}

// FindShotCamera locates the one .scd/.sco pair in a shot archive.
func FindShotCamera(files []lm.RARCFile) (scd, sco string, ok bool) {
	for _, f := range files {
		switch {
		case strings.HasSuffix(strings.ToLower(f.Name), ".scd"):
			scd = f.Name
		case strings.HasSuffix(strings.ToLower(f.Name), ".sco"):
			sco = f.Name
		}
	}
	return scd, sco, scd != "" && sco != ""
}
