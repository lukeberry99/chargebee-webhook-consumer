package models

import "time"

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
	ReceivedAt string         `json:"received_at"`
	Delta      string         `json:"delta"`
	Event      ChargebeeEvent `json:"event"`
}
