package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWebhookHandler(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		method         string
		payload        interface{}
		expectedStatus int
		validateFile   bool
	}{
		{
			name:   "Valid POST request",
			method: http.MethodPost,
			payload: ChargebeeEvent{
				ID:         "ev_test_123",
				OccurredAt: time.Now().Unix(),
				EventType:  "subscription_created",
				Content:    map[string]interface{}{"test": "data"},
			},
			expectedStatus: http.StatusOK,
			validateFile:   true,
		},
		{
			name:           "Invalid method GET",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateFile:   false,
		},
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			validateFile:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.payload != nil {
				if str, ok := tt.payload.(string); ok {
					body = []byte(str)
				} else {
					body, err = json.Marshal(tt.payload)
					if err != nil {
						t.Fatalf("Failed to marshal test payload: %v", err)
					}
				}
			}

			req := httptest.NewRequest(tt.method, "/", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			webhookHandler(w, req, tmpDir)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateFile {
				files, err := filepath.Glob(filepath.Join(tmpDir, "*"))
				if err != nil {
					t.Fatalf("Failed to read log directory: %v", err)
				}

				if len(files) != 1 {
					t.Errorf("Expected 1 log file, got %d", len(files))
					return
				}

				content, err := os.ReadFile(files[0])
				if err != nil {
					t.Fatalf("Failed to read log file: %v", err)
				}

				var webhookData WebhookData
				if err := json.Unmarshal(content, &webhookData); err != nil {
					t.Fatalf("Failed to parse log file content: %v", err)
				}

				if webhookData.ReceivedAt == "" {
					t.Error("ReceivedAt timestamp is empty")
				}

				if webhookData.Delta == "" {
					t.Error("Delta is empty")
				}

				if webhookData.Event.ID != "ev_test_123" {
					t.Errorf("Expected event ID 'ev_test_123', got '%s'", webhookData.Event.ID)
				}
			}
		})
	}
}

func TestWebhookDataTimestamps(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-5 * time.Minute)

	event := ChargebeeEvent{
		ID:         "ev_test_456",
		OccurredAt: pastTime.Unix(),
		EventType:  "test_event",
	}

	webhook := WebhookData{
		ReceivedAt: now.Format(time.RFC3339),
		Delta:      now.Sub(pastTime).String(),
		Event:      event,
	}

	expectedDelta := "5m0s"
	if !strings.Contains(webhook.Delta, "5m") {
		t.Errorf("Expected delta around %s, got %s", expectedDelta, webhook.Delta)
	}

	if _, err := time.Parse(time.RFC3339, webhook.ReceivedAt); err != nil {
		t.Errorf("ReceivedAt is not in RFC3339 format: %v", err)
	}
}

func TestWebhookHandlerConcurrency(t *testing.T) {
	tmpDir := t.TempDir()

	var wg sync.WaitGroup
	numRequests := 10
	errors := make(chan error, numRequests)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Add a random offset to the occurred_at timestamp to prevent multiple files getting created with the same name.
			// This is unlikely to ever happen in a realworld scenario
			randomOffset := r.Int63n(1000)
			occuredAt := time.Now().Unix() + randomOffset

			payload := ChargebeeEvent{
				ID:            "ev_test_concurrent_" + strconv.Itoa(i),
				OccurredAt:    occuredAt,
				Source:        "source_example",
				User:          "user_example",
				Object:        "object_example",
				APIVersion:    "v1",
				Content:       map[string]interface{}{"test": "data"},
				EventType:     "subscription_created",
				WebhookStatus: "active",
				Webhooks: []struct {
					ID            string `json:"id"`
					WebhookStatus string `json:"webhook_status"`
					Object        string `json:"object"`
				}{
					{ID: "webhook_id_example", WebhookStatus: "active", Object: "object_example"},
				},
			}

			body, err := json.Marshal(payload)
			if err != nil {
				errors <- fmt.Errorf("Failed to marshal test payload: %v", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			webhookHandler(w, req, tmpDir)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	files, err := filepath.Glob(filepath.Join(tmpDir, "*"))
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}

	if len(files) != numRequests {
		t.Errorf("Expected %d log files, got %d", numRequests, len(files))
	}
}
