---
allowed-tools: Bash(./scripts/record_memo.sh:*), Bash(scripts/record_memo.sh:*), Bash(ls:*), Glob
description: Record voice memo to queue
argument-hint: [duration_in_seconds]
---

# Record Voice Memo

Record audio from your microphone and save to `queue/` for transcription.

## Instructions

1. Run the recording script with optional duration (default 5 minutes):
   ```bash
   ./scripts/record_memo.sh $ARGUMENTS
   ```

2. The script will:
   - Display "Recording..." indicator
   - Record until Ctrl+C or max duration reached
   - Save to `queue/memo_YYYY-MM-DD_HH-MM-SS.wav`

3. After recording completes, show the user:
   - File location and size
   - Next steps: `/queue_status` and `/process_queue`

## Error Handling

If sox is not installed, the script will display installation instructions.

If recording fails, check:
- Microphone permissions (macOS: System Settings → Privacy & Security → Microphone)
- sox installation: `sox --version`

## Tips

- **Stop early**: Press Ctrl+C to stop before max duration
- **Custom duration**: `/record_memo 60` for 60-second max
- **Batch process**: Record multiple memos, then `/process_queue`
- **System hotkey**: For recording outside Claude, see docs/hotkey-setup-macos.md
