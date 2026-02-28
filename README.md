# vox

Speak into your terminal. Get text on your clipboard. Remember everything.

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

## Install

```bash
curl https://i.jpillora.com/cdimoush/vox! | bash
```

**From source** (requires Go 1.22+):

```bash
go install github.com/cdimoush/vox/cmd/vox@latest
```

**Pre-built binaries**: Download from the [Releases](https://github.com/cdimoush/vox/releases) page.

## Requirements

| Dependency | Required | Install |
|---|---|---|
| **SoX** | For recording | `brew install sox` (macOS) / `sudo apt install sox` (Linux) |
| **OpenAI API key** | Yes | `export OPENAI_API_KEY=your-key` |
| **Clipboard tool** | Yes | `pbcopy` (macOS, built-in) / `sudo apt install xsel` (Linux) |

## Usage

### `vox` — Record and transcribe

Start recording immediately. Press Enter or Ctrl+C to stop. Audio is sent to OpenAI Whisper, transcribed text is copied to your clipboard and saved to history.

```bash
$ vox
● Recording... (Enter to stop)
"Refactor the sensor config to use YAML"
✓ Copied to clipboard
```

### `vox file <path>` — Transcribe an audio file

```bash
$ vox file memo.m4a
⠋ Transcribing...
"Refactor the sensor config to use YAML"
✓ Copied to clipboard
```

Transcribes an existing audio file (.wav, .m4a, .mp3, .webm) without recording. Does not require SoX.

### `vox ls` — Show history

```bash
$ vox ls
#   When        Text
1   2m ago      Move the contact sensor config into YAML...
2   14m ago     Remind Nick about the gantry collision boundary...
3   1h ago      Need to add error handling for USD stage loading...
```

Flags:
- `vox ls -n 50` — show last 50 entries
- `vox ls --all` — show all entries

### `vox cp <n>` — Re-copy a history entry

```bash
$ vox cp 3
✓ Copied #3 to clipboard
```

### `vox show <n>` — Show full text

```bash
$ vox show 1
[2m ago]

Move the contact sensor config into a YAML file instead of hardcoding
the joint names. The current approach has a list of 14 joint names that
breaks every time someone adds a new robot model.
```

### `vox clear` — Clear history

```bash
$ vox clear
Delete all 47 transcriptions? [y/N] y
✓ History cleared
```

## Shell Aliases

```bash
# .zshrc / .bashrc
alias v="vox"
alias vl="vox ls"
alias vc="vox cp"
alias vf="vox file"
```

Four keystrokes to capture a thought: `v` → speak → Enter → paste.

## How It Works

vox shells out to SoX `rec` for audio capture, sends the WAV to the OpenAI Whisper API (`gpt-4o-mini-transcribe`), and pipes the result to your platform's clipboard tool. History is stored as append-only JSONL at `~/.vox/history.jsonl`.

No config file. No local transcription. No TUI framework. Just a CLI that runs and exits.

## Uninstall

```bash
rm $(which vox)
rm -rf ~/.vox/
```
