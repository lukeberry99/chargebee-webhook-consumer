package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type cloudflareTunnel struct {
	token     string
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	tunnelURL string
}

func NewCloudflare(token string) Tunnel {
	ctx, cancel := context.WithCancel(context.Background())
	return &cloudflareTunnel{
		token:  token,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *cloudflareTunnel) Start() (string, error) {
	cmd := exec.CommandContext(c.ctx, "cloudflared", "tunnel", "--url", "http://localhost:8080")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start cloudflared: %v", err)
	}

	c.cmd = cmd

	urlChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "https://") {
				fields := strings.Fields(line)
				for _, field := range fields {
					if strings.HasPrefix(field, "https://") {
						c.tunnelURL = field
						urlChan <- field
						return
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading cloudflared output: %v", err)
		}
	}()

	select {
	case url := <-urlChan:
		return url, nil
	case err := <-errChan:
		return "", err
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for tunnel URL")
	case <-c.ctx.Done():
		return "", fmt.Errorf("context cancelled while waiting for tunnel")
	}
}

func (c *cloudflareTunnel) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill cloudflared process: %v", err)
		}
		if err := c.cmd.Wait(); err != nil {
			if !strings.Contains(err.Error(), "signal: killed") {
				return fmt.Errorf("error waiting for cloudflared to stop: %v", err)
			}
		}
	}
	return nil
}
