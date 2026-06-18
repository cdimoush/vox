# Vox UI — Cross-Platform Dictation Overlay

## What we're designing

A lightweight, always-available dictation overlay for **vox** — the existing Go CLI that records speech via SoX, transcribes via OpenAI Whisper, and copies to clipboard. The user currently uses vox by opening a terminal, typing `vox`, speaking, then pasting. This design adds a system-level UI layer so the user can:

1. Press a global hotkey (e.g., Super+Shift+M) to start recording — no terminal window appears
2. See visual feedback via cursor changes or a minimal overlay showing recording state and audio levels
3. Press the hotkey again (or a stop key) to end recording and trigger transcription
4. Have the transcribed text auto-pasted at the current cursor position — no manual Ctrl+V needed
5. Pause or cancel a recording mid-stream

This must work on both **Ubuntu (GNOME/Wayland+X11)** and **macOS**, which are the user's two daily-driver platforms.

## Why it matters

The user already has vox integrated into daily workflows on both laptops — it replaced a set of bash/Python scripts (recordmemo.sh, instant_memo.sh, transcribe.py). The existing hotkey setup (Super+Shift+M for instant clipboard, Super+Shift+R for queue) proves the value, but the current approach still requires a terminal to appear, provides no mic-level feedback outside the terminal, and requires a manual paste step. These friction points matter because the tool is used dozens of times daily across contexts — chat apps, docs, code editors, presentations.

Separately, vox serves a second role as a headless audio file processor (used by Relay and other server-side projects via `vox file`). This headless pathway must remain clean and unaffected by the UI work.

(Context from cyborg: user already has GNOME hotkey bindings configured for the predecessor scripts, and has established the Super+Shift+{R,M} keybinding pattern.)

## Domain constraints

- **Platforms**: Ubuntu (GNOME, likely Wayland primary with X11 fallback) and macOS (Sonoma+)
- **Language**: Go (matching vox core). Platform-specific shims in Swift/ObjC (macOS) or C/Python (Linux) are acceptable if necessary.
- **Architecture**: Thin UI shell over existing Go core. `vox` binary and `vox file` must continue working standalone.
- **Distribution**: User installs `vox` CLI first (via go install / GoReleaser). UI overlay is an optional add-on — could be a separate binary (`vox-ui` or `voxd`) or a subcommand (`vox daemon`).
- **Development**: User's development environment is headless (remote server). Design must address whether this can be built by an agent on the server or requires human-in-the-loop development on a machine with display + audio.

## Key questions this design should answer

1. **Single codebase or two?** Can one Go binary provide global hotkeys, cursor feedback, and auto-paste on both macOS and Ubuntu, or do we need platform-specific shells?
2. **Cursor vs overlay?** Is changing the system cursor technically feasible for showing recording state + audio levels, or is a small floating overlay (like a pill or dot) more practical?
3. **Auto-paste mechanism?** What are the reliable ways to simulate keyboard input (Ctrl+V or direct text insertion) on each platform, and what are the gotchas (permissions, accessibility APIs, Wayland restrictions)?
4. **Daemon architecture?** Should the UI run as a persistent daemon/system tray app that listens for hotkeys, or as a one-shot process launched by the OS hotkey system?
5. **Agent-buildable?** What parts of this can be built and tested on a headless server, and what requires a human with a display and microphone?
6. **Installation UX?** How does the user go from having `vox` CLI installed to having the overlay working — what's the minimal setup?

## Success criteria for the design document

- Clear architectural recommendation with platform-specific details for both Ubuntu and macOS
- Concrete technology choices for hotkey registration, visual feedback, and auto-paste on each platform
- Honest assessment of what can be built by an agent on a headless server vs. what needs human testing
- A build-order roadmap that lets the user (or an agent) start implementing today
- Maintains clean separation between vox core (CLI + headless file processing) and the new UI layer
