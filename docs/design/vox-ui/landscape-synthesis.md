# Vox UI — Landscape Synthesis

## Executive Summary

The dictation overlay space is crowded but fragmented. There are ~15 active projects across Python, Rust, Swift, Electron, and Go — but no single tool delivers a lightweight, cross-platform (macOS + Linux) dictation overlay in Go. The closest analog is **Turbo Whisper** (Python, 27 stars), which has the right UX pattern (hotkey → waveform overlay → auto-paste) but is a Python app with limited adoption. **OpenWhispr** (Electron, 2.4k stars) is the most feature-complete cross-platform option but carries Electron's weight. **HyprVoice** (Go, 196 stars) proves Go can do dictation on Linux/Wayland but is Hyprland-only with no overlay UI.

The critical finding is that **Wayland global hotkeys are broken on most compositors** — only KDE has a working implementation. GNOME has no portal support and no ETA. The practical pattern used by successful tools (Voxtype, Vocalinux, waystt) is to have the user configure a compositor-level hotkey that signals the app via IPC. This is unavoidable.

The second critical finding is that **system cursor modification is not viable** for cross-application feedback on any platform (X11, Wayland, or macOS). All cursors are process-scoped. Every successful dictation tool uses either a floating overlay window, a system tray icon state change, or audio cues.

## Comparison Matrix

| Tool | Lang | Platform | Hotkey | Overlay | Auto-paste | Stars | Status |
|---|---|---|---|---|---|---|---|
| OpenWhispr | JS/TS (Electron) | Mac/Win/Linux | Yes | Glassmorphic panel | Yes (fallback chain) | 2,400 | Active |
| Voxtype | Rust | Linux only | evdev | None | wtype→dotool→ydotool | 616 | Active |
| OpenSuperWhisper | Swift | macOS only | Carbon | Minimal | Yes (Accessibility) | 737 | Active |
| Turbo Whisper | Python | Mac/Win/Linux | Global | Waveform popup | Yes | 27 | Moderate |
| HyprVoice | Go | Linux/Hyprland | Signal | None | ydotool→wtype→clip | 196 | Active |
| whisper-overlay | Rust | Linux/Wayland | evdev | Layer-shell | Clipboard | 82 | Low |
| nerd-dictation | Python | Linux | CLI | None | xdotool/ydotool | 1,800 | Low |
| Superwhisper | Swift | macOS | Single key | Rich panel | Direct injection | Commercial | Active |
| Wispr Flow | Native | Mac/Win/Android | Fn key | Flow Bar | Direct injection | Commercial | Active |
| macOS Dictation | Native | macOS | Fn×2 | Cursor-adjacent | Streaming inline | Built-in | Active |

## Build vs Buy Recommendation

**Build.** No existing tool fits vox's constraints:

1. **vox already exists as a Go core** — the transcription pipeline, history, clipboard, and file processing are built and working. The UI is an overlay on existing infrastructure, not a new product.
2. **No Go tool does this cross-platform** — HyprVoice is the only Go dictation app and it's Hyprland-only.
3. **Electron is too heavy** — OpenWhispr works but conflicts with vox's "single static binary" philosophy.
4. **The user needs both macOS and Ubuntu** — Swift-only (OpenSuperWhisper) and Linux-only (Voxtype) tools each solve only half the problem.
5. **vox's headless `vox file` pathway must remain unaffected** — bolting a UI onto someone else's tool would compromise this.

### What to build on

| Capability | Library | Rationale |
|---|---|---|
| Global hotkey | `golang.design/x/hotkey` | Only Go option for X11+macOS; Wayland falls back to IPC |
| System tray | `fyne.io/systray` | GTK-free, DBus-based, actively maintained |
| Overlay window | Gio (`gioui.org`) | Pure Go, native Wayland, pixel-level control |
| Audio levels | `gen2brain/malgo` | miniaudio bindings, no system deps on macOS |
| Clipboard | `golang.design/x/clipboard` | X11+macOS; shell to `wl-copy` on Wayland |
| Keyboard inject | xdotool/ydotool/CGEvent | Platform-specific, shell out |

## Key Technologies to Incorporate

1. **Gio over Fyne for the overlay** — Fyne's Material Design defaults fight a minimal overlay aesthetic. Gio's immediate-mode rendering gives pixel-level control for audio level bars and state indicators. Gio also has native Wayland support (no XWayland needed for the window itself).

2. **fyne-io/systray standalone** — Use the systray library independently (not the full Fyne framework). It provides `RunWithExternalLoop()` for integration with Gio's event loop.

3. **gen2brain/malgo for mic monitoring** — Replaces SoX for audio level feedback in the overlay. vox core still uses SoX for recording; the overlay monitors levels independently via malgo callbacks.

4. **Platform-specific paste helpers** — Small compiled helpers for each platform rather than runtime detection of xdotool/ydotool. On macOS, a tiny Obj-C snippet via CGO. On Linux, detect and use the best available tool.

5. **Unix socket IPC** — The daemon listens on `/tmp/vox.sock`. The CLI `vox` command, compositor hotkeys, and the tray menu all communicate through this socket. This is the glue that makes Wayland hotkeys work.
