# Prime
> Execute the following sections to understand the whisper project, then summarize your understanding.

## Run
List the project structure:
```bash
ls -la
ls scripts/
ls .claude/commands/
```

## Read
- README.md (project overview, setup, usage)
- pyproject.toml (Python dependencies)
- One spec file from specs/ to understand the planning workflow

## Summarize
After reading, provide a summary of:
1. What this tool does (audio transcription via OpenAI Whisper)
2. Key directories (scripts/, queue/, archive/, output/, specs/)
3. Main slash commands available (/transcribe, /queue_status, /process_queue)
4. How to set up the project (uv sync, ffmpeg, OPENAI_API_KEY)
