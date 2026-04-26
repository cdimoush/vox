# vox-core — specification

> The durable artifact. Whatever language vox-core is in, it must satisfy this spec and pass the tests in `tests.md`.
> Scope: vox-core only (CLI, transcribe, history, clipboard, config, recorder). UI/daemon is out of scope and being deleted.

## What vox-core is

- A CLI that turns audio into text via the OpenAI Whisper API and copies the result to the clipboard
- A pipeline tool that other projects shell out to (stdout text, optional `--json`)
- A history keeper (append-only JSONL at `~/.vox/history.jsonl`)
- One static binary, no display server, no daemon, no hotkey

## CLI surface

- `vox` — record from mic via SoX, Enter or Ctrl+C to stop, transcribe, write text to clipboard, append to history. stderr = chrome, stdout = nothing
- `vox file <path>` — transcribe an existing audio file. stdout = transcript text. stderr = spinner + status. Also writes to clipboard + appends history
- `vox file <path> --json` — same, but stdout = `{text, duration_s, chunks, error?}` and stderr is silent (no spinner)
- `vox file -` — read audio from stdin into a temp file, then transcribe. `--format=ogg` (default) sets the temp file extension
- `vox ls` — list history, most-recent first, default last 20. `-n N` limit, `--all` no limit. stdout = table
- `vox cp <n>` — re-copy history entry `n` (1-indexed against the `vox ls` ordering) to clipboard
- `vox show <n>` — print full text of history entry `n` to stdout
- `vox clear` — confirm-then-delete `~/.vox/history.jsonl`
- `vox login` — interactive prompt → write `OPENAI_API_KEY=…` to `~/.vox/config`
- `vox --version` / `vox -v` — print version, exit 0

## JSON contracts (frozen forever)

- `vox file --json` success: `{"text": string, "duration_s": number, "chunks": int}`
- `vox file --json` error: `{"text": "", "duration_s": 0, "chunks": 0, "error": string}` — exit code still set per error class
- history line: `{"ts": rfc3339, "text": string, "duration_s": number}` — one line per entry, `\n`-terminated, no trailing comma
- File ordering inside history: append-only, oldest first. `vox ls` reverses for display

## Config / API key discovery

- `config.FindAPIKey()` checks in order, returns the first non-empty match:
  1. `$OPENAI_API_KEY` env var
  2. `OPENAI_API_KEY=…` line in `~/.vox/config` (key=value, `#` comments, optional quotes)
  3. `export OPENAI_API_KEY=…` line in `~/.zshrc`, `~/.bashrc`, `~/.bash_profile`, `~/.profile` (in that order)
- `vox login` writes the key to `~/.vox/config` (mode 0600, dir mode 0700) and prints a one-line shell-profile hint to stderr
- Key must start with `sk-`; otherwise reject with exit 1

## Audio pipeline

- Recording: SoX `rec` shelled out at 16kHz, mono, 16-bit. SIGINT to stop (gives SoX time to finalize the WAV header). Output is a temp file the caller deletes
- File transcription: accepted formats `.wav .m4a .mp3 .webm .ogg`. Files >8 minutes are auto-chunked into 5-minute segments and stitched
- All transcription goes through OpenAI Whisper (`gpt-4o-mini-transcribe`). No local model, no other provider in vox-core today
- Provider abstraction is an implementation detail — the spec only cares that `audio in → text out` round-trips

## History

- Path: `~/.vox/history.jsonl`. Append-only. One JSON object per line
- Created lazily (vox creates `~/.vox/` mode 0700 and the file mode 0600 on first append)
- Concurrent appends: relies on POSIX append-mode atomicity for single-line writes. Don't promise more
- `vox clear` removes the file. Missing file is not an error anywhere

## Exit codes (frozen)

- `0` — success
- `1` — generic error (file not found, bad args, bad input format, missing dependency)
- `2` — OpenAI API error (rate limit, timeout, server error)
- `3` — no API key found anywhere

## Stderr / stdout discipline (frozen)

- Data → stdout
- Chrome (spinners, "Copied to clipboard", error prefixes) → stderr
- `--json` mode silences the chrome to stderr-of-its-own (no spinner)

## What's NOT in this spec (deferred)

- Provider abstraction (Deepgram, AssemblyAI) — spec the contract when the code lands
- Daemon / IPC / hotkey / overlay / tray — being deleted in vox-1yh phase 2

## Compatibility guarantees

- **Frozen forever**: `--json` schema, history line schema, exit code semantics, FindAPIKey order, command names
- **Free to change**: stderr text wording, spinner glyphs, internal file paths under `~/.vox/`
