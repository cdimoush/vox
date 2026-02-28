#!/usr/bin/env python3
"""Transcribe audio files using OpenAI's Whisper API."""

import os
import sys
import tempfile
from pathlib import Path

SUPPORTED_FORMATS = {"mp3", "mp4", "mpeg", "mpga", "m4a", "wav", "webm"}
MAX_FILE_SIZE_MB = 25
CHUNK_DURATION_MS = 5 * 60 * 1000  # 5 minutes in milliseconds
CHUNK_THRESHOLD_MS = 8 * 60 * 1000  # Only chunk files longer than 8 minutes

# Map file extensions to ffmpeg export format names (some differ from extension)
EXPORT_FORMAT_MAP = {"m4a": "ipod", "mpga": "mp3"}


def get_audio_duration_ms(path: str) -> int:
    """Get the duration of an audio file in milliseconds."""
    from pydub import AudioSegment

    audio = AudioSegment.from_file(path)
    return len(audio)


def split_audio_into_chunks(path: str, chunk_duration_ms: int = CHUNK_DURATION_MS) -> list[str]:
    """Split an audio file into chunks if it exceeds the threshold duration.

    Args:
        path: Path to the audio file
        chunk_duration_ms: Duration of each chunk in milliseconds

    Returns:
        List of file paths (original path if no splitting needed, or temp chunk paths)
    """
    from pydub import AudioSegment

    audio = AudioSegment.from_file(path)
    duration_ms = len(audio)

    if duration_ms <= CHUNK_THRESHOLD_MS:
        return [path]

    chunk_paths = []
    for i, start_ms in enumerate(range(0, duration_ms, chunk_duration_ms)):
        end_ms = min(start_ms + chunk_duration_ms, duration_ms)
        chunk = audio[start_ms:end_ms]

        # Create temp file with same extension for compatibility
        ext = Path(path).suffix.lower().lstrip(".")
        export_format = EXPORT_FORMAT_MAP.get(ext, ext)
        temp_file = tempfile.NamedTemporaryFile(suffix=f".{ext}", delete=False)
        chunk.export(temp_file.name, format=export_format)
        chunk_paths.append(temp_file.name)

    return chunk_paths


def transcribe_audio(path: str, model: str = "gpt-4o-mini-transcribe") -> str:
    """Transcribe an audio file using OpenAI's Whisper API.

    Args:
        path: Path to the audio file
        model: OpenAI model to use for transcription

    Returns:
        Transcribed text
    """
    from openai import OpenAI

    client = OpenAI()
    with open(path, "rb") as f:
        tx = client.audio.transcriptions.create(
            model=model,
            file=f,
        )
    return tx.text


def transcribe_chunks(chunk_paths: list[str], original_path: str, model: str = "gpt-4o-mini-transcribe") -> str:
    """Transcribe multiple audio chunks and concatenate the results.

    Args:
        chunk_paths: List of paths to audio chunk files
        original_path: Original audio file path (to know which files are temp)
        model: OpenAI model to use for transcription

    Returns:
        Concatenated transcribed text from all chunks
    """
    transcripts = []
    try:
        for i, chunk_path in enumerate(chunk_paths):
            print(f"Transcribing chunk {i + 1}/{len(chunk_paths)}...", file=sys.stderr)
            transcript = transcribe_audio(chunk_path, model)
            transcripts.append(transcript)
    finally:
        # Clean up temporary chunk files
        for chunk_path in chunk_paths:
            if chunk_path != original_path and os.path.exists(chunk_path):
                os.unlink(chunk_path)

    return " ".join(transcripts)


def main():
    # Load environment variables from .env file
    try:
        from dotenv import load_dotenv
        load_dotenv()
    except ImportError:
        pass  # dotenv not installed, rely on environment variables

    if len(sys.argv) < 2:
        print("Error: No audio file path provided", file=sys.stderr)
        print(f"Usage: {sys.argv[0]} <audio-file-path>", file=sys.stderr)
        sys.exit(1)

    audio_path = sys.argv[1]

    # Check if file exists
    if not os.path.exists(audio_path):
        print(f"Error: File not found: {audio_path}", file=sys.stderr)
        sys.exit(1)

    # Check file extension
    ext = Path(audio_path).suffix.lower().lstrip(".")
    if ext not in SUPPORTED_FORMATS:
        print(f"Error: Unsupported file format: .{ext}", file=sys.stderr)
        print(f"Supported formats: {', '.join(sorted(SUPPORTED_FORMATS))}", file=sys.stderr)
        sys.exit(1)

    # Check file size
    file_size_mb = os.path.getsize(audio_path) / (1024 * 1024)
    if file_size_mb > MAX_FILE_SIZE_MB:
        print(f"Error: File too large ({file_size_mb:.1f}MB). Maximum is {MAX_FILE_SIZE_MB}MB.", file=sys.stderr)
        print("Tip: Compress with ffmpeg: ffmpeg -i input.m4a -vn -ac 1 -ar 16000 -b:a 48k output.m4a", file=sys.stderr)
        sys.exit(1)

    # Check for API key
    if not os.environ.get("OPENAI_API_KEY"):
        print("Error: OPENAI_API_KEY environment variable not set", file=sys.stderr)
        print("Set it in your .env file or export it: export OPENAI_API_KEY=your-key", file=sys.stderr)
        sys.exit(1)

    # Check duration and split into chunks if needed
    try:
        duration_ms = get_audio_duration_ms(audio_path)
        duration_min = duration_ms / 1000 / 60

        if duration_ms > CHUNK_THRESHOLD_MS:
            num_chunks = (duration_ms + CHUNK_DURATION_MS - 1) // CHUNK_DURATION_MS
            print(f"Audio is {duration_min:.1f} minutes, splitting into {num_chunks} chunks...", file=sys.stderr)
            chunk_paths = split_audio_into_chunks(audio_path)
            transcript = transcribe_chunks(chunk_paths, audio_path)
        else:
            transcript = transcribe_audio(audio_path)

        print(transcript)
    except Exception as e:
        print(f"Error during transcription: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
