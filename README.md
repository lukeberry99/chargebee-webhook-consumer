# Chargebee Webhook Consumer

A local development tool for receiving and debugging Chargebee webhooks. This
application creates a publicly accessible endpoint that can receive Chargebee
webhook events and saves them locally for inspection and debugging.

## Features

- **Webhook Handling:** Receives and processes Chargebee webhook events via
  HTTP POST requests
- **Public URL Access:** Automatically creates a public URL using ngrok to
  expose your local server
- **Event Logging:** Saves each webhook event as a separate JSON file in a
  `logs` directory.

  Files are named using Chargebee's event occurence timestamp
  in the format `{chargebee_event_timestamp}_{event_type}.json`

- **Timestamp Tracking:** Each logged event includes two timestamps:
  - The original event occurence time from Chargbee (`event.occured_at`)
  - The local time when our server received the webhook (`received_at`)

## Prerequisites

- **Go:** Version 1.23.4 or higher
- **Ngrok:** Install Ngrok and ensure that you have created an account and
  authenticated.

## Installation

1. Clone the repository:

```bash
git clone git@github.com:lukeberry99/chargebee-webhook-consumer.git
cd chargebee-webhook-consumer
```

2. Install dependencies:

```bash
go mod tidy
```

## Usage

1. **Start the Application**:

   Run the application using the following command:

   ```bash
   go run main.go
   ```

   This will:

   - Start a local server on port 8080
   - Create a public URL using ngrok
   - Display the public URL for configuring in Chargebee

2. **Configure Chargebee**:

   - Go to your Chargebee dashboard
   - Navigate to Settings → Configure Chargebee → Webhooks
   - Add a new webhook endpoint using the displayed public URL
   - Select the events you want to receive

3. **Monitor Webhooks**:

   The application will:

   - Receive webhook events at the root endpoint (/)
   - Save each event as a JSON file in the `logs` directory
   - Use the naming format: `{timestamp}_{event_type}.json`
   - Display a console message for each received webhook

## Webhook Data Structure

Each logged webhook file contains:

```json
{
  "received_at": "2023-XX-XX:XX:XX:XXZ", // Local receipt time
  "event": {
    "id": "ev_xxx",
    "occurred_at": 1234567890,
    "event_type": "event_name",
    "content": {} // ...additional Chargebee event data
  }
}
```

## Troubleshooting

- **Port 8080 In Use**: Ensure no other service is using port 8080
- **Connection Issues**: Check your internet connection and firewall settings
- **Missing Logs**: Ensure the application has write permissions in the current
  directory
- **Ngrok Issues**: Ensure Ngrok is installed and the binary is in your
  system's PATH. Make sure that you have created an ngrok account and set up your
  local environment with the auth key. Verify that no other application is using
  port 4040, which Ngrok uses for its API.
- **Port Conflicts**: Ensure no other service is running on port 8080.
