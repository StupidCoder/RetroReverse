package psp

// mpeg.go is a minimal sceMpeg HLE — enough of the PSMF movie-player contract
// (platform spec: the pspsdk pspmpeg.h interfaces and PSMF container header)
// for a game to run its movie playback loop to the end of the stream. There is
// no AVC decoder here: every "decoded" frame is reported as produced but the
// pixels are left untouched, so a movie plays as fast as the game pumps it and
// the screen keeps whatever the game last drew. The load-bearing parts are the
// PSMF header fields, the ringbuffer accounting, and the end-of-data error
// (SCE_MPEG_ERROR_NO_DATA) that tells the player the stream is over.
//
// The ringbuffer is fed by the GAME's own packet-read callback (registered at
// sceMpegRingbufferConstruct): sceMpegRingbufferPut runs it in a nested guest
// frame (callGuest), so the data really is streamed by the game's file manager.

const (
	errMpegNoData  = 0x80618001 // SCE_MPEG_ERROR_NO_DATA: AU queue / stream empty
	errMpegInvalid = 0x806101FE // invalid PSMF header / value

	mpegPacketSize = 2048
	mpegPtsPerFrame = 3003 // 90 kHz clock at 29.97 fps
)

// mpegState is the one active sceMpeg session (games construct a single player).
type mpegState struct {
	Handle   uint32 // guest addr of the SceMpeg handle the game passed to Create
	Ringbuf  uint32 // guest addr of the SceMpegRingbuffer
	Packets  uint32 // ringbuffer capacity in 2048-byte packets
	In       uint32 // packets currently in the ringbuffer
	Cb       uint32 // the game's packet-read callback
	CbArg    uint32
	Data     uint32 // ringbuffer data area
	Pts      uint32 // last AU presentation timestamp handed out
	FedTotal uint32 // packets ever fed (diagnostics)
}

// beGuest32 reads a big-endian u32 from guest memory (PSMF headers are BE).
func (m *Machine) beGuest32(addr uint32) uint32 {
	return uint32(m.Read(addr))<<24 | uint32(m.Read(addr+1))<<16 |
		uint32(m.Read(addr+2))<<8 | uint32(m.Read(addr+3))
}

// writeMpegAu fills a SceMpegAu (6 words: ptsMsb, pts, dtsMsb, dts, esBuffer, esSize).
func (m *Machine) writeMpegAu(au, pts, esBuf, esSize uint32) {
	m.write32(au+0, 0)
	m.write32(au+4, pts)
	m.write32(au+8, 0)
	m.write32(au+12, pts)
	m.write32(au+16, esBuf)
	m.write32(au+20, esSize)
}

// mpegRingbufferConstruct records the ringbuffer and fills its guest struct
// (packets, read/written counters, free, packet size, data, callback, arg,
// data upper bound, sema, mpeg backlink).
func (m *Machine) mpegRingbufferConstruct(rb, packets, data, size, cb, cbArg uint32) uint32 {
	m.mpeg.Ringbuf, m.mpeg.Packets, m.mpeg.Data = rb, packets, data
	m.mpeg.Cb, m.mpeg.CbArg = cb, cbArg
	m.mpeg.In = 0
	m.write32(rb+0, packets)
	m.write32(rb+4, 0) // packetsRead
	m.write32(rb+8, 0) // packetsWritten
	m.write32(rb+12, packets)
	m.write32(rb+16, mpegPacketSize)
	m.write32(rb+20, data)
	m.write32(rb+24, cb)
	m.write32(rb+28, cbArg)
	m.write32(rb+32, data+size)
	m.write32(rb+36, 0)
	m.write32(rb+40, 0)
	m.note("sceMpegRingbufferConstruct: %d packets, data 0x%08X, cb 0x%08X", packets, data, cb)
	return 0
}

// mpegRingbufferPut asks the game's callback for up to n packets and accounts
// for what it delivered. The callback signature is (data, numPackets, arg) ->
// packets written (or a negative error, passed through).
func (m *Machine) mpegRingbufferPut(rb, n, avail uint32) uint32 {
	free := m.mpeg.Packets - m.mpeg.In
	if n > free {
		n = free
	}
	if n > avail {
		n = avail
	}
	if n == 0 || m.mpeg.Cb == 0 {
		return 0
	}
	got := m.callGuest(m.mpeg.Cb, m.mpeg.Data, n, m.mpeg.CbArg)
	if got&0x80000000 != 0 { // negative: the game's feeder reports an error
		m.note("sceMpegRingbufferPut: callback returned 0x%08X", got)
		return got
	}
	if got > n {
		got = n
	}
	m.mpeg.In += got
	m.mpeg.FedTotal += got
	m.write32(rb+12, m.mpeg.Packets-m.mpeg.In) // packetsFree
	m.write32(rb+8, m.read32(rb+8)+got)        // packetsWritten
	return got
}

// --- sceAtrac3plus ----------------------------------------------------------
//
// A matching minimal ATRAC3+ HLE: no codec, silence PCM. The frame accounting
// is real — the RIFF header the game hands over names the block align and data
// size, and the decode loop runs one block per call until the data is spent,
// then reports the end flag and SCE_ATRAC_ERROR_ALL_DATA_DECODED. That is what
// a player loop needs to run its audio stream to completion and move on.

const (
	errAtracAllDecoded = 0x80630002
	errAtracBadID      = 0x80630004
	atracMaxSamples    = 2048 // ATRAC3+ output samples per decode call
)

// atracState is one decoding stream (sceAtracSetDataAndGetID).
type atracState struct {
	Buf, Size uint32
	Frames    uint32 // total decode calls the data covers
	Pos       uint32 // decode calls already served
	Channels  uint32
}

// atracParseRiff derives the frame count from the RIFF/WAVE header: data-chunk
// size over the fmt block align. A buffer without a parseable header gets a
// conservative estimate.
func (m *Machine) atracParseRiff(a *atracState) {
	a.Channels = 2
	a.Frames = a.Size / 512
	if m.Read(a.Buf) != 'R' || m.Read(a.Buf+1) != 'I' || m.Read(a.Buf+2) != 'F' || m.Read(a.Buf+3) != 'F' {
		return
	}
	var blockAlign, dataSize uint32
	for p := a.Buf + 12; p+8 < a.Buf+a.Size; {
		tag := string([]byte{m.Read(p), m.Read(p + 1), m.Read(p + 2), m.Read(p + 3)})
		sz := m.read32(p + 4)
		switch tag {
		case "fmt ":
			a.Channels = uint32(m.Read(p+10)) | uint32(m.Read(p+11))<<8
			blockAlign = uint32(m.Read(p+20)) | uint32(m.Read(p+21))<<8
		case "data":
			dataSize = sz
		}
		p += 8 + (sz+1)&^1
	}
	if blockAlign != 0 && dataSize != 0 {
		a.Frames = dataSize / blockAlign
	}
	if a.Channels == 0 || a.Channels > 2 {
		a.Channels = 2
	}
}

// atracDecode serves one silent frame: zero PCM, the sample count, the end flag
// on the last frame, and the all-decoded error past it.
func (m *Machine) atracDecode(id, out, samplesPtr, endPtr, remainPtr uint32) uint32 {
	a := m.atrac[id]
	if a == nil {
		return errAtracBadID
	}
	if a.Pos >= a.Frames {
		if samplesPtr != 0 {
			m.write32(samplesPtr, 0)
		}
		if endPtr != 0 {
			m.write32(endPtr, 1)
		}
		if remainPtr != 0 {
			m.write32(remainPtr, 0)
		}
		return errAtracAllDecoded
	}
	a.Pos++
	for i := uint32(0); i < atracMaxSamples*2*a.Channels; i += 4 {
		m.write32(out+i, 0)
	}
	if samplesPtr != 0 {
		m.write32(samplesPtr, atracMaxSamples)
	}
	end := uint32(0)
	if a.Pos >= a.Frames {
		end = 1
	}
	if endPtr != 0 {
		m.write32(endPtr, end)
	}
	if remainPtr != 0 {
		m.write32(remainPtr, a.Frames-a.Pos)
	}
	return 0
}

// mpegGetAu consumes one buffered packet and fills the AU, or reports that the
// stream has no data. Consuming a packet per AU is a pacing approximation (real
// access units span a variable packet count); the player's end-of-stream logic
// only needs the counters to drain.
func (m *Machine) mpegGetAu(au uint32) uint32 {
	if m.mpeg.In == 0 {
		return errMpegNoData
	}
	m.mpeg.In--
	if rb := m.mpeg.Ringbuf; rb != 0 {
		m.write32(rb+12, m.mpeg.Packets-m.mpeg.In)
		m.write32(rb+4, m.read32(rb+4)+1) // packetsRead
	}
	m.mpeg.Pts += mpegPtsPerFrame
	m.writeMpegAu(au, m.mpeg.Pts, m.read32(au+16), mpegPacketSize)
	return 0
}
