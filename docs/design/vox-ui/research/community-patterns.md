# Community Patterns: vox-ui Research
**Bead:** system_designer-d8s.4
**Date:** 2026-04-12

---

## 1. Building Global Hotkey Apps in Go

### Primary Library: golang-design/hotkey
**URL:** https://github.com/golang-design/hotkey  
**Pkg:** https://pkg.go.dev/golang.design/x/hotkey

The de-facto cross-platform Go global hotkey library. Registers system-level hotkeys that fire even when your app is not focused.

**What works:**
- Windows: solid, well-supported
- macOS: works but requires main-thread event loop — `hotkey.MainTread(func(){...})` must wrap the registration call
- Linux X11: works, but `BadAccess (X_GrabKey)` errors occur when another app has already grabbed the same key combo (issue #11, Jan 2022, still open)

**Known pitfalls:**
- AutoRepeat on X11 sends continuous KeyDown+KeyUp pairs — your hold-to-talk logic must debounce
- Issue #30 (Mar 2025): multiple events emitted on keyDown→hold→keyUp — still open
- Issue #33 (Mar 2026): "Doesn't work on Linux Mint X11" — recent regression
- **No Wayland support** — there are zero Wayland-specific issues in the tracker; the library simply uses XGrabKey under the hood

**macOS thread requirement** is the most common tripping point. Every blog post that uses this library ends up wrapping the whole thing in `hotkey.MainThread`.

### Alternative: go-vgo/robotgo
**URL:** https://github.com/go-vgo/robotgo

Heavier-weight automation library with global event listener, keyboard/mouse simulation, and screen capture. Community reports:
- Requires Accessibility + Screen Recording permissions on macOS; Terminal prompts the user on first run
- Listening to more than one hotkey simultaneously is fragile — recommended workaround: use async event listener with a switch statement instead of AddEvent for multiple bindings
- CGO-heavy; cross-compilation is painful

### Takeaway
For vox-ui: `golang.design/x/hotkey` is the right choice for macOS and X11. On Wayland, it is a dead end — a completely different approach is required.

---

## 2. Wayland Global Hotkey Limitations (Critical)

### The Security Architecture
X11 allowed any app to call XGrabKey and intercept global input. Wayland deliberately removes this — each client only receives events targeted at its own windows. Input interception across clients requires compositor cooperation.

**The two protocols:**
- `org.freedesktop.portal.GlobalShortcuts` — D-Bus portal API; compositor registers the shortcut, fires a signal to your app
- `ext-global-shortcut-v1` — newer Wayland protocol (not yet in wayland-protocols stable)

### Current Compositor Support (as of early 2026)
Source: https://github.com/tauri-apps/global-hotkey/issues/28 + https://dec05eba.com/2024/03/29/wayland-global-hotkeys-shortcut-is-mostly-useless/

| Compositor | GlobalShortcuts Portal | Notes |
|---|---|---|
| KDE Plasma | Yes (full) | Shows GUI to configure/remap keys; reports bindings back to app |
| Hyprland | Partial / broken | Ignores default keys, no GUI, doesn't report back to app — "mostly useless" |
| GNOME | No (tracking issue only) | Wontfix for now; no implementation |
| wlroots (Sway, river) | No | Expected to follow Hyprland's broken pattern |
| xdg-desktop-portal-gtk | No | GNOME's GTK portal lacks implementation |

The GlobalShortcuts portal has been spec-stable since late 2023, but only KDE has done a real implementation. As of March 2025, a bug was filed that the Activated signal doesn't pass an xdg-activation token, meaning even on KDE the activated app may not be brought to focus properly.

### Practical State for vox-ui
- **X11:** hotkey works, no issues
- **KDE Wayland:** technically possible via D-Bus portal, but requires non-trivial D-Bus plumbing, and users must configure keys in system settings — app cannot propose a default
- **GNOME Wayland:** no global hotkey API whatsoever. The only escape hatch is:
  1. Tell the user to bind a key in GNOME Settings → Keyboard → Custom shortcuts, pointing to a script that signals vox-ui via a Unix socket or DBus
  2. Use a system-level daemon approach (like Vocalinux/Voxtype do via systemd + uinput — see section 6)
- **Wayland workaround pattern used in the wild:** Voxtype and Vocalinux both handle this by either having the user set their own compositor binding or by detecting XWayland and falling back to X11 hotkey APIs

**Recommended approach for vox-ui:** Register hotkey via `golang.design/x/hotkey` on X11 and macOS. On Wayland, fall back to a clear error message with instructions for the user to configure a compositor hotkey that sends a signal to vox-ui (e.g., `kill -SIGUSR1 $(cat /tmp/vox.pid)` or a Unix socket poke).

---

## 3. macOS Accessibility Permissions for Synthetic Input

### What's Required
Source: https://jano.dev/apple/macos/swift/2025/01/08/Accessibility-Permission.html  
Reference: https://developer.apple.com/forums/thread/122492

macOS gates synthetic keyboard/mouse input behind the **Accessibility** permission (`com.apple.security.accessibility`). Apps that bypass this (e.g., use the Fn/Globe key with Apple's private API) don't need it, but anything calling `CGEventPost` with keyboard events does.

**Three distinct permissions in modern macOS:**
1. **Accessibility** — synthetic input (CGEventPost, keyboard simulation)
2. **Input Monitoring** — reading input from other apps (global hotkey listening)
3. **Automation** — controlling other apps via AppleScript/Scripting Bridge

For vox-ui: Accessibility is needed for auto-paste via keystroke injection; Input Monitoring is needed for global hotkeys.

### Check and Request Flow
```
AXIsProcessTrusted()  → bool
AXIsProcessTrustedWithOptions({kAXTrustedCheckOptionPrompt: true})  → triggers system dialog
```

App must add to entitlements:
```xml
<key>com.apple.security.accessibility</key><true/>
```
And Info.plist:
```xml
<key>NSAccessibilityUsageDescription</key>
<string>Reason for the user</string>
```

**Critical: there is no programmatic way to auto-grant this.** The user must click "Allow" in System Settings → Privacy & Security → Accessibility. The system dialog can be triggered at runtime with `kAXTrustedCheckOptionPrompt`, but it just opens the pane — the user clicks.

### Go-specific notes
- robotgo surfaces this as: "Terminal wants to control your computer" on first run from a terminal
- Pindrop (Swift/macOS) documents the graceful fallback: clipboard-only mode works without Accessibility; cursor insertion requires it
- The common pattern: try `CGEventPost`, catch the permission denied condition, fall back to clipboard+Cmd+V

---

## 4. Cross-Platform System Tray in Go

### Primary Library: fyne-io/systray
**URL:** https://github.com/fyne-io/systray  
**Pkg:** https://pkg.go.dev/fyne.io/systray

Fork of getlantern/systray that removes GTK dependency. Used by Tailscale (they maintain their own fork: `github.com/tailscale/systray`).

**What works:**
- Windows: solid
- macOS: solid, tooltip supported
- Linux X11 with desktop environment: works via DBus/SystemNotifier/AppIndicator spec

**Linux/Wayland pitfalls:**
- Uses DBus AppIndicator spec — requires a compatible system tray host (GNOME with AppIndicator extension, KDE, XFCE work; plain Sway/Hyprland without a tray do not)
- Tooltip not available on Linux
- fyne-io/fyne issues show Wayland crashes (issue #5908): `PlatformError: Wayland: Focusing a window requires user interaction`
- Building with `-tags wayland` drops X11 support entirely — no graceful fallback
- If both Fyne GUI and systray are used together, they both try to own the main thread and deadlock — workaround is `RunWithExternalLoop`

### Tailscale's fork
`github.com/tailscale/systray` — more actively maintained, same API, used in production by a large product. Worth preferring over the base fyne-io version.

### Practical GNOME Wayland status
GNOME's top bar does not support AppIndicator natively. Users need the [AppIndicator and KStatusNotifierItem Support](https://extensions.gnome.org/extension/615/appindicator-support/) GNOME extension. This is a real friction point; a significant portion of Ubuntu/Fedora Wayland users don't have it installed.

### Alternative: status bar / floating window
Several dictation tools avoid the systray problem on Linux by using a small floating status window (always-on-top, no taskbar entry) or by doing audio-only feedback (beep on start/stop). This sidesteps the systray compatibility maze entirely.

---

## 5. Speech-to-Text UX Patterns: Recording State Indication

### Community consensus from HN / existing tools

**HN discussion: "Ghost Pepper – Local hold-to-talk STT for macOS"**
**URL:** https://news.ycombinator.com/item?id=47666024

Key community takeaways:
- The HN thread turned into "a support group for people who have each independently built the same macOS STT app" — there is a proliferation of nearly identical tools
- **Live transcription streaming** during recording is the most-requested missing feature across all tools; users report it helps them structure their speech
- **Auto-paste without extra keystrokes** is the second most-requested feature
- Hold-to-talk is preferred over toggle by most power users ("feels more natural, less error-prone")

### Pattern catalogue from real tools

**Menu bar icon state change (most common macOS pattern):**
- Pindrop: menu bar icon turns red during recording — no floating window, no dock icon
- WhisperType: microphone icon in menu bar, state-driven color change
- open-wispr: waveform icon in menu bar

**Floating overlay (macOS pattern, with Wayland caveats):**
Source: https://yuta-san.medium.com/building-sttinput-universal-voice-to-text-for-macos-080ca40cb9de
- STTInput uses a floating overlay at `.statusBar` window level (above all windows, below system UI)
- Three states: idle (hidden or "⌘×3 to Record" hint, auto-hides in 5s) → recording (red gradient + shimmer) → transcribing (blue gradient + rotating icon)
- On Linux: Vocalinux disables the recording overlay by default because compositors treat it as the active window, stealing focus and breaking the paste-back

**Audio feedback:**
- Vocalinux uses subtle start/stop sounds as primary state indicator (avoids focus-stealing overlay problem on Linux)
- Ghost Pepper does not mention audio cues; visual only

**Notification area (Linux):**
- Voxtype: status indicator in menu bar showing idle/recording/transcribing (via systray)
- Vocalinux: systray icon with state

**Key insight for vox-ui:** The floating overlay approach is the richest UX but is actively broken on some Linux compositors. The safest cross-platform pattern is menu bar/tray icon color change + optional audio beep. The overlay should be optional and off by default on Linux.

---

## 6. Auto-Paste Patterns: Getting Text Into Arbitrary Apps

### The Wayland auto-paste crisis (critical reading)
**URL:** https://github.com/OpenWhispr/openwhispr/issues/240

OpenWhispr's issue #240 documents the full failure cascade on GNOME/Wayland. Four compounding bugs:

1. **Silent xdotool failure:** xdotool exits 0 but silently does nothing for native Wayland windows (non-XWayland). Clipboard paste appears to succeed, text never appears.
2. **Socket permissions:** ydotoold started as root creates `/tmp/.ydotool_socket` with root-only perms; user-level app can't connect.
3. **Clipboard race condition:** 200ms restore delay is too short on Wayland — text flashes then disappears as the clipboard is restored before the target app has finished processing the paste.
4. **Environment variable inheritance:** `YDOTOOL_SOCKET` set in a terminal doesn't propagate to apps launched from app menus.

**Proposed resolution:** Use ydotool `type` (direct character injection via uinput) rather than clipboard+Ctrl+V; skip clipboard restore when using direct typing; start ydotoold as the user, not root.

### Tool landscape for Linux text injection

| Tool | X11 | Wayland | Root needed | Notes |
|---|---|---|---|---|
| xdotool | Yes | No (silently fails on native Wayland windows) | No | Works for XWayland apps only |
| ydotool | Yes | Yes | No (needs `/dev/uinput` perms or user-run daemon) | uinput-based; requires ydotoold daemon |
| dotool | Yes | Yes | No | Newer alternative; works without daemon |
| IBus API | X11+Wayland | Yes | No | Input method bus; Vocalinux approach; requires IBus running |
| zwp_virtual_keyboard | Wayland only | Yes | No | Protocol-level; only on wlroots compositors (Sway/Hyprland) |

### wayland-virtual-input-go (Go library, June 2025)
**URL:** https://github.com/bnema/wayland-virtual-input-go  
**Pkg:** https://pkg.go.dev/github.com/bnema/wayland-virtual-input-go/virtual_keyboard

Published June 2025. Implements `zwp_virtual_keyboard_v1` and `zwlr_virtual_pointer_v1` in pure Go.
- Works on Sway and Hyprland (wlroots-based)
- Does NOT work on GNOME Wayland (protocol not supported)
- No root required — works via Wayland socket
- API: `keyboard.TypeString("Hello World!")`

### bendahl/uinput (Go, Linux only)
**URL:** https://github.com/bendahl/uinput  
**Pkg:** https://pkg.go.dev/github.com/bendahl/uinput

Pure Go wrapper for Linux `/dev/uinput`. Creates a virtual input device.
- Requires write access to `/dev/uinput` (udev rule needed: `KERNEL=="uinput", GROUP="$USER", MODE:="0660"`)
- v1.6.1 attempted Wayland fix; v1.6.2 reverted due to regression — compatibility still uncertain
- More reliable: use ydotool (which also uses uinput) and shell out to it

### macOS auto-paste patterns
Source: Pindrop README, STTInput blog post

The standard pattern used by every macOS Whisper dictation tool:
1. Write text to clipboard (`NSPasteboard`)
2. Simulate Cmd+V via `CGEventPost` (requires Accessibility permission)
3. Restore original clipboard contents (with a delay — race condition risk!)

Pindrop's graceful degradation: "Clipboard output still works without Accessibility permission. Direct cursor insertion is an optional enhancement."

STTInput uses the same two-tier approach: clipboard+Cmd+V as primary; virtual keypress injection as fallback.

**Race condition warning:** Clipboard restore after paste is universally tricky. The correct approach is to delay restoration until after the paste has been processed, or skip restoration entirely (just leave the text in clipboard).

### Recommended vox-ui auto-paste strategy

**macOS:**
- Primary: set clipboard + simulate Cmd+V via CGO/CGEvent
- Fallback: clipboard-only (user pastes manually)
- Permission check at startup; graceful degradation without Accessibility

**Linux X11:**
- Primary: `xdotool type --clearmodifiers` or clipboard + `xdotool key ctrl+v`
- Secondary: ydotool/dotool for non-XWayland windows

**Linux Wayland (wlroots: Sway, Hyprland):**
- Primary: `wayland-virtual-input-go` (`zwp_virtual_keyboard_v1`) — no root, no daemon
- Fallback: ydotool type (requires daemon setup, udev rule)

**Linux Wayland (GNOME):**
- Only reliable option: clipboard + xdg-desktop-portal input capture (complex, rarely implemented)
- Practical fallback: clipboard-only — set clipboard, notify user to Ctrl+V
- IBus approach (Vocalinux): works if IBus is running, which it is by default on GNOME

---

## Summary of Key Findings

1. **Global hotkeys on Linux Wayland are fundamentally broken for most setups.** KDE is the one exception. For GNOME and wlroots, instruct users to configure a compositor-level hotkey that signals vox-ui via a simple IPC mechanism (Unix socket, SIGUSR1, etc.).

2. **golang.design/x/hotkey is the right Go library** for macOS and X11. Known quirks: main-thread requirement on macOS, key conflict errors on X11, hold-key event repetition bug.

3. **macOS requires explicit user action** to grant Accessibility permissions — no workaround. Design for graceful clipboard-only fallback when the permission is absent.

4. **fyne-io/systray works** on most desktop environments but requires AppIndicator support on GNOME. Consider the Tailscale fork. Audio feedback is a robust cross-platform alternative to a tray icon for recording state.

5. **The floating overlay approach** is richest UX on macOS but causes focus-stealing bugs on Linux compositors. Make it opt-in on Linux, default off.

6. **Auto-paste on Wayland is a multi-layer problem.** xdotool silently fails for native Wayland windows. ydotool/dotool via uinput works everywhere but requires daemon setup. `wayland-virtual-input-go` is the clean Go-native path for wlroots compositors (Sway/Hyprland) as of June 2025. GNOME Wayland paste is only reliable via IBus or clipboard-only.

---

## Source URLs

- https://github.com/golang-design/hotkey
- https://pkg.go.dev/golang.design/x/hotkey
- https://dec05eba.com/2024/03/29/wayland-global-hotkeys-shortcut-is-mostly-useless/
- https://github.com/tauri-apps/global-hotkey/issues/28
- https://github.com/albertlauncher/albert/issues/309
- https://github.com/warpdotdev/Warp/issues/4800
- https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.GlobalShortcuts.html
- https://jano.dev/apple/macos/swift/2025/01/08/Accessibility-Permission.html
- https://developer.apple.com/forums/thread/122492
- https://github.com/fyne-io/systray
- https://pkg.go.dev/github.com/tailscale/systray
- https://news.ycombinator.com/item?id=47666024
- https://yuta-san.medium.com/building-sttinput-universal-voice-to-text-for-macos-080ca40cb9de
- https://voxtype.io/
- https://vocalinux.com/
- https://github.com/jatinkrmalik/vocalinux
- https://github.com/OpenWhispr/openwhispr/issues/240
- https://github.com/ReimuNotMoe/ydotool
- https://github.com/bnema/wayland-virtual-input-go
- https://pkg.go.dev/github.com/bnema/wayland-virtual-input-go/virtual_keyboard
- https://github.com/bendahl/uinput
- https://pkg.go.dev/github.com/bendahl/uinput
- https://github.com/watzon/pindrop
- https://github.com/go-vgo/robotgo
