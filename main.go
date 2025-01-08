package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type WebhookData struct {
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
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

	webhook := WebhookData{
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      string(body),
	}

	webhookJSON, err := json.MarshalIndent(webhook, "", "  ")
	if err != nil {
		http.Error(w, "Error processing webhook", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received webhook:\n%s\n", string(webhookJSON))

	f, err := os.OpenFile("webhooks.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	http.HandleFunc("/webhook", webhookHandler)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
