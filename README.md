# Chargebee Webhook Consumer

## Features

- **Webhook Handling:** Receives and processes Chargebee webhook events.
- **Ngrok Integration:** Automatically sets up a public URL using Ngrok to expose the local server.
- **Logging:** Saves webhook data to JSON files in a `logs` directory.

## Prerequisites

- **Go:** Ensure you have Go version 1.23.x installed.
- **Ngrok:** Install Ngrok and ensure that you have created an account and authenticated.

## Installation

1. Clone the repository:

   ```bash
   git clone git@github.com:lukeberry99/chargebee-webhook-consumer
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

   This will start the web server on port 8080 and Ngrok will create a public URL to access it.

2. **Ngrok URL**:

   After starting the application, you will see an Ngrok URL in the terminal output. Use this URL to configure Chargebee to send webhooks to your local server.

3. **Receive Webhooks**:

   The application listens for incoming webhooks on the `/` endpoint. Webhook data is logged in the `logs` directory, with each event saved as a separate JSON file.

## Configuration

- **Ngrok Configuration**: You can customize Ngrok settings by modifying the `startNgrok` function if needed, such as adding authentication tokens for paid accounts.

## Troubleshooting

- **Ngrok Issues**: Ensure Ngrok is installed and the binary is in your system's PATH. Verify that no other application is using port 4040, which Ngrok uses for its API.
- **Port Conflicts**: Ensure no other service is running on port 8080.
