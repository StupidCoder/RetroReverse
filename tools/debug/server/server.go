// Package server hosts the frame debugger's user interface as a local web page.
//
// The debugger binary serves both the page (embedded, no build step, no network) and
// a WebSocket the page drives the emulator over. The browser earns its keep on the
// pixel side: it gets a frame's whole provenance buffer once, so hovering a pixel
// names the command that drew it with no round trip, and selecting a command can
// light up every pixel it touched.
//
// A Runner owns the target and outlives any connection, so reloading the page does
// not throw the machine away. The target tells the page what it can do (see
// debug.Capabilities) and the page builds itself from that list — a platform without
// per-pixel provenance simply has no overdraw panel, rather than an empty one.
//
// The server binds to loopback: it exposes a machine's whole memory and deserves no
// listener on a public interface.
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

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/wsock"
)

//go:embed web
var webFS embed.FS

// Server serves the debugger UI.
type Server struct {
	ws *Workspace
}

// New returns a Server with a target already open. The library is discovered from the
// games/ tree if one can be found, so a session that started on one game can still
// switch to another.
func New(tgt debug.Target) *Server {
	s := &Server{ws: newWorkspace(debug.FindGamesRoot("."))}
	if tgt != nil {
		s.ws.setTarget(newRunner(tgt), nil)
	}
	return s
}

// NewLibrary returns a Server with nothing open: the page starts at the library.
func NewLibrary(gamesRoot string) *Server {
	return &Server{ws: newWorkspace(gamesRoot)}
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
	conn, err := wsock.Upgrade(w, r)
	if err != nil {
		log.Printf("framedbg: upgrade: %v", err)
		return
	}
	defer conn.Close()

	// Several connections may attach at once — a second tab, or the same tab after a
	// reload. They all watch one machine; only its runner touches it.
	log.Printf("framedbg: session attached")
	(&session{conn: conn, ws: s.ws}).serve()
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

// ListenAndServe serves the debugger on addr and blocks. A bare port (":8088") binds
// loopback only.
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
