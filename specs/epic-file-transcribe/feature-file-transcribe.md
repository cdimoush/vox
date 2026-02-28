# Feature: File Transcribe Command

## Summary

Add `vox file <path>` command that transcribes an existing audio file using the OpenAI Whisper API, prints the transcription, copies it to clipboard, and saves it to history. This reuses the existing `transcribe`, `clipboard`, and `history` packages with no new dependencies.

## Motivation

Vox only supports live recording today. Users and agents often have pre-recorded audio files (phone memos, meeting clips, pipeline outputs) that need transcription. The `transcribe` package already accepts a file path — this feature exposes that capability as a CLI command.

## Detailed Design

### Command Interface

```
vox file <path>

Arguments:
  path    Path to an audio file (.wav, .m4a, .mp3, .webm)

Examples:
  vox file memo.m4a
  vox file ~/Downloads/meeting-notes.mp3
  vox file /tmp/vox-queue/recording.wav
```

### Behavior

1. Validate that exactly one argument (file path) is provided
2. Check that the file exists and is readable
3. Check that `OPENAI_API_KEY` is set
4. Check that clipboard is available (`clipboard.Detect()`)
5. Show spinner: `⠋ Transcribing...`
6. Call `transcribe.Transcribe(filePath)` (reuse existing package)
7. Stop spinner
8. Print transcription in quotes to stderr: `"transcribed text"`
9. Copy to clipboard via `clipboard.Write()`
10. Print `✓ Copied to clipboard` to stderr
11. Save to history via `history.Store.Append()` with duration_s set to 0 (no recording duration)

### Error Cases

| Condition | Message |
|---|---|
| No path argument | `Usage: vox file <path>` |
| File doesn't exist | `Error: file not found: <path>` |
| OPENAI_API_KEY not set | Same message as `run()`: includes `Set it with: export OPENAI_API_KEY=your-key` |
| Clipboard unavailable | Same error as `run()` |
| API error | `Error: transcription failed: <api error>` |

### Signal Handling

- Ctrl+C during transcription aborts cleanly (same `transcribeWithContext` pattern as `run.go`)
- No two-phase handling needed (there's no recording phase)

## Files to Create/Modify

### Create

- **`cmd/vox/file.go`** — `cmdFile()` function implementing the command
  - Follows the same pattern as `cmd/vox/cp.go` and `cmd/vox/show.go`
  - Uses spinner from `run.go` (duplicate the small spinner loop — it's ~15 lines)
  - Calls `transcribeWithContext()` already defined in `run.go`

- **`cmd/vox/file_test.go`** — Unit tests
  - Test: no arguments → usage error
  - Test: nonexistent file → file not found error
  - Test: argument parsing with valid path structure

### Modify

- **`cmd/vox/main.go`** — Add `case "file"` to the command switch, update usage string to include `file`

## Implementation Notes

- The `transcribeWithContext()` function in `run.go` is already exported at package level (lowercase, same package) — `file.go` can call it directly
- `duration_s` in the history entry should be `0` since we don't know the audio duration without reading the file header, and that's out of scope
- The spinner code is small enough to duplicate rather than extracting a shared helper (avoid over-abstraction per CLAUDE.md)
- No format validation beyond what the Whisper API does — if the API rejects the file, the error propagates naturally

## Testing Strategy

### Unit Tests (no API key needed)
1. `TestCmdFileNoArgs` — calling with no file path returns usage error
2. `TestCmdFileNonexistentFile` — calling with bad path returns file not found
3. `TestCmdFileArgParsing` — validates the argument is extracted correctly

### Manual Test (requires API key)
1. `vox file test.m4a` → verify transcription prints and clipboard has text
2. `vox ls` → verify the transcription appears in history
3. `vox file nonexistent.wav` → verify clear error message

## Dependencies

- No new Go dependencies
- No new system dependencies (SoX is NOT required for this command)
- Requires: `OPENAI_API_KEY` env var, clipboard tool

## Acceptance Criteria

- [ ] `vox file memo.m4a` transcribes and copies to clipboard
- [ ] Transcription saved to history with `duration_s: 0`
- [ ] Spinner displays during API call
- [ ] Clear error for missing argument, missing file, missing API key
- [ ] Ctrl+C during transcription exits cleanly
- [ ] All existing tests pass
- [ ] New unit tests pass
