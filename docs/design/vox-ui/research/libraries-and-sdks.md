# Libraries & SDKs — vox-ui

> Research bead: `system_designer-d8s.3`
> Date: 2026-04-12
> Topic: Go libraries and platform APIs for a global hotkey dictation overlay

---

## 1. Global Hotkey Registration

### 1.1 golang-design/hotkey

- **Repo**: https://github.com/golang-design/hotkey
- **Package**: `golang.design/x/hotkey`
- **Stars**: ~258
- **Latest release**: v0.4.1 — December 28, 2022
- **Maintenance**: Effectively dormant. 24 total commits. 8 open issues, some from 2026, no activity from maintainers.
- **License**: MIT

**What it does**: Cross-platform global hotkey registration (macOS, Linux/X11, Windows). Channel-based API — `hk.Keydown()` and `hk.Keyup()` return `<-chan Event`. No callback hell.

**API sketch**:
```go
hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS)
if err := hk.Register(); err != nil { ... }
<-hk.Keydown()  // blocks until pressed
```

**Linux (X11)**:
- Uses `XGrabKey` under the hood.
- Known pitfall: AutoRepeat enabled in X server causes continuous `Keyup` floods.
- Key mapping complexity: `Ctrl+Alt+S` may need to be registered as `Ctrl+Mod2+Mod4+S` depending on keyboard layout.
- Issue #11: `BadAccess (X_GrabKey)` if another application has already grabbed the same combo.
- **Wayland**: Not supported at all. No mention in README or issues of any Wayland path.

**macOS**:
- Requires hotkey events on the main thread. For standalone apps, must call `mainthread.Init(fn)`. For Fyne/Gio/Ebiten, this is handled automatically by the framework.
- Uses Carbon `RegisterEventHotKey` internally. Carbon is deprecated since macOS 10.8 but still works. As of macOS Sequoia (15), hotkeys with only Option or Option+Shift modifiers stopped working — Apple intentionally restricted this to limit keylogging.
- No Accessibility permissions required (unlike CGEventTap).

**Fyne integration**: The repo ships `examples/fyne/main.go` — spin up hotkey listener in a goroutine, no `mainthread` wrapper needed when Fyne owns the run loop.

**Verdict**: Best choice for X11 + macOS if you can tolerate the dormant maintenance. For push-to-talk (hold key), the AutoRepeat issue is a real problem on Linux — requires `xset -r <keycode>` workaround or careful debouncing.

---

### 1.2 go-vgo/robotgo (global hooks via gohook)

- **Repo**: https://github.com/go-vgo/robotgo
- **Stars**: ~10,700
- **Maintenance**: Active (used by 2024–2025 projects, recent CI updates)
- **License**: Apache-2.0
- **CGO**: Required. Heavy build deps.

**What it does**: Full RPA/automation library — mouse, keyboard, screen capture, image matching. Includes global event hooks via `robotn/gohook` (a separate sub-library wrapping `libuiohook`).

**gohook specifics**:
- **Repo**: https://github.com/robotn/gohook
- **Stars**: ~410
- **Last commit**: December 4, 2025 (v0.42.3)
- Provides `hook.Start()` event channel — global keyboard AND mouse events.
- Uses `libuiohook` (C), so requires CGO.

**Linux**: X11 only in open-source version. Wayland support is explicitly listed as a **RobotGo Pro** (commercial) feature. Compiling on Wayland-native distros (Fedora, Ubuntu 24.04+ without XWayland) fails unless `libx11-dev` and `libxtst-dev` are present and XWayland is running.

**macOS**: Requires Accessibility permission (Screen Recording + Accessibility in System Settings). CGEventTap-style under the hood.

**Verdict**: Overkill for just hotkeys. Use `golang-design/hotkey` instead for that. gohook is useful if you also need mouse event monitoring (e.g., detecting focus changes).

---

### 1.3 Linux X11: XGrabKey Direct (CGO)

If neither library fits, `XGrabKey` via CGO is straightforward:

```c
XGrabKey(display, keycode, modifier, root_window, False, GrabModeAsync, GrabModeAsync);
```

The `github.com/linuxdeepin/go-x11-client` package (published April 2024) provides pure-Go X11 protocol bindings without CGO, based on XCB. Usable for `XGrabKey` equivalents.

**Pitfall**: `BadAccess` error if another app has grabbed the key combo. Must handle gracefully.

---

### 1.4 Linux Wayland: The Global Hotkey Graveyard

**Bottom line**: Wayland global hotkeys are a mess in 2026. Here is the landscape:

**XDG Desktop Portal — GlobalShortcuts**:
- Spec exists (xdg-desktop-portal v1.16+).
- **GNOME**: Does not support the GlobalShortcuts portal at all as of early 2026.
- **KDE Plasma**: Works well — app registers preferred keys, user sees a GUI, settings persist.
- **wlr-based compositors (Sway, Hyprland via xdg-desktop-portal-wlr)**: The xdg-desktop-portal-wlr project (issue #240) explicitly blocked on a missing Wayland protocol. **Unimplemented, no ETA.**
- **Hyprland via xdg-desktop-portal-hyprland**: Partially works. `BindShortcuts` and `ListShortcuts` methods are broken; workaround is passing shortcuts during `CreateSession`. Closed as "completed" with the workaround as intended behavior.

No production-ready Go client library exists for the GlobalShortcuts portal as of this writing.

**`/dev/uinput` as input listener** (read side): Not applicable — uinput is for injecting events, not monitoring them.

**XWayland bridge**: Run as an X11 app under XWayland. `XGrabKey` works for XWayland clients. However, hotkeys won't fire when focus is on a native Wayland window — making it unreliable.

**Practical recommendation**: Ship X11 support and document Wayland as "use KDE Plasma, or run with XWayland." Wayland-native global hotkeys are not solvable reliably in 2026.

---

### 1.5 macOS: Carbon vs. CGEventTap vs. NSEvent

| API | Requires Accessibility | Works in Sandbox | Deprecated | Notes |
|-----|----------------------|-----------------|------------|-------|
| `RegisterEventHotKey` (Carbon) | No | Yes | Since 10.8 | Still works; golang-design/hotkey uses this. Broken for Option/Shift-only in macOS 15. |
| `CGEventTap` | Yes (Accessibility) | Yes (macOS 10.15+) | No | Most powerful; catches all keys including Option. Requires user to grant permission. |
| `NSEvent.addGlobalMonitor` | Yes (Input Monitoring) | App Store compatible | No | Swift/ObjC; can be called from CGO. |

For a non-sandboxed dictation app, `RegisterEventHotKey` (via `golang-design/hotkey`) is the path of least resistance. If Opt+key combos are needed, CGEventTap is the alternative — but requires user to approve Accessibility access on first run.

---

## 2. System Tray / Status Bar

### 2.1 fyne-io/systray (Recommended)

- **Repo**: https://github.com/fyne-io/systray
- **Package**: `fyne.io/systray`
- **Stars**: ~338
- **Used by**: 7,400+ projects
- **Last commit**: Active; latest published December 23, 2025
- **License**: BSD-3

**What it does**: Drop-in cross-platform system tray icon + menu. Fork of `getlantern/systray` with GTK dependency removed.

**Linux**: Uses DBus via SystemNotifier/AppIndicator spec. Requires `libayatana-appindicator3` or `libappindicator3` dev package. Old desktop environments may need `snixembed` proxy for compatibility.

**macOS**: Requires app bundle (`Info.plist`). `NSStatusItemBehavior` can be set via `systray.SetRemovalAllowed(true)` for icon draggability.

**Windows**: Compile with `-ldflags "-H=windowsgui"` to suppress console.

**CGO**: Required. `CGO_ENABLED=1`.

**API**: `systray.Run(onReady, onExit)` blocks the main thread. For use with another framework's main loop: `systray.RunWithExternalLoop()` returns start/end functions.

**Windowless operation**: Yes — `systray.Run()` is the entire app lifecycle. No window required.

---

### 2.2 getlantern/systray (Original)

- **Repo**: https://github.com/getlantern/systray
- **Stars**: ~3,700
- **Last commit**: May 2, 2023 — appears stale
- **Linux**: Requires libappindicator3 (GTK-based). Heavier dependency.

**Verdict**: Use `fyne-io/systray` instead. Same API, fewer deps, actively maintained.

---

### 2.3 energye/systray

- **Repo**: https://github.com/energye/systray
- **Stars**: ~166
- **Latest release**: v1.0.3, January 31, 2026
- **Maintenance**: Active
- **Origin**: Fork of `getlantern/systray`, GTK removed like fyne-io fork.
- **Language**: 89.7% Go, 9.0% Objective-C

Essentially equivalent to `fyne-io/systray` in features. Fewer dependents. No clear advantage unless you're in the energye ecosystem.

---

### 2.4 Fyne Framework System Tray (Built-in)

Since Fyne v2.2 (July 2022), Fyne has built-in system tray support via `app.DesktopApp` interface:

```go
if desk, ok := app.Current().(desktop.App); ok {
    desk.SetSystemTrayMenu(fyne.NewMenu("Vox", ...))
}
```

Fyne v2.7 adds `SetSystemTrayWindow()` — left-click on tray shows/hides a window.

Works on macOS, Windows, and "most Linux" (requires compositor with AppIndicator support).

**Caveat**: Global hotkey shortcuts registered with `canvas.AddShortcut()` only fire when the Fyne window is focused. For global hotkeys, must use `golang-design/hotkey` in a goroutine alongside Fyne. This works cleanly — the repo ships an example.

---

## 3. Cursor / Pointer Modification

### 3.1 Linux X11

**Within your own window**: Use `XDefineCursor(display, window, cursor)` with a cursor created via `XCreateGlyphCursor` or `XcursorLibraryLoadCursor`. Trivially doable.

**System-wide cursor replacement**: Not possible through standard APIs. `XFixesChangeCursorByName` exists in the XFIXES extension but only affects cursors for named resource; it does not change the cursor the user sees globally. Cursor themes are set via `XCURSOR_THEME` env var, which requires app restart.

**Cursor animation**: In X11, `XcursorAnimatedCreate` or multi-image `Xcursor` files support animated cursors within a single application's window. **Animating the system-wide cursor (i.e., changing the cursor the user sees over other apps) is not possible** through standard X11 APIs.

**Conclusion for vox-ui**: Can change cursor appearance within your own overlay window (if you show one) but cannot change the system cursor while the user is typing in another app's window.

---

### 3.2 Linux Wayland

Wayland apps set cursor shape by requesting the compositor draw it. The `cursor-shape-v1` protocol (wp_cursor_shape_manager_v1) was merged in 2024:
- GNOME/Mutter supports it as of 2024.
- SDL added support March 2024.
- GLFW implemented it.

**However**: This protocol only controls the cursor **for the client's own surfaces**. An app cannot change the cursor when the pointer is over a different app's window. The compositor is the sole arbiter of cursor appearance outside your surfaces.

**Cursor animation for audio levels**: Not feasible via cursor on Wayland. The cursor can be set to a static named shape (e.g., `wp_cursor_shape_device_v1_shape.crosshair`) but animated/custom cursor images require using `wl_pointer.set_cursor` with a `wl_buffer` — doable within your own surfaces only.

---

### 3.3 macOS

**Within your app**: `NSCursor` API is clean and well-documented. Create from `NSImage`, set with `cursor.set()`. Custom animated cursors require private APIs (see Mousecape) — not suitable for production apps.

**System-wide**: `NSCursor.current` is scoped to your process. `NSCursor.currentSystem` is read-only. **There is no public API to change the system cursor globally on macOS.** macOS security policies prevent this.

**macOS 14 regression**: Custom cursors are periodically reset to arrow when the pointer moves over certain view regions. Known Apple bug; no fix as of early 2025.

**Conclusion**: Cursor-based audio level indication is not viable as a cross-application overlay technique on any major platform. The overlay window approach (a small floating transparent window near the cursor or in a fixed screen position) is the correct design.

---

## 4. Synthetic Keyboard Input / Auto-Paste

The recommended pattern for speech-to-text auto-paste is:
1. Write transcribed text to clipboard.
2. Simulate Ctrl+V (or Cmd+V on macOS) to paste into the focused window.

### 4.1 Linux X11: xdotool

- **Binary**: `xdotool type --clearmodifiers "text"` or `xdotool key ctrl+v`
- **X11 only**. Returns exit code 0 even on Wayland (silent failure).
- Shell out via `exec.Command("xdotool", ...)`.
- Mature, widely available on X11 distros.

### 4.2 Linux Wayland: ydotool

- **Repo**: https://github.com/ReimuNotMoe/ydotool
- Uses Linux `uinput` kernel interface — works on both X11 and Wayland.
- **Requires**: Running `ydotoold` daemon. User must configure udev rules or run daemon as root.
- **Socket permissions**: When run with sudo, `/tmp/.ydotool_socket` has root-only permissions — user-level clients can't reach it. Must set `YDOTOOL_SOCKET` or configure udev.
- OpenWhispr (v1.4.9) solved this by implementing `/dev/uinput` injection directly in their binary.

### 4.3 Linux Wayland: wtype

- **Repo**: https://github.com/atx/wtype
- Uses `zwp_virtual_keyboard_v1` Wayland protocol.
- Types arbitrary text into the focused window. No daemon required.
- **Limitation**: Only works on compositors supporting `virtual-keyboard-unstable-v1` (wlroots: Sway, Hyprland; **not** GNOME as of early 2026).
- CLI tool; shell out from Go.

### 4.4 Go: bendahl/uinput (ARCHIVED)

- **Repo**: https://github.com/bendahl/uinput
- **Stars**: ~109
- **Status**: **Archived March 19, 2024.** Read-only. No further updates.
- **What it did**: Pure Go wrapper for Linux `/dev/uinput`. Wayland fix was released in v1.6.1 (April 2023).
- Can still be used (it works), but no maintenance path.
- Requires udev rule: `KERNEL=="uinput", GROUP="$USER", MODE:="0660"`

### 4.5 Go: bnema/wayland-virtual-input-go

- **Package**: `github.com/bnema/wayland-virtual-input-go/virtual_keyboard`
- **Version**: v0.2.0 (June 16, 2025)
- **Protocol**: `virtual-keyboard-unstable-v1` via `neurlang/wayland`
- **Compositors**: wlroots-based (Sway, Hyprland). Not GNOME.
- Pre-v1 stability but functional.

**API**:
```go
manager, _ := virtual_keyboard.NewVirtualKeyboardManager(ctx)
keyboard, _ := manager.CreateKeyboard()
keyboard.TypeString("Hello World!")
keyboard.TypeKey(KEY_ENTER)
```

Good choice for pure-Go Wayland input injection if wlroots compositors are the target.

### 4.6 Go: micmonay/keybd_event (ARCHIVED)

- **Repo**: https://github.com/micmonay/keybd_event
- **Stars**: ~414
- **Status**: **Archived November 20, 2025.**
- Linux backend uses `uinput` (requires root or udev rules). 2-second startup delay required on Linux.
- macOS: depends on Apple frameworks; no cross-compilation.
- Last release: October 20, 2023. Use only if uinput approach fits; prefer `bendahl/uinput` or `bnema/wayland-virtual-input-go`.

### 4.7 macOS: CGEventPost

```c
CGEventRef event = CGEventCreateKeyboardEvent(NULL, kVK_ANSI_V, true);
CGEventSetFlags(event, kCGEventFlagMaskCommand);
CGEventPost(kCGAnnotatedSessionEventTap, event);
```

- Requires Accessibility permission (or Input Monitoring for `listenOnly` tap).
- From Go, call via CGO or shell out to an AppleScript/osascript helper.
- macOS Sequoia (15) tightened CGEventPost restrictions — sandboxed apps cannot post events.
- **Alternative**: Write to pasteboard + simulate Cmd+V via CGEventPost. This is the reliable pattern.

### 4.8 golang-design/clipboard (Recommended Clipboard Library)

- **Repo**: https://github.com/golang-design/clipboard
- **Stars**: ~770
- **Latest**: v0.7.1, June 14, 2025
- **Maintenance**: Active
- macOS: CGO required. Linux: requires `libx11-dev`. **Wayland: unsupported** — requires XWayland bridge with `DISPLAY` set.
- Supports text and PNG image. Provides change notification channel.

**For Wayland clipboard**: Use `wl-copy` (from `wl-clipboard` package) via `exec.Command`. The `tiagomelo/go-clipboard` package abstracts this but is a thin shell wrapper.

### 4.9 atotto/clipboard (Simpler Alternative)

- **Repo**: https://github.com/atotto/clipboard
- **Stars**: ~1,400
- Requires `xclip` or `xsel` on Linux (CLI tools). No Wayland (`wl-clipboard`) support mentioned.
- Simpler API but less maintained than `golang-design/clipboard`.

### 4.10 Recommended Auto-Paste Strategy

```
X11:      golang-design/clipboard (write) + xdotool key ctrl+v (paste)
Wayland:  wl-copy (write) + ydotool/wtype key ctrl+v (paste, compositor-dependent)
macOS:    golang-design/clipboard (write) + CGEventPost Cmd+V
```

The OpenWhispr project's v1.4.9 solution (direct `/dev/uinput` injection) is the most reliable for Wayland Linux but requires udev setup from the user.

---

## 5. Cross-Platform GUI Frameworks

### 5.1 Fyne (v2.7.x)

- **Repo**: https://github.com/fyne-io/fyne
- **Stars**: ~28,000
- **Latest stable**: v2.7.3, February 21, 2025
- **Maintenance**: Excellent. Releases every 1–3 months. ~100 contributors.
- **License**: BSD-3 (free for non-commercial; commercial license available)
- **CGO**: Required (OpenGL/OpenGLES backend)

**Relevant capabilities**:
- System tray: Built-in since v2.2. `SetSystemTrayWindow()` since v2.7.
- Background mode: `Window.Hide()` + intercept close — app stays alive in tray.
- Global hotkeys: Natively only within focused window. External: `golang-design/hotkey` in goroutine (official example exists).
- Overlay window: Can create a window with no decorations, `SetFixedSize`, positioned near cursor.

**Platform support**: macOS, Linux (X11+Wayland via EGL), Windows, Android, iOS, WebAssembly.

**Gotchas**:
- Custom/branded look requires fighting Fyne's Material Design defaults.
- Performance on animations is adequate but not buttery.
- Single-maintainer community concern (historically) — now has broader team.

**Assessment for vox-ui**: Good fit if you want a proper GUI toolkit. System tray + overlay window + integration with golang-design/hotkey covers all vox-ui requirements. The "no visible window at startup" pattern is documented and supported.

---

### 5.2 Wails (v3 alpha)

- **Website**: https://wails.io / https://v3alpha.wails.io
- **Stars**: ~28,000+ (v2)
- **Status**: v3 in alpha as of 2025
- **License**: MIT

**What it is**: Build desktop apps with Go backend + web frontend (HTML/CSS/JS). Uses native WebView (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux).

**Relevant v3 capabilities**:
- System tray: Native support. `systray-basic` and `systray-menu` examples in v3.
- Background mode: `ActivationPolicyAccessory` for macOS (no Dock icon, no main window).
- Hidden window on startup: `WebviewWindowOptions{Hidden: true}` — known regression in Windows alpha 22 (issue #4498).
- Global hotkeys: Not built-in; would need external library.

**Assessment for vox-ui**: Attractive if you want a web-rendered overlay UI (audio levels visualized with JS canvas, etc.). However: v3 is alpha — production risk. WebView dependency adds weight. No built-in global hotkey support. **Not recommended** for a minimal dictation overlay; better suited to richer UIs.

---

### 5.3 Gio (gioui.org)

- **Repo**: https://github.com/gioui/gio (mirror of sr.ht)
- **Stars**: ~4,200
- **Maintenance**: Active — monthly newsletters, consistent development
- **License**: MIT + Unlicense
- **No CGO** — pure Go (uses platform-specific drawing APIs via purego/syscall)

**Relevant capabilities**:
- Immediate-mode rendering — full control over every pixel.
- Overlay windows: `app.Window` with no decorations, `pointer.PassOp` for click-through areas.
- System tray: **Not built-in.** Must combine with `fyne-io/systray` or similar.
- Global hotkeys: No built-in support. Use `golang-design/hotkey`.
- Platforms: Linux (X11+Wayland via Wayland protocol directly), macOS, Windows, Android, iOS, WASM.

**Wayland**: Gio communicates directly with the Wayland compositor — no XWayland needed for the window itself. This makes it the best choice for a native Wayland overlay window.

**Assessment for vox-ui**: Best for a pixel-perfect overlay widget (audio level bars, animated indicator). Composing Gio window + `fyne-io/systray` + `golang-design/hotkey` covers vox-ui needs, but requires wiring three separate systems. More work, more control.

---

### 5.4 Framework Comparison Summary

| Feature | Fyne v2.7 | Wails v3 | Gio |
|---------|-----------|----------|-----|
| System tray (built-in) | Yes | Yes | No (need fyne-io/systray) |
| No window at startup | Yes | Yes (alpha bugs) | Yes |
| Global hotkeys | External lib | External lib | External lib |
| Wayland native window | Via EGL | Via WebKitGTK | Native |
| CGO required | Yes | Yes (WebView) | No |
| Overlay/no-decor window | Yes | Possible | Yes (full control) |
| Stability | Production | Alpha | Production |
| Custom rendering | Limited | Full (web) | Full (immediate mode) |
| Stars | 28k | 28k | 4.2k |

**Recommendation**: Fyne is the practical choice for fastest time-to-working-app. Gio is the right choice if Wayland-native rendering and pixel-level control of the overlay are priorities.

---

## 6. Audio Level Monitoring

### 6.1 gen2brain/malgo

- **Repo**: https://github.com/gen2brain/malgo
- **Stars**: ~398
- **Latest**: Active; last release September 2025 according to pkg.go.dev
- **License**: MIT
- **CGO**: Required (wraps miniaudio C library)
- **External libs**: None on macOS/Windows. Linux needs `-ldl` only.
- **Platform audio backends**: WASAPI/DirectSound/WinMM (Windows), CoreAudio (macOS/iOS), PulseAudio/ALSA/JACK (Linux)

**Microphone capture API**: Callback-based. Raw PCM frames arrive as `[]byte`. 16-bit signed samples at 44100 Hz (configurable).

**RMS calculation**: Straightforward — convert bytes to int16 pairs, square, average, sqrt. No built-in RMS function but trivial to implement in ~10 lines of Go.

**Assessment**: Best choice for audio level monitoring in vox-ui. No system library installation required on macOS and Windows. Linux requires nothing beyond `-ldl`. Actively maintained. The capture example in `_examples/capture/capture.go` shows the full pattern.

---

### 6.2 gordonklaus/portaudio

- **Repo**: https://github.com/gordonklaus/portaudio
- **Stars**: ~834
- **Status**: Dormant (47 total commits, no releases published)
- **External dependency**: Requires `portaudio19-dev` installed on the system (or source build)
- **CGO**: Required

**Assessment**: Avoid. Requires system installation. Dormant. `malgo` is strictly better.

---

### 6.3 MarkKremer/microphone (PortAudio + beep)

- **Repo**: https://github.com/MarkKremer/microphone
- **Latest**: March 10, 2025
- Wraps gordonklaus/portaudio into a `beep.StreamCloser`. Inherits all PortAudio system dependency baggage.

**Assessment**: Same problems as portaudio. Not recommended.

---

### 6.4 Audio Level Architecture Recommendation

```go
// malgo capture loop with RMS computation
malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {})

config := malgo.DefaultDeviceConfig(malgo.Capture)
config.Capture.Format = malgo.FormatS16
config.Capture.Channels = 1
config.SampleRate = 16000

callbacks := malgo.DeviceCallbacks{
    Data: func(_, input []byte, frameCount uint32) {
        rms := computeRMS(input) // convert []byte -> []int16 -> RMS
        levelCh <- rms           // send to overlay goroutine
    },
}
device, _ := malgo.InitDevice(ctx.Context, config, callbacks)
device.Start()
```

For the overlay, poll `levelCh` at ~30 FPS and update the visual. No terminal involvement.

---

## Key Findings Summary

1. **Global hotkeys**: `golang-design/hotkey` (X11 + macOS). **Wayland: no good solution** — tell users to use KDE or XWayland.
2. **System tray**: `fyne-io/systray` — most maintained fork, GTK-free, DBus-based on Linux.
3. **Cursor modification**: Not viable as a cross-app overlay mechanism on any platform. Use a floating overlay window instead.
4. **Auto-paste Linux X11**: `golang-design/clipboard` + xdotool. Wayland: `wl-copy` + ydotool or `bnema/wayland-virtual-input-go` (wlroots only, not GNOME).
5. **Auto-paste macOS**: `golang-design/clipboard` + CGEventPost Cmd+V.
6. **GUI framework**: Fyne (production-ready, systray built-in, hotkey goroutine integration documented). Gio if Wayland-native rendering is critical.
7. **Audio levels**: `gen2brain/malgo` — CGO, no system deps on macOS/Windows, ALSA/PulseAudio on Linux. Compute RMS from PCM callback.

---

## Critical Dependency Tree

```
vox-ui binary
├── Global hotkey:       golang.design/x/hotkey  (X11 + macOS)
├── System tray:         fyne.io/systray  (CGO, DBus on Linux)
├── Overlay window:      fyne-io/fyne OR gioui.org  (choose one)
├── Clipboard write:     golang.design/x/clipboard  (CGO, X11+macOS; Wayland: shell out to wl-copy)
├── Keyboard inject:     xdotool (X11) / ydotool or wtype (Wayland) / CGEventPost (macOS)
└── Audio capture:       github.com/gen2brain/malgo  (CGO, no system deps on macOS/Windows)
```

**CGO is unavoidable** for this feature set on both platforms. Plan your build toolchain accordingly.
