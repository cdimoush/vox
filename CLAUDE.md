# vox — Development Sandbox

You are building **vox**, a Go CLI tool for voice-to-clipboard with persistent history.

## Start Here

1. Read `vision.md` — this is the complete, tight spec. All decisions are made.
2. Read `reference/links.md` — key libraries, APIs, and ecosystem links.
3. Browse `reference/` — existing Python/Bash scripts that vox replaces. Study the SoX recording flags, OpenAI API call pattern, and clipboard detection logic. Your Go code will do the same things.

## What This Directory Is

A development sandbox. Code is built here iteratively by you (the agent) and sub-agents. This is not a monorepo subfolder — it's a standalone Go project that will eventually get its own repo.

## Development Approach

This will be built in multiple steps across sessions. Each step should produce working, testable code. Expect to be guided by a staff-level developer who will:

- Break work into incremental steps
- Review and redirect between steps
- Use `/epic` or beads to track progress
- Delegate sub-tasks to sub-agents

You should:

- Keep the Go code simple and idiomatic
- Use the standard library where possible, minimize dependencies
- Follow the package structure in the vision (`cmd/vox/`, `recorder/`, `transcribe/`, `clipboard/`, `history/`)
- Write tests as you go
- Target macOS and Linux — no Windows support needed

## Key Technical Decisions (Already Made)

- **Language**: Go
- **Transcription**: OpenAI Whisper API (cloud only, no local, ever)
- **Audio capture**: SoX `rec` command (shelled out, not a Go audio library)
- **Clipboard**: `pbcopy` (macOS) / `xclip` or `xsel` (Linux), detected at runtime
- **History**: JSONL file at `~/.vox/history.jsonl`
- **Config**: `OPENAI_API_KEY` env var only. No config file.
- **Distribution**: GoReleaser for cross-platform binaries + Homebrew tap

## Don't

- Don't add features not in the vision
- Don't add a config file
- Don't add local transcription
- Don't add beads or pipeline integration (that's post-v1)
- Don't use Bubbletea or any TUI framework (v1 is a simple CLI)
- Don't over-abstract — this is a small tool, not a framework
