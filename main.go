package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type NgrokTunnel struct {
	PublicURL string `json:"public_url"`
}

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

	fmt.Printf("Webhook received: logs/%d_%s.json\n", webhook.Event.OccurredAt, webhook.Event.EventType)

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Error creating logs directory: %v", err)
		http.Error(w, "Error logging webhook", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("logs/%d_%s.json", webhook.Event.OccurredAt, webhook.Event.EventType)
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
}

func startNgrok() (string, error) {
	cmd := exec.Command("ngrok", "http", "8080")
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ngrok: %w", err)
	}

	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return "", fmt.Errorf("failed to query ngrok API: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Tunnels []NgrokTunnel `json:"tunnels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ngrok API response: %w", err)
	}

	for _, tunnel := range result.Tunnels {
		if tunnel.PublicURL != "" {
			return tunnel.PublicURL, nil
		}
	}

	return "", fmt.Errorf("ngrok URL not found")
}

func main() {
	url, err := startNgrok()
	if err != nil {
		log.Fatalf("Error starting Ngrok: %v", err)
	}

	fmt.Printf("Ngrok URL: %s\n", url)
	fmt.Println("Create a new webhook in Chargebee with this URL, and make sure that all events are sent. https://<team-name>.chargebee.com/apikeys_and_webhooks/webhooks")

	if err := http.ListenAndServe(":8080", http.HandlerFunc(webhookHandler)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
