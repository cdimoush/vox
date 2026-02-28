# vox — Reference Links

## Go Ecosystem

- **Go standard library**: https://pkg.go.dev/std
- **Go modules reference**: https://go.dev/ref/mod
- **GoReleaser** (cross-platform builds + Homebrew): https://goreleaser.com/

## TUI & CLI Libraries (for future reference, not v1)

- **Bubbletea** (TUI framework): https://github.com/charmbracelet/bubbletea
- **Bubbles** (TUI components): https://github.com/charmbracelet/bubbles
- **Lip Gloss** (TUI styling): https://github.com/charmbracelet/lipgloss
- **Cobra** (CLI framework): https://github.com/spf13/cobra

## Beads Ecosystem

- **Beads** (issue tracking): https://github.com/steveyegge/beads
- **Perles** (TUI for Beads, Go/Bubbletea): https://github.com/zjrosen/perles
  - Architecture reference for Go project structure, GoReleaser config, Homebrew tap
  - Uses Bubbletea for TUI, BQL for queries, direct beads DB integration

## OpenAI Audio API

- **Whisper API reference**: https://platform.openai.com/docs/guides/speech-to-text
- **Go OpenAI client** (community): https://github.com/sashabaranov/go-openai
- **Models**: `gpt-4o-mini-transcribe` (default, fast), `gpt-4o-transcribe` (higher quality)
- **Limits**: 25MB max file size, supported formats: mp3, mp4, mpeg, mpga, m4a, wav, webm

## Audio Recording

- **SoX** (Sound eXchange): https://sox.sourceforge.net/
  - `rec` command for recording, stderr shows volume levels
  - Voice-optimized flags: `-r 16000 -c 1 -b 16` (16kHz mono 16-bit)
  - Install: `brew install sox` (macOS), `apt install sox` (Linux)

## Clipboard

- **macOS**: `pbcopy` (built-in)
- **Linux**: `xclip` (`apt install xclip`) or `xsel` (`apt install xsel`)
  - xsel preferred — persists after process exits (xclip dies with terminal)

## Existing Whisper Scripts (copied to reference/)

- `transcribe_reference.py` — OpenAI Whisper API call, chunking for long audio
- `instant_memo_reference.sh` — Record + transcribe + clipboard (the flow vox replaces)
- `record_memo_reference.sh` — SoX recording with volume settings, signal handling
