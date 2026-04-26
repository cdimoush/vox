//go:build ui && linux

package overlay

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
#include <string.h>

// setOverlayHints configures the X11 window to behave as a non-focusable
// overlay: always on top, skip taskbar/pager, no input focus, notification type.
void setOverlayHints(Display *dpy, Window win) {
	// 1. Window type = notification (doesn't take focus, no decorations).
	Atom wmType = XInternAtom(dpy, "_NET_WM_WINDOW_TYPE", False);
	Atom typeNotif = XInternAtom(dpy, "_NET_WM_WINDOW_TYPE_NOTIFICATION", False);
	XChangeProperty(dpy, win, wmType, XA_ATOM, 32, PropModeReplace,
		(unsigned char *)&typeNotif, 1);

	// 2. State: above + skip taskbar + skip pager.
	Atom wmState = XInternAtom(dpy, "_NET_WM_STATE", False);
	Atom above   = XInternAtom(dpy, "_NET_WM_STATE_ABOVE", False);
	Atom skipTB  = XInternAtom(dpy, "_NET_WM_STATE_SKIP_TASKBAR", False);
	Atom skipP   = XInternAtom(dpy, "_NET_WM_STATE_SKIP_PAGER", False);
	Atom states[] = {above, skipTB, skipP};
	XChangeProperty(dpy, win, wmState, XA_ATOM, 32, PropModeReplace,
		(unsigned char *)states, 3);

	// 3. WM hints: don't accept input focus.
	XWMHints *hints = XAllocWMHints();
	hints->flags = InputHint;
	hints->input = False;
	XSetWMHints(dpy, win, hints);
	XFree(hints);

	// 4. Start unmapped (hidden) — the overlay shows on first toggle.
	XUnmapWindow(dpy, win);
	XFlush(dpy);
}

// positionOverlay moves the window to top-center of the screen.
void positionOverlay(Display *dpy, Window win, int width, int height) {
	int screen = DefaultScreen(dpy);
	int screenW = DisplayWidth(dpy, screen);
	int x = (screenW - width) / 2;
	int y = 20;  // 20px from top
	XMoveWindow(dpy, win, x, y);
	XFlush(dpy);
}

void mapWindow(Display *dpy, Window win)   { XMapRaised(dpy, win); XFlush(dpy); }
void unmapWindow(Display *dpy, Window win) { XUnmapWindow(dpy, win); XFlush(dpy); }
*/
import "C"

import (
	"log"
	"sync"

	"gioui.org/app"
	"gioui.org/io/event"
)

var (
	hintsOnce  sync.Once
	storedDpy  *C.Display
	storedWin  C.Window
	windowInit bool
)

// handleViewEvent is called for every Gio event. On Linux, it catches
// X11ViewEvent to set window manager hints and store handles for
// show/hide calls.
func handleViewEvent(e event.Event, width, height int) {
	ve, ok := e.(app.X11ViewEvent)
	if !ok || !ve.Valid() {
		return
	}
	hintsOnce.Do(func() {
		storedDpy = (*C.Display)(ve.Display)
		storedWin = C.Window(ve.Window)
		windowInit = true
		C.setOverlayHints(storedDpy, storedWin)
		C.positionOverlay(storedDpy, storedWin, C.int(width), C.int(height))
		log.Println("overlay: X11 hints applied (notification type, no-focus, above, hidden)")
	})
}

// platformShow maps the overlay window (makes it visible).
func platformShow() {
	if windowInit {
		C.mapWindow(storedDpy, storedWin)
	}
}

// platformHide unmaps the overlay window (makes it invisible).
func platformHide() {
	if windowInit {
		C.unmapWindow(storedDpy, storedWin)
	}
}
