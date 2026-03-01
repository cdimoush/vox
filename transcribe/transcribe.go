package transcribe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ErrNoAPIKey is returned when OPENAI_API_KEY is not set.
var ErrNoAPIKey = errors.New("OPENAI_API_KEY environment variable not set")

// ErrAPI is a sentinel for OpenAI API errors (rate limits, timeouts, server errors).
var ErrAPI = errors.New("API error")

// Transcribe sends the audio file at filePath to the OpenAI Whisper API
// and returns the transcribed text and audio duration in seconds.
// For files longer than 8 minutes, the audio is automatically chunked
// into 5-minute segments and transcribed sequentially.
func Transcribe(ctx context.Context, filePath string) (string, float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", 0, ErrNoAPIKey
	}

	if _, err := os.Stat(filePath); err != nil {
		return "", 0, fmt.Errorf("audio file: %w", err)
	}

	duration, err := GetDuration(filePath)
	if err != nil {
		// If we can't get duration, transcribe as single file.
		duration = 0
	}

	// Chunk if longer than 8 minutes.
	if duration > ChunkThreshold {
		return transcribeChunked(ctx, apiKey, filePath, duration)
	}

	text, err := transcribeSingle(ctx, apiKey, filePath)
	if err != nil {
		return "", duration, err
	}
	return text, duration, nil
}

// transcribeSingle transcribes a single audio file.
func transcribeSingle(ctx context.Context, apiKey, filePath string) (string, error) {
	client := openai.NewClient(apiKey)
	resp, err := client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    "gpt-4o-mini-transcribe",
		FilePath: filePath,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAPI, err)
	}
	return resp.Text, nil
}

// transcribeChunked splits the file into chunks and transcribes each.
func transcribeChunked(ctx context.Context, apiKey, filePath string, duration float64) (string, float64, error) {
	chunks, err := ChunkFile(filePath, duration)
	if err != nil {
		return "", duration, fmt.Errorf("chunking audio: %w", err)
	}
	defer func() {
		for _, c := range chunks {
			os.Remove(c)
		}
	}()

	var parts []string
	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return "", duration, ctx.Err()
		default:
		}

		text, err := transcribeSingle(ctx, apiKey, chunk)
		if err != nil {
			return strings.Join(parts, " "), duration, err
		}
		parts = append(parts, strings.TrimSpace(text))
	}

	return strings.Join(parts, " "), duration, nil
}
