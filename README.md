# 🎯 whook

> This tool is updating quickly, and the readme isn't always up to date.

A lightweight command-line tool for locally testing and debugging webhooks
during development. When you run this tool, it creates a temporary public URL
that receives webhook events and saves them locally for inspection - perfect
for development and debugging without deploying to a server.

## ⚡ What Does It Do?

When you start the tool, it:

1. Creates a local server on your machine
2. Establishes a secure tunnel using ngrok or cloudflare tunnels, or runs
   locally
3. Provides you with a temporary URL to configure in a third party
4. Saves all received webhooks as JSON files for easy inspection
5. Presents an interactive TUI for browsing and inspecting webhooks

Each webhook is saved with timing information of when the webhook was sent.

## 📦 Installation

You can install the tool in several ways:

Using Go:

```bash
go install github.com/lukeberry99/whook@latest
```

Using Homebrew (macOS and Linux):

```bash
brew install lukeberry99/tap/whook
```

## ✨ Prerequisites

For ngrok:

- Install ngrok (brew install ngrok or download from [ngrok.com](https://ngrok.com)
- Create an ngrok account
- Authenticate your ngrok installation

For Cloudflare:

- Install cloudflared (brew install cloudflared or download from [developers.cloudflare.com](https://developers.cloudflare.com)
- Create a Cloudflare account
- Generate an API token

## 🔧 Configuration

It looks for configuration files in the following location: `$HOME/.config/whook/config.yaml`

Example configuration file:

```yaml
server:
  port: 8080 # Default 8080
storage:
  path: "./logs" # Default ./logs
tunnel:
  driver: "ngrok" # Options: "ngrok", "cloudflare", "local" Default: "local"
  cloudflare_token: "your-token-here" # Only needed for cloudflare
services:
  chargebee:
    event_type_source: "json"
    event_type_location: "event_type"
```

Default configuration values:

- Server port: 8080
- Storage path:

Default configuration values:

- Server port: 8080
- Storage path: ./logs
- Tunnel driver: local

## 🚀 Usage

1. Start receiving webhooks:

```bash
wc
```

2. It will display a temporary URL

```bash
Ngrok URL: https://a1b2c3d4.ngrok.io
# or
Cloudflare URL: https://your-tunnel.trycloudflare.com
# or
Local URL: http://localhost:8080
```

3. Configure this URL in your third party service

4. It tool will now:

- Receive webhooks at this URL
- Save each webhook as a JSON file in the `logs` directory
- Display webhooks in the TUI
- Allow you to browse and inspect webhooks using keyboard navigation

## 🎮 Terminal UI Controls

- `↑`/`↓` or `j`/`k`: Navigate through webhooks
- `Tab`: Switch between webhook list and details panel
- `Enter`: View webhook details
- `e`: Open the current webhook in your `$EDITOR`
- `Esc`: Quit the application

## 📝 Understanding the Saved Webhooks

Each webhook is saved as a JSON file containing:

```jsonc
{
  "received_at": "2024-01-09T15:04:05Z",
  "event": {}, // The raw event payload
}
```

## 🔍 Troubleshooting

If you're having issues:

1. Check that your chosen tunnel provider (ngrok/cloudflared) is properly installed and authenticated
2. Ensure your configured port is available on your machine
3. Verify that port 4040 (used by ngrok's API) isn't in use when using ngrok
4. Make sure you have write permissions in the logs directory
5. Check your configuration file is properly formatted and in a valid location

## 🤝 Contributing

If you'd like to contribute to the project:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes
4. Push to your fork
5. Open a Pull Request

Make sure to:

- Follow existing code style
- Add tests for new features
- Update documentation as needed

## 💡 Need Help?

- File an issue on GitHub if you find a bug
- Star the repository if you find it useful
- Pull requests are welcome!
