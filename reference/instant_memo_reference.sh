#!/usr/bin/env bash
#
# instant_memo.sh - Record voice, transcribe, copy to clipboard
#
# Records audio, transcribes via OpenAI Whisper, and copies
# the result directly to your clipboard. No files saved.
#
# Usage:
#   ./scripts/instant_memo.sh           # Record for max 2 minutes
#   ./scripts/instant_memo.sh 30        # Record for max 30 seconds
#   ./scripts/instant_memo.sh --help    # Show help
#
# Press Ctrl+C to stop recording early.

set -euo pipefail

# Configuration
DEFAULT_DURATION=120  # 2 minutes max by default
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Temp file path (set in main, used by cleanup)
TEMP_FILE=""

# Show help
show_help() {
    cat << EOF
Usage: $(basename "$0") [duration_in_seconds]

Records audio, transcribes it, and copies the text to clipboard.
Press Ctrl+C to stop recording early.

Arguments:
  duration    Maximum recording duration in seconds (default: 120 = 2 minutes)

Examples:
  $(basename "$0")          # Record for max 2 minutes
  $(basename "$0") 30       # Record for max 30 seconds
  $(basename "$0") 60       # Record for max 1 minute

Requirements:
  - sox (for recording)
  - OPENAI_API_KEY in environment or .env
  - xclip or xsel (Linux) or pbcopy (macOS)
EOF
    exit 0
}

# Check for help flag
if [[ "${1:-}" == "--help" ]] || [[ "${1:-}" == "-h" ]]; then
    show_help
fi

# Check dependencies
check_sox() {
    if ! command -v sox &> /dev/null; then
        echo "Error: sox is not installed" >&2
        echo "" >&2
        echo "Install with:" >&2
        echo "  macOS:  brew install sox" >&2
        echo "  Ubuntu: sudo apt-get install sox libsox-fmt-all" >&2
        exit 1
    fi
}

# Get clipboard command for this platform
get_clipboard_cmd() {
    if command -v pbcopy &> /dev/null; then
        echo "pbcopy"
    elif command -v xsel &> /dev/null; then
        # Prefer xsel - it stores data persistently (xclip dies with terminal)
        echo "xsel --clipboard --input"
    elif command -v xclip &> /dev/null; then
        echo "xclip -selection clipboard"
    else
        echo ""
    fi
}

check_clipboard() {
    local clip_cmd
    clip_cmd=$(get_clipboard_cmd)
    if [[ -z "$clip_cmd" ]]; then
        echo "Error: No clipboard tool found" >&2
        echo "" >&2
        echo "Install with:" >&2
        echo "  macOS:  (pbcopy is built-in)" >&2
        echo "  Ubuntu: sudo apt-get install xclip" >&2
        exit 1
    fi
}

# Cleanup temp file
cleanup_temp() {
    if [[ -n "$TEMP_FILE" ]] && [[ -f "$TEMP_FILE" ]]; then
        rm -f "$TEMP_FILE"
    fi
}

# Transcribe and copy to clipboard
transcribe_and_copy() {
    if [[ ! -f "$TEMP_FILE" ]]; then
        echo "Error: No recording file" >&2
        exit 1
    fi

    local file_size
    file_size=$(stat -f%z "$TEMP_FILE" 2>/dev/null || stat -c%s "$TEMP_FILE" 2>/dev/null || echo "0")

    if [[ "$file_size" -lt 1000 ]]; then
        echo "Warning: Recording too short or empty" >&2
        cleanup_temp
        exit 1
    fi

    echo "Transcribing..." >&2

    local transcript
    transcript=$(cd "$SCRIPT_DIR/.." && uv run python scripts/transcribe.py "$TEMP_FILE")

    if [[ -z "$transcript" ]]; then
        echo "Error: Transcription returned empty" >&2
        cleanup_temp
        exit 1
    fi

    # Copy to clipboard
    local clip_cmd
    clip_cmd=$(get_clipboard_cmd)
    echo -n "$transcript" | $clip_cmd

    echo "" >&2
    echo "Copied to clipboard:" >&2
    echo "---" >&2
    echo "$transcript" >&2
    echo "---" >&2

    cleanup_temp
}

# Handle Ctrl+C - stop recording and transcribe
handle_interrupt() {
    echo "" >&2
    echo "Recording stopped" >&2
    transcribe_and_copy
    exit 0
}

trap handle_interrupt SIGINT SIGTERM
trap cleanup_temp EXIT

# Main
main() {
    local duration="${1:-$DEFAULT_DURATION}"

    # Validate duration
    if ! [[ "$duration" =~ ^[0-9]+$ ]]; then
        echo "Error: Duration must be a positive number" >&2
        exit 1
    fi

    check_sox
    check_clipboard

    # Create temp file (macOS and Linux compatible)
    TEMP_FILE=$(mktemp "${TMPDIR:-/tmp}/instant_memo.XXXXXX.wav")

    echo "Recording... Press Ctrl+C to stop (max ${duration}s)" >&2
    echo "" >&2

    # Record audio
    sox -d \
        -r 16000 \
        -c 1 \
        -b 16 \
        "$TEMP_FILE" \
        trim 0 "$duration" 2>&1 || true

    # If we get here, recording finished naturally
    transcribe_and_copy
}

main "$@"
