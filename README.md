# 🎯 Webhook Consumer

A lightweight command-line tool for locally testing and debugging webhooks during development. When you run this tool, it creates a temporary public URL that receives webhook events and saves them locally for inspection - perfect for development and debugging without deploying to a server.

## ⚡ What Does It Do?

When you start the tool, it:

1. Creates a local server on your machine
2. Establishes a secure tunnel using ngrok to make your local server publicly accessible
3. Provides you with a temporary URL to configure in a third party
4. Saves all received webhooks as JSON files for easy inspection

Each webhook is saved with timing information so you can analyze:

- When the event was sent
- When you received it
- The time difference between these points

## 📦 Installation

You can install the tool in several ways:

Using Go:

```bash
go install github.com/lukeberry99/webhook-consumer@latest
```

Using Homebrew (macOS and Linux):

```bash
brew install lukeberry99/tap/webhook-consumer
```

## ✨ Prerequisites

- ngrok - Make sure you have:
  1. Installed ngrok (`brew install ngrok` or download from ngrok.com)
  2. Created an ngrok account
  3. Authenticated your ngrok installation

## 🚀 Using the Tool

1. Start receiving webhooks:

```bash
wc
```

2. The tool will display a temporary URL, something like:

```
Ngrok URL: https://a1b2c3d4.ngrok.io
```

3. Configure this URL in your third party:

4. The tool will now:
   - Receive webhooks at this URL
   - Save each webhook as a JSON file in the `logs` directory
   - Name files as `{timestamp}_{event_type}.json`

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

1. Check that ngrok is properly installed and authenticated
2. Ensure port 8080 is available on your machine
3. Verify that port 4040 (used by ngrok's API) isn't in use
4. Make sure you have write permissions in the directory where you're running the tool

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
