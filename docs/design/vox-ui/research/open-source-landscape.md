# Open-Source Landscape: Dictation Overlays, Global Hotkey Tools, and Go Desktop Libraries

**Research bead:** system_designer-d8s.1
**Date:** 2026-04-12
**Topic:** vox-ui — cross-platform dictation UI overlay for a Go speech-to-text CLI

---

## 1. Complete Applications: Dictation Overlays with Global Hotkeys

### 1.1 OpenWhispr
- **Repo:** https://github.com/OpenWhispr/openwhispr
- **Stars:** 2,400
- **Last commit:** Active (early 2026)
- **Language:** JavaScript/TypeScript (Electron)
- **License:** MIT
- **Platforms:** macOS 12+, Windows 10+, Linux (X11/Wayland — deb, rpm, AppImage, Flatpak)
- **Key features:**
  - Global hotkey (default: backtick `` ` ``) triggers record/stop from anywhere
  - Auto-pastes transcribed text at cursor position
  - Local models: NVIDIA Parakeet, whisper.cpp; cloud: OpenAI, Groq, Anthropic, Google
  - Glassmorphic chat overlay with streaming AI agent mode
  - Google Calendar integration + live meeting transcription
  - Custom dictionary with auto-learn; full-text search; local vector search (Qdrant)
  - Native helpers per-platform: Globe key listener (macOS), fast-paste binary (Linux/Windows)
- **Limitations:**
  - Push-to-talk unavailable on GNOME Wayland (known issue #240)
  - Heavy — Electron runtime, Node.js 22+ for dev builds
  - Linux paste fallback chain: native binary → wtype → ydotool → xdotool
- **Maintenance:** Very active, largest community of the field

---

### 1.2 Turbo Whisper
- **Repo:** https://github.com/knowall-ai/turbo-whisper
- **Stars:** 27
- **Last commit:** January 22, 2025
- **Language:** Python
- **License:** MIT
- **Platforms:** Linux (Ubuntu PPA, AUR), macOS (Homebrew), Windows
- **Key features:**
  - Global hotkey Ctrl+Shift+Space from anywhere
  - Animated waveform visualization in system tray popup
  - OpenAI Whisper API or self-hosted faster-whisper-server
  - System tray icon with autostart support
  - Auto-types transcribed text into focused window
  - Clipboard integration
  - Closest analog to what vox-ui is targeting (SuperWhisper-like UX, Linux-friendly)
- **Limitations:**
  - Small community, limited to ~27 stars
  - Requires API key or self-hosted GPU/CPU whisper server
  - macOS needs accessibility permissions
  - PyAudio Windows install is awkward
- **Maintenance:** Moderate; last commit Jan 2025

---

### 1.3 Voxtype
- **Repo:** https://github.com/peteonrails/voxtype
- **Stars:** 616
- **Last commit:** Active (2025–2026)
- **Language:** Rust
- **License:** Not specified in search results
- **Platforms:** Linux (Wayland-native; X11 fallback); no macOS/Windows
- **Key features:**
  - Push-to-talk optimized for Wayland; X11 fallback via evdev
  - 7 transcription engines: Whisper, Parakeet, Moonshine, SenseVoice, Paraformer, Dolphin, Omnilingual
  - Fully offline by default (whisper.cpp local)
  - GPU acceleration: Vulkan, CUDA, Metal, ROCm
  - CJK + 1600+ language support
  - Meeting mode with speaker attribution and export
  - Text injection: wtype → dotool → ydotool → clipboard
- **Limitations:**
  - Linux-exclusive
  - Requires glibc 2.38+ (Ubuntu 24.04+, Fedora 39+)
  - Needs PipeWire or PulseAudio
  - Input group membership for X11 hotkeys
- **Maintenance:** Active

---

### 1.4 OpenSuperWhisper
- **Repo:** https://github.com/Starmel/OpenSuperWhisper
- **Stars:** 737
- **Last commit:** March 3, 2026
- **Language:** Swift (93.5%)
- **License:** MIT
- **Platforms:** macOS — Apple Silicon (ARM64) only; Intel incomplete
- **Key features:**
  - Global keyboard shortcuts with hold-to-record mode
  - Two transcription engines: Whisper and Parakeet
  - Drag-and-drop audio file support
  - Microphone selection (including Bluetooth)
  - Asian language autocorrect
  - 180 commits, 7 releases — stable and actively maintained
- **Limitations:**
  - macOS-only (no Linux, no Windows)
  - Intel Mac support incomplete (TODO #15)
  - No streaming transcription yet
  - No agent/LLM post-processing
- **Maintenance:** Active (March 2026)

---

### 1.5 WhisperWriter
- **Repo:** https://github.com/savbell/whisper-writer
- **Stars:** 1,000
- **Last commit:** May 28, 2024 (PyQt5 migration)
- **Language:** Python
- **License:** Not specified
- **Platforms:** Linux, macOS, Windows; Docker for GPU
- **Key features:**
  - Four recording modes: continuous, VAD, press-to-toggle, hold-to-record
  - PyQt5 GUI with settings panel
  - Faster-whisper (local) or OpenAI API
  - Customizable hotkey (default: Ctrl+Shift+Space)
  - Auto-types into active window
  - Voice activity detection filter
- **Limitations:**
  - Python 3.11+; GPU requires NVIDIA CUDA 12
  - Relatively simple; no overlay/animation UI
- **Maintenance:** Moderate (last significant update mid-2024)

---

### 1.6 whisper-overlay (Wayland)
- **Repo:** https://github.com/oddlama/whisper-overlay
- **Stars:** 82
- **Last commit:** June 16, 2024
- **Language:** Rust (client) + Python (server)
- **License:** MIT
- **Platforms:** Linux Wayland only (sway, Hyprland, etc.)
- **Key features:**
  - True overlay rendered via wlr-layer-shell (composited on top of everything)
  - Dual-model transcription: fast base model for live display + large-v3 for final output
  - Server-client split for remote GPU processing
  - Waybar integration with status indicators
  - evdev for global hotkey (no window focus needed)
- **Limitations:**
  - Requires custom RealtimeSTT fork
  - Wayland-only — X11 explicitly not planned
  - Single client at a time
  - Requires input group membership
- **Maintenance:** Low (last commit June 2024)

---

### 1.7 nerd-dictation
- **Repo:** https://github.com/ideasman42/nerd-dictation
- **Stars:** 1,800
- **Last commit:** January 2023
- **Language:** Python
- **License:** GPL-3.0
- **Platforms:** Linux (X11, Wayland via wtype/ydotool, TTY)
- **Key features:**
  - Single-file Python script, minimal dependencies
  - Offline via VOSK-API (not Whisper)
  - Multiple audio backends: PulseAudio, Sox, PipeWire
  - Multiple output methods: xdotool, ydotool, dotool, wtype, stdout
  - Hackable text post-processing via user config scripts
  - Suspend/resume, auto-timeout
- **Limitations:**
  - VOSK only (no Whisper support)
  - VOSK output is all-lowercase (no capitalization)
  - No GUI, no overlay, no tray icon — purely CLI activation
  - No daemon mode — must be manually invoked
  - Last real commit January 2023 (maintenance-only since)
- **Maintenance:** Low / archived-in-practice

---

### 1.8 Speech Note (dsnote)
- **Repo:** https://github.com/mkiol/dsnote
- **Stars:** 1,400
- **Last commit:** January 13, 2025
- **Language:** C++ / Qt (CMake)
- **License:** Not specified (Flatpak distributed)
- **Platforms:** Linux Desktop (Flatpak, Arch AUR, openSUSE Packman); Sailfish OS
- **Key features:**
  - Full Qt desktop app for note-taking + dictation + translation
  - Multiple STT engines: Whisper (whisper.cpp, Faster Whisper), Vosk, Coqui, april-asr
  - Multiple TTS engines: Piper, espeak-ng, Kokoro, F5-TTS, Parler-TTS, Mimic 3, S.A.M.
  - Machine translation via Bergamot Translator
  - 90+ languages
  - GPU acceleration via Vulkan (Intel, AMD, NVIDIA)
  - Fully offline
- **Limitations:**
  - Dedicated note-taking app, not a system-wide overlay/hotkey tool
  - No global push-to-talk while using other apps
  - Large install (1.2 GiB base Flatpak)
  - x86-64 only for Faster Whisper and Coqui
- **Maintenance:** Active (Jan 2025)

---

### 1.9 waystt
- **Repo:** https://github.com/sevos/waystt
- **Stars:** 120
- **Last commit:** Active (2025)
- **Language:** Rust
- **License:** GPL-3.0
- **Platforms:** Linux Wayland (Hyprland, Niri, GNOME, KDE); PipeWire required
- **Key features:**
  - Minimal signal-driven: starts via keybinding signal, transcribes, outputs to stdout, exits
  - UNIX philosophy — pipes output to ydotool, clipboard, or any command
  - Providers: OpenAI Whisper API, Google STT, local whisper-rs
  - Audio feedback via beeps
  - AUR package available
- **Limitations:**
  - Wayland-only
  - No overlay UI — stdout only
  - Requires PipeWire
  - Cloud providers need API keys; local is disk-heavy
- **Maintenance:** Active

---

### 1.10 HyprVoice
- **Repo:** https://github.com/LeonardoTrapani/hyprvoice
- **Stars:** 196
- **Last commit:** February 25, 2026
- **Language:** Go (99.3%)
- **License:** MIT
- **Platforms:** Linux / Wayland / Hyprland (AUR); PipeWire required
- **Key features:**
  - **Written in Go** — the only complete dictation application in Go found
  - 26 STT models across cloud + local providers (OpenAI, Groq, Mistral, ElevenLabs, Deepgram, whisper.cpp)
  - LLM post-processing option
  - Toggle-based recording workflow
  - Text injection: ydotool → wtype → clipboard fallback
  - Hot-reload config
- **Limitations:**
  - Hyprland/Wayland-only; no macOS or Windows
  - No overlay UI — no visual recording indicator beyond compositor signals
  - Requires PipeWire
  - Small community (196 stars)
- **Maintenance:** Active (Feb 2026)

---

### 1.11 BlahST
- **Repo:** https://github.com/QuantiusBenignus/BlahST
- **Stars:** 172
- **Last commit:** Active (2025)
- **Language:** Shell/Bash (zsh)
- **License:** Not specified
- **Platforms:** Linux (X11 + Wayland); multiple DEs: GNOME, KDE, XFCE4, Cinnamon, LXQt
- **Key features:**
  - Shell-script wrapper around whisper.cpp
  - `wsi` for fast single-shot STT; `blooper` for continuous hands-free dictation loop
  - `wsiAI` / `blahstbot` for local LLM integration via llama.cpp
  - `blahstream` for streaming speech-to-speech chat
  - Works on both X11 and Wayland
- **Limitations:**
  - Requires zsh shell
  - No GUI or overlay — purely CLI/scripting
  - Silence detection unreliable in noisy environments
  - Requires manual config of paths/server IPs
- **Maintenance:** Active

---

## 2. Go Libraries for Building the Stack

### 2.1 golang-design/hotkey
- **Repo:** https://github.com/golang-design/hotkey
- **Stars:** 258
- **Last release:** v0.4.1 (December 28, 2022)
- **Language:** Go (86%), Objective-C (8%), C (6%)
- **Platforms:** macOS, Linux (X11 only — no Wayland), Windows
- **Key features:**
  - Register system-level global hotkeys without window focus
  - Modifier + key combinations via `Keydown()` / `Keyup()` channels
  - `mainthread` utilities for macOS (hotkeys must run on main thread)
- **Limitations:**
  - No Wayland support on Linux — X11 only
  - Last release Dec 2022; low commit count (24 total)
  - AutoRepeat quirk on Linux may trigger spurious Keyup events
- **Maintenance:** Low / stable

---

### 2.2 getlantern/systray
- **Repo:** https://github.com/getlantern/systray
- **Stars:** 3,700
- **Last commit:** May 2, 2023
- **Language:** Go (78%), Objective-C (11%), C (10%)
- **License:** Apache-2.0
- **Platforms:** Windows, macOS, Linux (requires gtk3 + libayatana-appindicator3)
- **Key features:**
  - Place icon + menu in system notification area
  - Checkable and disableable menu items
  - Goroutine-safe
  - CGO-based (CGO_ENABLED=1 required)
- **Limitations:**
  - Linux requires GTK3 and libayatana headers — CGO dependency
  - Last commit May 2023; not actively maintained
  - Does not handle global hotkeys (separate concern from tray icon)
- **Maintenance:** Low / stable (used as dependency by many projects)
- **Forks of note:** `fyne-io/systray` (removed GTK dependency), `energye/systray`

---

### 2.3 fyne-io/systray (fork)
- **Repo:** https://pkg.go.dev/fyne.io/systray
- **Stars:** Part of Fyne ecosystem
- **Language:** Go
- **Platforms:** Windows, macOS, Linux
- **Key difference from getlantern/systray:** Removed GTK dependency; integrates cleanly with Fyne apps
- **Maintenance:** Active (maintained by Fyne team)

---

### 2.4 mutablelogic/go-whisper
- **Repo:** https://github.com/mutablelogic/go-whisper
- **Stars:** 183
- **Last release:** v0.0.39 (January 27, 2026)
- **Language:** Go (96%)
- **License:** Apache-2.0
- **Platforms:** Linux (AMD64, ARM64), macOS (Metal GPU)
- **Key features:**
  - Unified STT + translation service in Go wrapping whisper.cpp
  - Local GPU acceleration: CUDA, Vulkan, Metal
  - CLI tool + HTTP REST API server
  - Docker deployment
  - SRT/VTT/JSON/text output formats
  - Realtime transcription streaming
  - Speaker diarization
- **Application type:** Full service application — can run as a local daemon with HTTP API
- **Limitations:**
  - "Currently in development and subject to change"
  - No hotkey or overlay — pure transcription service
- **Maintenance:** Active (Jan 2026)

---

### 2.5 fyne-io/fyne
- **Repo:** https://github.com/fyne-io/fyne
- **Stars:** 28,100
- **Last release:** v2.7.3 (February 21, 2026)
- **Language:** Go (96%)
- **License:** BSD-3-Clause
- **Platforms:** Desktop (Windows, macOS, Linux) + Mobile (Android, iOS)
- **Key features:**
  - Full cross-platform UI toolkit in Go; Material Design-inspired
  - System tray via fyne-io/systray fork (GTK-free)
  - Custom canvas for overlay-style windows (transparency, frameless)
  - Actively maintained; 12,394 commits
- **Limitations:**
  - No built-in global hotkey support (needs golang-design/hotkey alongside)
  - CGO required for desktop targets
  - Heavier than a pure tray approach
- **Maintenance:** Very active

---

## 3. Key Patterns and Observations

### Text Injection Stack (Linux)
All serious Linux dictation tools use the same fallback chain:
1. **wtype** — Wayland virtual keyboard protocol (best Unicode/CJK support)
2. **dotool** — XKB-based; handles non-US keyboard layouts on Wayland
3. **ydotool** — uinput kernel device; works on X11, Wayland, TTY
4. **xdotool** — X11 only; last resort for XWayland apps
5. **clipboard paste** — wl-copy / xclip / xsel → simulated Ctrl+V

### Global Hotkey Approaches
| Approach | X11 | Wayland | Notes |
|---|---|---|---|
| golang-design/hotkey | Yes | No | XGrabKey; doesn't work on Wayland |
| evdev (kernel-level) | Yes | Yes | Requires input group; used by whisper-overlay, Voxtype |
| Compositor keybindings (sway/Hyprland) | No | Yes | External shell command trigger; no daemon needed |
| pynput / evdev in Python | Yes | Yes | Common in Python tools |

### The macOS Gap
No pure Go dictation overlay exists for macOS. Swift (OpenSuperWhisper) is the native option. Python tools (whisper-dictation, OpenWhispr) use accessibility APIs.

### No Go Dictation Overlay Exists Cross-Platform
HyprVoice (Go, Linux/Wayland only) is the closest. There is no Go project combining:
- Global hotkey (cross-platform)
- Visual overlay/tray feedback
- STT invocation + text injection
- macOS + Linux support

This is the gap vox-ui would fill.

---

## 4. Summary Table

| Project | Lang | Stars | Last Active | Platforms | Hotkey | Overlay UI | Tray |
|---|---|---|---|---|---|---|---|
| OpenWhispr | JS/TS | 2,400 | 2026 | Mac/Win/Linux | Yes | Glassmorphic | Yes |
| Turbo Whisper | Python | 27 | Jan 2025 | Mac/Win/Linux | Yes | Waveform | Yes |
| Voxtype | Rust | 616 | 2026 | Linux only | Yes | None | No |
| OpenSuperWhisper | Swift | 737 | Mar 2026 | macOS only | Yes | Minimal | No |
| WhisperWriter | Python | 1,000 | May 2024 | Mac/Win/Linux | Yes | PyQt5 | No |
| whisper-overlay | Rust | 82 | Jun 2024 | Linux Wayland | Yes | Full overlay | No |
| nerd-dictation | Python | 1,800 | Jan 2023 | Linux only | No | None | No |
| Speech Note | C++/Qt | 1,400 | Jan 2025 | Linux only | No | Full app | No |
| waystt | Rust | 120 | 2025 | Linux Wayland | Via signal | None | No |
| HyprVoice | **Go** | 196 | Feb 2026 | Linux Wayland | Via signal | None | No |
| BlahST | Shell | 172 | 2025 | Linux | No | None | No |
| golang-design/hotkey | **Go** | 258 | Dec 2022 | Mac/Win/Linux X11 | Library | — | — |
| getlantern/systray | **Go** | 3,700 | May 2023 | Mac/Win/Linux | No | — | Library |
| go-whisper | **Go** | 183 | Jan 2026 | Linux/macOS | No | None | No |
| fyne | **Go** | 28,100 | Feb 2026 | Mac/Win/Linux | No | UI toolkit | Via fork |
