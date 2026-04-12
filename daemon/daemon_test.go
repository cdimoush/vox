package daemon

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeRecorder blocks on ctx and returns wavPath (or recordErr) when cancelled.
type fakeRecorder struct {
	wavPath   string
	recordErr error
	// started is closed once Record has begun — lets tests wait for the
	// machine to reach Recording before sending Toggle.
	started chan struct{}
	once    sync.Once
}

func (f *fakeRecorder) Record(ctx context.Context) (string, error) {
	f.once.Do(func() { close(f.started) })
	<-ctx.Done()
	if f.recordErr != nil {
		return "", f.recordErr
	}
	return f.wavPath, nil
}

type fakeTranscriber struct {
	text string
	err  error
	// block, if non-nil, causes Transcribe to wait on ctx.Done() as well —
	// used for the cancel-during-transcribing test.
	block bool
}

func (f *fakeTranscriber) Transcribe(ctx context.Context, _ string) (string, error) {
	if f.block {
		<-ctx.Done()
		return "", ctx.Err()
	}
	return f.text, f.err
}

type fakePaster struct {
	got string
	err error
}

func (f *fakePaster) Paste(text string) error {
	f.got = text
	return f.err
}

func newFakes() (*fakeRecorder, *fakeTranscriber, *fakePaster) {
	return &fakeRecorder{wavPath: "/tmp/fake.wav", started: make(chan struct{})},
		&fakeTranscriber{text: "hello world"},
		&fakePaster{}
}

// collect drains events until the machine returns to Idle after having
// entered Recording. Returns the ordered list of observed transitions.
func collect(t *testing.T, d *Daemon, timeout time.Duration) []Event {
	t.Helper()
	var out []Event
	deadline := time.After(timeout)
	sawRecording := false
	for {
		select {
		case ev, ok := <-d.Events():
			if !ok {
				return out
			}
			out = append(out, ev)
			if ev.To == StateRecording {
				sawRecording = true
			}
			if sawRecording && ev.To == StateIdle {
				return out
			}
		case <-deadline:
			t.Fatalf("timed out collecting events; got %d so far: %+v", len(out), out)
		}
	}
}

func TestHappyPath(t *testing.T) {
	rec, tx, pas := newFakes()
	d := New(rec, tx, pas)

	d.Toggle() // idle → recording

	<-rec.started
	d.Toggle() // recording → (transcribe → done → idle)

	events := collect(t, d, 2*time.Second)

	wantSequence := []State{StateRecording, StateTranscribing, StateDone, StateIdle}
	if len(events) != len(wantSequence) {
		t.Fatalf("want %d events, got %d: %+v", len(wantSequence), len(events), events)
	}
	for i, want := range wantSequence {
		if events[i].To != want {
			t.Errorf("event %d: want To=%s, got To=%s", i, want, events[i].To)
		}
	}
	if pas.got != "hello world" {
		t.Errorf("paste text: want 'hello world', got %q", pas.got)
	}
	if d.State() != StateIdle {
		t.Errorf("final state: want idle, got %s", d.State())
	}
}

func TestCancelDuringTranscription(t *testing.T) {
	rec, _, pas := newFakes()
	tx := &fakeTranscriber{block: true}
	d := New(rec, tx, pas)

	d.Toggle()
	<-rec.started
	d.Toggle() // stop recording → enter transcribing

	// Wait until we observe the Transcribing event, then toggle again.
	timeout := time.After(2 * time.Second)
	for {
		select {
		case ev := <-d.Events():
			if ev.To == StateTranscribing {
				goto abort
			}
		case <-timeout:
			t.Fatal("never reached transcribing")
		}
	}
abort:
	d.Toggle() // transcribing → idle (abort)

	// Expect one more event with To=Idle and an Err (ctx canceled).
	select {
	case ev := <-d.Events():
		if ev.To != StateIdle {
			t.Errorf("want final Idle, got %s", ev.To)
		}
		if ev.Err == nil {
			t.Errorf("want Err on aborted transcribe, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no final event after abort")
	}
}

func TestRecordingFailureReturnsToIdle(t *testing.T) {
	rec, tx, pas := newFakes()
	rec.recordErr = errors.New("mic boom")
	d := New(rec, tx, pas)

	d.Toggle()
	<-rec.started
	d.Toggle()

	// Expect recording event then idle-with-err.
	ev1 := <-d.Events()
	if ev1.To != StateRecording {
		t.Fatalf("first event: want recording, got %s", ev1.To)
	}
	ev2 := <-d.Events()
	if ev2.To != StateIdle || ev2.Err == nil {
		t.Errorf("want idle+err, got %+v", ev2)
	}
	if pas.got != "" {
		t.Errorf("paster should not run on rec failure, got %q", pas.got)
	}
}

func TestToggleFromIdleOnlyStartsOnce(t *testing.T) {
	rec, tx, pas := newFakes()
	d := New(rec, tx, pas)

	d.Toggle()
	<-rec.started
	// A second Toggle before we've left Recording should stop, not spawn
	// a new session. We verify by counting Recording events in the log.
	d.Toggle()

	events := collect(t, d, 2*time.Second)
	recordingCount := 0
	for _, ev := range events {
		if ev.To == StateRecording {
			recordingCount++
		}
	}
	if recordingCount != 1 {
		t.Errorf("want exactly 1 recording transition, got %d", recordingCount)
	}
}
