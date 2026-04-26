package ipc

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeHandler struct {
	mu          sync.Mutex
	toggles     int32
	cancels     int32
	shutdowns   int32
	state       string
	shutdownCh  chan struct{}
}

func newFakeHandler() *fakeHandler {
	return &fakeHandler{state: "idle", shutdownCh: make(chan struct{})}
}

func (f *fakeHandler) Toggle()       { atomic.AddInt32(&f.toggles, 1) }
func (f *fakeHandler) Cancel()       { atomic.AddInt32(&f.cancels, 1) }
func (f *fakeHandler) State() string { f.mu.Lock(); defer f.mu.Unlock(); return f.state }
func (f *fakeHandler) Shutdown() {
	if atomic.AddInt32(&f.shutdowns, 1) == 1 {
		close(f.shutdownCh)
	}
}

func startServer(t *testing.T, h Handler) *Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vox.sock")
	srv, err := Listen(path, h)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go srv.Serve()
	t.Cleanup(func() { srv.Stop() })
	return srv
}

func TestDispatchToggleAndStatus(t *testing.T) {
	h := newFakeHandler()
	srv := startServer(t, h)

	resp, err := Dial(srv.Path(), Request{Cmd: CmdToggle})
	if err != nil {
		t.Fatalf("dial toggle: %v", err)
	}
	if !resp.OK || resp.State != "idle" {
		t.Errorf("toggle resp: %+v", resp)
	}
	if atomic.LoadInt32(&h.toggles) != 1 {
		t.Errorf("want 1 toggle, got %d", h.toggles)
	}

	resp, err = Dial(srv.Path(), Request{Cmd: CmdStatus})
	if err != nil {
		t.Fatalf("dial status: %v", err)
	}
	if !resp.OK || resp.State != "idle" {
		t.Errorf("status resp: %+v", resp)
	}
}

func TestDispatchCancel(t *testing.T) {
	h := newFakeHandler()
	srv := startServer(t, h)

	if _, err := Dial(srv.Path(), Request{Cmd: CmdCancel}); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if atomic.LoadInt32(&h.cancels) != 1 {
		t.Errorf("want 1 cancel, got %d", h.cancels)
	}
}

func TestDispatchUnknownCmd(t *testing.T) {
	h := newFakeHandler()
	srv := startServer(t, h)

	resp, err := Dial(srv.Path(), Request{Cmd: "bogus"})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if resp.OK || resp.Err == "" {
		t.Errorf("want failure, got %+v", resp)
	}
}

func TestShutdownReplies(t *testing.T) {
	h := newFakeHandler()
	srv := startServer(t, h)

	resp, err := Dial(srv.Path(), Request{Cmd: CmdShutdown})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if !resp.OK {
		t.Errorf("shutdown resp not ok: %+v", resp)
	}
	// Shutdown runs in a goroutine on the server side; give it a beat.
	select {
	case <-h.shutdownCh:
	case <-time.After(time.Second):
		t.Fatal("handler.Shutdown not called")
	}
}

func TestListenRefusesDoubleBind(t *testing.T) {
	h := newFakeHandler()
	path := filepath.Join(t.TempDir(), "vox.sock")

	srv1, err := Listen(path, h)
	if err != nil {
		t.Fatal(err)
	}
	go srv1.Serve()
	defer srv1.Stop()

	_, err = Listen(path, h)
	if err == nil {
		t.Fatal("expected Listen on live socket to fail")
	}
}

func TestListenReclaimsStaleSocket(t *testing.T) {
	h := newFakeHandler()
	path := filepath.Join(t.TempDir(), "vox.sock")

	// First daemon lives and dies.
	srv1, err := Listen(path, h)
	if err != nil {
		t.Fatal(err)
	}
	srv1.Stop()

	// The socket file has been removed on clean Stop, so this is a real
	// no-previous-run path. Simulate a stale file:
	if err := writeStaleFile(path); err != nil {
		t.Fatal(err)
	}

	srv2, err := Listen(path, h)
	if err != nil {
		t.Fatalf("want stale cleanup to succeed, got: %v", err)
	}
	srv2.Stop()
}

// writeStaleFile creates a socket-looking file that nothing is bound to.
func writeStaleFile(path string) error {
	f, err := openTruncate(path)
	if err != nil {
		return err
	}
	return f.Close()
}
