// Package wsock is a server-side WebSocket (RFC 6455) implementation in stdlib Go.
//
// It exists because the frame debugger streams framebuffers to a browser and the
// tools module carries no external dependencies. The surface is only what a
// localhost debugger needs: upgrade an HTTP request, read masked client messages,
// write unmasked text and binary ones. There is no client side, no extension
// negotiation (permessage-deflate would only burn CPU on incompressible pixels),
// and no subprotocol.
//
// A Conn is safe for one reader and any number of concurrent writers.
package wsock

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Opcodes, as on the wire.
const (
	OpText   byte = 0x1
	OpBinary byte = 0x2
	opClose  byte = 0x8
	opPing   byte = 0x9
	opPong   byte = 0xA
	opCont   byte = 0x0
)

// guid is the magic string RFC 6455 §1.3 concatenates with the client key.
const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// maxMessage caps a reassembled client message. Client messages here are short
// JSON commands, so this is a guard against a hostile peer, not a real limit.
const maxMessage = 1 << 20

// ErrClosed is returned by Read once the peer has sent a close frame.
var ErrClosed = errors.New("wsock: connection closed")

// Conn is an established WebSocket connection.
type Conn struct {
	c  net.Conn
	br *bufio.Reader

	wmu sync.Mutex // serialises writers; a message must not interleave with another
}

// Accept computes the Sec-WebSocket-Accept value for a client key.
func Accept(key string) string {
	h := sha1.New()
	io.WriteString(h, key+guid)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Upgrade completes the WebSocket handshake and hijacks the connection. On error
// it has already written an HTTP error response.
func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if !headerHasToken(r.Header, "Connection", "upgrade") ||
		!strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "expected a websocket upgrade", http.StatusBadRequest)
		return nil, errors.New("wsock: not an upgrade request")
	}
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		w.Header().Set("Sec-WebSocket-Version", "13")
		http.Error(w, "unsupported websocket version", http.StatusUpgradeRequired)
		return nil, errors.New("wsock: bad version")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return nil, errors.New("wsock: missing key")
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "connection cannot be hijacked", http.StatusInternalServerError)
		return nil, errors.New("wsock: ResponseWriter is not a Hijacker")
	}
	c, brw, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("wsock: hijack: %w", err)
	}
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + Accept(key) + "\r\n\r\n"
	if _, err := brw.WriteString(resp); err != nil {
		c.Close()
		return nil, err
	}
	if err := brw.Flush(); err != nil {
		c.Close()
		return nil, err
	}
	// brw.Reader may already hold bytes the client pipelined after the handshake,
	// so keep it rather than starting a fresh reader on the raw conn.
	return &Conn{c: c, br: brw.Reader}, nil
}

// headerHasToken reports whether a comma-separated header field contains a token.
func headerHasToken(h http.Header, name, token string) bool {
	for _, v := range h.Values(name) {
		for _, t := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(t), token) {
				return true
			}
		}
	}
	return false
}

// Read returns the next text or binary message, reassembling continuation frames.
// Control frames are handled inside: a ping is answered with a pong, a close ends
// the connection with ErrClosed.
func (c *Conn) Read() (op byte, payload []byte, err error) {
	var msg []byte
	var msgOp byte
	for {
		fin, opcode, frame, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch opcode {
		case opPing:
			if err := c.write(opPong, frame); err != nil {
				return 0, nil, err
			}
			continue
		case opPong:
			continue
		case opClose:
			c.write(opClose, frame) // best effort; the peer is going away regardless
			c.c.Close()
			return 0, nil, ErrClosed
		case OpText, OpBinary:
			if msg != nil {
				return 0, nil, errors.New("wsock: new message began mid-fragment")
			}
			msgOp = opcode
			msg = frame
		case opCont:
			if msg == nil {
				return 0, nil, errors.New("wsock: continuation with no message")
			}
			msg = append(msg, frame...)
		default:
			return 0, nil, fmt.Errorf("wsock: unknown opcode %#x", opcode)
		}
		if len(msg) > maxMessage {
			return 0, nil, fmt.Errorf("wsock: message over %d bytes", maxMessage)
		}
		if fin {
			return msgOp, msg, nil
		}
	}
}

// readFrame reads one frame off the wire and unmasks its payload.
func (c *Conn) readFrame() (fin bool, opcode byte, payload []byte, err error) {
	var h [2]byte
	if _, err := io.ReadFull(c.br, h[:]); err != nil {
		return false, 0, nil, err
	}
	fin = h[0]&0x80 != 0
	if h[0]&0x70 != 0 {
		return false, 0, nil, errors.New("wsock: reserved bits set (no extension negotiated)")
	}
	opcode = h[0] & 0x0F
	masked := h[1]&0x80 != 0
	if !masked {
		// RFC 6455 §5.1: a client-to-server frame MUST be masked.
		return false, 0, nil, errors.New("wsock: unmasked client frame")
	}
	n := uint64(h[1] & 0x7F)
	switch n {
	case 126:
		var b [2]byte
		if _, err := io.ReadFull(c.br, b[:]); err != nil {
			return false, 0, nil, err
		}
		n = uint64(binary.BigEndian.Uint16(b[:]))
	case 127:
		var b [8]byte
		if _, err := io.ReadFull(c.br, b[:]); err != nil {
			return false, 0, nil, err
		}
		n = binary.BigEndian.Uint64(b[:])
	}
	if n > maxMessage {
		return false, 0, nil, fmt.Errorf("wsock: frame of %d bytes over the %d limit", n, maxMessage)
	}
	var key [4]byte
	if _, err := io.ReadFull(c.br, key[:]); err != nil {
		return false, 0, nil, err
	}
	payload = make([]byte, n)
	if _, err := io.ReadFull(c.br, payload); err != nil {
		return false, 0, nil, err
	}
	for i := range payload {
		payload[i] ^= key[i&3]
	}
	return fin, opcode, payload, nil
}

// WriteText sends a text message.
func (c *Conn) WriteText(b []byte) error { return c.write(OpText, b) }

// WriteBinary sends a binary message.
func (c *Conn) WriteBinary(b []byte) error { return c.write(OpBinary, b) }

// write sends one unfragmented, unmasked frame. Server-to-client frames are never
// masked (RFC 6455 §5.1).
func (c *Conn) write(opcode byte, b []byte) error {
	hdr := make([]byte, 0, 10)
	hdr = append(hdr, 0x80|opcode) // FIN
	switch n := len(b); {
	case n < 126:
		hdr = append(hdr, byte(n))
	case n <= 0xFFFF:
		hdr = append(hdr, 126, byte(n>>8), byte(n))
	default:
		hdr = append(hdr, 127)
		hdr = binary.BigEndian.AppendUint64(hdr, uint64(n))
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if _, err := c.c.Write(hdr); err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	_, err := c.c.Write(b)
	return err
}

// Close sends a close frame and shuts the connection down.
func (c *Conn) Close() error {
	c.write(opClose, []byte{0x03, 0xE8}) // 1000 normal closure
	return c.c.Close()
}
