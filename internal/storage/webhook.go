package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileStorage struct {
	baseDir         string
	updates         chan []EventListItem
	selectedService string
}

type WebhookEvent struct {
	ReceivedAt time.Time
	RawEvent   interface{}
}

type WebhookStorage interface {
	Store(event *WebhookEvent, rawBody []byte) (string, error)
}

type EventListItem struct {
	Filename    string
	ReceivedAt  string
	ServiceName string
}

func getWebhookDataDirectory() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "whook", "webhooks")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "webhook-data")
	}

	return filepath.Join(home, ".local", "share", "whook", "webhooks")
}

func NewFileStorage(customPath string) (*FileStorage, error) {
	baseDir := customPath
	if baseDir == "" {
		baseDir = getWebhookDataDirectory()
	}
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

	// Watch base directory
	if err := watcher.Add(fs.baseDir); err != nil {
		fmt.Printf("Error watching base directory: %v\n", err)
		return
	}

	// Watch all existing service directories
	if err := filepath.Walk(fs.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				fmt.Printf("Error watching directory %s: %v\n", path, err)
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("Error setting up directory watchers: %v\n", err)
	}

	var timer *time.Timer
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := watcher.Add(event.Name); err != nil {
						fmt.Printf("Error watching new directory %s: %v\n", event.Name, err)
					}
				}
			}

			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Write) != 0 {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(100*time.Millisecond, func() {
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

func (fs *FileStorage) ListEvents() ([]EventListItem, error) {
	searchDir := fs.baseDir
	if fs.selectedService != "" && fs.selectedService != "All" {
		searchDir = filepath.Join(fs.baseDir, fs.selectedService)
		if err := os.MkdirAll(searchDir, 0750); err != nil {
			return nil, fmt.Errorf("creating service directory: %w", err)
		}
	}

	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		return []EventListItem{}, nil
	}

	var items []EventListItem

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		relPath, err := filepath.Rel(fs.baseDir, path)
		if err != nil {
			return nil
		}

		pathParts := strings.Split(filepath.Dir(relPath), string(filepath.Separator))
		var serviceName string
		if len(pathParts) > 0 && pathParts[0] != "." {
			serviceName = pathParts[0]
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var fileData struct {
			ReceivedAt string `json:"received_at"`
		}
		if err := json.Unmarshal(data, &fileData); err != nil {
			return nil
		}

		timestamp, err := time.Parse(time.RFC3339, fileData.ReceivedAt)
		if err != nil {
			return nil
		}

		formattedTime := timestamp.Format("02/01/2006 15:04:05")

		items = append(items, EventListItem{
			Filename:    filepath.Base(path),
			ReceivedAt:  formattedTime,
			ServiceName: serviceName,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	// Sort by filename (which includes timestamp), newest first
	sort.Slice(items, func(i, j int) bool {
		return items[i].Filename > items[j].Filename
	})

	return items, nil
}

func (fs *FileStorage) ReadEvent(filename string) ([]byte, error) {
	filepath := fs.GetFullPath(filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filepath, err)
	}

	return data, nil
}

func (fs *FileStorage) Store(event *WebhookEvent, rawBody []byte) (string, error) {
	storageDir := fs.baseDir
	if fs.selectedService != "" && fs.selectedService != "All" {
		storageDir = filepath.Join(fs.baseDir, fs.selectedService)
	}

	if err := os.MkdirAll(storageDir, 0750); err != nil {
		return "", fmt.Errorf("creating storage directory: %w", err)
	}

	filename := filepath.Join(storageDir, fmt.Sprintf("%s_%s.json",
		event.ReceivedAt.Format("150405"),
		fs.generateUniqueFilename(rawBody)))

	if fs.selectedService != "" && fs.selectedService != "All" {
		filename = filepath.Join(storageDir, fmt.Sprintf("%s_%s_%s.json",
			event.ReceivedAt.Format("150405"),
			fs.selectedService,
			fs.generateUniqueFilename(rawBody)))
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0640)
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

	// Trigger an immediate update after storing
	go func() {
		// Small delay to ensure file is written
		time.Sleep(50 * time.Millisecond)
		items, err := fs.ListEvents()
		if err == nil {
			fs.updates <- items
		}
	}()

	return filename, nil
}

func (fs *FileStorage) GetFullPath(filename string) string {
	if fs.selectedService != "" && fs.selectedService != "All" {
		return fmt.Sprintf("%s/%s/%s", fs.baseDir, fs.selectedService, filename)
	}
	return fmt.Sprintf("%s/%s", fs.baseDir, filename)
}

func (fs *FileStorage) SetSelectedService(service string) {
	fs.selectedService = service

	// Trigger an update of the file browser when the selected service changes
	items, err := fs.ListEvents()
	if err == nil {
		fs.updates <- items
	}
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
