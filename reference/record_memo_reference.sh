#!/usr/bin/env bash
#
# record_memo.sh - Record voice memo and save to queue/
#
# Records audio from the default microphone using sox and saves
# to the queue/ directory with a timestamped filename.
#
# Usage:
#   ./scripts/record_memo.sh           # Record for max 5 minutes
#   ./scripts/record_memo.sh 60        # Record for max 60 seconds
#   ./scripts/record_memo.sh --help    # Show help
#
# Press Ctrl+C to stop recording early.

set -euo pipefail

# Configuration
DEFAULT_DURATION=300  # 5 minutes max by default
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
QUEUE_DIR="$SCRIPT_DIR/../queue"

# Output file path (set in main, used by cleanup)
OUTPUT_FILE=""

# Show help
show_help() {
    cat << EOF
Usage: $(basename "$0") [duration_in_seconds]

Records audio from microphone and saves to queue/ directory.
Press Ctrl+C to stop recording early.

Arguments:
  duration    Maximum recording duration in seconds (default: 300 = 5 minutes)

Examples:
  $(basename "$0")          # Record for max 5 minutes
  $(basename "$0") 60       # Record for max 60 seconds
  $(basename "$0") 1800     # Record for max 30 minutes

After recording:
  - Check queue: /queue_status (or: uv run python scripts/queue_status.py)
  - Process queue: /process_queue
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
        echo "Error: sox is not installed"
        echo ""
        echo "Install with:"
        echo "  macOS:  brew install sox"
        echo "  Ubuntu: sudo apt-get install sox libsox-fmt-all"
        echo ""
        echo "See README.md for more details."
        exit 1
    fi
}

# Ensure queue directory exists
ensure_queue_dir() {
    if [[ ! -d "$QUEUE_DIR" ]]; then
        echo "Creating queue/ directory..."
        mkdir -p "$QUEUE_DIR"
    fi
}

# Get file size in human-readable format (cross-platform)
get_file_size() {
    local file="$1"
    if [[ -f "$file" ]]; then
        # Try macOS stat first, then Linux stat
        local size
        size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")

        if [[ "$size" -gt 1048576 ]]; then
            echo "$(( size / 1048576 )) MB"
        elif [[ "$size" -gt 1024 ]]; then
            echo "$(( size / 1024 )) KB"
        else
            echo "$size bytes"
        fi
    else
        echo "0 bytes"
    fi
}

# Verify recording was successful
verify_recording() {
    if [[ -f "$OUTPUT_FILE" ]]; then
        local file_size
        file_size=$(stat -f%z "$OUTPUT_FILE" 2>/dev/null || stat -c%s "$OUTPUT_FILE" 2>/dev/null || echo "0")

        if [[ "$file_size" -gt 1000 ]]; then
            echo ""
            echo "Saved: $OUTPUT_FILE ($(get_file_size "$OUTPUT_FILE"))"
            echo ""
            echo "Next steps:"
            echo "  Check queue:   /queue_status"
            echo "  Process queue: /process_queue"
        else
            echo ""
            echo "Warning: Recording file is very small ($file_size bytes)"
            echo "Check microphone permissions and try again."
        fi
    else
        echo ""
        echo "Error: Recording failed - no file created"
        exit 1
    fi
}

# Cleanup function for graceful shutdown
cleanup() {
    echo ""
    echo "Recording stopped"
    verify_recording
    exit 0
}

# Trap Ctrl+C and termination signals
trap cleanup SIGINT SIGTERM

# Main recording function
main() {
    local duration="${1:-$DEFAULT_DURATION}"

    # Validate duration is a number
    if ! [[ "$duration" =~ ^[0-9]+$ ]]; then
        echo "Error: Duration must be a positive number (seconds)"
        echo "Usage: $(basename "$0") [duration_in_seconds]"
        exit 1
    fi

    check_sox
    ensure_queue_dir

    # Generate timestamped filename
    local timestamp
    timestamp=$(date +%Y-%m-%d_%H-%M-%S)
    OUTPUT_FILE="$QUEUE_DIR/memo_$timestamp.wav"

    echo "Recording... Press Ctrl+C to stop (max ${duration}s)"
    echo "Output: $OUTPUT_FILE"
    echo ""

    # Record with settings optimized for voice transcription
    # -d: default input device
    # -r 16000: 16kHz sample rate (sufficient for voice, smaller files)
    # -c 1: mono (voice doesn't need stereo)
    # -b 16: 16-bit depth
    # trim 0 $duration: record for max duration seconds
    sox -d \
        -r 16000 \
        -c 1 \
        -b 16 \
        "$OUTPUT_FILE" \
        trim 0 "$duration" 2>&1 || true

    # If we get here without trap, recording completed naturally
    verify_recording
}

# Run main with all arguments
main "$@"
