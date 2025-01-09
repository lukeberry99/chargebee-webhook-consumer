package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type FileStorage struct {
	baseDir string
}

type WebhookEvent struct {
	EventName  string
	ReceivedAt time.Time
	RawEvent   interface{}
}

type WebhookStorage interface {
	Store(event *WebhookEvent) (string, error)
}

func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}
	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

func (fs *FileStorage) Store(event *WebhookEvent) (string, error) {
	filename := fmt.Sprintf("%s/%d_%s.json", fs.baseDir, event.ReceivedAt.Unix(), event.EventName)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	data := map[string]interface{}{
		"received_at": event.ReceivedAt.Format(time.RFC3339),
		"event":       event.RawEvent,
	}

	if err := encoder.Encode(data); err != nil {
		return "", fmt.Errorf("encoding JSON: %w", err)
	}

	return filename, nil
}
