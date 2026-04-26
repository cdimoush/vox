# vox — Agent Guide

vox is a shipped Go CLI tool for voice-to-clipboard with persistent history. This is the production repository.

## Codebase Map

```
cmd/vox/          CLI entrypoint — main.go + one file per command
config/           API key discovery (env → ~/.vox/config → shell profiles)
recorder/         SoX audio capture, volume metering
transcribe/       OpenAI Whisper API, chunking for long audio
clipboard/        pbcopy (macOS) / xclip|xsel (Linux)
history/          Append-only JSONL at ~/.vox/history.jsonl
```

## Key Technical Decisions

| Area | Decision |
|------|----------|
| Language | Go — single static binary, no runtime deps |
| Transcription | OpenAI Whisper API (`gpt-4o-mini-transcribe`) — cloud only, always |
| Audio capture | SoX `rec` command (shelled out) |
| Clipboard | `pbcopy` / `xclip` / `xsel` — detected at runtime |
| History | `~/.vox/history.jsonl` — append-only JSONL |
| API key | `config.FindAPIKey()` — env var → `~/.vox/config` → shell profiles |
| Distribution | GoReleaser — Homebrew tap + pre-built binaries |
| No local transcription | This is a settled decision, not a gap |
| No TUI framework | v1 is a simple CLI. No Bubbletea. |

## Development

Build a dev binary (never overwrites the installed one):
```bash
go build -o vox-dev ./cmd/vox
```

Run all tests:
```bash
go test ./...
```

**When writing tests that clear `OPENAI_API_KEY`**, also set `HOME` to a temp dir — otherwise `config.FindAPIKey()` will find the key in the real `~/.bashrc` and the test won't behave as expected:
```go
t.Setenv("OPENAI_API_KEY", "")
t.Setenv("HOME", t.TempDir())
```

## Skills

| Skill | When to use |
|-------|-------------|
| `/concept` | Capturing a new feature idea or design question |
| `/trade-study` | Comparing implementation approaches before committing |
| `/blueprint` | Breaking an approved concept into concrete tasks |
| `/build` | Executing a blueprint task by task |
| `/engineer` | Full pipeline (concept → trade study → blueprint → build → PR) for self-contained features |
| `/design` | Open-ended system design research |
| `/vox-up` | **Dev only** — rebuild + restart the `vox ui` daemon on this branch. Remove when the UI daemon ships. |

All work lives in beads (`bd`). No markdown planning files.

## Rules

- Always work on a feature branch. Never commit directly to main.
- Always end with a pull request — even if incomplete, open it with `[FAIL]` prefix.
- Keep Go simple and idiomatic. Minimize dependencies.
- Target macOS and Linux. No Windows.
- Don't add local transcription. Don't add a TUI. Don't over-abstract.


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
