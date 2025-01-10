package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lukeberry99/webhook-consumer/internal/storage"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request, store storage.WebhookStorage, logChan chan<- string) {
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

	var rawJSON interface{}
	if err := json.Unmarshal(rawBody, &rawJSON); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	event := &storage.WebhookEvent{
		ReceivedAt: receivedAt,
		RawEvent:   rawJSON,
	}

	filename, err := store.Store(event, rawBody)
	if err != nil {
		logChan <- fmt.Sprintf("Error storing webhook: %v", err)
		http.Error(w, "Error processing webhook", http.StatusInternalServerError)
		return
	}

	logChan <- fmt.Sprintf("Webhook processed: %s", filename)

	w.WriteHeader(http.StatusOK)
}
