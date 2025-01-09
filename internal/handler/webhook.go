package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lukeberry99/webhook-consumer/internal/storage"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request, store storage.WebhookStorage) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(rawBody))

	if _, err := decoder.Token(); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	receivedAt := time.Now()
	var eventName string

	var depth int
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			http.Error(w, "Error reading JSON", http.StatusBadRequest)
			return
		}

		switch token {
		case json.Delim('{'):
			depth++
			continue
		case json.Delim('}'):
			depth--
			continue
		case json.Delim('['):
			depth++
			continue
		case json.Delim(']'):
			depth--
			continue
		}

		if depth == 0 {
			if key, ok := token.(string); ok {
				if key == "event_type" {
					token, err := decoder.Token()
					if err != nil {
						http.Error(w, "Error reading JSON", http.StatusBadRequest)
						return
					}
					if str, ok := token.(string); ok {
						eventName = str
						break // Exit immediately if we find event_type
					}
				} else if eventName == "" { // Only check other fields if we haven't found event_type
					for _, field := range []string{"type", "title", "id"} {
						if key == field {
							token, err := decoder.Token()
							if err != nil {
								http.Error(w, "Error reading JSON", http.StatusBadRequest)
								return
							}
							if str, ok := token.(string); ok {
								eventName = str
							}
							break
						}
					}
				}
			}
		}
	}

	if eventName == "" {
		eventName = generateFilename(rawBody)
	}

	var rawJSON interface{}
	if err := json.Unmarshal(rawBody, &rawJSON); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	event := &storage.WebhookEvent{
		EventName:  eventName,
		ReceivedAt: receivedAt,
		RawEvent:   rawJSON,
	}

	filename, err := store.Store(event)
	if err != nil {
		log.Printf("Error storing webhook: %v", err)
		http.Error(w, "Error processing webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Webhook processed: %s\n", filename)

	w.WriteHeader(http.StatusOK)
}

func generateFilename(rawBody []byte) string {
	// Generate a hash of the request payload, and use that as a unique
	// identifier in the filename
	hasher := sha256.New()
	hasher.Write(rawBody)
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString[:8]
}
