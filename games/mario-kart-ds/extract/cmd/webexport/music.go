// music stage: render Mario Kart DS's sequenced music from sound_data.sdat to
// MP3, as cmd/musicrender (SSEQ bytecode → SBNK instruments → SWAR samples via
// tools/nds/sdat, then ffmpeg to MP3).
package main

import (
	"fmt"

	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	"retroreverse.com/tools/platform/nds"
	"retroreverse.com/tools/platform/nds/sdat"
)

const rate = 32768 // the DS mixer's output rate

func runMusic(ctx *cli.Context, rom *nds.ROM) error {
	data := rom.FileByPath("data/Sound/sound_data.sdat")
	if data == nil {
		return fmt.Errorf("data/Sound/sound_data.sdat not found in ROM")
	}
	s, err := sdat.Parse(data)
	if err != nil {
		return err
	}

	total := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID >= 0 {
			total++
		}
	}
	ctx.Logf("SDAT: %d sequences (%d renderable), %d banks, %d wave archives",
		len(s.Seqs), total, len(s.Banks), len(s.Wavearcs))

	n := 0
	for i := range s.Seqs {
		if s.Seqs[i].FileID < 0 {
			continue
		}
		n++
		stem := fmt.Sprintf("seq_%02d", i)
		L, R, err := s.Render(i, rate, 2, 180)
		if err != nil {
			ctx.Logf("%s: %v", stem, err)
			continue
		}
		if len(L) < rate { // sub-second jingles/stingers cut short
			continue
		}
		fadeOut(L, R)
		samples := make([]int16, len(L)*2)
		clip := func(v float64) int16 {
			if v > 1 {
				v = 1
			}
			if v < -1 {
				v = -1
			}
			return int16(v * 32767)
		}
		for k := range L {
			samples[k*2] = clip(L[k])
			samples[k*2+1] = clip(R[k])
		}
		out, err := ctx.Builder.Path("music", stem+".mp3")
		if err != nil {
			return err
		}
		wave := audio.PCM16{Rate: rate, Channels: 2, Samples: samples}
		if err := audio.EncodeMP3(wave, out); err != nil {
			ctx.Logf("%s: %v", stem, err)
			continue
		}
		secs := len(L) / rate
		ctx.Builder.AddMedia(schema.Asset{
			ID: stem, Category: schema.CategoryMusic,
			Name:     fmt.Sprintf("Sequence %02d", i),
			File:     "music/" + stem + ".mp3",
			Duration: wave.Duration(),
		})
		ctx.Progress("music", n, total, fmt.Sprintf("%s.mp3 (%d:%02d)", stem, secs/60, secs%60))
	}
	return nil
}

// fadeOut applies a 3-second fade at the end (the render stops after the loop
// count, mid-music).
func fadeOut(L, R []float64) {
	n := 3 * rate
	if n > len(L) {
		n = len(L)
	}
	for i := 0; i < n; i++ {
		g := float64(n-i) / float64(n)
		L[len(L)-n+i] *= g
		R[len(R)-n+i] *= g
	}
}
