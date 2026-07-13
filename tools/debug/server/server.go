// Package server hosts the frame debugger's user interface as a local web page.
//
// The debugger binary serves both the page (embedded, no build step, no network)
// and a WebSocket the page drives the emulator over. The browser earns its keep on
// the pixel side: it gets the frame's whole provenance buffer once, so hovering a
// pixel names the command that drew it with no round trip, and selecting a command
// can light up every pixel it touched.
//
// The server binds to loopback and accepts one session at a time — a DebugTarget is
// a single stateful machine, not something two pages can meaningfully share.
package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/wsock"
)

//go:embed web
var webFS embed.FS

// Server serves the debugger UI for one target.
type Server struct {
	tgt  debug.DebugTarget
	rom  string
	busy atomic.Bool // one session at a time
}

// New returns a Server for a target. rom is a label for the page's title bar.
func New(tgt debug.DebugTarget, rom string) *Server {
	return &Server{tgt: tgt, rom: rom}
}

// Handler is the debugger's HTTP surface: the embedded page, and the socket.
func (s *Server) Handler() http.Handler {
	web, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err) // the embed is compiled in; a failure here is a build bug
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(web)))
	mux.HandleFunc("/ws", s.serveWS)
	return mux
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	// A page on another origin must not be able to reach into a debugger running on
	// the developer's own machine.
	if !sameOrigin(r) {
		http.Error(w, "cross-origin websocket rejected", http.StatusForbidden)
		return
	}
	if !s.busy.CompareAndSwap(false, true) {
		http.Error(w, "a debugger session is already attached", http.StatusConflict)
		return
	}
	defer s.busy.Store(false)

	conn, err := wsock.Upgrade(w, r)
	if err != nil {
		log.Printf("framedbg: upgrade: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("framedbg: session attached")
	newSession(s.tgt, s.rom, conn).serve()
	log.Printf("framedbg: session ended")
}

// sameOrigin allows a request with no Origin (a non-browser client) and one whose
// Origin host matches the host it connected to.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

// ListenAndServe serves the debugger on addr and blocks. A bare port (":8088")
// binds loopback only: this exposes a machine's whole memory and deserves no
// listener on a public interface.
func (s *Server) ListenAndServe(addr string) error {
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Printf("framedbg: serving the debugger at http://%s\n", ln.Addr())
	return http.Serve(ln, s.Handler())
}
