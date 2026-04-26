//go:build ui

// Package overlay draws a floating pill-shaped window showing the
// daemon's recording state and (optionally) audio level bars.
//
// The overlay is hidden when idle — on composited desktops this is
// achieved by painting a fully transparent frame. On non-composited
// setups the window may appear as a small black rectangle; the user
// can disable the overlay via config.
package overlay

import (
	"image"
	"image/color"
	"log"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/cdimoush/vox/daemon"
)

const (
	doneFlashDuration = 2 * time.Second
)

// State colors.
var (
	bgDark        = color.NRGBA{R: 30, G: 30, B: 30, A: 220}
	bgDone        = color.NRGBA{R: 40, G: 160, B: 80, A: 220}
	dotRecording  = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
	dotTranscribe = color.NRGBA{R: 50, G: 120, B: 220, A: 255}
	textWhite     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	barFill       = color.NRGBA{R: 255, G: 255, B: 255, A: 180}
)

// shared is the thread-safe state written by waker goroutines and
// read by the Gio frame loop.
type shared struct {
	mu       sync.Mutex
	phase    daemon.State
	text     string    // transcript (set on Done)
	level    float64   // mic RMS 0.0–1.0
	doneAt   time.Time // when Done was received
	showDone bool      // true while done flash is visible
}

func (sh *shared) onEvent(ev daemon.Event) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.phase = ev.To
	if ev.To == daemon.StateDone {
		sh.text = ev.Text
		sh.doneAt = time.Now()
		sh.showDone = true
	}
	if ev.To == daemon.StateIdle && !sh.showDone {
		sh.text = ""
	}
}

func (sh *shared) onLevel(l float64) {
	sh.mu.Lock()
	sh.level = l
	sh.mu.Unlock()
}

type snapshot struct {
	phase    daemon.State
	text     string
	level    float64
	doneAt   time.Time
	showDone bool
	visible  bool
}

func (sh *shared) snap() snapshot {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	vis := false
	switch sh.phase {
	case daemon.StateRecording, daemon.StateTranscribing, daemon.StateDone:
		vis = true
	case daemon.StateIdle:
		if sh.showDone && time.Since(sh.doneAt) < doneFlashDuration {
			vis = true
		} else {
			sh.showDone = false
			sh.text = ""
		}
	}
	return snapshot{
		phase:    sh.phase,
		text:     sh.text,
		level:    sh.level,
		doneAt:   sh.doneAt,
		showDone: sh.showDone,
		visible:  vis,
	}
}

// Run creates the overlay window, subscribes to daemon events, and
// blocks in the Gio event loop. Call from a goroutine.
//
// levels may be nil if the mic-level monitor is not wired up yet; the
// overlay will show a static recording indicator instead of live bars.
//
// On macOS the Gio window requires the main thread — this goroutine
// approach works on Linux only. macOS support will need daemon
// restructuring so Gio owns the main thread.
func Run(d *daemon.Daemon, levels <-chan float64, width, height int) {
	w := new(app.Window)
	w.Option(
		app.Size(unit.Dp(width), unit.Dp(height)),
		app.MinSize(unit.Dp(width), unit.Dp(height)),
		app.MaxSize(unit.Dp(width), unit.Dp(height)),
		app.Decorated(false),
		app.Title("vox"),
	)

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var ops op.Ops
	var sh shared

	// Waker goroutines: update shared state + poke the Gio window.
	// w.Invalidate() is safe to call from any goroutine.
	events := d.Subscribe()
	go func() {
		for ev := range events {
			sh.onEvent(ev)
			w.Invalidate()
			// When entering done/idle, schedule another invalidation
			// after the flash duration so the hide transition fires.
			if ev.To == daemon.StateDone || ev.To == daemon.StateIdle {
				go func() {
					time.Sleep(doneFlashDuration + 100*time.Millisecond)
					w.Invalidate()
				}()
			}
		}
	}()
	if levels != nil {
		go func() {
			for l := range levels {
				sh.onLevel(l)
				w.Invalidate()
			}
		}()
	}

	log.Println("overlay: window created")

	wasVisible := true // starts mapped by Gio; first idle frame will unmap it

	for {
		e := w.Event()

		// Let platform-specific code handle view events (X11 hints, etc).
		handleViewEvent(e, width, height)

		switch e := e.(type) {
		case app.FrameEvent:
			s := sh.snap()

			// Show/hide the actual X11 window on visibility transitions.
			if s.visible && !wasVisible {
				platformShow()
			} else if !s.visible && wasVisible {
				platformHide()
			}
			wasVisible = s.visible

			gtx := app.NewContext(&ops, e)
			drawFrame(gtx, th, &s)
			e.Frame(gtx.Ops)

			// Keep redrawing while visible (for level bar animation).
			if s.visible {
				w.Invalidate()
			}

		case app.DestroyEvent:
			log.Println("overlay: window destroyed")
			return
		}
	}
}

func drawFrame(gtx layout.Context, th *material.Theme, s *snapshot) layout.Dimensions {
	size := gtx.Constraints.Max
	if !s.visible {
		return layout.Dimensions{Size: size}
	}

	radius := size.Y / 2

	// Background pill.
	bg := bgDark
	if s.showDone {
		bg = bgDone
	} else if s.phase == daemon.StateDone {
		bg = bgDone
	}
	rrect := clip.UniformRRect(image.Rectangle{Max: size}, radius)
	paint.FillShape(gtx.Ops, bg, rrect.Op(gtx.Ops))

	// Dot indicator (8px circle, left-aligned).
	const dotR = 4
	const dotMarginL = 14
	dotCX := dotMarginL + dotR
	dotCY := size.Y / 2
	dotRect := image.Rect(dotCX-dotR, dotCY-dotR, dotCX+dotR, dotCY+dotR)
	dotClip := clip.UniformRRect(dotRect, dotR)
	dotColor := dotRecording
	switch {
	case s.phase == daemon.StateTranscribing:
		dotColor = dotTranscribe
	case s.showDone || s.phase == daemon.StateDone:
		dotColor = textWhite
	}
	paint.FillShape(gtx.Ops, dotColor, dotClip.Op(gtx.Ops))

	// Content area: after dot.
	contentX := dotCX + dotR + 10
	contentW := size.X - contentX - 12

	switch {
	case s.phase == daemon.StateRecording:
		drawBars(gtx.Ops, contentX, size.Y, contentW, s.level)
	case s.phase == daemon.StateTranscribing:
		drawLabel(gtx, th, contentX, size, contentW, "Transcribing...")
	case s.showDone || s.phase == daemon.StateDone:
		txt := s.text
		if len(txt) > 40 {
			txt = txt[:40] + "..."
		}
		if txt == "" {
			txt = "Done"
		}
		drawLabel(gtx, th, contentX, size, contentW, txt)
	}

	return layout.Dimensions{Size: size}
}

// drawBars renders an audio-level visualisation: a row of thin vertical
// bars whose height scales with the current RMS level.
func drawBars(ops *op.Ops, x, h, w int, level float64) {
	const barW = 3
	const gap = 2
	const padY = 8
	maxH := h - padY*2

	n := w / (barW + gap)
	if n < 1 {
		n = 1
	}

	for i := 0; i < n; i++ {
		frac := level
		if i%2 == 0 {
			frac *= 0.8
		}
		if frac < 0.05 {
			frac = 0.05
		}
		if frac > 1 {
			frac = 1
		}

		barH := int(float64(maxH) * frac)
		if barH < 2 {
			barH = 2
		}

		bx := x + i*(barW+gap)
		by := h/2 - barH/2
		r := clip.UniformRRect(image.Rect(bx, by, bx+barW, by+barH), 1)
		paint.FillShape(ops, barFill, r.Op(ops))
	}
}

// drawLabel renders a single line of text inside the content area.
func drawLabel(gtx layout.Context, th *material.Theme, x int, size image.Point, w int, txt string) {
	defer op.Offset(image.Pt(x, 0)).Push(gtx.Ops).Pop()

	gtx.Constraints = layout.Exact(image.Pt(w, size.Y))
	lbl := material.Body2(th, txt)
	lbl.Color = textWhite
	lbl.MaxLines = 1
	layout.Center.Layout(gtx, lbl.Layout)
}
