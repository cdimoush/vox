# Vision: vox

> Speak into your terminal. Get text on your clipboard. Remember everything.

## The Problem

Capturing a voice memo today means: open terminal, run a script, wait, hope it landed on your clipboard. No feedback while recording. No way to get back something you said 10 minutes ago without digging through files. The transcription pipeline works but it's invisible and stateless.

## What vox Is

A Go CLI tool. You type `vox`, speak, hit Enter, and your words are on your clipboard. That's the core loop. It also keeps a history of everything you've ever said so you can search and re-copy old transcriptions.

Think of it like `pbcopy` for your voice — with memory.

```
$ vox
● Recording... (Enter to stop)
██▓░░██▓██░░▓██

⠋ Transcribing...

"Move the contact sensor config into a YAML file
 instead of hardcoding the joint names"

✓ Copied to clipboard
$
```

You're back at your prompt. Paste wherever.

## What vox Is Not

- Not a GUI overlay or floating panel
- Not a TUI app (no Bubbletea, no persistent terminal session)
- Not integrated with the Whisper `/act` pipeline, brain, or beads (v1)
- Not a local transcription engine — cloud only, always

## Decisions (All Resolved)

| Decision | Answer |
|---|---|
| **Transcription engine** | OpenAI Whisper API (`gpt-4o-mini-transcribe`). Cloud only. No local fallback. No offline mode. |
| **Platform** | macOS and Linux. Go cross-compiles natively. |
| **Technology** | Go. Single static binary. No Python dependency. |
| **History persistence** | Persistent across sessions. JSONL file at `~/.vox/history.jsonl`. |
| **History scope** | Unlimited, append-only. User can `vox clear` to reset. |
| **Pipeline integration** | None in v1. vox is clipboard-oriented. Pipeline integration is a future concern. |
| **Packaging** | `go install`, Homebrew tap, and pre-built binaries via GoReleaser. |
| **Audio capture** | SoX (`rec` command). Already a dependency in the current Whisper setup. |
| **Clipboard** | `pbcopy` on macOS, `xclip`/`xsel` on Linux. Detected at runtime. |
| **Configuration** | `OPENAI_API_KEY` env var. No config file for v1. |

## Commands

### `vox` — Record and transcribe

Start recording immediately. Show volume bars in terminal so user knows mic is hot. Press Enter (or Ctrl+C) to stop. Send audio to OpenAI Whisper API. Print transcription. Copy to clipboard. Append to history. Exit.

```
$ vox
● Recording... (Enter to stop)
██▓░░██▓██░░▓██

⠋ Transcribing...

"Move the contact sensor config into YAML"

✓ Copied to clipboard
```

### `vox ls` — Show history

Print recent transcriptions, numbered, most recent first. Show relative timestamp and truncated text. Default: last 20 entries.

```
$ vox ls
#   When        Text
1   2m ago      Move the contact sensor config into YAML...
2   14m ago     Remind Nick about the gantry collision boundary...
3   1h ago      Need to add error handling for USD stage loading...
```

Flags:
- `vox ls -n 50` — show last 50
- `vox ls --all` — show everything

### `vox cp <n>` — Re-copy a history entry

Copy history item #n back to the clipboard.

```
$ vox cp 3
✓ Copied #3 to clipboard
```

### `vox show <n>` — Show full text of a history entry

Print the complete transcription text for entry #n (not truncated).

```
$ vox show 1
[2m ago]

Move the contact sensor config into a YAML file instead of hardcoding
the joint names. The current approach in design_lab/sensors/contact.py
has a list of 14 joint names that breaks every time someone adds a new
robot model.
```

### `vox clear` — Clear history

Delete all history entries. Asks for confirmation.

```
$ vox clear
Delete all 47 transcriptions? [y/N] y
✓ History cleared
```

## Architecture

```
┌──────────────────────────────────────────┐
│                vox binary                │
│                                          │
│  ┌──────────┐  ┌────────┐  ┌─────────┐  │
│  │ recorder  │→│  api    │→│ output   │  │
│  │ (sox)     │  │(openai) │  │(clip+fs) │  │
│  └──────────┘  └────────┘  └─────────┘  │
│                                          │
│  ┌──────────────────────────────────────┐│
│  │ history (read/write ~/.vox/)         ││
│  └──────────────────────────────────────┘│
└──────────────────────────────────────────┘
```

### Internal Packages

- **`cmd/vox/`** — CLI entrypoint. Parses commands, dispatches to packages.
- **`recorder/`** — Wraps SoX `rec` command. Starts/stops recording to a temp WAV file. Reads audio levels from SoX stderr for volume display.
- **`transcribe/`** — Sends audio file to OpenAI Whisper API. Returns text. Handles chunking for files >25MB (unlikely for quick memos but defensive).
- **`clipboard/`** — Detects platform (`pbcopy` vs `xclip`). Writes text to system clipboard.
- **`history/`** — Append-only JSONL read/write. Each entry: timestamp, transcription text, duration.

### History Format

`~/.vox/history.jsonl`:

```jsonl
{"ts":"2026-02-28T14:30:00Z","text":"Move the contact sensor config into YAML...","duration_s":12.4}
{"ts":"2026-02-28T14:16:00Z","text":"Remind Nick about the gantry collision...","duration_s":8.1}
```

No audio files are persisted. History is text-only.

## Runtime Dependencies

| Dependency | Required | Notes |
|---|---|---|
| **SoX** | Yes | `brew install sox` / `apt install sox` — audio recording |
| **OpenAI API key** | Yes | `OPENAI_API_KEY` env var |
| **pbcopy** (macOS) or **xclip** (Linux) | Yes | Clipboard access. pbcopy is built-in on macOS. |

No ffmpeg. No Python. No other runtime dependencies.

## Build & Distribution

- **Source**: `go build -o vox ./cmd/vox`
- **Install**: `go install github.com/<owner>/vox@latest`
- **Homebrew**: `brew tap <owner>/vox && brew install vox`
- **Binaries**: GoReleaser builds for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- **Uninstall**: `rm $(which vox) && rm -rf ~/.vox/`

## User Workflow

### Quick dictation (the 90% case)

```bash
# You're in the middle of coding. You have a thought.
$ vox
● Recording...
"Refactor the sensor config to use YAML"
✓ Copied to clipboard

# Paste into a commit message, Slack, doc, wherever
```

### Retrieve something you said earlier

```bash
$ vox ls
#   When        Text
1   2m ago      Refactor the sensor config to use YAML...
2   1h ago      Tell Nick about the gantry collision boundary issue...
3   3h ago      Research USD stage loading error patterns...

$ vox cp 2
✓ Copied #2 to clipboard
```

### Shell alias for even less friction

```bash
# .zshrc
alias v="vox"
alias vl="vox ls"
alias vc="vox cp"
```

Now it's: `v` → speak → Enter → paste. Four keystrokes to capture a thought.

## What v1 Intentionally Skips

These are all valid future work, explicitly deferred:

- **No TUI mode** — no Bubbletea, no persistent UI. Just a CLI that runs and exits.
- **No beads integration** — transcriptions don't create issues.
- **No `/act` pipeline** — no brain context, no deliverable generation.
- **No search** — `vox ls` is chronological only. Full-text search is future.
- **No config file** — API key via env var only.
- **No audio persistence** — temp WAV is deleted after transcription. Text-only history.
- **No local transcription** — cloud only, always. This is a settled decision, not a deferral.

## Future Directions (Not v1)

Rough priority order for post-v1:

1. **`vox search <query>`** — full-text search across history
2. **`vox act`** — send a recording through the full Whisper `/act` pipeline
3. **`vox tag <n> <tag>`** — tag history entries for categorization
4. **Beads integration** — optionally create a bead from a transcription
5. **Config file** — model selection, default flags, custom SoX args
6. **TUI mode** — Bubbletea interactive history browser (if the CLI feels limiting)

---

*Tight vision. All decisions resolved. Ready for epic breakdown.*
