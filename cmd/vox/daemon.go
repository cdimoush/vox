package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cdimoush/vox/daemon"
	"github.com/cdimoush/vox/daemon/config"
	"github.com/cdimoush/vox/daemon/ipc"
	"github.com/cdimoush/vox/daemon/paste"
	"github.com/cdimoush/vox/history"
	"github.com/cdimoush/vox/recorder"
	"github.com/cdimoush/vox/transcribe"
)

// cmdDaemon is the hidden `vox __daemon` entrypoint invoked by `vox ui
// start`. Display-bound subsystems (overlay, tray, hotkey, audio cues,
// mic-level streaming) are NOT wired in here — see HANDOFF.md for the
// Phase-2/3 pieces that a human with a display + mic must add.
func cmdDaemon() error {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("vox daemon starting")

	// --- Config ----------------------------------------------------------
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return fmt.Errorf("config path: %w", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log.Printf("config: paste.method=%s auto_paste=%v overlay.enabled=%v",
		cfg.Paste.Method, cfg.Paste.AutoPaste, cfg.Overlay.Enabled)

	// --- State machine ---------------------------------------------------
	rec := recorderAdapter{}
	tx := transcriberAdapter{}
	pas := pasterAdapter{method: paste.Method(cfg.Paste.Method), auto: cfg.Paste.AutoPaste}

	d := daemon.New(rec, tx, pas)
	go logEvents(d) // also appends successful transcriptions to history

	// --- UI subsystems (hotkey, overlay, tray) ---------------------------
	// Built behind the `ui` tag; no-op in headless builds.
	quit := make(chan struct{})
	startUISubsystems(d, cfg, quit)

	// --- IPC -------------------------------------------------------------
	handler := &handlerAdapter{d: d, quit: quit}
	srv, err := ipc.Listen(ipc.DefaultSocketPath(), handler)
	if err != nil {
		return fmt.Errorf("ipc listen: %w", err)
	}
	defer srv.Stop()
	log.Printf("listening on %s", srv.Path())

	// --- Signals ---------------------------------------------------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve() }()

	select {
	case s := <-sigCh:
		log.Printf("caught %s, shutting down", s)
	case <-quit:
		log.Println("shutdown command received")
	case err := <-errCh:
		if err != nil {
			log.Printf("serve: %v", err)
		}
	}

	// Give Shutdown a moment to drain any in-flight session.
	d.Shutdown()
	_ = srv.Stop()
	_ = cleanupPidfile()
	time.Sleep(50 * time.Millisecond)
	log.Println("vox daemon exited cleanly")
	return nil
}

// cleanupPidfile removes ~/.vox/daemon.pid on clean exit. Best-effort.
func cleanupPidfile() error {
	p, err := pidPath()
	if err != nil {
		return err
	}
	return os.Remove(p)
}

// logEvents prints transitions and appends successful runs to history.
func logEvents(d *daemon.Daemon) {
	store := history.NewStore(history.DefaultPath())
	for ev := range d.Events() {
		if ev.Err != nil {
			log.Printf("%s → %s (err: %v)", ev.From, ev.To, ev.Err)
			continue
		}
		log.Printf("%s → %s", ev.From, ev.To)
		if ev.To == daemon.StateDone && ev.Text != "" {
			_ = store.Append(history.Entry{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Text:      ev.Text,
			})
		}
	}
}

// --- Adapters ----------------------------------------------------------------

// recorderAdapter bridges recorder.Record to daemon.Recorder.
type recorderAdapter struct{}

func (recorderAdapter) Record(ctx context.Context) (string, error) {
	r, err := recorder.Record(ctx)
	if err != nil {
		return "", err
	}
	return r.FilePath, nil
}

// transcriberAdapter bridges transcribe.Transcribe to daemon.Transcriber
// and cleans up the WAV file after the call.
type transcriberAdapter struct{}

func (transcriberAdapter) Transcribe(ctx context.Context, wavPath string) (string, error) {
	defer os.Remove(wavPath)
	text, _, err := transcribe.Transcribe(ctx, wavPath)
	return text, err
}

// pasterAdapter wraps paste.Injector with respect for the auto_paste flag.
type pasterAdapter struct {
	method paste.Method
	auto   bool
}

func (p pasterAdapter) Paste(text string) error {
	effective := p.method
	if !p.auto {
		effective = paste.MethodClipboardOnly
	}
	return paste.NewInjector(effective).Paste(text)
}

// handlerAdapter makes *daemon.Daemon satisfy ipc.Handler. Shutdown
// signals the main loop via quit; the main loop then stops the IPC
// server and the state machine in the right order.
type handlerAdapter struct {
	d    *daemon.Daemon
	quit chan struct{}
}

func (h *handlerAdapter) Toggle()       { h.d.Toggle() }
func (h *handlerAdapter) Cancel()       { h.d.Cancel() }
func (h *handlerAdapter) State() string { return h.d.State().String() }
func (h *handlerAdapter) Shutdown() {
	select {
	case <-h.quit:
	default:
		close(h.quit)
	}
}
