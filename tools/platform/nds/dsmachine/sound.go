package dsmachine

// The ARM7's sound hardware — sixteen channels of PCM8, PCM16, IMA-ADPCM and PSG,
// mixed to a stereo output.
//
// This is a deliberate, declared gap. The register file is real: the channel control
// words, the source addresses, the timers and the lengths are all latched (io.go), so
// the ARM7's sound driver runs, sequences its music and keys its channels on and off
// exactly as it does on hardware, and nothing in the boot or the frame loop blocks.
// What is missing is the *mixer*: no samples are fetched and no audio is produced.
//
// The distinction matters because it is the difference between a gap and a lie. A
// game whose sound driver is running has correct timing and correct IPC traffic
// whether or not anyone is listening; a game whose sound registers silently read back
// "ready" from a stub can hang the frame it waits on one. Nothing here fakes a status
// bit.
//
// When audio does land, this is where it goes, and the shape is known: 16 channels
// summed at 32,768 Hz, and the oracle's `-wav` flag captures the mix — the same
// instrument the 3DS oracle uses to verify anything that makes a sound.

const (
	regSOUNDxCNT = 0x04000400 // 16 channels, 16 bytes apart
	regSOUNDCNT  = 0x04000500 // master volume and enable
	regSOUNDBIAS = 0x04000504
)

// soundKeyed notes, once, that the game has started a voice we are not rendering, so
// a run's log says plainly that the music is playing and unheard rather than leaving
// a reader to infer it from silence.
func (m *Machine) soundKeyed() {
	m.note("ARM7: sound channels are being keyed on; the register file is modelled but the mixer is not (sound.go)")
}
