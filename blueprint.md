# Blueprint: vox

> Build a Go CLI that records voice, transcribes via OpenAI Whisper, copies to clipboard, and maintains searchable history.

**Vision**: [vox](vision.md)

## Overview

vox replaces the existing bash/Python voice memo scripts (`instant_memo.sh`, `record_memo.sh`, `transcribe.py`) with a single static Go binary. The core loop is: type `vox`, speak, press Enter, get text on your clipboard. History is persisted as append-only JSONL so users can retrieve past transcriptions.

The tool is intentionally minimal. No TUI framework, no config file, no local transcription. It shells out to SoX for recording, calls the OpenAI Whisper API for transcription, and uses platform-native clipboard tools (`pbcopy`/`xclip`/`xsel`). The entire tool is five internal packages plus a CLI entrypoint.

The build is split into three phases: foundation packages with no external dependencies, the core recording and transcription pipeline, and finally the subcommands and distribution packaging.

## Existing State

- Empty Go project (no `.go` files, no `go.mod`)
- `vision.md` with complete spec and all decisions resolved
- `CLAUDE.md` with development constraints
- `reference/` directory with Python transcription script, bash recording scripts, and ecosystem links
- No beads tasks exist

---

## Phase 1: Foundation

**Goal**: Go module initialized, `history` and `clipboard` packages implemented with tests, buildable but not yet functional as a CLI.

**Success Criteria**: `go test ./...` passes. History entries can be appended, read, and cleared. Clipboard write works on the current platform.

### Dependency Graph

```
WI-1.1 ──→ WI-1.2
WI-1.1 ──→ WI-1.3
```

### WI-1.1: Initialize Go module and project skeleton

```yaml
type: chore
complexity: S
depends_on: []
parallel_with: []
```

**Description**: Create the Go module, directory structure, and a minimal `main.go` that compiles and prints a version string. This establishes the build and gives all other work items a compilable base.

**Files to Create/Modify**:
- `go.mod` — (create) Go module definition (`github.com/conner/vox` or similar)
- `cmd/vox/main.go` — (create) minimal entrypoint that prints `vox v0.1.0-dev`
- `recorder/` — (create) empty directory with placeholder doc.go
- `transcribe/` — (create) empty directory with placeholder doc.go
- `clipboard/` — (create) empty directory with placeholder doc.go
- `history/` — (create) empty directory with placeholder doc.go

**Acceptance Criteria**:
- [ ] `go build ./cmd/vox` produces a `vox` binary
- [ ] `./vox` prints a version string and exits 0
- [ ] `go vet ./...` passes with no warnings
- [ ] Directory structure matches vision: `cmd/vox/`, `recorder/`, `transcribe/`, `clipboard/`, `history/`

### WI-1.2: Implement history package

```yaml
type: feature
complexity: M
depends_on: [WI-1.1]
parallel_with: [WI-1.3]
```

**Description**: Build the `history` package for append-only JSONL storage at `~/.vox/history.jsonl`. Each entry has three fields: `ts` (RFC3339 timestamp), `text` (transcription), and `duration_s` (float64 recording duration in seconds). The package must handle creating the `~/.vox/` directory if it does not exist, appending entries atomically, reading all entries, reading the last N entries, and clearing the file.

Reference the JSONL format from the vision:
```jsonl
{"ts":"2026-02-28T14:30:00Z","text":"...","duration_s":12.4}
```

**Files to Create/Modify**:
- `history/history.go` — (create) `Entry` struct, `Store` type with `Append`, `List`, `Clear` methods, `DefaultPath()` function
- `history/history_test.go` — (create) tests using a temp directory

**Acceptance Criteria**:
- [ ] `Append` writes a valid JSONL line and creates `~/.vox/` if missing
- [ ] `List(n)` returns the last N entries in reverse-chronological order; `List(0)` returns all
- [ ] `Clear` removes the history file
- [ ] All tests pass with `go test ./history/`
- [ ] Entries round-trip correctly through write and read (no data loss)

### WI-1.3: Implement clipboard package

```yaml
type: feature
complexity: S
depends_on: [WI-1.1]
parallel_with: [WI-1.2]
```

**Description**: Build the `clipboard` package that detects the platform clipboard command at runtime and writes text to it. Detection order: `pbcopy` (macOS) > `xsel --clipboard --input` (Linux, preferred) > `xclip -selection clipboard` (Linux, fallback). If none found, return a clear error. Follow the same detection logic as `instant_memo_reference.sh`.

**Files to Create/Modify**:
- `clipboard/clipboard.go` — (create) `Write(text string) error` function, `Detect() (string, error)` for finding clipboard binary
- `clipboard/clipboard_test.go` — (create) test that `Detect` finds a clipboard tool on the current platform; test `Write` with a known string

**Acceptance Criteria**:
- [ ] `Detect()` returns `"pbcopy"` on macOS, `"xsel"` or `"xclip"` on Linux
- [ ] `Detect()` returns a descriptive error if no clipboard tool is found
- [ ] `Write(text)` pipes text to the detected clipboard command via stdin
- [ ] All tests pass with `go test ./clipboard/`

---

## Phase 2: Core Pipeline

**Goal**: The `vox` command (no subcommand) records audio via SoX, transcribes via OpenAI Whisper API, copies to clipboard, and saves to history. The main use case works end-to-end.

**Success Criteria**: Running `vox` with a microphone and valid `OPENAI_API_KEY` records, transcribes, copies text to clipboard, and appends to `~/.vox/history.jsonl`.

### Dependency Graph

```
WI-2.1 ──→ WI-2.3
WI-2.2 ──→ WI-2.3
```

### WI-2.1: Implement recorder package

```yaml
type: feature
complexity: M
depends_on: [WI-1.1]
parallel_with: [WI-2.2]
```

**Description**: Build the `recorder` package that wraps the SoX `rec` command. It should start recording to a temporary WAV file using voice-optimized settings (`-r 16000 -c 1 -b 16`), display volume feedback by reading SoX's stderr output, and stop recording when the user presses Enter or Ctrl+C. The temp file path is returned for the transcription step. Follow the SoX flags from `record_memo_reference.sh`.

Key behaviors:
- Shell out to `rec` (part of SoX) with: `rec -r 16000 -c 1 -b 16 <tempfile.wav>`
- Read SoX stderr to extract volume levels and render a simple volume bar to the terminal
- Stop recording on SIGINT or when stdin receives a newline
- Return the temp file path and recording duration
- Check that `rec` is available before starting; return a clear error with install instructions if not

**Files to Create/Modify**:
- `recorder/recorder.go` — (create) `Record(ctx context.Context) (result Result, err error)` where `Result` has `FilePath` and `Duration`
- `recorder/volume.go` — (create) parse SoX stderr volume levels, render volume bar string
- `recorder/recorder_test.go` — (create) test SoX detection, test volume parsing logic

**Acceptance Criteria**:
- [ ] `Record` starts SoX, creates a temp WAV file, and returns its path
- [ ] Recording stops cleanly on context cancellation (Enter or Ctrl+C)
- [ ] Volume bar parsing extracts levels from SoX stderr format
- [ ] Returns a clear error with install instructions if `rec` is not found
- [ ] All unit tests pass with `go test ./recorder/`

### WI-2.2: Implement transcribe package

```yaml
type: feature
complexity: M
depends_on: [WI-1.1]
parallel_with: [WI-2.1]
```

**Description**: Build the `transcribe` package that sends a WAV file to the OpenAI Whisper API and returns the transcribed text. Use the `github.com/sashabaranov/go-openai` client library. Default model is `gpt-4o-mini-transcribe`. Read `OPENAI_API_KEY` from the environment. Follow the API call pattern from `transcribe_reference.py`.

Key behaviors:
- Open the audio file, send to `client.CreateTranscription` with model `gpt-4o-mini-transcribe`
- Return the transcribed text
- Return clear errors for: missing API key, file not found, API errors
- No chunking needed for v1 (quick memos are always under 25MB)

**Files to Create/Modify**:
- `transcribe/transcribe.go` — (create) `Transcribe(filePath string) (string, error)` function
- `transcribe/transcribe_test.go` — (create) test error cases (missing API key, missing file); integration test behind build tag
- `go.mod` — (modify) add `github.com/sashabaranov/go-openai` dependency
- `go.sum` — (auto-generated)

**Acceptance Criteria**:
- [ ] `Transcribe` sends audio to OpenAI and returns text
- [ ] Returns `ErrNoAPIKey` when `OPENAI_API_KEY` is not set
- [ ] Returns a clear error when the audio file does not exist
- [ ] Unit tests pass without an API key (error-path tests only)
- [ ] All tests pass with `go test ./transcribe/`

### WI-2.3: Wire up the main `vox` command

```yaml
type: feature
complexity: M
depends_on: [WI-2.1, WI-2.2]
parallel_with: []
```

**Description**: Connect all packages in `cmd/vox/main.go` to implement the core `vox` command (no subcommand). The flow is: check dependencies (SoX, clipboard, API key) -> print "Recording..." -> record via `recorder` -> print "Transcribing..." -> transcribe via `transcribe` -> print quoted text -> copy to clipboard -> append to history -> clean up temp file -> exit.

Handle signals gracefully: Ctrl+C during recording should stop recording and proceed to transcription (not abort). Ctrl+C during transcription should abort cleanly.

Match the UX from the vision:
```
$ vox
Recording... (Enter to stop)
[volume bars]
Transcribing...
"Move the contact sensor config into YAML"
Copied to clipboard
```

**Files to Create/Modify**:
- `cmd/vox/main.go` — (modify) implement full record-transcribe-copy-save flow
- `cmd/vox/run.go` — (create) extract the core `vox` run logic into a separate function for testability

**Acceptance Criteria**:
- [ ] Running `vox` with no args records, transcribes, copies to clipboard, and saves to history
- [ ] Ctrl+C during recording stops recording and proceeds to transcription
- [ ] Missing `OPENAI_API_KEY` prints a clear error and exits non-zero
- [ ] Missing `rec` (SoX) prints install instructions and exits non-zero
- [ ] Transcribed text is printed in quotes before the "Copied to clipboard" message
- [ ] UX matches vision: `●` recording indicator, spinner during transcription, `✓` on copy
- [ ] Temp WAV file is cleaned up after transcription (success or failure)

---

## Phase 3: CLI Commands & Polish

**Goal**: All five commands work (`vox`, `vox ls`, `vox cp`, `vox show`, `vox clear`). GoReleaser configured for distribution. README written.

**Success Criteria**: All commands match the UX described in the vision. `goreleaser check` passes. `go build` and `go test ./...` pass.

### Dependency Graph

```
WI-3.1 ──→ WI-3.2
            WI-3.2 ──→ WI-3.4
WI-3.3 (parallel with all)
```

### WI-3.1: Implement command routing and `vox ls`

```yaml
type: feature
complexity: M
depends_on: [WI-2.3]
parallel_with: [WI-3.3]
```

**Description**: Add argument parsing to `main.go` to route subcommands: no args runs the record flow, `ls` shows history, `cp` re-copies, `show` displays full text, `clear` wipes history. Use the standard library `os.Args` — no CLI framework needed for five commands.

Implement `vox ls` fully:
- Print recent entries in a table: `#  When  Text`
- Default to last 20 entries
- `-n <count>` flag to change the limit
- `--all` flag to show everything
- Relative timestamps: "2m ago", "1h ago", "3d ago"
- Truncate text to fit terminal width (or ~60 chars)
- Most recent first (entry #1 is the newest)

**Files to Create/Modify**:
- `cmd/vox/main.go` — (modify) add command routing switch
- `cmd/vox/ls.go` — (create) `vox ls` implementation with flag parsing
- `cmd/vox/format.go` — (create) relative time formatting, text truncation helpers
- `cmd/vox/format_test.go` — (create) tests for relative time and truncation

**Acceptance Criteria**:
- [ ] `vox ls` prints the last 20 history entries in a formatted table
- [ ] `vox ls -n 5` shows only the last 5 entries
- [ ] `vox ls --all` shows all entries
- [ ] Timestamps display as relative ("2m ago", "1h ago", "3d ago")
- [ ] Unknown subcommands print a usage message and exit non-zero

### WI-3.2: Implement `vox cp`, `vox show`, and `vox clear`

```yaml
type: feature
complexity: S
depends_on: [WI-3.1]
parallel_with: [WI-3.3]
```

**Description**: Implement the remaining three subcommands. These are all simple operations on top of the existing `history` and `clipboard` packages.

`vox cp <n>`: Look up history entry #n (1-indexed, 1 = most recent), copy its text to clipboard, print confirmation.

`vox show <n>`: Look up entry #n, print the relative timestamp and full untruncated text.

`vox clear`: Prompt "Delete all N transcriptions? [y/N]", read stdin, clear history if confirmed.

**Files to Create/Modify**:
- `cmd/vox/cp.go` — (create) `vox cp` implementation
- `cmd/vox/show.go` — (create) `vox show` implementation
- `cmd/vox/clear.go` — (create) `vox clear` implementation

**Acceptance Criteria**:
- [ ] `vox cp 1` copies the most recent entry to clipboard and prints confirmation
- [ ] `vox show 1` prints the full text of the most recent entry with timestamp
- [ ] `vox clear` prompts for confirmation and clears history on "y"
- [ ] All commands print useful errors for invalid input (missing arg, out-of-range index, empty history)

### WI-3.3: GoReleaser configuration

```yaml
type: chore
complexity: S
depends_on: [WI-1.1]
parallel_with: [WI-3.1, WI-3.2]
```

**Description**: Add GoReleaser configuration for cross-platform binary builds and Homebrew tap. Target: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`. No Windows. Reference the Perles project (linked in `reference/links.md`) for GoReleaser + Homebrew tap patterns.

**Files to Create/Modify**:
- `.goreleaser.yaml` — (create) build config for 4 platform/arch targets, archive settings, changelog generation
- `.github/workflows/release.yml` — (create) GitHub Actions workflow that runs GoReleaser on tag push

**Acceptance Criteria**:
- [ ] `goreleaser check` passes (valid config)
- [ ] Config builds for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- [ ] No Windows targets included
- [ ] Release workflow triggers on `v*` tag push

### WI-3.4: README and final polish

```yaml
type: chore
complexity: S
depends_on: [WI-3.2]
parallel_with: []
```

**Description**: Write a README.md covering installation (go install, Homebrew, binary download), usage for all five commands, runtime dependencies (SoX, API key, clipboard tool), and shell alias suggestions. Add a `.gitignore` for Go build artifacts. Do a final pass: ensure `go vet ./...` and `go test ./...` pass clean.

**Files to Create/Modify**:
- `README.md` — (create) installation, usage, dependencies, aliases
- `.gitignore` — (create) Go binary, vendor/, .env, ~/.vox/ references

**Acceptance Criteria**:
- [ ] README covers install, all 5 commands, dependencies, and aliases
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `.gitignore` covers build artifacts

---

## Summary

| ID | Title | Type | Complexity | Phase | Depends On |
|----|-------|------|-----------|-------|------------|
| WI-1.1 | Initialize Go module and project skeleton | chore | S | 1 | — |
| WI-1.2 | Implement history package | feature | M | 1 | WI-1.1 |
| WI-1.3 | Implement clipboard package | feature | S | 1 | WI-1.1 |
| WI-2.1 | Implement recorder package | feature | M | 2 | WI-1.1 |
| WI-2.2 | Implement transcribe package | feature | M | 2 | WI-1.1 |
| WI-2.3 | Wire up the main `vox` command | feature | M | 2 | WI-2.1, WI-2.2 |
| WI-3.1 | Implement command routing and `vox ls` | feature | M | 3 | WI-2.3 |
| WI-3.2 | Implement `vox cp`, `vox show`, `vox clear` | feature | S | 3 | WI-3.1 |
| WI-3.3 | GoReleaser configuration | chore | S | 3 | WI-1.1 |
| WI-3.4 | README and final polish | chore | S | 3 | WI-3.2 |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| SoX stderr volume format varies across versions | Medium | Low | Parse best-effort; degrade to no volume bars if format unrecognized |
| `go-openai` library API changes or lacks transcription support | Low | High | Library is mature and widely used; pin version in go.mod |
| xclip clipboard content lost when terminal closes | Medium | Low | Prefer xsel over xclip in detection order (already planned) |
| Long recordings exceed 25MB API limit | Low | Low | Quick memos are typically under 1 minute; defer chunking to post-v1 |

## Notes

- Phase 1 work items (WI-1.2 and WI-1.3) can be done in parallel by separate agents after WI-1.1 is complete.
- Phase 2 work items (WI-2.1 and WI-2.2) can also be parallelized — they share no code dependencies.
- WI-3.3 (GoReleaser) only needs WI-1.1 and can be done any time after the module exists.
- No CLI framework (Cobra, urfave/cli) is used. Five commands do not justify the dependency. Standard `os.Args` routing is sufficient.
- The `transcribe` package integration test requires a real API key and should be gated behind a build tag like `//go:build integration`.
