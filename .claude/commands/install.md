# Install & Prime
> Set up the whisper project for use, then prime yourself on the codebase.

## Check System Dependencies
Check if ffmpeg is installed (required for audio processing):
```bash
which ffmpeg && ffmpeg -version | head -1
```
If ffmpeg is not installed, note this for the user (install via `brew install ffmpeg` on macOS).

## Install Python Dependencies
```bash
uv sync
```

## Check Environment Configuration
Check if .env file exists:
```bash
ls -la .env 2>/dev/null || echo "No .env file found"
```

If .env doesn't exist or is missing OPENAI_API_KEY:
1. Create .env file with: `OPENAI_API_KEY=your-key-here`
2. Or copy from another location if available

## Read and Execute
.claude/commands/prime.md

## Report
Output a summary of:
- What was installed (Python dependencies via uv)
- System dependency status (ffmpeg installed or needs installation)
- Environment status (.env file present or needs configuration)
- Any manual steps the user needs to complete
