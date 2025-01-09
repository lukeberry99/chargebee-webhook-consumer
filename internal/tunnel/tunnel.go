package tunnel

import "fmt"

type Provider string

const (
	ProviderNgrok      Provider = "ngrok"
	ProviderCloudflare Provider = "cloudflare"
)

type Tunnel interface {
	Start() (string, error)
	Stop() error
}

type Config struct {
	Provider        Provider
	CloudflareToken string
	NgrokAuthToken  string
}

func New(config Config) (Tunnel, error) {
	switch config.Provider {
	case ProviderNgrok:
		return NewNgrok(config.NgrokAuthToken), nil
	case ProviderCloudflare:
		return NewCloudflare(config.CloudflareToken), nil
	default:
		return nil, fmt.Errorf("unsupported tunnel provider: %s", config.Provider)
	}
}
