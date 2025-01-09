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
	ReceivedAt time.Time
	OccurredAt int64
	EventType  string
	RawEvent   interface{}
}

type WebhookStorage interface {
	Store(event *WebhookEvent) error
}

func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}
	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

func (fs *FileStorage) Store(event *WebhookEvent) error {
	filename := fmt.Sprintf("%s/%d_%s.json", fs.baseDir, event.OccurredAt, event.EventType)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	data := map[string]interface{}{
		"received_at": event.ReceivedAt.Format(time.RFC3339),
		"event":       event.RawEvent,
	}

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
