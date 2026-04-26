// Package daemon implements the vox UI daemon's state machine and lifecycle.
//
// The daemon owns four concerns:
//  1. A state machine (idle → recording → transcribing → done → idle).
//  2. Event emission to subscribers (overlay, tray, audio cues).
//  3. Coordination of Recorder, Transcriber, and Paster — all injectable
//     interfaces so the state machine is testable without audio hardware
//     or network.
//
// This file has NO platform-specific imports. All display/audio/hotkey
// hookups live behind interfaces and the `ui` build tag.
package daemon

import (
	"context"
	"fmt"
	"sync"
)

// State is the current phase of the dictation lifecycle.
type State int

const (
	StateIdle State = iota
	StateRecording
	StateTranscribing
	StateDone
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRecording:
		return "recording"
	case StateTranscribing:
		return "transcribing"
	case StateDone:
		return "done"
	}
	return "unknown"
}

// Event is emitted on every state transition.
type Event struct {
	From State
	To   State
	// Text is populated only on Done transitions (the transcribed result).
	Text string
	// Err is populated when a transition is caused by a failure; the
	// machine always returns to Idle after surfacing the error.
	Err error
}

// Recorder captures audio. The real implementation wraps recorder.Record;
// tests inject a fake that just returns a preset path or error.
type Recorder interface {
	// Record blocks until ctx is cancelled, then returns the path to a
	// finalized WAV file. The caller is responsible for deletion.
	Record(ctx context.Context) (string, error)
}

// Transcriber converts a WAV path to text.
type Transcriber interface {
	Transcribe(ctx context.Context, wavPath string) (string, error)
}

// Paster writes text to the clipboard and (best-effort) simulates paste.
type Paster interface {
	Paste(text string) error
}

// Daemon is the state machine. Construct with New, call Toggle/Cancel
// from any goroutine. Subscribe once via Events before Run.
type Daemon struct {
	rec Recorder
	tx  Transcriber
	pas Paster

	mu    sync.Mutex
	state State
	// recCancel cancels the current recording context (set while in
	// StateRecording).
	recCancel context.CancelFunc
	// txCancel cancels the current transcription context (set while
	// in StateTranscribing).
	txCancel context.CancelFunc

	events chan Event
	subs   []chan Event // additional subscriber channels (from Subscribe)
}

// New builds a daemon. All three dependencies are required.
func New(rec Recorder, tx Transcriber, pas Paster) *Daemon {
	return &Daemon{
		rec:    rec,
		tx:     tx,
		pas:    pas,
		state:  StateIdle,
		events: make(chan Event, 16),
	}
}

// State returns the current state (snapshot).
func (d *Daemon) State() State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// Events returns the primary receive-only channel of transitions. Subscribers
// must drain the channel — slow consumers drop events (the channel is buffered
// but non-blocking on send). For additional consumers, use Subscribe.
func (d *Daemon) Events() <-chan Event {
	return d.events
}

// Subscribe creates and returns a new event channel. Each subscriber
// receives its own copy of every transition. The channel is closed on
// Shutdown. Safe to call from any goroutine before or after Run.
func (d *Daemon) Subscribe() <-chan Event {
	d.mu.Lock()
	defer d.mu.Unlock()
	ch := make(chan Event, 16)
	d.subs = append(d.subs, ch)
	return ch
}

// Toggle advances the state machine:
//   - idle → recording (starts rec goroutine)
//   - recording → transcribing → done → idle (stops rec; the rec goroutine
//     drives the remaining transitions)
//   - transcribing → idle (cancels tx; treated as user abort)
//
// Safe to call from any goroutine.
func (d *Daemon) Toggle() {
	d.mu.Lock()
	switch d.state {
	case StateIdle:
		d.startRecordingLocked()
	case StateRecording:
		// Signal rec goroutine to stop; it will drive the rest.
		if d.recCancel != nil {
			d.recCancel()
		}
		d.mu.Unlock()
	case StateTranscribing:
		// Second toggle during transcription = abort.
		if d.txCancel != nil {
			d.txCancel()
		}
		d.mu.Unlock()
	default:
		d.mu.Unlock()
	}
}

// Cancel forces an immediate return to idle from any state.
func (d *Daemon) Cancel() {
	d.mu.Lock()
	if d.recCancel != nil {
		d.recCancel()
	}
	if d.txCancel != nil {
		d.txCancel()
	}
	d.mu.Unlock()
}

// Shutdown closes the events channel and all subscriber channels. Call once, last.
func (d *Daemon) Shutdown() {
	d.Cancel()
	d.mu.Lock()
	defer d.mu.Unlock()
	// Only close once.
	select {
	case <-d.events:
	default:
	}
	close(d.events)
	for _, ch := range d.subs {
		close(ch)
	}
	d.subs = nil
}

// startRecordingLocked must be called with mu held. It flips state to
// Recording, launches the record→transcribe→paste goroutine, and releases
// the lock before returning.
func (d *Daemon) startRecordingLocked() {
	ctx, cancel := context.WithCancel(context.Background())
	d.recCancel = cancel
	d.transitionLocked(StateRecording, "", nil)
	d.mu.Unlock()

	go d.runSession(ctx)
}

// runSession drives the full lifecycle from a given recording context.
// It's the only place that writes state after startRecordingLocked.
func (d *Daemon) runSession(recCtx context.Context) {
	wavPath, err := d.rec.Record(recCtx)

	d.mu.Lock()
	d.recCancel = nil
	if err != nil {
		d.transitionLocked(StateIdle, "", fmt.Errorf("recording: %w", err))
		d.mu.Unlock()
		return
	}
	// Move to transcribing.
	txCtx, txCancel := context.WithCancel(context.Background())
	d.txCancel = txCancel
	d.transitionLocked(StateTranscribing, "", nil)
	d.mu.Unlock()

	text, err := d.tx.Transcribe(txCtx, wavPath)

	d.mu.Lock()
	d.txCancel = nil
	if err != nil {
		d.transitionLocked(StateIdle, "", fmt.Errorf("transcribe: %w", err))
		d.mu.Unlock()
		return
	}
	d.mu.Unlock()

	// Paste outside the lock — it may shell out to xdotool and we don't
	// want to block Toggle callers.
	pasteErr := d.pas.Paste(text)

	d.mu.Lock()
	d.transitionLocked(StateDone, text, pasteErr)
	d.transitionLocked(StateIdle, "", nil)
	d.mu.Unlock()
}

// transitionLocked updates state and emits an event to the primary channel
// and all subscribers. mu must be held.
func (d *Daemon) transitionLocked(to State, text string, err error) {
	ev := Event{From: d.state, To: to, Text: text, Err: err}
	d.state = to
	// Non-blocking send — slow consumers lose events.
	select {
	case d.events <- ev:
	default:
	}
	for _, ch := range d.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
