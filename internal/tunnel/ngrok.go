package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type ngrokTunnel struct {
	cmd       *exec.Cmd
	authToken string
}

func NewNgrok(authToken string) Tunnel {
	return &ngrokTunnel{
		authToken: authToken,
	}
}

func (n *ngrokTunnel) Start() (string, error) {
	args := []string{"http", "8080"}
	if n.authToken != "" {
		args = append([]string{"authtoken", n.authToken}, args...)
	}

	n.cmd = exec.Command("ngrok", args...)
	if err := n.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ngrok: %v", err)
	}

	time.Sleep(2 * time.Second)

	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return "", fmt.Errorf("failed to query ngrok API: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
		} `json:"tunnels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ngrok API response: %v", err)
	}

	for _, tunnel := range result.Tunnels {
		if tunnel.PublicURL != "" {
			return tunnel.PublicURL, nil
		}
	}

	return "", fmt.Errorf("ngrok URL not found")
}

func (n *ngrokTunnel) Stop() error {
	if n.cmd != nil && n.cmd.Process != nil {
		return n.cmd.Process.Kill()
	}
	return nil
}
