package transcribe

import (
	"context"
	"errors"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// ErrNoAPIKey is returned when OPENAI_API_KEY is not set.
var ErrNoAPIKey = errors.New("OPENAI_API_KEY environment variable not set")

// Transcribe sends the audio file at filePath to the OpenAI Whisper API
// and returns the transcribed text.
func Transcribe(filePath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", ErrNoAPIKey
	}

	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("audio file: %w", err)
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateTranscription(context.Background(), openai.AudioRequest{
		Model:    "gpt-4o-mini-transcribe",
		FilePath: filePath,
	})
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	return resp.Text, nil
}
