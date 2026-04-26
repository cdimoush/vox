---
name: vox-up
description: Rebuild the vox ui daemon with -tags=ui and (re)start the user systemd service. Use when the user wants to start dictation, pick up yesterday's setup after a reboot, or reload after code changes.
triggers:
  - vox up
  - start vox
  - bring up vox
  - restart vox
  - reload vox
  - good morning vox
allowed-tools: Bash, Read
---

# Bring up the vox ui daemon

This is a **development-only** convenience skill for the `feature/vox-ui-daemon`
branch. Once the UI daemon ships, this skill should be removed.

## Preconditions (ambient facts about this machine)

- Go toolchain lives at `/home/conner/.local/go/bin/go` and is **not** on the
  default PATH — always prepend it.
- `vox` binary is installed to `/home/conner/bin/vox`. The systemd unit
  (`~/.config/systemd/user/vox.service`) embeds this path, so do not relocate.
- User systemd unit was created by `vox ui install`. If it's missing, run
  `~/bin/vox ui install` once.
- X11 session (DISPLAY=:1). Hotkey registration uses X11 grabs — will not work
  under Wayland.

## The flow

Run these as a single chained command so the user gets one tidy confirmation:

```bash
PATH=/home/conner/.local/go/bin:$PATH go build -tags=ui -o ~/bin/vox ./cmd/vox \
  && systemctl --user restart vox.service \
  && sleep 1 \
  && systemctl --user is-active vox.service \
  && journalctl --user -u vox.service --no-pager -n 10
```

What to look for in the journal tail:
- `hotkey registered: Ctrl+Shift+Space` — X11 grab succeeded
- `tray: started` — system tray is up
- `overlay: window created` — Gio overlay window exists
- `level monitor: capturing mic levels` — mic is open

Then report to the user:
- Daemon PID
- Hotkey combo (pulled from the journal line or `~/.vox/config.toml`)
- Anything that logged at level WARN/ERROR since the last restart

## When things fail

- **`go: command not found`** — the `PATH=` prefix was dropped. Re-run with it.
- **Build fails with cgo / libX11 errors** — stop, report, ask the user. Do not
  try to auto-install system packages.
- **`systemctl restart` hangs** — the daemon may be wedged holding the X11
  display. `systemctl --user kill -s SIGKILL vox.service` then restart.
- **No `hotkey registered` line** — another app grabbed the combo first. Ask
  the user to pick a different key in `~/.vox/config.toml` `[hotkey]`.
- **First-time install (no unit file yet)** — run `~/bin/vox ui install` after
  the build, before `systemctl start`.

## What NOT to do

- Don't run `vox ui start` — that's the manual pidfile path, and it fights
  the systemd unit for the socket.
- Don't `go install` — it won't pass the `-tags=ui` flag and the daemon will
  silently start without hotkey/tray/overlay.
- Don't rebuild the `vox-dev` binary in the repo root; the systemd unit points
  at `~/bin/vox`.
