# Epic: File Transcription Command

## Epic Overview

Vox currently only supports live recording — you speak into your mic, it transcribes, and copies to clipboard. But there's a common need to transcribe an existing audio file: a voice memo recorded on your phone, an audio clip from another tool, or a file dropped into a queue by an agent pipeline.

This epic adds `vox file <path>` — a command that takes a path to an audio file, sends it through the same OpenAI Whisper transcription pipeline, copies the result to clipboard, and saves it to history. It reuses the existing `transcribe` and `clipboard` packages with zero new dependencies. The command is designed for both human use (quick CLI transcription) and agent use (programmatic transcription in automation pipelines).

The scope is intentionally small: one new command, one new file, tests, and documentation. No chunking, no batch processing, no format conversion — just the simplest path from file to clipboard.

## Vision Context

**Source Vision**: [Vision: vox](../../vision.md)

### Key Vision Points

**Current Pain Points Addressed:**
- No way to transcribe an existing audio file without recording live
- Agents and pipelines that produce audio files have no vox entry point
- Users with phone-recorded memos must use a separate transcription tool

**Vision Goals:**
1. Keep vox simple — a CLI that runs and exits
2. Clipboard-first output with persistent history
3. Future pipeline integration (this command is the natural hook point)

**Architecture Decision:**
- Reuse existing `transcribe.Transcribe()` — it already accepts a file path
- Reuse existing `clipboard.Write()` and `history.Store.Append()`
- No new packages or dependencies needed
- SoX is NOT required for this command (no recording involved)

## Specs in This Epic

### Phase 1: Core Implementation
- [ ] [Feature: File Transcribe Command](./feature-file-transcribe.md) - Add `vox file <path>` command

### Phase 2: Polish
- [ ] [Chore: Documentation Update](./chore-docs-update.md) - Update README and help text

## Execution Order

### Phase 1: Core Implementation (~30 min)
**Goal**: Working `vox file <path>` command that transcribes, copies, and saves to history.

Execute in order:
1. [Feature: File Transcribe Command](./feature-file-transcribe.md) - Core command implementation with tests

**Success Criteria**:
- `vox file recording.m4a` transcribes the file and copies text to clipboard
- Transcription is saved to history (visible in `vox ls`)
- Spinner shows during transcription
- Clear error messages for: missing file, missing API key, unsupported format
- All existing tests still pass
- New tests cover argument validation, file existence check, and error paths

**🔄 USER BREAKPOINT #1**: Manual test with a real audio file and API key. Verify the full flow: `vox file <path>` → transcription printed → clipboard populated → appears in `vox ls`.

---

### Phase 2: Polish (~10 min)
**Goal**: Documentation reflects the new command.

Execute in order:
1. [Chore: Documentation Update](./chore-docs-update.md) - Add `vox file` to README usage section and update help text in main.go

**Success Criteria**:
- README.md documents `vox file <path>` with example
- `vox` with unknown command shows `file` in the usage line
- vision.md is NOT modified (this is a post-v1 addition)

---

## Path Dependencies Diagram

```
Phase 1
    Feature: File Transcribe Command
        ↓
    🔄 USER BREAKPOINT: manual test
        ↓
Phase 2
    Chore: Documentation Update

Critical Path:
- Feature implementation → manual test → docs
```

## Implementation Notes

### Cross-Cutting Concerns

**Architecture Decisions:**
- New file: `cmd/vox/file.go` — follows existing pattern (one file per command: `ls.go`, `cp.go`, `show.go`, `clear.go`)
- Reuses `transcribeWithContext()` from `run.go` for context cancellation support
- Same spinner pattern as `run()` for consistent UX
- No new packages or dependencies

**Shared Dependencies:**
- `transcribe.Transcribe()` — already accepts arbitrary file paths
- `clipboard.Write()` — unchanged
- `history.Store.Append()` — unchanged
- Spinner code from `run.go` — can be extracted or duplicated (small enough to duplicate)

**Testing Strategy:**
- Unit tests for argument parsing and file validation (no API key needed)
- Integration test requires real API key (same pattern as existing transcribe tests)
- All existing tests must continue to pass

**Rollout Plan:**
- Feature branch: `feature/file-transcribe`
- Merge to main after user breakpoint passes
- No breaking changes to existing commands

### User Testing Breakpoints

This epic includes **1 explicit user testing breakpoint** (marked with 🔄):
1. After Phase 1: Manual end-to-end test with a real audio file to verify transcription, clipboard, and history integration

Each breakpoint is a decision point: proceed to next phase, iterate on current phase, or stop if goals are met.

## Success Metrics

- [ ] `vox file <path>` successfully transcribes and copies to clipboard
- [ ] Transcription appears in `vox ls` history
- [ ] Works with common formats: .wav, .m4a, .mp3, .webm
- [ ] Error messages are clear for bad paths and missing API key
- [ ] Zero regressions in existing commands
- [ ] README documents the new command

## Future Enhancements

Ideas that came up during planning but are out of scope for this epic:

1. **Batch file transcription** — `vox file *.m4a` or `vox file dir/` to process multiple files
2. **Format validation** — check file headers, not just extensions
3. **Large file chunking** — split files >25MB before sending to Whisper API
4. **Stdin support** — `cat audio.wav | vox file -` for piped input
5. **Output format flags** — `--json`, `--no-copy` for agent-friendly output
6. **Audio file duration display** — show duration before/after transcription
