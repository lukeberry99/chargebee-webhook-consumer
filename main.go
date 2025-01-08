package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/localtunnel/go-localtunnel"
)

type ChargebeeEvent struct {
	ID            string      `json:"id"`
	OccurredAt    int64       `json:"occurred_at"`
	Source        string      `json:"source"`
	User          string      `json:"user"`
	Object        string      `json:"object"`
	APIVersion    string      `json:"api_version"`
	Content       interface{} `json:"content"`
	EventType     string      `json:"event_type"`
	WebhookStatus string      `json:"webhook_status"`
	Webhooks      []struct {
		ID            string `json:"id"`
		WebhookStatus string `json:"webhook_status"`
		Object        string `json:"object"`
	} `json:"webhooks"`
}

type WebhookData struct {
	Timestamp string         `json:"timestamp"`
	Event     ChargebeeEvent `json:"event"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var chargebeeEvent ChargebeeEvent
	if err := json.Unmarshal(body, &chargebeeEvent); err != nil {
		log.Printf("Error parsing webhook: %v", err)
		http.Error(w, "Error parsing webhook", http.StatusBadRequest)
		return
	}

	webhook := WebhookData{
		Timestamp: time.Now().Format(time.RFC3339),
		Event:     chargebeeEvent,
	}

	webhookJSON, err := json.MarshalIndent(webhook, "", "  ")
	if err != nil {
		http.Error(w, "Error processing webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received webhook:\n%s\n", string(webhookJSON))

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Error creating logs directory: %v", err)
		http.Error(w, "Error logging webhook", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("logs/%s_%d.json", webhook.Event.EventType, webhook.Event.OccurredAt)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		http.Error(w, "Error logging webhook", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(string(webhookJSON) + "\n"); err != nil {
		log.Printf("Error writing to log file: %v", err)
		http.Error(w, "Error logging webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received successfully"))
}

func main() {
	tunnel, err := localtunnel.Listen(localtunnel.Options{})
	if err != nil {
		log.Fatalf("Unable to start localtunnel: %v", err)
	}

	fmt.Printf("Localtunnel URL: %s\n", tunnel.URL())

	if err := http.ListenAndServe(":8080", http.HandlerFunc(webhookHandler)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
