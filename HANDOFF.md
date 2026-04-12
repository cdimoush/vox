# vox UI daemon — agent → human handoff

This PR lands **Phase 1** of the vox UI dictation overlay described in
[`system-designer/designs/2026-04-12/vox-ui/system-design.md`](../system_designer/designs/2026-04-12/vox-ui/system-design.md).

It is **intentionally incomplete.** The agent built everything that can
be exercised on a headless server. The remaining work requires a human
with a display and a microphone. Do not merge until you finish Phase 2+.

---

## What landed (Phase 1 — agent-buildable)

End-to-end `vox ui start → toggle → stop` plumbing, minus the display/audio
bits. Run `go build -o vox-dev ./cmd/vox` and try:

```bash
./vox-dev ui start     # forks a detached daemon, writes pid to ~/.vox/daemon.pid
./vox-dev ui status    # pid NNNN, state=idle
./vox-dev ui toggle    # flips state to recording (will actually try to record!)
./vox-dev ui stop      # clean shutdown via IPC
```

| Area | File(s) | Notes |
|---|---|---|
| TOML config loader | `daemon/config/` | `~/.vox/config.toml`, defaults from design §8, validates paste-method enum |
| State machine | `daemon/daemon.go` | `idle → recording → transcribing → done → idle`, event channel, cancel-anywhere |
| Unix-socket IPC | `daemon/ipc/` | JSON line protocol; commands: toggle, cancel, status, shutdown |
| Paste dispatcher | `daemon/paste/` | Per-platform command selection (xdotool / wtype / ydotool / osascript); clipboard-only fallback; best-effort semantics |
| CLI subcommands | `cmd/vox/ui.go` | `vox ui {start,stop,toggle,status}`; forks self as `vox __daemon` |
| Daemon entrypoint | `cmd/vox/daemon.go` | Loads config → builds adapters → starts IPC → blocks on signal/shutdown; appends successful transcriptions to the existing `history` store |

All of it is unit-tested on a headless machine (`go test ./...` passes
with no CGO, no display, no mic).

**The headless pathways — `vox` and `vox file` — are untouched.** Run
them exactly as before; they do not pull in any of the new packages.

---

## What's NOT in this PR (Phase 2+ — human must do)

Design §Developability called these out as "requires a display + mic."
The interfaces are in place; you just need to plug real implementations
in behind a `ui` build tag so headless builds stay clean.

### 1. Global hotkey registration (design §3)
- **macOS / Linux X11**: wire `golang.design/x/hotkey` to call
  `daemon.Toggle()` directly (don't go through the socket).
- **Linux Wayland**: print the compositor keybinding instructions from
  design §3 on first run. No hotkey code needed — Wayland already calls
  `vox ui toggle` via the user's compositor config.

### 2. Gio overlay window (design §4)
- Pill-shaped floating window, 200×40.
- Subscribe to `daemon.Events()` to drive state-dependent visuals.
- Audio-level bars need a mic tap — see item 5.
- Off by default on Wayland (per config default; flip on after testing
  your compositor doesn't focus-steal).

### 3. System tray (design §5)
- `fyne.io/systray` (or tailscale fork) with `RunWithExternalLoop` so it
  co-exists with Gio's main-thread requirements.
- Icon colors map to `daemon.State`: gray/red/blue/green.
- Menu: Toggle / History / Settings / Quit.

### 4. Audio cues (design §6)
- `gen2brain/malgo` output device; play 200ms start/300ms done/200ms
  error tones.
- Gated by `config.Audio.CuesEnabled`.

### 5. Mic-level streaming (design §4 data flow)
- Open the mic with malgo concurrently with SoX — design §Risks flags
  this as needing a real test on PipeWire / CoreAudio. Validate on both
  your laptops.
- Stream RMS levels into the Gio overlay at ~30 FPS.

### 6. Paste injection — real keystroke path (design §7)
- What we shipped uses `osascript` on macOS. Swap to `CGEventPost` for a
  faster, no-AppleScript-permission path; add an `AXIsProcessTrusted()`
  probe at daemon startup so we can show a one-time Accessibility prompt.
- Linux: validate xdotool / wtype / ydotool actually paste into the
  apps you care about (Slack, VS Code, Chrome) on your specific
  GNOME + Wayland config.

### 7. Install helpers (design §Roadmap Phase 4)
- `vox ui install` → write a user systemd unit (Linux) or launchd plist
  (macOS) so the daemon autostarts on login.

---

## How to continue

1. Check out this branch: `git checkout feature/vox-ui-daemon`.
2. Read the design doc linked at the top of this file (it is the source
   of truth; this handoff is a map *into* it).
3. Move Phase 2+ work into beads as child tasks of **vox-6c5.1**
   (the blueprint epic). One bead per item above is reasonable.
4. Gate new display/audio packages behind a `//go:build ui` tag so
   `go build ./...` without `-tags=ui` continues to produce a binary
   that works in a headless CI.
5. When you replace the pasteAdapter / add the overlay / etc., prefer
   editing `cmd/vox/daemon.go`'s adapter struct fields over touching
   the `daemon/` packages — the state machine is a stable seam.

## Known rough edges

- `vox ui toggle` currently drives a *real* record → transcribe → paste
  pipeline if you press it. That is by design (the agent proved the
  plumbing works), but the first time you try it on your laptop you
  should have a terminal open to `tail -f ~/.vox/daemon.log` because
  there is no visual feedback yet.
- The daemon does not yet honor `SIGHUP` for config reload (design §8
  promises this). Add it when you add the tray's Settings menu item.
- `vox ui start` prints the daemon's pid to stderr but does not yet
  check whether the user has SoX + a working API key *inside the
  daemon*. It will fail inside `runSession` instead. Consider a startup
  preflight in Phase 2.

---

## Beads trail

- Concept → trade study → blueprint: `vox-6c5`, `vox-6c5.1`
- Variants considered: `vox-6c5.1`, `vox-6c5.2`, `vox-6c5.3`
- Phase-1 tasks (all closed): `vox-6c5.1.1` through `vox-6c5.1.7`

Run `bd show vox-6c5` for the full lineage.
