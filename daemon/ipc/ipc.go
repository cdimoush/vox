// Package ipc implements the Unix-domain-socket JSON protocol between the
// vox CLI and the running daemon.
//
// Wire format: one JSON object per line. Requests carry a "cmd" and
// optional "args"; responses carry "ok" and an optional "state" or "err".
// The protocol is deliberately tiny — every message fits in one syscall.
package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// Request is a command sent from CLI → daemon.
type Request struct {
	Cmd  string          `json:"cmd"`
	Args json.RawMessage `json:"args,omitempty"`
}

// Response is the daemon's reply.
type Response struct {
	OK    bool   `json:"ok"`
	State string `json:"state,omitempty"`
	Err   string `json:"err,omitempty"`
}

// Known commands.
const (
	CmdToggle   = "toggle"
	CmdCancel   = "cancel"
	CmdStatus   = "status"
	CmdShutdown = "shutdown"
)

// Handler is implemented by whatever owns the state machine (the Daemon).
// Returning an error surfaces as {ok:false, err:"..."} on the wire.
type Handler interface {
	Toggle()
	Cancel()
	State() string
	Shutdown()
}

// DefaultSocketPath returns the preferred socket location:
// $XDG_RUNTIME_DIR/vox.sock if set, else /tmp/vox-<uid>.sock.
func DefaultSocketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "vox.sock")
	}
	return filepath.Join(os.TempDir(), "vox-"+strconv.Itoa(os.Getuid())+".sock")
}

// Server owns the listener and accept loop. Stop idempotently closes the
// socket and removes the on-disk file.
type Server struct {
	path string
	ln   net.Listener
	h    Handler

	mu     sync.Mutex
	closed bool
}

// Listen creates the socket at path, removing any stale file first.
func Listen(path string, h Handler) (*Server, error) {
	// Clean up a stale socket from a previous crashed daemon. This is
	// safe: if another daemon is actually bound, the subsequent Listen
	// below would succeed — but we test-connect first to avoid killing
	// a live peer by accident.
	if _, err := os.Stat(path); err == nil {
		if conn, dialErr := net.Dial("unix", path); dialErr == nil {
			conn.Close()
			return nil, fmt.Errorf("socket %s already in use by another daemon", path)
		}
		_ = os.Remove(path)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", path, err)
	}
	// Tighten perms — only the owning user should be able to talk to the
	// daemon. net.Listen on Unix sockets respects umask, so chmod after.
	_ = os.Chmod(path, 0o600)
	return &Server{path: path, ln: ln, h: h}, nil
}

// Path returns the socket path (for logging / status checks).
func (s *Server) Path() string { return s.path }

// Serve runs the accept loop. Returns when Stop is called or the listener
// fails irrecoverably.
func (s *Server) Serve() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			return fmt.Errorf("accept: %w", err)
		}
		go s.handle(conn)
	}
}

// Stop closes the listener and removes the socket file. Safe to call
// multiple times.
func (s *Server) Stop() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	err := s.ln.Close()
	_ = os.Remove(s.path)
	return err
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(bufio.NewReader(conn))
	enc := json.NewEncoder(conn)

	var req Request
	if err := dec.Decode(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			_ = enc.Encode(Response{OK: false, Err: "bad request: " + err.Error()})
		}
		return
	}

	resp := s.dispatch(req)
	_ = enc.Encode(resp)
}

func (s *Server) dispatch(req Request) Response {
	switch req.Cmd {
	case CmdToggle:
		s.h.Toggle()
		return Response{OK: true, State: s.h.State()}
	case CmdCancel:
		s.h.Cancel()
		return Response{OK: true, State: s.h.State()}
	case CmdStatus:
		return Response{OK: true, State: s.h.State()}
	case CmdShutdown:
		// Reply before shutting down so the client doesn't see a
		// truncated response.
		state := s.h.State()
		go s.h.Shutdown()
		return Response{OK: true, State: state}
	default:
		return Response{OK: false, Err: "unknown cmd: " + req.Cmd}
	}
}

// Dial is a convenience used by the CLI side to send a single command and
// read a single response. Always closes the connection.
func Dial(path string, req Request) (Response, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return Response{}, err
	}
	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return Response{}, err
	}
	return resp, nil
}
