//go:build ui && !linux

package overlay

import "gioui.org/io/event"

func handleViewEvent(_ event.Event, _, _ int) {}
func platformShow()                           {}
func platformHide()                           {}
