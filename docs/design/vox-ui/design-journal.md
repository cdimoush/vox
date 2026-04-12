# Vox UI — Design Journal

## Research Synthesis — 2026-04-12

### Conclusions
- Cursor modification is not viable on any platform for cross-app feedback. Must use floating overlay or tray icon.
- Wayland global hotkeys are broken on GNOME and wlroots. Only KDE works. Must use IPC-based approach.
- No existing Go project solves this problem cross-platform. Must build.
- The Go library stack exists: golang-design/hotkey + fyne-io/systray + Gio + gen2brain/malgo.
- CGO is unavoidable across the feature set.
- Commercial tools (Superwhisper, Wispr Flow) establish the UX bar: hotkey → waveform → auto-paste → audio cues.

### Build rationale
vox core already exists and works. The UI is a thin daemon layer that wraps the existing Go packages. No existing tool provides this without either (a) being Electron-heavy, (b) being platform-locked, or (c) requiring Python.

### Starting assumptions
1. The overlay is a separate binary (`voxd`) that imports vox's Go packages as a library.
2. vox core needs minor refactoring to expose `recorder` and `transcribe` as importable APIs (they already are).
3. Two-platform support means platform-specific code behind build tags, not two codebases.
4. The user's existing GNOME hotkey setup (Super+Shift+M) can signal the daemon via socket/SIGUSR1.

## Iteration 1 — 2026-04-12

### Architecture: Daemon + Socket IPC

```
┌─────────────────────────────────────────────────┐
│                  voxd (daemon)                   │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ hotkey    │  │ socket   │  │ tray icon     │  │
│  │ listener  │  │ server   │  │ (systray)     │  │
│  │ (X11/Mac) │  │ /tmp/vox │  │               │  │
│  └─────┬─────┘  └─────┬────┘  └───────┬───────┘  │
│        │              │               │          │
│        └──────────────┼───────────────┘          │
│                       │                          │
│                       ▼                          │
│              ┌────────────────┐                  │
│              │  state machine │                  │
│              │  idle→rec→tx→  │                  │
│              │  paste→idle    │                  │
│              └───────┬────────┘                  │
│                      │                           │
│        ┌─────────────┼─────────────┐             │
│        ▼             ▼             ▼             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ malgo    │  │ recorder │  │ overlay  │       │
│  │ (levels) │  │ (SoX)    │  │ (Gio)    │       │
│  └──────────┘  └──────────┘  └──────────┘       │
│                      │                           │
│                      ▼                           │
│              ┌──────────────┐                    │
│              │ transcribe   │                    │
│              │ (OpenAI API) │                    │
│              └──────┬───────┘                    │
│                     │                            │
│           ┌─────────┼──────────┐                 │
│           ▼         ▼          ▼                 │
│     ┌──────────┐ ┌────────┐ ┌────────┐          │
│     │ clipboard│ │ paste  │ │ history│          │
│     │ write    │ │ inject │ │ append │          │
│     └──────────┘ └────────┘ └────────┘          │
└─────────────────────────────────────────────────┘

CLI:  vox          (standalone, unchanged)
      vox file     (standalone, unchanged)
      vox ui start (starts voxd)
      vox ui stop  (stops voxd)
```

### Key decisions

**Decision 1: Separate daemon binary (`voxd`) vs. subcommand (`vox daemon`)**
→ Subcommand (`vox ui start` / `vox ui stop`). Avoids a second binary in PATH. The daemon is started by `vox ui start` and runs in the background. PID file at `/tmp/vox.pid`, socket at `/tmp/vox.sock`.

**Decision 2: Gio vs. Fyne for overlay**
→ Gio. The overlay is a 200×40px floating bar with audio level visualization. Fyne's Material Design widgets are overkill. Gio's immediate-mode rendering gives direct control over the animation. Gio also renders natively on Wayland (no XWayland needed for the overlay window).

**Decision 3: Audio monitoring — malgo vs. SoX stderr parsing**
→ Both. The overlay uses malgo to show real-time levels *before* recording starts (confirming mic is live). Once recording starts, SoX handles capture (as today). The malgo stream provides level data for the overlay during recording too — both run concurrently. Alternative: use malgo for capture too and drop SoX. Deferred to implementation — SoX is proven and handles edge cases.

**Decision 4: Wayland hotkey strategy**
→ On Wayland, `voxd` does NOT register global hotkeys (can't). Instead:
1. User configures a compositor hotkey (e.g., in GNOME Settings, Hyprland config) that runs `vox ui toggle`
2. `vox ui toggle` sends a toggle command to the daemon via Unix socket
3. This works on every compositor — GNOME, Sway, Hyprland, KDE

**Decision 5: Auto-paste strategy**
→ Two-tier:
1. Write text to clipboard
2. Simulate Cmd+V (macOS via CGEventPost) or Ctrl+V (Linux via xdotool/ydotool)
3. If injection fails or isn't available: notify user text is on clipboard

**Decision 6: Overlay behavior**
→ Small floating pill (200×40px) anchored to a configurable screen position (default: top-center).
- **Idle**: hidden
- **Recording**: appears with red tint + animated audio level bars
- **Transcribing**: blue tint + spinner
- **Done**: green flash + text preview, auto-hides after 2s
- Audio cues: short tone on start, success chime on completion
- On Linux: overlay is off by default (focus-stealing risk). Tray icon + audio cues are primary. User can opt-in to overlay.

### Remaining unknowns
1. Can Gio create an always-on-top, no-taskbar-entry, click-through overlay window on both platforms?
2. Does malgo's capture callback conflict with SoX's exclusive mic access?
3. What's the right UX for the first-run Accessibility permission prompt on macOS?

## Iteration 2 — 2026-04-12

### Changes
- Resolved overlay window feasibility: Gio supports `app.Window` with no decorations on both platforms. On macOS, set `NSPanel` level via CGO for always-on-top without activation. On Linux/Wayland, use `wlr-layer-shell` protocol (Sway/Hyprland) or just a positioned window (GNOME).
- Resolved mic conflict: malgo and SoX can coexist — SoX uses ALSA/PulseAudio and malgo uses its own miniaudio backend. Both can open the same capture device concurrently on modern audio stacks (PipeWire, CoreAudio).
- Added: config file at `~/.vox/config.toml` for overlay position, hotkey, paste method, overlay enabled/disabled.

### Decisions made
- **Config file**: TOML at `~/.vox/config.toml`. Keeps the env-var-only rule for vox core, but the UI daemon needs persistent configuration. Overlay preferences, paste method, and platform-specific settings go here.
- **Overlay on Linux**: Default OFF on Wayland, ON on X11. Tray icon + audio cues are the primary feedback on Wayland. User can enable overlay if their compositor handles it well.
- **First-run macOS**: On first launch, `voxd` checks `AXIsProcessTrusted()`. If false, shows a Gio window explaining why Accessibility is needed and a button to open System Settings. Auto-paste degrades to clipboard-only until granted.

### Remaining unknowns
- Implementation-level: exact Gio widget layout, animation frame rate, audio cue file format
- These are not architectural — ready for final design.
