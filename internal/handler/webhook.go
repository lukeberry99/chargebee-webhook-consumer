package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lukeberry99/chargebee-webhook-consumer/internal/storage"
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
	var occurredAt int64
	var eventType string

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			http.Error(w, "Error reading JSON", http.StatusBadRequest)
			return
		}

		if key, ok := token.(string); ok {
			switch key {
			case "occurred_at":
				token, err := decoder.Token()
				if err != nil {
					http.Error(w, "Error reading JSON", http.StatusBadRequest)
					return
				}
				if num, ok := token.(float64); ok {
					occurredAt = int64(num)
				}
			case "event_type":
				token, err := decoder.Token()
				if err != nil {
					http.Error(w, "Error reading JSON", http.StatusBadRequest)
					return
				}
				if str, ok := token.(string); ok {
					eventType = str
				}
			default:
				if _, err := decoder.Token(); err != nil {
					http.Error(w, "Error reading JSON", http.StatusBadRequest)
					return
				}
			}
		}
	}

	if eventType == "" || occurredAt == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var rawJSON interface{}
	if err := json.Unmarshal(rawBody, &rawJSON); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	delta := receivedAt.Sub(time.Unix(occurredAt, 0))

	event := &storage.WebhookEvent{
		ReceivedAt: receivedAt,
		OccurredAt: occurredAt,
		EventType:  eventType,
		Delta:      delta,
		RawEvent:   rawJSON,
	}

	if err := store.Store(event); err != nil {
		log.Printf("Error storing webhook: %v", err)
		http.Error(w, "Error processing webhook", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
