package main

// The music stage: the 31 songs rendered through the tools/platform/n64/audio
// synth (the same pipeline cmd/musicrender validated A/B against the oracle's
// AI capture), encoded by the shared MP3 helper. It reads the music instrument
// bank the game DMAs from cart 0x62D460, that bank's own VADPCM sample table
// at 0x6314D0, and the "S1" sequence bank at 0x618B70.

import (
	"fmt"

	"retroreverse.com/tools/lib/retrox/audio"
	"retroreverse.com/tools/lib/retrox/cli"
	"retroreverse.com/tools/lib/retrox/schema"
	n64audio "retroreverse.com/tools/platform/n64/audio"
)

const (
	seqBankOff  = 0x618B70
	seqBankEnd  = 0x62D460
	musicBank   = 0x62D460
	musicTblOff = 0x6314D0

	musicRate  = 22050 // output sample rate
	musicLoops = 2     // stop looping songs after N repeats
	musicMax   = 150.0 // hard length cap per song (seconds)
	musicFade  = 3.0   // fade-out seconds
)

func exportMusic(ctx *cli.Context, rom []byte) error {
	bank, err := n64audio.ParseBankFile(rom[musicBank:])
	if err != nil {
		return fmt.Errorf("music bank: %w", err)
	}
	sb, err := n64audio.ParseSeqBank(rom[seqBankOff:seqBankEnd])
	if err != nil {
		return fmt.Errorf("seq bank: %w", err)
	}
	player := n64audio.NewPlayer(bank.Banks[0], rom[musicTblOff:], musicRate)

	for i, song := range sb.Songs {
		L, R := player.Render(song, musicLoops, musicMax)
		fadeOut(L, R, int(musicFade*musicRate))
		pcm := audio.PCM16{Rate: musicRate, Channels: 2, Samples: interleave(L, R)}
		id := fmt.Sprintf("song-%02d", i)
		out, err := ctx.Builder.Path("music", id+".mp3")
		if err != nil {
			return err
		}
		if err := audio.EncodeMP3(pcm, out); err != nil {
			return fmt.Errorf("song %d: %w", i, err)
		}
		ctx.Builder.AddMedia(schema.Asset{
			ID: id, Category: schema.CategoryMusic,
			Name:     fmt.Sprintf("Song %02d", i),
			File:     "music/" + id + ".mp3",
			Duration: pcm.Duration(),
		})
		ctx.Progress("music", i+1, len(sb.Songs),
			fmt.Sprintf("%s  %.1fs, %d notes", id, pcm.Duration(), player.NotesPlayed))
	}
	return nil
}

// fadeOut applies a linear fade to the last n samples.
func fadeOut(L, R []float64, n int) {
	if n <= 0 || n > len(L) {
		n = len(L)
	}
	start := len(L) - n
	for i := 0; i < n; i++ {
		g := 1 - float64(i)/float64(n)
		L[start+i] *= g
		R[start+i] *= g
	}
}

func interleave(L, R []float64) []int16 {
	out := make([]int16, 0, len(L)*2)
	for i := range L {
		out = append(out, clip16(L[i]), clip16(R[i]))
	}
	return out
}

func clip16(v float64) int16 {
	if v > 1 {
		v = 1
	}
	if v < -1 {
		v = -1
	}
	return int16(v * 32767)
}
