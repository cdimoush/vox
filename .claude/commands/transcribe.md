---
allowed-tools: Bash(python:*), Bash(uv:*)
description: Transcribe audio file to text
argument-hint: <audio-file-path>
---

Run the transcription script on the provided audio file and return the transcribed text.

```bash
uv run python scripts/transcribe.py $ARGUMENTS
```

After transcription completes, you can:
- Summarize the content
- Extract action items or tasks
- Create meeting notes
- Answer questions about what was discussed
