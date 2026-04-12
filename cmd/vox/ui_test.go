package main

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cdimoush/vox/daemon/ipc"
)

// fakeHandler implements ipc.Handler so we can drive the CLI subcommands
// against a real socket without spawning the daemon binary.
type fakeHandler struct {
	toggles  int32
	state    string
	shutdown chan struct{}
	once     sync.Once
}

func newFakeHandler() *fakeHandler {
	return &fakeHandler{state: "idle", shutdown: make(chan struct{})}
}

func (f *fakeHandler) Toggle()       { atomic.AddInt32(&f.toggles, 1) }
func (f *fakeHandler) Cancel()       {}
func (f *fakeHandler) State() string { return f.state }
func (f *fakeHandler) Shutdown()     { f.once.Do(func() { close(f.shutdown) }) }

// pointAllAt redirects socket + pidfile paths at a temp dir. Returns the
// socket path so tests can stand up their own server.
func pointAllAt(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("HOME", dir)
	// ~/.vox must exist so cmdUIStatus's pidPath lookup succeeds without
	// erroring on a missing parent.
	_ = os.MkdirAll(filepath.Join(dir, ".vox"), 0o755)
	return filepath.Join(dir, "vox.sock")
}

func TestStatusWhenNotRunning(t *testing.T) {
	pointAllAt(t)
	os.Args = []string{"vox", "ui", "status"}
	if err := cmdUIStatus(); err != nil {
		t.Fatalf("status: %v", err)
	}
}

func TestToggleFailsWhenNoDaemon(t *testing.T) {
	pointAllAt(t)
	os.Args = []string{"vox", "ui", "toggle"}
	if err := cmdUIToggle(); err == nil {
		t.Fatal("want error when daemon absent")
	}
}

func TestToggleHitsRunningDaemon(t *testing.T) {
	sockPath := pointAllAt(t)
	h := newFakeHandler()
	srv, err := ipc.Listen(sockPath, h)
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()
	defer srv.Stop()

	// Seed a pidfile so status sees the "daemon" as live (our own pid is
	// guaranteed to exist).
	pidF, _ := pidPath()
	os.WriteFile(pidF, []byte(strconv.Itoa(os.Getpid())), 0o600)

	os.Args = []string{"vox", "ui", "toggle"}
	if err := cmdUIToggle(); err != nil {
		t.Fatalf("toggle: %v", err)
	}
	if atomic.LoadInt32(&h.toggles) != 1 {
		t.Errorf("want 1 toggle, got %d", h.toggles)
	}
}

func TestStatusSeesRunningDaemon(t *testing.T) {
	sockPath := pointAllAt(t)
	h := newFakeHandler()
	h.state = "recording"
	srv, err := ipc.Listen(sockPath, h)
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve()
	defer srv.Stop()

	pidF, _ := pidPath()
	os.WriteFile(pidF, []byte(strconv.Itoa(os.Getpid())), 0o600)

	// cmdUIStatus prints to stdout; we just verify it doesn't error and
	// successfully spoke to the server.
	if err := cmdUIStatus(); err != nil {
		t.Fatalf("status: %v", err)
	}
}

func TestStopWhenNotRunningIsNoOp(t *testing.T) {
	pointAllAt(t)
	if err := cmdUIStop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
}

func TestWaitForSocketTimeout(t *testing.T) {
	pointAllAt(t)
	start := time.Now()
	err := waitForSocket("/tmp/definitely-not-a-socket-"+t.Name(), 150*time.Millisecond)
	if err == nil {
		t.Fatal("want timeout")
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Errorf("returned too fast: %v", elapsed)
	}
}

func TestReadLivePIDRejectsGarbage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pid")
	os.WriteFile(path, []byte("not-a-pid"), 0o600)
	if _, ok := readLivePID(path); ok {
		t.Error("garbage pidfile should not report live")
	}
}
