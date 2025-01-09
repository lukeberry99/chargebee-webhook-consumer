package ngrok

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/lukeberry99/chargebee-webhook-consumer/internal/models"
)

func Start() (string, error) {
	cmd := exec.Command("ngrok", "http", "8080")
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ngrok: %v", err)
	}

	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return "", fmt.Errorf("failed to query ngrok API: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Tunnels []models.NgrokTunnel `json:"tunnels"`
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
