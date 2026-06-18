# Vox UI — System Design

## Executive Summary

Vox UI adds a global hotkey-driven dictation overlay to vox, the existing Go CLI for speech-to-text. The user presses a hotkey from any application, speaks, and the transcribed text is auto-pasted at the cursor. A small floating overlay shows recording state and audio levels. This works on both macOS and Ubuntu via a single Go codebase with platform-specific build tags — not two separate applications. The architecture is a background daemon (`vox ui start`) that listens for hotkey/IPC signals, manages recording via the existing vox packages, and handles visual feedback through a Gio overlay + system tray icon.

## Problem Statement

The user has vox working as a CLI tool on both Ubuntu and macOS. The daily workflow is: open terminal → type `vox` → speak → paste from clipboard. This works but has three friction points:

1. **Terminal interruption** — must switch to a terminal to start recording, breaking flow in whatever app you're using
2. **No mic feedback** — if the mic isn't working, you don't know until transcription returns empty
3. **Manual paste** — must Ctrl+V/Cmd+V after recording, adding a step to every dictation

The user already has GNOME hotkeys (Super+Shift+M) configured for an older script, proving the value. This design replaces that ad-hoc setup with an integrated, cross-platform solution.

## Value Proposition

No existing tool provides a lightweight Go-based dictation overlay for both macOS and Linux. OpenWhispr (Electron) is too heavy. OpenSuperWhisper (Swift) is macOS-only. HyprVoice (Go) is Hyprland-only with no overlay. Vox UI fills this gap by layering a thin daemon on top of vox's proven transcription pipeline.

## User Stories

1. **As a user typing in any application**, I want to press a hotkey and start dictating, so that I don't need to switch to a terminal.

2. **As a user dictating**, I want to see animated audio level bars confirming my mic is active, so that I know the recording is working before I finish speaking.

3. **As a user who finishes dictating**, I want the transcribed text auto-pasted at my cursor position, so that I don't need to manually Ctrl+V.

4. **As a user running vox on a headless server** (Relay, audio file processing), I want `vox file` to continue working exactly as it does today, with no UI dependencies.

5. **As a user on either macOS or Ubuntu**, I want a single installation that works on my platform without maintaining two different tools.

## Landscape Context

15 projects surveyed across Python, Rust, Swift, Electron, and Go. Full analysis in [landscape-synthesis.md](landscape-synthesis.md). The key gap: no Go tool provides cross-platform (macOS + Linux) hotkey → overlay → auto-paste. The commercial UX bar is set by Superwhisper and Wispr Flow: single-key activation, waveform feedback, color-coded states, audio cues, direct cursor injection.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     vox binary (unchanged)                   │
│                                                              │
│  vox          → record + transcribe + clipboard (CLI)        │
│  vox file     → transcribe audio file (headless)             │
│  vox ls/cp/show/clear → history management                   │
│  vox ui start → launch daemon (new)                          │
│  vox ui stop  → stop daemon (new)                            │
│  vox ui toggle→ send toggle signal to daemon (new)           │
│  vox ui status→ check if daemon is running (new)             │
└──────────────────────────────┬──────────────────────────────┘
                               │
              ┌────────────────┼────────────────────┐
              │        vox daemon (voxd)            │
              │      (started by vox ui start)       │
              │                                      │
              │  Input Layer                         │
              │  ├─ Global hotkey (X11/macOS)        │
              │  ├─ Unix socket (/tmp/vox.sock)      │
              │  └─ System tray menu                 │
              │                                      │
              │  State Machine                       │
              │  idle → recording → transcribing     │
              │    ↑        ↓            ↓           │
              │    └── done ←────────────┘           │
              │                                      │
              │  Output Layer                        │
              │  ├─ Gio overlay (audio bars, state)  │
              │  ├─ System tray icon (state color)   │
              │  ├─ Audio cues (start/done tones)    │
              │  ├─ Clipboard + paste injection      │
              │  └─ History append                   │
              │                                      │
              │  Shared with vox core:               │
              │  ├─ recorder/ (SoX)                  │
              │  ├─ transcribe/ (OpenAI API)         │
              │  ├─ clipboard/ (platform detection)  │
              │  └─ history/ (JSONL storage)         │
              └──────────────────────────────────────┘
```

### Data flow

1. **Trigger**: User presses hotkey (captured by golang-design/hotkey on X11/macOS) or compositor signals `vox ui toggle` (on Wayland)
2. **vox ui toggle** sends `{"cmd":"toggle"}` to Unix socket at `/tmp/vox.sock`
3. **Daemon receives toggle**: transitions from idle → recording
4. **Recording**: SoX `rec` starts capturing audio. Concurrently, malgo opens the mic for level monitoring. Levels stream to the Gio overlay at ~30 FPS.
5. **Stop**: Second hotkey press (or `vox ui toggle`) sends stop signal. SoX receives SIGINT, writes WAV.
6. **Transcribe**: WAV sent to OpenAI Whisper API. Overlay shows "transcribing" state with spinner.
7. **Paste**: Text written to clipboard. Paste keystroke simulated (Cmd+V on macOS, Ctrl+V on Linux). If injection not available, tray notification says "Copied to clipboard."
8. **History**: Entry appended to `~/.vox/history.jsonl` (same store as CLI vox).

## Component Breakdown

### 1. CLI Extensions (`cmd/vox/ui.go`)

**Purpose**: Add `vox ui {start|stop|toggle|status}` subcommands to the existing vox binary.

**Interfaces**:
- `vox ui start` — forks the daemon process, writes PID to `/tmp/vox.pid`
- `vox ui stop` — sends shutdown command via socket, waits for clean exit
- `vox ui toggle` — sends toggle command via socket (used by compositor hotkeys on Wayland)
- `vox ui status` — checks if daemon is running, prints state

**Key decisions**: These are thin socket clients. No daemon code runs in-process — the daemon is a separate goroutine tree started by `start`.

### 2. Daemon Core (`daemon/`)

**Purpose**: Long-running process that manages the dictation lifecycle.

**Interfaces**:
```go
type Daemon struct {
    state    State           // idle, recording, transcribing, done
    config   *Config         // from ~/.vox/config.toml
    recorder *recorder.Recorder
    overlay  *overlay.Overlay
    tray     *tray.Tray
    sock     net.Listener    // Unix socket
}

func (d *Daemon) Run(ctx context.Context) error
func (d *Daemon) Toggle()           // start/stop recording
func (d *Daemon) Cancel()           // abort current operation
func (d *Daemon) Shutdown()
```

**Key decisions**:
- State machine with 4 states: `idle → recording → transcribing → done → idle`
- All state transitions emit events consumed by overlay, tray, and audio cue subsystems
- Cancel is always available (second hotkey press during transcription)
- Daemon writes logs to `~/.vox/daemon.log`

### 3. Hotkey Listener (`daemon/hotkey/`)

**Purpose**: Register global hotkeys on platforms that support it.

**Platform behavior**:
- **macOS**: `golang.design/x/hotkey` with Carbon `RegisterEventHotKey`. Default: Ctrl+Shift+Space. Must run on main thread (handled by Gio's event loop).
- **Linux X11**: `golang.design/x/hotkey` with `XGrabKey`. Same default hotkey.
- **Linux Wayland**: No hotkey registration. Prints setup instructions on first run:
  ```
  Wayland detected. Global hotkeys require compositor configuration.

  GNOME:  Settings → Keyboard → Custom Shortcuts
          Name: Vox Toggle
          Command: vox ui toggle

  Sway:   bindsym $mod+Shift+space exec vox ui toggle

  Hyprland: bind = $mainMod SHIFT, space, exec, vox ui toggle
  ```

**Key decisions**: Hotkey is configurable in `~/.vox/config.toml`. On Wayland, the daemon still starts — it just doesn't register hotkeys itself. The socket IPC handles all triggering.

### 4. Overlay Window (`daemon/overlay/`)

**Purpose**: Floating pill-shaped window showing recording state and audio levels.

**Technology**: Gio (`gioui.org`) — immediate-mode rendering, native Wayland support, no CGO for the rendering itself.

**Visual design**:
```
┌─────────────────────────────────┐
│  ● ████░░░░██████░░░░███░░      │   ← recording state (200×40px)
└─────────────────────────────────┘

┌─────────────────────────────────┐
│  ◌ Transcribing...              │   ← transcribing state
└─────────────────────────────────┘
```

**Behavior**:
- Hidden when idle
- Appears on recording start, anchored to configurable position (default: top-center, 20px from top)
- Shows real-time audio level bars (data from malgo callback)
- Transitions to "Transcribing..." with spinner animation
- Brief green flash + first ~40 chars of transcript on completion, then auto-hides after 2s
- No window decorations, no taskbar entry
- Click-through (input passes to app beneath) except for a small dismiss button

**Platform specifics**:
- macOS: `NSPanel` with `NSWindowLevelFloating` via CGO for always-on-top without activation
- Linux X11: `_NET_WM_STATE_ABOVE` + `_NET_WM_STATE_SKIP_TASKBAR` hints
- Linux Wayland: `wlr-layer-shell` overlay on wlroots compositors; regular always-on-top window on GNOME (may not stay on top — acceptable degradation)

**Key decisions**:
- Overlay is **off by default on Wayland** (focus-stealing risk on some compositors). Tray icon + audio cues are primary. User opts in via `overlay_enabled = true` in config.
- Overlay is **on by default on macOS and X11** where it works reliably.

### 5. System Tray (`daemon/tray/`)

**Purpose**: Persistent status indicator + menu for daemon control.

**Technology**: `fyne.io/systray` (or `tailscale/systray` fork) with `RunWithExternalLoop()` for Gio integration.

**Tray states** (icon color changes):
- Gray microphone: idle
- Red microphone: recording
- Blue microphone: transcribing
- Green checkmark: done (2s, then back to gray)

**Menu items**:
- Toggle Recording (with hotkey hint)
- History (opens terminal `vox ls`)
- Settings (opens `~/.vox/config.toml` in editor)
- Quit

**Platform notes**:
- macOS: works out of the box (NSStatusItem)
- Linux: requires AppIndicator support. GNOME users need the AppIndicator extension. If AppIndicator is unavailable, daemon still works — just no tray icon.

### 6. Audio Cues (`daemon/audio/`)

**Purpose**: Audible feedback for state transitions.

**Implementation**: Play short WAV/OGG samples via malgo output device:
- Recording start: short ascending tone (~200ms)
- Transcription complete: two-note chime (~300ms)
- Error: low tone (~200ms)

**Key decisions**: Audio cues are the primary feedback mechanism on Linux Wayland (where overlay may be off and tray may be unsupported). They are enabled by default on all platforms. Configurable volume and disable option in config.

### 7. Paste Injection (`daemon/paste/`)

**Purpose**: Simulate keyboard paste after clipboard write.

**Strategy per platform**:

| Platform | Primary | Fallback |
|---|---|---|
| macOS | CGEventPost Cmd+V (requires Accessibility) | Clipboard only |
| Linux X11 | `xdotool key ctrl+v` | Clipboard only |
| Linux Wayland (wlroots) | `wtype -k ctrl+v` or `ydotool key ctrl+v` | Clipboard only |
| Linux Wayland (GNOME) | `ydotool key ctrl+v` (if daemon running) | Clipboard only |

**Implementation**:
```go
func Paste(text string) error {
    clipboard.Write(text)
    if err := injectPaste(); err != nil {
        notify("Text copied to clipboard")
        return nil // graceful degradation
    }
    return nil
}
```

**Key decisions**:
- Always write to clipboard first (safe fallback)
- Paste injection is best-effort — never fail the whole operation if injection doesn't work
- No clipboard restoration (avoids race conditions; the transcribed text staying on clipboard is useful)
- macOS: check `AXIsProcessTrusted()` at daemon startup; if false, skip injection and show one-time notification about enabling Accessibility

### 8. Configuration (`~/.vox/config.toml`)

```toml
[hotkey]
# Modifier+Key combo (used on X11 and macOS; ignored on Wayland)
modifiers = ["ctrl", "shift"]
key = "space"

[overlay]
enabled = true          # false by default on Wayland
position = "top-center" # top-left, top-center, top-right, bottom-left, etc.
width = 200
height = 40

[audio]
cues_enabled = true
cue_volume = 0.5        # 0.0-1.0

[paste]
auto_paste = true       # attempt keystroke injection after clipboard write
method = "auto"         # auto, clipboard-only, xdotool, ydotool, wtype
```

**Key decisions**: Config file is only for the daemon — vox core remains env-var-only. Config is loaded at daemon startup and can be reloaded via tray menu or SIGHUP.

## Technology Choices

| Choice | Technology | Why |
|---|---|---|
| Language | Go (matching vox core) | Single codebase, shared packages, cross-compile |
| Overlay rendering | Gio | Pure Go rendering, native Wayland, pixel-level control, no Material Design baggage |
| System tray | fyne-io/systray | GTK-free, DBus-based, actively maintained, standalone use without Fyne framework |
| Global hotkey | golang-design/hotkey | Only Go option for X11+macOS; mature enough despite dormant maintenance |
| Audio levels | gen2brain/malgo | miniaudio bindings, callback-based, no system deps on macOS |
| Clipboard | golang-design/clipboard + wl-copy | Cross-platform with Wayland shell-out |
| IPC | Unix domain socket | Simple, reliable, works everywhere, no D-Bus complexity |
| Config | TOML | Human-readable, well-supported in Go (pelletier/go-toml) |

**CGO requirement**: CGO is unavoidable for systray (DBus/Cocoa), hotkey (X11/Carbon), malgo (miniaudio), and macOS paste injection (CGEventPost). Build with `CGO_ENABLED=1`. Cross-compilation requires platform-specific toolchains.

## Developability: Agent vs. Human

### What an agent CAN build on a headless server

| Component | Agent-buildable? | Notes |
|---|---|---|
| CLI extensions (`vox ui start/stop/toggle`) | Yes | Pure Go, socket IPC, fully testable headless |
| Daemon core + state machine | Yes | Pure Go logic, unit-testable |
| Socket IPC server | Yes | Standard net/unix, testable |
| Config parsing | Yes | TOML parsing, unit-testable |
| Paste injection logic | Partially | Can write the detection/fallback code; can't test actual paste |
| History integration | Yes | Already exists and works |
| Gio overlay widget | Partially | Can write the rendering code; can't verify visual output |
| System tray setup | No | Requires display server |
| Hotkey registration | No | Requires X11/macOS display |
| Audio cues | No | Requires audio output device |
| malgo mic monitoring | No | Requires microphone |

**Bottom line**: ~60% of the code is agent-buildable and testable on a headless server. The remaining ~40% (overlay visuals, tray, hotkey, audio) requires a human with a display + mic for testing. The recommended approach:

1. **Agent builds**: daemon core, state machine, IPC, config, paste logic, CLI extensions — all with comprehensive unit tests
2. **Human tests**: overlay rendering, tray behavior, hotkey registration, audio cues, end-to-end flow on both platforms
3. **Agent iterates**: fixes based on human feedback, adds edge case handling

### Development hardware needed
- An Ubuntu machine with display (for Linux testing) — the user's Ubuntu laptop
- A macOS machine (for macOS testing) — the user's MacBook
- Both need: microphone, Go 1.26+, SoX, platform-specific deps (libayatana-appindicator3 on Ubuntu)

## Risk Register

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| Wayland global hotkeys never improve | Medium | High | Already designed around this — socket IPC handles all Wayland triggering via compositor keybindings |
| Gio overlay causes focus-stealing on Wayland compositors | Medium | Medium | Overlay off by default on Wayland; tray + audio cues as primary feedback |
| malgo and SoX conflict on mic access | High | Low | Modern audio stacks (PipeWire, CoreAudio) allow concurrent capture; test and fall back to SoX-only level parsing if conflict detected |
| golang-design/hotkey stops working on new macOS | Medium | Low | Carbon API has been "deprecated" since 10.8 but still works in macOS 15; fallback to CGEventTap via CGO if needed |
| fyne-io/systray breaks on GNOME without AppIndicator | Low | Medium | Detect missing AppIndicator at startup; daemon works without tray (audio cues + overlay remain) |
| CGO cross-compilation pain | Medium | High | Don't cross-compile. Build on each target platform (user has both machines). CI can use platform-specific runners if needed later. |
| Auto-paste race condition (clipboard read before paste completes) | Medium | Medium | Don't restore clipboard; leave transcribed text on clipboard (it's useful there) |

## Implementation Roadmap

Build order optimized for incremental value and testability:

### Phase 1: Daemon Foundation (agent-buildable)
1. **Refactor vox packages for library use** — ensure `recorder`, `transcribe`, `clipboard`, `history` can be imported by the daemon without pulling in CLI code
2. **Daemon core + state machine** — idle/recording/transcribing/done transitions with event emission
3. **Unix socket IPC** — server in daemon, client in `vox ui toggle/start/stop/status`
4. **CLI extensions** — `vox ui` subcommand family
5. **Config file** — TOML parsing, defaults, platform detection

### Phase 2: Platform Integration (requires display)
6. **System tray** — icon + state-driven color changes + menu
7. **Global hotkey registration** — X11 and macOS, with Wayland detection and instructions
8. **Paste injection** — platform-specific paste with graceful fallback
9. **Audio cues** — tone playback via malgo output

### Phase 3: Overlay (requires display)
10. **Gio overlay window** — floating pill, no decorations, always-on-top
11. **Audio level visualization** — malgo mic capture → RMS → animated bars
12. **State-driven overlay** — recording/transcribing/done visual transitions

### Phase 4: Polish
13. **macOS Accessibility onboarding** — first-run permission check + explanation window
14. **Wayland first-run instructions** — detect compositor, print keybinding instructions
15. **Config hot-reload** — SIGHUP or tray menu trigger
16. **Installer script** — `vox ui install` that sets up autostart (systemd user service on Linux, launchd plist on macOS)

## Open Questions

1. **Should the daemon auto-start on login?** Probably yes, via systemd user service (Linux) or launchd plist (macOS). But this should be opt-in via `vox ui install`.

2. **Should malgo replace SoX entirely?** malgo can capture audio directly, removing the SoX dependency for recording. But SoX is proven, handles edge cases, and vox core already depends on it. Could be a future simplification.

3. **Should the overlay show live transcript (streaming)?** OpenAI's Whisper API doesn't support streaming transcription. Would require switching to a streaming-capable API or local model. Deferred.

4. **Push-to-talk vs. toggle?** The Superwhisper pattern (tap = toggle, hold = push-to-talk, same key) is ergonomically ideal but harder to implement with golang-design/hotkey's AutoRepeat issues on X11. Start with toggle-only; add push-to-talk as enhancement.

5. **Multi-monitor overlay positioning?** Which monitor does the overlay appear on? Default: follow the focused window's monitor. May require platform-specific monitor detection.

## Appendix

- [Scope](scope.md)
- [Landscape Synthesis](landscape-synthesis.md)
- [Design Journal](design-journal.md)
- Research: [OSS](research/open-source-landscape.md) | [Commercial](research/commercial-products.md) | [Libraries](research/libraries-and-sdks.md) | [Community](research/community-patterns.md)
