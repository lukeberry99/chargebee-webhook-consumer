package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileStorage struct {
	baseDir string
	updates chan []EventListItem
}

type WebhookEvent struct {
	ReceivedAt time.Time
	RawEvent   interface{}
}

type WebhookStorage interface {
	Store(event *WebhookEvent, rawBody []byte) (string, error)
}

func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}

	fs := &FileStorage{
		baseDir: baseDir,
		updates: make(chan []EventListItem, 1),
	}

	go fs.watchDirectory()

	return fs, nil
}

func (fs *FileStorage) WatchEvents() <-chan []EventListItem {
	return fs.updates
}

func (fs *FileStorage) watchDirectory() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(fs.baseDir)
	if err != nil {
		fmt.Printf("Error watching directory: %v\n", err)
		return
	}

	// Debounce timer to prevent rapid updates
	var timer *time.Timer
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only care about create, remove, or rename operations
			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if timer != nil {
					timer.Stop()
				}
				// Wait a short period before updating to batch rapid changes
				timer = time.AfterFunc(100*time.Millisecond, func() {
					// Get updated list
					items, err := fs.ListEvents()
					if err != nil {
						fmt.Printf("Error listing events: %v\n", err)
						return
					}
					fs.updates <- items
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

type EventListItem struct {
	Filename   string
	ReceivedAt string
}

func (fs *FileStorage) ListEvents() ([]EventListItem, error) {
	entries, err := os.ReadDir(fs.baseDir)
	if err != nil {
		return nil, fmt.Errorf("reading storage directory: %w", err)
	}

	items := make([]EventListItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Read and parse the file to get the received_at timestamp
		filePath := fmt.Sprintf("%s/%s", fs.baseDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", filePath, err)
		}

		var fileData struct {
			ReceivedAt string `json:"received_at"`
		}
		if err := json.Unmarshal(data, &fileData); err != nil {
			return nil, fmt.Errorf("parsing JSON from %s: %w", filePath, err)
		}

		timestamp, err := time.Parse(time.RFC3339, fileData.ReceivedAt)
		if err != nil {
			return nil, fmt.Errorf("parsing timestamp from %s: %w", filePath, err)
		}

		formattedTime := timestamp.Format("02/01/2006 15:04:05")

		items = append(items, EventListItem{
			Filename:   entry.Name(),
			ReceivedAt: formattedTime,
		})
	}

	// Sort by filename (which includes timestamp), newest first
	sort.Slice(items, func(i, j int) bool {
		return items[i].Filename > items[j].Filename
	})

	return items, nil
}

func (fs *FileStorage) ReadEvent(filename string) ([]byte, error) {
	filepath := fmt.Sprintf("%s/%s", fs.baseDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filepath, err)
	}

	return data, nil
}

func (fs *FileStorage) Store(event *WebhookEvent, rawBody []byte) (string, error) {
	filename := fmt.Sprintf("%s/%d_%s.json", fs.baseDir, event.ReceivedAt.Unix(), fs.generateUniqueFilename(rawBody))

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

func (fs *FileStorage) GetFullPath(filename string) string {
	return fmt.Sprintf("%s/%s", fs.baseDir, filename)
}

// Generate a hash of the request payload for unique filenames
func (fs *FileStorage) generateUniqueFilename(rawBody []byte) string {
	hasher := sha256.New()
	hasher.Write(rawBody)
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	shortHash := hashString[:8]

	return shortHash
}
