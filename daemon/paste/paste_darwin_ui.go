//go:build ui && darwin

package paste

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include <ApplicationServices/ApplicationServices.h>
#include <stdbool.h>

// axTrusted returns true if the process has Accessibility permissions.
bool axTrusted() {
	return AXIsProcessTrusted();
}

// postPasteKeystroke simulates Cmd+V via CGEventPost.
void postPasteKeystroke() {
	CGEventSourceRef src = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
	CGEventRef down = CGEventCreateKeyboardEvent(src, 9, true);  // 9 = 'v'
	CGEventRef up   = CGEventCreateKeyboardEvent(src, 9, false);
	CGEventSetFlags(down, kCGEventFlagMaskCommand);
	CGEventSetFlags(up,   kCGEventFlagMaskCommand);
	CGEventPost(kCGHIDEventTap, down);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	CFRelease(src);
}
*/
import "C"

import (
	"fmt"
	"log"
)

func init() {
	// Register CGEventPost as the preferred paste injection on macOS
	// when built with the ui tag. Falls back to osascript if
	// Accessibility is not granted.
	injectHook = cgEventPaste
}

func cgEventPaste() error {
	if !C.axTrusted() {
		return fmt.Errorf("accessibility not granted")
	}
	C.postPasteKeystroke()
	return nil
}

// CheckAccessibility logs a one-time status message about Accessibility
// permissions. Call at daemon startup on macOS.
func CheckAccessibility() {
	if !C.axTrusted() {
		log.Println("paste: Accessibility not granted — paste injection will use osascript fallback")
		log.Println("paste: grant access in System Settings → Privacy & Security → Accessibility")
	} else {
		log.Println("paste: Accessibility granted — using CGEventPost for paste injection")
	}
}
