# go-chatapp

A real-time chat application in Go that supports multiple clients connecting and chatting simultaneously using WebSockets.

## Features

- **Real-time Communication**: Built with Gorilla WebSocket for bi-directional communication
- **Multiple Concurrent Clients**: Supports unlimited simultaneous connections
- **Clean Architecture**: Well-structured Go code with separate concerns (Hub, Client, Server)
- **Automatic Reconnection**: Client UI automatically reconnects if connection is lost
- **Modern Web Interface**: Beautiful, responsive HTML/CSS/JavaScript client
- **Heartbeat Mechanism**: Ping/Pong to detect and handle dead connections

## Architecture

The application consists of three main components:

1. **Hub (`hub.go`)**: Manages all connected clients and broadcasts messages
2. **Client (`client.go`)**: Handles individual WebSocket connections with read/write pumps
3. **Server (`main.go`)**: HTTP server and WebSocket upgrade handler

## Getting Started

### Prerequisites

- Go 1.19 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/cloneyjay/go-chatapp.git
cd go-chatapp

# Install dependencies
go mod download

# Build the application
go build
```

### Running the Application

```bash
# Run on default port (8080)
./go-chatapp

# Run on a custom port
./go-chatapp -addr :9000
```

Then open your web browser and navigate to:
- Default: http://localhost:8080
- Custom port: http://localhost:9000

### Testing with Multiple Clients

Open multiple browser tabs or windows to simulate multiple chat clients. Messages sent by any client will be broadcast to all connected clients in real-time.

## Project Structure

```
.
├── client.go      # WebSocket client connection handler
├── hub.go         # Client manager and message broadcaster
├── main.go        # HTTP server and entry point
├── home.html      # Web-based chat client UI
├── go.mod         # Go module dependencies
└── README.md      # This file
```

## How It Works

1. **Client Connection**: When a user opens the web page, JavaScript establishes a WebSocket connection to `/ws`
2. **Registration**: The server creates a new Client and registers it with the Hub
3. **Message Broadcasting**: When a client sends a message, it's read by the client's `readPump()` and sent to the Hub's broadcast channel
4. **Distribution**: The Hub distributes the message to all registered clients through their `send` channels
5. **Delivery**: Each client's `writePump()` reads from its `send` channel and writes to the WebSocket connection

## WebSocket Protocol

The application uses the Gorilla WebSocket library with the following configuration:

- **Read Buffer**: 1024 bytes
- **Write Buffer**: 1024 bytes
- **Max Message Size**: 512 bytes
- **Pong Wait**: 60 seconds
- **Ping Period**: 54 seconds
- **Write Wait**: 10 seconds

## License

This project is licensed under the MIT License - see the LICENSE file for details.

